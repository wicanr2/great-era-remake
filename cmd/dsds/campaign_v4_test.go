package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/game"
)

func TestCampaignV4MigratesOldVersionsOnlyOnSave(t *testing.T) {
	for _, version := range []int{1, 2, 3} {
		t.Run(fmt.Sprint(version), func(t *testing.T) {
			a := newArmoryApp(t, 2, 1)
			base := a.campaignState()
			var dto any
			switch version {
			case 1:
				base.Identity.Mode = "original"
				dto = base
			case 2:
				dto = campaignExpansion{Base: base, Phase: "preparation", Armory: a.armory.Clone(), View: a.armoryViewSnapshot()}
			case 3:
				// Existing locked saves have no wave metadata. Upgrading must
				// retain that state without inventing armies or replaying an event.
				a.militaryRules.DeclaredAt = uint32(a.year)*12 + uint32(a.month)
				dto = campaignOptions{Base: base, Rules: *a.militaryRules, Armory: a.armory.Clone(), View: a.armoryViewSnapshot()}
			}
			raw, err := game.EncodeCampaignSaveVersion(dto, nil, version)
			if err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(a.savePath, raw, 0o600); err != nil {
				t.Fatal(err)
			}
			b := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
			if err := b.loadCampaign(raw); err != nil {
				t.Fatal(err)
			}
			untouched, _ := os.ReadFile(a.savePath)
			if !bytes.Equal(untouched, raw) || !reflect.DeepEqual(base, b.campaignState()) {
				t.Fatal("read migrated file or changed original world")
			}
			if err := b.autosave(); err != nil {
				t.Fatal(err)
			}
			upgraded, _ := os.ReadFile(a.savePath)
			var v4 campaignV4
			if _, err := game.DecodeCampaignSaveVersion(upgraded, &v4, 4); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(base, v4.Base) || len(v4.Actors) != 0 || (v4.Declaration != nil) != (version == 3) {
				t.Fatal("upgrade changed Base, spawned actors or lost declaration")
			}
			if version == 3 && (!v4.Declaration.Imported || !v4.Declaration.AnnouncementRead) {
				t.Fatal("legacy declaration was queued again")
			}
			c := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
			if err := c.loadCampaign(upgraded); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(base, c.campaignState()) || !reflect.DeepEqual(b.armory, c.armory) {
				t.Fatal("v4 restart lost world or equipment")
			}
			if err := c.autosave(); err != nil {
				t.Fatal(err)
			}
			again, _ := os.ReadFile(a.savePath)
			if !bytes.Equal(again, upgraded) {
				t.Fatal("second save changed imported state")
			}
		})
	}
}

// This explicitly registered storage sample is not a historical roster or
// evidence of a normally triggered declaration. Original MAN records stay intact.
func campaignV4ActorSample(t *testing.T) (*app, *campaignV4) {
	t.Helper()
	a := newArmoryApp(t, 2, 1)
	a.militaryRules.DeclaredAt = uint32(a.year)*12 + uint32(a.month)
	v, err := a.campaignV4State(a.campaignState())
	if err != nil {
		t.Fatal(err)
	}
	id := game.GeneralID(280)
	g := game.General{AbilityA: 85, AbilityB: 100, AbilityC: 80, Force: 12000,
		Branch: game.BranchInfantry, F19: 80, F20: 80, Stamina: 90, F30: 90, Province: 7}
	u := game.CombatUnit{General: id, Active: true, Faction: 107, Province: 7, Cell: game.NoCell, NextCell: game.NoCell}
	in := game.StrengthInput{Ability: 85, Force: 12000, Branch: game.BranchInfantry,
		F19: 80, F20: 80, F29: 90, F30: 90, General: id, Faction: 107}
	v.Actors = []campaignActor{{Key: "jp-itagaki-seishiro", ID: id, General: g, Unit: u, Strength: in}}
	_, units, strengths, _, err := projectCampaignActors(v, game.Scenarios[1])
	if err != nil {
		t.Fatal(err)
	}
	v.Armory, err = game.NewArmory(&game.AIWorld{Table: v.Base.State.Provinces, Units: units, Strengths: strengths}, 39, uint32(a.year)*12+uint32(a.month))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := game.EncodeCampaignSaveVersion(v, nil, 4)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.loadCampaign(raw); err != nil {
		t.Fatal(err)
	}
	return a, v
}

func TestCampaignV4ExternalActorBattleWithdrawalRestart(t *testing.T) {
	a, original := campaignV4ActorSample(t)
	id := game.GeneralID(280)
	if len(a.generals) != 280 || a.campaignActors[id] != "jp-itagaki-seishiro" || !reflect.DeepEqual(a.generals[:191], original.Base.State.Generals) {
		t.Fatal("actor projection overwrote original people")
	}
	if err := a.startDefensiveBattleWithRoster(11, 7, []game.GeneralID{id}); err != nil {
		t.Fatal(err)
	}
	if len(a.battle.sim.Attacker) != 1 || a.battle.sim.Attacker[0].General != id {
		t.Fatal("registered actor was replaced by a native unit")
	}
	a.battle.sim.Attacker[0].Current--
	a.pauseBattle()
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(a.savePath)
	var saved campaignV4
	if _, err := game.DecodeCampaignSaveVersion(raw, &saved, 4); err != nil {
		t.Fatal(err)
	}
	if len(saved.Base.State.Generals) != 191 || len(saved.Base.Units) != 191 || len(saved.Actors) != 1 || saved.Actors[0].ID != id {
		t.Fatal("projection holes were serialized as original people")
	}
	b := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
	if err := b.loadCampaign(raw); err != nil {
		t.Fatal(err)
	}
	if !b.battle.paused || !b.battle.humanDefender || b.battle.sim.Unit(id).Current != a.battle.sim.Unit(id).Current || !reflect.DeepEqual(a.campaignState(), b.campaignState()) {
		t.Fatal("actor battle was redeployed or lost on restart")
	}
	if err := b.withdrawDefenders(b.battle.retreatTargets[0]); err != nil {
		t.Fatal(err)
	}
	if b.generals[id-1].Province != 11 || b.world.Units[id-1].Province != 11 || b.armory.Units[id].Force == 0 {
		t.Fatal("actor settlement lost position or equipment")
	}
	if err := b.autosave(); err != nil {
		t.Fatal(err)
	}
	settled, _ := os.ReadFile(a.savePath)
	c := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
	if err := c.loadCampaign(settled); err != nil {
		t.Fatal(err)
	}
	if c.campaignActors[id] != a.campaignActors[id] || !reflect.DeepEqual(b.campaignState(), c.campaignState()) || !reflect.DeepEqual(b.armory, c.armory) {
		t.Fatal("settled actor was moved or equipped twice")
	}
	if output := os.Getenv("DSDS_V4_SAMPLE_DIR"); output != "" {
		if err := os.WriteFile(filepath.Join(output, "v4-actor-battle.gesave"), raw, 0o600); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(output, "v4-actor-settled.gesave"), settled, 0o600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestCampaignV4RejectsCorruptRegistrationAtomically(t *testing.T) {
	a, v := campaignV4ActorSample(t)
	good, err := game.EncodeCampaignSaveVersion(v, nil, 4)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(a.campaignState())
	army, _ := json.Marshal(a.armory)
	oldTable, oldRules, oldDeclaration := a.tbl, a.militaryRules, a.declaration
	for _, change := range []string{"key", "id", "duplicate", "faction", "force", "branch", "province", "target-hole", "original-target-hole", "raw", "missing-actors", "original-prefix", "event", "date", "rule", "unread-import", "too-many", "first-stage"} {
		t.Run(change, func(t *testing.T) {
			var x campaignV4
			if _, err := game.DecodeCampaignSaveVersion(good, &x, 4); err != nil {
				t.Fatal(err)
			}
			switch change {
			case "key":
				x.Actors[0].Key = "jp-muto-nobuyoshi"
			case "id":
				x.Actors[0].ID = 191
			case "duplicate":
				x.Actors = append(x.Actors, x.Actors[0])
			case "faction":
				x.Actors[0].Unit.Faction = 1
			case "force":
				x.Actors[0].Strength.Force++
			case "branch":
				x.Actors[0].General.Branch, x.Actors[0].Strength.Branch = 2, 2
			case "province":
				x.Actors[0].General.Province, x.Actors[0].Unit.Province = 40, 40
			case "target-hole":
				x.Actors[0].Unit.TargetUnit = 200
			case "original-target-hole":
				x.Base.Units[0].TargetUnit = 200
			case "raw":
				x.Actors[0].General.Raw[0] = 1
			case "missing-actors":
				x.Actors = nil
			case "original-prefix":
				x.Base.State.Generals = append(x.Base.State.Generals, game.General{})
			case "event":
				x.Declaration.ID = "other-event"
			case "date":
				x.Declaration.Date.RosterMonth++
			case "rule":
				x.Rules.DeclaredAt = 0
			case "unread-import":
				x.Declaration.AnnouncementRead = false
			case "too-many":
				for len(x.Actors) < 11 {
					x.Actors = append(x.Actors, x.Actors[0])
				}
			case "first-stage":
				x.Base.Identity.Stage = 1
			}
			bad, err := game.EncodeCampaignSaveVersion(x, nil, 4)
			if err != nil {
				t.Fatal(err)
			}
			if err := a.loadCampaign(bad); err == nil {
				t.Fatal("accepted corrupt v4 state")
			}
			after, _ := json.Marshal(a.campaignState())
			gear, _ := json.Marshal(a.armory)
			if !bytes.Equal(before, after) || !bytes.Equal(army, gear) || a.tbl != oldTable || a.militaryRules != oldRules || a.declaration != oldDeclaration {
				t.Fatal("failed v4 load partially applied state")
			}
		})
	}
	// A modified projection hole must fail before the old file is replaced.
	prior, _ := os.ReadFile(a.savePath)
	a.world.Strengths[199].Force = 1
	if err := a.autosave(); err == nil {
		t.Fatal("unregistered runtime soldier was serialized")
	}
	after, _ := os.ReadFile(a.savePath)
	if !bytes.Equal(prior, after) {
		t.Fatal("invalid writer replaced a usable save")
	}
}

func TestCampaignV4ActorSurvivesStrategyMonthAndAttackLedger(t *testing.T) {
	a, _ := campaignV4ActorSample(t)
	month := a.month
	a.endTurn()
	if !a.activeBattle() || a.strategyTurn == nil {
		t.Fatal("declared sample did not reach the normal Japanese AI battle path")
	}
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(a.savePath)
	b := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
	if err := b.loadCampaign(raw); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a.strategyTurn.JapaneseAttacks, b.strategyTurn.JapaneseAttacks) ||
		!reflect.DeepEqual(a.campaignState(), b.campaignState()) || b.campaignActors[280] != "jp-itagaki-seishiro" {
		t.Fatal("actor registry changed the interrupted strategy turn on restart")
	}
	// Controlled battle outcomes isolate the month iterator and its runtime holes.
	// This is not a campaign victory or normal-player battle completion receipt.
	for attempts := 0; b.month == month && attempts < 50; attempts++ {
		if b.activeBattle() {
			b.battle.finishWithWinner(game.BattleSideSecond)
			if err := b.persistFinishedBattle(); err != nil {
				t.Fatal(err)
			}
		}
		b.endTurn()
	}
	if b.month != month+1 || b.strategyTurn != nil {
		t.Fatal("registered actor prevented a single month settlement")
	}
	if err := b.autosave(); err != nil {
		t.Fatal("month settlement corrupted a projection hole:", err)
	}
	if err := b.armory.Validate(b.world); err != nil {
		t.Fatal(err)
	}
}
