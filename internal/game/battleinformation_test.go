package game

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestBattleInformationBothSidesAndLiveEquipment(t *testing.T) {
	w, a := armoryFixture(t)
	u := func(id GeneralID) *Combatant {
		return &Combatant{CombatUnit: w.Units[int(id)-1], Strength: w.Strengths[int(id)-1]}
	}
	s := &BattleSim{Attacker: []*Combatant{u(1), u(3), nil}, Defender: []*Combatant{u(2), u(4), u(6)}, Opts: StrengthOpts{Armory: a}}
	before, _ := json.Marshal(s)
	armyBefore, _ := json.Marshal(a)
	d := InspectBattle(s)
	if d.Sides[0].Force != 12600 || d.Sides[1].Force != 12640 || d.Sides[0].Units != 2 || d.Sides[1].Branches[BranchArmour].Force != 40 || d.Sides[0].Branches[BranchArmour].Force != 0 {
		t.Fatal(d)
	}
	if !d.Sides[0].Roster[0].Tracked || d.Sides[0].Roster[0].Equipped != 120 || d.Sides[0].Roster[0].Power != 100 {
		t.Fatal(d.Sides[0].Roster)
	}
	// A returned view owns its slices/maps; callers cannot mutate the battle or ledger.
	d.Sides[0].Roster[0].Equipment[0].Count = 999
	d.Sides[0].Roster[0].Strength.Force = 1
	after, _ := json.Marshal(s)
	armyAfter, _ := json.Marshal(a)
	if !reflect.DeepEqual(before, after) || !reflect.DeepEqual(armyBefore, armyAfter) {
		t.Fatal("inspection changed battle or equipment")
	}
	a.ApplyLoss(2, 12000, 6000)
	s.Defender[0].Strength.Force = 6000
	d = InspectBattle(s)
	if d.Sides[1].Force != 6640 || d.Sides[1].Roster[0].Equipped != 60 || d.Sides[0].Force != 12600 {
		t.Fatal("stale or crossed side data")
	}
	s.Defender[1].Strength.Force = 0
	d = InspectBattle(s)
	if d.Sides[1].Units != 2 || len(d.Sides[1].Roster) != 3 || d.Sides[1].Branches[BranchArtiller].Force != 0 {
		t.Fatal("dead records must remain inspectable without counting as active")
	}
	s.Opts.Armory = nil
	if InspectBattle(s).Sides[0].Roster[0].Tracked {
		t.Fatal("invented original equipment")
	}
	if got := InspectBattle(nil); got.Sides[0].Units != 0 {
		t.Fatal(got)
	}
}
