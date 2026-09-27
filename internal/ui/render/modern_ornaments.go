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
	}
	// 原版致敬：四角改繪花框角（暗紅外圈＋寶藍內圈）。
	if r.W >= 30 && r.H >= 30 {
		s := 9
		drawHomageCorner(c, r.X+1, r.Y+1, false, false, s, style.Ink, style.Focus)
		drawHomageCorner(c, r.Right()-2, r.Y+1, true, false, s, style.Ink, style.Focus)
		drawHomageCorner(c, r.X+1, r.Bottom()-2, false, true, s, style.Ink, style.Focus)
		drawHomageCorner(c, r.Right()-2, r.Bottom()-2, true, true, s, style.Ink, style.Focus)
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
	if r.W >= 40 && r.H >= 40 {
		s := 11
		drawHomageCorner(c, r.X+2, r.Y+2, false, false, s, style.Ink, style.Focus)
		drawHomageCorner(c, r.Right()-3, r.Y+2, true, false, s, style.Ink, style.Focus)
		drawHomageCorner(c, r.X+2, r.Bottom()-3, false, true, s, style.Ink, style.Focus)
		drawHomageCorner(c, r.Right()-3, r.Bottom()-3, true, true, s, style.Ink, style.Focus)
	}
	if style.Ornaments.Valid() {
		drawModernOrnamentMask(c, style.Ornaments.PanelFrame, r, style.Accent)
	}
}

// drawHomageCorner 在給定的外角像素處畫一個原版致敬花框角。(ax,ay) 是外角，
// flipX／flipY 把局部座標鏡射到其餘三個角。s 為邊長（建議 9）。
// 外圈暗紅雙線、內圈寶藍線、角端實心方塊，均為程式化重繪，不讀原版位元組。
func drawHomageCorner(c *Canvas, ax, ay int, flipX, flipY bool, s int, outer, inner assets.RGB) {
	if c == nil || s < 7 {
		return
	}
	put := func(i, j int, col assets.RGB) {
		x, y := ax, ay
		if flipX {
			x -= i
		} else {
			x += i
		}
		if flipY {
			y -= j
		} else {
			y += j
		}
		c.setPixel(x, y, col)
	}
	for i := 0; i < s; i++ {
		put(i, 0, outer)
		put(0, i, outer)
	}
	for i := 2; i < s; i++ {
		put(i, 2, outer)
		put(2, i, outer)
	}
	for i := 4; i < s; i++ {
		put(i, 4, inner)
		put(4, i, inner)
	}
	for di := 0; di < 3; di++ {
		for dj := 0; dj < 3; dj++ {
			put(di, dj, outer)
		}
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
