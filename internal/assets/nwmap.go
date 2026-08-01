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

// 地物編號 1–11 的名稱**出自原版自己的詞表**，不是視覺判讀：
// `2.15` 的詞條 195–205 依序是「平原 丘陵 河海 森林 城市 高山 沙漠
// 橋樑 橋樑 高原 關口」，前面第 194 條正是「地形」。
//
// 順序與編號 1–11 一一對應，連**橋樑重複兩次**（縱、橫兩種）都對上——
// 這是 confirmed，不是推測。
//
// 12–22 沒有對應的詞條，視覺上是長城的各段（橫、縱、彎角、斜坡）；
// `2.15` 詞條 120 有「長城」，但沒有逐段的名稱。
const (
	TilePlain    TileKind = 1  // 平原
	TileHill     TileKind = 2  // 丘陵
	TileWater    TileKind = 3  // 河海（湖北的長江與漢水全部是 3）
	TileForest   TileKind = 4  // 森林
	TileCity     TileKind = 5  // 城市（39/39 對上 TOWN 的城市數）
	TileMountain TileKind = 6  // 高山
	TileDesert   TileKind = 7  // 沙漠
	TileBridgeA  TileKind = 8  // 橋樑（縱）
	TileBridgeB  TileKind = 9  // 橋樑（橫）
	TilePlateau  TileKind = 10 // 高原
	TilePass     TileKind = 11 // 關口
)

// TileName 回傳地物的原版名稱。12–22（長城各段）與鐵路沒有詞條名稱。
func (k TileKind) TileName() string {
	switch k {
	case TilePlain:
		return "平原"
	case TileHill:
		return "丘陵"
	case TileWater:
		return "河海"
	case TileForest:
		return "森林"
	case TileCity:
		return "城市"
	case TileMountain:
		return "高山"
	case TileDesert:
		return "沙漠"
	case TileBridgeA, TileBridgeB:
		return "橋樑"
	case TilePlateau:
		return "高原"
	case TilePass:
		return "關口"
	}
	if k >= 12 && k <= TileKindMax {
		return "長城"
	}
	return ""
}

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
// 拆出來的底層地形只有 {1, 2, 3, 7, 10}（平原、丘陵、河海、沙漠、高原），
// 正是鐵路會經過的地形；河海的那 77 格就是鐵橋。
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
