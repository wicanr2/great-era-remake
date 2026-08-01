package game

import (
	"os"
	"path/filepath"
	"testing"
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
