package render

import (
	"testing"

	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

func TestWargameFretCornerMarksFourCorners(t *testing.T) {
	style := uitheme.UIStyle{}
	modern := uitheme.NewModern()
	style = modern.Style()
	c := NewCanvas(120, 90)
	r := uilayout.Rect{X: 10, Y: 10, W: 100, H: 70}
	drawModernDecoratedControl(c, r, style, style.Panel, style.Muted)
	want := colorOf(style.Focus)
	// 四個回紋角的外角像素應為銅章色。
	for _, p := range [][2]int{
		{11, 11}, {108, 11}, {11, 78}, {108, 78},
	} {
		if got := c.Image().RGBAAt(p[0], p[1]); got != want {
			t.Fatalf("回紋角 (%d,%d)=%v，應為 %v", p[0], p[1], got, want)
		}
	}
	// 中央仍是底色，不是框線色。
	if got := c.Image().RGBAAt(60, 45); got != colorOf(style.Panel) {
		t.Fatalf("中央 (60,45)=%v，應為底色 %v", got, colorOf(style.Panel))
	}
}

func TestWargameDecoratedPanelMarksCorners(t *testing.T) {
	modern := uitheme.NewModern()
	style := modern.Style()
	c := NewCanvas(120, 100)
	r := uilayout.Rect{X: 10, Y: 10, W: 100, H: 80}
	drawModernDecoratedPanel(c, r, style, style.Panel)
	want := colorOf(style.Focus)
	for _, p := range [][2]int{
		{12, 12}, {107, 12}, {12, 87}, {107, 87},
	} {
		if got := c.Image().RGBAAt(p[0], p[1]); got != want {
			t.Fatalf("面板回紋角 (%d,%d)=%v，應為 %v", p[0], p[1], got, want)
		}
	}
}

func TestWargameFretCornerSkipsTinyRects(t *testing.T) {
	modern := uitheme.NewModern()
	style := modern.Style()
	c := NewCanvas(60, 40)
	r := uilayout.Rect{X: 5, Y: 5, W: 20, H: 20}
	drawModernDecoratedControl(c, r, style, style.Panel, style.Muted)
	// 小矩形不畫回紋角：(6,6) 應維持底色。
	if got := c.Image().RGBAAt(6, 6); got != colorOf(style.Panel) {
		t.Fatalf("小矩形 (6,6)=%v，不應有回紋角", got)
	}
}
