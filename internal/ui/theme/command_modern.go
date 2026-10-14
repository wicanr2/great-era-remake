package theme

import (
	"fmt"
	"image"
	"image/color"
	"math"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/ui/hdart"
)

// CommandModern keeps the original asset indices, with newly drawn military
// counters and a muted terrain palette. No external game artwork is used.
type CommandModern struct {
	*Modern
	cache      map[[4]int]*image.RGBA
	cacheBytes int
}

func NewCommandModern() *CommandModern {
	m := NewModern()
	pal := commandPalette()
	for _, group := range [][]Bitmap{m.tiles, m.rails, m.units, m.resources, m.commands} {
		for i := range group {
			group[i].Palette = pal
		}
	}
	for i := range m.units {
		m.units[i].Image = earthIndexedUnit(i)
	}
	for i := range m.rails {
		m.rails[i].Image = earthRailBitmap(i)
	}
	// The original-layout HD option uses the same new terrain, sampled back
	// into the established 32×24 palette contract.
	for i := range m.tiles {
		im := image.NewRGBA(image.Rect(0, 0, 96, 72))
		base, accent := terrainColors(i)
		fillTerrainPalette(im, i, base, accent, pal)
		applyCommandTerrain(im, i, base)
		tile := m.tiles[i].Image
		for y := 0; y < tile.H; y++ {
			for x := 0; x < tile.W; x++ {
				p := im.RGBAAt(x*96/tile.W, y*72/tile.H)
				best, at := int(^uint(0)>>1), 1
				for k := 1; k < len(pal); k++ {
					r, g, b := int(p.R)-int(pal[k].R), int(p.G)-int(pal[k].G), int(p.B)-int(pal[k].B)
					d := r*r + g*g + b*b
					if d < best {
						best, at = d, k
					}
				}
				tile.Pix[y*tile.W+x] = byte(at)
			}
		}
	}
	return &CommandModern{Modern: m, cache: make(map[[4]int]*image.RGBA)}
}
func CommandStyle() UIStyle {
	s := (&Modern{}).Style()
	s.Ornaments = nil
	s.Textured = true
	s.Ink = assets.RGB{R: 230, G: 226, B: 212}
	s.Paper = assets.RGB{R: 30, G: 25, B: 21}
	s.Panel = assets.RGB{R: 47, G: 39, B: 31}
	s.Muted = assets.RGB{R: 148, G: 121, B: 84}
	s.Accent = assets.RGB{R: 207, G: 179, B: 121}
	s.Focus = assets.RGB{R: 235, G: 208, B: 151}
	s.AccentAlt = assets.RGB{R: 181, G: 97, B: 78}
	return s
}
func (m *CommandModern) Style() UIStyle                            { return CommandStyle() }
func (m *CommandModern) SetOriginalOrnaments(o *OriginalOrnaments) {} // New artwork has its own borders.
func commandPalette() assets.Palette {
	return assets.Palette{{R: 0, G: 0, B: 0}, {R: 226, G: 210, B: 180}, {R: 38, G: 31, B: 26}, {R: 167, G: 84, B: 65}, {R: 103, G: 137, B: 171}, {R: 105, G: 82, B: 54}, {R: 203, G: 181, B: 112}, {R: 152, G: 115, B: 75}, {R: 46, G: 77, B: 102}, {R: 138, G: 128, B: 116}, {R: 179, G: 148, B: 97}, {R: 139, G: 105, B: 66}, {R: 193, G: 185, B: 165}, {R: 85, G: 70, B: 51}, {R: 235, G: 221, B: 193}, {R: 45, G: 36, B: 29}, {R: 37, G: 30, B: 24}, {R: 173, G: 145, B: 102}}
}

func commandCounter(index int) *assets.Image {
	im := &assets.Image{W: 32, H: 17, Pix: make([]byte, 32*17)}
	fill := byte(4)
	if (index < 6 && index%2 == 1) || index >= 12 {
		fill = 3
	}
	set := func(x, y int, v byte) {
		if x >= 0 && x < 32 && y >= 0 && y < 17 {
			im.Pix[y*32+x] = v
		}
	}
	for y := 2; y < 15; y++ {
		for x := 4; x < 28; x++ {
			v := fill
			if x == 4 || x == 27 || y == 2 || y == 14 {
				v = 1
			}
			set(x, y, v)
		}
	}
	line := func(x0, y0, x1, y1 int) {
		steps := max(absCounter(x1-x0), absCounter(y1-y0))
		for i := 0; i <= steps; i++ {
			set(x0+(x1-x0)*i/max(1, steps), y0+(y1-y0)*i/max(1, steps), 1)
		}
	}
	switch {
	case index < 2:
		line(9, 5, 22, 11)
		line(22, 5, 9, 11)
	case index < 4:
		line(11, 5, 20, 5)
		line(11, 11, 20, 11)
		line(9, 7, 9, 9)
		line(22, 7, 22, 9)
		line(10, 6, 11, 5)
		line(21, 6, 20, 5)
		line(10, 10, 11, 11)
		line(21, 10, 20, 11)
	case index < 6:
		line(9, 11, 22, 5)
	default:
		for y := 7; y < 10; y++ {
			for x := 14; x < 18; x++ {
				set(x, y, 1)
			}
		}
		facing := (index - 6) % 6
		dirs := [6][2]int{{0, -5}, {8, -5}, {8, 0}, {0, 5}, {-8, 5}, {-8, 0}}
		line(16, 8, 16+dirs[facing][0], 8+dirs[facing][1])
	}
	return im
}
func absCounter(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func (m *CommandModern) high(kind, index, w, h int) (*image.RGBA, error) {
	if w < 1 || h < 1 || w > 512 || h > 512 {
		return nil, fmt.Errorf("theme: 圖塊尺寸 %dx%d 超出界限", w, h)
	}
	key := [4]int{kind, index, w, h}
	if im := m.cache[key]; im != nil {
		return im, nil
	}
	var b Bitmap
	var err error
	switch kind {
	case 0:
		b, err = m.Tile(index)
	case 1:
		b, err = m.Rail(index)
	case 2:
		b, err = m.Unit(index)
	case 3:
		b, err = m.ResourceIcon(index)
	case 4:
		b, err = m.CommandIcon(index)
	}
	if err != nil {
		return nil, err
	}
	if kind == 0 || kind == 1 || kind == 2 || kind == 3 || kind == 4 {
		var out *image.RGBA
		switch kind {
		case 0:
			hi := image.NewRGBA(image.Rect(0, 0, w*4, h*4))
			base, accent := terrainColors(index)
			fillTerrainPalette(hi, index, base, accent, commandPalette())
			applyCommandTerrain(hi, index, base)
			out = downsampleHUD(hi, w, h)
			earthTerrainTint(out, index)
		case 1:
			out = earthRailImage(index, w, h)
		case 2:
			out = hdart.EarthUnit(index, w, h)
		case 3:
			out, err = m.Modern.ResourceIconRGBA(index, w, h)
		case 4:
			out, err = m.Modern.CommandIconRGBA(index, w, h)
		}
		if err != nil {
			return nil, err
		}
		if kind == 3 || kind == 4 {
			recolorCommand(out)
		}
		if len(m.cache) >= 192 || m.cacheBytes+len(out.Pix) > 16*1024*1024 {
			clear(m.cache)
			m.cacheBytes = 0
		}
		m.cache[key] = out
		m.cacheBytes += len(out.Pix)
		return out, nil
	}
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			v := b.Image.Pix[(y*b.Image.H/h)*b.Image.W+x*b.Image.W/w]
			if v == 0 && kind != 0 {
				continue
			}
			c := b.Palette[v]
			out.SetRGBA(x, y, color.RGBA{c.R, c.G, c.B, 255})
		}
	}
	if len(m.cache) >= 192 {
		clear(m.cache)
	}
	m.cache[key] = out
	return out, nil
}

func applyCommandTerrain(im *image.RGBA, index int, base byte) {
	if index == 21 {
		return
	} // Preserve the existing unknown-terrain fallback.
	materials := [22]int{hdart.Plain, hdart.Grass, hdart.Coastal, hdart.Forest, hdart.Plain, hdart.Mountain, hdart.Desert, hdart.Coastal, hdart.Coastal, hdart.Limestone, hdart.Mountain, hdart.Limestone, hdart.Limestone, hdart.Limestone, hdart.Limestone, hdart.Limestone, hdart.Limestone, hdart.Limestone, hdart.Limestone, hdart.Limestone, hdart.Limestone, hdart.Canvas}
	bc := commandPalette()[base]
	for y := 0; y < im.Bounds().Dy(); y++ {
		for x := 0; x < im.Bounds().Dx(); x++ {
			p := hdart.Sample(materials[index], x/2, y/2)
			v := im.RGBAAt(x, y)
			distance := absCounter(int(v.R)-int(bc.R)) + absCounter(int(v.G)-int(bc.G)) + absCounter(int(v.B)-int(bc.B))
			amount := .72
			if distance > 60 {
				amount = .22
			} // Keep the established city, bridge and wall symbols legible.
			if index == 0 || index == 1 || index == 2 || index == 3 || index == 5 || index == 6 || index == 9 {
				// Natural terrain has continuous photographic detail, without
				// oversized triangle peaks or diagram trees from the old HD set.
				v = color.RGBA{bc.R, bc.G, bc.B, 255}
				amount = .85
			}
			im.SetRGBA(x, y, color.RGBA{byte(float64(v.R)*(1-amount) + float64(p.R)*amount), byte(float64(v.G)*(1-amount) + float64(p.G)*amount), byte(float64(v.B)*(1-amount) + float64(p.B)*amount), 255})
		}
	}
}
func recolorCommand(im *image.RGBA) {
	old, pal := modernPalette(), commandPalette()
	for y := 0; y < im.Bounds().Dy(); y++ {
		for x := 0; x < im.Bounds().Dx(); x++ {
			c := im.RGBAAt(x, y)
			if c.A == 0 {
				continue
			}
			best, at := math.MaxFloat64, 1
			for i := 1; i < len(old); i++ {
				p := old[i]
				r, g, b := float64(int(c.R)-int(p.R)), float64(int(c.G)-int(p.G)), float64(int(c.B)-int(p.B))
				d := r*r + g*g + b*b
				if d < best {
					best, at = d, i
				}
			}
			p := pal[at]
			im.SetRGBA(x, y, color.RGBA{p.R, p.G, p.B, c.A})
		}
	}
}
func highCommandCounter(index, w, h int) *image.RGBA {
	hi := image.NewRGBA(image.Rect(0, 0, w*4, h*4))
	pal := commandPalette()
	team := 4
	if (index < 6 && index%2 == 1) || index >= 12 {
		team = 3
	}
	wc, hc := w*4, h*4
	ink := toHighRGBA(pal[1])
	fillHighRect(hi, wc/8, hc/8, wc*7/8, hc*7/8, ink)
	fillHighRect(hi, wc/8+4, hc/8+4, wc*7/8-4, hc*7/8-4, toHighRGBA(pal[team]))
	line := func(x0, y0, x1, y1 int) { lineHUD(hi, x0*wc/32, y0*hc/17, x1*wc/32, y1*hc/17, max(3, wc/64), ink) }
	switch {
	case index < 2:
		line(9, 5, 23, 12)
		line(23, 5, 9, 12)
	case index < 4:
		for i := 0; i < 96; i++ {
			a, b := float64(i)*math.Pi/48, float64(i+1)*math.Pi/48
			lineHUD(hi, int((16+6.5*math.Cos(a))*float64(wc)/32), int((8.5+3.5*math.Sin(a))*float64(hc)/17), int((16+6.5*math.Cos(b))*float64(wc)/32), int((8.5+3.5*math.Sin(b))*float64(hc)/17), max(3, wc/64), ink)
		}
	case index < 6:
		line(9, 12, 23, 5)
	default:
		fillHighCircle(hi, wc/2, hc/2, max(5, hc/10), ink)
		dirs := [6][2]int{{0, -5}, {8, -5}, {8, 0}, {0, 5}, {-8, 5}, {-8, 0}}
		d := dirs[(index-6)%6]
		line(16, 8, 16+d[0], 8+d[1])
	}
	return downsampleHUD(hi, w, h)
}
func (m *CommandModern) HighTile(i, w, h int) (*image.RGBA, error)         { return m.high(0, i, w, h) }
func (m *CommandModern) HighRail(i, w, h int) (*image.RGBA, error)         { return m.high(1, i, w, h) }
func (m *CommandModern) HighUnit(i, w, h int) (*image.RGBA, error)         { return m.high(2, i, w, h) }
func (m *CommandModern) ResourceIconRGBA(i, w, h int) (*image.RGBA, error) { return m.high(3, i, w, h) }
func (m *CommandModern) CommandIconRGBA(i, w, h int) (*image.RGBA, error)  { return m.high(4, i, w, h) }
