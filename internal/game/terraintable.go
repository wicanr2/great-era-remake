package game

import "github.com/wicanr2/great-era-remake/internal/assets"

// 地形 × 兵種的兩張對照表，出自 `sub_503BB` 與 `sub_501A3`
// （`docs/re/08` §6）。兩支都是三百多行的比較樹，內容就是這兩張表。
//
// 它們餵給 `sub_51D68` 的兩個修正係數：
//
//	sub_50655(u) = 1.0 + 0.25 × (防禦表[u.格][u.兵種] − 1)
//	sub_50600(u) = 0.5  × 攻擊表[u.格][u.兵種]
//
// 而戰力的算法是
//
//	F 對 E 的攻擊值 = 強度(F) × sub_50600(F) ÷ sub_50655(E)
//
// ——**攻擊方看自己站的地形，防守方看自己站的地形**，一個當分子一個當分母。

// 兵種代號。**哪個號碼是哪個兵種沒有證據**（`docs/re/08` §4c），
// 所以用號碼命名，不用「步兵」「騎兵」這種會誤導的名字。
const (
	Branch1 uint8 = 1
	Branch4 uint8 = 4
	Branch5 uint8 = 5
	Branch6 uint8 = 6
)

// branchIdx 把兵種代號壓成 0..3，讓表可以用陣列寫。
func branchIdx(b uint8) int {
	switch b {
	case Branch1:
		return 0
	case Branch4:
		return 1
	case Branch5:
		return 2
	case Branch6:
		return 3
	}
	return -1
}

// 表的一列：[兵種1, 兵種4, 兵種5, 兵種6]。
type branchRow [4]int8

// noValue 標記原版沒有設定的格子——`sub_503BB`／`sub_501A3` 在那些
// 分支上直接跳到 return，`var_1` 是未初始化的 stack 值。
//
// 只有高山（地形 6）的兵種 4/5/6 會落到這裡，而**高山的移動成本是 255
// 等於不可通行**，那三個兵種根本站不上去。原版的「漏寫」與地形規則自洽。
const noValue int8 = -1

// defenceTable 是 `sub_503BB` 的內容：防守方的地形係數。
//
// 值域 0..5。愈大愈耐打——`sub_50655` 把它換成 `1 + 0.25×(v−1)`
// 當作分母，所以 5 → 2.0（傷害減半）、1 → 1.0（無修正）、0 → 0.75。
var defenceTable = map[assets.TileKind]branchRow{
	assets.TilePlain:    {2, 2, 5, 4},                   // 1 平原
	assets.TileHill:     {4, 2, 3, 3},                   // 2 丘陵
	assets.TileWater:    {1, 1, 1, 1},                   // 3 河海
	assets.TileForest:   {4, 1, 1, 2},                   // 4 森林
	assets.TileCity:     {5, 2, 2, 3},                   // 5 城市
	assets.TileMountain: {4, noValue, noValue, noValue}, // 6 高山，只有兵種 1 站得上去
	assets.TileDesert:   {1, 2, 5, 4},                   // 7 沙漠
	assets.TileBridgeA:  {1, 2, 2, 2},                   // 8 橋樑（縱）
	assets.TileBridgeB:  {1, 2, 2, 2},                   // 9 橋樑（橫）
	assets.TilePlateau:  {3, 2, 5, 4},                   // 10 高原
	assets.TilePass:     {5, 5, 5, 5},                   // 11 關口，不分兵種
	assets.TileKindMax:  {5, 5, 5, 5},                   // 22，與關口同組
}

// attackTable 是 `sub_501A3` 的內容：攻擊方的地形係數。
//
// 值域 0..5。`sub_50600` 把它換成 `0.5 × v` 當作乘數，
// 所以 5 → ×2.5、2 → ×1.0、0 → 打不出傷害。
var attackTable = map[assets.TileKind]branchRow{
	assets.TilePlain:    {3, 5, 5, 3},                   // 1 平原
	assets.TileHill:     {3, 5, 3, 2},                   // 2 丘陵
	assets.TileWater:    {1, 1, 3, 1},                   // 3 河海
	assets.TileForest:   {3, 1, 1, 1},                   // 4 森林
	assets.TileCity:     {3, 2, 2, 3},                   // 5 城市
	assets.TileMountain: {3, noValue, noValue, noValue}, // 6 高山
	assets.TileDesert:   {1, 5, 5, 3},                   // 7 沙漠
	assets.TileBridgeA:  {3, 3, 3, 3},                   // 8 橋樑（縱）
	assets.TileBridgeB:  {3, 3, 3, 3},                   // 9 橋樑（橫）
	assets.TilePlateau:  {3, 5, 5, 3},                   // 10 高原，與平原同組
	assets.TilePass:     {3, 5, 5, 3},                   // 11 關口，同上
	assets.TileKindMax:  {3, 5, 5, 3},                   // 22，同上
}

// 長城段（12..21）在兩張表裡各自是一列，不分段。
var (
	greatWallDefence = branchRow{5, 0, 0, 0}
	greatWallAttack  = branchRow{3, 0, 0, 0}
)

// 鐵路覆蓋地形——兩支函式都是**第一件事**就問 `sub_4FEF0`。
var (
	railDefence = branchRow{2, 2, 2, 2} // 不分兵種
	railAttack  = branchRow{3, 2, 3, 3} // 兵種 4 比較差
)

// DefenceFactor 回傳防守方站在這一格的地形係數（`sub_503BB` 的值）。
//
// 第二個回傳值為 false 表示原版在那個組合上沒有設定值
// （只發生在高山 × 兵種 4/5/6，而那三個兵種上不了高山）。
func DefenceFactor(t assets.Tile, branch uint8) (int, bool) {
	return lookupTerrain(defenceTable, greatWallDefence, railDefence, t, branch)
}

// AttackFactor 回傳攻擊方站在這一格的地形係數（`sub_501A3` 的值）。
func AttackFactor(t assets.Tile, branch uint8) (int, bool) {
	return lookupTerrain(attackTable, greatWallAttack, railAttack, t, branch)
}

func lookupTerrain(tbl map[assets.TileKind]branchRow, wall, rail branchRow,
	t assets.Tile, branch uint8) (int, bool) {
	i := branchIdx(branch)
	if i < 0 {
		return 0, false
	}
	var row branchRow
	switch {
	case t.HasRail(): // 鐵路先問，問到就不看地形
		row = rail
	case t.Kind.Blocks(): // 長城 12..21
		row = wall
	default:
		r, ok := tbl[t.Kind]
		if !ok {
			return 0, false
		}
		row = r
	}
	if row[i] == noValue {
		return 0, false
	}
	return int(row[i]), true
}

// DefenceCoefficient 是 `sub_50655`：`1.0 + 0.25 × (防禦表值 − 1)`。
// 它當分母，所以愈大愈耐打。
func DefenceCoefficient(t assets.Tile, branch uint8) float64 {
	v, ok := DefenceFactor(t, branch)
	if !ok {
		return 1
	}
	return 1 + 0.25*float64(v-1)
}

// AttackCoefficient 是 `sub_50600`：`0.5 × 攻擊表值`。
func AttackCoefficient(t assets.Tile, branch uint8) float64 {
	v, ok := AttackFactor(t, branch)
	if !ok {
		return 0
	}
	return 0.5 * float64(v)
}

// AttackValue 是 `sub_51D68` 算出的「攻擊方施加於防守方的值」：
//
//	強度(攻) × 攻擊係數(攻方站的格) ÷ 防禦係數(守方站的格)
//
// 攻守雙方各看**自己**站的地形，這是這條公式最容易搞錯的地方。
func AttackValue(attackerStrength int, attackerTile assets.Tile, attackerBranch uint8,
	defenderTile assets.Tile, defenderBranch uint8) int {
	den := DefenceCoefficient(defenderTile, defenderBranch)
	if den == 0 {
		return 0
	}
	return round(float64(attackerStrength) *
		AttackCoefficient(attackerTile, attackerBranch) / den)
}
