package render

import "testing"

func TestModernMusicMenuKeepsGameOutsideMenuVisible(t *testing.T) {
	v := ToolbarView{Desk: true, Menu: 3}
	v.Options[3] = make([]string, 15)
	c := NewCanvas(1280, 720)
	c.DrawDeskToolbar(nil, v)
	if c.Image().RGBAAt(100, 200).A != 0 {
		t.Fatal("menu canvas hides the game outside the menu")
	}
	if c.Image().RGBAAt(800, 200).A != 255 {
		t.Fatal("music menu is not visible")
	}
	if c.Image().RGBAAt(10, 10).A != 255 {
		t.Fatal("permanent toolbar is not visible")
	}
}
