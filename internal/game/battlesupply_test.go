package game

import "testing"

func TestSupplyPerTroopEdges(t *testing.T) {
	// §48 的兩個特例。
	if got := SupplyPerTroop(0, 60000); got != 0 {
		t.Errorf("資源 0 該回 0，實得 %d", got)
	}
	if got := SupplyPerTroop(5000, 0); got != AIBattleSupplyCap {
		t.Errorf("兵力 0 該回 %d，實得 %d", AIBattleSupplyCap, got)
	}
	// 兵力不到 150 時分母被夾成 1 → 直接回資源本身。
	if got := SupplyPerTroop(37, 100); got != 37 {
		t.Errorf("分母夾 1 時該回資源本身，實得 %d", got)
	}
}

func TestSupplyPerTroopMatchesFormula(t *testing.T) {
	// 正對照：三個師 60,000 兵 → 分母 Round(60000/150) = 400。
	const troops = 60000
	cases := []struct {
		res, want int
	}{
		{60000, 150}, // 60000/400
		{18050, 45},  // 實機湖北的糧食
		{4200, 10},   // 實機湖北的黃金：4200/400 = 10.5 → 成雙取 10
		{4000, 10},
	}
	for _, c := range cases {
		if got := SupplyPerTroop(c.res, troops); got != c.want {
			t.Errorf("SupplyPerTroop(%d, %d) = %d，預期 %d", c.res, troops, got, c.want)
		}
	}
}

func TestRatioGateIsTheCommonCase(t *testing.T) {
	// ⭐ §48 的結論：拿實機湖北的存量跟三個滿員師比，
	// 「糧食夠、黃金不夠」成立——那三個行動是主線不是例外。
	s := BattleSupply{Gold: 4200, Food: 18050, Troops: 60000}
	if !s.RatioGate(0) {
		t.Error("湖北的糧食 18050／黃金 4200 對 60000 兵，比率門檻該成立")
	}
	// 黃金拉到跟糧食一樣多 → 黃金不吃緊 → 不成立。
	s.Gold = 18050
	if s.RatioGate(0) {
		t.Error("黃金充足時比率門檻不該成立")
	}
	// 糧食也吃緊 → 前半的 `!糧食吃緊` 破掉 → 不成立。
	s = BattleSupply{Gold: 100, Food: 100, Troops: 60000}
	if s.RatioGate(0) {
		t.Error("兩項都吃緊時比率門檻不該成立")
	}
}

func TestTroopTotalSumsForce(t *testing.T) {
	us := []*Combatant{
		{Strength: StrengthInput{Force: 20000}},
		nil,
		{Strength: StrengthInput{Force: 10000}},
	}
	if got := TroopTotal(us); got != 30000 {
		t.Errorf("TroopTotal = %d，預期 30000", got)
	}
}

func TestRoundDivIsBankers(t *testing.T) {
	// Turbo Pascal 的 Round 是四捨六入五成雙。
	cases := [][3]int{
		{5, 2, 2},  // 2.5 → 2（偶數）
		{7, 2, 4},  // 3.5 → 4（偶數）
		{6, 4, 2},  // 1.5 → 2
		{10, 4, 2}, // 2.5 → 2
		{7, 3, 2},  // 2.33 → 2
		{8, 3, 3},  // 2.67 → 3
	}
	for _, c := range cases {
		if got := roundDiv(c[0], c[1]); got != c[2] {
			t.Errorf("roundDiv(%d, %d) = %d，預期 %d", c[0], c[1], got, c[2])
		}
	}
}
