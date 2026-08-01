package render

import (
	"fmt"
	"image/color"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

// 字模繪製。原版把 16×15 的點陣字模畫到畫面上，但**不是原樣畫**——
// 有兩個處理，兩個都是逐像素比對原版截圖量出來的：
//
//  1. 加粗：每列與自己右移一位 OR，筆劃水平膨脹 1 px
//  2. 字距 20 px，不是字模寬 16 px
//
// 驗證：政略畫面的「湖北省」三字在 (18,16)、(38,16)、(58,16)，
// 與加粗後的字模**逐像素差異 0**。不加粗的話最佳匹配還差 49 個像素
// （畫面上 132 個前景點 vs 原始字模 83 個）。
//
// remake 不沿用原版「每場景一份字模子集」的機制（CLAUDE.md §6），
// 但**繪製方式要一樣**，否則畫面對不上原版。

// GlyphAdvance 是字距：畫下一個字時 x 要前進多少。
//
// 不等於 assets.GlyphW（16）——原版留了 4 px 字間距。
const GlyphAdvance = 20

// Embolden 把一個字模加粗：每列與自己右移一位 OR。
//
// 這是原版繪字前做的處理。倚天 15 點只有一種字體（明體、偏細），
// 想要粗一點就得程式加粗——`eten-bitmap-font` kb 把這列為推薦做法，
// 原版正是這樣做的。
func Embolden(g assets.Glyph) assets.Glyph {
	var out assets.Glyph
	for y := 0; y < assets.GlyphH; y++ {
		row := uint16(g[y*assets.GlyphRowB])<<8 | uint16(g[y*assets.GlyphRowB+1])
		row |= row >> 1
		out[y*assets.GlyphRowB] = byte(row >> 8)
		out[y*assets.GlyphRowB+1] = byte(row)
	}
	return out
}

// DrawGlyph 把一個字模畫到 (x, y)。
//
// 字模是 1-bit：bit 為 1 畫 fg，為 0 **不畫**（透明，露出底色）。
// 原版的文字是疊在已經畫好的面板上的，不是整塊覆蓋。
//
// bold 為 true 時先加粗——原版畫中文就是加粗的，要對上畫面必須開。
func (c *Canvas) DrawGlyph(g assets.Glyph, fg assets.RGB, x, y int, bold bool) {
	if bold {
		g = Embolden(g)
	}
	col := color.RGBA{fg.R, fg.G, fg.B, 0xFF}
	for gy := 0; gy < assets.GlyphH; gy++ {
		for gx := 0; gx < assets.GlyphW; gx++ {
			if g.At(gx, gy) {
				c.img.SetRGBA(x+gx, y+gy, col)
			}
		}
	}
}

// DrawGlyphs 把一串字模橫向排開，字距 GlyphAdvance。
//
// 原版的排版是定長槽位，空白字模照樣佔一格——那是排版填充，
// 不是可以省略的空格。
func (c *Canvas) DrawGlyphs(gs []assets.Glyph, fg assets.RGB, x, y int, bold bool) {
	for i, g := range gs {
		c.DrawGlyph(g, fg, x+i*GlyphAdvance, y, bold)
	}
}

// DrawEntry 從字模檔取第 k 個詞條畫出來，保留空白填充。
//
// 與 assets.GlyphFile.Entry 不同：Entry 會濾掉空白字模（那是給文字還原用的），
// 這裡**不濾**，因為畫面上的空白格子是排版的一部分。
func (c *Canvas) DrawEntry(f *assets.GlyphFile, k, slotWidth int, fg assets.RGB, x, y int, bold bool) error {
	if slotWidth <= 0 {
		return fmt.Errorf("render: 槽寬必須為正，得到 %d", slotWidth)
	}
	start := k * slotWidth
	if start < 0 || start+slotWidth > len(f.Glyphs) {
		return fmt.Errorf("render: 詞條 %d 超出範圍（共 %d 條）",
			k, len(f.Glyphs)/slotWidth)
	}
	c.DrawGlyphs(f.Glyphs[start:start+slotWidth], fg, x, y, bold)
	return nil
}
