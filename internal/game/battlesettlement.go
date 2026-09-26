package game

import "fmt"

// BattleSettlementKind 是原版戰鬥收尾的已證實分類。
//
// 這裡刻意不把「攻方獲勝」或「守方獲勝」拆成兩套寫回規則：
// `byte_64901` 為 1／2 時，原版都不走平局的 `sub_3964E`；勝負的具體來源
// （全滅、補給或必勝結算）由呼叫端另行保存。
type BattleSettlementKind uint8

const (
	// BattleSettlementUnknown 不是合法收尾，只作零值與錯誤回傳的安全狀態。
	BattleSettlementUnknown BattleSettlementKind = iota
	// BattleSettlementRetreat 對應已閉合的立即撤退樣本。
	BattleSettlementRetreat
	// BattleSettlementDecisive 表示原版已分出第一方或第二方勝負。
	BattleSettlementDecisive
	// BattleSettlementTurnLimitDraw 表示碰到回合上限但尚未分出勝負。
	BattleSettlementTurnLimitDraw
)

// BattleSettlement 是戰鬥收尾的窄資料契約。
//
// RequiresDT2Writeback 只描述原版控制流程是否會進入 `sub_3964E`；它不是
// 「目前 Go 程式已經能正確寫出所有 469 bytes」的保證。正常平局的欄位同步
// 尚未有第二份完成戰鬥的 DOSBox oracle，呼叫端仍必須保留未解 bytes。
type BattleSettlement struct {
	Kind                 BattleSettlementKind
	Winner               BattleSide
	Turn                 int
	TurnCap              int
	RequiresDT2Writeback bool
}

// BattleSettlementImmediateRetreat 建立立即撤退的已證實結果。
func BattleSettlementImmediateRetreat() BattleSettlement {
	return BattleSettlement{
		Kind:                 BattleSettlementRetreat,
		Winner:               BattleSideNone,
		RequiresDT2Writeback: false,
	}
}

// ApplyBattleSettlement 把一場已結束的 BattleSim 投影回策略層的已知欄位。
//
// 這是「remake 可完成玩家路徑」的窄入口，不宣稱重建原版所有戰後副作用：
//
//   - 戰鬥勝負已分出：依 sub_54DAC 清除交戰旗；攻方勝時由來源省司令
//     接管目標省，存活攻方將領移防到目標省。
//   - 回合上限平局：只清除交戰旗，不改變司令或將領所屬省。
//   - 立即撤退：不做任何投影。
//
// `generals` 與 `units` 是可選的共享切片；提供時只改寫已證實的 Province
// 欄位，未解的 General.Raw 仍由既有 writer 保留。呼叫端應以自己的 id 對照
// 表傳入相同索引的切片，避免把戰場順序誤當成 MAN 槽位。
func ApplyBattleSettlement(sim *BattleSim, table *ProvinceTable,
	generals []General, units []CombatUnit, settlement BattleSettlement) error {
	if sim == nil {
		return fmt.Errorf("game: nil 戰鬥不能套用結算")
	}
	if table == nil {
		return fmt.Errorf("game: 戰鬥結算需要省份表")
	}
	if !sim.At.Valid() || !sim.From.Valid() {
		return fmt.Errorf("game: 戰鬥省份無效：來源 %d、目標 %d", sim.From, sim.At)
	}
	target, err := table.At(sim.At)
	if err != nil {
		return err
	}
	source, err := table.At(sim.From)
	if err != nil {
		return err
	}

	switch settlement.Kind {
	case BattleSettlementRetreat:
		return nil
	case BattleSettlementTurnLimitDraw:
		// sub_54DAC 的 winner=0 路徑：只清除「正在打仗」旗標。
		target.Capture(0)
		return nil
	case BattleSettlementDecisive:
		// 下面只接受已分出的第一／第二方；零值不能被當成守方勝利。
		if settlement.Winner != BattleSideFirst && settlement.Winner != BattleSideSecond {
			return fmt.Errorf("game: 已分勝負卻沒有有效勝方：%d", settlement.Winner)
		}
		if settlement.Winner == BattleSideSecond {
			// 守方勝：省份不易主，但戰鬥旗必須清掉。
			target.Capture(0)
			return nil
		}

		// 攻方勝：原版的 +20 寫入值是來源勢力司令，而不是第一個
		// 戰場單位的陣列序號。這個值也能與 startBattle 的 Faction 對照。
		winner := source.Commander
		if !winner.Valid() {
			return fmt.Errorf("game: 攻方勝但來源省 %d 沒有司令", sim.From)
		}
		target.Capture(winner)
		for _, u := range sim.Attacker {
			if u == nil || !u.Alive() || u.General == 0 {
				continue
			}
			i := int(u.General) - 1
			if generals != nil {
				if i < 0 || i >= len(generals) {
					return fmt.Errorf("game: 攻方將領 %d 超出 %d 個將領槽位", u.General, len(generals))
				}
				generals[i].Province = sim.At
			}
			if units != nil {
				if i < 0 || i >= len(units) {
					return fmt.Errorf("game: 執行期攻方將領 %d 超出 %d 個槽位", u.General, len(units))
				}
				units[i].Province = sim.At
			}
		}
		return nil
	default:
		return fmt.Errorf("game: 未知戰鬥結算類型：%d", settlement.Kind)
	}
}

// ClassifyBattleSettlement 依已證實的勝方旗標與回合上限分類戰鬥收尾。
//
// `winner` 對應原版 `byte_64901`：0 代表尚未分勝負、1／2 代表第一／第二方。
// 只有「回合已達上限且 winner==0」會標成需要 `.DT2` 寫回的平局。中途尚未
// 結束時回錯，避免 UI 或自動戰鬥把進行中的狀態誤當成可結算結果。
func ClassifyBattleSettlement(turn, turnCap int, winner BattleSide) (BattleSettlement, error) {
	if turnCap <= 0 {
		return BattleSettlement{}, fmt.Errorf("game: 戰鬥回合上限必須為正數：%d", turnCap)
	}
	if winner != BattleSideNone && winner != BattleSideFirst && winner != BattleSideSecond {
		return BattleSettlement{}, fmt.Errorf("game: 未知戰鬥勝方值：%d", winner)
	}
	if winner != BattleSideNone {
		return BattleSettlement{
			Kind:                 BattleSettlementDecisive,
			Winner:               winner,
			Turn:                 turn,
			TurnCap:              turnCap,
			RequiresDT2Writeback: false,
		}, nil
	}
	if turn < turnCap {
		return BattleSettlement{}, fmt.Errorf("game: 戰鬥尚未達回合上限：第 %d／%d 回合", turn, turnCap)
	}
	return BattleSettlement{
		Kind:                 BattleSettlementTurnLimitDraw,
		Winner:               BattleSideNone,
		Turn:                 turn,
		TurnCap:              turnCap,
		RequiresDT2Writeback: true,
	}, nil
}

// String 只供日誌與測試使用；玩家可見文字仍應走語系資料。
func (k BattleSettlementKind) String() string {
	switch k {
	case BattleSettlementRetreat:
		return "immediate-retreat"
	case BattleSettlementDecisive:
		return "decisive"
	case BattleSettlementTurnLimitDraw:
		return "turn-limit-draw"
	default:
		return "unknown"
	}
}
