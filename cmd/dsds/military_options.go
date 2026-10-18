package main

import (
	"fmt"

	"github.com/wicanr2/great-era-remake/internal/game"
)

func (a *app) militaryEnabled() bool {
	return a.militaryRules != nil && a.militaryRules.Enabled
}

func (a *app) militaryText(key string) string {
	s, err := a.wordingText("military." + key)
	if err != nil {
		return "[military." + key + "]"
	}
	return s
}

func (a *app) militaryStatus() string {
	if a.opening != nil {
		if a.opening.Mode == 1 {
			return "on"
		}
		return "off"
	}
	if a.militaryRules != nil && a.militaryRules.Locked() {
		return "locked"
	}
	if a.militaryEnabled() {
		return "on"
	}
	return "off"
}

// setMilitaryEnabled is the only runtime rules switch. Equipment already
// issued remains real equipment; disabling never refunds or deletes it.
func (a *app) setMilitaryEnabled(enabled bool) error {
	if a.opening != nil {
		mode := 0
		if enabled {
			mode = 1
		}
		a.openingActivate(mode)
		return nil
	}
	if a.campaign == nil || a.world == nil || a.rng == nil {
		return fmt.Errorf("%s", a.militaryText("campaign.required"))
	}
	next := game.MilitaryRules{}
	if a.militaryRules != nil {
		next = *a.militaryRules
	}
	if err := next.SetEnabled(enabled); err != nil {
		return fmt.Errorf("%s", a.militaryText("locked.note"))
	}
	if enabled && a.stage > 1 && a.playerCommander == 107 {
		return fmt.Errorf("%s", a.openingText("japan.note"))
	}
	if a.militaryEnabled() == enabled {
		return nil
	}
	if a.strategyLocked() || a.reorganization != nil || a.transferSession != nil || a.armoryView.PendingUnit != 0 || a.screen == screenQuit || a.screen == screenSaveConfirm || a.screen == screenLoadConfirm {
		return fmt.Errorf("%s", a.militaryText("busy"))
	}
	army := a.armory
	if enabled && army == nil {
		var err error
		army, err = game.NewArmory(a.world, a.provinceLimit, uint32(a.year)*12+uint32(a.month))
		if err != nil {
			return err
		}
	}
	a.armory, a.world.Opts.Armory = army, army
	a.militaryRules = &next
	if !enabled && a.armorySurface() {
		a.screen = screenMap
		a.applyWindowSize()
	}
	a.cancelGamePointer()
	a.dirty = true
	return nil
}

func (a *app) japaneseOffensiveAllowed(faction game.GeneralID) bool {
	return a.militaryRules == nil || a.militaryRules.JapaneseOffensiveAllowed(uint8(a.stage), a.playerCommander, faction)
}
