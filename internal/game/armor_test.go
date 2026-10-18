package game

import (
	"math"
	"testing"
)

func TestArmorDamageAdoptedExamplesAndBounds(t *testing.T) {
	for _, tc := range []struct {
		loss            uint16
		piercing, armor uint32
		want            uint16
	}{
		{100, 5, 50, 55}, {100, 10, 50, 60}, {100, 20, 50, 70},
		{100, 30, 50, 80}, {100, 40, 50, 90}, {100, 80, 50, 100},
		{100, 0, 50, 50}, {100, 0, 0, 100}, {1, 0, 70, 1}, {0, 5, 50, 0},
		{65535, 0, 70, 32768}, {65535, 80, 70, 65535},
	} {
		if got := ArmorDamage(tc.loss, tc.piercing*1000, tc.armor*1000); got != tc.want {
			t.Fatalf("%+v: got %d", tc, got)
		}
	}
	for _, armor := range []uint32{0, 1, 20000, 50000, 70000, math.MaxUint32} {
		for _, piercing := range []uint32{0, 5000, 80000, math.MaxUint32} {
			got := ArmorDamage(65535, piercing, armor)
			if got < 32768 {
				t.Fatalf("cap or integer overflow: %d", got)
			}
		}
	}
}

func TestProtectionActualMixedAndMissingEquipment(t *testing.T) {
	a := &Armory{Units: map[GeneralID]*UnitEquipment{1: {Items: map[EquipmentID]uint32{"light-tank": 10, "legacy-armor": 10}}}}
	if got := a.Protection(1, BranchArmour, 200); got != (EquipmentProtection{35000, 40000}) {
		t.Fatal("mixed loadout", got)
	}
	delete(a.Units[1].Items, "legacy-armor")
	if got := a.Protection(1, BranchArmour, 200); got != (EquipmentProtection{25000, 25000}) {
		t.Fatal("missing vehicles conferred free protection", got)
	}
	a.Units[1].Items["antitank"] = math.MaxUint32
	if got := a.Protection(1, BranchArmour, 200); got.Piercing != 25000 {
		t.Fatal("incompatible equipment", got)
	}
	a.Units[1].Items["light-tank"] = math.MaxUint32
	if got := a.Protection(1, BranchArmour, 200); got != (EquipmentProtection{50000, 50000}) {
		t.Fatal("overissue overflow", got)
	}
	for _, got := range []EquipmentProtection{a.Protection(1, 0, 200), a.Protection(1, BranchArmour, 0), a.Protection(2, BranchArmour, 200), (*Armory)(nil).Protection(1, BranchArmour, 200)} {
		if got != (EquipmentProtection{}) {
			t.Fatal("invalid/no-ledger state", got)
		}
	}
}

func armorTestLedger(t *testing.T, s *BattleSim) *Armory {
	t.Helper()
	w := &AIWorld{Table: &ProvinceTable{}}
	w.Table.Province[0].Commander = 1
	w.Table.Province[1].Commander = 2
	for _, u := range s.all() {
		u.Province = 1
		if !u.Attacking {
			u.Province = 2
		}
		w.Units = append(w.Units, u.CombatUnit)
		w.Strengths = append(w.Strengths, u.Strength)
	}
	a, err := NewArmory(w, 36, 187)
	if err != nil {
		t.Fatal(err)
	}
	s.Opts.Armory = a
	for _, u := range s.all() {
		u.armory = a
	}
	return a
}

func TestArmorRealStandardAttackAndOriginalVisualBranch(t *testing.T) {
	cell := CellIndex(100)
	other, _ := cell.Neighbour(DirDown)
	atk := effectsTestUnit(1, BranchInfantry, cell, true, 1)
	def := effectsTestUnit(2, BranchArmour, other, false, 2)
	def.Strength.Force = 200
	s := effectsTestSim(atk, def)
	if result, err := s.ResolveBattleAttack(atk, def); err != nil || !result.Effect.VisualOnly || result.LossTarget != 0 {
		t.Fatal("original unresolved branch changed", result, err)
	}
	a := armorTestLedger(t, s)
	copyA, copyD := *atk, *def
	preparePair(&copyA, &copyD, nil, 5, 1, 1)
	rawA, rawD := s.pairLosses(&copyA, &copyD)
	// Rifle 5 vs old armour 20 gives 5/8 of weapon casualties.
	want := (rawD*5 + 4) / 8
	result, err := s.ResolveBattleAttack(atk, def)
	if err != nil || result.Effect.VisualOnly || want <= 0 || result.LossTarget != want || result.LossAttacker != rawA {
		t.Fatal("equipped standard handler", result, want, err)
	}
	if def.Force() != uint16(200-want) || a.Units[2].Force != def.Force() || a.Lost["legacy-armor"] == 0 {
		t.Fatal("armour casualties not written to personnel and equipment")
	}
	before := a.Protection(2, def.Branch(), def.Force())
	if got := InspectBattle(s).Sides[1].Roster[0].Protection; got != before {
		t.Fatal("UI uses a different profile", got, before)
	}
}

func TestArmorCoordinatedAndRangedUseFinalWeaponDamage(t *testing.T) {
	cell := CellIndex(100)
	other, _ := cell.Neighbour(DirDown)
	supportCell, _ := cell.Neighbour(DirUp)
	atk := effectsTestUnit(1, BranchInfantry, cell, true, 1)
	def := effectsTestUnit(2, BranchArmour, other, false, 2)
	def.Strength.Force = 200
	support := effectsTestUnit(3, BranchInfantry, supportCell, false, 2)
	s := effectsTestSim(atk, def, support)
	a := armorTestLedger(t, s)
	a.Units[2].Items = map[EquipmentID]uint32{"light-tank": 20}
	copyA, copyD := *atk, *def
	preparePair(&copyA, &copyD, nil, 5, 1, 1)
	_, rawD := s.pairLosses(&copyA, &copyD)
	result, err := s.ResolveBattleAttack(atk, def)
	want := (rawD*11 + 10) / 20
	if err != nil || result.Effect.Kind != BattleEffectCoordinated || result.LossTarget != want || len(result.SupportLosses) != 1 {
		t.Fatal("coordinated protection", result, want, err)
	}

	f := rangedTestField()
	targetCell := NoCell
	for _, c := range RangedTargetCells(f, cell, 1) {
		if !Adjacent(cell, c) {
			targetCell = c
			break
		}
	}
	if !targetCell.Valid() {
		t.Fatal("no ranged target")
	}
	atk = effectsTestUnit(1, BranchArtiller, cell, true, 1)
	atk.Facing, atk.Strength.Force = 1, 600
	def = effectsTestUnit(2, BranchArmour, targetCell, false, 2)
	def.Strength.Force = 200
	s = effectsTestSim(atk, def)
	s.Field = f
	a = armorTestLedger(t, s)
	a.Units[2].Items = map[EquipmentID]uint32{"light-tank": 20}
	at, dt := s.tileOf(atk), s.tileOf(def)
	_, once := rangedLossOnce(atk, def,
		AttackValue(s.StrengthOf(atk), at, atk.Branch(), dt, def.Branch()),
		AttackValue(s.StrengthOf(def), dt, def.Branch(), at, atk.Branch())*rangedDefenceMultiplier(dt))
	base := clampLoss(once*3, def.Force())
	result, err = s.ResolveBattleAttack(atk, def)
	want = (base*7 + 5) / 10 // Piercing 20 vs armour 50.
	if err != nil || result.Effect.Kind != BattleEffectRanged || result.LossTarget != want || result.LossAttacker != 0 || want <= 0 {
		t.Fatal("ranged protection must follow the three-shot clamp", result, want, err)
	}
}

func TestArmorCavalryAndTwoSidedSymmetry(t *testing.T) {
	cell := CellIndex(100)
	other, _ := cell.Neighbour(DirDown)
	a := effectsTestUnit(1, BranchCavalry, cell, true, 1)
	b := effectsTestUnit(2, BranchInfantry, other, false, 2)
	s := effectsTestSim(a, b)
	armorTestLedger(t, s)
	copyA, copyB := *a, *b
	baseline := effectsTestSim(&copyA, &copyB)
	baseline.Opts.Armory = s.Opts.Armory.Clone()
	copyA.armory, copyB.armory = baseline.Opts.Armory, baseline.Opts.Armory
	// Repeat the existing charge formula with the same equipment losses and
	// firepower depletion, omitting only the new protection layer.
	wantA, wantB := 0, 0
	passes := int(copyB.Strength.Ability)/30 + int(copyB.Strength.F30)/40 + 1
	for i := 0; i < passes; i++ {
		rawA, rawB := baseline.pairLosses(&copyA, &copyB)
		rawB = roundDiv(rawB*2, 3)
		copyA.applyLoss(rawA)
		copyB.applyLoss(rawB)
		wantA, wantB = wantA+rawA, wantB+rawB
		if copyA.Force() <= 3 || copyB.Force() <= 3 {
			break
		}
	}
	got, err := s.ResolveBattleAttack(a, b)
	if err != nil || got.Effect.Kind != BattleEffectCavalryCharge || got.LossAttacker != wantA || got.LossTarget != wantB {
		t.Fatal("unarmoured cavalry changed", got, wantA, wantB, err)
	}
	a.Strength.Branch, b.Strength.Branch = BranchArmour, BranchArmour
	a.Strength.Force, b.Strength.Force = 200, 200
	s.Opts.Armory.Units[a.General].Items = map[EquipmentID]uint32{"armor-improved": 20}
	s.Opts.Armory.Units[b.General].Items = map[EquipmentID]uint32{"light-tank": 20}
	la, lb := s.protectedLosses(a, b, 100, 100)
	rb, ra := s.protectedLosses(b, a, 100, 100)
	if la != ra || lb != rb || lb != 100 || la != 86 {
		t.Fatal("nationality/order asymmetry", la, lb, ra, rb)
	}
}
