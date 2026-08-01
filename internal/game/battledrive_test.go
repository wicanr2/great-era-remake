package game

import "testing"

// 造一場湖北打河南的戰鬥，攻守兵力可調——用來觀測決策鏈在不同局勢的反應。
func mkTracedBattle(t *testing.T, atkForce, defForce uint16) *BattleSim {
	t.Helper()
	m := loadTestMap(t)
	const wu, zhang = GeneralID(58), GeneralID(166)

	attackers := []*Combatant{mkUnit(101, wu, Branch1, atkForce)}

	bf, err := m.Battlefield(19)
	if err != nil {
		t.Fatal(err)
	}
	var cell CellIndex = NoCell
	for i := 0; i < CellCount; i++ {
		c := CellIndex(i)
		col, row := c.ColRow()
		if bf.Owner[row][col] == 0 && bf.Tiles[row][col].MoveCost() < 255 {
			cell = c
			break
		}
	}
	if cell == NoCell {
		t.Fatal("找不到可站的守方格")
	}
	defenders := []*Combatant{mkUnit(201, zhang, Branch1, defForce)}
	defenders[0].Cell = cell

	sim, err := NewBattleSim(m, 19, 26, attackers, defenders, StrengthOpts{Stage: 1})
	if err != nil {
		t.Fatal(err)
	}
	return sim
}

func TestSideStrengthCountsOnlyLiving(t *testing.T) {
	sim := mkTracedBattle(t, 20000, 18000)
	full := sim.SideStrength(sim.Defender)
	if full <= 0 {
		t.Fatalf("守方戰力該是正數，實際 %d", full)
	}
	sim.Defender[0].Strength.Force = 0 // 打光
	if got := sim.SideStrength(sim.Defender); got != 0 {
		t.Errorf("陣亡的不該計入戰力，實際 %d", got)
	}
}

func TestLeaderOnField(t *testing.T) {
	sim := mkTracedBattle(t, 20000, 18000)
	// 守方單位的 General 是 201。
	if !LeaderOnField(sim.Defender, 201) {
		t.Error("領袖就在隊伍裡，該回 true")
	}
	if LeaderOnField(sim.Defender, 999) {
		t.Error("不在隊伍裡的領袖該回 false")
	}
	if LeaderOnField(sim.Defender, 0) {
		t.Error("領袖 ID 為 0（無主）一律 false")
	}
	// 陣亡的不算在場。
	sim.Defender[0].Strength.Force = 0
	if LeaderOnField(sim.Defender, 201) {
		t.Error("陣亡的領袖不該算在場")
	}
}

// ⭐ 這個測試是這一層存在的理由：看決策鏈在**真實局勢**下選什麼。
func TestDecideTurnOnLopsidedBattle(t *testing.T) {
	// 守方被輾壓（攻方 20000 對守方 1000）——分支 A 該走必勝結算。
	sim := mkTracedBattle(t, 20000, 1000)
	d := sim.DecideTurn(1, BattleChainGates{}, 201, 58)

	atk, def := sim.SideStrength(sim.Attacker), sim.SideStrength(sim.Defender)
	if def*5 > atk {
		t.Skipf("這組兵力沒有達到五倍差（攻 %d 守 %d），換個數字再測", atk, def)
	}
	if d.A.Action != ActADecisive {
		t.Errorf("守方被輾壓，分支 A 該走必勝結算，實際 %s（%s）",
			BattleActionName(d.A.Action), d.A.Step)
	}
	// ⚠️ §45：sub_53619 為真時就不准直接判勝負，改走別條。
	d2 := sim.DecideTurn(1, BattleChainGates{Sub53619: true}, 201, 58)
	if d2.A.Action == ActADecisive {
		t.Error("sub_53619 為真時不該必勝結算——那是 §45 的第二道門檻")
	}
}

func TestTraceDecisionsDoesNotChangeOutcome(t *testing.T) {
	// TraceDecisions 只在旁邊觀測，戰鬥結果必須與 AutoResolve 一致。
	a := mkTracedBattle(t, 20000, 18000)
	want := a.AutoResolve(50)

	b := mkTracedBattle(t, 20000, 18000)
	trace, got := b.TraceDecisions(50, BattleChainGates{}, 201, 58)

	if got.Turns != want.Turns || got.Decided != want.Decided ||
		got.AttackerWon != want.AttackerWon ||
		got.AttackerLoss != want.AttackerLoss || got.DefenderLoss != want.DefenderLoss {
		t.Errorf("觀測不該改變戰鬥結果\n  AutoResolve %+v\n  Traced      %+v", want, got)
	}
	if len(trace) != got.Turns {
		t.Errorf("每回合各一筆紀錄：%d 筆 vs %d 回合", len(trace), got.Turns)
	}
	if len(trace) > 0 && trace[0].Turn != 1 {
		t.Errorf("回合序號從 1 起，實際 %d", trace[0].Turn)
	}
}
