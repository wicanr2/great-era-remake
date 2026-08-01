package game

import "testing"

// 造一場**守方不在城市上**的戰鬥：值 3 挑不到城市，
// 所以會走 `sub_3B9D8` 那條追人的路（§52）。
func mkNoCityBattle(t *testing.T, nAtk, nDef int, defForce uint16) *BattleSim {
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
	cityset := map[CellIndex]bool{}
	for _, c := range CityCells(bf) {
		cityset[c] = true
	}
	var cells []CellIndex
	for i := CellCount - 1; i >= 0 && len(cells) < nDef; i-- {
		c := CellIndex(i)
		col, row := c.ColRow()
		if cityset[c] || bf.Tiles[row][col].MoveCost() >= 255 {
			continue
		}
		cells = append(cells, c)
	}
	if len(cells) < nDef {
		t.Fatalf("找不到 %d 個非城市的可站格", nDef)
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

func directRoute(to, from CellIndex) CellIndex {
	if !to.Valid() || !from.Valid() {
		return NoCell
	}
	if Adjacent(from, to) {
		return to
	}
	best, bestD := NoCell, 1<<30
	for _, n := range from.Neighbours() {
		if d := CellManhattan(n, to); d < bestD {
			best, bestD = n, d
		}
	}
	return best
}

func TestChaseAssignsWhenNoCityAvailable(t *testing.T) {
	sim := mkNoCityBattle(t, 3, 2, 5000)
	sim.BeginTurn()
	for _, u := range sim.Attacker {
		u.Command = BattleCmdSeekTarget
		u.ClearAssignment()
	}
	r := sim.ExecuteAction(ActBTakeCity, sim.Attacker, sim.Defender, directRoute)
	if !r.Implemented {
		t.Fatal("值 3 該是已實作的")
	}
	if r.Assigned == 0 {
		t.Fatalf("守方不在城市上，該改去追人；實得 %+v", r)
	}
	for _, u := range sim.Attacker {
		if u.TargetUnit == 0 {
			continue
		}
		if sim.Unit(u.TargetUnit) == nil {
			t.Errorf("將領 %d 的目標 %d 不存在", u.General, u.TargetUnit)
		}
	}
}

func TestChaseFirepowerCapSpreadsUnits(t *testing.T) {
	// ⭐ §52 的重點規則：一個敵方單位最多吸引三倍於它的我方戰力。
	// 攻方 20000×4 對守方 5000×2——若沒有上限，四個都會撲第一個敵人。
	sim := mkNoCityBattle(t, 4, 2, 5000)
	sim.BeginTurn()
	for _, u := range sim.Attacker {
		u.Command = BattleCmdSeekTarget
		u.ClearAssignment()
	}
	n := sim.ChaseAssign(sim.Attacker, sim.Defender, directRoute)
	if n == 0 {
		t.Fatal("該有人被指派")
	}
	count := map[GeneralID]int{}
	for _, u := range sim.Attacker {
		if u.TargetUnit != 0 {
			count[u.TargetUnit]++
		}
	}
	if len(count) < 2 {
		t.Errorf("火力上限該把單位分散到兩個敵人身上，實際只打了 %d 個：%v",
			len(count), count)
	}
}

func TestChaseSkipsUnitsNotAwaitingOrders(t *testing.T) {
	sim := mkNoCityBattle(t, 2, 2, 5000)
	sim.BeginTurn()
	for _, u := range sim.Attacker {
		u.Command = BattleCmdStandby // 待命，不是等待指派
		u.ClearAssignment()
	}
	if n := sim.ChaseAssign(sim.Attacker, sim.Defender, directRoute); n != 0 {
		t.Errorf("待命的單位不該被派工，實得 %d", n)
	}
}

func TestChaseNoFoes(t *testing.T) {
	sim := mkNoCityBattle(t, 2, 2, 5000)
	sim.BeginTurn()
	if n := sim.ChaseAssign(sim.Attacker, nil, directRoute); n != 0 {
		t.Errorf("沒有敵人該回 0，實得 %d", n)
	}
}

func TestAttackerFewerThanCities(t *testing.T) {
	// §53：`sub_56D13` = 攻方在場單位數 < 城市格數。
	sim := mkNoCityBattle(t, 1, 2, 5000)
	nCity := len(CityCells(sim.Field))
	if nCity == 0 {
		t.Skip("這張圖沒有城市格")
	}
	if got, want := sim.AttackerFewerThanCities(), 1 < nCity; got != want {
		t.Errorf("攻方 1 個、城市 %d 個 → 該回 %v，實得 %v", nCity, want, got)
	}
}

func TestAttackerNearCityDetectsApproach(t *testing.T) {
	// §6：城市格上或距離 2 以內站著攻方單位。
	sim := mkNoCityBattle(t, 1, 1, 5000)
	cities := CityCells(sim.Field)
	if len(cities) == 0 {
		t.Skip("這張圖沒有城市格")
	}
	// 把攻方挪到遠離所有城市的地方——找不到就跳過（圖太小）。
	far := NoCell
	for i := 0; i < CellCount; i++ {
		c := CellIndex(i)
		col, row := c.ColRow()
		if sim.Field.Tiles[row][col].MoveCost() >= 255 {
			continue
		}
		near := false
		for _, cc := range cities {
			if WithinTwoSteps(c, cc) {
				near = true
				break
			}
		}
		if !near {
			far = c
			break
		}
	}
	if far == NoCell {
		t.Skip("這張圖每一格都在城市兩步之內")
	}
	sim.Attacker[0].Cell = far
	if sim.AttackerNearCity() {
		t.Error("攻方離所有城市都超過兩步，不該算逼近")
	}
	// 挪到城市上 → 一定算逼近（WithinTwoSteps(c, c) 成立）。
	sim.Attacker[0].Cell = cities[0]
	if !sim.AttackerNearCity() {
		t.Error("攻方站在城市上該算逼近")
	}
}

func TestFallbackToStandbyDefenderHasNoGuards(t *testing.T) {
	// 第二方（守方）沒有保護：連首位單位都會被打回待命。
	us := []*Combatant{
		{CombatUnit: CombatUnit{General: 1, Command: BattleCmdSeekTarget}},
		{CombatUnit: CombatUnit{General: 2, Command: BattleCmdStandby}},
		{CombatUnit: CombatUnit{General: 3, Command: BattleCmdStandby}},
		{CombatUnit: CombatUnit{General: 4, Command: BattleCmdStandby}},
		{CombatUnit: CombatUnit{General: 5, Command: BattleCmdSeekTarget}},
	}
	if !FallbackToStandby(us, us[0], false) {
		t.Error("守方的首位單位也該被打回待命")
	}
	if us[0].Command != BattleCmdStandby {
		t.Errorf("命令該是待命，實得 %d", us[0].Command)
	}
	// 待命人數已滿也照打。
	if !FallbackToStandby(us, us[4], false) {
		t.Error("守方沒有待命人數上限")
	}
}

func TestFallbackToStandbyAttackerGuardsLeader(t *testing.T) {
	// ⭐ 第一方（攻方）的首位單位是主帥，永遠不打回。
	us := []*Combatant{
		{CombatUnit: CombatUnit{General: 1, Command: BattleCmdSeekTarget}},
		{CombatUnit: CombatUnit{General: 2, Command: BattleCmdSeekTarget}},
	}
	if FallbackToStandby(us, us[0], true) {
		t.Error("攻方的首位單位不該被打回待命")
	}
	if us[0].Command != BattleCmdSeekTarget {
		t.Error("首位單位的命令不該被改")
	}
	if !FallbackToStandby(us, us[1], true) {
		t.Error("非首位的單位該被打回")
	}
}

func TestFallbackToStandbyAttackerCapsStandbyCount(t *testing.T) {
	// ⭐ index 2..10 已經有 3 個待命就不再多一個（`> 2` 才擋 → 上限 3）。
	mk := func(n int) []*Combatant {
		us := []*Combatant{{CombatUnit: CombatUnit{General: 1,
			Command: BattleCmdSeekTarget}}}
		for i := 0; i < n; i++ {
			us = append(us, &Combatant{CombatUnit: CombatUnit{
				General: GeneralID(2 + i), Command: BattleCmdStandby}})
		}
		us = append(us, &Combatant{CombatUnit: CombatUnit{
			General: 99, Command: BattleCmdSeekTarget}})
		return us
	}
	for _, c := range []struct {
		standby int
		want    bool
	}{{0, true}, {1, true}, {2, true}, {3, false}, {4, false}} {
		us := mk(c.standby)
		if got := FallbackToStandby(us, us[len(us)-1], true); got != c.want {
			t.Errorf("已有 %d 個待命 → 該回 %v，實得 %v", c.standby, c.want, got)
		}
	}
}

func TestFallbackDoesNotTouchNextCell(t *testing.T) {
	// ⚠️ 原版只設 +9 與 +10，**不動 +12**——與 ResetToStandby 不同。
	u := &Combatant{CombatUnit: CombatUnit{General: 7,
		Command: BattleCmdSeekTarget, NextCell: 42, TargetUnit: 9}}
	FallbackToStandby([]*Combatant{{CombatUnit: CombatUnit{General: 1}}, u}, u, true)
	if u.NextCell != 42 {
		t.Errorf("下一跳不該被動，實得 %d", u.NextCell)
	}
	if u.TargetUnit != 0 {
		t.Errorf("目標該被清成 0，實得 %d", u.TargetUnit)
	}
}

// §57：值 4 的候選清單**只收對方單位**。初版漏了陣營條件，
// 會把我方單位列進候選然後派去打自己人。
func TestStrikeForcePoolExcludesOwnUnits(t *testing.T) {
	sim := mkNoCityBattle(t, 3, 2, 5000)
	sim.BeginTurn()
	for _, u := range sim.Attacker {
		u.Command = BattleCmdSeekTarget
		u.ClearAssignment()
	}
	// 把攻方全部擠到守方主力旁邊——沒有陣營過濾的話，
	// 他們會互相成為候選。
	lead := sim.Defender[0]
	n := 0
	for _, c := range lead.Cell.Neighbours() {
		if n >= len(sim.Attacker) {
			break
		}
		if sim.Occ[c] == 0 {
			sim.Move(sim.Attacker[n].General, dirToward(sim.Attacker[n].Cell, c))
			sim.Attacker[n].Cell = c
			sim.Occ[c] = sim.Attacker[n].General
			n++
		}
	}
	r := sim.ExecuteAction(ActBStrikeForce, sim.Attacker, sim.Defender, directRoute)
	if !r.Implemented {
		t.Fatal("值 4 該是已實作的")
	}
	atk := map[GeneralID]bool{}
	for _, u := range sim.Attacker {
		atk[u.General] = true
	}
	for _, u := range sim.Attacker {
		if u.TargetUnit != 0 && atk[u.TargetUnit] {
			t.Errorf("將領 %d 被派去打我方的 %d", u.General, u.TargetUnit)
		}
	}
}
