package game

import "sort"

// 戰鬥 AI 挑目標城市的那一段（`sub_3B492` 367 行 + `sub_3B400`）。
//
// `docs/re/31` §5 記過「戰鬥 AI 的每一個決策都圍繞城市」。這裡是它的核心：
//
//	掃全場城市（`sub_55FBE`）
//	  → 每個城市算一個 32-bit 分數（`sub_3B492`）
//	  → 照分數**升序**排
//	  → 取第一個正分的當目標（`sub_3C89A` 迴圈 2）
//
// 而分數就是**距離**（`sub_3B400`），所以整句話是：
//
//	**挑最近的、敵方佔據的城市。**

// CityScore 是一個城市格的評分結果。
type CityScore struct {
	// Cell 是城市所在的格。`NoCell` 表示這一格已被剔除。
	Cell CellIndex
	// Score 是分數，原版存成 32-bit（低位／高位兩個 word）。
	// 實際值是曼哈頓距離，用不到 32 bit，但照原版的寬度留著。
	Score int32
}

// CellManhattan 是 `sub_3B400`：兩格之間的**曼哈頓距離**。
//
//	col = cell mod 14        row = cell div 14
//	回 |row₁ − row₂| + |col₁ − col₂|
//
// ⚠️ **戰場是六角格，這個距離卻是矩形的曼哈頓距離。**
// 原版就是這樣估的（`ColRow` 用的 `GridW` 同樣是 14），不是筆誤——
// `sub_3B400` 從頭到尾只有兩次 `div 14` 與兩次絕對值，沒有六角修正。
// 照抄，不在 remake 改成正確的六角距離。
func CellManhattan(a, b CellIndex) int {
	ac, ar := a.ColRow()
	bc, br := b.ColRow()
	return absCell(ar-br) + absCell(ac-bc)
}

// absCell 不叫 abs：`deploy_test.go` 已經有一個同名的測試輔助函式。
func absCell(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// RankCityTargets 是 `sub_3B492`：給每個城市格打分並排序。
//
//	cities    `sub_55FBE` 掃出來的城市格清單（最多 10 個）
//	occupancy 格佔用表（`word_62A8`）
//	from      量距離的起點，即首位單位所在的格（`將領[word_64902].+5`）
//	unit      查一個單位的欄位；回 `nil` 表示查無此人
//
// 三種結果：
//
//	該格沒人站                    → 分數 0
//	站的是**攻方**（`+8 != 0`）    → 分數 0，而且可能整格被剔除
//	站的是**守方**（`+8 == 0`）    → 分數 = 到 `from` 的曼哈頓距離
//
// 剔除的條件（`loc_3B572`），兩者取一：
//
//	`+12`（下一跳格）== 這一格 且 `+9` == **命令 5**
//	`+12` == 0FFh（沒有下一跳）
//
// ⭐ 第一條是**整份程式唯一比較命令 5 的地方**（`docs/re/31` §21）。
// 讀作「已經有攻方單位朝這個城市去了，別再派人」。
//
// 排序是**升序**，而且 `NoCell` 不參與交換（`loc_3B6A7` 的前置）。
// 呼叫端取「第一個分數 > 0 的」，也就是**距離最近的敵方城市**——
// 距離 0 的那格（自己站著）會被 `> 0` 濾掉。
func RankCityTargets(cities []CellIndex, occupancy *Occupancy, from CellIndex,
	unit func(GeneralID) *CombatUnit) []CityScore {
	out := make([]CityScore, len(cities))
	for i, c := range cities {
		out[i] = CityScore{Cell: c}
		if !c.Valid() {
			continue
		}
		id := occupancy[c]
		if id == 0 {
			continue // 空格：分數 0
		}
		u := unit(id)
		if u == nil {
			continue
		}
		if u.Attacking {
			// 攻方站著。滿足剔除條件就把這一格拿掉。
			if (u.NextCell == c && u.Command == BattleCmdUnknown5) ||
				u.NextCell == NoCell {
				out[i].Cell = NoCell
			}
			continue // 不論剔除與否，分數都是 0
		}
		// 守方站著 → 這是可以打的城市，分數 = 距離。
		out[i].Score = int32(CellManhattan(from, c))
	}

	// 升序。`NoCell` 留在原位不參與交換——原版在交換前先檢查
	// `城市清單[j] == 0FFh` 就跳過，所以被剔除的格子不會被搬動。
	movable := make([]int, 0, len(out))
	for i := range out {
		if out[i].Cell != NoCell {
			movable = append(movable, i)
		}
	}
	vals := make([]CityScore, len(movable))
	for k, i := range movable {
		vals[k] = out[i]
	}
	sort.SliceStable(vals, func(a, b int) bool { return vals[a].Score < vals[b].Score })
	for k, i := range movable {
		out[i] = vals[k]
	}
	return out
}

// FirstPositiveCity 是 `sub_3C89A` 的第二個迴圈：取**第一個分數為正**的城市。
//
// 配上 `RankCityTargets` 的升序，這就是「最近的敵方城市」。
// 回 `NoCell` 表示沒有可打的城市。
func FirstPositiveCity(scores []CityScore) CellIndex {
	for _, s := range scores {
		if s.Cell != NoCell && s.Score > 0 {
			return s.Cell
		}
	}
	return NoCell
}

// WithinTwoSteps 是 `sub_55BCC(a, b)`：**a 在 b 的兩步之內嗎**。
//
//	sub_55BCC(a, b) = sub_510E0(a, b) 或 sub_55AB7(a, b)
//
// 前者是一步（邊界／相鄰規則），後者是兩層六方向掃描（`docs/re/31` §25）。
// 原版兩層迴圈各自跑 `sub_510E0` 過濾越界，這裡用既有的 `Neighbour`
// ——它本來就只回合法的鄰格。
//
// `sub_3BCED` 拿它篩候選：**距離目標城市兩格之內、站著單位的格。**
//
// ⭐ `a == b` **成立**——兩層掃描會走出去再走回來（n1 是 b 的鄰格，
// n2 是 n1 的鄰格，其中一個就是 b 自己）。原版沒有排除這個情形，照抄。
func WithinTwoSteps(a, b CellIndex) bool {
	if !a.Valid() || !b.Valid() {
		return false
	}
	if Adjacent(a, b) {
		return true
	}
	for _, n1 := range b.Neighbours() {
		for _, n2 := range n1.Neighbours() {
			if n2 == a {
				return true
			}
		}
	}
	return false
}

// ── 派工：從候選清單挑一個走得通的當目標 ────────────────────────────

// TargetAssignment 是一次派工的結果。
type TargetAssignment struct {
	// Target 是選中的目標單位，0 表示一個都走不到。
	Target GeneralID
	// NextCell 是往目標走的下一格，`NoCell` 表示走不到。
	NextCell CellIndex
}

// AssignTargetFrom 是**三支同構函式**的共同內容
// （`sub_3BF6A` §24、`sub_3CA09` §29、`sub_58209` §27）：
//
//	逐一試候選 → 第一個尋路走得通的就是目標
//
// 三支的差別只有兩項：清單從哪來（城市周圍／敵方主力周圍），
// 以及**要不要先照距離排序**（只有 `sub_58209` 排）。
//
//	candidates 候選單位，順序就是嘗試順序
//	from       出發格（`將領[u].+5`）
//	cellOf     查一個單位站在哪一格
//	route      尋路，回下一跳；走不到回 `NoCell`（原版的 `sub_567B9`）
//
// 回傳 `Target == 0` 表示全部走不到——原版這時另外呼叫 `sub_3B8B0`
// （`sub_3BF6A`）或什麼都不做（`sub_3CA09`），由呼叫端決定。
func AssignTargetFrom(candidates []GeneralID, from CellIndex,
	cellOf func(GeneralID) CellIndex,
	route func(to, fromCell CellIndex) CellIndex) TargetAssignment {
	for _, c := range candidates {
		next := route(cellOf(c), from)
		if next != NoCell {
			return TargetAssignment{Target: c, NextCell: next}
		}
	}
	return TargetAssignment{NextCell: NoCell}
}

// SortCandidatesByDistance 是 `sub_58209` 比另外兩支多做的那一步：
// 把候選照**到 `from` 的曼哈頓距離**由近而遠排。
//
// 原版是選擇排序，交換條件 `dA > dB`，所以**相等時不換**——
// 用 `SliceStable` 對應。站不在場上的候選（`+5 == 0xFF`）原版直接跳過
// 不參與比較，這裡排到最後面。
func SortCandidatesByDistance(candidates []GeneralID, from CellIndex,
	cellOf func(GeneralID) CellIndex) {
	sort.SliceStable(candidates, func(i, j int) bool {
		ci, cj := cellOf(candidates[i]), cellOf(candidates[j])
		if !ci.Valid() || !cj.Valid() {
			return ci.Valid() && !cj.Valid()
		}
		return CellManhattan(from, ci) < CellManhattan(from, cj)
	})
}
