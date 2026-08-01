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

	// Decisive 表示這個行動要求**直接判勝負**，不再逐格打（§16 的必勝結算）。
	// 為真時 `DecisiveAttackerWon` 才有意義，推進器看到就該結束戰鬥。
	Decisive            bool
	DecisiveAttackerWon bool
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
	case ActAReset:
		return s.execReset(units, route)
	case ActADefault:
		return s.execDefault(units, route)
	case ActADecapitateKeepOne:
		return s.execDecapitate(units, foes, route, true)
	case ActADecapitateKeepAll:
		return s.execDecapitate(units, foes, route, false)
	case ActBStrikeForce:
		return s.execStrikeForce(units, foes, route)
	case ActAStandbyOnly:
		return s.execStandbyOnly(units, foes, route)
	case ActARecompute:
		return s.execRecompute(units, route)
	case ActAWeakest:
		return s.execWeakest(units, foes, route)
	case ActADecisive:
		// 分支 A 選中它 = **第二方（守方）被壓到五分之一以下** → 攻方勝。
		// 原版 `sub_3B19C(0)` → `sub_54E3B(word_64944, word_64942)`，
		// 第二個參數是勝方 = `word_64942` = 第一方（§16／§37）。
		return BattleExecResult{Implemented: true, Decisive: true,
			DecisiveAttackerWon: true, Note: "必勝結算：守方被壓到 1/5 以下"}
	case ActBDecisive:
		// 分支 B 選中它 = **第一方（攻方）被壓到五分之一以下** → 守方勝。
		return BattleExecResult{Implemented: true, Decisive: true,
			DecisiveAttackerWon: false, Note: "必勝結算：攻方被壓到 1/5 以下"}
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
	// ⭐ §19 的骨架第一步：值 3 與值 4 都**先跑預備隊投入**（`sub_3C26A`），
	// 那支把待命(2) 的單位轉成命令 4，後面的派工才有東西可派。
	//
	// ⛔ 漏掉這一步的後果是整場戰鬥空轉——決策鏈每回合都選值 3，
	// 但所有單位都是待命，而值 3 只派命令 3 與 4
	// （2026-08-02 被 `TestChainRunActuallyMovesAndFights` 抓到）。
	s.activateReserves(units)

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


// execReset 是值 12：**推倒重來**（§33）。
//
//	先把每個單位打回待命（`sub_3B15E`：+9=2、+12=0xFF、+10=0）
//	→ 再掃格清單，找第一個「站著守方、而且走得通」的格
//
// ⚠️ **不看現有命令**——這是值 12 與值 13 最大的差別（13 依命令分流）。
func (s *BattleSim) execReset(units []*Combatant,
	route func(to, from CellIndex) CellIndex) BattleExecResult {
	n := 0
	for _, u := range units {
		if u == nil || !u.Alive() {
			continue
		}
		u.ResetToStandby() // sub_3B15E

		// 掃全場找站著守方的格。原版掃的是 `sub_560D7` 給的清單，
		// 那一支未讀（§32），這裡用全場掃描近似——**標為差異**。
		for i := 0; i < CellCount; i++ {
			c := CellIndex(i)
			v := s.Occ[c]
			if v == 0 {
				continue
			}
			t := s.Unit(v)
			if t == nil || t.Attacking { // 只找守方（+8 == 0）
				continue
			}
			next := route(c, u.Cell)
			if next == NoCell {
				continue
			}
			u.AssignTo(v, next)
			n++
			break
		}
	}
	return BattleExecResult{Assigned: n, Implemented: true,
		Note: "推倒重來：全體重置後重新找目標"}
}

// execDefault 是值 13：**預設分流**（§32）。依單位**現有的命令**決定做什麼。
//
//	命令 1  站在城市上 → 下一跳 = 原地；否則找空城
//	命令 2/3 壓成 2，清下一跳
//	命令 4  下一跳 = 原地
//
// ⚠️ 「找空城」原版取的是**最後一個**空城（迴圈不 break，後面覆蓋前面），
// 不是最近的——與值 3 的「照距離排序取最近」是完全不同的挑法。照抄。
func (s *BattleSim) execDefault(units []*Combatant,
	route func(to, from CellIndex) CellIndex) BattleExecResult {
	cities := CityCells(s.Field)
	n := 0
	for _, u := range units {
		if u == nil || !u.Alive() {
			continue
		}
		switch u.Command {
		case BattleCmdGarrison: // 命令 1
			if s.tileOf(u).Kind == assets.TileCity {
				u.NextCell = u.Cell // 已經到了，原地
				n++
				continue
			}
			u.NextCell = NoCell
			for _, c := range cities {
				if s.Occ[c] == 0 {
					u.NextCell = c // ⚠️ 不 break：取最後一個空城
				}
			}
			if u.NextCell == NoCell && len(cities) > 0 {
				u.NextCell = cities[0] // 一個空城都沒有 → 取第一個
			}
			n++
		case BattleCmdStandby, BattleCmdSeekTarget: // 命令 2 / 3
			u.Command = BattleCmdStandby // 統一壓成 2
			u.NextCell = NoCell
			n++
		case BattleCmdCommitted: // 命令 4
			u.NextCell = u.Cell
			n++
		}
	}
	_ = route // 命令 2/3 那條原版還會用 sub_560D7 找格，該支未讀
	return BattleExecResult{Assigned: n, Implemented: true,
		Note: "預設分流：依現有命令處理"}
}

// execDecapitate 是值 14／15：**斬首**（§34）——全軍撲向敵方首位單位。
//
//	keepOne = true （值 14）只留**一個**在城市駐守，其餘全部出擊
//	keepOne = false（值 15）**所有**駐守中的都留下
//
// ⚠️ **15 比 14 保守**，與編號直覺相反。原版是同一支函式，
// 差別只在「第一個命令 1 的單位有沒有設旗標」。
func (s *BattleSim) execDecapitate(units, foes []*Combatant,
	route func(to, from CellIndex) CellIndex, keepOne bool) BattleExecResult {
	lead := firstLiving(foes) // 敵方首位單位（原版 word_64902）
	if lead == nil {
		return BattleExecResult{Implemented: true, Note: "敵方沒有在場單位"}
	}

	kept := false
	n := 0
	for _, u := range units {
		if u == nil || !u.Alive() {
			continue
		}
		if u.Command == BattleCmdGarrison && !kept {
			if keepOne {
				kept = true // 值 14：設了旗標，後面的駐守單位就會被派出去
			}
			continue // 兩個值都跳過這一個
		}
		next := route(lead.Cell, u.Cell)
		if next == NoCell {
			continue
		}
		u.AssignTo(lead.General, next)
		n++
	}
	mode := "只留一個守城"
	if !keepOne {
		mode = "駐守的都留"
	}
	return BattleExecResult{Assigned: n, Implemented: true, Note: "斬首（" + mode + "）"}
}


// execStrikeForce 是值 4：**打敵方主力周邊**（§29）。
//
// 骨架與值 3（打城市）相同，只換中心點：
//
//	中心 = 敵方主力所在格（原版 word_64944 指到的單位）
//	候選 = 距離中心兩格之內、站著單位的格（§25 的 sub_55BCC）
//	→ 只派命令 3 與 4 的單位，逐一試路，第一個走得通的當目標
//
// ⚠️ 與值 3 的另一個差別：**成功後不標旗標**（§29 的對照表）。
func (s *BattleSim) execStrikeForce(units, foes []*Combatant,
	route func(to, from CellIndex) CellIndex) BattleExecResult {
	s.activateReserves(units) // 與值 3 同一個前置（§29 的呼叫序列比對）

	center := firstLiving(foes)
	if center == nil {
		return BattleExecResult{Implemented: true, Note: "敵方主力不在場"}
	}

	// 候選：距離中心兩格之內、站著單位的格（§25）。
	var pool []GeneralID
	for i := 0; i < CellCount; i++ {
		c := CellIndex(i)
		v := s.Occ[c]
		if v == 0 || !WithinTwoSteps(c, center.Cell) {
			continue
		}
		pool = append(pool, v)
	}
	if len(pool) == 0 {
		return BattleExecResult{Implemented: true, Note: "主力周邊沒有單位"}
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
		if u.Command != BattleCmdSeekTarget && u.Command != BattleCmdCommitted {
			continue // 只派命令 3 與 4（§23）
		}
		got := AssignTargetFrom(pool, u.Cell, cellOf, route)
		if got.Target == 0 {
			continue
		}
		// ⚠️ 值 4 不標 +13 bit 7，所以不用 AssignTo。
		u.TargetUnit, u.NextCell = got.Target, got.NextCell
		n++
	}
	return BattleExecResult{Assigned: n, Implemented: true,
		Note: "打主力周邊：候選 " + itoa(len(pool)) + " 個"}
}

// execStandbyOnly 是值 16：**只處理待命與已出發的單位**（§35）。
//
//	命令 2        無條件處理
//	命令 4 / 5    要 `byte_64900 ≥ 5` 才處理（`lateStage`）
//	其餘          跳過
//	已經與目標相鄰就不動
//
// `lateStage` 對應原版的 `byte_64900 ≥ 5`——那個變數第三次出現
// （§17 的 > 4、§35 的 ≥ 5、§42 的加項），**語意仍未解**，
// 所以由呼叫端傳。
func (s *BattleSim) execStandbyOnly(units, foes []*Combatant,
	route func(to, from CellIndex) CellIndex) BattleExecResult {
	return s.execStandbyOnlyStage(units, foes, route, true)
}

func (s *BattleSim) execStandbyOnlyStage(units, foes []*Combatant,
	route func(to, from CellIndex) CellIndex, lateStage bool) BattleExecResult {
	center := firstLiving(foes)
	if center == nil {
		return BattleExecResult{Implemented: true, Note: "敵方不在場"}
	}
	n, skipped := 0, 0
	for _, u := range units {
		if u == nil || !u.Alive() {
			continue
		}
		switch u.Command {
		case BattleCmdStandby: // 命令 2：無條件
		case BattleCmdCommitted, BattleCmdUnknown5: // 命令 4／5：要後期
			if !lateStage {
				skipped++
				continue
			}
		default:
			continue
		}
		if Adjacent(u.Cell, center.Cell) {
			continue // 已相鄰就不動（§35）
		}
		next := route(center.Cell, u.Cell)
		if next == NoCell {
			continue
		}
		u.AssignTo(center.General, next)
		n++
	}
	note := "只處理待命與已出發"
	if skipped > 0 {
		note += "（前期，跳過 " + itoa(skipped) + " 個命令 4／5）"
	}
	return BattleExecResult{Assigned: n, Implemented: true, Note: note}
}

// execRecompute 是值 17：**重算全軍的行動**（§14）。
//
// 原版三個迴圈，順序有意義：
//
//	1. 命令 3 且沒有目標格 → 降成 2
//	2. 命令 2／4／5 一批    → 重算下一格
//	3. 命令 1 單獨一批      → 重算下一格
//
// ⚠️ 命令 1 單獨最後處理，不與 2／4／5 混在一起——那是「守城」與
// 「移動」兩種性質的分界（§14 的語意線索）。
func (s *BattleSim) execRecompute(units []*Combatant,
	route func(to, from CellIndex) CellIndex) BattleExecResult {
	// 迴圈 1：命令 3 的清理。
	for _, u := range units {
		if u == nil || !u.Alive() {
			continue
		}
		if u.Command == BattleCmdSeekTarget && u.NextCell == NoCell {
			u.Command = BattleCmdStandby
		}
	}

	recalc := func(u *Combatant) bool {
		if u.TargetUnit == 0 {
			return false
		}
		t := s.Unit(u.TargetUnit)
		if t == nil || !t.Alive() {
			return false
		}
		next := route(t.Cell, u.Cell)
		u.NextCell = next
		return next != NoCell
	}

	n := 0
	// 迴圈 2：命令 2／4／5 一批。
	for _, u := range units {
		if u == nil || !u.Alive() {
			continue
		}
		switch u.Command {
		case BattleCmdStandby, BattleCmdCommitted, BattleCmdUnknown5:
			if recalc(u) {
				n++
			}
		}
	}
	// 迴圈 3：命令 1 單獨一批。
	for _, u := range units {
		if u == nil || !u.Alive() {
			continue
		}
		if u.Command == BattleCmdGarrison && recalc(u) {
			n++
		}
	}
	return BattleExecResult{Assigned: n, Implemented: true, Note: "重算全軍的下一格"}
}


// execWeakest 是值 18：**在圍著我方城市的敵軍裡，挑最弱的打**（§46）。
//
//	候選 = 對每個「守方佔據的城市」，收集它周圍的**攻方**單位
//	→ 照攻擊力**升序**排（弱的在前）
//	→ 派工
//
// ⭐ 這是分支 A 九種行動裡**唯一比戰力**的一支——其他行動比的是
// 距離或位置關係。三種目標選法各有性格：
//
//	值 3      最近的城市
//	值 14／15 敵方主帥（斬首）
//	值 18     圍住我的人裡最弱的
//
// ⚠️ `friendly` 判定用 `Attacking`：候選是攻方、城市要守方佔著。
// 這一支只在分支 A（守方）被選中，所以方向是固定的。
func (s *BattleSim) execWeakest(units, foes []*Combatant,
	route func(to, from CellIndex) CellIndex) BattleExecResult {
	cities := CityCells(s.Field)
	if len(cities) == 0 {
		return BattleExecResult{Implemented: true, Note: "戰場上沒有城市"}
	}

	seen := map[GeneralID]bool{}
	var pool []GeneralID
	for _, c := range cities {
		holder := s.Occ[c]
		if holder == 0 {
			continue
		}
		h := s.Unit(holder)
		if h == nil || h.Attacking {
			continue // 只看守方佔據的城市
		}
		for _, n := range RingCells(c, 1) {
			v := s.Occ[n]
			if v == 0 || seen[v] {
				continue
			}
			t := s.Unit(v)
			if t == nil || !t.Attacking {
				continue // 只收攻方單位
			}
			seen[v] = true
			pool = append(pool, v)
		}
	}
	if len(pool) == 0 {
		return BattleExecResult{Implemented: true, Note: "沒有敵軍圍在我方城市周圍"}
	}

	// 照攻擊力升序——弱的排前面（§38 的排序方向）。
	strengthOf := func(id GeneralID) int {
		if u := s.Unit(id); u != nil {
			return s.StrengthOf(u)
		}
		return 0
	}
	for i := 0; i < len(pool); i++ {
		for j := i + 1; j < len(pool); j++ {
			if strengthOf(pool[i]) > strengthOf(pool[j]) {
				pool[i], pool[j] = pool[j], pool[i]
			}
		}
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
		got := AssignTargetFrom(pool, u.Cell, cellOf, route)
		if got.Target == 0 {
			continue
		}
		u.AssignTo(got.Target, got.NextCell)
		n++
	}
	return BattleExecResult{Assigned: n, Implemented: true,
		Note: "挑最弱：圍城敵軍 " + itoa(len(pool)) + " 個"}
}


// activateReserves 是 `sub_3C26A` 的前半（§17）：**把待命的單位投入**。
//
// 留幾個待命由「首位單位對敵軍全體的優勢等級」決定：
//
//	劣勢 留 2      優勢 留 1      壓倒 留 0
//
// 值 3（打城市）與值 4（打主力周邊）都以這一步開頭（§19／§29 的呼叫序列）。
//
// ⚠️ 原版還有四道前置決定「要不要跑這一步」（§19），那四道包含
// 未解的 `byte_6AA84` bit 6/7，這裡一律跑——**標為差異**。
func (s *BattleSim) activateReserves(units []*Combatant) {
	lead := firstLiving(units)
	if lead == nil {
		return
	}
	// 對手是另一方——用 Attacking 判斷 `units` 是哪一方。
	foes := s.Defender
	if !lead.Attacking {
		foes = s.Attacker
	}
	adv := AdvantageLevel(s.StrengthOf(lead), s.SideStrength(foes))

	// `ActivateReserves` 吃 index 1..10 的命令陣列（原版的槽位佈局）。
	cmds := make([]uint8, UnitsPerSide+1)
	idx := make([]*Combatant, UnitsPerSide+1)
	n := 0
	for _, u := range units {
		if u == nil || !u.Alive() || n >= UnitsPerSide {
			continue
		}
		n++
		cmds[n], idx[n] = u.Command, u
	}
	ActivateReserves(cmds, adv)
	for i := 1; i <= n; i++ {
		if idx[i] != nil {
			idx[i].Command = cmds[i]
		}
	}
}
