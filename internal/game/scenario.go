package game

import "fmt"

// 三個歷史背景（劇本）的起始參數（`sub_38C01`，`docs/re/31` §50）。
//
// 原版的選單讀 `'1'`／`'2'`／`'3'`，一次設四個全域：
//
//	byte_6FE7D  年（民國紀年）
//	byte_6FE7E  月
//	word_7001E  可用省份數
//	byte_6FE88  時期編號，也就是 MAN(N).DAT／TOWN(N).DAT 的那個 N

// Scenario 是一個歷史背景的起始參數。
type Scenario struct {
	// Stage 是 `byte_6FE88`，同時也是資料檔後綴：`MAN(1).DAT` 對 Stage 1。
	Stage uint8
	// Name 是這一期的名字。
	Name string
	// Year / Month 是開局日期（民國紀年）。
	Year, Month uint8
	// Provinces 是這一期**可用**的省份數。
	//
	// ⚠️ 它不是常數——北伐時期只有 36，抗戰兩期是 39
	// （`ProvinceCount` 那個 39 是檔案裡的記錄數，不是可用數）。
	Provinces int
	// Generals 是這一期的**將領數**（原版 `word_6BC4A`，`sub_391E1` 依
	// `byte_6FE88` 設定）。
	//
	// ⚠️ **不是 274**：只有第一期是 274（`0x112`），第二、三期是 **191**（`0xBF`）。
	// `MAN(2).DAT`／`MAN(3).DAT` 檔案裡仍然有 274 筆的空間，但**第 192 筆
	// 之後是垃圾**——照 274 筆解會得到兵種 186、所屬省 232 這種值。
	Generals int
	// Historical 是對應的史實事件，寫給文件與 UI 用。
	Historical string
}

// Stage1GeneralCount 是第一期的將領數，給不想繞 `Scenario` 的呼叫端用。
const Stage1GeneralCount = 274

// Scenarios 是三個劇本，索引 0..2 對應原版選單的 '1'..'3'。
var Scenarios = [3]Scenario{
	{Stage: 1, Name: "北伐時期", Year: 15, Month: 7, Provinces: 36, Generals: 274,
		Historical: "1926 年 7 月，國民革命軍誓師北伐"},
	{Stage: 2, Name: "抗戰前期", Year: 26, Month: 7, Provinces: 39, Generals: 191,
		Historical: "1937 年 7 月，七七事變"},
	{Stage: 3, Name: "抗日後期", Year: 30, Month: 12, Provinces: 39, Generals: 191,
		Historical: "1941 年 12 月，太平洋戰爭爆發"},
}

// ScenarioByStage 取某一期的參數。
func ScenarioByStage(stage uint8) (Scenario, error) {
	for _, s := range Scenarios {
		if s.Stage == stage {
			return s, nil
		}
	}
	return Scenario{}, fmt.Errorf("game: 沒有第 %d 期，只有 1..3", stage)
}

// DataSuffix 是這一期資料檔的後綴：`MAN(1).DAT` 的 `1`。
func (s Scenario) DataSuffix() int { return int(s.Stage) }

// RandomProvinceRange 是隨機事件挑省份的範圍上限。
//
// 原版 `sub_19D6A` 用 `Random(word_7001E) + 1`，所以值域是 `1..Provinces`。
// ⭐ **北伐時期永遠挑不到最後三個省**——那三個是哪三個要對 `TOWN(1).DAT`
// 才知道，還沒查。
func (s Scenario) RandomProvinceRange() int { return s.Provinces }

// Date 是開局的遊戲內日期。
func (s Scenario) Date() GameDate {
	return GameDate{Year: uint16(s.Year), Month: s.Month}
}
