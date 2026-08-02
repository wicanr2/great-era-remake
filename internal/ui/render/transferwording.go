package render

import (
	"github.com/wicanr2/great-era-remake/internal/assets"
)

// DrawPlainTransferMode 畫現代白話的調動方式頁。這條路徑只換顯示文字，
// 數字鍵與狀態機仍由 cmd/dsds 的既有流程處理。
func (c *Canvas) DrawPlainTransferMode(fonts *assets.EtenFonts, fg, bg assets.RGB,
	x, y, w, h int, prompt, partial, all string) []rune {
	c.fillRect(x, y, w, h, bg)
	missing := c.DrawSemanticText(fonts, prompt, fg, x+24, y+20)
	for i, label := range []string{partial, all} {
		cy := y + 72 + i*46
		c.DrawNumber(uint32(i+1), fg, x+42, cy)
		missing = append(missing, c.DrawSemanticText(fonts, label, fg, x+74, cy)...)
	}
	return uniqueRunes(missing)
}

// DrawPlainTransferTarget 畫白話目標提示；省份候選仍使用原版姓名字模，
// 因為 wording 軸不應暗中改動姓名或圖形主題。
func (c *Canvas) DrawPlainTransferTarget(fonts *assets.EtenFonts, provinces *assets.GlyphFile,
	fg, bg assets.RGB, x, y, w, h int, prompt string, candidates []int, input uint32) ([]rune, error) {
	c.fillRect(x, y, w, h, bg)
	missing := c.DrawSemanticText(fonts, prompt, fg, x+24, y+20)
	c.DrawNumber(input, fg, x+w-28, y+20)
	for i, id := range candidates {
		cy := y + 64 + i*26
		c.DrawNumber(uint32(id), fg, x+44, cy)
		if id >= 1 {
			if err := c.DrawEntry(provinces, id-1, 3, fg, x+100, cy, true); err != nil {
				return nil, err
			}
		}
	}
	return uniqueRunes(missing), nil
}

// DrawPlainTransferSelection 畫白話選將提示與確認說明；候選、游標、勾選狀態
// 和原典路徑逐項共用，不能因文字較易懂而增加名單資訊。
func (c *Canvas) DrawPlainTransferSelection(fonts *assets.EtenFonts, generals *assets.GlyphFile,
	fg, bg assets.RGB, x, y, w, h int, prompt, confirm string, ids []int, selected []bool,
	cursor, page int) ([]rune, error) {
	c.fillRect(x, y, w, h, bg)
	missing := c.DrawSemanticText(fonts, prompt, fg, x+24, y+10)
	start := page * 20
	end := start + 20
	if end > len(ids) {
		end = len(ids)
	}
	for i := start; i < end; i++ {
		row := i - start
		cx := x + 26 + (row/10)*205
		cy := y + 50 + (row%10)*24
		if i == cursor {
			c.strokeRect(cx-8, cy-3, 190, 20, fg)
		}
		if i < len(selected) && selected[i] {
			c.fillRect(cx, cy+4, 7, 7, fg)
		}
		c.DrawNumber(uint32(i+1), fg, cx+34, cy)
		if ids[i] > 0 {
			if err := c.DrawEntry(generals, ids[i]-1, 3, fg, cx+48, cy, true); err != nil {
				return nil, err
			}
		}
	}
	missing = append(missing, c.DrawSemanticText(fonts, confirm, fg, x+24, y+h-30)...)
	return uniqueRunes(missing), nil
}

// DrawPlainTransferAmount 畫目前資源的完整白話問句與輸入值。
func (c *Canvas) DrawPlainTransferAmount(fonts *assets.EtenFonts, fg, bg assets.RGB,
	x, y, w, h int, prompt string, amount uint32) []rune {
	c.fillRect(x, y, w, h, bg)
	missing := c.DrawSemanticText(fonts, prompt, fg, x+24, y+34)
	c.DrawNumber(amount, fg, x+36, y+82)
	return uniqueRunes(missing)
}

// DrawPlainConfirm 畫白話確認頁。Y/N 仍是既有輸入，中文只說明按鍵語意。
func (c *Canvas) DrawPlainConfirm(fonts *assets.EtenFonts, fg, bg assets.RGB,
	x, y, w, h int, prompt string) []rune {
	c.fillRect(x, y, w, h, bg)
	boxX, boxY, boxW, boxH := x+58, y+104, w-116, 106
	c.strokeRect(boxX, boxY, boxW, boxH, fg)
	missing := c.DrawSemanticText(fonts, prompt, fg, boxX+24, boxY+18)
	missing = append(missing, c.DrawSemanticText(fonts, "Y 確定　N 返回", fg,
		boxX+24, boxY+60)...)
	return uniqueRunes(missing)
}

func uniqueRunes(in []rune) []rune {
	seen := map[rune]bool{}
	out := make([]rune, 0, len(in))
	for _, r := range in {
		if !seen[r] {
			seen[r] = true
			out = append(out, r)
		}
	}
	return out
}
