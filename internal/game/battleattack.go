package game

import "fmt"

// RepeatedAttackPasses 是 WAR.EXE 的 sub_53428 所見的重複正規攻擊次數。
//
// 這不是「六種攻擊」的選單編號：sub_4D585 的 1..6 是方向／目標輸入，
// sub_4BF27 之後才依狀態分派到數個內部 handler；第二層 1..5 的直接分支已由 IDA
// 證實，第 6 鍵與 handler 完整語意尚未由正常玩家 oracle 確認。sub_53428 的窄切片只
// 證實會連續呼叫五次 sub_51D68；其餘動畫、音效與呼叫時機仍見
// docs/re/41-block10-attack-dispatch.md。
const RepeatedAttackPasses = 5

// EngageRepeated 執行已證實的「五次正規戰損」切片（sub_53428）。
//
// 每次都沿用 ResolveBattleAttack 的相鄰檢查、戰力公式與戰損 writer，因此不複製另一份
// Casualties 規則。回傳的是五次累計的損失；若某一方在中途歸零，後續 pass
// 仍會走同一條純規則函式，但其損失會被兵力上限夾住。這保留原版「呼叫五次」
// 的可觀察邊界，同時不假定尚未解出的畫面／音效副作用。
//
// 證據等級：strong inference。IDA 已閉合 sub_53428 的五次呼叫與雙方損失
// 寫回，但 sub_51D68 周邊 helper 的所有 side effect、以及玩家何時選到這條
// handler，尚未由正常玩家 oracle 完整確認。
func (s *BattleSim) EngageRepeated(attacker, target *Combatant) (lossAttacker, lossTarget int, err error) {
	if s == nil || attacker == nil || target == nil {
		return 0, 0, fmt.Errorf("game: 重複交戰雙方不得為 nil")
	}
	if !attacker.Cell.Valid() || !target.Cell.Valid() || !Adjacent(attacker.Cell, target.Cell) {
		return 0, 0, fmt.Errorf("game: 重複交戰需要相鄰且有效的戰場格")
	}
	for pass := 0; pass < RepeatedAttackPasses; pass++ {
		result, callErr := s.ResolveBattleAttack(attacker, target)
		if callErr != nil {
			return lossAttacker, lossTarget, callErr
		}
		lossAttacker += result.LossAttacker
		lossTarget += result.LossTarget
	}
	return lossAttacker, lossTarget, nil
}
