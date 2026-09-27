package render

import (
	"testing"

	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

func TestHomageCornerMarksFourCorners(t *testing.T) {
	modern := uitheme.NewModern()
	style := modern.Style()
	c := NewCanvas(120, 90)
	r := uilayout.Rect{X: 10, Y: 10, W: 100, H: 70}
	drawModernDecoratedControl(c, r, style, style.Panel, style.Muted)
	// 四個花框角的外角像素應為暗紅，內圈為寶藍。
	for _, p := range [][2]int{
		{11, 11}, {108, 11}, {11, 78}, {108, 78},
	} {
		if got, want := c.Image().RGBAAt(p[0], p[1]), colorOf(style.Ink); got != want {
			t.Fatalf("花框角 (%d,%d)=%v，應為 %v", p[0], p[1], got, want)
		}
	}
	for _, p := range [][2]int{
		{15, 15}, {104, 15}, {15, 74}, {104, 74},
	} {
		if got, want := c.Image().RGBAAt(p[0], p[1]), colorOf(style.Focus); got != want {
			t.Fatalf("花框內圈 (%d,%d)=%v，應為 %v", p[0], p[1], got, want)
		}
	}
	// 中央仍是底色，不是框線色。
	if got := c.Image().RGBAAt(60, 45); got != colorOf(style.Panel) {
		t.Fatalf("中央 (60,45)=%v，應為底色 %v", got, colorOf(style.Panel))
	}
}

func TestHomageDecoratedPanelMarksCorners(t *testing.T) {
	modern := uitheme.NewModern()
	style := modern.Style()
	c := NewCanvas(120, 100)
	r := uilayout.Rect{X: 10, Y: 10, W: 100, H: 80}
	drawModernDecoratedPanel(c, r, style, style.Panel)
	if got, want := c.Image().RGBAAt(12, 12), colorOf(style.Ink); got != want {
		t.Fatalf("面板花框角 (12,12)=%v，應為 %v", got, want)
	}
	if got, want := c.Image().RGBAAt(16, 16), colorOf(style.Focus); got != want {
		t.Fatalf("面板花框內圈 (16,16)=%v，應為 %v", got, want)
	}
}

func TestHomageCornerSkipsTinyRects(t *testing.T) {
	modern := uitheme.NewModern()
	style := modern.Style()
	c := NewCanvas(60, 40)
	r := uilayout.Rect{X: 5, Y: 5, W: 20, H: 20}
	drawModernDecoratedControl(c, r, style, style.Panel, style.Muted)
	// 小矩形不畫花框角：(6,6) 應維持底色。
	if got := c.Image().RGBAAt(6, 6); got != colorOf(style.Panel) {
		t.Fatalf("小矩形 (6,6)=%v，不應有花框角", got)
	}
}
