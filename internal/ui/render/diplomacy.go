package render

import (
	"fmt"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

// DrawDiplomacyMenu 畫原版外交子選單的三個已在 DOSBox 實測看見的項目。
//
// 文字由 wording catalog 提供，讓原典／現代白話共用同一個版面與命中契約；
// 這是呈現層切換，不改外交規則或帳本。
func (c *Canvas) DrawDiplomacyMenu(fonts *assets.EtenFonts, fg, bg assets.RGB,
	x, y, w, h int, labels []string) []rune {
	c.fillRect(x, y, w, h, bg)
	missing := []rune{}
	for i, label := range labels {
		cy := y + 38 + i*42
		c.DrawNumber(uint32(i+1), fg, x+42, cy)
		missing = append(missing, c.DrawSemanticText(fonts, label, fg, x+82, cy)...)
	}
	return uniqueRunes(missing)
}

// DrawDiplomacyLoan 畫貸款額度與信用度。額度欄沿用數字輸入畫面的右側
// 位置，觸控數字鍵盤因此可以與所有其他數字頁共用。
func (c *Canvas) DrawDiplomacyLoan(fonts *assets.EtenFonts, fg, bg assets.RGB,
	x, y, w, h int, prompt, creditLabel string, credit uint8, amount uint32) []rune {
	return c.DrawDiplomacyAmount(fonts, fg, bg, x, y, w, h, prompt,
		fmt.Sprintf("%s %d", creditLabel, credit), amount)
}

// DrawDiplomacyAmount 畫外交額度輸入頁的共用版面。第二行由呼叫端提供
// 「信用度」或「目前外債／黃金」等資訊，讓貸款與償債共用同一組數字鍵盤
// 與命中幾何；規則層不依賴這個 renderer。
func (c *Canvas) DrawDiplomacyAmount(fonts *assets.EtenFonts, fg, bg assets.RGB,
	x, y, w, h int, prompt, detail string, amount uint32) []rune {
	c.fillRect(x, y, w, h, bg)
	missing := c.DrawSemanticText(fonts, prompt, fg, x+28, y+34)
	missing = append(missing, c.DrawSemanticText(fonts, detail, fg, x+28, y+68)...)
	c.DrawNumber(amount, fg, x+w-42, y+104)
	return uniqueRunes(missing)
}

// DrawCeasefireTarget 畫原版停火玩家輸入提示。單字「欲／在／？」沿用
// `1.15`，其餘沿用 `2.15`，保留 IDA `sub_211D5` 的詞條順序；數字輸入
// 仍由共用數字鍵盤提供。
func (c *Canvas) DrawCeasefireTarget(fonts PanelFonts, fg, bg assets.RGB,
	x, y, w, h int, amount uint32) error {
	c.fillRect(x, y, w, h, bg)
	type part struct {
		font      *assets.GlyphFile
		entry     int
		slotWidth int
	}
	parts := []part{
		{fonts.W2, 21, 2},  // 司令（2.15 #22）
		{fonts.W1, 30, 1},  // 欲（1.15 #31）
		{fonts.W1, 55, 1},  // 在（1.15 #56）
		{fonts.W2, 39, 2},  // 何省（2.15 #40）
		{fonts.W2, 9, 2},   // 談判（2.15 #10）
		{fonts.W2, 149, 2}, // 停火（2.15 #150）
		{fonts.W1, 17, 1},  // ？（1.15 #18）
	}
	wx := x + 24
	for _, p := range parts {
		if err := c.DrawEntry(p.font, p.entry, p.slotWidth, fg, wx, y+34, true); err != nil {
			return err
		}
		wx += p.slotWidth * GlyphAdvance
	}
	c.DrawNumber(amount, fg, x+w-42, y+104)
	return nil
}
