package game

import (
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

func newBattlePathTestSim() *BattleSim {
	bf := &Battlefield{}
	for y := 0; y < assets.GridH; y++ {
		for x := 0; x < assets.GridW; x++ {
			// 先全部封成高山，只把測試明確指定的走廊打開；這樣
			// 路徑選擇不會被未設定的旁路干擾。
			bf.Tiles[y][x] = assets.Tile{Kind: assets.TileMountain, Rail: assets.NoRail}
		}
	}
	return &BattleSim{Field: bf, byID: make(map[GeneralID]*Combatant)}
}

func setBattlePathTile(s *BattleSim, c CellIndex, kind assets.TileKind, rail int) {
	col, row := c.ColRow()
	s.Field.Tiles[row][col] = assets.Tile{Kind: kind, Rail: rail}
}

func TestBattlePathWeightUsesConfirmedAIPreferences(t *testing.T) {
	cases := []struct {
		name string
		tile assets.Tile
		want int
	}{
		{"鐵路", assets.Tile{Kind: assets.TilePlain, Rail: 0}, AIPathRailWeight},
		{"河海", assets.Tile{Kind: assets.TileWater, Rail: assets.NoRail}, AIPathWaterWeight},
		{"沙漠", assets.Tile{Kind: assets.TileDesert, Rail: assets.NoRail}, AIPathDesertWeight},
		{"平原", assets.Tile{Kind: assets.TilePlain, Rail: assets.NoRail}, 3},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := battlePathWeight(tc.tile); got != tc.want {
				t.Fatalf("battlePathWeight(%+v) = %d，要 %d", tc.tile, got, tc.want)
			}
		})
	}
}

func TestBattlePathPrefersLowWeightDetour(t *testing.T) {
	s := newBattlePathTestSim()
	from, _ := CellAt(2, 2)
	target, _ := CellAt(6, 2)
	setBattlePathTile(s, from, assets.TilePlain, assets.NoRail)
	setBattlePathTile(s, target, assets.TilePlain, assets.NoRail)

	// 直線只有四步，但三個中間格都是沙漠；上方繞行較長，卻因
	// AI 權重避開沙漠。這是實際矩陣權重，而不是曼哈頓距離。
	direct := []CellIndex{31, 32, 33}
	for _, c := range direct {
		setBattlePathTile(s, c, assets.TileDesert, assets.NoRail)
	}
	detour := []CellIndex{16, 17, 18, 19, 20}
	for _, c := range detour {
		setBattlePathTile(s, c, assets.TilePlain, assets.NoRail)
	}

	path, cost := s.weightedBattlePath(target, from, true)
	if len(path) == 0 {
		t.Fatal("應該能沿上方走廊抵達目標")
	}
	if cost >= 303 { // 直線至少要付 3 + 100×3 + 3
		t.Fatalf("路徑成本 = %d，沒有避開高權重沙漠", cost)
	}
	for _, c := range path {
		for _, desert := range direct {
			if c == desert {
				t.Fatalf("路徑不應穿過沙漠格 %d：%v", c, path)
			}
		}
	}
	if got := s.RouteNextCell(target, from); got != path[0] {
		t.Fatalf("RouteNextCell = %d，完整路徑第一格 = %d", got, path[0])
	}
}

func TestBattlePathSkipsOccupiedIntermediate(t *testing.T) {
	s := newBattlePathTestSim()
	from, _ := CellAt(2, 2)
	target, _ := CellAt(6, 2)
	setBattlePathTile(s, from, assets.TilePlain, assets.NoRail)
	setBattlePathTile(s, target, assets.TilePlain, assets.NoRail)
	for _, c := range []CellIndex{31, 32, 33} {
		setBattlePathTile(s, c, assets.TileDesert, assets.NoRail)
	}
	for _, c := range []CellIndex{16, 17, 18, 19, 20} {
		setBattlePathTile(s, c, assets.TilePlain, assets.NoRail)
	}
	s.Occ[16] = 99

	path, _ := s.weightedBattlePath(target, from, true)
	if len(path) == 0 {
		t.Fatal("中間格被佔用時，仍應有另一條走廊")
	}
	for _, c := range path {
		if c == 16 {
			t.Fatalf("路徑不應把被佔用的中間格當下一跳：%v", path)
		}
	}
	if got := s.RouteNextCell(target, from); got == 16 {
		t.Fatalf("RouteNextCell 不應選被佔用格 16")
	}
}

func TestBattlePathRespectsUnitGreatWallGate(t *testing.T) {
	s := newBattlePathTestSim()
	from, _ := CellAt(2, 2)
	target, _ := CellAt(4, 2)
	wall, _ := CellAt(3, 2)
	setBattlePathTile(s, from, assets.TilePlain, assets.NoRail)
	setBattlePathTile(s, wall, assets.TileGreatWallFirst, assets.NoRail)
	setBattlePathTile(s, target, assets.TilePlain, assets.NoRail)

	s.Occ[from] = 1
	u := &Combatant{CombatUnit: CombatUnit{General: 1, Cell: from, CanCross: false}}
	s.byID[1] = u
	if got := s.RouteNextCell(target, from); got != NoCell {
		t.Fatalf("不可穿越長城的單位不應得到路徑，實際 %d", got)
	}

	u.CanCross = true
	if got := s.RouteNextCell(target, from); got != wall {
		t.Fatalf("可穿越單位應走長城格 %d，實際 %d", wall, got)
	}
}
