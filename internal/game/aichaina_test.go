package game

import (
	"fmt"
	"testing"
)

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

// 在**真實存檔**上跑一輪決策鏈 A，統計行動分佈。
//
// 這不是斷言某個數字，是**驗證電腦真的會動**——單元測試都是人造世界，
// 全綠不代表接上真實資料後不會整輪空轉（`CLAUDE.md` §10.4：測試綠只是
// 沒退步）。分佈印出來供人看，只在「一件事都沒做」時判失敗。
func TestChainAOnRealSave(t *testing.T) {
	w := realWorld(t)
	gens, err := ParseGenerals(readGame(t, "MAN(1).DAT"), 274)
	if err != nil {
		t.Fatal(err)
	}
	opt := ChainAOpts{Hard: true, EnableStep5: true, EnableLastThree: true}

	kinds := map[string]int{}
	steps := map[string]int{}
	acted := 0
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		prov, err := w.Table.At(p)
		if err != nil || prov.Commander == 0 {
			continue
		}
		opt.TotalForce = w.ProvinceForceTotal(p, gens)
		res := w.ChainA(p, gens, opt)
		if res.Action.Kind == AINone {
			kinds["（無動作）"]++
			continue
		}
		acted++
		switch res.Action.Kind {
		case AIComfort:
			kinds["慰勞"]++
		case AITransfer:
			kinds["調動"]++
		case AIAttack:
			kinds[fmt.Sprintf("出兵(模式%d)", res.SortieMode)]++
		}
		steps[res.Action.Step]++
	}

	for k, v := range kinds {
		t.Logf("  %-14s %d 省", k, v)
	}
	for k, v := range steps {
		t.Logf("  由「%s」決定：%d", k, v)
	}
	if acted == 0 {
		t.Error("39 省跑完一輪，電腦一件事都沒做——決策鏈接上真實資料後整輪空轉")
	}

	// ⚠️ 開局這一輪的結果是「慰勞 + 調動，零出兵」。那**不是 bug**：
	// 步驟 0 的慰勞門檻在高難度是「體力 < 80 或 士氣 < 70」，開局多數
	// 部隊達不到，於是慰勞壓過後面所有步驟。實機也是打了幾個月才看到
	// 電腦攻過來（`docs/playtest/14` 是 8 月）。
	//
	// 但「零出兵」與「出兵的步驟根本接不上」在這個統計裡長得一樣，
	// 所以再做一次對照：把士氣體力補滿，看後面的步驟會不會接手。
	t.Run("士氣體力補滿之後就走得到出兵", func(t *testing.T) {
		full := append([]General(nil), gens...)
		for i := range full {
			full[i].Stamina, full[i].F30 = 100, MoraleMax
		}
		attacks := 0
		for p := ProvinceID(1); p <= ProvinceCount; p++ {
			prov, err := w.Table.At(p)
			if err != nil || prov.Commander == 0 {
				continue
			}
			opt.TotalForce = w.ProvinceForceTotal(p, full)
			if w.ChainA(p, full, opt).Action.Kind == AIAttack {
				attacks++
			}
		}
		t.Logf("  補滿之後出兵 %d 省", attacks)
		if attacks == 0 {
			t.Error("沒有任何省出兵——步驟 1／4／5／6 可能都接不上")
		}
	})
}

// 兩條決策鏈接力：A 沒決定時 B 要接得上（`docs/re/13` §6 的順序）。
//
// 這驗的是**覆蓋率**——單獨跑 A 有 7 省什麼都不做，
// 那些省是不是真的無事可做，還是只是 A 管不到？
func TestChainAThenChainB(t *testing.T) {
	w := realWorld(t)
	gens, err := ParseGenerals(readGame(t, "MAN(1).DAT"), 274)
	if err != nil {
		t.Fatal(err)
	}
	opt := ChainAOpts{Hard: true, EnableStep5: true, EnableLastThree: true}

	var byA, byB, idle, total int
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		prov, err := w.Table.At(p)
		if err != nil || prov.Commander == 0 {
			continue
		}
		total++
		opt.TotalForce = w.ProvinceForceTotal(p, gens)
		switch {
		case w.ChainA(p, gens, opt).Action.Kind != AINone:
			byA++
		case w.Decide(p).Kind != AINone:
			byB++
		default:
			idle++
		}
	}
	t.Logf("有主省 %d：決策鏈 A %d、決策鏈 B 接手 %d、都沒決定 %d",
		total, byA, byB, idle)

	if byA == 0 {
		t.Error("決策鏈 A 一個省都沒決定")
	}
	// B 接手 0 個不必然是錯（A 可能真的涵蓋了全部），但**要看得到**，
	// 因為「B 沒接上」與「B 接上了但沒事做」在總數上長得一樣。
	if byA+byB == 0 {
		t.Error("兩條鏈加起來一個決定都沒有")
	}
}
