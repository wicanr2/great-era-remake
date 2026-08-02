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
// **操作是 remake 的選擇，不是還原原版。** 原版的戰鬥選單是七項
// （`docs/mechanics/30-combat.md` §3），內容與按鍵還沒解出來；
// 這裡用數字鍵 1–6 對六個方向，因為六角格的方向編號本來就是 1–6
// （`docs/re/07` §6），對得起來又好記。

// 六角格的方向編號直接對應數字鍵 1–6。
var dirKeys = [...]ebiten.Key{
	ebiten.Key1, ebiten.Key2, ebiten.Key3,
	ebiten.Key4, ebiten.Key5, ebiten.Key6,
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
	sim      *game.BattleSim
	sel      int    // 目前選中第幾個攻方單位
	log      string // 最近一次動作的結果，畫在面板上
	finished bool

	// turn 是回合序號，決策鏈要用（每回合重跑一次）。
	turn int
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
}

// startBattle 從當前省對某個鄰省開戰。
//
// 玩家控制**攻方**，從 `from` 省打進 `at` 省。守方擺在腹地
// ——原版的守方佈署位置從哪來還沒解（`sub_41513` 只放攻方），
// 所以這裡挑腹地裡站得上去的格，並標記為 remake 差異。
func (a *app) startBattle(at, from game.ProvinceID) error {
	bf, err := a.m.Battlefield(at)
	if err != nil {
		return err
	}
	atkGens := game.GeneralsOf(a.generals, from)
	defGens := game.GeneralsOf(a.generals, at)
	if len(atkGens) == 0 {
		return fmt.Errorf("省 %d 沒有將領可以出兵", from)
	}
	if len(defGens) == 0 {
		return fmt.Errorf("省 %d 沒有守軍", at)
	}

	atk := combatants(atkGens, from)
	def := combatants(defGens, at)

	// 守方擺在腹地（WARPOS 值 0）且站得上去的格。
	placed := 0
	for i := 0; i < game.CellCount && placed < len(def); i++ {
		c := game.CellIndex(i)
		col, row := c.ColRow()
		if bf.Owner[row][col] != 0 || bf.Tiles[row][col].MoveCost() >= assets.MoveCostImpassable {
			continue
		}
		def[placed].Cell = c
		placed++
	}
	def = def[:placed]
	if placed == 0 {
		return fmt.Errorf("省 %d 的腹地放不下守軍", at)
	}

	sim, err := game.NewBattleSim(a.m, at, from, atk, def, game.StrengthOpts{Stage: 1})
	if err != nil {
		return err
	}
	sim.BeginTurn()

	// 守方司令＝被打的那個省的司令（省份記錄 `+20`）。
	// `sub_56D49`（§44）問的就是「他本人在不在守方隊伍裡」，
	// 那是電腦改變打法的觸發點。
	var leader game.GeneralID
	if p, err := a.tbl.At(at); err == nil {
		leader = p.Commander
	}
	// 雙方帶進戰場的資源（§48 的比率門檻要用）。
	//
	// ⚠️ **這是 remake 的取值方式，不是原版的。** 原版存在
	// `MEM_WAR.DAT` 的 `+0..+7`（`docs/re/05` §2），但「出兵時帶多少」
	// 是誰決定的還沒解。這裡直接取雙方所屬省的存量當近似，
	// 並把它標成已知差異——不要拿它對原版做行為驗收。
	b := &battleState{sim: sim, turn: 1, leader: leader, tbl: a.tbl, units: a.world.Units}
	b.supAtk = supplyOf(a.tbl, from, atk)
	b.supDef = supplyOf(a.tbl, at, def)
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
func combatants(gs []game.General, prov game.ProvinceID) []*game.Combatant {
	var out []*game.Combatant
	for i := range gs {
		if len(out) >= game.UnitsPerSide {
			break
		}
		g := &gs[i]
		id := game.GeneralID(i + 1)
		out = append(out, &game.Combatant{
			CombatUnit: game.CombatUnit{
				General: id, Faction: game.GeneralID(prov), Cell: game.NoCell,
				Province: prov, Max: 12, Current: 12, Active: true, Decaying: 80,
			},
			Strength: game.StrengthInput{
				Ability: g.AbilityA, Force: g.Force,
				F19: g.F19, F20: g.F20, F29: g.Stamina, F30: g.F30,
				Branch: g.Branch, General: id, Faction: game.GeneralID(prov),
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
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		a.battle = nil
		a.screen, a.dirty = screenMap, true
		return nil
	}
	if b.finished {
		return nil
	}

	// Tab：換一個攻方單位。
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
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

	// 1–6：往六個方向走一格。
	for i, k := range dirKeys {
		d := game.HexDir(i + 1)
		pointerDir, pointerMove := actions.BattleMoveDirection(a.pointerAction)
		if !inpututil.IsKeyJustPressed(k) && !(pointerMove && pointerDir == int(d)) {
			continue
		}
		if _, err := b.sim.Move(u.General, d); err != nil {
			b.log = err.Error()
		} else {
			b.log = fmt.Sprintf("移動到 %d，剩餘機動力 %d", u.Cell, u.Current)
		}
		a.dirty = true
		return nil
	}

	// Enter：對相鄰的敵人發動攻擊。
	if a.pointerAction == actions.BattleAttack ||
		inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
		inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
		target := b.adjacentEnemy(u)
		if target == nil {
			b.log = "旁邊沒有敵人"
			a.dirty = true
			return nil
		}
		la, lt, err := b.sim.Engage(u, target)
		if err != nil {
			b.log = err.Error()
		} else {
			b.log = fmt.Sprintf("攻損 %d／守損 %d", la, lt)
			b.sim.Sweep()
		}
		if over, won := b.sim.Over(); over {
			b.finished = true
			if won {
				b.log = "攻方獲勝"
			} else {
				b.log = "守方獲勝"
			}
		}
		a.dirty = true
		return nil
	}

	// 空白鍵：結束這一回合——**先讓守方 AI 動**，再換回合。
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		b.runDefenderAI()
		b.sim.Sweep()
		if over, won := b.sim.Over(); over {
			b.finished = true
			if won {
				b.log = "攻方獲勝"
			} else {
				b.log = "守方獲勝"
			}
			a.dirty = true
			return nil
		}
		// 回合結束的三件事，順序照原版 `sub_41D20`（§49）：
		// 先扣補給、再加回合數、再檢查上限。
		if w := game.TurnUpkeep(&b.supAtk, &b.supDef); w != game.BattleSideNone {
			b.finished = true
			if w == game.BattleSideFirst {
				b.log = "守方補給見底，攻方獲勝"
			} else {
				b.log = "攻方補給見底，守方獲勝"
			}
			a.dirty = true
			return nil
		}
		b.sim.EndTurn()
		b.sim.BeginTurn()
		b.turn++
		// 原版的回合上限是 16（`byte_64900 == 10h`），到了就結束、不判勝負。
		if b.turn >= game.BattleTurnCap(false) {
			b.finished = true
			b.log = "回合用盡，戰鬥結束"
			a.dirty = true
			return nil
		}
		b.log = fmt.Sprintf("第 %d／%d 回合｜攻方糧食可撐 %d 回合",
			b.turn, game.BattleTurnCap(false), b.supAtk.TurnsOfFood())
		a.dirty = true
	}
	return nil
}

// runDefenderAI 讓守方由**決策鏈**指揮走一回合。
//
// 三層都在規則層（`internal/game`）：
//
//	DecideTurn      決策鏈選一個行動（13 種，`docs/re/31` §41）
//	ExecuteAction   執行層寫命令／目標／下一跳
//	依 +12 移動、相鄰就交戰
//
// ⚠️ 那三個還沒解出來源的判斷（`BattleChainGates`）目前**全部給 false**
// ——比率門檻、佈防閘門、後援判斷。後果是守方偏向走「預設分流」，
// 補齊之前這是**已知落差**，不是最終行為。
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
		Deploy: b.supAtk.Ammo == 0,
	}
	d := b.sim.DecideTurn(b.turn, gates, b.leader, 0)

	route := func(to, from game.CellIndex) game.CellIndex {
		if game.Adjacent(from, to) {
			return to
		}
		best, bestD := game.NoCell, 1<<30
		for _, n := range from.Neighbours() {
			if b.sim.Occ[n] != 0 {
				continue
			}
			if dist := game.CellManhattan(n, to); dist < bestD {
				best, bestD = n, dist
			}
		}
		return best
	}

	r := b.sim.ExecuteAction(d.A.Action, b.sim.Defender, b.sim.Attacker, route)
	if r.Decisive {
		b.finished = true
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
	for _, d := range b.sim.Defender {
		if d.Alive() && d.Cell.Valid() && game.Adjacent(u.Cell, d.Cell) {
			return d
		}
	}
	return nil
}

// drawBattle 把戰場、單位與游標畫上去。
func (a *app) drawBattle(c *render.Canvas) error {
	b := a.battle
	if b == nil {
		return nil
	}
	bf, err := a.m.Battlefield(b.sim.At)
	if err != nil {
		return err
	}
	if err := c.DrawTiledBattlefield(bf, a.tiles, fieldX, fieldY); err != nil {
		return err
	}
	for _, u := range b.sim.Defender {
		if !u.Alive() || !u.Cell.Valid() {
			continue
		}
		idx := render.BranchIcon(u.Branch(), true, u.Facing)
		if err := c.DrawUnitAtCell(a.icons, idx, assets.EGADefaultPalette,
			fieldX, fieldY, u.Cell); err != nil {
			return err
		}
	}
	for _, u := range b.sim.Attacker {
		if !u.Alive() || !u.Cell.Valid() {
			continue
		}
		idx := render.BranchIcon(u.Branch(), false, u.Facing)
		if err := c.DrawUnitAtCell(a.icons, idx, assets.EGADefaultPalette,
			fieldX, fieldY, u.Cell); err != nil {
			return err
		}
	}
	// 選中的單位框黃色，它旁邊的敵人框紅色。
	if u := b.current(); u != nil {
		c.DrawCellCursor(fieldX, fieldY, u.Cell, cursorOurs)
		if t := b.adjacentEnemy(u); t != nil {
			c.DrawCellCursor(fieldX, fieldY, t.Cell, cursorTarget)
		}
	}
	return c.DrawBattlePanel(a.battlePanelData(), a.fonts)
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
