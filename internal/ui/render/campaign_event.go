package render

import (
	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
	t "github.com/wicanr2/great-era-remake/internal/ui/theme"
	"image"
)

type CampaignEventData struct {
	Fonts                                                             *assets.EtenFonts
	Style                                                             t.UIStyle
	Title, Date, Intro, RosterTitle, Status, Continue, Menu, ExitNote string
	Names                                                             []string
	Scene                                                             *image.RGBA
	Ending, Quit, Light                                               bool
	QuitTitle, QuitYes, QuitNo                                        string
}

func CampaignEventTargets(w, h int, ending, quit bool) []actions.Target {
	height, width := 38, 180
	if w < 1000 {
		height, width = 28, 134
	}
	y := h - height - 16
	if quit {
		return []actions.Target{
			{Action: actions.Confirm, Rect: actions.Rect{w/2 - width - 8, h/2 + 24, width, height}},
			{Action: actions.Cancel, Rect: actions.Rect{w/2 + 8, h/2 + 24, width, height}},
		}
	}
	out := []actions.Target{{Action: actions.Confirm, Rect: actions.Rect{w - width - 20, y, width, height}}}
	if ending {
		out = append(out, actions.Target{Action: actions.Back, Rect: actions.Rect{20, y, width, height}})
	}
	return out
}

func DrawCampaignEvent(w, h int, d CampaignEventData) *Canvas {
	c := NewCanvas(w, h)
	c.UseLightUI(d.Light)
	style := d.Style
	roomRect(c, l.Rect{0, 0, w, h}, style.Paper, style.Muted)
	pad, titleHeight := 24, 60
	if w < 1000 {
		pad, titleHeight = 12, 32
	}
	roomRect(c, l.Rect{pad, pad, w - pad*2, titleHeight}, style.Panel, style.Accent)
	titleScale := 2
	if w < 1000 {
		titleScale = 1
	}
	roomText(c, d.Fonts, d.Title, l.Rect{pad + 12, pad + 4, w - pad*2 - 24, titleHeight - 8}, style.Ink, titleScale)
	top := pad + titleHeight + 12
	roomText(c, d.Fonts, d.Date, l.Rect{pad, top, w - pad*2, 18}, style.Ink, 1)
	top += 26
	bottom := h - 78
	if w < 1000 {
		bottom = h - 58
	}
	leftW := (w - pad*3) * 2 / 3
	rightX := pad*2 + leftW
	rightW := w - pad - rightX
	left := l.Rect{pad, top, leftW, bottom - top}
	right := l.Rect{rightX, top, rightW, bottom - top}
	roomRect(c, left, style.Panel, style.Muted)
	roomRect(c, right, style.Panel, style.Muted)
	if d.Scene != nil {
		box := l.Rect{left.X + 6, left.Y + 6, left.W - 12, left.H - 54}
		if w < 1000 {
			box.H = left.H - 64
		}
		// Fit the entire approved composition. No crop or tank relocation.
		src := d.Scene.Bounds()
		pw, ph := box.W, box.W*src.Dy()/src.Dx()
		if ph > box.H {
			ph = box.H
			pw = ph * src.Dx() / src.Dy()
		}
		x0, y0 := box.X+(box.W-pw)/2, box.Y+(box.H-ph)/2
		for y := 0; y < ph; y++ {
			for x := 0; x < pw; x++ {
				c.img.SetRGBA(x0+x, y0+y, d.Scene.RGBAAt(src.Min.X+x*src.Dx()/pw, src.Min.Y+y*src.Dy()/ph))
			}
		}
		captionHeight := 36
		if w < 1000 {
			captionHeight = 54
		}
		roomParagraph(c, d.Fonts, d.Intro, l.Rect{left.X + 12, left.Bottom() - captionHeight - 4, left.W - 24, captionHeight}, style.Ink)
	} else {
		roomParagraph(c, d.Fonts, d.Intro, l.Rect{left.X + 12, left.Y + left.H/2 - 24, left.W - 24, 64}, style.Ink)
	}
	roomText(c, d.Fonts, d.RosterTitle, l.Rect{right.X + 10, right.Y + 6, right.W - 20, 20}, style.Accent, 1)
	rowH := min(34, (right.H-28)/max(1, len(d.Names)))
	for i, name := range d.Names {
		roomText(c, d.Fonts, name, l.Rect{right.X + 10, right.Y + 26 + i*rowH, right.W - 20, rowH}, style.Ink, 1)
	}
	roomText(c, d.Fonts, d.Status, l.Rect{pad, bottom + 5, w - pad*2, 18}, style.Ink, 1)
	if w >= 1000 {
		roomText(c, d.Fonts, d.ExitNote, l.Rect{pad, bottom + 24, w - pad*2, 16}, style.Muted, 1)
	}
	for _, target := range CampaignEventTargets(w, h, d.Ending, false) {
		label := d.Continue
		if target.Action == actions.Back {
			label = d.Menu
		}
		roomButton(c, d.Fonts, label, l.Rect{target.Rect.X, target.Rect.Y, target.Rect.W, target.Rect.H}, style, target.Action == actions.Confirm)
	}
	if d.Quit {
		r := l.Rect{w/2 - min(300, w/2-20), h/2 - 70, min(600, w-40), 140}
		roomRect(c, r, style.Panel, style.Accent)
		roomText(c, d.Fonts, d.QuitTitle, l.Rect{r.X + 12, r.Y + 10, r.W - 24, 40}, style.Ink, 1)
		for _, target := range CampaignEventTargets(w, h, d.Ending, true) {
			label := d.QuitYes
			if target.Action == actions.Cancel {
				label = d.QuitNo
			}
			roomButton(c, d.Fonts, label, l.Rect{target.Rect.X, target.Rect.Y, target.Rect.W, target.Rect.H}, style, target.Action == actions.Confirm)
		}
	}
	return c
}
