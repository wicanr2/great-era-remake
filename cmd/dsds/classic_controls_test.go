package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/i18n"
)

func TestKeyboardAliasesAcceptBothNumericKeyboards(t *testing.T) {
	for n, digit := range digitKeys() {
		want := []ebiten.Key{digit, keypadDigits[n]}
		if !reflect.DeepEqual(keyboardAliases(digit), want) || !reflect.DeepEqual(keyboardAliases(keypadDigits[n]), want) {
			t.Fatal(n, keyboardAliases(digit))
		}
	}
	if got := keyboardAliases(ebiten.KeyEscape); !reflect.DeepEqual(got, []ebiten.Key{ebiten.KeyEscape}) {
		t.Fatal(got)
	}
}
func TestZeroBackDoesNotConsumeNumericZeroOrTenthCommand(t *testing.T) {
	a := &app{}
	for _, s := range []screen{screenCommand, screenOtherOptions, screenRecruitAmount, screenTransferTarget, screenViewProvinceSelect, screenReorganizeAmount, screenTrainConfirm, screenQuit} {
		a.screen = s
		if a.classicZeroBack() {
			t.Fatal("zero must retain its numeric/tenth-command role", s)
		}
	}
	for _, s := range []screen{screenDevelop, screenViewGenerals, screenViewProvince, screenViewMenu} {
		a.screen = s
		if !a.classicZeroBack() {
			t.Fatal("missing menu return", s)
		}
	}
}
func TestClassicHomeIsReadonlyAndCannotDiscardTransactions(t *testing.T) {
	a := newArmoryApp(t, 1, 1)
	a.screen = screenViewProvince
	a.provinceMap = true
	before, _ := json.Marshal(a.campaignState())
	army, _ := json.Marshal(a.armory)
	current := a.current
	a.classicNation()
	after, _ := json.Marshal(a.campaignState())
	later, _ := json.Marshal(a.armory)
	if a.screen != screenMap || a.provinceMap || a.current != current || string(before) != string(after) || string(army) != string(later) {
		t.Fatal("return changed world, RNG, budget or equipment")
	}
	for _, s := range []screen{screenReorganizeAmount, screenTransferConfirm, screenTransferSelection, screenRecruitConfirm, screenBattle, screenSaveConfirm} {
		a.screen = s
		if a.classicHomeAllowed() {
			t.Fatal("Home bypasses transaction or battle", s)
		}
	}
	a.screen = screenRecruitAmount
	a.recruitAmount = 230
	if a.numericInput() != &a.recruitAmount {
		t.Fatal("wrong editable draft")
	}
	*a.numericInput() = 0
	after, _ = json.Marshal(a.campaignState())
	if string(before) != string(after) {
		t.Fatal("draft edit changed campaign")
	}
}
func TestClassicRosterContainsFullActualBasicRecords(t *testing.T) {
	a := newArmoryApp(t, 1, 1)
	var err error
	a.wording, err = i18n.LoadWording("../../translations/zh-Hant")
	if err != nil {
		t.Fatal(err)
	}
	a.wordingMode = i18n.WordingOriginal
	a.viewGenerals = []game.GeneralID{1, 2}
	before, _ := json.Marshal(a.campaignState())
	d := a.classicRosterData()
	if len(d.Rows) != 2 {
		t.Fatal("lost roster", len(d.Rows))
	}
	for i, row := range d.Rows {
		g := a.generals[i]
		if row.General != g || !strings.Contains(row.Cells[0][2], fmt.Sprint(g.Force)) || !strings.Contains(row.Cells[1][2], fmt.Sprint(g.AbilityC)) || !strings.Contains(row.Cells[3][1], fmt.Sprint(g.Stamina)) || !strings.Contains(row.Cells[3][2], fmt.Sprint(g.F30)) {
			t.Fatal("missing actual basic fields", row)
		}
		for _, cells := range row.Cells {
			for _, cell := range cells {
				if cell == "" {
					t.Fatal("empty basic information")
				}
			}
		}
	}
	after, _ := json.Marshal(a.campaignState())
	if string(before) != string(after) {
		t.Fatal("view consumed RNG or changed world")
	}
}
