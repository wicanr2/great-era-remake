package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/game"
)

func ledgerLandedApp(t *testing.T) *app {
	t.Helper()
	a := externalLandingApp(t) // Explicit declaration and post-reconquest fixture.
	if err := a.withdrawDefenders(a.battle.retreatTargets[0]); err != nil {
		t.Fatal(err)
	}
	// Legal player transfer creates a weak point without granting resources or
	// changing AI thresholds. It keeps the weakest original army in Hebei.
	selection, err := a.world.BeginPlayerTransfer(11, 16, game.PlayerTransferPartial, a.generals)
	if err != nil {
		t.Fatal(err)
	}
	candidates := selection.Candidates()
	if len(candidates) < 2 {
		t.Fatal("expected real transferred garrison")
	}
	keep := candidates[0]
	for _, id := range candidates {
		if a.generals[id-1].Force < a.generals[keep-1].Force {
			keep = id
		}
	}
	for _, id := range candidates {
		if id != keep {
			if err := selection.Toggle(id); err != nil {
				t.Fatal(err)
			}
		}
	}
	a.current, a.transferSession = 11, selection
	if a.execPlayerTransfer() == "" {
		t.Fatal("legal weak-point preparation failed")
	}
	a.transferSession = nil
	a.screen = screenMap
	return a
}

func restartLedger(t *testing.T, a *app, export string) *app {
	t.Helper()
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	if output := os.Getenv("DSDS_LEDGER_SAMPLE_DIR"); output != "" && export != "" {
		if err := os.WriteFile(filepath.Join(output, export), raw, 0o600); err != nil {
			t.Fatal(err)
		}
	}
	b := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
	if err := b.loadCampaign(raw); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(a.armory, b.armory) || !reflect.DeepEqual(a.campaignState(), b.campaignState()) {
		left, _ := json.Marshal(a.campaignState())
		right, _ := json.Marshal(b.campaignState())
		for i := 0; i < min(len(left), len(right)); i++ {
			if left[i] != right[i] {
				t.Fatalf("restart state diff at %d: %s / %s", i, left[max(0, i-100):min(len(left), i+150)], right[max(0, i-100):min(len(right), i+150)])
			}
		}
		t.Fatalf("restart private state mismatch: jsonEqual=%v armoryEqual=%v", string(left) == string(right), reflect.DeepEqual(a.armory, b.armory))
	}
	return b
}

func TestInvasionLedgerLandingAndNormalBlitzShareTwoAttacks(t *testing.T) {
	a := ledgerLandedApp(t)
	a = restartLedger(t, a, "ledger-before-turn.gesave")
	basis, err := a.armory.InvasionMonthlyAttacks(a.armory.LastMonth)
	if err != nil || len(basis) != 6 {
		t.Fatal(basis, err)
	}
	a.endTurn() // Normal selector, command budget and strategic cursor.
	if !a.activeBattle() || !a.battle.japaneseBlitz || a.battle.sim.From != 7 || a.battle.sim.At != 11 {
		t.Fatalf("normal post-landing AI did not select Hebei: cursor=%+v", a.strategyTurn)
	}
	imported := 0
	for _, u := range a.battle.sim.Attacker {
		if basis[u.General] > 0 {
			imported++
		}
		if a.strategyTurn.JapaneseAttacks[u.General] != basis[u.General]+1 {
			t.Fatal("landing attack was omitted or counted twice", u.General, a.strategyTurn.JapaneseAttacks)
		}
	}
	if imported != len(basis) {
		t.Fatal("normal roster omitted landing participants", imported, len(basis))
	}
	a.pauseBattle()
	a = restartLedger(t, a, "ledger-second-attack.gesave")
	if err := a.withdrawDefenders(a.battle.retreatTargets[0]); err != nil {
		t.Fatal(err)
	}
	a = restartLedger(t, a, "ledger-second-settled.gesave")
	plan := a.world.PlanJapaneseBlitz(11, 107, a.strategyTurn.JapaneseAttacks, a.provinceLimit)
	for _, id := range plan.Units {
		if basis[id] > 0 {
			t.Fatal("same unit could attack for a third time after landing", plan)
		}
	}
	before := a.armory.Clone()
	for i := 0; i < 3; i++ {
		if err := a.persistFinishedBattle(); err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(before, a.armory) {
			t.Fatal("finished battle changed offshore history")
		}
	}
	t.Logf("normal AI 7->11 uses all six landing participants: 1 offshore + 1 provincial; third attack rejected")
}

func TestInvasionLedgerRejectsMissingOrUnderreportedSavedCounts(t *testing.T) {
	a := ledgerLandedApp(t)
	a.endTurn()
	if !a.activeBattle() {
		t.Fatal("normal selector did not attack")
	}
	a = restartLedger(t, a, "")
	raw, _ := os.ReadFile(a.savePath)
	before, army := a.campaignState(), a.armory.Clone()
	for _, mode := range []string{"nil-ledger", "missing-offshore", "underreported-current-battle", "over-limit"} {
		var v campaignV4
		if _, err := game.DecodeCampaignSaveVersion(raw, &v, 4); err != nil {
			t.Fatal(err)
		}
		id := game.GeneralID(0)
		for _, u := range v.Base.Battle.Attacker {
			if a.campaignActors[u.General] != "" {
				id = u.General
				break
			}
		}
		if id == 0 {
			t.Fatal("normal battle has no imported participant")
		}
		switch mode {
		case "nil-ledger":
			v.Base.StrategyTurn.JapaneseAttacks = nil
		case "missing-offshore":
			delete(v.Base.StrategyTurn.JapaneseAttacks, id)
		case "underreported-current-battle":
			v.Base.StrategyTurn.JapaneseAttacks[id] = 1
		case "over-limit":
			v.Base.StrategyTurn.JapaneseAttacks[id] = 3
		}
		bad, err := game.EncodeCampaignSaveVersion(v, nil, 4)
		if err != nil {
			t.Fatal(err)
		}
		if err := a.loadCampaign(bad); err == nil {
			t.Fatal("bad shared ledger accepted", mode)
		}
		if !reflect.DeepEqual(before, a.campaignState()) || !reflect.DeepEqual(army, a.armory) {
			t.Fatal("bad ledger partially loaded", mode)
		}
	}
}
