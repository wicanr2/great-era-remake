// assetdump 把原版素材解碼成 PNG，用來驗證 internal/assets 這層。
//
//	tools/go.sh run ./cmd/assetdump -game workplace/orig/game -out workplace/godump
//
// 原版素材唯讀，只讀不寫。輸出目錄不要放進版控（含原版美術）。
package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

func main() {
	game := flag.String("game", "workplace/orig/game", "原版素材目錄（唯讀）")
	out := flag.String("out", "workplace/godump", "輸出目錄")
	flag.Parse()

	if err := run(*game, *out); err != nil {
		fmt.Fprintln(os.Stderr, "錯誤:", err)
		os.Exit(1)
	}
}

func run(game, out string) error {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return err
	}
	var nTPC, nGLB int

	// .TPC：用同名的 .RGB 當調色盤，沒有就用 EGA 預設
	entries, err := os.ReadDir(game)
	if err != nil {
		return fmt.Errorf("讀不到素材目錄 %s: %w", game, err)
	}
	for _, e := range entries {
		if !strings.EqualFold(filepath.Ext(e.Name()), ".TPC") {
			continue
		}
		data, err := os.ReadFile(filepath.Join(game, e.Name()))
		if err != nil {
			return err
		}
		set, err := assets.DecodeBGISet(data)
		if err != nil {
			fmt.Printf("  略過 %s: %v\n", e.Name(), err)
			continue
		}
		pal := paletteFor(game, e.Name())
		stem := strings.TrimSuffix(e.Name(), filepath.Ext(e.Name()))
		for i, im := range set {
			name := fmt.Sprintf("%s.png", stem)
			if len(set) > 1 {
				name = fmt.Sprintf("%s-%03d.png", stem, i)
			}
			if err := writePNG(filepath.Join(out, name), im.ToRGBA(pal)); err != nil {
				return err
			}
			nTPC++
		}
	}

	// .GLB 圖庫
	for _, stem := range []string{"GRT", "GRT2"} {
		gtb, err1 := os.ReadFile(filepath.Join(game, stem+".GTB"))
		glb, err2 := os.ReadFile(filepath.Join(game, stem+".GLB"))
		if err1 != nil || err2 != nil {
			continue
		}
		g, err := assets.OpenGLB(gtb, glb)
		if err != nil {
			return fmt.Errorf("%s: %w", stem, err)
		}
		for i, ent := range g.Entries {
			if ent.IsPalette() {
				continue
			}
			sp, err := g.Sprite(i)
			if err != nil {
				return fmt.Errorf("%s #%d: %w", stem, i, err)
			}
			pi := g.PaletteFor(i)
			if pi < 0 {
				continue
			}
			pal, err := g.Palette(pi)
			if err != nil {
				return err
			}
			im := &assets.Image{W: sp.W, H: sp.H, Pix: sp.Pix}
			name := fmt.Sprintf("%s-%02d-%dx%d.png", stem, i, sp.W, sp.H)
			if err := writePNG(filepath.Join(out, name), im.ToRGBA(pal)); err != nil {
				return err
			}
			nGLB++
		}
	}

	fmt.Printf("輸出 %d 張 .TPC、%d 張 .GLB → %s\n", nTPC, nGLB, out)
	return nil
}

// paletteFor 找同名的 .RGB 當調色盤。CHOOSE1.TPC → CHOOSE.RGB。
func paletteFor(game, tpc string) assets.Palette {
	stem := strings.TrimSuffix(tpc, filepath.Ext(tpc))
	for len(stem) > 0 {
		f, err := os.Open(filepath.Join(game, stem+".RGB"))
		if err == nil {
			defer f.Close()
			if pal, err := assets.ParseRGBFile(f); err == nil {
				return pal
			}
		}
		stem = stem[:len(stem)-1] // CHOOSE1 → CHOOSE
	}
	return assets.EGA16
}

func writePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()
	return png.Encode(f, img)
}
