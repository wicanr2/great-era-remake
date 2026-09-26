package game

import "fmt"

// DT1State 是一份 .DT1 的「已解欄位快照」。
//
// ParseDT1 會把目前已能以證據讀出的區塊全部裝入；WriteDT1 則以呼叫端
// 提供的 orig bytes 為基底，只寫這些已解欄位，未解欄位（包含區塊 7 的
// 非領袖殘留、戰爭記錄的未知 bytes，以及檔尾 7 bytes）一律保留；區塊 7
// 只寫目前勢力領袖的 confirmed 反查格。這個
// API 的目的，是讓遊戲 autosave 不必自行拼接多個 writer，也不代表 DT1
// 的所有欄位都已經解完。
type DT1State struct {
	Provinces         *ProvinceTable
	Generals          []General
	Factions          FactionTable
	FactionLeaders    FactionLeaders
	FactionOfGeneral  FactionOfGeneral
	WarRecords        [ProvinceCount + 1]WarRecord
	CeasefireStates   [ProvinceCount + 1]uint8
	MajorPowerLeaders MajorPowerLeaders
	Ledger            DiplomacyLedger
}

// ParseDT1 以該期實際將領數解析 .DT1 的完整已知快照。
//
// generalCount 必須是劇本明示的將領筆數，不可從姓名表猜測；這保留
// SAVE(1).DT1 後 85 筆無姓名部隊仍屬有效記錄的既有結論。
func ParseDT1(data []byte, generalCount int) (DT1State, error) {
	var state DT1State
	if len(data) < SaveFileSize {
		return state, fmt.Errorf("game: .DT1 只有 %d bytes，至少需要 %d", len(data), SaveFileSize)
	}
	provinces, err := ParseSaveProvinces(data)
	if err != nil {
		return state, fmt.Errorf("game: DT1 省份區：%w", err)
	}
	state.Provinces = provinces
	if state.Generals, err = ParseSaveGenerals(data, generalCount); err != nil {
		return DT1State{}, fmt.Errorf("game: DT1 將領區：%w", err)
	}
	if state.Factions, err = ParseFactionTable(data); err != nil {
		return DT1State{}, fmt.Errorf("game: DT1 勢力表：%w", err)
	}
	if state.FactionLeaders, err = ParseFactionLeaders(data); err != nil {
		return DT1State{}, fmt.Errorf("game: DT1 勢力領袖表：%w", err)
	}
	if state.FactionOfGeneral, err = ParseFactionOfGeneral(data); err != nil {
		return DT1State{}, fmt.Errorf("game: DT1 勢力反查表：%w", err)
	}
	if state.WarRecords, err = ParseWarRecords(data); err != nil {
		return DT1State{}, fmt.Errorf("game: DT1 戰爭記錄：%w", err)
	}
	if state.CeasefireStates, err = ParseCeasefireStates(data); err != nil {
		return DT1State{}, fmt.Errorf("game: DT1 停火表：%w", err)
	}
	if state.MajorPowerLeaders, err = ParseMajorPowerLeaders(data); err != nil {
		return DT1State{}, fmt.Errorf("game: DT1 十大勢力清單：%w", err)
	}
	if state.Ledger, err = ParseDiplomacyLedger(data); err != nil {
		return DT1State{}, fmt.Errorf("game: DT1 外交帳本：%w", err)
	}
	return state, nil
}

// WriteDT1 從 orig bytes 寫回 DT1State 的所有已知欄位。
//
// 呼叫順序固定且每個 writer 都以目前 out 的副本為基底；任一步失敗都
// 回傳錯誤，不產生半成品。傳入 nil Provinces 會 fail-closed，避免呼叫端
// 不小心以空的狀態覆蓋省份表。
func WriteDT1(orig []byte, state DT1State) ([]byte, error) {
	if state.Provinces == nil {
		return nil, fmt.Errorf("game: DT1State 缺少省份快照")
	}
	out, err := WriteSave(orig, state.Provinces, state.Generals)
	if err != nil {
		return nil, fmt.Errorf("game: DT1 省份／將領寫回：%w", err)
	}
	if out, err = WriteFactionTable(out, state.Factions); err != nil {
		return nil, fmt.Errorf("game: DT1 勢力表寫回：%w", err)
	}
	if out, err = WriteFactionLeaders(out, state.FactionLeaders); err != nil {
		return nil, fmt.Errorf("game: DT1 勢力領袖表寫回：%w", err)
	}
	if out, err = WriteFactionOfGeneralLeaders(out, state.FactionLeaders); err != nil {
		return nil, fmt.Errorf("game: 勢力領袖反查格寫回：%w", err)
	}
	if out, err = WriteWarRecords(out, state.WarRecords); err != nil {
		return nil, fmt.Errorf("game: DT1 戰爭記錄寫回：%w", err)
	}
	if out, err = WriteCeasefireStates(out, state.CeasefireStates); err != nil {
		return nil, fmt.Errorf("game: DT1 停火表寫回：%w", err)
	}
	if out, err = WriteMajorPowerLeaders(out, state.MajorPowerLeaders); err != nil {
		return nil, fmt.Errorf("game: DT1 十大勢力清單寫回：%w", err)
	}
	if out, err = WriteDiplomacyLedger(out, state.Ledger); err != nil {
		return nil, fmt.Errorf("game: DT1 外交帳本寫回：%w", err)
	}
	return out, nil
}
