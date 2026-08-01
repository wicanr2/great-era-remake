package game

import "testing"

func TestCityCellsFindsCities(t *testing.T) {
	m := loadTestMap(t)
	bf, err := m.Battlefield(19) // 河南
	if err != nil {
		t.Fatal(err)
	}
	cities := CityCells(bf)
	if len(cities) == 0 {
		t.Fatal("河南的戰場該有城市格")
	}
	if len(cities) > UnitsPerSide {
		t.Errorf("原版最多收 10 個，實際 %d", len(cities))
	}
	// 每一格都要真的是城市。
	for _, c := range cities {
		col, row := c.ColRow()
		if bf.Tiles[row][col].Kind != 5 {
			t.Errorf("格 %d 不是城市（Kind=%d）", c, bf.Tiles[row][col].Kind)
		}
	}
}

func TestRingCellsCountsMatchOriginal(t *testing.T) {
	// §12：原版 mode 1 收三圈共 37 格（1 + 6 + 12 + 18）。
	mid, err := CellAt(6, 6) // 挑中央，六個方向都不越界
	if err != nil {
		t.Fatal(err)
	}
	cases := []struct{ rings, want int }{{0, 1}, {1, 7}, {2, 19}, {3, 37}}
	for _, c := range cases {
		got := RingCells(mid, c.rings)
		if len(got) != c.want {
			t.Errorf("%d 圈該有 %d 格，實際 %d", c.rings, c.want, len(got))
		}
	}
	// 第一格永遠是中心自己（原版位置表的第 1 格給首位單位）。
	if r := RingCells(mid, 3); r[0] != mid {
		t.Errorf("第一格該是中心 %d，實際 %d", mid, r[0])
	}
	if RingCells(NoCell, 3) != nil {
		t.Error("無效的中心該回 nil")
	}
}

func TestExecuteActionReportsUnimplemented(t *testing.T) {
	// ⚠️ 沒實作的行動不得靜默跳過——那會讓實作缺口看起來像 AI 的決策結果。
	sim := mkTracedBattle(t, 20000, 18000)
	noRoute := func(to, from CellIndex) CellIndex { return NoCell }
	for _, a := range []BattleAction{ActADecisive, ActAReset, ActADefault,
		ActADecapitateKeepOne, ActAWeakest, ActBStrikeForce} {
		got := sim.ExecuteAction(a, sim.Defender, sim.Attacker, noRoute)
		if got.Implemented {
			t.Errorf("行動 %s 還沒實作，不該回 Implemented=true", BattleActionName(a))
		}
		if got.Note == "" {
			t.Errorf("行動 %s 未實作時要有說明", BattleActionName(a))
		}
	}
}

func TestExecTakeCityAssignsTargets(t *testing.T) {
	sim := mkTracedBattle(t, 20000, 18000)
	// 攻方單位要有命令 3 或 4 才會被派（§23）。
	for _, u := range sim.Attacker {
		u.Command = BattleCmdSeekTarget
	}
	// 把一個守方單位擺到城市上，讓那個城市有守軍。
	cities := CityCells(sim.Field)
	if len(cities) == 0 {
		t.Skip("這張戰場沒有城市")
	}
	d := sim.Defender[0]
	sim.Occ[d.Cell] = 0
	d.Cell = cities[0]
	sim.Occ[cities[0]] = d.General

	route := func(to, from CellIndex) CellIndex { return to } // 一步到位
	got := sim.ExecuteAction(ActBTakeCity, sim.Attacker, sim.Defender, route)
	if !got.Implemented {
		t.Fatal("打城市已經實作了")
	}
	if got.Assigned == 0 {
		t.Fatalf("該指派到目標，實際 0（%s）", got.Note)
	}
	// §23：目標單位是城市上的守軍，而且三個欄位一起設。
	a := sim.Attacker[0]
	if a.TargetUnit != d.General {
		t.Errorf("目標該是城市上的守軍 %d，實際 %d", d.General, a.TargetUnit)
	}
	if !a.Assigned() {
		t.Error("+13 bit 7 該被立起來（§31 那三個欄位一起設）")
	}
}

func TestExecEngageAllTargetsLivingFoes(t *testing.T) {
	sim := mkTracedBattle(t, 20000, 18000)
	route := func(to, from CellIndex) CellIndex { return to }
	got := sim.ExecuteAction(ActAEngageAll, sim.Defender, sim.Attacker, route)
	if !got.Implemented || got.Assigned == 0 {
		t.Fatalf("全面接戰該指派到目標：%+v", got)
	}
	d := sim.Defender[0]
	if d.Command != BattleCmdSeekTarget {
		t.Errorf("值 19 把命令設成 3，實際 %d", d.Command)
	}
	// 對方全部陣亡時沒有目標池。
	for _, a := range sim.Attacker {
		a.Strength.Force = 0
	}
	got = sim.ExecuteAction(ActAEngageAll, sim.Defender, sim.Attacker, route)
	if got.Assigned != 0 {
		t.Errorf("對方都陣亡了不該有指派，實際 %d", got.Assigned)
	}
}

func TestExecDeployFillsRing(t *testing.T) {
	sim := mkTracedBattle(t, 20000, 18000)
	// 多擺兩個守方單位，才看得出「依序發格子」。
	m := loadTestMap(t)
	bf, err := m.Battlefield(19)
	if err != nil {
		t.Fatal(err)
	}
	var free []CellIndex
	for i := 0; i < CellCount && len(free) < 2; i++ {
		c := CellIndex(i)
		col, row := c.ColRow()
		if sim.Occ[c] == 0 && bf.Owner[row][col] == 0 &&
			bf.Tiles[row][col].MoveCost() < 255 {
			free = append(free, c)
		}
	}
	for i, c := range free {
		u := mkUnit(GeneralID(300+i), 166, Branch1, 9000)
		u.Cell = c
		sim.Occ[c] = u.General
		sim.Defender = append(sim.Defender, u)
	}

	got := sim.ExecuteAction(ActBDeploy, sim.Defender, sim.Attacker, nil)
	if !got.Implemented {
		t.Fatal("佈防已經實作了")
	}
	if got.Assigned == 0 {
		t.Fatalf("該發出位置，實際 0（%s）", got.Note)
	}
	// §30：首位單位不參與（它分到位置表第一格 = 原地）。
	lead := sim.Defender[0]
	if lead.NextCell != NoCell {
		t.Errorf("首位單位不該被發位置，實際 %d", lead.NextCell)
	}
	// 發出去的格子不得重複。
	seen := map[CellIndex]bool{}
	for _, u := range sim.Defender {
		if u.NextCell == NoCell {
			continue
		}
		if seen[u.NextCell] {
			t.Errorf("格 %d 被發給兩個單位", u.NextCell)
		}
		seen[u.NextCell] = true
	}
	// 已經有去處的不動。
	sim.Defender[1].NextCell = 42
	before := sim.Defender[1].NextCell
	sim.ExecuteAction(ActBDeploy, sim.Defender, sim.Attacker, nil)
	if sim.Defender[1].NextCell != before {
		t.Error("已經有下一跳的單位不該被重新指派")
	}
}
