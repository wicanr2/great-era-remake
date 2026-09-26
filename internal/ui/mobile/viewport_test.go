package mobile

import (
	"math"
	"testing"

	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
)

func TestViewportRoundTripAndRejectsLetterbox(t *testing.T) {
	v, err := NewViewport(1920, 1200, 1, Insets{})
	if err != nil {
		t.Fatal(err)
	}
	if math.Abs(v.Scale-1.5) > 1e-9 {
		t.Fatalf("scale=%.3f，預期 1.5", v.Scale)
	}
	x, y, ok := v.SurfaceToDesign(v.OriginX+640*v.Scale+0.25, v.OriginY+360*v.Scale+0.25)
	if !ok || x != 640 || y != 360 {
		t.Fatalf("中心點反算=(%d,%d,%v)", x, y, ok)
	}
	if _, _, ok := v.SurfaceToDesign(1, 1); ok {
		t.Fatal("畫布外點不應命中")
	}
	r, ok := v.DesignToSurface(uilayout.Rect{X: 100, Y: 200, W: 48, H: 52})
	if !ok || r.W != 72 || r.H != 78 {
		t.Fatalf("設計矩形轉換=%+v", r)
	}
}

func TestViewportHonorsInsetsAnd48DPTarget(t *testing.T) {
	v, err := NewViewport(1920, 1080, 3, Insets{Top: 30, Bottom: 30, Left: 40, Right: 40})
	if err != nil {
		t.Fatal(err)
	}
	if v.OriginX <= 40 || v.OriginY < 30 {
		t.Fatalf("viewport 沒有保留安全區：origin=(%.1f,%.1f)", v.OriginX, v.OriginY)
	}
	r, ok := v.TouchTarget(uilayout.Rect{X: 600, Y: 300, W: 48, H: 48})
	if !ok {
		t.Fatal("48 dp touch target 不應失敗")
	}
	physical, ok := v.DesignToSurface(r)
	if !ok || physical.W < 144 || physical.H < 144 {
		t.Fatalf("觸控矩形=%+v，未達 48 dp：%+v", r, physical)
	}
}

func TestViewportRejectsInvalidDensityAndInsets(t *testing.T) {
	for _, tc := range []struct {
		name    string
		density float64
		insets  Insets
	}{
		{"zero density", 0, Insets{}},
		{"nan density", math.NaN(), Insets{}},
		{"negative inset", 2, Insets{Left: -1}},
		{"full width inset", 2, Insets{Left: 960, Right: 960}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := NewViewport(1920, 1080, tc.density, tc.insets); err == nil {
				t.Fatal("非法 viewport 應拒絕")
			}
		})
	}
}
