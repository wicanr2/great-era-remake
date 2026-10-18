package render

import (
	"fmt"
	"image"
	"image/draw"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
	"github.com/wicanr2/great-era-remake/internal/ui/textlayout"
)

// DrawClassicText leaves a small gap around 16/8 pixel bitmap glyphs.
func (c *Canvas) DrawClassicText(fonts *assets.EtenFonts, s string, r l.Rect, fg assets.RGB) []rune {
	if r.H < 16 || r.W <= 0 {
		return nil
	}
	if messageNeedsUnicode(fonts, s) {
		roomLabel(c, fonts, s, r, fg, 16, false)
		return nil
	}
	width := func(s string) int {
		n := 0
		for _, v := range s {
			n += textlayout.RuneHalfCells(v) * 9
		}
		return n
	}
	if width(s) > r.W {
		v := []rune(s)
		suffix := "..."
		for len(suffix) > 0 && width(suffix) > r.W {
			suffix = suffix[:len(suffix)-1]
		}
		for len(v) > 0 && width(string(v)+suffix) > r.W {
			v = v[:len(v)-1]
		}
		s = string(v) + suffix
	}
	x := r.X
	missing := []rune{}
	for _, v := range s {
		w := textlayout.RuneHalfCells(v) * 9
		if v != ' ' {
			if v >= 32 && v <= 126 {
				g, ok := fonts.GlyphASCII(v)
				if ok {
					c.drawASCIIGlyph(g, fg, x, r.Y)
				} else {
					missing = append(missing, v)
					c.drawMissingGlyph(fg, x, r.Y, w)
				}
			} else {
				g, ok := fonts.GlyphRune(v)
				if ok {
					c.DrawGlyph(g, fg, x, r.Y, true)
				} else {
					missing = append(missing, v)
					c.drawMissingGlyph(fg, x, r.Y, w)
				}
			}
		}
		x += w
	}
	return uniqueRunes(missing)
}

type ClassicMapMarker struct {
	ID, X, Y int
	Selected bool
}
type ClassicMapData struct {
	Fonts                                       *assets.EtenFonts
	Map                                         image.Image
	Title, Hint, Commands, Province, Home, News string
	Markers                                     []ClassicMapMarker
}

func ClassicMapRect() l.Rect { return l.Rect{X: 192, Y: 30, W: 446, H: 276} }
func ClassicMapButtons() []l.Rect {
	return []l.Rect{{200, 322, 100, 24}, {310, 322, 100, 24}, {420, 322, 100, 24}, {530, 322, 100, 24}}
}
func (c *Canvas) DrawClassicMap(d ClassicMapData, fg, bg assets.RGB) []rune {
	c.fillRect(190, 0, 450, 350, bg)
	missing := c.DrawClassicText(d.Fonts, d.Title, l.Rect{198, 5, 434, 18}, fg)
	r := ClassicMapRect()
	if d.Map != nil {
		draw.Draw(c.img, image.Rect(r.X, r.Y, r.Right(), r.Bottom()), d.Map, d.Map.Bounds().Min, draw.Src)
	}
	for _, m := range d.Markers {
		c.fillRect(m.X-9, m.Y-7, 20, 15, bg)
		if m.Selected {
			c.strokeRect(m.X-11, m.Y-9, 24, 19, fg)
		}
		missing = append(missing, c.DrawClassicText(d.Fonts, fmt.Sprint(m.ID), l.Rect{X: m.X - 9, Y: m.Y - 7, W: 20, H: 16}, fg)...)
	}
	missing = append(missing, c.DrawClassicText(d.Fonts, d.Hint, l.Rect{198, 306, 434, 16}, fg)...)
	for i, label := range []string{d.Commands, d.Province, d.Home, d.News} {
		if label == "" {
			continue
		}
		b := ClassicMapButtons()[i]
		c.strokeRect(b.X, b.Y, b.W, b.H, fg)
		missing = append(missing, c.DrawClassicText(d.Fonts, label, l.Rect{b.X + 5, b.Y + 4, b.W - 10, 16}, fg)...)
	}
	return uniqueRunes(missing)
}

type ClassicRosterRow struct {
	Name    string
	Cells   [4][3]string
	General game.General
	Fan     *assets.GlyphFile
}
type ClassicRosterData struct {
	Fonts                  *assets.EtenFonts
	Title, Hint, UnitTitle string
	Rows                   []ClassicRosterRow
	Selected               int
}

const ClassicRosterPageSize = 2

func (c *Canvas) DrawClassicRoster(d ClassicRosterData, fg, bg assets.RGB) []rune {
	c.fillRect(190, 0, 450, 350, bg)
	missing := c.DrawClassicText(d.Fonts, d.Title, l.Rect{198, 5, 330, 18}, fg)
	start := max(0, d.Selected) / ClassicRosterPageSize * ClassicRosterPageSize
	missing = append(missing, c.DrawClassicText(d.Fonts, fmt.Sprintf("%d/%d", start/2+1, max(1, (len(d.Rows)+1)/2)), l.Rect{550, 5, 82, 18}, fg)...)
	for i := start; i < min(start+ClassicRosterPageSize, len(d.Rows)); i++ {
		row := d.Rows[i]
		y := 30 + (i-start)*140
		if i == d.Selected {
			c.strokeRect(194, y, 442, 130, fg)
		}
		for line, cells := range row.Cells {
			for col, text := range cells {
				missing = append(missing, c.DrawClassicText(d.Fonts, text, l.Rect{200 + col*144, y + 6 + line*23, 136, 18}, fg)...)
			}
		}
		missing = append(missing, c.DrawClassicText(d.Fonts, d.UnitTitle, l.Rect{200, y + 101, 60, 18}, fg)...)
		wx := 264
		if row.General.TitlePrefix > 0 && row.Fan != nil {
			if err := c.DrawEntry(row.Fan, int(row.General.TitlePrefix)-1, 5, fg, wx, y+101, true); err == nil {
				wx += 5 * GlyphAdvance
			}
		}
		c.DrawNumber(uint32(row.General.TitleNumber), fg, wx+40, y+101)
		wx += 50
		if row.General.TitleSuffix > 0 && row.Fan != nil {
			_ = c.DrawEntry(row.Fan, int(row.General.TitleSuffix)-1, 5, fg, wx, y+101, true)
		}
	}
	missing = append(missing, c.DrawClassicText(d.Fonts, d.Hint, l.Rect{198, 314, 434, 32}, fg)...)
	return uniqueRunes(missing)
}
