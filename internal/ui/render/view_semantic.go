package render

import "github.com/wicanr2/great-era-remake/internal/assets"

// DrawSemanticProvinceSelect 保留原版省編號與省名字模，只把倒裝提示換成語意文字。
func (c *Canvas) DrawSemanticProvinceSelect(fonts *assets.EtenFonts, provinces *assets.GlyphFile,
	fg, bg assets.RGB, x, y, w, h int, prompt string, input uint32) ([]rune, error) {
	c.fillRect(x, y, w, h, bg)
	missing := c.DrawSemanticText(fonts, prompt, fg, x+22, y+14)
	c.DrawNumber(input, fg, x+w-132, y+14)
	for i := 0; i < 39; i++ {
		col, line := i/13, i%13
		cx, cy := x+18+col*145, y+48+line*22
		c.DrawNumber(uint32(i+1), fg, cx+28, cy)
		if err := c.DrawEntry(provinces, i, 3, fg, cx+38, cy, true); err != nil {
			return nil, err
		}
	}
	return uniqueRunes(missing), nil
}

// DrawSemanticProvinceChoice 畫選定省份後的兩條白話路徑。
func (c *Canvas) DrawSemanticProvinceChoice(fonts *assets.EtenFonts, provinces *assets.GlyphFile,
	fg, bg assets.RGB, x, y, w, h, province int, labels []string) ([]rune, error) {
	c.fillRect(x, y, w, h, bg)
	if province > 0 {
		if err := c.DrawEntry(provinces, province-1, 3, fg, x+28, y+18, true); err != nil {
			return nil, err
		}
	}
	missing := []rune{}
	for i, label := range labels {
		cy := y + 76 + i*56
		c.DrawNumber(uint32(i+1), fg, x+54, cy)
		missing = append(missing, c.DrawSemanticText(fonts, label, fg, x+82, cy)...)
	}
	return uniqueRunes(missing), nil
}

// DrawSemanticOwnedProvinceOverview 保留省名字模、排序與兵力，只換標題。
func (c *Canvas) DrawSemanticOwnedProvinceOverview(fonts *assets.EtenFonts, provinces *assets.GlyphFile,
	fg, bg assets.RGB, x, y, w, h int, title string, ids []int, forces []uint32) ([]rune, error) {
	c.fillRect(x, y, w, h, bg)
	missing := c.DrawSemanticText(fonts, title, fg, x+22, y+14)
	for i, id := range ids {
		col, line := i/10, i%10
		cx, cy := x+22+col*205, y+52+line*27
		c.DrawNumber(uint32(id), fg, cx+28, cy)
		if id > 0 {
			if err := c.DrawEntry(provinces, id-1, 3, fg, cx+38, cy, true); err != nil {
				return nil, err
			}
		}
		if i < len(forces) {
			c.DrawNumber(forces[i], fg, cx+194, cy)
		}
	}
	return uniqueRunes(missing), nil
}

// DrawSemanticProvinceNames 保留原版兩頁範圍與省名字模，只換對照表標題。
func (c *Canvas) DrawSemanticProvinceNames(fonts *assets.EtenFonts, provinces *assets.GlyphFile,
	fg, bg assets.RGB, x, y, w, h int, title string, page, limit int) ([]rune, error) {
	c.fillRect(x, y, w, h, bg)
	missing := c.DrawSemanticText(fonts, title, fg, x+22, y+14)
	start, end := 1, 20
	if page == 2 {
		start, end = 21, limit
	}
	if limit < end {
		end = limit
	}
	for id := start; id <= end; id++ {
		i := id - start
		col, line := i/10, i%10
		cx, cy := x+30+col*205, y+58+line*27
		c.DrawNumber(uint32(id), fg, cx+30, cy)
		if err := c.DrawEntry(provinces, id-1, 3, fg, cx+42, cy, true); err != nil {
			return nil, err
		}
	}
	c.DrawNumber(uint32(page), fg, x+w-132, y+14)
	return uniqueRunes(missing), nil
}
