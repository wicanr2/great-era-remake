package main

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	"github.com/wicanr2/great-era-remake/internal/ui/atlasmap"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
)

var keypadDigits = [...]ebiten.Key{ebiten.KeyNumpad0, ebiten.KeyNumpad1, ebiten.KeyNumpad2, ebiten.KeyNumpad3, ebiten.KeyNumpad4, ebiten.KeyNumpad5, ebiten.KeyNumpad6, ebiten.KeyNumpad7, ebiten.KeyNumpad8, ebiten.KeyNumpad9}

func keyboardAliases(key ebiten.Key) []ebiten.Key {
	for n, digit := range digitKeys() {
		if key == digit || key == keypadDigits[n] {
			return []ebiten.Key{digit, keypadDigits[n]}
		}
	}
	return []ebiten.Key{key}
}
func keyboardKeyPressed(key ebiten.Key) bool {
	for _, k := range keyboardAliases(key) {
		if inpututil.IsKeyJustPressed(k) {
			return true
		}
	}
	return false
}

func (a *app) numericInput() *uint32 {
	switch a.screen {
	case screenTransferTarget, screenTransferSelection, screenTransferAmount:
		return &a.transferInput
	case screenTradeAmount:
		return &a.tradeAmount
	case screenSupplyTarget, screenSupplyAmount:
		return &a.supplyInput
	case screenRecruitAmount:
		return &a.recruitAmount
	case screenReorganizeTarget, screenReorganizeAmount:
		return &a.reorganizeInput
	case screenCovertTarget:
		return &a.covertInput
	case screenLoanAmount, screenRepayAmount:
		return &a.diplomacyInput
	case screenCeasefireTarget:
		return &a.ceasefireInput
	case screenViewProvinceSelect:
		return &a.viewInput
	case screenAutonomy:
		return &a.autonomyInput
	case screenMessageTime:
		return &a.messageTimeInput
	case screenProduction:
		if a.productionItem != 0 {
			return &a.productionInput
		}
	case screenBattle:
		if a.battle != nil && a.battle.mode == battleModeRetreat {
			return &a.battle.retreatInput
		}
	}
	return nil
}
func (a *app) classicZeroBack() bool {
	if a.commandInterface() || a.numericInput() != nil {
		return false
	}
	switch a.screen {
	case screenDevelop, screenTransferMode, screenTradeMode, screenTradeGood,
		screenRecruitAction, screenRecruitBranch, screenReorganizeBranch, screenCovertAction,
		screenViewMenu, screenViewProvinceChoice, screenViewProvince, screenViewOverview,
		screenViewProvinceNames, screenViewGenerals, screenViewGeneral, screenBiography,
		screenPolicy, screenProduction, screenDiplomacy, screenDisplayOptions, screenResolutionOptions:
		return true
	}
	return false
}
func (a *app) classicHomeAllowed() bool {
	switch a.screen {
	case screenMap, screenCommand, screenViewMenu, screenViewProvinceSelect, screenViewProvinceChoice,
		screenViewProvince, screenViewOverview, screenViewProvinceNames, screenViewGenerals, screenViewGeneral, screenBiography:
		return true
	}
	return false
}
func (a *app) classicNation() {
	a.provinceMap = false
	a.screen, a.dirty = screenMap, true
}
func (a *app) updateClassicControls() bool {
	if a.commandInterface() {
		return false
	}
	if a.classicHomeAllowed() && keyboardKeyPressed(ebiten.KeyHome) {
		a.classicNation()
		return true
	}
	if input := a.numericInput(); input != nil && keyboardKeyPressed(ebiten.KeyDelete) {
		*input = 0
		a.dirty = true
		return true
	}
	if a.screen != screenMap {
		return false
	}
	switch a.pointerAction {
	case actions.Action("classic.province"):
		a.provinceMap = true
		a.dirty = true
		return true
	case actions.Action("classic.nation"):
		a.classicNation()
		return true
	}
	if keyboardKeyPressed(ebiten.KeyP) {
		a.provinceMap = !a.provinceMap
		a.dirty = true
		return true
	}
	if a.provinceMap && (a.pointerAction == actions.Back || keyboardKeyPressed(ebiten.KeyEscape) || keyboardKeyPressed(ebiten.KeyDigit0)) {
		a.classicNation()
		return true
	}
	if strings.HasPrefix(string(a.pointerAction), "classic.province.") {
		n, _ := strconv.Atoi(strings.TrimPrefix(string(a.pointerAction), "classic.province."))
		if n > 0 && n <= a.provinceLimit {
			a.current = game.ProvinceID(n)
			a.dirty = true
		}
		return true
	}
	return false
}

func (a *app) classicText(key string) string {
	s, e := a.wordingText("classic." + key)
	if e != nil {
		return "[classic." + key + "]"
	}
	return s
}
func (a *app) classicMapData() (render.ClassicMapData, error) {
	if err := a.ensureRoomMap(); err != nil {
		return render.ClassicMapData{}, err
	}
	camera := atlasmap.DefaultCamera()
	camera.Tilt = 1.08
	var factions [39]int
	for i, p := range a.tbl.Province {
		factions[i] = int(p.Commander)
	}
	r := render.ClassicMapRect()
	d := render.ClassicMapData{Fonts: a.eten, Title: a.classicText("map.title") + "  " + a.provinceName(a.current), Hint: a.classicText("map.hint"),
		Commands: a.classicText("commands"), Province: a.classicText("province"), Home: a.classicText("home")}
	if a.narrative != nil && len(a.narrativeImages) > 0 {
		label, _ := a.wordingText("narrative.open")
		d.News = "N " + label
	}
	d.Map = a.room.Map.Render(camera, r.W, r.H, a.themeMode == "retro", true, factions)
	for i, anchor := range atlasmap.Anchors {
		if i >= a.provinceLimit {
			break
		}
		p := a.room.Map.Project(anchor, camera, r.W, r.H)
		if p[0] < 12 || p[0] >= float64(r.W-12) || p[1] < 8 || p[1] >= float64(r.H-8) {
			continue
		}
		d.Markers = append(d.Markers, render.ClassicMapMarker{ID: i + 1, X: r.X + int(p[0]), Y: r.Y + int(p[1]), Selected: game.ProvinceID(i+1) == a.current})
	}
	return d, nil
}
func (a *app) classicRosterData() render.ClassicRosterData {
	d := render.ClassicRosterData{Fonts: a.eten, Title: a.classicText("roster.title"), Hint: a.classicText("roster.hint"), Selected: a.viewIndex}
	text := func(key string) string { s, _ := a.wordingText("view.general." + key); return s }
	for n, id := range a.viewGenerals {
		i := int(id) - 1
		if i < 0 || i >= len(a.generals) {
			continue
		}
		g := a.generals[i]
		attack := 0
		if a.world != nil && i < len(a.world.Strengths) {
			attack = game.Strength(a.world.Strengths[i], a.world.Opts)
		}
		value := func(key string, v any) string { return fmt.Sprintf("%s %v", text(key), v) }
		row := render.ClassicRosterRow{Name: fmt.Sprintf("%d %s", n+1, a.generalDisplayName(id)), General: g, Fan: a.fan}
		row.Cells = [4][3]string{
			{row.Name, a.deskBranchName(g.Branch), value("force", g.Force)},
			{value("lead", g.AbilityA), value("loyalty", g.AbilityB), value("politics", g.AbilityC)},
			{value("experience", g.Experience), value("attack", attack), value("armed", g.F20)},
			{value("skill", g.F19), value("stamina", g.Stamina), value("morale", g.F30)},
		}
		d.Rows = append(d.Rows, row)
	}
	d.UnitTitle = a.classicText("unit.title")
	return d
}
