package game

// 「被敵人完全包圍」的判定（`sub_5A9F6`）與它底下的敵對鄰省收集（`sub_5A907`）。
//
// 決策鏈 A 的**步驟 4 與步驟 6** 都用它：
//
//	步驟 6：`sub_5A9F6(當前省)` 為真 → 走 `sub_17236` → **`sub_174C9` 出兵**
//	步驟 4：`sub_170C0` 在鄰省裡找第一個滿足它的省
//
// 讀作：**被包圍就突圍**。

// EnemyNeighbours 收集某省的敵對鄰省（`sub_5A907`）。
//
// 篩選條件三個，缺一不可：
//
//	鄰省編號 != 0 且 != 255（SeaBorder）   ← 有效的鄰省
//	該省的司令 != 0                        ← **無主省不算敵人**
//	該省的司令 != faction                  ← 不是自己人
//
// [雷] **無主省不算敵對**。這與攻打候選的規則正好相反——
// 攻打要的是「司令不同**且非無主**」（`20-military.md` §1b），
// 而這裡把無主省從敵人裡排除，所以一個被無主省環繞的省**不算被包圍**。
func (w *AIWorld) EnemyNeighbours(p ProvinceID, faction GeneralID) []ProvinceID {
	prov, err := w.Table.At(p)
	if err != nil {
		return nil
	}
	var out []ProvinceID
	for _, n := range prov.Neighbours {
		np, err := w.Table.At(n)
		if err != nil || np.Commander == 0 || np.Commander == faction {
			continue
		}
		out = append(out, n)
	}
	return out
}

// Encircled 回報某省是不是**每一個有效鄰省都是敵方的**（`sub_5A9F6`）。
//
//	total = 有效鄰省數（非 0、非 255）
//	enemy = EnemyNeighbours 的數量
//	回 total == enemy
//
// ⚠️ 判斷用的是「該省自己的司令」當我方勢力（原版讀 `+20` 再傳給 `sub_5A907`），
// 所以無主省算不出有意義的結果——這裡直接回 false。
//
// ⚠️ 沒有敵對鄰省而且也沒有有效鄰省（例如四面環海）時，`0 == 0` 成立，
// **原版會回 true**。照抄這個行為：一個孤島省會被判成「被包圍」。
// 這看起來像邊界情況沒被想到，但它是原版的行為，不擅自修。
func (w *AIWorld) Encircled(p ProvinceID) bool {
	prov, err := w.Table.At(p)
	if err != nil || prov.Commander == 0 {
		return false
	}
	total := len(prov.Neighbours)
	return total == len(w.EnemyNeighbours(p, prov.Commander))
}

// FirstEncircledNeighbour 在某省的鄰省裡找第一個「被敵人完全包圍」的
// （`sub_170C0`，68 行——決策鏈 A 最短的一支）。
//
// 原版把省份的 8 byte 鄰省表整個複製到堆疊再逐一檢查，
// 找到就寫進 `var` 參數並回 true。回 0 表示沒有。
func (w *AIWorld) FirstEncircledNeighbour(p ProvinceID) ProvinceID {
	prov, err := w.Table.At(p)
	if err != nil {
		return 0
	}
	for _, n := range prov.Neighbours {
		if w.Encircled(n) {
			return n
		}
	}
	return 0
}
