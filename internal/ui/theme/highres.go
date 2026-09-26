package theme

import (
	"fmt"
	"image"
	"image/color"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

const highResolutionSamples = 4

// HighTile directly paints terrain geometry on a supersampled canvas. It does
// not call Tile or consume the low-resolution bitmap, so the high-resolution
// presentation remains an independent, program-authored asset path.
func (m *Modern) HighTile(index, width, height int) (*image.RGBA, error) {
	if index < 0 || index >= modernTileCount {
		return nil, fmt.Errorf("theme: 高解析地形索引 %d 超出 0..%d", index, modernTileCount-1)
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("theme: 高解析地形目的尺寸無效：%dx%d", width, height)
	}
	hi := image.NewRGBA(image.Rect(0, 0, width*highResolutionSamples, height*highResolutionSamples))
	base, accent := terrainColors(index)
	fillHighTerrain(hi, index, base, accent)
	return downsampleHUD(hi, width, height), nil
}

// HighRail directly paints the four-way connectivity mask. Index 0 remains a
// valid vertical rail and empty pixels remain transparent.
func (m *Modern) HighRail(index, width, height int) (*image.RGBA, error) {
	if index < 0 || index >= modernRailCount {
		return nil, fmt.Errorf("theme: 高解析鐵路索引 %d 超出 0..%d", index, modernRailCount-1)
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("theme: 高解析鐵路目的尺寸無效：%dx%d", width, height)
	}
	hi := image.NewRGBA(image.Rect(0, 0, width*highResolutionSamples, height*highResolutionSamples))
	ink := color.RGBA{R: 179, G: 61, B: 47, A: 255} // 朱紅運輸線
	sleeper := color.RGBA{R: 75, G: 47, B: 32, A: 255}
	cx, cy := hi.Bounds().Dx()/2, hi.Bounds().Dy()/2
	stroke := maxHUD(highResolutionSamples, minHUD(hi.Bounds().Dx(), hi.Bounds().Dy())/18)
	mask := [...]uint8{0b0101, 0b1010, 0b1010, 0b1001, 0b1001, 0b0011, 0b0011, 0b1100, 0b1100, 0b0110, 0b0110, 0b1011, 0b1011, 0b1110, 0b1110, 0b0111, 0b0111, 0b1101, 0b1101, 0b1111, 0b1111}[index]
	const top, right, bottom, left = 1, 2, 4, 8
	if mask&top != 0 {
		lineHUD(hi, cx, cy, cx, 0, stroke, ink)
		drawSleepers(hi, cx, cy, cx, 0, stroke, sleeper)
	}
	if mask&right != 0 {
		lineHUD(hi, cx, cy, hi.Bounds().Dx()-1, cy, stroke, ink)
		drawSleepers(hi, cx, cy, hi.Bounds().Dx()-1, cy, stroke, sleeper)
	}
	if mask&bottom != 0 {
		lineHUD(hi, cx, cy, cx, hi.Bounds().Dy()-1, stroke, ink)
		drawSleepers(hi, cx, cy, cx, hi.Bounds().Dy()-1, stroke, sleeper)
	}
	if mask&left != 0 {
		lineHUD(hi, cx, cy, 0, cy, stroke, ink)
		drawSleepers(hi, cx, cy, 0, cy, stroke, sleeper)
	}
	return downsampleHUD(hi, width, height), nil
}

// HighUnit directly paints the semantic infantry/armour/cavalry/artillery
// token and its six artillery facings, without reading Unit(index).
func (m *Modern) HighUnit(index, width, height int) (*image.RGBA, error) {
	if index < 0 || index >= modernUnitCount {
		return nil, fmt.Errorf("theme: 高解析部隊索引 %d 超出 0..%d", index, modernUnitCount-1)
	}
	if width <= 0 || height <= 0 {
		return nil, fmt.Errorf("theme: 高解析部隊目的尺寸無效：%dx%d", width, height)
	}
	hi := image.NewRGBA(image.Rect(0, 0, width*highResolutionSamples, height*highResolutionSamples))
	red, kind, facing := false, 0, 1
	switch {
	case index < 2:
		red = index == 1
	case index < 4:
		kind = 1
		red = index == 3
	case index < 6:
		kind = 2
		red = index == 5
	default:
		kind = 3
		red = index >= 12
		facing = (index % 6) + 1
	}
	team := modernPalette()[5]
	if red {
		team = modernPalette()[3]
	}
	ink, highlight := modernPalette()[2], modernPalette()[1]
	w, h := hi.Bounds().Dx(), hi.Bounds().Dy()
	cx, cy := w/2, h/2
	r := minHUD(w, h) * 36 / 100
	fillHighCircle(hi, cx, cy, r, toHighRGBA(ink))
	fillHighCircle(hi, cx, cy, r*86/100, toHighRGBA(team))
	lineHUD(hi, cx-r/2, cy+r/3, cx+r/2, cy+r/3, maxHUD(2, w/32), toHighRGBA(highlight))
	stroke := maxHUD(2, w/28)
	switch kind {
	case 0:
		fillHighRect(hi, cx-r/2, cy-r/2, cx+r/2, cy-r/8, toHighRGBA(highlight))
		lineHUD(hi, cx-r*3/4, cy, cx+r*3/4, cy, stroke, toHighRGBA(ink))
		fillHighRect(hi, cx-r/4, cy, cx+r/4, cy+r/2, toHighRGBA(highlight))
	case 1:
		fillHighRect(hi, cx-r*3/4, cy, cx+r*3/4, cy+r/3, toHighRGBA(ink))
		fillHighRect(hi, cx-r/2, cy-r/5, cx+r/3, cy+r/8, toHighRGBA(highlight))
		lineHUD(hi, cx, cy-r/5, cx+r, cy-r/5, stroke, toHighRGBA(highlight))
	case 2:
		lineHUD(hi, cx-r, cy+r/2, cx-r/4, cy-r/5, stroke, toHighRGBA(highlight))
		lineHUD(hi, cx-r/4, cy-r/5, cx+r/2, cy, stroke, toHighRGBA(highlight))
		lineHUD(hi, cx+r/4, cy, cx+r*3/4, cy-r*2/3, stroke, toHighRGBA(ink))
	case 3:
		fillHighRect(hi, cx-r/2, cy+r/5, cx+r/2, cy+r/2, toHighRGBA(ink))
		fillHighCircle(hi, cx-r/2, cy+r/2, r/5, toHighRGBA(ink))
		fillHighCircle(hi, cx+r/2, cy+r/2, r/5, toHighRGBA(ink))
		dirs := [...]struct{ dx, dy int }{{0, -1}, {1, -1}, {1, 0}, {0, 1}, {-1, 1}, {-1, 0}}
		d := dirs[facing-1]
		lineHUD(hi, cx, cy, cx+d.dx*r, cy+d.dy*r*2/3, stroke, toHighRGBA(highlight))
	}
	return downsampleHUD(hi, width, height), nil
}

func toHighRGBA(p assets.RGB) color.RGBA { return color.RGBA{R: p.R, G: p.G, B: p.B, A: 255} }
func fillHighRect(im *image.RGBA, x0, y0, x1, y1 int, c color.RGBA) {
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			if image.Pt(x, y).In(im.Bounds()) {
				im.SetRGBA(x, y, c)
			}
		}
	}
}
func fillHighCircle(im *image.RGBA, cx, cy, r int, c color.RGBA) { fillHUDCircle(im, cx, cy, r, c) }
func fillHighTerrain(im *image.RGBA, kind int, base, accent byte) {
	w, h := im.Bounds().Dx(), im.Bounds().Dy()
	pal := modernPalette()
	bc, ac := toHighRGBA(pal[base]), toHighRGBA(pal[accent])
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			im.SetRGBA(x, y, bc)
		}
	}
	ink := color.RGBA{R: 75, G: 47, B: 32, A: 255}
	green := color.RGBA{R: 100, G: 126, B: 78, A: 255}
	water := color.RGBA{R: 52, G: 127, B: 145, A: 255}
	cream := color.RGBA{R: 248, G: 232, B: 181, A: 255}
	gold := color.RGBA{R: 177, G: 123, B: 61, A: 255}
	red := color.RGBA{R: 179, G: 61, B: 47, A: 255}
	cx, cy := w/2, h/2
	stroke := maxHUD(highResolutionSamples, minHUD(w, h)/18)
	switch {
	case kind == 0: // 平原農地：低密度短斜線
		for y := h / 5; y < h; y += h / 4 {
			for x := w / 8; x < w; x += w / 4 {
				lineHUD(im, x, y, x+w/10, y-w/10, stroke/2, ac)
			}
		}
	case kind == 1: // 丘陵：兩道等高曲線
		lineHUD(im, w/10, h*3/5, w/3, h/3, stroke, green)
		lineHUD(im, w/3, h/3, w*3/5, h/2, stroke, green)
		lineHUD(im, w*3/5, h/2, w*9/10, h/4, stroke, green)
		lineHUD(im, w/10, h*4/5, w/3, h/2, stroke/2, ink)
		lineHUD(im, w/3, h/2, w*3/5, h*2/3, stroke/2, ink)
	case kind == 2: // 河海：連續波紋
		for y := h / 4; y < h; y += h / 4 {
			lineHUD(im, w/10, y, w/3, y-h/12, stroke/2, water)
			lineHUD(im, w/3, y-h/12, w*3/5, y+h/12, stroke/2, water)
			lineHUD(im, w*3/5, y+h/12, w*9/10, y, stroke/2, water)
		}
	case kind == 3: // 森林：樹冠群
		for _, p := range [][2]int{{w / 4, h / 3}, {w / 2, h / 2}, {w * 3 / 4, h / 3}} {
			fillHighCircle(im, p[0], p[1], w/7, green)
			fillHighCircle(im, p[0]-w/10, p[1]+h/12, w/8, green)
			fillHighRect(im, p[0]-stroke/2, p[1]+h/8, p[0]+stroke/2, h*4/5, ink)
		}
	case kind == 4: // 城市：城樓與屋脊
		fillHighRect(im, w/3, h/2, w*2/3, h*4/5, cream)
		fillHighRect(im, w*2/5, h/3, w*3/5, h/2, red)
		lineHUD(im, w*2/5, h/3, w/2, h/5, stroke, red)
		lineHUD(im, w/2, h/5, w*3/5, h/3, stroke, red)
	case kind == 5: // 高山：尖峰與雪脊
		polygonHUD(im, [][2]int{{w / 10, h * 4 / 5}, {w / 2, h / 5}, {w * 9 / 10, h * 4 / 5}}, stroke, ink)
		lineHUD(im, w/2, h/5, w*3/5, h*2/5, stroke, cream)
		lineHUD(im, w*3/5, h*2/5, w*7/10, h*7/10, stroke, cream)
	case kind == 6: // 沙漠：水平沙丘波紋
		for y := h / 3; y < h; y += h / 5 {
			lineHUD(im, w/10, y, w/3, y-h/10, stroke/2, gold)
			lineHUD(im, w/3, y-h/10, w*2/3, y+h/10, stroke/2, gold)
			lineHUD(im, w*2/3, y+h/10, w*9/10, y, stroke/2, gold)
		}
	case kind == 7 || kind == 8: // 橋：拱形與橋墩
		if kind == 7 {
			lineHUD(im, cx, 0, cx, h, stroke, gold)
			fillHighRect(im, cx-w/8, h/3, cx+w/8, h*2/3, ink)
		} else {
			lineHUD(im, 0, cy, w, cy, stroke, gold)
			fillHighRect(im, w/3, cy-h/8, w*2/3, cy+h/8, ink)
		}
	case kind == 9: // 高原：短脊線
		lineHUD(im, w/10, h*2/3, w/3, h/2, stroke, ac)
		lineHUD(im, w/3, h/2, w*2/3, h*3/5, stroke, ac)
		lineHUD(im, w*2/3, h*3/5, w*9/10, h/3, stroke, ac)
	case kind == 10: // 關口：門樓
		fillHighRect(im, w/4, h/3, w*3/4, h*4/5, gold)
		fillHighRect(im, w*2/5, h/2, w*3/5, h*4/5, bc)
		lineHUD(im, w/4, h/3, w/2, h/5, stroke, ink)
		lineHUD(im, w/2, h/5, w*3/4, h/3, stroke, ink)
	case kind >= 11 && kind <= 20: // 長城十種可辨變體；不宣稱原版逐段命名
		drawWallVariant(im, kind-11, stroke, ink)
	default: // TileKind 22 unknown：中性安全節點，不升格成長城／城市／橋
		fillHighRect(im, w/5, h/3, w*4/5, h*2/3, color.RGBA{R: 233, G: 212, B: 154, A: 255})
		lineHUD(im, w/5, h/3, w*4/5, h/3, stroke/2, ink)
		lineHUD(im, w/5, h*2/3, w*4/5, h*2/3, stroke/2, ink)
	}
}

func drawWallVariant(im *image.RGBA, variant, stroke int, wall color.RGBA) {
	w, h := im.Bounds().Dx(), im.Bounds().Dy()
	mx, my := w/10, h/2
	line := func(x0, y0, x1, y1 int) { lineHUD(im, x0, y0, x1, y1, stroke, wall) }
	block := func(x, y int) { fillHighRect(im, x-stroke, y-stroke*2, x+stroke, y, wall) }
	switch variant {
	case 0: // 水平直段
		line(mx, my, w-mx, my)
		for x := w / 5; x < w; x += w / 4 {
			block(x, my)
		}
	case 1: // 垂直直段
		line(w/2, h/10, w/2, h*9/10)
		for y := h / 4; y < h; y += h / 4 {
			block(w/2, y)
		}
	case 2: // 外轉角（右下）
		line(mx, my, w*3/5, my)
		line(w*3/5, my, w*3/5, h*9/10)
		block(w*3/5, my)
		block(w*3/5, h*3/5)
	case 3: // 內轉角（右上）
		line(mx, h*3/4, w*3/5, h*3/4)
		line(w*3/5, h*3/4, w*3/5, h/10)
		block(w*3/5, h*3/4)
		block(w*3/5, h/2)
	case 4: // 斜坡上行
		line(mx, h*4/5, w*4/5, h/5)
		for i := 1; i < 4; i++ {
			x := mx + (w*7/10)*i/4
			y := h*4/5 - (h*3/5)*i/4
			block(x, y)
		}
	case 5: // 斜坡下行
		line(mx, h/5, w*4/5, h*4/5)
		for i := 1; i < 4; i++ {
			x := mx + (w*7/10)*i/4
			y := h/5 + (h*3/5)*i/4
			block(x, y)
		}
	case 6: // 單端點
		line(mx, my, w*4/5, my)
		block(mx, my)
		fillHighRect(im, w*4/5, my-stroke*2, w*4/5+stroke*2, my+stroke, wall)
	case 7: // 雙端連接
		line(mx, my, w*9/10, my)
		block(mx, my)
		block(w*9/10, my)
	case 8: // 折返髮夾
		line(mx, h/4, w*3/4, h/4)
		line(w*3/4, h/4, w*3/4, h*3/4)
		line(w*3/4, h*3/4, mx, h*3/4)
		block(w*3/4, h/4)
		block(w*3/4, h*3/4)
	case 9: // 門洞段：兩側牆與中央空洞
		line(mx, h/2, w*2/5, h/2)
		line(w*3/5, h/2, w*9/10, h/2)
		block(mx, h/2)
		block(w*9/10, h/2)
		line(w*2/5, h/2-stroke*2, w*2/5, h/2+stroke*2)
		line(w*3/5, h/2-stroke*2, w*3/5, h/2+stroke*2)
	}
}

func drawSleepers(im *image.RGBA, x0, y0, x1, y1, stroke int, c color.RGBA) {
	steps := maxHUD(2, absInt(x1-x0+y1-y0)/8)
	for i := steps; i < steps*8; i += steps {
		x := x0 + (x1-x0)*i/(steps*8)
		y := y0 + (y1-y0)*i/(steps*8)
		if x0 == x1 {
			fillHighRect(im, x-stroke*2, y-stroke, x+stroke*2, y+stroke, c)
		} else {
			fillHighRect(im, x-stroke, y-stroke*2, x+stroke, y+stroke*2, c)
		}
	}
}
