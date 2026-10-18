package game

import (
	"reflect"
	"testing"
)

func blitzWorld() *AIWorld {
	w := &AIWorld{Table: &ProvinceTable{}, Opts: StrengthOpts{Stage: 2}}
	for p := 1; p <= 6; p++ {
		w.Table.Province[p-1] = Province{Commander: 107, Gold: 60000, Food: 60000, Ammo: 60000, Fuel: 60000}
	}
	return w
}
func blitzUnit(w *AIWorld, p ProvinceID, faction GeneralID, force uint16) GeneralID {
	id := GeneralID(len(w.Units) + 1)
	w.Units = append(w.Units, CombatUnit{General: id, Province: p, Faction: faction, Active: true})
	w.Strengths = append(w.Strengths, StrengthInput{General: id, Faction: faction, Branch: BranchInfantry, Ability: 80, F19: 80, F20: 80, F29: 81, F30: 85, Force: force})
	return id
}

func TestJapaneseBlitzWeakPointAndMonthlyLimit(t *testing.T) {
	w := blitzWorld()
	w.Table.Province[0].Neighbours = []ProvinceID{2, 3, 0, 255}
	w.Table.Province[1].Commander = 1
	w.Table.Province[2].Commander = 1
	first := blitzUnit(w, 1, 107, 12000)
	second := blitzUnit(w, 1, 107, 10000)
	blitzUnit(w, 2, 1, 18000)
	blitzUnit(w, 3, 1, 1000)
	before := append([]CombatUnit(nil), w.Units...)
	plan := w.PlanJapaneseBlitz(1, 107, nil, 6)
	if plan.Action.Kind != AIAttack || plan.Action.To != 3 || !reflect.DeepEqual(plan.Units, []GeneralID{first, second}) || !reflect.DeepEqual(before, w.Units) {
		t.Fatalf("wrong weak point: %+v", plan)
	}
	used := map[GeneralID]uint8{first: 2, second: 1}
	plan = w.PlanJapaneseBlitz(1, 107, used, 6)
	if !reflect.DeepEqual(plan.Units, []GeneralID{second}) {
		t.Fatalf("exhausted army reused: %+v", plan)
	}
	used[second] = 2
	if w.PlanJapaneseBlitz(1, 107, used, 6).Action.Kind != AINone {
		t.Fatal("third attack allowed")
	}
	used = map[GeneralID]uint8{}
	w.Table.Province[0].Food = 0
	if w.PlanJapaneseBlitz(1, 107, used, 6).Action.Kind != AINone {
		t.Fatal("starved army attacked")
	}
}

func TestJapaneseBlitzReservesAndSelectedFormation(t *testing.T) {
	w := blitzWorld()
	w.Table.Province[0].Neighbours = []ProvinceID{2}
	w.Table.Province[1].Commander = 1
	for i := 0; i < 11; i++ {
		blitzUnit(w, 1, 107, 1000+uint16(i)*1000)
	}
	for i := 0; i < 11; i++ {
		blitzUnit(w, 2, 1, 20000)
	}
	if w.PlanJapaneseBlitz(1, 107, nil, 6).Action.Kind != AINone {
		t.Fatal("ignored defensive reserves")
	}
	for i := 11; i < 22; i++ {
		w.Strengths[i].Force = 100
	}
	plan := w.PlanJapaneseBlitz(1, 107, nil, 6)
	if len(plan.Units) != 10 || plan.Units[0] != 11 || plan.Units[9] != 2 {
		t.Fatalf("formation not strongest ten: %+v", plan)
	}
	for _, id := range plan.Units {
		if _, _, err := w.SelectedCombatants(1, 107, []GeneralID{id}); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := w.SelectedCombatants(1, 107, []GeneralID{11, 11}); err == nil {
		t.Fatal("duplicate army accepted")
	}
	if _, _, err := w.SelectedCombatants(1, 107, []GeneralID{12}); err == nil {
		t.Fatal("enemy army accepted")
	}
}

func TestJapaneseBlitzConcentratesThroughFriendlyTerritory(t *testing.T) {
	w := blitzWorld()
	w.Table.Province[0].Neighbours = []ProvinceID{2, 5}
	w.Table.Province[1].Neighbours = []ProvinceID{1, 3}
	w.Table.Province[2].Neighbours = []ProvinceID{2, 4}
	w.Table.Province[3].Commander = 1
	w.Table.Province[4].Commander = 0
	strong := blitzUnit(w, 1, 107, 10000)
	keep := blitzUnit(w, 1, 107, 1000)
	blitzUnit(w, 3, 107, 1000)
	blitzUnit(w, 4, 1, 10000)
	plan := w.PlanJapaneseBlitz(1, 107, nil, 6)
	if plan.Action.Kind != AITransfer || plan.Action.To != 2 || !reflect.DeepEqual(plan.Units, []GeneralID{strong}) {
		t.Fatalf("wrong concentration route: %+v", plan)
	}
	w.Table.Province[1].Flags |= provinceFlagTransferBlocked
	if w.PlanJapaneseBlitz(1, 107, nil, 6).Action.Kind != AINone {
		t.Fatal("concentration crossed a blocked province")
	}
	if err := w.ApplyBlitzTransfer(plan, 107); err == nil {
		t.Fatal("blocked concentration destination was accepted")
	}
	w.Table.Province[1].Flags &^= provinceFlagTransferBlocked
	if err := w.ApplyBlitzTransfer(plan, 107); err != nil {
		t.Fatal(err)
	}
	if w.Units[strong-1].Province != 2 || w.Units[keep-1].Province != 1 {
		t.Fatal("concentration lost garrison")
	}
	w.Table.Province[1].Commander = 1
	if w.PlanJapaneseBlitz(1, 107, nil, 6).Action.Kind == AITransfer {
		t.Fatal("walked through enemy province")
	}
	before := append([]CombatUnit(nil), w.Units...)
	if err := w.ApplyBlitzTransfer(BlitzPlan{AIAction{Kind: AITransfer, From: 1, To: 5}, []GeneralID{keep}}, 107); err == nil || !reflect.DeepEqual(before, w.Units) {
		t.Fatal("invalid transfer mutated world")
	}
}
