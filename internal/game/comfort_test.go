package game

import "testing"

// 慰勞軍民對單一將領的效果（`docs/re/22`）：
// 體力 +10 夾 100、士氣 <80 時 += 帶兵能力÷5 夾 80、忠誠度 +20 夾 100。
func TestComfortGeneral(t *testing.T) {
	cases := []struct {
		stamina, morale, loyalty, ability uint8
		wantS, wantM, wantL               uint8
	}{
		// 吳佩孚：體力 55、士氣 74、帶兵 90 → 74 + 18 = 92 → 夾 80
		{55, 74, 100, 90, 65, 80, 100},
		// 士氣已達上限就完全不加
		{50, 80, 50, 90, 60, 80, 70},
		{50, 85, 50, 90, 60, 85, 70}, // 超過上限也不動它
		// 體力夾到 100
		{95, 40, 40, 50, 100, 50, 60},
		// 帶兵能力低 → 士氣加得少
		{10, 30, 10, 9, 20, 31, 30},
	}
	for _, c := range cases {
		s, m, l := ComfortGeneral(c.stamina, c.morale, c.loyalty, c.ability)
		if s != c.wantS || m != c.wantM || l != c.wantL {
			t.Errorf("體力 %d／士氣 %d／忠誠 %d／帶兵 %d → (%d,%d,%d)，應為 (%d,%d,%d)",
				c.stamina, c.morale, c.loyalty, c.ability,
				s, m, l, c.wantS, c.wantM, c.wantL)
		}
	}
}

// 士氣上限是 80 不是 100——實機五個將領全部低於 80（`docs/playtest/08`）。
func TestMoraleCapIsEighty(t *testing.T) {
	if MoraleMax != 80 {
		t.Fatalf("士氣上限是 %d，原版是 80", MoraleMax)
	}
	// 實機那五個值
	for _, m := range []uint8{74, 67, 42, 54, 30} {
		if m >= MoraleMax {
			t.Errorf("實機士氣 %d 不該達到或超過上限 %d", m, MoraleMax)
		}
	}
	// 從 79 開始慰勞，帶兵能力再高也只到 80
	if _, m, _ := ComfortGeneral(50, 79, 50, 100); m != MoraleMax {
		t.Errorf("士氣 79 + 帶兵 100 → %d，應該夾到 %d", m, MoraleMax)
	}
}

// 省的人民忠誠度 += 領袖政治手腕 ÷ 10，夾到 100。
func TestComfortProvinceLoyalty(t *testing.T) {
	cases := []struct {
		before, politics uint8
		wantGain         int
	}{
		{41, 94, 9}, // 吳佩孚政治手腕 94 → 9
		{41, 41, 4},
		{41, 9, 0},   // 不足 10 → 白跑
		{95, 94, 5},  // 夾到 100
		{100, 94, 0}, // 已滿
	}
	for _, c := range cases {
		w := realWorld(t)
		prov, _ := w.Table.At(19)
		prov.Loyalty = c.before
		res, err := w.Comfort(19, c.politics)
		if err != nil {
			t.Fatal(err)
		}
		if res.ProvinceLoyaltyGain != c.wantGain {
			t.Errorf("忠誠 %d／政治 %d → 增 %d，應為 %d",
				c.before, c.politics, res.ProvinceLoyaltyGain, c.wantGain)
		}
	}
}
