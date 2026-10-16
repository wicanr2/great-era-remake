package game

import "fmt"

// NewCampaignState 由原始劇本建立已知世界，不以任何附帶存檔為模板。
// 去重順序及初始化證據見 SPEC-54；不產生偽造的 DOS DT1 bytes。
func NewCampaignState(stage uint8, town *ProvinceTable, generals []General, player GeneralID) (DT1State, ProvinceID, error) {
	choices, err := ScenarioFactionChoices(stage, town, generals)
	if err != nil {
		return DT1State{}, 0, err
	}
	var home ProvinceID
	for _, f := range choices {
		if f.Leader == player {
			home = f.Home
			break
		}
	}
	if home == 0 {
		return DT1State{}, 0, fmt.Errorf("game: 領袖 %d 不是本劇本的可選勢力", player)
	}
	sc, _ := ScenarioByStage(stage)
	t := *town
	d := sc.Date()
	t.Date = &d
	s := DT1State{Provinces: &t, Generals: append([]General(nil), generals...), Ledger: NewDiplomacyLedger()}
	unknown := [FactionSlotCount]uint8{12, 9, 11, 14, 1, 10, 4, 7, 5, 2}
	if stage != 1 {
		unknown[1] = 3
	}
	for i := range s.Factions {
		s.Factions[i].Unknown2 = unknown[i]
	}
	seen := make(map[GeneralID]bool)
	n := 0
	for _, p := range t.Province[:sc.Provinces] {
		id := p.Commander
		if id == 0 || seen[id] {
			continue
		}
		if n == FactionSlotCount {
			return DT1State{}, 0, fmt.Errorf("game: 初始勢力超過 %d", FactionSlotCount)
		}
		seen[id] = true
		s.FactionLeaders[n] = id
		s.FactionOfGeneral[id-1] = uint8(n + 1)
		s.Factions[n].Leader = id
		for j := range s.Factions[n].Relations {
			if j != n {
				s.Factions[n].Relations[j] = FactionRelationDefault
			}
		}
		n++
	}
	s.MajorPowerLeaders[0] = player
	return s, home, nil
}
