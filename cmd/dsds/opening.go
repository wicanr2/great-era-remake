package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
)

type openingState struct {
	Stage, Selected, Mode, Focus int
	Choices                      [3][]game.ScenarioFaction
	Press                        int
	Message                      string
	CanContinue                  bool
}

func (a *app) prepareOpening() error {
	o := &openingState{Press: -1, Focus: -1}
	for i, sc := range game.Scenarios {
		t, err := a.dataRead(fmt.Sprintf("TOWN(%d).DAT", sc.Stage))
		if err != nil {
			return err
		}
		m, err := a.dataRead(fmt.Sprintf("MAN(%d).DAT", sc.Stage))
		if err != nil {
			return err
		}
		town, err := game.ParseTownFile(t)
		if err != nil {
			return err
		}
		gens, err := game.ParseGenerals(m, sc.Generals)
		if err != nil {
			return err
		}
		o.Choices[i], err = game.ScenarioFactionChoices(sc.Stage, town, gens)
		if err != nil {
			return err
		}
	}
	if st, err := os.Stat(a.savePath); err == nil {
		o.CanContinue = !st.IsDir()
	}
	a.opening = o
	a.applyWindowSize()
	return nil
}

func (a *app) openingText(key string) string {
	s, err := a.wordingText("opening." + key)
	if err != nil {
		return "[opening." + key + "]"
	}
	return s
}

func (a *app) openingData() render.OpeningData {
	o := a.opening
	d := render.OpeningData{LightUI: a.lightInterface(), Fonts: a.eten, Stage: o.Stage, Selected: o.Selected, Mode: o.Mode, Focus: o.Focus, CanContinue: o.CanContinue, Message: o.Message, Quit: a.screen == screenQuit, Labels: map[string]string{}}
	for _, key := range []string{"eyebrow", "title", "intro", "original", "original.hint", "resistance", "resistance.hint", "factions", "source", "provinces", "generals", "force", "start", "save.hint", "continue", "quit"} {
		d.Labels[key] = a.openingText(key)
	}
	for i, sc := range game.Scenarios {
		d.Stages = append(d.Stages, a.openingText(fmt.Sprintf("stage.%d", i+1)))
		d.Dates = append(d.Dates, fmt.Sprintf(a.openingText("date"), int(sc.Year)+1911, sc.Month))
	}
	choices := o.Choices[o.Stage]
	name := func(id game.GeneralID) string {
		if o.Stage > 0 && id == 107 {
			return a.openingText("japan")
		}
		if a.people != nil {
			if p, ok := a.people.PersonAt(o.Stage+1, int(id)); ok {
				return p.NameInGame
			}
		}
		return fmt.Sprintf(a.openingText("leader"), id)
	}
	for _, f := range choices {
		home := a.provinceName(f.Home)
		disabled := o.Mode == 1 && o.Stage > 0 && f.Leader == 107
		if disabled {
			home = a.openingText("original.only")
		}
		d.Factions = append(d.Factions, render.OpeningFaction{Name: name(f.Leader), Home: home, Summary: fmt.Sprintf(a.openingText("summary"), len(f.GeneralIDs), len(f.Provinces)), Disabled: disabled})
	}
	f := choices[o.Selected]
	sc := game.Scenarios[o.Stage]
	d.Name = name(f.Leader)
	d.Date = fmt.Sprintf("%s · %d", d.Stages[o.Stage], int(sc.Year)+1911)
	d.Home = fmt.Sprintf(a.openingText("home"), a.provinceName(f.Home))
	d.Provinces, d.Generals, d.Force = len(f.Provinces), len(f.GeneralIDs), f.Force
	var provinces []string
	for _, p := range f.Provinces {
		provinces = append(provinces, a.provinceName(p))
	}
	d.Territory = strings.Join(provinces, "、")
	d.Notice = a.openingText("original.note")
	if o.Mode == 1 {
		d.Notice = a.openingText("resistance.pending")
	}
	if o.Mode == 1 && o.Stage > 0 && f.Leader == 107 {
		d.Notice = a.openingText("japan.note")
	}
	return d
}

func (a *app) openingActivate(target int) {
	o := a.opening
	if o == nil {
		return
	}
	n := len(o.Choices[o.Stage])
	switch {
	case target < 0:
		return
	case target < 2:
		o.Mode = target
	case target < 5:
		o.Stage = target - 2
		o.Selected = 0
		o.Focus = -1
	case target < 5+n:
		f := o.Choices[o.Stage][target-5]
		if o.Mode == 1 && o.Stage > 0 && f.Leader == 107 {
			return
		}
		o.Selected = target - 5
	case target == 5+n:
		player := o.Choices[o.Stage][o.Selected].Leader
		if o.Mode == 1 && o.Stage > 0 && player == 107 {
			o.Message = a.openingText("japan.note")
			break
		}
		mode := "original"
		if o.Mode == 1 {
			mode = "unification-resistance"
		}
		if err := a.startCampaignMode(uint8(o.Stage+1), player, mode); err != nil {
			o.Message = err.Error()
		} else {
			a.applyWindowSize()
		}
	case target == 6+n:
		if !o.CanContinue {
			return
		}
		b, err := os.ReadFile(a.savePath)
		if err == nil {
			err = a.loadCampaign(b)
		}
		if err != nil {
			o.Message = err.Error()
		} else {
			a.applyWindowSize()
		}
	}
	a.dirty = true
}

func (a *app) updateOpening() error {
	o := a.opening
	n := len(o.Choices[o.Stage])
	targets := render.OpeningTargets(n)
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		o.Focus = (o.Focus + 1) % len(targets)
		a.dirty = true
	}
	for i, k := range []ebiten.Key{ebiten.KeyDigit1, ebiten.KeyDigit2, ebiten.KeyDigit3} {
		if keyboardKeyPressed(k) {
			a.openingActivate(2 + i)
			return nil
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyM) {
		a.openingActivate(1 - o.Mode)
		return nil
	}
	for _, item := range []struct {
		k     ebiten.Key
		delta int
	}{{ebiten.KeyArrowLeft, -1}, {ebiten.KeyArrowRight, 1}, {ebiten.KeyArrowUp, -2}, {ebiten.KeyArrowDown, 2}} {
		if inpututil.IsKeyJustPressed(item.k) {
			o.Selected = (o.Selected + item.delta + n) % n
			o.Focus = 5 + o.Selected
			a.dirty = true
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyC) {
		a.openingActivate(6 + n)
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		t := o.Focus
		if t < 0 || (t >= 5 && t < 5+n && t == 5+o.Selected) {
			t = 5 + n
		}
		a.openingActivate(t)
		return nil
	}
	x, y, valid := a.gameCursorPosition()
	hit := -1
	if valid {
		for i, r := range targets {
			if x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H {
				hit = i
				break
			}
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		o.Press = hit
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		pressed := o.Press
		o.Press = -1
		if hit >= 0 && hit == pressed {
			o.Focus = hit
			a.openingActivate(hit)
		}
	}
	return nil
}
