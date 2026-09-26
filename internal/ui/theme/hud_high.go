package theme

import (
	"fmt"
	"image"
	"image/color"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

const hudSupersample = 4

// ResourceIconRGBA 依 Modern 畫面的實際目的尺寸重繪資源圖示。
func (m *Modern) ResourceIconRGBA(index, width, height int) (*image.RGBA, error) {
	if index < 0 || index >= ResourceIconCount {
		return nil, fmt.Errorf("theme: modern 資源圖示索引 %d 超出 0..%d", index, ResourceIconCount-1)
	}
	return renderSmoothHUDIcon(index, width, height, false, modernPalette())
}

// CommandIconRGBA 依 Modern 畫面的實際目的尺寸重繪指令圖示。
func (m *Modern) CommandIconRGBA(index, width, height int) (*image.RGBA, error) {
	if index < 0 || index >= CommandIconCount {
		return nil, fmt.Errorf("theme: modern 指令圖示索引 %d 超出 0..%d", index, CommandIconCount-1)
	}
	return renderSmoothHUDIcon(index, width, height, true, modernPalette())
}

func renderSmoothHUDIcon(index, width, height int, command bool, palette assets.Palette) (*image.RGBA, error) {
	if width <= 0 || height <= 0 || len(palette) < 16 {
		return nil, fmt.Errorf("theme: high-resolution HUD 尺寸或調色盤無效：%dx%d", width, height)
	}
	w, h := width*hudSupersample, height*hudSupersample
	hi := image.NewRGBA(image.Rect(0, 0, w, h))
	toRGBA := func(p assets.RGB) color.RGBA { return color.RGBA{R: p.R, G: p.G, B: p.B, A: 0xff} }
	ink, accent, light := toRGBA(palette[2]), toRGBA(palette[commandAccentIndex(index)]), toRGBA(palette[1])
	cx, cy := w/2, h/2
	r := minHUD(w, h) * 34 / 100
	fillHUDCircle(hi, cx, cy, r, accent)
	fillHUDCircle(hi, cx, cy, r*76/100, ink)
	stroke := maxHUD(hudSupersample, minHUD(w, h)/18)
	if command {
		drawCommandMark(hi, index, cx, cy, r*68/100, stroke, light, accent)
	} else {
		drawResourceMark(hi, index, cx, cy, r*68/100, stroke, light, accent)
	}
	return downsampleHUD(hi, width, height), nil
}

func drawCommandMark(im *image.RGBA, index, cx, cy, r, stroke int, light, accent color.RGBA) {
	line := func(x0, y0, x1, y1 int, c color.RGBA) { lineHUD(im, x0, y0, x1, y1, stroke, c) }
	poly := func(points [][2]int, c color.RGBA) { polygonHUD(im, points, stroke, c) }
	switch index {
	case 0: // 兩省定位點與雙向箭頭
		strokeHUDCircle(im, cx-r*2/3, cy, r/4, stroke, light)
		strokeHUDCircle(im, cx+r*2/3, cy, r/4, stroke, light)
		line(cx-r/3, cy-r/5, cx+r/3, cy-r/5, light)
		line(cx+r/3, cy+r/5, cx-r/3, cy+r/5, light)
		line(cx+r/3, cy-r/5, cx+r/8, cy-r*2/5, light)
		line(cx-r/3, cy+r/5, cx-r/8, cy+r*2/5, light)
	case 1: // 軍旗與交叉軍刀
		line(cx-r*2/3, cy-r, cx-r*2/3, cy+r, light)
		poly([][2]int{{cx - r*2/3, cy - r}, {cx + r/2, cy - r*2/3}, {cx, cy - r/8}}, light)
		line(cx-r/2, cy-r/2, cx+r/2, cy+r/2, accent)
		line(cx+r/2, cy-r/2, cx-r/2, cy+r/2, accent)
	case 2: // 補給箱、道路與箭頭
		poly([][2]int{{cx - r*2/3, cy - r/4}, {cx, cy - r/2}, {cx + r*2/3, cy - r/4}, {cx + r/2, cy + r/2}, {cx - r*2/3, cy + r/2}}, light)
		line(cx-r/3, cy+r*2/3, cx+r, cy+r*2/3, light)
		line(cx+r/2, cy+r/2, cx+r, cy+r*2/3, light)
		line(cx+r/2, cy+r*5/6, cx+r, cy+r*2/3, light)
		strokeHUDCircle(im, cx-r/2, cy+r*2/3, r/8, stroke, accent)
	case 3: // 帳冊、錢幣與印章
		poly([][2]int{{cx - r*2/3, cy - r*2/3}, {cx - r/8, cy - r*2/3}, {cx - r/8, cy + r*2/3}, {cx - r*2/3, cy + r*2/3}}, light)
		line(cx-r/2, cy-r/3, cx-r/4, cy-r/3, accent)
		line(cx-r/2, cy, cx-r/4, cy, accent)
		strokeHUDCircle(im, cx+r/3, cy-r/8, r/3, stroke, light)
		line(cx+r/3, cy-r/3, cx+r/3, cy+r/8, accent)
		line(cx+r/6, cy-r/8, cx+r/2, cy-r/8, accent)
		strokeHUDCircle(im, cx+r/2, cy+r/2, r/5, stroke, accent)
	case 4: // 鋼盔、人形與加號
		poly([][2]int{{cx - r/2, cy - r/3}, {cx - r/3, cy - r*2/3}, {cx + r/3, cy - r*2/3}, {cx + r/2, cy - r/3}}, light)
		poly([][2]int{{cx - r/3, cy - r/5}, {cx + r/3, cy - r/5}, {cx + r/2, cy + r*2/3}, {cx - r/2, cy + r*2/3}}, light)
		line(cx-r*2/3, cy+r/3, cx+r*2/3, cy+r/3, accent)
		line(cx+r*2/3, cy-r/3, cx+r*2/3, cy+r/3, light)
		line(cx+r/2, cy, cx+r, cy, light)
	case 5: // 摺頁地圖與放大鏡
		poly([][2]int{{cx - r, cy - r*2/3}, {cx - r/3, cy - r}, {cx + r/3, cy - r*2/3}, {cx + r, cy - r}, {cx + r, cy + r}, {cx + r/3, cy + r*2/3}, {cx - r/3, cy + r}, {cx - r, cy + r*2/3}}, light)
		line(cx-r/3, cy-r, cx-r/3, cy+r, accent)
		line(cx+r/3, cy-r*2/3, cx+r/3, cy+r*2/3, accent)
		strokeHUDCircle(im, cx+r/3, cy+r/3, r/3, stroke, light)
		line(cx+r*2/3, cy+r*2/3, cx+r, cy+r, light)
	case 6: // 藍圖、磚塊與十字鎬
		poly([][2]int{{cx - r, cy - r*2/3}, {cx + r/3, cy - r*2/3}, {cx + r/3, cy + r/2}, {cx - r, cy + r/2}}, light)
		line(cx-r*2/3, cy-r/4, cx, cy-r/4, accent)
		line(cx-r*2/3, cy+r/8, cx, cy+r/8, accent)
		line(cx-r/2, cy+r*2/3, cx-r/8, cy+r*2/3, accent)
		line(cx+r/8, cy+r*2/3, cx+r/2, cy+r*2/3, accent)
		line(cx+r/3, cy+r/2, cx+r, cy-r, light)
		line(cx+r/8, cy-r/3, cx+r, cy-r, light)
	case 7: // 蓋章政策公文
		poly([][2]int{{cx - r*2/3, cy - r*2/3}, {cx + r/3, cy - r*2/3}, {cx + r/3, cy + r*2/3}, {cx - r*2/3, cy + r*2/3}}, light)
		line(cx-r/3, cy-r/3, cx+r/8, cy-r/3, accent)
		line(cx-r/3, cy, cx+r/4, cy, accent)
		line(cx-r/3, cy+r/3, cx+r/8, cy+r/3, accent)
		strokeHUDCircle(im, cx+r/2, cy+r/2, r/4, stroke, accent)
		line(cx+r/3, cy+r/2, cx+2*r/3, cy+r/2, light)
		line(cx+r/2, cy+r/3, cx+r/2, cy+2*r/3, light)
	case 8: // 兩旗與握手
		line(cx-r*2/3, cy-r, cx-r*2/3, cy+r, light)
		line(cx+r*2/3, cy-r, cx+r*2/3, cy+r, light)
		poly([][2]int{{cx - r*2/3, cy - r}, {cx, cy - r/2}, {cx - r*2/3, cy}}, light)
		poly([][2]int{{cx + r*2/3, cy - r}, {cx, cy - r/2}, {cx + r*2/3, cy}}, light)
		line(cx-r/2, cy+r/4, cx, cy, accent)
		line(cx+r/2, cy+r/4, cx, cy, accent)
		line(cx-r/8, cy+r/4, cx+r/8, cy+r/4, accent)
	case 9: // 交叉軍刀與停火條
		line(cx-r*2/3, cy-r*2/3, cx+r*2/3, cy+r*2/3, light)
		line(cx+r*2/3, cy-r*2/3, cx-r*2/3, cy+r*2/3, light)
		poly([][2]int{{cx - r, cy - r/5}, {cx + r, cy - r/5}, {cx + r, cy + r/5}, {cx - r, cy + r/5}}, accent)
		line(cx-r*5/4, cy, cx-r, cy, light)
		line(cx+r, cy, cx+r*5/4, cy, light)
	case 10: // 眼睛、虛線路徑與遮蔽角
		poly([][2]int{{cx - r, cy}, {cx - r/2, cy - r/2}, {cx + r/2, cy - r/2}, {cx + r, cy}, {cx + r/2, cy + r/2}, {cx - r/2, cy + r/2}}, light)
		strokeHUDCircle(im, cx, cy, r/4, stroke, accent)
		for i := -2; i <= 2; i++ {
			line(cx+i*r/2, cy+r*3/4, cx+i*r/2+r/8, cy+r*3/4, light)
		}
		poly([][2]int{{cx + r/2, cy + r/2}, {cx + r, cy + r/2}, {cx + r, cy + r}}, accent)
	case 11: // 商店櫃台與雙向箭頭
		poly([][2]int{{cx - r, cy - r/3}, {cx, cy - r}, {cx + r, cy - r/3}}, light)
		poly([][2]int{{cx - r*2/3, cy - r/6}, {cx + r*2/3, cy - r/6}, {cx + r*2/3, cy + r*2/3}, {cx - r*2/3, cy + r*2/3}}, light)
		line(cx-r, cy+r, cx+r, cy+r, light)
		line(cx-r*2/3, cy-r*2/3, cx+r*2/3, cy-r*2/3, accent)
		line(cx+r/3, cy-r, cx+r*2/3, cy-r*2/3, accent)
		line(cx+r/3, cy-r/3, cx-r*2/3, cy-r/3, accent)
		line(cx-r/3, cy-r/6, cx-r*2/3, cy-r/3, accent)
	case 12: // 三列隊形、旗標與靶心
		for _, x := range []int{cx - r*2/3, cx, cx + r*2/3} {
			strokeHUDCircle(im, x, cy-r/3, r/8, stroke, light)
			line(x, cy, x, cy+r/3, light)
			line(x-r/8, cy+r/8, x+r/8, cy+r/8, light)
		}
		line(cx-r, cy-r, cx-r, cy+r, accent)
		poly([][2]int{{cx - r, cy - r}, {cx - r/3, cy - r*2/3}, {cx - r, cy - r/3}}, accent)
		strokeHUDCircle(im, cx+r/2, cy+r/2, r/3, stroke, light)
		strokeHUDCircle(im, cx+r/2, cy+r/2, r/8, stroke, accent)
	case 13: // 慰問桌、兩人與心形
		strokeHUDCircle(im, cx-r/2, cy-r/2, r/8, stroke, light)
		strokeHUDCircle(im, cx+r/2, cy-r*2/3, r/8, stroke, light)
		line(cx-r/2, cy-r/3, cx-r/2, cy+r/4, light)
		line(cx+r/2, cy-r/2, cx+r/2, cy+r/4, light)
		line(cx-r, cy+r/2, cx+r, cy+r/2, light)
		line(cx-r/3, cy+r/2, cx-r/3, cy+r, accent)
		line(cx+r/3, cy+r/2, cx+r/3, cy+r, accent)
		poly([][2]int{{cx, cy - r/3}, {cx + r/4, cy - r/2}, {cx + r/2, cy - r/3}, {cx, cy + r/3}, {cx - r/2, cy - r/3}, {cx - r/4, cy - r/2}}, accent)
	case 14: // 三列設定方格與右下三點
		for _, y := range []int{cy - r*2/3, cy, cy + r*2/3} {
			line(cx-r*2/3, y, cx+r/2, y, light)
			strokeHUDCircle(im, cx-r/8, y, r/8, stroke, accent)
		}
		strokeHUDCircle(im, cx+r*2/3, cy+r*2/3, r/10, stroke, light)
		strokeHUDCircle(im, cx+r, cy+r*2/3, r/10, stroke, light)
		strokeHUDCircle(im, cx+r/3, cy+r*2/3, r/10, stroke, light)
	}
}

func drawResourceMark(im *image.RGBA, index, cx, cy, r, stroke int, light, accent color.RGBA) {
	switch index {
	case 0: // 金幣
		strokeHUDCircle(im, cx, cy, r, stroke, light)
		lineHUD(im, cx, cy-r/2, cx, cy+r/2, stroke, accent)
	case 1: // 穀穗
		lineHUD(im, cx, cy+r, cx, cy-r, stroke, light)
		for y := -r / 2; y <= r/2; y += maxHUD(stroke*2, 1) {
			lineHUD(im, cx, cy+y, cx-r/2, cy+y-r/3, stroke, accent)
			lineHUD(im, cx, cy+y, cx+r/2, cy+y-r/3, stroke, accent)
		}
	case 2: // 彈匣
		polygonHUD(im, [][2]int{{cx - r/2, cy - r}, {cx + r/2, cy - r}, {cx + r/2, cy + r}, {cx - r/2, cy + r}}, stroke, light)
		lineHUD(im, cx, cy-r/2, cx, cy+r/2, stroke, accent)
	case 3: // 燃料滴
		polygonHUD(im, [][2]int{{cx, cy - r}, {cx + r*2/3, cy}, {cx + r/2, cy + r*2/3}, {cx, cy + r}, {cx - r/2, cy + r*2/3}, {cx - r*2/3, cy}}, stroke, light)
	case 4: // 煤礦
		polygonHUD(im, [][2]int{{cx - r, cy + r/2}, {cx - r/2, cy - r}, {cx + r/2, cy - r*2/3}, {cx + r, cy + r/3}, {cx + r/2, cy + r}}, stroke, light)
	case 5: // 鐵錠
		polygonHUD(im, [][2]int{{cx - r, cy - r/2}, {cx + r/2, cy - r/2}, {cx + r, cy + r/2}, {cx + r/2, cy + r}, {cx - r*2/3, cy + r*2/3}}, stroke, light)
	}
}

func fillHUDCircle(im *image.RGBA, cx, cy, r int, c color.RGBA) {
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy <= r*r && image.Pt(x, y).In(im.Bounds()) {
				im.SetRGBA(x, y, c)
			}
		}
	}
}

func strokeHUDCircle(im *image.RGBA, cx, cy, r, stroke int, c color.RGBA) {
	outer, inner := r*r, (r-stroke)*(r-stroke)
	for y := cy - r; y <= cy+r; y++ {
		for x := cx - r; x <= cx+r; x++ {
			dx, dy := x-cx, y-cy
			d := dx*dx + dy*dy
			if d <= outer && d >= inner && image.Pt(x, y).In(im.Bounds()) {
				im.SetRGBA(x, y, c)
			}
		}
	}
}

func lineHUD(im *image.RGBA, x0, y0, x1, y1, stroke int, c color.RGBA) {
	dx, sx := absInt(x1-x0), 1
	if x0 > x1 {
		sx = -1
	}
	dy, sy := -absInt(y1-y0), 1
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		fillHUDCircle(im, x0, y0, maxHUD(1, stroke/2), c)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func polygonHUD(im *image.RGBA, points [][2]int, stroke int, c color.RGBA) {
	for i := range points {
		next := points[(i+1)%len(points)]
		lineHUD(im, points[i][0], points[i][1], next[0], next[1], stroke, c)
	}
}

func downsampleHUD(src *image.RGBA, width, height int) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			var r, g, b, a uint32
			for sy := 0; sy < hudSupersample; sy++ {
				for sx := 0; sx < hudSupersample; sx++ {
					p := src.RGBAAt(x*hudSupersample+sx, y*hudSupersample+sy)
					r += uint32(p.R) * uint32(p.A)
					g += uint32(p.G) * uint32(p.A)
					b += uint32(p.B) * uint32(p.A)
					a += uint32(p.A)
				}
			}
			if a == 0 {
				continue
			}
			n := uint32(hudSupersample * hudSupersample)
			dst.SetRGBA(x, y, color.RGBA{R: uint8(r / a), G: uint8(g / a), B: uint8(b / a), A: uint8((a + n/2) / n)})
		}
	}
	return dst
}

func minHUD(a, b int) int {
	if a < b {
		return a
	}
	return b
}
func maxHUD(a, b int) int {
	if a > b {
		return a
	}
	return b
}
