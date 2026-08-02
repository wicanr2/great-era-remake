package render

import "github.com/wicanr2/great-era-remake/internal/assets"

// RecruitWording 是指令 5 的穩定語意文字；只改顯示，不改兵種順序。
type RecruitWording struct {
	Recruit, Reorganize, Infantry, Armour, Artillery, Cavalry     string
	Amount, Limit, Cost, Gold, Confirm, Remaining, General, Force string
}

func (c *Canvas) DrawRecruitSemantic(fonts *assets.EtenFonts, generals *assets.GlyphFile,
	fg, bg assets.RGB, x, y, w, h int, kind string, labels RecruitWording,
	branch uint8, limit int, amount uint32, cost uint32, targets []int,
	remaining int, generalID int) []rune {
	c.fillRect(x, y, w, h, bg)
	missing := []rune{}
	draw := func(s string, px, py int) { missing = append(missing, c.DrawSemanticText(fonts, s, fg, px, py)...) }
	branches := []string{labels.Infantry, labels.Armour, labels.Artillery, labels.Cavalry}
	switch kind {
	case "action":
		for i, s := range []string{labels.Recruit, labels.Reorganize} {
			c.DrawNumber(uint32(i+1), fg, x+40, y+48+i*52)
			draw(s, x+76, y+48+i*52)
		}
	case "branch":
		for i, s := range branches {
			c.DrawNumber(uint32(i+1), fg, x+40, y+38+i*46)
			draw(s, x+76, y+38+i*46)
		}
	case "amount":
		name := labels.Infantry
		for i, b := range []uint8{1, 5, 4, 6} {
			if branch == b {
				name = branches[i]
			}
		}
		draw(labels.Amount+name, x+24, y+28)
		draw(labels.Limit, x+24, y+78)
		c.DrawNumber(uint32(limit), fg, x+190, y+78)
		c.DrawNumber(amount, fg, x+w-40, y+78)
	case "confirm":
		draw(labels.Cost, x+28, y+42)
		draw(labels.Gold, x+128, y+42)
		c.DrawNumber(cost, fg, x+250, y+42)
		draw(labels.Confirm, x+28, y+112)
	case "targets":
		draw(labels.Remaining, x+24, y+18)
		c.DrawNumber(uint32(remaining), fg, x+210, y+18)
		for i, id := range targets {
			cx, cy := x+30+(i/10)*205, y+56+(i%10)*24
			c.DrawNumber(uint32(i+1), fg, cx, cy)
			if id > 0 {
				_ = c.DrawEntry(generals, id-1, 3, fg, cx+38, cy, true)
			}
		}
		c.DrawNumber(amount, fg, x+w-24, y+18)
	case "reorganize-amount":
		draw(labels.General, x+24, y+26)
		if generalID > 0 {
			_ = c.DrawEntry(generals, generalID-1, 3, fg, x+112, y+26, true)
		}
		draw(labels.Amount+labels.Force, x+24, y+72)
		draw(labels.Remaining, x+24, y+118)
		c.DrawNumber(uint32(remaining), fg, x+220, y+118)
		c.DrawNumber(uint32(limit), fg, x+330, y+118)
		c.DrawNumber(amount, fg, x+w-30, y+118)
	}
	return uniqueRunes(missing)
}
