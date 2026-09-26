package game

import (
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

func effectsTestField() *Battlefield {
	f := &Battlefield{}
	for y := range f.Tiles {
		for x := range f.Tiles[y] {
			f.Tiles[y][x] = assets.Tile{Kind: assets.TilePlain, Rail: assets.NoRail}
		}
	}
	return f
}

func effectsTestUnit(id GeneralID, branch uint8, cell CellIndex, attacking bool, faction GeneralID) *Combatant {
	return &Combatant{
		CombatUnit: CombatUnit{
			General: id, Cell: cell, Attacking: attacking, Faction: faction,
			Max: 12, Current: 12, Active: true, Decaying: 80, NextCell: NoCell,
		},
		Strength: StrengthInput{
			Ability: 90, Force: 20000, F19: 60, F20: 60, F29: 64, F30: 80,
			Branch: branch, General: id, Faction: faction,
		},
	}
}

func effectsTestSim(units ...*Combatant) *BattleSim {
	s := &BattleSim{Field: effectsTestField(), Opts: StrengthOpts{Stage: 1}, byID: make(map[GeneralID]*Combatant)}
	for _, u := range units {
		s.byID[u.General] = u
		s.Occ[u.Cell] = u.General
		if u.Attacking {
			s.Attacker = append(s.Attacker, u)
		} else {
			s.Defender = append(s.Defender, u)
		}
	}
	return s
}

func TestEngageWithEffectsWritesConfirmedSideEffects(t *testing.T) {
	attCell := CellIndex(100)
	targetCell, ok := attCell.Neighbour(DirDown)
	if !ok {
		t.Fatal("測試格沒有相鄰格")
	}
	attacker := effectsTestUnit(1, BranchInfantry, attCell, true, 10)
	target := effectsTestUnit(2, BranchInfantry, targetCell, false, 20)
	s := effectsTestSim(attacker, target)
	result, err := s.EngageWithEffects(attacker, target)
	if err != nil {
		t.Fatal(err)
	}
	if result.Effect.Kind != BattleEffectStandard || result.Effect.ExperienceGainAttacker != 5 ||
		result.Effect.StaminaLossAttacker != 1 || result.Effect.MoraleLossAttacker != 1 {
		t.Fatalf("正規副作用未接上：%+v", result.Effect)
	}
	if attacker.Experience != 5 || attacker.Strength.F29 != 63 || attacker.Strength.F30 != 79 {
		t.Fatalf("攻方欄位未寫回：經驗=%d 體力=%d 士氣=%d", attacker.Experience, attacker.Strength.F29, attacker.Strength.F30)
	}
}

func TestResolveBattleAttackUsesCoordinationAndStopsOnDeath(t *testing.T) {
	attCell := CellIndex(100)
	targetCell, ok := attCell.Neighbour(DirDown)
	if !ok {
		t.Fatal("測試格沒有目標鄰格")
	}
	supportCell, ok := attCell.Neighbour(DirUp)
	if !ok || supportCell == targetCell {
		t.Fatal("測試格沒有支援鄰格")
	}
	attacker := effectsTestUnit(1, BranchInfantry, attCell, true, 10)
	target := effectsTestUnit(2, BranchInfantry, targetCell, false, 20)
	support := effectsTestUnit(3, BranchInfantry, supportCell, false, 20)
	s := effectsTestSim(attacker, target, support)
	result, err := s.ResolveBattleAttack(attacker, target)
	if err != nil {
		t.Fatal(err)
	}
	if result.Effect.Kind != BattleEffectCoordinated || len(result.Effect.Supporters) != 1 ||
		len(result.SupportLosses) != 1 {
		t.Fatalf("協同分派錯誤：%+v", result)
	}
	if result.LossTarget <= 0 || result.SupportLosses[0].Loss <= 0 {
		t.Fatalf("協同損失未寫回：%+v", result)
	}
}

func TestResolveBattleAttackCavalryChargeRecordsCellMotions(t *testing.T) {
	attCell := CellIndex(100)
	targetCell, ok := attCell.Neighbour(DirDown)
	if !ok {
		t.Fatal("測試格沒有目標鄰格")
	}
	attacker := effectsTestUnit(1, BranchCavalry, attCell, true, 10)
	target := effectsTestUnit(2, BranchInfantry, targetCell, false, 20)
	s := effectsTestSim(attacker, target)
	result, err := s.ResolveBattleAttack(attacker, target)
	if err != nil {
		t.Fatal(err)
	}
	if result.Effect.Kind != BattleEffectCavalryCharge || len(result.Effect.CellMotions) == 0 {
		t.Fatalf("衝鋒副作用未接上：%+v", result.Effect)
	}
	if target.Cell != targetCell || result.Effect.StaminaLossAttacker != 5 ||
		result.Effect.StaminaLossTarget != 2 || result.Effect.ExperienceGainAttacker != 15 {
		t.Fatalf("衝鋒不應永久搬格或漏欄位：目標格=%d 效果=%+v", target.Cell, result.Effect)
	}
}

func TestResolveBattleAttackVisualBranchDoesNotChangeForce(t *testing.T) {
	attCell := CellIndex(100)
	targetCell, ok := attCell.Neighbour(DirDown)
	if !ok {
		t.Fatal("測試格沒有目標鄰格")
	}
	attacker := effectsTestUnit(1, BranchInfantry, attCell, true, 10)
	target := effectsTestUnit(2, BranchArtiller, targetCell, false, 20)
	s := effectsTestSim(attacker, target)
	beforeA, beforeT := attacker.Force(), target.Force()
	result, err := s.ResolveBattleAttack(attacker, target)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Effect.VisualOnly || result.LossAttacker != 0 || result.LossTarget != 0 ||
		attacker.Force() != beforeA || target.Force() != beforeT {
		t.Fatalf("特殊視覺分支不應改兵力：%+v 攻=%d/%d 守=%d/%d", result, attacker.Force(), beforeA, target.Force(), beforeT)
	}
}

func TestSyncBattleStatsToGeneralsWritesOnlyKnownFields(t *testing.T) {
	u := effectsTestUnit(2, BranchInfantry, 100, true, 10)
	u.Experience = 44
	u.Strength.Ability, u.Strength.F19, u.Strength.F20, u.Strength.F29, u.Strength.F30 = 77, 61, 62, 63, 64
	s := effectsTestSim(u)
	generals := []General{{Raw: [GeneralRecordSize]byte{7}}, {Raw: [GeneralRecordSize]byte{8}, Force: 1}}
	updated, err := s.SyncBattleStatsToGenerals(generals)
	if err != nil {
		t.Fatal(err)
	}
	if updated != 1 {
		t.Fatalf("同步筆數=%d，預期 1", updated)
	}
	g := generals[1]
	if g.Force != 20000 || g.AbilityA != 77 || g.F19 != 61 || g.F20 != 62 ||
		g.Stamina != 63 || g.F30 != 64 || g.Experience != 44 || g.Raw[0] != 8 {
		t.Fatalf("已解欄位／Raw 同步錯誤：%+v", g)
	}
}
