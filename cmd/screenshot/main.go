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
	flag.Parse()

	if err := run(*dir, *out, game.ProvinceID(*prov)); err != nil {
		fmt.Fprintln(os.Stderr, "錯誤:", err)
		os.Exit(1)
	}
}

func run(dir, out string, only game.ProvinceID) error {
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
		if err := c.DrawTiledBattlefield(bf, ts, 190, 14); err != nil {
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

		path := filepath.Join(out, fmt.Sprintf("province-%02d.png", id))
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

