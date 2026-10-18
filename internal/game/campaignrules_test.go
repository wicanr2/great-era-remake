package game

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestMilitaryDeclarationLocksSwitchAndRejectsInvalidSaveState(t *testing.T) {
	const month = 26*12 + 7
	r := MilitaryRules{Enabled: true, DeclaredAt: month}
	if !r.Valid(month) || !r.Locked() {
		t.Fatal("declared rules rejected")
	}
	before := r
	if err := r.SetEnabled(false); err != ErrMilitaryLocked || r != before {
		t.Fatal("declaration lock changed state", r, err)
	}
	for _, invalid := range []MilitaryRules{{Enabled: false, DeclaredAt: month}, {Enabled: true, DeclaredAt: month + 1}} {
		if invalid.Valid(month) {
			t.Fatal("invalid declaration state accepted", invalid)
		}
	}
	if r.Valid(0) || r.Valid(65536*12+1) {
		t.Fatal("invalid accounting date accepted")
	}
}

func TestJapaneseOffensiveGateIsSpecificToEnabledChineseCampaign(t *testing.T) {
	r := MilitaryRules{Enabled: true}
	if r.JapaneseOffensiveAllowed(2, 1, 107) || r.JapaneseOffensiveAllowed(3, 1, 107) {
		t.Fatal("Japan can attack before declaration")
	}
	for _, tc := range []struct {
		stage           uint8
		player, faction GeneralID
	}{{1, 1, 107}, {2, 107, 1}, {2, 1, 2}} {
		if !r.JapaneseOffensiveAllowed(tc.stage, tc.player, tc.faction) {
			t.Fatal("unrelated faction or human Japan blocked", tc)
		}
	}
	r.Enabled = false
	if !r.JapaneseOffensiveAllowed(2, 1, 107) {
		t.Fatal("disabled original campaign changed")
	}
	r.Enabled, r.DeclaredAt = true, 26*12+7
	if !r.JapaneseOffensiveAllowed(2, 1, 107) {
		t.Fatal("declared Japan cannot attack")
	}
}

func TestUnificationAllFourteenFreshScenarioChoices(t *testing.T) {
	rules := MilitaryRules{Enabled: true}
	count := 0
	for _, sc := range Scenarios {
		read := func(name string) []byte {
			t.Helper()
			data, err := os.ReadFile(filepath.Join("..", "..", "workplace", "orig", "game", name))
			if os.IsNotExist(err) {
				t.Skip("private original data not installed")
			}
			if err != nil {
				t.Fatal(err)
			}
			return data
		}
		town, err := ParseTownFile(read(fmt.Sprintf("TOWN(%d).DAT", sc.Stage)))
		if err != nil {
			t.Fatal(err)
		}
		generals, err := ParseGenerals(read(fmt.Sprintf("MAN(%d).DAT", sc.Stage)), sc.Generals)
		if err != nil {
			t.Fatal(err)
		}
		choices, err := ScenarioFactionChoices(sc.Stage, town, generals)
		if err != nil {
			t.Fatal(err)
		}
		for _, choice := range choices {
			count++
			t.Run(fmt.Sprintf("stage%d/player%d", sc.Stage, choice.Leader), func(t *testing.T) {
				state, _, err := NewCampaignState(sc.Stage, town, generals, choice.Leader)
				if err != nil {
					t.Fatal(err)
				}
				before, err := json.Marshal(state)
				if err != nil {
					t.Fatal(err)
				}
				s, err := rules.AssessUnification(sc.Stage, choice.Leader, state.Provinces)
				want := sc.Stage > 1 && choice.Leader == 1
				if err != nil || s.Unified != want || s.CanDeclare != want {
					t.Fatal(s, err, "want", want)
				}
				if s.PlayerProvinces+s.JapaneseProvinces+s.UnownedProvinces+s.OtherChineseProvinces != sc.Provinces {
					t.Fatal("lost effective province", s)
				}
				if sc.Stage == 2 && (s.UnownedProvinces != 6 || s.JapaneseProvinces != 15) {
					t.Fatal("fresh 1937 facts changed", s)
				}
				after, err := json.Marshal(state)
				if err != nil || string(before) != string(after) {
					t.Fatal("assessment changed world, generals, resources or diplomatic state", err)
				}
			})
		}
	}
	if count != 14 {
		t.Fatal("scenario choices changed", count)
	}
}

func TestUnificationBoundaryAndAlreadyDeclaredState(t *testing.T) {
	r := MilitaryRules{Enabled: true}
	var table ProvinceTable
	// All unowned and a player with no land must not satisfy an empty condition.
	s, err := r.AssessUnification(2, 1, &table)
	if err != nil || s.Unified || s.CanDeclare || s.UnownedProvinces != 39 {
		t.Fatal(s, err)
	}
	table.Province[0].Commander = 1
	table.Province[1].Commander = 107
	s, err = r.AssessUnification(2, 1, &table)
	if err != nil || !s.Unified || !s.CanDeclare || s.UnownedProvinces != 37 {
		t.Fatal(s, err)
	}
	// The same ID is Chinese in stage 1 and cannot be ignored as Japan.
	s, err = r.AssessUnification(1, 1, &table)
	if err != nil || s.Unified || s.JapaneseProvinces != 0 || s.OtherChineseProvinces != 1 {
		t.Fatal(s, err)
	}
	table.Province[0].Commander = 107
	s, err = r.AssessUnification(1, 107, &table)
	if err != nil || !s.Unified || s.JapanesePlayer {
		t.Fatal(s, err)
	}
	s, err = r.AssessUnification(2, 107, &table)
	if err != nil || s.Unified || s.CanDeclare || !s.JapanesePlayer {
		t.Fatal(s, err)
	}
	table.Province[0].Commander, table.Province[1].Commander = 1, 107
	r.Enabled = false
	s, err = r.AssessUnification(2, 1, &table)
	if err != nil || !s.Unified || s.CanDeclare {
		t.Fatal("disabled rule opened declaration", s, err)
	}
	r.Enabled, r.DeclaredAt = true, 26*12+7
	locked := r
	s, err = r.AssessUnification(2, 1, &table)
	if err != nil || s.CanDeclare || r != locked {
		t.Fatal("declaration can repeat", s, err)
	}
	table.Province[0].Commander = 107
	s, err = r.AssessUnification(2, 1, &table)
	if err != nil || s.Unified || s.CanDeclare || r != locked || !r.Locked() {
		t.Fatal("losing territory undid the declaration", s, err)
	}
}

func TestUnificationRejectsInvalidInputsAndHonorsEffectiveProvinceLimit(t *testing.T) {
	r := MilitaryRules{Enabled: true}
	var table ProvinceTable
	table.Province[0].Commander = 1
	table.Province[38].Commander = 65535 // outside the first-stage effective range
	s, err := r.AssessUnification(1, 1, &table)
	if err != nil || !s.CanDeclare || s.UnownedProvinces != 35 {
		t.Fatal(s, err)
	}
	for _, tc := range []struct {
		stage  uint8
		player GeneralID
		table  *ProvinceTable
	}{{0, 1, &table}, {4, 1, &table}, {1, 0, &table}, {2, 192, &table}, {1, 1, nil}, {2, 1, &table}} {
		if _, err := r.AssessUnification(tc.stage, tc.player, tc.table); err == nil {
			t.Fatal("invalid unification state accepted", tc)
		}
	}
}
