package game

import (
	"fmt"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

// PlaceNameSlotWidth 是 TN15 地名表的槽寬：每個地名佔 5 個字模格。
const PlaceNameSlotWidth = 5

// TestPlaceIndexMatchesNameTable 是「TERNAME 是地名索引，不是地形編號」的證據。
//
// 每省戰場用到的最大索引，必須**恰好等於**該省 TN15 地名表的詞條數。
// 這個對應要是隨機的，39 個省全中的機率可以忽略不計。
//
// 舊說「格值 = 地形編號 0..22」解釋不了為什麼各省的上限都不一樣，
// 也解釋不了為什麼上限剛好落在該省的地名數上。
func TestPlaceIndexMatchesNameTable(t *testing.T) {
	m, err := LoadMap(readGame(t, "WARPOS.DAT"), readGame(t, "TERNAME.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	for id := ProvinceID(1); id <= ProvinceCount; id++ {
		gf, err := assets.ParseGlyphFile(readGame(t, fmt.Sprintf("TN15.%d", id)))
		if err != nil {
			t.Fatalf("省 %d 的地名表: %v", id, err)
		}
		if len(gf.Glyphs)%PlaceNameSlotWidth != 0 {
			t.Errorf("省 %d 的地名表有 %d 個字模，不是槽寬 %d 的整數倍",
				id, len(gf.Glyphs), PlaceNameSlotWidth)
			continue
		}
		names := len(gf.Glyphs) / PlaceNameSlotWidth

		bf, err := m.Battlefield(id)
		if err != nil {
			t.Fatal(err)
		}
		if got := int(bf.MaxPlace()); got != names {
			t.Errorf("省 %d：地名表有 %d 條，但戰場用到的最大索引是 %d",
				id, names, got)
		}
	}
}

// TestPlaceDistribution 檢查湖北的地名佔格數符合地理常識。
//
// 河流與山脈綿延多格、城市各佔一格——這是「地名索引」說的旁證：
// 如果格值是地形類型，就不會出現「某個值全省只有一格」這種分佈。
func TestPlaceDistribution(t *testing.T) {
	m, err := LoadMap(readGame(t, "WARPOS.DAT"), readGame(t, "TERNAME.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	bf, err := m.Battlefield(26) // 湖北
	if err != nil {
		t.Fatal(err)
	}
	count := map[PlaceIndex]int{}
	for y := range bf.Place {
		for _, v := range bf.Place[y] {
			count[v]++
		}
	}
	// TN15.26 的詞條：1 漢口、2 武昌、3 漢陽、4 襄陽（城市，各一格）；
	// 9 長江（橫貫全省，格數最多）；6 大別山（山脈，多格）。
	for _, city := range []PlaceIndex{1, 2, 3, 4} {
		if count[city] != 1 {
			t.Errorf("湖北的城市 %d 應佔 1 格，實得 %d 格", city, count[city])
		}
	}
	if count[9] <= count[6] {
		t.Errorf("湖北的長江（9，%d 格）應比大別山（6，%d 格）佔更多格",
			count[9], count[6])
	}
	// 沒有地名的腹地應該是最多的。
	for k, n := range count {
		if k != 0 && n >= count[0] {
			t.Errorf("地名 %d 佔 %d 格，不該多於無地名的腹地（%d 格）", k, n, count[0])
		}
	}
}
