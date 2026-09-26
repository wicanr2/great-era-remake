package render

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"

	"github.com/wicanr2/great-era-remake/internal/assets"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

// drawFilteredIndexedIcon 只平滑專案自行繪製的小型 HUD 圖示；地圖圖塊與
// 玩家原版資料產生的執行期裝飾仍保留像素遮罩，不進入本路徑。
func drawFilteredIndexedIcon(c *Canvas, b uitheme.Bitmap, dst uilayout.Rect) error {
	if !b.Valid() || dst.W <= 0 || dst.H <= 0 {
		return nil
	}
	for y := 0; y < dst.H; y++ {
		v := (float64(y)+0.5)*float64(b.Image.H)/float64(dst.H) - 0.5
		for x := 0; x < dst.W; x++ {
			u := (float64(x)+0.5)*float64(b.Image.W)/float64(dst.W) - 0.5
			col, ok := filteredIndexedPixel(b, u, v)
			if ok {
				c.blendPixel(dst.X+x, dst.Y+y,
					assets.RGB{R: col.R, G: col.G, B: col.B}, col.A)
			}
		}
	}
	return nil
}

func drawModernCommandIcon(c *Canvas, provider uitheme.HUDIconProvider, index int, dst uilayout.Rect) error {
	if high, ok := provider.(uitheme.HighResolutionHUDIconProvider); ok {
		icon, err := high.CommandIconRGBA(index, dst.W, dst.H)
		if err != nil {
			return err
		}
		return drawRGBAIcon(c, icon, dst)
	}
	icon, err := provider.CommandIcon(index)
	if err != nil {
		return err
	}
	return drawFilteredIndexedIcon(c, icon, dst)
}

func drawModernResourceIcon(c *Canvas, provider uitheme.HUDIconProvider, index int, dst uilayout.Rect) error {
	if high, ok := provider.(uitheme.HighResolutionHUDIconProvider); ok {
		icon, err := high.ResourceIconRGBA(index, dst.W, dst.H)
		if err != nil {
			return err
		}
		return drawRGBAIcon(c, icon, dst)
	}
	icon, err := provider.ResourceIcon(index)
	if err != nil {
		return err
	}
	return drawFilteredIndexedIcon(c, icon, dst)
}

func drawRGBAIcon(c *Canvas, icon *image.RGBA, dst uilayout.Rect) error {
	if c == nil || icon == nil || icon.Bounds().Dx() != dst.W || icon.Bounds().Dy() != dst.H {
		return fmt.Errorf("Modern high-resolution HUD 圖示尺寸不符：image=%v dst=%+v", icon, dst)
	}
	bounds := icon.Bounds()
	for y := 0; y < dst.H; y++ {
		for x := 0; x < dst.W; x++ {
			p := icon.RGBAAt(bounds.Min.X+x, bounds.Min.Y+y)
			c.blendPixel(dst.X+x, dst.Y+y, assets.RGB{R: p.R, G: p.G, B: p.B}, p.A)
		}
	}
	return nil
}

func drawModernTerrain(c *Canvas, provider uitheme.Theme, index int, dst uilayout.Rect) error {
	if high, ok := provider.(uitheme.HighResolutionTheme); ok {
		im, err := high.HighTile(index, dst.W, dst.H)
		if err != nil {
			return err
		}
		return drawRGBAIcon(c, im, dst)
	}
	b, err := provider.Tile(index)
	if err != nil {
		return err
	}
	return drawIndexedScaled(c, b, dst, false)
}

func drawModernRail(c *Canvas, provider uitheme.Theme, index int, dst uilayout.Rect) error {
	if high, ok := provider.(uitheme.HighResolutionTheme); ok {
		im, err := high.HighRail(index, dst.W, dst.H)
		if err != nil {
			return err
		}
		return drawRGBAIcon(c, im, dst)
	}
	b, err := provider.Rail(index)
	if err != nil {
		return err
	}
	return drawIndexedScaled(c, b, dst, true)
}

func drawModernUnit(c *Canvas, provider uitheme.UnitProvider, index int, dst uilayout.Rect) error {
	if high, ok := provider.(uitheme.HighResolutionUnitProvider); ok {
		im, err := high.HighUnit(index, dst.W, dst.H)
		if err != nil {
			return err
		}
		return drawRGBAIcon(c, im, dst)
	}
	b, err := provider.Unit(index)
	if err != nil {
		return err
	}
	return drawIndexedScaled(c, b, dst, true)
}

func filteredIndexedPixel(b uitheme.Bitmap, x, y float64) (color.RGBA, bool) {
	// 四點雙線性取樣保留索引 0 的透明語意，只在 RGBA 畫布中柔化程式化
	// 幾何邊緣；不寫出任何衍生資產。
	x0, y0 := int(math.Floor(x)), int(math.Floor(y))
	fx, fy := x-float64(x0), y-float64(y0)
	var rr, gg, bb, aa float64
	for j := 0; j < 2; j++ {
		for i := 0; i < 2; i++ {
			sx, sy := x0+i, y0+j
			if sx < 0 {
				sx = 0
			}
			if sy < 0 {
				sy = 0
			}
			if sx >= b.Image.W {
				sx = b.Image.W - 1
			}
			if sy >= b.Image.H {
				sy = b.Image.H - 1
			}
			idx := b.Image.Pix[sy*b.Image.W+sx]
			if int(idx) >= len(b.Palette) {
				continue
			}
			w := (1 - fx)
			if i == 1 {
				w = fx
			}
			wy := (1 - fy)
			if j == 1 {
				wy = fy
			}
			w *= wy
			if idx == 0 {
				continue
			}
			p := b.Palette[idx]
			rr += float64(p.R) * w
			gg += float64(p.G) * w
			bb += float64(p.B) * w
			aa += 255 * w
		}
	}
	if aa == 0 {
		return color.RGBA{}, false
	}
	return color.RGBA{R: uint8(rr*255/aa + 0.5), G: uint8(gg*255/aa + 0.5), B: uint8(bb*255/aa + 0.5), A: uint8(aa + 0.5)}, true
}

// drawSampledImage 放大照片時使用雙線性取樣，縮小時使用面積平均。結果只
// 寫進目前畫布，不保存、打包或散布人物照片的衍生檔。
func drawSampledImage(dstImg *image.RGBA, dst uilayout.Rect, src image.Image, crop image.Rectangle) {
	if dstImg == nil || src == nil || dst.W <= 0 || dst.H <= 0 || crop.Empty() {
		return
	}
	// image.Uniform 的 Bounds 是刻意設成極大值；若把它當一般照片做面積
	// 積分會產生無意義的巨量迴圈。測試與純色 fallback 直接填滿即可。
	if uniform, ok := src.(*image.Uniform); ok {
		draw.Draw(dstImg, image.Rect(dst.X, dst.Y, dst.Right(), dst.Bottom()), uniform, image.Point{}, draw.Src)
		return
	}
	sw, sh := crop.Dx(), crop.Dy()
	for y := 0; y < dst.H; y++ {
		for x := 0; x < dst.W; x++ {
			x0 := float64(crop.Min.X) + float64(x)*float64(sw)/float64(dst.W)
			x1 := float64(crop.Min.X) + float64(x+1)*float64(sw)/float64(dst.W)
			y0 := float64(crop.Min.Y) + float64(y)*float64(sh)/float64(dst.H)
			y1 := float64(crop.Min.Y) + float64(y+1)*float64(sh)/float64(dst.H)
			var c color.RGBA
			if sw > dst.W || sh > dst.H {
				c = areaSample(src, x0, y0, x1, y1)
			} else {
				c = bilinearSample(src, (x0+x1)/2, (y0+y1)/2)
			}
			dstImg.Set(dst.X+x, dst.Y+y, c)
		}
	}
}

func rgbaAt(src image.Image, x, y int) color.RGBA {
	r, g, b, a := src.At(x, y).RGBA()
	return color.RGBA{uint8(r >> 8), uint8(g >> 8), uint8(b >> 8), uint8(a >> 8)}
}
func bilinearSample(src image.Image, x, y float64) color.RGBA {
	b := src.Bounds()
	x0, y0 := int(math.Floor(x)), int(math.Floor(y))
	fx, fy := x-float64(x0), y-float64(y0)
	var rr, gg, bb, aa float64
	for j := 0; j < 2; j++ {
		for i := 0; i < 2; i++ {
			sx, sy := x0+i, y0+j
			if sx < b.Min.X {
				sx = b.Min.X
			}
			if sy < b.Min.Y {
				sy = b.Min.Y
			}
			if sx >= b.Max.X {
				sx = b.Max.X - 1
			}
			if sy >= b.Max.Y {
				sy = b.Max.Y - 1
			}
			p := rgbaAt(src, sx, sy)
			w := (1 - fx)
			if i == 1 {
				w = fx
			}
			wy := (1 - fy)
			if j == 1 {
				wy = fy
			}
			w *= wy
			rr += float64(p.R) * w
			gg += float64(p.G) * w
			bb += float64(p.B) * w
			aa += float64(p.A) * w
		}
	}
	return color.RGBA{uint8(rr + .5), uint8(gg + .5), uint8(bb + .5), uint8(aa + .5)}
}
func areaSample(src image.Image, x0, y0, x1, y1 float64) color.RGBA {
	b := src.Bounds()
	ix0, iy0 := int(math.Floor(x0)), int(math.Floor(y0))
	ix1, iy1 := int(math.Ceil(x1)), int(math.Ceil(y1))
	var rr, gg, bb, aa, weight float64
	for y := iy0; y < iy1; y++ {
		for x := ix0; x < ix1; x++ {
			ox := math.Max(0, math.Min(x1, float64(x+1))-math.Max(x0, float64(x)))
			oy := math.Max(0, math.Min(y1, float64(y+1))-math.Max(y0, float64(y)))
			w := ox * oy
			if w == 0 {
				continue
			}
			sx, sy := x, y
			if sx < b.Min.X {
				sx = b.Min.X
			}
			if sy < b.Min.Y {
				sy = b.Min.Y
			}
			if sx >= b.Max.X {
				sx = b.Max.X - 1
			}
			if sy >= b.Max.Y {
				sy = b.Max.Y - 1
			}
			p := rgbaAt(src, sx, sy)
			rr += float64(p.R) * w
			gg += float64(p.G) * w
			bb += float64(p.B) * w
			aa += float64(p.A) * w
			weight += w
		}
	}
	if weight == 0 {
		return color.RGBA{}
	}
	return color.RGBA{uint8(rr/weight + .5), uint8(gg/weight + .5), uint8(bb/weight + .5), uint8(aa/weight + .5)}
}
