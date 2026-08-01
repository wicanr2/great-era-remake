package game

// 電腦的**季節性內政**（`sub_1ACCC` 的第一段，`docs/re/32`）。
//
// `byte_6FE7E` 是月份，所以那七個分支是「一年裡七個節點做不同的事」。
// 這裡只接**規則已經全解、而且知道對價是什麼**的那幾條：
//
//	1／3／8／11 月   徵兵（`sub_19563`）
//	6 月            人民忠誠度 +20（`sub_1A100`）
//
// ⚠️ **沒接的**與理由：
//
//	3／7／11 月 將領 +1 加 20（`sub_18903`）——`+1` 是忠誠度，
//	            但那一支還沒讀完整條件
//	1／8／12 月 黃金支出（`sub_19CD4` −5,000／`sub_197DB` −10,000）——
//	            **只看到扣錢，不知道換到什麼**。接了會讓電腦白白變窮
//	6 月       地價成長（`sub_189C8`）
//
// 照實接一半勝過假裝跑了完整的內政。

// AIInteriorMonths 是「這個月要不要跑內政」的對照表（`docs/re/32` §1）。
//
// **2／4／5／9／10 月沒有內政。**
var AIInteriorMonths = map[uint8]bool{
	1: true, 3: true, 6: true, 7: true, 8: true, 11: true, 12: true,
}

// AIRecruitMonths 是徵兵的月份：**1／3／8／11**（`sub_19563` 的四個呼叫點）。
//
// 3 月與 11 月還有黃金門檻（5,000／3,000），1 月與 8 月無條件。
var AIRecruitMonths = map[uint8]int{
	1:  0,    // 無條件
	3:  5000, // 黃金 ≥ 5,000
	8:  0,    // 無條件
	11: 3000, // 黃金 ≥ 3,000
}

// AIComfortMonth 是人民忠誠度 +20 的月份（`sub_1A100`）。
const AIComfortMonth = 6

// AILoyaltyGain 是那一次加多少，夾到 100。
const AILoyaltyGain = 20

// InteriorResult 是一次內政的結果。
type InteriorResult struct {
	// Recruited 是這次補進去的兵。
	Recruited int
	// GoldSpent 是花掉的黃金。
	GoldSpent int
	// LoyaltyGained 為真表示跑了 6 月的忠誠度補充。
	LoyaltyGained bool
}

// RunInterior 跑某省這個月的內政（`docs/re/32`）。
//
// `gens` 會被就地修改（兵力 `+17`、士兵戰技 `+19`、`+20`）。
func (w *AIWorld) RunInterior(p ProvinceID, month uint8, gens []General,
	bonusMask uint8) InteriorResult {
	var out InteriorResult
	prov, err := w.Table.At(p)
	if err != nil || prov.Commander == 0 {
		return out
	}

	// 6 月：人民忠誠度 +20，夾 100。
	if month == AIComfortMonth {
		if prov.Loyalty < 100 {
			if v := int(prov.Loyalty) + AILoyaltyGain; v > 100 {
				prov.Loyalty = 100
			} else {
				prov.Loyalty = uint8(v)
			}
			out.LoyaltyGained = true
		}
	}

	// 1／3／8／11 月：徵兵。
	if minGold, ok := AIRecruitMonths[month]; ok && int(prov.Gold) >= minGold {
		r := w.recruitOnce(p, prov, gens, bonusMask)
		out.Recruited, out.GoldSpent = r.Recruited, r.GoldSpent
	}
	return out
}

// recruitOnce 是 `sub_19563` 的一次徵兵。
func (w *AIWorld) recruitOnce(p ProvinceID, prov *Province, gens []General,
	bonusMask uint8) InteriorResult {
	var out InteriorResult

	// 觸發：**只要有一個部隊沒滿員**（`sub_18CD5` + `sub_18A67`）。
	roster := w.rosterOfProvince(p, prov.Commander, gens)
	if !anyUnderStrength(roster, gens) {
		return out
	}

	// 門檻（`sub_19563`）。
	force := w.ProvinceForceTotal(p, gens)
	upkeep := force / 100 * 3
	if upkeep > AIRecruitUpkeepCap {
		upkeep = AIRecruitUpkeepCap
	}
	spare := int(prov.Gold) - upkeep
	if spare <= 0 || int(prov.Gold) < AIRecruitMinGold {
		return out
	}
	if int(prov.Population) < force/100 || prov.Population < AIRecruitMinPopulation {
		return out
	}

	// 預算（`sub_18D45`）：min(黃金 × 2/3, 步兵缺額 ÷ 10)，夾難度與絕對上限。
	budget := AIRecruitBudget(int(prov.Gold), infantryGap(roster, gens), bonusMask)
	if budget <= 0 {
		return out
	}

	// 分配（`sub_19443` → 三支策略）。這裡只實作第三支（平分），
	// 因為前兩支要「當前指定的將領」與排序後的清單，
	// 而 remake 的呼叫端還沒有那些狀態。**標為差異。**
	under := underStrengthUnits(roster, gens)
	if len(under) == 0 {
		return out
	}
	share := AIRecruitSplitShare(budget, len(under))
	for i, g := range under {
		if i >= AIRecruitSplitMax {
			break
		}
		idx := int(g) - 1
		added, spent := AIRecruitInto(gens[idx].Branch, gens[idx].Force, share)
		if added == 0 {
			continue
		}
		gens[idx].Force += uint16(added)
		gens[idx].F19, gens[idx].F20 = AIRecruitQualityCost(
			gens[idx].Branch, gens[idx].F19, gens[idx].F20)
		out.Recruited += added
		out.GoldSpent += spent
	}

	// 扣費（`sub_194F4`）：黃金 −花費、人口 −Round(花費 ÷ 10)。
	if out.GoldSpent > 0 {
		if int(prov.Gold) < out.GoldSpent {
			out.GoldSpent = int(prov.Gold)
		}
		prov.Gold -= uint16(out.GoldSpent)
		popCost := (out.GoldSpent + 5) / 10 // Round
		if uint32(popCost) > prov.Population {
			popCost = int(prov.Population)
		}
		prov.Population -= uint32(popCost)
	}
	return out
}

// 徵兵的門檻常數（`sub_19563`／`sub_18D45`，`docs/re/32` §6）。
const (
	// AIRecruitUpkeepCap 是養兵成本的上限：兵力 ÷ 100 × 3，夾在這裡。
	AIRecruitUpkeepCap = 4000
	// AIRecruitMinGold 是徵兵的黃金下限。
	AIRecruitMinGold = 500
	// AIRecruitMinPopulation 是人口下限。
	AIRecruitMinPopulation = 200
	// AIRecruitBudgetCap 是花費的絕對上限。
	AIRecruitBudgetCap = 100000
	// AIRecruitBudgetHard / AIRecruitBudgetEasy 是難度上限
	// （`byte_6FFCA` bit 4 或 bit 2 → 20,000；bit 5 → 40,000）。
	AIRecruitBudgetHard = 20000
	AIRecruitBudgetEasy = 40000
)

// AIRecruitBudget 算這次徵兵的花費上限（`sub_18D45`）。
//
//	min(黃金 × 2/3, 步兵缺額 ÷ 10)，夾難度上限與 100,000
func AIRecruitBudget(gold, infantryGap int, bonusMask uint8) int {
	budget := gold / 3 * 2
	if need := (infantryGap + 5) / 10; need < budget {
		budget = need
	}
	switch {
	case bonusMask&0x10 != 0 || bonusMask&0x04 != 0:
		if budget > AIRecruitBudgetHard {
			budget = AIRecruitBudgetHard
		}
	case bonusMask&0x20 != 0:
		if budget > AIRecruitBudgetEasy {
			budget = AIRecruitBudgetEasy
		}
	}
	if budget > AIRecruitBudgetCap {
		budget = AIRecruitBudgetCap
	}
	return budget
}

// rosterOfProvince 是某省效忠其司令的可用將領（與 `GeneralCount` 同一組條件）。
func (w *AIWorld) rosterOfProvince(p ProvinceID, faction GeneralID,
	gens []General) []GeneralID {
	var out []GeneralID
	for i := range w.Units {
		u := &w.Units[i]
		if !u.Active || u.Province != p || u.Faction != faction || i >= len(gens) {
			continue
		}
		out = append(out, GeneralID(i+1))
	}
	return out
}

// anyUnderStrength 回報有沒有任何一個部隊沒滿員（`sub_18A67` 的反面）。
func anyUnderStrength(roster []GeneralID, gens []General) bool {
	return len(underStrengthUnits(roster, gens)) > 0
}

// underStrengthUnits 挑出沒滿員的部隊。
func underStrengthUnits(roster []GeneralID, gens []General) []GeneralID {
	var out []GeneralID
	for _, g := range roster {
		i := int(g) - 1
		if i < 0 || i >= len(gens) {
			continue
		}
		full := BranchFullStrength(gens[i].Branch)
		if full == 0 {
			continue // 未知兵種，原版四個 cmp 都不中
		}
		if gens[i].Force < full {
			out = append(out, g)
		}
	}
	return out
}

// infantryGap 是**步兵**的總缺額（`sub_18D45` 第一段，只算兵種 1）。
func infantryGap(roster []GeneralID, gens []General) int {
	full := int(BranchFullStrength(BranchInfantry))
	gap := 0
	for _, g := range roster {
		i := int(g) - 1
		if i < 0 || i >= len(gens) || gens[i].Branch != BranchInfantry {
			continue
		}
		gap += full - int(gens[i].Force)
	}
	if gap < 0 {
		return 0
	}
	return gap
}
