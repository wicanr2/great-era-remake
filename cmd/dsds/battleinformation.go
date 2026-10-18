package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
	"strings"
)

func (a *app) battleInformationData() render.BattleInformationData {
	word := func(key string) string {
		text, err := a.wordingText("battleinfo." + key)
		if err != nil {
			return "[battleinfo." + key + "]"
		}
		return text
	}
	d := render.BattleInformationData{Fonts: a.eten, Labels: map[string]string{}, Page: a.battle.infoPage}
	for _, key := range []string{"title", "previous", "next", "back", "pages", "empty"} {
		d.Labels[key] = word(key)
	}
	data := game.InspectBattle(a.battle.sim)
	panel := a.battlePanelData()
	for side, raw := range data.Sides {
		out := &d.Sides[side]
		summary := panel.Attacker
		role := "attacker"
		if side == 1 {
			summary = panel.Defender
			role = "defender"
		}
		out.Name = word(role) + " · " + a.generalDisplayName(summary.Leader)
		out.Summary = append(out.Summary, fmt.Sprintf(word("total"), raw.Units, raw.Force))
		for _, branch := range []uint8{game.BranchInfantry, game.BranchCavalry, game.BranchArmour, game.BranchArtiller} {
			b := raw.Branches[branch]
			out.Summary = append(out.Summary, fmt.Sprintf(word("branch.total"), word(fmt.Sprintf("branch.%d", branch)), b.Units, b.Force))
		}
		out.Summary = append(out.Summary, fmt.Sprintf(word("supply.0"), summary.Gold, summary.Food), fmt.Sprintf(word("supply.1"), summary.Ammo, summary.Fuel))
		for _, unit := range raw.Roster {
			s := unit.Strength
			branchName := word("branch.unknown")
			if _, ok := game.EquipmentClassOf(s.Branch); ok {
				branchName = word(fmt.Sprintf("branch.%d", s.Branch))
			}
			position := fmt.Sprint(unit.Cell)
			if !unit.Cell.Valid() {
				position = word("offmap")
			}
			lines := []string{
				a.generalDisplayName(unit.General) + " · " + branchName,
				fmt.Sprintf(word("unit.force"), s.Force, s.Ability, s.F19),
				fmt.Sprintf(word("unit.condition"), s.F20, s.F29, s.F30),
				fmt.Sprintf(word("unit.field"), unit.Experience, unit.Mobility, unit.MaxMobility, position),
			}
			if unit.Tracked {
				lines = append(lines, fmt.Sprintf(word("unit.equipped"), unit.Equipped, unit.Need, max(int64(0), int64(unit.Need)-int64(unit.Equipped)), unit.Power))
				lines = append(lines, fmt.Sprintf(word("unit.protection"), float64(unit.Protection.Armor)/float64(game.EquipmentCoefficientScale), float64(unit.Protection.Piercing)/float64(game.EquipmentCoefficientScale)))
				items := []string{}
				for _, item := range unit.Equipment {
					items = append(items, fmt.Sprintf("%s %d", a.armoryText("model."+string(item.ID)), item.Count))
				}
				// A maximum of four models in a class fits two bounded rows.
				middle := (len(items) + 1) / 2
				lines = append(lines, strings.Join(items[:middle], " / "), strings.Join(items[middle:], " / "))
			} else {
				lines = append(lines, word("unit.untracked"), word("unit.unknown"))
			}
			out.Units = append(out.Units, lines)
			portrait, unavailable := a.portraitForGeneral(unit.General)
			out.Portraits = append(out.Portraits, render.PortraitView{Portrait: portrait, Unavailable: unavailable})
		}
	}
	pages := render.BattleInformationPages(d, a.highModernBattle())
	a.battle.infoPage = max(0, min(d.Page, pages-1))
	d.Page = a.battle.infoPage
	return d
}

func (a *app) updateBattleInformation() {
	b := a.battle
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || a.gameMouseEnabled() && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		b.mode = battleModeCommand
		a.dirty = true
		return
	}
	d := a.battleInformationData()
	w, h := 640, 350
	if a.highModernBattle() {
		w, h = 1280, 720
	}
	targets := render.BattleInformationTargets(d, w, h)
	perform := func(key string) {
		switch key {
		case "previous":
			b.infoPage = max(0, b.infoPage-1)
		case "next":
			b.infoPage = min(render.BattleInformationPages(d, w >= 1000)-1, b.infoPage+1)
		case "back":
			b.mode = battleModeCommand
		}
		a.dirty = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
		perform("previous")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
		perform("next")
	}
	x, y := ebiten.CursorPosition()
	x, y, valid := a.basePointerPosition(x, y)
	valid = valid && a.gameMouseEnabled()
	hit := ""
	if valid {
		for _, target := range targets {
			r := target.Rect
			if !target.Disabled && x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H {
				hit = target.Key
				break
			}
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		b.infoPress = hit
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		if hit != "" && hit == b.infoPress {
			perform(hit)
		}
		b.infoPress = ""
	}
}
