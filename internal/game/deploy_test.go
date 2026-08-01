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
	got, ok := occ.Deploy(zone, 58) // 吳佩孚
	if !ok || got != 186 {
		t.Errorf("Deploy 落點 = %d (%v)，應為 186", got, ok)
	}
	if occ[186] != 58 {
		t.Errorf("佔用表沒指回單位：occ[186] = %d", occ[186])
	}
	// 佔滿之後要回報放不下，不是靜靜覆蓋別人。
	for i := 1; i < len(zone); i++ {
		occ.Deploy(zone, GeneralID(100+i))
	}
	if _, ok := occ.Deploy(zone, 999); ok {
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
