// screenshot 把政略畫面合成成 PNG，不需要顯示器。
//
//	tools/go.sh run ./cmd/screenshot -game workplace/orig/game -province 26 -out workplace/shots
//
// internal/ui/render 不依賴 Ebiten（CLAUDE.md §11），所以這支可以在
// 無頭環境跑，用來做視覺驗收與對照原版截圖。
//
// 輸出含原版美術，不要放進版控。
package main

import (
	"flag"
	"fmt"
	"image/png"
	"os"
	"path/filepath"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
	uiresolution "github.com/wicanr2/great-era-remake/internal/ui/resolution"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

func main() {
	dir := flag.String("game", "workplace/orig/game", "原版素材目錄（唯讀）")
	out := flag.String("out", "workplace/shots", "輸出目錄")
	prov := flag.Int("province", 0, "只畫某一省（1-39），0 = 全部")
	menu := flag.Bool("menu", false, "右側改畫政略指令選單")
	units := flag.Bool("units", false, "在戰場上疊出參戰單位的圖示")
	battle := flag.Bool("battle", false, "畫一場實際的戰鬥（跑 BattleSim）")
	biography := flag.Bool("biography", false, "畫 Modern 高解析人物自傳（首頁與末頁）")
	themeFlag := flag.String("theme", "retro", "圖形主題：retro 或 modern")
	resolutionFlag := flag.String("resolution", string(uiresolution.ModeOriginal), "截圖畫布：original 或 high（high 支援 Modern 地圖／政略／戰鬥／自傳頁預覽）")
	etenDir := flag.String("eten", "workplace/eten", "Modern／語系面板使用的玩家自備倚天字庫目錄")
	localeDir := flag.String("locale", "translations/zh-Hant", "Modern 面板的語系目錄")
	flag.Parse()

	themeMode, err := uitheme.ParseMode(*themeFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, "錯誤:", err)
		os.Exit(1)
	}
	resolutionMode, err := uiresolution.Parse(*resolutionFlag)
	if err != nil {
		fmt.Fprintln(os.Stderr, "錯誤:", err)
		os.Exit(1)
	}
	if err := run(*dir, *out, game.ProvinceID(*prov), *menu, *units, *battle, *biography, themeMode,
		resolutionMode, *etenDir, *localeDir); err != nil {
		fmt.Fprintln(os.Stderr, "錯誤:", err)
		os.Exit(1)
	}
}

func run(dir, out string, only game.ProvinceID, menu, units, battle, biography bool,
	themeMode uitheme.Mode, resolutionMode uiresolution.Mode, etenDir, localeDir string) error {
	if biography && (themeMode != uitheme.ModeModern || resolutionMode != uiresolution.ModeHigh) {
		return fmt.Errorf("人物自傳截圖只支援 Modern high")
	}
	if biography && (menu || units || battle) {
		return fmt.Errorf("人物自傳截圖不可與 menu、units 或 battle 同時使用")
	}
	if resolutionMode == uiresolution.ModeHigh && (themeMode != uitheme.ModeModern || units) {
		return fmt.Errorf("高解析截圖目前只支援 Modern 地圖／政略／戰鬥／自傳頁預覽（不可同時使用 units）")
	}
	read := func(name string) ([]byte, error) {
		return os.ReadFile(filepath.Join(dir, name))
	}
	must := func(name string) []byte {
		b, err := read(name)
		if err != nil {
			fmt.Fprintln(os.Stderr, "讀不到", name, err)
			os.Exit(1)
		}
		return b
	}

	m, err := game.LoadMap(must("WARPOS.DAT"), must("TERNAME.DAT"), must("NWMAP.DAT"))
	if err != nil {
		return err
	}
	fonts, err := render.LoadPanelFonts(must("1.15"), must("2.15"), must("3.15"), must("MAN115"))
	if err != nil {
		return err
	}
	w4, err := assets.ParseGlyphFile(must("4.15"))
	if err != nil {
		return err
	}
	cmdFonts := render.CommandFonts{W2: fonts.W2, W4: w4}

	icons, err := render.LoadIcons(must("NEWICON.TPC"))
	if err != nil {
		return err
	}
	battles, err := game.ParseBattleStates(must("MEM_WAR.DAT"))
	if err != nil {
		return err
	}
	// 用 EGA 預設調色盤——原版戰場配哪個 .RGB 還沒查出來（8 個檔名都不像戰場），
	// 所以顏色不保證與實機逐像素相同（internal/assets/palette.go 的說明）。
	ts, err := render.LoadTileSet(must("NEWTERR.TPC"), must("RAIL.TPC"),
		assets.EGADefaultPalette)
	if err != nil {
		return err
	}
	var battlefieldTheme uitheme.Theme = render.NewRetroThemeWithIcons(ts, icons, assets.EGADefaultPalette)
	if themeMode == uitheme.ModeModern {
		modern := uitheme.NewModern()
		if ornaments, ornamentErr := uitheme.DecodeOriginalOrnaments(read); ornamentErr != nil {
			fmt.Fprintf(os.Stderr, "原版介面裝飾載入失敗（Modern A2 截圖改用程式化邊框）：%v\n", ornamentErr)
		} else {
			modern.SetOriginalOrnaments(ornaments)
		}
		battlefieldTheme = modern
	}
	var eten *assets.EtenFonts
	var locale *i18n.Locale
	var wording *i18n.WordingCatalog
	var people *i18n.PeopleDB
	var portraits *i18n.PortraitCatalog
	var labels render.PanelLabels
	commandLabels := make([]string, 15)
	narrativeLabel := ""
	if themeMode == uitheme.ModeModern {
		var err error
		eten, err = assets.LoadEtenFonts(etenDir)
		if err != nil {
			return fmt.Errorf("Modern 截圖需要完整倚天字庫：%w", err)
		}
		locale, err = i18n.Load(localeDir)
		if err != nil {
			return fmt.Errorf("Modern 截圖語系載入失敗：%w", err)
		}
		wording, err = i18n.LoadWording(localeDir)
		if err != nil {
			return fmt.Errorf("Modern 截圖用語載入失敗：%w", err)
		}
		sharedDir := filepath.Join(filepath.Dir(localeDir), "shared")
		people, err = i18n.LoadPeople(localeDir, sharedDir)
		if err != nil {
			return fmt.Errorf("Modern 截圖人物語系載入失敗：%w", err)
		}
		portraits, err = i18n.LoadPortraitCatalog(sharedDir, people)
		if err != nil {
			return fmt.Errorf("Modern 截圖肖像登錄載入失敗：%w", err)
		}
		values := []*string{
			&labels.Status, &labels.Commander, &labels.Governor,
			&labels.Gold, &labels.Food, &labels.Ammo, &labels.Fuel,
			&labels.Coal, &labels.Iron, &labels.Land, &labels.Population,
			&labels.Cities, &labels.Arsenal, &labels.Force, &labels.Generals,
			&labels.People, &labels.Loyalty, &labels.Commands, &labels.Count,
		}
		keys := []string{
			"panel.status.normal", "panel.commander", "panel.governor",
			"panel.gold", "panel.food", "panel.ammo", "panel.fuel",
			"panel.coal", "panel.iron", "panel.land", "panel.population",
			"panel.cities", "panel.arsenal", "panel.force", "panel.generals",
			"panel.people", "panel.loyalty", "panel.commands", "panel.count",
		}
		for i, key := range keys {
			entry, ok := wording.Text(key, i18n.WordingOriginal)
			if !ok {
				return fmt.Errorf("Modern 截圖缺少語意鍵 %q", key)
			}
			*values[i] = entry
		}
		for i := range commandLabels {
			entry, ok := wording.Text(fmt.Sprintf("command.%02d", i+1), i18n.WordingOriginal)
			if !ok {
				return fmt.Errorf("Modern 截圖缺少語意鍵 %q", fmt.Sprintf("command.%02d", i+1))
			}
			commandLabels[i] = entry
		}
		var wordingOK bool
		narrativeLabel, wordingOK = wording.Text("narrative.open", i18n.WordingOriginal)
		if !wordingOK {
			return fmt.Errorf("Modern 截圖缺少語意鍵 %q", "narrative.open")
		}
	}
	unitTheme, ok := battlefieldTheme.(uitheme.UnitProvider)
	if !ok || unitTheme == nil {
		return fmt.Errorf("主題 %s 的部隊圖示尚未載入", themeMode)
	}
	var hudIcons uitheme.HUDIconProvider
	if themeMode == uitheme.ModeModern {
		if hudIcons, ok = battlefieldTheme.(uitheme.HUDIconProvider); !ok || hudIcons == nil {
			return fmt.Errorf("主題 %s 的 HUD 圖示尚未載入", themeMode)
		}
	}
	tbl, err := game.ParseSaveProvinces(must("SAVE(1).DT1"))
	if err != nil {
		return err
	}
	generals, err := game.ParseGenerals(must("MAN(1).DAT"),
		len(fonts.Gen.Glyphs)/game.GeneralNameSlotWidth)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	if biography {
		return writeModernBiographyScreenshots(out, only, tbl, people, portraits, wording, eten,
			battlefieldTheme.(uitheme.StyleProvider).Style())
	}

	for id := game.ProvinceID(1); id <= game.ProvinceCount; id++ {
		if only != 0 && id != only {
			continue
		}
		bf, err := m.Battlefield(id)
		if err != nil {
			return err
		}
		p, err := tbl.At(id)
		if err != nil {
			return err
		}
		d := render.PanelData{
			ID: id, Province: p,
			Force:    game.ForceOf(generals, id),
			Generals: game.CountOf(generals, id),
			Icons:    hudIcons,
		}
		d.Style = uitheme.RetroStyle()
		if themeMode == uitheme.ModeModern {
			d.Style = battlefieldTheme.(uitheme.StyleProvider).Style()
			d.SemanticFonts = eten
			d.Labels = labels
			d.ProvinceName = locale.Province(int(id))
			// SAVE(1).DT1 是第一期截圖來源；人物槽位表以期別與槽位
			// 接合，找不到時保留空白而不猜人名。
			if person, ok := people.PersonAt(1, int(p.Commander)); ok {
				d.CommanderName = person.NameInGame
			}
			if person, ok := people.PersonAt(1, int(p.Governor)); ok {
				d.GovernorName = person.NameInGame
			}
		}
		d.Commands = d.Generals/8 + 1
		if tbl.Date != nil {
			d.Year, d.Month = tbl.Date.Year, tbl.Date.Month
		}
		var c *render.Canvas
		if resolutionMode == uiresolution.ModeHigh {
			c = render.NewCanvas(uiresolution.HighWidth, uiresolution.HighHeight)
			if battle {
				if err := drawModernLiveBattle(c, m, tbl, generals, battlefieldTheme, unitTheme,
					eten, wording, people, id); err != nil {
					return err
				}
			} else {
				portrait, portraitUnavailable := screenshotPortrait(people, portraits, wording, p.Commander)
				if err := c.DrawModernMapSurface(render.ModernMapSurfaceData{
					Battlefield:         bf,
					Theme:               battlefieldTheme,
					Style:               d.Style,
					Panel:               d,
					Fonts:               eten,
					Command:             labels.Commands,
					Narrative:           narrativeLabel,
					Portrait:            portrait,
					PortraitUnavailable: portraitUnavailable,
				}); err != nil {
					return err
				}
			}
			if menu && !battle {
				title, ok := wording.Text("command.title", i18n.WordingOriginal)
				if !ok {
					return fmt.Errorf("Modern 截圖缺少語意鍵 %q", "command.title")
				}
				if err := c.DrawModernCommandSurface(render.ModernCommandSurfaceData{
					Title: title, Labels: commandLabels, Fonts: eten, Icons: hudIcons, Style: d.Style,
				}); err != nil {
					return err
				}
			}
		} else {
			c = render.NewBGICanvas()
			if err := c.DrawThemedBattlefield(bf, battlefieldTheme, 190, 0); err != nil {
				return err
			}
			if err := c.DrawStrategyPanel(d, fonts); err != nil {
				return err
			}
		}
		if battle && resolutionMode != uiresolution.ModeHigh {
			// 跑一場真的戰鬥再畫——單位的位置由部署規則決定
			// （`docs/re/07` §5），不是排版擺的。
			if err := drawLiveBattle(c, m, tbl, generals, unitTheme, id); err != nil {
				return err
			}
		}
		if units {
			// 把參戰單位擺在戰場上。**位置是 remake 的排版選擇**——
			// 原版每個單位在哪一格由戰鬥狀態決定，那部分還沒解出來
			// （200 B 的單位詳細資料，docs/re/05）。
			st := battles[id-1]
			for i, gid := range st.Attackers() {
				_ = gid
				bitmap, err := unitTheme.Unit(0)
				if err != nil {
					return err
				}
				if err := c.DrawThemedUnitIcon(bitmap, 190, 0, i, 0); err != nil {
					return err
				}
			}
			for i, gid := range st.Defenders() {
				_ = gid
				bitmap, err := unitTheme.Unit(1)
				if err != nil {
					return err
				}
				if err := c.DrawThemedUnitIcon(bitmap, 190, 0, i, 13); err != nil {
					return err
				}
			}
		}
		if menu && resolutionMode != uiresolution.ModeHigh {
			// 指令選單蓋在戰場上，對照實機的「司令，請下命令？」清單。
			if err := c.DrawCommandPageWithIcons(cmdFonts, hudIcons,
				assets.RGB{R: 0xAE, G: 0x00, B: 0x00},
				assets.RGB{R: 0xFF, G: 0xFF, B: 0xA2},
				190, 0, render.ModeBGIW-190, render.ModeBGIH-14); err != nil {
				return err
			}
		}

		name := fmt.Sprintf("province-%02d.png", id)
		if battle {
			name = fmt.Sprintf("battle-%02d.png", id)
		} else if menu {
			name = fmt.Sprintf("menu-%02d.png", id)
		}
		path := filepath.Join(out, name)
		f, err := os.Create(path)
		if err != nil {
			return err
		}
		if err := png.Encode(f, c.Image()); err != nil {
			f.Close()
			return err
		}
		f.Close()
		fmt.Println("寫出", path)
	}
	return nil
}

func screenshotPortrait(people *i18n.PeopleDB, portraits *i18n.PortraitCatalog,
	wording *i18n.WordingCatalog, general game.GeneralID) (*i18n.Portrait, string) {
	var person *i18n.Person
	if people != nil {
		person, _ = people.PersonAt(1, int(general))
	}
	return screenshotPortraitForPerson(portraits, wording, person)
}

func screenshotPortraitForPerson(portraits *i18n.PortraitCatalog,
	wording *i18n.WordingCatalog, person *i18n.Person) (*i18n.Portrait, string) {
	placeholder := ""
	if wording != nil {
		placeholder, _ = wording.Text("biography.portrait_unavailable", i18n.WordingOriginal)
	}
	if portraits == nil || person == nil {
		return nil, placeholder
	}
	portrait, ok := portraits.PortraitFor(person.ID)
	if !ok {
		return nil, placeholder
	}
	return &portrait, placeholder
}

// writeModernBiographyScreenshots 只把現有 PeopleDB 的人物資料交給高解析
// renderer；它不建立新的人物名冊，也不執行任何規則／存檔操作。首頁與末頁
// 分開輸出，讓長文換行、來源提示與翻頁狀態都有可重生的視覺證據。
func writeModernBiographyScreenshots(out string, province game.ProvinceID,
	tbl *game.ProvinceTable, people *i18n.PeopleDB, portraits *i18n.PortraitCatalog, wording *i18n.WordingCatalog,
	fonts *assets.EtenFonts, style uitheme.UIStyle) error {
	if tbl == nil || people == nil || wording == nil {
		return fmt.Errorf("人物自傳截圖缺少 province table／PeopleDB／wording")
	}
	title, ok := wording.Text("biography.page", i18n.WordingOriginal)
	if !ok {
		return fmt.Errorf("Modern 自傳截圖缺少語意鍵 %q", "biography.page")
	}
	unavailable, ok := wording.Text("biography.unavailable", i18n.WordingOriginal)
	if !ok {
		return fmt.Errorf("Modern 自傳截圖缺少語意鍵 %q", "biography.unavailable")
	}
	if province == 0 {
		province = 26
	}
	p, err := tbl.At(province)
	if err != nil {
		return err
	}
	person, ok := people.PersonAt(1, int(p.Commander))
	if !ok || person == nil || person.Biography == "" {
		// 省份司令可能是沒有可接合傳記的槽位；為了不猜姓名，
		// 只在同一 PeopleDB 內依遞增 id 找第一筆有正文的人物。
		for id := 1; id <= 1000; id++ {
			candidate, found := people.PersonByID(id)
			if found && candidate != nil && candidate.Biography != "" {
				person = candidate
				ok = true
				break
			}
		}
	}
	if !ok || person == nil {
		return fmt.Errorf("找不到可展示的人物自傳")
	}
	sourceNotice := ""
	if person.BiographyStatus == "source-fallback" {
		if sourceNotice, ok = wording.Text("biography.source_fallback", i18n.WordingOriginal); !ok {
			return fmt.Errorf("Modern 自傳截圖缺少語意鍵 %q", "biography.source_fallback")
		}
	}
	view := render.BiographyView{
		Person: person, Title: title, Unavailable: unavailable, SourceNotice: sourceNotice,
	}
	portrait, portraitUnavailable := screenshotPortraitForPerson(portraits, wording, person)
	first := render.NewCanvas(uiresolution.HighWidth, uiresolution.HighHeight)
	result, err := first.DrawModernBiographySurface(fonts, view,
		style.Ink, style.Paper, style, portrait, portraitUnavailable, "返回", "上一頁", "下一頁")
	if err != nil {
		return err
	}
	if len(result.Missing) != 0 {
		return fmt.Errorf("Modern 自傳截圖缺字：%q", string(result.Missing))
	}
	if result.PageCount <= 0 {
		return fmt.Errorf("Modern 自傳沒有有效頁數")
	}
	writePage := func(page int, canvas *render.Canvas) error {
		path := filepath.Join(out, fmt.Sprintf("biography-%03d-%02d.png", person.ID, page+1))
		f, createErr := os.Create(path)
		if createErr != nil {
			return createErr
		}
		if encodeErr := png.Encode(f, canvas.Image()); encodeErr != nil {
			_ = f.Close()
			return encodeErr
		}
		if closeErr := f.Close(); closeErr != nil {
			return closeErr
		}
		fmt.Println("寫出", path)
		return nil
	}
	if err := writePage(0, first); err != nil {
		return err
	}
	if result.PageCount > 1 {
		last := render.NewCanvas(uiresolution.HighWidth, uiresolution.HighHeight)
		view.Page = result.PageCount - 1
		lastResult, err := last.DrawModernBiographySurface(fonts, view,
			style.Ink, style.Paper, style, portrait, portraitUnavailable, "返回", "上一頁", "下一頁")
		if err != nil {
			return err
		}
		if len(lastResult.Missing) != 0 || lastResult.PageCount != result.PageCount {
			return fmt.Errorf("Modern 自傳末頁缺字或頁數不一致：%+v", lastResult)
		}
		if err := writePage(result.PageCount-1, last); err != nil {
			return err
		}
	}
	return nil
}

// buildLiveBattle 用 BattleSim 建立一份只讀展示快照。
//
// 與 `-units` 的差別是**單位位置由規則決定**：攻方走部署掃描落在
// 進場區，守方在腹地。這是規則層與呈現層第一次串起來；呼叫端不得
// 由這份快照回寫規則或存檔。
func buildLiveBattle(m *game.Map, tbl *game.ProvinceTable,
	generals []game.General, at game.ProvinceID) (*game.BattleSim, error) {
	from := tbl.FirstAttackable(at)
	if from == 0 {
		// 沒有可攻打的鄰省就挑第一個鄰省，純粹為了畫得出東西。
		ns, err := m.Neighbours(at)
		if err != nil || len(ns) == 0 {
			return nil, nil
		}
		from = ns[0]
	}
	mk := func(gs []game.General, prov game.ProvinceID, base int) []*game.Combatant {
		var out []*game.Combatant
		for i := range gs {
			if len(out) >= game.UnitsPerSide {
				break
			}
			id := game.GeneralID(base + i + 1)
			out = append(out, &game.Combatant{
				CombatUnit: game.CombatUnit{
					General: id, Faction: game.GeneralID(prov), Cell: game.NoCell,
					Province: prov, Max: 12, Current: 12, Active: true,
					Decaying: 80, Facing: gs[i].Range,
				},
				Strength: game.StrengthInput{
					Ability: gs[i].AbilityA, Force: gs[i].Force,
					F19: gs[i].F19, F20: gs[i].F20, F29: gs[i].Stamina, F30: gs[i].F30,
					Branch: gs[i].Branch, General: id, Faction: game.GeneralID(prov),
				},
			})
		}
		return out
	}
	atk := mk(game.GeneralsOf(generals, from), from, 0)
	def := mk(game.GeneralsOf(generals, at), at, 500)
	if len(atk) == 0 || len(def) == 0 {
		return nil, nil
	}

	bf, err := m.Battlefield(at)
	if err != nil {
		return nil, err
	}
	placed := 0
	for i := 0; i < game.CellCount && placed < len(def); i++ {
		cc := game.CellIndex(i)
		col, row := cc.ColRow()
		if bf.Owner[row][col] != 0 || bf.Tiles[row][col].MoveCost() >= assets.MoveCostImpassable {
			continue
		}
		def[placed].Cell = cc
		placed++
	}
	def = def[:placed]
	if placed == 0 {
		return nil, nil
	}

	sim, err := game.NewBattleSim(m, at, from, atk, def, game.StrengthOpts{Stage: 1})
	if err != nil {
		return nil, err
	}
	return sim, nil
}

// drawLiveBattle 用 BattleSim 跑一場戰鬥，把佈署後的樣子畫出來。
func drawLiveBattle(c *render.Canvas, m *game.Map, tbl *game.ProvinceTable,
	generals []game.General, unitTheme uitheme.UnitProvider, at game.ProvinceID) error {
	sim, err := buildLiveBattle(m, tbl, generals, at)
	if err != nil || sim == nil {
		return err
	}
	bf := sim.Field
	// 印出佈署，讓落點能與 `docs/re/07` §5 的進場區對照。
	fmt.Printf("  戰鬥：省 %d ← 省 %d　攻 %d／守 %d\n",
		at, sim.From, len(sim.Attacker), len(sim.Defender))
	for _, u := range sim.Attacker {
		col, row := u.Cell.ColRow()
		fmt.Printf("    攻 %d 落在格 %d (欄%2d,列%2d)，該格的 WARPOS = %d\n",
			u.General, u.Cell, col, row, bf.Owner[row][col])
	}
	for _, u := range sim.Defender {
		idx := render.BranchIcon(u.Branch(), true, u.Facing)
		bitmap, err := unitTheme.Unit(idx)
		if err != nil {
			return err
		}
		if err := c.DrawThemedUnitAtCell(bitmap, 190, 0, u.Cell); err != nil {
			return err
		}
	}
	for _, u := range sim.Attacker {
		idx := render.BranchIcon(u.Branch(), false, u.Facing)
		bitmap, err := unitTheme.Unit(idx)
		if err != nil {
			return err
		}
		if err := c.DrawThemedUnitAtCell(bitmap, 190, 0, u.Cell); err != nil {
			return err
		}
	}
	return nil
}

// drawModernLiveBattle 把同一份 BattleSim 佈署快照交給 H1-c renderer。
// 它只整理面板／語系資料，不執行攻擊、AI 或存檔，讓 high screenshot 與
// cmd/dsds 的正常戰鬥路徑共用規則入口。
func drawModernLiveBattle(c *render.Canvas, m *game.Map, tbl *game.ProvinceTable,
	generals []game.General, theme uitheme.Theme, unitTheme uitheme.UnitProvider,
	fonts *assets.EtenFonts, wording *i18n.WordingCatalog, people *i18n.PeopleDB,
	at game.ProvinceID) error {
	sim, err := buildLiveBattle(m, tbl, generals, at)
	if err != nil || sim == nil {
		return err
	}
	labels, err := modernBattleLabels(wording, people, tbl, sim)
	if err != nil {
		return err
	}
	panel, err := modernBattlePanel(tbl, sim)
	if err != nil {
		return err
	}
	var current game.CellIndex
	if len(sim.Attacker) != 0 && sim.Attacker[0] != nil {
		current = sim.Attacker[0].Cell
	}
	style := uitheme.NewModern().Style()
	if provider, ok := theme.(uitheme.StyleProvider); ok {
		style = provider.Style()
	}
	return c.DrawModernBattleSurface(render.ModernBattleSurfaceData{
		Battlefield: sim.Field,
		Theme:       theme,
		Units:       unitTheme,
		Attackers:   sim.Attacker,
		Defenders:   sim.Defender,
		CurrentCell: current,
		Panel:       panel,
		Labels:      labels,
		Fonts:       fonts,
		Style:       style,
	})
}

func modernBattleLabels(wording *i18n.WordingCatalog, people *i18n.PeopleDB,
	tbl *game.ProvinceTable, sim *game.BattleSim) (render.ModernBattleLabels, error) {
	var labels render.ModernBattleLabels
	get := func(key string) (string, error) {
		if wording == nil {
			return "", fmt.Errorf("Modern 戰鬥截圖缺少 wording catalog")
		}
		value, ok := wording.Text(key, i18n.WordingOriginal)
		if !ok || value == "" {
			return "", fmt.Errorf("Modern 戰鬥截圖缺少語意鍵 %q", key)
		}
		return value, nil
	}
	var err error
	if labels.Title, err = get("battle.title"); err != nil {
		return labels, err
	}
	if labels.Attacker, err = get("battle.attacker"); err != nil {
		return labels, err
	}
	if labels.Defender, err = get("battle.defender"); err != nil {
		return labels, err
	}
	if labels.Attack, err = get("battle.attack"); err != nil {
		return labels, err
	}
	if labels.Province, err = get("battle.province"); err != nil {
		return labels, err
	}
	if labels.Date, err = get("battle.date"); err != nil {
		return labels, err
	}
	for i, key := range []string{"battle.units", "battle.soldiers", "battle.gold", "battle.food", "battle.ammo", "battle.fuel"} {
		value, getErr := get(key)
		if getErr != nil {
			return labels, getErr
		}
		switch i {
		case 0:
			labels.Units = value
		case 1:
			labels.Soldiers = value
		case 2:
			labels.Gold = value
		case 3:
			labels.Food = value
		case 4:
			labels.Ammo = value
		case 5:
			labels.Fuel = value
		}
	}
	for i, key := range []string{"battle.command.move", "battle.command.attack", "battle.command.retreat", "battle.command.garrison", "battle.command.inspect"} {
		if labels.Commands[i], err = get(key); err != nil {
			return labels, err
		}
	}
	for i, key := range []string{"battle.control.attack", "battle.control.next", "battle.control.end"} {
		if labels.Controls[i], err = get(key); err != nil {
			return labels, err
		}
	}
	if labels.Mode, err = get("battle.mode.retreat"); err != nil {
		return labels, err
	}
	if labels.Delete, err = get("battle.delete"); err != nil {
		return labels, err
	}
	if labels.Submit, err = get("battle.submit"); err != nil {
		return labels, err
	}
	if sim == nil || tbl == nil {
		return labels, fmt.Errorf("Modern 戰鬥截圖缺少 sim／province table")
	}
	name := func(id game.GeneralID) string {
		if people == nil {
			return ""
		}
		if person, ok := people.PersonAt(1, int(id)); ok && person != nil {
			return person.NameInGame
		}
		return ""
	}
	if p, lookupErr := tbl.At(sim.From); lookupErr == nil {
		labels.AttackerName = name(p.Commander)
	}
	if p, lookupErr := tbl.At(sim.At); lookupErr == nil {
		labels.DefenderName = name(p.Commander)
	}
	return labels, nil
}

func modernBattlePanel(tbl *game.ProvinceTable, sim *game.BattleSim) (render.BattlePanelData, error) {
	if tbl == nil || sim == nil {
		return render.BattlePanelData{}, fmt.Errorf("Modern 戰鬥截圖缺少 sim／province table")
	}
	month := uint8(0)
	if tbl.Date != nil {
		month = tbl.Date.Month
	}
	count := func(units []*game.Combatant) (uint32, uint32) {
		var alive, soldiers uint32
		for _, unit := range units {
			if unit == nil || !unit.Alive() {
				continue
			}
			alive++
			soldiers += uint32(unit.Strength.Force)
		}
		return alive, soldiers
	}
	attackerUnits, attackerSoldiers := count(sim.Attacker)
	defenderUnits, defenderSoldiers := count(sim.Defender)
	panel := render.BattlePanelData{
		Province:           sim.At,
		Month:              month,
		Day:                1,
		ShowBattleMenu:     true,
		ShowBattleControls: true,
		BattleMenuMode:     render.BattleMenuCommand,
		Style:              uitheme.NewModern().Style(),
	}
	panel.Attacker.Units = attackerUnits
	panel.Attacker.Soldiers = attackerSoldiers
	panel.Defender.Units = defenderUnits
	panel.Defender.Soldiers = defenderSoldiers
	if p, err := tbl.At(sim.From); err == nil {
		panel.Attacker.Leader = p.Commander
		panel.Attacker.Gold, panel.Attacker.Food = uint32(p.Gold), uint32(p.Food)
		panel.Attacker.Ammo, panel.Attacker.Fuel = uint32(p.Ammo), uint32(p.Fuel)
	}
	if p, err := tbl.At(sim.At); err == nil {
		panel.Defender.Leader = p.Commander
		panel.Defender.Gold, panel.Defender.Food = uint32(p.Gold), uint32(p.Food)
		panel.Defender.Ammo, panel.Defender.Fuel = uint32(p.Ammo), uint32(p.Fuel)
	}
	return panel, nil
}
