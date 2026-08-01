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
	// ⚠️ 這份清單是「執行層還缺什麼」的憑證。實作一個就從這裡移除一筆
	// ——測試會在你忘了移除時紅給你看（2026-08-02 值 12/13/14/15 實作後就發生過）。
	for _, a := range []BattleAction{
		ActADecisive, // 11／1 必勝結算：要接九步結算，不是指派命令
		ActAWeakest,  // 18 挑最弱（sub_3E24F 的統計段未讀完，先不做）
	} {
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

func TestExecDefaultBranchesByCommand(t *testing.T) {
	// §32：值 13 依單位**現有的命令**分流，保留狀態。
	sim := mkTracedBattle(t, 20000, 18000)
	cities := CityCells(sim.Field)
	if len(cities) == 0 {
		t.Skip("這張戰場沒有城市")
	}
	route := func(to, from CellIndex) CellIndex { return to }

	// 命令 1 且不在城市上 → 找空城。
	u := sim.Defender[0]
	u.Command = BattleCmdGarrison
	u.NextCell = NoCell
	sim.ExecuteAction(ActADefault, sim.Defender, sim.Attacker, route)
	if u.NextCell == NoCell {
		t.Error("命令 1 該找到城市當去處")
	}
	// ⚠️ §32：取的是**最後一個**空城，不是最近的。
	var lastEmpty CellIndex = NoCell
	for _, c := range cities {
		if sim.Occ[c] == 0 {
			lastEmpty = c
		}
	}
	if lastEmpty != NoCell && u.NextCell != lastEmpty {
		t.Errorf("原版取最後一個空城 %d，實際 %d", lastEmpty, u.NextCell)
	}

	// 命令 3 → 壓成 2 並清下一跳。
	u.Command = BattleCmdSeekTarget
	u.NextCell = 42
	sim.ExecuteAction(ActADefault, sim.Defender, sim.Attacker, route)
	if u.Command != BattleCmdStandby || u.NextCell != NoCell {
		t.Errorf("命令 3 該被壓成 2 並清下一跳，實際 cmd=%d next=%d", u.Command, u.NextCell)
	}

	// 命令 4 → 下一跳 = 原地。
	u.Command = BattleCmdCommitted
	sim.ExecuteAction(ActADefault, sim.Defender, sim.Attacker, route)
	if u.NextCell != u.Cell {
		t.Errorf("命令 4 該原地待命，實際 %d（自己在 %d）", u.NextCell, u.Cell)
	}
}

func TestExecDecapitateKeepOneVsKeepAll(t *testing.T) {
	// §34：15 比 14 保守——這是最容易讀反的一條。
	route := func(to, from CellIndex) CellIndex { return to }

	mk := func() *BattleSim {
		s := mkTracedBattle(t, 20000, 18000)
		// 造三個守方單位，全部是命令 1（駐守中）。
		m := loadTestMap(t)
		bf, _ := m.Battlefield(19)
		var free []CellIndex
		for i := 0; i < CellCount && len(free) < 2; i++ {
			c := CellIndex(i)
			col, row := c.ColRow()
			if s.Occ[c] == 0 && bf.Owner[row][col] == 0 &&
				bf.Tiles[row][col].MoveCost() < 255 {
				free = append(free, c)
			}
		}
		for i, c := range free {
			u := mkUnit(GeneralID(400+i), 166, Branch1, 9000)
			u.Cell, u.NextCell = c, NoCell
			s.Occ[c] = u.General
			s.Defender = append(s.Defender, u)
		}
		for _, u := range s.Defender {
			u.Command = BattleCmdGarrison
		}
		return s
	}

	// 值 14：只留第一個，其餘全部出擊。
	s14 := mk()
	got14 := s14.ExecuteAction(ActADecapitateKeepOne, s14.Defender, s14.Attacker, route)
	// 值 15：全部駐守的都留下 → 一個都不派。
	s15 := mk()
	got15 := s15.ExecuteAction(ActADecapitateKeepAll, s15.Defender, s15.Attacker, route)

	if got14.Assigned <= got15.Assigned {
		t.Errorf("值 14 該比值 15 派出更多（15 較保守）：14=%d 15=%d",
			got14.Assigned, got15.Assigned)
	}
	if got15.Assigned != 0 {
		t.Errorf("全部都是命令 1 時，值 15 一個都不該派，實際 %d", got15.Assigned)
	}
}

func TestExecResetClearsThenReassigns(t *testing.T) {
	// §33：值 12 先把單位打回待命，再重新找目標。
	sim := mkTracedBattle(t, 20000, 18000)
	route := func(to, from CellIndex) CellIndex { return to }
	a := sim.Attacker[0]
	a.Command = BattleCmdGarrison
	a.AssignTo(999, 5) // 先給一個假目標

	got := sim.ExecuteAction(ActAReset, sim.Attacker, sim.Defender, route)
	if !got.Implemented {
		t.Fatal("推倒重來已經實作了")
	}
	// 目標該被換成真的守方單位（不是原本那個假的 999）。
	if a.TargetUnit == 999 {
		t.Error("值 12 該先重置再重新指派，舊目標不該留著")
	}
	if got.Assigned > 0 && a.Command != BattleCmdStandby && !a.Assigned() {
		t.Error("重置後的單位狀態不一致")
	}
}

func TestExecRecomputeLoopOrder(t *testing.T) {
	// §14：三個迴圈，順序有意義。
	sim := mkTracedBattle(t, 20000, 18000)
	route := func(to, from CellIndex) CellIndex { return to }
	d := sim.Defender[0]

	// 迴圈 1：命令 3 且沒有目標格 → 降成 2。
	a := sim.Attacker[0]
	a.Command = BattleCmdSeekTarget
	a.NextCell = NoCell
	a.TargetUnit = 0
	sim.ExecuteAction(ActARecompute, sim.Attacker, sim.Defender, route)
	if a.Command != BattleCmdStandby {
		t.Errorf("命令 3 沒目標格該降成 2，實際 %d", a.Command)
	}

	// 有目標的會重算下一格。
	a.Command = BattleCmdCommitted
	a.TargetUnit = d.General
	a.NextCell = NoCell
	got := sim.ExecuteAction(ActARecompute, sim.Attacker, sim.Defender, route)
	if got.Assigned == 0 || a.NextCell != d.Cell {
		t.Errorf("該重算出下一格 %d，實際 %d（assigned=%d）", d.Cell, a.NextCell, got.Assigned)
	}

	// 目標已陣亡就算不出來。
	d.Strength.Force = 0
	a.NextCell = NoCell
	got = sim.ExecuteAction(ActARecompute, sim.Attacker, sim.Defender, route)
	if got.Assigned != 0 {
		t.Errorf("目標陣亡時不該算出下一格，實際 assigned=%d", got.Assigned)
	}
}

func TestExecStandbyOnlyFiltersByCommand(t *testing.T) {
	// §35：命令 2 無條件處理；命令 4／5 要後期；其餘跳過。
	sim := mkTracedBattle(t, 20000, 18000)
	route := func(to, from CellIndex) CellIndex { return to }

	a := sim.Attacker[0]
	// 命令 1 不在處理範圍。
	a.Command = BattleCmdGarrison
	a.NextCell = NoCell
	if got := sim.ExecuteAction(ActAStandbyOnly, sim.Attacker, sim.Defender, route); got.Assigned != 0 {
		t.Errorf("命令 1 不該被處理，實際 %d", got.Assigned)
	}
	// 命令 2 會。
	a.Command = BattleCmdStandby
	if got := sim.ExecuteAction(ActAStandbyOnly, sim.Attacker, sim.Defender, route); got.Assigned == 0 {
		t.Error("命令 2 該被處理")
	}
	// 前期時命令 4／5 被跳過（用內部的 stage 版本驗）。
	a.Command = BattleCmdUnknown5
	a.NextCell = NoCell
	early := sim.execStandbyOnlyStage(sim.Attacker, sim.Defender, route, false)
	if early.Assigned != 0 {
		t.Errorf("前期不該處理命令 5，實際 %d", early.Assigned)
	}
	late := sim.execStandbyOnlyStage(sim.Attacker, sim.Defender, route, true)
	if late.Assigned == 0 {
		t.Error("後期該處理命令 5")
	}
}

func TestExecStrikeForceDoesNotSetAssignedBit(t *testing.T) {
	// §29：值 4 與值 3 的差別之一——**成功後不標旗標**。
	sim := mkTracedBattle(t, 20000, 18000)
	route := func(to, from CellIndex) CellIndex { return to }
	a := sim.Attacker[0]
	a.Command = BattleCmdSeekTarget
	a.Flags13 = 0

	got := sim.ExecuteAction(ActBStrikeForce, sim.Attacker, sim.Defender, route)
	if !got.Implemented {
		t.Fatal("打主力周邊已經實作了")
	}
	if got.Assigned > 0 && a.Assigned() {
		t.Error("值 4 不該立 +13 bit 7——那是值 3 才做的（§29 的對照表）")
	}
	if got.Assigned > 0 && a.TargetUnit == 0 {
		t.Error("該設目標單位")
	}
}
