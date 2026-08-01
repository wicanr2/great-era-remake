package game

// 沒有城市可打時的派工（`sub_3B9D8`，`docs/re/31` §52）。
//
// 值 3（打城市）挑不到目標時，原版走的是這一條：**改去追人**。
// 規則兩層——
//
//	1. 逐一試敵方單位，取第一個**走得通**的當目標（同 §24 的 sub_3BF6A）
//	2. ⭐ 但每個敵方單位有**火力額度**：吸引到的我方戰力達到它自己的
//	   三倍就滿了，後面的單位改試下一個敵人
//
// 第二層是整條戰鬥 AI 裡**唯一一處主動分散兵力**的規則。
// 其他行動（打城市、斬首）都是全軍撲同一個目標。

// AIChaseFirepowerRatio 是火力分配上限：**3 倍**（`sub_3B961`）。
//
//	已派去打它的我方戰力 >= 該敵方單位戰力 × 3  → 額度滿
const AIChaseFirepowerRatio = 3

// ChaseAssign 把等待指派的單位派去追敵方單位，回傳指派了幾個。
//
//	units   我方（原版是 side 選出來的那一份）
//	foes    敵方
//	route   尋路，回下一跳；走不到回 NoCell
//
// ⚠️ **兩處刻意沒實作**（§52 記了原因）：
//
//	sub_56A57 的貢獻     呼叫端用我方 index 寫、判斷用敵方 index 讀，
//	                     兩者不一致而且會被覆寫。看不懂就不假裝知道
//	byte[65BAh] 佔用表   只有 fallback 那條路會設，主路徑不設。
//	                     這個不對稱照抄沒有意義，先不做
func (s *BattleSim) ChaseAssign(units, foes []*Combatant,
	route func(to, from CellIndex) CellIndex) int {
	// 攻方是第一方（§32），收尾的兩道保護只對它生效。
	isFirstSide := len(units) > 0 && len(s.Attacker) > 0 &&
		units[0] != nil && s.Attacker[0] != nil && units[0].General == s.Attacker[0].General
	// 敵方的非零槽位壓成連續清單（原版的第一個迴圈）。
	var list []*Combatant
	for _, f := range foes {
		if f != nil && f.Alive() && f.Cell.Valid() {
			list = append(list, f)
		}
	}
	if len(list) == 0 {
		return 0
	}

	// 每個敵方單位吸引到的我方戰力（原版的 `火力[]`）。
	drawn := make([]int, len(list))
	// 各自的戰力（原版的 `強度[]`），只算一次。
	power := make([]int, len(list))
	for j, f := range list {
		power[j] = s.StrengthOf(f)
	}

	n := 0
	for _, u := range units {
		if u == nil || !u.Alive() {
			continue
		}
		// 同 §23／§24：只派「等待指派目標」的單位。
		if u.Command != BattleCmdSeekTarget && u.Command != BattleCmdCommitted {
			continue
		}
		if u.Assigned() {
			continue
		}
		assigned := false
		for j, f := range list {
			// 火力額度滿了就換下一個敵人——**這一步是重點**。
			if drawn[j] >= power[j]*AIChaseFirepowerRatio {
				continue
			}
			next := route(f.Cell, u.Cell)
			if next == NoCell {
				continue
			}
			u.AssignTo(f.General, next)
			drawn[j] += s.StrengthOf(u)
			n++
			assigned = true
			break
		}
		if !assigned {
			// 一個目標都走不到 → `sub_3B8B0`（§54）。
			FallbackToStandby(units, u, isFirstSide)
		}
	}
	return n
}

// AttackerNearCity 是 `sub_3A766`（§6）：**城市格上、或距離 2 以內，
// 站著攻方單位**。
//
// 分支 A 第四步用它挑值 18（挑最弱的圍城者）——問的是
// 「我的城市被人逼近了嗎」。
func (s *BattleSim) AttackerNearCity() bool {
	// 原版兩關：那格站的是攻方（`+8 != 0`），或 `sub_55CEC(2, ...)` > 0
	// ——距離 2 以內有敵人。`WithinTwoSteps(c, c)` 成立（原版沒排除），
	// 所以「站在城市上」被第二關涵蓋，不必分開判。
	for _, c := range CityCells(s.Field) {
		for _, u := range s.Attacker {
			if u != nil && u.Alive() && u.Cell.Valid() && WithinTwoSteps(u.Cell, c) {
				return true
			}
		}
	}
	return false
}

// AttackerFewerThanCities 是 `sub_56D13`（§53）：
// **攻方在場的單位數 < 戰場上的城市格數**。
//
//	sub_56D13() = sub_56CD0() < 城市數
//	              └ 掃第一方的 10 個槽位數非零的
//
// 分支 A 第四步用它在 14（只留一個守城）與 15（駐守的都留）之間選。
// 語意很順：**敵人少就大膽出擊，敵人多就保留守軍。**
func (s *BattleSim) AttackerFewerThanCities() bool {
	n := 0
	for _, u := range s.Attacker {
		if u != nil && u.Alive() {
			n++
		}
	}
	return n < len(CityCells(s.Field))
}

// AIStandbyCap 是分支 B（第一方／攻方）**同時待命的人數上限**：3。
//
// `sub_3B8B0` 掃 index 2..10 數命令 2 的單位，`> 2` 就不再把人打回待命
// ——所以待命數 0/1/2 時可以再加一個，最多 3 個。
const AIStandbyCap = 3

// FallbackToStandby 是 `sub_3B8B0`（`docs/re/31` §54）：**一個目標都走不到時的收尾**。
//
// 值 3 那條路（打城市／追人）與 `sub_3BF6A` 共用它。兩方的行為**不對稱**：
//
//	第二方（守方，side == 0）  直接打回待命，沒有任何保護
//	第一方（攻方，side != 0）  兩道保護：
//	                            1. 首位單位（主帥）免疫，永遠不打回
//	                            2. index 2..10 已經有 3 個以上待命就不再多一個
//
// ⚠️ 原版**只設 `+9` 與 `+10`，不動 `+12`**——與 `ResetToStandby`（`sub_3B15E`）
// 不同，那支會把 `+12` 清成 `0xFF`。兩支不能合併。
//
// 回傳是否真的把單位打回待命。
func FallbackToStandby(units []*Combatant, u *Combatant, isFirstSide bool) bool {
	if u == nil {
		return false
	}
	demote := func() bool {
		u.Command = BattleCmdStandby
		u.TargetUnit = 0
		return true
	}
	if !isFirstSide {
		return demote()
	}
	// 首位單位是主帥，不打回。
	if len(units) > 0 && units[0] != nil && units[0].General == u.General {
		return false
	}
	n := 0
	for i := 1; i < len(units); i++ { // 原版從 index 2 起，跳過首位
		if v := units[i]; v != nil && v.Command == BattleCmdStandby {
			n++
		}
	}
	if n >= AIStandbyCap {
		return false
	}
	return demote()
}
