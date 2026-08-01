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

// TestAttackTargets 用 DOSBox 實機畫面對照。
//
// 實機：玩家控制湖北(26)，在河南(19) 的攻打選單顯示 11,16,18,20,21,22。
// 鄰接表是這六個再加 26，扣掉自己控制的 26 之後應該完全吻合。
func TestAttackTargets(t *testing.T) {
	m := loadTestMap(t)
	owned := map[ProvinceID]bool{26: true} // 湖北
	got, err := m.AttackTargets(19, owned)
	if err != nil {
		t.Fatal(err)
	}
	want := []ProvinceID{11, 16, 18, 20, 21, 22}
	if len(got) != len(want) {
		t.Fatalf("攻打候選 = %v，預期 %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("攻打候選 = %v，預期 %v", got, want)
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
