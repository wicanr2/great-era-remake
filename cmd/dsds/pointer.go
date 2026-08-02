package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
)

const pointerDragLimit = 8 // 640×350 邏輯畫布像素；超過就不當點擊。

type pointerPress struct {
	x, y   int
	screen screen
}

type pointerTracker struct {
	mouse   *pointerPress
	touches map[ebiten.TouchID]pointerPress
}

func (p *pointerTracker) cancel() {
	p.mouse = nil
	p.touches = nil
}

func withinClick(start pointerPress, x, y int, s screen) bool {
	if start.screen != s {
		return false
	}
	dx, dy := x-start.x, y-start.y
	return dx*dx+dy*dy <= pointerDragLimit*pointerDragLimit
}

// collectPointerAction 同時是桌面滑鼠與 Android 觸控的 Ebiten adapter。
// CursorPosition／TouchPosition 由 Ebiten Layout 先轉成 640×350 邏輯座標；
// 其他平台 adapter 可使用 actions.SurfaceToLogical 做同樣的黑邊反算。
func (a *app) collectPointerAction() actions.Action {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		a.pointer.mouse = &pointerPress{x: x, y: y, screen: a.screen}
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		start := a.pointer.mouse
		a.pointer.mouse = nil
		if start != nil && withinClick(*start, x, y, a.screen) {
			return actions.Hit(a.interactiveTargets(), x, y)
		}
	}

	pressed := inpututil.AppendJustPressedTouchIDs(nil)
	if len(pressed) > 0 && a.pointer.touches == nil {
		a.pointer.touches = map[ebiten.TouchID]pointerPress{}
	}
	for _, id := range pressed {
		x, y := ebiten.TouchPosition(id)
		a.pointer.touches[id] = pointerPress{x: x, y: y, screen: a.screen}
	}
	for _, id := range inpututil.AppendJustReleasedTouchIDs(nil) {
		start, ok := a.pointer.touches[id]
		delete(a.pointer.touches, id)
		if !ok {
			continue
		}
		x, y := ebiten.TouchPosition(id)
		if withinClick(start, x, y, a.screen) {
			if action := actions.Hit(a.interactiveTargets(), x, y); action != actions.None {
				return action
			}
		}
	}
	return actions.None
}

func (a *app) navigationActions() []actions.Action {
	switch a.screen {
	case screenMap, screenBattle, screenQuit:
		return nil
	case screenTransferConfirm, screenRecruitConfirm, screenTrainConfirm,
		screenSaveConfirm, screenLoadConfirm:
		return []actions.Action{actions.Cancel}
	case screenTransferSelection:
		out := []actions.Action{actions.Back}
		if a.transferSession != nil && len(a.transferSession.Selected()) > 0 {
			out = append(out, actions.Submit)
		}
		return out
	case screenBiography:
		out := []actions.Action{actions.Back}
		if a.bioPage+1 < a.bioPages {
			out = append(out, actions.NextPage)
		}
		if a.bioPage > 0 {
			out = append(out, actions.PreviousPage)
		}
		return out
	case screenViewProvinceNames:
		return []actions.Action{actions.Back, actions.NextPage}
	default:
		return []actions.Action{actions.Back}
	}
}

func (a *app) navigationTargets() []actions.Target {
	buttons := a.navigationActions()
	out := make([]actions.Target, len(buttons))
	for i, action := range buttons {
		p := uilayout.NavigationButton(640, a.navigationY(), i)
		out[i] = actions.Target{Action: action,
			Rect: actions.Rect{X: p.HitX, Y: p.HitY, W: p.HitW, H: p.HitH}}
	}
	return out
}

func (a *app) navigationY() int {
	if a.screen == screenCommand || a.screen == screenViewGeneral {
		return 294
	}
	if a.screen == screenViewGenerals || a.screen == screenTransferSelection {
		return 0
	}
	return 8
}

func (a *app) interactiveTargets() []actions.Target {
	if a.screen == screenBattle {
		return a.battlefieldTargets()
	}
	base := a.pointerTargets()
	out := append(a.numericKeypadTargets(), a.navigationTargets()...)
	return append(out, base...)
}

// battlefieldTargets 只公告目前選中單位周圍可由鍵盤完成的動作。
// 主畫面的同一張地圖刻意不走這條路：它是省內戰場展示，不是 39 省地圖。
func (a *app) battlefieldTargets() []actions.Target {
	if a.battle == nil || a.battle.finished {
		return nil
	}
	u := a.battle.current()
	if u == nil || !u.Cell.Valid() {
		return nil
	}
	attackTarget := a.battle.adjacentEnemy(u)
	out := make([]actions.Target, 0, 6)
	for d := game.DirLowerLeft; d <= game.DirUpperRight; d++ {
		cell, ok := u.Cell.Neighbour(d)
		if !ok {
			continue
		}
		dx, dy := cell.ScreenXY()
		action := actions.BattleMove(int(d))
		occupied := false
		for _, friend := range a.battle.sim.Attacker {
			if friend.Alive() && friend.Cell == cell {
				occupied = true
				break
			}
		}
		for _, enemy := range a.battle.sim.Defender {
			if enemy.Alive() && enemy.Cell == cell {
				occupied = true
				if enemy == attackTarget {
					action = actions.BattleAttack
				}
				break
			}
		}
		// Enter 固定攻擊掃描到的第一個相鄰敵軍；其他佔用格沒有鍵盤等價
		// 動作，所以不公告成可點目標。
		if occupied && action != actions.BattleAttack {
			continue
		}
		out = append(out, actions.Target{Action: action, Rect: actions.Rect{
			X: fieldX + dx, Y: fieldY + dy, W: game.HexCellW, H: game.HexCellH,
		}})
	}
	return out
}

func (a *app) numericKeypadVisible() bool {
	switch a.screen {
	case screenTransferTarget, screenTransferAmount, screenTradeAmount,
		screenSupplyTarget, screenSupplyAmount, screenRecruitAmount,
		screenReorganizeAmount, screenCovertTarget, screenMessageTime:
		return true
	case screenProduction:
		return a.productionItem != 0
	default:
		return false
	}
}

func (a *app) numericKeypadTargets() []actions.Target {
	if !a.numericKeypadVisible() {
		return nil
	}
	out := make([]actions.Target, 12)
	for i := range out {
		p := uilayout.NumericKeypadButton(i)
		action := actions.DeleteDigit
		switch {
		case i < 9:
			action = actions.Digit(i + 1)
		case i == 9:
			action = actions.Digit0
		case i == 11:
			action = actions.Submit
		}
		out[i] = actions.Target{Action: action,
			Rect: actions.Rect{X: p.HitX, Y: p.HitY, W: p.HitW, H: p.HitH}}
	}
	return out
}

func (a *app) actionPressed(action actions.Action, keys ...ebiten.Key) bool {
	if a.pointerAction == action {
		return true
	}
	for _, key := range keys {
		if inpututil.IsKeyJustPressed(key) {
			return true
		}
	}
	return false
}

func (a *app) digitPressed(digit int, key ebiten.Key) bool {
	return a.actionPressed(actions.Digit(digit), key)
}

func (a *app) deleteDigitPressed() bool {
	return a.actionPressed(actions.DeleteDigit, ebiten.KeyBackspace)
}

func (a *app) submitPressed() bool {
	return a.actionPressed(actions.Submit, ebiten.KeyEnter, ebiten.KeyKPEnter)
}

func (a *app) pointerTargets() []actions.Target {
	vertical := func(count, x, y, width, rowHeight int) []actions.Target {
		out := make([]actions.Target, count)
		for i := range out {
			out[i] = actions.Target{Action: actions.Selection(i + 1),
				Rect: actions.Rect{X: x, Y: y + i*rowHeight, W: width, H: rowHeight}}
		}
		return out
	}
	semanticList := func(count int) []actions.Target {
		return vertical(count, fieldX+24, fieldY+30, 426, 42)
	}
	originalList := func(count int) []actions.Target {
		return vertical(count, fieldX+24, fieldY+22, 426, 28)
	}

	switch a.screen {
	case screenMap:
		p := uilayout.OpenCommandButton(640)
		return []actions.Target{{Action: actions.OpenCommands,
			Rect: actions.Rect{X: p.HitX, Y: p.HitY, W: p.HitW, H: p.HitH}}}

	case screenCommand:
		rowH, y0 := 20, fieldY+16
		if a.wordingMode == i18n.WordingPlain {
			rowH, y0 = 38, fieldY+30
		}
		out := make([]actions.Target, 15)
		for i := range out {
			p := uilayout.Grid(i, fieldX+14, y0, 8, 210, rowH, 0, 205, rowH)
			out[i] = actions.Target{Action: actions.Selection(i + 1), Rect: actions.Rect{X: p.HitX, Y: p.HitY, W: p.HitW, H: p.HitH}}
		}
		return out

	case screenDevelop:
		return vertical(3, fieldX+14, fieldY+22, 425, 38)

	case screenTransferMode:
		if a.wordingMode == i18n.WordingPlain {
			return vertical(2, fieldX+24, fieldY+64, 426, 46)
		}
		return vertical(2, fieldX+24, fieldY+26, 426, 34)

	case screenTransferTarget:
		return vertical(len(a.transferTargets), fieldX+24, fieldY+56, 426, 24)

	case screenSupplyTarget:
		return vertical(len(a.supplyTargets), fieldX+24, fieldY+56, 426, 24)

	case screenTransferSelection:
		if a.transferSession == nil {
			return nil
		}
		cands := a.transferSession.Candidates()
		start := (a.transferCursor / 20) * 20
		end := start + 20
		if end > len(cands) {
			end = len(cands)
		}
		out := make([]actions.Target, 0, end-start)
		for i := start; i < end; i++ {
			row := i - start
			cx := fieldX + 26 + (row/10)*205
			cy := fieldY + 50 + (row%10)*24
			out = append(out, actions.Target{Action: actions.Selection(i + 1),
				Rect: actions.Rect{X: cx - 8, Y: cy - 3, W: 190, H: 22}})
		}
		return out

	case screenTradeMode:
		if a.wordingMode == i18n.WordingPlain {
			return semanticList(2)
		}
		return originalList(2)

	case screenTradeGood:
		count := 3
		if !a.tradeImport {
			count = 5
		}
		if a.wordingMode == i18n.WordingPlain {
			return semanticList(count)
		}
		return originalList(count)

	case screenRecruitAction:
		if a.wordingMode == i18n.WordingPlain {
			return vertical(2, fieldX+24, fieldY+40, 426, 52)
		}
		return vertical(2, fieldX+24, fieldY+26, 426, 32)

	case screenRecruitBranch, screenReorganizeBranch:
		if a.wordingMode == i18n.WordingPlain {
			return vertical(4, fieldX+24, fieldY+30, 426, 46)
		}
		return originalList(4)

	case screenCovertAction:
		if a.wordingMode == i18n.WordingPlain {
			return semanticList(2)
		}
		return vertical(2, fieldX+24, fieldY+22, 426, 32)

	case screenViewMenu:
		if a.wordingMode == i18n.WordingPlain {
			return semanticList(4)
		}
		return vertical(4, fieldX+24, fieldY+38, 426, 52)

	case screenViewGenerals:
		start := (a.viewIndex / 20) * 20
		end := start + 20
		if end > len(a.viewGenerals) {
			end = len(a.viewGenerals)
		}
		out := make([]actions.Target, 0, end-start)
		for i := start; i < end; i++ {
			visible := i - start
			cx := fieldX + 24 + (visible/10)*215
			cy := fieldY + 52 + (visible%10)*27
			out = append(out, actions.Target{Action: actions.Selection(i + 1),
				Rect: actions.Rect{X: cx - 8, Y: cy - 4, W: 198, H: 25}})
		}
		return out

	case screenViewProvinceChoice:
		return vertical(2, fieldX+24, fieldY+62, 426, 56)

	case screenViewProvinceSelect:
		out := make([]actions.Target, 39)
		for i := range out {
			col, line := i/13, i%13
			cx, cy := fieldX+18+col*145, fieldY+48+line*22
			out[i] = actions.Target{Action: actions.Selection(i + 1),
				Rect: actions.Rect{X: cx, Y: cy - 3, W: 142, H: 21}}
		}
		return out

	case screenReorganizeTarget:
		if a.reorganization == nil {
			return nil
		}
		targets := a.reorganization.Targets()
		out := make([]actions.Target, len(targets))
		for i := range out {
			cx := fieldX + 30 + (i/10)*205
			cy := fieldY + 56 + (i%10)*24
			out[i] = actions.Target{Action: actions.Selection(i + 1),
				Rect: actions.Rect{X: cx - 6, Y: cy - 3, W: 195, H: 22}}
		}
		return out

	case screenAutonomy:
		out := make([]actions.Target, len(a.autonomyTargets))
		for i := range out {
			col, line := i/19, i%19
			cx, cy := fieldX+col*225+10, fieldY+60+line*14
			out[i] = actions.Target{Action: actions.Selection(i + 1),
				Rect: actions.Rect{X: cx, Y: cy - 2, W: 220, H: 14}}
		}
		return out

	case screenPolicy:
		return vertical(2, fieldX+24, fieldY+66, 426, 52)

	case screenProduction:
		if a.productionItem != 0 {
			return nil
		}
		return vertical(4, fieldX+18, fieldY+100, 414, 42)

	case screenOtherOptions:
		out := make([]actions.Target, 0, 10)
		for i := 0; i < 9; i++ {
			p := uilayout.Grid(i, fieldX+12, fieldY+30, 5, 215, 55, 8, 210, 54)
			out = append(out, actions.Target{Action: actions.Selection(i + 1), Rect: actions.Rect{X: p.HitX, Y: p.HitY, W: p.HitW, H: p.HitH}})
		}
		return append(out, actions.Target{Action: actions.Back,
			Rect: actions.Rect{X: fieldX + 12, Y: 300, W: 160, H: 50}})

	case screenDisplayOptions:
		w := 640 - fieldX
		p1 := uilayout.Grid(0, fieldX+28, 112, 2, 0, 52, 15, w-56, 48)
		p2 := uilayout.Grid(1, fieldX+28, 112, 2, 0, 52, 15, w-56, 48)
		return []actions.Target{
			{Action: actions.Select1, Rect: actions.Rect{X: p1.HitX, Y: p1.HitY, W: p1.HitW, H: p1.HitH}},
			{Action: actions.Select2, Rect: actions.Rect{X: p2.HitX, Y: p2.HitY, W: p2.HitW, H: p2.HitH}},
			{Action: actions.Back, Rect: actions.Rect{X: fieldX + 16, Y: 292, W: 180, H: 58}},
		}

	case screenTransferConfirm, screenRecruitConfirm, screenTrainConfirm,
		screenSaveConfirm, screenLoadConfirm, screenQuit:
		return []actions.Target{
			{Action: actions.Confirm, Rect: actions.Rect{X: fieldX + 45, Y: 85, W: 165, H: 90}},
			{Action: actions.Cancel, Rect: actions.Rect{X: fieldX + 220, Y: 85, W: 165, H: 90}},
		}

	case screenMessageTime:
		return []actions.Target{{Action: actions.Back, Rect: actions.Rect{X: fieldX + 16, Y: 292, W: 180, H: 58}}}
	}
	return nil
}
