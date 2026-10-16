package render

import (
	"fmt"
	"github.com/wicanr2/great-era-remake/internal/assets"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
)

type BattleInformationSide struct {
	Name    string
	Summary []string
	Units   [][]string
}
type BattleInformationData struct {
	Fonts  *assets.EtenFonts
	Labels map[string]string
	Sides  [2]BattleInformationSide
	Page   int
}
type BattleInformationTarget struct {
	Key      string
	Rect     l.Rect
	Disabled bool
}

func BattleInformationPages(d BattleInformationData, high bool) int {
	n := 2
	if high {
		n = 3
	}
	return max(1, (max(len(d.Sides[0].Units), len(d.Sides[1].Units))+n-1)/n)
}
func BattleInformationTargets(d BattleInformationData, w, h int) []BattleInformationTarget {
	pages := BattleInformationPages(d, w >= 1000)
	pad, bw, bh := 8, 104, 24
	if w >= 1000 {
		pad, bw, bh = 20, 180, 36
	}
	return []BattleInformationTarget{{Key: "previous", Rect: l.Rect{pad, h - bh - pad, bw, bh}, Disabled: d.Page == 0}, {Key: "next", Rect: l.Rect{pad + bw + 8, h - bh - pad, bw, bh}, Disabled: d.Page+1 >= pages}, {Key: "back", Rect: l.Rect{w - pad - bw, h - bh - pad, bw, bh}}}
}

func (c *Canvas) DrawBattleInformation(d BattleInformationData) {
	w, h := c.Bounds().Dx(), c.Bounds().Dy()
	high := w >= 1000
	pad, header, summary, rowH, size, lineH, count := 8, 25, 109, 90, 10, 12, 2
	if high {
		pad, header, summary, rowH, size, lineH, count = 20, 48, 174, 144, 15, 19, 3
	}
	ink, muted, gold := openingColor(229, 212, 181), openingColor(179, 165, 142), openingColor(209, 170, 104)
	bg, panel, line := openingColor(25, 24, 23), openingColor(39, 36, 32), openingColor(101, 87, 71)
	openingFill(c, l.Rect{0, 0, w, h}, bg)
	label := func(text string, r l.Rect, col assets.RGB, s int) { roomLabel(c, d.Fonts, text, r, col, s, false) }
	label(d.Labels["title"], l.Rect{pad, 4, w - 2*pad, header - 6}, ink, size+4)
	column := (w - 3*pad) / 2
	for side, info := range d.Sides {
		x := pad + side*(column+pad)
		openingFill(c, l.Rect{x, header, column, h - header - 46}, panel)
		openingFill(c, l.Rect{x, header, column, 2}, line)
		label(info.Name, l.Rect{x + 8, header + 5, column - 16, lineH + 3}, gold, size+2)
		for i, text := range info.Summary {
			label(text, l.Rect{x + 8, header + lineH + 12 + i*lineH, column - 16, lineH}, muted, size)
		}
		start := d.Page * count
		if start >= len(info.Units) {
			label(d.Labels["empty"], l.Rect{x + 8, header + summary, column - 16, rowH}, muted, size)
		}
		for i := 0; i < count && start+i < len(info.Units); i++ {
			y := header + summary + i*rowH
			openingFill(c, l.Rect{x + 5, y, column - 10, 1}, line)
			for j, text := range info.Units[start+i] {
				col := ink
				if j == 0 {
					col = gold
				}
				label(text, l.Rect{x + 8, y + 4 + j*lineH, column - 16, lineH}, col, size)
			}
		}
	}
	for _, target := range BattleInformationTargets(d, w, h) {
		openingBox(c, target.Rect, false, target.Disabled, false)
		label(d.Labels[target.Key], l.Rect{target.Rect.X + 7, target.Rect.Y + 3, target.Rect.W - 14, target.Rect.H - 6}, ink, size)
	}
	pages := BattleInformationPages(d, high)
	label(fmt.Sprintf(d.Labels["pages"], d.Page+1, pages), l.Rect{w/2 - 60, h - 34, 120, 24}, muted, size)
}
