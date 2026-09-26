package render

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

// ModernBattleLabels 是戰鬥 Modern HUD 的語系邊界；renderer 不把中文／英文
// 硬編碼成規則分支。命令順序仍與 battleModeForCommand 的 1..5 對齊。
type ModernBattleLabels struct {
	Title, Attacker, Defender, Attack, Province, Date string
	AttackerName, DefenderName                        string
	Units, Soldiers, Gold, Food, Ammo, Fuel           string
	Commands                                          [5]string
	Controls                                          [3]string
	Mode                                              string
	Back, Delete, Submit                              string
}

// ModernBattleSurfaceData 是 H1-c 的純呈現資料；所有 combatant 欄位由 cmd/dsds
// 先整理，renderer 不執行移動、攻擊、AI 或存檔寫回。
type ModernBattleSurfaceData struct {
	Battlefield *game.Battlefield
	Theme       uitheme.Theme
	Units       uitheme.UnitProvider
	Attackers   []*game.Combatant
	Defenders   []*game.Combatant
	CurrentCell game.CellIndex
	TargetCells []game.CellIndex
	TargetNums  []int
	Panel       BattlePanelData
	Labels      ModernBattleLabels
	Fonts       *assets.EtenFonts
	Log         string
	Style       uitheme.UIStyle
}

// DrawModernBattleSurface 畫 1280×720 戰鬥 HUD。地形／鐵路／部隊圖示沿用已驗證
// 的 provider 索引，只替換顯示尺寸與面板層級；戰鬥規則仍由 BattleSim 完成。
func (c *Canvas) DrawModernBattleSurface(d ModernBattleSurfaceData) error {
	if err := validateModernPageCanvas(c, d.Style); err != nil {
		return err
	}
	if d.Battlefield == nil || d.Theme == nil || d.Units == nil {
		return fmt.Errorf("Modern 戰鬥缺少 battlefield、theme 或 unit provider")
	}
	l, err := uilayout.NewModernBattleLayout(uilayout.ModernDesignSurface())
	if err != nil {
		return err
	}
	c.fillRect(0, 0, uilayout.ModernDesignWidth, uilayout.ModernDesignHeight, d.Style.Paper)
	drawModernSatelliteChrome(c, l.Rail, l.Header, d.Style)
	drawModernDecoratedPanel(c, l.MapCard, d.Style, d.Style.Panel)
	drawModernSatelliteGrid(c, l.MapCard, d.Style)
	drawModernDecoratedPanel(c, l.Panel, d.Style, d.Style.Panel)
	if missing := drawModernTitleScaled(c, d.Fonts, d.Labels.Title, d.Style.Ink, l.Header.X+24, l.Header.Y+12, 2, 48); len(missing) != 0 {
		return fmt.Errorf("Modern 戰鬥標題缺字：%q", string(missing))
	}
	if err := c.drawModernBattlefield(d, l); err != nil {
		return err
	}
	if err := c.drawModernBattlePanel(d, l); err != nil {
		return err
	}
	return nil
}

func (c *Canvas) drawModernBattlefield(d ModernBattleSurfaceData, l uilayout.ModernBattleLayout) error {
	for gy := 0; gy < 14; gy++ {
		for gx := 0; gx < 14; gx++ {
			p, err := l.MapCellRect(gx, gy)
			if err != nil {
				return err
			}
			if err := drawModernTerrain(c, d.Theme, d.Battlefield.Tiles[gy][gx].Kind.TileIndex(), p); err != nil {
				return err
			}
			if d.Battlefield.Tiles[gy][gx].HasRail() {
				if err := drawModernRail(c, d.Theme, d.Battlefield.Tiles[gy][gx].Rail, p); err != nil {
					return err
				}
			}
			if d.Battlefield.Owner[gy][gx] != 0 {
				c.strokeRect(p.X, p.Y, p.W, p.H, d.Style.AccentAlt)
			}
		}
	}
	drawUnit := func(u *game.Combatant, red bool) error {
		if u == nil || !u.Alive() || !u.Cell.Valid() {
			return nil
		}
		col, row := u.Cell.ColRow()
		cell, err := l.MapCellRect(col, row)
		if err != nil {
			return err
		}
		return drawModernUnit(c, d.Units, BranchIcon(u.Branch(), red, u.Facing),
			uilayout.Rect{X: cell.X + 4, Y: cell.Y + 5, W: cell.W - 8, H: cell.H - 10})
	}
	for _, u := range d.Defenders {
		if err := drawUnit(u, true); err != nil {
			return err
		}
	}
	for _, u := range d.Attackers {
		if err := drawUnit(u, false); err != nil {
			return err
		}
	}
	if d.CurrentCell.Valid() {
		if cell, err := modernCellRect(l, d.CurrentCell); err == nil {
			c.strokeRect(cell.X, cell.Y, cell.W, cell.H, d.Style.Focus)
		}
	}
	for i, target := range d.TargetCells {
		cell, err := modernCellRect(l, target)
		if err != nil {
			continue
		}
		c.strokeRect(cell.X+3, cell.Y+3, cell.W-6, cell.H-6, d.Style.Accent)
		if i < len(d.TargetNums) {
			_ = drawModernTextScaled(c, d.Fonts, strconv.Itoa(d.TargetNums[i]), d.Style.Accent,
				cell.X+6, cell.Y+4, 1, 4)
		}
	}
	return nil
}

func modernCellRect(l uilayout.ModernBattleLayout, cell game.CellIndex) (uilayout.Rect, error) {
	if !cell.Valid() {
		return uilayout.Rect{}, fmt.Errorf("戰鬥格無效：%d", cell)
	}
	col, row := cell.ColRow()
	return l.MapCellRect(col, row)
}

func (c *Canvas) drawModernBattlePanel(d ModernBattleSurfaceData, l uilayout.ModernBattleLayout) error {
	p := l.Panel
	labels := d.Labels
	if missing := drawModernTextScaled(c, d.Fonts, labels.Province+" "+strconv.Itoa(int(d.Panel.Province)), d.Style.Ink, p.X+24, p.Y+20, 1, 30); len(missing) != 0 {
		return fmt.Errorf("Modern 戰鬥省名缺字：%q", string(missing))
	}
	date := labels.Date
	if date == "" {
		date = fmt.Sprintf("%d/%d", d.Panel.Month, d.Panel.Day)
	}
	_ = drawModernTextScaled(c, d.Fonts, date, d.Style.Muted, p.Right()-150, p.Y+20, 1, 18)
	attacker := labels.Attacker + " " + labels.AttackerName
	defender := labels.Defender + " " + labels.DefenderName
	if missing := drawModernTextScaled(c, d.Fonts, attacker+"  "+labels.Attack+"  "+defender,
		d.Style.Ink, p.X+24, p.Y+54, 1, 52); len(missing) != 0 {
		return fmt.Errorf("Modern 戰鬥雙方缺字：%q", string(missing))
	}
	c.strokeRect(p.X+16, p.Y+78, p.W-32, 1, d.Style.Muted)
	rows := []struct {
		label string
		atk   uint32
		def   uint32
	}{
		{labels.Units, d.Panel.Attacker.Units, d.Panel.Defender.Units},
		{labels.Soldiers, d.Panel.Attacker.Soldiers, d.Panel.Defender.Soldiers},
		{labels.Gold, d.Panel.Attacker.Gold, d.Panel.Defender.Gold},
		{labels.Food, d.Panel.Attacker.Food, d.Panel.Defender.Food},
		{labels.Ammo, d.Panel.Attacker.Ammo, d.Panel.Defender.Ammo},
		{labels.Fuel, d.Panel.Attacker.Fuel, d.Panel.Defender.Fuel},
	}
	for i, row := range rows {
		col, compactRow := i%2, i/2
		x := p.X + 24 + col*(p.W/2)
		y := p.Y + 94 + compactRow*18
		_ = drawModernTextScaled(c, d.Fonts, row.label, d.Style.Muted, x, y, 1, 10)
		_ = drawModernTextScaled(c, d.Fonts, strconv.FormatUint(uint64(row.atk), 10)+"/"+
			strconv.FormatUint(uint64(row.def), 10), d.Style.Ink, x+84, y, 1, 18)
	}
	drawModernDecoratedPanel(c, l.Log, d.Style, d.Style.Paper)
	if d.Panel.RetreatActive {
		return c.drawModernRetreatPanel(d, l)
	}
	for i, card := range l.CommandButtons {
		fill, ink := d.Style.Paper, d.Style.Ink
		if int(d.Panel.BattleMenuMode) == i+1 {
			fill, ink = d.Style.Accent, d.Style.Paper
		}
		drawModernDecoratedControl(c, card, d.Style, fill, d.Style.Muted)
		label := labels.Commands[i]
		_ = drawModernTextScaled(c, d.Fonts, strconv.Itoa(i+1)+"  "+label, ink, card.X+20, card.Y+9, 1, 30)
	}
	for i, card := range l.ControlButtons {
		drawModernDecoratedControl(c, card, d.Style, d.Style.Paper, d.Style.AccentAlt)
		_ = drawModernTextScaled(c, d.Fonts, labels.Controls[i], d.Style.Ink, card.X+18, card.Y+(card.H-16)/2, 1, 18)
	}
	return c.drawModernBattleLog(d, l)
}

func (c *Canvas) drawModernRetreatPanel(d ModernBattleSurfaceData, l uilayout.ModernBattleLayout) error {
	for i, key := range l.RetreatButtons {
		drawModernDecoratedControl(c, key, d.Style, d.Style.Paper, d.Style.AccentAlt)
		label := strconv.Itoa(i + 1)
		switch {
		case i == 9:
			label = "0"
		case i == 10:
			label = d.Labels.Delete
		case i == 11:
			label = d.Labels.Submit
		}
		_ = drawModernTextScaled(c, d.Fonts, label, d.Style.Ink, key.X+key.W/2-8, key.Y+13, 2, 12)
	}
	_ = drawModernTextScaled(c, d.Fonts, d.Labels.Mode+"  "+strconv.FormatUint(uint64(d.Panel.RetreatInput), 10),
		d.Style.Accent, l.Panel.X+24, l.Panel.Y+204, 1, 36)
	return c.drawModernBattleLog(d, l)
}

func (c *Canvas) drawModernBattleLog(d ModernBattleSurfaceData, l uilayout.ModernBattleLayout) error {
	if strings.TrimSpace(d.Log) == "" {
		return nil
	}
	returnValue := drawModernTextScaled(c, d.Fonts, d.Log, d.Style.Muted, l.Log.X+14, l.Log.Y+6, 1, 58)
	if len(returnValue) != 0 {
		return fmt.Errorf("Modern 戰鬥訊息缺字：%q", string(returnValue))
	}
	return nil
}
