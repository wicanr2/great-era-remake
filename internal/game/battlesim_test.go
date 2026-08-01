package game

import "testing"

// 造一個單位。能力值與屬性用中庸值，兵種可指定。
func mkUnit(id GeneralID, faction GeneralID, branch uint8, force uint16) *Combatant {
	return &Combatant{
		CombatUnit: CombatUnit{
			General: id, Faction: faction, Cell: NoCell, NextCell: NoCell,
			Max: 12, Current: 12, Active: true, Decaying: 80,
		},
		Strength: StrengthInput{
			Ability: 75, Force: force, F19: 60, F20: 60, F29: 64, F30: 80,
			Branch: branch, General: id, Faction: faction,
		},
	}
}

// 用原版資料跑一場完整的戰鬥：湖北(26) 打河南(19)。
// 這是規則層第一次端到端跑起來——部署、移動、交戰、戰損、判定結束。
func TestFullBattleHubeiAttacksHenan(t *testing.T) {
	m := loadTestMap(t)

	const wu, zhang = GeneralID(58), GeneralID(166) // 吳佩孚、張作霖
	attackers := []*Combatant{
		mkUnit(101, wu, Branch1, 20000),
		mkUnit(102, wu, Branch1, 15000),
	}
	// 守方擺在戰場中央的腹地（WARPOS 值 0 的那片）。
	defCells := []CellIndex{}
	bf, err := m.Battlefield(19)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < CellCount && len(defCells) < 2; i++ {
		c := CellIndex(i)
		col, row := c.ColRow()
		if bf.Owner[row][col] != 0 {
			continue
		}
		if bf.Tiles[row][col].MoveCost() >= 255 {
			continue // 高山站不上去
		}
		defCells = append(defCells, c)
	}
	defenders := []*Combatant{
		mkUnit(201, zhang, Branch1, 18000),
		mkUnit(202, zhang, Branch1, 12000),
	}
	for i, u := range defenders {
		u.Cell = defCells[i]
	}

	sim, err := NewBattleSim(m, 19, 26, attackers, defenders, StrengthOpts{Stage: 1})
	if err != nil {
		t.Fatal(err)
	}

	// 部署：攻方應該落在湖北的進場區裡，第一個是 cell 186。
	if attackers[0].Cell != 186 {
		t.Errorf("第一個攻方單位落在 %d，應為 186", attackers[0].Cell)
	}
	zone, _ := m.DeployZone(19, 26)
	inZone := map[CellIndex]bool{}
	for _, c := range zone {
		inZone[c] = true
	}
	for _, u := range attackers {
		if !inZone[u.Cell] {
			t.Errorf("攻方單位 %d 落在 %d，不在湖北的進場區", u.General, u.Cell)
		}
		// 佔用表要指回去。
		if sim.Occ[u.Cell] != u.General {
			t.Errorf("佔用表沒指回單位 %d", u.General)
		}
	}

	// 回合初始化：機動力回滿。
	attackers[0].Current = 3
	sim.BeginTurn()
	if attackers[0].Current != attackers[0].Max {
		t.Errorf("回合開始機動力應回滿，得到 %d/%d",
			attackers[0].Current, attackers[0].Max)
	}

	// 移動：往北走（DirUp），成本依地形扣。
	before := attackers[0].Current
	dst, err := sim.Move(101, DirUp)
	if err != nil {
		t.Fatalf("往北移動失敗：%v", err)
	}
	if attackers[0].Current >= before {
		t.Error("移動沒有扣機動力")
	}
	if sim.Occ[dst] != 101 || attackers[0].Cell != dst {
		t.Error("移動後佔用表或單位位置沒同步")
	}
}

// 交戰：把兩個單位擺成相鄰，打一次，驗雙方都掉血且守方掉得多。
func TestEngageBothSidesTakeLosses(t *testing.T) {
	m := loadTestMap(t)
	const wu, zhang = GeneralID(58), GeneralID(166)

	atk := mkUnit(101, wu, Branch1, 20000)
	def := mkUnit(201, zhang, Branch1, 20000)

	// 找一組相鄰、都站得上去的腹地格。
	bf, _ := m.Battlefield(19)
	var a, b CellIndex = NoCell, NoCell
	for i := 0; i < CellCount && a == NoCell; i++ {
		c := CellIndex(i)
		col, row := c.ColRow()
		if bf.Owner[row][col] != 0 || bf.Tiles[row][col].MoveCost() >= 255 {
			continue
		}
		for _, n := range c.Neighbours() {
			ncol, nrow := n.ColRow()
			if bf.Owner[nrow][ncol] == 0 && bf.Tiles[nrow][ncol].MoveCost() < 255 {
				a, b = c, n
				break
			}
		}
	}
	if a == NoCell {
		t.Fatal("找不到兩格相鄰的腹地，測試前提不成立")
	}
	def.Cell, atk.Cell = a, b

	sim, err := NewBattleSim(m, 19, 26, nil, []*Combatant{def}, StrengthOpts{Stage: 1})
	if err != nil {
		t.Fatal(err)
	}
	// 把攻方手動放上場（這一測不想走部署掃描）。
	atk.Attacking = true
	sim.Attacker = []*Combatant{atk}
	sim.Occ[atk.Cell] = atk.General
	sim.byID[atk.General] = atk

	lossA, lossD, err := sim.Engage(atk, def)
	if err != nil {
		t.Fatal(err)
	}
	if lossA <= 0 || lossD <= 0 {
		t.Errorf("一次交戰雙方都該有損失，得到 攻 %d／守 %d", lossA, lossD)
	}
	if atk.Force() != 20000-uint16(lossA) {
		t.Errorf("攻方兵力 = %d，應為 %d", atk.Force(), 20000-lossA)
	}
	if def.Force() != 20000-uint16(lossD) {
		t.Errorf("守方兵力 = %d，應為 %d", def.Force(), 20000-lossD)
	}
	t.Logf("攻方損失 %d、守方損失 %d（格 %d vs %d）", lossA, lossD, atk.Cell, def.Cell)

	// 不相鄰就打不到。
	def.Cell = 0
	atk.Cell = 195
	if _, _, err := sim.Engage(atk, def); err == nil {
		t.Error("隔那麼遠竟然打得到")
	}
}

// 打到全滅：戰鬥要能收斂，而且 Over() 要回報攻方獲勝。
func TestBattleConvergesToAnnihilation(t *testing.T) {
	m := loadTestMap(t)
	const wu, zhang = GeneralID(58), GeneralID(166)

	// 攻方壓倒性強（滿員 + 滿能力），守方弱小。
	atk := mkUnit(101, wu, Branch5, 20000)
	atk.Strength.Ability = 100
	def := mkUnit(201, zhang, Branch1, 500)
	def.Strength.Ability = 40

	bf, _ := m.Battlefield(19)
	var a, b CellIndex = NoCell, NoCell
	for i := 0; i < CellCount && a == NoCell; i++ {
		c := CellIndex(i)
		col, row := c.ColRow()
		if bf.Owner[row][col] != 0 || bf.Tiles[row][col].MoveCost() >= 255 {
			continue
		}
		for _, n := range c.Neighbours() {
			ncol, nrow := n.ColRow()
			if bf.Owner[nrow][ncol] == 0 && bf.Tiles[nrow][ncol].MoveCost() < 255 {
				a, b = c, n
				break
			}
		}
	}
	def.Cell, atk.Cell = a, b

	sim, err := NewBattleSim(m, 19, 26, nil, []*Combatant{def}, StrengthOpts{Stage: 1})
	if err != nil {
		t.Fatal(err)
	}
	atk.Attacking = true
	sim.Attacker = []*Combatant{atk}
	sim.Occ[atk.Cell] = atk.General
	sim.byID[atk.General] = atk

	// 最多打 50 回合。收斂不了就是規則有問題。
	rounds := 0
	for ; rounds < 50; rounds++ {
		if over, _ := sim.Over(); over {
			break
		}
		if _, _, err := sim.Engage(atk, def); err != nil {
			t.Fatal(err)
		}
		sim.Sweep()
	}
	over, attackerWon := sim.Over()
	if !over {
		t.Fatalf("打了 %d 回合還沒結束：攻 %d／守 %d",
			rounds, atk.Force(), def.Force())
	}
	if !attackerWon {
		t.Errorf("壓倒性的攻方竟然沒贏：攻 %d／守 %d", atk.Force(), def.Force())
	}
	t.Logf("%d 回合分出勝負，攻方剩 %d", rounds, atk.Force())

	// 陣亡的單位要從佔用表清掉。
	if sim.Occ[a] != 0 {
		t.Errorf("守方陣亡了，格 %d 卻還被佔著", a)
	}
}

// 部署放不下時要回報錯誤，不是靜靜吃掉。
func TestDeployOverflowIsReported(t *testing.T) {
	m := loadTestMap(t)
	// 進場區只有 10 格，塞 11 個會超過 UnitsPerSide 先被擋。
	var many []*Combatant
	for i := 0; i < UnitsPerSide+1; i++ {
		many = append(many, mkUnit(GeneralID(300+i), 58, Branch1, 1000))
	}
	if _, err := NewBattleSim(m, 19, 26, many, nil, StrengthOpts{}); err == nil {
		t.Error("超過每方 10 個部隊卻沒被擋")
	}
}
