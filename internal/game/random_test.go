package game

import "testing"

// LCG 的遞推式：seed = seed × 134775813 + 1（32-bit 環繞）。
// 這是整個亂數系統的地基，常數錯一位後面全錯。
func TestLCGRecurrence(t *testing.T) {
	r := NewRand(0)
	s := uint32(0)
	for i := 0; i < 5; i++ {
		s = s*0x08088405 + 1
		if got := r.next(); got != s {
			t.Fatalf("第 %d 步：%#x，應為 %#x", i+1, got, s)
		}
	}
	if NewRand(0).next() != 1 {
		t.Error("種子 0 的第一步應該是 1（0 × K + 1）")
	}
}

// 同樣的種子必須產生同樣的序列——這是「可重現」的定義。
func TestRandReproducible(t *testing.T) {
	a, b := NewRand(12345), NewRand(12345)
	for i := 0; i < 100; i++ {
		if x, y := a.Word(1000), b.Word(1000); x != y {
			t.Fatalf("第 %d 個值不同：%d vs %d", i, x, y)
		}
	}
	c, d := NewRand(12346), NewRand(12345)
	same := 0
	for i := 0; i < 100; i++ {
		if c.Word(1000) == d.Word(1000) {
			same++
		}
	}
	if same > 20 {
		t.Errorf("兩個不同種子的序列有 %d/100 個值相同，太像了", same)
	}
}

// Random(n) 用「乘上 n 取高 32 位」而不是取餘數——兩者分佈不同。
func TestRandWordRange(t *testing.T) {
	r := NewRand(999)
	for i := 0; i < 500; i++ {
		if v := r.Word(6); v >= 6 {
			t.Fatalf("Random(6) 回了 %d，超出範圍", v)
		}
	}
	r2 := NewRand(42)
	for i := 0; i < 20; i++ {
		if v := r2.Word(1); v != 0 {
			t.Fatalf("Random(1) 回了 %d，應為 0", v)
		}
	}
	if v := NewRand(1).Word(0); v != 0 {
		t.Errorf("Random(0) 回了 %d", v)
	}
}

// 六面骰的分佈應該大致均勻——LCG 常數錯的話這裡會明顯偏。
func TestRandDistribution(t *testing.T) {
	r := NewRand(2026)
	const n, rolls = 6, 60000
	buckets := make([]int, n)
	for i := 0; i < rolls; i++ {
		buckets[r.Word(n)]++
	}
	exp := rolls / n
	for i, c := range buckets {
		if d := c - exp; d < -exp/10 || d > exp/10 {
			t.Errorf("第 %d 面出現 %d 次，期望 %d ±10%%", i, c, exp)
		}
	}
	t.Logf("六面骰 %d 次：%v", rolls, buckets)
}

func TestRandReal(t *testing.T) {
	r := NewRand(7)
	for i := 0; i < 1000; i++ {
		if v := r.Real(); v < 0 || v >= 1 {
			t.Fatalf("Real() 回了 %f", v)
		}
	}
}
