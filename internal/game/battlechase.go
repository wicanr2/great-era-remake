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
			break
		}
	}
	return n
}
