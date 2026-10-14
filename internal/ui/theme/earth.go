package theme

import (
	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/ui/hdart"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
	"image"
	"image/color"
)

func earthIndexedUnit(index int) *assets.Image {
	src := hdart.EarthUnit(index, 32, 17)
	out := &assets.Image{W: 32, H: 17, Pix: make([]byte, 32*17)}
	if src == nil {
		return out
	}
	pal := commandPalette()
	for y := 0; y < 17; y++ {
		for x := 0; x < 32; x++ {
			p := src.RGBAAt(x, y)
			if p.A < 32 {
				continue
			}
			best, at := 1<<30, 1
			for i := 1; i < len(pal); i++ {
				r, g, b := int(p.R)-int(pal[i].R), int(p.G)-int(pal[i].G), int(p.B)-int(pal[i].B)
				d := r*r + g*g + b*b
				if d < best {
					best, at = d, i
				}
			}
			out.Pix[y*32+x] = byte(at)
		}
	}
	return out
}
func earthRailImage(index, w, h int) *image.RGBA {
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	ports := l.BattleRailPorts(index)
	if ports == nil {
		return out
	}
	cx, cy := w/2, h/2
	for _, dir := range ports {
		if dir == 4 || dir == 6 {
			cy = h / 4
			break
		}
		if dir == 1 || dir == 3 {
			cy = h * 3 / 4
		}
	}
	for _, dir := range ports {
		x, y := cx, cy
		switch dir {
		case 1:
			x, y = 0, h*3/4
		case 2:
			x, y = cx, h-1
		case 3:
			x, y = w-1, h*3/4
		case 4:
			x, y = 0, h/4
		case 5:
			x, y = cx, 0
		case 6:
			x, y = w-1, h/4
		}
		lineHUD(out, cx, cy, x, y, max(1, w/18), color.RGBA{190, 177, 152, 255})
	}
	if index >= 11 {
		fillHighRect(out, cx-2, cy-2, cx+3, cy+3, color.RGBA{163, 130, 84, 255})
	}
	return out
}
func earthRailBitmap(index int) *assets.Image {
	src := earthRailImage(index, 32, 24)
	out := &assets.Image{W: 32, H: 24, Pix: make([]byte, 32*24)}
	for y := 0; y < 24; y++ {
		for x := 0; x < 32; x++ {
			if src.RGBAAt(x, y).A > 0 {
				out.Pix[y*32+x] = 1
			}
		}
	}
	return out
}
func earthTerrainTint(im *image.RGBA, index int) {
	bases := [12][3]int{{139, 105, 66}, {152, 115, 75}, {46, 77, 102}, {85, 70, 51}, {149, 130, 106}, {138, 128, 116}, {179, 148, 97}, {143, 132, 110}, {143, 132, 110}, {155, 120, 81}, {120, 96, 71}, {123, 116, 100}}
	b := bases[min(index, 11)]
	for y := 0; y < im.Bounds().Dy(); y++ {
		for x := 0; x < im.Bounds().Dx(); x++ {
			p := im.RGBAAt(x, y)
			light := (int(p.R)+int(p.G)+int(p.B))/3 - 115
			im.SetRGBA(x, y, color.RGBA{uint8(max(0, min(255, b[0]+light/2))), uint8(max(0, min(255, b[1]+light/2))), uint8(max(0, min(255, b[2]+light/2))), p.A})
		}
	}
}
