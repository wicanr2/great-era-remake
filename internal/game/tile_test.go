package game

import (
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

// TestCityCountMatchesProvinceTable 是 NWMAP 解讀正確的決定性證據。
//
// 每省戰場上「地物 == 城市」的格數，必須等於 TOWN(1).DAT 的城市數欄位。
// 39/39 零誤差。這一條同時驗了三件事：NWMAP 的記錄切法（每省 392 B）、
// 地物編號 5 是城市、SPEC-03 的 +16 是城市數。
func TestCityCountMatchesProvinceTable(t *testing.T) {
	m := loadTestMap(t)
	tbl, err := ParseTownFile(readGame(t, "TOWN(1).DAT"))
	if err != nil {
		t.Fatal(err)
	}
	for id := ProvinceID(1); id <= ProvinceCount; id++ {
		bf, err := m.Battlefield(id)
		if err != nil {
			t.Fatal(err)
		}
		p, err := tbl.At(id)
		if err != nil {
			t.Fatal(err)
		}
		if got := bf.Tiles.CountKind(assets.TileCity); got != int(p.Cities) {
			t.Errorf("省 %d：戰場上有 %d 個城市格，省份表記 %d 個",
				id, got, p.Cities)
		}
	}
}

// TestWaterTilesAreRivers 驗證湖北的長江與漢水兩個地名，
// 每一格的地物都是水。
//
// 地名（TERNAME）與地物（NWMAP）是兩份獨立的檔案，它們在同一格上一致，
// 是兩邊都解對了才會有的結果。
func TestWaterTilesAreRivers(t *testing.T) {
	m := loadTestMap(t)
	bf, err := m.Battlefield(26) // 湖北
	if err != nil {
		t.Fatal(err)
	}
	// TN15.26 的 9 = 長江、10 = 漢水
	rivers := map[PlaceIndex]string{9: "長江", 10: "漢水"}
	seen := map[PlaceIndex]int{}
	for y := 0; y < assets.GridH; y++ {
		for x := 0; x < assets.GridW; x++ {
			name, ok := rivers[bf.Place[y][x]]
			if !ok {
				continue
			}
			seen[bf.Place[y][x]]++
			if k := bf.Tiles[y][x].Kind; k != assets.TileWater {
				t.Errorf("湖北 (%d,%d) 是%s，地物應為水，實得 %d", x, y, name, k)
			}
		}
	}
	if seen[9] != 33 || seen[10] != 10 {
		t.Errorf("湖北的長江應 33 格、漢水 10 格，實得 %d / %d", seen[9], seen[10])
	}
}

// TestTileKindRanges 檢查每一格都拆得出合法的地形，鐵路格拆得出合法的圖塊。
//
// 鐵路的編碼是 `v = 30 + 25×圖塊 + 地形`（docs/spec/04 §2）。
// 全 39 省 7,644 格用這個公式拆，零例外——公式錯的話會立刻拆出越界值。
func TestTileKindRanges(t *testing.T) {
	m := loadTestMap(t)
	rail := 0
	under := map[assets.TileKind]int{}
	for id := ProvinceID(1); id <= ProvinceCount; id++ {
		bf, _ := m.Battlefield(id)
		for y := 0; y < assets.GridH; y++ {
			for x := 0; x < assets.GridW; x++ {
				tile := bf.Tiles[y][x]
				if tile.Kind == 0 || tile.Kind > assets.TileKindMax {
					t.Fatalf("省 %d (%d,%d) 的地形 %d 超出 1..%d",
						id, x, y, tile.Kind, assets.TileKindMax)
				}
				if idx := tile.Kind.TileIndex(); idx < 0 || idx >= 22 {
					t.Fatalf("地形 %d 的圖塊索引 %d 超出 NEWTERR 的 22 張", tile.Kind, idx)
				}
				if tile.HasRail() {
					if tile.Rail < 0 || tile.Rail >= assets.RailTileCount {
						t.Fatalf("省 %d (%d,%d) 的鐵路圖塊 %d 超出 RAIL 的 %d 張",
							id, x, y, tile.Rail, assets.RailTileCount)
					}
					rail++
					under[tile.Kind]++
				}
			}
		}
	}
	// 鐵路格數是實測值，變了就代表解讀改了。
	if rail != 885 {
		t.Errorf("鐵路格數應為 885，實得 %d", rail)
	}
	// 鐵路只鋪在這五種地形上：平原、丘陵、河海（鐵橋）、沙漠、高原。
	// 地形名稱出自原版詞表 2.15 的 195–205（docs/spec/04 §2）。
	// 這是拆出來的結果，不是預先設定的過濾條件——公式錯就會冒出別的地形。
	for k := range under {
		switch k {
		case assets.TilePlain, assets.TileHill, assets.TileWater,
			assets.TileDesert, assets.TilePlateau:
		default:
			t.Errorf("鐵路底下出現非預期的地形 %d（%d 格）", k, under[k])
		}
	}
	if under[assets.TileWater] != 77 {
		t.Errorf("跨河海的鐵橋應有 77 格，實得 %d", under[assets.TileWater])
	}
}

// TestCitiesAlwaysFlagged 城市格一定帶旗標，高山與沙漠一定不帶。
//
// 旗標的語意未解（docs/spec/04 §4），這裡只把觀察到的規律釘住，
// 日後解出來時可以確認沒有理解偏移。
func TestCitiesAlwaysFlagged(t *testing.T) {
	m := loadTestMap(t)
	for id := ProvinceID(1); id <= ProvinceCount; id++ {
		bf, _ := m.Battlefield(id)
		for y := 0; y < assets.GridH; y++ {
			for x := 0; x < assets.GridW; x++ {
				tile := bf.Tiles[y][x]
				switch tile.Kind {
				case assets.TileCity:
					if tile.Flags == 0 {
						t.Errorf("省 %d (%d,%d) 是城市卻沒有旗標", id, x, y)
					}
				case assets.TileMountain, assets.TileDesert:
					if tile.Flags != 0 {
						t.Errorf("省 %d (%d,%d) 是高山／沙漠卻帶旗標 %#x",
							id, x, y, tile.Flags)
					}
				}
			}
		}
	}
}
