package layout

import "github.com/wicanr2/great-era-remake/internal/game"

import "github.com/wicanr2/great-era-remake/internal/assets"

// BattlefieldCellAt 將邏輯畫布上的點反算成 14×14 戰場格。
//
// 原版圖塊寬 32、高 24，奇數欄下移 12 像素；每張圖塊的矩形範圍彼此
// 不重疊，所以命中契約與 renderer 的 CellIndex.ScreenXY 完全相同。
func BattlefieldCellAt(x, y, originX, originY int) (game.CellIndex, bool) {
	lx, ly := x-originX, y-originY
	if lx < 0 || ly < 0 {
		return game.NoCell, false
	}
	col := lx / game.HexCellW
	if col < 0 || col >= assets.GridW {
		return game.NoCell, false
	}
	if col%2 == 1 {
		ly -= game.HexStep
	}
	if ly < 0 {
		return game.NoCell, false
	}
	row := ly / game.HexCellH
	cell, err := game.CellAt(col, row)
	if err != nil {
		return game.NoCell, false
	}
	return cell, true
}
