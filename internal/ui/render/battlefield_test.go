package render

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
)

func loadTestMap(t *testing.T) *game.Map {
	t.Helper()
	warpos, e1 := os.ReadFile(filepath.Join(gameDir, "WARPOS.DAT"))
	tername, e2 := os.ReadFile(filepath.Join(gameDir, "TERNAME.DAT"))
	nwmap, e3 := os.ReadFile(filepath.Join(gameDir, "NWMAP.DAT"))
	if e1 != nil || e2 != nil || e3 != nil {
		t.Skip("沒有原版素材，跳過")
	}
	m, err := game.LoadMap(warpos, tername, nwmap)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

// TestDrawBattlefield 確認 39 個省的戰場都畫得出來、不越界。
func TestDrawBattlefield(t *testing.T) {
	m := loadTestMap(t)
	w, h := assets.GridW*CellSize, assets.GridH*CellSize
	for p := game.ProvinceID(1); p <= game.ProvinceCount; p++ {
		bf, err := m.Battlefield(p)
		if err != nil {
			t.Fatal(err)
		}
		c := NewCanvas(w, h)
		c.DrawBattlefield(bf, 0, 0)
		// 每格都該被填到，畫布不該留下初始的全黑
		var black int
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				r, g, b, _ := c.Image().At(x, y).RGBA()
				if r == 0 && g == 0 && b == 0 {
					black++
				}
			}
		}
		if black == w*h {
			t.Errorf("省 %d 的戰場整張是黑的，沒畫到", p)
		}
	}
}

// TestDrawBattlefieldClipsOutOfBounds 畫到畫布外不該 panic。
func TestDrawBattlefieldClipsOutOfBounds(t *testing.T) {
	m := loadTestMap(t)
	bf, err := m.Battlefield(1)
	if err != nil {
		t.Fatal(err)
	}
	c := NewCanvas(10, 10) // 遠小於一張戰場
	c.DrawBattlefield(bf, -50, -50)
	c.DrawBattlefield(bf, 500, 500)
}
