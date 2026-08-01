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
	"github.com/wicanr2/great-era-remake/internal/ui/render"
)

func main() {
	dir := flag.String("game", "workplace/orig/game", "原版素材目錄（唯讀）")
	out := flag.String("out", "workplace/shots", "輸出目錄")
	prov := flag.Int("province", 0, "只畫某一省（1-39），0 = 全部")
	menu := flag.Bool("menu", false, "右側改畫政略指令選單")
	units := flag.Bool("units", false, "在戰場上疊出參戰單位的圖示")
	battle := flag.Bool("battle", false, "畫一場實際的戰鬥（跑 BattleSim）")
	flag.Parse()

	if err := run(*dir, *out, game.ProvinceID(*prov), *menu, *units, *battle); err != nil {
		fmt.Fprintln(os.Stderr, "錯誤:", err)
		os.Exit(1)
	}
}

func run(dir, out string, only game.ProvinceID, menu, units, battle bool) error {
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
	fonts, err := render.LoadPanelFonts(must("2.15"), must("3.15"), must("MAN115"))
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

	for id := game.ProvinceID(1); id <= game.ProvinceCount; id++ {
		if only != 0 && id != only {
			continue
		}
		c := render.NewBGICanvas()
		bf, err := m.Battlefield(id)
		if err != nil {
			return err
		}
		if err := c.DrawTiledBattlefield(bf, ts, 190, 0); err != nil {
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
		}
		if tbl.Date != nil {
			d.Year, d.Month = tbl.Date.Year, tbl.Date.Month
		}
		if err := c.DrawStrategyPanel(d, fonts); err != nil {
			return err
		}
		if battle {
			// 跑一場真的戰鬥再畫——單位的位置由部署規則決定
			// （`docs/re/07` §5），不是排版擺的。
			if err := drawLiveBattle(c, m, tbl, generals, icons, id); err != nil {
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
				if err := c.DrawUnitIcon(icons, 0, assets.EGADefaultPalette,
					190, 0, i, 0); err != nil {
					return err
				}
			}
			for i, gid := range st.Defenders() {
				_ = gid
				if err := c.DrawUnitIcon(icons, 1, assets.EGADefaultPalette,
					190, 0, i, 13); err != nil {
					return err
				}
			}
		}
		if menu {
			// 指令選單蓋在戰場上，對照實機的「司令，請下命令？」清單。
			if err := c.DrawCommandPage(cmdFonts,
				assets.RGB{R: 0xAE, G: 0x00, B: 0x00},
				assets.RGB{R: 0xFF, G: 0xFF, B: 0xA2},
				190, 0, render.ModeBGIW-190, render.ModeBGIH-14); err != nil {
				return err
			}
		}

		name := fmt.Sprintf("province-%02d.png", id)
		if menu {
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

// drawLiveBattle 用 BattleSim 跑一場戰鬥，把佈署後的樣子畫出來。
//
// 與 `-units` 的差別是**單位位置由規則決定**：攻方走部署掃描落在
// 進場區，守方在腹地。這是規則層與呈現層第一次串起來。
func drawLiveBattle(c *render.Canvas, m *game.Map, tbl *game.ProvinceTable,
	generals []game.General, icons []*assets.Image, at game.ProvinceID) error {
	from := tbl.FirstAttackable(at)
	if from == 0 {
		// 沒有可攻打的鄰省就挑第一個鄰省，純粹為了畫得出東西。
		ns, err := m.Neighbours(at)
		if err != nil || len(ns) == 0 {
			return nil
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
		return nil
	}

	bf, err := m.Battlefield(at)
	if err != nil {
		return err
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
		return nil
	}

	sim, err := game.NewBattleSim(m, at, from, atk, def, game.StrengthOpts{Stage: 1})
	if err != nil {
		return err
	}
	// 印出佈署，讓落點能與 `docs/re/07` §5 的進場區對照。
	fmt.Printf("  戰鬥：省 %d ← 省 %d　攻 %d／守 %d\n",
		at, from, len(sim.Attacker), len(sim.Defender))
	for _, u := range sim.Attacker {
		col, row := u.Cell.ColRow()
		fmt.Printf("    攻 %d 落在格 %d (欄%2d,列%2d)，該格的 WARPOS = %d\n",
			u.General, u.Cell, col, row, bf.Owner[row][col])
	}
	for _, u := range sim.Defender {
		idx := render.BranchIcon(u.Branch(), true, u.Facing)
		if err := c.DrawUnitAtCell(icons, idx, assets.EGADefaultPalette, 190, 0, u.Cell); err != nil {
			return err
		}
	}
	for _, u := range sim.Attacker {
		idx := render.BranchIcon(u.Branch(), false, u.Facing)
		if err := c.DrawUnitAtCell(icons, idx, assets.EGADefaultPalette, 190, 0, u.Cell); err != nil {
			return err
		}
	}
	return nil
}
