package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
)

// Real scenario data and normal battle construction, not a normal campaign AI
// receipt. The invasion/resumable strategy turn is deliberately not connected.
func defensiveTestApp(t *testing.T) *app {
	t.Helper()
	a := newArmoryApp(t, 2, 1)
	for from := game.ProvinceID(1); int(from) <= a.provinceLimit; from++ {
		p, _ := a.tbl.At(from)
		if p.Commander != 107 {
			continue
		}
		for _, at := range p.Neighbours {
			q, _ := a.tbl.At(at)
			if q.Commander != 1 {
				continue
			}
			if err := a.startDefensiveBattle(at, from); err == nil {
				return a
			}
		}
	}
	t.Fatal("real 1937 scenario has no constructible Japan-to-China battle")
	return nil
}

func TestDefensiveControlAndSaveKeepActualBattleSides(t *testing.T) {
	a := defensiveTestApp(t)
	b := a.battle
	if a.current != b.sim.At {
		t.Fatal("original strategy panel describes a different province than the defensive battle")
	}
	if !b.humanDefender || !reflect.DeepEqual(b.retreatTargets, game.DefenderRetreatTargets(a.tbl, b.sim.At, 1)) {
		t.Fatal("defensive role or friendly retreat targets lost")
	}
	if b.current().Faction != a.playerCommander || b.sim.Attacker[0].Faction != 107 {
		t.Fatal("player does not control actual defenders")
	}
	for _, u := range b.humanUnits() {
		if u.Attacking || u.Faction != 1 {
			t.Fatal("defenders were exchanged with attackers")
		}
	}
	for _, u := range b.enemyUnits() {
		if !u.Attacking || u.Faction != 107 {
			t.Fatal("attackers were exchanged with defenders")
		}
	}
	panel := a.battlePanelData()
	if panel.Attacker.Leader != 107 || panel.Defender.Leader != 1 {
		t.Fatal("inspection exchanged actual leaders")
	}
	b.sel = len(b.sim.Defender) - 1
	selected := b.current().General
	a.pauseBattle()
	state := a.campaignState()
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	restored := &app{m: a.m, dataRead: campaignTestRead}
	if err := restored.loadCampaign(raw); err != nil {
		t.Fatal(err)
	}
	if !restored.battle.humanDefender || restored.battle.current().General != selected || !restored.battle.paused || !reflect.DeepEqual(state, restored.campaignState()) {
		t.Fatal("save/restart lost human side, selected unit or world")
	}
	for _, change := range []string{"owner", "province-owner", "defender-faction", "attacker-faction", "selection", "retreat"} {
		var options campaignV4
		legacy, err := game.DecodeCampaignSaveVersion(raw, &options, 4)
		if err != nil {
			t.Fatal(err)
		}
		x := options.Base.Battle
		switch change {
		case "owner":
			x.AtCommander = 107
		case "province-owner":
			options.Base.State.Provinces.Province[x.At-1].Commander = 107
		case "defender-faction":
			x.Defender[0].Faction = 107
		case "attacker-faction":
			x.Attacker[0].Faction = 1
		case "selection":
			x.Selected = len(x.Defender)
		case "retreat":
			x.RetreatTargets = []game.ProvinceID{7}
		}
		bad, err := game.EncodeCampaignSaveVersion(options, legacy, 4)
		if err != nil {
			t.Fatal(err)
		}
		before, _ := json.Marshal(restored.campaignState())
		if err := restored.loadCampaign(bad); err == nil {
			t.Fatalf("accepted invalid human-side %s", change)
		}
		after, _ := json.Marshal(restored.campaignState())
		if string(before) != string(after) {
			t.Fatal("failed restore changed live state")
		}
	}
	t.Logf("real 1937 battle %d -> %d: selected defender %d, pause/save/restart and both roles verified", b.sim.From, b.sim.At, selected)
}

func TestDefensiveComputerTurnOnlyOrdersAttackers(t *testing.T) {
	a := defensiveTestApp(t)
	b := a.battle
	before, _ := json.Marshal(b.sim.Defender)
	b.runDefenderAI()
	// At the initial deployment the sides are not adjacent. Thus any change
	// in a defending unit's orders/mobility would be the computer controlling
	// the human army, rather than a legitimate casualty from an engagement.
	after, _ := json.Marshal(b.sim.Defender)
	if string(before) != string(after) {
		t.Fatal("computer ordered the human defenders")
	}
	if b.aiAction == 0 {
		t.Fatal("attacker computer decision was not executed")
	}
	if b.sim.Attacker[0].Faction != 107 || b.sim.Defender[0].Faction != 1 {
		t.Fatal("AI changed sides")
	}
}

func TestDefensiveAdjacentTargetsFollowHumanSide(t *testing.T) {
	a := defensiveTestApp(t)
	b := a.battle
	u := b.current()
	// A narrow target-selection probe, not a deployment or gameplay receipt.
	v := b.enemyUnits()[0]
	for _, c := range u.Cell.Neighbours() {
		v.Cell = c
		break
	}
	got := b.adjacentEnemies(u)
	if len(got) == 0 || got[0].Faction != 107 {
		t.Fatal("human defender targets its own army")
	}
	for _, v := range got {
		if v.Faction == 1 {
			t.Fatal("friendly unit listed as enemy")
		}
	}
}

func TestDefensiveEntryRejectsFriendlySourceWithoutChangingWorld(t *testing.T) {
	a := newArmoryApp(t, 2, 1)
	before, _ := json.Marshal(a.campaignState())
	for _, pair := range [][2]game.ProvinceID{{a.current, a.current}, {7, a.current}} {
		if err := a.startDefensiveBattle(pair[0], pair[1]); err == nil {
			t.Fatal("accepted invalid defensive entry", pair)
		}
	}
	after, _ := json.Marshal(a.campaignState())
	if string(before) != string(after) || a.battle != nil {
		t.Fatal("failed entry changed world or battle")
	}
}

func TestDefensiveWithdrawalAndRestartDoNotRepeatSettlement(t *testing.T) {
	a := defensiveTestApp(t)
	b := a.battle
	if len(b.retreatTargets) == 0 {
		t.Fatal("real fixture needs a friendly retreat province")
	}
	to := b.retreatTargets[0]
	armyBefore, _ := json.Marshal(a.armory)
	seed, budget := a.rng.Seed(), a.cmdBudget.Snapshot()
	var garrison []game.GeneralID
	for i, u := range a.world.Units {
		if u.Province == b.sim.At && u.Faction == 1 && a.generals[i].Force > 0 {
			garrison = append(garrison, u.General)
		}
	}
	b.mode = battleModeRetreat
	b.retreatInput = uint32(to)
	b.sim.EndTurn()
	b.forceDirty = false // Exercise the condition change after the usual sync.
	if err := a.withdrawDefenders(to); err != nil {
		t.Fatal(err)
	}
	if !b.finished || !b.settled || b.settlement.Winner != game.BattleSideFirst || a.current != to || a.screen != screenMap {
		t.Fatal("withdrawal did not finish at the chosen province")
	}
	p, _ := a.tbl.At(b.sim.At)
	if p.Commander != 107 || p.InBattle() {
		t.Fatal("attacker did not capture after withdrawal")
	}
	for _, id := range garrison {
		if a.generals[id-1].Province != to || a.world.Units[id-1].Province != to || a.world.Units[id-1].Faction != 1 {
			t.Fatal("living garrison left behind", id)
		}
	}
	for _, u := range b.sim.Defender {
		if a.generals[u.General-1].F30 != u.Strength.F30 || a.world.Strengths[u.General-1].F30 != u.Strength.F30 {
			t.Fatal("withdrawal lost the last tactical turn's morale", u.General)
		}
	}
	armyAfter, _ := json.Marshal(a.armory)
	if string(armyBefore) != string(armyAfter) || seed != a.rng.Seed() || !reflect.DeepEqual(budget, a.cmdBudget.Snapshot()) {
		t.Fatal("withdrawal changed equipment, RNG or command budget")
	}
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	restored := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
	if err := restored.loadCampaign(raw); err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(restored.campaignState())
	if err := restored.persistFinishedBattle(); err != nil {
		t.Fatal(err)
	}
	after, _ := json.Marshal(restored.campaignState())
	if string(before) != string(after) || restored.current != to || !restored.battle.humanDefender {
		t.Fatal("restart repeated settlement or changed human side")
	}
	t.Logf("real 1937 withdrawal: %d surviving garrison units moved from %d to %d; Japan captured; save/restart did not replay", len(garrison), b.sim.At, to)
}

// Explicit local QA export only. The normal executable has no debug entry or
// resource injection flag. This fixture is not a declaration/normal AI receipt.
func TestDefensiveSampleExport(t *testing.T) {
	dir := os.Getenv("DSDS_DEFENDER_SAMPLE_DIR")
	if dir == "" {
		return
	}
	info, err := os.Stat(dir)
	if err != nil || !info.IsDir() {
		t.Fatal("sample output directory must already exist", err)
	}
	a := defensiveTestApp(t)
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "late-defender-start.gesave"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	l, err := uilayout.NewEarthBattleLayout(uilayout.ModernDesignSurface())
	if err != nil {
		t.Fatal(err)
	}
	var enemies [][2]float64
	for _, u := range a.battle.enemyUnits() {
		col, row := u.Cell.ColRow()
		x, y := l.CellCenter(col, row)
		enemies = append(enemies, [2]float64{x, y})
	}
	b, err := json.MarshalIndent(map[string]any{"scope": "real 1937 data, programmatically constructed defender battle; not normal declaration or strategy AI", "from": a.battle.sim.From, "at": a.battle.sim.At, "layout": l, "enemy_centres": enemies, "retreat_targets": a.battle.retreatTargets}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "late-defender-sample.json"), b, 0644); err != nil {
		t.Fatal(err)
	}
}

func TestDefensiveWithdrawalMessagesHaveGlyphs(t *testing.T) {
	fonts, err := assets.LoadEtenFonts("../../workplace/eten")
	if err != nil {
		t.Fatal(err)
	}
	for _, lang := range []string{"zh-Hant", "en", "ja"} {
		wording, err := i18n.LoadWording("../../translations/" + lang)
		if err != nil {
			t.Fatal(err)
		}
		a := &app{wording: wording, wordingMode: i18n.WordingOriginal}
		for _, key := range []string{"withdraw.unavailable", "withdraw.failed"} {
			text := a.militaryText(key)
			missing := render.NewBGICanvas().DrawMessageOverlay(fonts, assets.RGB{R: 230, G: 220, B: 200}, assets.RGB{R: 30, G: 25, B: 20}, 0, 0, 640, 100, text)
			if len(missing) > 0 {
				t.Fatalf("%s %s missing %q", lang, key, string(missing))
			}
		}
	}
}
