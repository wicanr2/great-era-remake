package render

import (
	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/ui/hdart"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
	"image"
	"image/draw"
)

// Nine-slice corners remain small and sharp. Only the quiet center stretches.
func earthFrame(c *Canvas, r l.Rect, kind int) bool {
	if c.lightUI && r.W >= 4 && r.H >= 4 {
		ink := assets.RGB{185, 172, 148}
		if kind == 1 {
			ink = assets.RGB{192, 179, 155}
		}
		if kind == 3 {
			ink = assets.RGB{173, 145, 94}
		}
		c.fillUIRect(r.X, r.Y, r.W, r.H, ink)
		c.strokeUIRect(r.X, r.Y, r.W, r.H, ink)
		c.line(r.X+1, r.Y+1, r.Right()-2, r.Y+1, assets.RGB{224, 213, 189})
		c.line(r.X+1, r.Bottom()-2, r.Right()-2, r.Bottom()-2, assets.RGB{81, 74, 59})
		return true
	}
	src := hdart.EarthPanel(kind)
	if src == nil || r.W < 4 || r.H < 4 {
		return false
	}
	edge := min(10, min(r.W, r.H)/3)
	cutsX := [4]int{r.X, r.X + edge, r.Right() - edge, r.Right()}
	cutsY := [4]int{r.Y, r.Y + edge, r.Bottom() - edge, r.Bottom()}
	b := src.Bounds()
	sx := [4]int{b.Min.X, b.Min.X + 24, b.Max.X - 24, b.Max.X}
	sy := [4]int{b.Min.Y, b.Min.Y + 24, b.Max.Y - 24, b.Max.Y}
	for row := 0; row < 3; row++ {
		for col := 0; col < 3; col++ {
			dst := image.Rect(cutsX[col], cutsY[row], cutsX[col+1], cutsY[row+1]).Intersect(c.Bounds())
			if dst.Empty() {
				continue
			}
			for y := dst.Min.Y; y < dst.Max.Y; y++ {
				for x := dst.Min.X; x < dst.Max.X; x++ {
					px := sx[col] + (x-cutsX[col])*(sx[col+1]-sx[col])/max(1, cutsX[col+1]-cutsX[col])
					py := sy[row] + (y-cutsY[row])*(sy[row+1]-sy[row])/max(1, cutsY[row+1]-cutsY[row])
					c.img.SetRGBA(x, y, src.RGBAAt(px, py))
				}
			}
		}
	}
	return true
}
func earthUnit(c *Canvas, index int, r l.Rect) bool {
	im := hdart.EarthUnit(index, r.W, r.H)
	if im == nil {
		return false
	}
	draw.Draw(c.img, image.Rect(r.X, r.Y, r.Right(), r.Bottom()), im, image.Point{}, draw.Over)
	return true
}
