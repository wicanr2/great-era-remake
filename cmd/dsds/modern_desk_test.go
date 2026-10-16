package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	"github.com/wicanr2/great-era-remake/internal/ui/layout"
)

func TestModernBattlePauseRetainsStateAndRestart(t *testing.T) {
	a := &app{m: campaignTestMap(t), dataRead: campaignTestRead, initialSeed: 1, savePath: filepath.Join(t.TempDir(), "pause.gesave")}
	if err := a.startCampaign(1, 58); err != nil {
		t.Fatal(err)
	}
	target := a.tbl.FirstAttackable(a.current)
	if err := a.startBattle(target, a.current); err != nil {
		t.Fatal(err)
	}
	a.battle.turn = 2
	a.battle.sim.Attacker[0].Current--
	before := a.campaignState()
	battle := a.battle
	a.eten = &assets.EtenFonts{}
	r := layout.BattleReturnButton(a.highModernBattle())
	if actions.Hit(a.interactiveTargets(), r.X+r.W/2, r.Y+r.H/2) != actions.Back {
		t.Fatal("visible battle return lacks a pointer target")
	}
	a.pauseBattle()
	if a.screen != screenMap || a.battle != battle || !battle.paused {
		t.Fatal("pause discarded active battle")
	}
	paused := a.campaignState()
	a.endTurn()
	if err := a.startBattle(target, a.current); err == nil {
		t.Fatal("second battle replaced paused battle")
	}
	a.openArmory()
	if a.screen != screenMap || !reflect.DeepEqual(paused, a.campaignState()) {
		t.Fatal("paused world accepted a mutating command")
	}
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	b := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
	if err := b.loadCampaign(data); err != nil {
		t.Fatal(err)
	}
	if b.screen != screenMap || b.battle == nil || !b.battle.paused || !reflect.DeepEqual(paused, b.campaignState()) {
		t.Fatal("restart lost paused battle state")
	}
	b.resumeBattle()
	if b.screen != screenBattle || b.battle.paused || !reflect.DeepEqual(before, b.campaignState()) {
		t.Fatal("resume redeployed units or changed RNG/resources")
	}
}

func TestModernMapRejectsProvincesOutsideScenario(t *testing.T) {
	a := newArmoryApp(t, 1, 1)
	current := a.current
	for _, key := range []string{"province.39", "province.0", "pick.39"} {
		a.desk.Picker = "province"
		a.deskAction(key)
		if a.current != current {
			t.Fatalf("%s selected an unavailable province", key)
		}
	}
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	b := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
	if err := b.loadCampaign(data); err != nil {
		t.Fatal("map selection produced an unloadable save", err)
	}
}

func TestModernTradeLimitUsesActualAmountWithoutMutation(t *testing.T) {
	a := newArmoryApp(t, 2, 1)
	a.tradeGood = game.GoodFood
	p, _ := a.tbl.At(a.current)
	p.Gold = 1000
	p.Food = 321
	before := a.campaignState()
	if limit := a.deskTradeLimit(); limit != 321 {
		t.Fatalf("export offered %d, have 321", limit)
	}
	a.tradeImport = true
	limit := a.deskTradeLimit()
	table := *a.tbl
	world := *a.world
	world.Table = &table
	r, e := world.Import(a.current, a.tradeGood, limit)
	if e != nil || r.Amount != limit || limit == 0 {
		t.Fatalf("quote is not executable: %d %+v %v", limit, r, e)
	}
	if !reflect.DeepEqual(before, a.campaignState()) {
		t.Fatal("preview quote mutated world")
	}
}

func TestModernReturnCancelsConfirmAndRestoresEntireReorganization(t *testing.T) {
	a := newArmoryApp(t, 2, 1)
	a.screen = screenRecruitConfirm
	a.recruitAmount = 500
	a.deskRequestBack(false)
	if a.screen != screenRecruitAmount || a.recruitAmount != 500 {
		t.Fatal("cancel lost the draft")
	}
	a.deskRequestBack(true)
	if a.screen != screenMap {
		t.Fatal("home failed to reach map")
	}
	before := a.campaignState()
	if err := a.deskBeginReorganization(game.BranchInfantry); err != nil {
		t.Fatal(err)
	}
	id := a.reorganization.Targets()[0]
	if err := a.reorganization.Assign(id, 0); err != nil {
		t.Fatal(err)
	}
	a.syncReorganizationGenerals()
	a.deskRequestBack(true)
	if a.screen != screenReorganizeCancel {
		t.Fatal("home silently discarded a reorganization")
	}
	a.deskConfirm()
	if a.screen != screenMap || a.reorganization != nil || !reflect.DeepEqual(before, a.campaignState()) {
		t.Fatal("cancel failed to restore world, equipment or budget")
	}
}
