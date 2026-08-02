package actions

import "testing"

func TestSurfaceToLogicalHandlesLetterboxAndPillarbox(t *testing.T) {
	for _, tc := range []struct {
		name         string
		x, y, w, h   int
		wantX, wantY int
		wantOK       bool
	}{
		{"2x", 640, 350, 1280, 700, 320, 175, true},
		{"pillarbox", 60, 70, 1400, 700, 0, 35, true},
		{"pillarbox-black", 50, 350, 1400, 700, 0, 0, false},
		{"letterbox", 640, 10, 1280, 720, 320, 0, true},
		{"letterbox-black", 640, 5, 1280, 720, 0, 0, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			gotX, gotY, ok := SurfaceToLogical(tc.x, tc.y, tc.w, tc.h, 640, 350)
			if ok != tc.wantOK || ok && (gotX != tc.wantX || gotY != tc.wantY) {
				t.Fatalf("got=(%d,%d,%v) want=(%d,%d,%v)", gotX, gotY, ok, tc.wantX, tc.wantY, tc.wantOK)
			}
		})
	}
}

func TestHitUsesHalfOpenEdgesAndStableOrder(t *testing.T) {
	targets := []Target{
		{Select1, Rect{10, 20, 30, 40}},
		{Select2, Rect{40, 20, 30, 40}},
	}
	for _, p := range []struct {
		x, y int
		want Action
	}{{10, 20, Select1}, {39, 59, Select1}, {40, 20, Select2}, {70, 20, None}, {10, 60, None}} {
		if got := Hit(targets, p.x, p.y); got != p.want {
			t.Fatalf("Hit(%d,%d)=%q want %q", p.x, p.y, got, p.want)
		}
	}
}

func TestSelectionSupportsLongVisibleLists(t *testing.T) {
	for _, n := range []int{1, 15, 16, 39, 99} {
		action := Selection(n)
		got, ok := SelectionNumber(action)
		if !ok || got != n {
			t.Fatalf("Selection(%d)=%q parsed=(%d,%v)", n, action, got, ok)
		}
	}
	for _, action := range []Action{None, Back, Selection(0), Selection(100)} {
		if _, ok := SelectionNumber(action); ok {
			t.Fatalf("%q 不應是清單選擇", action)
		}
	}
}

func TestDigitActionsRoundTrip(t *testing.T) {
	for n := 0; n <= 9; n++ {
		got, ok := DigitNumber(Digit(n))
		if !ok || got != n {
			t.Fatalf("Digit(%d)=%q parsed=(%d,%v)", n, Digit(n), got, ok)
		}
	}
	for _, action := range []Action{None, Back, Digit(-1), Digit(10)} {
		if _, ok := DigitNumber(action); ok {
			t.Fatalf("%q 不應是數字輸入", action)
		}
	}
}
