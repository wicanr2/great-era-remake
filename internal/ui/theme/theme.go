// Package theme 定義呈現主題與 legacy 圖像之間的窄邊界。
//
// 這個 package 不依賴 Ebiten，也不依賴 render；因此可以在無頭測試中驗證
// 主題資產的尺寸、索引與透明規則。render 只消費 Theme，不知道資產來自
// 原版 TPC、現代程式化圖塊或未來的可再散布圖片。
package theme

import (
	"fmt"
	"image"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

// Mode 是玩家可選的主題 preset。Modern provider 替換戰場地形／鐵路、部隊／HUD
// 圖示與 1280×720 高解析 UI 外殼；高解析文字的可再散布比例字型仍是獨立發行 gate。
type Mode string

const (
	ModeRetro  Mode = "retro"
	ModeModern Mode = "modern"
)

// ParseMode 驗證命令列與偏好檔的主題值。
func ParseMode(raw string) (Mode, error) {
	if raw == "" {
		return ModeRetro, nil
	}
	switch Mode(raw) {
	case ModeRetro, ModeModern:
		return Mode(raw), nil
	default:
		return "", fmt.Errorf("theme: 未知主題 %q（只接受 retro 或 modern）", raw)
	}
}

// Bitmap 是主題的一張索引圖。鐵路圖塊內的像素索引 0 代表透明；鐵路圖塊
// 本身的索引 0 仍是有效的直線圖。地形圖塊則是完整覆蓋。Palette 與 Image
// 一起回傳，避免把 modern 的顏色錯套到 retro 圖上。
type Bitmap struct {
	Image   *assets.Image
	Palette assets.Palette
}

// Valid 檢查基本圖像不變量；更細的尺寸由 renderer 依資產種類檢查。
func (b Bitmap) Valid() bool {
	return b.Image != nil && b.Image.W > 0 && b.Image.H > 0 &&
		len(b.Image.Pix) == b.Image.W*b.Image.H && len(b.Palette) >= 16
}

// Theme 是戰場呈現所需的最小介面。index 全部採 0-based：
// terrain 0..21、rail 0..20，這正是 NWMAP 解碼後的索引。
type Theme interface {
	Name() string
	Tile(index int) (Bitmap, error)
	Rail(index int) (Bitmap, error)
}

// HighResolutionTheme 是 Modern 高解析戰場的可選圖片 provider。它不取代
// Theme：既有 32×24 indexed 圖塊仍供 retro 與舊 renderer 使用；高解析呼叫端
// 可要求任意目的尺寸，實作必須以純 Go 產生僅存於記憶體的 RGBA 圖像。
type HighResolutionTheme interface {
	HighTile(index, width, height int) (*image.RGBA, error)
	HighRail(index, width, height int) (*image.RGBA, error)
}

// UIStyle 是主題的完整外殼契約。它只描述顏色、語意強調色與排版密度，
// 不攜帶規則狀態；因此切換主題時可以原子交換 provider，而不會改動存檔、
// 亂數或命令流程。未提供 StyleProvider 的舊主題會由呼叫端採用 RetroStyle。
type UIStyle struct {
	Name        Mode
	Ink         assets.RGB
	Paper       assets.RGB
	Panel       assets.RGB
	Muted       assets.RGB
	Accent      assets.RGB
	AccentAlt   assets.RGB
	Focus       assets.RGB
	FactionTint [10]assets.RGB
	// Ornaments 是玩家原版資料於執行期解碼出的可選邊飾遮罩。nil 時 renderer
	// 必須畫可再散布的程式化 fallback；不得讓缺少裝飾破壞完整 theme。
	Ornaments *OriginalOrnaments
}

// OriginalOrnaments 只保存原版 BGI 的索引遮罩，不會寫檔或進入 release。
// MENU1/3 是左右邊、MENU2/4 是上下邊，MENU5 是中央飾帶；WARMENU 是面板框。
type OriginalOrnaments struct {
	ButtonLeft, ButtonTop, ButtonRight, ButtonBottom, ButtonBand *assets.Image
	PanelFrame                                                   *assets.Image
}

// Valid 驗證 A2 裝飾組必須原子完整，避免顯示半套原版邊框。
func (o *OriginalOrnaments) Valid() bool {
	if o == nil {
		return false
	}
	for _, im := range []*assets.Image{o.ButtonLeft, o.ButtonTop, o.ButtonRight,
		o.ButtonBottom, o.ButtonBand, o.PanelFrame} {
		if im == nil || im.W <= 0 || im.H <= 0 || len(im.Pix) != im.W*im.H {
			return false
		}
	}
	return true
}

// DecodeOriginalOrnaments 從呼叫端提供的唯讀原版檔案建立執行期遮罩。
// read 的檔名契約固定，讓 cmd/dsds 與 cmd/screenshot 共用相同 fail-closed 行為。
func DecodeOriginalOrnaments(read func(string) ([]byte, error)) (*OriginalOrnaments, error) {
	if read == nil {
		return nil, fmt.Errorf("theme: 原版裝飾 reader 不可為 nil")
	}
	names := []string{"MENU1.TPC", "MENU2.TPC", "MENU3.TPC", "MENU4.TPC", "MENU5.TPC", "WARMENU.TPC"}
	images := make([]*assets.Image, len(names))
	for i, name := range names {
		data, err := read(name)
		if err != nil {
			return nil, fmt.Errorf("theme: 讀取原版裝飾 %s: %w", name, err)
		}
		images[i], err = assets.DecodeBGI(data, 0)
		if err != nil {
			return nil, fmt.Errorf("theme: 解碼原版裝飾 %s: %w", name, err)
		}
	}
	o := &OriginalOrnaments{ButtonLeft: images[0], ButtonTop: images[1], ButtonRight: images[2],
		ButtonBottom: images[3], ButtonBand: images[4], PanelFrame: images[5]}
	if !o.Valid() {
		return nil, fmt.Errorf("theme: 原版裝飾組不完整")
	}
	return o, nil
}

// RetroStyle 保留原版米黃／暗紅外殼，供 legacy renderer 與測試明確使用。
func RetroStyle() UIStyle {
	return UIStyle{
		Name:      ModeRetro,
		Ink:       assets.RGB{R: 0xAE, G: 0x00, B: 0x00},
		Paper:     assets.RGB{R: 0xFF, G: 0xFF, B: 0xA2},
		Panel:     assets.RGB{R: 0xFF, G: 0xFF, B: 0xA2},
		Muted:     assets.RGB{R: 0x80, G: 0x5A, B: 0x52},
		Accent:    assets.RGB{R: 0x00, G: 0x00, B: 0xAA},
		AccentAlt: assets.RGB{R: 0xAA, G: 0x00, B: 0x00},
		Focus:     assets.RGB{R: 0x00, G: 0x00, B: 0xAA},
	}
}

// StyleProvider 是完整主題可選的 UI 外殼介面。保留 optional 形式，讓既有
// retro adapter 與最小測試 provider 不必在同一輪被迫改寫。
type StyleProvider interface {
	Style() UIStyle
}

// UnitProvider 是 P2 部隊圖示的可選呈現介面。index 沿用 NEWICON.TPC 的
// 0-based 18 張圖，讓兵種／攻守／砲兵朝向的唯一對照仍留在
// `render.BranchIcon`；主題只替換圖像，不複製規則層的兵種判斷。
//
// P1 的 Theme 不強迫所有 provider 立刻提供部隊圖示；在 icon 尚未載入時，
// 呼叫端必須 fail-closed，而不是靜默拿另一個索引頂替。
type UnitProvider interface {
	Unit(index int) (Bitmap, error)
}

// HighResolutionUnitProvider 是 Modern 高解析部隊圖示的可選 provider。
// index 仍維持原有 18 張圖的語意與順序；目的尺寸由 renderer 決定。
type HighResolutionUnitProvider interface {
	HighUnit(index, width, height int) (*image.RGBA, error)
}

// HUDIconProvider 是 P3 modern HUD 的輔助圖示介面。索引是穩定的呈現契約：
// 資源 0..5 對應黃金／糧食／彈藥／燃料／煤礦／鐵礦，指令 0..14 對應政略指令
// 1..15。provider 只提供圖像，不複製規則或語系文字。
type HUDIconProvider interface {
	ResourceIcon(index int) (Bitmap, error)
	CommandIcon(index int) (Bitmap, error)
}

// HighResolutionHUDIconProvider 是 Modern 高解析畫面的可選介面。既有
// HUDIconProvider 仍保留 16×16 契約給原解析度；支援者可依實際目的尺寸
// 直接產生帶 alpha 的程式化圖示，避免把低解析點陣任意放大。
type HighResolutionHUDIconProvider interface {
	ResourceIconRGBA(index, width, height int) (*image.RGBA, error)
	CommandIconRGBA(index, width, height int) (*image.RGBA, error)
}

const (
	ResourceIconCount = 6
	CommandIconCount  = 15
	HUDIconW          = 16
	HUDIconH          = 16
)
