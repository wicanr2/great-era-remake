package main

import (
	"fmt"
	"os"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
)

// 戰鬥畫面。規則全部來自 `internal/game`（`docs/re/07`／`08`／`09`），
// 這一層只做輸入與繪製。
//
// 原版的戰鬥主選單與五個分支已由 DOSBox／IDA 閉合（`docs/mechanics/30-combat.md`
// §3）：1 移動、2 攻擊、3 撤退、4 駐軍、5 查閱。方向鍵只在移動分支內使用；
// 目標／方向輸入的 1..6 不是六個有名稱的攻擊模式；原版另有第二層 "123456" 選單，
// 目前只把 1..5 的 handler 分支視為證據，第 6 鍵語意仍未知。
//
// 目前仍保留明確標記的 remake 差異：Enter（或相鄰敵軍的點擊）可直接嘗試一次
// 近身攻擊；攻擊子選單的六個選項以穩定目標順序選取目標。兵種 4 的非相鄰目標
// 已接到原版 `sub_58854`／`sub_57B15` 可證實的遠程戰損路徑；其彈藥消耗、動畫
// 與第二層六鍵選單的完整語意仍未冒充完成。撤退仍只有陝西 18 ← 河南 19 正常玩家樣本；
// 駐軍／查閱則提供可玩的外殼操作，不冒充原版尚未閉合的完整輸入分支。

// 六角格的方向編號在移動子狀態內直接對應數字鍵 1–6。
var dirKeys = [...]ebiten.Key{
	ebiten.Key1, ebiten.Key2, ebiten.Key3,
	ebiten.Key4, ebiten.Key5, ebiten.Key6,
}

// battleMode 是戰鬥畫面的輸入子狀態。主選單是 0；其餘值與
// `render.BattleMenuMode` 保持一一對應，供面板顯示目前正在等哪一層輸入。
type battleMode uint8

const (
	battleModeCommand battleMode = iota
	battleModeMove
	battleModeAttack
	battleModeRetreat
	battleModeGarrison
	battleModeInspect
)

// battleModeForCommand 是已由原版函式分派確認的五個主選單鍵。
// 回傳 false 代表不是戰鬥主選單命令，讓呼叫端不把方向鍵誤吃掉。
func battleModeForCommand(n int) (battleMode, bool) {
	if n < 1 || n > 5 {
		return battleModeCommand, false
	}
	return battleMode(n), true
}

func battleModeMessage(mode battleMode) string {
	switch mode {
	case battleModeMove:
		return "移動：請按 1–6 選擇方向"
	case battleModeAttack:
		return "攻擊：請按 1–6 選擇目標"
	case battleModeRetreat:
		return "撤退：請輸入可退省份編號"
	case battleModeGarrison:
		return "駐軍：選中單位前往第一個可達城市"
	case battleModeInspect:
		return "查閱：顯示選中單位摘要"
	default:
		return ""
	}
}

// commandNumberPressed 只讀戰鬥主選單的 1..5；移動／攻擊子狀態另由
// directionPressed／attackOptionPressed 讀取 1..6。
func commandNumberPressed() (int, bool) {
	keys := [...]ebiten.Key{ebiten.Key1, ebiten.Key2, ebiten.Key3, ebiten.Key4, ebiten.Key5}
	for i, key := range keys {
		if inpututil.IsKeyJustPressed(key) {
			return i + 1, true
		}
	}
	return 0, false
}

func attackOptionPressed() (int, bool) {
	for i, key := range dirKeys {
		if inpututil.IsKeyJustPressed(key) {
			return i + 1, true
		}
	}
	return 0, false
}

// 圖示由兵種決定（`render.BranchIcon`，`docs/formats/05` §3）：
// 步兵是鋼盔、裝甲兵是戰車、騎兵是馬頭、砲兵是大砲（六個朝向）。
// 綠是攻方、紅是守方。

var (
	cursorOurs   = assets.RGB{R: 0xFF, G: 0xFF, B: 0x55}
	cursorTarget = assets.RGB{R: 0xFF, G: 0x55, B: 0x55}
)

// battleState 是戰鬥畫面的介面狀態。規則狀態在 `sim` 裡。
type battleState struct {
	sim        *game.BattleSim
	sel        int    // 目前選中第幾個攻方單位
	log        string // 最近一次動作的結果，畫在面板上
	finished   bool
	settlement game.BattleSettlement // 已證實的收尾分類；未知時保持零值
	settled    bool                  // 世界層已套用結算；避免 F10／重複寫回重跑
	forceDirty bool                  // 戰鬥副本的已確認兵力尚未同步回將領表
	mode       battleMode            // 原版五項選單的目前輸入層
	// retreatInput／retreatTargets 服務原版樣本與省份表 fallback；來源標籤
	// 保留在 game.RetreatCandidates，不把 fallback 冒充跨戰鬥 oracle。
	retreatInput   uint32
	retreatTargets []game.ProvinceID

	// turn 是回合序號，決策鏈要用（每回合重跑一次）。
	turn    int
	turnCap int // 由開戰月份固定；二月 15，其餘月份 16（原版 byte_6FE7E）
	// aiLog 是守方 AI 這一回合做了什麼。⚠️ 目前只印到 stderr——
	// 畫面上畫不出自由組合的中文（原版字模是場景子集，見 `runDefenderAI`）。
	aiLog string
	// aiAction／aiMoves／aiFights 是同一件事的數字版，面板畫得出來。
	aiAction, aiMoves, aiFights int
	// leader 是這個省的司令，`sub_56D49`（§44）要問它在不在守方隊伍裡。
	leader game.GeneralID
	// tbl 是省份表，`sub_534FF`（§47）要掃鄰省找支援。
	tbl *game.ProvinceTable
	// supAtk／supDef 是雙方帶進這場戰鬥的資源與兵力總和，
	// 比率門檻（§48）要用。攻方＝第一方、守方＝第二方。
	supAtk, supDef game.BattleSupply
	// units 是全期將領表，`sub_5A881`（數某省的可用將領）要掃它。
	units []game.CombatUnit
	// enableLastSteps 對應 `byte_6FFCA & 4`；從策略層設定複製進戰鬥，
	// 不讓 UI fixture 的零值悄悄關閉 AI 後段行動。
	enableLastSteps bool
}

// finishWithWinner 只記錄已分出勝負的收尾分類；世界層投影由
// persistFinishedBattle 統一執行，避免每條勝負分支各自寫一份戰後規則。
// 原版 `byte_64901` 為 1／2 時不進平局寫回函式。
func (b *battleState) finishWithWinner(winner game.BattleSide) {
	if b == nil {
		return
	}
	settlement, err := game.ClassifyBattleSettlement(b.turn, b.turnLimit(), winner)
	if err != nil {
		// 這個呼叫端只傳入已驗證的 1／2；若未來改動破壞契約，
		// 保留未知零值而不是把錯誤當成可寫回狀態。
		b.settlement = game.BattleSettlement{}
		return
	}
	b.settlement = settlement
	b.finished = true
}

// finishByTurnLimit 記錄「無勝方、回合用盡」的已證實分類。這是目前唯一
// 標記原版需要進入 `sub_3964E` 的收尾，但實際 `.DT2` bytes 仍留給 oracle。
func (b *battleState) finishByTurnLimit() {
	if b == nil {
		return
	}
	settlement, err := game.ClassifyBattleSettlement(b.turn, b.turnLimit(), game.BattleSideNone)
	if err != nil {
		b.settlement = game.BattleSettlement{}
		return
	}
	b.settlement = settlement
	b.finished = true
}

// persistFinishedBattle 讓所有已結束分支共用同一個副本寫回入口。
// 沒有載入戰鬥副本的新局由 app.persistBattleSnapshot 安全略過；有來源
// 卻只載入一份時則回錯，避免玩家以為兩份已同步。
func (a *app) persistFinishedBattle() error {
	if a == nil || a.battle == nil || !a.battle.finished {
		return nil
	}
	if err := a.applyBattleSettlement(); err != nil {
		return err
	}
	return a.persistBattleSnapshot()
}

// clearBattleFlag 清除目前戰場省份的策略層交戰旗。它用於 ESC 放棄尚未
// 結算的戰鬥與已證實的立即撤退；兩者都不應把半成品旗標帶回下一回合。
func (a *app) clearBattleFlag() {
	if a == nil || a.battle == nil || a.tbl == nil || a.battle.sim == nil {
		return
	}
	if p, err := a.tbl.At(a.battle.sim.At); err == nil {
		p.Flags &^= game.ProvinceFlagInBattle
	}
}

// applyBattleSettlement 將已分類的收尾投影回策略層。這一步與 DT2／MEM_WAR
// writer 分離：即使玩家用初始 TOWN 檔開局、沒有副本輸出路徑，戰後省份與
// 將領所屬仍要在目前遊戲世界中一致；未解戰鬥 bytes 仍由 writer 原樣保留。
func (a *app) applyBattleSettlement() error {
	if a == nil {
		return nil
	}
	b := a.battle
	if b == nil || b.sim == nil || b.settled {
		return nil
	}
	if b.settlement.Kind == game.BattleSettlementUnknown {
		// 舊測試 fixture／尚未分類的異常結束不猜戰後結果；ESC 仍會
		// 由 clearBattleFlag 清理暫時旗標。
		return nil
	}
	if a.tbl == nil {
		return nil
	}
	var units []game.CombatUnit
	if a.world != nil {
		units = a.world.Units
	}
	if err := game.ApplyBattleSettlement(b.sim, a.tbl, a.generals, units, b.settlement); err != nil {
		return err
	}
	if b.settlement.Kind == game.BattleSettlementDecisive &&
		b.settlement.Winner == game.BattleSideFirst {
		// 攻方勝後，地圖游標跟著玩家軍隊進入新佔領省，避免
		// 「戰報說拿下、回地圖卻仍停在來源省」的半完成體驗。
		a.current = b.sim.At
	}
	b.settled = true
	return nil
}

// turnLimit 回傳開戰時固定的原版回合上限。零值只供舊測試 fixture 使用，
// fail-closed 地退回一般月份 16，不從未知狀態猜測二月。
func (b *battleState) turnLimit() int {
	if b != nil && b.turnCap > 0 {
		return b.turnCap
	}
	return game.BattleTurnCap(false)
}

// confirmedRetreatTargets 只回傳目前有正常 DOSBox 路徑閉合的候選清單。
//
// 2026-08-09 的固定樣本是「攻打陝西省（18），來源河南省（19）」；原版
// 在撤退分支列出 19、26、14、17。其他交戰組合的候選來源與順序仍未知，
// 因此不以鄰接、勢力或來源省假說補出清單。空回傳讓呼叫端保留在安全的
// 主選單，不把未證實的撤退規則當成完成。
func confirmedRetreatTargets(at, from game.ProvinceID) []game.ProvinceID {
	if at != 18 || from != 19 {
		return nil
	}
	return []game.ProvinceID{19, 26, 14, 17}
}

// appendProvinceInput 依既有政略省份輸入契約追加一位數字。超過 39 或
// 會溢位時保持原值，讓非法輸入不會污染戰鬥狀態。
func appendProvinceInput(current uint32, digit int) (uint32, bool) {
	if digit < 0 || digit > 9 {
		return current, false
	}
	if current > (uint32(game.ProvinceCount)-uint32(digit))/10 {
		return current, false
	}
	next := current*10 + uint32(digit)
	if next > uint32(game.ProvinceCount) {
		return current, false
	}
	return next, true
}

func retreatDestination(input uint32, targets []game.ProvinceID) (game.ProvinceID, bool) {
	if input == 0 {
		return 0, false
	}
	want := game.ProvinceID(input)
	return want, provinceIn(targets, want)
}

// startBattle 從當前省對某個鄰省開戰。
//
// 玩家控制**攻方**，從 `from` 省打進 `at` 省。自動守方部署依原版
// `sub_4166E` 掃 NWMAP `0x4000` 旗標候選；玩家手動部署仍是另一條
// `sub_42566` 互動路徑。
func (a *app) startBattle(at, from game.ProvinceID) error {
	fromProv, err := a.tbl.At(from)
	if err != nil {
		return err
	}
	atProv, err := a.tbl.At(at)
	if err != nil {
		return err
	}
	if atProv.InBattle() {
		return fmt.Errorf("省 %d 已在交戰中", at)
	}
	// `GeneralID` 是 MAN 槽位（1-based），不是「該省篩選後的序號」。
	// Faction 對應執行期記錄 +14，應填省份司令而非省編號；戰力公式的
	// 十大勢力加成與敵我判定都會讀這個欄位（docs/spec/02 §3、docs/re/08 §4）。
	atk := combatants(a.generals, from, fromProv.Commander)
	def := combatants(a.generals, at, atProv.Commander)
	if len(atk) == 0 {
		return fmt.Errorf("省 %d 沒有將領可以出兵", from)
	}
	if len(def) == 0 {
		return fmt.Errorf("省 %d 沒有守軍", at)
	}

	// 自動守方候選是 NWMAP 0x4000，掃描順序 0 → 195；不再以
	// WARPOS==0 的「腹地」假說取代已閉合的原版資料流。
	defZone, err := a.m.DefenderDeployZone(at)
	if err != nil {
		return err
	}
	placed := 0
	for _, c := range defZone {
		if placed >= len(def) {
			break
		}
		def[placed].Cell = c
		placed++
	}
	def = def[:placed]
	if placed == 0 {
		return fmt.Errorf("省 %d 的 NWMAP 0x4000 候選放不下守軍", at)
	}

	sim, err := game.NewBattleSim(a.m, at, from, atk, def, game.StrengthOpts{Stage: 1})
	if err != nil {
		return err
	}
	// 值 4（`sub_3CA09`）的 mode 1 候選會追加當前交戰省司令；
	// 這個欄位來自省份記錄 `+20`，不能從守方部隊順序猜。
	sim.AtCommander = atProv.Commander
	sim.BeginTurn()

	// 守方司令＝被打的那個省的司令（省份記錄 `+20`）。
	// `sub_56D49`（§44）問的就是「他本人在不在守方隊伍裡」，
	// 那是電腦改變打法的觸發點。
	leader := atProv.Commander
	// 雙方帶進戰場的資源（§48 的比率門檻要用）。
	//
	// ⚠️ **這是 remake 的取值方式，不是原版的。** 原版存在
	// `MEM_WAR.DAT` 的 `+0..+7`（`docs/re/05` §2），但「出兵時帶多少」
	// 是誰決定的還沒解。這裡直接取雙方所屬省的存量當近似，
	// 並把它標成已知差異——不要拿它對原版做行為驗收。
	enableLastSteps := true
	if a.world != nil {
		enableLastSteps = a.world.EnableExtra
	}
	b := &battleState{sim: sim, turn: 1, turnCap: game.BattleTurnCap(a.month == 2), mode: battleModeCommand,
		leader: leader, tbl: a.tbl, units: a.world.Units, enableLastSteps: enableLastSteps}
	// `RetreatCandidates` 會保留省份表的原始鄰接順序：18←19 的清單
	// 帶有固定 DOSBox oracle 標籤，其餘組合則明示為 remake fallback，
	// 不再讓正常玩家路徑永遠卡在單一樣本。
	retreat := game.RetreatCandidates(a.tbl, at, from)
	b.retreatTargets = retreat.Candidates
	if len(b.retreatTargets) == 0 {
		// 舊 fixture 沒有完整省份表時仍保留已驗證樣本入口。
		b.retreatTargets = confirmedRetreatTargets(at, from)
	}
	b.supAtk = supplyOf(a.tbl, from, atk)
	b.supDef = supplyOf(a.tbl, at, def)
	// ProvinceFlagInBattle 是策略層已證實的暫時旗標；它必須在成功建立
	// BattleSim 後才立起，避免部署失敗留下無法清除的假戰鬥。
	atProv.Flags |= game.ProvinceFlagInBattle
	a.battle = b
	a.screen, a.dirty = screenBattle, true
	return nil
}

// supplyOf 組出某一方的 `BattleSupply`：資源取所屬省的存量，
// 兵力總和照 `sub_3A4CE` 把 10 個槽位的 `+17` 加起來。
func supplyOf(tbl *game.ProvinceTable, prov game.ProvinceID, units []*game.Combatant) game.BattleSupply {
	s := game.BattleSupply{Troops: game.TroopTotal(units)}
	if p, err := tbl.At(prov); err == nil {
		s.Gold, s.Food = int(p.Gold), int(p.Food)
		s.Ammo, s.Fuel = int(p.Ammo), int(p.Fuel)
	}
	return s
}

// combatants 把將領表換成戰場單位。
//
// 屬性全部來自 `MAN(N).DAT`——執行期記錄與檔案記錄是同一套佈局
// （`docs/spec/02` §3），所以兵種、戰力欄位都直接讀檔案。
func combatants(gs []game.General, prov game.ProvinceID, faction game.GeneralID) []*game.Combatant {
	var out []*game.Combatant
	for i := range gs {
		if gs[i].Province != prov {
			continue
		}
		if len(out) >= game.UnitsPerSide {
			break
		}
		g := &gs[i]
		// 保留完整 MAN 表的槽位 ID；不能用篩選後的 out index 重新編號。
		id := game.GeneralID(i + 1)
		out = append(out, &game.Combatant{
			CombatUnit: game.CombatUnit{
				General: id, Faction: faction, Cell: game.NoCell,
				Experience: g.Experience,
				Province:   prov, Max: 12, Current: 12, Active: true, Decaying: 80,
			},
			Strength: game.StrengthInput{
				Ability: g.AbilityA, Force: g.Force,
				F19: g.F19, F20: g.F20, F29: g.Stamina, F30: g.F30,
				Branch: g.Branch, General: id, Faction: faction,
			},
		})
	}
	return out
}

// updateBattle 處理戰鬥畫面的輸入。
//
// ESC 退回政略畫面（`CLAUDE.md` §9：ESC 只取消／退回，不離開遊戲）。
func (a *app) updateBattle() error {
	b := a.battle
	if b == nil {
		a.screen, a.dirty = screenMap, true
		return nil
	}
	if err := a.syncBattleForces(); err != nil {
		return err
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		if b.mode != battleModeCommand {
			b.mode = battleModeCommand
			b.log = "返回戰鬥指令選單"
			a.dirty = true
			return nil
		}
		a.battle = nil
		a.screen, a.dirty = screenMap, true
		return nil
	}
	if b.finished {
		return nil
	}
	enterMode := func(mode battleMode) bool {
		// `RetreatCandidates` 已提供省份表 fallback；若仍為空，代表
		// 目前輸入沒有可重建的省份資料，安全停在主選單。
		if mode == battleModeRetreat && len(b.retreatTargets) == 0 {
			b.log = "撤退候選尚未由原版證據閉合"
			a.dirty = true
			return false
		}
		b.mode = mode
		b.retreatInput = 0
		b.log = battleModeMessage(mode)
		a.dirty = true
		return true
	}

	if b.mode == battleModeRetreat {
		changed := false
		if a.deleteDigitPressed() {
			b.retreatInput /= 10
			changed = true
		}
		for d, key := range digitKeys() {
			if !a.digitPressed(d, key) {
				continue
			}
			if next, ok := appendProvinceInput(b.retreatInput, d); ok {
				b.retreatInput = next
				changed = true
			}
		}
		if a.submitPressed() {
			to, ok := retreatDestination(b.retreatInput, b.retreatTargets)
			if !ok {
				b.log = "撤退省份不在原版候選清單"
				a.dirty = true
				return nil
			}
			// 立即撤退樣本在回到政略地圖時不改寫 DT2／MEM_WAR；
			// 這裡只移動目前畫面的省份，不呼叫任何戰後寫回入口。
			b.settlement = game.BattleSettlementImmediateRetreat()
			a.clearBattleFlag()
			a.current = to
			a.battle = nil
			a.screen, a.dirty = screenMap, true
			return nil
		}
		if changed {
			a.dirty = true
		}
		return nil
	}

	// Tab：換一個攻方單位。
	if a.actionPressed(actions.BattleNextUnit, ebiten.KeyTab) {
		if len(b.sim.Attacker) == 0 {
			return nil
		}
		b.sel = (b.sel + 1) % len(b.sim.Attacker)
		b.log = ""
		a.dirty = true
		return nil
	}
	u := b.current()
	if u == nil {
		b.finished = true
		a.dirty = true
		return nil
	}

	// 點擊周圍空格是既有的裝置無關捷徑；先進入移動層，再沿同一條
	// `sim.Move` 規則執行，避免滑鼠直接改寫戰鬥狀態。
	if b.mode == battleModeCommand {
		if n, ok := actions.BattleCommandNumber(a.pointerAction); ok {
			if mode, valid := battleModeForCommand(n); valid {
				enterMode(mode)
				return nil
			}
		}
		if _, pointerMove := actions.BattleMoveDirection(a.pointerAction); pointerMove {
			b.mode = battleModeMove
		}
		if n, ok := commandNumberPressed(); ok {
			if mode, valid := battleModeForCommand(n); valid {
				enterMode(mode)
				return nil
			}
		}
	}

	if b.mode == battleModeMove {
		if d, ok := a.directionPressed(); ok {
			if _, err := b.sim.Move(u.General, d); err != nil {
				b.log = err.Error()
			} else {
				b.log = fmt.Sprintf("移動到 %d，剩餘機動力 %d", u.Cell, u.Current)
			}
			// 原版移動分支完成一次輸入後回到五項主選單；失敗也要能
			// 重新選擇，不把玩家鎖在子狀態。
			b.mode = battleModeCommand
			a.dirty = true
			return nil
		}
		return nil
	}

	if b.mode == battleModeAttack {
		if n, ok := attackOptionPressed(); ok {
			return a.executeBattleAttack(n)
		}
		if n, ok := actions.BattleAttackTargetNumber(a.pointerAction); ok {
			return a.executeBattleAttack(n)
		}
		return nil
	}

	if b.mode == battleModeGarrison {
		a.executeBattleGarrison(u)
		a.dirty = true
		return nil
	}
	if b.mode == battleModeInspect {
		b.log = fmt.Sprintf("查閱：將領 %d｜兵力 %d｜格 %d｜命令 %d",
			u.General, u.Force(), u.Cell, u.Command)
		fmt.Fprintln(os.Stderr, "[battle]", b.log)
		b.mode = battleModeCommand
		a.dirty = true
		return nil
	}

	// Enter：remake 的直接攻擊捷徑，選取目前目標表的第一個目標；兵種 4
	// 在沒有相鄰目標時也可能選到已證實的遠程候選。原版主選單的「攻擊」
	// 分支不等同於這個呼叫，故只在主選單保留它，並在文件中標成差異。
	if b.mode == battleModeCommand && (a.pointerAction == actions.BattleAttack ||
		inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeyKPEnter)) {
		return a.executeBattleAttack(1)
	}

	// 未知的原版細節不阻止 remake 外殼操作；若未來擴充新的子狀態，
	// 仍保持 fail-closed 回到主選單。
	if b.mode != battleModeCommand {
		return nil
	}

	// 空白鍵：結束這一回合——**先讓守方 AI 動**，再換回合。
	if a.actionPressed(actions.BattleEndTurn, ebiten.KeySpace) {
		b.runDefenderAI()
		if b.finished {
			if err := a.persistFinishedBattle(); err != nil {
				return err
			}
			b.log = b.aiLog
			a.dirty = true
			return nil
		}
		// 守方 AI 可能在同一回合交戰；即使這次只移動，窄同步入口也會
		// 以兵力相同回傳 no-op，不會改動未知欄位。
		b.forceDirty = true
		if err := a.syncBattleForces(); err != nil {
			return err
		}
		b.sim.Sweep()
		if over, won := b.sim.Over(); over {
			if won {
				b.finishWithWinner(game.BattleSideFirst)
			} else {
				b.finishWithWinner(game.BattleSideSecond)
			}
			if won {
				b.log = "攻方獲勝"
			} else {
				b.log = "守方獲勝"
			}
			if err := a.persistFinishedBattle(); err != nil {
				return err
			}
			a.dirty = true
			return nil
		}
		// 回合結束的三件事，順序照原版 `sub_41D20`（§49）：
		// 先扣補給、再加回合數、再檢查上限。
		if w := game.TurnUpkeep(&b.supAtk, &b.supDef); w != game.BattleSideNone {
			b.finishWithWinner(w)
			if w == game.BattleSideFirst {
				b.log = "守方補給見底，攻方獲勝"
			} else {
				b.log = "攻方補給見底，守方獲勝"
			}
			if err := a.persistFinishedBattle(); err != nil {
				return err
			}
			a.dirty = true
			return nil
		}
		b.sim.EndTurn()
		b.sim.BeginTurn()
		b.turn++
		// 原版的回合上限是 16（`byte_64900 == 10h`），到了就結束、不判勝負。
		if b.turn >= b.turnLimit() {
			b.finishByTurnLimit()
			b.log = "回合用盡，戰鬥結束"
			if err := a.persistFinishedBattle(); err != nil {
				return err
			}
			a.dirty = true
			return nil
		}
		b.log = fmt.Sprintf("第 %d／%d 回合｜攻方糧食可撐 %d 回合",
			b.turn, b.turnLimit(), b.supAtk.TurnsOfFood())
		a.dirty = true
	}
	return nil
}

// syncBattleForces 把戰場副本的已確認兵力同步到玩家世界。
//
// 只在戰鬥副本真的改過兵力時執行；格位、命令、補給與 `.DT2` 未解區域不在
// 這個入口處理。這讓 F10 自動存檔或 ESC 返回地圖時，不會遺失已發生的戰損。
func (a *app) syncBattleForces() error {
	b := a.battle
	if b == nil || !b.forceDirty {
		return nil
	}
	if _, err := b.sim.SyncBattleStatsToGenerals(a.generals); err != nil {
		return err
	}
	// AIWorld 另有一份與 generals 對齊的戰力鏡像；若只改存檔側
	// `a.generals`，下一個政略判斷仍會讀到戰鬥前兵力。
	if a.world != nil {
		for _, u := range append(append([]*game.Combatant(nil), b.sim.Attacker...), b.sim.Defender...) {
			if u == nil || u.General == 0 {
				continue
			}
			i := int(u.General) - 1
			if i < 0 || i >= len(a.world.Strengths) {
				return fmt.Errorf("戰鬥將領 %d 超出策略戰力鏡像", u.General)
			}
			a.world.Strengths[i].Force = u.Strength.Force
			if u.Strength.General == u.General && u.General != 0 {
				a.world.Strengths[i].Ability = u.Strength.Ability
				a.world.Strengths[i].F19 = u.Strength.F19
				a.world.Strengths[i].F20 = u.Strength.F20
				a.world.Strengths[i].F29 = u.Strength.F29
				a.world.Strengths[i].F30 = u.Strength.F30
			}
			if i < len(a.world.Units) {
				a.world.Units[i].Experience = u.Experience
			}
		}
	}
	b.forceDirty = false
	return nil
}

// directionPressed 讀取鍵盤六方向，或既有的周圍格點擊動作。
func (a *app) directionPressed() (game.HexDir, bool) {
	pointerDir, pointerMove := actions.BattleMoveDirection(a.pointerAction)
	if pointerMove {
		return game.HexDir(pointerDir), true
	}
	for i, key := range dirKeys {
		if inpututil.IsKeyJustPressed(key) {
			return game.HexDir(i + 1), true
		}
	}
	return 0, false
}

// runDefenderAI 讓守方由**決策鏈**指揮走一回合。
//
// 三層都在規則層（`internal/game`）：
//
//	DecideTurn      決策鏈選一個行動（13 種，`docs/re/31` §41）
//	ExecuteAction   執行層寫命令／目標／下一跳
//	依 +12 移動、相鄰就交戰
//
// gates 的比率、彈藥與後援都由目前戰鬥資源與省份表計算；只剩
// `DefaultPostStage` 的外部預約表生命週期仍不在玩家可見路徑宣稱 parity。
func (b *battleState) runDefenderAI() {
	// ⭐ `sub_53619`（§47）接上了：問「守方在這個省有沒有可用的鄰省支援」。
	// 原版回的是**反相**（有支援回 0），所以這裡取反。
	//
	// 這個判斷控制兩處必勝結算與值 16／17 的分流：
	// **戰力差五倍而且有後援，才敢直接判勝負。**
	//
	// ⭐ 比率門檻（§48）也接上了：**糧食夠但黃金不夠**。
	// 分支 A 的「我方」是第二方（守方），分支 B 的是第一方（攻方）。
	//
	// ⭐ 那個加項就是**回合數**（原版 `byte_64900`，§49）：
	//
	//	還能撐幾回合 + 已經打了幾回合 < 15
	//	  = 補給撐不到第 15 回合（戰鬥上限 16）
	gates := game.BattleChainGates{
		Sub53619:  !game.HasBattleSupport(b.tbl, b.sim.At, b.leader, b.units),
		RatioSelf: b.supDef.RatioGate(b.turn),
		RatioFoe:  b.supAtk.RatioGate(b.turn),
		// §43 的 `word_6493A == 0` = **第一方（攻方）的彈藥為 0**。
		Deploy:          b.supAtk.Ammo == 0,
		EnableLastSteps: b.enableLastSteps,
	}
	d := b.sim.DecideTurn(b.turn, gates, b.leader, 0)

	route := func(to, from game.CellIndex) game.CellIndex {
		return b.sim.RouteNextCell(to, from)
	}

	r := b.sim.ExecuteAction(d.A.Action, b.sim.Defender, b.sim.Attacker, route)
	if r.Decisive {
		if r.DecisiveAttackerWon {
			b.finishWithWinner(game.BattleSideFirst)
		} else {
			b.finishWithWinner(game.BattleSideSecond)
		}
		if r.DecisiveAttackerWon {
			b.aiLog = "守方認輸：" + r.Note
		} else {
			b.aiLog = "攻方潰敗：" + r.Note
		}
		return
	}

	moves, fights := 0, 0
	for _, u := range b.sim.Defender {
		if b.sim.StepByOrder(u) {
			moves++
		}
		if b.sim.EngageIfAdjacent(u) {
			fights++
		}
	}
	b.aiLog = fmt.Sprintf("守方：%s（動 %d 打 %d）",
		game.BattleActionName(d.A.Action), moves, fights)
	if !r.Implemented {
		b.aiLog += " ⚠未實作"
	}
	b.aiAction, b.aiMoves, b.aiFights = int(d.A.Action), moves, fights

	// ⚠️ **畫面上還畫不出這行字。** 原版的字模是每個場景一份子集
	// （`CLAUDE.md` §3.5），畫不出自由組合的中文；完整字型是 M6 的事。
	// 在那之前先印到 stderr，讓行為至少是可觀測的——
	// 面板上用數字顯示（`aiAction`／`aiMoves`／`aiFights`）。
	fmt.Fprintln(os.Stderr, "[battle]", b.aiLog)
}

// current 回傳目前選中的攻方單位，跳過已陣亡的。
func (b *battleState) current() *game.Combatant {
	for i := 0; i < len(b.sim.Attacker); i++ {
		u := b.sim.Attacker[(b.sel+i)%len(b.sim.Attacker)]
		if u.Alive() && u.Cell.Valid() {
			b.sel = (b.sel + i) % len(b.sim.Attacker)
			return u
		}
	}
	return nil
}

// adjacentEnemy 找選中單位旁邊的第一個敵人。
func (b *battleState) adjacentEnemy(u *game.Combatant) *game.Combatant {
	if targets := b.adjacentEnemies(u); len(targets) > 0 {
		return targets[0]
	}
	return nil
}

// adjacentEnemies 是攻擊子選單公告的穩定目標順序：沿用原始守方陣列，
// 只列出存活且相鄰的單位。這讓數字鍵、滑鼠與觸控共用同一份目標表。
func (b *battleState) adjacentEnemies(u *game.Combatant) []*game.Combatant {
	if b == nil || b.sim == nil || u == nil || !u.Cell.Valid() {
		return nil
	}
	out := make([]*game.Combatant, 0, 6)
	for _, d := range b.sim.Defender {
		if d != nil && d.Alive() && d.Cell.Valid() && game.Adjacent(u.Cell, d.Cell) {
			out = append(out, d)
		}
	}
	return out
}

// battleAttackTargets 是鍵盤、滑鼠與觸控共用的穩定攻擊目標表。
// 相鄰目標沿用原本守方陣列順序；只有兵種 4 另追加已由靜態證據閉合的
// 非相鄰遠程候選，且最多公告六個，符合現有攻擊子選單的輸入寬度。
func (b *battleState) battleAttackTargets(u *game.Combatant) []*game.Combatant {
	if b == nil || b.sim == nil || u == nil {
		return nil
	}
	out := b.adjacentEnemies(u)
	if u.Branch() != game.BranchArtiller {
		if len(out) > 6 {
			return out[:6]
		}
		return out
	}
	for _, target := range b.sim.RangedTargets(u) {
		if game.Adjacent(u.Cell, target.Cell) || containsCombatant(out, target) {
			continue
		}
		out = append(out, target)
		if len(out) == 6 {
			break
		}
	}
	return out
}

func containsCombatant(list []*game.Combatant, want *game.Combatant) bool {
	for _, got := range list {
		if got == want {
			return true
		}
	}
	return false
}

// executeBattleGarrison 把駐軍命令接到既有的城市尋路／命令欄位。
// 原版駐軍的完整互動與城市選擇時機仍未有跨場景 oracle；remake 採
// 「第一個可達城市、先走一格」的明確差異，讓五項戰鬥命令不再是死路。
func (a *app) executeBattleGarrison(u *game.Combatant) {
	b := a.battle
	if b == nil || b.sim == nil || u == nil {
		return
	}
	for _, city := range game.CityCells(b.sim.Field) {
		if u.Cell == city {
			u.TargetUnit = 0
			u.NextCell = game.NoCell
			u.Flags13 &^= game.UnitAssignedBit
			u.Command = game.BattleCmdGarrisoned
			b.log = fmt.Sprintf("將領 %d 已在城市格 %d 駐軍", u.General, city)
			b.mode = battleModeCommand
			return
		}
		next := b.sim.RouteNextCell(city, u.Cell)
		if next == game.NoCell {
			continue
		}
		u.Command = game.BattleCmdGarrison
		u.AssignTo(0, next)
		moved := b.sim.StepByOrder(u)
		if moved && u.Cell == city {
			u.Command = game.BattleCmdGarrisoned
		}
		b.log = fmt.Sprintf("將領 %d 前往城市 %d（下一格 %d）", u.General, city, u.Cell)
		b.mode = battleModeCommand
		return
	}
	b.log = "沒有可達城市，駐軍命令未執行"
	b.mode = battleModeCommand
}

// executeBattleAttack 把 1..6 目標編號接到目前已確認的戰損路徑。
// 相鄰／非相鄰目標都交給共用的 handler 分派；目前已接協同、衝鋒、
// 遠程與特殊零損失視覺分支。原版第二層六鍵選單的完整名稱、副作用與第 6 鍵語意仍未知。
func (a *app) executeBattleAttack(number int) error {
	b := a.battle
	if b == nil || b.sim == nil {
		return nil
	}
	u := b.current()
	if u == nil {
		b.log = "沒有可操作的攻方單位"
		a.dirty = true
		return nil
	}
	targets := b.battleAttackTargets(u)
	if number < 1 || number > len(targets) {
		b.log = fmt.Sprintf("攻擊目標 %d 不在相鄰目標清單", number)
		b.mode = battleModeCommand
		a.dirty = true
		return nil
	}
	target := targets[number-1]
	result, err := b.sim.ResolveBattleAttack(u, target)
	if err != nil {
		b.log = err.Error()
		b.mode = battleModeCommand
		a.dirty = true
		return nil
	}
	b.forceDirty = true
	b.log = fmt.Sprintf("攻擊目標 %d｜攻損 %d／守損 %d（%s）", number,
		result.LossAttacker, result.LossTarget, result.Effect.Kind)
	b.sim.Sweep()
	if err := a.syncBattleForces(); err != nil {
		return err
	}
	if over, won := b.sim.Over(); over {
		if won {
			b.finishWithWinner(game.BattleSideFirst)
			b.log = "攻方獲勝"
		} else {
			b.finishWithWinner(game.BattleSideSecond)
			b.log = "守方獲勝"
		}
		if err := a.persistFinishedBattle(); err != nil {
			return err
		}
	}
	b.mode = battleModeCommand
	a.dirty = true
	return nil
}

// drawBattle 把戰場、單位與游標畫上去。
func (a *app) drawBattle(c *render.Canvas) error {
	b := a.battle
	if b == nil {
		return nil
	}
	if a.highModernBattle() {
		return a.drawModernBattle(c)
	}
	bf, err := a.m.Battlefield(b.sim.At)
	if err != nil {
		return err
	}
	if err := c.DrawThemedBattlefield(bf, a.battlefieldTheme, fieldX, fieldY); err != nil {
		return err
	}
	for _, u := range b.sim.Defender {
		if !u.Alive() || !u.Cell.Valid() {
			continue
		}
		idx := render.BranchIcon(u.Branch(), true, u.Facing)
		if err := a.drawThemedUnit(c, idx, u.Cell); err != nil {
			return err
		}
	}
	for _, u := range b.sim.Attacker {
		if !u.Alive() || !u.Cell.Valid() {
			continue
		}
		idx := render.BranchIcon(u.Branch(), false, u.Facing)
		if err := a.drawThemedUnit(c, idx, u.Cell); err != nil {
			return err
		}
	}
	// 選中的單位框黃色，它旁邊的敵人框紅色。
	if u := b.current(); u != nil {
		c.DrawCellCursor(fieldX, fieldY, u.Cell, cursorOurs)
		if t := b.adjacentEnemy(u); t != nil {
			c.DrawCellCursor(fieldX, fieldY, t.Cell, cursorTarget)
		}
		if b.mode == battleModeAttack {
			// 攻擊子選單的 1..6 現在是穩定目標序號；把同一序號
			// 疊在實際敵軍格上，讓鍵盤、滑鼠與 Android 觸控都能
			// 從畫面知道「數字會打誰」，而不是只靠游標顏色猜。
			for i, target := range b.battleAttackTargets(u) {
				if i >= 6 || target == nil || !target.Alive() || !target.Cell.Valid() {
					continue
				}
				dx, dy := target.Cell.ScreenXY()
				c.DrawSmallDigit(i+1, cursorTarget, fieldX+dx+2, fieldY+dy+2)
			}
		}
	}
	if err := c.DrawBattlePanel(a.battlePanelData(), a.fonts); err != nil {
		return err
	}
	// 戰鬥操作結果使用完整倚天字庫畫在 remake 面板底部；原典字模不
	// 能承載自由組合訊息時仍保留 stderr 診斷，不讓訊息列阻斷規則。
	if a.eten != nil && b.log != "" {
		c.DrawSemanticText(a.eten, b.log, a.uiStyle().Ink, render.BattlePanelX+8, 306)
	}
	return nil
}

// drawModernBattle 是 H1-c 的呈現切片。它只把現有 battleState 整理成
// ModernBattleSurfaceData；移動／攻擊／撤退／AI 仍由下面既有 updateBattle 路徑處理。
func (a *app) drawModernBattle(c *render.Canvas) error {
	if a == nil || a.battle == nil || a.battle.sim == nil {
		return nil
	}
	b := a.battle
	labels, err := a.modernBattleLabels()
	if err != nil {
		return err
	}
	d := a.battlePanelData()
	bf, err := a.m.Battlefield(b.sim.At)
	if err != nil {
		return err
	}
	current := b.current()
	var currentCell game.CellIndex
	if current != nil {
		currentCell = current.Cell
	}
	targets := make([]game.CellIndex, 0, 6)
	numbers := make([]int, 0, 6)
	if current != nil && b.mode == battleModeAttack {
		for i, target := range b.battleAttackTargets(current) {
			if i >= 6 || target == nil || !target.Alive() || !target.Cell.Valid() {
				continue
			}
			targets = append(targets, target.Cell)
			numbers = append(numbers, i+1)
		}
	}
	return c.DrawModernBattleSurface(render.ModernBattleSurfaceData{
		Battlefield: bf,
		Theme:       a.battlefieldTheme, Units: a.unitTheme,
		Attackers: b.sim.Attacker, Defenders: b.sim.Defender,
		CurrentCell: currentCell, TargetCells: targets, TargetNums: numbers,
		Panel: d, Labels: labels, Fonts: a.eten, Log: b.log, Style: a.uiStyle(),
	})
}

func (a *app) modernBattleLabels() (render.ModernBattleLabels, error) {
	get := func(key string) (string, error) { return a.wordingText(key) }
	var out render.ModernBattleLabels
	var err error
	if out.Title, err = get("battle.title"); err != nil {
		return out, err
	}
	if out.Attacker, err = get("battle.attacker"); err != nil {
		return out, err
	}
	if out.Defender, err = get("battle.defender"); err != nil {
		return out, err
	}
	if out.Attack, err = get("battle.attack"); err != nil {
		return out, err
	}
	if out.Province, err = get("battle.province"); err != nil {
		return out, err
	}
	if out.Date, err = get("battle.date"); err != nil {
		return out, err
	}
	for i, key := range []string{"battle.units", "battle.soldiers", "battle.gold", "battle.food", "battle.ammo", "battle.fuel"} {
		var value string
		if value, err = get(key); err != nil {
			return out, err
		}
		switch i {
		case 0:
			out.Units = value
		case 1:
			out.Soldiers = value
		case 2:
			out.Gold = value
		case 3:
			out.Food = value
		case 4:
			out.Ammo = value
		case 5:
			out.Fuel = value
		}
	}
	for i, key := range []string{"battle.command.move", "battle.command.attack", "battle.command.retreat", "battle.command.garrison", "battle.command.inspect"} {
		if out.Commands[i], err = get(key); err != nil {
			return out, err
		}
	}
	for i, key := range []string{"battle.control.attack", "battle.control.next", "battle.control.end"} {
		if out.Controls[i], err = get(key); err != nil {
			return out, err
		}
	}
	if out.Mode, err = get("battle.mode.retreat"); err != nil {
		return out, err
	}
	if out.Delete, err = get("battle.delete"); err != nil {
		return out, err
	}
	if out.Submit, err = get("battle.submit"); err != nil {
		return out, err
	}
	if a.battle != nil && a.battle.sim != nil {
		if p, lookupErr := a.tbl.At(a.battle.sim.From); lookupErr == nil {
			out.AttackerName = a.generalDisplayName(p.Commander)
		}
		if p, lookupErr := a.tbl.At(a.battle.sim.At); lookupErr == nil {
			out.DefenderName = a.generalDisplayName(p.Commander)
		}
	}
	return out, nil
}

// drawThemedUnit 讓部隊圖示與地形／鐵路在同一個主題 asset group 內切換。
// BranchIcon 仍是唯一的原版索引對照；這裡只取 bitmap 並繪製，不改規則狀態。
func (a *app) drawThemedUnit(c *render.Canvas, index int, cell game.CellIndex) error {
	if a.unitTheme == nil {
		return fmt.Errorf("戰鬥主題沒有部隊圖示 provider")
	}
	bitmap, err := a.unitTheme.Unit(index)
	if err != nil {
		return err
	}
	return c.DrawThemedUnitAtCell(bitmap, fieldX, fieldY, cell)
}

// battlePanelData 把戰鬥狀態整理成右側面板要顯示的內容。
//
// 版面與欄位出自實機截圖（`docs/playtest/14`）。四種資源目前**從省份記錄取**
// ——原版是從參戰部隊表（`ds:A358h`）取，那張表的完整語意還沒解完
// （`docs/re/29` §4），所以這裡先用省份的值。**這是已知的差異。**
func (a *app) battlePanelData() render.BattlePanelData {
	b := a.battle
	d := render.BattlePanelData{
		Province: b.sim.At,
		Month:    a.month,
		Day:      1,
		Style:    a.uiStyle(),
	}
	count := func(us []*game.Combatant) (units, soldiers uint32) {
		for _, u := range us {
			if u.Alive() {
				units++
				soldiers += uint32(u.Strength.Force)
			}
		}
		return
	}
	d.AIAction, d.AIMoves, d.AIFights = b.aiAction, b.aiMoves, b.aiFights
	d.ShowBattleMenu = true
	d.ShowBattleControls = !b.finished
	d.RetreatActive = b.mode == battleModeRetreat
	d.RetreatInput = b.retreatInput
	d.RetreatTargets = append([]game.ProvinceID(nil), b.retreatTargets...)
	switch b.mode {
	case battleModeMove:
		d.BattleMenuMode = render.BattleMenuMove
	case battleModeAttack:
		d.BattleMenuMode = render.BattleMenuAttack
	case battleModeRetreat:
		d.BattleMenuMode = render.BattleMenuRetreat
	case battleModeGarrison:
		d.BattleMenuMode = render.BattleMenuGarrison
	case battleModeInspect:
		d.BattleMenuMode = render.BattleMenuInspect
	default:
		d.BattleMenuMode = render.BattleMenuCommand
	}
	d.Attacker.Units, d.Attacker.Soldiers = count(b.sim.Attacker)
	d.Defender.Units, d.Defender.Soldiers = count(b.sim.Defender)

	if p, err := a.tbl.At(b.sim.From); err == nil {
		d.Attacker.Leader = p.Commander
		d.Attacker.Gold, d.Attacker.Food = uint32(p.Gold), uint32(p.Food)
		d.Attacker.Ammo, d.Attacker.Fuel = uint32(p.Ammo), uint32(p.Fuel)
	}
	if p, err := a.tbl.At(b.sim.At); err == nil {
		d.Defender.Leader = p.Commander
		d.Defender.Gold, d.Defender.Food = uint32(p.Gold), uint32(p.Food)
		d.Defender.Ammo, d.Defender.Fuel = uint32(p.Ammo), uint32(p.Fuel)
	}
	return d
}
