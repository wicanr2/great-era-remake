package main

import (
	"bytes"
	"os"
	"reflect"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/game"
)

func TestCampaignRosterMonthReferenceSurvivesV4(t *testing.T) {
	// Real scenario arrays, explicit imported declaration/date fixtures. This
	// checks storage semantics, not normally triggered war or historical casting.
	for _, tt := range []struct {
		name   string
		stage  uint8
		year   uint16
		month  uint8
		exit   game.RosterExitDate
		before bool
	}{
		{"muto-already-dead", 2, 26, 7, game.RosterExitDate{Year: 1933, Month: 7, Day: 28}, false},
		{"ishiwara-already-reserve", 3, 30, 12, game.RosterExitDate{Year: 1941, Month: 3}, false},
		{"anami-august-month-start", 3, 34, 8, game.RosterExitDate{Year: 1945, Month: 8, Day: 15}, true},
		{"anami-frozen-after-war", 3, 35, 1, game.RosterExitDate{Year: 1945, Month: 8, Day: 15}, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			a := newArmoryApp(t, tt.stage, 1)
			a.year, a.month = tt.year, tt.month
			month := uint32(a.year)*12 + uint32(a.month)
			a.armory.LastMonth = month
			a.militaryRules.DeclaredAt = month
			var err error
			a.declaration, err = importedCampaignDeclaration(*a.militaryRules)
			if err != nil {
				t.Fatal(err)
			}
			date := a.declaration.Date
			before, err := date.BeforeRosterExit(tt.exit)
			if err != nil || before != tt.before {
				t.Fatalf("initial exclusion: %v, %v", before, err)
			}
			original, armory := a.campaignState(), a.armory.Clone()
			if err := a.autosave(); err != nil {
				t.Fatal(err)
			}
			raw, err := os.ReadFile(a.savePath)
			if err != nil {
				t.Fatal(err)
			}
			b := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
			if err := b.loadCampaign(raw); err != nil {
				t.Fatal(err)
			}
			if b.declaration.Date != date || !b.declaration.Imported || len(b.campaignActors) != 0 ||
				!reflect.DeepEqual(original, b.campaignState()) || !reflect.DeepEqual(armory, b.armory) {
				t.Fatal("restart reselected the reference, spawned actors or altered the original world")
			}
			before, err = b.declaration.Date.BeforeRosterExit(tt.exit)
			if err != nil || before != tt.before {
				t.Fatalf("restored exclusion: %v, %v", before, err)
			}
			if err := b.autosave(); err != nil {
				t.Fatal(err)
			}
			again, err := os.ReadFile(a.savePath)
			if err != nil || !bytes.Equal(raw, again) {
				t.Fatal("second save changed the reference or world")
			}
		})
	}
}
