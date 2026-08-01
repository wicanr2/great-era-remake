// Package game 是規則層：只認識遊戲規則與狀態，不認識畫面。
//
// 這一層不得 import Ebiten 或任何繪圖套件（CLAUDE.md §11）——否則無頭環境
// 連測都測不了。呈現層在 internal/ui。
package game

import (
	"fmt"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

// 依據：docs/spec/01-map-and-terrain.md（READY）

// ProvinceID 是省編號，1-based，1..39。順序與 3.15 的省名表一致。
type ProvinceID int

// ProvinceCount 是省份數。
const ProvinceCount = assets.ProvinceCount

// Valid 回報這個省編號是否在合法範圍。
func (p ProvinceID) Valid() bool { return p >= 1 && p <= ProvinceCount }

// PlaceIndex 是格子上的**地名索引**，1-based；0 表示這一格沒有地名。
//
// [訂正] 這個欄位一度被當成「地形編號 0..22」（docs/spec/01 舊版）。
// 實際上 k > 0 代表該格是本省地名表 `TN15.<省編號>` 的第 k 個詞條。
// 證據：全 39 省的最大值都**恰好等於**該省地名表的詞條數，零例外
// （廣西 22 個地名 → 最大值 22；海南島 7 個 → 最大值 7）。
// 舊說的「值域 0..22」只是「地名最多的省有 22 個地名」的副作用。
//
// 地形性質由地名的種類隱含：湖北的 9 號是長江，佔 33 格；6 號是大別山，
// 佔 21 格；漢口、武昌這些城市各佔 1 格。
type PlaceIndex byte

// Battlefield 是一個省的 14×14 戰場格子。
type Battlefield struct {
	// Owner[y][x] 是該格所屬的省編號；0 表示本省腹地。
	// 非零格代表邊界，也就是敵方可能的進入方向。
	Owner [assets.GridH][assets.GridW]ProvinceID
	// Place[y][x] 是地名索引，指向該省的 TN15 地名表。
	Place [assets.GridH][assets.GridW]PlaceIndex
	// Tiles 是每格的地物（NWMAP.DAT，SPEC-04）：平原、水、城市、城牆…
	// 地名說「這一格叫什麼」，地物說「這一格該畫成什麼」。
	Tiles assets.TileGrid
}

// MaxPlace 回傳這張戰場用到的最大地名索引，也就是該省的地名數。
func (b *Battlefield) MaxPlace() PlaceIndex {
	var mx PlaceIndex
	for y := range b.Place {
		for _, v := range b.Place[y] {
			if v > mx {
				mx = v
			}
		}
	}
	return mx
}

// Map 是整張戰略地圖：39 個省的鄰接關係與各自的戰場格子。
type Map struct {
	adjacency [][]ProvinceID
	fields    []Battlefield
}

// LoadMap 從 WARPOS.DAT、TERNAME.DAT 與 NWMAP.DAT 建出地圖。
//
// 鄰接表是**從檔案導出**的，不在程式碼裡重打一份——重抄的那份一定會漂
// （CLAUDE.md §7 第 4 條）。
func LoadMap(warpos, tername, nwmap []byte) (*Map, error) {
	pos, err := assets.ParseProvinceGrids(warpos)
	if err != nil {
		return nil, fmt.Errorf("game: WARPOS.DAT: %w", err)
	}
	ter, err := assets.ParseProvinceGrids(tername)
	if err != nil {
		return nil, fmt.Errorf("game: TERNAME.DAT: %w", err)
	}

	tiles, err := assets.ParseNWMap(nwmap)
	if err != nil {
		return nil, fmt.Errorf("game: NWMAP.DAT: %w", err)
	}

	if bad := pos.IsSymmetric(); len(bad) != 0 {
		return nil, fmt.Errorf("game: 鄰接關係有 %d 個不對稱，第一個是 %v", len(bad), bad[0])
	}

	m := &Map{
		adjacency: make([][]ProvinceID, ProvinceCount),
		fields:    make([]Battlefield, ProvinceCount),
	}
	for k := 0; k < ProvinceCount; k++ {
		ns, err := pos.Neighbours(k + 1)
		if err != nil {
			return nil, err
		}
		ids := make([]ProvinceID, len(ns))
		for i, v := range ns {
			ids[i] = ProvinceID(v)
		}
		m.adjacency[k] = ids
		m.fields[k].Tiles = tiles[k]

		for y := 0; y < assets.GridH; y++ {
			for x := 0; x < assets.GridW; x++ {
				m.fields[k].Owner[y][x] = ProvinceID(pos[k].At(x, y))
				m.fields[k].Place[y][x] = PlaceIndex(ter[k].At(x, y))
			}
		}
	}
	return m, nil
}

// Neighbours 回傳某省的鄰省，遞增排序。回傳的是內部切片的副本。
func (m *Map) Neighbours(p ProvinceID) ([]ProvinceID, error) {
	if !p.Valid() {
		return nil, fmt.Errorf("game: 省編號 %d 超出 1..%d", p, ProvinceCount)
	}
	src := m.adjacency[p-1]
	out := make([]ProvinceID, len(src))
	copy(out, src)
	return out, nil
}

// Adjacent 回報兩個省是否相鄰。
func (m *Map) Adjacent(a, b ProvinceID) bool {
	if !a.Valid() || !b.Valid() {
		return false
	}
	for _, v := range m.adjacency[a-1] {
		if v == b {
			return true
		}
	}
	return false
}

// Battlefield 回傳某省的戰場格子。
func (m *Map) Battlefield(p ProvinceID) (*Battlefield, error) {
	if !p.Valid() {
		return nil, fmt.Errorf("game: 省編號 %d 超出 1..%d", p, ProvinceCount)
	}
	return &m.fields[p-1], nil
}

// AttackTargets 回傳從 from 出發可攻打的省。
//
// 規則：鄰省扣掉 owned 裡已控制的省。
//
// 兩個獨立樣本驗證過（docs/spec/01 §2）：玩家同時控制河南(19) 與湖北(26)，
// 從這兩省出發時，實機的攻打候選都正好等於鄰接表扣掉另一個己方省。
//
// 未涵蓋：跨海（臺灣、海南島）能不能攻打、盟省算不算 owned。
func (m *Map) AttackTargets(from ProvinceID, owned map[ProvinceID]bool) ([]ProvinceID, error) {
	ns, err := m.Neighbours(from)
	if err != nil {
		return nil, err
	}
	out := make([]ProvinceID, 0, len(ns))
	for _, v := range ns {
		if !owned[v] {
			out = append(out, v)
		}
	}
	return out, nil
}
