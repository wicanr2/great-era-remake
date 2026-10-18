package main

import (
	"fmt"
	"reflect"

	"github.com/wicanr2/great-era-remake/internal/game"
)

// The normal declaration/cast controller will use this entry. It never creates
// people or marks a declaration on its own, and cannot borrow provincial supply.
func (a *app) startExternalInvasionBattle() error {
	if a == nil || a.world == nil || a.armory == nil || a.activeBattle() || a.strategyTurn != nil ||
		a.campaign == nil || a.declaration == nil || a.declaration.Imported ||
		!a.militaryEnabled() || !a.militaryRules.Locked() {
		return game.ErrInvasionLanding
	}
	p, err := a.tbl.At(game.InvasionEntryProvince)
	if err != nil || p.Commander != a.playerCommander && !(a.invasionControlled() && p.Commander == 0) || p.InBattle() {
		return game.ErrInvasionLanding
	}
	next := a.armory.Clone()
	ids, err := next.BeginInvasionLanding(a.world)
	if err != nil {
		return err
	}
	faction := next.InvasionBatches[1].Faction
	atk, _, err := a.world.SelectedCombatants(game.ExternalInvasionSource, faction, ids)
	if err != nil {
		return err
	}
	var defenderIDs []game.GeneralID
	for i, u := range a.world.Units {
		if u.Active && u.Province == game.InvasionEntryProvince && u.Faction == p.Commander && a.world.Strengths[i].Force > 0 {
			defenderIDs = append(defenderIDs, u.General)
			if len(defenderIDs) == game.UnitsPerSide {
				break
			}
		}
	}
	var def []*game.Combatant
	if len(defenderIDs) == 0 && !a.invasionControlled() {
		return game.ErrInvasionLanding
	}
	if len(defenderIDs) > 0 {
		def, _, err = a.world.SelectedCombatants(game.InvasionEntryProvince, p.Commander, defenderIDs)
		if err != nil {
			return err
		}
	}
	zone, err := a.m.DefenderDeployZone(game.InvasionEntryProvince)
	if err != nil || len(zone) < len(def) {
		return game.ErrInvasionLanding
	}
	for i, unit := range def {
		unit.Cell = zone[i]
	}
	opt := a.world.Opts
	opt.Armory = a.armory
	sim, err := game.NewExternalBattleSim(a.m, game.InvasionEntryProvince, atk, def, opt)
	if err != nil {
		return err
	}
	sim.AtCommander = p.Commander
	sim.BeginTurn()
	supply, err := next.InvasionOffshoreSupply()
	if err != nil {
		return err
	}
	b := &battleState{sim: sim, turn: 1, turnCap: game.BattleTurnCap(a.month == 2), mode: battleModeCommand,
		leader: p.Commander, tbl: a.tbl, units: a.world.Units, enableLastSteps: a.world.EnableExtra,
		humanDefender: len(def) > 0, externalInvasion: true}
	b.retreatTargets = game.DefenderRetreatTargets(a.tbl, game.InvasionEntryProvince, p.Commander)
	b.supAtk = game.BattleSupply{Gold: int(supply.Gold), Food: int(supply.Food), Ammo: int(supply.Ammo),
		Fuel: int(supply.Fuel), Troops: game.TroopTotal(atk)}
	b.supDef = supplyOf(a.tbl, game.InvasionEntryProvince, def)
	if faction == a.playerCommander {
		return game.ErrInvasionLanding
	}
	*a.armory = *next
	b.sim.Opts.Armory = a.armory
	p.Flags |= game.ProvinceFlagInBattle
	a.battle, a.current = b, game.InvasionEntryProvince
	a.screen, a.dirty = screenBattle, true
	a.applyWindowSize()
	if len(def) == 0 {
		b.finished = true
		b.settlement = game.BattleSettlement{Kind: game.BattleSettlementDecisive, Winner: game.BattleSideFirst}
		if err := a.settleExternalInvasion(0); err != nil {
			return err
		}
	}
	return nil
}

// Settlement and defender withdrawal share a transaction on private world,
// armory and person copies. A failed validation cannot half-credit a landing.
func (a *app) settleExternalInvasion(to game.ProvinceID) error {
	b := a.battle
	if b == nil || !b.externalInvasion || b.sim == nil || a.armory == nil || a.world == nil {
		return game.ErrInvasionLanding
	}
	if b.settled {
		return nil
	}
	if err := a.syncBattleForces(); err != nil {
		return err
	}
	settlement := b.settlement
	if to != 0 {
		settlement = game.BattleSettlement{Kind: game.BattleSettlementDecisive, Winner: game.BattleSideFirst}
	}
	army := a.armory.Clone()
	world := *a.world
	table := *a.tbl
	world.Table, world.Opts.Armory = &table, army
	world.Units = append([]game.CombatUnit(nil), a.world.Units...)
	world.Strengths = append([]game.StrengthInput(nil), a.world.Strengths...)
	gs := append([]game.General(nil), a.generals...)
	if err := army.Sync(&world); err != nil {
		return err
	}
	if army.Landing == nil {
		return game.ErrInvasionLanding
	}
	last := army.Landing.Attempts[len(army.Landing.Attempts)-1]
	if last.Outcome != game.LandingActive || !sameInvasionRoster(last.Units, b.sim.Attacker) {
		return game.ErrInvasionLanding
	}
	available, err := army.InvasionOffshoreSupply()
	if err != nil {
		return err
	}
	left := [4]int{b.supAtk.Gold, b.supAtk.Food, b.supAtk.Ammo, b.supAtk.Fuel}
	remaining := [4]uint32{available.Gold, available.Food, available.Ammo, available.Fuel}
	var spent [4]uint32
	for i := range remaining {
		if left[i] < 0 || uint64(left[i]) > uint64(remaining[i]) {
			return fmt.Errorf("invalid external supply remainder")
		}
		spent[i] = remaining[i] - uint32(left[i])
	}
	for number := uint8(1); int(number) <= len(army.InvasionBatches); number++ {
		cap, _ := army.InvasionRemainingSupply(number)
		caps := [4]uint32{cap.Gold, cap.Food, cap.Ammo, cap.Fuel}
		var take [4]uint32
		for i := range take {
			take[i] = min(spent[i], caps[i])
			spent[i] -= take[i]
		}
		if err := army.SpendInvasionSupply(&world, number, game.InvasionSupply{Gold: take[0], Food: take[1], Ammo: take[2], Fuel: take[3]}); err != nil {
			return err
		}
	}
	p, err := table.At(game.InvasionEntryProvince)
	if err != nil {
		return err
	}
	def := [4]int{b.supDef.Gold, b.supDef.Food, b.supDef.Ammo, b.supDef.Fuel}
	original := [4]uint16{p.Gold, p.Food, p.Ammo, p.Fuel}
	for i, value := range def {
		if value < 0 || value > int(original[i]) {
			return fmt.Errorf("invalid defender supply remainder")
		}
	}
	faction := army.InvasionBatches[1].Faction
	if to != 0 {
		err = game.ApplyDefenderWithdrawal(b.sim, &table, gs, world.Units, to)
	} else {
		err = game.ApplyExternalBattleSettlement(b.sim, &table, gs, world.Units, faction, settlement)
	}
	if err != nil {
		return err
	}
	p.Gold, p.Food, p.Ammo, p.Fuel = uint16(def[0]), uint16(def[1]), uint16(def[2]), uint16(def[3])
	if err := army.FinishInvasionLanding(&world, settlement); err != nil {
		return err
	}
	if settlement.Kind == game.BattleSettlementDecisive && settlement.Winner == game.BattleSideFirst {
		for number := uint8(1); int(number) <= len(army.InvasionBatches); number++ {
			if _, err := army.DeliverInvasionSupply(&world, number); err != nil {
				return err
			}
		}
	}
	*a.tbl, *a.armory = table, *army
	copy(a.generals, gs)
	copy(a.world.Units, world.Units)
	b.settled = true
	return nil
}

func sameInvasionRoster(ids []game.GeneralID, units []*game.Combatant) bool {
	got := make([]game.GeneralID, len(units))
	for i, u := range units {
		if u == nil {
			return false
		}
		got[i] = u.General
	}
	return reflect.DeepEqual(ids, got)
}

func validCampaignLanding(v *campaignV4) bool {
	b := v.Base.Battle
	external := b != nil && b.ExternalInvasion
	if v.Armory == nil || v.Armory.Landing == nil {
		return !external
	}
	attempts := v.Armory.Landing.Attempts
	if len(attempts) == 0 || v.Declaration == nil || v.Declaration.Imported {
		return false
	}
	last := attempts[len(attempts)-1]
	if last.Outcome == game.LandingActive && (!external || b.Finished || b.Settled) {
		return false
	}
	for _, actor := range v.Actors {
		if actor.General.Force == 0 {
			continue
		}
		if last.Outcome == game.LandingLanded {
			if actor.Unit.Province == game.ExternalInvasionSource {
				return false
			}
		} else if actor.Unit.Province != game.ExternalInvasionSource ||
			last.Outcome == game.LandingWithdrawn && actor.Unit.Active {
			return false
		}
		if last.Outcome == game.LandingActive {
			found := false
			for _, id := range last.Units {
				found = found || id == actor.ID
			}
			if !found {
				return false
			}
		}
	}
	if !external {
		return true
	}
	if b.At != game.InvasionEntryProvince || b.From != game.ExternalInvasionSource || !b.HumanDefender ||
		b.JapaneseBlitz || v.Base.StrategyTurn != nil || !sameInvasionRoster(last.Units, b.Attacker) ||
		b.Settled != (last.Outcome != game.LandingActive) || b.Finished != b.Settled {
		return false
	}
	left := [4]int{b.SupplyAttacker.Gold, b.SupplyAttacker.Food, b.SupplyAttacker.Ammo, b.SupplyAttacker.Fuel}
	remaining, err := v.Armory.InvasionOffshoreSupply()
	if err != nil {
		return false
	}
	caps := [4]uint32{remaining.Gold, remaining.Food, remaining.Ammo, remaining.Fuel}
	for i, value := range left {
		if value < 0 || !b.Settled && uint64(value) > uint64(caps[i]) {
			return false
		}
	}
	if b.Settled {
		landed := b.Settlement.Kind == game.BattleSettlementDecisive && b.Settlement.Winner == game.BattleSideFirst
		if landed != (last.Outcome == game.LandingLanded) {
			return false
		}
	}
	return true
}
