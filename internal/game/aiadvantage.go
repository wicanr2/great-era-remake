package game

// 電腦的**優勢比門檻**（`sub_16F89`，83 行）。
//
// 決策鏈 A 的**步驟 1 與步驟 3 都用它**（`docs/re/28` §2）——
// 步驟 1 拿它決定「出兵（`sub_174C9`）還是只調動（`sub_14F9A`）」，
// 所以這是「電腦要不要打」最直接的一道判斷。

// 優勢倍數：**目標省的戰力乘上它，還小於等於本省戰力**才算有優勢。
//
//	高難度（byte_6FFCA & 1）：4 倍
//	低難度：                  3 倍
//
// ⚠️ **高難度反而更嚴格**（4 > 3），讀起來像反直覺，但與
// `sub_16B44` 的補給門檻是同一個方向（高難度要求更多存糧才敢動，
// `70-ai.md` §3x）。合起來的行為是：**高難度的電腦更謹慎，
// 只在絕對優勢時才出手**——它不亂送兵，實際上更難對付。
//
// 這是 `byte_6FFCA` bit 0 的**第四個用途**。
const (
	AIAdvantageRatioHard = 4
	AIAdvantageRatioEasy = 3
)

// AIHasAdvantage 判斷本省對目標省有沒有出兵的戰力優勢（`sub_16F89`）。
//
//	mine   = ProvinceStrength(from)     ; sub_5B983
//	theirs = ProvinceStrength(target)
//	回 theirs × 倍數 ≤ mine
//
// `sub_5B983` 的內容與 `ProvinceStrength` 完全相同：掃全部將領，
// 取 `+4`（所屬省）符合且 `+16` bit 0 為 1 的，累加 `sub_5A0B9`
// 的攻擊力。**AI 的評估函式就是攻擊力公式**——這是第三個獨立的例子
// （前兩個是調動 `70-ai.md` §6e、談判停火 `50-diplomacy.md` §1z）。
func (w *AIWorld) AIHasAdvantage(from, target ProvinceID, hard bool) bool {
	ratio := AIAdvantageRatioEasy
	if hard {
		ratio = AIAdvantageRatioHard
	}
	mine := ProvinceStrength(from, w.Units, w.Strengths, w.Opts)
	theirs := ProvinceStrength(target, w.Units, w.Strengths, w.Opts)
	return theirs*ratio <= mine
}

// AIAdvantageRatio 回傳當前難度要求的倍數，方便呼叫端顯示或除錯。
func AIAdvantageRatio(hard bool) int {
	if hard {
		return AIAdvantageRatioHard
	}
	return AIAdvantageRatioEasy
}
