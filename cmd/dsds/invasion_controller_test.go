package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/game"
	uiresolution "github.com/wicanr2/great-era-remake/internal/ui/resolution"
)

func processInvasion(t *testing.T, a *app) bool {
	t.Helper()
	changed, err := a.processCampaignEvents()
	if err != nil {
		t.Fatal(err)
	}
	return changed
}

func restartControlledInvasion(t *testing.T, a *app) *app {
	t.Helper()
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	b := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
	if err := b.loadCampaign(raw); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a.campaignState(), b.campaignState()) || !reflect.DeepEqual(a.armory, b.armory) || !reflect.DeepEqual(a.declaration, b.declaration) {
		t.Fatal("restart changed controller, army or world")
	}
	return b
}

func exportControlledInvasion(t *testing.T, a *app, name string) {
	t.Helper()
	if out := os.Getenv("DSDS_FINISH_SAMPLE_DIR"); out != "" {
		if err := a.autosave(); err != nil {
			t.Fatal(err)
		}
		raw, err := os.ReadFile(a.savePath)
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(out, name), raw, 0600); err != nil {
			t.Fatal(err)
		}
	}
}

func TestInvasionControllerOriginalOpeningAndEventWindowSize(t *testing.T) {
	a := &app{resolution: uiresolution.ModeOriginal, toolbar: newToolbarState(), opening: &openingState{}}
	if w, h := a.windowSize(); w != 1280 || h != 768 {
		t.Fatal("original opening was enlarged beyond a 1280 screen", w, h)
	}
	a.opening = nil
	a.screen = screenCampaignEvent
	if w, h := a.windowSize(); w != 1280 || h != 748 {
		t.Fatal("original event did not preserve classic scale", w, h)
	}
}

func TestInvasionControllerNormalOpeningAndThreeWaves(t *testing.T) {
	for _, stage := range []uint8{2, 3} {
		t.Run(string(rune('0'+stage)), func(t *testing.T) {
			a := newArmoryApp(t, stage, 1)
			native := append([]game.General(nil), a.generals...)
			initial := *a.tbl
			if !processInvasion(t, a) || a.screen != screenCampaignEvent || !a.invasionControlled() || len(a.campaignActors) != 0 {
				t.Fatal("normal scenario did not declare before deployment")
			}
			if !reflect.DeepEqual(*a.tbl, initial) || !reflect.DeepEqual(native, a.generals) {
				t.Fatal("declaration changed the original world")
			}
			date := a.declaration.Date
			cast := append([]game.GeneralID(nil), a.declaration.Cast...)
			a = restartControlledInvasion(t, a)
			if !processInvasion(t, a) || a.screen != screenCampaignEvent {
				t.Fatal("unread announcement was lost")
			}
			a.acknowledgeCampaignEvent()
			processInvasion(t, a)
			if len(a.campaignActors) != 6 || len(a.armory.InvasionBatches) != 1 || a.armory.Landing != nil {
				t.Fatal("first six not deployed to existing Japanese port")
			}
			if err := a.setMilitaryEnabled(false); err == nil {
				t.Fatal("declaration can be disabled")
			}
			if !reflect.DeepEqual(native, a.generals[:len(native)]) {
				t.Fatal("new officers overwrote MAN people")
			}
			for month := uint32(0); month <= 6; month++ {
				// Controlled calendar fixture: no AI or research claim. The production
				// controller, arrival rules and v4 persistence remain the real functions.
				target := date.DeclaredAt + month
				for a.armory.LastMonth < target {
					if err := a.armory.PauseMonth(a.world, a.armory.LastMonth+1); err != nil {
						t.Fatal(err)
					}
				}
				a.year, a.month = uint16((target-1)/12), uint8((target-1)%12+1)
				processInvasion(t, a)
				count := 6
				if month >= 3 {
					count = 8
				}
				if month >= 6 {
					count = 10
				}
				if len(a.campaignActors) != count || !reflect.DeepEqual(a.declaration.Cast, cast) {
					t.Fatal("wrong schedule or cast reselected", month, len(a.campaignActors))
				}
				before := a.armory.Clone()
				processInvasion(t, a)
				if !reflect.DeepEqual(before, a.armory) {
					t.Fatal("same month duplicated imports")
				}
				a = restartControlledInvasion(t, a)
			}
			var gold, food, ammo, fuel uint32
			for _, b := range a.armory.InvasionBatches {
				gold += b.Delivered.Gold
				food += b.Delivered.Food
				ammo += b.Delivered.Ammo
				fuel += b.Delivered.Fuel
			}
			if gold != 24000 || food != 32000 || ammo != 24000 || fuel != 12000 {
				t.Fatal("finite supply differs", gold, food, ammo, fuel)
			}
		})
	}
}

func TestInvasionControllerOptionalAndJapanesePlayer(t *testing.T) {
	for _, test := range []struct {
		stage   uint8
		player  game.GeneralID
		enabled bool
		want    bool
	}{{1, 1, false, false}, {1, 1, true, false}, {2, 1, false, false}, {2, 1, true, true}, {3, 1, true, true}, {2, 107, false, false}} {
		a := &app{m: campaignTestMap(t), dataRead: campaignTestRead, initialSeed: 1, savePath: filepath.Join(t.TempDir(), "current.gesave")}
		if err := a.startCampaign(test.stage, test.player); err != nil {
			t.Fatal(err)
		}
		if test.enabled {
			if err := a.setMilitaryEnabled(true); err != nil {
				t.Fatal(err)
			}
		}
		if got := processInvasion(t, a); got != test.want {
			t.Fatal("incorrect optional declaration", test, got)
		}
		if test.player == 107 && a.setMilitaryEnabled(true) == nil {
			t.Fatal("Japan can invade itself")
		}
	}
}

func firstStageUnifiedFixture(t *testing.T, empty bool) *app {
	t.Helper()
	a := newArmoryApp(t, 1, 1)
	// Explicit hypothetical unified snapshot. Never claimed as normal unification.
	for i := 0; i < a.provinceLimit; i++ {
		if a.tbl.Province[i].Commander != 0 {
			a.tbl.Province[i].Commander = 1
		}
	}
	if empty {
		for a.world.RosterOf(game.InvasionEntryProvince).Len() > 0 {
			if report := a.world.ApplyTransfer(game.InvasionEntryProvince, game.TransferFirst, 11); len(report.Moved) == 0 {
				t.Fatal("fixture transfer did not move any unit")
			}
		}
		a.syncExpansionGenerals()
	} else {
		for i := 0; i < 4 && a.world.RosterOf(a.current).Len() > 1; i++ {
			a.world.ApplyTransfer(a.current, game.TransferFirst, game.InvasionEntryProvince)
		}
		a.syncExpansionGenerals()
	}
	if err := a.armory.Sync(a.world); err != nil {
		t.Fatal(err)
	}
	processInvasion(t, a)
	a.acknowledgeCampaignEvent()
	processInvasion(t, a)
	return a
}

func TestInvasionControllerFirstStageDefenderWithdrawalAndContinuation(t *testing.T) {
	a := firstStageUnifiedFixture(t, false)
	if !a.activeBattle() || !a.battle.humanDefender || !a.battle.externalInvasion || len(a.battle.sim.Defender) == 0 {
		t.Fatal("full controller skipped a real garrison")
	}
	a.pauseBattle()
	a = restartControlledInvasion(t, a)
	exportControlledInvasion(t, a, "finish-stage1-active.gesave")
	if len(a.battle.retreatTargets) == 0 {
		t.Fatal("fixture has no legal withdrawal")
	}
	if err := a.withdrawDefenders(a.battle.retreatTargets[0]); err != nil {
		t.Fatal(err)
	}
	a = restartControlledInvasion(t, a)
	processInvasion(t, a)
	if a.invasionEnded() || a.armory.Landing.Attempts[0].Outcome != game.LandingLanded {
		t.Fatal("withdrawal did not continue the invasion")
	}
	// Same-month control boundary must neither re-enter nor duplicate landing.
	a = restartControlledInvasion(t, a)
	processInvasion(t, a)
	if len(a.armory.Landing.Attempts) != 1 || len(a.armory.InvasionBatches) != 1 {
		t.Fatal("landing replayed after restart")
	}
}

func TestInvasionControllerDefeatAndLostPortWait(t *testing.T) {
	a := newArmoryApp(t, 2, 1)
	processInvasion(t, a)
	a.acknowledgeCampaignEvent()
	processInvasion(t, a)
	// Explicit boundary fixture: loss of the entry port while other Japanese
	// territory still exists delays supply and reinforcements.
	p, _ := a.tbl.At(game.InvasionEntryProvince)
	p.Capture(1)
	target := a.declaration.Date.DeclaredAt + 3
	for a.armory.LastMonth < target {
		if err := a.armory.PauseMonth(a.world, a.armory.LastMonth+1); err != nil {
			t.Fatal(err)
		}
	}
	a.year, a.month = uint16((target-1)/12), uint8((target-1)%12+1)
	processInvasion(t, a)
	if len(a.armory.InvasionBatches) != 1 || a.invasionEnded() {
		t.Fatal("reinforcements teleported across an enemy port")
	}
	for i := 0; i < a.provinceLimit; i++ {
		if a.tbl.Province[i].Commander == 1 {
			a.tbl.Province[i].Capture(107)
		}
	}
	if !processInvasion(t, a) || a.declaration.Outcome != "defeat" {
		t.Fatal("loss of all Chinese provinces did not end campaign")
	}
	exportControlledInvasion(t, a, "finish-controlled-defeat.gesave")
	a = restartControlledInvasion(t, a)
	processInvasion(t, a)
	a.acknowledgeCampaignEvent()
	before := a.campaignState()
	a.endTurn()
	if !reflect.DeepEqual(before, a.campaignState()) || !a.invasionEnded() {
		t.Fatal("terminal defeat remains active")
	}
}

func TestInvasionControllerFirstStageHQEmptyPortAndVictory(t *testing.T) {
	a := firstStageUnifiedFixture(t, true)
	p, _ := a.tbl.At(game.InvasionEntryProvince)
	if p.Commander != 300 || len(a.world.Units) != 300 || len(a.campaignActors) != 6 || a.battle != nil {
		t.Fatal("empty port did not settle without an unplayable battle")
	}
	if a.world.Units[299].Active || a.generals[299] != (game.General{}) {
		t.Fatal("HQ became an eleventh soldier or historical person")
	}
	a = restartControlledInvasion(t, a)
	// Terminal-state boundary fixture, separately from actual battle settlement.
	for i := 0; i < a.provinceLimit; i++ {
		if a.tbl.Province[i].Commander == 300 {
			a.tbl.Province[i].Capture(1)
		}
	}
	if !processInvasion(t, a) || a.declaration.Outcome != "victory" {
		t.Fatal("liberated front did not end campaign")
	}
	exportControlledInvasion(t, a, "finish-controlled-victory.gesave")
	a = restartControlledInvasion(t, a)
	processInvasion(t, a)
	a.acknowledgeCampaignEvent()
	date := a.armory.LastMonth
	world := a.campaignState()
	a.endTurn()
	processInvasion(t, a)
	if a.armory.LastMonth != date || !reflect.DeepEqual(world, a.campaignState()) || len(a.campaignActors) != 6 {
		t.Fatal("ending kept simulating or issued future reinforcements")
	}
}

func TestInvasionControllerRejectsForgedStateAtomically(t *testing.T) {
	a := newArmoryApp(t, 2, 1)
	processInvasion(t, a)
	a.acknowledgeCampaignEvent()
	processInvasion(t, a)
	a = restartControlledInvasion(t, a)
	good, _ := os.ReadFile(a.savePath)
	before, army := a.campaignState(), a.armory.Clone()
	for _, kind := range []string{"cast", "controller", "imported", "outcome", "unread", "actor", "batch"} {
		var v campaignV4
		if _, err := game.DecodeCampaignSaveVersion(good, &v, 4); err != nil {
			t.Fatal(err)
		}
		switch kind {
		case "cast":
			v.Declaration.Cast[0] = 293
		case "controller":
			v.Declaration.Controller = 2
		case "imported":
			v.Declaration.Imported = true
		case "outcome":
			v.Declaration.Outcome = "victory"
		case "unread":
			v.Declaration.AnnouncementRead = false
		case "actor":
			v.Actors[0].Key = "unknown"
		case "batch":
			b := v.Armory.InvasionBatches[1]
			b.Units[0], b.Units[1] = b.Units[1], b.Units[0]
			v.Armory.InvasionBatches[1] = b
		}
		bad, err := game.EncodeCampaignSaveVersion(v, nil, 4)
		if err != nil {
			t.Fatal(err)
		}
		if a.loadCampaign(bad) == nil {
			t.Fatal("accepted forged controller", kind)
		}
		if !reflect.DeepEqual(before, a.campaignState()) || !reflect.DeepEqual(army, a.armory) {
			t.Fatal("failed load partly replaced campaign", kind)
		}
	}
}
