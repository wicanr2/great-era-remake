package layout

import "math"

func NewEarthBattleLayout(surface ModernSurface) (ModernBattleLayout, error) {
	l, err := NewModernBattleLayout(surface)
	if err != nil {
		return l, err
	}
	l.Hex = true
	l.Header = Rect{0, 0, 1280, 44}
	l.Map = Rect{242, 46, 500, 580}
	l.MapCard = l.Map
	l.Rail = Rect{17, 112, 192, 102}
	l.Panel = Rect{786, 112, 478, 516}
	l.Log = Rect{786, 600, 478, 28}
	for i := range l.CommandButtons {
		l.CommandButtons[i] = Rect{17 + i*120, 640, 112, 48}
	}
	l.ControlButtons = []Rect{{617, 640, 120, 48}, {745, 640, 140, 48}, {1127, 640, 136, 48}}
	for i := range l.RetreatButtons {
		l.RetreatButtons[i] = Rect{848 + i%3*120, 334 + i/3*56, 112, 48}
	}
	return l, nil
}
func (l ModernBattleLayout) EarthCellRect(col, row int) Rect {
	x, y := l.CellCenter(col, row)
	return Rect{int(math.Floor(x - BattleHexRX)), int(math.Floor(y - BattleHexRY)), int(math.Ceil(BattleHexRX*2)) + 1, 41}
}

// Original RAIL.TPC edge-pixel evidence: rail-original-reference.json.
// Upper/lower horizontal variants are distinct. Index zero is a vertical track.
var battleRailPorts = [21][]int{{2, 5}, {4, 6}, {1, 3}, {2, 4}, {1, 2}, {2, 6}, {2, 3}, {4, 5}, {1, 5}, {5, 6}, {3, 5}, {2, 4, 6}, {1, 2, 3}, {4, 5, 6}, {1, 3, 5}, {2, 5, 6}, {2, 3, 5}, {2, 4, 5}, {1, 2, 5}, {2, 4, 5, 6}, {1, 2, 3, 5}}

func BattleRailPorts(index int) []int {
	if index < 0 || index >= len(battleRailPorts) {
		return nil
	}
	return append([]int(nil), battleRailPorts[index]...)
}

const BattleHexRY = 20.0

var BattleHexRX = 40 / math.Sqrt(3)

func BattleRailOffset(dir int) (float64, float64) {
	switch dir {
	case 1:
		return -.75 * BattleHexRX, .5 * BattleHexRY
	case 2:
		return 0, BattleHexRY
	case 3:
		return .75 * BattleHexRX, .5 * BattleHexRY
	case 4:
		return -.75 * BattleHexRX, -.5 * BattleHexRY
	case 5:
		return 0, -BattleHexRY
	case 6:
		return .75 * BattleHexRX, -.5 * BattleHexRY
	}
	return 0, 0
}
func (l ModernBattleLayout) CellCenter(col, row int) (float64, float64) {
	return float64(l.Map.X) + BattleHexRX + float64(col)*1.5*BattleHexRX, float64(l.Map.Y) + BattleHexRY + float64(row)*2*BattleHexRY + float64(col%2)*BattleHexRY
}
func (l ModernBattleLayout) CellContains(col, row, x, y int) bool {
	if col < 0 || col >= 14 || row < 0 || row >= 14 {
		return false
	}
	cx, cy := l.CellCenter(col, row)
	dx, dy := math.Abs(float64(x)-cx), math.Abs(float64(y)-cy)
	return dy <= BattleHexRY && dx <= BattleHexRX*(1-.5*dy/BattleHexRY)
}
