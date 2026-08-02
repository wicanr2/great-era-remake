package render

import (
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
)

// 用實機那一場（蔣中正攻打孫傳芳的江西）的數值畫一次，
// 逐像素比對截圖裡的**標籤**。
//
// 只比標籤不比數字：數字字模是自己畫的（`smalldigits.go` 標了 remake 差異），
// 逐像素一定不同。標籤走的是原版字模，**必須零差異**。
func TestBattlePanelLabelsMatchOriginal(t *testing.T) {
	fonts, err := LoadPanelFonts(readFile(t, gameDir, "1.15"), readFile(t, gameDir, "2.15"),
		readFile(t, gameDir, "3.15"), readFile(t, gameDir, "MAN115"))
	if err != nil {
		t.Fatal(err)
	}

	shot := loadPNG(t, "m02.png")

	c := NewCanvas(ModeBGIW, ModeBGIH)
	d := BattlePanelData{
		Province: 25, Month: 8, Day: 1,
		Attacker: BattleSide{Units: 3, Soldiers: 33472, Gold: 335,
			Food: 3348, Ammo: 15300, Fuel: 23988},
		Defender: BattleSide{Units: 9, Soldiers: 28200, Gold: 3655,
			Food: 19089, Ammo: 16170, Fuel: 12047},
	}
	if err := c.DrawBattlePanel(d, fonts); err != nil {
		t.Fatal(err)
	}

	// 六個資料列的標籤。前兩列三字、後四列兩字（畫在中間一格）。
	labels := []struct {
		name   string
		row    int
		width  int
		x, adv int
	}{
		{"單位數", 0, 3, battlePanelLabelX, battlePanelLabelW},
		{"士兵數", 1, 3, battlePanelLabelX, battlePanelLabelW},
		{"黃金", 2, 2, battlePanelLabel2X, battlePanelLabel2W},
		{"糧食", 3, 2, battlePanelLabel2X, battlePanelLabel2W},
		{"彈藥", 4, 2, battlePanelLabel2X, battlePanelLabel2W},
		{"燃料", 5, 2, battlePanelLabel2X, battlePanelLabel2W},
	}
	for _, l := range labels {
		y := battlePanelRowY + l.row*battlePanelRowH
		x0 := l.x
		var diff int
		for i := 0; i < l.width; i++ {
			for gy := 0; gy < assets.GlyphH; gy++ {
				for gx := 0; gx < assets.GlyphW+1; gx++ {
					px := x0 + i*l.adv + gx
					if px >= ModeBGIW || y+gy >= ModeBGIH {
						continue
					}
					got := c.Image().At(px, y+gy)
					want := shot.At(px, y+gy)
					gr, gg, gb, _ := got.RGBA()
					wr, wg, wb, _ := want.RGBA()
					gIsBlue := uint8(gr>>8) == battlePanelLabel.R &&
						uint8(gg>>8) == battlePanelLabel.G &&
						uint8(gb>>8) == battlePanelLabel.B
					wIsBlue := uint8(wr>>8) == battlePanelLabel.R &&
						uint8(wg>>8) == battlePanelLabel.G &&
						uint8(wb>>8) == battlePanelLabel.B
					if gIsBlue != wIsBlue {
						diff++
					}
				}
			}
		}
		if diff != 0 {
			t.Errorf("標籤「%s」@y=%d 與原版相異 %d 個像素", l.name, y, diff)
		}
	}
}

// 版面常數與實機量到的座標一致——這是防止有人「順手調整」的護欄。
func TestBattlePanelLayoutConstants(t *testing.T) {
	// 六列的頂端，實機量到 49／65／81／97／113／129。
	want := []int{49, 65, 81, 97, 113, 129}
	for i, w := range want {
		if got := battlePanelRowY + i*battlePanelRowH; got != w {
			t.Errorf("第 %d 列的 y 是 %d，實機是 %d", i+1, got, w)
		}
	}
	// 三字標籤的左邊界，實機量到 520／538／556。
	for i, w := range []int{520, 538, 556} {
		if got := battlePanelLabelX + i*battlePanelLabelW; got != w {
			t.Errorf("三字標籤第 %d 字的 x 是 %d，實機是 %d", i+1, got, w)
		}
	}
	// 二字標籤是 530／550——**不是把三字置中**，字距反而更大。
	for i, w := range []int{530, 550} {
		if got := battlePanelLabel2X + i*battlePanelLabel2W; got != w {
			t.Errorf("二字標籤第 %d 字的 x 是 %d，實機是 %d", i+1, got, w)
		}
	}
}

// 小數字的字距與尺寸符合實機量到的值。
func TestSmallDigitMetrics(t *testing.T) {
	// 實機「33472」五個數字的左邊界是 470／478／486／494／502，右邊界 508。
	c := NewCanvas(ModeBGIW, ModeBGIH)
	c.DrawSmallNumber(33472, assets.RGB{}, 508, 65)
	if w := SmallNumberWidth(33472); w != 5*SmallDigitAdvance {
		t.Errorf("五位數寬度 %d，預期 %d", w, 5*SmallDigitAdvance)
	}
	if SmallDigitAdvance != 8 || SmallDigitW != 6 || SmallDigitH != 14 {
		t.Errorf("小數字規格 %d/%d/%d，實機量到 8/6/14",
			SmallDigitAdvance, SmallDigitW, SmallDigitH)
	}
}

// 每個小數字都畫得出東西——空字模會讓面板看起來像壞掉。
func TestSmallDigitsAllRender(t *testing.T) {
	for d := 0; d <= 9; d++ {
		c := NewCanvas(ModeBGIW, ModeBGIH)
		c.DrawSmallDigit(d, assets.RGB{R: 255}, 10, 10)
		n := 0
		for y := 10; y < 10+SmallDigitH; y++ {
			for x := 10; x < 10+SmallDigitW; x++ {
				r, _, _, _ := c.Image().At(x, y).RGBA()
				if uint8(r>>8) == 255 {
					n++
				}
			}
		}
		if n < 10 {
			t.Errorf("數字 %d 只畫了 %d 個像素，字模可能是空的", d, n)
		}
	}
}

var _ = game.ProvinceID(0)
