package render

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

// TestStrategyCommandsShape 指令表要是 1..15 連續，且每個詞條索引在字模檔範圍內。
func TestStrategyCommandsShape(t *testing.T) {
	if len(StrategyCommands) != 15 {
		t.Fatalf("政略指令應為 15 個，實得 %d", len(StrategyCommands))
	}
	for i, c := range StrategyCommands {
		if c.Num != i+1 {
			t.Errorf("第 %d 個指令的編號是 %d，預期 %d", i, c.Num, i+1)
		}
		if len(c.Parts) == 0 || len(c.Parts) > 2 {
			t.Errorf("指令 %d 由 %d 個詞條組成，預期 1 或 2", c.Num, len(c.Parts))
		}
	}
}

// TestCommandWordsResolve 每個詞條都要能從原版字模檔取到，不越界。
//
// 這一條同時保證「指令名是用原版詞表拼的」——如果哪天有人改成寫死字串，
// 這個測試就沒有東西可驗了。
func TestCommandWordsResolve(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	read := func(name string) *assets.GlyphFile {
		b, err := os.ReadFile(filepath.Join(gameDir, name))
		if err != nil {
			t.Skipf("沒有原版素材 %s，跳過", name)
		}
		gf, err := assets.ParseGlyphFile(b)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		return gf
	}
	f := CommandFonts{W2: read("2.15"), W4: read("4.15")}

	for _, c := range StrategyCommands {
		for _, p := range c.Parts {
			gf, w := f.W2, 2
			if p.FromW4 {
				gf, w = f.W4, 4
			}
			if end := (p.Index + 1) * w; end > len(gf.Glyphs) {
				t.Errorf("指令 %d 的詞條 %d（槽寬 %d）超出字模檔範圍（共 %d 條）",
					c.Num, p.Index, w, len(gf.Glyphs)/w)
			}
		}
	}
}

// TestCommandMenuDraws 畫得出來且不報錯，選單區域確實有前景像素。
func TestCommandMenuDraws(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	load := func(name string) *assets.GlyphFile {
		b, err := os.ReadFile(filepath.Join(gameDir, name))
		if err != nil {
			t.Skipf("沒有原版素材 %s，跳過", name)
		}
		gf, _ := assets.ParseGlyphFile(b)
		return gf
	}
	f := CommandFonts{W2: load("2.15"), W4: load("4.15")}

	c := NewBGICanvas()
	fg := assets.RGB{R: 0xAE}
	if err := c.DrawCommandMenu(f, fg, 20, 20); err != nil {
		t.Fatal(err)
	}
	n := 0
	b := c.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := c.Image().At(x, y).RGBA()
			if uint8(r>>8) == fg.R && uint8(g>>8) == fg.G && uint8(bl>>8) == fg.B {
				n++
			}
		}
	}
	if n < 1000 {
		t.Errorf("選單只畫出 %d 個前景像素，看起來沒畫成功", n)
	}
	t.Logf("十五個指令畫出 %d 個前景像素", n)
}
