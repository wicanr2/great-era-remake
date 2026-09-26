package render

import (
	"fmt"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

const narrativePerPage = 4

// DrawModernShell 在地圖／子畫面工作區外沿畫出一致的 modern 卡片邊界。
// 邏輯解析度與內容矩形不變，只增加可見層級；因此不會改變滑鼠／觸控命中。
func (c *Canvas) DrawModernShell(style uitheme.UIStyle, x, y, w, h int) {
	if style.Name != uitheme.ModeModern {
		return
	}
	c.strokeRect(x, y, w, h, style.Muted)
	c.fillRect(x, y, w, 2, style.Accent)
	c.fillRect(x, y+h-2, w, 2, style.AccentAlt)
}

// DrawNarrativeButton 是地圖上的新聞／史事入口。它只開啟資料瀏覽器，
// 不觸發規則或改寫存檔；命中幾何由 layout.NarrativeButton 共用。
func (c *Canvas) DrawNarrativeButton(fonts *assets.EtenFonts, label string,
	fg, bg assets.RGB, logicalWidth int) []rune {
	p := uilayout.NarrativeButton(logicalWidth)
	c.fillRect(p.X, p.Y, p.HitW, p.HitH, bg)
	c.strokeRect(p.X, p.Y, p.HitW, p.HitH, fg)
	if fonts == nil {
		// 沒有外部字庫時仍保留可見入口：N 的幾何標記不是原版資產。
		c.fillRect(p.X+16, p.Y+15, 4, 18, fg)
		c.fillRect(p.X+34, p.Y+15, 4, 18, fg)
		c.fillRect(p.X+20, p.Y+17, 14, 4, fg)
		c.fillRect(p.X+20, p.Y+27, 14, 4, fg)
		return nil
	}
	return c.DrawSemanticText(fonts, label, fg, p.X+12, p.Y+16)
}

// NarrativePageCount 回傳頁數；沒有來源圖像時回 0。
func NarrativePageCount(images []*assets.Image) int {
	if len(images) == 0 {
		return 0
	}
	return (len(images) + narrativePerPage - 1) / narrativePerPage
}

// DrawNarrativeGallery 顯示 NEWSDATA.DAT 的原版新聞圖像與已證實的語系
// 模板。未知文字只顯示 locale 的 fallback，保留圖像證據，不做猜測 OCR。
func (c *Canvas) DrawNarrativeGallery(fonts *assets.EtenFonts, catalog *i18n.NarrativeCatalog,
	images []*assets.Image, page int, title string, mode i18n.WordingMode,
	style uitheme.UIStyle) error {
	if catalog == nil || len(images) == 0 {
		return fmt.Errorf("敘事圖庫缺少 catalog 或 NEWSDATA 圖像")
	}
	pages := NarrativePageCount(images)
	if page < 0 || page >= pages {
		return fmt.Errorf("敘事頁 %d 超出 1..%d", page+1, pages)
	}
	if style.Name == "" {
		style = uitheme.RetroStyle()
	}
	c.fillRect(0, 0, ModeBGIW, ModeBGIH, style.Paper)
	c.fillRect(12, 8, ModeBGIW-24, 28, style.Panel)
	c.strokeRect(12, 8, ModeBGIW-24, 28, style.Accent)
	if fonts != nil {
		missing := c.DrawSemanticText(fonts, title, style.Ink, 24, 15)
		if len(missing) != 0 {
			return fmt.Errorf("敘事標題缺字：%q", string(missing))
		}
	}
	start := page * narrativePerPage
	end := start + narrativePerPage
	if end > len(images) {
		end = len(images)
	}
	for i := start; i < end; i++ {
		y := 48 + (i-start)*62
		c.fillRect(16, y-4, 608, 54, style.Panel)
		c.strokeRect(16, y-4, 608, 54, style.Muted)
		if err := c.DrawBGI(images[i], assets.EGADefaultPalette, 24, y); err != nil {
			return err
		}
		c.DrawSmallNumber(uint32(i), style.Accent, 250, y+2)
		text, status, ok := catalog.Text(i, mode)
		if !ok {
			text = catalog.Fallback
		}
		if fonts != nil {
			missing := c.DrawSemanticText(fonts, text, style.Ink, 278, y+2)
			if len(missing) != 0 && status == i18n.NarrativeTranslated {
				return fmt.Errorf("敘事 #%d 缺字：%q", i, string(missing))
			}
		}
	}
	if fonts != nil {
		footer := fmt.Sprintf("%d/%d　%s", page+1, pages, catalog.Source)
		if missing := c.DrawSemanticText(fonts, footer, style.Muted, 24, 326); len(missing) != 0 {
			return fmt.Errorf("敘事頁尾缺字：%q", string(missing))
		}
	}
	return nil
}
