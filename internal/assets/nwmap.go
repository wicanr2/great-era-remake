package assets

import (
	"encoding/binary"
	"fmt"
)

// 依據：docs/spec/04-battlefield-tiles.md（READY）

// NWMapRecordSize 是 NWMAP.DAT 每省一筆的大小。
//
// 392 = 196 × 2 = 14×14 格，每格一個 u16。這個值來自 WAR.EXE 的
// `Reset(f, 188h)`（sub_4ED9B），不是猜的。
const NWMapRecordSize = 196 * 2

// 每格 u16 的拆法：高 4 bit 是旗標，低 12 bit 是地物編號。
const (
	tileFlagMask = 0xF000
	tileKindMask = 0x0FFF
)

// TileKind 是地物編號，1..22，減 1 就是 NEWTERR.TPC 的圖塊索引。
type TileKind uint16

// 已定名的地物編號。1–5 有直接證據，其餘是圖塊視覺判讀。
const (
	TilePlain    TileKind = 1  // 平原
	TileHill     TileKind = 2  // 丘陵
	TileWater    TileKind = 3  // 水（湖北的長江與漢水全部是 3）
	TileForest   TileKind = 4  // 森林
	TileCity     TileKind = 5  // 城市（39/39 對上 TOWN 的城市數）
	TileMountain TileKind = 6  // 高山
	TileDesert   TileKind = 7  // 沙漠／黃土
	TileBridgeV  TileKind = 8  // 橋（縱）
	TileBridgeH  TileKind = 9  // 橋（橫）
	TileRock     TileKind = 10 // 岩地／荒地
	TileGate     TileKind = 11 // 關隘城門
)

// TileKindMax 是地物編號的上限，也是 NEWTERR.TPC 的圖塊數。
const TileKindMax TileKind = 22

// TileIndex 回傳對應的 NEWTERR.TPC 圖塊索引（0-based）。
func (k TileKind) TileIndex() int {
	if k == 0 || k > TileKindMax {
		return -1
	}
	return int(k) - 1
}

// 鐵路的編碼。原始值大於 TileKindMax 時，這一格有鐵路疊在地形上：
//
//	鐵路圖塊 = (v - railBase) / railStride     0..20，對應 RAIL.TPC 的 21 張
//	底層地形 = (v - railBase) % railStride + 1  1..22，照常查 NEWTERR
//
// railStride 是 25，不是 22——編碼留了幾個空號。實測全 39 省 7,644 格
// 用這個公式拆出來，地形與鐵路索引**全部落在合法範圍，零例外**。
//
// 拆出來的底層地形只有 {1, 2, 3, 7, 10}（平原、丘陵、水、沙漠、岩地），
// 正是鐵路會經過的地形；水的那 77 格就是鐵橋。
const (
	railBase   = 31
	railStride = 25
	// RailTileCount 是 RAIL.TPC 的圖塊數（縱、橫、彎角、T 字、十字、車站）。
	RailTileCount = 21
	// NoRail 表示這一格沒有鐵路。
	NoRail = -1
)

// Tile 是戰場上的一格。
type Tile struct {
	Kind TileKind
	// Rail 是疊在地形上的鐵路圖塊索引（0..20），NoRail 表示沒有鐵路。
	Rail int
	// Flags 是高 4 bit（0x1000/0x2000/0x4000/0x8000）。**語意未解**
	// （docs/spec/04 §4）——城市一定帶旗標，高山與沙漠一定不帶，
	// 但 0x4000 不是兵工廠（39 省只有 9 個吻合，是巧合）。
	Flags uint16
	// Raw 是原始的 u16，寫回時用。
	Raw uint16
}

// HasRail 回報這一格有沒有鐵路。
func (t Tile) HasRail() bool { return t.Rail != NoRail }

// TileGrid 是一個省的 14×14 圖塊。
type TileGrid [GridH][GridW]Tile

// At 取 (x, y) 的格子。
func (g *TileGrid) At(x, y int) Tile { return g[y][x] }

// CountKind 數某種地物有幾格。
func (g *TileGrid) CountKind(k TileKind) int {
	n := 0
	for y := range g {
		for x := range g[y] {
			if g[y][x].Kind == k {
				n++
			}
		}
	}
	return n
}

// ParseNWMap 解 NWMAP.DAT，回傳 39 省的圖塊。
//
// 檔案尾端多出來的 bytes 會被忽略——Turbo Pascal 的 typed file 是按 recsize
// 定位的，檔尾多餘的資料不屬於任何一筆記錄。
func ParseNWMap(data []byte) ([]TileGrid, error) {
	need := ProvinceCount * NWMapRecordSize
	if len(data) < need {
		return nil, fmt.Errorf("assets: NWMAP.DAT 需要至少 %d bytes（%d 省 × %d），只有 %d",
			need, ProvinceCount, NWMapRecordSize, len(data))
	}
	out := make([]TileGrid, ProvinceCount)
	for k := 0; k < ProvinceCount; k++ {
		base := k * NWMapRecordSize
		for y := 0; y < GridH; y++ {
			for x := 0; x < GridW; x++ {
				v := binary.LittleEndian.Uint16(data[base+(y*GridW+x)*2:])
				kind := TileKind(v & tileKindMask)
				rail := NoRail
				if kind > TileKindMax {
					n := int(kind) - railBase
					if n < 0 {
						return nil, fmt.Errorf(
							"assets: 省 %d (%d,%d) 的地物編號 %d 落在 %d..%d 的空號區",
							k+1, x, y, kind, TileKindMax+1, railBase-1)
					}
					rail = n / railStride
					kind = TileKind(n%railStride) + 1
					if rail >= RailTileCount || kind > TileKindMax {
						return nil, fmt.Errorf(
							"assets: 省 %d (%d,%d) 的鐵路編碼 %d 拆出圖塊 %d／地形 %d，超出範圍",
							k+1, x, y, v&tileKindMask, rail, kind)
					}
				}
				out[k][y][x] = Tile{
					Kind:  kind,
					Rail:  rail,
					Flags: v & tileFlagMask,
					Raw:   v,
				}
			}
		}
	}
	return out, nil
}
