package render

import (
	"fmt"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
)

// 政略畫面的左側面板。
//
// 版面依據 DOSBox 實機截圖（`docs/playtest/02` §2）：面板寬約 190 px，
// 標籤靠左、數值靠右對齊。13 個欄位的順序照畫面，**不是照資料**——
// 資料是「煤礦、燃料」，畫面印的是「燃料、煤礦」（`docs/spec/03` §2 的雷）。
//
// 標籤全部取自原版詞表，不是重打的字串：
// `2.15` 的 20–35、68，`3.15` 的 39–40（`docs/mechanics/01-vocabulary.md`）。

// 面板版面。**座標是 remake 的排版選擇**，不是還原原版的精確數值。
const (
	panelX     = 8
	panelLabel = panelX + 8   // 標籤起點
	panelValue = panelX + 182 // 數值右邊界
	panelTop   = 8
	lineH      = 18 // 行高
)

// 2.15 的詞條編號。
const (
	w1Count    = 27 // 數
	w2Command  = 18 // 指令
	w2Sikrei   = 21 // 司令
	w2Gold     = 22 // 黃金
	w2Food     = 23 // 糧食
	w2Ammo     = 24 // 彈藥
	w2Coal     = 25 // 煤礦
	w2Fuel     = 26 // 燃料
	w2Iron     = 27 // 鐵礦
	w2Land     = 28 // 地價
	w2Pop      = 29 // 人口
	w2City     = 30 // 城市
	w2Force    = 31 // 兵力
	w2General  = 32 // 將領
	w2People   = 33 // 人民
	w2Normal   = 35 // 正常
	w2Governor = 68 // 省長
)

// 3.15 的詞條編號（前 39 條是省名）。
const (
	w3Loyalty = 39 // 忠誠度
	w3Arsenal = 40 // 兵工廠
)

// 面板配色，取自實機截圖的實際像素值。
var (
	panelFG    = assets.RGB{R: 0xAE, G: 0x00, B: 0x00} // 數值與標籤的暗紅
	panelTitle = assets.RGB{R: 0x00, G: 0x00, B: 0xAA} // 省名的藍
	panelBG    = assets.RGB{R: 0xFF, G: 0xFF, B: 0xA2} // 面板底色的米黃
)

// PanelFonts 是畫面板需要的四個字模檔。
type PanelFonts struct {
	W1  *assets.GlyphFile // 1.15，單字詞（「數」）
	W2  *assets.GlyphFile // 2.15，二字詞
	W3  *assets.GlyphFile // 3.15，三字詞（前 39 條是省名）
	Gen *assets.GlyphFile // MAN{期}15，將領姓名（三字）
}

// PanelData 是面板要顯示的一個省的狀態。
type PanelData struct {
	ID       game.ProvinceID
	Province *game.Province
	Force    uint32 // 該省將領的兵力加總
	Generals int    // 該省的將領數
	Year     uint16
	Month    uint8
	Commands int // 本省本月剩餘指令數
}

// DrawStrategyPanel 把政略畫面的左側面板畫到畫布上。
func (c *Canvas) DrawStrategyPanel(d PanelData, f PanelFonts) error {
	c.fillRect(panelX, 0, panelValue-panelX+8, ModeBGIH, panelBG)

	y := panelTop
	// 省名（三字，藍色）+ 狀態
	if err := c.DrawEntry(f.W3, int(d.ID)-1, 3, panelTitle, panelLabel, y, true); err != nil {
		return err
	}
	if err := c.DrawEntry(f.W2, w2Normal, 2, panelTitle, panelLabel+72, y, true); err != nil {
		return err
	}
	y += lineH + 4

	// 司令與省長
	for _, row := range []struct {
		label int
		gen   game.GeneralID
	}{
		{w2Sikrei, d.Province.Commander},
		{w2Governor, d.Province.Governor},
	} {
		if err := c.DrawEntry(f.W2, row.label, 2, panelTitle, panelLabel, y, true); err != nil {
			return err
		}
		if row.gen.Valid() {
			if err := c.DrawEntry(f.Gen, int(row.gen)-1, 3, panelFG, panelLabel+72, y, true); err != nil {
				return err
			}
		}
		y += lineH
	}
	y += 10

	// 六種資源 + 地價 + 人口 —— 順序照畫面，不是照資料
	p := d.Province
	for _, row := range []struct {
		label int
		value uint32
	}{
		{w2Gold, uint32(p.Gold)},
		{w2Food, uint32(p.Food)},
		{w2Ammo, uint32(p.Ammo)},
		{w2Fuel, uint32(p.Fuel)}, // 畫面上燃料在煤礦之前
		{w2Coal, uint32(p.Coal)},
		{w2Iron, uint32(p.Iron)},
		{w2Land, uint32(p.LandValue)},
		{w2Pop, p.PopulationWan()},
	} {
		if err := c.DrawEntry(f.W2, row.label, 2, panelFG, panelLabel, y, true); err != nil {
			return err
		}
		c.DrawNumber(row.value, panelFG, panelValue, y)
		y += lineH
	}

	// 城市數（二字 + 數）、兵工廠數（三字 + 數）——「數」在原版是獨立的一格，
	// 這裡只畫詞條本身，省略「數」字（remake 排版差異）。
	if err := c.DrawEntry(f.W2, w2City, 2, panelFG, panelLabel, y, true); err != nil {
		return err
	}
	c.DrawNumber(uint32(p.Cities), panelFG, panelValue, y)
	y += lineH
	if err := c.DrawEntry(f.W3, w3Arsenal, 3, panelFG, panelLabel, y, true); err != nil {
		return err
	}
	c.DrawNumber(uint32(p.Arsenals), panelFG, panelValue, y)
	y += lineH + 10

	// 兵力、將領數、人民忠誠度
	if err := c.DrawEntry(f.W2, w2Force, 2, panelFG, panelLabel, y, true); err != nil {
		return err
	}
	c.DrawNumber(d.Force, panelFG, panelValue, y)
	y += lineH
	if err := c.DrawEntry(f.W2, w2General, 2, panelFG, panelLabel, y, true); err != nil {
		return err
	}
	c.DrawNumber(uint32(d.Generals), panelFG, panelValue, y)
	y += lineH
	if err := c.DrawEntry(f.W2, w2People, 2, panelFG, panelLabel, y, true); err != nil {
		return err
	}
	if err := c.DrawEntry(f.W3, w3Loyalty, 3, panelFG, panelLabel+40, y, true); err != nil {
		return err
	}
	c.DrawNumber(uint32(p.Loyalty), panelFG, panelValue, y)
	y += lineH

	// 原版右下角的「指令數」。`sub_10193` 由 `2.15` 詞條 18「指令」與
	// `1.15` 詞條 27「數」拼成（`docs/re/27`）。
	if err := c.DrawEntry(f.W2, w2Command, 2, panelFG, panelLabel, y, true); err != nil {
		return err
	}
	if err := c.DrawEntry(f.W1, w1Count, 1, panelFG, panelLabel+40, y, true); err != nil {
		return err
	}
	remaining := d.Commands
	if remaining < 0 {
		remaining = 0
	}
	c.DrawNumber(uint32(remaining), panelFG, panelValue, y)

	return nil
}

// LoadPanelFonts 讀進面板需要的字模檔。
func LoadPanelFonts(w1, w2, w3, gen []byte) (PanelFonts, error) {
	var f PanelFonts
	var err error
	if f.W1, err = assets.ParseGlyphFile(w1); err != nil {
		return f, fmt.Errorf("render: 1.15: %w", err)
	}
	if f.W2, err = assets.ParseGlyphFile(w2); err != nil {
		return f, fmt.Errorf("render: 2.15: %w", err)
	}
	if f.W3, err = assets.ParseGlyphFile(w3); err != nil {
		return f, fmt.Errorf("render: 3.15: %w", err)
	}
	if f.Gen, err = assets.ParseGlyphFile(gen); err != nil {
		return f, fmt.Errorf("render: 將領名表: %w", err)
	}
	return f, nil
}
