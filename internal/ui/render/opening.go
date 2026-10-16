package render

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/wicanr2/great-era-remake/internal/assets"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
)

type OpeningFaction struct {
	Name, Summary, Home string
	Disabled            bool
}
type OpeningData struct {
	Fonts                        *assets.EtenFonts
	Labels                       map[string]string
	Stages, Dates                []string
	Factions                     []OpeningFaction
	Stage, Selected, Mode, Focus int
	Name, Home, Territory, Date  string
	Provinces, Generals          int
	Force                        uint32
	Notice, Message              string
	CanContinue, Quit            bool
}

// OpeningTargets 與繪製共用幾何；座標不含常駐工具列。
func OpeningTargets(factions int) []l.Rect {
	out := []l.Rect{{18, 115, 616, 56}, {646, 115, 616, 56}, {18, 184, 407, 57}, {437, 184, 406, 57}, {855, 184, 407, 57}}
	for i := 0; i < factions; i++ {
		out = append(out, l.Rect{35 + (i%2)*355, 307 + (i/2)*64, 346, 54})
	}
	return append(out, l.Rect{789, 599, 456, 48}, l.Rect{1114, 20, 148, 38})
}

func openingColor(r, g, b uint8) assets.RGB { return assets.RGB{R: r, G: g, B: b} }

func openingFill(c *Canvas, r l.Rect, fill assets.RGB) {
	draw.Draw(c.img, image.Rect(r.X, r.Y, r.X+r.W, r.Y+r.H), image.NewUniform(color.RGBA{R: fill.R, G: fill.G, B: fill.B, A: 255}), image.Point{}, draw.Src)
}

func openingBox(c *Canvas, r l.Rect, active, disabled, action bool) {
	top, bottom, border := openingColor(65, 58, 50), openingColor(41, 38, 34), openingColor(116, 96, 71)
	if active {
		top, bottom, border = openingColor(119, 92, 60), openingColor(73, 60, 45), openingColor(210, 173, 113)
	}
	if action {
		top, bottom, border = openingColor(139, 73, 53), openingColor(98, 51, 38), openingColor(186, 119, 81)
	}
	if disabled {
		top, bottom, border = openingColor(47, 42, 37), openingColor(35, 32, 29), openingColor(83, 71, 55)
	}
	openingFill(c, r, border)
	for y := 1; y < r.H-1; y++ {
		blend := func(a, b uint8) uint8 { return uint8(int(a) + (int(b)-int(a))*y/max(1, r.H-1)) }
		openingFill(c, l.Rect{r.X + 1, r.Y + y, r.W - 2, 1}, openingColor(blend(top.R, bottom.R), blend(top.G, bottom.G), blend(top.B, bottom.B)))
	}
	if active {
		openingFill(c, l.Rect{r.X + 1, r.Y + 1, 4, r.H - 2}, openingColor(187, 110, 69))
	}
}

// DrawOpening 沿用已採用的 1280×768 構圖，工具列由 app 額外繪製。
func DrawOpening(d OpeningData) *Canvas {
	c := NewCanvas(1280, 720)
	ink, muted, gold := openingColor(225, 207, 173), openingColor(196, 183, 163), openingColor(198, 148, 89)
	openingFill(c, l.Rect{0, 0, 1280, 720}, openingColor(25, 23, 22))
	label := func(s string, r l.Rect, col assets.RGB, size int) { roomLabel(c, d.Fonts, s, r, col, size, false) }
	label(d.Labels["eyebrow"], l.Rect{18, 18, 500, 20}, gold, 13)
	label(d.Labels["title"], l.Rect{18, 40, 900, 40}, openingColor(240, 219, 180), 28)
	label(d.Labels["intro"], l.Rect{18, 82, 1150, 24}, muted, 13)
	targets := OpeningTargets(len(d.Factions))
	for i, key := range []string{"original", "resistance"} {
		r := targets[i]
		openingBox(c, r, d.Mode == i, false, false)
		label(d.Labels[key], l.Rect{r.X + 13, r.Y + 7, r.W - 26, 22}, ink, 16)
		label(d.Labels[key+".hint"], l.Rect{r.X + 13, r.Y + 29, r.W - 26, 20}, muted, 12)
	}
	for i, name := range d.Stages {
		r := targets[2+i]
		openingBox(c, r, d.Stage == i, false, false)
		label(name, l.Rect{r.X + 13, r.Y + 6, r.W - 26, 23}, ink, 16)
		label(d.Dates[i], l.Rect{r.X + 13, r.Y + 29, r.W - 26, 20}, muted, 12)
	}
	for _, r := range []l.Rect{{18, 253, 734, 455}, {772, 253, 490, 455}} {
		openingFill(c, r, openingColor(119, 96, 67))
		openingFill(c, l.Rect{r.X + 1, r.Y + 1, r.W - 2, r.H - 2}, openingColor(39, 35, 31))
	}
	label(fmt.Sprintf(d.Labels["factions"], len(d.Factions)), l.Rect{35, 270, 650, 27}, ink, 20)
	for i, f := range d.Factions {
		r := targets[5+i]
		openingBox(c, r, d.Selected == i, f.Disabled, false)
		col := ink
		if f.Disabled {
			col = muted
		}
		label(f.Name, l.Rect{r.X + 12, r.Y + 5, r.W - 115, 24}, col, 17)
		label(f.Summary, l.Rect{r.X + 12, r.Y + 29, r.W - 24, 18}, muted, 12)
		label(f.Home, l.Rect{r.X + r.W - 105, r.Y + 13, 95, 24}, gold, 12)
	}
	label(d.Labels["source"], l.Rect{35, 633, 700, 25}, muted, 12)
	openingFill(c, l.Rect{789, 270, 456, 252}, openingColor(217, 185, 128))
	openingFill(c, l.Rect{790, 271, 454, 250}, openingColor(185, 155, 108))
	paperInk := openingColor(39, 30, 22)
	label(d.Date, l.Rect{810, 288, 410, 25}, openingColor(95, 52, 36), 14)
	label(d.Name, l.Rect{810, 313, 410, 42}, paperInk, 28)
	label(d.Home, l.Rect{810, 356, 410, 30}, paperInk, 16)
	openingFill(c, l.Rect{810, 391, 414, 1}, openingColor(120, 99, 67))
	for i, s := range []string{fmt.Sprint(d.Provinces), fmt.Sprint(d.Generals), fmt.Sprintf("%.1f", float64(d.Force)/10000)} {
		x := 810 + i*141
		label(s, l.Rect{x, 401, 135, 34}, paperInk, 26)
		label(d.Labels[[]string{"provinces", "generals", "force"}[i]], l.Rect{x, 439, 135, 22}, paperInk, 12)
	}
	openingFill(c, l.Rect{810, 468, 414, 1}, openingColor(120, 99, 67))
	roomParagraph(c, d.Fonts, d.Territory, l.Rect{810, 477, 414, 38}, paperInk)
	openingFill(c, l.Rect{789, 536, 3, 44}, openingColor(147, 74, 53))
	roomParagraph(c, d.Fonts, d.Notice, l.Rect{804, 536, 429, 48}, ink)
	start := targets[len(targets)-2]
	openingBox(c, start, false, d.Factions[d.Selected].Disabled, true)
	label(d.Labels["start"], l.Rect{start.X + 12, start.Y + 5, start.W - 24, start.H - 10}, ink, 18)
	roomParagraph(c, d.Fonts, d.Labels["save.hint"], l.Rect{789, 657, 456, 44}, muted)
	if d.CanContinue {
		r := targets[len(targets)-1]
		openingBox(c, r, false, false, false)
		label(d.Labels["continue"], l.Rect{r.X + 10, r.Y + 2, r.W - 20, r.H - 4}, ink, 14)
	}
	if d.Focus >= 0 && d.Focus < len(targets) {
		r := targets[d.Focus]
		openingFill(c, l.Rect{r.X + 3, r.Y + r.H - 4, r.W - 6, 2}, openingColor(227, 182, 97))
	}
	if d.Message != "" {
		openingFill(c, l.Rect{35, 666, 700, 33}, openingColor(25, 23, 22))
		label(d.Message, l.Rect{42, 670, 686, 25}, gold, 13)
	}
	if d.Quit {
		openingBox(c, l.Rect{380, 285, 520, 150}, true, false, false)
		label(d.Labels["quit"], l.Rect{400, 308, 480, 80}, ink, 20)
	}
	return c
}
