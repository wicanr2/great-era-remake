package game

import (
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

// 六角格幾何：值全部出自 sub_50FF5，這裡驗算式，不驗「看起來對」。
func TestHexGeometry(t *testing.T) {
	cases := []struct {
		cell     CellIndex
		col, row int
		x, y     int
	}{
		{0, 0, 0, 0, 0},        // 左上，偶數欄不下移
		{1, 1, 0, 32, 12},      // 奇數欄下移半格
		{13, 13, 0, 416, 12},   // 最右欄（13 是奇數）
		{14, 0, 1, 0, 24},      // 換列
		{186, 4, 13, 128, 312}, // 湖北打河南時的第一個部署格
		{195, 13, 13, 416, 324},
	}
	for _, c := range cases {
		col, row := c.cell.ColRow()
		if col != c.col || row != c.row {
			t.Errorf("cell %d 的欄列 = (%d,%d)，應為 (%d,%d)", c.cell, col, row, c.col, c.row)
		}
		x, y := c.cell.ScreenXY()
		if x != c.x || y != c.y {
			t.Errorf("cell %d 的座標 = (%d,%d)，應為 (%d,%d)", c.cell, x, y, c.x, c.y)
		}
		back, err := CellAt(c.col, c.row)
		if err != nil || back != c.cell {
			t.Errorf("CellAt(%d,%d) = %d, %v，應回 %d", c.col, c.row, back, err, c.cell)
		}
	}
}

// 戰場塞得進 BGI 640×350。這不是巧合，是驗證幾何解讀正確的旁證：
// 若欄寬或列高解錯，整張圖會超出畫面或空一大塊。
func TestBattlefieldFitsBGIScreen(t *testing.T) {
	if BattlefieldW != 448 || BattlefieldH != 348 {
		t.Fatalf("戰場範圍 = %d×%d，應為 448×348", BattlefieldW, BattlefieldH)
	}
	if BattlefieldH > 350 {
		t.Errorf("戰場高 %d 超出 BGI 640×350", BattlefieldH)
	}
}

// 每個鄰省的進場區恰好 10 格，正好對上戰鬥記錄的 10 個部隊槽。
// 這條在全 39 省成立，零例外——是「WARPOS 是戰場分區圖」最硬的資料證據。
func TestDeployZoneIsAlwaysTenCells(t *testing.T) {
	m := loadTestMap(t)
	checked := 0
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		ns, err := m.Neighbours(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, n := range ns {
			zone, err := m.DeployZone(p, n)
			if err != nil {
				t.Fatalf("省 %d 對 %d: %v", p, n, err)
			}
			if len(zone) != DeployZoneSize {
				t.Errorf("省 %d 上屬於省 %d 的格數 = %d，應為 %d",
					p, n, len(zone), DeployZoneSize)
			}
			checked++
		}
	}
	if checked < 100 {
		t.Fatalf("只檢查了 %d 組鄰接，樣本太少", checked)
	}
	t.Logf("檢查 %d 組（省, 鄰省）進場區", checked)
}

// 掃描方向是 195 → 0。方向錯了部署落點就錯，所以要驗第一格。
func TestDeployScanOrderMatchesOriginal(t *testing.T) {
	m := loadTestMap(t)
	// 湖北(26) 打河南(19)：原版從 cell 195 往下掃，第一個 26 是 cell 186。
	zone, err := m.DeployZone(19, 26)
	if err != nil {
		t.Fatal(err)
	}
	if zone[0] != 186 {
		t.Errorf("湖北打河南的第一個部署格 = %d，應為 186", zone[0])
	}
	for i := 1; i < len(zone); i++ {
		if zone[i] >= zone[i-1] {
			t.Fatalf("掃描順序不是遞減：zone[%d]=%d >= zone[%d]=%d",
				i, zone[i], i-1, zone[i-1])
		}
	}

	var occ Occupancy
	got, ok := occ.Deploy(zone, 58, nil) // 吳佩孚
	if !ok || got != 186 {
		t.Errorf("Deploy 落點 = %d (%v)，應為 186", got, ok)
	}
	if occ[186] != 58 {
		t.Errorf("佔用表沒指回單位：occ[186] = %d", occ[186])
	}
	// 佔滿之後要回報放不下，不是靜靜覆蓋別人。
	for i := 1; i < len(zone); i++ {
		occ.Deploy(zone, GeneralID(100+i), nil)
	}
	if _, ok := occ.Deploy(zone, 999, nil); ok {
		t.Error("進場區已滿卻還放得進去")
	}
}

// 進場區都在戰場邊緣：14×14 的中央腹地不該有任何鄰省的格。
// 這是「攻方從自己那一側進場」的空間性質。
func TestDeployZonesHugTheEdges(t *testing.T) {
	m := loadTestMap(t)
	bf, err := m.Battlefield(19)
	if err != nil {
		t.Fatal(err)
	}
	const margin = 4 // 距邊界 4 格以內才算邊緣
	for row := margin; row < assets.GridH-margin; row++ {
		for col := margin; col < assets.GridW-margin; col++ {
			if bf.Owner[row][col] != 0 {
				t.Errorf("河南戰場中央 (%d,%d) 竟屬於省 %d", col, row, bf.Owner[row][col])
			}
		}
	}
}

func TestCellIndexBounds(t *testing.T) {
	if NoCell.Valid() {
		t.Error("0xFF 不該是合法格")
	}
	if !CellIndex(195).Valid() || CellIndex(196).Valid() {
		t.Error("合法範圍應為 0..195")
	}
	if _, err := CellAt(14, 0); err == nil {
		t.Error("欄 14 應該越界")
	}
}

// 六角鄰接：位移表出自 sub_4E65C，邊界規則出自 sub_510E0。
func TestHexNeighbours(t *testing.T) {
	// 正中央的格子：偶欄與奇欄各驗一個，都該有 6 個鄰居。
	center, _ := CellAt(6, 6) // 偶數欄
	if got := center.Neighbours(); len(got) != 6 {
		t.Errorf("偶數欄中央格的鄰居數 = %d，應為 6：%v", len(got), got)
	}
	oddCenter, _ := CellAt(7, 6) // 奇數欄
	if got := oddCenter.Neighbours(); len(got) != 6 {
		t.Errorf("奇數欄中央格的鄰居數 = %d，應為 6：%v", len(got), got)
	}

	// 位移對不對，逐個方向驗。cell 90 = col 6, row 6。
	want := map[HexDir]CellIndex{
		DirLowerLeft: 89, DirDown: 104, DirLowerRight: 91,
		DirUpperLeft: 75, DirUp: 76, DirUpperRight: 77,
	}
	for d, w := range want {
		got, ok := center.Neighbour(d)
		if !ok || got != w {
			t.Errorf("cell 90 方向 %d 的鄰格 = %d (%v)，應為 %d", d, got, ok, w)
		}
	}
}

// 鄰接關係必須對稱。不對稱代表位移表或邊界規則抄錯了——
// 這是最便宜也最狠的檢查，全 196 格跑一遍。
func TestHexAdjacencyIsSymmetric(t *testing.T) {
	for i := 0; i < CellCount; i++ {
		c := CellIndex(i)
		for _, n := range c.Neighbours() {
			if !Adjacent(c, n) {
				t.Fatalf("cell %d 說 %d 是鄰居，反過來卻不成立", c, n)
			}
		}
	}
}

// 幾何自洽：鄰格的螢幕座標必須真的貼著本格。
// 六角格的鄰居距離不是「同一列同一欄 ±1」，所以這條抓得到位移表寫反。
func TestHexNeighboursAreGeometricallyClose(t *testing.T) {
	for i := 0; i < CellCount; i++ {
		c := CellIndex(i)
		x, y := c.ScreenXY()
		for _, n := range c.Neighbours() {
			nx, ny := n.ScreenXY()
			dx, dy := abs(nx-x), abs(ny-y)
			if dx > HexCellW || dy > HexCellH {
				t.Errorf("cell %d 的鄰格 %d 距離 (%d,%d) 太遠", c, n, dx, dy)
			}
		}
	}
}

// 邊緣的鄰居數：原版只裁左右，上下靠 0..195 的範圍檢查擋掉。
func TestHexEdgeNeighbourCounts(t *testing.T) {
	cases := []struct {
		col, row int
		want     int
		why      string
	}{
		{0, 0, 2, "左上角：偶欄左緣裁兩個，再被上邊界擋掉一個"},
		{0, 6, 4, "左緣偶欄：左下、左上被裁，剩下、右下、上、右上"},
		{13, 6, 4, "右緣奇欄：右下、右上被裁"},
		{13, 13, 2, "右下角"},
		{6, 0, 3, "上緣偶欄：左上、上、右上都出界"},
		{6, 13, 5, "下緣偶欄：只有正下方出界"},
	}
	for _, c := range cases {
		cell, err := CellAt(c.col, c.row)
		if err != nil {
			t.Fatal(err)
		}
		if got := len(cell.Neighbours()); got != c.want {
			t.Errorf("(%d,%d) 的鄰居數 = %d，應為 %d（%s）：%v",
				c.col, c.row, got, c.want, c.why, cell.Neighbours())
		}
	}
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// 部署的第一輪要避開「旁邊已經有敵人」的格，掃不到才放寬。
// 兩輪都實作了才不會把原版的讓步規則寫成硬性限制。
func TestDeployAvoidsCellsNextToEnemies(t *testing.T) {
	m := loadTestMap(t)
	zone, err := m.DeployZone(19, 26) // 湖北打河南，10 格
	if err != nil {
		t.Fatal(err)
	}
	isEnemy := func(g GeneralID) bool { return g >= 900 }

	// 把敵人擺在第一個部署格的鄰格上，那一格就該被跳過。
	var occ Occupancy
	nb := zone[0].Neighbours()
	if len(nb) == 0 {
		t.Fatal("第一個部署格沒有鄰格，測試前提不成立")
	}
	occ[nb[0]] = 999
	got, ok := occ.Deploy(zone, 58, isEnemy)
	if !ok {
		t.Fatal("應該還有別的格可放")
	}
	if got == zone[0] {
		t.Errorf("落點 %d 旁邊有敵人，第一輪不該選它", got)
	}

	// 每一格旁邊都有敵人時，第二輪必須硬塞——不能回報失敗。
	var packed Occupancy
	for _, c := range zone {
		for _, n := range c.Neighbours() {
			if packed[n] == 0 {
				packed[n] = 999
			}
		}
	}
	// 清掉被擺到 zone 自己身上的敵人，留下空的落點。
	for _, c := range zone {
		packed[c] = 0
	}
	if _, ok := packed.Deploy(zone, 58, isEnemy); !ok {
		t.Error("第一輪找不到安全格時，原版會硬塞，不該失敗")
	}
}

// 移動：格子轉移、佔用表兩邊都要更新、機動力照扣。
func TestMoveConsumesMovementAllowance(t *testing.T) {
	var occ Occupancy
	start, _ := CellAt(6, 6)
	u := &CombatUnit{General: 58, Cell: start, Max: 10, Current: 10}
	occ[start] = u.General

	dst, err := occ.Move(nil, u, DirDown, 3)
	if err != nil {
		t.Fatal(err)
	}
	if want, _ := CellAt(6, 7); dst != want {
		t.Errorf("落點 = %d，應為 %d", dst, want)
	}
	if u.Cell != dst {
		t.Errorf("單位的格沒更新：%d", u.Cell)
	}
	if u.Current != 7 {
		t.Errorf("剩餘機動力 = %d，應為 7", u.Current)
	}
	if occ[start] != 0 {
		t.Errorf("舊格沒清空：occ[%d] = %d", start, occ[start])
	}
	if occ[dst] != u.General {
		t.Errorf("新格沒指回單位：occ[%d] = %d", dst, occ[dst])
	}

	// 機動力不足要擋下，而且不能動到任何狀態。
	before := *u
	if _, err := occ.Move(nil, u, DirDown, 99); err == nil {
		t.Error("機動力不足卻走得動")
	}
	if *u != before {
		t.Error("移動失敗卻改了單位狀態")
	}

	// 目標格有人也要擋。
	nb, _ := u.Cell.Neighbour(DirUp)
	occ[nb] = 999
	if _, err := occ.Move(nil, u, DirUp, 1); err == nil {
		t.Error("目標格有人卻走得進去")
	}

	// 出界要擋。
	edge, _ := CellAt(0, 0)
	e := &CombatUnit{General: 1, Cell: edge, Max: 9, Current: 9}
	occ[edge] = 1
	if _, err := occ.Move(nil, e, DirUp, 1); err == nil {
		t.Error("走出上邊界卻沒被擋")
	}
}

// 長城擋路：只有能穿越的單位過得去。省 11（河北）有 8 格長城。
func TestMoveBlockedByGreatWall(t *testing.T) {
	m := loadTestMap(t)
	bf, err := m.Battlefield(11)
	if err != nil {
		t.Fatal(err)
	}
	// 找一組「相鄰、其中一格是長城、另一格不是」的格子。
	var from, wall CellIndex = NoCell, NoCell
	var dir HexDir
	for i := 0; i < CellCount && from == NoCell; i++ {
		c := CellIndex(i)
		col, row := c.ColRow()
		if bf.Tiles[row][col].Kind.Blocks() {
			continue
		}
		for d := DirLowerLeft; d <= DirUpperRight; d++ {
			n, ok := c.Neighbour(d)
			if !ok {
				continue
			}
			ncol, nrow := n.ColRow()
			if bf.Tiles[nrow][ncol].Kind.Blocks() {
				from, wall, dir = c, n, d
				break
			}
		}
	}
	if from == NoCell {
		t.Fatal("河北的戰場上找不到緊鄰長城的格，測試前提不成立")
	}

	var occ Occupancy
	u := &CombatUnit{General: 58, Cell: from, Max: 9, Current: 9}
	occ[from] = u.General
	if _, err := occ.Move(bf, u, dir, 1); err == nil {
		t.Errorf("從格 %d 走進長城格 %d 竟然通過", from, wall)
	}
	if u.Cell != from || u.Current != 9 {
		t.Error("被擋下卻改了單位狀態")
	}

	u.CanCross = true
	if _, err := occ.Move(bf, u, dir, 1); err != nil {
		t.Errorf("能穿越的單位仍被擋：%v", err)
	}
	if u.Cell != wall {
		t.Errorf("穿越後的位置 = %d，應為 %d", u.Cell, wall)
	}
}
