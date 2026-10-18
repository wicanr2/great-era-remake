package game

import (
	"bytes"
	"encoding/json"
	"reflect"
	"testing"
)

func externalBattleFixture(t *testing.T, guards int) (*Map, []*Combatant, []*Combatant) {
	t.Helper()
	m := loadTestMap(t)
	plan, err := FixedInvasionPlan(InvasionEarly)
	if err != nil {
		t.Fatal(err)
	}
	var atk, def []*Combatant
	for i, f := range plan.Waves[0].Formations {
		u := mkUnit(GeneralID(275+i), 275, f.Branch, f.Force)
		u.Strength.Ability, u.Strength.F19, u.Strength.F20 = f.Ability, f.Skill, f.Arming
		u.Strength.F29, u.Strength.F30 = f.Stamina, f.Morale
		atk = append(atk, u)
	}
	zone, err := m.DefenderDeployZone(7)
	if err != nil {
		t.Fatal(err)
	}
	if len(zone) < guards {
		t.Fatal("real map cannot hold fixture guards")
	}
	for i := 0; i < guards; i++ {
		u := mkUnit(GeneralID(i+1), 1, BranchInfantry, 12000)
		u.Cell = zone[i]
		def = append(def, u)
	}
	return m, atk, def
}

func battleInputBytes(t *testing.T, a, d []*Combatant) []byte {
	t.Helper()
	b, err := json.Marshal([2][]*Combatant{a, d})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestExternalBattleRealLiaoningDeploymentAndRestore(t *testing.T) {
	for _, guards := range []int{0, 1, 6, 10} {
		t.Run(string(rune('A'+guards)), func(t *testing.T) {
			m, atk, def := externalBattleFixture(t, guards)
			before := battleInputBytes(t, atk, def)
			sim, err := NewExternalBattleSim(m, 7, atk, def, StrengthOpts{Stage: 1})
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(before, battleInputBytes(t, atk, def)) {
				t.Fatal("constructor modified caller units")
			}
			want := []CellIndex{194, 193, 192, 191, 181, 180}
			for i, u := range sim.Attacker {
				if u.Cell != want[i] || !u.Attacking || sim.Occ[u.Cell] != u.General || sim.Unit(u.General) != u {
					t.Fatalf("bad deployment %+v", u)
				}
			}
			for i, u := range sim.Defender {
				if u.Cell != def[i].Cell || u.Force() != def[i].Force() || u.Attacking {
					t.Fatal("defender displaced or erased")
				}
			}
			// Persist actual combatant values, including a non-default movement
			// balance and dead off-board unit. Restore must not deploy again.
			sim.Attacker[0].Current = 3
			sim.Attacker[1].Strength.Force = 0
			sim.Sweep()
			b := battleInputBytes(t, sim.Attacker, sim.Defender)
			var saved [2][]*Combatant
			if err := json.Unmarshal(b, &saved); err != nil {
				t.Fatal(err)
			}
			restored, err := RestoreExternalBattleSim(m, 7, 1, saved[0], saved[1], sim.Opts, 284)
			if err != nil {
				t.Fatal(err)
			}
			if !bytes.Equal(b, battleInputBytes(t, restored.Attacker, restored.Defender)) {
				t.Fatal("restore reset combatant state")
			}
			if restored.Occ != sim.Occ {
				t.Fatal("restore changed occupancy")
			}
			if _, err := NewBattleSim(m, 7, 40, nil, nil, sim.Opts); err == nil {
				t.Fatal("ordinary constructor accepted external source")
			}
			if _, err := RestoreBattleSim(m, 7, 40, 1, saved[0], saved[1], sim.Opts, 284); err == nil {
				t.Fatal("ordinary restore accepted external source")
			}
			t.Logf("real province 7: %d attackers / %d defenders; deployment, no-input-mutation and restore passed", len(sim.Attacker), guards)
		})
	}
}

func TestExternalBattleErrorsDoNotMutateUnits(t *testing.T) {
	for _, kind := range []string{"andong", "duplicate", "too-many", "blocked", "occupied-entry", "nil-map"} {
		t.Run(kind, func(t *testing.T) {
			m, atk, def := externalBattleFixture(t, 1)
			at := ProvinceID(7)
			switch kind {
			case "andong":
				at = 9
			case "duplicate":
				atk[0].General = def[0].General
			case "too-many":
				for len(atk) < 11 {
					u := mkUnit(GeneralID(280+len(atk)), 275, BranchInfantry, 12000)
					atk = append(atk, u)
				}
			case "blocked":
				bf, _ := m.Battlefield(7)
				found := false
				for i := 0; i < CellCount; i++ {
					c := CellIndex(i)
					x, y := c.ColRow()
					if bf.Tiles[y][x].Kind.Blocks() {
						def[0].Cell = c
						found = true
						break
					}
				}
				if !found {
					t.Fatal("expected blocked map tile")
				}
			case "occupied-entry":
				zone, err := externalInvasionZone(m, 7)
				if err != nil {
					t.Fatal(err)
				}
				def = nil
				for i, c := range zone {
					u := mkUnit(GeneralID(i+1), 1, BranchInfantry, 12000)
					u.Cell = c
					def = append(def, u)
				}
			case "nil-map":
				m = nil
			}
			before := battleInputBytes(t, atk, def)
			if _, err := NewExternalBattleSim(m, at, atk, def, StrengthOpts{}); err == nil {
				t.Fatal("invalid battle accepted")
			}
			if !bytes.Equal(before, battleInputBytes(t, atk, def)) {
				t.Fatal("failed deployment changed caller units")
			}
		})
	}
}

func TestExternalSettlementPreservesDefendersAndIsAtomic(t *testing.T) {
	for _, kind := range []string{"attacker-win", "defender-win", "draw", "retreat", "invalid-id", "wrong-faction", "ordinary-reject"} {
		t.Run(kind, func(t *testing.T) {
			m, atk, def := externalBattleFixture(t, 6)
			sim, err := NewExternalBattleSim(m, 7, atk, def, StrengthOpts{})
			if err != nil {
				t.Fatal(err)
			}
			table := &ProvinceTable{}
			table.Province[6].Commander = 1
			table.Province[6].Flags = ProvinceFlagInBattle | ProvinceFlagTaxed
			gens := make([]General, 284)
			units := make([]CombatUnit, 284)
			for i := range gens {
				gens[i].Raw[10] = byte(i)
				gens[i].Province = 2
				units[i].Province = 2
			}
			before := append([]General(nil), gens...)
			s := BattleSettlement{Kind: BattleSettlementDecisive, Winner: BattleSideFirst}
			switch kind {
			case "defender-win":
				s.Winner = BattleSideSecond
			case "draw":
				s.Kind = BattleSettlementTurnLimitDraw
				s.Winner = BattleSideNone
			case "retreat":
				s = BattleSettlementImmediateRetreat()
			case "invalid-id":
				sim.Attacker[5].General = 285
			case "wrong-faction":
				sim.Attacker[5].Faction = 1
			}
			if kind == "ordinary-reject" {
				err = ApplyBattleSettlement(sim, table, gens, units, s)
			} else {
				err = ApplyExternalBattleSettlement(sim, table, gens, units, 275, s)
			}
			bad := kind == "invalid-id" || kind == "wrong-faction" || kind == "ordinary-reject"
			if bad {
				if err == nil || table.Province[6].Commander != 1 || table.Province[6].Flags != ProvinceFlagInBattle|ProvinceFlagTaxed || !reflect.DeepEqual(before, gens) {
					t.Fatal("invalid settlement partly changed world", err)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if kind == "attacker-win" {
				if table.Province[6].Commander != 275 {
					t.Fatal("winning invasion did not capture")
				}
				for _, u := range sim.Attacker {
					if gens[u.General-1].Province != 7 || units[u.General-1].Province != 7 {
						t.Fatal("survivor not moved")
					}
				}
			} else if table.Province[6].Commander != 1 {
				t.Fatal("non-winning invasion captured province")
			}
			for i := 0; i < 274; i++ {
				if gens[i] != before[i] {
					t.Fatal("settlement erased a legacy defender")
				}
			}
			for i := range gens {
				if gens[i].Raw != before[i].Raw {
					t.Fatal("raw legacy bytes changed")
				}
			}
		})
	}
}
