package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/game"
)

func TestOriginalCampaignMilitaryPauseRestartAndResume(t *testing.T) {
	a := &app{m: campaignTestMap(t), dataRead: campaignTestRead, initialSeed: 1, savePath: filepath.Join(t.TempDir(), "options.gesave")}
	if err := a.startCampaign(2, 1); err != nil {
		t.Fatal(err)
	}
	before := a.campaignState()
	if a.militaryEnabled() || a.armory != nil {
		t.Fatal("disabled opening unexpectedly has military equipment")
	}
	if err := a.setMilitaryEnabled(true); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(before, a.campaignState()) || a.campaign.Mode != "original" {
		t.Fatal("switch changed world, RNG, budget or campaign origin")
	}
	if err := a.armory.Research(a.world, 1, "rifle-modern"); err != nil {
		t.Fatal(err)
	}
	if err := a.armory.Assign(a.world, 1, "rifle-old", 1); err != nil {
		t.Fatal(err)
	}
	// State-level pause/resume test before the map declaration safe point.
	a.screen = screenArmory
	a.endTurn()
	n := a.armory.Nations[1]
	progress, stock := n.Progress["rifle-modern"], n.Stock["rifle-old"]
	if progress != 1 || stock != 30 {
		t.Fatal("initial active month", progress, stock)
	}
	army := a.armory
	if err := a.setMilitaryEnabled(false); err != nil {
		t.Fatal(err)
	}
	a.endTurn()
	a.endTurn()
	// Sync replaces the nested nation objects after validating a clone.
	// Read the current ledger, not the nation pointer from before the pause.
	n = a.armory.Nations[1]
	if a.armory != army || n.Progress["rifle-modern"] != progress || n.Stock["rifle-old"] != stock || len(n.Last) != 0 || a.armory.LastMonth != uint32(a.year)*12+uint32(a.month) {
		t.Fatal("paused months produced equipment, research or reset the ledger")
	}
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(a.savePath)
	version, err := game.CampaignSaveVersion(data)
	if err != nil || version != game.CampaignSaveCurrentVersion {
		t.Fatal(version, err)
	}
	b := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
	if err := b.loadCampaign(data); err != nil {
		t.Fatal(err)
	}
	if b.militaryEnabled() || !reflect.DeepEqual(a.armory, b.armory) || !reflect.DeepEqual(a.campaignState(), b.campaignState()) {
		t.Fatal("disabled restart lost world or preserved equipment")
	}
	b.openArmory()
	if b.screen == screenArmory {
		t.Fatal("disabled armory opened")
	}
	if err := b.setMilitaryEnabled(true); err != nil {
		t.Fatal(err)
	}
	b.screen = screenArmory
	b.endTurn()
	if !b.armory.Nations[1].Known["rifle-modern"] || b.armory.Nations[1].Stock["rifle-old"] != 60 {
		t.Fatal("resume caught up paused months or lost progress")
	}
}

func TestMilitaryLockSaveAndMalformedLoadAreAtomic(t *testing.T) {
	a := newArmoryApp(t, 2, 1)
	// State-level save test only. This is not evidence of a playable invasion.
	a.militaryRules.DeclaredAt = uint32(a.year)*12 + uint32(a.month)
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(a.savePath)
	b := &app{m: a.m, dataRead: campaignTestRead}
	if err := b.loadCampaign(data); err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(b.armory)
	rules, snapshot := *b.militaryRules, b.campaignState()
	if err := b.setMilitaryEnabled(false); err == nil || *b.militaryRules != rules || !reflect.DeepEqual(snapshot, b.campaignState()) {
		t.Fatal("save/restart bypassed declaration lock")
	}
	for _, mutate := range []func(*campaignV4){
		func(s *campaignV4) { s.Rules.Enabled = false },
		func(s *campaignV4) { s.Rules.DeclaredAt++ },
		func(s *campaignV4) { s.Armory = nil },
		func(s *campaignV4) { s.Rules = game.MilitaryRules{}; s.View.Open = true },
	} {
		var s campaignV4
		if _, err := game.DecodeCampaignSaveVersion(data, &s, 4); err != nil {
			t.Fatal(err)
		}
		mutate(&s)
		bad, err := game.EncodeCampaignSaveVersion(s, nil, 4)
		if err != nil {
			t.Fatal(err)
		}
		if err := b.loadCampaign(bad); err == nil {
			t.Fatal("invalid military state accepted")
		}
		after, _ := json.Marshal(b.armory)
		if string(before) != string(after) || *b.militaryRules != rules || !reflect.DeepEqual(snapshot, b.campaignState()) {
			t.Fatal("failed load partially replaced equipment, rules or world")
		}
	}
}

func TestMilitaryV1V2Compatibility(t *testing.T) {
	a := newArmoryApp(t, 1, 1)
	for _, version := range []int{1, 2} {
		snapshot := a.campaignState()
		var state any
		if version == 1 {
			snapshot.Identity.Mode = "original"
			state = snapshot
		} else {
			state = campaignExpansion{Base: snapshot, Phase: "preparation", Armory: a.armory.Clone(), View: a.armoryViewSnapshot()}
		}
		data, err := game.EncodeCampaignSaveVersion(state, nil, version)
		if err != nil {
			t.Fatal(err)
		}
		b := &app{m: a.m, dataRead: campaignTestRead}
		if err := b.loadCampaign(data); err != nil {
			t.Fatal(err)
		}
		if b.militaryEnabled() != (version == 2) || b.militaryRules != nil && b.militaryRules.Locked() {
			t.Fatal("legacy military option migration", version)
		}
	}
}
