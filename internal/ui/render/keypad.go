package render

import (
	"github.com/wicanr2/great-era-remake/internal/assets"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
)

// DrawNumericKeypad 畫 1..9、0、刪除與送出。數字沿用原版數字字模；
// 末兩鍵以程式圖形表示，因此 Android 不需要額外授權字型才能輸入。
func (c *Canvas) DrawNumericKeypad(fg, bg assets.RGB) {
	for i := 0; i < 12; i++ {
		p := uilayout.NumericKeypadButton(i)
		c.fillRect(p.X, p.Y, p.HitW, p.HitH, bg)
		c.strokeRect(p.X, p.Y, p.HitW, p.HitH, fg)
		switch {
		case i < 9:
			c.DrawNumber(uint32(i+1), fg, p.X+32, p.Y+15)
		case i == 9:
			c.DrawNumber(0, fg, p.X+32, p.Y+15)
		case i == 10:
			// 刪除：向左箭頭加尾端叉記。
			c.fillRect(p.X+15, p.Y+22, 28, 4, fg)
			for j := 0; j < 4; j++ {
				c.fillRect(p.X+16-j*2, p.Y+22-j*2, 4, 4, fg)
				c.fillRect(p.X+16-j*2, p.Y+22+j*2, 4, 4, fg)
			}
			c.fillRect(p.X+46, p.Y+17, 3, 14, fg)
		case i == 11:
			c.drawCheckmark(fg, p.X, p.Y)
		}
	}
}

func (c *Canvas) drawCheckmark(fg assets.RGB, x, y int) {
	for j := 0; j < 4; j++ {
		c.fillRect(x+14+j*3, y+24+j*3, 4, 4, fg)
	}
	for j := 0; j < 7; j++ {
		c.fillRect(x+25+j*3, y+33-j*3, 4, 4, fg)
	}
}

// DrawSubmitButton 畫單獨的 48×48 勾號送出鍵，供多選清單使用。
func (c *Canvas) DrawSubmitButton(fg, bg assets.RGB, logicalWidth, y, slot int) {
	p := uilayout.NavigationButton(logicalWidth, y, slot)
	c.fillRect(p.X, p.Y, p.HitW, p.HitH, bg)
	c.strokeRect(p.X, p.Y, p.HitW, p.HitH, fg)
	c.drawCheckmark(fg, p.X-8, p.Y)
}
