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
