package game

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

func readGame(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(gameDir, name))
	if err != nil {
		t.Skipf("沒有原版素材 %s，跳過", name)
	}
	return b
}

// TestGeneralRoundTrip 驗證「改寫而非重建」：不改任何欄位時，
// Bytes() 必須與原始 33 bytes byte-for-byte 相同（CLAUDE.md §9）。
func TestGeneralRoundTrip(t *testing.T) {
	data := readGame(t, "MAN(1).DAT")
	gs, err := ParseGenerals(data, len(data)/GeneralRecordSize)
	if err != nil {
		t.Fatal(err)
	}
	for i := range gs {
		out := gs[i].Bytes()
		orig := data[i*GeneralRecordSize : (i+1)*GeneralRecordSize]
		for k := 0; k < GeneralRecordSize; k++ {
			if out[k] != orig[k] {
				t.Fatalf("第 %d 位將領 round-trip 在 offset %d 不同：%#x vs %#x",
					i, k, out[k], orig[k])
			}
		}
	}
	t.Logf("%d 位將領 byte-for-byte round-trip 通過", len(gs))
}

// TestManDatStructure 驗證 MAN(1).DAT 是 274 × 33，且蔣中正三項全 100。
func TestManDatStructure(t *testing.T) {
	data := readGame(t, "MAN(1).DAT")
	if len(data)%GeneralRecordSize != 0 {
		t.Fatalf("MAN(1).DAT 大小 %d 不是 %d 的倍數", len(data), GeneralRecordSize)
	}
	if n := len(data) / GeneralRecordSize; n != 274 {
		t.Errorf("將領數 = %d，預期 274（與 MAN115 的 274 個人名一致）", n)
	}
	gs, err := ParseGenerals(data, 3)
	if err != nil {
		t.Fatal(err)
	}
	// 蔣中正是第 1 筆，三項能力全滿
	if gs[0].AbilityA != 100 || gs[0].AbilityB != 100 || gs[0].AbilityC != 100 {
		t.Errorf("蔣中正能力 = %d,%d,%d，預期 100,100,100",
			gs[0].AbilityA, gs[0].AbilityB, gs[0].AbilityC)
	}
	if gs[0].Force != 20000 {
		t.Errorf("蔣中正兵力 = %d，預期 20000（步兵師滿員）", gs[0].Force)
	}
	// 能力值必須都在 0-100
	all, err := ParseGenerals(data, len(data)/GeneralRecordSize)
	if err != nil {
		t.Fatal(err)
	}
	for i, g := range all {
		if g.AbilityA > 100 || g.AbilityB > 100 || g.AbilityC > 100 {
			t.Errorf("第 %d 位將領能力值超出 0-100：%d,%d,%d",
				i, g.AbilityA, g.AbilityB, g.AbilityC)
		}
	}
}

// TestTotalForce 用實機畫面對照：湖北省 15 位將領的兵力總和 = 97500。
//
// 存檔 SAVE(1).DT1 的將領區起點 7084（實測反推），stride 33。
func TestTotalForce(t *testing.T) {
	data := readGame(t, "SAVE(1).DT1")
	const hubeiBase = 7084
	gs, err := ParseGenerals(data[hubeiBase:], 15)
	if err != nil {
		t.Fatal(err)
	}
	want := []uint16{20000, 6000, 4500, 6000, 5500, 5500, 7000, 10000,
		7000, 7000, 3500, 4500, 3500, 3500, 4000}
	for i, w := range want {
		if gs[i].Force != w {
			t.Errorf("第 %d 位將領兵力 = %d，實機顯示 %d", i+1, gs[i].Force, w)
		}
	}
	if got := TotalForce(gs); got != 97500 {
		t.Errorf("兵力總和 = %d，實機畫面顯示 97500", got)
	}
	// 吳佩孚是第 1 位，能力 90,100,94
	if gs[0].AbilityA != 90 || gs[0].AbilityB != 100 || gs[0].AbilityC != 94 {
		t.Errorf("吳佩孚能力 = %d,%d,%d，預期 90,100,94",
			gs[0].AbilityA, gs[0].AbilityB, gs[0].AbilityC)
	}
}

func TestParseGeneralsRejectsShortData(t *testing.T) {
	if _, err := ParseGenerals(make([]byte, 10), 1); err == nil {
		t.Fatal("資料不足應該報錯")
	}
}

// TestGeneralProvinceAgainstScreen 是「+4 是所屬省」的決定性驗證。
//
// 兩個省的將領數與兵力加總都要對上 DOSBox 實機截圖，零誤差。
// 這兩個數字是獨立的：將領數來自計數、兵力來自加總，欄位抓錯不可能兩個都對。
func TestGeneralProvinceAgainstScreen(t *testing.T) {
	data := readGame(t, "MAN(1).DAT")
	gs, err := ParseGenerals(data, len(data)/GeneralRecordSize)
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range []struct {
		prov  ProvinceID
		count int
		force uint32
		name  string
	}{
		{26, 15, 97500, "湖北"},
		{19, 25, 140500, "河南"},
	} {
		if got := CountOf(gs, c.prov); got != c.count {
			t.Errorf("%s 的將領數：實機 %d，解出 %d", c.name, c.count, got)
		}
		if got := ForceOf(gs, c.prov); got != c.force {
			t.Errorf("%s 的兵力：實機 %d，解出 %d", c.name, c.force, got)
		}
	}
	// 實機派將列表第一行是「1. 吳佩孚 20000」，吳佩孚是 MAN115 的第 58 位。
	hubei := GeneralsOf(gs, 26)
	if len(hubei) == 0 || hubei[0].Force != 20000 {
		t.Errorf("湖北派將列表第一位的兵力應為 20000")
	}
	if gs[57].Province != 26 {
		t.Errorf("吳佩孚（第 58 位）的所屬省應為 26（湖北），解出 %d", gs[57].Province)
	}
}

// TestGeneralProvinceRange 所屬省必須是 0 或 1..39。
//
// [雷] 三個 MAN(N).DAT 都是 9,042 B（274 筆的空間），**但實際筆數不同**：
// 第一期 274 位、第二三期各 106 位，其餘是殘料（`+4` 會讀出 180 這種越界值）。
// 筆數要以對應的 MAN{N}15 名表為準，不能用「檔案大小 ÷ 33」。
func TestGeneralProvinceRange(t *testing.T) {
	for _, c := range []struct {
		dat, names string
	}{
		{"MAN(1).DAT", "MAN115"},
		{"MAN(2).DAT", "MAN215"},
		{"MAN(3).DAT", "MAN315"},
	} {
		gf, err := assets.ParseGlyphFile(readGame(t, c.names))
		if err != nil {
			t.Fatalf("%s: %v", c.names, err)
		}
		count := len(gf.Glyphs) / GeneralNameSlotWidth

		data := readGame(t, c.dat)
		gs, err := ParseGenerals(data, count)
		if err != nil {
			t.Fatal(err)
		}
		unassigned := 0
		for i := range gs {
			p := gs[i].Province
			if p == 0 {
				unassigned++
				continue
			}
			if !p.Valid() {
				t.Fatalf("%s 第 %d 位將領的所屬省 %d 超出 1..%d",
					c.dat, i+1, p, ProvinceCount)
			}
		}
		t.Logf("%s：名表 %d 位，%d 位無所屬", c.dat, count, unassigned)

		// 殘料驗證：名表之後的空間確實有越界值，證明不能整檔解析。
		if extra := len(data)/GeneralRecordSize - count; extra > 0 {
			bad := 0
			for i := count; i < len(data)/GeneralRecordSize; i++ {
				if data[i*GeneralRecordSize+4] > ProvinceCount {
					bad++
				}
			}
			if bad == 0 {
				t.Errorf("%s 名表之後的 %d 筆沒有越界值，筆數上限可能抓錯",
					c.dat, extra)
			}
		}
	}
}
