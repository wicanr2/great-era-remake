// Package render 是呈現層的合成核心：把解好的資產畫成一張點陣圖。
//
// **這一層不依賴 Ebiten**（CLAUDE.md §11）。理由不是潔癖——Ebiten 在 init 期
// 就要求顯示器，無頭環境連測都測不了。把合成邏輯留在純 image 這一層，
// 就能在 CI 上逐像素比對原版截圖；Ebiten 只負責把成品貼到視窗。
//
// 原版的結構也是這樣分的：戰鬥模組（docs/re/04）裡沒有任何 Graph unit 呼叫，
// 規則與繪圖本來就是分開的。
package render

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

// 原版的兩種顯示模式（docs/playtest/01-dosbox-probe.md）。
const (
	// ModeBGIW / ModeBGIH 是主遊戲的 BGI 模式：640×350、16 色 planar。
	ModeBGIW, ModeBGIH = 640, 350
	// Mode13W / Mode13H 是開場與標題的 mode 13h：320×200、256 色。
	Mode13W, Mode13H = 320, 200
)

// Canvas 是一張可以疊圖的畫布。
type Canvas struct {
	img *image.RGBA
}

// NewCanvas 開一張指定大小的畫布，初始全黑。
func NewCanvas(w, h int) *Canvas {
	c := &Canvas{img: image.NewRGBA(image.Rect(0, 0, w, h))}
	draw.Draw(c.img, c.img.Bounds(), image.NewUniform(color.RGBA{A: 0xFF}),
		image.Point{}, draw.Src)
	return c
}

// NewBGICanvas 開一張主遊戲尺寸（640×350）的畫布。
func NewBGICanvas() *Canvas { return NewCanvas(ModeBGIW, ModeBGIH) }

// Image 回傳底層的 image.RGBA，供比對或交給 Ebiten。
func (c *Canvas) Image() *image.RGBA { return c.img }

// Bounds 回傳畫布尺寸。
func (c *Canvas) Bounds() image.Rectangle { return c.img.Bounds() }

// DrawBGI 把一張 BGI 16 色圖畫到 (x, y)，不做透明處理。
//
// `.TPC` 沒有透明色的概念——BGI 的 putimage 是整塊覆蓋。
func (c *Canvas) DrawBGI(im *assets.Image, pal assets.Palette, x, y int) error {
	if len(pal) < 16 {
		return fmt.Errorf("render: BGI 需要 16 色調色盤，只有 %d 色", len(pal))
	}
	for iy := 0; iy < im.H; iy++ {
		for ix := 0; ix < im.W; ix++ {
			v := im.Pix[iy*im.W+ix]
			if int(v) >= len(pal) {
				return fmt.Errorf("render: 索引 %d 超出調色盤 %d 色", v, len(pal))
			}
			p := pal[v]
			c.img.SetRGBA(x+ix, y+iy, color.RGBA{p.R, p.G, p.B, 0xFF})
		}
	}
	return nil
}

// DrawSprite 把一張 .GLB 的 256 色圖疊到 (x, y)，索引 0 視為透明。
//
// 索引 0 透明是實測出來的：書法標題就是這樣疊在開場圖上的
// （docs/formats/02-glb-image-library.md §5）。
func (c *Canvas) DrawSprite(sp *assets.Sprite, pal assets.Palette, x, y int) error {
	if len(pal) < 256 {
		return fmt.Errorf("render: 256 色圖需要 256 色調色盤，只有 %d 色", len(pal))
	}
	for iy := 0; iy < sp.H; iy++ {
		for ix := 0; ix < sp.W; ix++ {
			v := sp.Pix[iy*sp.W+ix]
			if v == assets.TransparentIndex {
				continue
			}
			p := pal[v]
			c.img.SetRGBA(x+ix, y+iy, color.RGBA{p.R, p.G, p.B, 0xFF})
		}
	}
	return nil
}

// DiffCount 回傳兩張圖有多少像素不同，用於對照原版截圖。
//
// 尺寸不同直接回錯——比對前必須先確認範圍一致，否則「差異少」沒有意義。
func DiffCount(a, b image.Image) (int, error) {
	ra, rb := a.Bounds(), b.Bounds()
	if ra.Dx() != rb.Dx() || ra.Dy() != rb.Dy() {
		return 0, fmt.Errorf("render: 尺寸不同，%dx%d vs %dx%d",
			ra.Dx(), ra.Dy(), rb.Dx(), rb.Dy())
	}
	var n int
	for y := 0; y < ra.Dy(); y++ {
		for x := 0; x < ra.Dx(); x++ {
			ar, ag, ab, _ := a.At(ra.Min.X+x, ra.Min.Y+y).RGBA()
			br, bg, bb, _ := b.At(rb.Min.X+x, rb.Min.Y+y).RGBA()
			if ar != br || ag != bg || ab != bb {
				n++
			}
		}
	}
	return n, nil
}
