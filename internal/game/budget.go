package game

import "fmt"

// 每個省的**指令數上限**（`docs/re/13` §2）。
//
// 原版：`sub_13D23` 算 `將領數 ÷ 8 + 1` 存進 `狀態[-232h]`，
// 主迴圈跑到它歸零才換下一個省。實機畫面右下角就顯示剩餘指令數
// （`docs/playtest/02`，湖北開局是 2）。
//
// ⚠️ 這一層放在 `internal/game` 而不是 `cmd/`，理由是 `AGENTS.md` §8：
// **Ebiten 在 init 期就要求顯示器**，放進 `cmd/dsds` 的話連測都測不了
// （第一版就是放那裡，`go test` 直接 panic 在 GLFW 初始化）。

// CommandBudget 記每個省這個月還剩幾個指令。
//
// 沒有 entry 表示這個月還沒動過——第一次查詢時用 `CommandsFor` 現算。
type CommandBudget struct {
	left map[ProvinceID]int
	// commandsFor 是「這個省有幾個指令數」。獨立成欄位是為了可測；
	// 正常用 `NewCommandBudget` 接 `AIWorld.CommandsFor`。
	commandsFor func(ProvinceID) int
	// remainingAfter 套用執行一項後的原版特例；自治省會直接歸零。
	remainingAfter func(ProvinceID, int) int
}

// NewCommandBudget 用某個世界的將領分佈建一份指令數帳。
func NewCommandBudget(w *AIWorld) *CommandBudget {
	return &CommandBudget{commandsFor: w.CommandsFor, remainingAfter: w.CommandsRemainingAfter}
}

// NewCommandBudgetFunc 讓呼叫端自己決定上限怎麼算（測試用）。
func NewCommandBudgetFunc(f func(ProvinceID) int) *CommandBudget {
	return &CommandBudget{commandsFor: f}
}

// Peek reads a displayed budget without freezing a previously unvisited province.
func (b *CommandBudget) Peek(p ProvinceID) int {
	if n, ok := b.left[p]; ok {
		return n
	}
	if b.commandsFor != nil {
		return b.commandsFor(p)
	}
	return 0
}

// Remaining 回傳某省剩餘的指令數，並固定本月首次執行時的上限。
func (b *CommandBudget) Remaining(p ProvinceID) int {
	if b.left == nil {
		b.left = map[ProvinceID]int{}
	}
	if n, ok := b.left[p]; ok {
		return n
	}
	n := b.Peek(p)
	b.left[p] = n
	return n
}

// Spend 消耗一個指令數。回 false 表示已經用完，指令不該執行。
func (b *CommandBudget) Spend(p ProvinceID) bool {
	before := b.Remaining(p)
	if before <= 0 {
		return false
	}
	if b.remainingAfter != nil {
		b.left[p] = b.remainingAfter(p, before)
	} else {
		b.left[p] = before - 1
	}
	return true
}

// Reset 在換月時把所有省的指令數清掉，下次查詢會重新算。
//
// ⚠️ 原版在哪裡重置**還沒讀到**。`sub_13D23` 是電腦回合的第一步，
// 玩家那一側的對應位置未知。呼叫端目前放在月結算，因為徵稅的
// 「每月限一次」也在那裡（`docs/re/18`），兩者節奏一致——**標為推論**。
func (b *CommandBudget) Reset() {
	b.left = nil
}

// Snapshot 保留已查詢省份的剩餘數，不把尚未查詢的省份具體化。
func (b *CommandBudget) Snapshot() map[ProvinceID]int {
	out := make(map[ProvinceID]int, len(b.left))
	for p, n := range b.left {
		out[p] = n
	}
	return out
}

// Restore 在驗證後一次替換，保留原有的規則 callback。
func (b *CommandBudget) Restore(left map[ProvinceID]int, provinceLimit int) error {
	out := make(map[ProvinceID]int, len(left))
	for p, n := range left {
		if !p.Valid() || int(p) > provinceLimit || n < 0 || n > SaveGeneralCount/8+1 {
			return fmt.Errorf("game: 不合法的指令預算 %d/%d", p, n)
		}
		out[p] = n
	}
	b.left = out
	return nil
}
