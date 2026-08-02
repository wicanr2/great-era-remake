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

// RecruitSoldiersPerGold 是一塊黃金能徵幾個**步兵**。
//
// ⚠️ **只對步兵成立。** 實機那兩個樣本測的都是徵步兵，而
// `sub_18AED`（電腦側，`docs/re/32` §8）顯示四個兵種的匯率各不同：
// 步兵 ×10、騎兵 ×2、砲兵 ÷5、裝甲 ÷10。
// 其餘三個兵種請用 `airecruit.go` 的 `AIRecruitSoldiers`。
//
// 實機兩個樣本：
//
//	湖北 黃金 4,150 → 徵兵上限 (0-41,500)
//	河南 黃金 5,950 → 徵兵上限 (0-59,500)
//
// 加上輸入 1,000 兵時畫面顯示「共須黃金 100」——**10 兵 = 1 金**。
const RecruitSoldiersPerGold = 10

// RecruitAffordableLimit 回傳該兵種由黃金決定的徵兵上限。
//
// 四條玩家徵兵函式已逐條確認：步兵 `sub_260C7`、砲兵 `sub_271F8`、
// 裝甲兵 `sub_28259`、騎兵 `sub_29494`。玩家側匯率為：
// 步兵 10 人／金、騎兵 5 人／金、砲兵 1 人／金、裝甲兵 1 人／10 金。
// 注意電腦側騎兵是 2 人／金；這是原版兩條路徑的實際差異。
func RecruitAffordableLimit(branch uint8, gold int) int {
	switch branch {
	case BranchInfantry:
		return gold * 10
	case BranchCavalry:
		return gold * 5
	case BranchArtiller:
		return gold
	case BranchArmour:
		return gold / 10
	default:
		return 0
	}
}

// RecruitLimit 是這個省這次最多能徵幾個指定兵種的兵。
// 原版取「黃金可負擔量」與「該兵種所有可用部隊的缺額」兩者較小值。
func (w *AIWorld) RecruitLimit(p ProvinceID, branch uint8) int {
	prov, err := w.Table.At(p)
	if err != nil {
		return 0
	}
	full := int(BranchFullStrength(branch))
	if full == 0 {
		return 0
	}
	room := 0
	for _, i := range w.RosterOf(p).ids {
		if i >= len(w.Strengths) || w.Strengths[i].Branch != branch {
			continue
		}
		if gap := full - int(w.Strengths[i].Force); gap > 0 {
			room += gap
		}
	}
	affordable := RecruitAffordableLimit(branch, int(prov.Gold))
	if room < affordable {
		return room
	}
	return affordable
}

// RecruitCost 是徵 n 個指定兵種的兵要花多少黃金。
func RecruitCost(branch uint8, n int) int {
	switch branch {
	case BranchInfantry:
		return n / 10
	case BranchCavalry:
		return n / 5
	case BranchArtiller:
		return n
	case BranchArmour:
		return n * 10
	default:
		return 0
	}
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
	if limit := w.RecruitLimit(p, branch); n > limit {
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
		cost := RecruitCost(branch, added)
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
	if limit := w.RecruitLimit(p, branch); need > limit {
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

// ---------------------------------------------------------------------------
// 開發的另外兩項：墾地（`sub_241D0`）與挖金礦（`sub_24535`）。
// 見 `docs/re/21`。
// ---------------------------------------------------------------------------

const (
	// LandValueMax 是地價上限（`cmp byte ptr [di-6224h], 0C8h`）。
	LandValueMax = 200
	// ReclaimStaminaCost 是墾地消耗的體力（`sub byte ptr [di+7A9Ah], 5`）。
	ReclaimStaminaCost = 5
	// ReclaimStaminaNeed 是墾地的體力門檻（`cmp byte ptr [di+7A9Ah], 5`）。
	ReclaimStaminaNeed = 5
	// MineStaminaNeed 是挖金礦的體力門檻（`cmp byte ptr [di+7A9Ah], 14h`）
	// ——比墾地高四倍。
	MineStaminaNeed = 20
	// DevelopPoliticsDivisor 是開發的成效除數：效果 = 政治手腕 ÷ 10
	// （`mov cx, 0Ah / div cx`，兩項都用）。
	DevelopPoliticsDivisor = 10
)

// ReclaimLand 是墾地：派一個將領去，地價 += 政治手腕 ÷ 10，體力 −5。
//
// 地價夾在 200 以下。回傳實際增加的地價。
//
// `politics` 是將領記錄的 `+2`（政治手腕，`docs/playtest/08` §2），
// `stamina` 是 `+29`（體力）。**兩個都由呼叫端傳進來**——規則層的
// 將領狀態還沒有統一的容器，等 `+16` 的語意解出來再收攏。
func (w *AIWorld) ReclaimLand(p ProvinceID, politics, stamina uint8) (int, uint8, error) {
	prov, err := w.Table.At(p)
	if err != nil {
		return 0, stamina, err
	}
	if stamina < ReclaimStaminaNeed {
		return 0, stamina, fmt.Errorf("game: 體力 %d 不足 %d，不能墾地",
			stamina, ReclaimStaminaNeed)
	}
	gain := int(politics) / DevelopPoliticsDivisor
	v := int(prov.LandValue) + gain
	if v > LandValueMax {
		v = LandValueMax
	}
	actual := v - int(prov.LandValue)
	prov.LandValue = uint8(v)
	return actual, stamina - ReclaimStaminaCost, nil
}

const (
	// MineStaminaCost 是挖金礦消耗的體力（`sub byte ptr [di+7A9Ah], 5`）。
	// **門檻是 20 但只扣 5**，兩個數字不同，不要混用。
	MineStaminaCost = 5
	// MineYieldScale / MineYieldBase 是挖金礦的產出係數：
	//
	//	黃金 += Random(礦藏 + 政治手腕÷10) × 50 + 200
	//
	// `Random(n)` 回 `[0, n)`，所以產出範圍是 `[200, 200 + 50×(n−1)]`——
	// **保底 200**，運氣差也不會空手而回。
	MineYieldScale = 50
	MineYieldBase  = 200
)

// MineGold 是挖金礦（開發指令的第二項，`sub_24535`）。
//
// `deposit` 是那張以省編號為索引的 byte 表的值（`ds:11h` 起，stride 1）。
// ⚠️ **那張表的語意未解**——位置與形狀對得上「各省礦藏量」，但沒有直接證據，
// 所以參數名只是暫用，不要當成已確認的欄位。
//
// `politics` 是將領 `+2`（政治手腕），`stamina` 是 `+29`（體力）。
// 回傳實際入帳的黃金與扣完的體力。
func (w *AIWorld) MineGold(p ProvinceID, deposit, politics, stamina uint8, rng *Rand) (int, uint8, error) {
	prov, err := w.Table.At(p)
	if err != nil {
		return 0, stamina, err
	}
	if stamina < MineStaminaNeed {
		return 0, stamina, fmt.Errorf("game: 體力 %d 不足 %d，不能挖礦",
			stamina, MineStaminaNeed)
	}
	n := int(deposit) + int(politics)/DevelopPoliticsDivisor
	yield := int(rng.Int(n))*MineYieldScale + MineYieldBase

	before := prov.Gold
	prov.Gold = AddResource(prov.Gold, uint16(yield))
	return int(prov.Gold - before), stamina - MineStaminaCost, nil
}

// ---------------------------------------------------------------------------
// 慰勞軍民（政略指令 14），出自 `sub_3412B`（`docs/re/22`）。
// ---------------------------------------------------------------------------

const (
	// ComfortStaminaGain 是慰勞回補的體力（`add byte ptr es:[di+1Dh], 0Ah`）。
	ComfortStaminaGain = 10
	// StaminaMax 是體力上限。
	StaminaMax = 100
	// MoraleMax 是士氣上限——**80 不是 100**（`cmp es:[di+1Eh], 50h`）。
	//
	// 與實機吻合：`docs/playtest/08` 的五個將領士氣是
	// 74／67／42／54／30，全部低於 80。
	MoraleMax = 80
	// MoraleAbilityDivisor 是士氣回補的除數：帶兵能力 ÷ 5。
	MoraleAbilityDivisor = 5
	// ComfortLoyaltyGain 是慰勞回補的將領忠誠度（`add byte ptr es:[di+1], 14h`）。
	ComfortLoyaltyGain = 20
	// GeneralLoyaltyMax 是將領忠誠度上限。
	GeneralLoyaltyMax = 100
)

// ComfortResult 記錄慰勞軍民影響了誰。
type ComfortResult struct {
	// Generals 是被回補的將領數（`Units` 索引）。
	Generals []int
	// ProvinceLoyaltyGain 是省的人民忠誠度增量。
	ProvinceLoyaltyGain int
}

// Comfort 是慰勞軍民，語意照 `sub_3412B`：
//
//	該省每個「可用 + 同勢力」的將領：
//	    體力 += 10          夾到 100
//	    士氣 < 80 時 += 帶兵能力 ÷ 5   夾到 80
//	    忠誠度 += 20        夾到 100
//	省的人民忠誠度 += 領袖的政治手腕 ÷ 10   夾到 100
//
// `leaderPolitics` 是勢力領袖的 `+2`（政治手腕）——原版拿的是
// `arg_0`（勢力 ID）那個將領的欄位，不是執行慰勞的人。
//
// ⚠️ 規則層的將領狀態（體力／士氣／忠誠度）還沒收攏進一個容器，
// 所以這裡回報「哪些將領被影響」，實際的欄位更新由呼叫端做。
// 等 `+16` 的語意解出來再一起重整。
func (w *AIWorld) Comfort(p ProvinceID, leaderPolitics uint8) (ComfortResult, error) {
	prov, err := w.Table.At(p)
	if err != nil {
		return ComfortResult{}, err
	}
	var res ComfortResult
	res.Generals = append(res.Generals, w.RosterOf(p).ids...)

	gain := int(leaderPolitics) / DevelopPoliticsDivisor
	v := int(prov.Loyalty) + gain
	if v > LoyaltyMax {
		v = LoyaltyMax
	}
	res.ProvinceLoyaltyGain = v - int(prov.Loyalty)
	prov.Loyalty = uint8(v)
	return res, nil
}

// ComfortGeneral 算一個將領被慰勞之後的體力／士氣／忠誠度。
//
// 分開寫是因為原版的三格在 `MAN` 記錄裡，而規則層目前把它們
// 散在 `StrengthInput`（`F29`／`F30`）與 `General`（`AbilityB`）。
func ComfortGeneral(stamina, morale, loyalty, ability uint8) (uint8, uint8, uint8) {
	if s := int(stamina) + ComfortStaminaGain; s > StaminaMax {
		stamina = StaminaMax
	} else {
		stamina = uint8(s)
	}
	// 士氣只有**未達上限**時才加——原版先檢查再算。
	if morale < MoraleMax {
		if m := int(morale) + int(ability)/MoraleAbilityDivisor; m > MoraleMax {
			morale = MoraleMax
		} else {
			morale = uint8(m)
		}
	}
	if l := int(loyalty) + ComfortLoyaltyGain; l > GeneralLoyaltyMax {
		loyalty = GeneralLoyaltyMax
	} else {
		loyalty = uint8(l)
	}
	return stamina, morale, loyalty
}

// ApplyComfortGenerals 套用慰勞的三項將領效果，並同步 AIWorld 使用的
// StrengthInput 鏡像。AbilityA 是帶兵能力；AbilityB 是將領忠誠度。
// 回傳實際更新的有效索引數；越界索引會被忽略。
func ApplyComfortGenerals(gens []General, strengths []StrengthInput, indices []int) int {
	n := 0
	for _, i := range indices {
		if i < 0 || i >= len(gens) {
			continue
		}
		g := &gens[i]
		g.Stamina, g.F30, g.AbilityB = ComfortGeneral(
			g.Stamina, g.F30, g.AbilityB, g.AbilityA)
		if i < len(strengths) {
			strengths[i].F29 = g.Stamina
			strengths[i].F30 = g.F30
		}
		n++
	}
	return n
}
