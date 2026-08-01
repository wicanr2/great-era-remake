package render

import (
	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
)

// 戰場格子的視覺化。
//
// 依據 docs/spec/01-map-and-terrain.md：每省一張 14×14 格，
// `WARPOS` 的格值是省編號（0 = 本省腹地），`TERNAME` 的格值是**地名索引**
// ——指向該省的 `TN15.<省編號>` 地名表，不是地形編號（見 §4 的訂正）。
//
// 原版怎麼把地名索引畫成地形圖塊（`NEWTERR.TPC` 的 22 張）還沒解出來：
// 缺的是「地名 → 圖塊」的對應，湖北的「長江」該用水、「大別山」該用山，
// 但那份對照表還沒找到。**在找到之前這裡不猜**——每個地名給一個可區分的
// 顏色，讓格子的分區看得出來，但不宣稱哪一格是山、哪一格是水。

// CellSize 是一格畫多大。原版戰場的實際格子尺寸未知，
// 這是 remake 的呈現選擇，不是還原值。
const CellSize = 16

// plainColor 是沒有地名的格子（腹地）的底色。
var plainColor = assets.RGB{R: 60, G: 110, B: 60}

// placeShades 給地名索引 1..22 各一個可區分的顏色。
//
// 純粹是為了看得出哪些格子屬於同一個地名，**沒有地形語意**。
var placeShades = [22]assets.RGB{
	{R: 200, G: 180, B: 120}, {R: 170, G: 150, B: 100}, {R: 140, G: 190, B: 90},
	{R: 110, G: 160, B: 70}, {R: 70, G: 120, B: 200}, {R: 60, G: 100, B: 170},
	{R: 190, G: 170, B: 90}, {R: 150, G: 130, B: 70}, {R: 120, G: 110, B: 100},
	{R: 100, G: 90, B: 80}, {R: 90, G: 140, B: 90}, {R: 80, G: 120, B: 80},
	{R: 160, G: 140, B: 110}, {R: 130, G: 115, B: 90}, {R: 200, G: 120, B: 80},
	{R: 180, G: 100, B: 70}, {R: 210, G: 200, B: 190}, {R: 190, G: 180, B: 170},
	{R: 160, G: 60, B: 60}, {R: 140, G: 50, B: 50}, {R: 220, G: 220, B: 120},
	{R: 200, G: 200, B: 100},
}

// PlaceColor 回傳地名索引對應的格子顏色。
func PlaceColor(p game.PlaceIndex) assets.RGB {
	if p == 0 || int(p) > len(placeShades) {
		return plainColor
	}
	return placeShades[p-1]
}

// borderColor 是邊界格（屬於鄰省）的疊色。
var borderColor = assets.RGB{R: 220, G: 60, B: 60}

// DrawBattlefield 把一個省的戰場格子畫到 (x, y)。
//
// 每格先依地名上底色，屬於鄰省的格子（Owner != 0）再疊一層邊框色，
// 讓「敵方可能的進入方向」看得出來。
func (c *Canvas) DrawBattlefield(bf *game.Battlefield, x, y int) {
	for gy := 0; gy < assets.GridH; gy++ {
		for gx := 0; gx < assets.GridW; gx++ {
			c.fillCell(x+gx*CellSize, y+gy*CellSize, PlaceColor(bf.Place[gy][gx]))
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
