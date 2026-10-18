package render

import (
	"fmt"
	"github.com/wicanr2/great-era-remake/internal/assets"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
	"github.com/wicanr2/great-era-remake/internal/ui/textlayout"
	"image"
	"image/color"
	"image/draw"
	"strconv"
	"strings"
)

type DeskControl struct {
	Key, Text, Kind          string
	Rect                     l.Rect
	Disabled, Selected, Body bool
	Value, Maximum           int
}
type DeskText struct {
	Text                   string
	Rect                   l.Rect
	Size                   int
	Muted, Gold, Paragraph bool
}
type DeskUnit struct{ Name, Detail, Force string }
type DeskIllustration struct {
	Image *image.RGBA
	Rect  l.Rect
}
type DeskDialog struct {
	Title, Text, Yes, No string
	YesDisabled          bool
}
type DeskData struct {
	LightUI                                                bool
	Fonts                                                  *assets.EtenFonts
	Labels                                                 map[string]string
	Title, Prompt, Province, Commander, Date, Page, Status string
	Resources                                              [6]string
	Force, Arsenals, Coal, Iron                            string
	Units                                                  []DeskUnit
	Controls                                               []DeskControl
	Texts                                                  []DeskText
	Cards                                                  []l.Rect
	Map                                                    *image.RGBA
	Illustrations                                          []DeskIllustration
	Scroll, ContentHeight                                  int
	Focus                                                  string
	Dialog                                                 *DeskDialog
	Portrait                                               *PortraitView
}

var deskInk = assets.RGB{R: 238, G: 230, B: 215}
var deskMuted = assets.RGB{R: 188, G: 176, B: 155}
var deskGold = assets.RGB{R: 206, G: 177, B: 122}
var deskBackground = assets.RGB{R: 25, G: 26, B: 28}
var deskPanel = assets.RGB{R: 38, G: 40, B: 43}
var deskLine = assets.RGB{R: 94, G: 85, B: 70}
var deskAccent = assets.RGB{R: 189, G: 114, B: 80}

func DeskParagraphHeight(s string, width int) int {
	doc, err := textlayout.LayoutProportional(s, textlayout.ProportionalOptions{Width: width, Rows: max(1, len([]rune(s))+1), Advance: modernBiographyAdvance})
	if err != nil || len(doc.Pages) == 0 {
		return 20
	}
	return max(20, len(doc.Pages[0].Lines)*20)
}

func DeskLabelWidth(text string, size int) int {
	f, err := assets.EmbeddedModernFontFamily(assets.ModernFontSans)
	if err != nil {
		return len([]rune(text)) * size
	}
	return roomLabelWidth(f, text, size)
}
func DeskTargets(d DeskData) []DeskControl {
	if d.Dialog != nil {
		return []DeskControl{{Key: "dialog.no", Text: d.Dialog.No, Rect: l.DeskDialogNo}, {Key: "dialog.yes", Text: d.Dialog.Yes, Rect: l.DeskDialogYes, Disabled: d.Dialog.YesDisabled}}
	}
	out := make([]DeskControl, 0, len(d.Controls))
	for _, v := range d.Controls {
		if v.Body {
			v.Rect.Y -= d.Scroll
			r := v.Rect
			left, top := max(r.X, l.DeskContent.X), max(r.Y, l.DeskContent.Y)
			right, bottom := min(r.Right(), l.DeskContent.Right()), min(r.Bottom(), l.DeskContent.Bottom())
			if right <= left || bottom <= top {
				continue
			}
			v.Rect = l.Rect{left, top, right - left, bottom - top}
		}
		out = append(out, v)
	}
	return out
}
func DeskHit(d DeskData, x, y int) string {
	ts := DeskTargets(d)
	for i := len(ts) - 1; i >= 0; i-- {
		v := ts[i]
		if v.Rect.Contains(x, y) {
			if v.Disabled {
				return ""
			}
			return v.Key
		}
	}
	return ""
}
func deskFill(c *Canvas, r l.Rect, col assets.RGB) { c.fillUIRect(r.X, r.Y, r.W, r.H, col) }
func deskLabel(c *Canvas, f *assets.EtenFonts, text string, r l.Rect, col assets.RGB, size int) {
	roomLabel(c, f, text, r, col, size, false)
}
func deskGradient(c *Canvas, r l.Rect, top, bottom assets.RGB) {
	top, bottom = c.uiSurface(top), c.uiSurface(bottom)
	for y := 0; y < r.H; y++ {
		mix := func(a, b byte) byte { return byte((int(a)*(r.H-1-y) + int(b)*y) / max(1, r.H-1)) }
		c.fillUIRect(r.X, r.Y+y, r.W, 1, assets.RGB{R: mix(top.R, bottom.R), G: mix(top.G, bottom.G), B: mix(top.B, bottom.B)})
	}
}
func deskButton(c *Canvas, f *assets.EtenFonts, v DeskControl, focus string) {
	r := v.Rect
	top, bottom := assets.RGB{R: 54, G: 54, B: 56}, assets.RGB{R: 38, G: 37, B: 39}
	edge, ink := deskLine, deskInk
	if v.Selected {
		top, bottom = assets.RGB{R: 88, G: 70, B: 51}, assets.RGB{R: 88, G: 70, B: 51}
		edge = deskGold
	}
	if v.Kind == "primary" {
		top, bottom = assets.RGB{R: 152, G: 84, B: 58}, assets.RGB{R: 104, G: 58, B: 43}
		edge = assets.RGB{R: 180, G: 123, B: 86}
	}
	if v.Disabled {
		ink = assets.RGB{R: 110, G: 107, B: 100}
	}
	if v.Kind == "slider" {
		y := r.Y + r.H/2
		deskFill(c, l.Rect{r.X, y - 3, r.W, 6}, assets.RGB{R: 74, G: 74, B: 74})
		x := r.X
		if v.Maximum > 0 {
			x += min(v.Maximum, max(0, v.Value)) * (r.W - 10) / v.Maximum
		}
		deskFill(c, l.Rect{r.X, y - 3, max(0, x-r.X), 6}, deskAccent)
		deskFill(c, l.Rect{x, y - 6, 10, 12}, deskAccent)
		return
	}
	deskGradient(c, r, top, bottom)
	c.strokeUIRect(r.X, r.Y, r.W, r.H, edge)
	if v.Key == focus {
		c.strokeUIRect(r.X-2, r.Y-2, r.W+4, r.H+4, deskGold)
	}
	text := v.Text
	if v.Kind == "checkbox" {
		check := l.Rect{r.X + 12, r.Y + (r.H-18)/2, 18, 18}
		c.strokeUIRect(check.X, check.Y, 18, 18, deskGold)
		if v.Selected {
			deskFill(c, l.Rect{check.X + 4, check.Y + 4, 10, 10}, deskAccent)
		}
		r.X += 29
		r.W -= 29
	}
	size := 15
	if v.Kind == "marker" {
		size = 11
	}
	deskLabel(c, f, text, l.Rect{r.X + 12, r.Y, r.W - 24, r.H}, ink, size)
}
func DrawDesk(d DeskData) *Canvas {
	c := NewCanvas(l.DeskWidth, l.DeskHeight)
	c.UseLightUI(d.LightUI)
	deskFill(c, l.Rect{W: l.DeskWidth, H: l.DeskHeight}, deskBackground)
	word := func(k string) string { return d.Labels[k] }
	deskGradient(c, l.DeskHeader, assets.RGB{R: 40, G: 42, B: 45}, assets.RGB{R: 28, G: 30, B: 32})
	c.strokeUIRect(0, 75, 1280, 1, deskLine)
	deskLabel(c, d.Fonts, "GREAT ERA · REMAKE", l.Rect{20, 8, 210, 20}, deskMuted, 11)
	deskLabel(c, d.Fonts, word("title"), l.Rect{20, 29, 222, 36}, deskInk, 22)
	for i, value := range d.Resources[:5] {
		x := 266 + i*113
		deskLabel(c, d.Fonts, word(fmt.Sprintf("resource.%d", i)), l.Rect{x, 8, 110, 19}, deskMuted, 11)
		deskLabel(c, d.Fonts, value, l.Rect{x, 27, 110, 30}, deskInk, 19)
	}
	deskLabel(c, d.Fonts, d.Date, l.Rect{1085, 8, 175, 25}, deskInk, 15)
	deskLabel(c, d.Fonts, word("mode"), l.Rect{1120, 36, 140, 22}, deskGold, 11)
	deskFill(c, l.DeskSidebar, assets.RGB{R: 32, G: 34, B: 37})
	deskFill(c, l.DeskInspector, assets.RGB{R: 34, G: 36, B: 38})
	c.strokeUIRect(179, 76, 1, 633, deskLine)
	c.strokeUIRect(1020, 76, 1, 633, deskLine)
	deskLabel(c, d.Fonts, word("headquarters"), l.Rect{20, 96, 150, 24}, deskMuted, 11)
	deskLabel(c, d.Fonts, word("progress"), l.Rect{20, 421, 150, 24}, deskMuted, 11)
	roomParagraph(c, d.Fonts, word("mouse.help"), l.Rect{20, 575, 150, 100}, deskMuted)
	deskFill(c, l.Rect{180, 76, 840, 64}, assets.RGB{R: 36, G: 37, B: 40})
	c.strokeUIRect(180, 139, 840, 1, deskLine)
	deskLabel(c, d.Fonts, word("breadcrumb"), l.Rect{276, 87, 170, 42}, deskMuted, 13)
	deskLabel(c, d.Fonts, d.Title, l.Rect{446, 87, 420, 42}, deskInk, 13)
	deskFill(c, l.DeskActions, assets.RGB{R: 37, G: 37, B: 40})
	c.strokeUIRect(180, 642, 840, 1, deskLine)
	deskFill(c, l.DeskFooter, assets.RGB{R: 32, G: 33, B: 35})
	c.strokeUIRect(0, 709, 1280, 1, deskLine)
	deskLabel(c, d.Fonts, word("footer"), l.Rect{18, 714, 850, 22}, deskMuted, 12)
	deskLabel(c, d.Fonts, word("selected"), l.Rect{1038, 95, 224, 24}, deskGold, 15)
	deskLabel(c, d.Fonts, d.Province, l.Rect{1038, 130, 224, 40}, deskInk, 26)
	deskLabel(c, d.Fonts, d.Commander, l.Rect{1038, 185, 224, 25}, deskMuted, 15)
	for i, fact := range []struct{ Label, Value string }{{word("force"), d.Force}, {word("arsenal"), d.Arsenals}, {word("coal"), d.Coal}, {word("iron"), d.Iron}} {
		x, y := 1038+(i%2)*112, 220+(i/2)*56
		deskLabel(c, d.Fonts, fact.Label, l.Rect{x, y, 108, 20}, deskMuted, 12)
		deskLabel(c, d.Fonts, fact.Value, l.Rect{x, y + 20, 108, 25}, deskInk, 16)
	}
	c.strokeUIRect(1038, 334, 224, 1, deskLine)
	deskLabel(c, d.Fonts, word("stationed"), l.Rect{1038, 350, 224, 24}, deskGold, 15)
	for i, u := range d.Units {
		if i >= 4 {
			break
		}
		y := 386 + i*72
		deskLabel(c, d.Fonts, u.Name, l.Rect{1038, y, 120, 24}, deskInk, 15)
		deskLabel(c, d.Fonts, u.Force, l.Rect{1160, y, 102, 24}, deskInk, 14)
		deskLabel(c, d.Fonts, u.Detail, l.Rect{1038, y + 25, 224, 24}, deskMuted, 11)
		c.strokeUIRect(1038, y+61, 224, 1, deskLine)
	}
	body := NewCanvas(l.DeskWidth, l.DeskHeight)
	body.UseLightUI(d.LightUI)
	deskFill(body, l.DeskContent, deskBackground)
	deskLabel(body, d.Fonts, d.Title, l.Rect{200, 160, 650, 34}, deskInk, 24)
	deskLabel(body, d.Fonts, d.Prompt, l.Rect{200, 198, 800, 26}, deskMuted, 13)
	if d.Map != nil {
		r := l.DeskMap
		r.Y -= d.Scroll
		for y := 0; y < r.H; y++ {
			for x := 0; x < r.W; x++ {
				body.img.Set(r.X+x, r.Y+y, d.Map.At(x*d.Map.Bounds().Dx()/r.W, y*d.Map.Bounds().Dy()/r.H))
			}
		}
		body.strokeUIRect(r.X, r.Y, r.W, r.H, deskLine)
	}
	for _, r := range d.Cards {
		r.Y -= d.Scroll
		deskGradient(body, r, assets.RGB{R: 48, G: 46, B: 43}, assets.RGB{R: 38, G: 39, B: 42})
		body.strokeUIRect(r.X, r.Y, r.W, r.H, deskLine)
	}
	for _, v := range d.Texts {
		r := v.Rect
		r.Y -= d.Scroll
		ink := deskInk
		if v.Muted {
			ink = deskMuted
		}
		if v.Gold {
			ink = deskGold
		}
		if v.Paragraph {
			roomParagraph(body, d.Fonts, v.Text, r, ink)
		} else {
			deskLabel(body, d.Fonts, v.Text, r, ink, max(11, v.Size))
		}
	}
	for _, v := range d.Controls {
		if v.Body {
			v.Rect.Y -= d.Scroll
			illustrated := false
			for _, im := range d.Illustrations {
				r := im.Rect
				r.Y -= d.Scroll
				if r != v.Rect {
					continue
				}
				illustrated = true
				bounds := im.Image.Bounds()
				for y := 0; y < r.H; y++ {
					for x := 0; x < r.W; x++ {
						body.img.Set(r.X+x, r.Y+y, im.Image.At(bounds.Min.X+x*bounds.Dx()/r.W, bounds.Min.Y+y*bounds.Dy()/r.H))
					}
				}
				body.strokeUIRect(r.X, r.Y, r.W, r.H, deskGold)
				if v.Key == d.Focus {
					body.strokeUIRect(r.X+2, r.Y+2, r.W-4, r.H-4, deskAccent)
				}
				deskFill(body, l.Rect{r.X + 1, r.Bottom() - 43, r.W - 2, 42}, deskPanel)
				deskLabel(body, d.Fonts, v.Text, l.Rect{r.X + 12, r.Bottom() - 43, r.W - 24, 42}, deskInk, 16)
			}
			if !illustrated {
				deskButton(body, d.Fonts, v, d.Focus)
			}
		}
	}
	clip := image.Rect(l.DeskContent.X, l.DeskContent.Y, l.DeskContent.Right(), l.DeskContent.Bottom())
	draw.Draw(c.img, clip, body.img, clip.Min, draw.Over)
	if d.Portrait != nil {
		drawPortraitView(c, d.Fonts, *d.Portrait, l.Rect{X: 800, Y: 244, W: 184, H: 350})
	}
	if d.Status != "" {
		deskFill(c, l.Rect{200, 612, 800, 26}, deskBackground)
		deskLabel(c, d.Fonts, d.Status, l.Rect{206, 612, 788, 26}, deskGold, 12)
	}
	for _, v := range d.Controls {
		if !v.Body {
			deskButton(c, d.Fonts, v, d.Focus)
		}
	}
	if d.ContentHeight > l.DeskContent.H {
		h := max(25, l.DeskContent.H*l.DeskContent.H/d.ContentHeight)
		y := l.DeskContent.Y + d.Scroll*(l.DeskContent.H-h)/max(1, d.ContentHeight-l.DeskContent.H)
		deskFill(c, l.Rect{1014, y, 4, h}, deskLine)
	}
	if d.Dialog != nil {
		for y := 0; y < l.DeskHeight; y++ {
			for x := 0; x < l.DeskWidth; x++ {
				p := c.img.RGBAAt(x, y)
				p.R /= 3
				p.G /= 3
				p.B /= 3
				c.img.SetRGBA(x, y, p)
			}
		}
		r := l.DeskDialog
		deskFill(c, r, deskPanel)
		c.strokeUIRect(r.X, r.Y, r.W, r.H, deskGold)
		deskFill(c, l.Rect{r.X, r.Y, r.W, 3}, deskAccent)
		deskLabel(c, d.Fonts, d.Dialog.Title, l.Rect{r.X + 25, r.Y + 20, r.W - 50, 40}, deskInk, 24)
		roomParagraph(c, d.Fonts, d.Dialog.Text, l.Rect{r.X + 25, r.Y + 72, r.W - 50, 94}, deskMuted)
		for _, v := range DeskTargets(d) {
			if v.Key == "dialog.yes" {
				v.Kind = "primary"
			}
			deskButton(c, d.Fonts, v, d.Focus)
		}
	}
	return c
}

func (c *Canvas) DrawBattleReturn(fonts *assets.EtenFonts, high bool, text string) {
	deskButton(c, fonts, DeskControl{Key: "battle.back", Text: text, Rect: l.BattleReturnButton(high)}, "")
}

func (c *Canvas) DrawDeskToolbar(fonts *assets.EtenFonts, v ToolbarView) {
	draw.Draw(c.img, c.Bounds(), image.NewUniform(color.RGBA{}), image.Point{}, draw.Src)
	deskFill(c, l.Rect{W: l.DeskWidth, H: l.DeskToolbarHeight}, assets.RGB{R: 41, G: 38, B: 36})
	c.strokeUIRect(0, 57, 1280, 1, deskGold)
	for i := 0; i < l.ToolbarFields; i++ {
		r := l.DeskToolbarField(i)
		label := l.Rect{r.X - 35, 11, 32, 36}
		if i == 5 {
			label = l.Rect{r.X - 91, 11, 84, 36}
		}
		deskLabel(c, fonts, v.Labels[i], label, deskMuted, 13)
		deskButton(c, fonts, DeskControl{Key: strconv.Itoa(i), Text: v.Values[i] + "  v", Rect: r, Selected: v.Menu == i}, "")
	}
	if v.Menu < 0 || v.Menu >= l.ToolbarFields {
		return
	}
	r := v.MenuRect()
	deskFill(c, r, deskPanel)
	c.strokeUIRect(r.X, r.Y, r.W, r.H, deskGold)
	for i, s := range v.Options[v.Menu] {
		selected := i == v.Selected[v.Menu]
		if len(v.Checked[v.Menu]) == len(v.Options[v.Menu]) {
			selected = v.Checked[v.Menu][i]
		}
		if selected {
			s = "* " + s
		}
		row := l.Rect{r.X + 5, r.Y + 5 + i*l.ToolbarRow, r.W - 10, l.ToolbarRow}
		if v.Hover == i {
			deskFill(c, row, assets.RGB{R: 88, G: 70, B: 51})
		}
		deskLabel(c, fonts, s, l.Rect{row.X + 10, row.Y, row.W - 20, row.H}, deskInk, 14)
	}
	if len(v.Notes) > 0 {
		y := r.Y + 10 + len(v.Options[v.Menu])*l.ToolbarRow
		roomParagraph(c, fonts, strings.Join(v.Notes, "\n"), l.Rect{r.X + 12, y, r.W - 24, r.Bottom() - y - 6}, deskMuted)
	}
}
