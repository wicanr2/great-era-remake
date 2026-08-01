package game

import (
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

func tile(k assets.TileKind) assets.Tile {
	return assets.Tile{Kind: k, Rail: assets.NoRail}
}

// 兩張表逐格對照反組譯。這是純資料，抄錯了只有逐格比才抓得到。
func TestTerrainTablesMatchDisassembly(t *testing.T) {
	branches := []uint8{Branch1, Branch4, Branch5, Branch6}

	def := map[assets.TileKind][4]int{
		assets.TilePlain:   {2, 2, 5, 4},
		assets.TileHill:    {4, 2, 3, 3},
		assets.TileWater:   {1, 1, 1, 1},
		assets.TileForest:  {4, 1, 1, 2},
		assets.TileCity:    {5, 2, 2, 3},
		assets.TileDesert:  {1, 2, 5, 4},
		assets.TileBridgeA: {1, 2, 2, 2},
		assets.TileBridgeB: {1, 2, 2, 2},
		assets.TilePlateau: {3, 2, 5, 4},
		assets.TilePass:    {5, 5, 5, 5},
		assets.TileKindMax: {5, 5, 5, 5},
	}
	for k, row := range def {
		for i, b := range branches {
			got, ok := DefenceFactor(tile(k), b)
			if !ok || got != row[i] {
				t.Errorf("防禦表[地形 %d][兵種 %d] = %d(%v)，應為 %d", k, b, got, ok, row[i])
			}
		}
	}

	atk := map[assets.TileKind][4]int{
		assets.TilePlain:   {3, 5, 5, 3},
		assets.TileHill:    {3, 5, 3, 2},
		assets.TileWater:   {1, 1, 3, 1},
		assets.TileForest:  {3, 1, 1, 1},
		assets.TileCity:    {3, 2, 2, 3},
		assets.TileDesert:  {1, 5, 5, 3},
		assets.TileBridgeA: {3, 3, 3, 3},
		assets.TileBridgeB: {3, 3, 3, 3},
		assets.TilePlateau: {3, 5, 5, 3},
		assets.TilePass:    {3, 5, 5, 3},
		assets.TileKindMax: {3, 5, 5, 3},
	}
	for k, row := range atk {
		for i, b := range branches {
			got, ok := AttackFactor(tile(k), b)
			if !ok || got != row[i] {
				t.Errorf("攻擊表[地形 %d][兵種 %d] = %d(%v)，應為 %d", k, b, got, ok, row[i])
			}
		}
	}
}

// 高山只有兵種 1 有值——而那三個沒值的兵種本來就上不了高山。
// 原版的「漏寫」與移動規則自洽，不是 bug。
func TestMountainOnlyDefinedForBranchOne(t *testing.T) {
	if v, ok := DefenceFactor(tile(assets.TileMountain), Branch1); !ok || v != 4 {
		t.Errorf("高山 × 兵種 1 的防禦係數 = %d(%v)，應為 4", v, ok)
	}
	for _, b := range []uint8{Branch4, Branch5, Branch6} {
		if _, ok := DefenceFactor(tile(assets.TileMountain), b); ok {
			t.Errorf("高山 × 兵種 %d 原版沒設值，不該回報有值", b)
		}
	}
	if tile(assets.TileMountain).MoveCost() != assets.MoveCostImpassable {
		t.Error("高山應該不可通行——兩件事要對得起來")
	}
}

// 鐵路覆蓋地形：兩張表都是第一件事就問鐵路。
func TestRailOverridesTerrainInBothTables(t *testing.T) {
	railed := assets.Tile{Kind: assets.TileWater, Rail: 7}
	if v, _ := DefenceFactor(railed, Branch1); v != 2 {
		t.Errorf("鐵路上的防禦係數 = %d，應為 2", v)
	}
	if v, _ := AttackFactor(railed, Branch1); v != 3 {
		t.Errorf("鐵路上的攻擊係數 = %d，應為 3", v)
	}
	for _, b := range []uint8{Branch1, Branch4, Branch5, Branch6} {
		if v, _ := DefenceFactor(railed, b); v != 2 {
			t.Errorf("鐵路防禦不該分兵種，兵種 %d 得到 %d", b, v)
		}
	}
	if v, _ := AttackFactor(railed, Branch4); v != 2 {
		t.Errorf("鐵路上兵種 4 的攻擊係數 = %d，應為 2", v)
	}
}

// 長城：兵種 1 以外的攻擊係數是 0——站上去打不出傷害。
func TestGreatWallZeroesAttackForMostBranches(t *testing.T) {
	for k := assets.TileGreatWallFirst; k <= assets.TileGreatWallLast; k++ {
		if v, _ := DefenceFactor(tile(k), Branch1); v != 5 {
			t.Errorf("長城 %d 對兵種 1 的防禦 = %d，應為 5", k, v)
		}
		for _, b := range []uint8{Branch4, Branch5, Branch6} {
			if v, _ := AttackFactor(tile(k), b); v != 0 {
				t.Errorf("長城 %d 對兵種 %d 的攻擊 = %d，應為 0", k, b, v)
			}
		}
	}
	wall := tile(assets.TileGreatWallFirst)
	if got := AttackValue(10000, wall, Branch5, tile(assets.TilePlain), Branch1); got != 0 {
		t.Errorf("長城上的兵種 5 攻擊值 = %d，應為 0", got)
	}
}

// 兩個係數的換算：防禦當分母、攻擊當乘數。
func TestCoefficientConversion(t *testing.T) {
	if got := DefenceCoefficient(tile(assets.TileCity), Branch1); got != 2 {
		t.Errorf("城市 × 兵種 1 的防禦係數 = %v，應為 2.0", got)
	}
	if got := DefenceCoefficient(tile(assets.TileWater), Branch1); got != 1 {
		t.Errorf("河海 × 兵種 1 的防禦係數 = %v，應為 1.0", got)
	}
	if got := AttackCoefficient(tile(assets.TilePlain), Branch4); got != 2.5 {
		t.Errorf("平原 × 兵種 4 的攻擊係數 = %v，應為 2.5", got)
	}
}

// 攻守各看自己站的格——這條公式最容易搞錯的地方。
func TestAttackValueUsesEachSideOwnTerrain(t *testing.T) {
	const s = 1000
	plain, city := tile(assets.TilePlain), tile(assets.TileCity)

	flat := AttackValue(s, plain, Branch4, plain, Branch1)
	inCity := AttackValue(s, plain, Branch4, city, Branch1)
	if inCity >= flat {
		t.Errorf("守方進城後攻擊值 = %d，應低於平原的 %d", inCity, flat)
	}
	if want := round(float64(s) * 2.5 / 2.0); inCity != want {
		t.Errorf("攻擊值 = %d，應為 %d", inCity, want)
	}

	inForest := AttackValue(s, tile(assets.TileForest), Branch4, plain, Branch1)
	if inForest >= flat {
		t.Errorf("攻方在森林的攻擊值 = %d，應低於平原的 %d", inForest, flat)
	}
}
