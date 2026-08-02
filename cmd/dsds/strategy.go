package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
)

func (a *app) recruitWording() (render.RecruitWording, error) {
	keys := []string{"recruit.action", "recruit.reorganize", "recruit.infantry", "recruit.armour", "recruit.artillery", "recruit.cavalry", "recruit.amount", "recruit.limit", "recruit.cost", "recruit.gold", "recruit.confirm", "recruit.remaining", "recruit.general", "recruit.force"}
	values := make([]string, len(keys))
	for i, key := range keys {
		value, err := a.wordingText(key)
		if err != nil {
			return render.RecruitWording{}, err
		}
		values[i] = value
	}
	return render.RecruitWording{Recruit: values[0], Reorganize: values[1], Infantry: values[2], Armour: values[3], Artillery: values[4], Cavalry: values[5], Amount: values[6], Limit: values[7], Cost: values[8], Gold: values[9], Confirm: values[10], Remaining: values[11], General: values[12], Force: values[13]}, nil
}

// 政略指令的執行層：把 `internal/game` 已解出的規則接到介面上。
//
// **只接規則層已經 confirmed 且有測試的指令**——公式沒讀出來的
// （外交、停火）一律不接，按鍵沒反應勝過假裝有效果。
// 政策的授權自治與產能分配均已接通；後者對應見 docs/re/26。
// 各指令的來歷見 `docs/mechanics/10-political.md` §3。

// buildWorld 從省份表與將領表組出規則層要的 AIWorld。
//
// 效忠對象推導自「所屬省的司令」——原版的將領 `+14` 是執行期才填的欄位，
// 檔案裡沒有（`internal/game/ai_test.go` 的 realWorld 用同一套推導）。
//
// `factions` 是存檔裡的勢力表（`docs/formats/07` §4）。傳 nil 表示沒有——
// ⚠️ **那不是無害的**：`MajorPower` 為 nil 時十大勢力的戰力加成不會生效
// （`internal/game/strength.go`），勝負判定也會保守地一律回「還沒結束」。
// 在解出勢力表之前這個欄位只能是 nil，現在有資料源了。
func buildWorld(tbl *game.ProvinceTable, gens []game.General, factions *game.FactionTable, stage uint8) *game.AIWorld {
	opts := game.StrengthOpts{Stage: stage}
	if factions != nil {
		opts.MajorPower = factions.MajorPower
	}
	w := &game.AIWorld{Table: tbl, Opts: opts, EnableExtra: true}
	for i := range gens {
		g := &gens[i]
		var faction game.GeneralID
		if p, err := tbl.At(g.Province); err == nil {
			faction = p.Commander
		}
		w.Units = append(w.Units, game.CombatUnit{
			General:  game.GeneralID(i + 1),
			Active:   g.Raw[16]&1 == 1,
			Cell:     game.NoCell,
			Faction:  faction,
			Province: g.Province,
		})
		w.Strengths = append(w.Strengths, game.StrengthInput{
			Ability: g.AbilityA, Force: g.Force, Branch: g.Branch,
			F19: g.F19, F20: g.F20, F29: g.Stamina, F30: g.F30,
			General: game.GeneralID(i + 1), Faction: faction,
		})
	}
	return w
}

// firstGeneralOf 回傳該省的第一位將領（索引與 generals 對齊），沒有回 -1。
//
// 原版的開發指令會問「司令欲派遣何將？」，玩家自己挑。
// remake 目前自動挑第一位——**這是 remake 差異**，等派將 UI 做好再換掉。
func (a *app) firstGeneralOf(p game.ProvinceID) int {
	for i := range a.generals {
		if a.generals[i].Province == p {
			return i
		}
	}
	return -1
}

// execTax 是徵稅（指令 4）。公式與兩段忠誠度扣減見 `docs/re/18`。
func (a *app) execTax() string {
	prov, err := a.tbl.At(a.current)
	if err != nil {
		return err.Error()
	}
	if prov.Taxed() {
		return "本月已徵過稅"
	}
	res, err := a.world.Tax(a.current, a.rng)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("共徵得黃金 %d、糧食 %d，人民忠誠度 %d",
		res.Gold, res.Food, res.LoyaltyAfter)
}

// execComfort 是慰勞軍民（指令 14）。四件事見 `docs/re/22`。
//
// 規則層的將領體力／士氣／忠誠度目前分散在 `General` 與
// `AIWorld.Strengths`；`Comfort` 回報受影響索引後，由
// `applyComfortGenerals` 同步兩份執行期狀態。
func (a *app) execComfort() string {
	leader, err := a.tbl.At(a.current)
	if err != nil {
		return err.Error()
	}
	var politics uint8
	if i := int(leader.Commander) - 1; i >= 0 && i < len(a.generals) {
		politics = a.generals[i].AbilityC
	}
	res, err := a.world.Comfort(a.current, politics)
	if err != nil {
		return err.Error()
	}
	n := game.ApplyComfortGenerals(a.generals, a.world.Strengths, res.Generals)
	return fmt.Sprintf("慰勞了 %d 位將領，人民忠誠度 +%d",
		n, res.ProvinceLoyaltyGain)
}

// execDevelop 是開發（指令 7）的三個子項。
//
// `sub` 是子項編號，照原版選單：1 墾地、2 建兵工廠、3 挖金礦。
func (a *app) execDevelop(sub int) string {
	gi := a.firstGeneralOf(a.current)
	if gi < 0 && sub != 2 {
		return "本省沒有將領可派遣"
	}

	switch sub {
	case 1: // 墾地
		g := &a.generals[gi]
		gain, stam, err := a.world.ReclaimLand(a.current, g.AbilityC, g.Stamina)
		if err != nil {
			return err.Error()
		}
		g.Stamina = stam
		prov, _ := a.tbl.At(a.current)
		return fmt.Sprintf("地價提昇為 %d（+%d），體力剩 %d", prov.LandValue, gain, stam)

	case 2: // 建兵工廠
		built, err := a.world.BuildArsenal(a.current, 1)
		if err != nil {
			return err.Error()
		}
		prov, _ := a.tbl.At(a.current)
		return fmt.Sprintf("建了 %d 座，目前兵工廠有 %d 座", built, prov.Arsenals)

	case 3: // 挖金礦
		g := &a.generals[gi]
		// ⚠️ 礦藏表（`ds:11h` 起，以省編號為索引）還沒從執行檔抽出來，
		// 這裡暫時用固定值。**這是已知的偏差，不是解出來的規則**
		// （`docs/mechanics/40-economy.md` §7b）。
		const depositPlaceholder = 5
		got, stam, err := a.world.MineGold(a.current, depositPlaceholder,
			g.AbilityC, g.Stamina, a.rng)
		if err != nil {
			return err.Error()
		}
		g.Stamina = stam
		return fmt.Sprintf("共得黃金 %d，體力剩 %d（礦藏用暫定值）", got, stam)
	}
	return ""
}

// execTrade 執行商業活動（指令 12）。匯率與可買賣品項見
// `internal/game/trade.go`／`docs/mechanics/40-economy.md` §8。
func (a *app) execTrade(importing bool, good game.TradeGood, amount int) string {
	if amount <= 0 {
		return ""
	}
	if importing {
		res, err := a.world.Import(a.current, good, amount)
		if err != nil {
			return err.Error()
		}
		return fmt.Sprintf("購買%s %d，共用黃金 %d",
			game.TradeGoodName(good), res.Amount, res.Gold)
	}
	res, err := a.world.Export(a.current, good, amount)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("拋售%s %d，共得黃金 %d",
		game.TradeGoodName(good), res.Amount, res.Gold)
}

func (a *app) supplyLimit(good int) int {
	p, err := a.tbl.At(a.current)
	if err != nil {
		return 0
	}
	switch good {
	case 0:
		return int(p.Gold)
	case 1:
		return int(p.Food)
	case 2:
		return int(p.Ammo)
	case 3:
		return int(p.Fuel)
	}
	return 0
}

func (a *app) transferSupplyLimit(good int) int {
	return a.supplyLimit(good)
}

// execPlayerTransfer 套用已確認的玩家調動；只有成功才消耗指令數。
func (a *app) execPlayerTransfer() string {
	if a.transferSession == nil {
		return "沒有可確認的調動"
	}
	if a.cmdBudget.Remaining(a.current) <= 0 {
		return fmt.Sprintf("%s 這個月的指令數用完了", a.provinceName(a.current))
	}
	from, to := a.transferSession.From(), a.transferSession.To()
	rep, err := a.world.ApplyPlayerTransfer(a.transferSession, a.generals, a.transferAmounts)
	if err != nil {
		return err.Error()
	}
	a.cmdBudget.Spend(from)
	return fmt.Sprintf("調動 %d 位將領至%s；黃金 %d、糧食 %d、彈藥 %d、燃料 %d（剩 %d）",
		len(rep.Moved), a.provinceName(to), rep.Supplies[0], rep.Supplies[1],
		rep.Supplies[2], rep.Supplies[3], a.cmdBudget.Remaining(from))
}

// execSupply 執行指令 3 的四項運補。全 0 視為取消，不消耗指令數。
func (a *app) execSupply(target game.ProvinceID, amounts [4]int) string {
	if amounts == [4]int{} {
		return ""
	}
	moved, err := a.world.SupplyAll(a.current, target, amounts)
	if err != nil {
		return err.Error()
	}
	return fmt.Sprintf("運補至%s：黃金 %d、糧食 %d、彈藥 %d、燃料 %d",
		a.provinceName(target), moved[0], moved[1], moved[2], moved[3])
}

// execRecruit 執行指令 5 的徵兵分支，並把規則層的兵力同步回可存檔的將領表。
func (a *app) execRecruit(branch uint8, amount int) string {
	added, err := a.world.Recruit(a.current, branch, amount)
	if err != nil {
		return err.Error()
	}
	if added == 0 {
		return ""
	}
	for i := range a.generals {
		if i < len(a.world.Strengths) {
			a.generals[i].Force = a.world.Strengths[i].Force
		}
	}
	return fmt.Sprintf("徵%s %d，共用黃金 %d",
		game.BranchName(branch), added, game.RecruitCost(branch, added))
}

// execTrain 執行指令 13「練兵」，並同步士兵戰技回可存檔的將領表。
func (a *app) execTrain() string {
	res, err := a.world.Train(a.current)
	if err != nil {
		return err.Error()
	}
	for i := range a.generals {
		if i < len(a.world.Strengths) {
			a.generals[i].F19 = a.world.Strengths[i].F19
		}
	}
	return fmt.Sprintf("練兵完成：%d 位將領，士兵戰技合計提升 %d", res.Generals, res.Gain)
}

// execStudentProtest 執行秘密行動的「鼓動學潮」。固定成本、20% 成功率與
// 三項 ×0.8 效果均已由 `sub_20447` 證實；成敗都消耗黃金與一個指令數。
func (a *app) execStudentProtest(target game.ProvinceID) string {
	res, err := a.world.IncitStudentProtest(a.current, target, a.generals, a.rng)
	if err != nil {
		return err.Error()
	}
	for _, i := range res.Demoralised {
		if i >= 0 && i < len(a.generals) && i < len(a.world.Strengths) {
			a.world.Strengths[i].F30 = a.generals[i].F30
		}
	}
	if res.Success {
		return fmt.Sprintf("%s發生學潮（影響 %d 位將領，花費黃金 %d）",
			a.provinceName(target), len(res.Demoralised), res.Cost)
	}
	return fmt.Sprintf("%s沒有學潮發生（花費黃金 %d）", a.provinceName(target), res.Cost)
}

func (a *app) syncReorganizationGenerals() {
	for i := range a.generals {
		if i < len(a.world.Strengths) {
			a.generals[i].Force = a.world.Strengths[i].Force
			a.generals[i].F20 = a.world.Strengths[i].F20
		}
	}
}

func (a *app) finishReorganization() {
	if a.reorganization != nil && a.reorganization.Changed() {
		a.report(a.withBudget(a.current, func() string { return "重新整編完成" }))
	}
	a.reorganization = nil
	a.reorganizeInput = 0
	a.screen, a.dirty = screenCommand, true
}

// endTurn 推進一個月：先讓電腦行動，再推進時間。跨年跑年度結算（`docs/re/25`）。
//
// 原版的回合結構是「逐省下指令、每省有指令數上限」（`docs/re/13`）。
// 玩家側已接上每省預算與面板剩餘值；目前仍待逐一確認的是哪些指令會消耗額度。
// 電腦側也已照命令數公式跑（`cmd/dsds/computer.go`）。
func (a *app) endTurn() string {
	// 電腦回合：決策鏈 A（`docs/re/28`）。玩家勢力是目前所在省的司令。
	var ai string
	if prov, err := a.tbl.At(a.current); err == nil && prov.Commander != 0 {
		ai = describeComputerTurn(a.runComputerTurn(prov.Commander))
	}

	a.month++
	if a.month <= 12 {
		// 換月：指令數重新算（`docs/re/13` §2）。
		a.cmdBudget.Reset()
		// 跨月清徵稅旗標。原版在哪裡做還沒讀到，但「每月限一次」
		// 這條是 confirmed（`docs/re/18`），沒有月結算的話徵稅只能做一次。
		for p := game.ProvinceID(1); p <= game.ProvinceCount; p++ {
			if prov, err := a.tbl.At(p); err == nil {
				prov.Flags &^= game.ProvinceFlagTaxed
			}
		}
		return joinTurn(fmt.Sprintf("民國 %d 年 %d 月", a.year, a.month), ai)
	}

	a.month = 1
	a.cmdBudget.Reset()
	st := &game.GameState{Stage: 1, Year: uint8(a.year), Month: a.month,
		MajorPower: a.world.Opts.MajorPower}
	rep := a.world.AdvanceYear(st, a.generals, 0, nil)
	a.year = uint16(st.Year)
	return joinTurn(fmt.Sprintf("民國 %d 年 1 月：%d 省人口成長共 %d 人",
		a.year, rep.Grown, rep.Growth), ai)
}

// joinTurn 把時間訊息與電腦回合的訊息接起來。
func joinTurn(when, ai string) string {
	if ai == "" {
		return when
	}
	return when + "｜" + ai
}

// describeComputerTurn 把電腦回合的結果講成一句話。
//
// 一件事都沒做時回空字串——**不要印「電腦無動作」**，那會讓
// 「規則還沒接」與「規則接了但這回合沒觸發」看起來一樣。
func describeComputerTurn(rep computerTurnReport) string {
	var parts []string
	parts = append(parts, rep.relocations...)
	parts = append(parts, rep.recruits...)
	if rep.comforts > 0 {
		parts = append(parts, fmt.Sprintf("慰勞 %d", rep.comforts))
	}
	if rep.transfers > 0 {
		parts = append(parts, fmt.Sprintf("調動 %d", rep.transfers))
	}
	parts = append(parts, rep.attacks...)
	if len(parts) == 0 {
		return ""
	}
	return "電腦：" + strings.Join(parts, "、")
}

// report 將指令結果排入畫面訊息佇列，並保留 stderr 作可重現診斷。
func (a *app) report(msg string) {
	if msg == "" {
		return
	}
	a.msg, a.dirty = msg, true
	// 完整字庫載入失敗時仍保留原本的 stderr 診斷，不可排入一則
	// 畫不出來卻又會擋住輸入的隱形訊息。
	if a.eten != nil {
		if a.messages == nil {
			a.messages = newMessageQueue(a.preferences.MessageTime)
		}
		a.messages.Push(msg)
	}
	fmt.Fprintln(os.Stderr, msg)
}
