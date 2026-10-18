package main

import (
	"fmt"
	"os"
	"sort"

	"github.com/wicanr2/great-era-remake/internal/game"
)

// strategyTurn retains the next unexecuted action across a defensive battle.
type strategyTurn struct {
	Year            uint16
	Month           uint8
	Player          game.GeneralID
	Province        game.ProvinceID
	Started         bool
	Remaining       int
	Report          computerTurnReport
	JapaneseAttacks map[game.GeneralID]uint8 `json:",omitempty"`
}

func (a *app) strategyLocked() bool {
	return a.activeBattle() || a.strategyTurn != nil || a.invasionEnded()
}

func (a *app) strategyLockText() string {
	if a.invasionEnded() {
		return a.invasionText("ended")
	}
	if a.activeBattle() {
		return a.deskText("battle.locked")
	}
	return a.militaryText("turn.pending")
}

func (a *app) advanceComputerTurn(turn *strategyTurn) {
	a.runComputerActions(turn, a.runOneComputerAction)
}

// The executor boundary lets tests replay an explicit selected action without
// changing production AI thresholds or adding executable debug flags.
func (a *app) runComputerActions(turn *strategyTurn, execute func(game.ProvinceID, *computerTurnReport) bool) {
	limit := a.provinceLimit
	if limit == 0 {
		limit = game.ProvinceCount
	}
	for int(turn.Province) <= limit {
		p := turn.Province
		if !turn.Started {
			prov, err := a.tbl.At(p)
			if err != nil || prov.Commander == 0 || prov.Commander == turn.Player {
				turn.Province++
				continue
			}
			turn.Started, turn.Remaining = true, a.world.CommandsFor(p)
		}
		if turn.Remaining > 0 {
			acted := execute(p, &turn.Report)
			turn.Remaining--
			if !acted {
				turn.Remaining = 0
			}
			if a.activeBattle() {
				return
			}
			continue
		}
		turn.Province++
		turn.Started = false
	}
	sort.Strings(turn.Report.Attacks)
}

func (a *app) computerAttack(action game.AIAction, faction game.GeneralID, rep *computerTurnReport) bool {
	return a.computerAttackSelected(action, faction, rep, nil)
}

func (a *app) computerAttackSelected(action game.AIAction, faction game.GeneralID, rep *computerTurnReport, roster []game.GeneralID) bool {
	if !a.japaneseOffensiveAllowed(faction) {
		return false
	}
	target, err := a.tbl.At(action.To)
	if err != nil {
		return false
	}
	humanDefense := a.strategyTurn != nil && a.campaign != nil && (a.stage > 1 || a.invasionControlled()) &&
		a.militaryEnabled() && a.militaryRules.Locked() && faction == a.invasionFaction() &&
		target.Commander == a.playerCommander
	if humanDefense {
		for i, u := range a.world.Units {
			if !u.Active || u.Province != action.To || u.Faction != a.playerCommander || i >= len(a.world.Strengths) || a.world.Strengths[i].Force == 0 {
				continue
			}
			if err := a.startDefensiveBattleWithRoster(action.To, action.From, roster); err != nil {
				fmt.Fprintf(os.Stderr, "defensive battle %d -> %d: %v\n", action.From, action.To, err)
				a.report(a.militaryText("turn.failed"))
				return false
			}
			return true
		}
	}
	var out game.BattleOutcome
	if roster != nil {
		out, err = a.world.ResolveSelectedAttack(action.From, action.To, roster)
	} else {
		out, err = a.world.ResolveAttack(action.From, action.To)
	}
	if err != nil {
		return false
	}
	rep.Attacks = append(rep.Attacks, fmt.Sprintf("%s 攻 %s（%s）",
		a.provinceName(action.From), a.provinceName(action.To), outcomeWord(out)))
	return true
}

// A completed defensive battle is reported once before the next AI action.
func (a *app) finishComputerBattle() error {
	b := a.battle
	if b == nil || !b.humanDefender || !b.finished {
		return nil
	}
	// The last tactical EndTurn can change morale after the usual force sync.
	// Commit those final conditions before AI resumes or this copy is cleared.
	b.forceDirty = true
	if err := a.syncBattleForces(); err != nil {
		return err
	}
	if err := a.persistFinishedBattle(); err != nil {
		return err
	}
	if a.strategyTurn != nil {
		out := game.BattleOutcome{Decided: b.settlement.Kind == game.BattleSettlementDecisive,
			AttackerWon: b.settlement.Winner == game.BattleSideFirst}
		a.strategyTurn.Report.Attacks = append(a.strategyTurn.Report.Attacks,
			fmt.Sprintf("%s 攻 %s（%s）", a.provinceName(b.sim.From), a.provinceName(b.sim.At), outcomeWord(out)))
	}
	a.battle = nil
	return nil
}

func validStrategyTurn(turn *strategyTurn, saved *campaignSnapshot, options *campaignOptions, sc game.Scenario) bool {
	if turn == nil {
		return saved.Battle == nil || !saved.Battle.JapaneseBlitz
	}
	cap := sc.Generals/8 + 1
	faction, factionErr := game.JapaneseInvasionFaction(sc.Stage)
	if options == nil || !options.Rules.Enabled || !options.Rules.Locked() || factionErr != nil ||
		sc.Stage == 1 && len(saved.Units) < 300 || saved.Identity.Player == faction || turn.Player != saved.Identity.Player ||
		turn.Year != saved.Year || turn.Month != saved.Month || turn.Province < 1 ||
		int(turn.Province) > sc.Provinces+1 || turn.Remaining < 0 || turn.Remaining > cap ||
		!turn.Started && turn.Remaining != 0 || turn.Started && int(turn.Province) > sc.Provinces ||
		saved.Reorganization != nil || options.View.Open {
		return false
	}
	if b := saved.Battle; b != nil {
		if !b.HumanDefender || !turn.Started || b.From != turn.Province || b.Settled && !b.Finished || b.Finished && b.Settlement.Kind != game.BattleSettlementDecisive && b.Settlement.Kind != game.BattleSettlementTurnLimitDraw {
			return false
		}
		for _, u := range b.Attacker {
			if u == nil || u.Faction != faction {
				return false
			}
		}
	} else if int(turn.Province) != sc.Provinces+1 || turn.Started {
		return false
	}
	maximum := sc.Provinces * cap
	offshore, err := options.Armory.InvasionMonthlyAttacks(uint32(saved.Year)*12 + uint32(saved.Month))
	if err != nil {
		return false
	}
	for id, count := range offshore {
		if turn.JapaneseAttacks[id] < count {
			return false
		}
	}
	for id, count := range turn.JapaneseAttacks {
		if id < 1 || int(id) > sc.Generals || count < 1 || count > game.JapaneseAttackLimit ||
			int(id) > len(saved.Units) || saved.Units[id-1].Faction != faction {
			return false
		}
	}
	if b := saved.Battle; b != nil && b.JapaneseBlitz {
		for _, u := range b.Attacker {
			if turn.JapaneseAttacks[u.General] < offshore[u.General]+1 {
				return false
			}
		}
	}
	r := turn.Report
	if r.Comforts < 0 || r.Transfers < 0 || r.Comforts > maximum || r.Transfers > maximum {
		return false
	}
	for _, items := range [][]string{r.Attacks, r.Relocations, r.Recruits} {
		if len(items) > maximum {
			return false
		}
		for _, text := range items {
			if len(text) > 4096 {
				return false
			}
		}
	}
	return true
}
