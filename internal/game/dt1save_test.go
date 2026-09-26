package game

import "testing"

func TestWriteFactionTablePreservesUnknownColumns(t *testing.T) {
	orig := readGame(t, "SAVE(1).DT1")
	table, err := ParseFactionTable(orig)
	if err != nil {
		t.Fatal(err)
	}
	const slot = 0
	table[slot].Leader ^= 0x0100
	table[slot].Relations[1] = 123
	out, err := WriteFactionTable(orig, table)
	if err != nil {
		t.Fatal(err)
	}
	blk, err := SaveBlockByGlobal("byte_6EFAA")
	if err != nil {
		t.Fatal(err)
	}
	row := blk.Offset + slot*FactionSlotSize
	allowed := map[int]bool{row: true, row + 1: true, row + facOffRelation + 1: true}
	for _, off := range DiffBytes(orig, out, 0) {
		if !allowed[off] {
			t.Fatalf("勢力表 writer 改到未知欄位 offset %d", off)
		}
	}
	back, err := ParseFactionTable(out)
	if err != nil {
		t.Fatal(err)
	}
	if back[slot].Leader != table[slot].Leader || back[slot].Relations[1] != 123 {
		t.Fatalf("已知勢力欄位沒有寫回：%+v", back[slot])
	}
	if back[slot].Unknown2 != table[slot].Unknown2 || back[slot].Trailing != table[slot].Trailing {
		t.Fatal("未知勢力欄位不應被 writer 重建")
	}
}

func TestWriteFactionLeadersOnlyChangesKnownBlock(t *testing.T) {
	orig := readGame(t, "SAVE(1).DT1")
	leaders, err := ParseFactionLeaders(orig)
	if err != nil {
		t.Fatal(err)
	}
	leaders[0] ^= 0x0100
	out, err := WriteFactionLeaders(orig, leaders)
	if err != nil {
		t.Fatal(err)
	}
	blk, err := SaveBlockByGlobal("byte_6EE68")
	if err != nil {
		t.Fatal(err)
	}
	for _, off := range DiffBytes(orig, out, 0) {
		if off < blk.Offset || off >= blk.Offset+blk.Size {
			t.Fatalf("勢力領袖 writer 改到區塊外 offset %d", off)
		}
	}
	back, err := ParseFactionLeaders(out)
	if err != nil {
		t.Fatal(err)
	}
	if back != leaders {
		t.Fatalf("勢力領袖表寫回後不一致：%v vs %v", back, leaders)
	}
}

func TestWriteDT1RoundTripsCompleteKnownSnapshot(t *testing.T) {
	orig := readGame(t, "SAVE(1).DT1")
	state, err := ParseDT1(orig, 274)
	if err != nil {
		t.Fatal(err)
	}
	out, err := WriteDT1(orig, state)
	if err != nil {
		t.Fatal(err)
	}
	if diff := DiffBytes(orig, out, 8); len(diff) != 0 {
		t.Fatalf("完整 DT1 snapshot round-trip 不應改變 bytes：%v", diff)
	}
}

func TestWriteDT1WritesKnownSnapshotWithoutTouchingUnknownBytes(t *testing.T) {
	orig := readGame(t, "SAVE(1).DT1")
	state, err := ParseDT1(orig, 274)
	if err != nil {
		t.Fatal(err)
	}
	state.Provinces.Date.Month++
	state.Generals[0].Experience ^= 0x01
	state.Factions[0].Relations[1] ^= 0x01
	state.FactionLeaders[0] = 167 // 仍在本期 274 筆將領範圍內
	state.WarRecords[1].SideALeader ^= 0x0100
	state.CeasefireStates[1]++
	state.MajorPowerLeaders[0] ^= 0x0100
	state.Ledger.Credit[1]++
	state.Ledger.Debt[1]++
	out, err := WriteDT1(orig, state)
	if err != nil {
		t.Fatal(err)
	}
	allowed := func(off int) bool {
		if off < SaveHeaderSize {
			return off == 2 // 本測試只改月份
		}
		if off >= SaveArrayOffset && off < SaveArrayOffset+ProvinceCount*ProvinceRecordSize {
			return true
		}
		if off >= SaveGeneralsOffset && off < SaveGeneralsOffset+274*GeneralRecordSize {
			return true
		}
		for _, name := range []string{"byte_6EFAA", "byte_6F532", "byte_6FE56", "byte_6EE68", "byte_6EE98", "word_70026", "byte_6FF96", "byte_6FF8C"} {
			blk, e := SaveBlockByGlobal(name)
			if e == nil && off >= blk.Offset && off < blk.Offset+blk.Size {
				return true
			}
		}
		return false
	}
	for _, off := range DiffBytes(orig, out, 0) {
		if !allowed(off) {
			t.Fatalf("WriteDT1 改到未知 offset %d", off)
		}
	}
	back, err := ParseDT1(out, 274)
	if err != nil {
		t.Fatal(err)
	}
	if back.Provinces.Date.Month != state.Provinces.Date.Month ||
		back.FactionLeaders[0] != state.FactionLeaders[0] ||
		back.WarRecords[1].SideALeader != state.WarRecords[1].SideALeader ||
		back.CeasefireStates[1] != state.CeasefireStates[1] ||
		back.Ledger.Debt[1] != state.Ledger.Debt[1] {
		t.Fatal("DT1 已知欄位寫回後無法讀回相同快照")
	}
}

func TestWriteDT1RequiresProvinceSnapshot(t *testing.T) {
	if _, err := WriteDT1(make([]byte, SaveFileSize), DT1State{}); err == nil {
		t.Fatal("缺少省份快照的 DT1State 應 fail-closed")
	}
}
