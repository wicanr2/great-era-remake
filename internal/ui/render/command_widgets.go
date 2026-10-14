package render

import (
	"fmt"
	"image"
	"math"
	"strings"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	"github.com/wicanr2/great-era-remake/internal/ui/hdart"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
	"github.com/wicanr2/great-era-remake/internal/ui/textlayout"
	t "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

var roomIcons = t.NewCommandModern()

func roomLabelWidth(f *assets.ModernFont, text string, size int) int {
	n := 0
	for _, ch := range text {
		if g, ok := f.Glyph(ch); ok {
			n += max(1, g.PixelAdvance()*size/32)
		} else {
			n += size
		}
	}
	return n
}

// Controls fit their existing rectangles. Long translations first use a smaller
// readable size, then wrap when the control has room for a second line.
func roomLabelPlan(f *assets.ModernFont, text string, r l.Rect, size int) ([]string, int) {
	for size > 12 && roomLabelWidth(f, text, size) > r.W {
		size--
	}
	if roomLabelWidth(f, text, size) <= r.W {
		return []string{text}, size
	}
	if r.H >= 2*(size+2) {
		if at := strings.IndexAny(text, ":："); at >= 0 {
			v := []rune(text[at:])
			split := at + len(string(v[0]))
			left, right := strings.TrimSpace(text[:split]), strings.TrimSpace(text[split:])
			if left != "" && right != "" && roomLabelWidth(f, left, size) <= r.W && roomLabelWidth(f, right, size) <= r.W {
				return []string{left, right}, size
			}
		}
		doc, err := textlayout.LayoutProportional(text, textlayout.ProportionalOptions{Width: r.W, Rows: r.H / (size + 2), Advance: func(ch rune) int { return roomLabelWidth(f, string(ch), size) }})
		if err == nil && len(doc.Pages) == 1 {
			lines := []string{}
			for _, line := range doc.Pages[0].Lines {
				lines = append(lines, strings.TrimSpace(line.Text))
			}
			return lines, size
		}
	}
	v := []rune(text)
	for len(v) > 0 && roomLabelWidth(f, string(v)+"...", size) > r.W {
		v = v[:len(v)-1]
	}
	return []string{string(v) + "..."}, size
}

// Text uses the embedded font's native advances. Sizes between 16 and 32 do
// not enlarge the old bitmap font and do not truncate resource numbers.
func roomLabel(c *Canvas, fonts *assets.EtenFonts, text string, r l.Rect, ink assets.RGB, size int, serif bool) {
	family := assets.ModernFontSans
	if serif {
		family = assets.ModernFontSerif
	}
	f, err := assets.EmbeddedModernFontFamily(family)
	if err != nil {
		roomText(c, fonts, text, r, ink, 1)
		return
	}
	lines, size := roomLabelPlan(f, text, r, size)
	lineHeight := size + 2
	y := r.Y + (r.H-(len(lines)*lineHeight-2))/2
	for _, line := range lines {
		x := r.X
		for _, ch := range line {
			g, ok := f.Glyph(ch)
			if !ok {
				// Missing characters remain visible; coverage is checked for all locales.
				c.strokeRect(x, y, max(1, size-2), max(1, size-2), ink)
				x += size
				continue
			}
			sw, sh := g.Size()
			for yy := 0; yy < size; yy++ {
				for xx := 0; xx < size; xx++ {
					coverage := modernGlyphCoverage(g, xx, yy, size, size, sw, sh)
					if coverage > 0 && x+xx < r.Right() && y+yy < r.Bottom() {
						c.blendPixel(x+xx, y+yy, ink, coverage)
					}
				}
			}
			x += max(1, g.PixelAdvance()*size/32)
		}
		y += lineHeight
	}
}

func roomSurface(c *Canvas, r l.Rect, material int, tint assets.RGB, strength float64, edge assets.RGB) {
	bounds := image.Rect(r.X+1, r.Y+1, r.Right()-1, r.Bottom()-1).Intersect(c.Bounds())
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		for x := bounds.Min.X; x < bounds.Max.X; x++ {
			p := hdart.Sample(material, x, y)
			shade := 1.05 - float64(y-r.Y)/float64(max(1, r.H))*.12
			rgb := assets.RGB{R: byte(math.Min(255, (float64(tint.R)*(1-strength)+float64(p.R)*strength)*shade)), G: byte(math.Min(255, (float64(tint.G)*(1-strength)+float64(p.G)*strength)*shade)), B: byte(math.Min(255, (float64(tint.B)*(1-strength)+float64(p.B)*strength)*shade))}
			c.blendPixel(x, y, rgb, 255)
		}
	}
	c.line(r.X, r.Y, r.Right()-1, r.Y, edge)
	c.line(r.X, r.Y, r.X, r.Bottom()-1, edge)
	c.line(r.Right()-1, r.Y, r.Right()-1, r.Bottom()-1, assets.RGB{R: 8, G: 9, B: 9})
	c.line(r.X, r.Bottom()-1, r.Right()-1, r.Bottom()-1, assets.RGB{R: 8, G: 9, B: 9})
	if r.W > 20 && r.H > 20 {
		c.line(r.X+2, r.Y+2, r.Right()-3, r.Y+2, assets.RGB{R: 70, G: 70, B: 62})
		c.line(r.X+2, r.Bottom()-3, r.Right()-3, r.Bottom()-3, assets.RGB{R: 14, G: 16, B: 15})
	}
}

func roomRivets(c *Canvas, r l.Rect, s t.UIStyle) {
	for _, p := range []image.Point{{r.X + 6, r.Y + 6}, {r.Right() - 7, r.Y + 6}, {r.X + 6, r.Bottom() - 7}, {r.Right() - 7, r.Bottom() - 7}} {
		c.fillRect(p.X-2, p.Y-2, 4, 4, assets.RGB{R: 9, G: 10, B: 10})
		c.fillRect(p.X-1, p.Y-1, 2, 2, s.Muted)
	}
}

func roomIcon(c *Canvas, index int, r l.Rect, resource bool) {
	var im *image.RGBA
	var err error
	if resource {
		im, err = roomIcons.ResourceIconRGBA(index, r.W, r.H)
	} else {
		im, err = roomIcons.CommandIconRGBA(index, r.W, r.H)
	}
	if err == nil {
		drawSampledImage(c.img, r, im, im.Bounds())
	}
}

func roomPortrait(c *Canvas, p *i18n.Portrait, placeholder string, r l.Rect, s t.UIStyle) {
	roomSurface(c, r, hdart.Paper, assets.RGB{R: 116, G: 108, B: 86}, .6, s.Muted)
	if p == nil || p.Image == nil {
		// A file card is deliberate: an unverified face must not stand in for a person.
		c.fillRect(r.X+14, r.Y+18, r.W-28, r.H-36, assets.RGB{R: 161, G: 150, B: 122})
		for i := 0; i < 4; i++ {
			c.line(r.X+23, r.Y+37+i*12, r.Right()-24, r.Y+37+i*12, assets.RGB{R: 100, G: 96, B: 78})
		}
		if placeholder != "" {
			roomLabel(c, nil, placeholder, l.Rect{r.X + 6, r.Bottom() - 27, r.W - 12, 20}, assets.RGB{R: 43, G: 42, B: 35}, 12, false)
		}
		return
	}
	b := p.Image.Bounds()
	src := image.Rect(b.Min.X+int(float64(b.Dx())*p.Crop.X), b.Min.Y+int(float64(b.Dy())*p.Crop.Y), b.Min.X+int(float64(b.Dx())*(p.Crop.X+p.Crop.Width)), b.Min.Y+int(float64(b.Dy())*(p.Crop.Y+p.Crop.Height))).Intersect(b)
	dst := l.Rect{r.X + 3, r.Y + 3, r.W - 6, r.H - 6}
	// Contain the verified crop without stretching or replacing the face.
	scale := math.Min(float64(dst.W)/float64(src.Dx()), float64(dst.H)/float64(src.Dy()))
	w, h := int(float64(src.Dx())*scale), int(float64(src.Dy())*scale)
	dst = l.Rect{dst.X + (dst.W-w)/2, dst.Y + (dst.H-h)/2, w, h}
	drawSampledImage(c.img, dst, p.Image, src)
}

func roomForce(force int) string {
	// Display the real troop count; the counter never invents a division count.
	return fmt.Sprint(force)
}
