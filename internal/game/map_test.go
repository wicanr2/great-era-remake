package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

const gameDir = "../../workplace/orig/game"

func loadTestMap(t *testing.T) *Map {
	t.Helper()
	warpos, err1 := os.ReadFile(filepath.Join(gameDir, "WARPOS.DAT"))
	tername, err2 := os.ReadFile(filepath.Join(gameDir, "TERNAME.DAT"))
	if err1 != nil || err2 != nil {
		t.Skip("沒有原版素材，跳過")
	}
	m, err := LoadMap(warpos, tername)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// TestLoadMapRejectsBadSize 確認尺寸不對時會報錯而不是默默吃掉。
func TestLoadMapRejectsBadSize(t *testing.T) {
	if _, err := LoadMap(make([]byte, 100), make([]byte, 100)); err == nil {
		t.Fatal("尺寸不對應該報錯")
	}
}

// TestAdjacencySymmetry 規則層也要守住對稱性——LoadMap 會拒收不對稱的資料。
func TestAdjacencySymmetry(t *testing.T) {
	m := loadTestMap(t)
	for a := ProvinceID(1); a <= ProvinceCount; a++ {
		ns, err := m.Neighbours(a)
		if err != nil {
			t.Fatal(err)
		}
		for _, b := range ns {
			if !m.Adjacent(b, a) {
				t.Errorf("%d → %d 但 %d ↛ %d", a, b, b, a)
			}
		}
	}
}

// TestAttackTargets 用 DOSBox 實機畫面對照，兩個獨立樣本。
//
// 存檔 SAVE(1) 裡玩家（吳佩孚）同時控制湖北(26) 與河南(19)——兩省的司令
// 都是吳佩孚。實機的攻打子選單在這兩個省分別顯示：
//
//	河南(19)：11,16,18,20,21,22      鄰接表多一個 26
//	湖北(26)：18,22,25,27,29         鄰接表多一個 19
//
// 兩次少掉的都正好是玩家控制的另一個省，所以「鄰省扣掉已控制」成立。
func TestAttackTargets(t *testing.T) {
	m := loadTestMap(t)
	owned := map[ProvinceID]bool{19: true, 26: true} // 河南、湖北
	for _, tc := range []struct {
		from ProvinceID
		want []ProvinceID
	}{
		{19, []ProvinceID{11, 16, 18, 20, 21, 22}},
		{26, []ProvinceID{18, 22, 25, 27, 29}},
	} {
		got, err := m.AttackTargets(tc.from, owned)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != len(tc.want) {
			t.Fatalf("從省 %d 的攻打候選 = %v，預期 %v", tc.from, got, tc.want)
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Fatalf("從省 %d 的攻打候選 = %v，預期 %v", tc.from, got, tc.want)
			}
		}
	}
}

// TestBattlefieldGrid 確認格子圖載進來了，且方位與地理一致。
//
// 河南(19) 的 WARPOS 圖：西北角是陜西(18)、東北角是山東(20)。
func TestBattlefieldGrid(t *testing.T) {
	m := loadTestMap(t)
	bf, err := m.Battlefield(19)
	if err != nil {
		t.Fatal(err)
	}
	if got := bf.Owner[0][0]; got != 18 {
		t.Errorf("河南戰場 (0,0) 的所屬省 = %d，預期 18（陜西，西北角）", got)
	}
	if got := bf.Owner[0][assets.GridW-1]; got != 20 {
		t.Errorf("河南戰場 (13,0) 的所屬省 = %d，預期 20（山東，東北角）", got)
	}
	// 中央應該是本省腹地
	if got := bf.Owner[5][5]; got != 0 {
		t.Errorf("河南戰場 (5,5) 的所屬省 = %d，預期 0（本省腹地）", got)
	}
	// 地形值域
	for y := 0; y < assets.GridH; y++ {
		for x := 0; x < assets.GridW; x++ {
			if bf.Terrain[y][x] > 22 {
				t.Fatalf("地形編號 %d 超出 0..22 於 (%d,%d)", bf.Terrain[y][x], x, y)
			}
		}
	}
}

// TestIslandProvinces 離島的度數應該是 1。
func TestIslandProvinces(t *testing.T) {
	m := loadTestMap(t)
	for _, tc := range []struct {
		p    ProvinceID
		want ProvinceID
		note string
	}{
		{37, 24, "臺灣省 → 福建"},
		{38, 36, "海南島 → 廣東"},
		{39, 34, "緬甸 → 雲南"},
	} {
		ns, err := m.Neighbours(tc.p)
		if err != nil {
			t.Fatal(err)
		}
		if len(ns) != 1 || ns[0] != tc.want {
			t.Errorf("%s：鄰省 = %v，預期 [%d]", tc.note, ns, tc.want)
		}
	}
}
