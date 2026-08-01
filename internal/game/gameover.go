package game

// 遊戲結束的判定，出自 `sub_1060A`（`docs/re/16`）。
//
// 主迴圈每一輪跑完就呼叫它，回非 0 就 `CLOSEGRAPH` + `Halt`：
//
//	loc_11343:
//	    call    sub_1060A
//	    or      al, al
//	    jnz     short loc_11360      ; ← 結束
//	    ...
//	    jmp     loc_10E55            ; 否則繼續主迴圈
//
// 結束時把一個**結局代碼**寫進 `Config.dat`（`sub_10541`），
// 由 `GRTE.EXE` 讀去播對應的結局——這印證了 `CLAUDE.md` §3.7
// 「模組間靠檔案交接狀態」的假說。

// EndingCode 是寫進 `Config.dat` 的結局代碼。
type EndingCode uint8

const (
	// EndingNone 表示遊戲還沒結束。
	EndingNone EndingCode = 0
	// EndingUnifiedStage1 是第一期統一（`sub_10541(1)`）。
	EndingUnifiedStage1 EndingCode = 1
	// EndingUnifiedLater 是其他期統一（`sub_10541(2)`）。
	EndingUnifiedLater EndingCode = 2
	// EndingDefeatStage1 是第一期覆滅（`sub_10541(20h)`）。
	EndingDefeatStage1 EndingCode = 0x20
	// EndingDefeatLater 是其他期覆滅（`sub_10541(40h)`）。
	EndingDefeatLater EndingCode = 0x40
)

// Stage1Deadline 是第一期的時限：**民國 26 年 7 月**。
//
//	cmp     byte_6FE7D, 1Ah         ; 年 == 26
//	cmp     byte_6FE7E, 7           ; 月 >= 7
//
// 民國 26 年 7 月 = **1937 年 7 月 = 盧溝橋事變**。北伐那一期到此
// 強制結束，接抗戰——與 `CLAUDE.md` §1.5 記的三期分法完全吻合。
const (
	Stage1DeadlineYear  = 26
	Stage1DeadlineMonth = 7
)

// GameState 是判定要用的世界狀態。
type GameState struct {
	// Stage 是 `byte_6FE88`。1 = 第一期。
	Stage uint8
	// Year / Month 是 `byte_6FE7D` / `byte_6FE7E`（民國紀年）。
	Year, Month uint8
	// Watching 是 `byte_7003A`（觀戰模式）。
	Watching bool
	// MajorPower 回報某個將領是不是十大勢力的領袖（`sub_5C7FE`）。
	// nil 表示沒有這份資訊，判定會保守地回「還沒結束」。
	MajorPower func(GeneralID) bool
	// LeaderAvailable 回報某個領袖還在不在（將領 `+16` bit 0）。
	LeaderAvailable func(GeneralID) bool
}

// CheckGameOver 判定遊戲是否結束，並回傳要寫進 `Config.dat` 的結局代碼。
//
// 兩條路徑，照 `sub_1060A`：
//
//	觀戰模式：只看「天下是不是統一」，統一就印 "Over !!" 收工
//	正常遊戲：玩家覆滅 或 統一 或 第一期時限到
func (w *AIWorld) CheckGameOver(st GameState) (over bool, code EndingCode) {
	if st.Watching {
		// 觀戰：掃全部有主省，只要出現第二個不同的司令就不算統一。
		var last GeneralID
		unified := true
		for p := ProvinceID(1); p <= ProvinceCount; p++ {
			prov, err := w.Table.At(p)
			if err != nil || prov.Commander == 0 || prov.Commander == last {
				continue
			}
			if last != 0 {
				unified = false
			}
			last = prov.Commander
		}
		// 原版這條路徑只印 "Over !!"，不寫結局代碼。
		return unified, EndingNone
	}

	if st.MajorPower == nil || st.LeaderAvailable == nil {
		return false, EndingNone
	}

	// 一、玩家覆滅：沒有任何十大勢力的領袖還可用。
	defeated := true
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		prov, err := w.Table.At(p)
		if err != nil {
			continue
		}
		if st.MajorPower(prov.Commander) && st.LeaderAvailable(prov.Commander) {
			defeated = false
			break
		}
	}

	// 二、第一期的時限：民國 26 年 7 月一到就強制結束，
	//     **走的是「覆滅」那條路**（原版直接把旗標設為 1）。
	if st.Stage == 1 && st.Year == Stage1DeadlineYear && st.Month >= Stage1DeadlineMonth {
		defeated = true
	}

	// 三、統一：十大勢力裡只剩一個還可用，而且所有有主省都歸十大勢力。
	unified := false
	if !defeated {
		alive := 0
		for _, id := range w.majorLeaders(st) {
			if st.LeaderAvailable(id) {
				alive++
			}
		}
		if alive == 1 {
			unified = true
			for p := ProvinceID(1); p <= ProvinceCount; p++ {
				prov, err := w.Table.At(p)
				if err != nil || prov.Commander == 0 {
					continue
				}
				if !st.MajorPower(prov.Commander) {
					unified = false
					break
				}
			}
		}
	}

	switch {
	case defeated:
		if st.Stage == 1 {
			return true, EndingDefeatStage1
		}
		return true, EndingDefeatLater
	case unified:
		if st.Stage == 1 {
			return true, EndingUnifiedStage1
		}
		return true, EndingUnifiedLater
	}
	return false, EndingNone
}

// majorLeaders 列出目前還掛在地圖上的十大勢力領袖。
//
// ⚠️ 原版掃的是一張 **24 格的 word 表**（`[-534Ah]`，`sub_1060A`
// 與 `sub_15018` 都用它），內容是執行期填的、靜態看不到
// （`docs/re/06` §3 那條保留仍然成立）。這裡從省份的司令欄位反推，
// **標為 remake 差異**——若原版那張表包含已經失去所有省份的領袖，
// 兩者的「還剩幾個」會不同。
func (w *AIWorld) majorLeaders(st GameState) []GeneralID {
	seen := map[GeneralID]bool{}
	var out []GeneralID
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		prov, err := w.Table.At(p)
		if err != nil || prov.Commander == 0 || seen[prov.Commander] {
			continue
		}
		if st.MajorPower(prov.Commander) {
			seen[prov.Commander] = true
			out = append(out, prov.Commander)
		}
	}
	return out
}
