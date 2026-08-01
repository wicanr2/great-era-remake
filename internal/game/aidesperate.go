package game

// 「這個省已經走投無路」的判定（`sub_17019`，84 行），以及它底下的
// 省兵力總和（`sub_306CF`，98 行）。
//
// 兩個呼叫端：
//
//	決策鏈 A 步驟 3（`sub_17ADA+314`）—— 掃鄰省時與 `sub_16F89`（優勢比）並列
//	出兵準備 `sub_174C9+CC`
//
// 讀作：**補給撐不住又被包圍 → 只能拚**。

// ProvinceForceTotal 是某省效忠其司令的可用部隊的**兵力總和**（`sub_306CF`）。
//
// 篩選條件與 `GeneralCount`（`sub_5A881`）逐字相同：
//
//	將領 +4  == 省編號
//	將領 +16 bit 0 == 1
//	將領 +14 == 該省的司令（省份 +20）
//
// 通過的累加 `+17`（兵力）成 **32-bit 總和**——原版特地用 `add`/`adc`
// 湊成 long 回傳，代表它預期總和會超過 65,535（20,000 的步兵師三個就滿了）。
//
// ⚠️ 原版開頭有一段**死碼**：先讀省份 `+32` 檢查 bit 6（交戰中），
// 若成立就把回傳值清零——但 `loc_3077C` 之後又無條件用累加結果覆寫，
// 那次清零完全沒有效果。所以**交戰中的省不會被特別處理**，
// 這裡照抄實際行為（不模擬那段死碼）。
//
// 這支是通用工具，`sub_13D23`／`sub_13EA8` 等多處都在用，
// 不是決策鏈 A 專屬。
func ProvinceForceTotal(p ProvinceID, units []CombatUnit, gens []General, commander GeneralID) int {
	total := 0
	for i := range units {
		u := &units[i]
		if !u.Active || u.Province != p || u.Faction != commander {
			continue
		}
		if i >= len(gens) {
			continue
		}
		total += int(gens[i].Force)
	}
	return total
}

// ProvinceForceTotal 是 AIWorld 上的包裝，司令直接從省份記錄讀。
func (w *AIWorld) ProvinceForceTotal(p ProvinceID, gens []General) int {
	prov, err := w.Table.At(p)
	if err != nil {
		return 0
	}
	return ProvinceForceTotal(p, w.Units, gens, prov.Commander)
}

// 撐不住的門檻：拿兵力總和換算成「這些兵最起碼要有多少存糧」。
//
//	黃金需求 = 兵力總和 ÷ 100
//	糧食需求 = 兵力總和 ÷ 10
//
// 比值是 **1:10**，與 `AISupplyOK` 進階模式的 `need` 與 `need*10`
// 完全一致（`aisupply.go`）——同一套「黃金:糧食 = 1:10」的比例，
// 只是除數不同：`sub_16B44` 用 50／70（動不動的門檻），
// 這裡用 100（撐不撐得住的門檻），所以**這一關比補給門檻寬鬆一倍以上**。
//
// 讀作：`sub_16B44` 問「有本錢行動嗎」，`sub_17019` 問「快餓死了嗎」。
const (
	AIDesperateGoldDiv = 100
	AIDesperateFoodDiv = 10
)

// AISupplyStarving 回報某省的存糧是不是撐不住它的部隊（`sub_17019` 前半）。
//
//	黃金 < 兵力總和/100  或  糧食 < 兵力總和/10   → true
//
// **是「或」不是「且」**，任一項不足就算。
func (w *AIWorld) AISupplyStarving(p ProvinceID, gens []General) bool {
	prov, err := w.Table.At(p)
	if err != nil {
		return false
	}
	force := w.ProvinceForceTotal(p, gens)
	if int(prov.Gold) < force/AIDesperateGoldDiv {
		return true
	}
	return int(prov.Food) < force/AIDesperateFoodDiv
}

// AIDesperate 是 `sub_17019` 本身：**撐不住而且被包圍**。
//
//	AISupplyStarving(p) && Encircled(p)
//
// 兩個條件是「且」——只缺糧不算，只被包圍也不算。
// 這與步驟 6 的 `sub_5A9F6`（單看被包圍就突圍）形成對照：
// 步驟 3 掃鄰省時要求更嚴，因為那是「要不要為了別人的省動用自己的兵」。
//
// ⚠️ `Encircled` 對四面環海的孤島省會回 true（原版行為，見 `aiencircled.go`），
// 所以一個缺糧的孤島省會被判成 desperate。照抄，不修。
func (w *AIWorld) AIDesperate(p ProvinceID, gens []General) bool {
	return w.AISupplyStarving(p, gens) && w.Encircled(p)
}
