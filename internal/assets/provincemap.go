package assets

import "fmt"

// `WARPOS.DAT` 與 `TERNAME.DAT`：每省一張 14×14 格子圖。
// 規格見 docs/spec/01-map-and-terrain.md（READY）。
//
//	WARPOS.DAT   格值 = 省編號（1-39），0 = 本省腹地。用來導鄰接表與戰場邊界
//	TERNAME.DAT  格值 = 地形編號（0-22），語意未解，保留原始編號不做語意化命名

const (
	// ProvinceCount 是省份數。39 這個數字同時出現在 TN15.1-39、3.15 的省名表、
	// 以及 WARPOS/TERNAME/MEM_WAR 三份資料表，是全遊戲的核心維度。
	ProvinceCount = 39

	// GridW / GridH 是每省戰場格子圖的尺寸。196 = 14×14，
	// 記錄大小來自 WAR.EXE 的 Reset(f, 196)，不是猜的。
	GridW = 14
	GridH = 14

	gridSize = GridW * GridH // 196
)

// Grid 是一張 14×14 的格子圖，row-major：index = y*GridW + x。
type Grid [gridSize]byte

// At 取 (x, y) 的格值。超出範圍回 0。
func (g *Grid) At(x, y int) byte {
	if x < 0 || x >= GridW || y < 0 || y >= GridH {
		return 0
	}
	return g[y*GridW+x]
}

// ProvinceGrids 是 39 個省各一張格子圖。索引 0-based，省編號 = 索引 + 1。
type ProvinceGrids [ProvinceCount]Grid

// ParseProvinceGrids 解析 WARPOS.DAT 或 TERNAME.DAT。
func ParseProvinceGrids(data []byte) (*ProvinceGrids, error) {
	want := ProvinceCount * gridSize
	if len(data) != want {
		return nil, fmt.Errorf("assets: 省份格子圖應為 %d bytes（%d 省 × %d），實為 %d",
			want, ProvinceCount, gridSize, len(data))
	}
	out := new(ProvinceGrids)
	for k := 0; k < ProvinceCount; k++ {
		copy(out[k][:], data[k*gridSize:])
	}
	return out, nil
}

// Neighbours 從 WARPOS 的格子圖導出第 province 省（1-based）的鄰省清單，遞增排序。
//
// 規則：格子圖上出現的其他省編號就是鄰省，本省的格子填 0。
// 導出的結果對 39 個省完全對稱（A→B 必然 B→A），這是格式判定的強驗證。
func (p *ProvinceGrids) Neighbours(province int) ([]int, error) {
	if province < 1 || province > ProvinceCount {
		return nil, fmt.Errorf("assets: 省編號 %d 超出 1..%d", province, ProvinceCount)
	}
	seen := make(map[byte]bool)
	for _, v := range p[province-1] {
		if v >= 1 && v <= ProvinceCount && int(v) != province {
			seen[v] = true
		}
	}
	out := make([]int, 0, len(seen))
	for v := 1; v <= ProvinceCount; v++ {
		if seen[byte(v)] {
			out = append(out, v)
		}
	}
	return out, nil
}

// AdjacencyList 導出全部 39 省的鄰接表，索引 0-based（第 k 項是省 k+1 的鄰省）。
func (p *ProvinceGrids) AdjacencyList() [][]int {
	out := make([][]int, ProvinceCount)
	for k := 1; k <= ProvinceCount; k++ {
		out[k-1], _ = p.Neighbours(k)
	}
	return out
}

// IsSymmetric 檢查鄰接關係是否對稱。回傳所有 A→B 但 B↛A 的配對。
//
// 這是驗證解析正確與否最有力的檢查：任何一個 off-by-one 或
// row/column 搞反都會讓對稱性破裂。
func (p *ProvinceGrids) IsSymmetric() [][2]int {
	adj := p.AdjacencyList()
	has := func(a, b int) bool {
		for _, v := range adj[a-1] {
			if v == b {
				return true
			}
		}
		return false
	}
	var bad [][2]int
	for a := 1; a <= ProvinceCount; a++ {
		for _, b := range adj[a-1] {
			if !has(b, a) {
				bad = append(bad, [2]int{a, b})
			}
		}
	}
	return bad
}
