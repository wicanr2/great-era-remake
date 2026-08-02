package render

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/game"
)

func TestStrategyPanelDrawsCommandCount(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(gameDir, name))
		if err != nil {
			t.Skipf("沒有原版素材 %s，跳過", name)
		}
		return b
	}
	fonts, err := LoadPanelFonts(read("1.15"), read("2.15"), read("3.15"), read("MAN115"))
	if err != nil {
		t.Fatal(err)
	}
	if w1Count >= len(fonts.W1.Glyphs) || (w2Command+1)*2 > len(fonts.W2.Glyphs) {
		t.Fatal("「指令數」詞條超出原版字模範圍")
	}

	p := &game.Province{Commander: 1, Governor: 1}
	c := NewBGICanvas()
	if err := c.DrawStrategyPanel(PanelData{
		ID: 1, Province: p, Commands: 4,
	}, fonts); err != nil {
		t.Fatal(err)
	}

	// 指令數是面板最後一行；右側數值區必須真的畫出「4」，不能只把值傳進資料。
	pixels := 0
	for y := 320; y < 340; y++ {
		for x := panelValue - 20; x <= panelValue; x++ {
			r, g, b, _ := c.Image().At(x, y).RGBA()
			if uint8(r>>8) == panelFG.R && uint8(g>>8) == panelFG.G && uint8(b>>8) == panelFG.B {
				pixels++
			}
		}
	}
	if pixels == 0 {
		t.Error("指令數的右側數值區沒有任何前景像素")
	}
}
