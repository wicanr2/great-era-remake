package game

import "testing"

func TestFixedInvasionBudgetMatchesEquipmentNeedsAndBranchCapacity(t *testing.T) {
	for _, era := range []InvasionEra{InvasionEarly, InvasionMid, InvasionLate} {
		plan, err := FixedInvasionPlan(era)
		if err != nil {
			t.Fatal(err)
		}
		var total InvasionSupply
		seen := map[uint8]bool{}
		for _, wave := range plan.Waves {
			total.Gold += wave.Supply.Gold
			total.Food += wave.Supply.Food
			total.Ammo += wave.Supply.Ammo
			total.Fuel += wave.Supply.Fuel
			need := map[EquipmentID]uint32{}
			for _, unit := range wave.Formations {
				if seen[unit.Formation] || unit.Formation < 1 || unit.Formation > 10 {
					t.Fatal("duplicate or invalid stable formation", unit)
				}
				seen[unit.Formation] = true
				model, ok := EquipmentByID(unit.Equipment)
				class, valid := EquipmentClassOf(unit.Branch)
				capacity := map[uint8]uint16{BranchInfantry: 20000, BranchArtiller: 2000, BranchArmour: 200}[unit.Branch]
				if !ok || !valid || model.Class != class || unit.Force == 0 || unit.Force > capacity {
					t.Fatal("invalid formation or class", unit)
				}
				need[unit.Equipment] += EquipmentNeed(unit.Branch, unit.Force)
			}
			if len(need) != len(wave.Equipment) {
				t.Fatal("phantom stock", wave)
			}
			for id, n := range need {
				if wave.Equipment[id] != n {
					t.Fatal("stock differs from real requirement", id, n, wave.Equipment)
				}
			}
		}
		if total != (InvasionSupply{Gold: 24000, Food: 32000, Ammo: 24000, Fuel: 12000}) || len(seen) != 10 {
			t.Fatal("finite accepted budget changed", total, seen)
		}
	}
}

func TestFixedInvasionPlanCopiesCannotCreditOrPoisonAnotherWave(t *testing.T) {
	plan, err := FixedInvasionPlan(InvasionMid)
	if err != nil {
		t.Fatal(err)
	}
	plan.Waves[0].Equipment["rifle-modern"] = 999999
	plan.Waves[0].Formations[0].Force = 1
	if plan.Waves[1].Formations[0].Force != 12000 || plan.Waves[1].Equipment["rifle-modern"] != 120 {
		t.Fatal("waves alias mutable data")
	}
	fresh, err := FixedInvasionPlan(InvasionMid)
	if err != nil || fresh.Waves[0].Equipment["rifle-modern"] != 480 || fresh.Waves[0].Formations[0].Force != 12000 {
		t.Fatal("caller mutated global catalog", fresh, err)
	}
	for _, invalid := range []InvasionEra{0, 4, 255} {
		if _, err := FixedInvasionPlan(invalid); err == nil {
			t.Fatal("invalid era accepted", invalid)
		}
	}
}
