package game

import "testing"

// TestWriteProvincesUnchanged 不改任何欄位時，寫回的檔案必須與原始
// byte-for-byte 相同——包含 1447 之後那 13,236 bytes 未解區域。
func TestWriteProvincesUnchanged(t *testing.T) {
	orig := readGame(t, "SAVE(1).DT1")
	tbl, err := ParseSaveProvinces(orig)
	if err != nil {
		t.Fatal(err)
	}
	out, err := WriteProvinces(orig, tbl)
	if err != nil {
		t.Fatal(err)
	}
	if d := DiffBytes(orig, out, 8); len(d) != 0 {
		t.Fatalf("不改欄位卻有 %d 處不同，前幾個 offset：%v", len(d), d)
	}
	if len(out) != len(orig) {
		t.Fatalf("檔案長度變了：%d → %d", len(orig), len(out))
	}
}

// TestWriteProvincesTouchesOnlyTarget 改一個欄位時，**只有那兩個 byte 變**。
//
// 這才是「改寫而非重建」的真正驗收（CLAUDE.md §9）：
// round-trip 相同只證明沒壞掉，這一條證明沒有波及未解區域。
func TestWriteProvincesTouchesOnlyTarget(t *testing.T) {
	orig := readGame(t, "SAVE(1).DT1")
	tbl, err := ParseSaveProvinces(orig)
	if err != nil {
		t.Fatal(err)
	}
	p, err := tbl.At(26) // 湖北
	if err != nil {
		t.Fatal(err)
	}
	p.Gold = 12345
	out, err := WriteProvinces(orig, tbl)
	if err != nil {
		t.Fatal(err)
	}

	// 湖北是 index 25，黃金在記錄 +0
	want := SaveArrayOffset + 25*ProvinceRecordSize
	d := DiffBytes(orig, out, 16)
	if len(d) != 2 || d[0] != want || d[1] != want+1 {
		t.Fatalf("改黃金應該只動 offset %d..%d 兩個 byte，實際動了 %v", want, want+1, d)
	}

	// 讀回來要是新值，而且其他欄位不變
	back, err := ParseSaveProvinces(out)
	if err != nil {
		t.Fatal(err)
	}
	bp, _ := back.At(26)
	if bp.Gold != 12345 {
		t.Errorf("讀回來的黃金是 %d，預期 12345", bp.Gold)
	}
	if bp.Food != 18050 || bp.Commander != 58 {
		t.Errorf("其他欄位被波及：糧食 %d、司令 %d", bp.Food, bp.Commander)
	}
}

// TestWriteTownFileUnchanged TOWN 檔的相位不同（從 offset 0 起），一樣要 round-trip。
func TestWriteTownFileUnchanged(t *testing.T) {
	for _, name := range []string{"TOWN(1).DAT", "TOWN(2).DAT", "TOWN(3).DAT"} {
		orig := readGame(t, name)
		tbl, err := ParseTownFile(orig)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		out, err := WriteTownFile(orig, tbl)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		if d := DiffBytes(orig, out, 8); len(d) != 0 {
			t.Fatalf("%s 不改欄位卻有 %d 處不同：%v", name, len(d), d)
		}
	}
}

// TestDateRoundTrip 檔頭的年月也要能改寫回去。
func TestDateRoundTrip(t *testing.T) {
	orig := readGame(t, "SAVE(1).DT1")
	tbl, err := ParseSaveProvinces(orig)
	if err != nil {
		t.Fatal(err)
	}
	tbl.Date.Month = 8
	out, err := WriteProvinces(orig, tbl)
	if err != nil {
		t.Fatal(err)
	}
	d := DiffBytes(orig, out, 8)
	if len(d) != 1 || d[0] != 2 {
		t.Fatalf("改月份應該只動 offset 2，實際動了 %v", d)
	}
	back, _ := ParseSaveProvinces(out)
	if back.Date.Month != 8 || back.Date.Year != 15 {
		t.Errorf("讀回來是民國 %d 年 %d 月", back.Date.Year, back.Date.Month)
	}
}

// TestAutosaveDoesNotTouchOriginal 模擬離開時的自動存檔：
// 寫出的是副本，原始 bytes 不變。
//
// CLAUDE.md §9：原版資產唯讀，測試存檔一律寫到明確的輸出目錄。
func TestAutosaveDoesNotTouchOriginal(t *testing.T) {
	orig := readGame(t, "SAVE(1).DT1")
	snapshot := make([]byte, len(orig))
	copy(snapshot, orig)

	tbl, err := ParseSaveProvinces(orig)
	if err != nil {
		t.Fatal(err)
	}
	p, _ := tbl.At(26)
	p.Gold = 999

	out, err := WriteProvinces(orig, tbl)
	if err != nil {
		t.Fatal(err)
	}
	if d := DiffBytes(orig, snapshot, 4); len(d) != 0 {
		t.Fatalf("WriteProvinces 改動了傳入的原始 bytes，offset %v", d)
	}
	back, _ := ParseSaveProvinces(out)
	bp, _ := back.At(26)
	if bp.Gold != 999 {
		t.Errorf("副本沒有寫入新值")
	}
}

func TestWriteSavePersistsGeneralStateAndPreservesUnknownBytes(t *testing.T) {
	orig := readGame(t, "SAVE(1).DT1")
	tbl, err := ParseSaveProvinces(orig)
	if err != nil {
		t.Fatal(err)
	}
	gens, err := ParseSaveGenerals(orig, 274)
	if err != nil {
		t.Fatal(err)
	}

	const i = 57 // 湖北第一位將領（吳佩孚），有實機欄位錨點。
	gens[i].Experience = 31
	gens[i].AbilityB = 87
	gens[i].Force = 12345
	gens[i].Stamina = 76
	gens[i].F30 = 68
	out, err := WriteSave(orig, tbl, gens)
	if err != nil {
		t.Fatal(err)
	}
	back, err := ParseSaveGenerals(out, 274)
	if err != nil {
		t.Fatal(err)
	}
	g := back[i]
	if g.Experience != 31 || g.AbilityB != 87 || g.Force != 12345 ||
		g.Stamina != 76 || g.F30 != 68 {
		t.Fatalf("將領狀態沒有完整寫回：%+v", g)
	}

	base := SaveGeneralsOffset + i*GeneralRecordSize
	allowed := map[int]bool{
		base + genOffExperience: true,
		base + genOffAbilityB:   true,
		base + genOffForce:      true,
		base + genOffForce + 1:  true,
		base + genOffStamina:    true,
		base + genOffF30:        true,
	}
	for _, off := range DiffBytes(orig, out, 0) {
		if !allowed[off] {
			t.Errorf("WriteSave 動到未授權 offset %d", off)
		}
	}
}
