package game

import (
	"reflect"
	"testing"
)

// TestApproximateDT2DiffMaskAllRecords 是免 DOSBox 的 DT2/MEM_WAR 差分閘門。
// 它只驗證 remake writer 的非破壞性契約：未知的 +8/+18 與其他保留 bytes
// 不得因為 Go 零值或整筆重建而被改掉。這不是原版 sub_3964E 的 parity 測試。
func TestApproximateDT2DiffMaskAllRecords(t *testing.T) {
	orig := make([]byte, ProvinceCount*BattleStateSize)
	for i := range orig {
		orig[i] = byte((i*73 + 19) & 0xff)
	}
	states, err := ParseBattleStates(orig)
	if err != nil {
		t.Fatal(err)
	}
	for i := range states {
		states[i].Header = [4]uint16{uint16(i + 1), uint16(i + 101), uint16(i + 201), uint16(i + 301)}
		states[i].RosterA[0] = uint16(1 + i)
		states[i].RosterB[0] = uint16(101 + i)
		states[i].UnitsA[0] = uint16(1 + i)
		states[i].UnitsB[0] = uint16(101 + i)
		states[i].Trailing = byte(0x80 | (i & 0x3f))
	}
	out, err := WriteBattleStates(orig, states)
	if err != nil {
		t.Fatal(err)
	}
	for off, before := range orig {
		if out[off] == before {
			continue
		}
		within := off % BattleStateSize
		known := (within >= bsOffHeader && within < bsOffHeader+8) ||
			(within >= bsOffRosterA && within < bsOffRosterA+BattleSlots*2) ||
			(within >= bsOffRosterB && within < bsOffRosterB+BattleSlots*2) ||
			(within >= bsOffDetailA && within < bsOffDetailA+BattleUnitArea) ||
			(within >= bsOffDetailB && within < bsOffDetailB+BattleUnitArea) ||
			within == bsOffTrailing
		if !known {
			t.Fatalf("DT2 近似 writer 改到保留 offset %d（record +%d）", off, within)
		}
	}
	// 額外把兩段未解槽位明確列出；上面的全檔遮罩已涵蓋它們，這裡讓失敗訊息
	// 直接表達本規格最重要的邊界。
	for i := 0; i < ProvinceCount; i++ {
		base := i * BattleStateSize
		for _, span := range [][2]int{{bsOffSlotsA, BattleSlots}, {bsOffSlotsB, BattleSlots}} {
			for j := 0; j < span[1]; j++ {
				if out[base+span[0]+j] != orig[base+span[0]+j] {
					t.Fatalf("DT2 未解槽位被改寫：record=%d +%d", i+1, span[0]+j)
				}
			}
		}
	}
}

// TestApproximateDT1KeepsUnresolvedSentinels 固定幾個最容易被重建 writer
// 誤蓋的 DT1 未解區：檔頭第 4 byte、戰爭記錄尾端、勢力表 +2／尾端、
// 以及檔案中夾著的未映射 runtime bytes。它接受已知欄位差分，但不接受未知
// byte 被 Go 零值清除。
func TestApproximateDT1KeepsUnresolvedSentinels(t *testing.T) {
	orig := readGame(t, "SAVE(1).DT1")
	state, err := ParseDT1(orig, 274)
	if err != nil {
		t.Fatal(err)
	}
	state.Provinces.Date.Month++
	state.Generals[0].Experience++
	state.WarRecords[1].SideALeader ^= 0x0100
	state.Factions[0].Relations[1]++
	state.CeasefireStates[1]++
	state.Ledger.Credit[1]++
	out, err := WriteDT1(orig, state)
	if err != nil {
		t.Fatal(err)
	}
	war, err := SaveBlockByGlobal("byte_6F532")
	if err != nil {
		t.Fatal(err)
	}
	factions, err := SaveBlockByGlobal("byte_6EFAA")
	if err != nil {
		t.Fatal(err)
	}
	block10, err := SaveBlockByGlobal("word_70026")
	if err != nil {
		t.Fatal(err)
	}
	unknown := []int{
		3, // header fourth byte
		war.Offset + WarRecordSize - 2,
		war.Offset + WarRecordSize - 1,
		factions.Offset + facOffUnknown2,
		factions.Offset + facOffTrailing,
		block10.Offset,
		block10.Offset + block10.Size - 1,
	}
	for _, off := range unknown {
		if out[off] != orig[off] {
			t.Fatalf("DT1 未解 sentinel offset %d 被改寫：%#x -> %#x", off, orig[off], out[off])
		}
	}
	// SaveTrailingBytes 不是連續尾端；用「已知 writer 欄位」遮罩，避免近似
	// 驗證把未命名的七個 runtime byte 或區塊內 unknown 欄位當成可寫欄位。
	knownWrite := func(off int) bool {
		if off == 2 {
			return true // 本測試只改月份
		}
		if off >= SaveArrayOffset && off < SaveArrayOffset+ProvinceCount*ProvinceRecordSize {
			return true
		}
		if off >= SaveGeneralsOffset && off < SaveGeneralsOffset+SaveGeneralCount*GeneralRecordSize {
			return true
		}
		if off >= war.Offset && off < war.Offset+war.Size {
			row := (off - war.Offset) % WarRecordSize
			return row < 4 // WriteWarRecords 只覆蓋 +0/+2
		}
		if off >= factions.Offset && off < factions.Offset+factions.Size {
			row := (off - factions.Offset) % FactionSlotSize
			return row < 2 || (row >= facOffSentinels && row < facOffTrailing) ||
				(row >= facOffRelation && row < facOffTrailing)
		}
		for _, name := range []string{"byte_6EE68", "byte_6EE98", "byte_6FE56", "byte_6FF96", "byte_6FF8C", "word_70026"} {
			blk, e := SaveBlockByGlobal(name)
			if e == nil && off >= blk.Offset && off < blk.Offset+blk.Size {
				return true
			}
		}
		return false
	}
	for off := 0; off < len(orig); off++ {
		if out[off] == orig[off] || knownWrite(off) {
			continue
		}
		// 本測試改動的 block10／戰爭分支欄位有各自窄 writer，這些未解位置
		// 若變動就是 writer 越界；逐 byte fail-closed。
		if out[off] != orig[off] {
			t.Fatalf("DT1 非可寫 offset %d 發生差分：%#x -> %#x", off, orig[off], out[off])
		}
	}
}

// TestApproximateAIDecisionTraceIsDeterministic 把 AI 時機縮成事件序列，
// 驗證同一合成狀態不會因觀測或呼叫順序而漂移。這是 remake 的 deterministic
// contract，不是原版 sub_3A9F4/sub_3AABA 的逐回合 oracle。
func TestApproximateAIDecisionTraceIsDeterministic(t *testing.T) {
	a := mkTracedBattle(t, 20000, 18000)
	b := mkTracedBattle(t, 20000, 18000)
	ta, oa := a.TraceDecisions(16, BattleChainGates{}, 201, 58)
	tb, ob := b.TraceDecisions(16, BattleChainGates{}, 201, 58)
	if !reflect.DeepEqual(ta, tb) || oa != ob {
		t.Fatalf("相同合成狀態的 AI trace 不具決定性\nA=%+v / %+v\nB=%+v / %+v", ta, oa, tb, ob)
	}
	for _, row := range ta {
		if row.Turn < 1 || row.Turn > BattleTurnLimit {
			t.Fatalf("AI trace 回合超出近似契約：%+v", row)
		}
		if BattleActionName(row.A.Action) == "未知行動" || BattleActionName(row.B.Action) == "未知行動" {
			t.Fatalf("AI trace 含未接行動：%+v", row)
		}
	}
}
