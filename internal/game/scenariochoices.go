package game

import (
	"fmt"
	"sort"
)

// ScenarioFaction 是原始劇本中一個有領地勢力的唯讀開局目錄。
// 姓名、外交、保存及玩家控制權不由本型別推測，契約見 SPEC-53。
type ScenarioFaction struct {
	Leader     GeneralID
	Home       ProvinceID
	Provinces  []ProvinceID
	GeneralIDs []GeneralID
	Force      uint32
}

// ScenarioFactionChoices 只接受 TOWN 與該期有效 MAN 記錄。
// 不借用附帶存檔的勢力表，也不把姓名表長度當有效將領數。
func ScenarioFactionChoices(stage uint8, town *ProvinceTable, generals []General) ([]ScenarioFaction, error) {
	sc, err := ScenarioByStage(stage)
	if err != nil {
		return nil, err
	}
	if town == nil || town.Date != nil {
		return nil, fmt.Errorf("game: 開局目錄需要原始 TOWN 省表，不能使用存檔快照")
	}
	if len(generals) != sc.Generals {
		return nil, fmt.Errorf("game: 第 %d 期目錄需要 %d 筆有效將領，實得 %d", stage, sc.Generals, len(generals))
	}
	active := *town
	for i := range active.Province {
		if i >= sc.Provinces {
			active.Province[i].Commander = 0
		} else if int(active.Province[i].Commander) > sc.Generals {
			return nil, fmt.Errorf("game: 第 %d 省領袖 %d 超出第 %d 期將領範圍", i+1, active.Province[i].Commander, stage)
		}
	}
	groups := active.Factions()
	if len(groups) == 0 {
		return nil, fmt.Errorf("game: 第 %d 期沒有有領地的勢力", stage)
	}
	ids := make([]int, 0, len(groups))
	for id := range groups {
		ids = append(ids, int(id))
	}
	sort.Ints(ids)
	out := make([]ScenarioFaction, len(ids))
	index := make(map[GeneralID]int, len(ids))
	for i, id := range ids {
		leader := GeneralID(id)
		home := generals[id-1].Province
		if !home.Valid() || int(home) > sc.Provinces || active.Province[home-1].Commander != leader {
			return nil, fmt.Errorf("game: 第 %d 期領袖 %d 的初始省份 %d 不在其轄區", stage, leader, home)
		}
		out[i] = ScenarioFaction{Leader: leader, Home: home, Provinces: groups[leader]}
		index[leader] = i
	}
	for i, general := range generals {
		p := general.Province
		if !p.Valid() || int(p) > sc.Provinces {
			continue
		}
		owner := active.Province[p-1].Commander
		if owner == 0 {
			continue
		}
		f := &out[index[owner]]
		f.GeneralIDs = append(f.GeneralIDs, GeneralID(i+1))
		f.Force += uint32(general.Force)
	}
	return out, nil
}
