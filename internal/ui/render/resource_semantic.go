package render

import "github.com/wicanr2/great-era-remake/internal/assets"

func (c *Canvas) DrawSemanticList(fonts *assets.EtenFonts, fg, bg assets.RGB,
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

func (c *Canvas) DrawSemanticAmount(fonts *assets.EtenFonts, fg, bg assets.RGB,
	x, y, w, h int, prompt string, amount uint32) []rune {
	c.fillRect(x, y, w, h, bg)
	missing := c.DrawSemanticText(fonts, prompt, fg, x+28, y+38)
	c.DrawNumber(amount, fg, x+w-42, y+88)
	return uniqueRunes(missing)
}

func (c *Canvas) DrawSemanticSupplyTarget(fonts *assets.EtenFonts, provinces *assets.GlyphFile,
	fg, bg assets.RGB, x, y, w, h int, prompt string, candidates []int, input uint32) ([]rune, error) {
	c.fillRect(x, y, w, h, bg)
	missing := c.DrawSemanticText(fonts, prompt, fg, x+24, y+24)
	for i, id := range candidates {
		cy := y + 70 + i*28
		c.DrawNumber(uint32(id), fg, x+48, cy)
		if id > 0 {
			if err := c.DrawEntry(provinces, id-1, 3, fg, x+102, cy, true); err != nil {
				return nil, err
			}
		}
	}
	c.DrawNumber(input, fg, x+w-42, y+24)
	return uniqueRunes(missing), nil
}
