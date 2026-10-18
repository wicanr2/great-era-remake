package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	"github.com/wicanr2/great-era-remake/internal/ui/hdart"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
)

func (a *app) campaignEventSurface() bool {
	return a.screen == screenCampaignEvent || a.screen == screenQuit && a.quitBack == screenCampaignEvent
}

func (a *app) acknowledgeCampaignEvent() {
	if !a.invasionControlled() {
		return
	}
	if a.declaration.Outcome != "" {
		a.declaration.OutcomeRead = true
	} else {
		a.declaration.AnnouncementRead = true
	}
	a.screen, a.dirty = screenMap, true
	a.cancelGamePointer()
	a.applyWindowSize()
}

func (a *app) updateCampaignEvent() error {
	if a.actionPressed(actions.Confirm, ebiten.KeyEnter, ebiten.KeySpace) {
		a.acknowledgeCampaignEvent()
	} else if a.actionPressed(actions.Back, ebiten.KeyEscape) {
		if !a.invasionEnded() {
			a.acknowledgeCampaignEvent()
			return nil
		}
		if err := a.autosave(); err != nil {
			a.report(err.Error())
			return nil
		}
		if err := a.prepareOpening(); err != nil {
			a.report(err.Error())
			return nil
		}
		a.dirty = true
	}
	return nil
}

func (a *app) composeCampaignEvent() error {
	if !a.invasionControlled() {
		return fmt.Errorf("missing campaign event")
	}
	d := a.declaration
	data := render.CampaignEventData{Fonts: a.eten, Style: a.uiStyle(), Light: a.lightInterface(),
		Title: a.invasionText("title"), Date: fmt.Sprintf(a.invasionText("date"), d.Date.GregorianYear, d.Date.Month),
		Intro: a.invasionText("intro"), RosterTitle: a.invasionText("roster"), Continue: a.invasionText("continue"),
		Menu: a.invasionText("menu"), ExitNote: a.invasionText("exit.note"), Scene: hdart.DeclarationScene(),
		Ending: d.Outcome != "", Quit: a.screen == screenQuit, QuitTitle: a.invasionText("quit.title"),
		QuitYes: a.invasionText("quit.yes"), QuitNo: a.invasionText("quit.no")}
	if d.Date.AlternateHistory {
		data.Date += " · " + a.invasionText("alternate")
	}
	for i, id := range d.Cast {
		data.Names = append(data.Names, fmt.Sprintf("%d  %s", i+1, a.invasionOfficerName(id)))
	}
	data.Status = fmt.Sprintf(a.invasionText("batch.status"), len(a.armory.InvasionBatches), len(a.campaignActors))
	if data.Ending {
		data.Title = a.invasionText(d.Outcome + ".title")
		data.Intro = a.invasionText(d.Outcome + ".body")
		data.Continue = a.invasionText("view")
		data.Scene = nil
	}
	w, h := a.gameSurfaceSize()
	a.frame = ebiten.NewImageFromImage(render.DrawCampaignEvent(w, h, data).Image())
	a.dirty = false
	return nil
}
