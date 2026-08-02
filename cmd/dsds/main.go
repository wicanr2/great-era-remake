// dsds 是《大時代的故事》remake 的執行檔。
//
// 目前做到的：載入原版資料，用**原版的地形圖塊**畫出各省的 14×14 戰場，
// 側欄顯示省份狀態，可切換省份、打仗、下已解出的政略指令、存檔。
//
// **只接規則層已 confirmed 的指令**（運補／徵稅／開發／秘密行動的學潮／
// 商業／慰勞／攻打）。政策、外交、停火尚未接上完整操作流程，按了沒反應——
// 那勝過假裝有效果（`cmd/dsds/strategy.go` 開頭）。
//
//	tools/go.sh run ./cmd/dsds -game workplace/orig/game
//
// 操作（CLAUDE.md §9：**ESC 只取消／退回上一層，F10 才離開**）：
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

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	userprefs "github.com/wicanr2/great-era-remake/internal/prefs"
	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
	"github.com/wicanr2/great-era-remake/internal/ui/textlayout"
)

// scale 是視窗放大倍率。
//
// remake 差異（CLAUDE.md §1「外殼允許現代化」）：只放大視窗，
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
	screenPolicy                           // 指令 8：授權自治／產能分配
	screenAutonomy                         // 政策：切換其他省份自治狀態
	screenProduction                       // 政策：調整本省產能分配
	screenOtherOptions                     // 指令 15：原版八項 + remake 顯示設定
	screenSaveConfirm                      // 其他選項 1：儲存確認
	screenLoadConfirm                      // 其他選項 2：載入確認
	screenMessageTime                      // 其他選項 6：訊息停留時間 1..10
	screenDisplayOptions                   // remake 新增：顯示設定（由指令 15 區域進入）
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
	fan              *assets.GlyphFile // FAN(1).15，部隊番號字模
	icons            []*assets.Image   // NEWICON.TPC 的兵種圖示
	battle           *battleState      // 非 nil 表示正在打仗
	world            *game.AIWorld     // 規則層：政略指令都經過它
	loc              *i18n.Locale      // 語系表：省名與 UI 詞彙（nil 表示沒載到）
	people           *i18n.PeopleDB    // 人物自傳：(期別, 將領槽位) → 語系人物資料
	eten             *assets.EtenFonts // 完整繁中字庫；載不到時停用自傳入口
	wording          *i18n.WordingCatalog
	wordingMode      i18n.WordingMode
	preferences      userprefs.Preferences
	prefsPath        string
	stage            int
	cmdBudget        *game.CommandBudget // 每省這個月剩餘的指令數（docs/re/13 §2）
	rng              *game.Rand          // 原版的 LCG（docs/re/17），固定種子才可重現
	year             uint16
	month            uint8
	msg              string // 上一個指令的結果，印到 stderr
	messages         *messageQueue
	messageTimeInput uint32
	pointer          pointerTracker
	pointerAction    actions.Action
	transferTargets  []game.ProvinceID
	transferMode     game.PlayerTransferMode
	transferSession  *game.PlayerTransferSelection
	transferInput    uint32
	transferCursor   int
	transferAmounts  [4]int
	transferGood     int
	tradeImport      bool
	tradeGood        game.TradeGood
	tradeAmount      uint32
	supplyTargets    []game.ProvinceID
	supplyTarget     game.ProvinceID
	supplyAmounts    [4]int
	supplyGood       int
	supplyInput      uint32
	recruitBranch    uint8
	recruitLimit     int
	recruitAmount    uint32
	reorganization   *game.Reorganization
	reorganizeID     game.GeneralID
	reorganizeInput  uint32
	viewGenerals     []game.GeneralID
	viewIndex        int
	viewGeneralBack  screen
	bioBack          screen
	bioPage          int
	bioPages         int
	viewProvince     game.ProvinceID
	viewInput        uint32
	viewPage         int
	covertAction     int
	covertInput      uint32
	autonomyTargets  []game.ProvinceID
	autonomyInput    uint32
	autonomySpent    bool
	productionItem   int
	productionInput  uint32
	productionSpent  bool
	provinceLimit    int
	current          game.ProvinceID
	playerCommander  game.GeneralID // 載入時用來重新找到玩家所屬的合法省份
	screen           screen
	quitBack         screen // 進入離開確認前的畫面；N／ESC 必須回到原處
	savePath         string // 離開時自動存檔的目標（不覆蓋原版）
	saveErr          error  // 存檔失敗就不離開
	dirty            bool
	frame            *ebiten.Image
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
	// F10 是唯一的離開鍵，而且要先確認（CLAUDE.md §9）。
	if inpututil.IsKeyJustPressed(ebiten.KeyF10) && a.screen != screenQuit {
		a.quitBack, a.screen, a.dirty = a.screen, screenQuit, true
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
				return nil
			}
			// 出兵消耗一個指令數（`sub_174C9`，`docs/re/13` §2）。
			if !a.cmdBudget.Spend(a.current) {
				a.report(fmt.Sprintf("%s 這個月的指令數用完了",
					a.provinceName(a.current)))
				return nil
			}
			if err := a.startBattle(target, a.current); err != nil {
				fmt.Fprintln(os.Stderr, "開戰失敗:", err)
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
		if inpututil.IsKeyJustPressed(ebiten.KeyB) {
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
		if inpututil.IsKeyJustPressed(ebiten.KeyB) {
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
		var mode i18n.WordingMode
		if a.actionPressed(actions.Select1, ebiten.KeyDigit1) {
			mode = i18n.WordingOriginal
		} else if a.actionPressed(actions.Select2, ebiten.KeyDigit2) {
			mode = i18n.WordingPlain
		} else {
			return nil
		}
		if err := a.setWordingPreference(mode); err != nil {
			a.report("顯示用語未變更：" + err.Error())
		}
		return nil
	}

	switch {
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

func (a *app) openBiography(back screen) {
	p, ok := a.currentBiography()
	if !ok {
		a.report("這位將領目前沒有可顯示的人物資料")
		return
	}
	body := p.Biography
	if body == "" {
		var found bool
		body, found = a.wording.Text("biography.unavailable", a.wordingMode)
		if !found {
			a.report("人物自傳用語資料不完整")
			return
		}
	}
	doc, err := textlayout.Layout(body, textlayout.DefaultBiographyOptions)
	if err != nil {
		a.report(err.Error())
		return
	}
	a.bioBack, a.bioPage, a.bioPages = back, 0, len(doc.Pages)
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

// autosave 把當前的省份與將領狀態寫回一份**副本**。
//
// CLAUDE.md §9：原版資產唯讀，測試存檔一律寫到明確的輸出目錄，
// 不覆蓋原版的 SAVE(1).DT1。寫回是「改寫」不是「重建」——
// 未解區域一個 byte 都不動（internal/game/save.go）。
func (a *app) autosave() error {
	if a.savePath == "" || a.origSave == nil {
		return nil // 沒有存檔來源（用初始檔開的），不寫
	}
	out, err := game.WriteSave(a.origSave, a.tbl, a.generals)
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
	c := render.NewBGICanvas()

	if a.screen == screenBattle {
		if err := a.drawBattle(c); err != nil {
			return err
		}
	} else {
		bf, err := a.m.Battlefield(a.current)
		if err != nil {
			return err
		}
		// 用原版的 NEWTERR 圖塊畫戰場，有鐵路的格子疊 RAIL.TPC。
		if err := c.DrawTiledBattlefield(bf, a.tiles, fieldX, fieldY); err != nil {
			return err
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
	}
	data.Year, data.Month = a.year, a.month
	if err := c.DrawStrategyPanel(data, a.fonts); err != nil {
		return err
	}

	switch a.screen {
	case screenCommand:
		if a.wordingMode == i18n.WordingPlain {
			labels := make([]string, 15)
			for i := range labels {
				key := fmt.Sprintf("command.%02d", i+1)
				var err error
				labels[i], err = a.wordingText(key)
				if err != nil {
					return err
				}
			}
			if err := checkWordingGlyphs("政略主選單", c.DrawSemanticCommandPage(a.eten, panelInk, panelPaper,
				fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY, labels)); err != nil {
				return err
			}
			break
		}
		if err := c.DrawCommandPage(a.cmdFonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY); err != nil {
			return err
		}
	case screenDevelop:
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		if a.wordingMode == i18n.WordingPlain {
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
		result, err := c.DrawBiography(a.eten, render.BiographyView{Person: p,
			Page: a.bioPage, Title: title, Unavailable: unavailable}, panelInk, panelPaper)
		if err != nil {
			return err
		}
		if len(result.Missing) != 0 {
			fmt.Fprintf(os.Stderr, "人物自傳缺字（%s）：%q\n", p.NameInGame, string(result.Missing))
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
			"settings.wording.original", "settings.wording.plain"}
		labels := make([]string, len(keys))
		for i, key := range keys {
			label, err := a.wordingText(key)
			if err != nil {
				return err
			}
			labels[i] = label
		}
		selected := 0
		if a.wordingMode == i18n.WordingPlain {
			selected = 1
		}
		missing := c.DrawDisplayOptions(a.eten, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY,
			labels[0], labels[1], labels[2], labels[3], selected)
		if err := checkWordingGlyphs("顯示設定", missing); err != nil {
			return err
		}
	case screenOtherOptions:
		keys := []string{"other.save", "other.load", "other.sound", "other.command_art",
			"other.music", "other.message_time", "other.spectate", "other.quit", "other.display"}
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
			false, false, false, true, false, true, true}
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
	if a.screen == screenMap {
		if err := c.DrawOpenCommandButton(a.cmdFonts, panelInk, panelPaper, render.ModeBGIW); err != nil {
			return err
		}
	}
	for i, action := range a.navigationActions() {
		if action == actions.Submit {
			c.DrawSubmitButton(panelInk, panelPaper, render.ModeBGIW, a.navigationY(), i)
		} else {
			c.DrawNavigationButton(panelInk, panelPaper, render.ModeBGIW, a.navigationY(), i,
				action == actions.NextPage)
		}
	}
	if a.numericKeypadVisible() {
		c.DrawNumericKeypad(panelInk, panelPaper)
	}
	if a.messages != nil && a.messages.Active() {
		missing := c.DrawMessageOverlay(a.eten, panelInk, panelPaper,
			fieldX+8, render.ModeBGIH-92, render.ModeBGIW-fieldX-16, 82, a.messages.Current())
		if err := checkWordingGlyphs("結果訊息", missing); err != nil {
			return err
		}
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
	dst.DrawImage(a.frame, nil)
}

func (a *app) Layout(_, _ int) (int, int) {
	return render.ModeBGIW, render.ModeBGIH
}

func main() {
	gameDir := flag.String("game", "workplace/orig/game", "原版素材目錄（唯讀）")
	start := flag.Int("province", 26, "起始省編號（1-39），預設 26 = 湖北省")
	// CLAUDE.md §9：原版資產唯讀，存檔一律寫到別的地方。
	savePath := flag.String("save", "workplace/saves/SAVE(1).DT1",
		"離開時自動存檔的路徑（**不會**覆蓋原版）")
	// 固定亂數種子是 CLAUDE.md §9 的硬規則：截圖驗收要能重現。
	seed := flag.Uint("seed", 1, "亂數種子（原版 LCG，docs/re/17）")
	localeDir := flag.String("locale", "translations/zh-Hant",
		"語系資料目錄。換一個目錄就換一種語言（CLAUDE.md §6）")
	etenDir := flag.String("eten", "workplace/eten",
		"使用者提供的倚天 STDFONT.15／SPCFONT.15／ASCFONT.15 目錄（不隨遊戲散布）")
	wording := flag.String("wording", "", "顯示用語：original 或 plain；空白沿用 prefs.json")
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

	if err := run(*gameDir, game.ProvinceID(*start), *savePath, uint32(*seed),
		*localeDir, *etenDir, wordingMode, prefsPath, preferences); err != nil {
		fmt.Fprintln(os.Stderr, "錯誤:", err)
		os.Exit(1)
	}
}

func run(dir string, start game.ProvinceID, savePath string, seed uint32,
	localeDir, etenDir string, wordingMode i18n.WordingMode,
	prefsPath string, preferences userprefs.Preferences) error {
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

	ebiten.SetWindowSize(render.ModeBGIW*scale, render.ModeBGIH*scale)
	ebiten.SetWindowTitle("大時代的故事")
	a := &app{
		m: m, tbl: tbl, generals: generals, fonts: fonts, cmdFonts: cmdFonts, fan: fan,
		tiles: ts, icons: icons, origSave: origSave, savePath: savePath,
		current: start, dirty: true, provinceLimit: sc.Provinces,
		stage: int(sc.Stage), wordingMode: wordingMode,
		prefsPath: prefsPath, preferences: preferences,
		// 固定種子：`CLAUDE.md` §9 要求截圖驗收可重現。
		rng:      game.NewRand(seed),
		messages: newMessageQueue(preferences.MessageTime),
	}
	if startupSession != nil {
		a.world, a.cmdBudget = startupSession.world, startupSession.cmdBudget
		a.origSave = startupSession.origSave
		a.playerCommander = startupSession.playerCommander
	} else {
		// TOWN(N).DAT 是新局初值，沒有 .DT1 後半的勢力與停火區塊。
		a.world = buildWorld(tbl, generals, nil, sc.Stage)
		a.cmdBudget = game.NewCommandBudget(a.world)
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
	}
	if fonts, err := assets.LoadEtenFonts(etenDir); err != nil {
		fmt.Fprintf(os.Stderr, "倚天完整字庫載入失敗（自傳入口停用）：%v\n", err)
	} else {
		a.eten = fonts
	}
	if wordingMode == i18n.WordingPlain && (a.wording == nil || a.eten == nil) {
		return fmt.Errorf("現代白話模式需要完整 wording.json 與倚天三套字庫")
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
