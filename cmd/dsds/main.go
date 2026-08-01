// dsds 是《大時代的故事》remake 的執行檔。
//
// 目前做到的：載入原版資料，用**原版的地形圖塊**畫出各省的 14×14 戰場，
// 側欄顯示省份狀態，可切換省份、打仗、下已解出的政略指令、存檔。
//
// **只接規則層已 confirmed 的指令**（徵稅／開發三項／慰勞／攻打）。
// 運補、商業、外交、秘密行動的公式還沒讀出來，按了沒反應——
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
//	4        徵稅（每月限一次）
//	7        開發 → 1 墾地　2 建兵工廠　3 挖金礦
//	C        慰勞軍民（原版是指令 14）
//	A        對第一個可攻打的鄰省開戰
//	E        結束回合，推進一個月；跨年跑年度結算
//
// 指令結果印在 stderr，面板的數值會即時更新。
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
	"github.com/wicanr2/great-era-remake/internal/ui/render"
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
	screenMap     screen = iota // 戰場 + 省份面板
	screenCommand               // 政略指令選單
	screenBattle                // 戰鬥（見 battle.go）
	screenDevelop               // 開發的三個子項（見 strategy.go）
	screenQuit                  // 離開確認
)

type app struct {
	m        *game.Map
	tbl      *game.ProvinceTable // 39 省的狀態（存檔或初始檔）
	generals []game.General      // 該期的將領表
	fonts    render.PanelFonts   // 面板用的三個字模檔
	tiles    *render.TileSet     // NEWTERR + RAIL 的圖塊
	origSave []byte              // 原始存檔內容，寫回時當基底
	cmdFonts render.CommandFonts
	icons    []*assets.Image // NEWICON.TPC 的兵種圖示
	battle   *battleState    // 非 nil 表示正在打仗
	world    *game.AIWorld   // 規則層：政略指令都經過它
	loc      *i18n.Locale    // 語系表：省名與 UI 詞彙（nil 表示沒載到）
	cmdBudget *game.CommandBudget // 每省這個月剩餘的指令數（docs/re/13 §2）
	rng      *game.Rand      // 原版的 LCG（docs/re/17），固定種子才可重現
	year     uint16
	month    uint8
	msg      string // 上一個指令的結果，印到 stderr
	current  game.ProvinceID
	screen   screen
	savePath string // 離開時自動存檔的目標（不覆蓋原版）
	saveErr  error  // 存檔失敗就不離開
	dirty    bool
	frame    *ebiten.Image
}

func (a *app) Update() error {
	// F10 是唯一的離開鍵，而且要先確認（CLAUDE.md §9）。
	if inpututil.IsKeyJustPressed(ebiten.KeyF10) && a.screen != screenQuit {
		a.screen, a.dirty = screenQuit, true
		return nil
	}

	switch a.screen {
	case screenQuit:
		switch {
		case inpututil.IsKeyJustPressed(ebiten.KeyY):
			// 存檔失敗就不離開。
			if err := a.autosave(); err != nil {
				a.saveErr, a.screen, a.dirty = err, screenMap, true
				fmt.Fprintln(os.Stderr, "自動存檔失敗，不離開:", err)
				return nil
			}
			return ebiten.Termination
		case inpututil.IsKeyJustPressed(ebiten.KeyN),
			inpututil.IsKeyJustPressed(ebiten.KeyEscape):
			a.screen, a.dirty = screenMap, true
		}
		return nil

	case screenBattle:
		return a.updateBattle()

	case screenCommand:
		// ESC 只退回上一層。
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			a.screen, a.dirty = screenMap, true
			return nil
		}
		// A：對第一個可攻打的鄰省開戰。攻打候選的規則見
		// `docs/spec/01` §2（鄰省 − 自己控制的省）。
		if inpututil.IsKeyJustPressed(ebiten.KeyA) {
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
		case inpututil.IsKeyJustPressed(ebiten.KeyDigit4):
			a.report(a.withBudget(a.current, a.execTax))
		case inpututil.IsKeyJustPressed(ebiten.KeyDigit7):
			a.screen, a.dirty = screenDevelop, true
		case inpututil.IsKeyJustPressed(ebiten.KeyC):
			a.report(a.withBudget(a.current, a.execComfort))
		case inpututil.IsKeyJustPressed(ebiten.KeyE):
			a.report(a.endTurn())
		}
		return nil

	case screenDevelop:
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			a.screen, a.dirty = screenCommand, true
			return nil
		}
		for i, k := range []ebiten.Key{ebiten.KeyDigit1, ebiten.KeyDigit2, ebiten.KeyDigit3} {
			if inpututil.IsKeyJustPressed(k) {
				sub := i + 1
				a.report(a.withBudget(a.current,
					func() string { return a.execDevelop(sub) }))
				a.screen = screenCommand
			}
		}
		return nil
	}

	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyEnter),
		inpututil.IsKeyJustPressed(ebiten.KeyKPEnter):
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

// autosave 把當前的省份狀態寫回一份**副本**。
//
// CLAUDE.md §9：原版資產唯讀，測試存檔一律寫到明確的輸出目錄，
// 不覆蓋原版的 SAVE(1).DT1。寫回是「改寫」不是「重建」——
// 未解區域一個 byte 都不動（internal/game/save.go）。
func (a *app) autosave() error {
	if a.savePath == "" || a.origSave == nil {
		return nil // 沒有存檔來源（用初始檔開的），不寫
	}
	out, err := game.WriteProvinces(a.origSave, a.tbl)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(a.savePath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(a.savePath, out, 0o644)
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
	}
	data.Year, data.Month = a.year, a.month
	if err := c.DrawStrategyPanel(data, a.fonts); err != nil {
		return err
	}

	switch a.screen {
	case screenCommand:
		if err := c.DrawCommandPage(a.cmdFonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY); err != nil {
			return err
		}
	case screenDevelop:
		if err := c.DrawCommandPage(a.cmdFonts, panelInk, panelPaper,
			fieldX, fieldY, render.ModeBGIW-fieldX, render.ModeBGIH-fieldY); err != nil {
			return err
		}
	case screenQuit:
		// 離開確認。用原版詞表的「您確定嗎」（4.15 詞條 0）。
		c.DrawConfirmBox(a.cmdFonts.W4, panelInk, panelPaper,
			fieldX+60, fieldY+120)
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
	flag.Parse()

	if err := run(*gameDir, game.ProvinceID(*start), *savePath, uint32(*seed),
		*localeDir); err != nil {
		fmt.Fprintln(os.Stderr, "錯誤:", err)
		os.Exit(1)
	}
}

func run(dir string, start game.ProvinceID, savePath string, seed uint32,
	localeDir string) error {
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
	fonts, err := render.LoadPanelFonts(w2, w3, gnames)
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

	// 省份狀態：優先讀存檔，沒有就用第一期的初始檔。
	tbl, origSave, err := loadProvinces(read)
	if err != nil {
		return err
	}

	mandat, err := read("MAN(1).DAT")
	if err != nil {
		return err
	}
	// 筆數以名表為準——MAN(N).DAT 都是 274 筆的空間，但二三期只有 106 位。
	generals, err := game.ParseGenerals(mandat,
		len(fonts.Gen.Glyphs)/game.GeneralNameSlotWidth)
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
		m: m, tbl: tbl, generals: generals, fonts: fonts, cmdFonts: cmdFonts,
		tiles: ts, icons: icons, origSave: origSave, savePath: savePath,
		current: start, dirty: true,
		// 固定種子：`CLAUDE.md` §9 要求截圖驗收可重現。
		rng: game.NewRand(seed),
	}
	a.world = buildWorld(tbl, generals)
	a.cmdBudget = game.NewCommandBudget(a.world)
	// 語系表載不到不是致命錯誤——省名會退回「省 N」，其餘照跑。
	// **不要靜默**：印到 stderr，否則「沒有語系表」與「語系表是壞的」
	// 在畫面上長得一樣。
	if loc, err := i18n.Load(localeDir); err != nil {
		fmt.Fprintf(os.Stderr, "語系表載入失敗（省名會顯示成編號）：%v\n", err)
	} else {
		a.loc = loc
	}
	if tbl.Date != nil {
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
