package hdart

import "testing"

func TestEarthArtMemoryAndAlphaRemainBounded(t *testing.T) {
	if n := Bytes() + EarthBytes(); n <= 0 || n > 2*1024*1024 {
		t.Fatal("material memory", n)
	}
	for i := 0; i < 18; i++ {
		im := EarthUnit(i, 64, 64)
		if im == nil {
			t.Fatal("missing original unit index", i)
		}
		for x := 0; x < 64; x++ {
			if im.RGBAAt(x, 0).A != 0 || im.RGBAAt(x, 63).A != 0 || im.RGBAAt(0, x).A != 0 || im.RGBAAt(63, x).A != 0 {
				t.Fatal("clipped unit", i)
			}
		}
	}
	if EarthUnit(18, 64, 64) != nil || EarthUnit(0, 8192, 8192) != nil {
		t.Fatal("unbounded or invalid index accepted")
	}
}
