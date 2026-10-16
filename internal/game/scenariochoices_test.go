package game

import (
	"fmt"
	"reflect"
	"testing"
)

func TestScenarioChoicesFromFreshOriginalData(t *testing.T) {
	want := [][]GeneralID{{1, 3, 4, 58, 98, 141, 146, 156, 157, 166}, {1, 107}, {1, 107}}
	for i, sc := range Scenarios {
		t.Run(fmt.Sprintf("stage-%d", sc.Stage), func(t *testing.T) {
			town, err := ParseTownFile(readGame(t, fmt.Sprintf("TOWN(%d).DAT", sc.Stage)))
			if err != nil {
				t.Fatal(err)
			}
			generals, err := ParseGenerals(readGame(t, fmt.Sprintf("MAN(%d).DAT", sc.Stage)), sc.Generals)
			if err != nil {
				t.Fatal(err)
			}
			beforeTown, beforeGenerals := *town, append([]General(nil), generals...)
			choices, err := ScenarioFactionChoices(sc.Stage, town, generals)
			if err != nil {
				t.Fatal(err)
			}
			var leaders []GeneralID
			for _, choice := range choices {
				leaders = append(leaders, choice.Leader)
				if choice.Home != generals[choice.Leader-1].Province {
					t.Errorf("leader %d: home=%d", choice.Leader, choice.Home)
				}
				for _, id := range choice.GeneralIDs {
					if id < 1 || int(id) > sc.Generals {
						t.Fatalf("invalid general %d", id)
					}
				}
				if sc.Stage > 1 && choice.Leader == 107 && (len(choice.GeneralIDs) != 85 || choice.GeneralIDs[0] != 107 || choice.GeneralIDs[84] != 191) {
					t.Fatalf("enemy records lost: %+v", choice.GeneralIDs)
				}
			}
			if !reflect.DeepEqual(leaders, want[i]) {
				t.Errorf("leaders=%v want=%v", leaders, want[i])
			}
			if !reflect.DeepEqual(*town, beforeTown) || !reflect.DeepEqual(generals, beforeGenerals) {
				t.Fatal("catalog mutated original input")
			}
			again, err := ScenarioFactionChoices(sc.Stage, town, generals)
			if err != nil || !reflect.DeepEqual(again, choices) {
				t.Fatal("unstable catalog")
			}
		})
	}
}

func catalogFixture() (*ProvinceTable, []General) {
	town := &ProvinceTable{}
	town.Province[0].Commander = 1
	town.Province[1].Commander = 1
	gens := make([]General, Stage1GeneralCount)
	gens[0] = General{Province: 2, Force: 65535}
	gens[1] = General{Province: 1, Force: 65535}
	return town, gens
}

func TestScenarioChoicesKeepHomeAndLargeForce(t *testing.T) {
	town, gens := catalogFixture()
	choices, err := ScenarioFactionChoices(1, town, gens)
	if err != nil {
		t.Fatal(err)
	}
	if len(choices) != 1 || choices[0].Home != 2 || choices[0].Force != 131070 {
		t.Fatalf("home must follow leader, force must not truncate: %+v", choices)
	}
	choices[0].Provinces[0] = 36
	choices[0].GeneralIDs[0] = 274
	again, err := ScenarioFactionChoices(1, town, gens)
	if err != nil || again[0].Provinces[0] != 1 || again[0].GeneralIDs[0] != 1 || gens[0].Province != 2 {
		t.Fatal("returned choices alias source or cached output")
	}
}

func TestScenarioChoicesIgnoreInactiveProvinceRecords(t *testing.T) {
	town, gens := catalogFixture()
	town.Province[36].Commander = 65000
	gens[2] = General{Province: 37, Force: 60000}
	choices, err := ScenarioFactionChoices(1, town, gens)
	if err != nil || len(choices) != 1 || choices[0].Force != 131070 || len(choices[0].GeneralIDs) != 2 {
		t.Fatalf("inactive records leaked: choices=%+v err=%v", choices, err)
	}
	if town.Province[36].Commander != 65000 {
		t.Fatal("inactive source record was rewritten")
	}
}

func TestScenarioChoicesRejectInvalidSources(t *testing.T) {
	for _, name := range []string{"stage", "nil", "save", "short", "tail", "leader", "home-zero", "home-outside", "home-foreign", "empty"} {
		t.Run(name, func(t *testing.T) {
			town, gens := catalogFixture()
			stage := uint8(1)
			switch name {
			case "stage":
				stage = 0
			case "nil":
				town = nil
			case "save":
				town.Date = &GameDate{Year: 15, Month: 7}
			case "short":
				gens = gens[:191]
			case "tail":
				stage = 2
			case "leader":
				town.Province[0].Commander = 275
			case "home-zero":
				gens[0].Province = 0
			case "home-outside":
				gens[0].Province = 37
			case "home-foreign":
				town.Province[1].Commander = 2
			case "empty":
				town = &ProvinceTable{}
			}
			if choices, err := ScenarioFactionChoices(stage, town, gens); err == nil || choices != nil {
				t.Fatalf("accepted invalid source: %+v %v", choices, err)
			}
		})
	}
}
