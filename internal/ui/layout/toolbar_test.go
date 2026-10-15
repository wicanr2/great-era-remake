package layout

import "testing"

func TestToolbarKeepsCompleteOriginalAndHighGameSurface(t *testing.T) {
	for _, tc := range []struct{ w, h, bar int }{{640, 350, 24}, {1280, 720, 48}} {
		if got := ToolbarHeightForWidth(tc.w); got != tc.bar {
			t.Fatalf("height=%d", got)
		}
		for _, point := range [][2]int{{0, 0}, {tc.w - 1, tc.h - 1}, {tc.w / 2, tc.h / 2}} {
			x, y, ok := ToolbarToGame(point[0], point[1]+tc.bar, tc.w, tc.h, true)
			if !ok || x != point[0] || y != point[1] {
				t.Fatalf("lost game point: %d %d %v", x, y, ok)
			}
		}
		for _, point := range [][2]int{{0, tc.bar - 1}, {0, tc.h + tc.bar}, {tc.w, tc.bar}, {-1, tc.bar}} {
			if _, _, ok := ToolbarToGame(point[0], point[1], tc.w, tc.h, true); ok {
				t.Fatalf("outside point accepted: %v", point)
			}
		}
	}
}

func TestToolbarDropdownRowsDoNotIncludeHelpOrBorders(t *testing.T) {
	r := ToolbarMenu(2, 1, 6)
	for _, tc := range []struct{ x, y, row int }{{r.X + 12, r.Y + 5, 0}, {r.X + 12, r.Y + 40, 0}, {r.X + 12, r.Y + 41, -2}, {r.X + 12, r.Bottom() - 1, -2}} {
		field, row := ToolbarHit(tc.x, tc.y, 2, 1, 6)
		if field != 2 || row != tc.row {
			t.Fatalf("hit=%d/%d wanted 2/%d", field, row, tc.row)
		}
	}
	if field, _ := ToolbarHit(r.Right(), r.Bottom()-1, 2, 1, 6); field != -1 {
		t.Fatal("right border leaked")
	}
	if r := ToolbarMenu(-1, 0, 0); r.W != 0 {
		t.Fatal("invalid menu has geometry")
	}
}
