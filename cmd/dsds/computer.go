package main

import (
	"fmt"
	"sort"

	"github.com/wicanr2/great-era-remake/internal/game"
)

// 電腦回合：對每個非玩家勢力的每個省跑決策鏈 A（`docs/re/28`）。
//
// 原版的回合結構是 `sub_1ACCC`：建清單 → 內政 → 遷都 → 兩條決策鏈 →
// 迴圈直到命令數用完。這裡做的是**決策鏈 A 那一段**，因為它是唯一會
// 出兵的地方，也是玩家最有感的部分。
//
// ⚠️ **還沒接的**：`sub_19B89`、收尾 `sub_1AC01`，以及內政裡「只看到扣錢
// 不知道換到什麼」的那幾條（1／8／12 月的黃金支出）。
// 照實標記，不假裝跑了完整的回合。

// computerTurnReport 是一次電腦回合的結果，用來組訊息。
type computerTurnReport struct {
	comforts    int
	transfers   int
	attacks     []string
	relocations []string
	recruits    []string
}

// runComputerTurn 讓所有非玩家勢力行動一輪。
//
// 每個省的行動次數上限用原版的命令數公式（將領數 ÷ 8 + 1，`sub_13D23`），
// 決策鏈本身回 `AINone` 就提早收工。
func (a *app) runComputerTurn(playerFaction game.GeneralID) computerTurnReport {
	var rep computerTurnReport

	// 依省編號跑，順序固定——原版是照它自己的清單，我們沒有那份清單，
	// 但**固定順序**至少讓行為可重現（`CLAUDE.md` §9 截圖驗收要可重現）。
	for p := game.ProvinceID(1); p <= game.ProvinceCount; p++ {
		prov, err := a.tbl.At(p)
		if err != nil || prov.Commander == 0 || prov.Commander == playerFaction {
			continue
		}
		budget := a.world.CommandsFor(p)
		for i := 0; i < budget; i++ {
			if !a.runOneComputerAction(p, &rep) {
				break
			}
		}
	}
	sort.Strings(rep.attacks)
	return rep
}

// runOneComputerAction 跑一次決策鏈並執行結果，回報有沒有真的做到事。
func (a *app) runOneComputerAction(p game.ProvinceID, rep *computerTurnReport) bool {
	prov, err := a.tbl.At(p)
	if err != nil {
		return false
	}
	opt := game.ChainAOpts{
		// ⚠️ 難度旗標對應 `byte_6FFCA`。原版由誰寫還沒追到
		// （`CONTEXT.md` §6d），這裡先固定成「高難度 + 啟用後面幾步」，
		// 因為那是規則解得最完整的一組。
		Hard:            true,
		EnableStep5:     true,
		EnableLastThree: true,
		FirstStage:      false,
		// TotalForce 就是該省的兵力總和——`sub_13D23` 呼叫 `sub_306CF`
		// 存進 `[-236h]:[-234h]`（`docs/re/13` §3，confirmed）。
		// 這裡用同一支 `ProvinceForceTotal`，**不是代入值**。
		TotalForce: a.world.ProvinceForceTotal(p, a.generals),
	}
	// 原版的順序是 `sub_19B89` → 遷都 → 內政 → 兩條決策鏈（`docs/re/13` §6）。
	//
	// 遷都會消耗一個命令數。
	if r := a.world.Relocate(p, prov.Commander, a.generals); r.Moved {
		rep.relocations = append(rep.relocations,
			fmt.Sprintf("%s → %s 遷都", a.provinceName(r.From), a.provinceName(r.To)))
		return true
	}

	// 季節性內政（`docs/re/32`）：徵兵在 1／3／8／11 月，忠誠度在 6 月。
	//
	// ⚠️ **內政不消耗命令數**——原版那一段在主迴圈裡、在命令數檢查之前，
	// 沒有 `dec 狀態[-232h]`。所以它每一輪都會跑，這裡照抄。
	if in := a.world.RunInterior(p, a.month, a.generals, 0); in.Recruited > 0 {
		rep.recruits = append(rep.recruits,
			fmt.Sprintf("%s 徵兵 %d（%d 金）",
				a.provinceName(p), in.Recruited, in.GoldSpent))
	}

	res := a.world.ChainA(p, a.generals, opt)
	if res.Action.Kind == game.AINone {
		// 決策鏈 A 沒做出決定 → 換決策鏈 B（`sub_15F3C`，六步全是調動）。
		// 原版的順序就是 A 在前、B 在後（`docs/re/13` §6）。
		//
		// ⚠️ 原版兩條鏈都還有一道 `狀態[-6] == 0` 的 gate
		// （`sub_5C7FE`，疑似「是不是十大勢力」）。**沒有實作**——
		// `docs/re/13` §6 自己標了「這讀起來很怪（主要軍閥反而不做決策？），
		// `sub_5C7FE` 的語意可能需要重新驗」。照抄一個可疑的 gate
		// 會讓大部分省不動，那比漏掉它更難發現。
		res = game.ChainAResult{Action: a.world.Decide(p)}
	}

	switch res.Action.Kind {
	case game.AIComfort:
		if _, err := a.world.Comfort(p, comfortPolitics(a, prov)); err != nil {
			return false
		}
		rep.comforts++
		return true

	case game.AITransfer:
		if len(a.world.ApplyAction(res.Action).Moved) == 0 {
			return false // 決策說要調、執行挑不到人 → 旗標沒立，換下一步
		}
		rep.transfers++
		return true

	case game.AIAttack:
		out, err := a.world.ResolveAttack(res.Action.From, res.Action.To)
		if err != nil {
			return false
		}
		rep.attacks = append(rep.attacks,
			fmt.Sprintf("%s 攻 %s（%s）",
				a.provinceName(res.Action.From), a.provinceName(res.Action.To),
				outcomeWord(out)))
		return true
	}
	return false
}

// comfortPolitics 取該省司令的政治力（`Comfort` 要它算效果）。
func comfortPolitics(a *app, prov *game.Province) uint8 {
	i := int(prov.Commander) - 1
	if i < 0 || i >= len(a.generals) {
		return 0
	}
	return a.generals[i].AbilityB
}

// outcomeWord 把戰鬥結果講成一句話。
func outcomeWord(out game.BattleOutcome) string {
	switch {
	case !out.Decided:
		return "未分勝負"
	case out.AttackerWon:
		return "攻方勝，易主"
	default:
		return "守方守住"
	}
}

// provinceName 取省名。語系表沒載到就退回「省 N」——
// **看得出是哪一省，也看得出語系表沒生效**。
func (a *app) provinceName(p game.ProvinceID) string {
	if a.loc != nil {
		return a.loc.Province(int(p))
	}
	return fmt.Sprintf("省 %d", p)
}
