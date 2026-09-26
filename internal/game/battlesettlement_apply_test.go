package game

import "testing"

func settlementProvinceTable() *ProvinceTable {
	t := &ProvinceTable{}
	t.Province[17] = Province{Commander: 20, Flags: ProvinceFlagInBattle}
	t.Province[18] = Province{Commander: 30}
	return t
}

func settlementSim() *BattleSim {
	attacker := &Combatant{
		CombatUnit: CombatUnit{General: 1, Province: 19, Faction: 30, Cell: 1},
		Strength:   StrengthInput{Force: 1000},
	}
	defender := &Combatant{
		CombatUnit: CombatUnit{General: 2, Province: 18, Faction: 20, Cell: 2},
		Strength:   StrengthInput{Force: 1000},
	}
	return &BattleSim{From: 19, At: 18, Attacker: []*Combatant{attacker}, Defender: []*Combatant{defender}}
}

func TestApplyBattleSettlementAttackerCaptureMovesLivingGenerals(t *testing.T) {
	table := settlementProvinceTable()
	sim := settlementSim()
	generals := []General{{Province: 19}, {Province: 18}}
	units := []CombatUnit{{General: 1, Province: 19}, {General: 2, Province: 18}}
	settlement := BattleSettlement{Kind: BattleSettlementDecisive, Winner: BattleSideFirst}
	if err := ApplyBattleSettlement(sim, table, generals, units, settlement); err != nil {
		t.Fatal(err)
	}
	target, _ := table.At(18)
	if target.InBattle() || target.Commander != 30 {
		t.Fatalf("攻方勝結算 = commander %d flags %#x", target.Commander, target.Flags)
	}
	if generals[0].Province != 18 || units[0].Province != 18 {
		t.Fatalf("存活攻方沒有移防：general=%d unit=%d", generals[0].Province, units[0].Province)
	}
	if generals[1].Province != 18 || units[1].Province != 18 {
		t.Fatalf("守方資料被意外改動：general=%d unit=%d", generals[1].Province, units[1].Province)
	}
}

func TestApplyBattleSettlementDefenderAndDrawOnlyClearBattleFlag(t *testing.T) {
	for _, tc := range []struct {
		name string
		kind BattleSettlementKind
		win  BattleSide
	}{
		{name: "守方勝", kind: BattleSettlementDecisive, win: BattleSideSecond},
		{name: "回合平局", kind: BattleSettlementTurnLimitDraw, win: BattleSideNone},
	} {
		t.Run(tc.name, func(t *testing.T) {
			table := settlementProvinceTable()
			sim := settlementSim()
			generals := []General{{Province: 19}, {Province: 18}}
			units := []CombatUnit{{General: 1, Province: 19}, {General: 2, Province: 18}}
			if err := ApplyBattleSettlement(sim, table, generals, units,
				BattleSettlement{Kind: tc.kind, Winner: tc.win}); err != nil {
				t.Fatal(err)
			}
			target, _ := table.At(18)
			if target.InBattle() || target.Commander != 20 {
				t.Fatalf("不應易主：commander %d flags %#x", target.Commander, target.Flags)
			}
			if generals[0].Province != 19 || units[0].Province != 19 {
				t.Fatalf("非攻方勝不應移防：general=%d unit=%d", generals[0].Province, units[0].Province)
			}
		})
	}
}

func TestApplyBattleSettlementRetreatDoesNotMutate(t *testing.T) {
	table := settlementProvinceTable()
	sim := settlementSim()
	generals := []General{{Province: 19}, {Province: 18}}
	if err := ApplyBattleSettlement(sim, table, generals, nil, BattleSettlementImmediateRetreat()); err != nil {
		t.Fatal(err)
	}
	target, _ := table.At(18)
	if !target.InBattle() || target.Commander != 20 || generals[0].Province != 19 {
		t.Fatalf("立即撤退不應投影：province=%+v generals=%+v", target, generals)
	}
}

func TestApplyBattleSettlementRejectsInvalidWinner(t *testing.T) {
	if err := ApplyBattleSettlement(settlementSim(), settlementProvinceTable(), nil, nil,
		BattleSettlement{Kind: BattleSettlementDecisive}); err == nil {
		t.Fatal("未分出的勝方不應被接受")
	}
}
