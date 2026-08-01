package game

// 戰場上的**戰力比判定**（`sub_3A320`）與**首位單位優勢等級**（`sub_58D4A`）。
//
// 這兩支是戰鬥 AI 的共用度量衡：戰鬥 AI 幾乎每個判斷都先問一句
// 「現在誰占上風、差多少」，答案就出自這裡。
//
// 兩者都建立在攻擊力公式 `sub_5A0B9`（`strength.go`）之上——
// 那支函式在整份程式裡出現六次，是**電腦唯一的評價函式**（`docs/re/29`）。

// 戰場上的兩個單位陣列（各 10 格，index 1..10）：
//
//	word[750h + i×2]  第一方
//	word[764h + i×2]  第二方
//
// 誰是攻方誰是守方由 `+8` 決定（`docs/spec/02`），不是由陣列決定。

// ForceRatioLE 回答 `sub_3A320(side, ratio)`：
//
//	我方戰力總和 × ratio ≤ 敵方戰力總和
//
// 成立回 true（原版回 1）。`side` 只決定哪一邊算「我方」，
// 在這裡由呼叫端自己把 `mine`／`foe` 擺對。
//
// ⚠️ **倍率用分數表示**。原版是 Turbo Pascal 的 48-bit Real 乘法，
// 而 0.67 那個常數的實際值是 0.670000000000073（Real 存不下 0.67）。
// 用 `67/100` 是精確的 0.67，兩者在**剛好相等**的邊界上會分歧。
// 這是已知偏差，不是 bug——要完全一致得把 48-bit Real 的乘法也照抄。
func ForceRatioLE(mine, foe, num, den int) bool {
	// mine × (num/den) ≤ foe  ⟺  mine × num ≤ foe × den
	return mine*num <= foe*den
}

// 六個呼叫端的倍率（全部從 48-bit Real 常數解出來，`docs/re/31` §16）。
//
// 讀法一律是「我方 × 倍率 ≤ 敵方」成立才回 true，
// 所以**倍率越大，越是在問「我方是不是被輾壓」**。
const (
	// AIBattleRatioCollapseNum/Den 是 **5.0**：我方 × 5 ≤ 敵方，
	// 即**我方被壓到敵方的五分之一以下**。這是**必勝門檻**。
	//
	// 兩個呼叫端對稱（`sub_3A817` 看第二方、`sub_3AA51` 看第一方），
	// 成立之後設的決策值都導向 `sub_3B19C` → `sub_54E3B`，
	// 也就是**九步結算**（`docs/mechanics/30-combat.md` §3）。
	//
	// 讀作：**一方戰力不到對方的五分之一，電腦不打了，直接判勝負。**
	AIBattleRatioCollapseNum = 5
	AIBattleRatioCollapseDen = 1

	// AIBattleRatioEvenNum/Den 是 **0.67**（`sub_3A8F7`）：
	// 我方 × 0.67 ≤ 敵方。呼叫端要的是**不成立**的那邊，
	// 也就是「敵方 < 我方的三分之二」。
	AIBattleRatioEvenNum = 67
	AIBattleRatioEvenDen = 100

	// AIBattleRatioAheadNum/Den 是 **0.5**（`sub_3C26A`）：
	// 我方 × 0.5 ≤ 敵方。呼叫端要的同樣是**不成立**：
	// 敵方 < 我方的一半，也就是**我方大幅領先**。
	AIBattleRatioAheadNum = 1
	AIBattleRatioAheadDen = 2

	// AIBattleRatioBehindNum/Den 是 **1.5**（`sub_3C89A`／`sub_3CA09`）：
	// 我方 × 1.5 ≤ 敵方，即**敵方是我方的 1.5 倍以上**。
	AIBattleRatioBehindNum = 3
	AIBattleRatioBehindDen = 2
)

// ── 首位單位優勢等級（`sub_58D4A`）──────────────────────────────────

// 三個等級。
const (
	// AIAdvantageBehind 是首位單位戰力 < 敵軍總戰力。
	AIAdvantageBehind = 0
	// AIAdvantageAhead 是首位單位戰力 ≥ 敵軍總戰力。
	AIAdvantageAhead = 1
	// AIAdvantageOverwhelming 是首位單位戰力 ≥ 敵軍總戰力 × 2。
	AIAdvantageOverwhelming = 2
)

// AdvantageLevel 是 `sub_58D4A`：拿**首位單位一個人**的攻擊力，
// 跟**敵方全部單位**的攻擊力總和比。
//
// ⚠️ 「首位單位」是 `word_64902`，即第一方單位陣列的第 1 格。
// 把它當成「主將」是**推論**（`docs/re/31` §20）：`sub_3C26A` 的兩個
// 計數迴圈都從 index 2 起、第 1 格單獨處理，形狀對得上而已。
//
//	敵方總和 × 2 ≤ 首位單位   → 2（壓倒）
//	敵方總和     ≤ 首位單位   → 1（優勢）
//	否則                  → 0（劣勢）
//
// ⚠️ 一人對全軍，看起來不公平，但首位單位的戰力通常是全軍最高
// （攻擊力公式吃能力值與兵力，`strength.go`）。這是原版的寫法，照抄。
//
// 原版的 `×2` 走 `Round(Real(敵方總和 × 2))`；因為乘的是整數 2，
// `Round` 不會改變結果，這裡直接用整數。
func AdvantageLevel(firstUnitStrength, foeTotal int) int {
	if foeTotal*2 <= firstUnitStrength {
		return AIAdvantageOverwhelming
	}
	if foeTotal <= firstUnitStrength {
		return AIAdvantageAhead
	}
	return AIAdvantageBehind
}

// ── 預備隊的投入（`sub_3C26A` 前半）──────────────────────────────────

// AIReserveKeep 回傳**要留幾個單位待命**，由首位單位的優勢等級決定：
//
//	劣勢(0) → 留 2      優勢(1) → 留 1      壓倒(2) → 留 0
//
// 越占上風，預備隊留得越少——壓倒性優勢時全部投入。
//
// 原版沒有這張表，是三道 `cmp` 的**穩定點**：迴圈倒著掃，
// 每把一個待命單位轉成命令 4 就 `dec` 計數，掃完之後計數必然停在這個值。
//
//	if 待命數 > 2                     → 轉
//	if 待命數 == 2 且 優勢等級 ≥ 1     → 轉
//	if 待命數 == 1 且 優勢等級 == 2    → 轉
func AIReserveKeep(advantage int) int {
	switch {
	case advantage >= AIAdvantageOverwhelming:
		return 0
	case advantage >= AIAdvantageAhead:
		return 1
	default:
		return 2
	}
}

// ActivateReserves 把待命（命令 2）的單位轉成命令 4，直到剩下
// `AIReserveKeep(advantage)` 個為止（`sub_3C26A` 的第三個迴圈）。
//
// `cmds` 是 index 1..10 的命令欄位（`+9`），index 0 不用。
//
// ⚠️ **倒著掃**（10 → 2），而且**從 index 2 起**——index 1 不碰。
// 原版兩個迴圈的邊界都是 `[2, 10]`，index 1 在別處另外處理
// （`sub_3C26A` 後半只動 `word_64902` 指到的首位單位）。
func ActivateReserves(cmds []uint8, advantage int) {
	standby := 0
	for i := 2; i <= 10 && i < len(cmds); i++ {
		if cmds[i] == BattleCmdStandby {
			standby++
		}
	}
	keep := AIReserveKeep(advantage)
	for i := 10; i >= 2; i-- {
		if i >= len(cmds) {
			continue
		}
		if cmds[i] != BattleCmdStandby || standby <= keep {
			continue
		}
		cmds[i] = BattleCmdCommitted
		standby--
	}
}

// ── 命令類型（`+9`）─────────────────────────────────────────────────
//
// 語意來自 `docs/re/31` §15／§16。**4 與 5 仍未完全確定**。

const (
	// BattleCmdGarrison 是 **1：前往城市駐守**（首位單位專屬）。
	// `sub_47EAA` 只在 `u == word_64902` 時設它，且已經站在城市上就不設；
	// `sub_3D0AC` 看到「命令 1 且站在城市上」就交接並改成 6。
	BattleCmdGarrison uint8 = 1
	// BattleCmdStandby 是 **2：待命**。不需要目標格。
	BattleCmdStandby uint8 = 2
	// BattleCmdSeekTarget 是 **3：需要（重新）找目標**。
	// 沒有目標格時 `sub_47EAA` 設它；`sub_3DED9` 重算路徑後仍找不到就退回 2。
	BattleCmdSeekTarget uint8 = 3
	// BattleCmdCommitted 是 **4：預備隊被投入**。
	//
	// ⚠️ 語意是**推論**。已知的是：待命(2) 的單位在優勢時會被轉成 4
	// （`ActivateReserves`），而首位單位在優勢等級 1 時也設 4、等級 2 時設 3。
	// 從「壓倒性優勢才設 3」推得 **4 比 3 保守**。
	BattleCmdCommitted uint8 = 4
	// BattleCmdUnknown5 是 **5：語意未知**。
	//
	// 值域確定（`sub_3DED9` 與 2／4 同批處理），但**整份程式沒有任何地方
	// 寫立即數 5 進 `+9`**。它可能來自存檔、或來自某個計算結果。
	BattleCmdUnknown5 uint8 = 5
	// BattleCmdGarrisoned 是 **6：駐守完成**（終態）。
	// `sub_3D0AC` 交接之後寫它，而**沒有任何地方比較它**——寫進去就不再變。
	BattleCmdGarrisoned uint8 = 6
)
