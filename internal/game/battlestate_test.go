package game

import "testing"

// TestBattleStateRoundTrip 驗證「改寫而非重建」：不改欄位時 Bytes() 必須
// byte-for-byte 相同（CLAUDE.md §9）。三個檔案都要過。
func TestBattleStateRoundTrip(t *testing.T) {
	for _, name := range []string{"MEM_WAR.DAT", "SAVE(1).DT2", "SAVE(2).DT2"} {
		data := readGame(t, name)
		sts, err := ParseBattleStates(data)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		for i := range sts {
			out := sts[i].Bytes()
			orig := data[i*BattleStateSize : (i+1)*BattleStateSize]
			for k := 0; k < BattleStateSize; k++ {
				if out[k] != orig[k] {
					t.Fatalf("%s 第 %d 省 round-trip 在 offset %d 不同：%#x vs %#x",
						name, i+1, k, out[k], orig[k])
				}
			}
		}
	}
}

func TestWriteBattleStatesRoundTrip(t *testing.T) {
	for _, name := range []string{"MEM_WAR.DAT", "SAVE(1).DT2", "SAVE(2).DT2"} {
		orig := readGame(t, name)
		states, err := ParseBattleStates(orig)
		if err != nil {
			t.Fatalf("%s: 解析失敗：%v", name, err)
		}
		out, err := WriteBattleStates(orig, states)
		if err != nil {
			t.Fatalf("%s: 寫回失敗：%v", name, err)
		}
		if len(out) != len(orig) {
			t.Fatalf("%s: 長度改變 %d → %d", name, len(orig), len(out))
		}
		if d := DiffBytes(orig, out, 8); len(d) != 0 {
			t.Fatalf("%s: 不改欄位卻有差異：%v", name, d)
		}
	}
}

func TestWriteBattleStatesTouchesOnlyParsedUnit(t *testing.T) {
	orig := readGame(t, "SAVE(1).DT2")
	states, err := ParseBattleStates(orig)
	if err != nil {
		t.Fatal(err)
	}
	const province = 17 // 陝西省，1-based 省號 18 的記錄 index
	const unit = 7
	states[province].UnitsA[unit] = 0xBEEF
	out, err := WriteBattleStates(orig, states)
	if err != nil {
		t.Fatal(err)
	}
	base := province*BattleStateSize + bsOffDetailA + unit*2
	d := DiffBytes(orig, out, 8)
	if len(d) != 2 || d[0] != base || d[1] != base+1 {
		t.Fatalf("改單一已解析 u16 應只動 %d、%d，實際：%v", base, base+1, d)
	}
	back, err := ParseBattleStates(out)
	if err != nil {
		t.Fatal(err)
	}
	if got := back[province].UnitsA[unit]; got != 0xBEEF {
		t.Fatalf("寫回後單位值為 %#x，預期 %#x", got, uint16(0xBEEF))
	}
}

func TestWriteBattleStatesRejectsWrongLength(t *testing.T) {
	var states [ProvinceCount]BattleState
	if _, err := WriteBattleStates(make([]byte, ProvinceCount*BattleStateSize-1), states); err == nil {
		t.Fatal("少一個 byte 的戰鬥狀態檔應拒絕寫回")
	}
	if _, err := WriteBattleStates(make([]byte, ProvinceCount*BattleStateSize+1), states); err == nil {
		t.Fatal("多一個 byte 的戰鬥狀態檔應拒絕寫回")
	}
}

func TestApplyRemakeSnapshotPreservesUnknownBytes(t *testing.T) {
	var b BattleState
	for i := range b.Raw {
		b.Raw[i] = byte((i*37 + 11) & 0xff)
	}
	// SlotsA/B 的 20 bytes 元素語意仍然未知；把它們設成非零哨兵，確認寫回不猜。
	copy(b.SlotsA[:], b.Raw[bsOffSlotsA:bsOffSlotsA+BattleSlots])
	copy(b.SlotsB[:], b.Raw[bsOffSlotsB:bsOffSlotsB+BattleSlots])
	before := b.Raw
	if err := b.ApplyRemakeSnapshot(19, BattleResources{Gold: 10, Food: 20, Ammo: 30, Fuel: 40},
		[]GeneralID{7, 9}, []GeneralID{12}); err != nil {
		t.Fatal(err)
	}
	out := b.Bytes()
	for i := range out {
		if (i >= bsOffHeader && i < bsOffHeader+8) ||
			(i >= bsOffRosterA && i < bsOffRosterA+BattleSlots*2) ||
			(i >= bsOffRosterB && i < bsOffRosterB+BattleSlots*2) ||
			(i >= bsOffDetailA && i < bsOffDetailA+BattleUnitArea) ||
			(i >= bsOffDetailB && i < bsOffDetailB+BattleUnitArea) || i == bsOffTrailing {
			continue
		}
		if out[i] != before[i] {
			t.Fatalf("未知 byte %#x 被改寫：%#x -> %#x", i, before[i], out[i])
		}
	}
	got, err := ParseBattleState(out[:])
	if err != nil {
		t.Fatal(err)
	}
	if got.Header != [4]uint16{10, 20, 30, 40} || got.RosterA[0] != 7 || got.RosterA[1] != 9 ||
		got.RosterB[0] != 12 || got.UnitsA[0] != 7 || got.UnitsA[1] != 9 ||
		got.UnitsB[0] != 12 || got.Trailing != 19 {
		t.Fatalf("remake snapshot 欄位錯誤：%+v", got)
	}
}

func TestApplyRemakeSnapshotRejectsInvalidInput(t *testing.T) {
	cases := []struct {
		name string
		from ProvinceID
		atk  []GeneralID
		def  []GeneralID
	}{
		{name: "bad province", from: 0},
		{name: "zero attacker", from: 1, atk: []GeneralID{0}},
		{name: "duplicate defender", from: 1, def: []GeneralID{3, 3}},
		{name: "too many", from: 1, atk: []GeneralID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var b BattleState
			if err := b.ApplyRemakeSnapshot(tc.from, BattleResources{}, tc.atk, tc.def); err == nil {
				t.Fatal("預期非法輸入被拒絕")
			}
		})
	}
}

// TestMemWarResidue 記錄 MEM_WAR.DAT 裡的未初始化殘料。
//
// 31 個省的兩個 200 B 單位區是乾淨的 0，**但有 8 個省不是**。
// 緬甸（省 39）那筆裡甚至有 8086 機器碼（`b8 00 25 cd 21` =
// mov ax,0025h / int 21h），顯然不是遊戲資料——是寫檔時把未初始化的
// 記憶體一起寫進去了。Turbo Pascal 的變數不會自動清零，1992 年的程式常見。
//
// [記帳] 這一條原本寫成「39 省全部乾淨」，因為我只看了湖北與河南兩個省。
// 又一次「用少數資料點驗證自己」（docs/playtest/02 §6）。
//
// **後果是實質的**：不能拿 MEM_WAR.DAT 當戰前基準去 diff 出部隊欄位，
// 那 8 個省會餵進垃圾。要打進實際戰鬥取 .DT2 才可信。
func TestMemWarResidue(t *testing.T) {
	sts, err := ParseBattleStates(readGame(t, "MEM_WAR.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	var dirty []int
	for i := range sts {
		if sts[i].Engaged() {
			dirty = append(dirty, i+1)
		}
	}
	want := []int{2, 3, 15, 17, 18, 33, 37, 39}
	if len(dirty) != len(want) {
		t.Fatalf("有殘料的省應為 %v，實得 %v", want, dirty)
	}
	for i := range want {
		if dirty[i] != want[i] {
			t.Fatalf("有殘料的省應為 %v，實得 %v", want, dirty)
		}
	}
	t.Logf("%d/%d 省的單位區是乾淨的", ProvinceCount-len(dirty), ProvinceCount)
}

// TestBattleSlotsUseEmptyMarker 驗證 10 B 的槽位用 0xFF 當空槽。
//
// 若偏移抓錯，落進來的會是 +0 那四個 u16 的碎片，不會呈現
// 「不是 0xFF 就是小值」這種分佈。
//
// 實測 638/780 個槽是 0xFF、57 個是 0，其餘 85 個散落在 1..190。
// 數字是釘住現況用的回歸檢查——這裡**不宣稱**非空槽是什麼。
// 戰前狀態下兩個 200 B 單位區全是 0，所以這些槽裝的不是戰鬥部隊。
func TestBattleSlotsUseEmptyMarker(t *testing.T) {
	sts, err := ParseBattleStates(readGame(t, "MEM_WAR.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	var empty, other, maxV int
	for i := range sts {
		for _, s := range [][BattleSlots]byte{sts[i].SlotsA, sts[i].SlotsB} {
			for _, v := range s {
				if v == EmptySlot {
					empty++
					continue
				}
				other++
				if int(v) > maxV {
					maxV = int(v)
				}
			}
		}
	}
	if empty != 638 {
		t.Errorf("MEM_WAR.DAT 的 0xFF 槽應為 638 個，實得 %d——偏移或解讀改了", empty)
	}
	if other != 142 {
		t.Errorf("非 0xFF 的槽應為 142 個，實得 %d", other)
	}
	// 非 0xFF 的值全部落在 0..190，遠低於 byte 上限——是編號不是旗標位元。
	if maxV != 190 {
		t.Errorf("非 0xFF 槽的最大值應為 190，實得 %d", maxV)
	}
}

// TestUnitIDProjectionPreservesOrderAndOnlyDropsZero 驗證 +68/+268 的已證實
// runtime 將領清單投影：0 是空槽，非零值保留原始順序，不在 parser 層做值域
// 清洗。這點很重要，因為初始 MEM_WAR.DAT 有未初始化殘料。
func TestUnitIDProjectionPreservesOrderAndOnlyDropsZero(t *testing.T) {
	var b BattleState
	b.UnitsA[0] = 58
	b.UnitsA[2] = 0xBEEF // 故意超出第一期將領數，投影仍不得擅自丟掉
	b.UnitsA[5] = 166
	b.UnitsB[1] = 127
	b.UnitsB[4] = 162

	if got, want := b.AttackerUnitIDs(), []GeneralID{58, 0xBEEF, 166}; !equalGeneralIDs(got, want) {
		t.Fatalf("攻方 unit ID 清單 = %v，預期 %v", got, want)
	}
	if got, want := b.DefenderUnitIDs(), []GeneralID{127, 162}; !equalGeneralIDs(got, want) {
		t.Fatalf("守方 unit ID 清單 = %v，預期 %v", got, want)
	}
}

func equalGeneralIDs(a, b []GeneralID) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestRosterIsAttackerDefender 驗證 +28/+48 是攻守雙方的參戰單位。
//
// 河南那筆分得乾乾淨淨：攻方全是張作霖系的東北／河北將領，
// 守方全是河南本地。用將領記錄的所屬省欄位（docs/spec/02 附錄）交叉比對。
func TestRosterIsAttackerDefender(t *testing.T) {
	sts, err := ParseBattleStates(readGame(t, "MEM_WAR.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	data := readGame(t, "MAN(1).DAT")
	gs, err := ParseGenerals(data, 274)
	if err != nil {
		t.Fatal(err)
	}
	provinceOf := func(id GeneralID) ProvinceID {
		if id == 0 || int(id) > len(gs) {
			return 0
		}
		return gs[id-1].Province
	}

	henan := sts[18] // 省 19
	att, def := henan.Attackers(), henan.Defenders()
	if len(att) != 6 || len(def) != 7 {
		t.Fatalf("河南的攻守單位數應為 6 / 7，實得 %d / %d", len(att), len(def))
	}
	// 守方全部來自河南本省
	for _, id := range def {
		if p := provinceOf(id); p != 19 {
			t.Errorf("守方將領 %d 的所屬省是 %d，預期 19（河南）", id, p)
		}
	}
	// 攻方全部不是河南的
	for _, id := range att {
		if p := provinceOf(id); p == 19 {
			t.Errorf("攻方將領 %d 的所屬省是河南，不該出現在攻方", id)
		}
	}
	// 攻方應該同屬一個勢力（張作霖系：東北九省 + 河北）
	zhang := map[ProvinceID]bool{1: true, 2: true, 3: true, 4: true, 5: true,
		6: true, 7: true, 8: true, 9: true, 11: true, 20: true}
	for _, id := range att {
		if p := provinceOf(id); !zhang[p] {
			t.Errorf("攻方將領 %d 來自省 %d，不在張作霖的轄區內", id, p)
		}
	}
	t.Logf("河南：攻方 %v，守方 %v", att, def)
}

// TestUnitAreaIsWordArray 驗證 200 B 的單位區切成 100 個 u16 之後，
// 仍然通過 byte-for-byte round-trip。
//
// [訂正] 這一區原本切成 [200]byte 並假設「10 個單位 × 20 B」。
// 實際上 sub_545B0／sub_5446D 遍歷它時迴圈上限是 100、以 word 為單位，
// 所以是 100 個 u16。切法改了但 round-trip 必須照樣過——
// 那是「改寫而非重建」的底線。
func TestUnitAreaIsWordArray(t *testing.T) {
	if BattleUnits*2 != BattleUnitArea {
		t.Fatalf("100 個 u16 應該正好是 %d bytes，實得 %d",
			BattleUnitArea, BattleUnits*2)
	}
	data := readGame(t, "MEM_WAR.DAT")
	sts, err := ParseBattleStates(data)
	if err != nil {
		t.Fatal(err)
	}
	// 緬甸（省 39）那筆有未初始化殘料，正好拿來驗非零資料的 round-trip
	st := sts[38]
	nonzero := 0
	for _, v := range st.UnitsA {
		if v != 0 {
			nonzero++
		}
	}
	if nonzero == 0 {
		t.Fatal("省 39 的 UnitsA 應該有殘料，測試前提不成立")
	}
	out := st.Bytes()
	orig := data[38*BattleStateSize : 39*BattleStateSize]
	for k := 0; k < BattleStateSize; k++ {
		if out[k] != orig[k] {
			t.Fatalf("省 39 round-trip 在 offset %d 不同：%#x vs %#x",
				k, out[k], orig[k])
		}
	}
	t.Logf("省 39 的 UnitsA 有 %d 個非零 word，round-trip 通過", nonzero)
}
