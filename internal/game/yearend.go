package game

// 年度結算，出自 `sub_30003`（`docs/re/25`）。跨年時跑一次，不是每回合。
//
// 這一支的存在解掉了一個久懸的誤判：`docs/re/21` §3 曾寫「人口只減不增」，
// 那是因為當時的掃描 pattern 漏掉相對定址（`docs/re/22`）。人口確實會長，
// 只是一年一次，藏在這裡。

// PopulationGrowthDivisor 是人口年成長的除數：每年 **+1/10**。
//
//	Real(人口) / Real(84h,0,2000h)   ← 48-bit Real 常數 84h/0/2000h = 10
//	Trunc
//	人口 += 商
//
// `Trunc` 是無條件捨去，所以人口小於 10 的省一個人都不會長。
const PopulationGrowthDivisor = 10

// PopulationGrowthCap 是成長的上限：人口 **< 1,000,000** 才成長。
//
//	cmp  word ptr es:[di+0Eh], 0Fh      ; 高位
//	cmp  word ptr es:[di+0Ch], 4240h    ; 低位   → 0x000F4240
//
// 到頂就完全停住——不是「成長變慢」，是不再執行那一段。
const PopulationGrowthCap = 1_000_000

// YearEndFlagMask 是跨年時保留的旗標位元（`and al, 7Dh`）。
//
// 清掉 bit 1（`02h`，語意未解）與 bit 7（`80h` = 本月已徵稅）。
// 徵稅旗標除了每月重置，**跨年也會再清一次**。
// bit 2（本回合已處理）與 bit 6（交戰中）不動。
const YearEndFlagMask = 0x7D

// HistoryCheckPenaltyDivisor 是歷史檢查沒過時的懲罰：帶兵能力 `shr 1`。
const HistoryCheckPenaltyDivisor = 2

// 三個分期的年份門檻（民國紀年）。`byte_6FE7D < 門檻` 時才做歷史檢查。
//
//	1 → 民國 17 年 = 1928，北伐完成／東北易幟
//	2 → 民國 28 年 = 1939
//	3 → 民國 33 年 = 1944
//
// ⚠️ 與 `Stage1Deadline`（民國 26 年 7 月，`gameover.go`）是兩件事：
// 那個結束遊戲，這個只是每年的懲罰閘。
var StageHistoryDeadline = map[uint8]uint8{1: 17, 2: 28, 3: 33}

// ManchurianFlagEventYear 是「東北易幟」事件的年份：民國 17 年。
//
// 效忠將領 166（張作霖）且還在場的將領，番號勢力名改成 3（`FAN(1).15` = 安國軍）。
// 畫面印「張作霖　更改　番號　為　安　國　軍」——程式碼、詞表、史實三方吻合
// （`docs/re/25` §5）。
const ManchurianFlagEventYear = 17

// ManchurianFlagEventLeader 是被檢查的效忠勢力值。
//
// 這個值同時是拿去查 `MAN115` 的索引（1-based）——也就是說
// **將領 `+0Eh`「效忠勢力」存的就是該勢力領袖自己的將領編號**。
const ManchurianFlagEventLeader = 166

// ManchurianFlagEventFaction 是事件後寫進將領 `+1Ch` 的番號勢力名。
const ManchurianFlagEventFaction = 3

// YearEndReport 記錄一次年度結算做了什麼，方便測試與逐項對照原版。
type YearEndReport struct {
	// Year 是結算後的年份（原版先 `inc` 再做事）。
	Year uint8
	// Grown 是人口有成長的省數，Growth 是總成長人數。
	Grown  int
	Growth uint32
	// FlagsCleared 是被清掉旗標的省數（全部有效省都會清）。
	FlagsCleared int
	// Reflagged 是東北易幟改到的將領數。
	Reflagged int
	// Penalised 是歷史檢查沒過而被砍能力的將領數。
	Penalised int
}

// AdvanceYear 跑一次年度結算。
//
// `st` 進來時是**結算前**的狀態；回傳的 report 帶結算後的年份。
//
// `gens` 與 `w.Units` 一一對應：兩者是**同一張 33 byte 表**的兩種視角
// （基址 `7A7Dh`）——能力值與番號在 `General`，效忠勢力 `+0Eh` 與在場旗標
// `+10h` 在 `CombatUnit`。年度結算兩邊都要改，所以兩個都得傳。
//
// `historyCheckPassed` 回報該分期的歷史條件過了沒——
// 原版比的是六個語意未知的全域變數（`docs/re/25` §6），
// 所以這裡把它做成外部注入的述詞，**不假裝知道那六個變數是什麼**。
// 傳 nil 視同通過（不懲罰）。
func (w *AIWorld) AdvanceYear(st *GameState, gens []General, playerFaction GeneralID, historyCheckPassed func(stage uint8) bool) YearEndReport {
	st.Year++
	rep := YearEndReport{Year: st.Year}

	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		prov, err := w.Table.At(p)
		if err != nil {
			continue
		}
		if prov.Population < PopulationGrowthCap {
			d := prov.Population / PopulationGrowthDivisor
			if d > 0 {
				prov.Population += d
				rep.Grown++
				rep.Growth += d
			}
		}
		prov.Flags &= YearEndFlagMask
		rep.FlagsCleared++
	}

	if st.Year == ManchurianFlagEventYear {
		rep.Reflagged = w.applyManchurianFlagEvent(gens)
	}

	// 歷史檢查：年份還沒到門檻、而且條件沒達成 → 玩家全將領帶兵能力減半。
	// 每年檢查一次，所以會反覆觸發、指數衰減。
	if deadline, ok := StageHistoryDeadline[st.Stage]; ok && st.Year < deadline {
		if historyCheckPassed != nil && !historyCheckPassed(st.Stage) {
			rep.Penalised = w.halveFactionAbility(gens, playerFaction)
		}
	}
	return rep
}

// applyManchurianFlagEvent 是民國 17 年的東北易幟。
//
// 原版另外畫一個框並印一行字，那是呈現層的事（`internal/ui`），
// 規則層只負責改資料。
func (w *AIWorld) applyManchurianFlagEvent(gens []General) int {
	n := 0
	for i := range w.Units {
		u := &w.Units[i]
		if u.Faction != ManchurianFlagEventLeader || !u.Active || i >= len(gens) {
			continue
		}
		gens[i].FactionName = ManchurianFlagEventFaction
		n++
	}
	return n
}

// halveFactionAbility 把某個勢力所有將領的帶兵能力砍半（`shr 1`）。
//
// ⚠️ 原版不看 `Active`，在場與否都砍。
func (w *AIWorld) halveFactionAbility(gens []General, faction GeneralID) int {
	n := 0
	for i := range w.Units {
		if w.Units[i].Faction != faction || i >= len(gens) {
			continue
		}
		gens[i].AbilityA /= HistoryCheckPenaltyDivisor
		n++
	}
	return n
}
