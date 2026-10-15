package render

import (
	"image"
	"image/color"
	"image/draw"
	"strings"

	"github.com/wicanr2/great-era-remake/internal/assets"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
	"github.com/wicanr2/great-era-remake/internal/ui/textlayout"
)

type ToolbarView struct {
	Labels, Values     [l.ToolbarFields]string
	Options            [l.ToolbarFields][]string
	Selected           [l.ToolbarFields]int
	Checked            [l.ToolbarFields][]bool
	Notes              []string
	Menu, Hover, Focus int
}

func ToolbarNoteLines(text string, width int) []string {
	if text == "" {
		return nil
	}
	f, err := assets.EmbeddedModernFontFamily(assets.ModernFontSans)
	if err != nil {
		return strings.Split(text, "\n")
	}
	doc, err := textlayout.LayoutProportional(text, textlayout.ProportionalOptions{
		Width: width, Rows: 100, Advance: func(r rune) int { return roomLabelWidth(f, string(r), 13) },
	})
	if err != nil || len(doc.Pages) == 0 {
		return nil
	}
	var lines []string
	for _, page := range doc.Pages {
		for _, line := range page.Lines {
			lines = append(lines, strings.TrimSpace(line.Text))
		}
	}
	return lines
}

func (v ToolbarView) MenuRect() l.Rect {
	if v.Menu < 0 || v.Menu >= l.ToolbarFields {
		return l.Rect{}
	}
	return l.ToolbarMenu(v.Menu, len(v.Options[v.Menu]), len(v.Notes))
}

func (v ToolbarView) Hit(x, y int) (int, int) {
	rows := 0
	if v.Menu >= 0 && v.Menu < l.ToolbarFields {
		rows = len(v.Options[v.Menu])
	}
	return l.ToolbarHit(x, y, v.Menu, rows, len(v.Notes))
}

func toolbarTextRects(r l.Rect, label string) (l.Rect, l.Rect) {
	width := 44
	if f, err := assets.EmbeddedModernFontFamily(assets.ModernFontSans); err == nil {
		width = min(r.W/3, roomLabelWidth(f, label, 13))
	}
	name := l.Rect{X: r.X + 12, Y: r.Y, W: width, H: r.H}
	value := l.Rect{X: name.Right() + 12, Y: r.Y, W: r.Right() - 28 - (name.Right() + 12), H: r.H}
	return name, value
}

// DrawToolbar only fills the strip and open menu. The remaining pixels stay
// transparent so the full game remains visible, including in original mode.
func (c *Canvas) DrawToolbar(fonts *assets.EtenFonts, v ToolbarView) {
	draw.Draw(c.img, c.Bounds(), image.NewUniform(color.RGBA{}), image.Point{}, draw.Src)
	ink, muted := assets.RGB{R: 233, G: 223, B: 204}, assets.RGB{R: 199, G: 182, B: 155}
	edge, accent := assets.RGB{R: 118, G: 99, B: 78}, assets.RGB{R: 177, G: 139, B: 88}
	gradient := func(r l.Rect, top, bottom assets.RGB) {
		for y := 0; y < r.H; y++ {
			mix := func(a, b byte) byte { return byte((int(a)*(r.H-1-y) + int(b)*y) / max(1, r.H-1)) }
			c.fillRect(r.X, r.Y+y, r.W, 1, assets.RGB{R: mix(top.R, bottom.R), G: mix(top.G, bottom.G), B: mix(top.B, bottom.B)})
		}
	}
	gradient(l.Rect{W: l.ToolbarWidth, H: l.ToolbarHeight}, assets.RGB{R: 60, G: 52, B: 44}, assets.RGB{R: 39, G: 35, B: 31})
	c.fillRect(0, 46, l.ToolbarWidth, 2, assets.RGB{R: 146, G: 115, B: 72})
	for i := 0; i < l.ToolbarFields; i++ {
		r := l.ToolbarField(i)
		border := edge
		if v.Menu == i || v.Focus == i {
			c.fillRect(r.X, r.Y, r.W, r.H, assets.RGB{R: 101, G: 80, B: 58})
			border = accent
		} else {
			gradient(r, assets.RGB{R: 74, G: 65, B: 55}, assets.RGB{R: 51, G: 45, B: 39})
		}
		c.strokeRect(r.X, r.Y, r.W, r.H, border)
		name, value := toolbarTextRects(r, v.Labels[i])
		roomLabel(c, fonts, v.Labels[i], name, muted, 13, false)
		roomLabel(c, fonts, v.Values[i], value, ink, 16, false)
		for line := 0; line < 5; line++ {
			c.fillRect(r.Right()-22+line, r.Y+13+line, 10-2*line, 1, accent)
		}
	}
	if v.Menu < 0 || v.Menu >= l.ToolbarFields {
		return
	}
	r := v.MenuRect()
	c.fillRect(r.X, r.Y, r.W, r.H, assets.RGB{R: 41, G: 36, B: 31})
	c.strokeRect(r.X, r.Y, r.W, r.H, accent)
	for row, label := range v.Options[v.Menu] {
		y := r.Y + 5 + row*l.ToolbarRow
		checked := row == v.Selected[v.Menu]
		if v.Checked[v.Menu] != nil {
			checked = row < len(v.Checked[v.Menu]) && v.Checked[v.Menu][row]
		}
		if row == v.Hover || checked {
			c.fillRect(r.X+5, y, r.W-10, l.ToolbarRow, assets.RGB{R: 112, G: 88, B: 60})
		}
		if checked {
			// Draw a check without adding a new glyph or asset dependency.
			c.line(r.X+12, y+18, r.X+16, y+22, ink)
			c.line(r.X+16, y+22, r.X+23, y+13, ink)
		}
		roomLabel(c, fonts, label, l.Rect{X: r.X + 30, Y: y, W: r.W - 40, H: l.ToolbarRow}, ink, 16, false)
	}
	if len(v.Notes) > 0 {
		y := r.Y + 10 + len(v.Options[v.Menu])*l.ToolbarRow
		c.fillRect(r.X+11, y, r.W-22, 1, edge)
		for i, line := range v.Notes {
			roomLabel(c, fonts, line, l.Rect{X: r.X + 12, Y: y + 9 + i*22, W: r.W - 24, H: 22}, muted, 13, false)
		}
	}
}
