package render

import (
	"github.com/wicanr2/great-era-remake/internal/assets"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
)

// DrawOtherOptions 畫原版八項與明示的 remake 顯示／解析度入口。available 為 false
// 時只顯示不可用狀態，輸入層也必須拒絕執行，避免做出假開關。
func (c *Canvas) DrawOtherOptions(fonts *assets.EtenFonts, fg, bg assets.RGB,
	x, y, w, h int, labels []string, available []bool, unavailable string) []rune {
	c.fillRect(x, y, w, h, bg)
	missing := []rune{}
	const perCol, colW, rowH = 5, 215, 55
	for i, label := range labels {
		p := uilayout.Grid(i, x+12, y+30, perCol, colW, rowH, 8, 210, 54)
		cx, cy := p.X, p.Y
		c.DrawNumber(uint32(i+1), fg, cx+24, cy)
		missing = append(missing, c.DrawSemanticText(fonts, label, fg, cx+48, cy)...)
		if i >= len(available) || !available[i] {
			missing = append(missing, c.DrawSemanticText(fonts, unavailable, fg, cx+48, cy+23)...)
		}
	}
	missing = append(missing, c.DrawSemanticText(fonts, "ESC 返回", fg, x+24, y+h-34)...)
	return uniqueRunes(missing)
}

// DrawDisplayOptions 畫 remake 新增的「顯示設定」子頁。用語與圖形主題各自
// 有明確選取框；未完成的字型／版面軸不顯示假選項。
func (c *Canvas) DrawDisplayOptions(fonts *assets.EtenFonts, fg, bg assets.RGB,
	x, y, w, h int, title, wording, original, plain, theme, retro, modern string,
	wordingSelected, themeSelected int) []rune {
	c.fillRect(x, y, w, h, bg)
	missing := c.DrawSemanticText(fonts, title, fg, x+24, y+18)
	missing = append(missing, c.DrawSemanticText(fonts, wording, fg, x+24, y+62)...)
	for i, label := range []string{original, plain} {
		p := uilayout.DisplayWordingOption(i, x+28, y, w-56)
		cy := p.Y
		if wordingSelected == i {
			c.strokeRect(p.X, cy-8, p.HitW, 32, fg)
		}
		c.DrawNumber(uint32(i+1), fg, p.X+22, cy)
		missing = append(missing, c.DrawSemanticText(fonts, label, fg, p.X+60, cy)...)
	}

	missing = append(missing, c.DrawSemanticText(fonts, theme, fg, x+24, y+184)...)
	for i, label := range []string{retro, modern} {
		p := uilayout.DisplayThemeOption(i, x+28, y, w-56)
		if themeSelected == i {
			c.strokeRect(p.X, p.Y-8, p.HitW, 32, fg)
		}
		c.DrawNumber(uint32(i+3), fg, p.X+22, p.Y)
		missing = append(missing, c.DrawSemanticText(fonts, label, fg, p.X+60, p.Y)...)
	}
	missing = append(missing, c.DrawSemanticText(fonts, "ESC 返回", fg, x+24, y+h-34)...)
	return uniqueRunes(missing)
}

// DrawResolutionOptions 畫原版／高解析畫布選擇頁。它獨立於 Modern theme，
// 讓 retro 玩家也能先選畫布再切換圖形；F3 與這兩個觸控列走同一個 setter。
func (c *Canvas) DrawResolutionOptions(fonts *assets.EtenFonts, fg, bg assets.RGB,
	x, y, w, h int, title, original, high string, selected int) []rune {
	c.fillRect(x, y, w, h, bg)
	missing := c.DrawSemanticText(fonts, title, fg, x+24, y+24)
	for i, label := range []string{original, high} {
		cy := y + 82 + i*52
		if selected == i {
			c.strokeRect(x+18, cy-10, w-36, 36, fg)
		}
		c.DrawNumber(uint32(i+1), fg, x+42, cy)
		missing = append(missing, c.DrawSemanticText(fonts, label, fg, x+82, cy)...)
	}
	missing = append(missing, c.DrawSemanticText(fonts, "ESC 返回", fg, x+24, y+h-34)...)
	return uniqueRunes(missing)
}

// DrawMessageOverlay 在右側工作區底部畫最多兩行訊息。原版會在畫面上
// 停留 byte_6FE85×400 ms；計時與佇列由輸入層管理，renderer 只畫當前文字。
func (c *Canvas) DrawMessageOverlay(fonts *assets.EtenFonts, fg, bg assets.RGB,
	x, y, w, h int, msg string) []rune {
	c.fillRect(x, y, w, h, bg)
	c.strokeRect(x, y, w, h, fg)
	lines := wrapSemanticMessage(msg, w-24, 3)
	missing := []rune{}
	for i, line := range lines {
		missing = append(missing, c.DrawSemanticText(fonts, line, fg, x+12, y+10+i*24)...)
	}
	return uniqueRunes(missing)
}

func wrapSemanticMessage(s string, maxPixels, maxLines int) []string {
	if maxPixels <= 0 || maxLines <= 0 {
		return nil
	}
	lines, line, width := []string{}, []rune{}, 0
	for _, r := range []rune(s) {
		advance := 20
		if r < 0x80 {
			advance = 10
		}
		if width+advance > maxPixels && len(line) > 0 {
			lines = append(lines, string(line))
			if len(lines) == maxLines {
				lines[len(lines)-1] = fitMessageEllipsis(lines[len(lines)-1], maxPixels)
				return lines
			}
			line, width = nil, 0
		}
		line, width = append(line, r), width+advance
	}
	if len(line) > 0 && len(lines) < maxLines {
		lines = append(lines, string(line))
	}
	return lines
}

func fitMessageEllipsis(s string, maxPixels int) string {
	runes := []rune(s)
	width := func(rs []rune) int {
		n := 0
		for _, r := range rs {
			if r < 0x80 {
				n += 10
			} else {
				n += 20
			}
		}
		return n
	}
	for len(runes) > 0 && width(runes)+30 > maxPixels {
		runes = runes[:len(runes)-1]
	}
	return string(runes) + "..."
}
