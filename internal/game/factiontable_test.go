package game

import "testing"

// 11 + 24 + 24 = 59，而且 24 × 59 要對得上區塊大小。
func TestFactionSlotArithmetic(t *testing.T) {
	if facOffRelation+2*FactionSlotCount != FactionSlotSize {
		t.Errorf("11 + 24 + 24 該等於 %d，實得 %d",
			FactionSlotSize, facOffRelation+2*FactionSlotCount)
	}
	blk, err := SaveBlockByGlobal("byte_6EFAA")
	if err != nil {
		t.Fatal(err)
	}
	if blk.Size != FactionSlotCount*FactionSlotSize {
		t.Errorf("勢力表區塊 %d B，24 × 59 = %d", blk.Size, FactionSlotCount*FactionSlotSize)
	}
}

// ⭐ 這是解開這一塊的關鍵判準：外交矩陣每一列的 0 要精準落在對角線上。
// 十個有效槽全中，才算這個切法是對的。
func TestFactionRelationDiagonalIsZero(t *testing.T) {
	tbl, err := ParseFactionTable(readGame(t, "SAVE(1).DT1"))
	if err != nil {
		t.Fatal(err)
	}
	active := tbl.ActiveFactions()
	if len(active) != 9 {
		// 北伐十槽裡有一個是空槽（領袖 ID 0），所以有勢力的是 9 個。
		t.Fatalf("有勢力的槽該是 9 個，實得 %d：%v", len(active), active)
	}
	for _, i := range active {
		if got := tbl[i].Relations[i]; got != 0 {
			t.Errorf("槽 %d 對自己的關係值該是 0，實得 %d", i, got)
		}
		for j := 0; j < 10; j++ {
			if i == j || !tbl.HasFaction(j) {
				continue
			}
			if got := tbl[i].Relations[j]; got != FactionRelationDefault {
				t.Errorf("開局時槽 %d 對槽 %d 該是 %d，實得 %d",
					i, j, FactionRelationDefault, got)
			}
		}
	}
}

// 領袖 ID 要對得上名冊裡的十大勢力（`docs/formats/07` §4）。
func TestFactionLeadersMatchHistory(t *testing.T) {
	tbl, err := ParseFactionTable(readGame(t, "SAVE(1).DT1"))
	if err != nil {
		t.Fatal(err)
	}
	want := []GeneralID{166, 4, 3, 58, 98, 0, 146, 157, 156, 1}
	for i, w := range want {
		if got := tbl[i].Leader; got != w {
			t.Errorf("槽 %d 的領袖該是 %d，實得 %d", i, w, got)
		}
	}
	// 空槽不算勢力領袖。
	if tbl.MajorPower(0) {
		t.Error("將領 ID 0 不該被當成勢力領袖")
	}
	if !tbl.MajorPower(58) {
		t.Error("吳佩孚（58）該是勢力領袖")
	}
	if tbl.SlotOfLeader(999) != -1 {
		t.Error("不存在的將領該回 -1")
	}
}

// `+3..+10` 那 8 個哨兵在 24 個槽都要有——這是「後 14 槽不是完全沒碰過」的證據。
func TestFactionSentinelsPresentInAllSlots(t *testing.T) {
	data := readGame(t, "SAVE(1).DT1")
	blk, err := SaveBlockByGlobal("byte_6EFAA")
	if err != nil {
		t.Fatal(err)
	}
	b := data[blk.Offset : blk.Offset+blk.Size]
	for i := 0; i < FactionSlotCount; i++ {
		row := b[i*FactionSlotSize : (i+1)*FactionSlotSize]
		for k := facOffSentinels; k < facOffRelation; k++ {
			if row[k] != FactionSentinel {
				t.Fatalf("槽 %d 的 +%d 該是 %#x，實得 %#x", i, k, FactionSentinel, row[k])
			}
		}
	}
}

// ⚠️ 後 14 槽的 `+11` 之後是殘留，不是資料。這個測試把那件事釘住，
// 免得之後有人拿 tbl[10..23] 的關係值去做判斷。
func TestFactionInactiveSlotsAreNotData(t *testing.T) {
	tbl, err := ParseFactionTable(readGame(t, "SAVE(1).DT1"))
	if err != nil {
		t.Fatal(err)
	}
	// ⛔ 這些槽的**領袖 ID 是非零的**（103、12、142…），拿去查名冊還查得到人。
	// 第一版的 Active() 用「領袖非零」判定，就是在這裡爆掉的。
	for i := 10; i < FactionSlotCount; i++ {
		if tbl[i].Leader == 0 {
			t.Errorf("槽 %d 的領袖竟然是 0——那這個測試就沒有在測它該測的東西", i)
		}
		if tbl.Initialized(i) {
			t.Errorf("槽 %d 不該被判成初始化過（對角線 %d）", i, tbl[i].Relations[i])
		}
		if tbl.HasFaction(i) {
			t.Errorf("槽 %d 是殘留，不該被當成勢力", i)
		}
	}
	if got := tbl.InitializedSlots(); got != 10 {
		t.Errorf("北伐該有 10 個初始化過的槽，實得 %d", got)
	}
}

func TestFactionRelationRejectsBadIndex(t *testing.T) {
	var tbl FactionTable
	if _, err := tbl.Relation(-1, 0); err == nil {
		t.Error("槽編號 -1 該回錯")
	}
	if _, err := tbl.Relation(0, FactionSlotCount); err == nil {
		t.Error("槽編號超界該回錯")
	}
}
