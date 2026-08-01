package game

import "fmt"

// 向列強貸款（政略指令 9「外交」的一項），出自 `sub_2164A`（726 行）。
//
// 畫面上的話是「無法貸款／目前信用度〈N〉／司令欲貸款多少？／
// 同意貸款／各國均拒絕提供貸款」（`docs/re/27`）。

// LoanUnit 是貸款的計價單位：**每 500 黃金算一「筆」**。
//
//	mov  cx, 1F4h            ; 500
//	call @$brdiv$q7Longintt1 ; 額度 ÷ 500
//
// 這個商同時決定兩件事：擲骰的加成，以及信用度要扣多少。
const LoanUnit = 500

// LoanRollRange / LoanRejectAbove 是核貸的判定：
//
//	var_7 = Random(10) + 額度 ÷ 500
//	若 var_7 > 12 → 各國均拒絕
//
// 讀出來的行為很直觀——**借越多越容易被拒**：
//
//	借 ≤ 500      加成 0 或 1，最大 10  → 幾乎必過
//	借 6,000      加成 12，只有骰到 0 才過
//	借 ≥ 6,500    加成 ≥ 13，**必被拒**
//
// 所以單次貸款的硬上限是 6,499（實務上 6,000 已經只有 1/10 機會）。
const (
	LoanRollRange   = 10
	LoanRejectAbove = 12
)

// CreditTableAddr 是「信用度」表在 `WAR.EXE` 資料段的位址。
//
//	sub  [di-4225h], dl      ; di = 索引，dl = 額度 ÷ 500
//
// `-4225h` 當 16-bit 是 `0xBDDB`。索引來自 `ss:[di-6]`——
// ⚠️ **那個索引是什麼還沒解**（形狀像列強編號 0–3，但沒有證據）。
//
// 記著這個位址是為了以後動態驗證：`docs/playtest/13` 的實機流程
// 可以開貸款畫面看「目前信用度」的顯示值，與這張表對照。
const CreditTableAddr = 0xBDDB

// LoanResult 記錄一次貸款申請的結果。
type LoanResult struct {
	// Approved 是核准與否。
	Approved bool
	// Roll 是擲出的亂數，Units 是 `額度 ÷ 500`。
	Roll  int
	Units int
	// Amount 是實際入帳的黃金（被拒時為 0）。
	Amount int
	// CreditCost 是信用度被扣掉的量（＝ Units）。
	CreditCost int
}

// RequestLoan 向列強借 amount 黃金。
//
// 判定與效果照 `sub_2164A`：
//
//	units = amount ÷ 500
//	roll  = Random(10)
//	若 roll + units > 12 → 拒絕，什麼都不做
//	否則  信用度 −= units、省份黃金 += amount
//
// `credit` 是當前的信用度，回傳扣完的值。
//
// ⚠️ **原版還有兩段這裡沒實作的邏輯**：
//
//   - 開頭用司令 ID 查一張 byte 表（`ds:ACE7h` 起），值等於 10 時
//     把 `var_6` 設成 99／100；分期不是 1 時設 99。
//     `var_6` **在核貸判定裡沒有被讀**，看起來是拿去顯示的
//     （可能就是畫面上的「目前信用度」），但沒追到用它的地方。
//   - 信用度本身會不會擋住貸款——畫面有「無法貸款」的訊息，
//     但那條分支的條件還沒對上。
//
// 這兩段沒解之前，**本函式只實作核貸判定與入帳**。
func (w *AIWorld) RequestLoan(p ProvinceID, amount int, credit uint8, rng *Rand) (LoanResult, uint8, error) {
	prov, err := w.Table.At(p)
	if err != nil {
		return LoanResult{}, credit, err
	}
	if amount <= 0 {
		return LoanResult{}, credit, fmt.Errorf("game: 貸款額要為正（%d）", amount)
	}

	res := LoanResult{Units: amount / LoanUnit}
	res.Roll = rng.Int(LoanRollRange)
	if res.Roll+res.Units > LoanRejectAbove {
		return res, credit, nil // 各國均拒絕提供貸款
	}

	res.Approved = true
	res.Amount = amount
	res.CreditCost = res.Units
	if int(credit) < res.CreditCost {
		credit = 0
	} else {
		credit -= uint8(res.CreditCost)
	}
	before := prov.Gold
	prov.Gold = AddResource(prov.Gold, uint16(amount))
	res.Amount = int(prov.Gold - before)
	return res, credit, nil
}

// LoanAlwaysApproved 回報某個額度是不是「一定會過」。
//
// `Random(10)` 最大是 9，所以 `9 + units ≤ 12` 時必過——也就是 units ≤ 3，
// 額度 ≤ 1,999。
func LoanAlwaysApproved(amount int) bool {
	return amount > 0 && (LoanRollRange-1)+amount/LoanUnit <= LoanRejectAbove
}

// LoanAlwaysRejected 回報某個額度是不是「一定會被拒」。
//
// `Random(10)` 最小是 0，所以 `0 + units > 12` 時必拒——units ≥ 13，
// 額度 ≥ 6,500。
func LoanAlwaysRejected(amount int) bool {
	return amount/LoanUnit > LoanRejectAbove
}
