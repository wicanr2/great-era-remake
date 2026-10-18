package main

import (
	"encoding/json"
	"github.com/wicanr2/great-era-remake/internal/game"
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func normalBlitzApp(t *testing.T) *app {
	t.Helper()
	a := declaredStrategyApp(t)
	a.endTurn() // Real selector and executor; no injected target/action.
	if !a.activeBattle() || !a.battle.japaneseBlitz || !a.battle.humanDefender {
		t.Fatalf("normal Japanese selector did not start a defensive battle: turn=%+v", a.strategyTurn)
	}
	t.Logf("normal AI chose %d -> %d; %d armies, remaining %d", a.battle.sim.From, a.battle.sim.At, len(a.battle.sim.Attacker), a.strategyTurn.Remaining)
	return a
}

func TestJapaneseBlitzNormalSelectionAndRestart(t *testing.T) {
	a := normalBlitzApp(t)
	for _, u := range a.battle.sim.Attacker {
		if a.strategyTurn.JapaneseAttacks[u.General] != 1 {
			t.Fatal("actual formation was not charged once")
		}
	}
	a.pauseBattle()
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(a.savePath)
	b := &app{m: a.m, dataRead: campaignTestRead, savePath: filepath.Join(t.TempDir(), "restart.gesave")}
	if err := b.loadCampaign(raw); err != nil {
		t.Fatal(err)
	}
	if strategyState(t, a) != strategyState(t, b) || !b.battle.japaneseBlitz {
		t.Fatal("restart lost monthly attack ledger or formation")
	}
	// The real upkeep formula, not an invented resource toll.
	game.TurnUpkeep(&a.battle.supAtk, &a.battle.supDef)
	game.TurnUpkeep(&b.battle.supAtk, &b.battle.supDef)
	from, at := a.battle.sim.From, a.battle.sim.At
	expectedAtk, expectedDef := a.battle.supAtk, a.battle.supDef
	for _, x := range []*app{a, b} {
		if err := x.withdrawDefenders(x.battle.retreatTargets[0]); err != nil {
			t.Fatal(err)
		}
		if err := x.persistFinishedBattle(); err != nil {
			t.Fatal(err)
		}
		p, _ := x.tbl.At(from)
		q, _ := x.tbl.At(at)
		if int(p.Gold) != expectedAtk.Gold || int(p.Food) != expectedAtk.Food || int(q.Gold) != expectedDef.Gold || int(q.Food) != expectedDef.Food {
			t.Fatal("tactical supplies were not committed")
		}
		if err := x.autosave(); err != nil {
			t.Fatal(err)
		}
		saved, _ := os.ReadFile(x.savePath)
		if err := x.loadCampaign(saved); err != nil {
			t.Fatal(err)
		}
		before := strategyState(t, x)
		if err := x.persistFinishedBattle(); err != nil {
			t.Fatal(err)
		}
		if before != strategyState(t, x) {
			t.Fatal("restart debited supplies twice")
		}
	}
	if strategyState(t, a) != strategyState(t, b) {
		left, right := strategyState(t, a), strategyState(t, b)
		for i := 0; i < min(len(left), len(right)); i++ {
			if left[i] != right[i] {
				t.Fatalf("saved continuation diverged at %d: %s / %s", i, left[max(0, i-100):min(len(left), i+150)], right[max(0, i-100):min(len(right), i+150)])
			}
		}
		t.Fatal("saved continuation length diverged")
	}
	// Complete real selected attacks. Every participating ID stays capped at 2.
	for n := 0; a.strategyTurn != nil && n < 100; n++ {
		for id, count := range a.strategyTurn.JapaneseAttacks {
			if count > 2 {
				t.Fatal("monthly attack cap exceeded", id, count)
			}
		}
		a.endTurn()
		b.endTurn()
		if strategyState(t, a) != strategyState(t, b) {
			t.Fatal("restart changed target selection")
		}
		if a.activeBattle() {
			for _, x := range []*app{a, b} {
				if len(x.battle.retreatTargets) > 0 {
					if err := x.withdrawDefenders(x.battle.retreatTargets[0]); err != nil {
						t.Fatal(err)
					}
				} else {
					x.battle.finishWithWinner(game.BattleSideSecond)
					if err := x.persistFinishedBattle(); err != nil {
						t.Fatal(err)
					}
				}
			}
		}
	}
	if a.strategyTurn != nil || a.month != 8 {
		t.Fatal("blitz turn did not finish once")
	}
}

func TestJapaneseBlitzRejectsInvalidSavedLedger(t *testing.T) {
	a := normalBlitzApp(t)
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(a.savePath)
	for _, mode := range []string{"third", "zero", "range", "missing", "no-turn", "supply", "settled-supply"} {
		var options campaignV4
		legacy, err := game.DecodeCampaignSaveVersion(raw, &options, 4)
		if err != nil {
			t.Fatal(err)
		}
		turn := options.Base.StrategyTurn
		id := options.Base.Battle.Attacker[0].General
		switch mode {
		case "third":
			turn.JapaneseAttacks[id] = 3
		case "zero":
			turn.JapaneseAttacks[id] = 0
		case "range":
			turn.JapaneseAttacks[999] = 1
		case "missing":
			delete(turn.JapaneseAttacks, id)
		case "no-turn":
			options.Base.StrategyTurn = nil
		case "supply":
			options.Base.Battle.SupplyAttacker.Gold = 70000
		case "settled-supply":
			b := options.Base.Battle
			b.Finished, b.Settled = true, true
			b.Settlement = game.BattleSettlement{Kind: game.BattleSettlementDecisive, Winner: game.BattleSideSecond}
			b.RetreatTargets = nil
			b.SupplyAttacker.Gold = 70000
		}
		bad, err := game.EncodeCampaignSaveVersion(options, legacy, 4)
		if err != nil {
			t.Fatal(err)
		}
		before := strategyState(t, a)
		if err := a.loadCampaign(bad); err == nil || before != strategyState(t, a) {
			t.Fatal("invalid saved ledger was accepted or partially loaded", mode)
		}
	}
}

func TestJapaneseBlitzScopeAndSettlementFailure(t *testing.T) {
	a := normalBlitzApp(t)
	turn := a.strategyTurn
	for _, mode := range []string{"prewar", "disabled", "japan-player", "stage-one", "no-cursor"} {
		yearStage, player, rules := a.stage, a.playerCommander, *a.militaryRules
		switch mode {
		case "prewar":
			a.militaryRules.DeclaredAt = 0
		case "disabled":
			a.militaryRules.Enabled = false
		case "japan-player":
			a.playerCommander = 107
		case "stage-one":
			a.stage = 1
		case "no-cursor":
			a.strategyTurn = nil
		}
		if a.japaneseBlitzActive() {
			t.Fatal("special policy escaped declared Chinese campaign", mode)
		}
		a.stage, a.playerCommander, a.strategyTurn = yearStage, player, turn
		*a.militaryRules = rules
	}
	a.battle.finishWithWinner(game.BattleSideFirst)
	a.battle.supDef.Food = -1
	before := strategyState(t, a)
	if err := a.persistFinishedBattle(); err == nil || strategyState(t, a) != before {
		t.Fatal("failed supply validation partially committed settlement")
	}
	if !reflect.DeepEqual(turn.JapaneseAttacks, a.strategyTurn.JapaneseAttacks) {
		t.Fatal("validation changed attack count")
	}
}

func TestJapaneseBlitzExportSample(t *testing.T) {
	dir := os.Getenv("DSDS_BLITZ_SAMPLE_DIR")
	if dir == "" {
		return
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		t.Fatal("existing sample directory required", err)
	}
	a := declaredStrategyApp(t)
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(a.savePath)
	if err := os.WriteFile(filepath.Join(dir, "blitz-before-turn.gesave"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	a.endTurn()
	if !a.activeBattle() || !a.battle.japaneseBlitz {
		t.Fatal("normal AI did not attack")
	}
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	raw, _ = os.ReadFile(a.savePath)
	if err := os.WriteFile(filepath.Join(dir, "blitz-selected.gesave"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	meta := map[string]any{"seed": 1, "scope": "real 1937 data, legal Chinese transfer, premarked declaration; normal endTurn and Japanese target selection; declaration and finite waves remain unconnected", "from": a.battle.sim.From, "at": a.battle.sim.At, "cursor": a.strategyTurn}
	raw, err = json.MarshalIndent(meta, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "blitz-sample.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
}
