package game

import (
	"testing"
)

// 驗收基準是 DOSBox 實機截圖（docs/playtest/02），不是內部一致性。
// 湖北省是唯一有完整畫面數值的省，所以拿它當金標準。
var hubeiJuly = struct {
	gold, food, ammo, coal, fuel, iron uint16
	popWan                             uint32
	cities, landValue, arsenals, loyal uint8
}{4200, 18050, 8787, 13000, 12048, 14031, 1825, 5, 22, 3, 79}

// TestSaveProvincesAgainstScreen 拿實機畫面的 13 個欄位對存檔。
func TestSaveProvincesAgainstScreen(t *testing.T) {
	tbl, err := ParseSaveProvinces(readGame(t, "SAVE(1).DT1"))
	if err != nil {
		t.Fatal(err)
	}
	if tbl.Date.Year != 15 || tbl.Date.Month != 7 {
		t.Errorf("存檔日期應為民國 15 年 7 月，得到 %d 年 %d 月",
			tbl.Date.Year, tbl.Date.Month)
	}

	p, err := tbl.At(26) // 湖北
	if err != nil {
		t.Fatal(err)
	}
	w := hubeiJuly
	for _, c := range []struct {
		name     string
		got, exp uint32
	}{
		{"黃金", uint32(p.Gold), uint32(w.gold)},
		{"糧食", uint32(p.Food), uint32(w.food)},
		{"彈藥", uint32(p.Ammo), uint32(w.ammo)},
		{"煤礦", uint32(p.Coal), uint32(w.coal)},
		{"燃料", uint32(p.Fuel), uint32(w.fuel)},
		{"鐵礦", uint32(p.Iron), uint32(w.iron)},
		{"人口（萬）", p.PopulationWan(), w.popWan},
		{"城市數", uint32(p.Cities), uint32(w.cities)},
		{"地價", uint32(p.LandValue), uint32(w.landValue)},
		{"兵工廠數", uint32(p.Arsenals), uint32(w.arsenals)},
		{"人民忠誠度", uint32(p.Loyalty), uint32(w.loyal)},
	} {
		if c.got != c.exp {
			t.Errorf("湖北 %s：實機 %d，解出 %d", c.name, c.exp, c.got)
		}
	}
	// 實機頂端顯示「司令 吳佩孚、省長 吳佩孚」，吳佩孚在 MAN115 是第 58 位。
	if p.Commander != 58 || p.Governor != 58 {
		t.Errorf("湖北司令/省長應同為 58（吳佩孚），得到 %d/%d",
			p.Commander, p.Governor)
	}
}

// TestTownFactions 驗證 TOWN(1).DAT 的司令欄位排得出 1926 年的割據圖。
//
// 這比任何內部一致性檢查都硬——欄位解錯不可能剛好排出一張對的歷史地圖。
func TestTownFactions(t *testing.T) {
	tbl, err := ParseTownFile(readGame(t, "TOWN(1).DAT"))
	if err != nil {
		t.Fatal(err)
	}
	if tbl.Date != nil {
		t.Error("TOWN 檔沒有日期檔頭，Date 應為 nil")
	}
	f := tbl.Factions()

	// 孫傳芳的五省聯軍：蘇、皖、浙、閩、贛，一個不多一個不少。
	want := map[GeneralID][]ProvinceID{
		98:  {21, 22, 23, 24, 25},                     // 孫傳芳
		58:  {19, 26},                                 // 吳佩孚
		1:   {35, 36},                                 // 蔣中正（兩廣）
		166: {1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 20},      // 張作霖
		4:   {16},                                     // 閰錫山（山西）
		3:   {18},                                     // 馮玉祥（陝西）
		141: {27}, 146: {28}, 157: {29}, 156: {34},    // 趙恆惕/袁祖銘/劉湘/唐繼堯
	}
	for id, exp := range want {
		got := f[id]
		if len(got) != len(exp) {
			t.Errorf("將領 %d 應轄 %v，解出 %v", id, exp, got)
			continue
		}
		for i := range exp {
			if got[i] != exp[i] {
				t.Errorf("將領 %d 應轄 %v，解出 %v", id, exp, got)
				break
			}
		}
	}
	if len(f) != len(want) {
		t.Errorf("第一期應有 %d 個勢力，解出 %d 個", len(want), len(f))
	}

	// 初始狀態下司令與省長相同；存檔裡才會分化。
	for i := range tbl.Province {
		p := &tbl.Province[i]
		if p.Commander != p.Governor {
			t.Errorf("TOWN(1) 第 %d 省的司令 %d 與省長 %d 不同",
				i+1, p.Commander, p.Governor)
		}
	}
}

// TestProvinceNeighboursVsMap 比對記錄裡的鄰省表與 WARPOS 的幾何鄰接。
//
// 兩者不完全相同（spec §4），所以這裡驗的是「差異只落在已知的兩類」，
// 不是驗相等——把已知差異寫進測試，日後真的變了才會被抓到。
func TestProvinceNeighboursVsMap(t *testing.T) {
	tbl, err := ParseTownFile(readGame(t, "TOWN(1).DAT"))
	if err != nil {
		t.Fatal(err)
	}
	m, err := LoadMap(readGame(t, "WARPOS.DAT"), readGame(t, "TERNAME.DAT"))
	if err != nil {
		t.Fatal(err)
	}

	// 隔海／隔境的三筆：TOWN 表不列，WARPOS 的格子上有。
	acrossSea := map[ProvinceID]ProvinceID{24: 37, 36: 38, 34: 39} // 閩→臺、粵→瓊、滇→緬
	exact := 0
	for id := ProvinceID(1); id <= ProvinceCount; id++ {
		p, err := tbl.At(id)
		if err != nil {
			t.Fatal(err)
		}
		listed := map[ProvinceID]bool{}
		for _, n := range p.Neighbours {
			listed[n] = true
		}
		geo, err := m.Neighbours(id)
		if err != nil {
			t.Fatal(err)
		}
		missing := 0
		for _, n := range geo {
			if n > ProvinceCount { // WARPOS 自己的哨兵值（40）
				continue
			}
			if !listed[n] {
				if acrossSea[id] == n {
					missing++
					continue
				}
				t.Errorf("省 %d 的鄰省表少了 %d，且不在已知的隔海清單裡", id, n)
			}
		}
		for _, n := range p.Neighbours {
			found := false
			for _, g := range geo {
				if g == n {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("省 %d 的鄰省表有 %d，但 WARPOS 的格子上沒有", id, n)
			}
		}
		if missing == 0 && len(p.Neighbours) == countReal(geo) {
			exact++
		}
	}
	if exact < 28 {
		t.Errorf("與 WARPOS 完全一致的省應至少 28 個，只有 %d 個", exact)
	}
	t.Logf("鄰省表與 WARPOS 完全一致 %d/%d 省", exact, ProvinceCount)
}

func countReal(ids []ProvinceID) int {
	n := 0
	for _, id := range ids {
		if id <= ProvinceCount {
			n++
		}
	}
	return n
}

// TestProvinceCoastal 驗證帶 SeaBorder 的省全部沿海或臨國境。
func TestProvinceCoastal(t *testing.T) {
	tbl, err := ParseTownFile(readGame(t, "TOWN(1).DAT"))
	if err != nil {
		t.Fatal(err)
	}
	// 遼寧、安東、河北、山東、江蘇、浙江、福建、廣東、臺灣、緬甸
	want := map[ProvinceID]bool{7: true, 9: true, 11: true, 20: true, 21: true,
		23: true, 24: true, 36: true, 37: true, 39: true}
	for id := ProvinceID(1); id <= ProvinceCount; id++ {
		p, _ := tbl.At(id)
		if p.Coastal != want[id] {
			t.Errorf("省 %d 的 Coastal 應為 %v，解出 %v", id, want[id], p.Coastal)
		}
	}
}

// TestProvinceRoundTrip 驗證「改寫而非重建」：不改欄位時 Bytes() 必須
// byte-for-byte 相同（CLAUDE.md §9）。TOWN 與 .DT1 兩種相位都要過。
func TestProvinceRoundTrip(t *testing.T) {
	for _, c := range []struct {
		file string
		base int
	}{
		{"TOWN(1).DAT", TownArrayOffset},
		{"TOWN(2).DAT", TownArrayOffset},
		{"TOWN(3).DAT", TownArrayOffset},
		{"SAVE(1).DT1", SaveArrayOffset},
		{"SAVE(2).DT1", SaveArrayOffset},
	} {
		data := readGame(t, c.file)
		var tbl *ProvinceTable
		var err error
		if c.base == TownArrayOffset {
			tbl, err = ParseTownFile(data)
		} else {
			tbl, err = ParseSaveProvinces(data)
		}
		if err != nil {
			t.Fatalf("%s: %v", c.file, err)
		}
		for i := range tbl.Province {
			out := tbl.Province[i].Bytes()
			orig := data[c.base+i*ProvinceRecordSize : c.base+(i+1)*ProvinceRecordSize]
			for k := 0; k < ProvinceRecordSize; k++ {
				if out[k] != orig[k] {
					t.Fatalf("%s 第 %d 省 round-trip 在 offset %d 不同：%#x vs %#x",
						c.file, i+1, k, out[k], orig[k])
				}
			}
		}
	}
}
