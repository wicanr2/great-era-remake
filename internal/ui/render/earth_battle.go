package render

import (
	"fmt"
	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/ui/hdart"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
	t "github.com/wicanr2/great-era-remake/internal/ui/theme"
	"math"
	"strconv"
)

func earthHex(c *Canvas, x, y float64, ink assets.RGB) {
	rx, ry := l.BattleHexRX, l.BattleHexRY
	p := [6][2]float64{{x - rx, y}, {x - rx/2, y - ry}, {x + rx/2, y - ry}, {x + rx, y}, {x + rx/2, y + ry}, {x - rx/2, y + ry}}
	for i := range p {
		q := p[(i+1)%6]
		c.line(int(math.Round(p[i][0])), int(math.Round(p[i][1])), int(math.Round(q[0])), int(math.Round(q[1])), ink)
	}
}
func earthRail(c *Canvas, g l.ModernBattleLayout, col, row, index int) {
	ports := l.BattleRailPorts(index)
	if ports == nil {
		return
	}
	cx, cy := g.CellCenter(col, row)
	hub := 0.0
	for _, d := range ports {
		if d == 4 || d == 6 {
			hub = -l.BattleHexRY / 3
			break
		}
		if d == 1 || d == 3 {
			hub = l.BattleHexRY / 3
		}
	}
	for _, dir := range ports {
		dx, dy := l.BattleRailOffset(dir)
		vx, vy := dx, dy-hub
		length := math.Hypot(vx, vy)
		nx, ny := -vy/length, vx/length
		line := func(offset float64, ink assets.RGB) {
			c.line(int(math.Round(cx+nx*offset)), int(math.Round(cy+hub+ny*offset)), int(math.Round(cx+dx+nx*offset)), int(math.Round(cy+dy+ny*offset)), ink)
		}
		for d := 2.0; d < length-1; d += 4 {
			x, y := cx+vx*d/length, cy+hub+vy*d/length
			c.line(int(math.Round(x-nx*3)), int(math.Round(y-ny*3)), int(math.Round(x+nx*3)), int(math.Round(y+ny*3)), assets.RGB{83, 65, 46})
		}
		for _, offset := range []float64{-2, -1, 0, 1, 2} {
			line(offset, assets.RGB{37, 35, 33})
		}
		line(-1, assets.RGB{204, 194, 172})
		line(1, assets.RGB{204, 194, 172})
	}
	if index >= 11 {
		c.fillRect(int(cx)-3, int(cy+hub)-3, 7, 7, assets.RGB{49, 40, 31})
		c.strokeRect(int(cx)-3, int(cy+hub)-3, 7, 7, assets.RGB{183, 148, 95})
		c.fillRect(int(cx)-1, int(cy+hub)-1, 3, 3, assets.RGB{215, 193, 153})
	}
}
func (c *Canvas) drawEarthBattle(d ModernBattleSurfaceData) error {
	g, err := l.NewEarthBattleLayout(l.ModernDesignSurface())
	if err != nil {
		return err
	}
	roomSurface(c, l.Rect{0, 44, 1280, 590}, hdart.Plain, assets.RGB{112, 85, 58}, .35, d.Style.Muted)
	modernArt := false
	if _, ok := d.Theme.(*t.CommandModern); ok {
		modernArt = true
	}
	bases := [12]assets.RGB{{133, 98, 62}, {152, 115, 75}, {170, 133, 88}, {46, 77, 102}, {85, 70, 51}, {149, 130, 106}, {138, 128, 116}, {179, 148, 97}, {143, 132, 110}, {143, 132, 110}, {155, 120, 81}, {120, 96, 71}}
	for row := 0; row < 14; row++ {
		for col := 0; col < 14; col++ {
			tile := d.Battlefield.Tiles[row][col]
			r := g.EarthCellRect(col, row)
			base := bases[min(int(tile.Kind), 11)]
			material := hdart.Plain
			switch tile.Kind {
			case assets.TileWater:
				material = hdart.Ocean
			case assets.TileForest:
				material = hdart.Forest
			case assets.TileMountain:
				material = hdart.Mountain
			case assets.TileDesert:
				material = hdart.Desert
			}
			var bitmap t.Bitmap
			if !modernArt {
				bitmap, err = d.Theme.Tile(tile.Kind.TileIndex())
				if err != nil {
					return err
				}
			}
			for y := r.Y; y < r.Bottom(); y++ {
				for x := r.X; x < r.Right(); x++ {
					if !g.CellContains(col, row, x, y) {
						continue
					}
					ink := base
					if modernArt {
						p := hdart.Sample(material, x, y)
						grain := (int(p.R)+int(p.G)+int(p.B))/3 - 100
						ink = assets.RGB{uint8(max(0, min(255, int(base.R)+grain/3))), uint8(max(0, min(255, int(base.G)+grain/3))), uint8(max(0, min(255, int(base.B)+grain/3)))}
					} else if bitmap.Valid() {
						v := bitmap.Image.Pix[(y-r.Y)*bitmap.Image.H/r.H*bitmap.Image.W+(x-r.X)*bitmap.Image.W/r.W]
						ink = bitmap.Palette[v]
					}
					c.setPixel(x, y, ink)
				}
			}
			x, y := g.CellCenter(col, row)
			if modernArt && tile.Kind == assets.TileCity {
				for _, p := range [][2]int{{-10, -10}, {2, -10}, {-10, 2}, {2, 2}} {
					c.fillRect(int(x)+p[0], int(y)+p[1], 8, 8, assets.RGB{184, 158, 116})
					c.strokeRect(int(x)+p[0], int(y)+p[1], 8, 8, assets.RGB{88, 72, 50})
				}
			}
			if modernArt && tile.Kind == assets.TileForest {
				for i := 0; i < 5; i++ {
					xx, yy := int(x)-12+i*5, int(y)-7+(i*7)%15
					c.fillRect(xx, yy, 5, 5, assets.RGB{64, 54, 40})
				}
			}
			earthHex(c, x, y, assets.RGB{182, 157, 117})
		}
	}
	for row := 0; row < 14; row++ {
		for col := 0; col < 14; col++ {
			index := d.Battlefield.Tiles[row][col].Rail
			if index >= 0 {
				earthRail(c, g, col, row, index)
			}
		}
	}
	drawUnit := func(u *game.Combatant, red bool) error {
		if u == nil || !u.Alive() || !u.Cell.Valid() {
			return nil
		}
		col, row := u.Cell.ColRow()
		x, y := g.CellCenter(col, row)
		index := BranchIcon(u.Branch(), red, u.Facing)
		side := assets.RGB{137, 164, 190}
		if red {
			side = assets.RGB{195, 113, 83}
		}
		if u.Cell == d.CurrentCell {
			for _, shift := range []float64{-1, 0, 1} {
				earthHex(c, x+shift, y, d.Style.Focus)
			}
		}
		if err := drawModernUnit(c, d.Units, index, l.Rect{int(x) - 18, int(y) - 19, 36, 36}); err != nil {
			return err
		}
		c.line(int(x)-14, int(y)+15, int(x)+14, int(y)+15, side)
		roomLabel(c, d.Fonts, strconv.Itoa(int(u.Current)), l.Rect{int(x) - 21, int(y) - 20, 19, 14}, d.Style.Ink, 10, false)
		return nil
	}
	for _, u := range d.Defenders {
		if err := drawUnit(u, !d.HumanDefender); err != nil {
			return err
		}
	}
	for _, u := range d.Attackers {
		if err := drawUnit(u, d.HumanDefender); err != nil {
			return err
		}
	}
	for i, cell := range d.TargetCells {
		if !cell.Valid() {
			continue
		}
		col, row := cell.ColRow()
		x, y := g.CellCenter(col, row)
		earthHex(c, x, y, d.Style.AccentAlt)
		if i < len(d.TargetNums) {
			roomLabel(c, d.Fonts, strconv.Itoa(d.TargetNums[i]), l.Rect{int(x) - 7, int(y) - 12, 20, 24}, d.Style.Ink, 18, false)
		}
	}
	earthFrame(c, g.Header, 0)
	roomLabel(c, d.Fonts, d.Labels.Title, l.Rect{22, 4, 235, 36}, d.Style.Accent, 23, true)
	roomLabel(c, d.Fonts, d.Labels.AttackerName+" / "+d.Labels.DefenderName, l.Rect{290, 5, 460, 34}, d.Style.Ink, 17, false)
	roomLabel(c, d.Fonts, fmt.Sprintf("%s %d   %d/%d", d.Labels.Province, d.Panel.Province, d.Panel.Month, d.Panel.Day), l.Rect{960, 5, 300, 34}, d.Style.Ink, 14, false)
	resource := l.Rect{17, 220, 192, 230}
	earthFrame(c, resource, 0)
	roomLabel(c, d.Fonts, d.Labels.Attacker+" / "+d.Labels.Defender, l.Rect{30, 230, 165, 26}, d.Style.Accent, 12, false)
	rows := []struct {
		label string
		a, b  uint32
	}{{d.Labels.Units, d.Panel.Attacker.Units, d.Panel.Defender.Units}, {d.Labels.Soldiers, d.Panel.Attacker.Soldiers, d.Panel.Defender.Soldiers}, {d.Labels.Gold, d.Panel.Attacker.Gold, d.Panel.Defender.Gold}, {d.Labels.Food, d.Panel.Attacker.Food, d.Panel.Defender.Food}, {d.Labels.Ammo, d.Panel.Attacker.Ammo, d.Panel.Defender.Ammo}, {d.Labels.Fuel, d.Panel.Attacker.Fuel, d.Panel.Defender.Fuel}}
	for i, v := range rows {
		y := 265 + i*27
		roomLabel(c, d.Fonts, v.label, l.Rect{30, y, 55, 22}, d.Style.Ink, 11, false)
		roomLabel(c, d.Fonts, strconv.FormatUint(uint64(v.a), 10), l.Rect{85, y, 53, 22}, assets.RGB{149, 177, 202}, 11, false)
		roomLabel(c, d.Fonts, strconv.FormatUint(uint64(v.b), 10), l.Rect{141, y, 56, 22}, assets.RGB{213, 151, 115}, 11, false)
	}
	roster := l.Rect{786, 163, 180, 372}
	earthFrame(c, roster, 0)
	humanRole, enemyRole := d.Labels.Attacker, d.Labels.Defender
	humanUnits := d.Attackers
	if d.HumanDefender {
		humanRole, enemyRole = d.Labels.Defender, d.Labels.Attacker
		humanUnits = d.Defenders
	}
	roomLabel(c, d.Fonts, humanRole, l.Rect{800, 173, 150, 24}, d.Style.Accent, 15, false)
	var current *game.Combatant
	for i, u := range humanUnits {
		if u == nil {
			continue
		}
		if u.Cell == d.CurrentCell {
			current = u
		}
		if i >= 10 {
			continue
		}
		y := 205 + i*31
		drawModernUnit(c, d.Units, BranchIcon(u.Branch(), false, u.Facing), l.Rect{800, y, 25, 25})
		roomLabel(c, d.Fonts, earthGeneralName(d, u.General), l.Rect{831, y, 60, 25}, d.Style.Ink, 12, false)
		roomLabel(c, d.Fonts, strconv.Itoa(int(u.Strength.Force)), l.Rect{893, y, 60, 25}, d.Style.Ink, 11, false)
	}
	card := l.Rect{984, 268, 278, 324}
	earthFrame(c, card, 0)
	role := humanRole
	if d.HoveredDefender != nil {
		current = d.HoveredDefender
		role = enemyRole
	}
	if current != nil {
		roomLabel(c, d.Fonts, role+" · "+earthGeneralName(d, current.General), l.Rect{1000, 282, 246, 30}, d.Style.Accent, 19, false)
		drawPortraitView(c, d.Fonts, d.Portraits[current.General], l.Rect{X: 1004, Y: 318, W: 100, H: 116})
		drawModernUnit(c, d.Units, BranchIcon(current.Branch(), d.HoveredDefender != nil, current.Facing), l.Rect{1128, 319, 56, 56})
		roomLabel(c, d.Fonts, strconv.Itoa(int(current.Strength.Force)), l.Rect{1120, 386, 126, 32}, d.Style.Ink, 19, false)
		earthFrame(c, l.Rect{998, 438, 250, 138}, 2)
		ink := assets.RGB{35, 45, 55}
		col, row := current.Cell.ColRow()
		for i, v := range []struct{ label, value string }{{d.Labels.Detail[0], fmt.Sprintf("%d / %d", current.Current, current.Max)}, {d.Labels.Detail[1], strconv.Itoa(int(current.Experience))}, {d.Labels.Detail[2], fmt.Sprintf("%d, %d", col+1, row+1)}, {d.Labels.Detail[3], d.Labels.TerrainNames[min(12, int(d.Battlefield.Tiles[row][col].Kind))]}} {
			y := 449 + i*28
			roomLabel(c, d.Fonts, v.label, l.Rect{1010, y, 108, 24}, ink, 12, false)
			roomLabel(c, d.Fonts, v.value, l.Rect{1120, y, 116, 24}, ink, 14, false)
		}
	}
	earthFrame(c, l.Rect{0, 634, 1280, 62}, 0)
	for i, r := range g.CommandButtons {
		kind := 1
		if int(d.Panel.BattleMenuMode) == i+1 {
			kind = 3
		}
		earthFrame(c, r, kind)
		roomLabel(c, d.Fonts, strconv.Itoa(i+1)+" "+d.Labels.Commands[i], l.Rect{r.X + 9, r.Y + 5, r.W - 18, r.H - 10}, d.Style.Ink, 14, false)
	}
	for i, r := range g.ControlButtons {
		kind := 1
		if i == 2 {
			kind = 3
		}
		earthFrame(c, r, kind)
		roomLabel(c, d.Fonts, d.Labels.Controls[i], l.Rect{r.X + 9, r.Y + 5, r.W - 18, r.H - 10}, d.Style.Ink, 14, false)
	}
	if d.Panel.RetreatActive {
		earthFrame(c, g.Panel, 0)
		return c.drawModernRetreatPanel(d, g)
	}
	earthFrame(c, g.Log, 0)
	roomLabel(c, d.Fonts, d.Log, l.Rect{g.Log.X + 9, g.Log.Y + 2, g.Log.W - 18, g.Log.H - 4}, d.Style.Ink, 11, false)
	return nil
}

func earthGeneralName(d ModernBattleSurfaceData, id game.GeneralID) string {
	if name := d.GeneralNames[id]; name != "" {
		return name
	}
	return fmt.Sprintf("#%d", id)
}
