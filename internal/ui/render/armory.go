package render

import (
	"fmt"
	"math"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
)

type ArmoryModel struct {
	Model                                        game.EquipmentModel
	Name, Role, Description, Status, Terms, Last string
	Known, Enabled, Researching                  bool
	Progress, Factories                          int
	Stock                                        uint32
}
type ArmoryUnit struct {
	ID                       game.GeneralID
	Name, Summary, Equipment string
	Need, Selected           uint32
	Enabled                  bool
	Convert, Reinforce       bool
}
type ArmoryData struct {
	Fonts                                      *assets.EtenFonts
	Labels                                     map[string]string
	Page, Category, Focus, UnitPage, UnitPages int
	Resources                                  game.ArmoryCost
	Factories, Assigned                        uint32
	Date, Message, Queue, QueueNote            string
	Progress                                   int
	Models                                     []ArmoryModel
	Selected                                   ArmoryModel
	Units                                      []ArmoryUnit
	Quit                                       bool
	Pending                                    bool
	Confirmation                               string
	Reserve                                    uint32
}
type ArmoryTarget struct {
	Key      string
	Rect     l.Rect
	Disabled bool
}

func ArmoryTargets(d ArmoryData) []ArmoryTarget {
	if d.Pending {
		return []ArmoryTarget{{Key: "convert.confirm", Rect: l.Rect{400, 430, 230, 40}}, {Key: "convert.cancel", Rect: l.Rect{650, 430, 230, 40}}}
	}
	out := []ArmoryTarget{}
	for i := 0; i < 4; i++ {
		out = append(out, ArmoryTarget{Key: fmt.Sprintf("page.%d", i), Rect: l.Rect{20 + i*310, 126, 310, 52}})
	}
	for i := 0; i < 3; i++ {
		out = append(out, ArmoryTarget{Key: fmt.Sprintf("category.%d", i), Rect: l.Rect{32, 235 + i*61, 126, 53}})
	}
	row := 0
	for _, m := range d.Models {
		if d.Page == 0 && !m.Model.Legacy {
			y := 236
			if d.Category == 2 {
				y = 272
			}
			out = append(out, ArmoryTarget{Key: "model." + string(m.Model.ID), Rect: l.Rect{194 + (row%2)*398, y + (row/2)*122, 388, 112}})
			row++
		}
		if d.Page == 1 && m.Known && !m.Model.Legacy {
			y := 276 + row*64
			out = append(out, ArmoryTarget{Key: "minus." + string(m.Model.ID), Rect: l.Rect{864, y + 13, 34, 30}, Disabled: m.Factories == 0}, ArmoryTarget{Key: "plus." + string(m.Model.ID), Rect: l.Rect{936, y + 13, 34, 30}, Disabled: d.Assigned >= d.Factories})
			row++
		}
		if d.Page == 2 {
			out = append(out, ArmoryTarget{Key: "model." + string(m.Model.ID), Rect: l.Rect{194, 236 + row*54, 776, 49}})
			row++
		}
	}
	if d.Page == 3 {
		for i, u := range d.Units {
			key := "issue"
			if u.Convert {
				key = "convert"
			}
			out = append(out, ArmoryTarget{Key: fmt.Sprintf("%s.%d", key, u.ID), Rect: l.Rect{817, 283 + i*72, 146, 32}, Disabled: !u.Enabled})
			if u.Reinforce && !u.Convert {
				out = append(out, ArmoryTarget{Key: fmt.Sprintf("reserve.%d", u.ID), Rect: l.Rect{817, 317 + i*72, 146, 22}})
			}
		}
		out = append(out, ArmoryTarget{Key: "previous", Rect: l.Rect{194, 508, 110, 28}, Disabled: d.UnitPage == 0}, ArmoryTarget{Key: "next", Rect: l.Rect{860, 508, 110, 28}, Disabled: d.UnitPage+1 >= d.UnitPages})
	}
	disabled := !d.Selected.Known && !d.Selected.Enabled && !d.Selected.Researching
	out = append(out, ArmoryTarget{Key: "primary", Rect: l.Rect{1012, 478, 236, 34}, Disabled: disabled})
	if d.Page == 3 {
		for i, m := range d.Models {
			out = append(out, ArmoryTarget{Key: "model." + string(m.Model.ID), Rect: l.Rect{1012 + (i%2)*120, 519 + (i/2)*18, 116, 16}})
		}
	}
	return append(out, ArmoryTarget{Key: "month", Rect: l.Rect{1067, 590, 180, 38}}, ArmoryTarget{Key: "save", Rect: l.Rect{950, 674, 96, 26}}, ArmoryTarget{Key: "load", Rect: l.Rect{1054, 674, 96, 26}}, ArmoryTarget{Key: "back", Rect: l.Rect{1158, 674, 102, 26}})
}

func DrawArmory(d ArmoryData) *Canvas {
	c := NewCanvas(1280, 720)
	ink, muted, gold := openingColor(229, 212, 181), openingColor(179, 165, 142), openingColor(209, 170, 104)
	bg, panel, line := openingColor(25, 24, 23), openingColor(39, 36, 32), openingColor(101, 87, 71)
	openingFill(c, l.Rect{0, 0, 1280, 720}, bg)
	label := func(text string, r l.Rect, col assets.RGB, size int) {
		roomLabel(c, d.Fonts, text, r, col, size, false)
	}
	word := func(key string) string { return d.Labels[key] }
	openingFill(c, l.Rect{0, 0, 1280, 38}, openingColor(32, 30, 27))
	for i, value := range []uint32{d.Resources.Gold, d.Resources.Iron, d.Resources.Coal, d.Resources.Fuel} {
		label(fmt.Sprintf("%s  %d", word(fmt.Sprintf("resource.%d", i)), value), l.Rect{20 + i*151, 7, 149, 24}, ink, 15)
	}
	label(fmt.Sprintf(word("factories"), d.Assigned, d.Factories), l.Rect{642, 7, 270, 24}, gold, 14)
	label(d.Date, l.Rect{1012, 7, 248, 24}, muted, 13)
	label(word("kicker"), l.Rect{20, 48, 680, 17}, gold, 11)
	label(word("title"), l.Rect{20, 65, 700, 31}, ink, 24)
	label(word("intro"), l.Rect{20, 99, 1000, 17}, muted, 12)
	label(word("mode"), l.Rect{1000, 65, 260, 32}, gold, 12)
	for _, r := range []l.Rect{{20, 188, 150, 377}, {182, 188, 806, 377}, {1000, 188, 260, 377}} {
		openingFill(c, r, line)
		openingFill(c, l.Rect{r.X + 1, r.Y + 1, r.W - 2, r.H - 2}, panel)
	}
	targets := ArmoryTargets(d)
	find := func(key string) ArmoryTarget {
		for _, t := range targets {
			if t.Key == key {
				return t
			}
		}
		return ArmoryTarget{}
	}
	button := func(key, text string, active, primary bool) {
		t := find(key)
		if t.Rect.W == 0 {
			return
		}
		openingBox(c, t.Rect, active, t.Disabled, primary)
		color := ink
		if t.Disabled {
			color = muted
		}
		label(text, l.Rect{t.Rect.X + 8, t.Rect.Y + 2, t.Rect.W - 16, t.Rect.H - 4}, color, 13)
	}
	for i := 0; i < 4; i++ {
		t := find(fmt.Sprintf("page.%d", i))
		openingBox(c, t.Rect, d.Page == i, false, false)
		label(fmt.Sprintf("%02d  %s", i+1, word(fmt.Sprintf("page.%d", i))), l.Rect{t.Rect.X + 14, t.Rect.Y + 6, t.Rect.W - 28, 21}, ink, 14)
		label(word(fmt.Sprintf("step.%d", i)), l.Rect{t.Rect.X + 39, t.Rect.Y + 28, t.Rect.W - 50, 17}, muted, 11)
	}
	label(word("categories"), l.Rect{32, 201, 126, 25}, ink, 17)
	for i := 0; i < 3; i++ {
		t := find(fmt.Sprintf("category.%d", i))
		openingBox(c, t.Rect, d.Category == i, false, false)
		label(word(fmt.Sprintf("category.%d", i)), l.Rect{t.Rect.X + 8, t.Rect.Y + 6, t.Rect.W - 16, 23}, ink, 14)
		label(word(fmt.Sprintf("category.hint.%d", i)), l.Rect{t.Rect.X + 8, t.Rect.Y + 29, t.Rect.W - 16, 17}, muted, 10)
	}
	roomParagraph(c, d.Fonts, word("sidebar.note"), l.Rect{32, 437, 126, 110}, muted)
	label(word(fmt.Sprintf("page.%d", d.Page)), l.Rect{194, 201, 490, 26}, ink, 18)
	selected := d.Selected
	if d.Page == 0 {
		if d.Category == 2 {
			openingFill(c, l.Rect{194, 235, 776, 28}, openingColor(53, 43, 32))
			label(word("artillery.note"), l.Rect{204, 238, 756, 22}, muted, 12)
		}
		for _, m := range d.Models {
			if m.Model.Legacy {
				continue
			}
			t := find("model." + string(m.Model.ID))
			openingBox(c, t.Rect, m.Model.ID == selected.Model.ID, false, false)
			label(m.Status, l.Rect{t.Rect.X + 8, t.Rect.Y + 8, t.Rect.W - 16, 16}, muted, 10)
			label(m.Name, l.Rect{t.Rect.X + 8, t.Rect.Y + 27, t.Rect.W - 16, 23}, ink, 14)
			armoryWeapon(c, m.Model.Class, l.Rect{t.Rect.X + 8, t.Rect.Y + 50, 130, 25}, ink)
			label(m.Role, l.Rect{t.Rect.X + 8, t.Rect.Y + 76, t.Rect.W - 16, 16}, muted, 10)
			label(m.Terms, l.Rect{t.Rect.X + 8, t.Rect.Y + 94, t.Rect.W - 16, 16}, gold, 10)
		}
		label(word("catalog.note"), l.Rect{194, 537, 776, 20}, muted, 11)
	}
	if d.Page == 1 {
		label(fmt.Sprintf(word("capacity"), d.Factories, d.Assigned), l.Rect{194, 236, 776, 30}, gold, 13)
		row := 0
		for _, m := range d.Models {
			if !m.Known || m.Model.Legacy {
				continue
			}
			y := 276 + row*64
			openingFill(c, l.Rect{194, y, 776, 58}, line)
			openingFill(c, l.Rect{195, y + 1, 774, 56}, openingColor(47, 41, 34))
			label(m.Name, l.Rect{205, y + 3, 610, 22}, ink, 14)
			label(m.Terms, l.Rect{205, y + 25, 610, 17}, muted, 11)
			label(m.Last, l.Rect{205, y + 42, 610, 15}, gold, 10)
			button("minus."+string(m.Model.ID), "−", false, false)
			label(fmt.Sprint(m.Factories), l.Rect{902, y + 13, 30, 30}, ink, 15)
			button("plus."+string(m.Model.ID), "+", false, false)
			row++
		}
		if row == 0 {
			roomParagraph(c, d.Fonts, word("production.locked"), l.Rect{194, 285, 760, 65}, muted)
		}
		label(word("production.note"), l.Rect{194, 540, 776, 18}, muted, 11)
	}
	if d.Page == 2 {
		for _, m := range d.Models {
			t := find("model." + string(m.Model.ID))
			openingBox(c, t.Rect, m.Model.ID == selected.Model.ID, false, false)
			label(m.Name, l.Rect{t.Rect.X + 10, t.Rect.Y + 3, 600, 22}, ink, 14)
			label(m.Role, l.Rect{t.Rect.X + 10, t.Rect.Y + 26, 600, 17}, muted, 11)
			label(fmt.Sprintf("%d %s", m.Stock, word("unit")), l.Rect{t.Rect.X + 610, t.Rect.Y + 8, 155, 28}, gold, 18)
		}
		label(word("stock.note"), l.Rect{194, 540, 776, 18}, muted, 11)
	}
	if d.Page == 3 {
		label(fmt.Sprintf(word("selected.stock"), selected.Name, selected.Stock), l.Rect{194, 236, 776, 28}, gold, 13)
		for i, u := range d.Units {
			y := 274 + i*72
			openingFill(c, l.Rect{194, y, 776, 65}, line)
			openingFill(c, l.Rect{195, y + 1, 774, 63}, openingColor(47, 41, 34))
			label(u.Name, l.Rect{205, y + 3, 600, 21}, ink, 14)
			label(u.Summary, l.Rect{205, y + 26, 600, 17}, muted, 11)
			label(u.Equipment, l.Rect{205, y + 44, 600, 17}, gold, 10)
			key := "issue"
			if u.Convert {
				key = "convert"
			}
			button(fmt.Sprintf("%s.%d", key, u.ID), word(key), false, false)
			if u.Reinforce && !u.Convert {
				button(fmt.Sprintf("reserve.%d", u.ID), word("reserve.issue"), false, false)
			}
		}
		if len(d.Units) == 0 {
			roomParagraph(c, d.Fonts, word("units.empty"), l.Rect{194, 282, 750, 80}, muted)
		}
		button("previous", word("previous"), false, false)
		button("next", word("next"), false, false)
		label(fmt.Sprintf(word("pages"), d.UnitPage+1, max(1, d.UnitPages)), l.Rect{320, 508, 520, 28}, muted, 12)
		label(word("equipment.note"), l.Rect{194, 540, 776, 18}, muted, 11)
	}
	label(word(fmt.Sprintf("category.%d", d.Category)), l.Rect{1012, 200, 236, 15}, gold, 11)
	label(selected.Name, l.Rect{1012, 217, 236, 29}, ink, 20)
	openingFill(c, l.Rect{1012, 254, 236, 60}, openingColor(40, 40, 39))
	for x := 1012; x < 1248; x += 16 {
		c.line(x, 254, x, 314, line)
	}
	for y := 254; y < 314; y += 16 {
		c.line(1012, y, 1248, y, line)
	}
	armoryWeapon(c, selected.Model.Class, l.Rect{1022, 259, 216, 50}, ink)
	progress := selected.Status
	if !selected.Known {
		progress = fmt.Sprintf(word("months.progress"), selected.Progress, selected.Model.Months)
	}
	values := []string{selected.Role, progress, fmt.Sprintf("%d%% / %d%%", selected.Model.Power, selected.Model.RoughPower), fmt.Sprintf(word("output"), selected.Model.Output), fmt.Sprintf("%d %s", selected.Stock, word("unit"))}
	for i, value := range values {
		label(word(fmt.Sprintf("fact.%d", i)), l.Rect{1012, 323 + i*18, 85, 18}, muted, 11)
		label(value, l.Rect{1100, 323 + i*18, 148, 18}, ink, 11)
	}
	roomParagraph(c, d.Fonts, selected.Description, l.Rect{1012, 417, 236, 54}, muted)
	action := word("research.start")
	if selected.Researching {
		action = word("research.pause")
	}
	if selected.Known {
		action = word("production.open")
		if selected.Model.Legacy || d.Page == 1 {
			action = word("stock.open")
		}
	}
	button("primary", action, false, true)
	if d.Page == 3 {
		for _, m := range d.Models {
			button("model."+string(m.Model.ID), m.Name, m.Model.ID == selected.Model.ID, false)
		}
	}
	openingFill(c, l.Rect{20, 575, 1240, 65}, line)
	openingFill(c, l.Rect{21, 576, 1238, 63}, panel)
	label(d.Queue, l.Rect{32, 584, 1000, 22}, ink, 13)
	label(d.QueueNote, l.Rect{32, 606, 1000, 17}, muted, 11)
	openingFill(c, l.Rect{32, 626, 1020, 4}, bg)
	openingFill(c, l.Rect{32, 626, 1020 * min(100, max(0, d.Progress)) / 100, 4}, gold)
	button("month", word("month"), false, true)
	label(d.Message, l.Rect{20, 646, 1240, 22}, gold, 12)
	if d.Reserve > 0 {
		label(fmt.Sprintf(word("reserve"), d.Reserve), l.Rect{32, 409, 126, 24}, gold, 12)
	}
	label(word("footer"), l.Rect{20, 674, 910, 26}, muted, 11)
	button("save", word("save"), false, false)
	button("load", word("load"), false, false)
	button("back", word("back"), false, false)
	if d.Focus >= 0 && d.Focus < len(targets) {
		r := targets[d.Focus].Rect
		openingFill(c, l.Rect{r.X + 3, r.Y + r.H - 4, r.W - 6, 2}, gold)
	}
	if d.Pending {
		openingBox(c, l.Rect{360, 220, 560, 280}, true, false, false)
		roomParagraph(c, d.Fonts, d.Confirmation, l.Rect{390, 245, 500, 165}, ink)
		button("convert.confirm", word("convert.confirm"), false, true)
		button("convert.cancel", word("convert.cancel"), false, false)
	}
	if d.Quit {
		openingBox(c, l.Rect{380, 285, 520, 150}, true, false, false)
		roomParagraph(c, d.Fonts, word("quit"), l.Rect{405, 313, 470, 90}, ink)
	}
	return c
}

func armoryWeapon(c *Canvas, class game.EquipmentClass, r l.Rect, ink assets.RGB) {
	paths := [][][2]int{}
	switch class {
	case game.EquipmentRifle:
		paths = [][][2]int{{{18, 39}, {54, 31}, {80, 31}, {84, 25}, {167, 25}, {167, 29}, {91, 31}, {83, 36}, {58, 37}, {27, 53}, {18, 51}, {18, 39}}, {{85, 33}, {96, 33}, {93, 44}, {86, 44}}, {{136, 23}, {136, 16}, {141, 16}, {141, 23}}, {{157, 23}, {157, 18}, {161, 18}, {161, 23}}}
	case game.EquipmentArmor:
		paths = [][][2]int{{{35, 38}, {45, 30}, {116, 30}, {133, 38}, {134, 47}, {35, 47}, {35, 38}}, {{65, 29}, {71, 17}, {101, 17}, {112, 29}}, {{103, 20}, {168, 20}, {168, 24}, {108, 24}}, {{34, 43}, {137, 43}, {137, 63}, {34, 63}, {34, 43}}}
		for _, x := range []int{47, 67, 87, 107, 125} {
			paths = append(paths, armoryCircle(x, 53, 6))
		}
	case game.EquipmentArtillery:
		paths = [][][2]int{{{64, 34}, {143, 14}, {145, 22}, {69, 41}}, {{71, 41}, {117, 56}}, {{69, 42}, {42, 58}}, {{112, 51}, {130, 59}}, {{74, 35}, {74, 62}}, {{61, 48}, {88, 48}}, armoryCircle(74, 48, 14), armoryCircle(74, 48, 4)}
	}
	for _, path := range paths {
		for i := 1; i < len(path); i++ {
			p, q := path[i-1], path[i]
			c.line(r.X+p[0]*r.W/190, r.Y+p[1]*r.H/75, r.X+q[0]*r.W/190, r.Y+q[1]*r.H/75, ink)
		}
	}
}
func armoryCircle(x, y, radius int) [][2]int {
	out := make([][2]int, 17)
	for i := range out {
		angle := float64(i) * math.Pi / 8
		out[i] = [2]int{x + int(math.Cos(angle)*float64(radius)), y + int(math.Sin(angle)*float64(radius))}
	}
	return out
}
