package main

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/i18n"
)

func loadViewPortraits(t *testing.T, a *app) {
	t.Helper()
	var err error
	a.people, err = i18n.LoadPeople("../../translations/zh-Hant", "../../translations/shared")
	if err != nil {
		t.Fatal(err)
	}
	a.eten = &assets.EtenFonts{}
	a.wording, err = i18n.LoadWording("../../translations/zh-Hant")
	if err != nil {
		t.Fatal(err)
	}
	a.portraits, err = i18n.LoadPortraitCatalog("../../translations/shared", a.people)
	if err != nil {
		t.Fatal(err)
	}
}

func TestBiographyPortraitUsesSelectedPersonAndKeepsFullText(t *testing.T) {
	a := newArmoryApp(t, 1, 1)
	loadViewPortraits(t, a)
	a.viewGenerals = []game.GeneralID{1, 2}
	a.viewIndex, a.screen = 1, screenBiography
	d := a.deskData()
	if d.Portrait == nil || d.Portrait.Portrait == nil || d.Portrait.Portrait.PersonID != 2 {
		t.Fatal("biography substituted the province leader for the selected general")
	}
	p, _ := a.people.PersonAt(1, 2)
	if !strings.Contains(d.Texts[len(d.Texts)-1].Text, p.Biography) || !strings.Contains(d.Texts[len(d.Texts)-1].Text, d.Portrait.Portrait.SourcePage) {
		t.Fatal("portrait column lost biography text or its source")
	}
}

func TestBattlePortraitsFollowActualRostersWithoutChangingSavedState(t *testing.T) {
	a := newArmoryApp(t, 1, 58)
	loadViewPortraits(t, a)
	if err := a.startBattle(18, a.current); err != nil {
		t.Fatal(err)
	}
	before, _ := json.Marshal(a.campaignState())
	d := a.battleInformationData()
	for side, units := range [][]*game.Combatant{a.battle.sim.Attacker, a.battle.sim.Defender} {
		if len(d.Sides[side].Portraits) != len(units) {
			t.Fatal("portrait count does not match the actual roster")
		}
		for i, unit := range units {
			want, _ := a.portraitForGeneral(unit.General)
			if !reflect.DeepEqual(d.Sides[side].Portraits[i].Portrait, want) {
				t.Fatalf("wrong portrait for side %d, general %d", side, unit.General)
			}
		}
	}
	after, _ := json.Marshal(a.campaignState())
	if string(before) != string(after) {
		t.Fatal("photo lookup changed game or RNG state")
	}
	if d.Sides[0].Portraits[0].Portrait.PersonID == d.Sides[1].Portraits[0].Portrait.PersonID {
		t.Fatal("opposing leaders share an incorrect photo")
	}
}

func TestRepeatedSaveShortcutPreservesOriginalReturnPage(t *testing.T) {
	a := newArmoryApp(t, 1, 1)
	a.screen = screenMap
	for i := 0; i < 3; i++ {
		if err := a.deskOpenPage("save"); err != nil {
			t.Fatal(err)
		}
	}
	if a.desk.Return != screenMap {
		t.Fatal("repeated shortcut replaced the map with its own save dialog")
	}
	a.deskRequestBack(false)
	if a.screen != screenMap {
		t.Fatal("cancel no longer returns to the original page")
	}
}
