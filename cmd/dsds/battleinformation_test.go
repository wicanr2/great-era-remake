package main

import (
	"encoding/json"
	"github.com/wicanr2/great-era-remake/internal/game"
	"os"
	"reflect"
	"testing"
)

func TestBattleInspectionSaveRestoresBothSidesWithoutMutation(t *testing.T) {
	a := newArmoryApp(t, 1, 1)
	if err := a.startBattle(25, a.current); err != nil {
		t.Fatal(err)
	}
	sim := a.battle.sim
	a.battle.mode = battleModeInspect
	before, _ := json.Marshal(a.campaignState())
	armyBefore, _ := json.Marshal(a.armory)
	info := game.InspectBattle(sim)
	if len(info.Sides[0].Roster) != len(sim.Attacker) || len(info.Sides[1].Roster) != len(sim.Defender) {
		t.Fatal("lost participants")
	}
	// The dedicated view must compose without the unrelated legacy strategy
	// panel's fonts. That panel used to cover the entire attacker column.
	if err := a.compose(); err != nil {
		t.Fatal(err)
	}
	if a.frame == nil || a.frame.Bounds().Dx() != 640 {
		t.Fatal("missing original-size inspection")
	}
	a.frame.Dispose()
	a.frame = nil
	after, _ := json.Marshal(a.campaignState())
	armyAfter, _ := json.Marshal(a.armory)
	if !reflect.DeepEqual(before, after) || !reflect.DeepEqual(armyBefore, armyAfter) {
		t.Fatal("inspection mutated game or RNG")
	}
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	restored := &app{m: a.m, dataRead: campaignTestRead}
	if err := restored.loadCampaign(data); err != nil {
		t.Fatal(err)
	}
	if restored.battle.mode != battleModeInspect || !reflect.DeepEqual(info, game.InspectBattle(restored.battle.sim)) {
		t.Fatal("saved inspection lost live units or equipment")
	}
}
