package main

import (
	"fmt"
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	"github.com/wicanr2/great-era-remake/internal/ui/atlasmap"
	"github.com/wicanr2/great-era-remake/internal/ui/hdart"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
	t "github.com/wicanr2/great-era-remake/internal/ui/theme"
	"strconv"
	"strings"
)

type modernDeskState struct {
	Press, Focus, Picker, Pending, Message, Slider string
	PressScreen, Return                            screen
	PressX, PressY, Scroll                         int
	Commands, Drag                                 bool
	DragX, DragY                                   int
	Screen                                         screen
	Known                                          bool
}

func (a *app) modernDeskVisible() bool {
	return a != nil && a.commandInterface() && a.opening == nil && a.world != nil && a.eten != nil && a.wording != nil && !a.armorySurface() && a.screen != screenBattle && a.screen != screenNarrative && !a.campaignEventSurface() && !(a.screen == screenQuit && a.quitBack == screenBattle)
}
func (a *app) deskText(key string) string {
	v, e := a.wordingText("desk." + key)
	if e != nil {
		return "[desk." + key + "]"
	}
	return v
}
func (a *app) deskBranchName(branch uint8) string {
	key := map[uint8]string{game.BranchInfantry: "recruit.infantry", game.BranchArmour: "recruit.armour", game.BranchArtiller: "recruit.artillery", game.BranchCavalry: "recruit.cavalry"}[branch]
	value, _ := a.wordingText(key)
	return value
}
func (a *app) deskGoodName(good game.TradeGood) string {
	key := map[game.TradeGood]string{game.GoodFood: "trade.food", game.GoodAmmo: "trade.ammo", game.GoodFuel: "trade.fuel", game.GoodCoal: "trade.coal", game.GoodIron: "trade.iron"}[good]
	value, _ := a.wordingText(key)
	return value
}
func deskNumber(v int) string {
	s := strconv.Itoa(v)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
func (a *app) deskUnitIDs() []game.GeneralID { return a.world.ActiveGeneralsAt(a.current) }
func (a *app) deskSetScreen(s screen) {
	width, height := a.gameSurfaceSize()
	a.screen = s
	a.desk.Scroll = 0
	a.desk.Focus = ""
	a.desk.Press = ""
	a.desk.Picker = ""
	a.dirty = true
	nextWidth, nextHeight := a.gameSurfaceSize()
	if width != nextWidth || height != nextHeight {
		a.applyWindowSize()
	}
}
func (a *app) deskStatus(s string) { a.desk.Message = s; a.dirty = true }
func (a *app) deskInitialize() {
	switch a.screen {
	case screenRecruitAction, screenRecruitBranch:
		if a.recruitBranch == 0 {
			a.recruitBranch = game.BranchInfantry
		}
		a.recruitLimit = a.world.RecruitLimit(a.current, a.recruitBranch)
		a.recruitAmount = min(a.recruitAmount, uint32(a.recruitLimit))
		a.screen = screenRecruitAmount
	case screenTransferMode, screenTransferTarget:
		if a.transferMode == 0 {
			a.transferMode = game.PlayerTransferPartial
		}
		a.transferTargets, _ = a.world.PlayerTransferTargets(a.current)
		if a.transferSession == nil && len(a.transferTargets) > 0 {
			a.deskTransferTarget(a.transferTargets[0])
		}
		a.screen = screenTransferSelection
	case screenSupplyTarget:
		a.supplyTargets, _ = a.world.SupplyTargets(a.current)
		if len(a.supplyTargets) > 0 && !provinceIn(a.supplyTargets, a.supplyTarget) {
			a.supplyTarget = a.supplyTargets[0]
		}
		a.screen = screenSupplyAmount
	case screenTradeMode, screenTradeGood:
		if a.tradeGood == 0 {
			a.tradeGood = game.GoodFood
		}
		a.screen = screenTradeAmount
	}
}
func (a *app) deskTransferTarget(id game.ProvinceID) {
	s, e := a.world.BeginPlayerTransfer(a.current, id, a.transferMode, a.generals)
	if e != nil {
		a.deskStatus(e.Error())
		return
	}
	a.transferSession = s
	a.transferAmounts = [4]int{}
	a.transferCursor = 0
}
func (a *app) deskBeginReorganization(branch uint8) error {
	remaining := a.cmdBudget.Peek(a.current)
	if remaining <= 0 {
		return fmt.Errorf("%s", a.deskText("no.commands"))
	}
	if a.armory != nil {
		if e := a.armory.Sync(a.world); e != nil {
			return e
		}
	}
	r, e := a.world.BeginReorganization(a.current, branch)
	if e != nil {
		return e
	}
	a.reorganization = r
	a.reorganizeInput = 0
	a.reorganizeID = 0
	a.deskSetScreen(screenReorganizeTarget)
	return nil
}
func (a *app) cancelReorganization() error {
	if a.reorganization == nil {
		return nil
	}
	if e := a.reorganization.Cancel(); e != nil {
		return e
	}
	a.syncReorganizationGenerals()
	a.reorganization = nil
	a.reorganizeInput = 0
	a.reorganizeID = 0
	a.deskSetScreen(screenMap)
	return nil
}
func (a *app) deskRequestBack(home bool) (bool, error) {
	if a.desk.Picker != "" {
		a.desk.Picker = ""
		a.desk.Scroll = 0
		a.dirty = true
		return true, nil
	}
	if a.desk.Pending != "" {
		a.desk.Pending = ""
		a.desk.Focus = ""
		a.dirty = true
		return true, nil
	}
	if a.reorganization != nil && (a.screen == screenReorganizeTarget || a.screen == screenReorganizeAmount) {
		if !a.reorganization.Changed() && a.reorganization.Remaining() == 0 {
			a.reorganization = nil
			a.deskSetScreen(screenMap)
			return true, nil
		}
		a.deskSetScreen(screenReorganizeCancel)
		return true, nil
	}
	if home && a.screen != screenReorganizeCancel {
		a.transferSession = nil
		a.desk.Commands = false
		a.deskSetScreen(screenMap)
		return true, nil
	}
	switch a.screen {
	case screenQuit:
		a.deskSetScreen(a.quitBack)
		return true, nil
	case screenReorganizeCancel:
		a.deskSetScreen(screenReorganizeTarget)
		return true, nil
	case screenSaveConfirm, screenLoadConfirm:
		a.deskSetScreen(a.desk.Return)
		return true, nil
	case screenRecruitConfirm:
		a.deskSetScreen(screenRecruitAmount)
		return true, nil
	case screenTransferConfirm:
		a.deskSetScreen(screenTransferSelection)
		return true, nil
	case screenRecruitAmount, screenTransferSelection, screenTransferAmount, screenSupplyAmount, screenTradeAmount:
		a.transferSession = nil
		a.deskSetScreen(screenMap)
		return true, nil
	case screenMap:
		a.desk.Commands = false
		return true, nil
	case screenCommand:
		a.desk.Commands = false
		a.deskSetScreen(screenMap)
		return true, nil
	}
	if home {
		a.deskSetScreen(screenMap)
		return true, nil
	}
	a.pointerAction = a.pointerBackAction()
	return false, nil
}
func (a *app) deskOpenPage(page string) error {
	// Repeated save/load shortcuts must not replace the original return page
	// with the confirmation dialog itself while a slow frame is being drawn.
	if page == "save" && a.screen == screenSaveConfirm || page == "load" && a.screen == screenLoadConfirm {
		return nil
	}
	if page != "save" && a.reorganization != nil && (a.reorganization.Changed() || a.reorganization.Remaining() > 0) {
		a.deskSetScreen(screenReorganizeCancel)
		return nil
	}
	if page != "save" && a.reorganization != nil {
		a.reorganization = nil
	}
	a.desk.Pending = ""
	a.desk.Picker = ""
	a.desk.Scroll = 0
	a.desk.Commands = false
	if a.strategyLocked() && page != "map" && page != "people" && page != "resume" && page != "save" && page != "load" && !(page == "commands" && !a.activeBattle()) {
		a.deskStatus(a.strategyLockText())
		return nil
	}
	switch page {
	case "map":
		a.transferSession = nil
		a.deskSetScreen(screenMap)
	case "commands":
		a.desk.Commands = true
		a.deskSetScreen(screenCommand)
	case "recruit":
		if !a.ownsProvince(a.current) {
			return nil
		}
		a.recruitBranch = game.BranchInfantry
		a.recruitLimit = a.world.RecruitLimit(a.current, a.recruitBranch)
		a.recruitAmount = uint32(min(500, a.recruitLimit))
		a.deskSetScreen(screenRecruitAmount)
	case "transfer":
		if !a.ownsProvince(a.current) {
			return nil
		}
		a.transferSession = nil
		a.transferMode = game.PlayerTransferPartial
		a.deskSetScreen(screenTransferTarget)
		a.deskInitialize()
	case "reorganize":
		if !a.ownsProvince(a.current) {
			return nil
		}
		return a.deskBeginReorganization(game.BranchInfantry)
	case "people":
		a.viewGenerals = a.world.ActiveGeneralsAt(a.current)
		a.viewIndex = 0
		a.viewGeneralBack = screenMap
		a.deskSetScreen(screenViewGenerals)
	case "save":
		if !a.canSaveSession() {
			a.deskStatus(a.deskText("save.unavailable"))
			return nil
		}
		a.desk.Return = a.screen
		a.deskSetScreen(screenSaveConfirm)
	case "load":
		a.desk.Return = a.screen
		a.deskSetScreen(screenLoadConfirm)
	case "resume":
		a.resumeBattle()
	case "armory":
		a.openArmory()
	}
	return nil
}
func (a *app) deskNumeric() (int, int, func(int)) {
	switch a.screen {
	case screenRecruitAmount, screenRecruitConfirm:
		limit := a.world.RecruitLimit(a.current, a.recruitBranch)
		return int(a.recruitAmount), limit, func(n int) { a.recruitAmount = uint32(n) }
	case screenReorganizeAmount:
		limit := a.reorganization.Limit(a.reorganizeID)
		return int(a.reorganizeInput), limit, func(n int) { a.reorganizeInput = uint32(n) }
	case screenTradeAmount:
		return int(a.tradeAmount), a.deskTradeLimit(), func(n int) { a.tradeAmount = uint32(n) }
	case screenCovertTarget:
		return int(a.covertInput), 39, func(n int) { a.covertInput = uint32(n) }
	case screenCeasefireTarget:
		return int(a.ceasefireInput), 39, func(n int) { a.ceasefireInput = uint32(n) }
	case screenLoanAmount:
		return int(a.diplomacyInput), int(a.ledger.Credit[a.diplomacySlot]), func(n int) { a.diplomacyInput = uint32(n) }
	case screenRepayAmount:
		p, _ := a.tbl.At(a.current)
		return int(a.diplomacyInput), min(int(p.Gold), int(a.ledger.Debt[a.diplomacySlot])), func(n int) { a.diplomacyInput = uint32(n) }
	case screenMessageTime:
		return int(a.messageTimeInput), 10, func(n int) { a.messageTimeInput = uint32(n) }
	case screenProduction:
		return int(a.productionInput), 100, func(n int) { a.productionInput = uint32(n) }
	case screenAutonomy:
		return int(a.autonomyInput), 39, func(n int) { a.autonomyInput = uint32(n) }
	}
	return 0, 0, func(int) {}
}
func (a *app) deskTradeLimit() int {
	lo, hi := 0, 65535
	for lo < hi {
		n := (lo + hi + 1) / 2
		table := *a.tbl
		world := *a.world
		world.Table = &table
		var e error
		var result game.TradeResult
		if a.tradeImport {
			result, e = world.Import(a.current, a.tradeGood, n)
		} else {
			result, e = world.Export(a.current, a.tradeGood, n)
		}
		if e == nil && result.Amount == n && result.Gold == game.TradePrice(a.tradeGood, n) {
			lo = n
		} else {
			hi = n - 1
		}
	}
	return lo
}
func (a *app) deskResource(index int) (int, int, func(int)) {
	if a.screen == screenSupplyAmount {
		return a.supplyAmounts[index], a.supplyLimit(index), func(n int) { a.supplyAmounts[index] = n }
	}
	return a.transferAmounts[index], a.transferSupplyLimit(index), func(n int) { a.transferAmounts[index] = n }
}
func (a *app) deskAdjust(key string, delta int, setValue *int) {
	var value, limit int
	var set func(int)
	if strings.HasPrefix(key, "resource.") {
		i, _ := strconv.Atoi(strings.TrimPrefix(key, "resource."))
		if i < 0 || i > 3 {
			return
		}
		value, limit, set = a.deskResource(i)
	} else {
		value, limit, set = a.deskNumeric()
	}
	n := value + delta
	if setValue != nil {
		n = *setValue
	}
	set(max(0, min(limit, n)))
	a.dirty = true
}
func (a *app) deskConfirm() (bool, error) {
	switch a.screen {
	case screenQuit:
		a.pointerAction = actions.Confirm
		return false, nil
	case screenReorganizeCancel:
		if a.reorganization != nil && !a.reorganization.CanCancel() {
			a.quitBack = screenReorganizeTarget
			a.deskSetScreen(screenQuit)
			return true, nil
		}
		if e := a.cancelReorganization(); e != nil {
			a.deskStatus(e.Error())
			a.deskSetScreen(screenReorganizeTarget)
		}
		return true, nil
	case screenSaveConfirm:
		if e := a.autosave(); e != nil {
			a.deskStatus(a.deskText("save.failed") + e.Error())
		} else {
			a.deskStatus(a.deskText("save.done"))
		}
		a.deskSetScreen(a.desk.Return)
		return true, nil
	case screenLoadConfirm:
		a.pointerAction = actions.Confirm
		return false, nil
	case screenRecruitConfirm, screenTransferConfirm, screenTrainConfirm:
		a.pointerAction = actions.Confirm
		return false, nil
	}
	pending := a.desk.Pending
	a.desk.Pending = ""
	switch pending {
	case "supply":
		a.report(a.withBudget(a.current, func() string { return a.execSupply(a.supplyTarget, a.supplyAmounts) }))
		a.deskSetScreen(screenMap)
	case "trade":
		a.report(a.withBudget(a.current, func() string { return a.execTrade(a.tradeImport, a.tradeGood, int(a.tradeAmount)) }))
		a.deskSetScreen(screenMap)
	case "reorganize":
		a.finishReorganization()
		a.deskSetScreen(screenMap)
	case "month":
		message := a.endTurn()
		if a.screen != screenBattle && a.screen != screenCampaignEvent {
			a.deskSetScreen(screenMap)
		}
		a.report(message)
	default:
		return true, nil
	}
	return true, nil
}
func (a *app) deskAction(key string) (bool, error) {
	a.dirty = true
	if key == "back" {
		return a.deskRequestBack(false)
	}
	if key == "home" {
		return a.deskRequestBack(true)
	}
	if key == "dialog.no" {
		return a.deskRequestBack(false)
	}
	if key == "dialog.yes" {
		return a.deskConfirm()
	}
	if strings.HasPrefix(key, "page.") {
		return true, a.deskOpenPage(strings.TrimPrefix(key, "page."))
	}
	if key == "primary" {
		switch a.screen {
		case screenMap:
			if a.activeBattle() {
				a.resumeBattle()
			} else {
				return true, a.deskOpenPage("commands")
			}
		case screenRecruitAmount:
			if a.recruitAmount > 0 {
				a.deskSetScreen(screenRecruitConfirm)
			}
		case screenTransferSelection, screenTransferAmount:
			if a.transferSession != nil && len(a.transferSession.Selected()) > 0 {
				a.deskSetScreen(screenTransferConfirm)
			}
		case screenSupplyAmount:
			if a.supplyAmounts != [4]int{} {
				a.desk.Pending = "supply"
			}
		case screenTradeAmount:
			if a.tradeAmount > 0 {
				a.desk.Pending = "trade"
			}
		case screenReorganizeTarget:
			if a.reorganization.CanFinish() {
				a.desk.Pending = "reorganize"
			}
		case screenReorganizeAmount:
			if e := a.reorganization.Assign(a.reorganizeID, int(a.reorganizeInput)); e != nil {
				a.deskStatus(e.Error())
			}
			a.syncReorganizationGenerals()
			a.deskSetScreen(screenReorganizeTarget)
		default:
			a.pointerAction = actions.Submit
			return false, nil
		}
		return true, nil
	}
	if key == "month" {
		a.desk.Pending = "month"
		return true, nil
	}
	if strings.HasPrefix(key, "command.") {
		n, _ := strconv.Atoi(strings.TrimPrefix(key, "command."))
		a.desk.Commands = false
		a.screen = screenCommand
		a.pointerAction = actions.Selection(n)
		return false, nil
	}
	if strings.HasPrefix(key, "option.") {
		n, _ := strconv.Atoi(strings.TrimPrefix(key, "option."))
		if a.screen == screenOtherOptions && n == 1 {
			return true, a.deskOpenPage("save")
		}
		if a.screen == screenOtherOptions && n == 2 {
			return true, a.deskOpenPage("load")
		}
		a.pointerAction = actions.Selection(n)
		return false, nil
	}
	if key == "province" || key == "branch" || key == "target" || key == "good" || key == "covert" || key == "ceasefire" || key == "autonomy" {
		a.desk.Picker = key
		a.desk.Scroll = 0
		return true, nil
	}
	if strings.HasPrefix(key, "pick.") {
		n, _ := strconv.Atoi(strings.TrimPrefix(key, "pick."))
		picker := a.desk.Picker
		a.desk.Picker = ""
		a.desk.Focus = ""
		a.desk.Scroll = 0
		switch picker {
		case "province":
			if n < 1 || n > int(a.provinceLimit) {
				return true, nil
			}
			if a.reorganization != nil {
				if a.reorganization.Changed() || a.reorganization.Remaining() > 0 {
					a.deskSetScreen(screenReorganizeCancel)
					return true, nil
				}
				a.reorganization = nil
			}
			a.current = game.ProvinceID(n)
			a.transferSession = nil
			a.recruitAmount = 0
			a.supplyTarget = 0
			if a.screen == screenReorganizeTarget {
				return true, a.deskBeginReorganization(game.BranchInfantry)
			}
			if a.screen == screenTransferSelection {
				a.screen = screenTransferTarget
				a.deskInitialize()
			}
			if a.screen == screenViewGenerals {
				a.viewGenerals = a.deskUnitIDs()
			}
			if a.screen == screenSupplyAmount {
				a.screen = screenSupplyTarget
				a.deskInitialize()
			}
		case "branch":
			if a.reorganization != nil {
				if a.reorganization.Changed() || a.reorganization.Remaining() > 0 {
					a.deskSetScreen(screenReorganizeCancel)
					return true, nil
				}
				return true, a.deskBeginReorganization(uint8(n))
			}
			a.recruitBranch = uint8(n)
			a.recruitAmount = 0
		case "target":
			if a.screen == screenSupplyAmount {
				a.supplyTarget = game.ProvinceID(n)
			} else {
				a.deskTransferTarget(game.ProvinceID(n))
			}
		case "good":
			a.tradeGood = game.TradeGood(n)
			a.tradeAmount = 0
		case "covert":
			a.covertInput = uint32(n)
		case "ceasefire":
			a.ceasefireInput = uint32(n)
		case "autonomy":
			a.autonomyInput = uint32(n)
		}
		return true, nil
	}
	if strings.HasPrefix(key, "province.") {
		n, _ := strconv.Atoi(strings.TrimPrefix(key, "province."))
		if n < 1 || n > int(a.provinceLimit) {
			return true, nil
		}
		a.current = game.ProvinceID(n)
		return true, nil
	}
	if strings.HasPrefix(key, "toggle.") {
		n, _ := strconv.Atoi(strings.TrimPrefix(key, "toggle."))
		if a.transferSession != nil {
			if e := a.transferSession.Toggle(game.GeneralID(n)); e != nil {
				a.deskStatus(e.Error())
			}
		}
		return true, nil
	}
	if key == "select.all" && a.transferSession != nil {
		for _, id := range a.transferSession.Candidates() {
			found := false
			for _, s := range a.transferSession.Selected() {
				found = found || s == id
			}
			if !found {
				_ = a.transferSession.Toggle(id)
			}
		}
		return true, nil
	}
	if strings.HasPrefix(key, "reorg.") {
		parts := strings.Split(key, ".")
		if len(parts) != 3 {
			return true, nil
		}
		id, _ := strconv.Atoi(parts[1])
		delta, _ := strconv.Atoi(parts[2])
		i := id - 1
		if i >= 0 && i < len(a.world.Strengths) {
			n := int(a.world.Strengths[i].Force) + delta
			if e := a.reorganization.Assign(game.GeneralID(id), n); e != nil {
				a.deskStatus(e.Error())
			}
			a.syncReorganizationGenerals()
		}
		return true, nil
	}
	if strings.HasPrefix(key, "unit.") {
		id, _ := strconv.Atoi(strings.TrimPrefix(key, "unit."))
		a.reorganizeID = game.GeneralID(id)
		a.reorganizeInput = uint32(a.world.Strengths[id-1].Force)
		a.deskSetScreen(screenReorganizeAmount)
		return true, nil
	}
	if strings.HasPrefix(key, "person.") {
		index, _ := strconv.Atoi(strings.TrimPrefix(key, "person."))
		a.viewIndex = index
		a.deskSetScreen(screenViewGeneral)
		return true, nil
	}
	if key == "biography" {
		a.openBiography(screenViewGeneral)
		a.desk.Scroll = 0
		return true, nil
	}
	if key == "import" || key == "export" {
		a.tradeImport = key == "import"
		if a.tradeImport && !game.Importable(a.tradeGood) {
			a.tradeGood = game.GoodFood
		}
		a.tradeAmount = 0
		return true, nil
	}
	if key == "quantity.minus" || key == "quantity.plus" {
		step := 100
		if a.screen == screenMessageTime || a.screen == screenProduction {
			step = 1
		}
		if key == "quantity.minus" {
			step = -step
		}
		a.deskAdjust("quantity", step, nil)
		return true, nil
	}
	if strings.HasPrefix(key, "preset.") {
		v := strings.TrimPrefix(key, "preset.")
		_, limit, _ := a.deskNumeric()
		n, _ := strconv.Atoi(v)
		if v == "max" {
			n = limit
		}
		a.deskAdjust("quantity", 0, &n)
		return true, nil
	}
	if strings.HasPrefix(key, "resource.minus.") || strings.HasPrefix(key, "resource.plus.") {
		parts := strings.Split(key, ".")
		delta := 100
		if parts[1] == "minus" {
			delta = -delta
		}
		a.deskAdjust("resource."+parts[2], delta, nil)
		return true, nil
	}
	if key == "map.in" || key == "map.out" {
		factor := 1.2
		if key == "map.out" {
			factor = 1 / 1.2
		}
		a.room.Camera.ZoomAt(factor, 500, 300, 1000, 600)
		return true, nil
	}
	if key == "map.reset" {
		a.room.Camera = atlasmap.DefaultCamera()
		return true, nil
	}
	return true, nil
}
func (a *app) updateModernDesk() (bool, error) {
	if !a.modernDeskVisible() {
		return false, nil
	}
	a.pointerAction = actions.None
	if !a.desk.Known || a.desk.Screen != a.screen {
		a.desk.Screen, a.desk.Known = a.screen, true
		a.desk.Focus = ""
		a.desk.Scroll = 0
	}
	a.deskInitialize()
	data := a.deskData()
	x, y, mouse := a.gameCursorPosition()
	if mouse && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		return a.deskRequestBack(false)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return a.deskRequestBack(false)
	}
	if data.Dialog != nil {
		if inpututil.IsKeyJustPressed(ebiten.KeyY) {
			return a.deskConfirm()
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyN) {
			return a.deskRequestBack(false)
		}
	}
	ctrl := ebiten.IsKeyPressed(ebiten.KeyControl) || ebiten.IsKeyPressed(ebiten.KeyMeta)
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyS) {
		return true, a.deskOpenPage("save")
	}
	if ctrl && inpututil.IsKeyJustPressed(ebiten.KeyL) {
		return true, a.deskOpenPage("load")
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		ts := render.DeskTargets(data)
		index := -1
		for i, v := range ts {
			if v.Key == a.desk.Focus {
				index = i
			}
		}
		step := 1
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			step = -1
		}
		for n := 0; n < len(ts); n++ {
			index = (index + step + len(ts)) % len(ts)
			if !ts[index].Disabled {
				a.desk.Focus = ts[index].Key
				break
			}
		}
		a.dirty = true
		return true, nil
	}
	if mouse {
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			a.desk.Press = render.DeskHit(data, x, y)
			a.desk.PressScreen = a.screen
			a.desk.PressX, a.desk.PressY = x, y
			a.desk.Focus = a.desk.Press
			for _, v := range render.DeskTargets(data) {
				if v.Key == a.desk.Press && v.Kind == "slider" {
					a.desk.Slider = v.Key
				}
			}
			if a.desk.Press == "" && data.Map != nil && l.DeskMap.Contains(x, y) {
				a.desk.Drag = true
				a.desk.DragX, a.desk.DragY = x, y
			}
		}
		if a.desk.Slider != "" && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			for _, v := range render.DeskTargets(data) {
				if v.Key == a.desk.Slider {
					n := max(0, min(v.Maximum, (x-v.Rect.X)*v.Maximum/max(1, v.Rect.W)))
					a.deskAdjust(v.Key, 0, &n)
				}
			}
		}
		if a.desk.Drag && ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			dx, dy := x-a.desk.DragX, y-a.desk.DragY
			if dx != 0 || dy != 0 {
				a.room.Camera.Pan(float64(dx)*1000/float64(l.DeskMap.W), float64(dy)*600/float64(l.DeskMap.H), 1000, 600)
				a.desk.DragX, a.desk.DragY = x, y
				a.dirty = true
			}
		}
		if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
			key := a.desk.Press
			dx, dy := x-a.desk.PressX, y-a.desk.PressY
			slider := a.desk.Slider
			a.desk.Press, a.desk.Slider = "", ""
			a.desk.Drag = false
			if slider != "" {
				return true, nil
			}
			if key != "" && a.desk.PressScreen == a.screen && dx*dx+dy*dy <= 64 && render.DeskHit(data, x, y) == key {
				return a.deskAction(key)
			}
		}
		_, wheel := ebiten.Wheel()
		if wheel != 0 {
			if data.Map != nil && l.DeskMap.Contains(x, y) {
				factor := 1.1
				if wheel < 0 {
					factor = 1 / 1.1
				}
				a.room.Camera.ZoomAt(factor, float64(x-l.DeskMap.X)*1000/float64(l.DeskMap.W), float64(y-l.DeskMap.Y)*600/float64(l.DeskMap.H), 1000, 600)
			} else if l.DeskContent.Contains(x, y) {
				a.desk.Scroll = max(0, min(max(0, data.ContentHeight-l.DeskContent.H), a.desk.Scroll-int(wheel*48)))
			}
			a.dirty = true
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
		a.desk.Scroll = min(max(0, data.ContentHeight-l.DeskContent.H), a.desk.Scroll+360)
		a.dirty = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
		a.desk.Scroll = max(0, a.desk.Scroll-360)
		a.dirty = true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if a.desk.Focus != "" {
			return a.deskAction(a.desk.Focus)
		}
		if data.Dialog != nil {
			return a.deskConfirm()
		}
		return a.deskAction("primary")
	}
	if data.Dialog == nil && a.desk.Picker == "" {
		value, limit, set := a.deskNumeric()
		if limit > 0 {
			changed := false
			for _, ch := range ebiten.AppendInputChars(nil) {
				if ch >= '0' && ch <= '9' {
					n := value*10 + int(ch-'0')
					if n <= limit {
						value = n
						changed = true
					}
				}
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
				value /= 10
				changed = true
			}
			if changed {
				set(value)
				a.dirty = true
			}
		}
		if limit == 0 && a.screen != screenMap && a.screen != screenCommand {
			for _, k := range inpututil.AppendJustPressedKeys(nil) {
				if k < ebiten.KeyDigit0 || k > ebiten.KeyDigit9 {
					continue
				}
				n := int(k - ebiten.KeyDigit0)
				if n == 0 {
					n = 10
				}
				target := fmt.Sprintf("option.%d", n)
				for _, v := range data.Controls {
					if v.Key == target && !v.Disabled {
						return a.deskAction(target)
					}
				}
			}
		}
		if a.screen == screenMap || a.screen == screenCommand {
			for _, k := range inpututil.AppendJustPressedKeys(nil) {
				if k == ebiten.KeyLeft || k == ebiten.KeyRight {
					delta := 1
					if k == ebiten.KeyLeft {
						delta = -1
					}
					limit := max(1, int(a.provinceLimit))
					n := (int(a.current)-1+delta+limit)%limit + 1
					a.current = game.ProvinceID(n)
					a.dirty = true
					return true, nil
				}
				if (k >= ebiten.KeyDigit0 && k <= ebiten.KeyDigit9) || k == ebiten.KeyR || k == ebiten.KeyO || k == ebiten.KeyT || k == ebiten.KeyS || k == ebiten.KeyV || k == ebiten.KeyA || k == ebiten.KeyE {
					a.screen = screenCommand
					return false, nil
				}
			}
		}
	}
	return true, nil
}

func (a *app) deskData() render.DeskData {
	d := render.DeskData{LightUI: a.lightInterface(), Fonts: a.eten, Labels: map[string]string{}, Scroll: a.desk.Scroll, Focus: a.desk.Focus, ContentHeight: l.DeskContent.H, Page: "map", Status: a.desk.Message}
	for _, key := range []string{"title", "mode", "headquarters", "progress", "mouse.help", "breadcrumb", "footer", "selected", "force", "arsenal", "coal", "iron", "stationed", "resource.0", "resource.1", "resource.2", "resource.3", "resource.4"} {
		d.Labels[key] = a.deskText(key)
	}
	word := a.deskText
	control := func(key, text, kind string, r l.Rect, body, disabled, selected bool) {
		d.Controls = append(d.Controls, render.DeskControl{Key: key, Text: text, Kind: kind, Rect: r, Body: body, Disabled: disabled, Selected: selected})
	}
	text := func(s string, r l.Rect, size int, muted, gold bool) {
		d.Texts = append(d.Texts, render.DeskText{Text: s, Rect: r, Size: size, Muted: muted, Gold: gold})
	}
	button := func(key, label string, r l.Rect, body, disabled bool) {
		control(key, label, "button", r, body, disabled, false)
	}
	card := func(r l.Rect) { d.Cards = append(d.Cards, r) }
	primary := func(label string, disabled bool) {
		control("primary", label, "primary", l.Rect{200, 655, 210, 42}, false, disabled, false)
		button("back", word("cancel"), l.Rect{422, 655, 82, 42}, false, false)
	}
	p, _ := a.tbl.At(a.current)
	d.Province = a.provinceName(a.current)
	d.Commander = a.generalDisplayName(p.Commander)
	d.Date = fmt.Sprintf(word("date"), int(a.year)+1911, a.month)
	d.Resources = [6]string{deskNumber(int(p.Gold)), deskNumber(int(p.Food)), deskNumber(int(p.Ammo)), deskNumber(int(p.Fuel)), strconv.Itoa(int(p.Arsenals)), deskNumber(int(p.Iron))}
	force := 0
	for _, u := range a.world.Units {
		if !u.Active || u.Province != a.current || u.Faction != p.Commander {
			continue
		}
		i := int(u.General) - 1
		if i < 0 || i >= len(a.world.Strengths) {
			continue
		}
		s := a.world.Strengths[i]
		force += int(s.Force)
		if len(d.Units) < 4 {
			d.Units = append(d.Units, render.DeskUnit{Name: a.generalDisplayName(u.General), Detail: fmt.Sprintf(word("unit.detail"), a.deskBranchName(s.Branch), s.Ability, s.F20), Force: deskNumber(int(s.Force))})
		}
	}
	d.Force, d.Arsenals, d.Coal, d.Iron = deskNumber(force), strconv.Itoa(int(p.Arsenals)), deskNumber(int(p.Coal)), deskNumber(int(p.Iron))
	d.Title = word("map")
	d.Prompt = word("map.hint")
	button("back", word("back"), l.DeskBack, false, false)
	button("home", word("map"), l.DeskHome, false, false)
	for i, key := range []string{"map", "recruit", "transfer", "reorganize", "people"} {
		disabled := i > 0 && i < 4 && (!a.ownsProvince(a.current) || a.strategyLocked())
		control("page."+key, word(key), "button", l.DeskSidebarButton(i), false, disabled, false)
	}
	button("page.save", word("save"), l.Rect{12, 449, 155, 45}, false, !a.canSaveSession())
	button("page.load", word("load"), l.Rect{12, 503, 155, 45}, false, a.savePath == "")
	provinceField := func() {
		text(word("province"), l.Rect{218, 274, 80, 42}, 15, true, false)
		button("province", d.Province+" · "+d.Commander, l.Rect{306, 274, 676, 42}, true, false)
	}
	quantity := func() {
		value, limit, _ := a.deskNumeric()
		text(word("quantity"), l.Rect{218, 405, 350, 26}, 15, false, true)
		text(fmt.Sprintf(word("limit"), deskNumber(limit)), l.Rect{748, 405, 230, 26}, 12, true, false)
		text(deskNumber(value), l.Rect{218, 438, 600, 56}, 38, false, true)
		button("quantity.minus", "-", l.Rect{218, 496, 42, 42}, true, value == 0)
		button("quantity.plus", "+", l.Rect{940, 496, 42, 42}, true, value >= limit)
		d.Controls = append(d.Controls, render.DeskControl{Key: "quantity", Kind: "slider", Rect: l.Rect{272, 496, 656, 42}, Body: true, Value: value, Maximum: limit})
		for i, v := range []string{"500", "1000", "5000", "max"} {
			label := v
			if v == "max" {
				label = word("maximum")
			}
			button("preset."+v, label, l.Rect{218 + i*193, 553, 184, 42}, true, limit == 0)
		}
	}
	switch a.screen {
	case screenMap, screenCommand:
		if a.desk.Commands {
			d.Page = "commands"
			d.Title = word("commands")
			d.Prompt = word("commands.hint")
			for i := 1; i <= 15; i++ {
				label, _ := a.wordingText(fmt.Sprintf("command.%02d", i))
				r := l.Rect{200 + ((i-1)%2)*405, 241 + ((i-1)/2)*57, 395, 48}
				button(fmt.Sprintf("command.%d", i), label, r, true, a.strategyLocked() && i != 6 && i != 15)
			}
			d.ContentHeight = 590
			break
		}
		if a.room.Map == nil {
			a.room.Map, _ = atlasmap.New()
			a.room.Camera = atlasmap.DefaultCamera()
		}
		var factions [39]int
		for i, v := range a.tbl.Province {
			factions[i] = int(v.Commander)
		}
		d.Map = a.room.Map.Render(a.room.Camera, 1000, 600, a.themeMode == t.ModeRetro, true, factions)
		var markers []l.Rect
		for _, selected := range []bool{true, false} {
			for i, anchor := range atlasmap.Anchors {
				if i >= int(a.provinceLimit) {
					continue
				}
				if (game.ProvinceID(i+1) == a.current) != selected {
					continue
				}
				v := a.room.Map.Project(anchor, a.room.Camera, 1000, 600)
				name := strings.TrimSuffix(a.provinceName(game.ProvinceID(i+1)), "省")
				w := render.DeskLabelWidth(name, 11) + 24
				r := l.Rect{l.DeskMap.X + int(v[0])*l.DeskMap.W/1000 - w/2, l.DeskMap.Y + int(v[1])*l.DeskMap.H/600 - 15, w, 30}
				if r.X < l.DeskMap.X+4 || r.Right() > l.DeskMap.Right()-4 || r.Y < l.DeskMap.Y+4 || r.Bottom() > l.DeskMap.Bottom()-52 {
					continue
				}
				overlap := false
				for _, m := range markers {
					if r.Intersects(m) {
						overlap = true
					}
				}
				if overlap {
					continue
				}
				markers = append(markers, r)
				control(fmt.Sprintf("province.%d", i+1), name, "marker", r, true, false, selected)
			}
		}
		button("province", d.Province, l.Rect{900, 181, 100, 28}, true, false)
		button("map.out", "-", l.Rect{806, 570, 38, 38}, true, false)
		button("map.reset", word("map.reset"), l.Rect{852, 570, 89, 38}, true, false)
		button("map.in", "+", l.Rect{949, 570, 38, 38}, true, false)
		if a.activeBattle() {
			control("page.resume", word("battle.resume"), "primary", l.Rect{200, 655, 190, 42}, false, false, false)
			d.Status = word("battle.locked")
		} else {
			control("page.recruit", word("recruit"), "primary", l.Rect{200, 655, 82, 42}, false, !a.ownsProvince(a.current) || a.strategyLocked(), false)
			button("page.transfer", word("transfer"), l.Rect{292, 655, 112, 42}, false, !a.ownsProvince(a.current) || a.strategyLocked())
			button("page.people", word("people"), l.Rect{414, 655, 112, 42}, false, false)
		}
		button("page.commands", word("commands"), l.Rect{536, 655, 112, 42}, false, false)
		button("page.armory", word("armory"), l.Rect{658, 655, 112, 42}, false, !a.militaryEnabled() || a.strategyLocked())
		monthLabel := word("month")
		if a.strategyTurn != nil && !a.activeBattle() {
			monthLabel, d.Status = a.militaryText("turn.continue"), a.strategyLockText()
		}
		button("month", monthLabel, l.Rect{780, 655, 170, 42}, false, a.activeBattle())
	case screenRecruitAmount, screenRecruitConfirm:
		d.Page = "recruit"
		d.Title = word("recruit")
		d.Prompt = word("form.hint")
		card(l.Rect{200, 241, 800, 362})
		provinceField()
		text(word("branch"), l.Rect{218, 331, 80, 42}, 15, true, false)
		button("branch", a.deskBranchName(a.recruitBranch), l.Rect{306, 331, 676, 42}, true, false)
		quantity()
		cost := game.RecruitCost(a.recruitBranch, int(a.recruitAmount))
		text(word("cost"), l.Rect{610, 438, 170, 20}, 12, true, false)
		text(deskNumber(cost), l.Rect{610, 460, 170, 28}, 20, false, false)
		text(word("after.gold"), l.Rect{800, 438, 182, 20}, 12, true, false)
		text(deskNumber(int(p.Gold)-cost), l.Rect{800, 460, 182, 28}, 20, false, false)
		d.ContentHeight = 502
		primary(word("recruit.review"), a.recruitAmount == 0 || int(a.recruitAmount) > a.world.RecruitLimit(a.current, a.recruitBranch) || a.activeBattle())
		if a.screen == screenRecruitConfirm {
			d.Dialog = &render.DeskDialog{Title: word("recruit.confirm"), Text: fmt.Sprintf(word("recruit.terms"), d.Province, a.deskBranchName(a.recruitBranch), deskNumber(int(a.recruitAmount)), deskNumber(cost)), Yes: word("confirm"), No: word("adjust")}
		}
	case screenTransferSelection, screenTransferAmount, screenTransferConfirm, screenSupplyAmount:
		supply := a.screen == screenSupplyAmount
		d.Page = "transfer"
		d.Title = word("transfer")
		if supply {
			d.Title = word("supply")
		}
		d.Prompt = word("form.hint")
		provinceField()
		target := ""
		if supply && a.supplyTarget.Valid() {
			target = a.provinceName(a.supplyTarget)
		} else if a.transferSession != nil {
			target = a.provinceName(a.transferSession.To())
		}
		text(word("destination"), l.Rect{218, 331, 80, 42}, 15, true, false)
		button("target", target, l.Rect{306, 331, 676, 42}, true, false)
		y := 405
		if !supply {
			button("select.all", word("select.all"), l.Rect{800, 379, 182, 30}, true, a.transferSession == nil)
			if a.transferSession != nil {
				selected := map[game.GeneralID]bool{}
				for _, id := range a.transferSession.Selected() {
					selected[id] = true
				}
				for _, id := range a.transferSession.Candidates() {
					i := int(id) - 1
					g := a.generals[i]
					control(fmt.Sprintf("toggle.%d", id), fmt.Sprintf(word("general.row"), a.generalDisplayName(id), a.deskBranchName(g.Branch), deskNumber(int(a.world.Strengths[i].Force)), g.Stamina), "checkbox", l.Rect{218, y, 764, 54}, true, g.Stamina < 30, selected[id])
					y += 60
				}
			}
		}
		text(word("resources"), l.Rect{218, y + 8, 764, 28}, 15, false, true)
		y += 42
		for i := 0; i < 4; i++ {
			value, limit, _ := a.deskResource(i)
			text(word(fmt.Sprintf("resource.%d", i))+"  "+deskNumber(value), l.Rect{218, y, 210, 42}, 15, false, false)
			button(fmt.Sprintf("resource.minus.%d", i), "-", l.Rect{450, y, 42, 42}, true, value == 0)
			d.Controls = append(d.Controls, render.DeskControl{Key: fmt.Sprintf("resource.%d", i), Kind: "slider", Rect: l.Rect{504, y, 416, 42}, Body: true, Value: value, Maximum: limit})
			button(fmt.Sprintf("resource.plus.%d", i), "+", l.Rect{932, y, 42, 42}, true, value >= limit)
			y += 54
		}
		d.ContentHeight = y - 140 + 25
		disabled := a.transferSession == nil || len(a.transferSession.Selected()) == 0
		if supply {
			disabled = a.supplyAmounts == [4]int{} || !a.supplyTarget.Valid()
		}
		primary(word("review"), disabled || a.activeBattle())
		if a.screen == screenTransferConfirm {
			d.Dialog = &render.DeskDialog{Title: word("transfer.confirm"), Text: fmt.Sprintf(word("transfer.terms"), d.Province, target, len(a.transferSession.Selected()), a.transferAmounts[0], a.transferAmounts[1], a.transferAmounts[2], a.transferAmounts[3]), Yes: word("confirm"), No: word("adjust")}
		}
	case screenReorganizeTarget, screenReorganizeAmount, screenReorganizeCancel:
		d.Page = "reorganize"
		d.Title = word("reorganize")
		d.Prompt = word("reorganize.hint")
		provinceField()
		r := a.reorganization
		if r != nil {
			text(fmt.Sprintf(word("pool"), deskNumber(r.Remaining())), l.Rect{218, 338, 764, 42}, 24, false, true)
			button("branch", a.deskBranchName(r.Snapshot().Branch), l.Rect{800, 386, 182, 36}, true, r.Changed() || r.Remaining() > 0)
			if a.screen == screenReorganizeAmount {
				text(a.generalDisplayName(a.reorganizeID), l.Rect{218, 380, 764, 28}, 20, false, false)
				quantity()
				d.ContentHeight = 480
				primary(word("assign"), false)
			} else {
				y := 432
				for _, id := range r.Targets() {
					i := int(id) - 1
					force := int(a.world.Strengths[i].Force)
					button(fmt.Sprintf("unit.%d", id), a.generalDisplayName(id)+"  "+deskNumber(force), l.Rect{218, y, 490, 42}, true, false)
					button(fmt.Sprintf("reorg.%d.-100", id), "-100", l.Rect{722, y, 120, 42}, true, force < 100)
					button(fmt.Sprintf("reorg.%d.100", id), "+100", l.Rect{854, y, 128, 42}, true, r.Remaining() < 100 || force+100 > r.Limit(id))
					y += 60
				}
				d.ContentHeight = y - 140 + 25
				primary(word("reorganize.finish"), !r.CanFinish() || !r.Changed())
			}
			if a.screen == screenReorganizeCancel {
				title, terms, yes := word("reorganize.cancel"), word("reorganize.cancel.terms"), word("reorganize.cancel.yes")
				if !r.CanCancel() {
					title, terms, yes = word("reorganize.legacy"), word("reorganize.legacy.terms"), word("save.exit")
				}
				d.Dialog = &render.DeskDialog{Title: title, Text: terms, Yes: yes, No: word("keep.editing")}
			}
		}
	case screenTradeAmount:
		d.Page = "trade"
		d.Title = word("trade")
		d.Prompt = word("form.hint")
		provinceField()
		control("import", word("import"), "button", l.Rect{218, 331, 190, 42}, true, false, a.tradeImport)
		control("export", word("export"), "button", l.Rect{420, 331, 190, 42}, true, false, !a.tradeImport)
		button("good", a.deskGoodName(a.tradeGood), l.Rect{622, 331, 360, 42}, true, false)
		quantity()
		text(word("cost")+"  "+deskNumber(game.TradePrice(a.tradeGood, int(a.tradeAmount))), l.Rect{218, 620, 764, 42}, 20, false, true)
		d.ContentHeight = 550
		primary(word("review"), a.tradeAmount == 0 || int(a.tradeAmount) > a.deskTradeLimit() || a.activeBattle())
	case screenCovertTarget, screenCeasefireTarget:
		flow, _ := a.modernFlowData()
		d.Title, d.Prompt = flow.Title, flow.Prompt
		value, _, _ := a.deskNumeric()
		key := "covert"
		if a.screen == screenCeasefireTarget {
			key = "ceasefire"
		}
		label := word("choose")
		if value > 0 && value <= 39 {
			label = a.provinceName(game.ProvinceID(value))
		}
		button(key, label, l.Rect{218, 274, 764, 50}, true, false)
		primary(word("confirm"), value < 1 || value > int(a.provinceLimit))
	case screenViewGenerals:
		d.Page = "people"
		d.Title = word("people")
		d.Prompt = word("people.hint")
		provinceField()
		y := 340
		for i, id := range a.viewGenerals {
			g := a.generals[int(id)-1]
			button(fmt.Sprintf("person.%d", i), fmt.Sprintf(word("general.row"), a.generalDisplayName(id), a.deskBranchName(g.Branch), deskNumber(int(g.Force)), g.Stamina), l.Rect{218, y, 764, 54}, true, false)
			y += 60
		}
		d.ContentHeight = y - 140 + 25
	case screenViewGeneral, screenBiography:
		d.Page = "people"
		d.Title = word("people")
		d.Prompt = word("people.hint")
		if a.viewIndex >= 0 && a.viewIndex < len(a.viewGenerals) {
			id := a.viewGenerals[a.viewIndex]
			g := a.generals[int(id)-1]
			portrait, unavailable := a.portraitForGeneral(id)
			d.Portrait = &render.PortraitView{Portrait: portrait, Unavailable: unavailable}
			text(a.generalDisplayName(id), l.Rect{218, 241, 562, 42}, 26, false, false)
			body := fmt.Sprintf(word("person.stats"), a.deskBranchName(g.Branch), g.Force, g.AbilityA, g.F19, g.F20, g.Stamina, g.F30)
			if a.screen == screenBiography {
				if p, ok := a.currentBiography(); ok {
					body = p.Biography
					if body == "" {
						body = a.biographyUnavailable()
					}
					body += "\n\n" + strings.Join(p.Sources, "\n")
					if portrait != nil {
						body += "\n\n" + portrait.Attribution + "\n" + portrait.SourcePage + "\n" + portrait.LicenseURL
					}
				}
			}
			h := render.DeskParagraphHeight(body, 562)
			d.Texts = append(d.Texts, render.DeskText{Text: body, Rect: l.Rect{218, 304, 562, h}, Paragraph: true})
			d.ContentHeight = max(502, h+190)
			if a.screen == screenViewGeneral {
				button("biography", word("biography"), l.Rect{200, 655, 210, 42}, false, false)
			}
		}
	case screenSaveConfirm, screenLoadConfirm, screenQuit, screenTrainConfirm:
		title, terms := word("save.confirm"), word("save.terms")
		if a.screen == screenLoadConfirm {
			title, terms = word("load.confirm"), word("load.terms")
		}
		if a.screen == screenQuit {
			title, terms = word("quit.confirm"), word("quit.terms")
		}
		if a.screen == screenTrainConfirm {
			title, terms = word("train.confirm"), word("train.terms")
		}
		d.Title = title
		d.Dialog = &render.DeskDialog{Title: title, Text: terms, Yes: word("confirm"), No: word("cancel")}
	default:
		flow, e := a.modernFlowData()
		if e != nil {
			d.Status = e.Error()
			break
		}
		d.Title, d.Prompt = flow.Title, flow.Prompt
		y := 241
		for i, s := range flow.Options {
			r := l.Rect{200 + (i%2)*405, y + (i/2)*57, 395, 48}
			if i < len(flow.Illustrations) {
				r.Y = y + (i/2)*228
				r.H = 216
				if im := hdart.DevelopmentScene(flow.Illustrations[i]); im != nil {
					d.Illustrations = append(d.Illustrations, render.DeskIllustration{Image: im, Rect: r})
				}
			}
			disabled := false
			if a.screen == screenOtherOptions {
				disabled = i == 2 || i == 3 || i == 4 || i == 6 || (i == 0 && !a.canSaveSession()) || (i == 1 && a.savePath == "")
			}
			button(fmt.Sprintf("option.%d", i+1), s, r, true, disabled)
		}
		d.ContentHeight = max(502, 160+(len(flow.Options)+1)/2*57)
		if len(flow.Illustrations) > 0 {
			d.ContentHeight = max(502, 160+(len(flow.Options)+1)/2*228)
		}
		if flow.InputMode {
			card(l.Rect{200, 241, 800, 400})
			text(flow.Prompt, l.Rect{218, 274, 764, 40}, 16, false, false)
			quantity()
			primary(word("confirm"), false)
		}
	}
	if a.desk.Picker != "" {
		d.Map = nil
		d.Controls = d.Controls[:9]
		d.Texts = nil
		d.Cards = nil
		d.Illustrations = nil
		d.Title = word("choose")
		d.Prompt = word("choose.hint")
		y := 241
		add := func(id int, name string) {
			button(fmt.Sprintf("pick.%d", id), name, l.Rect{218, y, 764, 42}, true, false)
			y += 49
		}
		switch a.desk.Picker {
		case "covert", "ceasefire":
			for id := game.ProvinceID(1); int(id) <= int(a.provinceLimit); id++ {
				add(int(id), a.provinceName(id))
			}
		case "province":
			for id := game.ProvinceID(1); int(id) <= int(a.provinceLimit); id++ {
				if (a.screen == screenMap || a.screen == screenCommand) || a.ownsProvince(id) {
					add(int(id), a.provinceName(id))
				}
			}
		case "target":
			ids := a.transferTargets
			if a.screen == screenSupplyAmount {
				ids = a.supplyTargets
			}
			for _, id := range ids {
				add(int(id), a.provinceName(id))
			}
		case "branch":
			for _, b := range game.RecruitBranchOrder {
				add(int(b), a.deskBranchName(b))
			}
		case "good":
			for _, g := range []game.TradeGood{game.GoodFood, game.GoodAmmo, game.GoodFuel, game.GoodCoal, game.GoodIron} {
				if !a.tradeImport || game.Importable(g) {
					add(int(g), a.deskGoodName(g))
				}
			}
		}
		d.ContentHeight = max(502, y-140+25)
	}
	if a.desk.Pending != "" {
		terms := word("review.terms")
		switch a.desk.Pending {
		case "trade":
			terms = fmt.Sprintf(word("trade.terms"), a.deskGoodName(a.tradeGood), a.tradeAmount, game.TradePrice(a.tradeGood, int(a.tradeAmount)))
		case "supply":
			terms = fmt.Sprintf(word("supply.terms"), a.provinceName(a.supplyTarget), a.supplyAmounts[0], a.supplyAmounts[1], a.supplyAmounts[2], a.supplyAmounts[3])
		case "reorganize":
			terms = word("reorganize.finish.terms")
		case "month":
			terms = word("month.terms")
		}
		d.Dialog = &render.DeskDialog{Title: word("review"), Text: terms, Yes: word("confirm"), No: word("adjust")}
	}
	if a.messages != nil && a.messages.Active() {
		d.Status = a.messages.Current()
	}
	return d
}
func (a *app) composeModernDesk() error {
	c := render.DrawDesk(a.deskData())
	if a.frame != nil {
		a.frame.Dispose()
	}
	a.frame = ebiten.NewImageFromImage(c.Image())
	a.dirty = false
	return nil
}
