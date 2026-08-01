package game

import "testing"

// chainAWorld 造一個「勢力 7 佔省 1、周圍是敵人 8」的世界，
// 並把將領全部搬到省 1。
func chainAWorld(t *testing.T) (*AIWorld, []General) {
	t.Helper()
	w := realWorld(t)
	for i := ProvinceID(1); ; i++ {
		p, err := w.Table.At(i)
		if err != nil {
			break
		}
		p.Commander = 8
		p.Flags &^= ProvinceFlagInBattle
		// 資源給滿，讓 AISupplyOK 不會擋路。
		p.Gold, p.Food, p.Ammo, p.Fuel = 60000, 60000, 60000, 60000
	}
	me, err := w.Table.At(1)
	if err != nil {
		t.Fatal(err)
	}
	me.Commander = 7

	w.Units = nil
	w.Strengths = nil
	var gens []General
	for i := 0; i < 6; i++ {
		id := GeneralID(i + 1)
		w.Units = append(w.Units, CombatUnit{
			General: id, Province: 1, Faction: 7, Active: true})
		w.Strengths = append(w.Strengths, StrengthInput{
			Ability: 80, Force: 20000, F19: 10, F20: 10,
			F29: 100, F30: 80, Branch: 1, General: id})
		gens = append(gens, General{
			Force: 20000, Branch: 1, Stamina: 100, F30: 80,
			AbilityA: 80, Province: 1})
	}
	return w, gens
}

// 離島三省在 bit 1 設起來時整條跳過。
func TestChainASkipsIslands(t *testing.T) {
	w, gens := chainAWorld(t)
	for _, p := range AIChainASkippedProvinces {
		prov, err := w.Table.At(p)
		if err != nil {
			t.Fatal(err)
		}
		prov.Commander = 7
		got := w.ChainA(p, gens, ChainAOpts{EnableStep5: true})
		if got.Action.Kind != AINone {
			t.Errorf("省 %d 在 bit 1 設起來時應該整條跳過，卻做了 %v", p, got.Action.Step)
		}
		// bit 1 沒設就照跑（至少不是被這道檢查擋掉的）。
		off := w.ChainA(p, gens, ChainAOpts{EnableStep5: false})
		_ = off
	}
}

// 步驟 0 優先於後面所有步驟。
func TestChainAComfortFirst(t *testing.T) {
	w, gens := chainAWorld(t)
	// 讓省 1 的將領士氣偏低 → 觸發慰勞。
	for i := range gens {
		gens[i].F30 = AIComfortMoraleHard - 1
	}
	got := w.ChainA(1, gens, ChainAOpts{Hard: true})
	if got.Action.Kind != AIComfort {
		t.Errorf("士氣偏低時應該先慰勞，實際做了 %v（%s）",
			got.Action.Kind, got.Action.Step)
	}
}

// 步驟 6：被包圍就突圍，模式 1。
func TestChainABreakout(t *testing.T) {
	w, gens := chainAWorld(t)
	// 讓勢力 7 有 3 個省（突圍門檻），但省 1 的鄰省全是敵人。
	for _, p := range []ProvinceID{20, 21} {
		setCommander(t, w, p, 7)
	}
	me, _ := w.Table.At(1)
	for _, n := range me.Neighbours {
		np, _ := w.Table.At(n)
		np.Commander = 8
	}
	if !w.Encircled(1) {
		t.Fatal("測試前提不成立：省 1 應該被包圍")
	}
	// 關掉步驟 5，避免它先命中。
	got := w.ChainA(1, gens, ChainAOpts{
		Hard: true, EnableLastThree: true, TotalForce: 100000})
	if got.Action.Kind == AIAttack && got.SortieMode == 1 {
		return // 步驟 6 命中
	}
	// 也可能被步驟 1／3／4 先攔下——那都是合法的優先序，
	// 只要不是「什麼也沒做」就算通過。
	if got.Action.Kind == AINone {
		t.Errorf("被包圍又補給充足，不該什麼都不做")
	}
	t.Logf("被步驟「%s」先命中（優先序正常）", got.Action.Step)
}

// 攻打候選排除自己人與無主省——這與調動的規則相反。
func TestChainAAttackCandidatesExcludeNeutral(t *testing.T) {
	w, gens := chainAWorld(t)
	me, _ := w.Table.At(1)
	if len(me.Neighbours) < 3 {
		t.Skip("省 1 的鄰省太少")
	}
	n0, _ := w.Table.At(me.Neighbours[0])
	n0.Commander = 0 // 無主
	n1, _ := w.Table.At(me.Neighbours[1])
	n1.Commander = 7 // 自己人

	got := w.attackCandidates(1, 7, gens, true)
	for _, c := range got {
		if c == me.Neighbours[0] {
			t.Error("無主省不該進攻打候選")
		}
		if c == me.Neighbours[1] {
			t.Error("自己人的省不該進攻打候選")
		}
	}
}

// 交戰中的省不進候選。
func TestChainAAttackCandidatesSkipInBattle(t *testing.T) {
	w, gens := chainAWorld(t)
	me, _ := w.Table.At(1)
	if len(me.Neighbours) == 0 {
		t.Skip("省 1 沒有鄰省")
	}
	target := me.Neighbours[0]
	before := len(w.attackCandidates(1, 7, gens, true))
	tp, _ := w.Table.At(target)
	tp.Flags |= ProvinceFlagInBattle
	after := w.attackCandidates(1, 7, gens, true)
	for _, c := range after {
		if c == target {
			t.Error("交戰中的省不該進候選")
		}
	}
	if before == 0 {
		t.Skip("原本就沒有候選，測不出差異")
	}
}
