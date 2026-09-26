package render

import (
	"fmt"
	"sort"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

// ModernMapSurfaceData 是 H1-a 地圖／資訊卡預覽的資料邊界。它只讀規則快照與
// theme provider；不持有 Ebiten、不寫存檔，也不重新計算任何玩法。
type ModernMapSurfaceData struct {
	Battlefield         *game.Battlefield
	Theme               uitheme.Theme
	Style               uitheme.UIStyle
	Panel               PanelData
	Fonts               *assets.EtenFonts
	Command             string
	Narrative           string
	Portrait            *i18n.Portrait
	PortraitUnavailable string
}

// DrawModernMapSurface 在 1280×720 image Canvas 上畫 H1-a 的地圖／資訊頁。
// 地形／鐵路仍取原有 Theme 索引，資訊卡只讀 PanelData；未提供字庫時保留
// 面板與地圖幾何但不畫字，讓無頭 metrics 預覽不必偷帶原版字型。其他 Modern
// 高解析頁面各自使用 H1-b／H1-c／H2 renderer；這個函式不執行規則或存檔。
func (c *Canvas) DrawModernMapSurface(d ModernMapSurfaceData) error {
	if c == nil || c.img == nil {
		return fmt.Errorf("Modern H1 Canvas 不可為 nil")
	}
	if c.img.Bounds().Dx() != uilayout.ModernDesignWidth ||
		c.img.Bounds().Dy() != uilayout.ModernDesignHeight {
		return fmt.Errorf("Modern H1 需要 %dx%d Canvas，得到 %dx%d",
			uilayout.ModernDesignWidth, uilayout.ModernDesignHeight,
			c.img.Bounds().Dx(), c.img.Bounds().Dy())
	}
	if d.Battlefield == nil {
		return fmt.Errorf("Modern H1 缺少戰場快照")
	}
	if d.Panel.Province == nil {
		return fmt.Errorf("Modern H1 缺少資訊卡省份快照")
	}
	if d.Theme == nil {
		return fmt.Errorf("Modern H1 主題不可為 nil")
	}
	if d.Style.Name == "" {
		if provider, ok := d.Theme.(uitheme.StyleProvider); ok {
			d.Style = provider.Style()
		} else {
			d.Style = uitheme.RetroStyle()
		}
	}
	if d.Style.Name != uitheme.ModeModern {
		return fmt.Errorf("Modern H1 只接受 modern style，得到 %q", d.Style.Name)
	}
	l, err := uilayout.NewModernMapLayout(uilayout.ModernDesignSurface())
	if err != nil {
		return err
	}

	c.fillRect(0, 0, uilayout.ModernDesignWidth, uilayout.ModernDesignHeight, d.Style.Paper)
	// H4 軍事衛星作戰室把省份上下文固定在工作區頂列；導覽軌只為既有入口
	// 提供第二個命中區，沒有假快捷鍵或未證實軍情。
	drawModernSatelliteChrome(c, l.Rail, l.Header, d.Style)
	drawModernMapRailMarker(c, l.MapRailMarker, d.Style)
	drawModernRailAction(c, l.CommandRailButton, d.Style, true, true)
	drawModernRailAction(c, l.NarrativeRailButton, d.Style, d.Narrative != "", false)
	if d.Panel.ProvinceName != "" {
		if missing := drawModernTitleScaled(c, d.Fonts, d.Panel.ProvinceName, d.Style.Ink,
			l.Header.X+28, l.Header.Y+16, 2, 40); len(missing) != 0 {
			return fmt.Errorf("Modern 儀表板標題缺字：%q", string(missing))
		}
	}
	if d.Panel.Status != "" {
		if missing := drawModernTextScaled(c, d.Fonts, d.Panel.Status, d.Style.Muted,
			l.Header.X+280, l.Header.Y+24, 1, 24); len(missing) != 0 {
			return fmt.Errorf("Modern 儀表板狀態缺字：%q", string(missing))
		}
	}
	if d.Panel.Year != 0 {
		if missing := drawModernTextScaled(c, d.Fonts, fmt.Sprintf("%04d/%02d", d.Panel.Year, d.Panel.Month),
			d.Style.Ink, l.Header.Right()-136, l.Header.Y+24, 1, 16); len(missing) != 0 {
			return fmt.Errorf("Modern 儀表板日期缺字：%q", string(missing))
		}
	}
	drawModernDecoratedPanel(c, l.MapCard, d.Style, d.Style.Panel)
	drawModernSatelliteGrid(c, l.MapCard, d.Style)
	drawModernDecoratedPanel(c, l.InfoPanel, d.Style, d.Style.Panel)
	c.fillRect(l.InfoPanel.X, l.InfoPanel.Y, 4, l.InfoPanel.H, d.Style.AccentAlt)
	drawModernDecoratedPanel(c, l.EventStrip, d.Style, d.Style.Panel)
	c.fillRect(l.EventStrip.X, l.EventStrip.Y, 4, l.EventStrip.H, d.Style.Accent)
	if err := c.drawModernHighBattlefield(d.Battlefield, d.Theme, l, d.Style); err != nil {
		return err
	}
	if err := c.drawModernStrategyDashboard(d.Panel, d.Fonts, l, d.Style, d.Portrait, d.PortraitUnavailable); err != nil {
		return err
	}
	drawModernSignalStrip(c, l.SignalStrip, d.Style)
	if err := c.drawModernDashboardButton(l.CommandButton, d.Command, d.Fonts, d.Style, true); err != nil {
		return err
	}
	if d.Narrative == "" {
		return nil
	}
	return c.drawModernDashboardButton(l.NarrativeButton, d.Narrative, d.Fonts, d.Style, false)
}

func (c *Canvas) drawModernHighBattlefield(bf *game.Battlefield, th uitheme.Theme,
	l uilayout.ModernMapLayout, style uitheme.UIStyle) error {
	for gy := 0; gy < 14; gy++ {
		for gx := 0; gx < 14; gx++ {
			cell, err := game.CellAt(gx, gy)
			if err != nil {
				return err
			}
			p, err := l.MapCellRect(gx, gy)
			if err != nil {
				return err
			}
			if err := drawModernTerrain(c, th, bf.Tiles[gy][gx].Kind.TileIndex(), p); err != nil {
				return fmt.Errorf("Modern H1 地形 (%d,%d)：%w", gx, gy, err)
			}
			if bf.Tiles[gy][gx].HasRail() {
				if err := drawModernRail(c, th, bf.Tiles[gy][gx].Rail, p); err != nil {
					return fmt.Errorf("Modern H1 鐵路 (%d,%d)：%w", gx, gy, err)
				}
			}
			if bf.Owner[gy][gx] != 0 {
				c.strokeRect(p.X, p.Y, p.W, p.H, style.AccentAlt)
			}
			_ = cell // 保留 CellAt 的範圍檢查；規則格號不進呈現計算。
		}
	}
	return nil
}

func drawIndexedScaled(c *Canvas, b uitheme.Bitmap, dst uilayout.Rect, transparentZero bool) error {
	if !b.Valid() {
		return fmt.Errorf("主題圖像無效")
	}
	for y := 0; y < dst.H; y++ {
		sy := y * b.Image.H / dst.H
		for x := 0; x < dst.W; x++ {
			sx := x * b.Image.W / dst.W
			v := b.Image.Pix[sy*b.Image.W+sx]
			if transparentZero && v == 0 {
				continue
			}
			if int(v) >= len(b.Palette) {
				return fmt.Errorf("圖像索引 %d 超出調色盤 %d 色", v, len(b.Palette))
			}
			c.setPixel(dst.X+x, dst.Y+y, b.Palette[v])
		}
	}
	return nil
}

// drawModernStrategyDashboard 是 H4 軍情檢視器。它把原本一路垂直堆疊的 13 個
// 原典欄位收斂成「本回合先看得到」的六個摘要；後勤移入獨立操作／事件列，
// 既有入口則由呼叫端繪製。資料、Action 與存檔欄位一律不變。
func (c *Canvas) drawModernStrategyDashboard(d PanelData, fonts *assets.EtenFonts,
	l uilayout.ModernMapLayout, style uitheme.UIStyle, portrait *i18n.Portrait, portraitUnavailable string) error {
	panel := l.InfoPanel
	draw := func(s string, x, y, maxHalf int, color assets.RGB) error {
		if s == "" {
			return nil
		}
		// H1 高解析頁優先使用隨程式散布的 GEMF atlas；Eten 只作
		// 缺字／玩家自備字庫 fallback，避免資訊卡仍被固定倚天字寬綁住。
		missing := drawModernTextScaled(c, fonts, trimHalfCells(s, maxHalf), color, x, y, 1, maxHalf)
		if len(missing) != 0 {
			return fmt.Errorf("Modern H1 資訊卡缺字：%q", string(missing))
		}
		return nil
	}
	accent := style.Accent
	if d.Faction > 0 && d.Faction <= len(style.FactionTint) {
		accent = style.FactionTint[d.Faction-1]
	}
	portraitMissing := map[rune]bool{}
	if err := c.drawModernPortraitCard(l.CommanderPortrait, portrait, portraitUnavailable, fonts, style, portraitMissing); err != nil {
		return err
	}
	if len(portraitMissing) != 0 {
		missingRunes := make([]rune, 0, len(portraitMissing))
		for r := range portraitMissing {
			missingRunes = append(missingRunes, r)
		}
		sort.Slice(missingRunes, func(i, j int) bool { return missingRunes[i] < missingRunes[j] })
		return fmt.Errorf("Modern 司令肖像欄缺字：%q", string(missingRunes))
	}
	if err := draw(d.Status, panel.X+24, panel.Y+22, 18, style.Muted); err != nil {
		return err
	}
	if missing := drawModernTitleScaled(c, fonts, trimHalfCells(d.ProvinceName, 14), style.Ink,
		panel.X+24, panel.Y+46, 2, 14); len(missing) != 0 {
		return fmt.Errorf("Modern A2 省份標題缺字：%q", string(missing))
	}
	c.strokeRect(panel.X+24, panel.Y+86, panel.W-48, 1, style.Muted)
	if err := draw(d.Labels.Commander, panel.X+24, panel.Y+98, 10, style.Muted); err != nil {
		return err
	}
	if missing := drawModernTextScaled(c, fonts, trimHalfCells(d.CommanderName, 10), style.Ink,
		panel.X+24, panel.Y+120, 2, 10); len(missing) != 0 {
		return fmt.Errorf("Modern A2 司令官姓名缺字：%q", string(missing))
	}
	if err := draw(d.Labels.Governor, panel.X+24, panel.Y+158, 10, style.Muted); err != nil {
		return err
	}
	if err := draw(d.GovernorName, panel.X+112, panel.Y+158, 12, style.Ink); err != nil {
		return err
	}

	commands := d.Commands
	if commands < 0 {
		// 規則層正常情形不會給負值；呈現層仍不可把它轉成巨大的 uint32。
		commands = 0
	}
	metrics := []struct {
		label string
		value uint32
		icon  int
	}{
		{d.Labels.Commands, uint32(commands), -1},
		{d.Labels.Force, d.Force, -1},
		{d.Labels.Generals, uint32(d.Generals), -1},
		{d.Labels.Gold, uint32(d.Province.Gold), 0},
		{d.Labels.Food, uint32(d.Province.Food), 1},
		{d.Labels.Loyalty, uint32(d.Province.Loyalty), -1},
	}
	if len(l.MetricCards) != len(metrics) {
		return fmt.Errorf("Modern 儀表板摘要卡數=%d，預期 %d", len(l.MetricCards), len(metrics))
	}
	for i, metric := range metrics {
		card := l.MetricCards[i]
		drawModernDecoratedPanel(c, card, style, style.Paper)
		c.fillRect(card.X, card.Y, 5, card.H, accent)
		labelX := card.X + 14
		if d.Icons != nil && metric.icon >= 0 {
			if err := drawModernResourceIcon(c, d.Icons, metric.icon,
				uilayout.Rect{X: card.X + 14, Y: card.Y + 12, W: 20, H: 20}); err != nil {
				return err
			}
			labelX += 28
		}
		if err := draw(metric.label, labelX, card.Y+12, 10, style.Muted); err != nil {
			return err
		}
		if err := draw(fmt.Sprintf("%d", metric.value), card.X+14, card.Y+36, 14, style.Ink); err != nil {
			return err
		}
	}
	logistics := []struct {
		label string
		value uint32
	}{
		{d.Labels.Ammo, uint32(d.Province.Ammo)},
		{d.Labels.Fuel, uint32(d.Province.Fuel)},
		{d.Labels.Coal, uint32(d.Province.Coal)},
		{d.Labels.Iron, uint32(d.Province.Iron)},
	}
	events := l.EventStrip
	for i, item := range logistics {
		col, row := i%2, i/2
		x := events.X + 20 + col*(events.W/2)
		y := events.Y + 50 + row*20
		if err := draw(item.label, x, y, 10, style.Muted); err != nil {
			return err
		}
		if err := draw(fmt.Sprintf("%d", item.value), x+92, y, 12, style.Ink); err != nil {
			return err
		}
	}
	return nil
}

func (c *Canvas) drawModernDashboardButton(r uilayout.Rect, label string,
	fonts *assets.EtenFonts, style uitheme.UIStyle, primary bool) error {
	if label == "" {
		return nil
	}
	fill, ink, border := style.Paper, style.Ink, style.AccentAlt
	if primary {
		fill, ink, border = style.Accent, style.Paper, style.Accent
	}
	drawModernDecoratedControl(c, r, style, fill, border)
	// 右側 chevron 是純圖形的前進提示，不另加未語系化的文字。
	for i := 0; i < 8; i++ {
		c.fillRect(r.Right()-32+i, r.Y+r.H/2-8+i, 2, 2, ink)
		c.fillRect(r.Right()-32+i, r.Y+r.H/2+8-i, 2, 2, ink)
	}
	scale := 1
	maxHalf := 24
	y := r.Y + (r.H-16)/2
	if len([]rune(label)) <= 8 {
		scale, maxHalf, y = 2, 12, r.Y+(r.H-30)/2
	}
	missing := drawModernTextScaled(c, fonts, trimHalfCells(label, maxHalf), ink, r.X+22, y, scale, maxHalf)
	if len(missing) != 0 {
		return fmt.Errorf("Modern 儀表板按鈕缺字：%q", string(missing))
	}
	return nil
}
