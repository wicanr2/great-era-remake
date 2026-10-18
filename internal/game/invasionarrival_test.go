package game

import (
	"bytes"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"
)

func arrivalFixture(t *testing.T, stage uint8) (*AIWorld, *Armory, []General, InvasionDate) {
	t.Helper()
	sc, err := ScenarioByStage(stage)
	if err != nil {
		t.Fatal(err)
	}
	table, err := ParseTownFile(readGame(t, fmt.Sprintf("TOWN(%d).DAT", stage)))
	if err != nil {
		t.Fatal(err)
	}
	gens, err := ParseGenerals(readGame(t, fmt.Sprintf("MAN(%d).DAT", stage)), sc.Generals)
	if err != nil {
		t.Fatal(err)
	}
	w := &AIWorld{Table: table, Opts: StrengthOpts{Stage: stage}}
	for i, g := range gens {
		faction := GeneralID(0)
		if p, err := table.At(g.Province); err == nil {
			faction = p.Commander
		}
		id := GeneralID(i + 1)
		w.Units = append(w.Units, CombatUnit{General: id, Faction: faction, Province: g.Province, Cell: NoCell, NextCell: NoCell})
		w.Strengths = append(w.Strengths, StrengthInput{General: id, Faction: faction, Ability: g.AbilityA,
			Force: g.Force, Branch: g.Branch, F19: g.F19, F20: g.F20, F29: g.Stamina, F30: g.F30})
	}
	d, err := InvasionDateFor(uint16(sc.Year), sc.Month)
	if err != nil {
		t.Fatal(err)
	}
	a, err := NewArmory(w, sc.Provinces, d.DeclaredAt)
	if err != nil {
		t.Fatal(err)
	}
	w.Opts.Armory = a
	return w, a, gens, d
}

func arrivalState(t *testing.T, w *AIWorld, a *Armory, gs []General) []byte {
	t.Helper()
	b, err := json.Marshal(struct {
		Table     *ProvinceTable
		Units     []CombatUnit
		Strengths []StrengthInput
		Generals  []General
		Army      *Armory
	}{w.Table, w.Units, w.Strengths, gs, a})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestInvasionArrivalFiniteThreeWavesAndLossConservation(t *testing.T) {
	for _, stage := range []uint8{1, 2, 3} {
		t.Run(fmt.Sprint(stage), func(t *testing.T) {
			w, army, gs, date := arrivalFixture(t, stage)
			original := arrivalState(t, w, army, gs)
			initial, _ := json.Marshal(army.Initial)
			produced, _ := json.Marshal(army.Produced)
			plan, _ := FixedInvasionPlan(date.Era)
			native := len(gs)
			id := GeneralID(275)
			var supply InvasionSupply
			for number := uint8(1); number <= 3; number++ {
				for army.LastMonth < date.DeclaredAt+uint32(plan.Waves[number-1].MonthOffset) {
					if err := army.PauseMonth(w, army.LastMonth+1); err != nil {
						t.Fatal(err)
					}
				}
				ids := make([]GeneralID, len(plan.Waves[number-1].Formations))
				for i := range ids {
					ids[i] = id
					id++
				}
				before := arrivalState(t, w, army, gs)
				out, err := AssembleInvasionWave(w, army, gs, date, number, ids, ExternalInvasionSource)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(before, arrivalState(t, w, army, gs)) || !reflect.DeepEqual(gs[:native], out.Generals[:native]) {
					t.Fatal("assembly mutated source or previous formations")
				}
				for prior := GeneralID(275); prior < ids[0]; prior++ {
					if out.Generals[prior-1] != gs[prior-1] {
						t.Fatal("a previous arrival was recreated")
					}
				}
				w, army, gs = out.World, out.Armory, out.Generals
				for i, id := range ids {
					u, template := army.Units[id], plan.Waves[number-1].Formations[i]
					if u.Branch != template.Branch || u.Force != template.Force || u.Items[template.Equipment] != EquipmentNeed(template.Branch, template.Force) {
						t.Fatal("formation or equipment differs from approved catalog")
					}
					if gs[id-1].Raw != ([GeneralRecordSize]byte{}) || w.Units[id-1].Province != ExternalInvasionSource {
						t.Fatal("original record was borrowed or unit was put in player land")
					}
				}
				left, err := army.InvasionRemainingSupply(number)
				if err != nil {
					t.Fatal(err)
				}
				supply.Gold += left.Gold
				supply.Food += left.Food
				supply.Ammo += left.Ammo
				supply.Fuel += left.Fuel
				if _, err := AssembleInvasionWave(w, army, gs, date, number, ids, ExternalInvasionSource); err == nil {
					t.Fatal("restart could issue an existing wave again")
				}
			}
			if supply != (InvasionSupply{Gold: 24000, Food: 32000, Ammo: 24000, Fuel: 12000}) {
				t.Fatal("finite supply allowance changed", supply)
			}
			in, _ := json.Marshal(army.Initial)
			pr, _ := json.Marshal(army.Produced)
			if !bytes.Equal(initial, in) || !bytes.Equal(produced, pr) {
				t.Fatal("imports were disguised as initial stock or factory output")
			}
			// Real loss writer must balance against the import receipt.
			army.ApplyLoss(275, 12000, 9000)
			gs[274].Force, w.Strengths[274].Force = 9000, 9000
			if err := army.Sync(w); err != nil {
				t.Fatal(err)
			}
			if err := army.Validate(w); err != nil || army.Lost[plan.Waves[0].Formations[0].Equipment] != 30 {
				t.Fatal("imported rifles bypassed loss conservation", err)
			}
			var source struct{ Generals []General }
			if err := json.Unmarshal(original, &source); err != nil || !reflect.DeepEqual(source.Generals, gs[:native]) {
				t.Fatal("native people changed during finite imports", err)
			}
		})
	}
}

func TestInvasionArrivalRejectsInvalidAndPrematureAssemblyAtomically(t *testing.T) {
	w, army, gs, date := arrivalFixture(t, 2)
	ids := []GeneralID{275, 276, 277, 278, 279, 280}
	before := arrivalState(t, w, army, gs)
	for _, change := range []string{"native-id", "duplicate", "hq-id", "capacity", "short", "date", "future", "second-before-first", "enemy-landing", "other-province"} {
		t.Run(change, func(t *testing.T) {
			binding, d, number, at := append([]GeneralID(nil), ids...), date, uint8(1), ExternalInvasionSource
			switch change {
			case "native-id":
				binding[0] = 191
			case "duplicate":
				binding[1] = binding[0]
			case "hq-id":
				binding[0] = 300
			case "capacity":
				binding[0] = 513
			case "short":
				binding = binding[:5]
			case "date":
				d.Era = InvasionLate
			case "future":
				d, _ = InvasionDateFor(26, 8)
			case "second-before-first":
				number, binding = 2, binding[:2]
			case "enemy-landing":
				at = 11
			case "other-province":
				at = 6
			}
			if _, err := AssembleInvasionWave(w, army, gs, d, number, binding, at); err == nil {
				t.Fatal("invalid arrival accepted")
			}
			if !bytes.Equal(before, arrivalState(t, w, army, gs)) {
				t.Fatal("failed arrival partially changed the source")
			}
		})
	}
	out, err := AssembleInvasionWave(w, army, gs, date, 1, ids, InvasionEntryProvince)
	if err != nil {
		t.Fatal(err)
	}
	w, army, gs = out.World, out.Armory, out.Generals
	before = arrivalState(t, w, army, gs)
	if _, err := AssembleInvasionWave(w, army, gs, date, 2, []GeneralID{281, 282}, InvasionEntryProvince); err == nil || !bytes.Equal(before, arrivalState(t, w, army, gs)) {
		t.Fatal("month-three reinforcements arrived in month zero")
	}
	for i := 0; i < 3; i++ {
		if err := army.PauseMonth(w, army.LastMonth+1); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := AssembleInvasionWave(w, army, gs, date, 2, []GeneralID{275, 282}, InvasionEntryProvince); err == nil {
		t.Fatal("a previous actor was issued a second formation")
	}
}

func TestInvasionArrivalSupplyCapacitySpendAndRestart(t *testing.T) {
	w, army, gs, date := arrivalFixture(t, 2)
	out, err := AssembleInvasionWave(w, army, gs, date, 1, []GeneralID{275, 276, 277, 278, 279, 280}, InvasionEntryProvince)
	if err != nil {
		t.Fatal(err)
	}
	w, army, gs = out.World, out.Armory, out.Generals
	p, _ := w.Table.At(InvasionEntryProvince)
	p.Gold, p.Food, p.Ammo, p.Fuel = 65530, 65534, 65535, 65000
	credit, err := army.DeliverInvasionSupply(w, 1)
	if err != nil || credit != (InvasionSupply{Gold: 5, Food: 1, Fuel: 535}) || p.Gold != 65535 || p.Food != 65535 || p.Ammo != 65535 || p.Fuel != 65535 {
		t.Fatal("delivery overflowed or lost excess", credit, err)
	}
	zero, err := army.DeliverInvasionSupply(w, 1)
	if err != nil || zero != (InvasionSupply{}) {
		t.Fatal("a full destination received supplies twice", zero, err)
	}
	if err := army.SpendInvasionSupply(w, 1, InvasionSupply{Gold: 105, Food: 1047}); err != nil {
		t.Fatal(err)
	}
	before := arrivalState(t, w, army, gs)
	if err := army.SpendInvasionSupply(w, 1, InvasionSupply{Fuel: 1 << 31}); err == nil || !bytes.Equal(before, arrivalState(t, w, army, gs)) {
		t.Fatal("supply overdraft was accepted or partially spent")
	}
	archive, err := EncodeCampaignSaveVersion(army, nil, 4)
	if err != nil {
		t.Fatal(err)
	}
	var restored Armory
	if _, err := DecodeCampaignSaveVersion(archive, &restored, 4); err != nil || restored.Validate(w) != nil {
		t.Fatal("receipt or supply counters did not survive container restart", err)
	}
	left, _ := restored.InvasionRemainingSupply(1)
	if left != (InvasionSupply{Gold: 14400 - 5 - 105, Food: 19200 - 1 - 1047, Ammo: 14400, Fuel: 7200 - 535}) {
		t.Fatal("restart refilled external supply", left)
	}
	p.Gold, p.Food, p.Ammo, p.Fuel = 0, 0, 0, 0
	delivered, err := restored.DeliverInvasionSupply(w, 1)
	if err != nil || delivered != left {
		t.Fatal("remaining supplies were not delivered exactly once", delivered, err)
	}
	empty, _ := restored.InvasionRemainingSupply(1)
	if empty != (InvasionSupply{}) {
		t.Fatal("external supply did not exhaust")
	}
	if err := restored.SpendInvasionSupply(w, 1, InvasionSupply{Gold: 1}); err == nil {
		t.Fatal("exhausted budget supplied another battle")
	}
}

func TestInvasionArrivalRejectsCorruptReceiptsAndClonesIndependently(t *testing.T) {
	w, army, gs, date := arrivalFixture(t, 2)
	out, err := AssembleInvasionWave(w, army, gs, date, 1, []GeneralID{275, 276, 277, 278, 279, 280}, ExternalInvasionSource)
	if err != nil {
		t.Fatal(err)
	}
	w, army = out.World, out.Armory
	for _, change := range []string{"missing", "skip", "future", "foreign", "date", "duplicate", "native", "hq", "missing-unit", "spent-overflow", "double-used"} {
		t.Run(change, func(t *testing.T) {
			x := army.Clone()
			batch := x.InvasionBatches[1]
			switch change {
			case "missing":
				x.InvasionBatches = nil
			case "skip":
				delete(x.InvasionBatches, 1)
				x.InvasionBatches[2] = batch
			case "future":
				batch.ArrivedAt++
			case "foreign":
				batch.Faction = 1
			case "date":
				batch.Date.RosterMonth++
			case "duplicate":
				batch.Units[1] = batch.Units[0]
			case "native":
				batch.Units[0] = 1
			case "hq":
				batch.Units[0] = 300
			case "missing-unit":
				delete(x.Units, batch.Units[0])
			case "spent-overflow":
				batch.Spent.Gold = 1 << 31
			case "double-used":
				batch.Spent.Gold, batch.Delivered.Gold = 8000, 8000
			}
			if change != "missing" && change != "skip" {
				x.InvasionBatches[1] = batch
			}
			if err := x.Validate(w); err == nil {
				t.Fatal("corrupt import receipt explained equipment")
			}
			if err := army.Validate(w); err != nil {
				t.Fatal("clone shared mutable receipt IDs", err)
			}
		})
	}
}

func TestInvasionArrivalExternalDeploymentWithdrawalAndDelivery(t *testing.T) {
	// Controlled rule sample on the real first-scenario map. No claim that
	// unification or the normal declaration event has occurred.
	w, army, gs, date := arrivalFixture(t, 1)
	before := arrivalState(t, w, army, gs)
	out, err := AssembleInvasionWave(w, army, gs, date, 1, []GeneralID{275, 276, 277, 278, 279, 280}, ExternalInvasionSource)
	if err != nil {
		t.Fatal(err)
	}
	m := loadTestMap(t)
	province, _ := out.World.Table.At(InvasionEntryProvince)
	zone, err := m.DefenderDeployZone(InvasionEntryProvince)
	if err != nil {
		t.Fatal(err)
	}
	var attacker, defender []*Combatant
	for _, id := range out.Armory.InvasionBatches[1].Units {
		attacker = append(attacker, &Combatant{CombatUnit: out.World.Units[id-1], Strength: out.World.Strengths[id-1]})
	}
	for i, unit := range out.World.Units {
		if unit.Province != InvasionEntryProvince || unit.Faction != province.Commander || out.World.Strengths[i].Force == 0 || len(defender) >= min(UnitsPerSide, len(zone)) {
			continue
		}
		u := &Combatant{CombatUnit: unit, Strength: out.World.Strengths[i]}
		u.Cell = zone[len(defender)]
		defender = append(defender, u)
	}
	if len(defender) == 0 {
		t.Fatal("actual first scenario has no Liaoning defenders")
	}
	sim, err := NewExternalBattleSim(m, InvasionEntryProvince, attacker, defender, out.World.Opts)
	if err != nil {
		t.Fatal(err)
	}
	sim.AtCommander = province.Commander
	province.Flags |= ProvinceFlagInBattle
	targets := DefenderRetreatTargets(out.World.Table, InvasionEntryProvince, sim.AtCommander)
	if len(targets) == 0 {
		t.Fatal("actual first scenario has no friendly retreat route")
	}
	if err := ApplyDefenderWithdrawal(sim, out.World.Table, out.Generals, out.World.Units, targets[0]); err != nil {
		t.Fatal(err)
	}
	if province.Commander != 300 {
		t.Fatal("external settlement did not assign the new Japanese headquarters")
	}
	for _, id := range out.Armory.InvasionBatches[1].Units {
		if out.Generals[id-1].Province != InvasionEntryProvince || out.World.Units[id-1].Province != InvasionEntryProvince {
			t.Fatal("surviving external formation did not enter captured province")
		}
	}
	if err := out.Armory.Sync(out.World); err != nil {
		t.Fatal(err)
	}
	if _, err := out.Armory.DeliverInvasionSupply(out.World, 1); err != nil {
		t.Fatal(err)
	}
	if err := out.Armory.Validate(out.World); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(before, arrivalState(t, w, army, gs)) {
		t.Fatal("private assembly table was shared with the original world")
	}
}
