package game

import "testing"

// 兩檔門檻的機率：佔上風 70%、劣勢 20%。
func TestCeasefireRates(t *testing.T) {
	cases := []struct {
		name string
		min  int
		want float64
	}{
		{"佔上風（roll ≥ 3）", CeasefireStrongMin, 0.7},
		{"居劣勢（roll ≥ 8）", CeasefireWeakMin, 0.2},
	}
	for _, c := range cases {
		rng := NewRand(2024)
		const n = 20000
		ok := 0
		for i := 0; i < n; i++ {
			if rng.Int(CeasefireRollRange) >= c.min {
				ok++
			}
		}
		rate := float64(ok) / n
		if rate < c.want-0.03 || rate > c.want+0.03 {
			t.Errorf("%s 的同意率 %.3f，預期 %.1f", c.name, rate, c.want)
		}
	}
}

// 佔上風的判定：守方請求時比守方戰力，攻方請求時比攻方戰力。
func TestCeasefireStrongerSide(t *testing.T) {
	w := realWorld(t)
	prov, err := w.Table.At(1)
	if err != nil {
		t.Fatal(err)
	}
	defender := GeneralID(7)
	attacker := GeneralID(8)
	prov.Commander = defender

	// 造兩個在場單位：守方強、攻方弱。
	w.Units = []CombatUnit{
		{General: 1, Province: 1, Cell: 0, Faction: defender, Active: true},
		{General: 2, Province: 1, Cell: 1, Faction: attacker, Active: true},
	}
	w.Strengths = []StrengthInput{
		{Ability: 90, Force: 20000, Branch: BranchInfantry, F19: 60, F20: 60, F29: 80, F30: 80},
		{Ability: 30, Force: 2000, Branch: BranchInfantry, F19: 30, F20: 30, F29: 40, F30: 40},
	}

	atk, def := w.BattleForces(1)
	if def <= atk {
		t.Fatalf("測試前提不成立：守方 %d 沒有比攻方 %d 強", def, atk)
	}

	// 守方（省司令）來談 → 佔上風。
	res, err := w.NegotiateCeasefire(1, defender, NewRand(1))
	if err != nil {
		t.Fatal(err)
	}
	if !res.RequesterStronger {
		t.Error("守方較強時，守方請求應該算佔上風")
	}

	// 攻方來談 → 劣勢。
	res, err = w.NegotiateCeasefire(1, attacker, NewRand(1))
	if err != nil {
		t.Fatal(err)
	}
	if res.RequesterStronger {
		t.Error("守方較強時，攻方請求應該算劣勢")
	}
}

// 同一顆種子下，佔上風的一方成功率必須不低於劣勢方。
func TestCeasefireStrongerIsEasier(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	defender, attacker := GeneralID(7), GeneralID(8)
	prov.Commander = defender
	w.Units = []CombatUnit{
		{General: 1, Province: 1, Cell: 0, Faction: defender, Active: true},
		{General: 2, Province: 1, Cell: 1, Faction: attacker, Active: true},
	}
	w.Strengths = []StrengthInput{
		{Ability: 90, Force: 20000, Branch: BranchInfantry, F19: 60, F20: 60, F29: 80, F30: 80},
		{Ability: 30, Force: 2000, Branch: BranchInfantry, F19: 30, F20: 30, F29: 40, F30: 40},
	}

	strongOK, weakOK := 0, 0
	const n = 300
	for seed := uint32(1); seed <= n; seed++ {
		if r, _ := w.NegotiateCeasefire(1, defender, NewRand(seed)); r.Agreed {
			strongOK++
		}
		if r, _ := w.NegotiateCeasefire(1, attacker, NewRand(seed)); r.Agreed {
			weakOK++
		}
	}
	if strongOK <= weakOK {
		t.Errorf("佔上風談成 %d 次、劣勢 %d 次——強勢方應該比較容易談成",
			strongOK, weakOK)
	}
}

// 不在場的單位（Cell 無效）不計入戰力。
func TestBattleForcesIgnoresOffField(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	prov.Commander = 7
	w.Units = []CombatUnit{
		{General: 1, Province: 1, Cell: NoCell, Faction: 7, Active: true},
		{General: 2, Province: 1, Cell: NoCell, Faction: 8, Active: true},
	}
	w.Strengths = []StrengthInput{
		{Ability: 90, Force: 20000, Branch: BranchInfantry, F19: 60, F20: 60, F29: 80, F30: 80},
		{Ability: 90, Force: 20000, Branch: BranchInfantry, F19: 60, F20: 60, F29: 80, F30: 80},
	}
	if atk, def := w.BattleForces(1); atk != 0 || def != 0 {
		t.Errorf("不在場的單位被計入了：攻 %d／守 %d", atk, def)
	}
}

// 無主省談不了停火。
func TestCeasefireNeedsCommander(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	prov.Commander = 0
	if _, err := w.NegotiateCeasefire(1, 7, NewRand(1)); err == nil {
		t.Error("無主省應該報錯")
	}
}
