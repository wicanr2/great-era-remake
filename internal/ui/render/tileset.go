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
// 邊界格（Owner != 0，屬於鄰省）疊一圈邊框色標示進入方向。
func (c *Canvas) DrawTiledBattlefield(bf *game.Battlefield, ts *TileSet, x, y int) error {
	for gy := 0; gy < assets.GridH; gy++ {
		for gx := 0; gx < assets.GridW; gx++ {
			px, py := x+gx*TileW, y+gy*TileH
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
