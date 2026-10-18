package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
)

// Declaration is premarked only to isolate the strategy-turn adapter. These
// tests do not prove the still-unconnected declaration/finite-wave campaign.
func declaredStrategyApp(t *testing.T) *app {
	t.Helper()
	a := newArmoryApp(t, 2, 1)
	a.militaryRules.DeclaredAt = uint32(a.year)*12 + uint32(a.month)
	// Legal partial transfer leaves the weakest existing army to defend Hebei.
	// No forces, supplies, AI gates or commander ownership are injected.
	selection, err := a.world.BeginPlayerTransfer(11, 16, game.PlayerTransferPartial, a.generals)
	if err != nil {
		t.Fatal(err)
	}
	candidates := selection.Candidates()
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
		t.Fatal("legal setup transfer was not executed")
	}
	a.transferSession = nil
	return a
}

func strategyState(t *testing.T, a *app) string {
	t.Helper()
	b, err := json.Marshal(campaignOptions{Base: a.campaignState(), Rules: *a.militaryRules, Armory: a.armory})
	if err != nil {
		t.Fatal(err)
	}
	return string(b)
}

func firstStrategyAttack(t *testing.T) (*app, []byte) {
	t.Helper()
	a := declaredStrategyApp(t)
	a.armory.PlanAI(a.world, a.playerCommander)
	a.strategyTurn = &strategyTurn{Year: a.year, Month: a.month, Player: a.playerCommander, Province: 1}
	year, month, settled := a.year, a.month, a.armory.LastMonth
	// Replay the selected action at the executor boundary. The existing AI
	// prefers rescue transfers in this original scenario; this is a controlled
	// action input, not a normal AI target-selection or declaration receipt.
	a.runComputerActions(a.strategyTurn, func(p game.ProvinceID, rep *computerTurnReport) bool {
		if p == 7 {
			return a.computerAttack(game.AIAction{Kind: game.AIAttack, From: 7, To: 11}, 107, rep)
		}
		return a.runOneComputerAction(p, rep)
	})
	if !a.activeBattle() || !a.battle.humanDefender || a.year != year || a.month != month || a.armory.LastMonth != settled || a.screen != screenBattle {
		t.Fatal("selected attack failed to interrupt before month settlement")
	}
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("controlled 1937 attack %d/%d: %d -> %d remaining=%d", year, month, a.battle.sim.From, a.battle.sim.At, a.strategyTurn.Remaining)
	return a, raw
}

func TestStrategyTurnSelectedAttackAndSavedContinuation(t *testing.T) {
	a, _ := firstStrategyAttack(t)
	before := strategyState(t, a)
	if a.endTurn() == "" || strategyState(t, a) != before {
		t.Fatal("end-turn changed an active battle")
	}
	a.pauseBattle()
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	b := &app{m: a.m, dataRead: campaignTestRead, savePath: filepath.Join(t.TempDir(), "restored.gesave")}
	if err := b.loadCampaign(raw); err != nil {
		t.Fatal(err)
	}
	if strategyState(t, a) != strategyState(t, b) || !b.battle.paused {
		t.Fatal("save lost cursor, report, army or RNG")
	}
	year, month, armyMonth := a.year, a.month, a.armory.LastMonth
	seed := a.rng.Seed()
	battles := 0
	for a.strategyTurn != nil && battles < 100 {
		if !a.activeBattle() || !b.activeBattle() || !reflect.DeepEqual(a.strategyTurn, b.strategyTurn) {
			t.Fatal("continuations disagree before battle")
		}
		if a.year != year || a.month != month || a.armory.LastMonth != armyMonth {
			t.Fatal("month advanced before all computer actions finished")
		}
		for _, x := range []*app{a, b} {
			if len(x.battle.retreatTargets) > 0 {
				if err := x.withdrawDefenders(x.battle.retreatTargets[0]); err != nil {
					t.Fatal(err)
				}
			} else {
				x.battle.turn = x.battle.turnLimit()
				x.battle.finishByTurnLimit()
				if err := x.persistFinishedBattle(); err != nil {
					t.Fatal(err)
				}
			}
		}
		before := strategyState(t, b)
		called := false
		b.withBudget(b.current, func() string { called = true; return "unexpected command" })
		b.armoryAction("research")
		if called || strategyState(t, b) != before {
			t.Fatal("new transaction changed an interrupted AI turn")
		}
		if err := b.autosave(); err != nil {
			t.Fatal(err)
		}
		raw, _ := os.ReadFile(b.savePath)
		if err := b.loadCampaign(raw); err != nil {
			t.Fatal(err)
		}
		if strategyState(t, b) != before {
			t.Fatal("restart repeated finished battle settlement")
		}
		a.endTurn()
		b.endTurn()
		if strategyState(t, a) != strategyState(t, b) {
			t.Fatal("save/restart repeated or skipped computer actions")
		}
		battles++
	}
	if a.strategyTurn != nil || battles == 0 || a.armory.LastMonth != armyMonth+1 || a.rng.Seed() != seed {
		t.Fatalf("turn failed to complete once: battles=%d, turn=%+v", battles, a.strategyTurn)
	}
	if month < 12 && (a.year != year || a.month != month+1) {
		t.Fatal("month did not advance exactly once")
	}
	t.Logf("%d controlled-input defensive battles; remaining real AI resumed after restart; month settled once", battles)
}

func TestStrategyTurnTwoInterruptionsKeepRemainingBudget(t *testing.T) {
	a, _ := firstStrategyAttack(t)
	turn := a.strategyTurn
	if turn.Province != 7 || turn.Remaining != a.world.CommandsFor(7)-1 {
		t.Fatal("attack was not charged once")
	}
	remaining := turn.Remaining
	reported := len(turn.Report.Attacks)
	a.battle.sim.EndTurn()
	a.battle.forceDirty = false
	lastMorale := a.battle.sim.Defender[0].Strength.F30
	defender := a.battle.sim.Defender[0].General
	a.battle.turn = a.battle.turnLimit()
	a.battle.finishByTurnLimit()
	if err := a.finishComputerBattle(); err != nil {
		t.Fatal(err)
	}
	if len(turn.Report.Attacks) != reported+1 {
		t.Fatal("first battle result missing")
	}
	if a.generals[defender-1].F30 != lastMorale || a.world.Strengths[defender-1].F30 != lastMorale {
		t.Fatal("continuation discarded the final tactical morale change")
	}
	a.runComputerActions(turn, func(p game.ProvinceID, rep *computerTurnReport) bool {
		if p != 7 {
			t.Fatal("skipped the province's remaining action")
		}
		return a.computerAttack(game.AIAction{Kind: game.AIAttack, From: 7, To: 11}, 107, rep)
	})
	if !a.activeBattle() || turn.Remaining != remaining-1 || len(turn.Report.Attacks) != reported+1 {
		t.Fatal("second interruption reset or replayed actions")
	}
	a.battle.turn = a.battle.turnLimit()
	a.battle.finishByTurnLimit()
	if err := a.persistFinishedBattle(); err != nil {
		t.Fatal(err)
	}
	a.pauseBattle()
	if a.battle == nil || !a.battle.paused {
		t.Fatal("finished battle return discarded pending result")
	}
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(a.savePath)
	if err := a.loadCampaign(raw); err != nil {
		t.Fatal(err)
	}
	if err := a.finishComputerBattle(); err != nil {
		t.Fatal(err)
	}
	if len(a.strategyTurn.Report.Attacks) != reported+2 {
		t.Fatal("completed battle report duplicated or lost")
	}
	if err := a.finishComputerBattle(); err != nil {
		t.Fatal(err)
	}
	if len(a.strategyTurn.Report.Attacks) != reported+2 {
		t.Fatal("settled result counted twice")
	}
}

func TestStrategyTurnMonthFailureRetryAndYearBoundary(t *testing.T) {
	a := declaredStrategyApp(t)
	a.month = 12
	a.armory.LastMonth = uint32(a.year)*12 + 12
	a.strategyTurn = &strategyTurn{Year: a.year, Month: 12, Player: 1, Province: game.ProvinceID(a.provinceLimit + 1)}
	// Deliberately damage only the mock armory ledger to exercise the error
	// path after a finished AI loop. No gameplay fixture uses this corruption.
	original := a.armory.Produced["rifle-old"]
	a.armory.Produced["rifle-old"] = ^uint64(0)
	before := strategyState(t, a)
	a.endTurn()
	if strategyState(t, a) != before || a.strategyTurn == nil {
		t.Fatal("failed settlement partially changed month, world or cursor")
	}
	a.armory.Produced["rifle-old"] = original
	year, settled, seed := a.year, a.armory.LastMonth, a.rng.Seed()
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(a.savePath)
	if err := a.loadCampaign(raw); err != nil {
		t.Fatal(err)
	}
	a.endTurn()
	if a.strategyTurn != nil || a.year != year+1 || a.month != 1 || a.armory.LastMonth != settled+1 || a.rng.Seed() != seed {
		t.Fatal("retry reran AI or failed to settle year boundary once")
	}
}

func TestStrategyTurnRejectsMalformedCursorAtomically(t *testing.T) {
	a, _ := firstStrategyAttack(t)
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	raw, _ := os.ReadFile(a.savePath)
	for _, change := range []string{"year", "month", "player", "province", "remaining", "started", "rule", "role", "source", "report", "missing-battle"} {
		var options campaignV4
		legacy, err := game.DecodeCampaignSaveVersion(raw, &options, 4)
		if err != nil {
			t.Fatal(err)
		}
		x := options.Base.StrategyTurn
		switch change {
		case "year":
			x.Year++
		case "month":
			x.Month++
		case "player":
			x.Player = 107
		case "province":
			x.Province = 0
		case "remaining":
			x.Remaining = 1000
		case "started":
			x.Started = false
		case "rule":
			options.Rules.DeclaredAt = 0
		case "role":
			options.Base.Battle.HumanDefender = false
		case "source":
			x.Province++
		case "report":
			x.Report.Comforts = -1
		case "missing-battle":
			options.Base.Battle = nil
		}
		bad, err := game.EncodeCampaignSaveVersion(options, legacy, 4)
		if err != nil {
			t.Fatal(err)
		}
		before := strategyState(t, a)
		if err := a.loadCampaign(bad); err == nil {
			t.Fatal("accepted malformed cursor", change)
		}
		if strategyState(t, a) != before {
			t.Fatal("bad cursor partially replaced live game", change)
		}
	}
}

func TestStrategyTurnLegacyLoopMatchesExistingActions(t *testing.T) {
	for _, stage := range []uint8{1, 2, 3} {
		a, b := newArmoryApp(t, stage, 1), newArmoryApp(t, stage, 1)
		var rep computerTurnReport
		for p := game.ProvinceID(1); int(p) <= b.provinceLimit; p++ {
			prov, _ := b.tbl.At(p)
			if prov.Commander == 0 || prov.Commander == b.playerCommander {
				continue
			}
			budget := b.world.CommandsFor(p)
			for i := 0; i < budget; i++ {
				if !b.runOneComputerAction(p, &rep) {
					break
				}
			}
		}
		a.runComputerTurn(a.playerCommander)
		if strategyState(t, a) != strategyState(t, b) || a.battle != nil || b.battle != nil {
			t.Fatal("resumable loop changed undeclared AI actions", stage)
		}
	}
}

func TestStrategyTurnModernConfirmationKeepsBattleScreen(t *testing.T) {
	_, before := firstStrategyAttack(t)
	a := &app{m: campaignTestMap(t), dataRead: campaignTestRead}
	if err := a.loadCampaign(before); err != nil {
		t.Fatal(err)
	}
	a.desk.Pending = "month"
	if _, err := a.deskConfirm(); err != nil {
		t.Fatal(err)
	}
	if !a.activeBattle() || a.screen != screenBattle {
		t.Fatal("Modern confirmation covered new battle with map")
	}
}

func TestStrategyTurnMessagesHaveGlyphs(t *testing.T) {
	fonts, err := assets.LoadEtenFonts("../../workplace/eten")
	if err != nil {
		t.Fatal(err)
	}
	for _, lang := range []string{"zh-Hant", "en", "ja"} {
		wording, err := i18n.LoadWording("../../translations/" + lang)
		if err != nil {
			t.Fatal(err)
		}
		for _, key := range []string{"pending", "continue", "defend", "failed"} {
			value, ok := wording.Text("military.turn."+key, i18n.WordingOriginal)
			if !ok || strings.TrimSpace(value) == "" {
				t.Fatal(lang, key, ok)
			}
			missing := render.NewBGICanvas().DrawMessageOverlay(fonts, assets.RGB{R: 230, G: 220, B: 200}, assets.RGB{R: 30, G: 25, B: 20}, 0, 0, 640, 100, value)
			if len(missing) > 0 {
				t.Fatalf("%s %s missing %q", lang, key, string(missing))
			}
		}
	}
}

func TestStrategyTurnSampleExport(t *testing.T) {
	dir := os.Getenv("DSDS_STRATEGY_SAMPLE_DIR")
	if dir == "" {
		return
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		t.Fatal("existing output directory required", err)
	}
	a, before := firstStrategyAttack(t)
	if err := os.WriteFile(filepath.Join(dir, "resume-selected-attack.gesave"), before, 0600); err != nil {
		t.Fatal(err)
	}
	meta := map[string]any{"scope": "real 1937 data, legal Chinese partial transfer, premarked declaration and controlled selected Japanese attack; not normal AI target selection, declaration or finite-wave receipt", "seed": 1, "year": a.year, "month": a.month, "from": a.battle.sim.From, "at": a.battle.sim.At, "retreat_targets": a.battle.retreatTargets, "cursor": a.strategyTurn}
	raw, err := json.MarshalIndent(meta, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "resume-sample.json"), raw, 0644); err != nil {
		t.Fatal(err)
	}
}
