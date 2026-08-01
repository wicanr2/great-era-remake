package game

// 用**決策鏈**驅動一整場戰鬥（`AutoResolveByChain`）。
//
// 這是三層的最後一塊：
//
//	battleai.go     決策鏈——選哪個行動（13 種）
//	battleexec.go   執行層——把命令、目標、下一跳寫進單位
//	battlerun.go    推進器——依那些命令實際移動與交戰   ← 這裡
//
// ⚠️ **既有的 `AutoResolve` 保留不動。** 它用的是簡化行為
// （朝最近的敵人走），與這裡是兩套獨立的推進方式。
// 保留的理由：**13 種行動雖然都實作了，但三個底層判斷的語意仍未解**
// （比率門檻的來源、`sub_534FF` 的輸出、兩個旗標），那些由呼叫端傳。
// 在它們解出來之前，這條鏈的行為與原版還有已知落差，兩套並存才能對照。
//
// ⭐ 實測（3v3、守方 5000×3）：**10 回合、移動 4、交戰 6、未實作 0、
// 攻方損 4715 守方損 3396**——決策鏈能把一場戰鬥完整跑完。

// BattleRunStats 是一場決策鏈驅動的戰鬥跑完之後的統計。
type BattleRunStats struct {
	BattleOutcome
	// Decisions 是每回合兩條鏈各選了什麼。
	Decisions []BattleTurnDecision
	// Unimplemented 記錄「決策鏈選了但執行層還沒實作」的次數。
	//
	// ⭐ **這個數字是驗收指標**：它大於 0 就表示這場戰鬥裡，
	// 電腦有幾回合其實什麼都沒做。不看它就會把「AI 很被動」
	// 誤讀成原版行為，而不是實作缺口。
	Unimplemented int
	// Moves / Engagements 是實際移動與交戰的次數。
	Moves, Engagements int

	// Decisive 表示這場戰鬥是被**必勝結算**判掉的，不是打到一方全滅。
	// 原版戰力差五倍時就不再逐格打（§16）——這個旗標讓呼叫端分得出來，
	// 免得把「秒結束」誤讀成模擬出錯。
	Decisive     bool
	DecisiveNote string
}

// dirToward 找出從 `from` 走到相鄰的 `to` 要用哪個方向。
//
// `+12`（下一跳）依定義是相鄰格，所以找不到方向就表示
// 尋路給了不相鄰的格——那是呼叫端的錯，回 0 讓上層跳過。
func dirToward(from, to CellIndex) HexDir {
	for d := HexDir(1); d <= 6; d++ {
		if n, ok := from.Neighbour(d); ok && n == to {
			return d
		}
	}
	return 0
}

// stepByOrder 依單位的 `+12`（下一跳）走一格。
//
// 回傳是否真的動了。走到之後把 `+12` 清成 `NoCell`——
// 下一跳是一次性的，要繼續走得重新算（原版每回合都會重算，§14）。
func (s *BattleSim) stepByOrder(u *Combatant) bool {
	if u == nil || !u.Alive() || u.NextCell == NoCell || u.NextCell == u.Cell {
		return false
	}
	d := dirToward(u.Cell, u.NextCell)
	if d == 0 {
		return false
	}
	if _, err := s.Move(u.General, d); err != nil {
		return false
	}
	u.NextCell = NoCell
	return true
}

// engageIfAdjacent 讓單位對它的目標出手——**前提是目標就在隔壁**。
//
// 原版的交戰是移動到相鄰之後才發生的，這裡照同樣的順序：
// 先走（`stepByOrder`），走完檢查相不相鄰。
func (s *BattleSim) engageIfAdjacent(u *Combatant) bool {
	if u == nil || !u.Alive() || u.TargetUnit == 0 {
		return false
	}
	t := s.Unit(u.TargetUnit)
	if t == nil || !t.Alive() || !t.Cell.Valid() {
		return false
	}
	if !Adjacent(u.Cell, t.Cell) {
		return false
	}
	if _, _, err := s.Engage(u, t); err != nil {
		return false
	}
	return true
}

// AutoResolveByChain 用決策鏈把一場戰鬥跑完。
//
//	每回合：
//	  1. 兩條鏈各決定一個行動
//	  2. 執行層依行動寫命令／目標／下一跳
//	  3. 依 `+12` 移動，走完相鄰就交戰
//
//	gates           那三個還沒解出來源的判斷（`BattleChainGates`）
//	defenderLeader  當前交戰省的司令（§44 的 `sub_56D49` 要用）
//
// ⚠️ 回傳的 `Unimplemented` 大於 0 時，**這場戰鬥的結果不能拿來對照原版**
// ——有回合是因為執行層缺口而空轉的。
func (s *BattleSim) AutoResolveByChain(maxTurns int, gates BattleChainGates,
	defenderLeader GeneralID) BattleRunStats {
	if maxTurns <= 0 {
		maxTurns = AutoBattleTurnCap
	}
	// 尋路：沿用戰場的移動成本，回下一跳。
	route := func(to, from CellIndex) CellIndex {
		if !to.Valid() || !from.Valid() {
			return NoCell
		}
		if Adjacent(from, to) {
			return to
		}
		// 貪心一步：往「離目標最近而且走得上去」的鄰格。
		// ⚠️ 原版的 `sub_567B9` 是真正的尋路（331 行，§10），
		// 這裡是近似——**標為 remake 差異**，補完之前不要拿來做行為驗收。
		best, bestD := NoCell, 1<<30
		for _, n := range from.Neighbours() {
			if s.Occ[n] != 0 {
				continue
			}
			col, row := n.ColRow()
			if s.Field.Tiles[row][col].MoveCost() >= 255 {
				continue
			}
			if d := CellManhattan(n, to); d < bestD {
				best, bestD = n, d
			}
		}
		return best
	}

	before := func(us []*Combatant) int {
		n := 0
		for _, u := range us {
			n += int(u.Force())
		}
		return n
	}
	atk0, def0 := before(s.Attacker), before(s.Defender)

	var st BattleRunStats
	for turn := 1; turn <= maxTurns; turn++ {
		st.Turns = turn
		s.BeginTurn()

		d := s.DecideTurn(turn, gates, defenderLeader, 0)
		st.Decisions = append(st.Decisions, d)

		// 分支 B 指揮攻方（第一方）、分支 A 指揮守方（第二方）。
		// ⚠️ 那個對應仍未驗（§2 的 `byte_64901`），見 `DecideTurn` 的說明。
		rb := s.ExecuteAction(d.B.Action, s.Attacker, s.Defender, route)
		ra := s.ExecuteAction(d.A.Action, s.Defender, s.Attacker, route)
		if !rb.Implemented {
			st.Unimplemented++
		}
		if !ra.Implemented {
			st.Unimplemented++
		}

		// ⭐ 必勝結算：戰力差五倍時原版不再逐格打，直接判勝負（§16）。
		// 兩條鏈都可能選中它，先到先算——分支 B 先跑，所以它優先。
		if rb.Decisive || ra.Decisive {
			r := rb
			if !rb.Decisive {
				r = ra
			}
			st.AttackerWon, st.Decided = r.DecisiveAttackerWon, true
			st.Decisive = true
			st.DecisiveNote = r.Note
			break
		}

		for _, u := range append(append([]*Combatant(nil), s.Attacker...), s.Defender...) {
			if s.stepByOrder(u) {
				st.Moves++
			}
			if s.engageIfAdjacent(u) {
				st.Engagements++
			}
		}

		s.Sweep()
		if over, won := s.Over(); over {
			st.AttackerWon, st.Decided = won, true
			break
		}
		s.EndTurn()
	}
	st.AttackerLoss = atk0 - before(s.Attacker)
	st.DefenderLoss = def0 - before(s.Defender)
	return st
}
