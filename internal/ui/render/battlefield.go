package render

import (
	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
)

// 戰場格子的視覺化。
//
// 依據 docs/spec/01-map-and-terrain.md（READY）：每省一張 14×14 格，
// WARPOS 的格值是省編號（0 = 本省腹地）、TERNAME 的格值是地形編號（0-22）。
//
// **地形編號的語意未解**，所以這裡只做「不同編號給不同顏色」的中性呈現，
// 不做「山=灰、河=藍」這種語意化上色——那會是編出來的
// （CLAUDE.md §9：不准為了讓東西看起來完整而編規則）。

// CellSize 是一格畫多大。原版戰場的實際格子尺寸未知，
// 這是 remake 的呈現選擇，不是還原值。
const CellSize = 16

// terrainShades 給 0-22 的地形編號各一個可區分的顏色。
//
// 純粹是為了看得出分區，**沒有語意**。查明語意之前不要照這個配色下結論。
var terrainShades = [23]assets.RGB{
	{R: 20, G: 20, B: 30}, {R: 200, G: 180, B: 120}, {R: 170, G: 150, B: 100},
	{R: 140, G: 190, B: 90}, {R: 110, G: 160, B: 70}, {R: 70, G: 120, B: 200},
	{R: 60, G: 100, B: 170}, {R: 190, G: 170, B: 90}, {R: 150, G: 130, B: 70},
	{R: 120, G: 110, B: 100}, {R: 100, G: 90, B: 80}, {R: 90, G: 140, B: 90},
	{R: 80, G: 120, B: 80}, {R: 160, G: 140, B: 110}, {R: 130, G: 115, B: 90},
	{R: 200, G: 120, B: 80}, {R: 180, G: 100, B: 70}, {R: 210, G: 200, B: 190},
	{R: 190, G: 180, B: 170}, {R: 160, G: 60, B: 60}, {R: 140, G: 50, B: 50},
	{R: 220, G: 220, B: 120}, {R: 200, G: 200, B: 100},
}

// borderColor 是邊界格（屬於鄰省）的疊色。
var borderColor = assets.RGB{R: 220, G: 60, B: 60}

// DrawBattlefield 把一個省的戰場格子畫到 (x, y)。
//
// 每格先依地形上底色，屬於鄰省的格子（Owner != 0）再疊一層邊框色，
// 讓「敵方可能的進入方向」看得出來。
func (c *Canvas) DrawBattlefield(bf *game.Battlefield, x, y int) {
	for gy := 0; gy < assets.GridH; gy++ {
		for gx := 0; gx < assets.GridW; gx++ {
			t := bf.Terrain[gy][gx]
			col := terrainShades[0]
			if int(t) < len(terrainShades) {
				col = terrainShades[t]
			}
			c.fillCell(x+gx*CellSize, y+gy*CellSize, col)
			if bf.Owner[gy][gx] != 0 {
				c.strokeCell(x+gx*CellSize, y+gy*CellSize, borderColor)
			}
		}
	}
}

func (c *Canvas) fillCell(px, py int, col assets.RGB) {
	for dy := 0; dy < CellSize; dy++ {
		for dx := 0; dx < CellSize; dx++ {
			c.setPixel(px+dx, py+dy, col)
		}
	}
}

// strokeCell 在格子邊緣畫一圈，用來標示邊界格。
func (c *Canvas) strokeCell(px, py int, col assets.RGB) {
	for d := 0; d < CellSize; d++ {
		c.setPixel(px+d, py, col)
		c.setPixel(px+d, py+CellSize-1, col)
		c.setPixel(px, py+d, col)
		c.setPixel(px+CellSize-1, py+d, col)
	}
}

func (c *Canvas) setPixel(x, y int, col assets.RGB) {
	b := c.img.Bounds()
	if x < b.Min.X || x >= b.Max.X || y < b.Min.Y || y >= b.Max.Y {
		return
	}
	c.img.SetRGBA(x, y, colorOf(col))
}
