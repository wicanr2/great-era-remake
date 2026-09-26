package resolution

import "testing"

func TestParseAndToggle(t *testing.T) {
	for _, raw := range []string{"original", "high"} {
		if _, err := Parse(raw); err != nil {
			t.Fatalf("Parse(%q): %v", raw, err)
		}
	}
	if got := Toggle(ModeOriginal); got != ModeHigh {
		t.Fatalf("Toggle(original) = %q", got)
	}
	if got := Toggle(ModeHigh); got != ModeOriginal {
		t.Fatalf("Toggle(high) = %q", got)
	}
	if _, err := Parse("4k"); err == nil {
		t.Fatal("未知解析度應失敗")
	}
}

func TestSurfaceToBaseKeepsHighDPIAspectAndRejectsBars(t *testing.T) {
	if x, y, ok := SurfaceToBase(640, 360, HighWidth, HighHeight); !ok || x != 320 || y != 175 {
		t.Fatalf("高解析中心反算 = (%d,%d,%t)，預期 (320,175,true)", x, y, ok)
	}
	if _, _, ok := SurfaceToBase(640, 5, HighWidth, HighHeight); ok {
		t.Fatal("上下安全邊界不應命中")
	}
	if x, y, ok := SurfaceToBase(320, 175, OriginalWidth, OriginalHeight); !ok || x != 320 || y != 175 {
		t.Fatalf("原版座標反算 = (%d,%d,%t)", x, y, ok)
	}
}
