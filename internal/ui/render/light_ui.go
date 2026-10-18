package render

import (
	"image"
	"image/color"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

// UseLightUI affects only explicitly marked UI drawing. Image copying, terrain,
// railways and unit sprites keep their original pixels. No package-global
// palette or gameplay RNG is changed.
func (c *Canvas) UseLightUI(enabled bool) { c.lightUI = enabled }

var lightSurfaces = map[assets.RGB]assets.RGB{
	{25, 26, 28}: {211, 198, 173}, {25, 24, 23}: {211, 198, 173}, {25, 23, 22}: {211, 198, 173},
	{38, 40, 43}: {185, 172, 148}, {39, 36, 32}: {185, 172, 148}, {39, 35, 31}: {185, 172, 148},
	{40, 42, 45}: {157, 148, 132}, {28, 30, 32}: {141, 133, 120},
	{32, 34, 37}: {185, 172, 148}, {34, 36, 38}: {185, 172, 148},
	{36, 37, 40}: {185, 172, 148}, {37, 37, 40}: {185, 172, 148}, {32, 33, 35}: {146, 138, 124},
	{41, 38, 36}: {146, 138, 124}, {32, 30, 27}: {146, 138, 124},
	{54, 54, 56}: {202, 189, 164}, {38, 37, 39}: {183, 170, 147},
	{88, 70, 51}: {173, 145, 94}, {119, 92, 60}: {181, 154, 106}, {73, 60, 45}: {162, 134, 86},
	{152, 84, 58}: {141, 80, 55}, {104, 58, 43}: {113, 62, 44},
	{189, 114, 80}: {129, 69, 47},
	{139, 73, 53}:  {141, 80, 55}, {98, 51, 38}: {113, 62, 44},
	{48, 46, 43}: {195, 182, 157}, {38, 39, 42}: {185, 172, 148},
	{217, 185, 128}: {211, 198, 173}, {185, 155, 108}: {204, 190, 164},
}

func (c *Canvas) uiSurface(ink assets.RGB) assets.RGB {
	if !c.lightUI {
		return ink
	}
	if light, ok := lightSurfaces[ink]; ok {
		return light
	}
	// Gradients between the old charcoal endpoints remain warm metal, without
	// applying a transform to any picture or to arbitrary terrain colors.
	if ink.R < 80 && ink.G < 80 && ink.B < 80 {
		shade := (int(ink.R) + int(ink.G) + int(ink.B)) / 12
		return assets.RGB{uint8(179 + shade), uint8(166 + shade), uint8(143 + shade)}
	}
	return ink
}

func (c *Canvas) uiText(ink assets.RGB, x, y int) assets.RGB {
	if !c.lightUI {
		return ink
	}
	bg := c.img.RGBAAt(x, y)
	if int(bg.R)+int(bg.G)+int(bg.B) < 280 {
		// Rust-red primary buttons and dark counters need light lettering.
		return assets.RGB{255, 246, 232}
	}
	if ink.B > ink.R+15 && ink.G > ink.R {
		return assets.RGB{43, 74, 99}
	}
	if int(ink.R)-int(ink.G) > 45 && int(ink.G)-int(ink.B) < 50 {
		return assets.RGB{121, 66, 46}
	}
	if int(ink.R)-int(ink.B) > 65 {
		return assets.RGB{104, 66, 29}
	}
	if int(ink.R)+int(ink.G)+int(ink.B) >= 580 {
		return assets.RGB{33, 31, 26}
	}
	if int(ink.R)+int(ink.G)+int(ink.B) >= 400 {
		return assets.RGB{68, 62, 53}
	}
	return ink
}

func (c *Canvas) uiEdge(ink assets.RGB) assets.RGB {
	if c.lightUI {
		return assets.RGB{98, 89, 74}
	}
	return ink
}

func (c *Canvas) fillUIRect(x, y, w, h int, ink assets.RGB) {
	if !c.lightUI {
		c.fillRect(x, y, w, h, ink)
		return
	}
	ink = c.uiSurface(ink)
	r := image.Rect(x, y, x+w, y+h).Intersect(c.Bounds())
	for yy := r.Min.Y; yy < r.Max.Y; yy++ {
		for xx := r.Min.X; xx < r.Max.X; xx++ {
			grain := int((uint32(xx)*73856093^uint32(yy)*19349663)%3) - 1
			if yy%4 == 0 {
				grain--
			}
			if (xx+37*yy)%1201 < 2 {
				grain += 4
			}
			mix := func(v byte) byte { return byte(max(0, min(255, int(v)+grain))) }
			c.img.SetRGBA(xx, yy, color.RGBA{mix(ink.R), mix(ink.G), mix(ink.B), 255})
		}
	}
}

func (c *Canvas) strokeUIRect(x, y, w, h int, ink assets.RGB) {
	c.strokeRect(x, y, w, h, c.uiEdge(ink))
}

func (c *Canvas) lineUI(x1, y1, x2, y2 int, ink assets.RGB) {
	c.line(x1, y1, x2, y2, c.uiEdge(ink))
}
