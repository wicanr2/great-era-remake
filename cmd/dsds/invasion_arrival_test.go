package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/game"
)

// Explicit storage/formation fixture. Declaration and cast are not normally
// triggered here; the actual game rules, battle and save readers are used.
func invasionArrivalApp(t *testing.T) *app {
	t.Helper()
	a := newArmoryApp(t, 2, 1)
	date, err := game.InvasionDateFor(a.year, a.month)
	if err != nil {
		t.Fatal(err)
	}
	a.militaryRules.DeclaredAt = date.DeclaredAt
	a.declaration = &campaignDeclaration{ID: campaignDeclarationID, Date: date, AnnouncementRead: true}
	ids := []game.GeneralID{278, 279, 280, 281, 285, 287}
	out, err := game.AssembleInvasionWave(a.world, a.armory, a.generals, date, 1, ids, game.InvasionEntryProvince)
	if err != nil {
		t.Fatal(err)
	}
	*a.tbl, *a.world, *a.armory = *out.World.Table, *out.World, *out.Armory
	a.world.Table, a.world.Opts.Armory = a.tbl, a.armory
	a.generals = out.Generals
	a.campaignActors = map[game.GeneralID]string{}
	for _, id := range ids {
		a.campaignActors[id] = campaignActorKey(id)
	}
	if _, err := a.armory.DeliverInvasionSupply(a.world, 1); err != nil {
		t.Fatal(err)
	}
	return a
}

func TestInvasionArrivalCampaignBattleSaveAndNoDuplicateImport(t *testing.T) {
	a := invasionArrivalApp(t)
	original := a.campaignState()
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	ready, _ := os.ReadFile(a.savePath)
	b := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
	if err := b.loadCampaign(ready); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a.armory, b.armory) || !reflect.DeepEqual(original, b.campaignState()) || len(b.campaignActors) != 6 {
		t.Fatal("restart lost arrivals, resources, original people or identity")
	}
	batch := b.armory.InvasionBatches[1]
	if _, err := game.AssembleInvasionWave(b.world, b.armory, b.generals, batch.Date, 1, batch.Units, game.InvasionEntryProvince); err == nil {
		t.Fatal("restart reissued the first six formations")
	}
	credit, err := b.armory.DeliverInvasionSupply(b.world, 1)
	if err != nil || credit != (game.InvasionSupply{}) || !reflect.DeepEqual(original, b.campaignState()) {
		t.Fatal("restart refilled already delivered provincial resources", credit, err)
	}
	if err := b.startDefensiveBattleWithRoster(11, 7, batch.Units); err != nil {
		t.Fatal(err)
	}
	if len(b.battle.sim.Attacker) != 6 {
		t.Fatal("first wave was replaced by original units")
	}
	b.strategyTurn = &strategyTurn{Year: b.year, Month: b.month, Player: b.playerCommander,
		Province: 7, Started: true, Remaining: max(0, b.world.CommandsFor(7)-1), JapaneseAttacks: map[game.GeneralID]uint8{}}
	for _, id := range batch.Units {
		b.strategyTurn.JapaneseAttacks[id] = 1
	}
	b.battle.japaneseBlitz = true
	protection := b.armory.Protection(batch.Units[5], game.BranchArmour, 120)
	if protection.Armor != 50000 || protection.Piercing != 50000 {
		t.Fatal("imported light tanks did not reach shared battle armor", protection)
	}
	b.pauseBattle()
	if err := b.autosave(); err != nil {
		t.Fatal(err)
	}
	active, _ := os.ReadFile(a.savePath)
	c := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
	if err := c.loadCampaign(active); err != nil {
		t.Fatal(err)
	}
	if !c.battle.paused || !c.battle.humanDefender || !reflect.DeepEqual(b.armory, c.armory) || !reflect.DeepEqual(b.campaignState(), c.campaignState()) {
		t.Fatal("active wave battle lost deployment or finite imports")
	}
	if err := c.withdrawDefenders(c.battle.retreatTargets[0]); err != nil {
		t.Fatal(err)
	}
	if err := c.autosave(); err != nil {
		t.Fatal(err)
	}
	settled, _ := os.ReadFile(a.savePath)
	d := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
	if err := d.loadCampaign(settled); err != nil || !reflect.DeepEqual(c.armory, d.armory) || !reflect.DeepEqual(c.campaignState(), d.campaignState()) {
		t.Fatal("withdrawal and restart changed the arrival ledger", err)
	}
	if output := os.Getenv("DSDS_WAVE_SAMPLE_DIR"); output != "" {
		for name, raw := range map[string][]byte{"wave-ready.gesave": ready, "wave-active.gesave": active, "wave-settled.gesave": settled} {
			if err := os.WriteFile(filepath.Join(output, name), raw, 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}
}

func TestInvasionArrivalCampaignRejectsCrossEventAndUnfundedActor(t *testing.T) {
	a := invasionArrivalApp(t)
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	good, _ := os.ReadFile(a.savePath)
	before, _ := json.Marshal(a.armory)
	snapshot := a.campaignState()
	for _, change := range []string{"other-declaration", "imported-declaration", "unfunded-actor"} {
		var v campaignV4
		if _, err := game.DecodeCampaignSaveVersion(good, &v, 4); err != nil {
			t.Fatal(err)
		}
		switch change {
		case "other-declaration":
			batch := v.Armory.InvasionBatches[1]
			batch.Date, _ = game.InvasionDateFor(26, 6)
			v.Armory.InvasionBatches[1] = batch
		case "imported-declaration":
			v.Declaration.Imported = true
		case "unfunded-actor":
			actor := v.Actors[0]
			actor.ID, actor.Key = 284, campaignActorKey(284)
			actor.Unit.General, actor.Strength.General = 284, 284
			v.Actors = append(v.Actors, actor)
			v.Armory.Units[284] = &game.UnitEquipment{Faction: 107, Branch: actor.Strength.Branch, Force: actor.Strength.Force, Items: map[game.EquipmentID]uint32{}}
		}
		bad, err := game.EncodeCampaignSaveVersion(v, nil, 4)
		if err != nil {
			t.Fatal(err)
		}
		if err := a.loadCampaign(bad); err == nil {
			t.Fatal("accepted mismatched declaration, imported event or unfunded actor", change)
		}
		after, _ := json.Marshal(a.armory)
		if !bytes.Equal(before, after) || !reflect.DeepEqual(snapshot, a.campaignState()) {
			t.Fatal("bad arrival partially replaced live campaign", change)
		}
	}
}
