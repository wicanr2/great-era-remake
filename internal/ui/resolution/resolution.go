// Package resolution 定義 remake 外殼的兩種顯示畫布。
//
// 原版模式保留 640×350 的 BGI 設計座標；高解析模式以 1280×720、16:9
// 作為 Modern／Android 的第一版設計畫布。這個 package 不依賴 Ebiten，
// 讓偏好、座標換算與無頭測試可以共用。
package resolution

import "fmt"

// Mode 是玩家顯示解析度偏好。它只影響呈現與輸入座標，不進遊戲存檔。
type Mode string

const (
	ModeOriginal Mode = "original"
	ModeHigh     Mode = "high"
)

const (
	OriginalWidth  = 640
	OriginalHeight = 350
	HighWidth      = 1280
	HighHeight     = 720
)

// Parse 驗證命令列或 prefs.json 的解析度值。
func Parse(raw string) (Mode, error) {
	switch Mode(raw) {
	case ModeOriginal, ModeHigh:
		return Mode(raw), nil
	default:
		return "", fmt.Errorf("resolution: 未知顯示解析度 %q（應為 original 或 high）", raw)
	}
}

// Toggle 在原版與高解析畫布間循環。
func Toggle(mode Mode) Mode {
	if mode == ModeHigh {
		return ModeOriginal
	}
	return ModeHigh
}

// Size 回傳該模式的設計畫布尺寸。
func (m Mode) Size() (int, int) {
	if m == ModeHigh {
		return HighWidth, HighHeight
	}
	return OriginalWidth, OriginalHeight
}

// BaseToSurface 將 640×350 原版內容以等比方式放進指定的設計畫布。
// 高解析 1280×720 會留下上下各 10 像素安全邊界；原版模式則是原座標。
func BaseToSurface(x, y, surfaceW, surfaceH int) (int, int, int, int, bool) {
	if surfaceW <= 0 || surfaceH <= 0 {
		return 0, 0, 0, 0, false
	}
	scaleX := float64(surfaceW) / float64(OriginalWidth)
	scaleY := float64(surfaceH) / float64(OriginalHeight)
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}
	drawnW := int(float64(OriginalWidth) * scale)
	drawnH := int(float64(OriginalHeight) * scale)
	offX := (surfaceW - drawnW) / 2
	offY := (surfaceH - drawnH) / 2
	return offX, offY, drawnW, drawnH, drawnW > 0 && drawnH > 0
}

// SurfaceToBase 把高解析設計畫布的座標反算回現有 640×350 命中座標。
// 黑邊／安全邊界回傳 false；避免高解析過渡期把點擊送到錯誤按鈕。
func SurfaceToBase(x, y, surfaceW, surfaceH int) (int, int, bool) {
	offX, offY, drawnW, drawnH, ok := BaseToSurface(0, 0, surfaceW, surfaceH)
	if !ok || x < offX || y < offY || x >= offX+drawnW || y >= offY+drawnH {
		return 0, 0, false
	}
	scale := float64(drawnW) / float64(OriginalWidth)
	if scale <= 0 {
		return 0, 0, false
	}
	return int(float64(x-offX) / scale), int(float64(y-offY) / scale), true
}
