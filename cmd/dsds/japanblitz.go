package main

import (
	"fmt"
	"github.com/wicanr2/great-era-remake/internal/game"
)

func (a *app) japaneseBlitzActive() bool {
	return a != nil && a.campaign != nil && a.strategyTurn != nil && (a.stage > 1 || a.invasionControlled()) &&
		a.playerCommander != a.invasionFaction() && a.militaryEnabled() && a.militaryRules.Locked()
}

func (a *app) runJapaneseBlitz(from game.ProvinceID, report *computerTurnReport) bool {
	turn := a.strategyTurn
	faction := a.invasionFaction()
	plan := a.world.PlanJapaneseBlitz(from, faction, turn.JapaneseAttacks, a.provinceLimit)
	switch plan.Action.Kind {
	case game.AITransfer:
		if err := a.world.ApplyBlitzTransfer(plan, faction); err != nil {
			return false
		}
		report.Transfers++
		a.syncExpansionGenerals()
		return true
	case game.AIAttack:
		if !a.computerAttackSelected(plan.Action, faction, report, plan.Units) {
			return false
		}
		if turn.JapaneseAttacks == nil {
			turn.JapaneseAttacks = map[game.GeneralID]uint8{}
		}
		for _, id := range plan.Units {
			turn.JapaneseAttacks[id]++
		}
		if a.activeBattle() {
			a.battle.japaneseBlitz = true
		}
		return true
	}
	return false
}

// Validate both copies before settlement changes ownership or moves armies.
func (a *app) prepareJapaneseSupply(b *battleState) (func(), error) {
	noop := func() {}
	if b == nil || !b.japaneseBlitz {
		return noop, nil
	}
	if a.tbl == nil || b.sim == nil {
		return nil, fmt.Errorf("Japanese battle requires a supply source")
	}
	src, err := a.tbl.At(b.sim.From)
	if err != nil {
		return nil, err
	}
	dst, err := a.tbl.At(b.sim.At)
	if err != nil {
		return nil, err
	}
	valid := func(p *game.Province, s game.BattleSupply) bool {
		return s.Gold >= 0 && s.Gold <= int(p.Gold) && s.Food >= 0 && s.Food <= int(p.Food) &&
			s.Ammo >= 0 && s.Ammo <= int(p.Ammo) && s.Fuel >= 0 && s.Fuel <= int(p.Fuel)
	}
	if !valid(src, b.supAtk) || !valid(dst, b.supDef) {
		return nil, fmt.Errorf("Japanese battle supply exceeds its province ledger")
	}
	if b.settled {
		return noop, nil
	}
	write := func(p *game.Province, s game.BattleSupply) {
		p.Gold, p.Food, p.Ammo, p.Fuel = uint16(s.Gold), uint16(s.Food), uint16(s.Ammo), uint16(s.Fuel)
	}
	return func() { write(src, b.supAtk); write(dst, b.supDef) }, nil
}
