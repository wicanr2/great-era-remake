package game

// 「那個省孤立無援」的判定（`sub_17019`，84 行），以及它底下的
// 省兵力總和（`sub_306CF`，98 行）。
//
// 兩個呼叫端傳進來的都是**敵方的目標省**，不是自己的省：
//
//	決策鏈 A 步驟 3（`sub_17ADA+314`）：掃鄰省時對每個**敵省**問，
//	    與 `sub_16F89`（我方有沒有 3–4 倍優勢）**二選一成立就收進候選**
//	出兵準備 `sub_174C9+CC`：對已經挑好的**目標省**再問一次
//
// 所以它是**軟柿子判定**：那個省缺糧、而且四面都是它的敵人 → 孤立無援，
// 打它划算。與 `sub_16F89` 是同一類判斷的兩種角度：
//
//	sub_16F89  我方壓倒它
//	sub_17019  它自己撐不住
//
// ⚠️ `Encircled` 用的是**該省自己的司令**當基準，所以問敵省時算的是
// 「它被它的敵人包圍」——我方是那些敵人之一。這正是想要的語意。

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
// 讀作：`sub_16B44` 問「**我**有本錢行動嗎」，`sub_17019` 問「**它**快餓死了嗎」。
// 兩支的對象不同——一個問自己，一個問目標。
const (
	AIDesperateGoldDiv = 100
	AIDesperateFoodDiv = 10
)

// AISupplyStarving 回報某省的存糧是不是撐不住它自己的部隊（`sub_17019` 前半）。
//
// 用在敵省身上就是「它快斷炊了」。
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

// AIDesperate 是 `sub_17019` 本身：**那個省缺糧而且被包圍**。
//
//	AISupplyStarving(p) && Encircled(p)
//
// 兩個條件是「且」——只缺糧不算，只被包圍也不算。
//
// ⚠️ 傳進來的通常是**敵方的目標省**（見檔頭）。這與步驟 6 的
// `sub_5A9F6` 形成對照：那裡問的是**自己的省**被不被包圍（要突圍），
// 這裡問的是**別人的省**撐不撐得住（要不要打它）。
// 同一個 `Encircled` 換個對象問，意思完全相反。
//
// ⚠️ `Encircled` 對四面環海的孤島省會回 true（原版行為，見 `aiencircled.go`），
// 所以一個缺糧的孤島省會被判成 desperate。照抄，不修。
func (w *AIWorld) AIDesperate(p ProvinceID, gens []General) bool {
	return w.AISupplyStarving(p, gens) && w.Encircled(p)
}
