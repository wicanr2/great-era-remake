package game

// 自動推進一場戰鬥，直到分出勝負。
//
// ⚠️ **推進策略不是原版的**。原版戰場上的行動決策還沒解出來
// （`docs/mechanics/70-ai.md` §7），這裡用的是最簡版本：
// 每個單位朝最近的敵人走，接觸就打。**這是 remake 差異**，
// 目的是讓政略層跑得下去，不是還原原版的戰術。
//
// 規則本身（移動成本、攻擊力、戰損、衰減）全部來自反組譯，
// 只有「先動誰、往哪走」是我們補的。
//
// 這份實作原本在 `cmd/battlesim`，移進規則層是為了不讓
// `cmd/aisim` 抄第二份——`CLAUDE.md` §7 第 4 條。

// BattleOutcome 是一場自動戰鬥的結果。
type BattleOutcome struct {
	// AttackerWon 表示攻方獲勝。雙方都還活著（打到回合上限）時為 false。
	AttackerWon bool
	// Decided 表示真的分出了勝負，不是撞到回合上限。
	Decided bool
	// Turns 是實際打了幾個回合。
	Turns int
	// AttackerLoss / DefenderLoss 是雙方的總兵力損失。
	AttackerLoss, DefenderLoss int
}

// AutoBattleTurnCap 是回合上限。
//
// 原版的戰鬥有沒有回合上限還沒解——`docs/re/05` 的戰鬥狀態 469 B
// 裡沒找到回合計數。這個數字純粹是防止無限迴圈的保險，
// **不是原版規則**：`+30` 每回合衰減 20% 且有下限（`sub_54826`），
// 所以拖久了雙方都打不動，戰鬥可能真的不會結束。
const AutoBattleTurnCap = 200

// AutoResolve 讓雙方自動打完。
func (s *BattleSim) AutoResolve(maxTurns int) BattleOutcome {
	if maxTurns <= 0 {
		maxTurns = AutoBattleTurnCap
	}
	before := func(us []*Combatant) int {
		n := 0
		for _, u := range us {
			n += int(u.Force())
		}
		return n
	}
	atk0, def0 := before(s.Attacker), before(s.Defender)

	var out BattleOutcome
	for turn := 1; turn <= maxTurns; turn++ {
		out.Turns = turn
		s.BeginTurn()
		actedA := s.advance(s.Attacker, s.Defender)
		actedD := s.advance(s.Defender, s.Attacker)
		s.Sweep()
		if over, won := s.Over(); over {
			out.AttackerWon, out.Decided = won, true
			break
		}
		s.EndTurn()
		if !actedA && !actedD {
			// 雙方都動不了（機動力耗盡且沒接觸）——僵局，不算分出勝負。
			break
		}
	}
	out.AttackerLoss = atk0 - before(s.Attacker)
	out.DefenderLoss = def0 - before(s.Defender)
	return out
}

// advance 讓 side 的每個單位朝最近的敵人推進，接觸就交戰。
func (s *BattleSim) advance(side, foes []*Combatant) bool {
	acted := false
	for _, u := range side {
		if !u.Alive() || !u.Cell.Valid() {
			continue
		}
		target := nearestLiving(u, foes)
		if target == nil {
			continue
		}
		if Adjacent(u.Cell, target.Cell) {
			if _, _, err := s.Engage(u, target); err == nil {
				acted = true
			}
			continue
		}
		// 一個回合可以連走好幾格，走到機動力用完或接觸為止。
		for s.stepToward(u, target) {
			acted = true
			if Adjacent(u.Cell, target.Cell) {
				break // 接觸就停，交戰留到下一回合
			}
		}
	}
	return acted
}

// stepToward 往目標的方向走一格，成功回 true。
func (s *BattleSim) stepToward(u, target *Combatant) bool {
	best, bestDist := HexDir(0), 1<<30
	tx, ty := target.Cell.ScreenXY()
	for d := DirLowerLeft; d <= DirUpperRight; d++ {
		n, ok := u.Cell.Neighbour(d)
		if !ok {
			continue
		}
		nx, ny := n.ScreenXY()
		if dist := absInt(nx-tx) + absInt(ny-ty); dist < bestDist {
			best, bestDist = d, dist
		}
	}
	if best == 0 {
		return false
	}
	_, err := s.Move(u.General, best)
	return err == nil
}

// nearestLiving 找還活著的最近敵人，距離用螢幕座標的曼哈頓距離。
//
// 用螢幕座標而不是格編號，是因為六角格的編號差不反映實際距離
// （奇數欄下移半格，`docs/re/07` §3）。
func nearestLiving(u *Combatant, list []*Combatant) *Combatant {
	var best *Combatant
	bestDist := 1 << 30
	ax, ay := u.Cell.ScreenXY()
	for _, o := range list {
		if !o.Alive() || !o.Cell.Valid() {
			continue
		}
		x, y := o.Cell.ScreenXY()
		if d := absInt(x-ax) + absInt(y-ay); d < bestDist {
			best, bestDist = o, d
		}
	}
	return best
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}
