package main

import (
	"fmt"

	"github.com/wicanr2/great-era-remake/internal/game"
)

func (b *battleState) humanUnits() []*game.Combatant {
	if b == nil || b.sim == nil {
		return nil
	}
	if b.humanDefender {
		return b.sim.Defender
	}
	return b.sim.Attacker
}

func (b *battleState) enemyUnits() []*game.Combatant {
	if b == nil || b.sim == nil {
		return nil
	}
	if b.humanDefender {
		return b.sim.Attacker
	}
	return b.sim.Defender
}

// The resumable strategy-AI and invasion event will call this explicit entry.
// It cannot be used to give a player control of someone else's defending army.
func (a *app) startDefensiveBattle(at, from game.ProvinceID) error {
	return a.startDefensiveBattleWithRoster(at, from, nil)
}

func (a *app) startDefensiveBattleWithRoster(at, from game.ProvinceID, roster []game.GeneralID) error {
	if a == nil || a.tbl == nil {
		return fmt.Errorf("defensive battle requires a campaign world")
	}
	p, err := a.tbl.At(at)
	if err != nil {
		return err
	}
	if a.campaign == nil || p.Commander != a.playerCommander {
		return fmt.Errorf("defensive battle requires a player-owned target province")
	}
	source, err := a.tbl.At(from)
	if err != nil {
		return err
	}
	if source.Commander == 0 || source.Commander == a.playerCommander {
		return fmt.Errorf("defensive battle requires a hostile source province")
	}
	if err := a.startBattleWithRoster(at, from, roster); err != nil {
		return err
	}
	a.battle.humanDefender = true
	a.current = at // Original strategy panel must describe the defended province.
	a.battle.retreatTargets = game.DefenderRetreatTargets(a.tbl, at, a.playerCommander)
	return nil
}

func (a *app) withdrawDefenders(to game.ProvinceID) error {
	b := a.battle
	if b == nil || !b.humanDefender || b.finished || a.world == nil {
		return fmt.Errorf("invalid player defender withdrawal")
	}
	// EndTurn changes morale after the usual force synchronization. Include
	// that final combat state before moving survivors out of this battlefield.
	b.forceDirty = true
	if err := a.syncBattleForces(); err != nil {
		return err
	}
	if b.externalInvasion {
		if err := a.settleExternalInvasion(to); err != nil {
			return err
		}
		b.finishWithWinner(game.BattleSideFirst)
		b.paused = true
		b.mode, b.retreatInput, b.retreatTargets = battleModeCommand, 0, nil
		a.current, a.screen, a.dirty = to, screenMap, true
		a.applyWindowSize()
		return nil
	}
	commitSupply, err := a.prepareJapaneseSupply(b)
	if err != nil {
		return err
	}
	if err := game.ApplyDefenderWithdrawal(b.sim, a.tbl, a.generals, a.world.Units, to); err != nil {
		return err
	}
	commitSupply()
	b.finishWithWinner(game.BattleSideFirst)
	b.settled = true
	b.paused = true // The completed withdrawal returns to strategy, including restart.
	b.mode, b.retreatInput, b.retreatTargets = battleModeCommand, 0, nil
	a.current = to
	a.screen, a.dirty = screenMap, true
	a.applyWindowSize()
	return nil
}
