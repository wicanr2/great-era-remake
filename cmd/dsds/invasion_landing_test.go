package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/game"
)

// Explicit post-reconquest snapshot, not a normal unification/declaration test.
func externalLandingApp(t *testing.T) *app {
	t.Helper()
	a := newArmoryApp(t, 2, 1)
	for i := 0; i < len(a.world.Units) && a.world.RosterOf(7).Len() > 0; i++ {
		a.world.ApplyTransfer(7, game.TransferFirst, 8)
	}
	p, _ := a.tbl.At(7)
	p.Capture(1)
	for i := 0; i < 6 && a.world.RosterOf(11).Len() > 1; i++ {
		a.world.ApplyTransfer(11, game.TransferFirst, 7)
	}
	a.syncExpansionGenerals()
	if err := a.armory.Sync(a.world); err != nil {
		t.Fatal(err)
	}
	date, _ := game.InvasionDateFor(a.year, a.month)
	a.militaryRules.DeclaredAt = date.DeclaredAt
	a.declaration = &campaignDeclaration{ID: campaignDeclarationID, Date: date, AnnouncementRead: true}
	appendExternalWave(t, a, 1, []game.GeneralID{278, 279, 280, 281, 285, 287})
	if err := a.startExternalInvasionBattle(); err != nil {
		t.Fatal(err)
	}
	for _, u := range a.battle.sim.Defender {
		if u.General == 160 || u.General == 183 || a.world.Units[u.General-1].Faction != 1 || !a.world.Units[u.General-1].Active {
			t.Fatal("inactive Japanese record was turned into a Chinese defender")
		}
	}
	a.pauseBattle()
	return a
}

func appendExternalWave(t *testing.T, a *app, number uint8, ids []game.GeneralID) {
	t.Helper()
	out, err := game.AssembleInvasionWave(a.world, a.armory, a.generals, a.declaration.Date, number, ids, game.ExternalInvasionSource)
	if err != nil {
		t.Fatal(err)
	}
	*a.tbl, *a.world, *a.armory = *out.World.Table, *out.World, *out.Armory
	a.world.Table, a.world.Opts.Armory = a.tbl, a.armory
	a.generals = out.Generals
	if a.campaignActors == nil {
		a.campaignActors = map[game.GeneralID]string{}
	}
	for _, id := range ids {
		a.campaignActors[id] = campaignActorKey(id)
	}
}

func restartLanding(t *testing.T, a *app, export string) *app {
	t.Helper()
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	if output := os.Getenv("DSDS_RETRY_SAMPLE_DIR"); output != "" && export != "" {
		if err := os.WriteFile(filepath.Join(output, export), raw, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	b := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
	if err := b.loadCampaign(raw); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a.armory, b.armory) || !reflect.DeepEqual(a.campaignState(), b.campaignState()) {
		t.Fatal("offshore restart reset army, supply, attempt or battlefield")
	}
	return b
}

func TestInvasionLandingCampaignThreeFailedAttemptsPersist(t *testing.T) {
	a := externalLandingApp(t)
	a = restartLanding(t, a, "retry-active.gesave")
	panel := a.battlePanelData()
	if panel.Attacker.Gold != 14400 || panel.Attacker.Food != 19200 || panel.Attacker.Leader != 107 || panel.Defender.Leader != 1 {
		t.Fatal("external panel did not show actual offshore supply and both factions", panel)
	}
	var goldSpent, foodSpent uint32
	for number := uint8(1); number <= 3; number++ {
		b := a.battle
		if number > 1 {
			if err := a.startExternalInvasionBattle(); err != nil {
				t.Fatal(err)
			}
			b = a.battle
		}
		before := b.supAtk
		game.TurnUpkeep(&b.supAtk, &b.supDef)
		goldSpent += uint32(before.Gold - b.supAtk.Gold)
		foodSpent += uint32(before.Food - b.supAtk.Food)
		if number == 1 {
			b.sim.Attacker[0].Strength.Force -= 1000
			b.forceDirty = true
			if err := a.syncBattleForces(); err != nil {
				t.Fatal(err)
			}
		}
		b.turn = b.turnLimit()
		b.finishByTurnLimit()
		if err := a.persistFinishedBattle(); err != nil {
			t.Fatal(err)
		}
		a.pauseBattle()
		var gold, food uint32
		for _, batch := range a.armory.InvasionBatches {
			gold += batch.Spent.Gold
			food += batch.Spent.Food
		}
		if gold != goldSpent || food != foodSpent {
			t.Fatal("finite supplies not charged exactly once")
		}
		var file string
		if number == 1 {
			file = "retry-waiting.gesave"
		}
		if number == 3 {
			file = "retry-final.gesave"
		}
		a = restartLanding(t, a, file)
		beforeSave, _ := os.ReadFile(a.savePath)
		if err := a.startExternalInvasionBattle(); err == nil {
			t.Fatal("retry without fresh reinforcement")
		}
		if number == 3 {
			if a.armory.Landing.Attempts[2].Outcome != game.LandingWithdrawn {
				t.Fatal("last wave did not withdraw")
			}
			for _, actor := range a.campaignActors {
				if actor == "" {
					t.Fatal("lost actor identity")
				}
			}
			break
		}
		a.battle = nil
		target := a.declaration.Date.DeclaredAt + uint32(number)*3
		for a.armory.LastMonth < target {
			if err := a.armory.PauseMonth(a.world, a.armory.LastMonth+1); err != nil {
				t.Fatal(err)
			}
			a.year, a.month = uint16((a.armory.LastMonth-1)/12), uint8((a.armory.LastMonth-1)%12+1)
			if err := a.startExternalInvasionBattle(); err == nil {
				t.Fatal("month reset retry gate")
			}
		}
		if !bytes.Equal(beforeSave, func() []byte { x, _ := os.ReadFile(a.savePath); return x }()) {
			t.Fatal("failed retry changed disk save")
		}
		ids := []game.GeneralID{284, 289}
		if number == 2 {
			ids = []game.GeneralID{283, 288}
		}
		appendExternalWave(t, a, number+1, ids)
		if a.generals[277].Force != 11000 {
			t.Fatal("reinforcements healed old losses")
		}
	}
}

func TestInvasionLandingCampaignWithdrawalAndAtomicFailure(t *testing.T) {
	a := externalLandingApp(t)
	a = restartLanding(t, a, "")
	b := a.battle
	world, army := a.campaignState(), a.armory.Clone()
	b.supAtk.Gold = -1
	if err := a.withdrawDefenders(b.retreatTargets[0]); err == nil || !reflect.DeepEqual(army, a.armory) ||
		!reflect.DeepEqual(world.State.Provinces, a.campaignState().State.Provinces) {
		t.Fatal("bad supply partially settled")
	}
	b.supAtk.Gold = 14400
	if err := a.withdrawDefenders(b.retreatTargets[0]); err != nil {
		for _, u := range b.sim.Defender {
			t.Logf("defender %d combat=%+v world=%+v general=%+v", u.General, u, a.world.Units[u.General-1], a.generals[u.General-1])
		}
		t.Fatal(err)
	}
	if a.armory.Landing.Attempts[0].Outcome != game.LandingLanded {
		t.Fatal("withdrawal did not settle landing")
	}
	a = restartLanding(t, a, "retry-landed.gesave")
	p, _ := a.tbl.At(7)
	if p.Commander != 107 {
		t.Fatal("landing did not capture through settlement")
	}
	before := a.armory.Clone()
	if err := a.applyBattleSettlement(); err != nil || !reflect.DeepEqual(before, a.armory) {
		t.Fatal("settlement repeated", err)
	}
	if err := a.startExternalInvasionBattle(); err == nil {
		t.Fatal("landed army attacked offshore again")
	}
}

func TestInvasionLandingCampaignRejectsCorruptOffshoreStateAtomically(t *testing.T) {
	a := externalLandingApp(t)
	a = restartLanding(t, a, "")
	good, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	before, army := a.campaignState(), a.armory.Clone()
	for _, change := range []string{"missing-front", "missing-battle", "missing-external-marker", "unknown-through", "foreign-roster", "finished-active", "moved-offshore", "supply-overflow", "stale-phase"} {
		var v campaignV4
		if _, err := game.DecodeCampaignSaveVersion(good, &v, 4); err != nil {
			t.Fatal(err)
		}
		switch change {
		case "missing-front":
			v.Armory.Landing = nil
		case "missing-battle":
			v.Base.Battle = nil
		case "missing-external-marker":
			v.Base.Battle.ExternalInvasion = false
		case "unknown-through":
			v.Armory.Landing.Attempts[0].Through = 2
		case "foreign-roster":
			v.Armory.Landing.Attempts[0].Units[0] = 1
		case "finished-active":
			v.Base.Battle.Finished = true
		case "moved-offshore":
			v.Actors[0].General.Province, v.Actors[0].Unit.Province = 7, 7
		case "supply-overflow":
			v.Base.Battle.SupplyAttacker.Gold = 14401
		case "stale-phase":
			v.Armory.Landing.Attempts[0].Outcome, v.Armory.Landing.Attempts[0].FinishedAt = game.LandingRepulsed, v.Armory.LastMonth
		}
		bad, err := game.EncodeCampaignSaveVersion(v, nil, 4)
		if err != nil {
			t.Fatal(err)
		}
		if err := a.loadCampaign(bad); err == nil {
			t.Fatal("accepted corrupt offshore state", change)
		}
		if !reflect.DeepEqual(before, a.campaignState()) || !reflect.DeepEqual(army, a.armory) {
			t.Fatal("bad save partially changed live world", change)
		}
	}
}

func TestInvasionLandingEmptyDefenceDoesNotCreateUnplayableBattle(t *testing.T) {
	a := externalLandingApp(t)
	a.battle, a.armory.Landing = nil, nil
	p, _ := a.tbl.At(7)
	p.Flags &^= game.ProvinceFlagInBattle
	for i := 0; i < len(a.world.Units) && a.world.RosterOf(7).Len() > 0; i++ {
		a.world.ApplyTransfer(7, game.TransferFirst, 11)
	}
	a.syncExpansionGenerals()
	before, army := a.campaignState(), a.armory.Clone()
	if err := a.startExternalInvasionBattle(); err == nil {
		t.Fatal("empty defence opened a battle with no selectable human unit")
	}
	if !reflect.DeepEqual(before, a.campaignState()) || !reflect.DeepEqual(army, a.armory) {
		t.Fatal("failed deployment changed state")
	}
}
