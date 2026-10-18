package main

import (
	"os"
	"reflect"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/game"
)

func TestPaidArmorConversionProtectionSurvivesDisabledSave(t *testing.T) {
	a := newArmoryApp(t, 2, 1)
	// This test isolates pre-declaration production and disabling. The normal
	// map declaration path is covered by invasion_controller_test.go.
	a.screen = screenArmory
	if err := a.armory.Research(a.world, 1, "light-tank"); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		a.endTurn()
	}
	if err := a.armory.Assign(a.world, 1, "light-tank", 2); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 10; i++ {
		a.endTurn()
	}
	if _, err := a.armory.ConvertToArmor(a.world, 1, 1, "light-tank"); err != nil {
		t.Fatal(err)
	}
	a.syncExpansionGenerals()
	unit := a.armory.Units[1]
	profile := a.armory.Protection(1, unit.Branch, unit.Force)
	if profile != (game.EquipmentProtection{Armor: 50000, Piercing: 50000}) {
		t.Fatal("paid conversion profile", profile)
	}
	if err := a.setMilitaryEnabled(false); err != nil {
		t.Fatal(err)
	}
	state := a.campaignState()
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	b := &app{m: a.m, dataRead: campaignTestRead}
	if err := b.loadCampaign(raw); err != nil {
		t.Fatal(err)
	}
	if b.militaryEnabled() || b.armory != b.world.Opts.Armory || !reflect.DeepEqual(state, b.campaignState()) || !reflect.DeepEqual(a.armory, b.armory) || b.armory.Protection(1, unit.Branch, unit.Force) != profile {
		t.Fatal("disabled save/restart lost equipment, profile, world or RNG")
	}
}
