package main

import (
	"fmt"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/i18n"
)

func (a *app) invasionOfficerName(id game.GeneralID) string {
	if a.people != nil {
		if p, ok := a.people.InvasionPerson(game.InvasionActorKey(id)); ok {
			if a.people.Language != "zh-Hant" && p.NameCommon != "" {
				return p.NameCommon
			}
			return p.NameInGame
		}
	}
	return fmt.Sprintf("%s %d", a.openingText("japan"), id)
}

func (a *app) generalPerson(id game.GeneralID) (*i18n.Person, bool) {
	if a.people == nil {
		return nil, false
	}
	if key := a.campaignActors[id]; key != "" {
		return a.people.InvasionPerson(key)
	}
	return a.people.PersonAt(a.stage, int(id))
}
