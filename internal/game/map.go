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

// Terrain 是地形編號，0..22。
//
// **語意未解**（docs/spec/01 §4）：目前只知道值域，不知道哪個數字是山、河、平原。
// 在查明之前一律以原始編號傳遞，不做語意化命名，也不要自行賦予防禦修正。
type Terrain byte

// Battlefield 是一個省的 14×14 戰場格子。
type Battlefield struct {
	// Owner[y][x] 是該格所屬的省編號；0 表示本省腹地。
	// 非零格代表邊界，也就是敵方可能的進入方向。
	Owner [assets.GridH][assets.GridW]ProvinceID
	// Terrain[y][x] 是地形編號。
	Terrain [assets.GridH][assets.GridW]Terrain
}

// Map 是整張戰略地圖：39 個省的鄰接關係與各自的戰場格子。
type Map struct {
	adjacency [][]ProvinceID
	fields    []Battlefield
}

// LoadMap 從 WARPOS.DAT 與 TERNAME.DAT 建出地圖。
//
// 鄰接表是**從檔案導出**的，不在程式碼裡重打一份——重抄的那份一定會漂
// （CLAUDE.md §7 第 4 條）。
func LoadMap(warpos, tername []byte) (*Map, error) {
	pos, err := assets.ParseProvinceGrids(warpos)
	if err != nil {
		return nil, fmt.Errorf("game: WARPOS.DAT: %w", err)
	}
	ter, err := assets.ParseProvinceGrids(tername)
	if err != nil {
		return nil, fmt.Errorf("game: TERNAME.DAT: %w", err)
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

		for y := 0; y < assets.GridH; y++ {
			for x := 0; x < assets.GridW; x++ {
				m.fields[k].Owner[y][x] = ProvinceID(pos[k].At(x, y))
				m.fields[k].Terrain[y][x] = Terrain(ter[k].At(x, y))
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
// **這是假說**（docs/spec/01 §2）：只用單一樣本驗證過——實機顯示河南(19) 可攻打
// 11,16,18,20,21,22，而鄰接表是這六個再加湖北(26)，26 正是當時玩家控制的省。
// 要換一個控制不同省份的存檔再驗。
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
