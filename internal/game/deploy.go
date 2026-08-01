package game

import (
	"fmt"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

// 戰場的六角格幾何與部署規則。
//
// 依據 docs/re/07-battlefield-geometry.md：全部取自 `WAR.EXE` 的
// `sub_50FF5`（格 → 螢幕座標）與 `sub_41513`（部署掃描），不是從版面猜的。

// 六角格的版面常數 —— **這幾個值是原版的，不是 remake 的排版選擇**。
//
//	sub_50FF5:  col = cell mod 14
//	            row = cell div 14
//	            x   = col << 5                       ; col * 32
//	            y   = 24*row + (col 為奇數 ? 12 : 0)
//
// 奇數欄整體下移半格 = column-staggered 六角格排列。這是「戰鬥用六角格」
// （CLAUDE.md §1.5 原本只是社群說法）的第一份程式碼證據。
const (
	HexCellW = 32 // 一格的寬（sub_50FF5 的 shl ax, 5）
	HexCellH = 24 // 一格的高（sub_50FF5 的 mul 18h）
	HexStep  = 12 // 奇數欄的下移量，正好半格

	// BattlefieldW / BattlefieldH 是整張戰場佔的像素範圍。
	// 348 塞得進 BGI 640×350 —— 幾何與畫面模式互相印證。
	BattlefieldW = (assets.GridW-1)*HexCellW + HexCellW           // 448
	BattlefieldH = (assets.GridH-1)*HexCellH + HexStep + HexCellH // 348
	CellCount    = assets.GridW * assets.GridH                    // 196
)

// CellIndex 是戰場格編號，0..195，row-major：index = row*14 + col。
//
// 這個編號就是執行期單位記錄 `+5` 存的值（`docs/re/07` §2）。
type CellIndex uint8

// NoCell 是「不在場上」的哨兵。單位記錄 `+5` 用 0xFF 表示空槽。
const NoCell CellIndex = 0xFF

// Valid 回報這個格編號是否落在戰場內。
func (c CellIndex) Valid() bool { return c < CellCount }

// ColRow 把格編號拆成欄、列。
func (c CellIndex) ColRow() (col, row int) {
	return int(c) % assets.GridW, int(c) / assets.GridW
}

// ScreenXY 回傳這一格在原版 640×350 畫面上的左上角座標。
//
// 奇數欄下移半格 —— 這是六角格，不是方格。
func (c CellIndex) ScreenXY() (x, y int) {
	col, row := c.ColRow()
	y = row * HexCellH
	if col%2 == 1 {
		y += HexStep
	}
	return col * HexCellW, y
}

// CellAt 由欄、列算回格編號。
func CellAt(col, row int) (CellIndex, error) {
	if col < 0 || col >= assets.GridW || row < 0 || row >= assets.GridH {
		return NoCell, fmt.Errorf("game: 戰場座標 (%d,%d) 超出 %d×%d",
			col, row, assets.GridW, assets.GridH)
	}
	return CellIndex(row*assets.GridW + col), nil
}

// DeployZone 是一方的進場區：戰場上標著該省編號的那些格。
//
// `WARPOS.DAT` 的每省 14×14 圖不只用來導鄰接表——它就是戰場的分區圖。
// 攻方從自己來源省的那一片格子進場，所以圖上「河南的南邊是湖北」
// 這件事同時決定了地理鄰接與戰術進場方向。
//
// **每個鄰省恰好 10 格，39 省零例外**，正好對上戰鬥記錄的 10 個部隊槽。
const DeployZoneSize = 10

// DeployZone 回傳 defender 省的戰場上屬於 from 省的那些格，
// 順序照原版的掃描方向（cell 195 → 0）。
//
// 依據 `sub_41513`：部署時從 195 往下掃，挑第一個
// `warpos[cell] == 攻方來源省` 且沒人佔的格。掃描方向會影響誰站哪一格，
// 所以這裡不排序成遞增。
func (m *Map) DeployZone(defender, from ProvinceID) ([]CellIndex, error) {
	bf, err := m.Battlefield(defender)
	if err != nil {
		return nil, err
	}
	if !from.Valid() {
		return nil, fmt.Errorf("game: 來源省 %d 超出 1..%d", from, ProvinceCount)
	}
	out := make([]CellIndex, 0, DeployZoneSize)
	for i := CellCount - 1; i >= 0; i-- {
		c := CellIndex(i)
		col, row := c.ColRow()
		if bf.Owner[row][col] == from {
			out = append(out, c)
		}
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("game: 省 %d 的戰場沒有屬於省 %d 的進場區", defender, from)
	}
	return out, nil
}

// Occupancy 是戰場上每格的佔用者，對應執行期的 `word_62A8[196]`。
// 0 表示空格。索引就是 CellIndex。
type Occupancy [CellCount]GeneralID

// Deploy 把一個單位放進 zone 裡第一個空格，回傳落點。
//
// 對應 `sub_41513` 的主迴圈：掃到第一個空格就寫
// `unit[+5] = cell` 與 `word_62A8[cell] = unit`，兩邊互指。
//
// **這裡只實作「空格」這一關。** 原版在寫入前還過一道 `sub_55CEC`
// （地形是否可站），那支還沒讀完；原版找不到合格格子時會退回第二輪，
// 放寬成只檢查空格。等 `sub_55CEC` 解出來再補，不先猜規則。
func (o *Occupancy) Deploy(zone []CellIndex, unit GeneralID) (CellIndex, bool) {
	for _, c := range zone {
		if o[c] == 0 {
			o[c] = unit
			return c, true
		}
	}
	return NoCell, false
}
