package game

import (
	"encoding/json"
	"reflect"
	"testing"
)

func withdrawalFixture(t *testing.T, external bool) (*BattleSim, *ProvinceTable, []General, []CombatUnit, ProvinceID) {
	t.Helper()
	m, atk, def := externalBattleFixture(t, 10)
	at, from := ProvinceID(7), ExternalInvasionSource
	if !external {
		at, from = 11, 7
		zone, err := m.DefenderDeployZone(at)
		if err != nil {
			t.Fatal(err)
		}
		for i, u := range def {
			u.Cell = zone[i]
		}
	}
	var sim *BattleSim
	var err error
	if external {
		sim, err = NewExternalBattleSim(m, at, atk, def, StrengthOpts{})
	} else {
		sim, err = NewBattleSim(m, at, from, atk, def, StrengthOpts{})
	}
	if err != nil {
		t.Fatal(err)
	}
	sim.AtCommander = 1
	table := &ProvinceTable{}
	target := &table.Province[at-1]
	target.Commander, target.Flags = 1, ProvinceFlagInBattle|ProvinceFlagTaxed
	target.Gold, target.Food = 1234, 5678
	target.Neighbours, err = m.Neighbours(at)
	if err != nil {
		t.Fatal(err)
	}
	to := ProvinceID(0)
	for _, id := range target.Neighbours {
		table.Province[id-1].Commander = 275
		if id != from && to == 0 {
			to = id
		}
	}
	if to == 0 {
		t.Fatal("real map needs retreat neighbour")
	}
	table.Province[to-1].Commander = 1
	gens := make([]General, 284)
	units := make([]CombatUnit, 284)
	for i := range gens {
		gens[i].Raw[10] = byte(i)
		units[i].General = GeneralID(i + 1)
	}
	for _, u := range sim.Attacker {
		gens[u.General-1].Force = u.Force()
		sourceProvince := from
		if external {
			sourceProvince = 0
		}
		gens[u.General-1].Province = sourceProvince
		units[u.General-1] = u.CombatUnit
		units[u.General-1].Province = sourceProvince
	}
	// Four surviving strategic reserves are deliberately outside tactical slots.
	for i := 0; i < 14; i++ {
		gens[i].Province, gens[i].Force = at, 12000
		units[i].Province, units[i].Faction, units[i].Active = at, 1, true
	}
	gens[14].Province, gens[14].Force = at, 0
	units[14].Province, units[14].Faction = at, 1
	return sim, table, gens, units, to
}

func TestDefenderWithdrawalMovesEntireLivingGarrison(t *testing.T) {
	for _, external := range []bool{false, true} {
		sim, table, gens, units, to := withdrawalFixture(t, external)
		before := append([]General(nil), gens...)
		gold, food := table.Province[sim.At-1].Gold, table.Province[sim.At-1].Food
		if err := ApplyDefenderWithdrawal(sim, table, gens, units, to); err != nil {
			t.Fatal(err)
		}
		p := table.Province[sim.At-1]
		if p.Commander != 275 || p.InBattle() || p.Flags&ProvinceFlagTaxed == 0 || p.Gold != gold || p.Food != food {
			t.Fatal("withdrawal captured incorrectly or moved province resources")
		}
		for i := 0; i < 14; i++ {
			if gens[i].Province != to || units[i].Province != to || units[i].Faction != 1 {
				t.Fatal("surviving garrison or reserve left behind", i)
			}
		}
		if gens[14].Province != sim.At {
			t.Fatal("zero-force unit was moved as a survivor")
		}
		for i := range gens {
			if gens[i].Raw != before[i].Raw || gens[i].Force != before[i].Force {
				t.Fatal("withdrawal changed raw bytes or force")
			}
		}
		for _, u := range sim.Attacker {
			if gens[u.General-1].Province != sim.At || units[u.General-1].Province != sim.At {
				t.Fatal("attacker survivor did not occupy target")
			}
		}
		t.Logf("external=%t: all 14 surviving defenders including 4 reserves moved %d -> %d; both rosters and raw bytes preserved", external, sim.At, to)
	}
}

func TestDefenderWithdrawalFailuresAreAtomic(t *testing.T) {
	for _, kind := range []string{"enemy-destination", "destination-in-battle", "non-neighbour", "invalid-attacker-id", "unsynchronized-defender", "attacker-identity", "bad-reserve-index", "changed-owner", "not-in-battle", "changed-source", "dead-attacker"} {
		sim, table, gens, units, to := withdrawalFixture(t, false)
		switch kind {
		case "enemy-destination":
			table.Province[to-1].Commander = 275
		case "destination-in-battle":
			table.Province[to-1].Flags |= ProvinceFlagInBattle
		case "non-neighbour":
			to = 39
			table.Province[38].Commander = 1
		case "invalid-attacker-id":
			sim.Attacker[5].General = 285
		case "attacker-identity":
			units[sim.Attacker[0].General-1].Faction = 1
		case "unsynchronized-defender":
			gens[0].Force--
		case "bad-reserve-index":
			units[13].General = 0
		case "changed-owner":
			table.Province[sim.At-1].Commander = 275
		case "not-in-battle":
			table.Province[sim.At-1].Flags &^= ProvinceFlagInBattle
		case "changed-source":
			table.Province[sim.From-1].Commander = 1
		case "dead-attacker":
			for _, u := range sim.Attacker {
				u.Strength.Force = 0
			}
		}
		snapshot := func() []byte {
			b, err := json.Marshal([]any{table, gens, units, sim.Attacker, sim.Defender})
			if err != nil {
				t.Fatal(err)
			}
			return b
		}
		before := snapshot()
		if err := ApplyDefenderWithdrawal(sim, table, gens, units, to); err == nil {
			t.Fatal("invalid withdrawal accepted", kind)
		}
		if !reflect.DeepEqual(before, snapshot()) {
			t.Fatal("failed withdrawal partly changed world", kind)
		}
	}
}

func TestDefenderRetreatTargetsExcludeEnemyNeutralAndBattle(t *testing.T) {
	table := &ProvinceTable{}
	table.Province[6].Neighbours = []ProvinceID{5, 9, 10, 11, 5, 0, 40, 255, 7}
	table.Province[4].Commander = 1
	table.Province[8].Commander = 275
	table.Province[10].Commander = 1
	table.Province[10].Flags = ProvinceFlagInBattle
	if got := DefenderRetreatTargets(table, 7, 1); !reflect.DeepEqual(got, []ProvinceID{5}) {
		t.Fatal(got)
	}
	for _, got := range [][]ProvinceID{DefenderRetreatTargets(nil, 7, 1), DefenderRetreatTargets(table, 40, 1), DefenderRetreatTargets(table, 7, 0), DefenderRetreatTargets(table, 7, 2)} {
		if len(got) != 0 {
			t.Fatal("invented retreat candidate", got)
		}
	}
}
