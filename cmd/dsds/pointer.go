package main

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
	uiresolution "github.com/wicanr2/great-era-remake/internal/ui/resolution"
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
// retro／非 Modern 高解析頁面沿用 640×350 命中矩形並扣除上下安全邊界；Modern
// 高解析 renderer 直接使用 1280×720 layout。兩條路徑都只送同一套 Action 識別碼。
func (a *app) collectPointerAction() actions.Action {
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		if x, y, ok := a.basePointerPosition(x, y); ok {
			a.pointer.mouse = &pointerPress{x: x, y: y, screen: a.screen}
		} else {
			a.pointer.mouse = nil
		}
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()
		start := a.pointer.mouse
		a.pointer.mouse = nil
		if x, y, ok := a.basePointerPosition(x, y); ok && start != nil && withinClick(*start, x, y, a.screen) {
			return actions.Hit(a.interactiveTargets(), x, y)
		}
	}

	pressed := inpututil.AppendJustPressedTouchIDs(nil)
	if len(pressed) > 0 && a.pointer.touches == nil {
		a.pointer.touches = map[ebiten.TouchID]pointerPress{}
	}
	for _, id := range pressed {
		x, y := ebiten.TouchPosition(id)
		if x, y, ok := a.basePointerPosition(x, y); ok {
			a.pointer.touches[id] = pointerPress{x: x, y: y, screen: a.screen}
		}
	}
	for _, id := range inpututil.AppendJustReleasedTouchIDs(nil) {
		start, ok := a.pointer.touches[id]
		delete(a.pointer.touches, id)
		if !ok {
			continue
		}
		x, y := ebiten.TouchPosition(id)
		if x, y, ok := a.basePointerPosition(x, y); ok && withinClick(start, x, y, a.screen) {
			if action := actions.Hit(a.interactiveTargets(), x, y); action != actions.None {
				return action
			}
		}
	}
	return actions.None
}

func (a *app) basePointerPosition(x, y int) (int, int, bool) {
	if a.highModernSurface() {
		// 高解析 Modern 頁面直接使用 1280×720 設計座標；retro 或非 Modern
		// 頁面才走下面的 Surface→640×350 反算。
		return x, y, true
	}
	if a.resolution == uiresolution.ModeHigh {
		return uiresolution.SurfaceToBase(x, y, uiresolution.HighWidth, uiresolution.HighHeight)
	}
	return x, y, true
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
	case screenNarrative:
		out := []actions.Action{actions.Back}
		pages := render.NarrativePageCount(a.narrativeImages)
		if a.narrativePage+1 < pages {
			out = append(out, actions.NextPage)
		}
		if a.narrativePage > 0 {
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
	if a.highModernFlow() {
		page, err := uilayout.NewModernPageLayout(uilayout.ModernDesignSurface())
		if err != nil {
			return nil
		}
		out := []actions.Target{{Action: actions.Back,
			Rect: actions.Rect{X: page.Back.X, Y: page.Back.Y, W: page.Back.W, H: page.Back.H}}}
		if a.screen == screenBiography || a.screen == screenNarrative {
			if (a.screen == screenBiography && a.bioPage > 0) || (a.screen == screenNarrative && a.narrativePage > 0) {
				out = append(out, actions.Target{Action: actions.PreviousPage,
					Rect: actions.Rect{X: page.Previous.X, Y: page.Previous.Y, W: page.Previous.W, H: page.Previous.H}})
			}
			if (a.screen == screenBiography && a.bioPage+1 < a.bioPages) ||
				(a.screen == screenNarrative && a.narrativePage+1 < render.NarrativePageCount(a.narrativeImages)) {
				out = append(out, actions.Target{Action: actions.NextPage,
					Rect: actions.Rect{X: page.Next.X, Y: page.Next.Y, W: page.Next.W, H: page.Next.H}})
			}
		} else if a.screen == screenViewGeneral {
			if _, ok := a.currentBiography(); ok {
				out = append(out, actions.Target{Action: actions.OpenBiography,
					Rect: actions.Rect{X: page.Next.X, Y: page.Next.Y, W: page.Next.W, H: page.Next.H}})
			}
		}
		return out
	}
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
	if a.highModernBattle() {
		return a.highModernBattlefieldTargets()
	}
	out := make([]actions.Target, 0, 16)
	if a.battle.mode == battleModeRetreat {
		for i := 0; i < 12; i++ {
			p := uilayout.BattleRetreatKeypadButton(i)
			action := actions.DeleteDigit
			switch {
			case i < 9:
				action = actions.Digit(i + 1)
			case i == 9:
				action = actions.Digit0
			case i == 11:
				action = actions.Submit
			}
			out = append(out, actions.Target{Action: action,
				Rect: actions.Rect{X: p.HitX, Y: p.HitY, W: p.HitW, H: p.HitH}})
		}
		return out
	}
	if a.battle.mode == battleModeCommand {
		for i := 0; i < 5; i++ {
			p := uilayout.BattleCommandButton(i)
			out = append(out, actions.Target{
				Action: actions.BattleCommand(i + 1),
				Rect:   actions.Rect{X: p.HitX, Y: p.HitY, W: p.HitW, H: p.HitH},
			})
		}
		for i, action := range []actions.Action{
			actions.BattleAttack, actions.BattleNextUnit, actions.BattleEndTurn,
		} {
			p := uilayout.BattleControlButton(i)
			out = append(out, actions.Target{
				Action: action,
				Rect:   actions.Rect{X: p.HitX, Y: p.HitY, W: p.HitW, H: p.HitH},
			})
		}
	}
	u := a.battle.current()
	if u == nil || !u.Cell.Valid() {
		return out
	}
	attackTarget := a.battle.adjacentEnemy(u)
	attackTargets := a.battle.battleAttackTargets(u)
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
				if a.battle.mode == battleModeAttack {
					for i, candidate := range attackTargets {
						if candidate == enemy && i < 6 {
							action = actions.BattleAttackTarget(i + 1)
							break
						}
					}
				} else if enemy == attackTarget {
					action = actions.BattleAttack
				}
				break
			}
		}
		if a.battle.mode == battleModeAttack && !occupied {
			continue
		}
		// 主選單的 Enter 固定攻擊掃描到的第一個相鄰敵軍；攻擊
		// 子狀態則公告 `battle.attack-target.N`，不能被這個主選單
		// 的「只有 Enter 可點」護欄過濾掉。
		if a.battle.mode == battleModeCommand && occupied && action != actions.BattleAttack {
			continue
		}
		out = append(out, actions.Target{Action: action, Rect: actions.Rect{
			X: fieldX + dx, Y: fieldY + dy, W: game.HexCellW, H: game.HexCellH,
		}})
	}
	// 兵種 4 的遠程候選不一定落在六個相鄰格；攻擊子狀態仍用同一份
	// 1..6 目標表，讓桌面滑鼠與 Android 觸控可直接點選非相鄰目標。
	if a.battle.mode == battleModeAttack {
		for i, enemy := range attackTargets {
			if i >= 6 || enemy == nil || !enemy.Alive() || game.Adjacent(u.Cell, enemy.Cell) {
				continue
			}
			dx, dy := enemy.Cell.ScreenXY()
			out = append(out, actions.Target{
				Action: actions.BattleAttackTarget(i + 1),
				Rect:   actions.Rect{X: fieldX + dx, Y: fieldY + dy, W: game.HexCellW, H: game.HexCellH},
			})
		}
	}
	return out
}

func (a *app) highModernBattlefieldTargets() []actions.Target {
	l, err := uilayout.NewModernBattleLayout(uilayout.ModernDesignSurface())
	if err != nil || a.battle == nil || a.battle.finished {
		return nil
	}
	out := make([]actions.Target, 0, 24)
	if a.battle.mode == battleModeRetreat {
		for i, key := range l.RetreatButtons {
			action := actions.DeleteDigit
			switch {
			case i < 9:
				action = actions.Digit(i + 1)
			case i == 9:
				action = actions.Digit0
			case i == 11:
				action = actions.Submit
			}
			out = append(out, actions.Target{Action: action,
				Rect: actions.Rect{X: key.X, Y: key.Y, W: key.W, H: key.H}})
		}
		return out
	}
	if a.battle.mode == battleModeCommand {
		for i, card := range l.CommandButtons {
			out = append(out, actions.Target{Action: actions.BattleCommand(i + 1),
				Rect: actions.Rect{X: card.X, Y: card.Y, W: card.W, H: card.H}})
		}
		for i, action := range []actions.Action{actions.BattleAttack, actions.BattleNextUnit, actions.BattleEndTurn} {
			card := l.ControlButtons[i]
			out = append(out, actions.Target{Action: action,
				Rect: actions.Rect{X: card.X, Y: card.Y, W: card.W, H: card.H}})
		}
	}
	u := a.battle.current()
	if u == nil || !u.Cell.Valid() {
		return out
	}
	attackTarget := a.battle.adjacentEnemy(u)
	attackTargets := a.battle.battleAttackTargets(u)
	for d := game.DirLowerLeft; d <= game.DirUpperRight; d++ {
		cell, ok := u.Cell.Neighbour(d)
		if !ok {
			continue
		}
		occupied := false
		action := actions.BattleMove(int(d))
		for _, friend := range a.battle.sim.Attacker {
			if friend.Alive() && friend.Cell == cell {
				occupied = true
				break
			}
		}
		for _, enemy := range a.battle.sim.Defender {
			if enemy.Alive() && enemy.Cell == cell {
				occupied = true
				if a.battle.mode == battleModeAttack {
					for i, candidate := range attackTargets {
						if candidate == enemy && i < 6 {
							action = actions.BattleAttackTarget(i + 1)
							break
						}
					}
				} else if enemy == attackTarget {
					action = actions.BattleAttack
				}
				break
			}
		}
		if a.battle.mode == battleModeAttack && !occupied {
			continue
		}
		if a.battle.mode == battleModeCommand && occupied && action != actions.BattleAttack {
			continue
		}
		col, row := cell.ColRow()
		p, rectErr := l.MapCellRect(col, row)
		if rectErr != nil {
			continue
		}
		out = append(out, actions.Target{Action: action,
			Rect: actions.Rect{X: p.X, Y: p.Y, W: p.W, H: p.H}})
	}
	if a.battle.mode == battleModeAttack {
		for i, enemy := range attackTargets {
			if i >= 6 || enemy == nil || !enemy.Alive() || game.Adjacent(u.Cell, enemy.Cell) {
				continue
			}
			col, row := enemy.Cell.ColRow()
			p, rectErr := l.MapCellRect(col, row)
			if rectErr != nil {
				continue
			}
			out = append(out, actions.Target{Action: actions.BattleAttackTarget(i + 1),
				Rect: actions.Rect{X: p.X, Y: p.Y, W: p.W, H: p.H}})
		}
	}
	return out
}

func (a *app) numericKeypadVisible() bool {
	switch a.screen {
	case screenTransferTarget, screenTransferAmount, screenTradeAmount,
		screenSupplyTarget, screenSupplyAmount, screenRecruitAmount,
		screenReorganizeAmount, screenCovertTarget, screenLoanAmount,
		screenRepayAmount, screenCeasefireTarget, screenMessageTime:
		return true
	case screenProduction:
		return a.productionItem != 0
	default:
		return false
	}
}

func (a *app) numericKeypadTargets() []actions.Target {
	if a.highModernFlow() {
		return nil
	}
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
			a.playInputEffect(action)
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
	if a.highModernFlow() && !a.highModernPage() {
		return a.modernFlowTargets()
	}

	switch a.screen {
	case screenMap:
		if a.highModernMap() {
			l, err := uilayout.NewModernMapLayout(uilayout.ModernDesignSurface())
			if err != nil {
				return nil
			}
			out := []actions.Target{
				{Action: actions.OpenCommands, Rect: actions.Rect{X: l.CommandButton.X, Y: l.CommandButton.Y,
					W: l.CommandButton.W, H: l.CommandButton.H}},
				{Action: actions.OpenCommands, Rect: actions.Rect{X: l.CommandRailButton.X, Y: l.CommandRailButton.Y,
					W: l.CommandRailButton.W, H: l.CommandRailButton.H}},
			}
			if a.narrative != nil && len(a.narrativeImages) > 0 {
				out = append(out, actions.Target{Action: actions.OpenNarrative,
					Rect: actions.Rect{X: l.NarrativeButton.X, Y: l.NarrativeButton.Y,
						W: l.NarrativeButton.W, H: l.NarrativeButton.H}}, actions.Target{Action: actions.OpenNarrative,
					Rect: actions.Rect{X: l.NarrativeRailButton.X, Y: l.NarrativeRailButton.Y,
						W: l.NarrativeRailButton.W, H: l.NarrativeRailButton.H}})
			}
			return out
		}
		p := uilayout.OpenCommandButton(640)
		out := []actions.Target{{Action: actions.OpenCommands,
			Rect: actions.Rect{X: p.HitX, Y: p.HitY, W: p.HitW, H: p.HitH}}}
		if a.narrative != nil && len(a.narrativeImages) > 0 {
			n := uilayout.NarrativeButton(640)
			out = append(out, actions.Target{Action: actions.OpenNarrative,
				Rect: actions.Rect{X: n.HitX, Y: n.HitY, W: n.HitW, H: n.HitH}})
		}
		return out

	case screenCommand:
		if a.highModernPage() {
			l, err := uilayout.NewModernCommandLayout(uilayout.ModernDesignSurface())
			if err != nil {
				return nil
			}
			out := make([]actions.Target, 0, len(l.Cards))
			for i, card := range l.Cards {
				out = append(out, actions.Target{Action: actions.Selection(i + 1),
					Rect: actions.Rect{X: card.X, Y: card.Y, W: card.W, H: card.H}})
			}
			return out
		}
		rowH, y0 := 20, fieldY+16
		if a.semanticWording() {
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
		if a.semanticWording() {
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
		if a.semanticWording() {
			return semanticList(2)
		}
		return originalList(2)

	case screenTradeGood:
		count := 3
		if !a.tradeImport {
			count = 5
		}
		if a.semanticWording() {
			return semanticList(count)
		}
		return originalList(count)

	case screenRecruitAction:
		if a.semanticWording() {
			return vertical(2, fieldX+24, fieldY+40, 426, 52)
		}
		return vertical(2, fieldX+24, fieldY+26, 426, 32)

	case screenRecruitBranch, screenReorganizeBranch:
		if a.semanticWording() {
			return vertical(4, fieldX+24, fieldY+30, 426, 46)
		}
		return originalList(4)

	case screenCovertAction:
		if a.semanticWording() {
			return semanticList(2)
		}
		return vertical(2, fieldX+24, fieldY+22, 426, 32)

	case screenDiplomacy:
		return semanticList(3)

	case screenViewMenu:
		if a.semanticWording() {
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

	case screenViewGeneral:
		// 自傳按鈕只有在同一張畫面確實能畫出語系字與倚天字庫時才公告；
		// 排除槽（例如「無省長」）也不公告隱形命中區；unknown 人物仍
		// 有入口，頁面會顯示「查無可靠傳記記載」而非空白。
		if a.eten == nil || a.wording == nil {
			return nil
		}
		if _, ok := a.currentBiography(); !ok {
			return nil
		}
		p := uilayout.BiographyButton(640, a.navigationY())
		return []actions.Target{{Action: actions.OpenBiography,
			Rect: actions.Rect{X: p.HitX, Y: p.HitY, W: p.HitW, H: p.HitH}}}

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
		if a.highModernPage() {
			l, err := uilayout.NewModernOptionLayout(uilayout.ModernDesignSurface(), 10)
			if err != nil {
				return nil
			}
			out := make([]actions.Target, 0, len(l.Options))
			for i, option := range l.Options {
				out = append(out, actions.Target{Action: actions.Selection(i + 1),
					Rect: actions.Rect{X: option.X, Y: option.Y, W: option.W, H: option.H}})
			}
			return out
		}
		out := make([]actions.Target, 0, 11)
		for i := 0; i < 10; i++ {
			p := uilayout.Grid(i, fieldX+12, fieldY+30, 5, 215, 55, 8, 210, 54)
			out = append(out, actions.Target{Action: actions.Selection(i + 1), Rect: actions.Rect{X: p.HitX, Y: p.HitY, W: p.HitW, H: p.HitH}})
		}
		return append(out, actions.Target{Action: actions.Back,
			Rect: actions.Rect{X: fieldX + 12, Y: 300, W: 160, H: 50}})

	case screenDisplayOptions:
		if a.highModernPage() {
			l, err := uilayout.NewModernOptionLayout(uilayout.ModernDesignSurface(), 4)
			if err != nil {
				return nil
			}
			return []actions.Target{
				{Action: actions.Select1, Rect: actions.Rect{X: l.Options[0].X, Y: l.Options[0].Y, W: l.Options[0].W, H: l.Options[0].H}},
				{Action: actions.Select2, Rect: actions.Rect{X: l.Options[1].X, Y: l.Options[1].Y, W: l.Options[1].W, H: l.Options[1].H}},
				{Action: actions.Select3, Rect: actions.Rect{X: l.Options[2].X, Y: l.Options[2].Y, W: l.Options[2].W, H: l.Options[2].H}},
				{Action: actions.Select4, Rect: actions.Rect{X: l.Options[3].X, Y: l.Options[3].Y, W: l.Options[3].W, H: l.Options[3].H}},
			}
		}
		w := 640 - fieldX
		p1 := uilayout.DisplayWordingOption(0, fieldX+28, fieldY, w-56)
		p2 := uilayout.DisplayWordingOption(1, fieldX+28, fieldY, w-56)
		p3 := uilayout.DisplayThemeOption(0, fieldX+28, fieldY, w-56)
		p4 := uilayout.DisplayThemeOption(1, fieldX+28, fieldY, w-56)
		return []actions.Target{
			{Action: actions.Select1, Rect: actions.Rect{X: p1.HitX, Y: p1.HitY, W: p1.HitW, H: p1.HitH}},
			{Action: actions.Select2, Rect: actions.Rect{X: p2.HitX, Y: p2.HitY, W: p2.HitW, H: p2.HitH}},
			{Action: actions.Select3, Rect: actions.Rect{X: p3.HitX, Y: p3.HitY, W: p3.HitW, H: p3.HitH}},
			{Action: actions.Select4, Rect: actions.Rect{X: p4.HitX, Y: p4.HitY, W: p4.HitW, H: p4.HitH}},
			{Action: actions.Back, Rect: actions.Rect{X: fieldX + 16, Y: 292, W: 180, H: 58}},
		}

	case screenResolutionOptions:
		if a.highModernPage() {
			l, err := uilayout.NewModernOptionLayout(uilayout.ModernDesignSurface(), 2)
			if err != nil {
				return nil
			}
			return []actions.Target{
				{Action: actions.Select1, Rect: actions.Rect{X: l.Options[0].X, Y: l.Options[0].Y, W: l.Options[0].W, H: l.Options[0].H}},
				{Action: actions.Select2, Rect: actions.Rect{X: l.Options[1].X, Y: l.Options[1].Y, W: l.Options[1].W, H: l.Options[1].H}},
			}
		}
		return append(vertical(2, fieldX+24, fieldY+64, 426, 52),
			actions.Target{Action: actions.Back, Rect: actions.Rect{X: fieldX + 16, Y: 292, W: 180, H: 58}})

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
