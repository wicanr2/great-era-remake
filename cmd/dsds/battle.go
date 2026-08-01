package main

import (
	"fmt"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
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

// 圖示編號。綠紅成對是敵我兩方（`docs/formats/05` §3）；
// 兵種編號與圖示編號的對應**還沒解出來**，所以只用第一對鋼盔，
// 不假裝畫得出兵種差異。
const (
	iconOurs   = 0 // 鋼盔（綠）
	iconTheirs = 1 // 鋼盔（紅）
)

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
	a.battle = &battleState{sim: sim}
	a.screen, a.dirty = screenBattle, true
	return nil
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
				F19: g.F19, F20: g.F20, F29: g.F29, F30: g.F30,
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
		if !inpututil.IsKeyJustPressed(k) {
			continue
		}
		d := game.HexDir(i + 1)
		if _, err := b.sim.Move(u.General, d); err != nil {
			b.log = err.Error()
		} else {
			b.log = fmt.Sprintf("移動到 %d，剩餘機動力 %d", u.Cell, u.Current)
		}
		a.dirty = true
		return nil
	}

	// Enter：對相鄰的敵人發動攻擊。
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) ||
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

	// 空白鍵：結束這一回合（衰減 + 機動力回滿）。
	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		b.sim.EndTurn()
		b.sim.BeginTurn()
		b.log = "換下一回合"
		a.dirty = true
	}
	return nil
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
		if err := c.DrawUnitAtCell(a.icons, iconTheirs, assets.EGADefaultPalette,
			fieldX, fieldY, u.Cell); err != nil {
			return err
		}
	}
	for _, u := range b.sim.Attacker {
		if !u.Alive() || !u.Cell.Valid() {
			continue
		}
		if err := c.DrawUnitAtCell(a.icons, iconOurs, assets.EGADefaultPalette,
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
	return nil
}
