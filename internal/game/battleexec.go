package game

import "github.com/wicanr2/great-era-remake/internal/assets"

// 把決策鏈選出的行動**真正執行**——也就是把命令、目標、下一跳寫進單位。
//
// 決策層在 `battleai.go`（選什麼）、接線在 `battledrive.go`（在真實狀態上跑），
// 這一層是「選完之後做什麼」。
//
// ⚠️ **13 種行動只實作了一部分**，見 `ExecuteAction` 的回傳值。
// 沒實作的不會靜默跳過，會明確回報——`AutoResolve` 那套推進行為
// 在全部補齊之前仍然是主力（`battledrive.go` 的說明）。

// CityCells 掃出戰場上所有的城市格（`sub_55FBE`，`docs/re/31` §5）。
//
// 原版掃 0..195 找 `sub_50151(格) == 5`，最多收 10 個，不足補 0xFF。
// **戰鬥 AI 的每一個決策都圍繞城市**，所以這是很多行動的第一步。
//
// 這裡不補 0xFF 哨兵——Go 用切片長度表達個數，呼叫端不需要那個約定。
func CityCells(bf *Battlefield) []CellIndex {
	var out []CellIndex
	for i := 0; i < CellCount && len(out) < UnitsPerSide; i++ {
		c := CellIndex(i)
		col, row := c.ColRow()
		if bf.Tiles[row][col].Kind == assets.TileCity {
			out = append(out, c)
		}
	}
	return out
}

// RingCells 收集某一格周圍 `rings` 圈的格（`sub_55632`，§12）。
//
// 原版 `mode = 1` 收**三圈共 37 格**（1 + 6 + 12 + 18），第一格是中心自己。
// 佈防（值 2，§30）用的就是這個 37 格清單。
//
// ⚠️ 原版另有 `mode = 0`（值 3 的候選清單用），**兩者的差別未驗**（§29）。
// 這裡只實作三圈那一種，`rings` 給呼叫端調是為了測試方便，不是原版有這個參數。
func RingCells(center CellIndex, rings int) []CellIndex {
	if !center.Valid() {
		return nil
	}
	out := []CellIndex{center}
	seen := map[CellIndex]bool{center: true}
	frontier := []CellIndex{center}
	for r := 0; r < rings; r++ {
		var next []CellIndex
		for _, c := range frontier {
			for _, n := range c.Neighbours() {
				if seen[n] {
					continue
				}
				seen[n] = true
				out = append(out, n)
				next = append(next, n)
			}
		}
		frontier = next
	}
	return out
}

// BattleExecResult 是執行一個行動之後的結果。
type BattleExecResult struct {
	// Assigned 是這一輪被指派了目標的單位數。
	Assigned int
	// Implemented 表示這個行動有實作。false 時什麼都沒做——
	// **呼叫端必須自己決定要不要退回舊的推進行為**。
	Implemented bool
	// Note 是給人看的說明，未實作時寫「還沒實作」與出處。
	Note string
}

// ExecuteAction 執行決策鏈選出的行動。
//
//	units    要下命令的那一方（分支 A 給守方、分支 B 給攻方）
//	foes     對方
//	route    尋路，回下一跳；走不到回 NoCell（原版 `sub_567B9`）
//
// ⚠️ **回傳 `Implemented == false` 時什麼都沒做。** 這是刻意的：
// 靜默跳過會讓「AI 這回合沒動」看起來像是決策結果，而不是實作缺口。
func (s *BattleSim) ExecuteAction(a BattleAction, units, foes []*Combatant,
	route func(to, from CellIndex) CellIndex) BattleExecResult {
	switch a {
	case ActBTakeCity:
		return s.execTakeCity(units, route)
	case ActBDeploy:
		return s.execDeploy(units)
	case ActAEngageAll:
		return s.execEngageAll(units, foes, route)
	}
	return BattleExecResult{
		Note: "行動「" + BattleActionName(a) + "」還沒實作（docs/re/31 §41 有語意，執行層待補）",
	}
}

// execTakeCity 是值 3：**挑最近的敵方城市**（§19／§21）。
//
//	掃全場城市 → 每個算距離 → 升序排 → 取第一個正分
//	→ 城市有守軍：全員往城市走，目標是那個守軍
//	→ 城市空著：改追附近的人（§24，這裡未實作）
func (s *BattleSim) execTakeCity(units []*Combatant,
	route func(to, from CellIndex) CellIndex) BattleExecResult {
	cities := CityCells(s.Field)
	if len(cities) == 0 {
		return BattleExecResult{Implemented: true, Note: "戰場上沒有城市"}
	}
	lead := firstLiving(units)
	if lead == nil {
		return BattleExecResult{Implemented: true, Note: "我方沒有存活單位"}
	}

	scores := RankCityTargets(cities, &s.Occ, lead.Cell, func(id GeneralID) *CombatUnit {
		if u := s.Unit(id); u != nil {
			return &u.CombatUnit
		}
		return nil
	})
	target := FirstPositiveCity(scores)
	if target == NoCell {
		return BattleExecResult{Implemented: true, Note: "沒有可打的敵方城市"}
	}

	// 城市上的守軍就是目標單位（§23）。
	garrison := s.Occ[target]
	n := 0
	for _, u := range units {
		if u == nil || !u.Alive() {
			continue
		}
		// 只派命令 3 與 4 的單位——兩者功能上同類（§23）。
		if u.Command != BattleCmdSeekTarget && u.Command != BattleCmdCommitted {
			continue
		}
		next := route(target, u.Cell)
		if next == NoCell {
			continue
		}
		u.AssignTo(garrison, next)
		n++
	}
	return BattleExecResult{Assigned: n, Implemented: true,
		Note: "打城市：目標格 " + itoa(int(target))}
}

// execDeploy 是值 2：**佈防**（§30）。
//
//	以首位單位所在格為中心取周圍三圈 37 格當位置表
//	→ 依序發給「還沒有下一跳」的單位
//	→ 已經站在城市上的不動
func (s *BattleSim) execDeploy(units []*Combatant) BattleExecResult {
	lead := firstLiving(units)
	if lead == nil {
		return BattleExecResult{Implemented: true, Note: "我方沒有存活單位"}
	}
	slots := RingCells(lead.Cell, 3) // 三圈 = 37 格
	// 位置表的第一格是中心（首位單位自己），所以從 index 1 開始發。
	next := 1
	n := 0
	for _, u := range units {
		if u == nil || !u.Alive() || u == lead {
			continue
		}
		if u.NextCell != NoCell {
			continue // 已經有去處
		}
		if s.tileOf(u).Kind == assets.TileCity {
			continue // 已經站在城市上就不動
		}
		for next < len(slots) && s.Occ[slots[next]] != 0 {
			next++
		}
		if next >= len(slots) {
			break
		}
		u.NextCell = slots[next]
		next++
		n++
	}
	return BattleExecResult{Assigned: n, Implemented: true,
		Note: "佈防：位置表 " + itoa(len(slots)) + " 格"}
}

// execEngageAll 是值 19：**全面接戰**（§40）。
//
//	把對方還在場上的單位壓實成清單（原版是把陣列往前擠）
//	→ 每個我方單位跑 `sub_58209`：照距離排序、逐一試路
//	→ 成功則設命令 3 + 目標 + 下一跳
func (s *BattleSim) execEngageAll(units, foes []*Combatant,
	route func(to, from CellIndex) CellIndex) BattleExecResult {
	// 壓實：只留還在場上的。
	var pool []GeneralID
	for _, f := range foes {
		if f != nil && f.Alive() && f.Cell.Valid() {
			pool = append(pool, f.General)
		}
	}
	if len(pool) == 0 {
		return BattleExecResult{Implemented: true, Note: "對方沒有在場的單位"}
	}

	cellOf := func(id GeneralID) CellIndex {
		if u := s.Unit(id); u != nil {
			return u.Cell
		}
		return NoCell
	}

	n := 0
	for _, u := range units {
		if u == nil || !u.Alive() {
			continue
		}
		// 每個單位各自照「離我多近」排一次候選（§27）。
		cand := append([]GeneralID(nil), pool...)
		SortCandidatesByDistance(cand, u.Cell, cellOf)
		got := AssignTargetFrom(cand, u.Cell, cellOf, route)
		if got.Target == 0 {
			continue
		}
		u.AssignTo(got.Target, got.NextCell)
		u.Command = BattleCmdSeekTarget // 值 19 把命令設成 3
		n++
	}
	return BattleExecResult{Assigned: n, Implemented: true,
		Note: "全面接戰：目標池 " + itoa(len(pool)) + " 個"}
}

func firstLiving(us []*Combatant) *Combatant {
	for _, u := range us {
		if u != nil && u.Alive() && u.Cell.Valid() {
			return u
		}
	}
	return nil
}

// itoa 避免為了組訊息把 strconv 拉進來——這一層只需要小整數。
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [8]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
