package layout_test

import (
	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
	"math"
	"os"
	"reflect"
	"sort"
	"testing"
)

func TestEarthHexUsesExistingGameNeighborsAndSharedRailEndpoints(t *testing.T) {
	view, err := l.NewEarthBattleLayout(l.ModernDesignSurface())
	if err != nil {
		t.Fatal(err)
	}
	opposite := map[int]int{1: 6, 2: 5, 3: 4, 4: 3, 5: 2, 6: 1}
	for cell := game.CellIndex(0); cell < 196; cell++ {
		col, row := cell.ColRow()
		x, y := view.CellCenter(col, row)
		if !view.CellContains(col, row, int(math.Round(x)), int(math.Round(y))) {
			t.Fatal("center not clickable", cell)
		}
		rect, _ := view.MapCellRect(col, row)
		if view.CellContains(col, row, rect.X, rect.Y) {
			t.Fatal("bounding corner steals hex click", cell)
		}
		for dir := game.DirLowerLeft; dir <= game.DirUpperRight; dir++ {
			n, ok := cell.Neighbour(dir)
			if !ok {
				continue
			}
			nc, nr := n.ColRow()
			nx, ny := view.CellCenter(nc, nr)
			dx, dy := l.BattleRailOffset(int(dir))
			ex, ey := l.BattleRailOffset(opposite[int(dir)])
			if gap := math.Hypot(x+dx-nx-ex, y+dy-ny-ey); gap > 1e-7 {
				t.Fatal("rail seam", cell, n, gap)
			}
		}
	}
}

func TestEarthRailPortsMatchOriginalDecodedPixels(t *testing.T) {
	raw, err := os.ReadFile("/src/workplace/orig/game/RAIL.TPC")
	if os.IsNotExist(err) {
		t.Skip("optional local original artwork reference")
	}
	if err != nil {
		t.Fatal(err)
	}
	tiles, err := assets.ParseRailTiles(raw)
	if err != nil {
		t.Fatal(err)
	}
	for index, pix := range tiles {
		ports := []int{}
		for edge := 0; edge < 4; edge++ {
			sum, count := 0, 0
			for n := 0; n < 32; n++ {
				x, y := n, 0
				switch edge {
				case 1:
					y = 23
				case 2:
					x, y = 0, n
				case 3:
					x, y = 31, n
				}
				if y >= 24 {
					continue
				}
				if pix[y*32+x] != 0 {
					sum += n
					count++
				}
			}
			if count == 0 {
				continue
			}
			dir := 5
			switch edge {
			case 1:
				dir = 2
			case 2:
				dir = 4
				if sum/count >= 12 {
					dir = 1
				}
			case 3:
				dir = 6
				if sum/count >= 12 {
					dir = 3
				}
			}
			ports = append(ports, dir)
		}
		sort.Ints(ports)
		if !reflect.DeepEqual(ports, l.BattleRailPorts(index)) {
			t.Fatal("original rail shape differs", index, ports, l.BattleRailPorts(index))
		}
	}
}
