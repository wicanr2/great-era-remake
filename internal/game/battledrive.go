package game

// 把決策鏈（`battleai.go`）接上真實的戰鬥狀態（`BattleSim`）。
//
// ⚠️ **這一層只做「決定」，不做「執行」。**
//
// 13 種行動的語意都解出來了（`docs/re/31` §41），但真正實作出來的只有
// 少數幾種（值 3 的挑城市在 `battlecity.go`、值 2 的佈防規則在 §30）。
// 在全部補齊之前，**不要拿決策鏈去取代 `AutoResolve`**——那會用一個
// 「只會選、不會做」的 AI 換掉一個能把戰鬥跑完的近似版，是退步。
//
// 所以這一層的用途是**觀測**：讓決策鏈在真實戰鬥狀態上跑，
// 看它每回合選什麼，跟原版的預期對照。等行動都實作完再換掉推進器。

// BattleTurnDecision 是某一回合兩條鏈各自的決定。
type BattleTurnDecision struct {
	// Turn 是回合序號，從 1 起。
	Turn int
	// A 是分支 A（第二方）的決定，B 是分支 B（第一方）的決定。
	A, B BattleDecision
}

// BattleChainGates 是決策鏈那幾個**還沒解出來源**的判斷。
//
// 每一個都對應 `docs/re/31` 裡一條標「未讀」的底層函式。
// 由呼叫端提供，而不是在這裡瞎猜——來源解出來之後，
// 把算法補進 `BattleSim.chainGates` 就好，決策鏈本身不用動。
type BattleChainGates struct {
	// RatioSelf / RatioFoe 對應 §42 的比率門檻（`sub_3A63C` and `sub_3A730`）。
	// 來源是 `word_64932/34/36/38` 與 `sub_3A4CE`，**未解**。
	RatioSelf, RatioFoe bool
	// Deploy 對應 §43 的 `word_6493A == 0`，分支 B 值 2 獨有的閘門。**未解**。
	Deploy bool
	// Sub53619 對應 §45／§47 的 `sub_53619`——**兩處「必勝結算」都要它為 false**。
	//
	// ⭐ 語意已解：它是「**沒有**可用的我方鄰省支援」的反相判斷。
	// 用 `!HasBattleSupport(tbl, at, leader)` 算（`battlesupport.go`）。
	Sub53619 bool
	// EnableLastSteps 是 `byte_6FFCA & 4`，這個**有解**（難度／階段旗標）。
	EnableLastSteps bool
}

// SideStrength 回傳某一方所有存活單位的攻擊力總和。
//
// 對應原版 `sub_3A320` 裡那兩個迴圈（掃 `0x750` 或 `0x764`，
// 逐個累加 `sub_5A0B9`）。
func (s *BattleSim) SideStrength(units []*Combatant) int {
	n := 0
	for _, u := range units {
		if u == nil || !u.Alive() {
			continue
		}
		n += s.StrengthOf(u)
	}
	return n
}

// LeaderOnField 回答 §44 的 `sub_56D49`：**某個勢力領袖本人在不在這批單位裡**。
//
// 原版掃該方的 10 個槽位，看有沒有一個就是領袖的將領 ID。
// 領袖是誰由呼叫端給——分支 A 用的是當前交戰省的司令
// （省份記錄 `+20`），分支 B 用的是首位單位效忠的領袖（`+14`）。
func LeaderOnField(units []*Combatant, leader GeneralID) bool {
	if leader == 0 {
		return false
	}
	for _, u := range units {
		if u != nil && u.Alive() && u.General == leader {
			return true
		}
	}
	return false
}

// DecideTurn 跑一次兩條決策鏈，回傳這一回合雙方各自選了什麼行動。
//
//	defenderLeader  當前交戰省的司令（分支 A 問它在不在守方隊伍裡）
//	attackerLeader  攻方首位單位效忠的勢力領袖（分支 B 用）
//
// ⚠️ **兩條鏈各管一方**（§32）：分支 A 管第二方、分支 B 管第一方。
// 這裡把守方對到第二方、攻方對到第一方——
// **那個對應本身還沒驗**（§2 的 `byte_64901` 對照仍未解），
// 所以這是一個明確標記的假設，不是結論。
func (s *BattleSim) DecideTurn(turn int, g BattleChainGates,
	defenderLeader, attackerLeader GeneralID) BattleTurnDecision {
	atk := s.SideStrength(s.Attacker)
	def := s.SideStrength(s.Defender)

	// 分支 A（第二方 ≈ 守方）：「我方」是守方。
	inA := BattleAIInput{
		SideStrength:     def,
		FoeStrength:      atk,
		RatioGateSelf:    g.RatioSelf,
		RatioGateFoe:     g.RatioFoe,
		FoeLeaderOnField: LeaderOnField(s.Defender, defenderLeader),
		Sub53619:         g.Sub53619,
		EnableLastSteps:  g.EnableLastSteps,
	}
	// 分支 B（第一方 ≈ 攻方）：「我方」是攻方。
	inB := BattleAIInput{
		SideStrength:     atk,
		FoeStrength:      def,
		RatioGateSelf:    g.RatioSelf,
		RatioGateFoe:     g.RatioFoe,
		DeployGateOpen:   g.Deploy,
		FoeLeaderOnField: LeaderOnField(s.Defender, defenderLeader),
		Sub53619:         g.Sub53619,
		EnableLastSteps:  g.EnableLastSteps,
	}
	_ = attackerLeader // 分支 B 的領袖判斷要等 sub_534FF 解出來才用得上

	return BattleTurnDecision{Turn: turn, A: DecideBattleA(inA), B: DecideBattleB(inB)}
}

// TraceDecisions 把一場戰鬥從頭跑到尾，記錄**每一回合決策鏈選了什麼**。
//
// 推進仍然用 `AutoResolve` 那一套簡化行為（朝最近的敵人走），
// 決策鏈只在旁邊觀測——所以這支不會改變戰鬥結果，可以安全地
// 拿去跟現有的驗收截圖對照。
//
// 用途：看決策鏈在真實局勢下的行為分佈。例如「一場一面倒的戰鬥裡，
// 必勝結算是不是真的會在前幾回合就觸發」。
func (s *BattleSim) TraceDecisions(maxTurns int, g BattleChainGates,
	defenderLeader, attackerLeader GeneralID) ([]BattleTurnDecision, BattleOutcome) {
	if maxTurns <= 0 {
		maxTurns = AutoBattleTurnCap
	}
	var trace []BattleTurnDecision

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
		trace = append(trace, s.DecideTurn(turn, g, defenderLeader, attackerLeader))

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
			break
		}
	}
	out.AttackerLoss = atk0 - before(s.Attacker)
	out.DefenderLoss = def0 - before(s.Defender)
	return trace, out
}
