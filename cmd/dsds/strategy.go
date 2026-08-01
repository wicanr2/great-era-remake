package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/wicanr2/great-era-remake/internal/game"
)

// 政略指令的執行層：把 `internal/game` 已解出的規則接到介面上。
//
// **只接規則層已經 confirmed 且有測試的指令**——公式沒讀出來的
// （運補、商業、外交、秘密行動）一律不接，按 F 鍵沒反應勝過假裝有效果。
// 各指令的來歷見 `docs/mechanics/10-political.md` §3。

// buildWorld 從省份表與將領表組出規則層要的 AIWorld。
//
// 效忠對象推導自「所屬省的司令」——原版的將領 `+14` 是執行期才填的欄位，
// 檔案裡沒有（`internal/game/ai_test.go` 的 realWorld 用同一套推導）。
func buildWorld(tbl *game.ProvinceTable, gens []game.General) *game.AIWorld {
	w := &game.AIWorld{Table: tbl, Opts: game.StrengthOpts{Stage: 1}, EnableExtra: true}
	for i := range gens {
		g := &gens[i]
		var faction game.GeneralID
		if p, err := tbl.At(g.Province); err == nil {
			faction = p.Commander
		}
		w.Units = append(w.Units, game.CombatUnit{
			General:  game.GeneralID(i + 1),
			Active:   true,
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
// ⚠️ 規則層的將領體力／士氣／忠誠度還沒收攏成一個容器，
// `Comfort` 只回報「影響了誰」，所以這裡自己套用體力那一項，
// 士氣與將領忠誠度**尚未接上**——照實顯示，不假裝做完了。
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
	for _, i := range res.Generals {
		if i < 0 || i >= len(a.generals) {
			continue
		}
		g := &a.generals[i]
		if v := int(g.Stamina) + game.ComfortStaminaGain; v > game.StaminaMax {
			g.Stamina = game.StaminaMax
		} else {
			g.Stamina = uint8(v)
		}
	}
	return fmt.Sprintf("慰勞了 %d 位將領，人民忠誠度 +%d（士氣尚未接上）",
		len(res.Generals), res.ProvinceLoyaltyGain)
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

// endTurn 推進一個月：先讓電腦行動，再推進時間。跨年跑年度結算（`docs/re/25`）。
//
// ⚠️ 原版的回合結構是「逐省下指令、每省有指令數上限」（`docs/re/13`），
// 玩家這一側還沒接上指令數上限——**照實標記**。
// 電腦那一側已經照命令數公式跑（`cmd/dsds/computer.go`）。
func (a *app) endTurn() string {
	// 電腦回合：決策鏈 A（`docs/re/28`）。玩家勢力是目前所在省的司令。
	var ai string
	if prov, err := a.tbl.At(a.current); err == nil && prov.Commander != 0 {
		ai = describeComputerTurn(a.runComputerTurn(prov.Commander))
	}

	a.month++
	if a.month <= 12 {
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
	st := &game.GameState{Stage: 1, Year: uint8(a.year), Month: a.month}
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

// report 記錄指令結果。畫面上還沒有訊息列，先印到 stderr——
// **面板的黃金／糧食／忠誠度會即時更新**，效果本來就看得見。
func (a *app) report(msg string) {
	if msg == "" {
		return
	}
	a.msg, a.dirty = msg, true
	fmt.Fprintln(os.Stderr, msg)
}
