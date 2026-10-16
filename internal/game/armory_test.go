package game

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func armoryFixture(t *testing.T) (*AIWorld, *Armory) {
	t.Helper()
	w := &AIWorld{Table: &ProvinceTable{}}
	w.Table.Province[0] = Province{Commander: 1, Gold: 9000, Iron: 180, Coal: 90, Fuel: 24, Arsenals: 2}
	w.Table.Province[1] = Province{Commander: 2, Gold: 9000, Iron: 180, Coal: 90, Fuel: 24, Arsenals: 2}
	for i, branch := range []uint8{BranchInfantry, BranchInfantry, BranchArtiller, BranchArtiller, BranchArmour, BranchArmour} {
		force := uint16(12000)
		if branch == BranchArtiller {
			force = 600
		}
		if branch == BranchArmour {
			force = 40
		}
		faction := GeneralID(i%2 + 1)
		id := GeneralID(i + 1)
		w.Units = append(w.Units, CombatUnit{General: id, Faction: faction, Province: ProvinceID(faction), Active: true})
		w.Strengths = append(w.Strengths, StrengthInput{General: id, Faction: faction, Branch: branch, Force: force, Ability: 80, F19: 60, F20: 40, F29: 80, F30: 80})
	}
	a, err := NewArmory(w, 36, 12*15+7)
	if err != nil {
		t.Fatal(err)
	}
	w.Opts.Armory = a
	return w, a
}
func armoryMonth(t *testing.T, w *AIWorld, a *Armory) {
	t.Helper()
	if err := a.SettleMonth(w, a.LastMonth+1); err != nil {
		t.Fatal(err)
	}
}
func armoryResearch(t *testing.T, w *AIWorld, a *Armory, id EquipmentID) {
	t.Helper()
	if err := a.Research(w, 1, id); err != nil {
		t.Fatal(err)
	}
	m, _ := EquipmentByID(id)
	for i := uint16(0); i < m.Months; i++ {
		armoryMonth(t, w, a)
	}
}

func TestArmoryArmorConversionAtomicAndPersonnel(t *testing.T) {
	w, a := armoryFixture(t)
	w.Strengths[0].Force = 12053
	if err := a.Sync(w); err != nil {
		t.Fatal(err)
	}
	armoryResearch(t, w, a, "light-tank")
	if err := a.Assign(w, 1, "light-tank", 2); err != nil {
		t.Fatal(err)
	}
	armoryMonth(t, w, a)
	before, _ := json.Marshal(a)
	gold := a.Resources(w, 1).Gold
	if _, err := a.ConvertToArmor(w, 1, 1, "light-tank"); err == nil {
		t.Fatal("accepted incomplete vehicle stock")
	}
	after, _ := json.Marshal(a)
	if !reflect.DeepEqual(before, after) || a.Resources(w, 1).Gold != gold {
		t.Fatal("failed conversion changed state")
	}
	// Enough resources and six paid production months supply the full formation.
	w.Table.Province[0].Iron, w.Table.Province[0].Coal, w.Table.Province[0].Fuel = 500, 200, 100
	for i := 0; i < 5; i++ {
		armoryMonth(t, w, a)
	}
	preview, err := a.ConversionPlan(w, 1, 1, "light-tank")
	if err != nil {
		t.Fatal(err)
	}
	gold = a.Resources(w, 1).Gold
	oldItems := a.Units[1].Items["legacy-rifle"]
	oldUnit := w.Units[0]
	plan, err := a.ConvertToArmor(w, 1, 1, "light-tank")
	if err != nil || plan != preview || plan.Force != 120 || plan.Reserve != 53 || plan.Groups != 12 {
		t.Fatal(plan, err)
	}
	if w.Units[0] != oldUnit || w.Strengths[0].Branch != BranchArmour || w.Strengths[0].Force != 120 || a.Resources(w, 1).Gold != gold-1000 || a.ReservePersonnel[1] != 53 || a.Nations[1].Stock["legacy-rifle"] != oldItems {
		t.Fatal("identity, personnel, gear or funds not preserved")
	}
	before, _ = json.Marshal(a)
	if _, err := a.ConvertToArmor(w, 1, 1, "light-tank"); err == nil {
		t.Fatal("duplicate conversion")
	}
	after, _ = json.Marshal(a)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("duplicate conversion mutated state")
	}
	if _, err := a.ReinforceReserve(w, 1, 2); err == nil {
		t.Fatal("enemy received reserves")
	}
	// Another friendly infantry can receive exactly the retained personnel, without free rifles.
	w.Units[1].Faction = 1
	w.Units[1].Province = 1
	if err := a.Sync(w); err != nil {
		t.Fatal(err)
	}
	items := a.Units[2].Items["legacy-rifle"]
	count, err := a.ReinforceReserve(w, 1, 2)
	if err != nil || count != 53 || w.Strengths[1].Force != 12053 || a.Units[2].Items["legacy-rifle"] != items || a.ReservePersonnel[1] != 0 {
		t.Fatal(count, err)
	}
	if err := a.Validate(w); err != nil {
		t.Fatal(err)
	}
}

func TestArmoryResearchProductionRefitConservation(t *testing.T) {
	w, a := armoryFixture(t)
	n := a.Nations[1]
	before := a.Resources(w, 1)
	if err := a.Assign(w, 1, "rifle-modern", 1); !errors.Is(err, ErrArmoryLocked) {
		t.Fatal(err)
	}
	if err := a.Research(w, 1, "rifle-modern"); err != nil {
		t.Fatal(err)
	}
	armoryMonth(t, w, a)
	if n.Known["rifle-modern"] || n.Stock["rifle-modern"] != 0 {
		t.Fatal("research created equipment")
	}
	if err := a.Research(w, 1, ""); err != nil {
		t.Fatal(err)
	}
	armoryMonth(t, w, a)
	if n.Progress["rifle-modern"] != 1 {
		t.Fatal("paused progress changed")
	}
	if err := a.Research(w, 1, "rifle-modern"); err != nil {
		t.Fatal(err)
	}
	armoryMonth(t, w, a)
	if !n.Known["rifle-modern"] || a.Resources(w, 1).Gold != before.Gold-360 || a.Power(1, BranchInfantry, 12000, false) != 100 {
		t.Fatal("unlock changed stock or firepower")
	}
	if err := a.Assign(w, 1, "rifle-modern", 2); err != nil {
		t.Fatal(err)
	}
	if err := a.Assign(w, 1, "rifle-old", 1); !errors.Is(err, ErrArmoryFactories) {
		t.Fatal(err)
	}
	armoryMonth(t, w, a)
	if n.Stock["rifle-modern"] != 48 || a.Resources(w, 1).Iron != before.Iron-48 {
		t.Fatal("production not funded")
	}
	count, err := a.Issue(w, 1, 1, "rifle-modern")
	if err != nil || count != 48 {
		t.Fatal(count, err)
	}
	if n.Stock["legacy-rifle"] != 48 || a.Units[1].Items["legacy-rifle"] != 72 || a.Power(1, BranchInfantry, 12000, false) != 110 {
		t.Fatal("partial replacement lost old equipment")
	}
	if _, err := a.Issue(w, 1, 1, "rifle-modern"); !errors.Is(err, ErrArmoryStock) {
		t.Fatal(err)
	}
	if _, err := a.Issue(w, 2, 1, "rifle-old"); !errors.Is(err, ErrArmoryUnit) {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		armoryMonth(t, w, a)
		if _, err := a.Issue(w, 1, 1, "rifle-modern"); err != nil {
			t.Fatal(err)
		}
	}
	if a.Units[1].Items["rifle-modern"] != 120 || n.Stock["rifle-modern"] != 24 || n.Stock["legacy-rifle"] != 120 {
		t.Fatal("full refit or surplus incorrect")
	}
	armoryMonth(t, w, a)
	armoryMonth(t, w, a)
	if !n.Last["rifle-modern"].Shortage || a.Resources(w, 1).Iron != 0 {
		t.Fatal("missing-material line did not stop")
	}
	if err := a.Validate(w); err != nil {
		t.Fatal(err)
	}
	snapshot, _ := json.Marshal(a)
	if err := a.SettleMonth(w, a.LastMonth); !errors.Is(err, ErrArmoryMonth) {
		t.Fatal(err)
	}
	after, _ := json.Marshal(a)
	if string(snapshot) != string(after) {
		t.Fatal("duplicate settlement changed state")
	}
}
func TestArmoryLossAndReorganization(t *testing.T) {
	w, a := armoryFixture(t)
	armoryResearch(t, w, a, "rifle-modern")
	_ = a.Assign(w, 1, "rifle-modern", 2)
	armoryMonth(t, w, a)
	_, _ = a.Issue(w, 1, 1, "rifle-modern")
	before := a.Units[1].Items["rifle-modern"]
	a.ApplyLoss(1, 12000, 6000)
	w.Strengths[0].Force = 6000
	if a.Units[1].Items["rifle-modern"] != before/2 || a.Lost["rifle-modern"] != uint64(before/2) {
		t.Fatal("casualties did not destroy proportional equipment")
	}
	stock := a.Nations[1].Stock["legacy-rifle"]
	if err := a.Sync(w); err != nil {
		t.Fatal(err)
	}
	if a.Nations[1].Stock["legacy-rifle"] != stock {
		t.Fatal("casualties refunded to stock")
	}
	w.Strengths[0].Force = 3000
	if err := a.Sync(w); err != nil {
		t.Fatal(err)
	}
	if a.Nations[1].Stock["rifle-modern"] == 0 {
		t.Fatal("noncombat excess was destroyed")
	}
	w.Strengths[0].Force = 12000
	if err := a.Sync(w); err != nil {
		t.Fatal(err)
	}
	if a.Power(1, BranchInfantry, 12000, false) >= 100 {
		t.Fatal("reinforcements received free equipment")
	}
	if err := a.Validate(w); err != nil {
		t.Fatal(err)
	}
}
func TestArmoryArmorArtilleryAndFactoryLoss(t *testing.T) {
	w, a := armoryFixture(t)
	if err := a.Research(w, 1, "armor-improved"); !errors.Is(err, ErrArmoryLocked) {
		t.Fatal(err)
	}
	armoryResearch(t, w, a, "light-tank")
	_ = a.Assign(w, 1, "light-tank", 2)
	armoryMonth(t, w, a)
	if a.Resources(w, 1).Fuel != 20 {
		t.Fatal("armor did not consume fuel")
	}
	if _, err := a.Issue(w, 1, 3, "light-tank"); !errors.Is(err, ErrArmoryUnit) {
		t.Fatal(err)
	}
	if count, err := a.Issue(w, 1, 5, "light-tank"); err != nil || count != 2 {
		t.Fatal(count, err)
	}
	w.Table.Province[0].Arsenals = 1
	armoryMonth(t, w, a)
	if a.Nations[1].Assigned() != 1 {
		t.Fatal("destroyed factory still produced")
	}
	_ = a.Assign(w, 1, "light-tank", -1)
	armoryResearch(t, w, a, "antitank")
	_ = a.Assign(w, 1, "antitank", 1)
	armoryMonth(t, w, a)
	if count, err := a.Issue(w, 1, 3, "antitank"); err != nil || count != 2 {
		t.Fatal(count, err)
	}
	if a.Units[3].Items["legacy-artillery"] != 4 || a.Nations[1].Stock["legacy-artillery"] != 2 {
		t.Fatal("artillery replacement lost equipment")
	}
	if err := a.Validate(w); err != nil {
		t.Fatal(err)
	}
}
func TestArmoryValidationAndUnchangedOriginal(t *testing.T) {
	w, a := armoryFixture(t)
	in := w.Strengths[0]
	original := Strength(in, StrengthOpts{Stage: 1})
	if got := Strength(in, StrengthOpts{Stage: 1, Armory: a}); got != original {
		t.Fatal("expansion gave free initial bonus", got, original)
	}
	armoryResearch(t, w, a, "rifle-modern")
	_ = a.Assign(w, 1, "rifle-modern", 2)
	armoryMonth(t, w, a)
	_, _ = a.Issue(w, 1, 1, "rifle-modern")
	if got := Strength(in, StrengthOpts{Stage: 1, Armory: a}); got <= original {
		t.Fatal("actual equipment has no firepower effect")
	}
	if got := Strength(in, StrengthOpts{Stage: 1}); got != original {
		t.Fatal("original formula changed")
	}
	raw, _ := json.Marshal(a)
	var saved Armory
	if err := json.Unmarshal(raw, &saved); err != nil {
		t.Fatal(err)
	}
	if err := saved.Validate(w); err != nil {
		t.Fatal(err)
	}
	saved.Nations[1].Stock["rifle-modern"]++
	if err := saved.Validate(w); !errors.Is(err, ErrArmoryState) {
		t.Fatal("fabricated stock accepted", err)
	}
	if !reflect.DeepEqual(a.Nations[1].Stock, w.Opts.Armory.Nations[1].Stock) {
		t.Fatal("world is not sharing armory")
	}
}
func TestArmoryAIUsesPaidProduction(t *testing.T) {
	w, a := armoryFixture(t)
	player := a.Nations[1]
	before, _ := json.Marshal(player)
	for i := 0; i < 9; i++ {
		if err := a.Sync(w); err != nil {
			t.Fatal(err)
		}
		a.PlanAI(w, 1)
		armoryMonth(t, w, a)
	}
	if !a.Nations[2].Known["rifle-modern"] || a.Units[2].Items["rifle-modern"] == 0 || a.Resources(w, 2).Gold >= 9000 {
		t.Fatal("AI did not research, pay and refit")
	}
	after, _ := json.Marshal(a.Nations[1])
	if string(before) != string(after) {
		t.Fatal("AI changed player procurement")
	}
	if err := a.Validate(w); err != nil {
		t.Fatal(err)
	}
}

func TestArmoryReorganizationFailureIsAtomic(t *testing.T) {
	w, a := armoryFixture(t)
	a.Nations[1].Stock["legacy-rifle"] = ArmoryStockLimit
	a.Initial["legacy-rifle"] += uint64(ArmoryStockLimit)
	if err := a.Validate(w); err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(a)
	w.Strengths[0].Force = 0
	if err := a.Sync(w); !errors.Is(err, ErrArmoryStock) {
		t.Fatal(err)
	}
	after, _ := json.Marshal(a)
	if string(before) != string(after) {
		t.Fatal("failed return partially changed equipment")
	}
}

func TestArmoryQuickAndTacticalBattleShareLossWriter(t *testing.T) {
	for _, quick := range []bool{false, true} {
		t.Run(fmt.Sprintf("quick-%t", quick), func(t *testing.T) {
			w, a := armoryFixture(t)
			w.Map = loadTestMap(t)
			w.Table.Province[25] = Province{Commander: 1, Gold: 9000, Iron: 180, Coal: 90, Fuel: 60000, Food: 60000, Ammo: 60000, Arsenals: 2}
			w.Table.Province[18] = Province{Commander: 2, Gold: 9000, Iron: 180, Coal: 90, Fuel: 60000, Food: 60000, Ammo: 60000, Arsenals: 2}
			for i := range w.Units {
				if w.Units[i].Faction == 1 {
					w.Units[i].Province = 26
				} else {
					w.Units[i].Province = 19
				}
			}
			armoryResearch(t, w, a, "rifle-modern")
			if err := a.Assign(w, 1, "rifle-modern", 1); err != nil {
				t.Fatal(err)
			}
			armoryMonth(t, w, a)
			if _, err := a.Issue(w, 1, 1, "rifle-modern"); err != nil {
				t.Fatal(err)
			}
			var forceBefore, forceAfter uint64
			for _, s := range w.Strengths {
				forceBefore += uint64(s.Force)
			}
			if quick {
				if _, err := w.ResolveAttack(26, 19); err != nil {
					t.Fatal(err)
				}
			} else {
				atk, _ := w.combatants(26, 1)
				def, _ := w.combatants(19, 2)
				field, _ := w.Map.Battlefield(19)
				placed := 0
				for i := 0; i < CellCount && placed < len(def); i++ {
					cell := CellIndex(i)
					x, y := cell.ColRow()
					if field.Owner[y][x] == 0 && !field.Tiles[y][x].Kind.Blocks() {
						def[placed].Cell = cell
						placed++
					}
				}
				def = def[:placed]
				sim, err := NewBattleSim(w.Map, 19, 26, atk, def, w.Opts)
				if err != nil {
					t.Fatal(err)
				}
				unit := sim.Unit(1)
				if unit == nil || unit.armory != a || sim.Opts.Armory != a {
					t.Fatal("tactical simulation did not attach the shared armory")
				}
				plain := sim.Opts
				plain.Armory = nil
				if sim.StrengthOf(unit) <= Strength(unit.Strength, plain) {
					t.Fatal("equipped firepower missing in tactical path")
				}
				unit.applyLoss(1200)
			}
			var losses uint64
			for _, v := range a.Lost {
				losses += v
			}
			for _, s := range w.Strengths {
				forceAfter += uint64(s.Force)
			}
			if !quick && losses == 0 {
				t.Fatal("battle did not record equipment casualties")
			}
			if quick && ((forceAfter < forceBefore && losses == 0) || (forceAfter == forceBefore && losses != 0)) {
				t.Fatal("automatic battle equipment losses disagree with actual force changes")
			}
			if err := a.Validate(w); err != nil {
				t.Fatal(err)
			}
		})
	}
}
