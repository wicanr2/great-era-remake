package game

import "fmt"

// 外交（政略指令 9）的另外兩項：請求外援與償還外債。
// 貸款在 `loan.go`。
//
//	外援      `sub_21D1D`（738 行）
//	償還外債  `sub_223ED`（556 行）

// CreditMax 是信用度的上限（`cmp byte ptr [di-4225h], 64h`）。
//
// 償還外債會加信用度，加完夾到 100。與貸款的扣是**完全對稱**的：
//
//	貸款      信用度 −= 額度 ÷ 500
//	償還外債  信用度 += 額度 ÷ 500，夾到 100
//
// 兩邊用同一個 `LoanUnit`（500），所以借多少還多少就回到原點。
const CreditMax = 100

// 外援的成功判定：`Random(10) ≤ 6` → **70%**。
const (
	AidRollRange  = 10
	AidSuccessMax = 6
)

// AidGoldCap 是外援黃金的上限（`cmp [bp+var_4], 1770h` → 6000）。
//
// 只有黃金被夾——糧食／彈藥／燃料那三格沒有對應的檢查。
const AidGoldCap = 6000

// 援助國的慷慨程度。原版用一個代碼（`var_14`）分流出**除數**，
// 援助量是 `Random(基準) ÷ 除數`，所以**除數越小給越多**。
//
//	代碼 99 (63h) 或 142 (8Eh) → 除數 1   最慷慨
//	代碼 100 (64h)             → 除數 3
//	其他（101、141…）          → 除數 5
//
// 那些代碼看起來是 `2.15` 的詞條索引（1-based 99／100／101 依序是
// 美國／英國／俄國，`50-diplomacy.md` §3）。若真是如此，
// **美國最慷慨**——與 `sub_37A81`「美國援助軍事物資中」那個事件呼應。
//
// ⚠️ **代碼與國家的對應標為假說**：數值對得上詞條位置，但沒有直接證據，
// 而且 141／142 對不上「法國」的詞條編號。
const (
	AidDonorGenerous = 99  // 除數 1
	AidDonorMedium   = 100 // 除數 3
	AidDonorGenerou2 = 142 // 除數 1
)

// AidDivisor 回傳某個援助國代碼對應的除數。
func AidDivisor(donor int) int {
	switch donor {
	case AidDonorGenerous, AidDonorGenerou2:
		return 1
	case AidDonorMedium:
		return 3
	}
	return 5
}

// 外援的史實特例：**張作霖在民國 17 年 2–6 月一律拿不到援助**。
//
//	cmp  word ptr [di-6221h], 0A6h   ; 省份司令 == 166（張作霖）
//	cmp  byte_6FE7D, 11h             ; 年 == 17
//	cmp  byte_6FE7E, 2 / 6           ; 月在 2..6
//	→ 直接跳到拒絕
//
// 民國 17 年 = 1928，北伐軍進逼、列強不再支持奉系，皇姑屯事件就在那年 6 月
// （`docs/re/25` §5 的東北易幟是同一年的事）。**這條是寫死的劇本**，
// 不是通用規則。
const (
	AidEmbargoLeader     = 166
	AidEmbargoYear       = 17
	AidEmbargoMonthFirst = 2
	AidEmbargoMonthLast  = 6
)

// AidResult 記錄一次外援請求的結果。
type AidResult struct {
	Approved bool
	Roll     int
	// Donor 是援助國代碼，Divisor 是它對應的除數。
	Donor, Divisor int
	// 四種資源實際入帳的量。
	Gold, Food, Ammo, Fuel int
	// Embargoed 表示被史實特例擋下（張作霖那條）。
	Embargoed bool
}

// RequestAid 向列強請求援助（`sub_21D1D`）。
//
// `base` 是四種資源各自的亂數基準——原版從 `var_4`／`var_8`／`var_C`／`var_10`
// 取，**那四個值怎麼來的還沒讀**（在函式的 200–360 行之間），
// 所以由呼叫端傳進來。
//
// 流程：
//
//	Random(10) ≤ 6 → 同意（70%）
//	張作霖 + 民國 17 年 2–6 月 → 一律拒絕
//	同意：四種資源各 += Random(base[i]) ÷ 除數，黃金夾 6000
func (w *AIWorld) RequestAid(p ProvinceID, st GameState, donor int,
	base [4]int, rng *Rand) (AidResult, error) {
	prov, err := w.Table.At(p)
	if err != nil {
		return AidResult{}, err
	}

	res := AidResult{Donor: donor, Divisor: AidDivisor(donor)}
	res.Roll = rng.Int(AidRollRange)
	if res.Roll > AidSuccessMax {
		return res, nil
	}

	// 史實特例：張作霖在民國 17 年 2–6 月拿不到援助。
	if prov.Commander == AidEmbargoLeader && st.Year == AidEmbargoYear &&
		st.Month >= AidEmbargoMonthFirst && st.Month <= AidEmbargoMonthLast {
		res.Embargoed = true
		return res, nil
	}

	res.Approved = true
	give := func(field *uint16, b int, cap int) int {
		if b <= 0 {
			return 0
		}
		v := rng.Int(b) / res.Divisor
		if cap > 0 && v > cap {
			v = cap
		}
		before := *field
		*field = AddResource(*field, uint16(v))
		return int(*field - before)
	}
	res.Gold = give(&prov.Gold, base[0], AidGoldCap)
	res.Food = give(&prov.Food, base[1], 0)
	res.Ammo = give(&prov.Ammo, base[2], 0)
	res.Fuel = give(&prov.Fuel, base[3], 0)
	return res, nil
}

// RepayDebt 償還外債（`sub_223ED`）。
//
// 與貸款完全對稱：每 500 黃金加 1 點信用度，加完夾到 100。
//
//	額度 ÷ 500 → dl
//	信用度 += dl
//	若 信用度 > 100 → 100
//
// ⚠️ **「外債」本身的位址還沒找到**（省份記錄裡沒有，是全域狀態），
// 所以這裡只處理「付錢 + 加信用度」，**沒有扣減外債餘額**。
// 原版一定有那一步，只是還沒讀到。
func (w *AIWorld) RepayDebt(p ProvinceID, amount int, credit uint8) (int, uint8, error) {
	prov, err := w.Table.At(p)
	if err != nil {
		return 0, credit, err
	}
	if amount <= 0 {
		return 0, credit, fmt.Errorf("game: 償還額要為正（%d）", amount)
	}
	if int(prov.Gold) < amount {
		return 0, credit, fmt.Errorf("game: 黃金 %d 不足 %d", prov.Gold, amount)
	}
	prov.Gold -= uint16(amount)

	units := amount / LoanUnit
	v := int(credit) + units
	if v > CreditMax {
		v = CreditMax
	}
	return units, uint8(v), nil
}
