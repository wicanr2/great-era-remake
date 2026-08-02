package render

import (
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
)

func TestDrawNumericKeypadDrawsEveryButtonBorder(t *testing.T) {
	bg := assets.RGB{R: 255, G: 255, B: 162}
	fg := assets.RGB{R: 174}
	c := NewCanvas(640, 350)
	c.fillRect(0, 0, 640, 350, bg)
	c.DrawNumericKeypad(fg, bg)
	for i := 0; i < 12; i++ {
		p := uilayout.NumericKeypadButton(i)
		if got := c.Image().RGBAAt(p.X, p.Y); got.R != fg.R || got.G != fg.G || got.B != fg.B {
			t.Fatalf("第 %d 鍵左上邊框=%v", i, got)
		}
	}
}
