package main

import (
	"fmt"

	"github.com/wicanr2/great-era-remake/internal/game"
)

func (a *app) invasionFaction() game.GeneralID {
	id, _ := game.JapaneseInvasionFaction(uint8(a.stage))
	return id
}

func (a *app) invasionControlled() bool {
	return a != nil && a.declaration != nil && a.declaration.Controller == 1
}

func (a *app) invasionEnded() bool {
	return a.invasionControlled() && a.declaration.Outcome != ""
}

func (a *app) invasionText(key string) string {
	value, err := a.wordingText("invasion." + key)
	if err != nil {
		return "[invasion." + key + "]"
	}
	return value
}

func (a *app) showCampaignEvent() {
	a.screen, a.dirty = screenCampaignEvent, true
	a.cancelGamePointer()
	a.applyWindowSize()
}

func (a *app) invasionSafePoint() bool {
	return a.campaign != nil && a.world != nil && a.armory != nil && a.militaryEnabled() &&
		!a.activeBattle() && a.strategyTurn == nil && a.reorganization == nil && a.transferSession == nil &&
		a.armoryView.PendingUnit == 0 && a.desk.Pending == "" && a.desk.Picker == "" &&
		(a.screen == screenMap || a.screen == screenCommand)
}

func invasionTerritory(table *game.ProvinceTable, limit int, player, japan game.GeneralID) (own, enemy int) {
	for _, p := range table.Province[:limit] {
		if p.Commander == player {
			own++
		}
		if p.Commander == japan {
			enemy++
		}
	}
	return
}

func invasionEstablished(d *campaignDeclaration, army *game.Armory) bool {
	if d.StartedWithPort {
		return true
	}
	if army == nil || army.Landing == nil {
		return false
	}
	for _, attempt := range army.Landing.Attempts {
		if attempt.Outcome == game.LandingLanded {
			return true
		}
	}
	return false
}

func invasionResult(d *campaignDeclaration, army *game.Armory, table *game.ProvinceTable, limit int, player, japan game.GeneralID) string {
	own, enemy := invasionTerritory(table, limit, player, japan)
	if own == 0 {
		return "defeat"
	}
	withdrawn := army != nil && army.Landing != nil && len(army.Landing.Attempts) > 0 &&
		army.Landing.Attempts[len(army.Landing.Attempts)-1].Outcome == game.LandingWithdrawn
	if enemy == 0 && (invasionEstablished(d, army) || withdrawn) {
		return "victory"
	}
	return ""
}

// processCampaignEvents is shared by normal Update and monthly turn boundaries.
// All values are prepared before committing the one-time declaration.
func (a *app) processCampaignEvents() (bool, error) {
	if !a.invasionSafePoint() {
		return false, nil
	}
	if a.declaration == nil {
		status, err := a.militaryRules.AssessUnification(uint8(a.stage), a.playerCommander, a.tbl)
		if err != nil || !status.CanDeclare {
			return false, err
		}
		date, err := game.InvasionDateFor(a.year, a.month)
		if err != nil {
			return false, err
		}
		cast, err := game.SelectInvasionCast(date)
		if err != nil {
			return false, err
		}
		ids := make([]game.GeneralID, len(cast))
		for i, p := range cast {
			ids[i] = p.ID
		}
		port, err := a.tbl.At(game.InvasionEntryProvince)
		if err != nil {
			return false, err
		}
		rules := *a.militaryRules
		rules.DeclaredAt = date.DeclaredAt
		d := &campaignDeclaration{ID: campaignDeclarationID, Date: date, Controller: 1,
			Cast: ids, StartedWithPort: port.Commander == a.invasionFaction()}
		if !validCampaignDeclaration(d, rules, a.year, a.month) {
			return false, fmt.Errorf("invalid prepared declaration")
		}
		a.declaration, a.militaryRules = d, &rules
		a.showCampaignEvent()
		return true, nil
	}
	if !a.invasionControlled() {
		return false, nil
	} // Do not replay old imported/probe saves.
	d := a.declaration
	if !d.AnnouncementRead || d.Outcome != "" && !d.OutcomeRead {
		a.showCampaignEvent()
		return true, nil
	}
	if d.Outcome != "" {
		return false, nil
	}
	if b := a.battle; b != nil && b.finished {
		if err := a.finishComputerBattle(); err != nil {
			return false, err
		}
		if a.battle != nil {
			if err := a.persistFinishedBattle(); err != nil {
				return false, err
			}
			a.battle = nil
		}
	}
	if result := invasionResult(d, a.armory, a.tbl, a.provinceLimit, a.playerCommander, a.invasionFaction()); result != "" {
		d.Outcome = result
		a.showCampaignEvent()
		return true, nil
	}
	plan, _ := game.FixedInvasionPlan(d.Date.Era)
	for number := uint8(len(a.armory.InvasionBatches) + 1); number <= 3; number++ {
		if uint32(a.year)*12+uint32(a.month) < d.Date.DeclaredAt+uint32(plan.Waves[number-1].MonthOffset) {
			break
		}
		port, err := a.tbl.At(game.InvasionEntryProvince)
		if err != nil {
			return false, err
		}
		established := invasionEstablished(d, a.armory)
		if established && (port.Commander != a.invasionFaction() || port.InBattle()) {
			break
		}
		if a.armory.Landing != nil {
			last := a.armory.Landing.Attempts[len(a.armory.Landing.Attempts)-1]
			if last.Outcome == game.LandingWithdrawn {
				break
			}
		}
		destination := game.ExternalInvasionSource
		if port.Commander == a.invasionFaction() {
			destination = game.InvasionEntryProvince
		}
		start, end := 0, 6
		if number == 2 {
			start, end = 6, 8
		}
		if number == 3 {
			start, end = 8, 10
		}
		assembled, err := game.AssembleInvasionWave(a.world, a.armory, a.generals, d.Date, number, d.Cast[start:end], destination)
		if err != nil {
			return false, err
		}
		// Preserve stable world/armory pointers used by command budgets and units.
		oldGenerals, oldUnits, oldStrengths := a.generals, a.world.Units, a.world.Strengths
		oldArmy, oldTable, oldActors := a.armory.Clone(), *a.tbl, a.campaignActors
		oldCurrent, oldScreen := a.current, a.screen
		actors := make(map[game.GeneralID]string, len(oldActors)+end-start)
		for id, key := range oldActors {
			actors[id] = key
		}
		for _, id := range d.Cast[start:end] {
			actors[id] = campaignActorKey(id)
		}
		a.generals, a.world.Units, a.world.Strengths = assembled.Generals, assembled.World.Units, assembled.World.Strengths
		*a.armory, *a.tbl = *assembled.Armory, *assembled.World.Table
		a.campaignActors = actors
		a.world.Opts.Armory = a.armory
		if destination == game.InvasionEntryProvince {
			_, err = a.armory.DeliverInvasionSupply(a.world, number)
		} else {
			used, lookupErr := a.armory.InvasionMonthlyAttacks(a.armory.LastMonth)
			err = lookupErr
			blocked := false
			for id, n := range used {
				if n >= game.JapaneseAttackLimit && a.world.Strengths[id-1].Force > 0 {
					blocked = true
				}
			}
			if err == nil && !blocked {
				err = a.startExternalInvasionBattle()
			}
		}
		if err != nil {
			a.generals, a.world.Units, a.world.Strengths = oldGenerals, oldUnits, oldStrengths
			*a.armory, *a.tbl, a.campaignActors = *oldArmy, oldTable, oldActors
			a.current, a.screen, a.battle = oldCurrent, oldScreen, nil
			return false, err
		}
		a.dirty = true
		if a.activeBattle() {
			return true, nil
		}
		if a.battle != nil && a.battle.finished {
			a.battle = nil
			a.current, a.screen = oldCurrent, screenMap
		}
	}
	// A delayed third merged attempt may wait for next month's shared allowance.
	if a.armory.Landing != nil {
		last := a.armory.Landing.Attempts[len(a.armory.Landing.Attempts)-1]
		if last.Outcome == game.LandingRepulsed && int(last.Through) < len(a.armory.InvasionBatches) {
			used, err := a.armory.InvasionMonthlyAttacks(a.armory.LastMonth)
			if err != nil {
				return false, err
			}
			blocked := false
			for id, n := range used {
				if n >= game.JapaneseAttackLimit && a.world.Strengths[id-1].Force > 0 {
					blocked = true
				}
			}
			if !blocked {
				if err := a.startExternalInvasionBattle(); err != nil {
					return false, err
				}
				return a.activeBattle(), nil
			}
		}
	}
	if port, err := a.tbl.At(game.InvasionEntryProvince); err == nil && port.Commander == a.invasionFaction() && !port.InBattle() {
		for number := uint8(1); int(number) <= len(a.armory.InvasionBatches); number++ {
			if _, err := a.armory.DeliverInvasionSupply(a.world, number); err != nil {
				return false, err
			}
		}
	}
	return false, nil
}

func validCampaignController(v *campaignV4, sc game.Scenario) bool {
	d := v.Declaration
	if d == nil || d.Controller == 0 {
		return true
	}
	if sc.Stage == 1 && d.StartedWithPort || !d.AnnouncementRead && (len(v.Actors) != 0 || d.Outcome != "") {
		return false
	}
	if len(v.Actors) != 0 && len(v.Armory.InvasionBatches) == 0 {
		return false
	}
	if d.StartedWithPort && v.Armory.Landing != nil {
		return false
	}
	if d.Outcome != "" {
		japan, _ := game.JapaneseInvasionFaction(sc.Stage)
		if invasionResult(d, v.Armory, v.Base.State.Provinces, sc.Provinces, v.Base.Identity.Player, japan) != d.Outcome {
			return false
		}
	}
	// Each new unit identity is exactly the corresponding frozen cast entry.
	var expected []game.GeneralID
	for n := uint8(1); int(n) <= len(v.Armory.InvasionBatches); n++ {
		expected = append(expected, v.Armory.InvasionBatches[n].Units...)
	}
	if len(expected) != len(v.Actors) {
		return false
	}
	actual := map[game.GeneralID]bool{}
	for _, person := range v.Actors {
		actual[person.ID] = true
	}
	for _, id := range expected {
		if !actual[id] {
			return false
		}
	}
	for i, id := range expected {
		if id != d.Cast[i] {
			return false
		}
	}
	return true
}
