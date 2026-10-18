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

func newArmoryApp(t *testing.T, stage uint8, player game.GeneralID) *app {
	t.Helper()
	a := &app{m: campaignTestMap(t), dataRead: campaignTestRead, initialSeed: 1, savePath: filepath.Join(t.TempDir(), "campaign.gesave")}
	if err := a.startCampaignMode(stage, player, "unification-resistance"); err != nil {
		t.Fatal(err)
	}
	a.world.Map = a.m
	return a
}

func TestReorganizationCancelRestoresEquipmentAfterSavingDraft(t *testing.T) {
	a := newArmoryApp(t, 2, 1)
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	initial, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	a.reorganization, err = a.world.BeginReorganization(a.current, game.BranchInfantry)
	if err != nil {
		t.Fatal(err)
	}
	id := a.reorganization.Targets()[0]
	if err := a.reorganization.Assign(id, 100); err != nil {
		t.Fatal(err)
	}
	a.syncReorganizationGenerals()
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	draft, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	restored := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
	if err := restored.loadCampaign(draft); err != nil {
		t.Fatal(err)
	}
	if !restored.reorganization.CanCancel() {
		t.Fatal("new snapshot lost rollback")
	}
	if err := restored.reorganization.Cancel(); err != nil {
		t.Fatal(err)
	}
	restored.syncReorganizationGenerals()
	restored.reorganization = nil
	restored.screen = screenMap
	if restored.armory != restored.world.Opts.Armory {
		t.Fatal("cancel broke shared equipment object")
	}
	if err := restored.autosave(); err != nil {
		t.Fatal(err)
	}
	after, err := os.ReadFile(restored.savePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(initial, after) {
		t.Fatal("cancel after save/restart changed force, arms, equipment, budget or RNG")
	}
}
func TestArmoryAllExpansionFactionsSave(t *testing.T) {
	count := 0
	for _, sc := range game.Scenarios {
		b, _ := campaignTestRead(fmt.Sprintf("TOWN(%d).DAT", sc.Stage))
		town, err := game.ParseTownFile(b)
		if err != nil {
			t.Fatal(err)
		}
		b, _ = campaignTestRead(fmt.Sprintf("MAN(%d).DAT", sc.Stage))
		gens, err := game.ParseGenerals(b, sc.Generals)
		if err != nil {
			t.Fatal(err)
		}
		choices, err := game.ScenarioFactionChoices(sc.Stage, town, gens)
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range choices {
			if sc.Stage > 1 && f.Leader == 107 {
				continue
			}
			t.Run(fmt.Sprintf("stage%d-leader%d", sc.Stage, f.Leader), func(t *testing.T) {
				a := newArmoryApp(t, sc.Stage, f.Leader)
				if a.armory == nil || a.world.Opts.Armory != a.armory {
					t.Fatal("no shared equipment state")
				}
				if err := a.armory.Research(a.world, f.Leader, "rifle-modern"); err != nil {
					t.Fatal(err)
				}
				if err := a.autosave(); err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(a.savePath)
				if err != nil {
					t.Fatal(err)
				}
				v, err := game.CampaignSaveVersion(data)
				if err != nil || v != game.CampaignSaveCurrentVersion {
					t.Fatal(v, err)
				}
				restored := &app{m: a.m, dataRead: campaignTestRead}
				if err := restored.loadCampaign(data); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(a.armory, restored.armory) || !reflect.DeepEqual(a.campaignState(), restored.campaignState()) {
					t.Fatal("expansion restart changed world or equipment")
				}
				if restored.armory != restored.world.Opts.Armory {
					t.Fatal("restored world has another armory")
				}
			})
			count++
		}
	}
	if count != 12 {
		t.Fatal(count)
	}
	a := &app{m: campaignTestMap(t), dataRead: campaignTestRead}
	if err := a.startCampaignMode(2, 107, "unification-resistance"); err == nil {
		t.Fatal("Japanese faction entered resistance campaign")
	}
}
func TestArmoryResearchProductionEquipmentRestart(t *testing.T) {
	a := newArmoryApp(t, 2, 1)
	// Stay in the preparation surface to isolate research from declaration.
	a.screen = screenArmory
	if err := a.armory.Research(a.world, 1, "rifle-modern"); err != nil {
		t.Fatal(err)
	}
	a.endTurn()
	a.endTurn()
	n := a.armory.Nations[1]
	if !n.Known["rifle-modern"] || n.Stock["rifle-modern"] != 0 {
		t.Fatal("research did not unlock design only")
	}
	if err := a.armory.Assign(a.world, 1, "rifle-modern", 1); err != nil {
		t.Fatal(err)
	}
	a.endTurn()
	n = a.armory.Nations[1]
	if n.Stock["rifle-modern"] != 24 {
		t.Fatal(n.Stock)
	}
	var selected game.GeneralID
	for id := game.GeneralID(1); int(id) <= len(a.generals); id++ {
		u := a.armory.Units[id]
		if u.Faction == 1 && u.Branch == game.BranchInfantry && u.Force >= 2400 {
			p, _ := a.tbl.At(a.world.Units[int(id)-1].Province)
			if p != nil && p.Commander == 1 {
				selected = id
				break
			}
		}
	}
	if selected == 0 {
		t.Fatal("no eligible infantry")
	}
	if count, err := a.armory.Issue(a.world, 1, selected, "rifle-modern"); err != nil || count != 24 {
		t.Fatal(count, err)
	}
	a.armoryView = defaultArmoryView()
	a.armoryView.Page = 3
	a.screen = screenArmory
	before, _ := json.Marshal(a.armory)
	rng := a.rng.Seed()
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(a.savePath)
	restored := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
	if err := restored.loadCampaign(data); err != nil {
		t.Fatal(err)
	}
	after, _ := json.Marshal(restored.armory)
	if !bytes.Equal(before, after) || restored.screen != screenArmory || restored.armoryView.Page != 3 || restored.rng.Seed() != rng {
		t.Fatal("restart lost equipped gear, UI, date or RNG")
	}
	if err := restored.autosave(); err != nil {
		t.Fatal(err)
	}
	saved, _ := os.ReadFile(a.savePath)
	if !bytes.Equal(saved, data) {
		t.Fatal("saving again changed or reissued equipment")
	}
	if restored.world.Opts.Armory.Power(selected, game.BranchInfantry, restored.armory.Units[selected].Force, false) <= 100 {
		t.Fatal("equipped rifles do not affect shared strength")
	}
}
func TestArmoryDamagedLoadAndFailedWriteKeepState(t *testing.T) {
	a := newArmoryApp(t, 1, 1)
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	good, _ := os.ReadFile(a.savePath)
	before := a.campaignState()
	army, _ := json.Marshal(a.armory)
	var snapshot campaignV4
	if _, err := game.DecodeCampaignSaveVersion(good, &snapshot, 4); err != nil {
		t.Fatal(err)
	}
	snapshot.Armory.Nations[1].Stock["rifle-modern"]++
	bad, err := game.EncodeCampaignSaveVersion(snapshot, nil, 4)
	if err != nil {
		t.Fatal(err)
	}
	if err := a.loadCampaign(bad); err == nil {
		t.Fatal("fabricated stock accepted")
	}
	after, _ := json.Marshal(a.armory)
	if !bytes.Equal(army, after) || !reflect.DeepEqual(before, a.campaignState()) {
		t.Fatal("bad load partially replaced state")
	}
	path := a.savePath
	a.savePath = t.TempDir()
	if err := a.autosave(); err == nil {
		t.Fatal("directory save accepted")
	}
	old, _ := os.ReadFile(path)
	if !bytes.Equal(old, good) {
		t.Fatal("failed save damaged previous file")
	}
	a.savePath = path
	if err := a.startCampaign(1, 1); err != nil {
		t.Fatal(err)
	}
	if a.armory != nil || a.world.Opts.Armory != nil {
		t.Fatal("original campaign retained expansion")
	}
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	original, _ := os.ReadFile(path)
	version, _ := game.CampaignSaveVersion(original)
	if version != game.CampaignSaveCurrentVersion {
		t.Fatal("new campaign lacks live rules", version)
	}
}
