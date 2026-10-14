package render

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"strings"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	"github.com/wicanr2/great-era-remake/internal/ui/hdart"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
	"github.com/wicanr2/great-era-remake/internal/ui/textlayout"
	t "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

type RoomMarker struct {
	ID, X, Y, Faction, Force int
	Name                     string
	Selected                 bool
}
type RoomArchiveData struct {
	People                              []*i18n.Person
	Selected                            *i18n.Person
	Query                               string
	ListPage, Page                      int
	Searching                           bool
	Unavailable, Sources, Title, Search string
	Portrait                            *i18n.Portrait
	PortraitUnavailable                 string
}
type CommandRoomData struct {
	Map                                  *image.RGBA
	Fonts                                *assets.EtenFonts
	Style                                t.UIStyle
	Panel                                PanelData
	Title, Date, Interface, Art, Message string
	Tabs, Commands, Tools                []string
	Tab                                  int
	CommandScroll                        int
	Portrait                             *i18n.Portrait
	PortraitUnavailable                  string
	Markers                              []RoomMarker
	Labels                               map[string]string
	Archive                              *RoomArchiveData
	OriginalUnits                        t.UnitProvider
}

func roomRect(c *Canvas, r l.Rect, fill, border assets.RGB) {
	kind := 0
	if fill.R > 150 && fill.G > 140 {
		kind = 2
	}
	if earthFrame(c, r, kind) {
		return
	}
	material, strength := hdart.Steel, .32
	if fill.R > 150 && fill.G > 140 {
		material, strength = hdart.Paper, .35
	}
	roomSurface(c, r, material, fill, strength, border)
}
func roomText(c *Canvas, fonts *assets.EtenFonts, s string, r l.Rect, ink assets.RGB, scale int) {
	if r.H < 15*scale {
		return
	}
	s = modernEllipsize(s, r.W, scale)
	// Width has already been measured in actual proportional glyph advances.
	// A second half-cell limit would silently remove digits and biography text.
	drawModernTextScaled(c, fonts, s, ink, r.X, r.Y+(r.H-15*scale)/2, scale, 100000)
}
func roomButton(c *Canvas, fonts *assets.EtenFonts, s string, r l.Rect, style t.UIStyle, active bool) {
	border, ink := style.Muted, style.Ink
	if active {
		border, ink = style.Accent, style.Accent
	}
	kind := 1
	if active {
		kind = 3
	}
	if !earthFrame(c, r, kind) {
		roomRect(c, r, style.Panel, border)
	}
	size := 18
	if r.W < 110 {
		size = 14
	}
	if r.W < 40 {
		size = 18
	}
	pad := 8
	if r.W < 40 {
		pad = 3
	}
	if r.W >= 80 && r.W < 110 {
		size = 12
	}
	roomLabel(c, fonts, s, l.Rect{r.X + pad, r.Y, r.W - pad*2, r.H}, ink, size, false)
}
func roomParagraph(c *Canvas, fonts *assets.EtenFonts, s string, r l.Rect, ink assets.RGB) {
	doc, err := textlayout.LayoutProportional(s, textlayout.ProportionalOptions{Width: r.W, Rows: max(1, r.H/20), Advance: modernBiographyAdvance})
	if err != nil || len(doc.Pages) == 0 {
		return
	}
	for i, line := range doc.Pages[0].Lines {
		roomText(c, fonts, line.Text, l.Rect{r.X, r.Y + i*20, r.W, 20}, ink, 1)
	}
}
func (c *Canvas) DrawCommandMessage(fonts *assets.EtenFonts, style t.UIStyle, message string) {
	roomRect(c, l.Rect{386, 550, 680, 124}, style.Paper, style.Muted)
	roomParagraph(c, fonts, message, l.Rect{398, 560, 656, 108}, style.Ink)
}
func RoomBiographyDocument(p *i18n.Person, unavailable, sources string) (textlayout.Document, error) {
	body := p.Biography
	if strings.TrimSpace(body) == "" {
		body = unavailable
	}
	if len(p.Sources) > 0 {
		body += "\n\n" + sources + "\n" + strings.Join(p.Sources, "\n")
	}
	return textlayout.LayoutProportional(body, textlayout.ProportionalOptions{Width: 508, Rows: 15, Advance: modernBiographyAdvance})
}

func RoomArchiveDocument(a *RoomArchiveData) (textlayout.Document, error) {
	p := *a.Selected
	if a.Portrait != nil {
		p.Sources = append(append([]string{}, p.Sources...), a.Portrait.Attribution, a.Portrait.SourcePage, a.Portrait.LicenseURL)
	}
	return RoomBiographyDocument(&p, a.Unavailable, a.Sources)
}

// Rendering and hit testing share the same visible markers. Hidden, overlapping
// counters must not intercept a click on a different province.
func VisibleRoomMarkers(markers []RoomMarker) []RoomMarker {
	out := make([]RoomMarker, 0, len(markers))
	for _, selected := range []bool{true, false} {
		for _, m := range markers {
			if m.Selected != selected || !l.RoomMarkerVisible(m.X, m.Y) {
				continue
			}
			overlap := false
			for _, p := range out {
				if absIntLocal(p.X-m.X) < 72 && absIntLocal(p.Y-m.Y) < 50 {
					overlap = true
					break
				}
			}
			if !overlap {
				out = append(out, m)
			}
		}
	}
	return out
}
func (c *Canvas) DrawCommandRoom(d CommandRoomData) error {
	if c.Bounds() != image.Rect(0, 0, 1280, 720) {
		return fmt.Errorf("command room requires 1280x720")
	}
	s := d.Style
	label := func(k string) string { return d.Labels[k] }
	c.fillRect(0, 0, 1280, 720, s.Paper)
	if d.Map != nil {
		draw.Draw(c.img, image.Rect(0, l.CommandRoom.Map.Y, 1280, 720), d.Map, image.Point{}, draw.Src)
	}
	var originalUnit *image.RGBA
	if d.OriginalUnits != nil {
		b, err := d.OriginalUnits.Unit(0)
		if err != nil {
			return err
		}
		if b.Valid() && b.Image.W <= 64 && b.Image.H <= 64 {
			originalUnit = image.NewRGBA(image.Rect(0, 0, b.Image.W, b.Image.H))
			for y := 0; y < b.Image.H; y++ {
				for x := 0; x < b.Image.W; x++ {
					v := b.Image.Pix[y*b.Image.W+x]
					if v != 0 {
						p := b.Palette[v]
						originalUnit.SetRGBA(x, y, color.RGBA{p.R, p.G, p.B, 255})
					}
				}
			}
		}
	}
	// Live provinces and forces are independent of the relief artwork.
	for _, m := range VisibleRoomMarkers(d.Markers) {
		fg := s.Ink
		if m.Selected {
			fg = s.Accent
		}
		name := strings.TrimSuffix(m.Name, "省")
		r := l.Rect{m.X - 45, m.Y - 25, 100, 23}
		roomLabel(c, d.Fonts, name, l.Rect{r.X + 1, r.Y + 1, r.W, r.H}, assets.RGB{R: 18, G: 22, B: 19}, 19, true)
		roomLabel(c, d.Fonts, name, r, fg, 19, true)
		c.line(m.X-3, m.Y, m.X+3, m.Y, fg)
		c.line(m.X, m.Y-3, m.X, m.Y+3, fg)
		if m.Force <= 0 {
			continue
		}
		if originalUnit != nil {
			p := image.Pt(m.X-originalUnit.Bounds().Dx()/2, m.Y+7)
			draw.Draw(c.img, originalUnit.Bounds().Add(p), originalUnit, image.Point{}, draw.Over)
			continue
		}
		r = l.Rect{m.X - 40, m.Y + 7, 80, 25}
		border := s.Muted
		if m.Selected {
			border = s.Accent
		}
		roomSurface(c, r, hdart.Steel, assets.RGB{R: 35, G: 29, B: 24}, .3, border)
		tint := s.FactionTint[m.Faction%10]
		c.fillRect(r.X+3, r.Y+4, 15, 16, tint)
		c.line(r.X+5, r.Y+7, r.X+15, r.Y+17, fg)
		c.line(r.X+15, r.Y+7, r.X+5, r.Y+17, fg)
		roomLabel(c, d.Fonts, roomForce(m.Force), l.Rect{r.X + 22, r.Y + 1, 56, 23}, fg, 13, false)
	}
	roomRect(c, l.CommandRoom.Header, s.Panel, s.Muted)
	roomLabel(c, d.Fonts, d.Title, l.Rect{12, 4, 204, 36}, s.Accent, 28, true)
	resources := []struct {
		label string
		value int
	}{
		{d.Panel.Labels.Gold, int(d.Panel.Province.Gold)}, {d.Panel.Labels.Food, int(d.Panel.Province.Food)},
		{d.Panel.Labels.Ammo, int(d.Panel.Province.Ammo)}, {d.Panel.Labels.Coal, int(d.Panel.Province.Coal)},
		{d.Panel.Labels.Fuel, int(d.Panel.Province.Fuel)}, {d.Panel.Labels.Iron, int(d.Panel.Province.Iron)},
	}
	for i, r := range resources {
		x := 222 + i*104
		c.line(x-5, 6, x-5, 39, assets.RGB{R: 66, G: 65, B: 57})
		roomIcon(c, []int{0, 1, 2, 4, 3, 5}[i], l.Rect{x, 12, 25, 25}, true)
		roomLabel(c, d.Fonts, r.label, l.Rect{x + 30, 3, 68, 17}, s.Muted, 12, false)
		roomLabel(c, d.Fonts, fmt.Sprint(r.value), l.Rect{x + 30, 20, 68, 22}, s.Ink, 18, false)
	}
	roomLabel(c, d.Fonts, d.Date, l.Rect{858, 8, 166, 30}, s.Ink, 22, false)
	roomButton(c, d.Fonts, d.Interface, l.RoomInterface, s, false)
	roomButton(c, d.Fonts, d.Art, l.RoomArt, s, false)
	roomRect(c, l.CommandRoom.Toolbar, s.Paper, s.Muted)
	tabIcons := []int{0, 1, 6, 8, 11, 14}
	for i, v := range d.Tabs {
		r := l.RoomTab(i)
		roomButton(c, d.Fonts, "", r, s, i == d.Tab)
		roomIcon(c, tabIcons[i%6], l.Rect{r.X + 7, r.Y + 6, 26, 25}, false)
		fg := s.Ink
		if i == d.Tab {
			fg = s.Accent
		}
		roomLabel(c, d.Fonts, v, l.Rect{r.X + 40, r.Y, r.W - 46, r.H}, fg, 18, false)
	}
	roomRect(c, l.CommandRoom.Sidebar, s.Paper, s.Muted)
	roomRivets(c, l.CommandRoom.Sidebar, s)
	roomRect(c, l.Rect{14, 112, 342, 58}, s.Panel, s.Muted)
	if d.Map != nil {
		var src image.Rectangle
		for _, m := range d.Markers {
			if m.Selected {
				src = image.Rect(m.X-90, m.Y-l.CommandRoom.Map.Y-55, m.X+90, m.Y-l.CommandRoom.Map.Y+55).Intersect(d.Map.Bounds())
			}
		}
		if src.Empty() {
			src = d.Map.Bounds()
		}
		drawSampledImage(c.img, l.Rect{20, 118, 86, 46}, d.Map, src)
	}
	roomLabel(c, d.Fonts, d.Panel.ProvinceName, l.Rect{118, 118, 177, 42}, s.Ink, 30, true)
	roomButton(c, d.Fonts, "<", l.RoomPrevious, s, false)
	roomButton(c, d.Fonts, ">", l.RoomNext, s, false)
	roomRect(c, l.Rect{14, 174, 342, 138}, s.Panel, s.Muted)
	roomPortrait(c, d.Portrait, d.PortraitUnavailable, l.Rect{20, 180, 94, 124}, s)
	roomButton(c, d.Fonts, d.Panel.CommanderName, l.RoomCommander, s, false)
	metrics := []string{
		fmt.Sprintf("%s  %d", d.Panel.Labels.Force, d.Panel.Force),
		fmt.Sprintf("%s  %d", d.Panel.Labels.Loyalty, d.Panel.Province.Loyalty),
		fmt.Sprintf("%s  %d", d.Panel.Labels.Commands, d.Panel.Commands),
	}
	for i, v := range metrics {
		roomIcon(c, []int{1, 11, 0}[i], l.Rect{128, 216 + i*28, 22, 22}, false)
		roomLabel(c, d.Fonts, v, l.Rect{158, 212 + i*28, 186, 27}, s.Ink, 18, false)
	}
	for row := 0; row < l.RoomVisibleCommands; row++ {
		i := d.CommandScroll + row
		if i >= len(d.Commands) {
			break
		}
		r := l.RoomCommand(row)
		active := d.Tab == 1 && (i == 1 || i == 2 || i == 4 || i == 12 || i == 13)
		roomButton(c, d.Fonts, "", r, s, active)
		roomIcon(c, i, l.Rect{r.X + 9, r.Y + 4, 29, 27}, false)
		fg := s.Ink
		if active {
			fg = s.Accent
		}
		roomLabel(c, d.Fonts, d.Commands[i], l.Rect{r.X + 54, r.Y, r.W - 80, r.H}, fg, 19, false)
		roomLabel(c, d.Fonts, ">", l.Rect{r.Right() - 22, r.Y, 18, r.H}, s.Muted, 17, false)
	}
	c.fillRect(348, 346, 5, 252, assets.RGB{R: 10, G: 11, B: 10})
	y := 346 + d.CommandScroll*116/max(1, len(d.Commands)-l.RoomVisibleCommands)
	roomRect(c, l.Rect{346, y, 9, 134}, s.Muted, s.Muted)
	roomButton(c, d.Fonts, "", l.RoomCommandScrollUp, s, false)
	c.line(348, 335, 351, 331, s.Muted)
	c.line(351, 331, 354, 335, s.Muted)
	roomButton(c, d.Fonts, "", l.RoomCommandScrollDown, s, false)
	c.line(348, 612, 351, 616, s.Muted)
	c.line(351, 616, 354, 612, s.Muted)
	for i, key := range []string{"people", "save", "end_turn"} {
		r := l.RoomBottomActions[i]
		roomButton(c, d.Fonts, "", r, s, false)
		roomIcon(c, []int{11, 0, 14}[i], l.Rect{r.X + 7, r.Y + 8, 25, 25}, false)
		roomLabel(c, d.Fonts, label(key), l.Rect{r.X + 38, r.Y, r.W - 42, r.H}, s.Ink, 15, false)
	}
	roomRect(c, l.Rect{1078, 586, 198, 110}, s.Paper, s.Muted)
	roomRivets(c, l.Rect{1078, 586, 198, 110}, s)
	if d.Map != nil {
		drawSampledImage(c.img, l.RoomMiniMap, d.Map, d.Map.Bounds())
	}
	for i, v := range d.Tools {
		roomButton(c, d.Fonts, v, l.RoomMapTools[i], s, false)
	}
	if d.Message != "" {
		roomRect(c, l.CommandRoom.Footer, s.Paper, s.Muted)
		roomLabel(c, d.Fonts, d.Message, l.CommandRoom.Footer, s.Ink, 13, false)
	}
	if d.Archive != nil {
		return c.drawRoomArchive(d)
	}
	return nil
}
func (c *Canvas) drawRoomArchive(d CommandRoomData) error {
	a, s := d.Archive, d.Style
	roomRect(c, l.RoomArchive, s.Paper, s.Muted)
	roomText(c, d.Fonts, fmt.Sprintf("%s · %d", a.Title, len(a.People)), l.Rect{310, 132, 620, 36}, s.Ink, 2)
	roomButton(c, d.Fonts, d.Labels["close"], l.RoomArchiveBack, s, false)
	query := a.Query
	if query == "" {
		query = a.Search
	}
	if a.Searching {
		query += "|"
	}
	roomButton(c, d.Fonts, query, l.RoomSearch, s, a.Searching)
	start := a.ListPage * 11
	for i := 0; i < 11 && start+i < len(a.People); i++ {
		p := a.People[start+i]
		roomButton(c, d.Fonts, p.NameInGame, l.RoomPerson(i), s, a.Selected != nil && p.ID == a.Selected.ID)
	}
	roomButton(c, d.Fonts, "<", l.RoomArchiveListPrevious, s, false)
	roomButton(c, d.Fonts, ">", l.RoomArchiveListNext, s, false)
	roomButton(c, d.Fonts, "<", l.RoomArchivePrevious, s, false)
	roomButton(c, d.Fonts, ">", l.RoomArchiveNext, s, false)
	if a.Selected == nil {
		return nil
	}
	p := a.Selected
	paper, ink := assets.RGB{R: 224, G: 219, B: 199}, assets.RGB{R: 52, G: 56, B: 46}
	roomRect(c, l.Rect{502, 178, 552, 448}, paper, paper)
	roomLabel(c, d.Fonts, p.NameInGame, l.Rect{522, 184, 440, 38}, ink, 30, true)
	roomPortrait(c, a.Portrait, a.PortraitUnavailable, l.Rect{982, 184, 64, 72}, s)
	roomText(c, d.Fonts, p.Faction, l.Rect{522, 228, 450, 24}, ink, 1)
	doc, err := RoomArchiveDocument(a)
	if err != nil {
		return err
	}
	page := max(0, min(a.Page, len(doc.Pages)-1))
	for i, line := range doc.Pages[page].Lines {
		roomText(c, d.Fonts, line.Text, l.Rect{522, 260 + i*23, 512, 22}, ink, 1)
	}
	roomText(c, d.Fonts, fmt.Sprintf("%d / %d", page+1, len(doc.Pages)), l.Rect{650, 638, 280, 26}, s.Ink, 1)
	return nil
}
func absIntLocal(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

// Keep image/color available for the bounded line primitive below.
func (c *Canvas) line(x0, y0, x1, y1 int, ink assets.RGB) {
	dx, dy := absIntLocal(x1-x0), absIntLocal(y1-y0)
	steps := max(dx, dy)
	for i := 0; i <= steps; i++ {
		x, y := x0+(x1-x0)*i/max(1, steps), y0+(y1-y0)*i/max(1, steps)
		c.img.SetRGBA(x, y, color.RGBA{ink.R, ink.G, ink.B, 255})
	}
}
