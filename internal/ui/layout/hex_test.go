package layout

import (
	"testing"

	"github.com/wicanr2/great-era-remake/internal/game"
)

func TestBattlefieldCellAt(t *testing.T) {
	tests := []struct {
		name string
		x, y int
		want game.CellIndex
		ok   bool
	}{
		{"偶數欄左上", 190, 0, 0, true},
		{"偶數欄右下仍在同格", 221, 23, 0, true},
		{"奇數欄上方凹口不命中", 222, 11, game.NoCell, false},
		{"奇數欄第一格", 222, 12, 1, true},
		{"奇數欄下一列邊界", 222, 36, 15, true},
		{"最後一格", 637, 347, 195, true},
		{"戰場左外", 189, 0, game.NoCell, false},
		{"戰場右外", 638, 0, game.NoCell, false},
		{"偶數欄下外", 190, 336, game.NoCell, false},
		{"奇數欄下外", 222, 348, game.NoCell, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, ok := BattlefieldCellAt(tt.x, tt.y, 190, 0)
			if got != tt.want || ok != tt.ok {
				t.Fatalf("BattlefieldCellAt(%d,%d) = (%d,%v)，要 (%d,%v)", tt.x, tt.y, got, ok, tt.want, tt.ok)
			}
		})
	}
}
