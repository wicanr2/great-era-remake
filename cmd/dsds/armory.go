package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
)

type armoryView struct {
	Page, Category, UnitPage, Focus, Press int
	Selected                               game.EquipmentID
	Message                                string
	PendingUnit                            game.GeneralID
}
type armoryViewSnapshot struct {
	Open bool
	View armoryView
}

func defaultArmoryView() armoryView {
	return armoryView{Selected: "rifle-modern", Focus: -1, Press: -1}
}
func (a *app) armorySurface() bool {
	return a != nil && (a.screen == screenArmory || a.screen == screenQuit && a.quitBack == screenArmory)
}

// The expansion uses the authoritative world after AI battles for both its
// equipment ledger and the generals consumed by the player battle screen.
func (a *app) syncExpansionGenerals() {
	if a.armory == nil || a.world == nil {
		return
	}
	for i := range a.generals {
		if i >= len(a.world.Units) || i >= len(a.world.Strengths) {
			break
		}
		g := &a.generals[i]
		s := a.world.Strengths[i]
		g.Force = s.Force
		g.Branch = s.Branch
		g.Province = a.world.Units[i].Province
		g.AbilityA = s.Ability
		g.F19, g.F20, g.Stamina, g.F30 = s.F19, s.F20, s.F29, s.F30
		g.Experience = a.world.Units[i].Experience
	}
}
func (a *app) armoryViewSnapshot() armoryViewSnapshot {
	v := a.armoryView
	v.Focus, v.Press, v.Message = -1, -1, ""
	v.PendingUnit = 0
	return armoryViewSnapshot{Open: a.armorySurface(), View: v}
}
func validArmoryView(s armoryViewSnapshot, army *game.Armory, player game.GeneralID) bool {
	v := s.View
	m, ok := game.EquipmentByID(v.Selected)
	return army != nil && army.Nations[player] != nil && v.Page >= 0 && v.Page < 4 && v.Category >= 0 && v.Category < 3 && v.UnitPage >= 0 && v.UnitPage <= len(army.Units)/3 && v.Focus == -1 && v.Press == -1 && v.Message == "" && v.PendingUnit == 0 && ok && int(m.Class) == v.Category
}
func (a *app) armoryText(key string) string {
	s, err := a.wordingText("armory." + key)
	if err != nil {
		return "[armory." + key + "]"
	}
	return s
}
func (a *app) openArmory() {
	if a.activeBattle() {
		a.report(a.deskText("battle.locked"))
		return
	}
	if a.armory == nil || a.campaign == nil || a.campaign.Mode != "unification-resistance" {
		a.report(a.armoryText("original.only"))
		return
	}
	if err := a.armory.Sync(a.world); err != nil {
		a.report(a.armoryText("error"))
		return
	}
	if _, ok := game.EquipmentByID(a.armoryView.Selected); !ok {
		a.armoryView = defaultArmoryView()
	}
	a.armoryView.Press = -1
	a.screen, a.dirty = screenArmory, true
	a.applyWindowSize()
}
func (a *app) armoryData() render.ArmoryData {
	v := &a.armoryView
	d := render.ArmoryData{Fonts: a.eten, Labels: map[string]string{}, Page: v.Page, Category: v.Category, Focus: v.Focus, UnitPage: v.UnitPage, Message: v.Message, Quit: a.screen == screenQuit}
	for _, key := range []string{"kicker", "title", "intro", "mode", "factories", "categories", "sidebar.note", "artillery.note", "catalog.note", "capacity", "production.locked", "production.note", "stock.note", "selected.stock", "unit", "issue", "previous", "next", "pages", "units.empty", "equipment.note", "months.progress", "output", "research.start", "research.pause", "production.open", "stock.open", "month", "footer", "save", "load", "back", "quit"} {
		d.Labels[key] = a.armoryText(key)
	}
	for _, group := range []struct {
		key   string
		count int
	}{{"resource", 4}, {"page", 4}, {"step", 4}, {"category", 3}, {"category.hint", 3}, {"fact", 5}} {
		for i := 0; i < group.count; i++ {
			key := fmt.Sprintf("%s.%d", group.key, i)
			d.Labels[key] = a.armoryText(key)
		}
	}
	for _, key := range []string{"convert", "convert.confirm", "convert.cancel", "reserve", "reserve.issue"} {
		d.Labels[key] = a.armoryText(key)
	}
	d.Reserve = a.armory.ReservePersonnel[a.playerCommander]
	if v.PendingUnit != 0 {
		d.Pending = true
		if plan, err := a.armory.ConversionPlan(a.world, a.playerCommander, v.PendingUnit, v.Selected); err == nil {
			d.Confirmation = fmt.Sprintf(a.armoryText("convert.review"), a.generalDisplayName(v.PendingUnit), plan.Before, plan.Force, plan.Groups, plan.Gold, plan.Reserve)
		} else {
			d.Confirmation = a.armoryText("error")
		}
	}
	d.Resources = a.armory.Resources(a.world, a.playerCommander)
	d.Factories = a.armory.Factories(a.world, a.playerCommander)
	n := a.armory.Nations[a.playerCommander]
	d.Assigned = n.Assigned()
	d.Date = fmt.Sprintf(a.armoryText("date"), int(a.year)+1911, a.month)
	for _, m := range game.EquipmentCatalog() {
		if int(m.Class) != v.Category {
			continue
		}
		known := m.Legacy || n.Known[m.ID]
		card := render.ArmoryModel{Model: m, Name: a.armoryText("model." + string(m.ID)), Role: a.armoryText("role." + string(m.ID)), Description: a.armoryText("description." + string(m.ID)), Known: known, Stock: n.Stock[m.ID], Progress: int(n.Progress[m.ID]), Factories: int(n.Lines[m.ID]), Researching: n.Active == m.ID}
		card.Enabled = !m.Legacy && !known && (m.Prerequisite == "" || n.Known[m.Prerequisite]) && (n.Active == "" || n.Active == m.ID)
		card.Status = a.armoryText("status.available")
		if known {
			card.Status = a.armoryText("status.known")
		} else if card.Researching {
			card.Status = a.armoryText("status.researching")
		} else if !card.Enabled {
			card.Status = a.armoryText("status.locked")
		}
		card.Terms = fmt.Sprintf(a.armoryText("research.terms"), m.Months)
		if known {
			card.Terms = a.armoryText("status.known")
		}
		if v.Page == 1 {
			card.Terms = fmt.Sprintf(a.armoryText("production.terms"), m.Output, m.Cost.Gold, m.Cost.Iron, m.Cost.Coal, m.Cost.Fuel)
		}
		if last, ok := n.Last[m.ID]; ok {
			card.Last = fmt.Sprintf(a.armoryText("last"), last.Count)
			if last.Shortage {
				card.Last += " · " + a.armoryText("shortage")
			}
		} else {
			card.Last = fmt.Sprintf(a.armoryText("stock.amount"), n.Stock[m.ID])
		}
		d.Models = append(d.Models, card)
		if m.ID == v.Selected {
			d.Selected = card
		}
	}
	var units []render.ArmoryUnit
	for id := game.GeneralID(1); int(id) <= len(a.world.Units); id++ {
		u := a.armory.Units[id]
		if u == nil || u.Faction != a.playerCommander || u.Force == 0 {
			continue
		}
		class, ok := game.EquipmentClassOf(u.Branch)
		convert := v.Category == int(game.EquipmentArmor) && u.Branch == game.BranchInfantry
		if !ok || int(class) != v.Category && !convert {
			continue
		}
		worldUnit := a.world.Units[int(id)-1]
		if !worldUnit.Province.Valid() || int(worldUnit.Province) > a.provinceLimit {
			continue
		}
		p, _ := a.tbl.At(worldUnit.Province)
		if p.Commander != a.playerCommander {
			continue
		}
		need := game.EquipmentNeed(u.Branch, u.Force)
		selected := u.Items[v.Selected]
		parts := []string{}
		for _, m := range game.EquipmentCatalog() {
			if count := u.Items[m.ID]; count > 0 {
				parts = append(parts, fmt.Sprintf(a.armoryText("equipment.item"), a.armoryText("model."+string(m.ID)), count))
			}
		}
		name := fmt.Sprintf(a.armoryText("unit.name"), a.generalDisplayName(id), a.provinceName(worldUnit.Province))
		row := render.ArmoryUnit{ID: id, Name: name, Need: need, Selected: selected, Summary: fmt.Sprintf(a.armoryText("unit.summary"), u.Force, need, selected), Equipment: strings.Join(parts, " / "), Enabled: d.Selected.Known && n.Stock[v.Selected] > 0 && selected < need, Convert: convert, Reinforce: u.Branch == game.BranchInfantry && u.Force < 20000 && d.Reserve > 0}
		if convert {
			_, err := a.armory.ConversionPlan(a.world, a.playerCommander, id, v.Selected)
			row.Enabled = err == nil
			row.Summary = fmt.Sprintf(a.armoryText("convert.summary"), u.Force, min(uint16(200), u.Force/100))
		}
		units = append(units, row)
	}
	d.UnitPages = max(1, (len(units)+2)/3)
	v.UnitPage = max(0, min(v.UnitPage, d.UnitPages-1))
	d.UnitPage = v.UnitPage
	start := v.UnitPage * 3
	if start < len(units) {
		d.Units = units[start:min(start+3, len(units))]
	}
	d.Queue = a.armoryText("queue.empty")
	d.QueueNote = a.armoryText("queue.note")
	if n.Active != "" {
		m, _ := game.EquipmentByID(n.Active)
		d.Queue = fmt.Sprintf(a.armoryText("queue.active"), a.armoryText("model."+string(m.ID)), n.Progress[m.ID], m.Months)
		d.Progress = int(n.Progress[m.ID]) * 100 / int(m.Months)
		if d.Resources.Gold < 180 {
			d.QueueNote = a.armoryText("research.shortage")
		}
	}
	return d
}
func (a *app) composeArmory() error {
	if a.armory == nil {
		return game.ErrArmoryState
	}
	if a.frame != nil {
		a.frame.Dispose()
	}
	a.frame = ebiten.NewImageFromImage(render.DrawArmory(a.armoryData()).Image())
	a.dirty = false
	return nil
}
func (a *app) armoryAction(key string) {
	v := &a.armoryView
	v.Press = -1
	v.Focus = -1
	if v.PendingUnit != 0 {
		if key == "convert.confirm" {
			_, err := a.armory.ConvertToArmor(a.world, a.playerCommander, v.PendingUnit, v.Selected)
			v.Message = a.armoryText("convert.done")
			if err != nil {
				v.Message = a.armoryText("error")
			}
			a.syncExpansionGenerals()
			v.PendingUnit = 0
		} else if key == "convert.cancel" || key == "back" {
			v.PendingUnit = 0
		}
		a.dirty = true
		return
	}
	if strings.HasPrefix(key, "page.") {
		number, err := strconv.Atoi(strings.TrimPrefix(key, "page."))
		if err == nil && number >= 0 && number < 4 {
			v.Page = number
		}
		a.dirty = true
		return
	}
	if strings.HasPrefix(key, "category.") {
		number, err := strconv.Atoi(strings.TrimPrefix(key, "category."))
		if err == nil && number >= 0 && number < 3 {
			v.Category = number
			v.UnitPage = 0
			for _, m := range game.EquipmentCatalog() {
				if int(m.Class) == number && m.Months > 0 {
					v.Selected = m.ID
					break
				}
			}
		}
		a.dirty = true
		return
	}
	if strings.HasPrefix(key, "model.") {
		id := game.EquipmentID(strings.TrimPrefix(key, "model."))
		if m, ok := game.EquipmentByID(id); ok && int(m.Class) == v.Category {
			v.Selected = id
		}
		a.dirty = true
		return
	}
	var err error
	switch key {
	case "primary":
		n := a.armory.Nations[a.playerCommander]
		m, _ := game.EquipmentByID(v.Selected)
		if m.Legacy || n.Known[m.ID] {
			previous := v.Page
			v.Page = 1
			if m.Legacy || previous == 1 {
				v.Page = 2
			}
		} else if n.Active == m.ID {
			err = a.armory.Research(a.world, a.playerCommander, "")
			v.Message = a.armoryText("research.paused")
		} else {
			err = a.armory.Research(a.world, a.playerCommander, m.ID)
			v.Message = a.armoryText("research.started")
		}
	case "month":
		message := a.endTurn()
		fmt.Fprintln(os.Stderr, message)
		v.Message = fmt.Sprintf(a.armoryText("month.done"), int(a.year)+1911, a.month)
	case "previous":
		v.UnitPage = max(0, v.UnitPage-1)
	case "next":
		v.UnitPage++
	case "save":
		err = a.autosave()
		v.Message = a.armoryText("saved")
	case "load":
		var data []byte
		data, err = os.ReadFile(a.savePath)
		if err == nil {
			err = a.loadCampaign(data)
		}
		if err == nil {
			a.armoryView.Message = a.armoryText("loaded")
			a.applyWindowSize()
		}
	case "back":
		a.screen = screenMap
		a.applyWindowSize()
	default:
		if strings.HasPrefix(key, "convert.") {
			id, parseErr := strconv.Atoi(strings.TrimPrefix(key, "convert."))
			if parseErr == nil {
				_, err = a.armory.ConversionPlan(a.world, a.playerCommander, game.GeneralID(id), v.Selected)
				if err == nil {
					v.PendingUnit = game.GeneralID(id)
				}
			}
		}
		if strings.HasPrefix(key, "reserve.") {
			id, parseErr := strconv.Atoi(strings.TrimPrefix(key, "reserve."))
			if parseErr == nil {
				_, err = a.armory.ReinforceReserve(a.world, a.playerCommander, game.GeneralID(id))
				a.syncExpansionGenerals()
			}
		}
		if strings.HasPrefix(key, "plus.") || strings.HasPrefix(key, "minus.") {
			delta := 1
			prefix := "plus."
			if strings.HasPrefix(key, "minus.") {
				delta = -1
				prefix = "minus."
			}
			err = a.armory.Assign(a.world, a.playerCommander, game.EquipmentID(strings.TrimPrefix(key, prefix)), delta)
			v.Message = a.armoryText("assigned")
		}
		if strings.HasPrefix(key, "issue.") {
			id, parseErr := strconv.Atoi(strings.TrimPrefix(key, "issue."))
			if parseErr == nil {
				var count uint32
				count, err = a.armory.Issue(a.world, a.playerCommander, game.GeneralID(id), v.Selected)
				v.Message = fmt.Sprintf(a.armoryText("issued"), count)
			}
		}
	}
	if err != nil {
		a.armoryView.Message = a.armoryText("error")
		fmt.Fprintln(os.Stderr, "[armory]", err)
	}
	a.dirty = true
}
func (a *app) updateArmory() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) && a.gameMouseEnabled() {
		a.armoryAction("back")
		return nil
	}
	if a.armoryView.PendingUnit != 0 {
		if enterPressed() || inpututil.IsKeyJustPressed(ebiten.KeyY) {
			a.armoryAction("convert.confirm")
			return nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyN) {
			a.armoryAction("convert.cancel")
			return nil
		}
		return a.updateArmoryPointer()
	}
	for i, key := range []ebiten.Key{ebiten.KeyDigit1, ebiten.KeyDigit2, ebiten.KeyDigit3, ebiten.KeyDigit4} {
		if inpututil.IsKeyJustPressed(key) {
			a.armoryAction(fmt.Sprintf("page.%d", i))
			return nil
		}
	}
	for _, entry := range []struct {
		key    ebiten.Key
		action string
	}{{ebiten.KeyE, "month"}, {ebiten.KeyPageUp, "previous"}, {ebiten.KeyPageDown, "next"}} {
		if inpututil.IsKeyJustPressed(entry.key) {
			a.armoryAction(entry.action)
			return nil
		}
	}
	if ebiten.IsKeyPressed(ebiten.KeyControl) {
		if inpututil.IsKeyJustPressed(ebiten.KeyS) {
			a.armoryAction("save")
			return nil
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyL) {
			a.armoryAction("load")
			return nil
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) {
		a.armoryAction(fmt.Sprintf("category.%d", (a.armoryView.Category+2)%3))
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) {
		a.armoryAction(fmt.Sprintf("category.%d", (a.armoryView.Category+1)%3))
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
		a.armoryAction(fmt.Sprintf("page.%d", (a.armoryView.Page+3)%4))
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		a.armoryAction(fmt.Sprintf("page.%d", (a.armoryView.Page+1)%4))
		return nil
	}
	targets := render.ArmoryTargets(a.armoryData())
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		delta := 1
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			delta = -1
		}
		for i := 0; i < len(targets); i++ {
			a.armoryView.Focus = (a.armoryView.Focus + delta + len(targets)) % len(targets)
			if !targets[a.armoryView.Focus].Disabled {
				break
			}
		}
		a.dirty = true
	}
	if enterPressed() || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		focus := a.armoryView.Focus
		if focus >= 0 && focus < len(targets) {
			if !targets[focus].Disabled {
				a.armoryAction(targets[focus].Key)
			}
		} else {
			for _, t := range targets {
				if t.Key == "primary" && !t.Disabled {
					a.armoryAction(t.Key)
					break
				}
			}
		}
		return nil
	}
	return a.updateArmoryPointer()
}

func (a *app) updateArmoryPointer() error {
	targets := render.ArmoryTargets(a.armoryData())
	x, y, valid := a.gameCursorPosition()
	hit := -1
	if valid {
		for i, t := range targets {
			r := t.Rect
			if x >= r.X && x < r.X+r.W && y >= r.Y && y < r.Y+r.H {
				hit = i
				break
			}
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		a.armoryView.Press = hit
	}
	if inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft) {
		pressed := a.armoryView.Press
		a.armoryView.Press = -1
		if pressed >= 0 && pressed == hit && !targets[hit].Disabled {
			a.armoryAction(targets[hit].Key)
		}
	}
	return nil
}
