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

// 六角鄰接。位移表出自 `sub_4E65C`，邊界規則出自 `sub_510E0`
// （`docs/re/07` §6）。

// HexDir 是六個方向，編號與原版一致（`sub_55CEC` 的迴圈跑 1..6）。
type HexDir int

// 兩組位移的方向語意是一致的，所以可以給共用的名字。
const (
	DirLowerLeft  HexDir = 1
	DirDown       HexDir = 2
	DirLowerRight HexDir = 3
	DirUpperLeft  HexDir = 4
	DirUp         HexDir = 5
	DirUpperRight HexDir = 6
)

// hexDelta[col 的奇偶][方向] 是格編號的位移。
//
// 索引 0 空著是為了讓方向編號與原版對齊（原版陣列基址就是 dir=0 的位置，
// 但迴圈從 1 開始，那 2 bytes 被拿去當別的旗標用了）。
//
// 因為每列 14 格是偶數，`cell % 2` 就等於 `col % 2`——原版用了這個捷徑。
var hexDelta = [2][7]int{
	// 偶數欄：左下、下、右下、左上、上、右上
	{0, -1, +14, +1, -15, -14, -13},
	// 奇數欄：同樣的六個方向，但因為整欄下移半格，位移不同
	{0, +13, +14, +15, -1, -14, +1},
}

// Neighbour 回傳某格朝某方向的鄰格。
//
// 越界（含左右緣跨列）回傳 NoCell, false。
func (c CellIndex) Neighbour(d HexDir) (CellIndex, bool) {
	if d < 1 || d > 6 || !c.Valid() {
		return NoCell, false
	}
	n := int(c) + hexDelta[int(c)%2][d]
	if n < 0 || n >= CellCount {
		return NoCell, false
	}
	if !Adjacent(CellIndex(n), c) {
		return NoCell, false
	}
	return CellIndex(n), true
}

// Neighbours 回傳某格所有合法的六角鄰格，照方向編號 1..6 的順序。
func (c CellIndex) Neighbours() []CellIndex {
	out := make([]CellIndex, 0, 6)
	for d := DirLowerLeft; d <= DirUpperRight; d++ {
		if n, ok := c.Neighbour(d); ok {
			out = append(out, n)
		}
	}
	return out
}

// Adjacent 回報 a 與 b 是否六角相鄰，語意照 `sub_510E0`。
//
// 上下邊界靠「格編號落在 0..195」擋掉，左右邊界靠欄位檢查擋掉——
// 原版就是這樣分工的，不是兩套都做。所以：
//
//   - 偶數欄且 col == 0：只有 4 個鄰居（左邊兩個會跨到前一列，裁掉）
//   - 奇數欄且 col == 13：只有 4 個鄰居（右邊兩個會跨到下一列，裁掉）
//
// 參數順序與原版一致：b 是基準格，a 是候選。實際上這個關係是對稱的。
func Adjacent(a, b CellIndex) bool {
	if !a.Valid() || !b.Valid() {
		return false
	}
	col := int(b) % assets.GridW
	for d := DirLowerLeft; d <= DirUpperRight; d++ {
		if col == 0 && int(b)%2 == 0 && (d == DirLowerLeft || d == DirUpperLeft) {
			continue // 左緣的偶數欄：左下、左上會跨列
		}
		if col == assets.GridW-1 && int(b)%2 == 1 && (d == DirLowerRight || d == DirUpperRight) {
			continue // 右緣的奇數欄：右下、右上會跨列
		}
		n := int(b) + hexDelta[int(b)%2][d]
		if n >= 0 && n < CellCount && CellIndex(n) == a {
			return true
		}
	}
	return false
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
// 原版還有一關：`sub_55CEC(unit, cell, 1)` 要回 0，也就是**落點周圍六格
// 不能已經站著敵人**。第一輪掃不到就跑第二輪，跳過這一關硬塞。
// 這裡照樣實作兩輪——但敵我判定需要單位的陣營欄位（`+14`，語意未證實），
// 所以 `enemy` 由呼叫端提供。傳 nil 就退化成只看空格。
func (o *Occupancy) Deploy(zone []CellIndex, unit GeneralID, enemy func(GeneralID) bool) (CellIndex, bool) {
	if enemy != nil {
		for _, c := range zone {
			if o[c] == 0 && o.enemyNeighbours(c, enemy) == 0 {
				o[c] = unit
				return c, true
			}
		}
	}
	// 第二輪：放寬成只要空格。
	for _, c := range zone {
		if o[c] == 0 {
			o[c] = unit
			return c, true
		}
	}
	return NoCell, false
}

// enemyNeighbours 數某格周圍六格裡的敵方單位，對應 `sub_55CEC` 的 `arg_0 == 1`。
func (o *Occupancy) enemyNeighbours(c CellIndex, enemy func(GeneralID) bool) int {
	n := 0
	for _, nb := range c.Neighbours() {
		if u := o[nb]; u != 0 && enemy(u) {
			n++
		}
	}
	return n
}

// Move 把單位從現在的格移到相鄰的一格，語意照 `sub_4A1C0`
// 與其呼叫端 `sub_4ABFD` 的三道 gate（`docs/re/07` §7）：
//
//  1. 目標格必須六角相鄰且在界內（`sub_510E0`）
//  2. 目標格必須可進入（`sub_4A583`）：長城段（地物 12..21）擋路，
//     除非單位能穿越；另外目標格不能已經有人站
//  3. 剩餘機動力 `+7` 必須 >= 該步的成本
//
// **成本怎麼算未解**，所以由呼叫端傳進來。原版的成本是一張以方向為索引的
// 表，在移動前算好，推測與目標格的地形有關，但沒有追到填表處。
//
// `bf` 可以傳 nil，那就跳過地形檢查（純狀態轉移測試用）。
func (o *Occupancy) Move(bf *Battlefield, u *CombatUnit, d HexDir, cost uint8) (CellIndex, error) {
	if u.General == 0 || !u.Cell.Valid() {
		return NoCell, fmt.Errorf("game: 單位不在場上，無法移動")
	}
	dst, ok := u.Cell.Neighbour(d)
	if !ok {
		return NoCell, fmt.Errorf("game: 格 %d 的方向 %d 出界", u.Cell, d)
	}
	if o[dst] != 0 {
		return NoCell, fmt.Errorf("game: 格 %d 已被單位 %d 佔住", dst, o[dst])
	}
	if bf != nil && !u.CanCross {
		col, row := dst.ColRow()
		if bf.Tiles[row][col].Kind.Blocks() {
			return NoCell, fmt.Errorf("game: 格 %d 是長城，擋住了", dst)
		}
	}
	if u.Current < cost {
		return NoCell, fmt.Errorf("game: 剩餘機動力 %d 不足 %d", u.Current, cost)
	}
	o[u.Cell] = 0
	u.Cell = dst
	u.Current -= cost
	o[dst] = u.General
	return dst, nil
}
