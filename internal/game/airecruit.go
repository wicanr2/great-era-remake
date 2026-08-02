package game

// 電腦徵兵的**各兵種匯率**（`sub_18AED`，`docs/re/32` §8）。
//
// `RecruitSoldiersPerGold = 10`（`recruit.go`）是從實機觀察來的，
// 而實機只測過**徵步兵**（湖北黃金 4,150 → 上限 41,500）。
// `sub_18AED` 顯示四個兵種的匯率**各不相同**：
//
//	兵種 1 步兵   1 金 → 10 兵     徵滿 20,000 要 2,000 金
//	兵種 6 騎兵   1 金 →  2 兵     徵滿 10,000 要 5,000 金
//	兵種 4 砲兵   5 金 →  1 兵     徵滿  2,000 要 10,000 金
//	兵種 5 裝甲  10 金 →  1 兵     徵滿    200 要 2,000 金
//
// ⚠️ 這是**電腦側**（`sub_18AED`）的證據。玩家側走 `sub_28259`，
// 兩者的匯率是否相同**未驗**——只有步兵那一項在兩邊都對得上
// （實機 ×10 ＝ `sub_18AED` 的兵種 1）。

// aiRecruitPerGold 是「一塊黃金買幾個兵」，用分子分母表示，
// 因為砲兵與裝甲是**不到一個兵**。
//
//	兵數 = 黃金 × num / den
var aiRecruitPerGold = map[uint8][2]int{
	BranchInfantry: {10, 1}, // ×10
	BranchCavalry:  {2, 1},  // ×2
	BranchArtiller: {1, 5},  // ÷5
	BranchArmour:   {1, 10}, // ÷10
}

// AIRecruitSoldiers 回傳「花這些黃金能徵幾個某兵種的兵」（`sub_18AED`）。
//
// 未知兵種回 0——原版對這四個以外的兵種**什麼也不做**
// （四個 `cmp` 都不中就直接跳到出口）。
func AIRecruitSoldiers(branch uint8, gold int) int {
	r, ok := aiRecruitPerGold[branch]
	if !ok {
		return 0
	}
	return gold * r[0] / r[1]
}

// AIRecruitGoldFor 回傳「徵這麼多兵要花多少黃金」，是上面的反函數。
func AIRecruitGoldFor(branch uint8, soldiers int) int {
	r, ok := aiRecruitPerGold[branch]
	if !ok {
		return 0
	}
	return soldiers * r[1] / r[0]
}

// AIRecruitPerCall 是**單次徵兵的黃金上限：2,000**。
//
//	cmp     word ptr es:[di], 7D0h
//	jb      ...
//	mov     word ptr es:[di], 7D0h      ; 夾到 2,000
//	mov     [bp+var_9], 1               ; 記下「夾過」
//	...
//	if 夾過: *預算 = 原值 − 2,000 + 剩餘   ; ← 把沒用完的還回去
//
// 所以一次呼叫最多動用 2,000 金，**剩下的還給預算**讓下一個部隊用。
// 2,000 金剛好是徵滿一個步兵師（20,000 ÷ 10）。
const AIRecruitPerCall = 2000

// AIRecruitInto 把黃金投進一個部隊，回傳（實際補的兵、實際花的金）。
//
// 照 `sub_18AED` 的順序：
//
//  1. 預算夾到 2,000（超過的部分記下來，最後還回去）
//  2. 依兵種匯率算出兵數，加到 `+17`，**夾滿員數**（`sub_5A467`）
//  3. 用「實際加了多少兵」反算花費，從預算扣掉
//
// 第 3 步是關鍵：夾到滿員之後**只扣實際用掉的錢**，不是扣整筆預算。
func AIRecruitInto(branch uint8, force uint16, budget int) (added, spent int) {
	full := int(BranchFullStrength(branch))
	if full == 0 {
		return 0, 0
	}
	use := budget
	if use > AIRecruitPerCall {
		use = AIRecruitPerCall
	}
	want := AIRecruitSoldiers(branch, use)
	after := int(force) + want
	if after > full {
		after = full
	}
	added = after - int(force)
	if added < 0 {
		added = 0
	}
	return added, AIRecruitGoldFor(branch, added)
}

// ── 徵新兵的素質代價（`sub_18F5B`，`docs/re/32` §9）─────────────────────

// AIRecruitSkillDiv / AIRecruitF20Div 是徵新兵之後兩個素質欄位的扣減除數。
//
//	士兵戰技（+19）−= 戰技 ÷ 5      **−20%**
//	+20            −= +20  ÷ 4      **−25%**，而且要 +20 ≥ 20 才扣
//
// **新兵拉低整體素質**——這是原版對「一次補一大批人」的制衡。
const (
	AIRecruitSkillDiv = 5
	AIRecruitF20Div   = 4
	// AIRecruitF20Min 是 `+20` 的扣減門檻：低於 20 就不再往下扣。
	AIRecruitF20Min = 20
)

// AIRecruitQualityCost 回傳補完兵之後的（士兵戰技, F20）。
//
// ⚠️ **砲兵與裝甲不受影響**（原版兩道扣減都排除兵種 4 與 5）。
// 那兩個兵種的徵兵單價高得多（徵滿 10,000／2,000 金），不掉素質是配套。
func AIRecruitQualityCost(branch, skill, f20 uint8) (uint8, uint8) {
	if branch == BranchArtiller || branch == BranchArmour {
		return skill, f20
	}
	if f20 >= AIRecruitF20Min {
		f20 -= f20 / AIRecruitF20Div
	}
	skill -= skill / AIRecruitSkillDiv
	return skill, f20
}

// AIRecruitSplitMax 是第三支分配策略（`sub_191A6`）的部隊數上限：**5**。
//
// 那一支先數有幾個沒滿員的部隊（數到 5 就不再數），再把預算**平分**。
// 硬編碼在 `cmp [bp+var_6], 5 / jge`。
const AIRecruitSplitMax = 5

// AIRecruitSplitShare 回傳「平分給 n 個部隊時每份多少」（`sub_191A6` 第二階段）。
//
// `n` 會先夾到 `AIRecruitSplitMax`。n == 0 回 0（原版直接 return）。
//
// 原版用 `Round(Real(預算) / Real(n))`，所以是四捨五入不是截斷。
func AIRecruitSplitShare(budget, n int) int {
	if n <= 0 {
		return 0
	}
	if n > AIRecruitSplitMax {
		n = AIRecruitSplitMax
	}
	q, r := budget/n, budget%n
	if r*2 >= n {
		q++
	}
	return q
}
