package game

import "testing"

// 倍數：高難度 4、低難度 3——高難度更嚴格。
func TestAIAdvantageRatio(t *testing.T) {
	if AIAdvantageRatio(true) != 4 || AIAdvantageRatio(false) != 3 {
		t.Errorf("倍數 %d／%d，原版是 4／3",
			AIAdvantageRatio(true), AIAdvantageRatio(false))
	}
	if AIAdvantageRatio(true) <= AIAdvantageRatio(false) {
		t.Error("高難度的倍數應該比低難度大（更保守）")
	}
}

// 門檻的邊界：theirs × 倍數 ≤ mine 才算有優勢。
func TestAIHasAdvantage(t *testing.T) {
	// 造兩個省各一個單位，用兵力控制戰力比。
	build := func(t *testing.T, myForce, theirForce uint16) *AIWorld {
		t.Helper()
		w := realWorld(t)
		w.Units = []CombatUnit{
			{General: 1, Province: 1, Faction: 7, Active: true},
			{General: 2, Province: 2, Faction: 8, Active: true},
		}
		mk := func(f uint16) StrengthInput {
			return StrengthInput{Ability: 60, Force: f, Branch: BranchInfantry,
				F19: 60, F20: 60, F29: 60, F30: 60}
		}
		w.Strengths = []StrengthInput{mk(myForce), mk(theirForce)}
		return w
	}

	w := build(t, 20000, 2000)
	mine := ProvinceStrength(1, w.Units, w.Strengths, w.Opts)
	theirs := ProvinceStrength(2, w.Units, w.Strengths, w.Opts)
	if theirs == 0 {
		t.Fatal("測試前提不成立：對方戰力為 0")
	}

	// 依實際算出來的戰力驗證兩個方向。
	t.Run("低難度 3 倍", func(t *testing.T) {
		want := theirs*3 <= mine
		if got := w.AIHasAdvantage(1, 2, false); got != want {
			t.Errorf("mine=%d theirs=%d：得到 %v，預期 %v", mine, theirs, got, want)
		}
	})
	t.Run("高難度 4 倍", func(t *testing.T) {
		want := theirs*4 <= mine
		if got := w.AIHasAdvantage(1, 2, true); got != want {
			t.Errorf("mine=%d theirs=%d：得到 %v，預期 %v", mine, theirs, got, want)
		}
	})
}

// 高難度成立時，低難度必然也成立（4 倍 ⊂ 3 倍）。
func TestAIAdvantageHardImpliesEasy(t *testing.T) {
	w := realWorld(t)
	mk := func(f uint16) StrengthInput {
		return StrengthInput{Ability: 60, Force: f, Branch: BranchInfantry,
			F19: 60, F20: 60, F29: 60, F30: 60}
	}
	for _, theirForce := range []uint16{500, 1000, 2000, 5000, 10000, 20000} {
		w.Units = []CombatUnit{
			{General: 1, Province: 1, Faction: 7, Active: true},
			{General: 2, Province: 2, Faction: 8, Active: true},
		}
		w.Strengths = []StrengthInput{mk(20000), mk(theirForce)}
		if w.AIHasAdvantage(1, 2, true) && !w.AIHasAdvantage(1, 2, false) {
			t.Errorf("對方兵力 %d：高難度成立但低難度不成立，倍數關係錯了", theirForce)
		}
	}
}

// 對方沒有可行動的單位時，戰力 0，一定有優勢。
func TestAIAdvantageAgainstEmpty(t *testing.T) {
	w := realWorld(t)
	w.Units = []CombatUnit{{General: 1, Province: 1, Faction: 7, Active: true}}
	w.Strengths = []StrengthInput{{Ability: 60, Force: 20000,
		Branch: BranchInfantry, F19: 60, F20: 60, F29: 60, F30: 60}}
	if !w.AIHasAdvantage(1, 2, true) {
		t.Error("對空省應該永遠有優勢")
	}
}
