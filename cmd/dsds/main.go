// dsds 是《大時代的故事》remake 的執行檔。
//
// 目前做到的：載入原版資料，以 retro 或 P1 modern 地形／鐵路主題畫出各省的
// 14×14 戰場；側欄顯示省份狀態，可切換省份、打仗、下已解出的政略指令、存檔。
//
// **只接規則層已 confirmed 的指令**（運補／徵稅／開發／秘密行動的學潮／
// 商業／慰勞／攻打／外交窄切片／第 10 項停火窄切片）。完整外交與所有原版
// 玩法仍未宣告等價，未閉合的路徑保持 fail-closed——
// 那勝過假裝有效果（`cmd/dsds/strategy.go` 開頭）。
//
//	tools/go.sh run ./cmd/dsds -game workplace/orig/game
//	tools/go.sh run ./cmd/dsds -game workplace/orig/game -audio retro
//	tools/go.sh run ./cmd/dsds -game workplace/orig/game -audio modern -modern-audio assets/music/modern
//
// 操作（AGENTS.md §8：**ESC 只取消／退回上一層，F10 才離開**）：
//
//	← →      切換省份
//	Enter    叫出政略指令選單
//	ESC      關掉選單／退回上一層／取消離開
//	F10      離開，跳 Y／N 確認並自動存檔；存檔失敗就不離開
//
// 指令選單裡：
//
//	1        調動 → 部份／全部 → 目標 → 選將 → 四種物資 → 確認
//	4        徵稅（每月限一次）
//	7        開發 → 1 墾地　2 建兵工廠　3 挖金礦
//	6        查閱本省將領；上下選人、Enter 看詳細、左右切換
//	B        在將領清單／詳細頁開啟人物自傳；Space／PgUp／PgDn 翻頁
//	O        開啟 remake 顯示設定（目前可切原典用語／現代白話）
//	T        商業活動 → 進口／出口 → 品項 → 輸入數量
//	S        運補 → 目標鄰省 → 黃金／糧食／彈藥／燃料數量
//	R        練兵 → Y 確認
//	V        秘密行動 → 2 鼓動學潮 → 目標省編號
//	C        慰勞軍民（原版是指令 14）
//	A        對第一個可攻打的鄰省開戰
//	E        結束回合，推進一個月；跨年跑年度結算
//
// 指令結果會顯示在畫面訊息列，並保留 stderr 診斷；面板數值即時更新。
//
// 需要顯示器。無頭環境請跑 internal/ui/render 的測試，
// 那一層不依賴 Ebiten，會逐像素比對原版截圖。
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	userprefs "github.com/wicanr2/great-era-remake/internal/prefs"
	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	uiaudio "github.com/wicanr2/great-era-remake/internal/ui/audio"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
	uiresolution "github.com/wicanr2/great-era-remake/internal/ui/resolution"
	"github.com/wicanr2/great-era-remake/internal/ui/textlayout"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

// scale 是視窗放大倍率。
//
// remake 差異（AGENTS.md §2「外殼允許現代化」）：只放大視窗，
// 邏輯解析度仍是原版的 640×350。
const scale = 2

// 版面。原版政略畫面是左側面板（約 190 寬）+ 右側地圖。
//
// 戰場的實際尺寸是 **448×348**（14 欄 × 32、13 列 × 24 + 半格 + 24，
// `docs/re/07` §3），高度只差 2 px 就頂到 BGI 640×350 的底部
// ——所以 y 必須是 0，不能留上邊距，否則最下面那一列會被切掉。
// x 是 remake 的排版選擇（原版戰場畫面從 x=0 起，面板在右側）。
const (
	fieldX, fieldY = 190, 0
)

// 配色取自實機截圖的實際像素值（面板的暗紅字與米黃底）。
var (
	panelInk   = assets.RGB{R: 0xAE, G: 0x00, B: 0x00}
	panelPaper = assets.RGB{R: 0xFF, G: 0xFF, B: 0xA2}
)

// screen 是介面的狀態。ESC 一律退回上一層，不會直接離開。
type screen int

const (
	screenMap                screen = iota // 戰場 + 省份面板
	screenCommand                          // 政略指令選單
	screenBattle                           // 戰鬥（見 battle.go）
	screenDevelop                          // 開發的三個子項（見 strategy.go）
	screenTransferMode                     // 調動：部份／全部
	screenTransferTarget                   // 調動：輸入目標省
	screenTransferSelection                // 調動：勾選將領
	screenTransferAmount                   // 調動：四項物資
	screenTransferConfirm                  // 調動：最後確認
	screenTradeMode                        // 商業活動：進口／出口
	screenTradeGood                        // 商業活動：選品項
	screenTradeAmount                      // 商業活動：輸入數量
	screenSupplyTarget                     // 運補：輸入目標省編號
	screenSupplyAmount                     // 運補：依序輸入四項資源
	screenRecruitAction                    // 徵兵／重新整編
	screenRecruitBranch                    // 徵兵：選兵種
	screenRecruitAmount                    // 徵兵：輸入人數
	screenRecruitConfirm                   // 徵兵：確認成本
	screenReorganizeBranch                 // 重新整編：選兵種
	screenReorganizeTarget                 // 重新整編：選將領
	screenReorganizeAmount                 // 重新整編：指派兵力
	screenTrainConfirm                     // 練兵：確認
	screenCovertAction                     // 秘密行動：游擊隊／學潮
	screenCovertTarget                     // 秘密行動：輸入目標省
	screenViewMenu                         // 查閱：四項原版選單
	screenViewProvinceSelect               // 查閱他省：輸入省編號
	screenViewProvinceChoice               // 查閱他省：概況／將領
	screenViewProvince                     // 查閱他省：省份詳細資料
	screenViewOverview                     // 查閱所屬各省
	screenViewProvinceNames                // 查閱省名：兩頁
	screenViewGenerals                     // 查閱：本省將領清單
	screenViewGeneral                      // 查閱：將領詳細狀態
	screenBiography                        // remake 新增：人物自傳全頁
	screenNarrative                        // remake 新增：原版新聞／史事圖庫
	screenPolicy                           // 指令 8：授權自治／產能分配
	screenDiplomacy                        // 指令 9：貸款／外援／償還外債
	screenCeasefireTarget                  // 指令 10：輸入談判停火目標省
	screenLoanAmount                       // 外交：輸入貸款額度
	screenRepayAmount                      // 外交：輸入償還外債額度
	screenAutonomy                         // 政策：切換其他省份自治狀態
	screenProduction                       // 政策：調整本省產能分配
	screenOtherOptions                     // 指令 15：原版八項 + remake 顯示設定
	screenSaveConfirm                      // 其他選項 1：儲存確認
	screenLoadConfirm                      // 其他選項 2：載入確認
	screenMessageTime                      // 其他選項 6：訊息停留時間 1..10
	screenDisplayOptions                   // remake 新增：顯示設定（由指令 15 區域進入）
	screenResolutionOptions                // remake 新增：原版／高解析畫布
	screenQuit                             // 離開確認
)

type app struct {
	m                *game.Map
	tbl              *game.ProvinceTable // 39 省的狀態（存檔或初始檔）
	generals         []game.General      // 該期的將領表
	fonts            render.PanelFonts   // 面板用的三個字模檔
	tiles            *render.TileSet     // NEWTERR + RAIL 的圖塊
	origSave         []byte              // 原始存檔內容，寫回時當基底
	cmdFonts         render.CommandFonts
	fan              *assets.GlyphFile       // FAN(1).15，部隊番號字模
	icons            []*assets.Image         // NEWICON.TPC 的兵種圖示
	battlefieldTheme uitheme.Theme           // 戰場主題；retro／modern 共用同一套索引
	unitTheme        uitheme.UnitProvider    // 與戰場主題成組切換的部隊圖示
	hudIcons         uitheme.HUDIconProvider // modern 資源／指令輔助圖示
	audio            *uiaudio.Manager        // 開局 BGM；off 模式為 nil
	audioMode        uiaudio.Mode
	audioTrack       uiaudio.Track
	audioWarned      map[uiaudio.Track]bool
	// battleDT2／battleMemWar 是原始戰鬥狀態的副本；只在戰鬥結束時以
	// internal/game 的非破壞性 writer 寫到 savePath 同目錄，絕不碰原版素材。
	battleDT2         []byte
	battleMemWar      []byte
	battleDT2States   [game.ProvinceCount]game.BattleState
	battleMemStates   [game.ProvinceCount]game.BattleState
	battleDT2Path     string
	battleMemWarPath  string
	retroTheme        uitheme.Theme // 已載入的原版主題，供 F2 原子切換
	modernTheme       uitheme.Theme // 已載入的 P1 modern 地形／鐵路主題
	themeMode         uitheme.Mode
	resolution        uiresolution.Mode
	battle            *battleState                           // 非 nil 表示正在打仗
	world             *game.AIWorld                          // 規則層：政略指令都經過它
	factions          game.FactionTable                      // 存檔中的勢力表；外交帳本索引需由它反查
	factionLeaders    game.FactionLeaders                    // `.DT1` 區塊 6；空槽 0 也要原樣保存
	factionOf         game.FactionOfGeneral                  // `.DT1` 區塊 7；modern 色帶只讀此反查
	warRecords        [game.ProvinceCount + 1]game.WarRecord // `.DT1` 區塊 2 已知欄位快照
	majorPowerLeaders game.MajorPowerLeaders                 // `.DT1` 區塊 10 的 runtime 快照
	loc               *i18n.Locale                           // 語系表：省名與 UI 詞彙（nil 表示沒載到）
	people            *i18n.PeopleDB                         // 人物自傳：(期別, 將領槽位) → 語系人物資料
	portraits         *i18n.PortraitCatalog                  // Modern 已驗證離線肖像；缺登錄顯示檔案卡
	narrative         *i18n.NarrativeCatalog                 // NEWSDATA.DAT 的已證實模板與來源圖像索引
	narrativeImages   []*assets.Image                        // 17 張原版新聞圖像；唯讀，不改寫
	narrativePage     int
	narrativeBack     screen
	eten              *assets.EtenFonts // 完整繁中字庫；載不到時停用自傳入口
	wording           *i18n.WordingCatalog
	wordingMode       i18n.WordingMode
	preferences       userprefs.Preferences
	prefsPath         string
	stage             int
	cmdBudget         *game.CommandBudget // 每省這個月剩餘的指令數（docs/re/13 §2）
	rng               *game.Rand          // 原版的 LCG（docs/re/17），固定種子才可重現
	year              uint16
	month             uint8
	msg               string // 上一個指令的結果，印到 stderr
	messages          *messageQueue
	messageTimeInput  uint32
	pointer           pointerTracker
	pointerAction     actions.Action
	transferTargets   []game.ProvinceID
	transferMode      game.PlayerTransferMode
	transferSession   *game.PlayerTransferSelection
	transferInput     uint32
	transferCursor    int
	transferAmounts   [4]int
	transferGood      int
	tradeImport       bool
	tradeGood         game.TradeGood
	tradeAmount       uint32
	supplyTargets     []game.ProvinceID
	supplyTarget      game.ProvinceID
	supplyAmounts     [4]int
	supplyGood        int
	supplyInput       uint32
	recruitBranch     uint8
	recruitLimit      int
	recruitAmount     uint32
	reorganization    *game.Reorganization
	reorganizeID      game.GeneralID
	reorganizeInput   uint32
	viewGenerals      []game.GeneralID
	viewIndex         int
	viewGeneralBack   screen
	bioBack           screen
	bioPage           int
	bioPages          int
	viewProvince      game.ProvinceID
	viewInput         uint32
	viewPage          int
	covertAction      int
	covertInput       uint32
	diplomacySlot     int
	diplomacyInput    uint32
	ceasefireInput    uint32
	autonomyTargets   []game.ProvinceID
	autonomyInput     uint32
	autonomySpent     bool
	productionItem    int
	productionInput   uint32
	productionSpent   bool
	provinceLimit     int
	current           game.ProvinceID
	playerCommander   game.GeneralID // 載入時用來重新找到玩家所屬的合法省份
	ledger            game.DiplomacyLedger
	screen            screen
	quitBack          screen // 進入離開確認前的畫面；N／ESC 必須回到原處
	savePath          string // 離開時自動存檔的目標（不覆蓋原版）
	saveErr           error  // 存檔失敗就不離開
	dirty             bool
	frame             *ebiten.Image
}

// desiredAudioTrack 把畫面情境映射到可選的曲目。它只描述外殼情境，
// 不改變任何遊戲規則；缺少玩家自備檔案時由 syncAudioForScreen 保留目前曲目。
func (a *app) desiredAudioTrack() uiaudio.Track {
	if a == nil || a.audio == nil {
		return ""
	}
	if a.screen == screenBattle {
		for _, track := range []uiaudio.Track{uiaudio.TrackBattle1, uiaudio.TrackBattle2, uiaudio.TrackBattleAlt} {
			if a.audio.HasTrack(track) {
				return track
			}
		}
		return uiaudio.TrackBattle1
	}
	if a.screen == screenBiography || a.screen == screenViewGeneral ||
		a.screen == screenViewGenerals {
		if a.audio.HasTrack(uiaudio.TrackWall) {
			return uiaudio.TrackWall
		}
	}
	if a.screen == screenNarrative {
		if a.audio.HasTrack(uiaudio.TrackScene) {
			return uiaudio.TrackScene
		}
	}
	if a.screen == screenQuit {
		if a.audio.HasTrack(uiaudio.TrackFinal) {
			return uiaudio.TrackFinal
		}
	}
	if a.screen == screenMap || a.screen == screenCommand {
		if a.audio.HasTrack(uiaudio.TrackMainTheme) {
			return uiaudio.TrackMainTheme
		}
	}
	if a.audio.HasTrack(uiaudio.TrackStrategy) {
		return uiaudio.TrackStrategy
	}
	if a.audio.HasTrack(uiaudio.TrackScene) {
		return uiaudio.TrackScene
	}
	return uiaudio.TrackStrategy
}

// syncAudioForScreen 在每幀入口檢查情境切換，讓所有鍵盤／滑鼠／觸控
// 路徑都共用同一個音訊切換點。缺少 optional 曲目只警告一次，不阻斷遊戲。
func (a *app) syncAudioForScreen() {
	if a == nil || a.audio == nil || (a.audioMode != uiaudio.ModeRetro && a.audioMode != uiaudio.ModeModern) {
		return
	}
	a.audio.Update()
	if a.audioWarned == nil {
		a.audioWarned = make(map[uiaudio.Track]bool)
	}
	want := a.desiredAudioTrack()
	if want == "" || want == a.audioTrack || !a.audio.HasTrack(want) {
		if want != "" && want != a.audioTrack && !a.audio.HasTrack(want) && !a.audioWarned[want] {
			fmt.Fprintf(os.Stderr, "音訊曲目 %q 未提供，沿用 %q\n", want, a.audioTrack)
			a.audioWarned[want] = true
		}
		return
	}
	// 18 UI ticks 約 300 ms；這是 remake 外殼的聽感選擇，不宣稱原版
	// AdLib 驅動具備同一段淡入淡出。
	if err := a.audio.PlayTrackWithFade(want, 0.7, 18); err != nil {
		if !a.audioWarned[want] {
			fmt.Fprintf(os.Stderr, "音訊曲目 %q 切換失敗，沿用 %q：%v\n", want, a.audioTrack, err)
			a.audioWarned[want] = true
		}
		return
	}
	a.audioTrack = want
}

func digitKeys() []ebiten.Key {
	return []ebiten.Key{ebiten.KeyDigit0, ebiten.KeyDigit1, ebiten.KeyDigit2,
		ebiten.KeyDigit3, ebiten.KeyDigit4, ebiten.KeyDigit5, ebiten.KeyDigit6,
		ebiten.KeyDigit7, ebiten.KeyDigit8, ebiten.KeyDigit9}
}

func enterPressed() bool {
	return inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeyKPEnter)
}

func provinceIn(ids []game.ProvinceID, want game.ProvinceID) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

func (a *app) Update() error {
	a.syncAudioForScreen()
	// 原版結果訊息會執行 DELAY(byte_6FE85×400 ms)，期間不接受指令。
	// remake 不阻塞繪圖執行緒，但在對應 Update ticks 內同樣擋住輸入。
	if a.messages != nil && a.messages.Active() {
		a.pointer.cancel() // 等待期間的按下不得在訊息消失後變成點擊。
		a.pointerAction = actions.None
		if a.messages.Tick() {
			a.dirty = true
		}
		return nil
	}
	a.pointerAction = a.collectPointerAction()
	if a.pointerAction != actions.None {
		a.playInputEffect(a.pointerAction)
	}
	// F10 是唯一的離開鍵，而且要先確認（AGENTS.md §8）。
	if inpututil.IsKeyJustPressed(ebiten.KeyF10) && a.screen != screenQuit {
		a.quitBack, a.screen, a.dirty = a.screen, screenQuit, true
		return nil
	}
	// F2 是 remake 外殼捷徑：只交換已完整建好的 Theme 指標，不碰規則層、
	// 亂數或存檔 bytes；偏好另存於 prefs.json。
	if inpututil.IsKeyJustPressed(ebiten.KeyF2) {
		if err := a.toggleTheme(); err != nil {
			a.report("主題切換失敗：" + err.Error())
		}
		return nil
	}
	// F3 是跨桌面的解析度捷徑：原版 640×350 ↔ Modern 1280×720。
	// 它只改呈現／輸入座標，不碰規則、亂數或存檔 bytes；觸控裝置則
	// 由「其他選項 → 解析度」進入同一個 setter。
	if inpututil.IsKeyJustPressed(ebiten.KeyF3) {
		if err := a.toggleResolution(); err != nil {
			a.report("解析度切換失敗：" + err.Error())
		}
		return nil
	}

	switch a.screen {
	case screenQuit:
		switch {
		case a.actionPressed(actions.Confirm, ebiten.KeyY):
			// 存檔失敗就不離開。
			if err := a.autosave(); err != nil {
				a.saveErr, a.screen, a.dirty = err, a.quitBack, true
				fmt.Fprintln(os.Stderr, "自動存檔失敗，不離開:", err)
				return nil
			}
			return ebiten.Termination
		case a.actionPressed(actions.Cancel, ebiten.KeyN, ebiten.KeyEscape):
			a.screen, a.dirty = a.quitBack, true
		}
		return nil

	case screenBattle:
		return a.updateBattle()

	case screenCommand:
		// ESC 只退回上一層。
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenMap, true
			return nil
		}
		// A：對第一個可攻打的鄰省開戰。攻打候選的規則見
		// `docs/spec/01` §2（鄰省 − 自己控制的省）。
		if a.actionPressed(actions.Select2, ebiten.KeyA) {
			target := a.tbl.FirstAttackable(a.current)
			if target == 0 {
				a.report("目前沒有可攻打的鄰省")
				return nil
			}
			// 先驗證能否建立戰鬥，成功後才消耗指令數。這遵守「做成事才扣」
			// 的實機契約（`docs/playtest/02` §282）；失敗不能留下隱形扣款。
			if a.cmdBudget.Remaining(a.current) <= 0 {
				a.report(fmt.Sprintf("%s 這個月的指令數用完了",
					a.provinceName(a.current)))
				return nil
			}
			if err := a.startBattle(target, a.current); err != nil {
				a.report("開戰失敗：" + err.Error())
				return nil
			}
			// 單執行緒輸入下這裡不會失敗；保留防線以免未來改成非同步
			// 建戰鬥後出現狀態不一致。
			if !a.cmdBudget.Spend(a.current) {
				a.battle = nil
				a.screen, a.dirty = screenCommand, true
				a.report(fmt.Sprintf("%s 這個月的指令數用完了",
					a.provinceName(a.current)))
			}
			return nil
		}
		// 已接上的政略指令。**只接規則層 confirmed 的**，
		// 其餘按了沒反應——那勝過假裝有效果（strategy.go 開頭）。
		switch {
		case a.actionPressed(actions.Select4, ebiten.KeyDigit4):
			a.report(a.withBudget(a.current, a.execTax))
		case a.actionPressed(actions.Select5, ebiten.KeyDigit5):
			a.screen, a.dirty = screenRecruitAction, true
		case a.actionPressed(actions.Select6, ebiten.KeyDigit6):
			a.screen, a.dirty = screenViewMenu, true
		case a.actionPressed(actions.Select7, ebiten.KeyDigit7):
			a.screen, a.dirty = screenDevelop, true
		case a.actionPressed(actions.Select8, ebiten.KeyDigit8):
			if a.wording == nil || a.eten == nil {
				a.report("政策畫面需要完整語系資料與倚天字庫")
			} else if a.cmdBudget.Remaining(a.current) <= 0 {
				a.report(fmt.Sprintf("%s 這個月的指令數用完了", a.provinceName(a.current)))
			} else {
				a.screen, a.dirty = screenPolicy, true
			}
		case a.actionPressed(actions.Select9, ebiten.KeyDigit9):
			if a.wording == nil || a.eten == nil {
				a.report("外交畫面需要完整語系資料與倚天字庫")
			} else if a.cmdBudget.Remaining(a.current) <= 0 {
				a.report(fmt.Sprintf("%s 這個月的指令數用完了", a.provinceName(a.current)))
			} else if _, err := a.currentDiplomacySlot(); err != nil {
				a.report(err.Error())
			} else {
				a.diplomacyInput = 0
				a.screen, a.dirty = screenDiplomacy, true
			}
		case a.actionPressed(actions.Select10, ebiten.KeyDigit0):
			if a.world == nil {
				a.report("停火需要有效的遊戲世界")
				return nil
			}
			if _, err := a.world.CeasefireRequester(uint8(a.stage), a.current); err != nil {
				if a.stage != int(game.CeasefireStage) {
					a.report(a.ceasefireText("ceasefire.unavailable", "停火在本期無法使用"))
				} else {
					a.report(a.ceasefireText("ceasefire.no_commander", "司令不在本省"))
				}
				return nil
			}
			a.ceasefireInput = 0
			a.screen, a.dirty = screenCeasefireTarget, true
		case a.actionPressed(actions.Select1, ebiten.KeyDigit1):
			targets, err := a.world.PlayerTransferTargets(a.current)
			if err != nil || len(targets) == 0 {
				a.report("無法調動：沒有合法鄰省")
				return nil
			}
			a.transferTargets, a.transferInput = targets, 0
			a.screen, a.dirty = screenTransferMode, true
		case a.actionPressed(actions.Select14, ebiten.KeyC):
			a.report(a.withBudget(a.current, a.execComfort))
		case a.actionPressed(actions.Select12, ebiten.KeyT):
			a.screen, a.dirty = screenTradeMode, true
		case a.actionPressed(actions.Select3, ebiten.KeyS):
			targets, err := a.world.SupplyTargets(a.current)
			if err != nil || len(targets) == 0 {
				a.report("無法運補：沒有同司令且未交戰的鄰省")
				return nil
			}
			a.supplyTargets, a.supplyInput = targets, 0
			a.screen, a.dirty = screenSupplyTarget, true
		case a.actionPressed(actions.Select13, ebiten.KeyR):
			a.screen, a.dirty = screenTrainConfirm, true
		case a.actionPressed(actions.Select11, ebiten.KeyV):
			a.screen, a.dirty = screenCovertAction, true
		case a.actionPressed(actions.Select15, ebiten.KeyO):
			if a.wording == nil || a.eten == nil {
				a.report("顯示設定需要完整語系資料與倚天字庫")
			} else {
				a.screen, a.dirty = screenOtherOptions, true
			}
		case inpututil.IsKeyJustPressed(ebiten.KeyE):
			a.report(a.endTurn())
		}
		return nil

	case screenTransferMode:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		if a.actionPressed(actions.Select1, ebiten.KeyDigit1) {
			a.transferMode = game.PlayerTransferPartial
		} else if a.actionPressed(actions.Select2, ebiten.KeyDigit2) {
			a.transferMode = game.PlayerTransferAll
		} else {
			return nil
		}
		a.transferInput = 0
		a.screen, a.dirty = screenTransferTarget, true
		return nil

	case screenTransferTarget:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenTransferMode, true
			return nil
		}
		if a.deleteDigitPressed() {
			a.transferInput /= 10
			a.dirty = true
		}
		for d, key := range digitKeys() {
			if a.digitPressed(d, key) {
				next := a.transferInput*10 + uint32(d)
				if next <= game.ProvinceCount {
					a.transferInput, a.dirty = next, true
				}
			}
		}
		direct := false
		if n, ok := actions.SelectionNumber(a.pointerAction); ok && n <= len(a.transferTargets) {
			a.transferInput, direct = uint32(a.transferTargets[n-1]), true
		}
		if a.submitPressed() || direct {
			to := game.ProvinceID(a.transferInput)
			if !provinceIn(a.transferTargets, to) {
				return nil
			}
			s, err := a.world.BeginPlayerTransfer(a.current, to, a.transferMode, a.generals)
			if err != nil {
				a.report(err.Error())
				return nil
			}
			a.transferSession, a.transferCursor = s, 0
			a.screen, a.dirty = screenTransferSelection, true
		}
		return nil

	case screenTransferSelection:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.transferSession = nil
			a.transferInput = 0
			a.screen, a.dirty = screenTransferTarget, true
			return nil
		}
		cands := a.transferSession.Candidates()
		if n, ok := actions.SelectionNumber(a.pointerAction); ok && n <= len(cands) {
			a.transferCursor = n - 1
			if err := a.transferSession.Toggle(cands[a.transferCursor]); err != nil {
				a.report(err.Error())
			} else {
				a.dirty = true
			}
			return nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) && a.transferCursor > 0 {
			a.transferCursor--
			a.dirty = true
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyDown) && a.transferCursor+1 < len(cands) {
			a.transferCursor++
			a.dirty = true
		}
		if inpututil.IsKeyJustPressed(ebiten.KeySpace) && len(cands) > 0 {
			if err := a.transferSession.Toggle(cands[a.transferCursor]); err != nil {
				a.report(err.Error())
			} else {
				a.dirty = true
			}
		}
		// Enter 是現代操作修正：DOS 版沒有提交非全選集合的可達按鍵。
		if a.actionPressed(actions.Submit, ebiten.KeyEnter, ebiten.KeyKPEnter) && len(a.transferSession.Selected()) > 0 {
			a.transferAmounts, a.transferGood, a.transferInput = [4]int{}, 0, 0
			a.screen, a.dirty = screenTransferAmount, true
		}
		return nil

	case screenTransferAmount:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.transferInput = 0
			a.screen, a.dirty = screenTransferSelection, true
			return nil
		}
		if a.deleteDigitPressed() {
			a.transferInput /= 10
			a.dirty = true
		}
		limit := a.transferSupplyLimit(a.transferGood)
		for d, key := range digitKeys() {
			if a.digitPressed(d, key) {
				next := a.transferInput*10 + uint32(d)
				if next <= uint32(limit) {
					a.transferInput, a.dirty = next, true
				}
			}
		}
		if a.submitPressed() {
			a.transferAmounts[a.transferGood] = int(a.transferInput)
			a.transferGood++
			a.transferInput = 0
			if a.transferGood == len(a.transferAmounts) {
				a.screen = screenTransferConfirm
			}
			a.dirty = true
		}
		return nil

	case screenTransferConfirm:
		if a.actionPressed(actions.Cancel, ebiten.KeyN, ebiten.KeyEscape) {
			a.transferGood = len(a.transferAmounts) - 1
			a.transferInput = uint32(a.transferAmounts[a.transferGood])
			a.screen, a.dirty = screenTransferAmount, true
		} else if a.actionPressed(actions.Confirm, ebiten.KeyY) {
			a.report(a.execPlayerTransfer())
			a.transferSession = nil
			a.screen, a.dirty = screenCommand, true
		}
		return nil

	case screenDevelop:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		for i, k := range []ebiten.Key{ebiten.KeyDigit1, ebiten.KeyDigit2, ebiten.KeyDigit3} {
			if a.actionPressed(actions.Selection(i+1), k) {
				sub := i + 1
				a.report(a.withBudget(a.current,
					func() string { return a.execDevelop(sub) }))
				a.screen = screenCommand
			}
		}
		return nil

	case screenTradeMode:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		if a.actionPressed(actions.Select1, ebiten.KeyDigit1) {
			a.tradeImport, a.screen, a.dirty = true, screenTradeGood, true
		} else if a.actionPressed(actions.Select2, ebiten.KeyDigit2) {
			a.tradeImport, a.screen, a.dirty = false, screenTradeGood, true
		}
		return nil

	case screenTradeGood:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenTradeMode, true
			return nil
		}
		var goods []game.TradeGood
		if a.tradeImport {
			goods = []game.TradeGood{game.GoodFood, game.GoodAmmo, game.GoodFuel}
		} else {
			goods = []game.TradeGood{game.GoodFood, game.GoodAmmo, game.GoodCoal, game.GoodIron, game.GoodFuel}
		}
		keys := []ebiten.Key{ebiten.KeyDigit1, ebiten.KeyDigit2, ebiten.KeyDigit3,
			ebiten.KeyDigit4, ebiten.KeyDigit5}
		for i := range goods {
			if a.actionPressed(actions.Selection(i+1), keys[i]) {
				a.tradeGood, a.tradeAmount = goods[i], 0
				a.screen, a.dirty = screenTradeAmount, true
				break
			}
		}
		return nil

	case screenTradeAmount:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenTradeGood, true
			return nil
		}
		if a.deleteDigitPressed() {
			a.tradeAmount /= 10
			a.dirty = true
		}
		for d, k := range []ebiten.Key{ebiten.KeyDigit0, ebiten.KeyDigit1, ebiten.KeyDigit2,
			ebiten.KeyDigit3, ebiten.KeyDigit4, ebiten.KeyDigit5, ebiten.KeyDigit6,
			ebiten.KeyDigit7, ebiten.KeyDigit8, ebiten.KeyDigit9} {
			if a.digitPressed(d, k) {
				next := a.tradeAmount*10 + uint32(d)
				if next <= 65535 {
					a.tradeAmount = next
					a.dirty = true
				}
			}
		}
		if a.submitPressed() {
			amount := int(a.tradeAmount)
			a.report(a.withBudget(a.current, func() string {
				return a.execTrade(a.tradeImport, a.tradeGood, amount)
			}))
			a.screen, a.dirty = screenCommand, true
		}
		return nil

	case screenSupplyTarget:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		if a.deleteDigitPressed() {
			a.supplyInput /= 10
			a.dirty = true
		}
		for d, k := range []ebiten.Key{ebiten.KeyDigit0, ebiten.KeyDigit1, ebiten.KeyDigit2,
			ebiten.KeyDigit3, ebiten.KeyDigit4, ebiten.KeyDigit5, ebiten.KeyDigit6,
			ebiten.KeyDigit7, ebiten.KeyDigit8, ebiten.KeyDigit9} {
			if a.digitPressed(d, k) {
				next := a.supplyInput*10 + uint32(d)
				if next <= game.ProvinceCount {
					a.supplyInput, a.dirty = next, true
				}
			}
		}
		direct := false
		if n, ok := actions.SelectionNumber(a.pointerAction); ok && n <= len(a.supplyTargets) {
			a.supplyInput, direct = uint32(a.supplyTargets[n-1]), true
		}
		if a.submitPressed() || direct {
			chosen := game.ProvinceID(a.supplyInput)
			for _, id := range a.supplyTargets {
				if id == chosen {
					a.supplyTarget, a.supplyAmounts = chosen, [4]int{}
					a.supplyGood, a.supplyInput = 0, 0
					a.screen, a.dirty = screenSupplyAmount, true
					break
				}
			}
		}
		return nil

	case screenSupplyAmount:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.supplyInput = 0
			a.screen, a.dirty = screenSupplyTarget, true
			return nil
		}
		if a.deleteDigitPressed() {
			a.supplyInput /= 10
			a.dirty = true
		}
		limit := a.supplyLimit(a.supplyGood)
		for d, k := range []ebiten.Key{ebiten.KeyDigit0, ebiten.KeyDigit1, ebiten.KeyDigit2,
			ebiten.KeyDigit3, ebiten.KeyDigit4, ebiten.KeyDigit5, ebiten.KeyDigit6,
			ebiten.KeyDigit7, ebiten.KeyDigit8, ebiten.KeyDigit9} {
			if a.digitPressed(d, k) {
				next := a.supplyInput*10 + uint32(d)
				if next <= uint32(limit) {
					a.supplyInput, a.dirty = next, true
				}
			}
		}
		if a.submitPressed() {
			a.supplyAmounts[a.supplyGood] = int(a.supplyInput)
			a.supplyGood++
			a.supplyInput = 0
			if a.supplyGood == len(a.supplyAmounts) {
				a.report(a.withBudget(a.current, func() string {
					return a.execSupply(a.supplyTarget, a.supplyAmounts)
				}))
				a.screen = screenCommand
			}
			a.dirty = true
		}
		return nil

	case screenRecruitAction:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenCommand, true
		} else if a.actionPressed(actions.Select1, ebiten.KeyDigit1) {
			a.screen, a.dirty = screenRecruitBranch, true
		} else if a.actionPressed(actions.Select2, ebiten.KeyDigit2) {
			if a.cmdBudget.Remaining(a.current) <= 0 {
				a.report(fmt.Sprintf("%s 這個月的指令數用完了", a.provinceName(a.current)))
				a.screen, a.dirty = screenCommand, true
			} else {
				a.screen, a.dirty = screenReorganizeBranch, true
			}
		}
		return nil

	case screenRecruitBranch:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenRecruitAction, true
			return nil
		}
		branches := game.RecruitBranchOrder
		keys := []ebiten.Key{ebiten.KeyDigit1, ebiten.KeyDigit2, ebiten.KeyDigit3, ebiten.KeyDigit4}
		for i, key := range keys {
			if !a.actionPressed(actions.Selection(i+1), key) {
				continue
			}
			a.recruitBranch = branches[i]
			a.recruitLimit = a.world.RecruitLimit(a.current, a.recruitBranch)
			if a.recruitLimit <= 0 {
				a.report("無法徵" + game.BranchName(a.recruitBranch))
				a.screen, a.dirty = screenCommand, true
				return nil
			}
			if a.recruitLimit > 99999 {
				a.recruitLimit = 99999 // 原版輸入框的明確上限 0x1869F。
			}
			a.recruitAmount = 0
			a.screen, a.dirty = screenRecruitAmount, true
			return nil
		}
		return nil

	case screenRecruitAmount:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenRecruitBranch, true
			return nil
		}
		if a.deleteDigitPressed() {
			a.recruitAmount /= 10
			a.dirty = true
		}
		for d, k := range []ebiten.Key{ebiten.KeyDigit0, ebiten.KeyDigit1, ebiten.KeyDigit2,
			ebiten.KeyDigit3, ebiten.KeyDigit4, ebiten.KeyDigit5, ebiten.KeyDigit6,
			ebiten.KeyDigit7, ebiten.KeyDigit8, ebiten.KeyDigit9} {
			if a.digitPressed(d, k) {
				next := a.recruitAmount*10 + uint32(d)
				if next <= uint32(a.recruitLimit) {
					a.recruitAmount, a.dirty = next, true
				}
			}
		}
		if a.submitPressed() {
			if a.recruitAmount == 0 {
				a.screen = screenRecruitBranch
			} else {
				a.screen = screenRecruitConfirm
			}
			a.dirty = true
		}
		return nil

	case screenRecruitConfirm:
		if a.actionPressed(actions.Cancel, ebiten.KeyN, ebiten.KeyEscape) {
			a.screen, a.dirty = screenRecruitAmount, true
		} else if a.actionPressed(actions.Confirm, ebiten.KeyY) {
			branch, amount := a.recruitBranch, int(a.recruitAmount)
			a.report(a.withBudget(a.current, func() string { return a.execRecruit(branch, amount) }))
			a.screen, a.dirty = screenCommand, true
		}
		return nil

	case screenReorganizeBranch:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenRecruitAction, true
			return nil
		}
		keys := []ebiten.Key{ebiten.KeyDigit1, ebiten.KeyDigit2, ebiten.KeyDigit3, ebiten.KeyDigit4}
		for i, key := range keys {
			if !a.actionPressed(actions.Selection(i+1), key) {
				continue
			}
			r, err := a.world.BeginReorganization(a.current, game.RecruitBranchOrder[i])
			if err != nil {
				a.report(err.Error())
				a.screen, a.dirty = screenCommand, true
				return nil
			}
			a.reorganization, a.reorganizeInput = r, 0
			a.screen, a.dirty = screenReorganizeTarget, true
			return nil
		}
		return nil

	case screenReorganizeTarget:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			if a.reorganization.CanFinish() {
				a.finishReorganization()
			}
			return nil
		}
		if a.deleteDigitPressed() {
			a.reorganizeInput /= 10
			a.dirty = true
		}
		targets := a.reorganization.Targets()
		direct := false
		if n, ok := actions.SelectionNumber(a.pointerAction); ok && n <= len(targets) {
			a.reorganizeInput, direct = uint32(n), true
		}
		for d, k := range []ebiten.Key{ebiten.KeyDigit0, ebiten.KeyDigit1, ebiten.KeyDigit2,
			ebiten.KeyDigit3, ebiten.KeyDigit4, ebiten.KeyDigit5, ebiten.KeyDigit6,
			ebiten.KeyDigit7, ebiten.KeyDigit8, ebiten.KeyDigit9} {
			if a.digitPressed(d, k) {
				next := a.reorganizeInput*10 + uint32(d)
				if next <= uint32(len(targets)) {
					a.reorganizeInput, a.dirty = next, true
				}
			}
		}
		if a.submitPressed() || direct {
			if a.reorganizeInput == 0 {
				if a.reorganization.CanFinish() {
					a.finishReorganization()
				}
			} else if int(a.reorganizeInput) <= len(targets) {
				a.reorganizeID = targets[int(a.reorganizeInput)-1]
				a.reorganizeInput = 0
				a.screen, a.dirty = screenReorganizeAmount, true
			}
		}
		return nil

	case screenReorganizeAmount:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.reorganizeInput = 0
			a.screen, a.dirty = screenReorganizeTarget, true
			return nil
		}
		if a.deleteDigitPressed() {
			a.reorganizeInput /= 10
			a.dirty = true
		}
		limit := a.reorganization.Limit(a.reorganizeID)
		for d, k := range []ebiten.Key{ebiten.KeyDigit0, ebiten.KeyDigit1, ebiten.KeyDigit2,
			ebiten.KeyDigit3, ebiten.KeyDigit4, ebiten.KeyDigit5, ebiten.KeyDigit6,
			ebiten.KeyDigit7, ebiten.KeyDigit8, ebiten.KeyDigit9} {
			if a.digitPressed(d, k) {
				next := a.reorganizeInput*10 + uint32(d)
				if next <= uint32(limit) {
					a.reorganizeInput, a.dirty = next, true
				}
			}
		}
		if a.submitPressed() {
			if err := a.reorganization.Assign(a.reorganizeID, int(a.reorganizeInput)); err != nil {
				a.report(err.Error())
			}
			a.syncReorganizationGenerals()
			a.reorganizeInput = 0
			a.screen, a.dirty = screenReorganizeTarget, true
		}
		return nil

	case screenTrainConfirm:
		if a.actionPressed(actions.Cancel, ebiten.KeyN, ebiten.KeyEscape) {
			a.screen, a.dirty = screenCommand, true
		} else if a.actionPressed(actions.Confirm, ebiten.KeyY) {
			a.report(a.withBudget(a.current, a.execTrain))
			a.screen, a.dirty = screenCommand, true
		}
		return nil

	case screenCovertAction:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenCommand, true
		} else if a.actionPressed(actions.Select1, ebiten.KeyDigit1) {
			// 游擊隊的成效已解，但派遣成本仍未知；不可臆造數值。
			a.report("派遣游擊隊尚待原版成本證據")
			a.screen, a.dirty = screenCommand, true
		} else if a.actionPressed(actions.Select2, ebiten.KeyDigit2) {
			a.covertAction, a.covertInput = 2, 0
			a.screen, a.dirty = screenCovertTarget, true
		}
		return nil

	case screenCovertTarget:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenCovertAction, true
			return nil
		}
		if a.deleteDigitPressed() {
			a.covertInput /= 10
			a.dirty = true
		}
		for d, key := range []ebiten.Key{ebiten.KeyDigit0, ebiten.KeyDigit1, ebiten.KeyDigit2,
			ebiten.KeyDigit3, ebiten.KeyDigit4, ebiten.KeyDigit5, ebiten.KeyDigit6,
			ebiten.KeyDigit7, ebiten.KeyDigit8, ebiten.KeyDigit9} {
			if a.digitPressed(d, key) {
				next := a.covertInput*10 + uint32(d)
				if next <= game.ProvinceCount {
					a.covertInput, a.dirty = next, true
				}
			}
		}
		if a.submitPressed() {
			target := game.ProvinceID(a.covertInput)
			if target.Valid() && a.covertAction == 2 {
				src, _ := a.tbl.At(a.current)
				dst, _ := a.tbl.At(target)
				if dst.InBattle() {
					a.report("目標省目前正在戰爭")
					a.screen, a.dirty = screenCommand, true
					return nil
				}
				if int(src.Gold) < game.StudentProtestCost {
					a.report(fmt.Sprintf("資金不足（有 %d，要 %d）", src.Gold, game.StudentProtestCost))
					a.screen, a.dirty = screenCommand, true
					return nil
				}
				a.report(a.withBudget(a.current, func() string { return a.execStudentProtest(target) }))
				a.screen, a.dirty = screenCommand, true
			}
		}
		return nil

	case screenDiplomacy:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		switch {
		case a.actionPressed(actions.Select1, ebiten.KeyDigit1):
			slot, err := a.currentDiplomacySlot()
			if err != nil {
				a.report(err.Error())
				return nil
			}
			// `sub_2164A` 在進入額度輸入前先檢查信用度表；0 會直接
			// 顯示「無法貸款」並返回主迴圈，不會消耗貸款亂數或指令。
			if a.ledger.Credit[slot] == 0 {
				a.report(a.loanText("diplomacy.loan.unavailable", "無法貸款"))
				a.screen, a.dirty = screenCommand, true
				return nil
			}
			a.diplomacySlot, a.diplomacyInput = slot, 0
			a.screen, a.dirty = screenLoanAmount, true
		case a.actionPressed(actions.Select2, ebiten.KeyDigit2):
			if a.cmdBudget.Remaining(a.current) <= 0 {
				a.report(fmt.Sprintf("%s 這個月的指令數用完了", a.provinceName(a.current)))
				a.screen, a.dirty = screenCommand, true
				return nil
			}
			a.report(a.executeDiplomacyAid())
			a.screen, a.dirty = screenCommand, true
		case a.actionPressed(actions.Select3, ebiten.KeyDigit3):
			slot, err := a.currentDiplomacySlot()
			if err != nil {
				a.report(err.Error())
				return nil
			}
			if a.ledger.Debt[slot] == 0 {
				msg, msgErr := a.wordingText("diplomacy.repay.none")
				if msgErr != nil {
					a.report("償還外債失敗：" + msgErr.Error())
				} else {
					a.report(msg)
				}
				a.screen, a.dirty = screenCommand, true
				return nil
			}
			a.diplomacySlot, a.diplomacyInput = slot, 0
			a.screen, a.dirty = screenRepayAmount, true
		}
		return nil

	case screenCeasefireTarget:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.ceasefireInput = 0
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		if a.deleteDigitPressed() {
			a.ceasefireInput /= 10
			a.dirty = true
		}
		limit := a.provinceLimit
		if limit <= 0 {
			limit = game.CeasefireProvinceLimit(uint8(a.stage))
		}
		for d, key := range digitKeys() {
			if !a.digitPressed(d, key) {
				continue
			}
			next := a.ceasefireInput*10 + uint32(d)
			if next >= a.ceasefireInput && (limit == 0 || next <= uint32(limit)) {
				a.ceasefireInput, a.dirty = next, true
			}
		}
		if !a.submitPressed() {
			return nil
		}
		if a.ceasefireInput == 0 || limit == 0 || int(a.ceasefireInput) > limit {
			a.report(a.ceasefireText("ceasefire.invalid", "停火目標省編號無效"))
			return nil
		}
		if a.cmdBudget == nil || a.cmdBudget.Remaining(a.current) <= 0 {
			a.report(fmt.Sprintf("%s 這個月的指令數用完了", a.provinceName(a.current)))
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		if a.world == nil || a.rng == nil {
			a.report("停火需要有效的遊戲世界與可重現的亂數來源")
			return nil
		}
		target := game.ProvinceID(a.ceasefireInput)
		requester, err := a.world.ValidateCeasefireTarget(uint8(a.stage), a.current, target)
		if err != nil {
			switch {
			case strings.Contains(err.Error(), "並無戰事"):
				a.report(a.ceasefireText("ceasefire.no_battle", "本省並無戰事"))
				a.screen, a.dirty = screenCommand, true
			case strings.Contains(err.Error(), "司令不在") || strings.Contains(err.Error(), "目前省沒有"):
				a.report(a.ceasefireText("ceasefire.no_commander", "司令不在本省"))
				a.screen, a.dirty = screenCommand, true
			default:
				// 原版對超界／無主目標會留在輸入迴圈；玩家可以修正
				// 數字，不把一次無效輸入誤算成已完成指令。
				a.report(a.ceasefireText("ceasefire.invalid", "停火目標省編號無效"))
			}
			return nil
		}
		beforeState, seed := a.world.CeasefireState[target], a.rng.Seed()
		res, err := a.world.NegotiateCeasefire(target, requester, a.rng)
		if err != nil {
			a.report("談判停火失敗：" + err.Error())
			return nil
		}
		// sub_20E05 對同意／拒絕都會立起 byte_6FE81，原版主迴圈
		// 因而都扣一次指令；前置拒絕則不會走到這裡。
		if !a.cmdBudget.Spend(a.current) {
			a.world.CeasefireState[target] = beforeState
			a.rng.SetSeed(seed)
			a.report("指令數已用完，停火狀態未能提交")
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		key := "ceasefire.refused"
		fallback := fmt.Sprintf("拒絕在省 %d 停火（剩餘指令 %d）", target, a.cmdBudget.Remaining(a.current))
		if res.Agreed {
			key = "ceasefire.agreed"
			fallback = fmt.Sprintf("同意在省 %d 停火（剩餘指令 %d）", target, a.cmdBudget.Remaining(a.current))
		}
		msg, msgErr := a.wordingFormat(key, target, a.cmdBudget.Remaining(a.current))
		if msgErr != nil {
			msg = fallback
		}
		a.report(msg)
		a.ceasefireInput = 0
		a.screen, a.dirty = screenCommand, true
		return nil

	case screenLoanAmount:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.diplomacyInput = 0
			a.screen, a.dirty = screenDiplomacy, true
			return nil
		}
		if a.deleteDigitPressed() {
			a.diplomacyInput /= 10
			a.dirty = true
		}
		for d, key := range digitKeys() {
			if a.digitPressed(d, key) {
				next := a.diplomacyInput*10 + uint32(d)
				if next <= 5000 {
					a.diplomacyInput, a.dirty = next, true
				}
			}
		}
		if !a.submitPressed() {
			return nil
		}
		if a.diplomacyInput == 0 {
			a.report("貸款額要為正")
			return nil
		}
		if a.cmdBudget.Remaining(a.current) <= 0 {
			a.report(fmt.Sprintf("%s 這個月的指令數用完了", a.provinceName(a.current)))
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		if a.rng == nil {
			a.report("貸款需要可重現的亂數來源")
			return nil
		}
		prov, err := a.tbl.At(a.current)
		if err != nil {
			a.report("貸款失敗：" + err.Error())
			return nil
		}
		beforeGold := prov.Gold
		beforeLedger := a.ledger
		beforeSeed := a.rng.Seed()
		res, err := a.ledger.RequestLoan(a.world, a.current, a.diplomacySlot,
			int(a.diplomacyInput), a.rng)
		if err != nil {
			a.report("貸款失敗：" + err.Error())
			return nil
		}
		if res.CreditBlocked {
			a.report(a.loanText("diplomacy.loan.unavailable", "無法貸款"))
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		if !res.Approved {
			if res.CommandCompleted && !a.cmdBudget.Spend(a.current) {
				a.rng.SetSeed(beforeSeed)
				a.report("指令數已用完，貸款拒絕狀態未能提交")
				a.screen, a.dirty = screenCommand, true
				return nil
			}
			a.report(a.loanText("diplomacy.loan.refused", "各國均拒絕提供貸款"))
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		if !a.cmdBudget.Spend(a.current) {
			// 進入前已檢查剩餘數；這只防護未來非同步改動，並回復
			// 貸款可能已寫入的兩個狀態來源。
			prov.Gold = beforeGold
			a.ledger = beforeLedger
			a.rng.SetSeed(beforeSeed)
			a.report("指令數已用完，貸款狀態未能提交")
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		a.report(fmt.Sprintf("貸款核准：黃金 +%d，外債 +%d（信用度 %d；剩 %d）",
			res.Amount, res.Amount, a.ledger.Credit[a.diplomacySlot],
			a.cmdBudget.Remaining(a.current)))
		a.diplomacyInput = 0
		a.screen, a.dirty = screenCommand, true
		return nil

	case screenRepayAmount:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.diplomacyInput = 0
			a.screen, a.dirty = screenDiplomacy, true
			return nil
		}
		if a.deleteDigitPressed() {
			a.diplomacyInput /= 10
			a.dirty = true
		}
		for d, key := range digitKeys() {
			if !a.digitPressed(d, key) {
				continue
			}
			// 償還金額由規則層驗證是否超過外債；輸入層只防止 uint32
			// 溢位，讓玩家能看見並修正「超過外債」的錯誤輸入。
			next := a.diplomacyInput*10 + uint32(d)
			if next >= a.diplomacyInput {
				a.diplomacyInput, a.dirty = next, true
			}
		}
		if !a.submitPressed() {
			return nil
		}
		if a.diplomacyInput == 0 {
			msg, err := a.wordingText("diplomacy.repay.invalid")
			if err != nil {
				a.report("償還外債失敗：" + err.Error())
			} else {
				a.report(msg)
			}
			return nil
		}
		if a.cmdBudget.Remaining(a.current) <= 0 {
			a.report(fmt.Sprintf("%s 這個月的指令數用完了", a.provinceName(a.current)))
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		if a.world == nil || a.tbl == nil {
			a.report("償還外債需要有效的遊戲世界")
			return nil
		}
		prov, err := a.tbl.At(a.current)
		if err != nil {
			a.report("償還外債失敗：" + err.Error())
			return nil
		}
		beforeGold, beforeLedger := prov.Gold, a.ledger
		res, err := a.ledger.RepayDebt(a.world, a.current, a.diplomacySlot,
			int(a.diplomacyInput))
		if err != nil {
			a.report("償還外債失敗：" + err.Error())
			return nil
		}
		if !a.cmdBudget.Spend(a.current) {
			prov.Gold, a.ledger = beforeGold, beforeLedger
			a.report("指令數已用完，償還狀態未能提交")
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		msg, msgErr := a.wordingFormat("diplomacy.repay.result", res.Amount,
			res.DebtAfter, res.CreditAfter, a.cmdBudget.Remaining(a.current))
		if msgErr != nil {
			a.report("償還外債完成，但結果用語缺失：" + msgErr.Error())
		} else {
			a.report(msg)
		}
		a.diplomacyInput = 0
		a.screen, a.dirty = screenCommand, true
		return nil

	case screenViewGenerals:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = a.viewGeneralBack, true
			return nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyUp) && a.viewIndex > 0 {
			a.viewIndex--
			a.dirty = true
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyDown) && a.viewIndex+1 < len(a.viewGenerals) {
			a.viewIndex++
			a.dirty = true
		}
		if n, ok := actions.SelectionNumber(a.pointerAction); ok && n <= len(a.viewGenerals) {
			a.viewIndex = n - 1
			a.screen, a.dirty = screenViewGeneral, true
			return nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
			a.screen, a.dirty = screenViewGeneral, true
		}
		if a.actionPressed(actions.OpenBiography, ebiten.KeyB) {
			a.openBiography(screenViewGenerals)
		}
		return nil

	case screenViewMenu:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		switch {
		case a.actionPressed(actions.Select1, ebiten.KeyDigit1):
			a.viewInput = 0
			a.screen, a.dirty = screenViewProvinceSelect, true
		case a.actionPressed(actions.Select2, ebiten.KeyDigit2):
			a.screen, a.dirty = screenViewOverview, true
		case a.actionPressed(actions.Select3, ebiten.KeyDigit3):
			a.viewGenerals = a.world.ActiveGeneralsAt(a.current)
			a.viewIndex = 0
			a.viewGeneralBack = screenViewMenu
			if len(a.viewGenerals) == 0 {
				a.report("本省並無可查閱將領")
			} else {
				a.screen, a.dirty = screenViewGenerals, true
			}
		case a.actionPressed(actions.Select4, ebiten.KeyDigit4):
			a.viewPage = 1
			a.screen, a.dirty = screenViewProvinceNames, true
		}
		return nil

	case screenViewProvinceSelect:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenViewMenu, true
			return nil
		}
		if a.deleteDigitPressed() {
			a.viewInput /= 10
			a.dirty = true
		}
		for d, key := range []ebiten.Key{ebiten.KeyDigit0, ebiten.KeyDigit1, ebiten.KeyDigit2,
			ebiten.KeyDigit3, ebiten.KeyDigit4, ebiten.KeyDigit5, ebiten.KeyDigit6,
			ebiten.KeyDigit7, ebiten.KeyDigit8, ebiten.KeyDigit9} {
			if inpututil.IsKeyJustPressed(key) {
				next := a.viewInput*10 + uint32(d)
				if next <= game.ProvinceCount {
					a.viewInput, a.dirty = next, true
				}
			}
		}
		direct := false
		if n, ok := actions.SelectionNumber(a.pointerAction); ok && n <= game.ProvinceCount {
			a.viewInput, direct = uint32(n), true
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) || direct {
			p := game.ProvinceID(a.viewInput)
			if p.Valid() {
				a.viewProvince = p
				a.screen, a.dirty = screenViewProvinceChoice, true
			}
		}
		return nil

	case screenViewProvinceChoice:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenViewProvinceSelect, true
			return nil
		}
		if a.actionPressed(actions.Select1, ebiten.KeyDigit1) {
			a.screen, a.dirty = screenViewProvince, true
		} else if a.actionPressed(actions.Select2, ebiten.KeyDigit2) {
			a.viewGenerals = a.world.ActiveGeneralsAt(a.viewProvince)
			a.viewIndex = 0
			a.viewGeneralBack = screenViewProvinceChoice
			if len(a.viewGenerals) == 0 {
				a.report("該省並無可查閱將領")
			} else {
				a.screen, a.dirty = screenViewGenerals, true
			}
		}
		return nil

	case screenViewProvince:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenViewProvinceChoice, true
		}
		return nil

	case screenViewOverview:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenViewMenu, true
		}
		return nil

	case screenViewProvinceNames:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) ||
			inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
			a.screen, a.dirty = screenViewMenu, true
			return nil
		}
		if a.actionPressed(actions.NextPage, ebiten.KeySpace) {
			if a.viewPage == 1 {
				a.viewPage = 2
			} else {
				a.viewPage = 1
			}
			a.dirty = true
		}
		return nil

	case screenViewGeneral:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenViewGenerals, true
			return nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) && a.viewIndex > 0 {
			a.viewIndex--
			a.dirty = true
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyRight) && a.viewIndex+1 < len(a.viewGenerals) {
			a.viewIndex++
			a.dirty = true
		}
		if a.actionPressed(actions.OpenBiography, ebiten.KeyB) {
			a.openBiography(screenViewGeneral)
		}
		return nil

	case screenBiography:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyB) {
			a.screen, a.dirty = a.bioBack, true
			return nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyLeft) && a.viewIndex > 0 {
			a.viewIndex--
			a.openBiography(a.bioBack)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyRight) && a.viewIndex+1 < len(a.viewGenerals) {
			a.viewIndex++
			a.openBiography(a.bioBack)
		}
		if a.actionPressed(actions.PreviousPage, ebiten.KeyPageUp) && a.bioPage > 0 {
			a.bioPage--
			a.dirty = true
		}
		if (a.actionPressed(actions.NextPage, ebiten.KeySpace, ebiten.KeyPageDown)) && a.bioPage+1 < a.bioPages {
			a.bioPage++
			a.dirty = true
		}
		return nil

	case screenNarrative:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyN) {
			a.screen, a.dirty = a.narrativeBack, true
			return nil
		}
		pages := render.NarrativePageCount(a.narrativeImages)
		if a.actionPressed(actions.PreviousPage, ebiten.KeyPageUp) && a.narrativePage > 0 {
			a.narrativePage--
			a.dirty = true
		}
		if a.actionPressed(actions.NextPage, ebiten.KeySpace, ebiten.KeyPageDown) && a.narrativePage+1 < pages {
			a.narrativePage++
			a.dirty = true
		}
		return nil

	case screenPolicy:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		if a.actionPressed(actions.Select2, ebiten.KeyDigit2) {
			a.productionItem, a.productionInput, a.productionSpent = 0, 0, false
			a.screen, a.dirty = screenProduction, true
			return nil
		}
		if !a.actionPressed(actions.Select1, ebiten.KeyDigit1) {
			return nil
		}
		targets, err := a.world.AutonomyTargets(a.current)
		if err != nil {
			a.report(err.Error())
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		if len(targets) == 0 {
			a.report("目前沒有可切換自治的省份")
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		a.autonomyTargets = targets
		a.autonomyInput, a.autonomySpent = 0, false
		a.screen, a.dirty = screenAutonomy, true
		return nil

	case screenProduction:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			if a.productionItem != 0 {
				a.productionItem, a.productionInput, a.dirty = 0, 0, true
				return nil
			}
			if a.productionSpent {
				a.screen = screenCommand
			} else {
				a.screen = screenPolicy
			}
			a.dirty = true
			return nil
		}
		if a.deleteDigitPressed() && a.productionItem != 0 {
			a.productionInput /= 10
			a.dirty = true
			return nil
		}
		for d, key := range digitKeys() {
			if !a.digitPressed(d, key) {
				continue
			}
			if a.productionItem == 0 {
				if d >= 1 && d <= 4 {
					a.productionItem, a.productionInput = d, 0
					a.dirty = true
				}
			} else if a.productionInput < 100 {
				a.productionInput = a.productionInput*10 + uint32(d)
				a.dirty = true
			}
			return nil
		}
		if n, ok := actions.SelectionNumber(a.pointerAction); ok && a.productionItem == 0 && n <= 4 {
			a.productionItem, a.productionInput, a.dirty = n, 0, true
			return nil
		}
		if !a.submitPressed() || a.productionItem == 0 {
			return nil
		}
		p, err := a.tbl.At(a.current)
		if err != nil {
			return err
		}
		before := p.ProductionAllocation().Value(a.productionItem)
		if err := p.SetProductionAllocation(a.productionItem, uint8(a.productionInput)); err != nil {
			a.report(err.Error())
			a.dirty = true
			return nil
		}
		if !a.productionSpent {
			if !a.cmdBudget.Spend(a.current) {
				_ = p.SetProductionAllocation(a.productionItem, before)
				a.report("指令數已用完，產能未變更")
				a.screen, a.dirty = screenCommand, true
				return nil
			}
			a.productionSpent = true
		}
		a.report(fmt.Sprintf("產能已調整為 %d%%（剩 %d）", a.productionInput, a.cmdBudget.Remaining(a.current)))
		a.productionItem, a.productionInput, a.dirty = 0, 0, true
		return nil

	case screenAutonomy:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			if a.autonomySpent {
				a.screen = screenCommand
			} else {
				a.screen = screenPolicy
			}
			a.dirty = true
			return nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
			a.autonomyInput /= 10
			a.dirty = true
			return nil
		}
		for d, key := range digitKeys() {
			if !inpututil.IsKeyJustPressed(key) {
				continue
			}
			if a.autonomyInput < 100 {
				a.autonomyInput = a.autonomyInput*10 + uint32(d)
				a.dirty = true
			}
			return nil
		}
		direct := false
		if n, ok := actions.SelectionNumber(a.pointerAction); ok && n <= len(a.autonomyTargets) {
			a.autonomyInput, direct = uint32(a.autonomyTargets[n-1]), true
		}
		if !a.submitPressed() && !direct {
			return nil
		}
		target := game.ProvinceID(a.autonomyInput)
		a.autonomyInput = 0
		if !provinceIn(a.autonomyTargets, target) {
			a.report("該省份不在可授權自治的名單中")
			a.dirty = true
			return nil
		}
		on, err := a.world.TogglePlayerAutonomy(a.current, target)
		if err != nil {
			a.report(err.Error())
			a.dirty = true
			return nil
		}
		if !a.autonomySpent {
			// 同一次 sub_22E25 可切換多省，整個政策指令只扣一次。
			if !a.cmdBudget.Spend(a.current) {
				// 單執行緒下理論上不會發生；若發生就將 toggle 回滾。
				_, _ = a.world.ToggleAutonomy(target)
				a.report("指令數已用完，自治狀態未變更")
				a.screen, a.dirty = screenCommand, true
				return nil
			}
			a.autonomySpent = true
		}
		state := "正常"
		if on {
			state = "自治"
		}
		a.report(fmt.Sprintf("%s：%s（剩 %d）", a.provinceName(target), state,
			a.cmdBudget.Remaining(a.current)))
		a.dirty = true
		return nil

	case screenOtherOptions:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		if a.actionPressed(actions.Select1, ebiten.KeyDigit1) {
			if a.savePath == "" || a.origSave == nil {
				a.report("目前沒有可寫回的存檔基底")
				return nil
			}
			a.screen, a.dirty = screenSaveConfirm, true
			return nil
		}
		if a.actionPressed(actions.Select2, ebiten.KeyDigit2) {
			if a.savePath == "" {
				a.report("沒有設定載入路徑")
				return nil
			}
			a.screen, a.dirty = screenLoadConfirm, true
			return nil
		}
		if a.actionPressed(actions.Select6, ebiten.KeyDigit6) {
			a.messageTimeInput = 0
			a.screen, a.dirty = screenMessageTime, true
			return nil
		}
		if a.actionPressed(actions.Select8, ebiten.KeyDigit8) {
			a.quitBack, a.screen, a.dirty = screenOtherOptions, screenQuit, true
			return nil
		}
		if a.actionPressed(actions.Select9, ebiten.KeyDigit9) {
			a.screen, a.dirty = screenDisplayOptions, true
			return nil
		}
		if a.actionPressed(actions.Select10, ebiten.KeyDigit0) {
			a.screen, a.dirty = screenResolutionOptions, true
			return nil
		}
		for i, key := range []ebiten.Key{ebiten.KeyDigit3, ebiten.KeyDigit4,
			ebiten.KeyDigit5, ebiten.KeyDigit7} {
			n := []int{3, 4, 5, 7}[i]
			if a.actionPressed(actions.Selection(n), key) {
				a.report("這項原版設定尚未完成，不會變更狀態")
				return nil
			}
		}
		return nil

	case screenMessageTime:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.messageTimeInput = 0
			a.screen, a.dirty = screenOtherOptions, true
			return nil
		}
		if a.deleteDigitPressed() {
			a.messageTimeInput /= 10
			a.dirty = true
			return nil
		}
		for d, key := range digitKeys() {
			if !a.digitPressed(d, key) {
				continue
			}
			if a.messageTimeInput < 10 {
				a.messageTimeInput = a.messageTimeInput*10 + uint32(d)
				a.dirty = true
			}
			return nil
		}
		if !a.submitPressed() {
			return nil
		}
		if a.messageTimeInput < 1 || a.messageTimeInput > 10 {
			a.report("訊息時間要在 1 到 10 之間")
			return nil
		}
		if err := a.setMessageTimePreference(int(a.messageTimeInput)); err != nil {
			a.report("訊息時間未變更：" + err.Error())
			return nil
		}
		a.messageTimeInput = 0
		a.report("訊息時間已設定")
		a.screen, a.dirty = screenOtherOptions, true
		return nil

	case screenSaveConfirm:
		switch {
		case a.actionPressed(actions.Confirm, ebiten.KeyY):
			if err := a.autosave(); err != nil {
				a.report("儲存失敗：" + err.Error())
			} else {
				a.report("遊戲已儲存")
			}
			a.screen, a.dirty = screenOtherOptions, true
		case a.actionPressed(actions.Cancel, ebiten.KeyN, ebiten.KeyEscape):
			a.screen, a.dirty = screenOtherOptions, true
		}
		return nil

	case screenLoadConfirm:
		switch {
		case a.actionPressed(actions.Confirm, ebiten.KeyY):
			b, err := os.ReadFile(a.savePath)
			if err == nil {
				err = a.loadSessionBytes(b)
			}
			if err != nil {
				a.report("載入失敗，目前遊戲未變更：" + err.Error())
				a.screen, a.dirty = screenOtherOptions, true
				return nil
			}
			a.report("遊戲已載入")
			a.screen, a.dirty = screenCommand, true
		case a.actionPressed(actions.Cancel, ebiten.KeyN, ebiten.KeyEscape):
			a.screen, a.dirty = screenOtherOptions, true
		}
		return nil

	case screenDisplayOptions:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenOtherOptions, true
			return nil
		}
		switch {
		case a.actionPressed(actions.Select1, ebiten.KeyDigit1):
			if err := a.setWordingPreference(i18n.WordingOriginal); err != nil {
				a.report("顯示用語未變更：" + err.Error())
			}
		case a.actionPressed(actions.Select2, ebiten.KeyDigit2):
			if err := a.setWordingPreference(i18n.WordingPlain); err != nil {
				a.report("顯示用語未變更：" + err.Error())
			}
		case a.actionPressed(actions.Select3, ebiten.KeyDigit3):
			if err := a.setThemePreference(uitheme.ModeRetro); err != nil {
				a.report("圖形主題未變更：" + err.Error())
			}
		case a.actionPressed(actions.Select4, ebiten.KeyDigit4):
			if err := a.setThemePreference(uitheme.ModeModern); err != nil {
				a.report("圖形主題未變更：" + err.Error())
			}
		}
		return nil

	case screenResolutionOptions:
		if a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.screen, a.dirty = screenOtherOptions, true
			return nil
		}
		switch {
		case a.actionPressed(actions.Select1, ebiten.KeyDigit1):
			if err := a.setResolutionPreference(uiresolution.ModeOriginal); err != nil {
				a.report("解析度未變更：" + err.Error())
			}
		case a.actionPressed(actions.Select2, ebiten.KeyDigit2):
			if err := a.setResolutionPreference(uiresolution.ModeHigh); err != nil {
				a.report("解析度未變更：" + err.Error())
			}
		}
		return nil
	}

	switch {
	case a.actionPressed(actions.OpenNarrative, ebiten.KeyN):
		if a.narrative == nil || len(a.narrativeImages) == 0 {
			a.report("史事新聞資料尚未載入")
			break
		}
		a.narrativeBack, a.narrativePage, a.screen, a.dirty = screenMap, 0, screenNarrative, true
	case a.actionPressed(actions.OpenCommands, ebiten.KeyEnter, ebiten.KeyKPEnter):
		a.screen, a.dirty = screenCommand, true
	case inpututil.IsKeyJustPressed(ebiten.KeyRight):
		a.current++
		if a.current > game.ProvinceCount {
			a.current = 1
		}
		a.dirty = true
	case inpututil.IsKeyJustPressed(ebiten.KeyLeft):
		a.current--
		if a.current < 1 {
			a.current = game.ProvinceCount
		}
		a.dirty = true
	}
	return nil
}

func (a *app) currentBiography() (*i18n.Person, bool) {
	if a.people == nil || a.eten == nil || a.wording == nil ||
		a.viewIndex < 0 || a.viewIndex >= len(a.viewGenerals) {
		return nil, false
	}
	return a.people.PersonAt(a.stage, int(a.viewGenerals[a.viewIndex]))
}

func (a *app) portraitForPerson(person *i18n.Person) (*i18n.Portrait, string) {
	placeholder := ""
	if a.wording != nil {
		placeholder, _ = a.wording.Text("biography.portrait_unavailable", a.wordingMode)
	}
	if person == nil || a.portraits == nil {
		return nil, placeholder
	}
	portrait, ok := a.portraits.PortraitFor(person.ID)
	if !ok {
		return nil, placeholder
	}
	return &portrait, placeholder
}

func (a *app) portraitForGeneral(id game.GeneralID) (*i18n.Portrait, string) {
	if a.people == nil {
		return a.portraitForPerson(nil)
	}
	person, _ := a.people.PersonAt(a.stage, int(id))
	return a.portraitForPerson(person)
}

// biographyFallbackBody 組出無正文槽的有據檔案卡：第一行是既有的
// 待考宣告，第二行只列登錄資料（派系／時期，均出自 people.json），
// 句句可溯源，不寫推定句。SPEC-46 R2 的關閉條件就靠這一頁。
func biographyFallbackBody(p *i18n.Person, unavailable, recordLabel string) string {
	faction := p.Faction
	if faction == "" {
		faction = "—"
	}
	periods := "—"
	if len(p.Periods) > 0 {
		periods = strings.Join(p.Periods, "／")
	}
	return unavailable + "\n" + recordLabel + "：" + faction + "／" + periods
}

func (a *app) openBiography(back screen) {
	p, ok := a.currentBiography()
	if !ok {
		a.report("這位將領目前沒有可顯示的人物資料")
		return
	}
	body := p.Biography
	if body == "" {
		unavailable, found := a.wording.Text("biography.unavailable", a.wordingMode)
		if !found {
			a.report("人物自傳用語資料不完整")
			return
		}
		recordLabel, found := a.wording.Text("biography.record", a.wordingMode)
		if !found {
			a.report("人物自傳用語資料不完整")
			return
		}
		body = biographyFallbackBody(p, unavailable, recordLabel)
	}
	doc, err := textlayout.Layout(body, textlayout.DefaultBiographyOptions)
	if err != nil {
		a.report(err.Error())
		return
	}
	pageCount := len(doc.Pages)
	if a.resolution == uiresolution.ModeHigh && a.themeMode == uitheme.ModeModern {
		// 高解析 H4 的比例字距、欄寬與復古 28 格版不同；必須使用 renderer
		// 共用的頁數入口，否則 NextPage 可能指向畫面不存在的一頁。
		pageCount, err = render.ModernBiographyPageCount(body)
		if err != nil {
			a.report(err.Error())
			return
		}
	}
	a.bioBack, a.bioPage, a.bioPages = back, 0, pageCount
	a.screen, a.dirty = screenBiography, true
}

func (a *app) wordingText(key string) (string, error) {
	if a.wording == nil {
		return "", fmt.Errorf("顯示用語資料未載入")
	}
	text, ok := a.wording.Text(key, a.wordingMode)
	if !ok {
		return "", fmt.Errorf("顯示用語缺少 %q/%s", key, a.wordingMode)
	}
	return text, nil
}

// uiStyle 讓主題圖集與整個畫面外殼採同一份 style。若外部測試 provider
// 尚未實作 StyleProvider，保留 retro-safe 預設，不因 UI polish 破壞窄測試。
func (a *app) uiStyle() uitheme.UIStyle {
	if a != nil && a.battlefieldTheme != nil {
		if provider, ok := a.battlefieldTheme.(uitheme.StyleProvider); ok {
			return provider.Style()
		}
	}
	return uitheme.RetroStyle()
}

// highModernMap 開啟已完成 H1-a metrics 的地圖／資訊卡頁。
func (a *app) highModernMap() bool {
	return a != nil && a.resolution == uiresolution.ModeHigh &&
		a.themeMode == uitheme.ModeModern && a.screen == screenMap
}

// highModernPage 是已有專用 renderer 的文件／設定／戰鬥頁集合。
func (a *app) highModernPage() bool {
	if a == nil || a.resolution != uiresolution.ModeHigh || a.themeMode != uitheme.ModeModern {
		return false
	}
	switch a.screen {
	case screenCommand, screenBiography, screenNarrative, screenBattle,
		screenDisplayOptions, screenResolutionOptions, screenOtherOptions:
		return true
	default:
		return false
	}
}

// highModernFlow 把其餘 Modern 政略／查閱流程也放到 1280×720 設計畫布；
// 專用頁維持自己的 renderer，其餘頁面使用 ModernFlowSurface 與同一份
// Selection／Digit／Confirm 命中契約，不再回退到 640×350 frame。
func (a *app) highModernFlow() bool {
	return a != nil && a.resolution == uiresolution.ModeHigh &&
		a.themeMode == uitheme.ModeModern && a.screen != screenMap && a.screen != screenBattle
}

func (a *app) highModernBattle() bool {
	return a != nil && a.resolution == uiresolution.ModeHigh &&
		a.themeMode == uitheme.ModeModern && a.screen == screenBattle
}

func (a *app) highModernSurface() bool {
	return a.highModernMap() || a.highModernFlow() || a.highModernBattle()
}

// panelLabels 是 modern 側欄唯一的語系入口。所有欄位都由 wording catalog
// 提供；缺一項就 fail-closed，不能在英文／日文畫面中偷偷混入繁中詞表。
func (a *app) panelLabels() (render.PanelLabels, error) {
	var labels render.PanelLabels
	values := []*string{
		&labels.Status, &labels.Commander, &labels.Governor,
		&labels.Gold, &labels.Food, &labels.Ammo, &labels.Fuel,
		&labels.Coal, &labels.Iron, &labels.Land, &labels.Population,
		&labels.Cities, &labels.Arsenal, &labels.Force, &labels.Generals,
		&labels.People, &labels.Loyalty, &labels.Commands, &labels.Count,
	}
	keyList := []string{
		"panel.status.normal", "panel.commander", "panel.governor",
		"panel.gold", "panel.food", "panel.ammo", "panel.fuel",
		"panel.coal", "panel.iron", "panel.land", "panel.population",
		"panel.cities", "panel.arsenal", "panel.force", "panel.generals",
		"panel.people", "panel.loyalty", "panel.commands", "panel.count",
	}
	for i, key := range keyList {
		value, err := a.wordingText(key)
		if err != nil {
			return render.PanelLabels{}, err
		}
		*values[i] = value
	}
	return labels, nil
}

func (a *app) generalDisplayName(id game.GeneralID) string {
	if a != nil && a.people != nil {
		if person, ok := a.people.PersonAt(a.stage, int(id)); ok && person != nil {
			return person.NameInGame
		}
	}
	return ""
}

// currentFaction 只取已證實的勢力領袖反查槽位；一般將領或新局未知表回 0，
// modern 面板就不繪製可能誤導的色帶。
func (a *app) currentFaction(id game.GeneralID) int {
	if a == nil {
		return 0
	}
	slot := a.factionOf.SlotOf(id, a.factionLeaders)
	if slot == 0 {
		return 0
	}
	return int(slot)
}

// semanticWording 決定畫面是否走真正的語系字串路徑。繁中原典模式保留
// 原版 15 點字模與版面；現代白話，以及 en／ja 語系，都必須走 wording
// catalog，否則語系切換只會改省名、命令仍偷偷顯示繁中原版字模。
func (a *app) semanticWording() bool {
	if a.wordingMode == i18n.WordingPlain {
		return true
	}
	return a.loc != nil && a.loc.Language != "" && a.loc.Language != "zh-Hant"
}

func (a *app) wordingFormat(key string, args ...any) (string, error) {
	text, err := a.wordingText(key)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf(text, args...), nil
}

// ceasefireText 讓停火前置錯誤在 wording catalog 缺失時仍 fail-closed 地
// 顯示診斷，不把原版規則判定改成另一條硬編 UI 分支。
func (a *app) ceasefireText(key, fallback string) string {
	text, err := a.wordingText(key)
	if err != nil {
		return fallback
	}
	return text
}

// loanText 讓信用度為零的原版早期 gate 在兩種用語模式都使用語系資料；
// 深層錯誤不應暗中退回另一種語言，因此只回傳明確的短 fallback。
func (a *app) loanText(key, fallback string) string {
	text, err := a.wordingText(key)
	if err != nil {
		return fallback
	}
	return text
}

// executeDiplomacyAid 是 M1 外援的唯一 UI→規則接點。第一期援助國代碼現在依
// 原版 `sub_21D1D` 的勢力反查值與同一次 Random(10) 選擇；沒有 `.DT1` 勢力表時
// 傳 0 走原版 default 分支，不把缺資料誤命名成某個國家。第二／三期仍由原版
// 直接覆成 99。
func (a *app) executeDiplomacyAid() string {
	if a.cmdBudget == nil {
		return "外援需要有效的指令數狀態"
	}
	if a.world == nil || a.tbl == nil {
		return "外援需要有效的遊戲世界"
	}
	if a.rng == nil {
		return "外援需要可重現的亂數來源"
	}
	prov, err := a.tbl.At(a.current)
	if err != nil {
		return "外援失敗：" + err.Error()
	}
	base, err := a.world.AidResourceBases(a.current)
	if err != nil {
		return "外援失敗：" + err.Error()
	}
	// 原版在第一次援助亂數前就立起完成旗標；因此核准、隨機拒絕與特殊核准分支
	// 都要扣指令。這兩份快照讓防禦性提交失敗時，省份與亂數仍可原子回復。
	before, seed := *prov, a.rng.Seed()
	factionCode := 0
	if slot := a.factions.SlotOfLeader(prov.Commander); slot >= 0 {
		// FactionTable 與原版區塊 7 都是 1-based 對外代碼。
		factionCode = slot + 1
	}
	res, err := a.world.RequestAidForFaction(a.current, game.GameState{
		Stage: uint8(a.stage), Year: uint8(a.year), Month: a.month,
	}, factionCode, base, a.rng)
	if err != nil {
		return "外援失敗：" + err.Error()
	}
	if !res.Approved {
		if res.CommandCompleted && !a.cmdBudget.Spend(a.current) {
			*prov = before
			a.rng.SetSeed(seed)
			return "指令數已用完，外援結果未能提交"
		}
		msg, msgErr := a.wordingText("diplomacy.aid.refused")
		if msgErr != nil {
			return "外援遭拒：" + msgErr.Error()
		}
		return msg
	}
	if !a.cmdBudget.Spend(a.current) {
		*prov = before
		a.rng.SetSeed(seed)
		return "指令數已用完，外援狀態未能提交"
	}
	msg, msgErr := a.wordingFormat("diplomacy.aid.result", res.Gold, res.Food,
		res.Ammo, res.Fuel, a.cmdBudget.Remaining(a.current))
	if msgErr != nil {
		return "外援已完成，但結果用語缺失：" + msgErr.Error()
	}
	return msg
}

// currentDiplomacySlot 把目前省份的司令反查成 .DT1 外交帳本的 1-based
// 勢力索引。勢力表是存檔後半才有的資料；以 TOWN(N).DAT 開新局時沒有
// 可安全推導的外交槽位，必須 fail-closed，而不是把第 0 格當成第一國。
func (a *app) currentDiplomacySlot() (int, error) {
	if a.tbl == nil {
		return 0, fmt.Errorf("外交需要已載入的省份狀態")
	}
	p, err := a.tbl.At(a.current)
	if err != nil {
		return 0, fmt.Errorf("外交無法讀取目前省份：%w", err)
	}
	slot := a.factions.SlotOfLeader(p.Commander)
	if slot < 0 || slot >= game.DiplomacyLedgerSlots {
		return 0, fmt.Errorf("目前司令沒有可用的外交勢力槽（新局初始檔尚未載入外交表）")
	}
	return slot + 1, nil
}

func checkWordingGlyphs(screen string, missing []rune) error {
	if len(missing) != 0 {
		return fmt.Errorf("%s 的顯示用語缺字：%q", screen, string(missing))
	}
	return nil
}

func (a *app) setWordingPreference(mode i18n.WordingMode) error {
	if _, err := i18n.ParseWordingMode(string(mode)); err != nil {
		return err
	}
	if mode == i18n.WordingPlain && (a.wording == nil || a.eten == nil) {
		return fmt.Errorf("現代白話需要完整語系資料與倚天字庫")
	}
	if mode == a.wordingMode && a.preferences.Wording == string(mode) {
		return nil
	}
	next := a.preferences
	next.Wording = string(mode)
	if err := userprefs.Save(a.prefsPath, next); err != nil {
		return err
	}
	a.preferences, a.wordingMode, a.dirty = next, mode, true
	fmt.Fprintf(os.Stderr, "顯示用語已切換為 %s（偏好：%s）\n", mode, a.prefsPath)
	return nil
}

// setThemePreference 以完整的主題／部隊圖示組為單位切換顯示主題。設定頁與
// F2 都走同一入口，確保偏好寫回、provider 檢查與畫面指標交換保持原子性。
func (a *app) setThemePreference(mode uitheme.Mode) error {
	parsed, err := uitheme.ParseMode(string(mode))
	if err != nil {
		return err
	}
	var selected uitheme.Theme
	switch parsed {
	case uitheme.ModeRetro:
		selected = a.retroTheme
	case uitheme.ModeModern:
		selected = a.modernTheme
	}
	if selected == nil {
		return fmt.Errorf("主題 %s 尚未載入", parsed)
	}
	selectedUnits, ok := selected.(uitheme.UnitProvider)
	if !ok || selectedUnits == nil {
		return fmt.Errorf("主題 %s 的部隊圖示尚未載入", parsed)
	}
	var selectedHUD uitheme.HUDIconProvider
	if parsed == uitheme.ModeModern {
		selectedHUD, ok = selected.(uitheme.HUDIconProvider)
		if !ok || selectedHUD == nil {
			return fmt.Errorf("主題 %s 的 HUD 圖示尚未載入", parsed)
		}
	}
	if parsed == a.themeMode && a.preferences.Theme == string(parsed) {
		return nil
	}
	nextPrefs := a.preferences
	nextPrefs.Theme = string(parsed)
	if err := userprefs.Save(a.prefsPath, nextPrefs); err != nil {
		return err
	}
	a.preferences, a.themeMode = nextPrefs, parsed
	a.battlefieldTheme, a.unitTheme, a.hudIcons = selected, selectedUnits, selectedHUD
	a.dirty = true
	fmt.Fprintf(os.Stderr, "顯示主題已切換為 %s（偏好：%s）\n", parsed, a.prefsPath)
	return nil
}

// toggleTheme 在 retro／modern 之間循環。兩套主題都在 run 初始化時完整
// 建好，因此這裡只替換指標；不會出現半載入的圖集或改變規則狀態。
func (a *app) toggleTheme() error {
	next := uitheme.ModeModern
	if a.themeMode == uitheme.ModeModern {
		next = uitheme.ModeRetro
	}
	return a.setThemePreference(next)
}

// setResolutionPreference 切換顯示畫布並以偏好檔保存；它不寫入原版存檔。
// 原版模式的實體桌面視窗仍保留目前 2 倍放大，邏輯畫布是 640×350；高解析
// 模式則使用 1280×720 設計畫布。高解析過渡期的 frame 仍以等比安全區繪製，
// 不把黑邊座標誤派成按鈕命中。
func (a *app) setResolutionPreference(mode uiresolution.Mode) error {
	parsed, err := uiresolution.Parse(string(mode))
	if err != nil {
		return err
	}
	if parsed == a.resolution && a.preferences.Resolution == string(parsed) {
		return nil
	}
	next := a.preferences
	next.Resolution = string(parsed)
	if err := userprefs.Save(a.prefsPath, next); err != nil {
		return err
	}
	a.preferences, a.resolution = next, parsed
	a.pointer.cancel()
	a.dirty = true
	if a.frame != nil {
		a.applyWindowSize()
	}
	fmt.Fprintf(os.Stderr, "顯示解析度已切換為 %s（偏好：%s）\n", parsed, a.prefsPath)
	return nil
}

func (a *app) toggleResolution() error {
	return a.setResolutionPreference(uiresolution.Toggle(a.resolution))
}

func (a *app) applyWindowSize() {
	if a.resolution == uiresolution.ModeHigh {
		ebiten.SetWindowSize(uiresolution.HighWidth, uiresolution.HighHeight)
		return
	}
	ebiten.SetWindowSize(uiresolution.OriginalWidth*scale, uiresolution.OriginalHeight*scale)
}

func (a *app) setMessageTimePreference(units int) error {
	if units < 1 || units > 10 {
		return fmt.Errorf("訊息時間要在 1..10")
	}
	next := a.preferences
	next.MessageTime = units
	if err := userprefs.Save(a.prefsPath, next); err != nil {
		return err
	}
	a.preferences = next
	if a.messages == nil {
		a.messages = newMessageQueue(units)
	} else {
		a.messages.SetUnits(units)
	}
	a.dirty = true
	return nil
}

// autosave 把當前 DT1State 的已解欄位寫回一份**副本**。
//
// AGENTS.md §8：原版資產唯讀，測試存檔一律寫到明確的輸出目錄，
// 不覆蓋原版的 SAVE(1).DT1。寫回是「改寫」不是「重建」——
// 未解區域一個 byte 都不動（internal/game/save.go）。
func (a *app) autosave() error {
	if a.savePath == "" || a.origSave == nil {
		return nil // 沒有存檔來源（用初始檔開的），不寫
	}
	// 正常結算已即時寫過；這裡再走同一入口，涵蓋玩家在結算畫面按
	// F10、或未來新增結算分支卻忘了掛接即時寫回的情況。
	if a.battle != nil && a.battle.finished {
		if err := a.persistBattleSnapshot(); err != nil {
			return err
		}
	}
	if err := a.syncBattleForces(); err != nil {
		return err
	}
	if a.world == nil {
		return fmt.Errorf("DT1 autosave 需要有效世界快照")
	}
	state := game.DT1State{
		Provinces:         a.tbl,
		Generals:          a.generals,
		Factions:          a.factions,
		FactionLeaders:    a.factionLeaders,
		WarRecords:        a.warRecords,
		CeasefireStates:   a.world.CeasefireState,
		MajorPowerLeaders: a.majorPowerLeaders,
		Ledger:            a.ledger,
	}
	out, err := game.WriteDT1(a.origSave, state)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(a.savePath), 0o755); err != nil {
		return err
	}
	if err := writeFileAtomic(a.savePath, out, 0o644); err != nil {
		return err
	}
	a.origSave = out
	return nil
}

// persistBattleSnapshot 把已結束戰鬥的目前狀態寫到玩家副本的 DT2 與
// MEM_WAR。這是 remake 的持久化入口，不宣稱等同原版每一條
// sub_3964E 呼叫時機；立即撤退不會呼叫本函式。
func (a *app) persistBattleSnapshot() error {
	b := a.battle
	if b == nil || !b.finished || b.sim == nil {
		return nil
	}
	if a.battleDT2Path == "" || a.battleMemWarPath == "" {
		// 呼叫端明確沒有 -save 輸出目標時，保留可遊玩但不宣稱持久化。
		return nil
	}
	// 新局可能只有 TOWN(1).DAT，沒有戰鬥副本。這時保持可遊玩，但不能
	// 假裝已完成檔案持久化；正常有來源的遊戲則要求兩份一起存在。
	if a.battleDT2 == nil && a.battleMemWar == nil {
		return nil
	}
	if a.battleDT2 == nil || a.battleMemWar == nil {
		return fmt.Errorf("戰鬥寫回需要同時載入 DT2 與 MEM_WAR.DAT（DT2=%t、MEM_WAR=%t）",
			a.battleDT2 != nil, a.battleMemWar != nil)
	}
	if !b.sim.At.Valid() {
		return fmt.Errorf("戰場省 %d 無法寫回戰鬥狀態", b.sim.At)
	}
	if err := a.syncBattleForces(); err != nil {
		return err
	}
	idx := int(b.sim.At) - 1
	attackers := battleSnapshotIDs(b.sim.Attacker)
	defenders := battleSnapshotIDs(b.sim.Defender)
	resources := game.BattleResources{
		Gold: uint16(clampBattleResource(b.supAtk.Gold)),
		Food: uint16(clampBattleResource(b.supAtk.Food)),
		Ammo: uint16(clampBattleResource(b.supAtk.Ammo)),
		Fuel: uint16(clampBattleResource(b.supAtk.Fuel)),
	}
	if err := a.battleDT2States[idx].ApplyRemakeSnapshot(b.sim.From, resources, attackers, defenders); err != nil {
		return fmt.Errorf("DT2 第 %d 省：%w", b.sim.At, err)
	}
	if err := a.battleMemStates[idx].ApplyRemakeSnapshot(b.sim.From, resources, attackers, defenders); err != nil {
		return fmt.Errorf("MEM_WAR 第 %d 省：%w", b.sim.At, err)
	}
	dt2, err := game.WriteBattleStates(a.battleDT2, a.battleDT2States)
	if err != nil {
		return fmt.Errorf("產生 DT2 寫回內容：%w", err)
	}
	mem, err := game.WriteBattleStates(a.battleMemWar, a.battleMemStates)
	if err != nil {
		return fmt.Errorf("產生 MEM_WAR 寫回內容：%w", err)
	}
	if err := os.MkdirAll(filepath.Dir(a.battleDT2Path), 0o755); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(a.battleMemWarPath), 0o755); err != nil {
		return err
	}
	if err := writeFileAtomic(a.battleDT2Path, dt2, 0o644); err != nil {
		return fmt.Errorf("寫回 %s：%w", a.battleDT2Path, err)
	}
	if err := writeFileAtomic(a.battleMemWarPath, mem, 0o644); err != nil {
		return fmt.Errorf("寫回 %s：%w", a.battleMemWarPath, err)
	}
	// 下一次結算仍應以最新副本為 Raw 基底，避免連續寫回退回開局 bytes。
	if parsed, parseErr := game.ParseBattleStates(dt2); parseErr != nil {
		return fmt.Errorf("重新解析 DT2 寫回內容：%w", parseErr)
	} else {
		a.battleDT2States = parsed
	}
	if parsed, parseErr := game.ParseBattleStates(mem); parseErr != nil {
		return fmt.Errorf("重新解析 MEM_WAR 寫回內容：%w", parseErr)
	} else {
		a.battleMemStates = parsed
	}
	a.battleDT2, a.battleMemWar = dt2, mem
	fmt.Fprintf(os.Stderr, "[battle] 已寫回 DT2 與 MEM_WAR：省 %d（攻 %d／守 %d）\n",
		b.sim.At, len(attackers), len(defenders))
	return nil
}

func battleSnapshotIDs(units []*game.Combatant) []game.GeneralID {
	out := make([]game.GeneralID, 0, len(units))
	for _, u := range units {
		if u != nil && u.General != 0 {
			out = append(out, u.General)
		}
	}
	return out
}

func clampBattleResource(v int) int {
	if v < 0 {
		return 0
	}
	if v > int(^uint16(0)) {
		return int(^uint16(0))
	}
	return v
}

func writeFileAtomic(path string, data []byte, mode os.FileMode) (err error) {
	dir := filepath.Dir(path)
	f, err := os.CreateTemp(dir, ".dsds-save-*")
	if err != nil {
		return err
	}
	tmp := f.Name()
	defer func() {
		_ = f.Close()
		if err != nil {
			_ = os.Remove(tmp)
		}
	}()
	if err = f.Chmod(mode); err != nil {
		return err
	}
	if _, err = f.Write(data); err != nil {
		return err
	}
	if err = f.Sync(); err != nil {
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func (a *app) compose() error {
	highModernMap := a.highModernMap()
	highModernPage := a.highModernPage()
	highModernSurface := a.highModernSurface()
	c := render.NewBGICanvas()
	if highModernSurface {
		c = render.NewCanvas(uiresolution.HighWidth, uiresolution.HighHeight)
	}
	style := a.uiStyle()
	// 在 compose 內遮蔽 legacy 常數，讓所有子畫面採同一套 modern
	// 語意色；drawBattle 另以 a.uiStyle().Ink 取色，避免跨函式全域狀態。
	panelInk, panelPaper := style.Ink, style.Paper
	var modernLabels render.PanelLabels
	if style.Name == uitheme.ModeModern && a.eten != nil && a.wording != nil {
		var err error
		modernLabels, err = a.panelLabels()
		if err != nil {
			return err
		}
	}

	var currentBattlefield *game.Battlefield
	if a.screen == screenBattle {
		if err := a.drawBattle(c); err != nil {
			return err
		}
	} else {
		bf, err := a.m.Battlefield(a.current)
		if err != nil {
			return err
		}
		currentBattlefield = bf
		if !highModernSurface {
			// 用原版的 NEWTERR 圖塊畫戰場，有鐵路的格子疊 RAIL.TPC。
			if err := c.DrawThemedBattlefield(bf, a.battlefieldTheme, fieldX, fieldY); err != nil {
				return err
			}
		}
	}

	p, err := a.tbl.At(a.current)
	if err != nil {
		return err
	}
	data := render.PanelData{
		ID:       a.current,
		Province: p,
		Force:    game.ForceOf(a.generals, a.current),
		Generals: game.CountOf(a.generals, a.current),
		Commands: a.cmdBudget.Remaining(a.current),
		Icons:    a.hudIcons,
		Style:    style,
	}
	if style.Name == uitheme.ModeModern && a.eten != nil && a.wording != nil {
		data.SemanticFonts = a.eten
		data.Labels = modernLabels
		data.ProvinceName = a.provinceName(a.current)
		data.CommanderName = a.generalDisplayName(p.Commander)
		data.GovernorName = a.generalDisplayName(p.Governor)
		data.Status = modernLabels.Status
		data.Faction = a.currentFaction(p.Commander)
	}
	data.Year, data.Month = a.year, a.month
	if highModernMap {
		narrative := ""
		if a.narrative != nil && len(a.narrativeImages) > 0 && a.wording != nil {
			var err error
			narrative, err = a.wordingText("narrative.open")
			if err != nil {
				return err
			}
		}
		portrait, portraitUnavailable := a.portraitForGeneral(p.Commander)
		if err := c.DrawModernMapSurface(render.ModernMapSurfaceData{
			Battlefield:         currentBattlefield,
			Theme:               a.battlefieldTheme,
			Style:               style,
			Panel:               data,
			Fonts:               a.eten,
			Command:             modernLabels.Commands,
			Narrative:           narrative,
			Portrait:            portrait,
			PortraitUnavailable: portraitUnavailable,
		}); err != nil {
			return err
		}
	} else if !highModernSurface {
		if err := c.DrawStrategyPanel(data, a.fonts); err != nil {
			return err
		}
	}
	// 尚未需要專用插圖的流程頁仍要使用高解析度 Modern surface；在這裡
	// 提前收束，避免後面的 legacy switch 把 640×350 元件畫回高解析畫布。
	if a.highModernFlow() && !highModernPage {
		flow, err := a.modernFlowData()
		if err != nil {
			return err
		}
		if err := c.DrawModernFlowSurface(flow); err != nil {
			return err
		}
		a.frame = ebiten.NewImageFromImage(c.Image())
		a.dirty = false
		return nil
	}

	switch a.screen {
	case screenCommand:
		if highModernPage {
			labels := make([]string, 15)
			for i := range labels {
				var err error
				labels[i], err = a.wordingText(fmt.Sprintf("command.%02d", i+1))
				if err != nil {
					return err
				}
			}
			title, err := a.wordingText("command.title")
			if err != nil {
				return err
			}
			back, _, _, err := a.modernNavigationLabels()
			if err != nil {
				return err
			}
			if err := c.DrawModernCommandSurface(render.ModernCommandSurfaceData{
				Title: title, Labels: labels, Fonts: a.eten, Icons: a.hudIcons, Style: style, Back: back,
			}); err != nil {
				return err
			}
			break
		}
		if a.semanticWording() {
			labels := make([]string, 15)
			for i := range labels {
				key := fmt.Sprintf("command.%02d", i+1)
				var err error
				labels[i], err = a.wordingText(key)
				if err != nil {
					return err
				}
			}
			if err := checkWordingGlyphs("政略主選單", c.DrawSemanticCommandPageWithIcons(a.eten, a.hudIcons, panelInk, panelPaper,
				fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, labels)); err != nil {
				return err
			}
			break
		}
		if err := c.DrawCommandPageWithIcons(a.cmdFonts, a.hudIcons, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY); err != nil {
			return err
		}
	case screenDevelop:
		if a.semanticWording() {
			labels := make([]string, 3)
			for i, key := range []string{"develop.reclaim", "develop.arsenal", "develop.mine"} {
				var err error
				labels[i], err = a.wordingText(key)
				if err != nil {
					return err
				}
			}
			if err := checkWordingGlyphs("發展選單", c.DrawSemanticCommandPage(a.eten, panelInk, panelPaper,
				fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, labels)); err != nil {
				return err
			}
			break
		}
		if err := c.DrawDevelopPage(a.fonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY); err != nil {
			return err
		}
	case screenTransferMode:
		if a.semanticWording() {
			prompt, err := a.wordingText("transfer.mode")
			if err != nil {
				return err
			}
			partial, err := a.wordingText("transfer.mode.partial")
			if err != nil {
				return err
			}
			all, err := a.wordingText("transfer.mode.all")
			if err != nil {
				return err
			}
			missing := c.DrawPlainTransferMode(a.eten, panelInk, panelPaper,
				fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
				prompt, partial, all)
			if err := checkWordingGlyphs("調動方式", missing); err != nil {
				return err
			}
		} else if err := c.DrawPlayerTransferMode(a.cmdFonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY); err != nil {
			return err
		}
	case screenTransferTarget:
		ids := make([]int, len(a.transferTargets))
		for i, id := range a.transferTargets {
			ids[i] = int(id)
		}
		if a.semanticWording() {
			prompt, err := a.wordingText("transfer.target")
			if err != nil {
				return err
			}
			missing, err := c.DrawPlainTransferTarget(a.eten, a.fonts.W3, panelInk, panelPaper,
				fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
				prompt, ids, a.transferInput)
			if err != nil {
				return err
			}
			if err := checkWordingGlyphs("調動目標", missing); err != nil {
				return err
			}
		} else if err := c.DrawPlayerTransferTarget(a.cmdFonts, a.fonts.W3, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			ids, a.transferInput); err != nil {
			return err
		}
	case screenTransferSelection:
		cands := a.transferSession.Candidates()
		selectedIDs := a.transferSession.Selected()
		ids, selected := make([]int, len(cands)), make([]bool, len(cands))
		for i, id := range cands {
			ids[i] = int(id)
			for _, chosen := range selectedIDs {
				selected[i] = selected[i] || chosen == id
			}
		}
		if a.semanticWording() {
			key := "transfer.select.partial"
			if a.transferMode == game.PlayerTransferAll {
				key = "transfer.select.all"
			}
			prompt, err := a.wordingText(key)
			if err != nil {
				return err
			}
			confirm, err := a.wordingText("transfer.selection.confirm")
			if err != nil {
				return err
			}
			missing, err := c.DrawPlainTransferSelection(a.eten, a.fonts.Gen,
				panelInk, panelPaper, fieldX, fieldY,
				render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
				prompt, confirm, ids, selected, a.transferCursor,
				a.transferCursor/20)
			if err != nil {
				return err
			}
			if err := checkWordingGlyphs("調動選將", missing); err != nil {
				return err
			}
		} else if err := c.DrawPlayerTransferSelection(a.cmdFonts, a.fonts.Gen, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			ids, selected, a.transferCursor, a.transferCursor/20, a.transferMode); err != nil {
			return err
		}
	case screenTransferAmount:
		if a.semanticWording() {
			keys := []string{"transfer.resource.gold", "transfer.resource.food",
				"transfer.resource.ammo", "transfer.resource.fuel"}
			good := a.transferGood
			if good < 0 || good >= len(keys) {
				good = 0
			}
			prompt, err := a.wordingText(keys[good])
			if err != nil {
				return err
			}
			missing := c.DrawPlainTransferAmount(a.eten, panelInk, panelPaper,
				fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
				prompt, a.transferInput)
			if err := checkWordingGlyphs("調動物資", missing); err != nil {
				return err
			}
		} else if err := c.DrawSupplyAmount(a.cmdFonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			a.transferGood, a.transferInput); err != nil {
			return err
		}
	case screenTransferConfirm:
		if a.semanticWording() {
			prompt, err := a.wordingText("common.confirm")
			if err != nil {
				return err
			}
			missing := c.DrawPlainConfirm(a.eten, panelInk, panelPaper,
				fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, prompt)
			if err := checkWordingGlyphs("調動確認", missing); err != nil {
				return err
			}
		} else {
			c.DrawConfirmBox(a.cmdFonts.W4, panelInk, panelPaper,
				fieldX+60, fieldY+120)
		}
	case screenTradeMode:
		if a.semanticWording() {
			labels := make([]string, 2)
			for i, key := range []string{"trade.import", "trade.export"} {
				var err error
				labels[i], err = a.wordingText(key)
				if err != nil {
					return err
				}
			}
			if err := checkWordingGlyphs("商業方式", c.DrawSemanticList(a.eten, panelInk, panelPaper, fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, labels)); err != nil {
				return err
			}
			break
		}
		if err := c.DrawTradeMenu(a.cmdFonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, nil); err != nil {
			return err
		}
	case screenTradeGood:
		if a.semanticWording() {
			keys := []string{"trade.food", "trade.ammo", "trade.fuel"}
			if !a.tradeImport {
				keys = []string{"trade.food", "trade.ammo", "trade.coal", "trade.iron", "trade.fuel"}
			}
			labels := make([]string, len(keys))
			for i, key := range keys {
				var err error
				labels[i], err = a.wordingText(key)
				if err != nil {
					return err
				}
			}
			if err := checkWordingGlyphs("商業品項", c.DrawSemanticList(a.eten, panelInk, panelPaper, fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, labels)); err != nil {
				return err
			}
			break
		}
		if err := c.DrawTradeMenu(a.cmdFonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, &a.tradeImport); err != nil {
			return err
		}
	case screenTradeAmount:
		if a.semanticWording() {
			promptKey := "trade.sell_amount"
			if a.tradeImport {
				promptKey = "trade.buy_amount"
			}
			prompt, err := a.wordingText(promptKey)
			if err != nil {
				return err
			}
			goodKeys := []string{"trade.food", "trade.ammo", "trade.fuel", "trade.coal", "trade.iron"}
			gi := int(a.tradeGood)
			if gi < 0 || gi >= len(goodKeys) {
				gi = 0
			}
			good, err := a.wordingText(goodKeys[gi])
			if err != nil {
				return err
			}
			if err := checkWordingGlyphs("商業數量", c.DrawSemanticAmount(a.eten, panelInk, panelPaper, fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, prompt+good+"？", a.tradeAmount)); err != nil {
				return err
			}
			break
		}
		if err := c.DrawTradeAmount(a.cmdFonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			a.tradeImport, int(a.tradeGood), a.tradeAmount); err != nil {
			return err
		}
	case screenSupplyTarget:
		ids := make([]int, len(a.supplyTargets))
		for i, id := range a.supplyTargets {
			ids[i] = int(id)
		}
		if a.semanticWording() {
			prompt, err := a.wordingText("supply.target")
			if err != nil {
				return err
			}
			missing, err := c.DrawSemanticSupplyTarget(a.eten, a.fonts.W3, panelInk, panelPaper, fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, prompt, ids, a.supplyInput)
			if err != nil {
				return err
			}
			if err := checkWordingGlyphs("運補目標", missing); err != nil {
				return err
			}
			break
		}
		if err := c.DrawSupplyTarget(a.cmdFonts, a.fonts.W3, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			ids, a.supplyInput); err != nil {
			return err
		}
	case screenSupplyAmount:
		if a.semanticWording() {
			keys := []string{"supply.gold", "supply.food", "supply.ammo", "supply.fuel"}
			good := a.supplyGood
			if good < 0 || good >= len(keys) {
				good = 0
			}
			prompt, err := a.wordingText(keys[good])
			if err != nil {
				return err
			}
			if err := checkWordingGlyphs("運補數量", c.DrawSemanticAmount(a.eten, panelInk, panelPaper, fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, prompt, a.supplyInput)); err != nil {
				return err
			}
			break
		}
		if err := c.DrawSupplyAmount(a.cmdFonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			a.supplyGood, a.supplyInput); err != nil {
			return err
		}
	case screenRecruitAction:
		if a.semanticWording() {
			labels, err := a.recruitWording()
			if err != nil {
				return err
			}
			missing := c.DrawRecruitSemantic(a.eten, a.fonts.Gen, panelInk, panelPaper, fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, "action", labels, 0, 0, 0, 0, nil, 0, 0)
			if err := checkWordingGlyphs("徵兵", missing); err != nil {
				return err
			}
			break
		}
		if err := c.DrawRecruitAction(a.cmdFonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY); err != nil {
			return err
		}
	case screenRecruitBranch:
		if a.semanticWording() {
			labels, err := a.recruitWording()
			if err != nil {
				return err
			}
			if err := checkWordingGlyphs("徵兵兵種", c.DrawRecruitSemantic(a.eten, a.fonts.Gen, panelInk, panelPaper, fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, "branch", labels, 0, 0, 0, 0, nil, 0, 0)); err != nil {
				return err
			}
			break
		}
		if err := c.DrawRecruitBranch(a.cmdFonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY); err != nil {
			return err
		}
	case screenRecruitAmount:
		if a.semanticWording() {
			labels, err := a.recruitWording()
			if err != nil {
				return err
			}
			if err := checkWordingGlyphs("徵兵數量", c.DrawRecruitSemantic(a.eten, a.fonts.Gen, panelInk, panelPaper, fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, "amount", labels, a.recruitBranch, a.recruitLimit, a.recruitAmount, 0, nil, 0, 0)); err != nil {
				return err
			}
			break
		}
		if err := c.DrawRecruitAmount(a.cmdFonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			a.recruitBranch, a.recruitLimit, a.recruitAmount); err != nil {
			return err
		}
	case screenRecruitConfirm:
		if a.semanticWording() {
			labels, err := a.recruitWording()
			if err != nil {
				return err
			}
			if err := checkWordingGlyphs("徵兵確認", c.DrawRecruitSemantic(a.eten, a.fonts.Gen, panelInk, panelPaper, fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, "confirm", labels, a.recruitBranch, 0, 0, uint32(game.RecruitCost(a.recruitBranch, int(a.recruitAmount))), nil, 0, 0)); err != nil {
				return err
			}
			break
		}
		if err := c.DrawRecruitConfirm(a.cmdFonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			uint32(game.RecruitCost(a.recruitBranch, int(a.recruitAmount)))); err != nil {
			return err
		}
	case screenReorganizeBranch:
		if a.semanticWording() {
			labels, err := a.recruitWording()
			if err != nil {
				return err
			}
			if err := checkWordingGlyphs("整編兵種", c.DrawRecruitSemantic(a.eten, a.fonts.Gen, panelInk, panelPaper, fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, "branch", labels, 0, 0, 0, 0, nil, 0, 0)); err != nil {
				return err
			}
			break
		}
		if err := c.DrawRecruitBranch(a.cmdFonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY); err != nil {
			return err
		}
	case screenReorganizeTarget:
		targets := a.reorganization.Targets()
		ids := make([]int, len(targets))
		for i, id := range targets {
			ids[i] = int(id)
		}
		if a.semanticWording() {
			labels, err := a.recruitWording()
			if err != nil {
				return err
			}
			if err := checkWordingGlyphs("整編對象", c.DrawRecruitSemantic(a.eten, a.fonts.Gen, panelInk, panelPaper, fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, "targets", labels, 0, 0, a.reorganizeInput, 0, ids, a.reorganization.Remaining(), 0)); err != nil {
				return err
			}
			break
		}
		if err := c.DrawReorganizationTarget(a.cmdFonts, a.fonts.Gen, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			ids, a.reorganization.Remaining(), a.reorganizeInput); err != nil {
			return err
		}
	case screenReorganizeAmount:
		if a.semanticWording() {
			labels, err := a.recruitWording()
			if err != nil {
				return err
			}
			if err := checkWordingGlyphs("整編兵力", c.DrawRecruitSemantic(a.eten, a.fonts.Gen, panelInk, panelPaper, fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, "reorganize-amount", labels, 0, a.reorganization.Limit(a.reorganizeID), a.reorganizeInput, 0, nil, a.reorganization.Remaining(), int(a.reorganizeID))); err != nil {
				return err
			}
			break
		}
		if err := c.DrawReorganizationAmount(a.cmdFonts, a.fonts.Gen, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			int(a.reorganizeID), a.reorganization.Remaining(),
			a.reorganization.Limit(a.reorganizeID), a.reorganizeInput); err != nil {
			return err
		}
	case screenTrainConfirm:
		if a.semanticWording() {
			prompt, err := a.wordingText("train.confirm")
			if err != nil {
				return err
			}
			if err := checkWordingGlyphs("練兵確認", c.DrawSemanticConfirm(a.eten, panelInk, panelPaper,
				fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, prompt)); err != nil {
				return err
			}
			break
		}
		if err := c.DrawTrainConfirm(a.cmdFonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY); err != nil {
			return err
		}
	case screenCovertAction:
		if a.semanticWording() {
			labels := make([]string, 2)
			for i, key := range []string{"covert.guerrilla", "covert.student"} {
				var err error
				labels[i], err = a.wordingText(key)
				if err != nil {
					return err
				}
			}
			if err := checkWordingGlyphs("秘密行動", c.DrawSemanticList(a.eten, panelInk, panelPaper,
				fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, labels)); err != nil {
				return err
			}
			break
		}
		if err := c.DrawCovertAction(a.cmdFonts, a.fonts.W3, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY); err != nil {
			return err
		}
	case screenCovertTarget:
		if a.semanticWording() {
			prompt, err := a.wordingText("covert.student.target")
			if err != nil {
				return err
			}
			if err := checkWordingGlyphs("學潮目標", c.DrawSemanticAmount(a.eten, panelInk, panelPaper,
				fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, prompt, a.covertInput)); err != nil {
				return err
			}
			break
		}
		if err := c.DrawCovertTarget(a.cmdFonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			a.covertAction, a.covertInput); err != nil {
			return err
		}
	case screenDiplomacy:
		keys := []string{"diplomacy.loan", "diplomacy.aid", "diplomacy.repay"}
		labels := make([]string, len(keys))
		for i, key := range keys {
			label, err := a.wordingText(key)
			if err != nil {
				return err
			}
			labels[i] = label
		}
		missing := c.DrawDiplomacyMenu(a.eten, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, labels)
		if err := checkWordingGlyphs("外交選單", missing); err != nil {
			return err
		}
	case screenCeasefireTarget:
		limit := a.provinceLimit
		if limit <= 0 {
			limit = game.CeasefireProvinceLimit(uint8(a.stage))
		}
		if a.semanticWording() {
			prompt, err := a.wordingText("ceasefire.prompt")
			if err != nil {
				return err
			}
			rangeText, err := a.wordingFormat("ceasefire.range", limit)
			if err != nil {
				return err
			}
			missing := c.DrawDiplomacyAmount(a.eten, panelInk, panelPaper,
				fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
				prompt, rangeText, a.ceasefireInput)
			if err := checkWordingGlyphs("停火目標", missing); err != nil {
				return err
			}
			break
		}
		if err := c.DrawCeasefireTarget(a.fonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			a.ceasefireInput); err != nil {
			return err
		}
	case screenLoanAmount:
		prompt, err := a.wordingText("diplomacy.loan.prompt")
		if err != nil {
			return err
		}
		creditLabel, err := a.wordingText("diplomacy.loan.credit")
		if err != nil {
			return err
		}
		slot := a.diplomacySlot
		if slot < 1 || slot > game.DiplomacyLedgerSlots {
			slot, err = a.currentDiplomacySlot()
			if err != nil {
				return err
			}
		}
		missing := c.DrawDiplomacyLoan(a.eten, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			prompt, creditLabel, a.ledger.Credit[slot], a.diplomacyInput)
		if err := checkWordingGlyphs("貸款額度", missing); err != nil {
			return err
		}
	case screenRepayAmount:
		prompt, err := a.wordingText("diplomacy.repay.prompt")
		if err != nil {
			return err
		}
		slot := a.diplomacySlot
		if slot < 1 || slot > game.DiplomacyLedgerSlots {
			slot, err = a.currentDiplomacySlot()
			if err != nil {
				return err
			}
		}
		prov, err := a.tbl.At(a.current)
		if err != nil {
			return err
		}
		debt, detailErr := a.wordingFormat("diplomacy.repay.debt",
			a.ledger.Debt[slot], prov.Gold)
		if detailErr != nil {
			return detailErr
		}
		missing := c.DrawDiplomacyAmount(a.eten, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			prompt, debt, a.diplomacyInput)
		if err := checkWordingGlyphs("償還外債額度", missing); err != nil {
			return err
		}
	case screenViewGenerals:
		ids := make([]int, len(a.viewGenerals))
		forces := make([]uint16, len(a.viewGenerals))
		for i, id := range a.viewGenerals {
			ids[i] = int(id)
			if gi := int(id) - 1; gi >= 0 && gi < len(a.generals) {
				forces[i] = a.generals[gi].Force
			}
		}
		if err := c.DrawGeneralList(a.cmdFonts, a.fonts.Gen, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			ids, forces, a.viewIndex); err != nil {
			return err
		}
	case screenViewMenu:
		if a.semanticWording() {
			labels := make([]string, 4)
			for i, key := range []string{"view.other", "view.owned", "view.generals", "view.province_names"} {
				var err error
				labels[i], err = a.wordingText(key)
				if err != nil {
					return err
				}
			}
			if err := checkWordingGlyphs("查閱選單", c.DrawSemanticList(a.eten, panelInk, panelPaper,
				fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, labels)); err != nil {
				return err
			}
			break
		}
		if err := c.DrawViewMenu(a.cmdFonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY); err != nil {
			return err
		}
	case screenViewProvinceSelect:
		if a.semanticWording() {
			prompt, err := a.wordingText("view.select_prompt")
			if err != nil {
				return err
			}
			missing, err := c.DrawSemanticProvinceSelect(a.eten, a.fonts.W3, panelInk, panelPaper,
				fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, prompt, a.viewInput)
			if err != nil {
				return err
			}
			if err := checkWordingGlyphs("查閱省份", missing); err != nil {
				return err
			}
			break
		}
		if err := c.DrawProvinceSelect(a.cmdFonts, a.fonts.W3, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, a.viewInput); err != nil {
			return err
		}
	case screenViewProvinceChoice:
		if a.semanticWording() {
			labels := make([]string, 2)
			for i, key := range []string{"view.choice.overview", "view.choice.generals"} {
				var err error
				labels[i], err = a.wordingText(key)
				if err != nil {
					return err
				}
			}
			missing, err := c.DrawSemanticProvinceChoice(a.eten, a.fonts.W3, panelInk, panelPaper,
				fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, int(a.viewProvince), labels)
			if err != nil {
				return err
			}
			if err := checkWordingGlyphs("查閱省份選項", missing); err != nil {
				return err
			}
			break
		}
		if err := c.DrawProvinceChoice(a.cmdFonts, a.fonts.W3, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			int(a.viewProvince)); err != nil {
			return err
		}
	case screenViewProvince:
		vp, err := a.tbl.At(a.viewProvince)
		if err != nil {
			return err
		}
		data := render.PanelData{ID: a.viewProvince, Province: vp,
			Force:    game.ForceOf(a.generals, a.viewProvince),
			Generals: game.CountOf(a.generals, a.viewProvince)}
		if err := c.DrawProvinceDetail(a.cmdFonts, a.fonts, data, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY); err != nil {
			return err
		}
	case screenViewOverview:
		provinces := a.world.OwnedProvinces(a.current)
		ids := make([]int, len(provinces))
		forces := make([]uint32, len(provinces))
		for i, id := range provinces {
			ids[i] = int(id)
			forces[i] = game.ForceOf(a.generals, id)
		}
		if a.semanticWording() {
			title, err := a.wordingText("view.owned.title")
			if err != nil {
				return err
			}
			missing, err := c.DrawSemanticOwnedProvinceOverview(a.eten, a.fonts.W3, panelInk, panelPaper,
				fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, title, ids, forces)
			if err != nil {
				return err
			}
			if err := checkWordingGlyphs("所屬省份概況", missing); err != nil {
				return err
			}
			break
		}
		if err := c.DrawOwnedProvinceOverview(a.cmdFonts, a.fonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, ids, forces); err != nil {
			return err
		}
	case screenViewProvinceNames:
		if a.semanticWording() {
			title, err := a.wordingText("view.names.title")
			if err != nil {
				return err
			}
			missing, err := c.DrawSemanticProvinceNames(a.eten, a.fonts.W3, panelInk, panelPaper,
				fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
				title, a.viewPage, a.provinceLimit)
			if err != nil {
				return err
			}
			if err := checkWordingGlyphs("省份編號對照", missing); err != nil {
				return err
			}
			break
		}
		if err := c.DrawProvinceNames(a.cmdFonts, a.fonts.W3, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			a.viewPage, a.provinceLimit); err != nil {
			return err
		}
	case screenViewGeneral:
		id := a.viewGenerals[a.viewIndex]
		gi := int(id) - 1
		if gi >= 0 && gi < len(a.generals) && gi < len(a.world.Strengths) {
			attack := game.Strength(a.world.Strengths[gi], a.world.Opts)
			if err := c.DrawGeneralDetail(a.cmdFonts, a.fonts, a.fan, panelInk, panelPaper,
				fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
				int(id), a.generals[gi], attack); err != nil {
				return err
			}
		}
	case screenBiography:
		p, ok := a.currentBiography()
		if !ok {
			return fmt.Errorf("人物自傳入口狀態失效")
		}
		title, okTitle := a.wording.Text("biography.page", a.wordingMode)
		unavailable, okUnavailable := a.wording.Text("biography.unavailable", a.wordingMode)
		if !okTitle || !okUnavailable {
			return fmt.Errorf("人物自傳用語資料不完整")
		}
		sourceNotice := ""
		if p.BiographyStatus == "source-fallback" {
			sourceNotice, ok = a.wording.Text("biography.source_fallback", a.wordingMode)
			if !ok {
				return fmt.Errorf("人物自傳來源提示資料不完整")
			}
		}
		if highModernPage {
			back, previous, next, navErr := a.modernNavigationLabels()
			if navErr != nil {
				return navErr
			}
			portrait, portraitUnavailable := a.portraitForPerson(p)
			result, err := c.DrawModernBiographySurface(a.eten, render.BiographyView{Person: p,
				Page: a.bioPage, Title: title, Unavailable: unavailable,
				SourceNotice: sourceNotice}, panelInk, panelPaper, style, portrait, portraitUnavailable, back, previous, next)
			if err != nil {
				return err
			}
			if len(result.Missing) != 0 {
				fmt.Fprintf(os.Stderr, "Modern 人物自傳缺字（%s）：%q\n", p.NameInGame, string(result.Missing))
			}
			break
		}
		result, err := c.DrawBiography(a.eten, render.BiographyView{Person: p,
			Page: a.bioPage, Title: title, Unavailable: unavailable,
			SourceNotice: sourceNotice}, panelInk, panelPaper)
		if err != nil {
			return err
		}
		if len(result.Missing) != 0 {
			fmt.Fprintf(os.Stderr, "人物自傳缺字（%s）：%q\n", p.NameInGame, string(result.Missing))
		}
	case screenNarrative:
		title := "News archive"
		if a.wording != nil {
			var err error
			title, err = a.wordingText("narrative.title")
			if err != nil {
				return err
			}
		}
		var err error
		if highModernPage {
			back, previous, next, navErr := a.modernNavigationLabels()
			if navErr != nil {
				return navErr
			}
			err = c.DrawModernNarrativeSurface(a.eten, a.narrative, a.narrativeImages,
				a.narrativePage, title, a.wordingMode, style, back, previous, next)
		} else {
			err = c.DrawNarrativeGallery(a.eten, a.narrative, a.narrativeImages,
				a.narrativePage, title, a.wordingMode, style)
		}
		if err != nil {
			return err
		}
	case screenPolicy:
		keys := []string{"policy.title", "policy.autonomy", "policy.production",
			"policy.production.unavailable"}
		labels := make([]string, len(keys))
		for i, key := range keys {
			label, err := a.wordingText(key)
			if err != nil {
				return err
			}
			labels[i] = label
		}
		missing := c.DrawPolicyMenu(a.eten, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			labels[0], labels[1], labels[2], labels[3])
		if err := checkWordingGlyphs("政策", missing); err != nil {
			return err
		}
	case screenAutonomy:
		keys := []string{"autonomy.title", "autonomy.normal", "autonomy.enabled", "autonomy.prompt"}
		labels := make([]string, len(keys))
		for i, key := range keys {
			label, err := a.wordingText(key)
			if err != nil {
				return err
			}
			labels[i] = label
		}
		rows := make([]render.AutonomyRow, 0, len(a.autonomyTargets))
		for _, id := range a.autonomyTargets {
			prov, err := a.tbl.At(id)
			if err != nil {
				return err
			}
			rows = append(rows, render.AutonomyRow{Province: int(id),
				Name: a.provinceName(id), Autonomous: prov.Autonomous()})
		}
		missing := c.DrawAutonomyPolicy(a.eten, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			labels[0], labels[1], labels[2], labels[3], rows, a.autonomyInput)
		if err := checkWordingGlyphs("授權自治", missing); err != nil {
			return err
		}
	case screenProduction:
		keys := []string{"production.title", "production.gold", "production.iron", "production.coal", "production.oil", "production.food"}
		labels := make([]string, len(keys))
		for i, key := range keys {
			var err error
			labels[i], err = a.wordingText(key)
			if err != nil {
				return err
			}
		}
		promptKey := "production.select"
		if a.productionItem != 0 {
			promptKey = "production.value"
		}
		prompt, err := a.wordingText(promptKey)
		if err != nil {
			return err
		}
		p, err := a.tbl.At(a.current)
		if err != nil {
			return err
		}
		pa := p.ProductionAllocation()
		missing := c.DrawProductionPolicy(a.eten, panelInk, panelPaper, fieldX, fieldY,
			render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, labels[0],
			[5]string{labels[1], labels[2], labels[3], labels[4], labels[5]},
			[5]uint8{pa.Gold(), pa.Iron, pa.Coal, pa.Oil, pa.Food}, a.productionItem, prompt, a.productionInput)
		if err := checkWordingGlyphs("產能分配", missing); err != nil {
			return err
		}
	case screenDisplayOptions:
		keys := []string{"settings.title", "settings.wording",
			"settings.wording.original", "settings.wording.plain",
			"settings.theme", "settings.theme.retro", "settings.theme.modern"}
		labels := make([]string, len(keys))
		for i, key := range keys {
			label, err := a.wordingText(key)
			if err != nil {
				return err
			}
			labels[i] = label
		}
		selected := 0
		if a.semanticWording() {
			selected = 1
		}
		themeSelected := 0
		if a.themeMode == uitheme.ModeModern {
			themeSelected = 1
		}
		if highModernPage {
			back, _, _, err := a.modernNavigationLabels()
			if err != nil {
				return err
			}
			if err := c.DrawModernOptionsSurface(render.ModernOptionSurfaceData{
				Title: labels[0], Options: []string{labels[2], labels[3], labels[5], labels[6]},
				Selected: []bool{selected == 0, selected == 1, themeSelected == 0, themeSelected == 1},
				Fonts:    a.eten, Style: style, Back: back,
			}); err != nil {
				return err
			}
			break
		}
		missing := c.DrawDisplayOptions(a.eten, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			labels[0], labels[1], labels[2], labels[3], labels[4], labels[5], labels[6],
			selected, themeSelected)
		if err := checkWordingGlyphs("顯示設定", missing); err != nil {
			return err
		}
	case screenResolutionOptions:
		keys := []string{"settings.resolution", "settings.resolution.original", "settings.resolution.high"}
		labels := make([]string, len(keys))
		for i, key := range keys {
			label, err := a.wordingText(key)
			if err != nil {
				return err
			}
			labels[i] = label
		}
		selected := 0
		if a.resolution == uiresolution.ModeHigh {
			selected = 1
		}
		if highModernPage {
			back, _, _, err := a.modernNavigationLabels()
			if err != nil {
				return err
			}
			if err := c.DrawModernOptionsSurface(render.ModernOptionSurfaceData{
				Title: labels[0], Options: []string{labels[1], labels[2]},
				Selected: []bool{selected == 0, selected == 1}, Fonts: a.eten, Style: style, Back: back,
			}); err != nil {
				return err
			}
			break
		}
		missing := c.DrawResolutionOptions(a.eten, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			labels[0], labels[1], labels[2], selected)
		if err := checkWordingGlyphs("解析度設定", missing); err != nil {
			return err
		}
	case screenOtherOptions:
		keys := []string{"other.save", "other.load", "other.sound", "other.command_art",
			"other.music", "other.message_time", "other.spectate", "other.quit", "other.display",
			"other.resolution"}
		labels := make([]string, len(keys))
		for i, key := range keys {
			label, err := a.wordingText(key)
			if err != nil {
				return err
			}
			labels[i] = label
		}
		labels[5] = fmt.Sprintf("%s %d", labels[5], a.preferences.MessageTime)
		unavailable, err := a.wordingText("other.unavailable")
		if err != nil {
			return err
		}
		available := []bool{a.savePath != "" && a.origSave != nil, a.savePath != "",
			false, false, false, true, false, true, true, true}
		if highModernPage {
			otherTitle, titleErr := a.wordingText("other.title")
			if titleErr != nil {
				return titleErr
			}
			back, _, _, navErr := a.modernNavigationLabels()
			if navErr != nil {
				return navErr
			}
			if err := c.DrawModernOptionsSurface(render.ModernOptionSurfaceData{
				Title: otherTitle, Options: labels, Fonts: a.eten, Style: style, Back: back,
			}); err != nil {
				return err
			}
			break
		}
		missing := c.DrawOtherOptions(a.eten, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			labels, available, unavailable)
		if err := checkWordingGlyphs("其他選項", missing); err != nil {
			return err
		}
	case screenMessageTime:
		prompt, err := a.wordingText("other.message_time.prompt")
		if err != nil {
			return err
		}
		missing := c.DrawSemanticAmount(a.eten, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			prompt, a.messageTimeInput)
		if err := checkWordingGlyphs("訊息時間", missing); err != nil {
			return err
		}
	case screenSaveConfirm, screenLoadConfirm:
		key := "other.save.confirm"
		if a.screen == screenLoadConfirm {
			key = "other.load.confirm"
		}
		prompt, err := a.wordingText(key)
		if err != nil {
			return err
		}
		missing := c.DrawSemanticConfirm(a.eten, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, prompt)
		if err := checkWordingGlyphs("儲存／載入確認", missing); err != nil {
			return err
		}
	case screenQuit:
		// 離開確認。用原版詞表的「您確定嗎」（4.15 詞條 0）。
		c.DrawConfirmBox(a.cmdFonts.W4, panelInk, panelPaper,
			fieldX+60, fieldY+120)
	}
	// 排除槽（例如第一期第 274 格「無省長」）不公告虛假的自傳入口；
	// 有人物但正文為空的 unknown 槽位仍會顯示入口，交給自傳頁的
	// 「查無可靠傳記記載」fallback，避免把「沒有這個人」與「資料未定」混為一談。
	if a.screen == screenViewGeneral && a.eten != nil && a.wording != nil {
		if _, ok := a.currentBiography(); ok {
			label, ok := a.wording.Text("biography.page", a.wordingMode)
			if !ok {
				return fmt.Errorf("人物自傳入口用語資料不完整")
			}
			missing := c.DrawBiographyButton(a.eten, label, panelInk, panelPaper,
				render.ModeBGIW, a.navigationY())
			if err := checkWordingGlyphs("人物自傳入口", missing); err != nil {
				return err
			}
		}
	}
	if a.screen == screenMap && !highModernMap {
		if err := c.DrawOpenCommandButton(a.cmdFonts, panelInk, panelPaper, render.ModeBGIW); err != nil {
			return err
		}
		if a.narrative != nil && len(a.narrativeImages) > 0 {
			label := "News"
			if a.wording != nil {
				var err error
				label, err = a.wordingText("narrative.open")
				if err != nil {
					return err
				}
			}
			if missing := c.DrawNarrativeButton(a.eten, label, panelInk, panelPaper, render.ModeBGIW); len(missing) != 0 {
				return fmt.Errorf("史事新聞入口缺字：%q", string(missing))
			}
		}
	}
	if !highModernSurface {
		for i, action := range a.navigationActions() {
			if action == actions.Submit {
				c.DrawSubmitButton(panelInk, panelPaper, render.ModeBGIW, a.navigationY(), i)
			} else {
				c.DrawNavigationButton(panelInk, panelPaper, render.ModeBGIW, a.navigationY(), i,
					action == actions.NextPage)
			}
		}
	}
	if a.numericKeypadVisible() && !highModernSurface {
		c.DrawNumericKeypad(panelInk, panelPaper)
	}
	if a.messages != nil && a.messages.Active() && !highModernSurface {
		missing := c.DrawMessageOverlay(a.eten, panelInk, panelPaper,
			fieldX+8, render.ModeBGIH-92, render.ModeBGIW-fieldX-16, 82, a.messages.Current())
		if err := checkWordingGlyphs("結果訊息", missing); err != nil {
			return err
		}
	}
	if style.Name == uitheme.ModeModern && !highModernSurface {
		renderWidth := render.ModeBGIW - fieldX
		renderHeight := render.ModeBGIH - fieldY
		c.DrawModernShell(style, fieldX, fieldY, renderWidth, renderHeight)
	}

	a.frame = ebiten.NewImageFromImage(c.Image())
	a.dirty = false
	return nil
}

func (a *app) Draw(dst *ebiten.Image) {
	if a.dirty || a.frame == nil {
		if err := a.compose(); err != nil {
			fmt.Fprintln(os.Stderr, "合成失敗:", err)
			return
		}
	}
	// 合成失敗時不得再解參照 nil frame，否則二次 panic 會遮掉
	// 真正的缺字／資產診斷。
	if a.frame == nil {
		return
	}
	if a.resolution == uiresolution.ModeHigh {
		if bounds := a.frame.Bounds(); bounds.Dx() == uiresolution.HighWidth &&
			bounds.Dy() == uiresolution.HighHeight {
			// H1-a 地圖頁已直接以設計畫布合成，不再二次縮放。
			dst.DrawImage(a.frame, nil)
			return
		}
		// H0 過渡橋接：先把既有 640×350 frame 等比放進 1280×720 設計畫布，
		// 並把高解析 Surface 的輸入反算回原版命中座標。這不是最終 Modern
		// renderer；H1 會以 1280×720 layout／字型替換此橋接，而不改規則層。
		offX, offY, drawnW, drawnH, ok := uiresolution.BaseToSurface(0, 0,
			uiresolution.HighWidth, uiresolution.HighHeight)
		if !ok {
			return
		}
		op := &ebiten.DrawImageOptions{}
		op.GeoM.Scale(float64(drawnW)/float64(uiresolution.OriginalWidth),
			float64(drawnH)/float64(uiresolution.OriginalHeight))
		op.GeoM.Translate(float64(offX), float64(offY))
		op.Filter = ebiten.FilterLinear
		dst.DrawImage(a.frame, op)
		return
	}
	dst.DrawImage(a.frame, nil)
}

func (a *app) Layout(_, _ int) (int, int) {
	return a.resolution.Size()
}

func main() {
	gameDir := flag.String("game", "workplace/orig/game", "原版素材目錄（唯讀）")
	start := flag.Int("province", 26, "起始省編號（1-39），預設 26 = 湖北省")
	// AGENTS.md §8：原版資產唯讀，存檔一律寫到別的地方。
	savePath := flag.String("save", "workplace/saves/SAVE(1).DT1",
		"離開時自動存檔的路徑（**不會**覆蓋原版）")
	// 固定亂數種子是 AGENTS.md §8 的硬規則：截圖驗收要能重現。
	seed := flag.Uint("seed", 1, "亂數種子（原版 LCG，docs/re/17）")
	localeDir := flag.String("locale", "translations/zh-Hant",
		"語系資料目錄。換一個目錄就換一種語言（AGENTS.md §6）")
	etenDir := flag.String("eten", "workplace/eten",
		"使用者提供的倚天 STDFONT.15／SPCFONT.15／ASCFONT.15 目錄（不隨遊戲散布）")
	wording := flag.String("wording", "", "顯示用語：original 或 plain；空白沿用 prefs.json")
	themeFlag := flag.String("theme", "", "圖形主題：retro 或 modern；空白沿用 prefs.json")
	resolutionFlag := flag.String("resolution", "", "顯示解析度：original 或 high；空白沿用 prefs.json")
	audioFlag := flag.String("audio", string(uiaudio.ModeOff), "音訊：off、retro 或 modern（retro 需玩家自備 MUS/TIM；modern 優先讀 manifest/Ogg，缺少時使用原創純 Go fallback）")
	modernAudioDir := flag.String("modern-audio", "assets/music/modern", "Modern Ogg 音樂目錄（有通過 manifest 的 Ogg 時優先使用；缺檔安全退回原創純 Go loop）")
	prefsFile := flag.String("prefs", "", "偏好檔路徑；空白使用 XDG_CONFIG_HOME/dsds/prefs.json")
	flag.Parse()
	prefsPath := *prefsFile
	if prefsPath == "" {
		var err error
		prefsPath, err = userprefs.DefaultPath()
		if err != nil {
			fmt.Fprintln(os.Stderr, "錯誤:", err)
			os.Exit(1)
		}
	}
	preferences, prefsErr := userprefs.Load(prefsPath)
	if prefsErr != nil {
		fmt.Fprintf(os.Stderr, "偏好檔無效，使用內建預設：%v\n", prefsErr)
	}
	wordingValue := preferences.Wording
	if *wording != "" {
		wordingValue = *wording
	}
	if wordingValue == "" {
		wordingValue = string(i18n.WordingOriginal)
	}
	wordingMode, err := i18n.ParseWordingMode(wordingValue)
	if err != nil {
		fmt.Fprintln(os.Stderr, "錯誤:", err)
		os.Exit(1)
	}
	themeValue := preferences.Theme
	if *themeFlag != "" {
		themeValue = *themeFlag
	}
	themeMode, themeErr := uitheme.ParseMode(themeValue)
	if themeErr != nil {
		if *themeFlag != "" {
			fmt.Fprintln(os.Stderr, "錯誤:", themeErr)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "偏好檔的主題無效，使用 retro：%v\n", themeErr)
		themeMode = uitheme.ModeRetro
	}
	resolutionValue := preferences.Resolution
	if *resolutionFlag != "" {
		resolutionValue = *resolutionFlag
	}
	if resolutionValue == "" {
		resolutionValue = string(uiresolution.ModeOriginal)
	}
	resolutionMode, resolutionErr := uiresolution.Parse(resolutionValue)
	if resolutionErr != nil {
		if *resolutionFlag != "" {
			fmt.Fprintln(os.Stderr, "錯誤:", resolutionErr)
			os.Exit(1)
		}
		fmt.Fprintf(os.Stderr, "偏好檔的解析度無效，使用 original：%v\n", resolutionErr)
		resolutionMode = uiresolution.ModeOriginal
	}
	audioMode, audioErr := uiaudio.ParseMode(*audioFlag)
	if audioErr != nil {
		fmt.Fprintln(os.Stderr, "錯誤:", audioErr)
		os.Exit(1)
	}

	if err := run(*gameDir, game.ProvinceID(*start), *savePath, uint32(*seed),
		*localeDir, *etenDir, wordingMode, themeMode, resolutionMode, audioMode, *modernAudioDir, prefsPath, preferences); err != nil {
		fmt.Fprintln(os.Stderr, "錯誤:", err)
		os.Exit(1)
	}
}

func run(dir string, start game.ProvinceID, savePath string, seed uint32,
	localeDir, etenDir string, wordingMode i18n.WordingMode,
	themeMode uitheme.Mode, resolutionMode uiresolution.Mode, audioMode uiaudio.Mode,
	modernAudioDir string, prefsPath string, preferences userprefs.Preferences) error {
	if !start.Valid() {
		return fmt.Errorf("省編號 %d 超出 1..%d", start, game.ProvinceCount)
	}
	read := func(name string) ([]byte, error) {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("讀不到 %s: %w", name, err)
		}
		return b, nil
	}

	warpos, err := read("WARPOS.DAT")
	if err != nil {
		return err
	}
	tername, err := read("TERNAME.DAT")
	if err != nil {
		return err
	}
	nwmap, err := read("NWMAP.DAT")
	if err != nil {
		return err
	}
	m, err := game.LoadMap(warpos, tername, nwmap)
	if err != nil {
		return err
	}

	w1, err := read("1.15")
	if err != nil {
		return err
	}
	w2, err := read("2.15")
	if err != nil {
		return err
	}
	w3, err := read("3.15")
	if err != nil {
		return err
	}
	gnames, err := read("MAN115")
	if err != nil {
		return err
	}
	fonts, err := render.LoadPanelFonts(w1, w2, w3, gnames)
	if err != nil {
		return err
	}

	w4, err := read("4.15")
	if err != nil {
		return err
	}
	w4f, err := assets.ParseGlyphFile(w4)
	if err != nil {
		return err
	}
	cmdFonts := render.CommandFonts{W2: fonts.W2, W4: w4f}
	fanData, err := read("FAN(1).15")
	if err != nil {
		return err
	}
	fan, err := assets.ParseGlyphFile(fanData)
	if err != nil {
		return fmt.Errorf("解析 FAN(1).15: %w", err)
	}

	// 省份狀態：優先讀存檔，沒有就用第一期的初始檔。
	tbl, origSave, err := loadProvinces(read)
	if err != nil {
		return err
	}
	// 戰鬥副本與省份存檔分開載入。來源一律唯讀取自 gameDir；輸出路徑
	// 則落在 -save 同目錄，避免把原版 SAVE(1).DT2／MEM_WAR.DAT 改掉。
	var battleDT2 []byte
	var battleMemWar []byte
	var battleDT2States [game.ProvinceCount]game.BattleState
	var battleMemStates [game.ProvinceCount]game.BattleState
	battleDT2Path, battleMemWarPath := "", ""
	if savePath != "" {
		battleDT2Path = filepath.Join(filepath.Dir(savePath), battleSourceName(savePath))
		battleMemWarPath = filepath.Join(filepath.Dir(savePath), "MEM_WAR.DAT")
	}
	if data, states, _, loadErr := loadBattleRecord(read, battleDT2Path,
		battleSourceName(savePath), "SAVE(1).DT2"); loadErr != nil {
		fmt.Fprintf(os.Stderr, "DT2 載入略過（戰鬥結算將不寫回）：%v\n", loadErr)
	} else {
		battleDT2, battleDT2States = data, states
	}
	if data, states, _, loadErr := loadBattleRecord(read, battleMemWarPath,
		"MEM_WAR.DAT", ""); loadErr != nil {
		fmt.Fprintf(os.Stderr, "MEM_WAR.DAT 載入略過（戰鬥結算將不寫回）：%v\n", loadErr)
	} else {
		battleMemWar, battleMemStates = data, states
	}

	mandat, err := read("MAN(1).DAT")
	if err != nil {
		return err
	}
	// ⛔ 筆數**不能以名表為準**。名表（`MAN215` 106 個名字）比程式實際
	// 會掃到的筆數少——第二、三期是 191（`word_6BC4A`），
	// 第 107..191 筆是**沒有姓名的番號部隊**（勢力名 `+28` = 0）。
	// 照名表解會靜默丟掉 85 支部隊。
	sc, err := game.ScenarioByStage(1)
	if err != nil {
		return err
	}
	var generals []game.General
	var startupSession *gameSession
	if origSave != nil {
		// 存檔必須在純記憶體快照中一次解完；不得像舊路徑那樣
		// 勢力表或停火表壞掉時只印警告、卻繼續帶著半套狀態運行。
		startupSession, err = buildSession(origSave, sc, start, 0)
		if err == nil {
			tbl, generals, start = startupSession.tbl, startupSession.generals, startupSession.current
		}
	} else {
		generals, err = game.ParseGenerals(mandat, sc.Generals)
	}
	if err != nil {
		return err
	}

	newterr, err := read("NEWTERR.TPC")
	if err != nil {
		return err
	}
	rail, err := read("RAIL.TPC")
	if err != nil {
		return err
	}
	ts, err := render.LoadTileSet(newterr, rail, assets.EGADefaultPalette)
	if err != nil {
		return err
	}
	newicon, err := read("NEWICON.TPC")
	if err != nil {
		return err
	}
	icons, err := render.LoadIcons(newicon)
	if err != nil {
		return err
	}
	var narrativeImages []*assets.Image
	if newsData, newsErr := read("NEWSDATA.DAT"); newsErr != nil {
		fmt.Fprintf(os.Stderr, "NEWSDATA.DAT 未載入（史事圖庫停用）：%v\n", newsErr)
	} else if decoded, decodeErr := assets.DecodeBGISet(newsData); decodeErr != nil || len(decoded) != i18n.NarrativeCount {
		if decodeErr != nil {
			fmt.Fprintf(os.Stderr, "NEWSDATA.DAT 解碼失敗（史事圖庫停用）：%v\n", decodeErr)
		} else {
			fmt.Fprintf(os.Stderr, "NEWSDATA.DAT 圖像數 %d，不是 %d（史事圖庫停用）\n", len(decoded), i18n.NarrativeCount)
		}
	} else {
		valid := true
		for i, image := range decoded {
			if image == nil || image.W != 215 || image.H != 16 {
				fmt.Fprintf(os.Stderr, "NEWSDATA.DAT 圖像 #%d 尺寸不是 215x16（史事圖庫停用）\n", i)
				valid = false
				break
			}
		}
		if valid {
			narrativeImages = decoded
		}
	}
	retroTheme := render.NewRetroThemeWithIcons(ts, icons, assets.EGADefaultPalette)
	modernTheme := uitheme.NewModern()
	if ornaments, ornamentErr := uitheme.DecodeOriginalOrnaments(read); ornamentErr != nil {
		fmt.Fprintf(os.Stderr, "原版介面裝飾載入失敗（Modern A2 改用程式化邊框）：%v\n", ornamentErr)
	} else {
		modernTheme.SetOriginalOrnaments(ornaments)
	}
	var battlefieldTheme uitheme.Theme
	var unitTheme uitheme.UnitProvider
	switch themeMode {
	case uitheme.ModeRetro:
		battlefieldTheme = retroTheme
	case uitheme.ModeModern:
		battlefieldTheme = modernTheme
	default:
		return fmt.Errorf("未知主題 %q", themeMode)
	}
	var ok bool
	if unitTheme, ok = battlefieldTheme.(uitheme.UnitProvider); !ok || unitTheme == nil {
		return fmt.Errorf("主題 %s 的部隊圖示尚未載入", themeMode)
	}
	var hudIcons uitheme.HUDIconProvider
	if themeMode == uitheme.ModeModern {
		if hudIcons, ok = battlefieldTheme.(uitheme.HUDIconProvider); !ok || hudIcons == nil {
			return fmt.Errorf("主題 %s 的 HUD 圖示尚未載入", themeMode)
		}
	}

	windowW, windowH := render.ModeBGIW*scale, render.ModeBGIH*scale
	if resolutionMode == uiresolution.ModeHigh {
		windowW, windowH = uiresolution.HighWidth, uiresolution.HighHeight
	}
	ebiten.SetWindowSize(windowW, windowH)
	ebiten.SetWindowTitle("大時代的故事")
	a := &app{
		m: m, tbl: tbl, generals: generals, fonts: fonts, cmdFonts: cmdFonts, fan: fan,
		tiles: ts, icons: icons, origSave: origSave, savePath: savePath,
		battleDT2: battleDT2, battleMemWar: battleMemWar,
		battleDT2States: battleDT2States, battleMemStates: battleMemStates,
		battleDT2Path: battleDT2Path, battleMemWarPath: battleMemWarPath,
		battlefieldTheme: battlefieldTheme, unitTheme: unitTheme, hudIcons: hudIcons,
		retroTheme: retroTheme, modernTheme: modernTheme,
		narrativeImages: narrativeImages,
		themeMode:       themeMode,
		resolution:      resolutionMode,
		current:         start, dirty: true, provinceLimit: sc.Provinces,
		stage: int(sc.Stage), wordingMode: wordingMode,
		prefsPath: prefsPath, preferences: preferences,
		// 固定種子：`AGENTS.md` §8 要求截圖驗收可重現。
		rng:      game.NewRand(seed),
		messages: newMessageQueue(preferences.MessageTime),
	}
	if audioMode == uiaudio.ModeRetro {
		// SCENE 是策略畫面的初始曲目；其他曲目存在就加入延遲切換庫，
		// 缺少時沿用目前曲目並列印一次診斷，不把玩家自備素材誤當必要檔。
		trackStems := []struct {
			track uiaudio.Track
			stem  string
		}{
			{uiaudio.TrackScene, "SCENE"},
			{uiaudio.TrackStrategy, "STRATEGY"},
			{uiaudio.TrackMainTheme, "MAINTHEM"},
			{uiaudio.TrackBattle1, "BATTLE1"},
			{uiaudio.TrackBattle2, "BATTLE2"},
			{uiaudio.TrackBattleAlt, "BT02"},
			{uiaudio.TrackWall, "WALL"},
			{uiaudio.TrackFinal, "FINAL"},
		}
		sources := make(map[uiaudio.Track]uiaudio.TrackSource)
		for _, item := range trackStems {
			musData, musErr := read(item.stem + ".MUS")
			timData, timErr := read(item.stem + ".TIM")
			if musErr != nil || timErr != nil {
				if item.track == uiaudio.TrackScene {
					if musErr != nil {
						return fmt.Errorf("retro 音訊需要 SCENE.MUS：%w", musErr)
					}
					return fmt.Errorf("retro 音訊需要 SCENE.TIM：%w", timErr)
				}
				fmt.Fprintf(os.Stderr, "音訊曲目 %s 缺少 MUS/TIM，略過\n", item.stem)
				continue
			}
			sources[item.track] = uiaudio.TrackSource{MUS: musData, TIM: timData}
		}
		manager, err := uiaudio.NewRetroTracks(sources, uiaudio.TrackScene, 0)
		if err != nil {
			return err
		}
		a.audio = manager
		a.audioMode = audioMode
		a.audioTrack = uiaudio.TrackScene
		a.audioWarned = make(map[uiaudio.Track]bool)
		defer manager.Close()
		if err := manager.Start(0.7); err != nil {
			return err
		}
	}
	if audioMode == uiaudio.ModeModern {
		set, loadErr := uiaudio.LoadModernTracks(modernAudioDir)
		if loadErr != nil {
			// Modern 音樂是新增外殼；缺少尚未清權的 Ogg 不應讓玩家無聲。
			// 這裡改用原創純 Go loop，並保留 manifest 失敗原因，避免把
			// technical reference 誤當成可發行音檔。
			fmt.Fprintf(os.Stderr, "modern manifest 未載入（%v），改用原創純 Go 音樂 fallback\n", loadErr)
			manager, managerErr := uiaudio.NewModernProceduralTracks(0)
			if managerErr != nil {
				return fmt.Errorf("建立 modern 純 Go 音訊：%w", managerErr)
			}
			a.audio = manager
			a.audioMode = audioMode
			a.audioTrack = uiaudio.TrackScene
			a.audioWarned = make(map[uiaudio.Track]bool)
			defer manager.Close()
			if startErr := manager.Start(0.7); startErr != nil {
				return fmt.Errorf("啟動 modern 純 Go 音訊：%w", startErr)
			}
		} else {
			for _, warning := range set.Warnings {
				fmt.Fprintln(os.Stderr, "modern 音訊：", warning)
			}
			manager, managerErr := uiaudio.NewModernTracks(set.Sources, uiaudio.TrackScene)
			if managerErr != nil {
				fmt.Fprintf(os.Stderr, "modern Ogg 不可播放（%v），改用原創純 Go 音樂 fallback\n", managerErr)
				manager, fallbackErr := uiaudio.NewModernProceduralTracks(0)
				if fallbackErr != nil {
					return fmt.Errorf("建立 modern 純 Go 音訊：%w", fallbackErr)
				}
				a.audio = manager
				a.audioMode = audioMode
				a.audioTrack = uiaudio.TrackScene
				a.audioWarned = make(map[uiaudio.Track]bool)
				defer manager.Close()
				if startErr := manager.Start(0.7); startErr != nil {
					return fmt.Errorf("啟動 modern 純 Go 音訊：%w", startErr)
				}
			} else {
				a.audio = manager
				a.audioMode = audioMode
				a.audioTrack = uiaudio.TrackScene
				a.audioWarned = make(map[uiaudio.Track]bool)
				defer manager.Close()
				if startErr := manager.Start(0.7); startErr != nil {
					fmt.Fprintf(os.Stderr, "modern Ogg 啟動失敗（%v），改用原創純 Go 音樂 fallback\n", startErr)
					fallback, fallbackErr := uiaudio.NewModernProceduralTracks(0)
					if fallbackErr != nil {
						return fmt.Errorf("建立 modern 純 Go 音訊：%w", fallbackErr)
					}
					_ = manager.Close()
					a.audio = fallback
					a.audioMode = audioMode
					a.audioTrack = uiaudio.TrackScene
					defer fallback.Close()
					if fallbackErr = fallback.Start(0.7); fallbackErr != nil {
						return fmt.Errorf("啟動 modern 純 Go 音訊：%w", fallbackErr)
					}
				}
			}
		}
	}
	if startupSession != nil {
		a.factions, a.factionLeaders, a.factionOf, a.warRecords = startupSession.factions, startupSession.leaders, startupSession.factionOf, startupSession.warRecords
		a.world, a.cmdBudget = startupSession.world, startupSession.cmdBudget
		a.ledger = startupSession.ledger
		a.origSave = startupSession.origSave
		a.playerCommander = startupSession.playerCommander
	} else {
		// TOWN(N).DAT 是新局初值，沒有 .DT1 後半的勢力與停火區塊。
		a.world = buildWorld(tbl, generals, nil, sc.Stage)
		a.cmdBudget = game.NewCommandBudget(a.world)
		a.ledger = game.NewDiplomacyLedger()
		if prov, err := tbl.At(start); err == nil {
			a.playerCommander = prov.Commander
		}
	}
	// 語系表載不到不是致命錯誤——省名會退回「省 N」，其餘照跑。
	// **不要靜默**：印到 stderr，否則「沒有語系表」與「語系表是壞的」
	// 在畫面上長得一樣。
	if loc, err := i18n.Load(localeDir); err != nil {
		fmt.Fprintf(os.Stderr, "語系表載入失敗（省名會顯示成編號）：%v\n", err)
	} else {
		a.loc = loc
	}
	if narrative, err := i18n.LoadNarrative(localeDir); err != nil {
		fmt.Fprintf(os.Stderr, "敘事語系載入失敗（史事圖庫停用）：%v\n", err)
	} else {
		a.narrative = narrative
	}
	if wording, err := i18n.LoadWording(localeDir); err != nil {
		fmt.Fprintf(os.Stderr, "顯示用語載入失敗（自傳入口停用）：%v\n", err)
	} else {
		a.wording = wording
	}
	sharedDir := filepath.Join(filepath.Dir(localeDir), "shared")
	if people, err := i18n.LoadPeople(localeDir, sharedDir); err != nil {
		fmt.Fprintf(os.Stderr, "人物語系資料載入失敗（自傳入口停用）：%v\n", err)
	} else {
		a.people = people
		if portraits, portraitErr := i18n.LoadPortraitCatalog(sharedDir, people); portraitErr != nil {
			fmt.Fprintf(os.Stderr, "人物肖像登錄載入失敗（Modern 改用檔案卡）：%v\n", portraitErr)
		} else {
			a.portraits = portraits
		}
	}
	if fonts, err := assets.LoadEtenFonts(etenDir); err != nil {
		fmt.Fprintf(os.Stderr, "倚天完整字庫載入失敗（自傳入口停用）：%v\n", err)
	} else {
		a.eten = fonts
	}
	if a.semanticWording() && (a.wording == nil || a.eten == nil) {
		return fmt.Errorf("英日語系／現代白話模式需要完整 wording.json 與倚天三套字庫")
	}
	if themeMode == uitheme.ModeModern && (a.wording == nil || a.eten == nil) {
		return fmt.Errorf("Modern 主題需要完整 wording.json 與倚天三套字庫；請以 -eten 提供玩家自備字庫")
	}
	if startupSession != nil {
		a.year, a.month = startupSession.year, startupSession.month
	} else if tbl.Date != nil {
		a.year, a.month = tbl.Date.Year, tbl.Date.Month
	}
	return ebiten.RunGame(a)
}

// loadProvinces 讀省份狀態：優先用存檔 SAVE(1).DT1，讀不到就退回
// 第一期的初始檔 TOWN(1).DAT。
//
// 兩個檔案是同一個結構，只差 4 bytes 的相位（docs/spec/03 §1）。
// 第二個回傳值是原始存檔的 bytes，寫回時當基底；用初始檔開的話是 nil。
func loadProvinces(read func(string) ([]byte, error)) (*game.ProvinceTable, []byte, error) {
	if b, err := read("SAVE(1).DT1"); err == nil {
		t, err := game.ParseSaveProvinces(b)
		return t, b, err
	}
	b, err := read("TOWN(1).DAT")
	if err != nil {
		return nil, nil, err
	}
	t, err := game.ParseTownFile(b)
	return t, nil, err
}

// battleSourceName 依 -save 的 .DT1 名稱推導對應的 .DT2 名稱。
// 自訂副檔名不偷偷改寫，回退到原版預設 SAVE(1).DT2，讓測試與新局仍能啟動。
func battleSourceName(savePath string) string {
	base := filepath.Base(savePath)
	ext := filepath.Ext(base)
	if strings.EqualFold(ext, ".DT1") {
		return base[:len(base)-len(ext)] + ".DT2"
	}
	return "SAVE(1).DT2"
}

// loadBattleRecord 解析一份 39×469 bytes 的戰鬥副本。若 outputPath 已有上次
// 執行產生的玩家副本，先以它作 Raw 基底；否則 preferred 找不到時再用 fallback。
// outputPath／fallback 可為空，表示不走該層回退。
func loadBattleRecord(read func(string) ([]byte, error), outputPath, preferred, fallback string) (
	[]byte, [game.ProvinceCount]game.BattleState, string, error) {
	var zero [game.ProvinceCount]game.BattleState
	parse := func(data []byte, source string) ([]byte, [game.ProvinceCount]game.BattleState, string, error) {
		states, err := game.ParseBattleStates(data)
		if err != nil {
			return nil, zero, source, fmt.Errorf("解析 %s：%w", source, err)
		}
		return append([]byte(nil), data...), states, source, nil
	}
	if outputPath != "" {
		data, err := os.ReadFile(outputPath)
		if err == nil {
			return parse(data, outputPath)
		}
		if !os.IsNotExist(err) {
			return nil, zero, outputPath, fmt.Errorf("讀取既有戰鬥副本 %s：%w", outputPath, err)
		}
	}
	if preferred == "" {
		return nil, zero, "", fmt.Errorf("戰鬥副本名稱為空")
	}
	data, err := read(preferred)
	source := preferred
	if err != nil && fallback != "" && fallback != preferred {
		data, err = read(fallback)
		source = fallback
	}
	if err != nil {
		return nil, zero, "", err
	}
	return parse(data, source)
}
