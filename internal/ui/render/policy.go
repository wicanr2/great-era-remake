package render

import "github.com/wicanr2/great-era-remake/internal/assets"

// AutonomyRow 是授權自治畫面的一個候選省。
type AutonomyRow struct {
	Province   int
	Name       string
	Autonomous bool
}

// DrawPolicyMenu 畫原版指令 8 的兩項政策選單。
func (c *Canvas) DrawPolicyMenu(fonts *assets.EtenFonts, fg, bg assets.RGB,
	x, y, w, h int, title, autonomy, production, unavailable string) []rune {
	c.fillRect(x, y, w, h, bg)
	missing := c.DrawSemanticText(fonts, title, fg, x+24, y+18)
	for i, label := range []string{autonomy, production} {
		cy := y + 82 + i*52
		c.DrawNumber(uint32(i+1), fg, x+44, cy)
		missing = append(missing, c.DrawSemanticText(fonts, label, fg, x+84, cy)...)
	}
	missing = append(missing, c.DrawSemanticText(fonts, unavailable, fg, x+24, y+218)...)
	missing = append(missing, c.DrawSemanticText(fonts, "ESC 返回", fg, x+24, y+h-34)...)
	return uniqueRunes(missing)
}

// DrawAutonomyPolicy 畫兩欄省份候選；選擇後留在本頁，讓玩家
// 在同一個政策指令內切換多省，對齊 `sub_22E25` 的迴圈。
func (c *Canvas) DrawAutonomyPolicy(fonts *assets.EtenFonts, fg, bg assets.RGB,
	x, y, w, h int, title, normal, autonomous, prompt string,
	rows []AutonomyRow, input uint32) []rune {
	c.fillRect(x, y, w, h, bg)
	missing := c.DrawSemanticText(fonts, title, fg, x+18, y+12)

	const perColumn = 19
	columnWidth := w / 2
	for i, row := range rows {
		col, line := i/perColumn, i%perColumn
		if col > 1 {
			break
		}
		cx, cy := x+col*columnWidth+10, y+60+line*14
		c.DrawNumber(uint32(row.Province), fg, cx, cy)
		missing = append(missing, c.DrawSemanticText(fonts, row.Name, fg, cx+42, cy)...)
		state := normal
		if row.Autonomous {
			state = autonomous
		}
		missing = append(missing, c.DrawSemanticText(fonts, state, fg, cx+120, cy)...)
	}

	missing = append(missing, c.DrawSemanticText(fonts, prompt, fg, x+18, y+h-30)...)
	c.DrawNumber(input, fg, x+w-62, y+h-30)
	return uniqueRunes(missing)
}

// DrawProductionPolicy 畫產能分配。黃金是四項分配後的剩餘值。
func (c *Canvas) DrawProductionPolicy(fonts *assets.EtenFonts, fg, bg assets.RGB,
	x, y, w, h int, title string, names [5]string, values [5]uint8,
	selected int, prompt string, input uint32) []rune {
	c.fillRect(x, y, w, h, bg)
	missing := c.DrawSemanticText(fonts, title, fg, x+24, y+18)
	for i := range names {
		cy := y + 66 + i*42
		if selected == i && i > 0 {
			c.strokeRect(x+18, cy-8, w-36, 34, fg)
		}
		if i > 0 {
			c.DrawNumber(uint32(i), fg, x+28, cy)
		}
		missing = append(missing, c.DrawSemanticText(fonts, names[i], fg, x+70, cy)...)
		c.DrawNumber(uint32(values[i]), fg, x+w-92, cy)
		missing = append(missing, c.DrawSemanticText(fonts, "％", fg, x+w-42, cy)...)
	}
	if selected != 0 {
		missing = append(missing, c.DrawSemanticText(fonts, prompt, fg, x+24, y+42)...)
		c.DrawNumber(input, fg, x+w-132, y+42)
	} else {
		missing = append(missing, c.DrawSemanticText(fonts, prompt, fg, x+24, y+h-54)...)
	}
	missing = append(missing, c.DrawSemanticText(fonts, "ESC 完成", fg, x+24, y+h-28)...)
	return uniqueRunes(missing)
}
