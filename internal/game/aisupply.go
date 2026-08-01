package game

// 電腦決策鏈的補給門檻（`sub_16B44`，179 行）。
//
// **這一支在決策鏈 A 的六個步驟裡出現四次**（步驟 1、3、5、6，`docs/re/28` §2）——
// 因為每個「要花資源的步驟」都先問一句「這個省的補給夠不夠」。
// 解開它等於一次補上四步的前提。

// 補給門檻的常數。分兩套：`byte_6FFCA & 4` 決定走哪一套。
//
// bit 2 就是 `70-ai.md` §6d 記的「啟用決策鏈最後三步」那個位元——
// 原來它同時也切換補給判定的嚴謹度。
const (
	// AISupplySimpleGold / AISupplySimpleFood 是**簡單模式**（bit 2 未設）的門檻：
	// 黃金 ≥ 1,000 且 糧食 ≥ 10,000 就算補給充足。
	AISupplySimpleGold = 1000
	AISupplySimpleFood = 10000

	// AISupplyUpkeepDivHard / AISupplyUpkeepDivEasy 是**進階模式**下
	// 「養這些兵要多少黃金」的除數（`byte_6FFCA & 1` 選一個）：
	//
	//	需求 = 兵力總和 ÷ 除數
	//
	// 除數越小需求越大，所以**高難度（50）比低難度（70）要求更多存糧**。
	AISupplyUpkeepDivHard = 50
	AISupplyUpkeepDivEasy = 70

	// AISupplyRichGold / AISupplyRichFood 是進階模式的另一條捷徑：
	// 黃金 ≥ 4,000 且 糧食 ≥ 25,000 就直接算充足，不看兵力。
	AISupplyRichGold = 4000
	AISupplyRichFood = 25000

	// AISupplyAmmo / AISupplyFuel 是**最後一關**，兩套模式都要過：
	// 彈藥 ≥ 15,000 且 燃料 ≥ 10,000。
	//
	// 前面的黃金糧食判斷再怎麼寬鬆，這兩個不夠就一律回 false。
	AISupplyAmmo = 15000
	AISupplyFuel = 10000
)

// AISupplyOK 判斷某省的補給夠不夠讓電腦行動（`sub_16B44`）。
//
//	簡單模式（bit 2 未設）：
//	    黃金 ≥ 1000 且 糧食 ≥ 10000
//
//	進階模式（bit 2 已設）：
//	    需求 = 兵力總和 ÷ (高難度 50 ／ 低難度 70)
//	    需求 ≤ 黃金 且（需求 × 10 ≤ 糧食 或 糧食 ≥ 25000）  → 充足
//	    或 黃金 ≥ 4000 且 糧食 ≥ 25000                      → 充足
//
//	兩套都要再過：彈藥 ≥ 15000 且 燃料 ≥ 10000
//
// `totalForce` 是該勢力的兵力總和——原版讀狀態結構的 `ss:[di-236h]`
// （32-bit）。⚠️ **那一格的語意是假說**：形狀與用法都像兵力總和
// （步驟 5 也拿它跟 63,392 比），但沒有直接證據，所以由呼叫端傳入。
//
// `advanced` 來自 `byte_6FFCA & 4`，`hard` 來自 `byte_6FFCA & 1`。
func (w *AIWorld) AISupplyOK(p ProvinceID, totalForce int, advanced, hard bool) bool {
	prov, err := w.Table.At(p)
	if err != nil {
		return false
	}
	gold, food := int(prov.Gold), int(prov.Food)

	ok := false
	if advanced {
		div := AISupplyUpkeepDivEasy
		if hard {
			div = AISupplyUpkeepDivHard
		}
		need := totalForce / div
		if need <= gold && (need*10 <= food || food >= AISupplyRichFood) {
			ok = true
		}
		if gold >= AISupplyRichGold && food >= AISupplyRichFood {
			ok = true
		}
	} else {
		ok = gold >= AISupplySimpleGold && food >= AISupplySimpleFood
	}
	if !ok {
		return false
	}
	// 最後一關：兩套模式都要過。
	return int(prov.Ammo) >= AISupplyAmmo && int(prov.Fuel) >= AISupplyFuel
}

// AIUpkeepNeed 回傳「養這些兵需要多少黃金」，方便呼叫端顯示或除錯。
func AIUpkeepNeed(totalForce int, hard bool) int {
	div := AISupplyUpkeepDivEasy
	if hard {
		div = AISupplyUpkeepDivHard
	}
	return totalForce / div
}
