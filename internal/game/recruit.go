package game

import "fmt"

// 徵兵（政略指令 5）。
//
// `docs/playtest/09` 的端到端實跑暴露了一個真實缺口：**規則層只有消耗
// 沒有補充**，300 回合後 31 個有主省裡 20 個一個將領都沒有。
// 徵兵是原版五條補充管道裡最直接的一條，先補它。
//
// 成本與上限來自實機（`docs/playtest/10` §2），兩個獨立樣本 + 一次
// 實際計算，零誤差。**分配規則只解了一半**，見 `Recruit` 的說明。

// RecruitSoldiersPerGold 是一塊黃金能徵幾個兵。
//
// 實機兩個樣本：
//
//	湖北 黃金 4,150 → 徵兵上限 (0-41,500)
//	河南 黃金 5,950 → 徵兵上限 (0-59,500)
//
// 加上輸入 1,000 兵時畫面顯示「共須黃金 100」——**10 兵 = 1 金**。
const RecruitSoldiersPerGold = 10

// RecruitLimit 是這個省這次最多能徵幾個兵。
//
// 原版把它直接印在提示裡：「司令，欲徵多少步兵？ (0-41500)」。
func (w *AIWorld) RecruitLimit(p ProvinceID) int {
	prov, err := w.Table.At(p)
	if err != nil {
		return 0
	}
	return int(prov.Gold) * RecruitSoldiersPerGold
}

// RecruitCost 是徵 n 個兵要花多少黃金。
func RecruitCost(n int) int {
	return n / RecruitSoldiersPerGold
}

// RecruitBranchOrder 是徵兵子選單的四個兵種，**照畫面順序**：
//
//  1. 徵步兵    2. 徵裝甲兵
//  3. 徵砲兵    4. 徵騎兵
//
// ⚠️ 選單順序與**兵種編號**不同（步 1、砲 4、裝甲 5、騎 6，
// `docs/mechanics/20-military.md` §0）。UI 照這個順序，規則層用編號。
var RecruitBranchOrder = [4]uint8{
	BranchInfantry, BranchArmour, BranchArtiller, BranchCavalry,
}

// Recruit 對某省徵 n 個某兵種的兵，回傳實際補進去的人數。
//
// **分配規則只解了一半。** `sub_28259` 掃該省的將領時檢查五個條件：
//
//	將領.+21 == 兵種        同兵種
//	將領.+4  == 目標省
//	將領.+16 == 1           可用
//	將領.+14 == 我方勢力
//	將領.+17 與滿員數比較   ← 補到滿為止
//
// 這裡照這五條實作「依清單順序補到滿」，但**原版補的先後次序沒解**
// ——`sub_28259` 有 630 行，中間還有幾段 Real 運算沒讀。
// 順序會影響誰先補滿，所以標為 remake 差異。
func (w *AIWorld) Recruit(p ProvinceID, branch uint8, n int) (int, error) {
	prov, err := w.Table.At(p)
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		return 0, nil
	}
	if limit := w.RecruitLimit(p); n > limit {
		return 0, fmt.Errorf("game: 省 %d 最多徵 %d 人，要求 %d", p, limit, n)
	}
	full := int(BranchFullStrength(branch))
	if full == 0 {
		return 0, fmt.Errorf("game: 兵種 %d 沒有滿員數", branch)
	}

	added := 0
	for _, i := range w.RosterOf(p).ids {
		if n <= 0 {
			break
		}
		if i >= len(w.Strengths) {
			continue
		}
		s := &w.Strengths[i]
		if s.Branch != branch {
			continue
		}
		room := full - int(s.Force)
		if room <= 0 {
			continue
		}
		take := room
		if take > n {
			take = n
		}
		s.Force += uint16(take)
		n -= take
		added += take
	}
	if added > 0 {
		cost := RecruitCost(added)
		if int(prov.Gold) < cost {
			cost = int(prov.Gold)
		}
		prov.Gold -= uint16(cost)
	}
	return added, nil
}

// RecruitToFull 把某省所有同兵種的將領補到滿，受黃金上限約束。
//
// 這是規則層自己的便利函式（原版沒有這個一鍵操作），
// `cmd/aisim` 用它讓長跑不會枯竭。**標為 remake 行為。**
func (w *AIWorld) RecruitToFull(p ProvinceID, branch uint8) int {
	need := 0
	full := int(BranchFullStrength(branch))
	for _, i := range w.RosterOf(p).ids {
		if i < len(w.Strengths) && w.Strengths[i].Branch == branch {
			if room := full - int(w.Strengths[i].Force); room > 0 {
				need += room
			}
		}
	}
	if limit := w.RecruitLimit(p); need > limit {
		need = limit
	}
	added, err := w.Recruit(p, branch, need)
	if err != nil {
		return 0
	}
	return added
}

// ---------------------------------------------------------------------------
// 徵稅（政略指令 4），出自 `sub_303CD`（`docs/re/18`）。
// ---------------------------------------------------------------------------

// 徵稅的常數，全部從 `sub_303CD` 讀出來。
const (
	// TaxGoldBase 是黃金收入的固定加成（`add ax, 12Ch`）。
	TaxGoldBase = 300
	// TaxFoodBase 是糧食收入的固定加成（`add ax, 0BB8h`）。
	TaxFoodBase = 3000
	// TaxGoldMax / TaxGoldMin 是黃金收入的上下限（`1964h` / `5DCh`）。
	//
	// **下限 1,500 解開了實機三個樣本裡「+1500 出現兩次」的謎**
	// （`docs/playtest/12` §3）——那不是巧合，是被夾到下限。
	TaxGoldMax = 6500
	TaxGoldMin = 1500
	// TaxLoyaltyDrop 是徵稅的**第一段**忠誠度代價（`sub_303CD` 的
	// `sub ax, 1Eh`），不足 30 就歸 0。
	TaxLoyaltyDrop = 30
	// TaxLoyaltySecondDivisor 是**第二段**：`sub_2C6C6` 尾段再扣一次
	// 「剩餘忠誠度 ÷ 5」（`mov cx, 5 / div cx / sub [di-6222h], cl`）。
	//
	// 兩段合起來對上實機零誤差（`docs/re/18` §3b）：
	//
	//	河南 41 → 41−30 = 11 → 11÷5 = 2 → 11−2 = 9   ✓
	//	湖北 79 → 79−30 = 49 → 49÷5 = 9 → 49−9 = 40  ✓
	TaxLoyaltySecondDivisor = 5
	// 資源上限用 `resource.go` 的 `ResourceCap`（同一個 `sub_5A467`）。
	// TaxRandScale 是丟進 `Random()` 的倍率（Real 常數 `0x83/0/0x2000`）。
	TaxRandScale = 5
	// TaxFoodDivisor 是糧食那條的除數（Real 常數 `0x8E/0/0x1C40`）。
	TaxFoodDivisor = 10000
)

// TaxResult 是一次徵稅的結果。
type TaxResult struct {
	Gold, Food   int
	LoyaltyAfter uint8
}

// Tax 對某省徵稅，語意照 `sub_303CD`：
//
//	黃金 = clamp(Random(Round(Sqrt(人口) × 5)) + 300, 1500, 6500)
//	糧食 = Random(Round(人口 ÷ 10000 × 5)) + 3000
//	忠誠度 -= 30（不足 30 歸 0），然後再 -= 剩餘 ÷ 5
//
// 兩筆收入都經 `sub_5A467` 加上去，欄位上限 60,000。
// 徵稅完成後 `sub_2C6C6` 會設省份 `+32` 的 **bit 7 =「已徵過稅」**。
//
// ⚠️ **公式的結構與常數是 confirmed，但代入實機樣本對不上**
// （`docs/re/18` §4）：三個樣本裡兩個落在下限、一個在中段，
// 與均勻亂數的期望不符。可能是人口欄位的單位、`Random` 的呼叫次序，
// 或樣本 1 那次還有別的收入來源。**照反組譯實作，差異記在文件裡。**
func (w *AIWorld) Tax(p ProvinceID, rng *Rand) (TaxResult, error) {
	prov, err := w.Table.At(p)
	if err != nil {
		return TaxResult{}, err
	}
	pop := float64(prov.Population)

	gold := rng.Int(round(sqrt(pop)*TaxRandScale)) + TaxGoldBase
	if gold > TaxGoldMax {
		gold = TaxGoldMax
	}
	if gold < TaxGoldMin {
		gold = TaxGoldMin
	}
	food := rng.Int(round(pop/TaxFoodDivisor*TaxRandScale)) + TaxFoodBase

	prov.Gold = AddResource(prov.Gold, uint16(gold))
	prov.Food = AddResource(prov.Food, uint16(food))
	// 第一段：固定扣 30（`sub_303CD`）。
	if prov.Loyalty < TaxLoyaltyDrop {
		prov.Loyalty = 0
	} else {
		prov.Loyalty -= TaxLoyaltyDrop
	}
	// 第二段：再扣剩餘的五分之一（`sub_2C6C6` 尾段）。
	prov.Loyalty -= prov.Loyalty / TaxLoyaltySecondDivisor
	// 標記「本月已徵過稅」。
	prov.Flags |= ProvinceFlagTaxed
	return TaxResult{Gold: gold, Food: food, LoyaltyAfter: prov.Loyalty}, nil
}

// ---------------------------------------------------------------------------
// 忠誠度：全檔只有五個地方會改它（`docs/re/19`）。
// ---------------------------------------------------------------------------

// LoyaltyMax 是人民忠誠度的上限（`cmp byte ptr [di-6222h], 64h`）。
const LoyaltyMax = 100

// LoyaltyRestGain 是電腦內政「休養」一次回補的忠誠度
// （`sub_1A100` 的 `add byte ptr [di-6222h], 14h`），加完夾到 100。
//
// 它由 `sub_1ACCC` 在 `byte_6FE7E == 6` 那條內政分支呼叫
// （`docs/re/13` §6），**是電腦專屬的行為**——玩家那邊沒有對應的指令
// （慰勞軍民不改省的忠誠度，見 `docs/re/19` §3）。
const LoyaltyRestGain = 20

// RestoreLoyalty 是 `sub_1A100`：忠誠度 +20，夾到 100。
func (w *AIWorld) RestoreLoyalty(p ProvinceID) error {
	prov, err := w.Table.At(p)
	if err != nil {
		return err
	}
	v := int(prov.Loyalty) + LoyaltyRestGain
	if v > LoyaltyMax {
		v = LoyaltyMax
	}
	prov.Loyalty = uint8(v)
	return nil
}

// ---------------------------------------------------------------------------
// 開發（政略指令 7）的「建兵工廠」，出自 `sub_248A8`（`docs/re/20`）。
// ---------------------------------------------------------------------------

const (
	// ArsenalCost 是蓋一座兵工廠的黃金（`mov cx, 1388h`）。
	ArsenalCost = 5000
	// ArsenalMax 是一個省的兵工廠上限（`cmp byte ptr [di-6223h], 5`）。
	ArsenalMax = 5
)

// ArsenalAffordable 是這個省現在最多能蓋幾座，語意照 `sub_248A8`：
// 黃金不足 5,000 就一座都蓋不了，而且不能超過上限 5。
func (w *AIWorld) ArsenalAffordable(p ProvinceID) int {
	prov, err := w.Table.At(p)
	if err != nil {
		return 0
	}
	n := int(prov.Gold) / ArsenalCost
	if room := ArsenalMax - int(prov.Arsenals); n > room {
		n = room
	}
	if n < 0 {
		n = 0
	}
	return n
}

// BuildArsenal 蓋 n 座兵工廠，回傳實際蓋成幾座。
//
// 蓋好之後每座每月把煤鐵轉成彈藥——那條是實機量的
// （`docs/playtest/06` §3，湖北 3 座 → 煤 −6,000、鐵 −6,000、彈藥 +6,000），
// **公式還沒從反組譯確認**，所以不在這裡實作。
func (w *AIWorld) BuildArsenal(p ProvinceID, n int) (int, error) {
	prov, err := w.Table.At(p)
	if err != nil {
		return 0, err
	}
	if n <= 0 {
		return 0, nil
	}
	if max := w.ArsenalAffordable(p); n > max {
		return 0, fmt.Errorf("game: 省 %d 最多蓋 %d 座（黃金 %d、已有 %d 座），要求 %d",
			p, max, prov.Gold, prov.Arsenals, n)
	}
	prov.Gold -= uint16(n * ArsenalCost)
	prov.Arsenals += uint8(n)
	return n, nil
}
