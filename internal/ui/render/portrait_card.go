package render

import (
	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
	t "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

// PortraitView contains only the verified catalog result for this person.
// Missing records deliberately remain an archive card, never another face.
type PortraitView struct {
	Portrait    *i18n.Portrait
	Unavailable string
}

func drawPortraitView(c *Canvas, fonts *assets.EtenFonts, v PortraitView, r l.Rect) {
	photo := l.Rect{X: r.X, Y: r.Y, W: r.W, H: r.H - 24}
	roomPortrait(c, v.Portrait, v.Unavailable, photo, t.UIStyle{Muted: deskLine})
	caption := v.Unavailable
	if v.Portrait != nil {
		caption = v.Portrait.Attribution
	}
	roomLabel(c, fonts, caption, l.Rect{X: r.X + 2, Y: r.Bottom() - 22, W: r.W - 4, H: 22}, deskMuted, 11, false)
}
