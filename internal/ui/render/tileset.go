package render

import (
	"fmt"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
)

// 戰場圖塊的繪製。依據 docs/spec/04-battlefield-tiles.md（READY）
// 與 docs/formats/05-tpc-tilesets.md。
//
// `NWMAP.DAT` 的地物編號 1..22 減 1 就是 `NEWTERR.TPC` 的圖塊索引，
// 所以這一層不需要任何對照表——直接拿編號去取圖。

// TileW / TileH 是一張地形圖塊的尺寸。
const TileW, TileH = 32, 24

// TileSet 是解好的地形圖塊組，含疊在上面的鐵路。
type TileSet struct {
	Tiles []*assets.Image
	Rails assets.RailTiles
	Pal   assets.Palette
}

// LoadTileSet 解 NEWTERR.TPC。
//
// 圖集的結構是 [圖][palt 20 B] 反覆，DecodeBGISet 已經處理。
// 預期 22 張 32×24——數量或尺寸不對就報錯，不要默默接受。
func LoadTileSet(newterr, rail []byte, pal assets.Palette) (*TileSet, error) {
	ims, err := assets.DecodeBGISet(newterr)
	if err != nil {
		return nil, fmt.Errorf("render: NEWTERR.TPC: %w", err)
	}
	if len(ims) != int(assets.TileKindMax) {
		return nil, fmt.Errorf("render: NEWTERR.TPC 應有 %d 張圖塊，解出 %d 張",
			assets.TileKindMax, len(ims))
	}
	for i, im := range ims {
		if im.W != TileW || im.H != TileH {
			return nil, fmt.Errorf("render: 第 %d 張圖塊是 %dx%d，預期 %dx%d",
				i, im.W, im.H, TileW, TileH)
		}
	}
	rails, err := assets.ParseRailTiles(rail)
	if err != nil {
		return nil, fmt.Errorf("render: RAIL.TPC: %w", err)
	}
	return &TileSet{Tiles: ims, Rails: rails, Pal: pal}, nil
}

// DrawTiledBattlefield 用原版圖塊把一個省的戰場畫到 (x, y)。
//
// 版面照原版的**六角格**排列——奇數欄下移半格（`game.CellIndex.ScreenXY`，
// 出自 `sub_50FF5`）。這裡曾經是方格排列，那是排版猜測，不是原版。
//
// 邊界格（Owner != 0，屬於鄰省）疊一圈邊框色標示進入方向。
func (c *Canvas) DrawTiledBattlefield(bf *game.Battlefield, ts *TileSet, x, y int) error {
	for gy := 0; gy < assets.GridH; gy++ {
		for gx := 0; gx < assets.GridW; gx++ {
			cell, err := game.CellAt(gx, gy)
			if err != nil {
				return err
			}
			dx, dy := cell.ScreenXY()
			px, py := x+dx, y+dy
			t := bf.Tiles[gy][gx]
			idx := t.Kind.TileIndex()
			if idx < 0 {
				return fmt.Errorf("render: 省內 (%d,%d) 的地物編號 %d 沒有對應圖塊",
					gx, gy, t.Kind)
			}
			if err := c.DrawBGI(ts.Tiles[idx], ts.Pal, px, py); err != nil {
				return err
			}
			// 鐵路疊在地形上，索引 0 透明。
			if t.HasRail() {
				c.drawRail(ts.Rails[t.Rail], ts.Pal, px, py)
			}
			if bf.Owner[gy][gx] != 0 {
				c.strokeRect(px, py, TileW, TileH, borderColor)
			}
		}
	}
	return nil
}

// drawRail 把一張鐵路圖塊疊上去，索引 0 視為透明。
func (c *Canvas) drawRail(pix []byte, pal assets.Palette, px, py int) {
	for y := 0; y < assets.RailH; y++ {
		for x := 0; x < assets.RailW; x++ {
			v := pix[y*assets.RailW+x]
			if v == 0 || int(v) >= len(pal) {
				continue
			}
			c.setPixel(px+x, py+y, pal[v])
		}
	}
}

func (c *Canvas) fillRect(px, py, w, h int, col assets.RGB) {
	for dy := 0; dy < h; dy++ {
		for dx := 0; dx < w; dx++ {
			c.setPixel(px+dx, py+dy, col)
		}
	}
}

func (c *Canvas) strokeRect(px, py, w, h int, col assets.RGB) {
	for dx := 0; dx < w; dx++ {
		c.setPixel(px+dx, py, col)
		c.setPixel(px+dx, py+h-1, col)
	}
	for dy := 0; dy < h; dy++ {
		c.setPixel(px, py+dy, col)
		c.setPixel(px+w-1, py+dy, col)
	}
}

// 兵種圖示（NEWICON.TPC，18 張 32×17）。
//
// 綠紅成對，應該就是敵我兩方（docs/formats/05 §3）：
//
//	0/1   鋼盔（步兵）綠／紅
//	2/3   戰車（裝甲）綠／紅
//	4/5   馬（騎兵）綠／紅
//	6-11  飛機，綠，六個朝向
//	12-17 飛機，紅，六個朝向
//
// ⚠️ 圖示編號與遊戲裡的兵種編號**還沒對應起來**——詞表有步兵、砲兵、
// 騎兵、裝甲兵、游擊隊五個兵種（docs/mechanics/20-military.md §5），
// 但圖示裡看不到明顯的砲兵。所以這裡只提供「畫第 N 張」，不做語意映射。

// IconW / IconH 是一張兵種圖示的尺寸。
const IconW, IconH = 32, 17

// IconCount 是 NEWICON.TPC 的圖示數。
const IconCount = 18

// LoadIcons 解 NEWICON.TPC。
func LoadIcons(newicon []byte) ([]*assets.Image, error) {
	ims, err := assets.DecodeBGISet(newicon)
	if err != nil {
		return nil, fmt.Errorf("render: NEWICON.TPC: %w", err)
	}
	if len(ims) != IconCount {
		return nil, fmt.Errorf("render: NEWICON.TPC 應有 %d 張，解出 %d 張",
			IconCount, len(ims))
	}
	for i, im := range ims {
		if im.W != IconW || im.H != IconH {
			return nil, fmt.Errorf("render: 第 %d 張圖示是 %dx%d，預期 %dx%d",
				i, im.W, im.H, IconW, IconH)
		}
	}
	return ims, nil
}

// DrawUnitIcon 把第 idx 張兵種圖示疊到格子 (gx, gy) 上。
//
// 圖示比格子矮（17 vs 24），所以垂直置中；索引 0 視為透明，
// 讓底下的地形透出來。
func (c *Canvas) DrawUnitIcon(icons []*assets.Image, idx int, pal assets.Palette,
	originX, originY, gx, gy int) error {
	if idx < 0 || idx >= len(icons) {
		return fmt.Errorf("render: 圖示編號 %d 超出 0..%d", idx, len(icons)-1)
	}
	im := icons[idx]
	px := originX + gx*TileW
	py := originY + gy*TileH + (TileH-IconH)/2
	for y := 0; y < im.H; y++ {
		for x := 0; x < im.W; x++ {
			v := im.Pix[y*im.W+x]
			if v == 0 || int(v) >= len(pal) {
				continue
			}
			c.setPixel(px+x, py+y, pal[v])
		}
	}
	return nil
}

// DrawUnitAtCell 把兵種圖示疊到六角格 cell 上。
//
// 與 DrawUnitIcon 的差別是座標用 `game.CellIndex.ScreenXY()`
// ——奇數欄下移半格（`docs/re/07` §3），不是方格排列。
func (c *Canvas) DrawUnitAtCell(icons []*assets.Image, idx int, pal assets.Palette,
	originX, originY int, cell game.CellIndex) error {
	if idx < 0 || idx >= len(icons) {
		return fmt.Errorf("render: 圖示編號 %d 超出 0..%d", idx, len(icons)-1)
	}
	dx, dy := cell.ScreenXY()
	im := icons[idx]
	px, py := originX+dx, originY+dy+(TileH-IconH)/2
	for y := 0; y < im.H; y++ {
		for x := 0; x < im.W; x++ {
			v := im.Pix[y*im.W+x]
			if v == 0 || int(v) >= len(pal) {
				continue
			}
			c.setPixel(px+x, py+y, pal[v])
		}
	}
	return nil
}

// DrawCellCursor 在六角格上畫一圈框，標示游標或選中的單位。
func (c *Canvas) DrawCellCursor(originX, originY int, cell game.CellIndex, col assets.RGB) {
	dx, dy := cell.ScreenXY()
	c.strokeRect(originX+dx, originY+dy, TileW, TileH, col)
}

// 兵種 → 圖示的對應。四個兵種全部對上（`docs/formats/05` §3）：
//
//	步兵   → 鋼盔  0 / 1
//	裝甲兵 → 戰車  2 / 3
//	騎兵   → 馬頭  4 / 5
//	砲兵   → 大砲  6–11 / 12–17（六個朝向）
//
// 綠是第一方、紅是第二方。
const (
	iconInfantry = 0 // 綠，+1 是紅
	iconArmour   = 2
	iconCavalry  = 4
	iconArtiller = 6 // 綠的六個朝向 6..11，紅的 12..17
)

// BranchIcon 回傳某個兵種該畫哪一張圖示。
//
// `red` 為 true 取紅色那一份。`facing` 只有砲兵用得到——它有六個朝向
// （`+31`，`docs/re/09`），其餘兵種忽略。facing 超出 1..6 一律當 1。
func BranchIcon(branch uint8, red bool, facing uint8) int {
	if facing < 1 || facing > 6 {
		facing = 1
	}
	switch branch {
	case game.BranchInfantry:
		return iconInfantry + boolToInt(red)
	case game.BranchArmour:
		return iconArmour + boolToInt(red)
	case game.BranchCavalry:
		return iconCavalry + boolToInt(red)
	case game.BranchArtiller:
		// 砲兵佔 12 張：綠 6..11、紅 12..17。
		base := iconArtiller
		if red {
			base += 6
		}
		return base + int(facing) - 1
	}
	return iconInfantry + boolToInt(red)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
