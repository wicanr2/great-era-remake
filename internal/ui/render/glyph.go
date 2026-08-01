package render

import (
	"fmt"
	"image/color"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

// 字模繪製。原版把 16×15 的點陣字模直接送 BGI putimage
// （docs/formats/01-glyph-text.md），這裡重現同一件事。
//
// remake 不沿用原版的「每場景一份字模子集」機制（CLAUDE.md §6）——
// 那是磁片容量的產物。但**繪製方式要一樣**，否則畫面對不上原版。

// DrawGlyph 把一個字模畫到 (x, y)。
//
// 字模是 1-bit：bit 為 1 畫 fg，為 0 **不畫**（透明，露出底色）。
// 原版的文字是疊在已經畫好的面板上的，不是整塊覆蓋。
func (c *Canvas) DrawGlyph(g assets.Glyph, fg assets.RGB, x, y int) {
	col := color.RGBA{fg.R, fg.G, fg.B, 0xFF}
	for gy := 0; gy < assets.GlyphH; gy++ {
		for gx := 0; gx < assets.GlyphW; gx++ {
			if g.At(gx, gy) {
				c.img.SetRGBA(x+gx, y+gy, col)
			}
		}
	}
}

// DrawGlyphs 把一串字模橫向排開，字距固定 GlyphW。
//
// 原版的排版是定長槽位（每格 16 px），不是比例字。
// 空白字模照樣佔一格——這是排版填充，不是可以省略的空格。
func (c *Canvas) DrawGlyphs(gs []assets.Glyph, fg assets.RGB, x, y int) {
	for i, g := range gs {
		c.DrawGlyph(g, fg, x+i*assets.GlyphW, y)
	}
}

// DrawEntry 從字模檔取第 k 個詞條畫出來，保留空白填充。
//
// 與 assets.GlyphFile.Entry 不同：Entry 會濾掉空白字模（給文字還原用），
// 這裡**不濾**，因為畫面上的空白格子是排版的一部分。
func (c *Canvas) DrawEntry(f *assets.GlyphFile, k, slotWidth int, fg assets.RGB, x, y int) error {
	if slotWidth <= 0 {
		return fmt.Errorf("render: 槽寬必須為正，得到 %d", slotWidth)
	}
	start := k * slotWidth
	if start < 0 || start+slotWidth > len(f.Glyphs) {
		return fmt.Errorf("render: 詞條 %d 超出範圍（共 %d 條）",
			k, len(f.Glyphs)/slotWidth)
	}
	c.DrawGlyphs(f.Glyphs[start:start+slotWidth], fg, x, y)
	return nil
}
