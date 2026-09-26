// Package mobile 定義 Modern 高解析畫布與行動裝置 Surface 之間的純 Go 邊界。
//
// 這裡不匯入 Android、Ebiten 或 cgo；平台 adapter 只要把實際 Surface 尺寸、
// density 與安全區交給 Viewport，再把觸控座標轉回 layout 的設計座標即可。
package mobile

import (
	"fmt"
	"math"

	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
)

const minTouchDP = 48

// Insets 是裝置像素中的系統列／瀏海安全區。它不是遊戲設計座標，不能直接
// 當成 renderer 的 24 px safe inset。
type Insets struct {
	Top, Right, Bottom, Left float64
}

// Viewport 是一個固定設計畫布在實際 Surface 中的等比 letterbox。
type Viewport struct {
	SurfaceW, SurfaceH int
	Density            float64
	Insets             Insets
	Design             uilayout.ModernSurface
	OriginX, OriginY   float64
	Scale              float64
}

// NewViewport 建立 Modern 1280×720 的行動裝置 viewport。Surface 必須是橫向
// 或至少能容納設計畫布；直向裝置不會被拉伸，呼叫端可依回傳錯誤顯示旋轉提示。
func NewViewport(surfaceW, surfaceH int, density float64, insets Insets) (Viewport, error) {
	if surfaceW <= 0 || surfaceH <= 0 {
		return Viewport{}, fmt.Errorf("mobile viewport 尺寸必須為正：%dx%d", surfaceW, surfaceH)
	}
	if density <= 0 || math.IsNaN(density) || math.IsInf(density, 0) {
		return Viewport{}, fmt.Errorf("mobile viewport density 必須為正有限值：%v", density)
	}
	if insets.Top < 0 || insets.Right < 0 || insets.Bottom < 0 || insets.Left < 0 {
		return Viewport{}, fmt.Errorf("mobile viewport 安全區不可為負：%+v", insets)
	}
	if insets.Left+insets.Right >= float64(surfaceW) || insets.Top+insets.Bottom >= float64(surfaceH) {
		return Viewport{}, fmt.Errorf("mobile viewport 安全區吃掉 Surface：%+v", insets)
	}
	design := uilayout.ModernDesignSurface()
	safeW := float64(surfaceW) - insets.Left - insets.Right
	safeH := float64(surfaceH) - insets.Top - insets.Bottom
	scale := math.Min(safeW/float64(design.Width), safeH/float64(design.Height))
	if scale <= 0 {
		return Viewport{}, fmt.Errorf("mobile viewport 無可用縮放：%.3fx%.3f", safeW, safeH)
	}
	drawnW := float64(design.Width) * scale
	drawnH := float64(design.Height) * scale
	return Viewport{
		SurfaceW: surfaceW, SurfaceH: surfaceH, Density: density, Insets: insets,
		Design:  design,
		OriginX: insets.Left + (safeW-drawnW)/2,
		OriginY: insets.Top + (safeH-drawnH)/2,
		Scale:   scale,
	}, nil
}

// SurfaceToDesign 將實際觸控／滑鼠座標轉回 Modern 設計座標。系統安全區外、
// letterbox 黑邊與設計畫布外的點一律拒絕。
func (v Viewport) SurfaceToDesign(x, y float64) (int, int, bool) {
	if v.Scale <= 0 || x < v.OriginX || y < v.OriginY ||
		x >= v.OriginX+float64(v.Design.Width)*v.Scale ||
		y >= v.OriginY+float64(v.Design.Height)*v.Scale {
		return 0, 0, false
	}
	dx := int(math.Floor((x - v.OriginX) / v.Scale))
	dy := int(math.Floor((y - v.OriginY) / v.Scale))
	if dx < 0 || dy < 0 || dx >= v.Design.Width || dy >= v.Design.Height {
		return 0, 0, false
	}
	return dx, dy, true
}

// DesignToSurface 將設計座標矩形轉成裝置像素矩形，供平台 debug overlay 或
// 無障礙焦點框使用。回傳值採半開區間，並保證正尺寸。
func (v Viewport) DesignToSurface(r uilayout.Rect) (uilayout.Rect, bool) {
	if v.Scale <= 0 || r.W <= 0 || r.H <= 0 || r.X < 0 || r.Y < 0 ||
		r.Right() > v.Design.Width || r.Bottom() > v.Design.Height {
		return uilayout.Rect{}, false
	}
	x := int(math.Floor(v.OriginX + float64(r.X)*v.Scale))
	y := int(math.Floor(v.OriginY + float64(r.Y)*v.Scale))
	right := int(math.Ceil(v.OriginX + float64(r.Right())*v.Scale))
	bottom := int(math.Ceil(v.OriginY + float64(r.Bottom())*v.Scale))
	return uilayout.Rect{X: x, Y: y, W: right - x, H: bottom - y}, right > x && bottom > y
}

// TouchTarget 擴大設計座標中的命中區，使其在目前 density／scale 下至少達到
// 48 dp。視覺卡片不變；只有平台 adapter 使用這個矩形做觸控命中，避免把桌面
// 1280×720 的像素尺寸誤稱成 Android dp。
func (v Viewport) TouchTarget(r uilayout.Rect) (uilayout.Rect, bool) {
	if r.W <= 0 || r.H <= 0 || v.Scale <= 0 {
		return uilayout.Rect{}, false
	}
	minDesignW := int(math.Ceil(minTouchDP * v.Density / v.Scale))
	minDesignH := int(math.Ceil(minTouchDP * v.Density / v.Scale))
	if minDesignW < r.W {
		minDesignW = r.W
	}
	if minDesignH < r.H {
		minDesignH = r.H
	}
	if minDesignW > v.Design.Width || minDesignH > v.Design.Height {
		return uilayout.Rect{}, false
	}
	x := r.X - (minDesignW-r.W)/2
	y := r.Y - (minDesignH-r.H)/2
	if x < 0 {
		x = 0
	}
	if y < 0 {
		y = 0
	}
	if x+minDesignW > v.Design.Width {
		x = v.Design.Width - minDesignW
	}
	if y+minDesignH > v.Design.Height {
		y = v.Design.Height - minDesignH
	}
	return uilayout.Rect{X: x, Y: y, W: minDesignW, H: minDesignH}, true
}
