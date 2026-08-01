package main

import (
	"fmt"

	"github.com/wicanr2/great-era-remake/internal/game"
)

// 玩家指令與指令數上限的接點。規則本體在 `internal/game/budget.go`
// ——那裡不依賴 Ebiten，才測得了（`CLAUDE.md` §11）。
//
// ⚠️ **哪些指令消耗指令數還沒全部確認**。已知的兩條：
//
//	`sub_1398D`（遷都）消耗一次      `docs/re/13` §4
//	`sub_174C9`（出兵）消耗一次      `docs/re/13` §2
//
// 實機還看到「查閱之後指令數沒減少」（`docs/playtest/02` §282，等於取消）。
// 所以這裡的規則是**做成事才扣**，沒做成不扣——那是目前證據支持的
// 最保守讀法，不是猜的規則。

// withBudget 包一個政略指令：先檢查指令數，做成事才扣。
func (a *app) withBudget(p game.ProvinceID, run func() string) string {
	if a.cmdBudget.Remaining(p) <= 0 {
		return fmt.Sprintf("%s 這個月的指令數用完了", a.provinceName(p))
	}
	msg := run()
	if msg == "" {
		return "" // 沒做成事就不扣，也不報告
	}
	a.cmdBudget.Spend(p)
	return fmt.Sprintf("%s（剩 %d）", msg, a.cmdBudget.Remaining(p))
}
