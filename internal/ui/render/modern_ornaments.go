package render

import (
	"github.com/wicanr2/great-era-remake/internal/assets"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

// drawModernDecoratedControl 是 A2 所有可操作控制項的共用底板。原版遮罩
// 只在 runtime 可用時疊上；否則仍以雙線與折角構成完整、可再散布的 fallback。
func drawModernDecoratedControl(c *Canvas, r uilayout.Rect, style uitheme.UIStyle,
	fill, border assets.RGB) {
	c.fillRect(r.X, r.Y, r.W, r.H, fill)
	c.strokeRect(r.X, r.Y, r.W, r.H, border)
	if r.W > 8 && r.H > 8 {
		c.strokeRect(r.X+4, r.Y+4, r.W-8, r.H-8, border)
		for _, p := range []struct{ x, y int }{
			{r.X + 2, r.Y + 2}, {r.Right() - 5, r.Y + 2},
			{r.X + 2, r.Bottom() - 5}, {r.Right() - 5, r.Bottom() - 5},
		} {
			c.fillRect(p.x, p.y, 3, 3, style.Focus)
		}
	}
	o := style.Ornaments
	if !o.Valid() {
		return
	}
	edgeX := min(10, max(4, r.W/12))
	edgeY := min(9, max(4, r.H/6))
	drawModernOrnamentMask(c, o.ButtonLeft, uilayout.Rect{X: r.X, Y: r.Y, W: edgeX, H: r.H}, border)
	drawModernOrnamentMask(c, o.ButtonRight, uilayout.Rect{X: r.Right() - edgeX, Y: r.Y, W: edgeX, H: r.H}, border)
	drawModernOrnamentMask(c, o.ButtonTop, uilayout.Rect{X: r.X, Y: r.Y, W: r.W, H: edgeY}, border)
	drawModernOrnamentMask(c, o.ButtonBottom, uilayout.Rect{X: r.X, Y: r.Bottom() - edgeY, W: r.W, H: edgeY}, border)
	bandW := min(r.W/3, 116)
	if bandW > 12 {
		drawModernOrnamentMask(c, o.ButtonBand,
			uilayout.Rect{X: r.X + (r.W-bandW)/2, Y: r.Y + 2, W: bandW, H: max(4, edgeY-2)}, style.Focus)
	}
}

// drawModernDecoratedPanel 以 WARMENU 的原版外框遮罩強化 A2 面板，但不把
// 原始點陣或調色盤寫入任何輸出資產。
func drawModernDecoratedPanel(c *Canvas, r uilayout.Rect, style uitheme.UIStyle, fill assets.RGB) {
	c.fillRect(r.X, r.Y, r.W, r.H, fill)
	c.strokeRect(r.X, r.Y, r.W, r.H, style.Muted)
	if style.Ornaments.Valid() {
		drawModernOrnamentMask(c, style.Ornaments.PanelFrame, r, style.Accent)
	}
}

func drawModernOrnamentMask(c *Canvas, src *assets.Image, dst uilayout.Rect, ink assets.RGB) {
	if c == nil || src == nil || src.W <= 0 || src.H <= 0 || dst.W <= 0 || dst.H <= 0 {
		return
	}
	background := dominantPaletteIndex(src.Pix)
	for y := 0; y < dst.H; y++ {
		sy := y * src.H / dst.H
		for x := 0; x < dst.W; x++ {
			sx := x * src.W / dst.W
			if src.Pix[sy*src.W+sx] != background {
				c.setPixel(dst.X+x, dst.Y+y, ink)
			}
		}
	}
}

func dominantPaletteIndex(pix []byte) byte {
	var counts [256]int
	for _, p := range pix {
		counts[p]++
	}
	best := byte(0)
	for i := 1; i < len(counts); i++ {
		if counts[i] > counts[best] {
			best = byte(i)
		}
	}
	return best
}
