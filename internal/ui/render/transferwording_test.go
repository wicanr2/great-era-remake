package render

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

func TestPlainTransferScreensHaveNoMissingGlyphs(t *testing.T) {
	fonts, err := assets.LoadEtenFonts(filepath.Join("..", "..", "..", "workplace", "eten"))
	if err != nil {
		t.Skipf("沒有使用者倚天字庫：%v", err)
	}
	fg, bg := assets.RGB{R: 174}, assets.RGB{R: 255, G: 255, B: 162}
	tests := []struct {
		name string
		draw func(*Canvas) []rune
	}{
		{"方式", func(c *Canvas) []rune {
			return c.DrawPlainTransferMode(fonts, fg, bg, 190, 0, 450, 350, "要怎麼調動？", "選擇部分將領", "調動全部可用將領")
		}},
		{"物資", func(c *Canvas) []rune {
			return c.DrawPlainTransferAmount(fonts, fg, bg, 190, 0, 450, 350, "要一起運送多少黃金？", 1234)
		}},
		{"確認", func(c *Canvas) []rune {
			return c.DrawPlainConfirm(fonts, fg, bg, 190, 0, 450, 350, "確定要執行嗎？")
		}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if missing := tc.draw(NewBGICanvas()); len(missing) != 0 {
				t.Fatalf("缺字：%q", string(missing))
			}
		})
	}
}

func TestPlainTransferTargetAndSelectionKeepOriginalNames(t *testing.T) {
	fonts, err := assets.LoadEtenFonts(filepath.Join("..", "..", "..", "workplace", "eten"))
	if err != nil {
		t.Skipf("沒有使用者倚天字庫：%v", err)
	}
	load := func(name string) *assets.GlyphFile {
		b, err := os.ReadFile(filepath.Join("..", "..", "..", "workplace", "orig", "game", name))
		if err != nil {
			t.Skipf("沒有原版素材 %s：%v", name, err)
		}
		f, err := assets.ParseGlyphFile(b)
		if err != nil {
			t.Fatal(err)
		}
		return f
	}
	fg, bg := assets.RGB{R: 174}, assets.RGB{R: 255, G: 255, B: 162}
	c := NewBGICanvas()
	missing, err := c.DrawPlainTransferTarget(fonts, load("3.15"), fg, bg,
		190, 0, 450, 350, "要調往哪個省？", []int{19, 28}, 19)
	if err != nil || len(missing) != 0 {
		t.Fatalf("白話目標頁：err=%v，缺字=%q", err, string(missing))
	}
	c = NewBGICanvas()
	missing, err = c.DrawPlainTransferSelection(fonts, load("MAN115"), fg, bg,
		190, 0, 450, 350, "要調動哪位將領？", "確認這份名單",
		[]int{58, 59, 60}, []bool{true, false, true}, 1, 0)
	if err != nil || len(missing) != 0 {
		t.Fatalf("白話選將頁：err=%v，缺字=%q", err, string(missing))
	}
}
