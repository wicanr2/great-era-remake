package game

import "testing"

// 造一場多單位的戰鬥。**單一單位不夠**——原版的預備隊投入
// （`ActivateReserves`，§17）不碰 index 1，所以 1v1 的場面裡
// 值 3 什麼都不會做，整場空轉。這是原版行為的直接後果，不是 bug。
//
// ⚠️ **兵力對比也要設計**。預備隊投入與主將親征都看「首位單位戰力
// vs 敵軍總戰力」的優勢等級（§17）——攻方劣勢時原版**正確地什麼都不做**，
// 那是「打不過就不打」，不是實作缺口。所以這裡讓攻方壓倒性優勢。
func mkMultiBattle(t *testing.T, nAtk, nDef int, defForce uint16) *BattleSim {
	t.Helper()
	m := loadTestMap(t)
	const wu, zhang = GeneralID(58), GeneralID(166)

	var attackers []*Combatant
	for i := 0; i < nAtk; i++ {
		attackers = append(attackers, mkUnit(GeneralID(101+i), wu, Branch1, 20000))
	}

	bf, err := m.Battlefield(19)
	if err != nil {
		t.Fatal(err)
	}
	// ⚠️ **守方要站在城市上**。值 3（打城市，分支 B 的預設行動）只打
	// 「敵方佔據的城市」（§21）——守方不在城市上，值 3 就正確地什麼都不做，
	// 整場空轉。那是原版行為，不是實作缺口。
	// 測試場面必須讓行動的前提成立，否則驗的是空氣。
	cells := CityCells(bf)
	for i := 0; i < CellCount && len(cells) < nDef; i++ {
		c := CellIndex(i)
		col, row := c.ColRow()
		if bf.Owner[row][col] == 0 && bf.Tiles[row][col].MoveCost() < 255 {
			cells = append(cells, c)
		}
	}
	if len(cells) < nDef {
		t.Fatalf("找不到 %d 個可站的守方格", nDef)
	}
	var defenders []*Combatant
	for i := 0; i < nDef; i++ {
		u := mkUnit(GeneralID(201+i), zhang, Branch1, defForce)
		u.Cell = cells[i]
		defenders = append(defenders, u)
	}

	sim, err := NewBattleSim(m, 19, 26, attackers, defenders, StrengthOpts{Stage: 1})
	if err != nil {
		t.Fatal(err)
	}
	return sim
}

func TestAutoResolveByChainRuns(t *testing.T) {
	// 守方 5000×3：攻方首位單位取得優勢（預備隊會投入），
	// 但沒到五倍差，所以不會被必勝結算秒判。
	sim := mkMultiBattle(t, 3, 3, 5000)
	st := sim.AutoResolveByChain(60, BattleChainGates{}, 201)

	if st.Turns == 0 {
		t.Fatal("該至少跑一回合")
	}
	if len(st.Decisions) != st.Turns {
		t.Errorf("每回合各一筆決策：%d 筆 vs %d 回合", len(st.Decisions), st.Turns)
	}
	// ⭐ 這場戰鬥有沒有因為執行層缺口而空轉。
	t.Logf("回合 %d／移動 %d／交戰 %d／未實作 %d／攻方損 %d 守方損 %d",
		st.Turns, st.Moves, st.Engagements, st.Unimplemented,
		st.AttackerLoss, st.DefenderLoss)

	// 決策鏈選出的行動要落在已知的 13 種裡。
	for _, d := range st.Decisions {
		if BattleActionName(d.A.Action) == "未知行動" {
			t.Errorf("回合 %d 分支 A 選了未知行動 %d", d.Turn, d.A.Action)
		}
		if BattleActionName(d.B.Action) == "未知行動" {
			t.Errorf("回合 %d 分支 B 選了未知行動 %d", d.Turn, d.B.Action)
		}
	}
}

func TestChainRunActuallyMovesAndFights(t *testing.T) {
	// 決策鏈驅動的推進器必須真的讓單位動起來、真的打起來——
	// 否則它只是一個很貴的空轉迴圈。
	sim := mkMultiBattle(t, 3, 3, 5000)
	st := sim.AutoResolveByChain(60, BattleChainGates{}, 201)

	// ⚠️ **不要求整場不觸發必勝結算**——打到後來一方被打殘，
	// 五倍門檻就會在戰鬥中途成立，那是原版的「打殘了就收工」（§16）。
	// 這裡只要求「真的動起來、真的打起來」。
	if st.Moves == 0 {
		t.Error("整場沒有任何移動——推進器沒接上命令")
	}
	if st.AttackerLoss == 0 && st.DefenderLoss == 0 && st.Engagements == 0 {
		t.Error("整場沒有任何交戰或損失")
	}
}

func TestDirToward(t *testing.T) {
	mid, err := CellAt(6, 6)
	if err != nil {
		t.Fatal(err)
	}
	for d := HexDir(1); d <= 6; d++ {
		n, ok := mid.Neighbour(d)
		if !ok {
			continue
		}
		if got := dirToward(mid, n); got != d {
			t.Errorf("往 %d 的方向該是 %d，實際 %d", n, d, got)
		}
	}
	// 不相鄰的格找不到方向。
	far, err := CellAt(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if got := dirToward(mid, far); got != 0 {
		t.Errorf("不相鄰該回 0，實際 %d", got)
	}
}

func TestStepByOrderClearsNextCell(t *testing.T) {
	sim := mkTracedBattle(t, 20000, 18000)
	u := sim.Attacker[0]
	ns := u.Cell.Neighbours()
	var free CellIndex = NoCell
	for _, n := range ns {
		col, row := n.ColRow()
		if sim.Occ[n] == 0 && sim.Field.Tiles[row][col].MoveCost() < 255 {
			free = n
			break
		}
	}
	if free == NoCell {
		t.Skip("周圍沒有可走的空格")
	}
	u.NextCell = free
	if !sim.StepByOrder(u) {
		t.Fatal("該走得動")
	}
	if u.Cell != free {
		t.Errorf("該走到 %d，實際在 %d", free, u.Cell)
	}
	// 下一跳是一次性的，走完要清掉（原版每回合重算，§14）。
	if u.NextCell != NoCell {
		t.Errorf("走完該清掉下一跳，實際 %d", u.NextCell)
	}
	// 沒有下一跳就不動。
	if sim.StepByOrder(u) {
		t.Error("沒有下一跳不該移動")
	}
}


func TestChainRunDecisiveResolution(t *testing.T) {
	// ⭐ §16：戰力差五倍時原版不再逐格打，直接判勝負。
	// 守方 2500×3 對攻方 20000×3，五倍門檻成立。
	sim := mkMultiBattle(t, 3, 3, 2500)
	st := sim.AutoResolveByChain(60, BattleChainGates{}, 201)

	if !st.Decisive {
		t.Fatalf("戰力差五倍該被必勝結算判掉，實際跑了 %d 回合", st.Turns)
	}
	if !st.Decided {
		t.Error("必勝結算該算分出勝負")
	}
	if !st.AttackerWon {
		t.Errorf("守方被輾壓該是攻方勝（%s）", st.DecisiveNote)
	}
	// 秒判：不該打滿很多回合。
	if st.Turns > 2 {
		t.Errorf("必勝結算該在頭一兩回合就觸發，實際 %d 回合", st.Turns)
	}
	// ⚠️ 這種結束方式沒有戰損——原版是「不打了直接判」，不是打光對方。
	t.Logf("必勝結算於第 %d 回合：%s", st.Turns, st.DecisiveNote)
}
