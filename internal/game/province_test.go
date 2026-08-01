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
	m, err := LoadMap(readGame(t, "WARPOS.DAT"), readGame(t, "TERNAME.DAT"), readGame(t, "NWMAP.DAT"))
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

// TestAttackableMatchesScreen 攻打候選要對上實機的攻打子選單。
//
// 實機在湖北(26) 時顯示 (18,22,25,27,29)，在河南(19) 時顯示
// (11,16,18,20,21,22)——都是鄰省扣掉同勢力的那一個。
func TestAttackableMatchesScreen(t *testing.T) {
	tbl, err := ParseSaveProvinces(readGame(t, "SAVE(1).DT1"))
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		from ProvinceID
		want []ProvinceID
		name string
	}{
		{26, []ProvinceID{18, 22, 25, 27, 29}, "湖北"},
		{19, []ProvinceID{11, 16, 18, 20, 21, 22}, "河南"},
	} {
		got, err := tbl.AttackableFrom(c.from)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != len(c.want) {
			t.Errorf("%s 的攻打候選：實機 %v，解出 %v", c.name, c.want, got)
			continue
		}
		for i := range c.want {
			if got[i] != c.want[i] {
				t.Errorf("%s 的攻打候選：實機 %v，解出 %v", c.name, c.want, got)
				break
			}
		}
	}
}

// TestAttackableSkipsUnowned 無主的省不能攻打。
//
// 這一條是從 sub_5B7DC 讀到的，**實機沒有驗過**——湖北與河南的鄰省
// 剛好全部有主。四川(29) 鄰接甘肅(17) 與青海(30)，兩者在第一期都無主。
func TestAttackableSkipsUnowned(t *testing.T) {
	tbl, err := ParseTownFile(readGame(t, "TOWN(1).DAT"))
	if err != nil {
		t.Fatal(err)
	}
	sichuan, err := tbl.At(29)
	if err != nil {
		t.Fatal(err)
	}
	got, err := tbl.AttackableFrom(29)
	if err != nil {
		t.Fatal(err)
	}
	inList := func(p ProvinceID) bool {
		for _, v := range got {
			if v == p {
				return true
			}
		}
		return false
	}
	// 甘肅與青海在四川的鄰省表裡，但第一期無主
	for _, p := range []ProvinceID{17, 30} {
		neighbour := false
		for _, n := range sichuan.Neighbours {
			if n == p {
				neighbour = true
			}
		}
		if !neighbour {
			t.Fatalf("省 %d 不在四川的鄰省表裡，測試前提不成立", p)
		}
		q, _ := tbl.At(p)
		if q.Commander.Valid() {
			t.Fatalf("省 %d 在第一期有主（司令 %d），測試前提不成立", p, q.Commander)
		}
		if inList(p) {
			t.Errorf("無主的省 %d 不該出現在攻打候選裡：%v", p, got)
		}
	}
	t.Logf("四川的攻打候選：%v（鄰省表 %v）", got, sichuan.Neighbours)
}

// TestProvinceFlags 記錄 +32 旗標的實測分佈。
//
// 初始檔全 0；SAVE(1) 有 25 個省設了 bit 2、SAVE(2) 只有 4 個，
// 而且**無主的省一律沒設**。這是「已徵過稅」假說的依據
// （docs/spec/03 §2）。數字是釘住現況用的回歸檢查。
func TestProvinceFlags(t *testing.T) {
	for _, name := range []string{"TOWN(1).DAT", "TOWN(2).DAT", "TOWN(3).DAT"} {
		tbl, err := ParseTownFile(readGame(t, name))
		if err != nil {
			t.Fatal(err)
		}
		for i := range tbl.Province {
			if f := tbl.Province[i].Flags; f != 0 {
				t.Errorf("%s 第 %d 省的旗標是 %#x，初始檔應全為 0", name, i+1, f)
			}
		}
	}

	for _, c := range []struct {
		file  string
		taxed int
	}{
		{"SAVE(1).DT1", 25},
		{"SAVE(2).DT1", 4},
	} {
		tbl, err := ParseSaveProvinces(readGame(t, c.file))
		if err != nil {
			t.Fatal(err)
		}
		n := 0
		for i := range tbl.Province {
			p := &tbl.Province[i]
			if !p.Taxed() {
				continue
			}
			n++
			if !p.Commander.Valid() {
				t.Errorf("%s 第 %d 省無主卻設了已徵稅旗標", c.file, i+1)
			}
		}
		if n != c.taxed {
			t.Errorf("%s 設了 bit 2 的省應為 %d 個，實得 %d", c.file, c.taxed, n)
		}
		// bit 6 在兩份存檔裡都沒出現過
		for i := range tbl.Province {
			if tbl.Province[i].Flags&ProvinceFlagExcluded != 0 {
				t.Errorf("%s 第 %d 省出現了 bit 6，先前的觀察是兩份存檔都沒有",
					c.file, i+1)
			}
		}
	}
}
