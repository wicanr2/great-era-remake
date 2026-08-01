package game

import (
	"encoding/binary"
	"fmt"
)

// 依據：docs/re/05-mem-war-record.md
//
// MEM_WAR.DAT（初始戰鬥狀態）與 SAVE(N).DT2（存檔的戰鬥狀態）是同一個結構：
// 39 省 × 469 bytes。欄位切法從 WAR.EXE 的 sub_4F468（讀）與 sub_3964E（寫）
// 兩邊對稱驗來，零剩餘。
//
// **欄位語意大多未解**——兩個 200 B 區在初始狀態下全是 0，要打進一場實際
// 戰鬥拿到對照才驗得到。所以這裡只切結構、保留原始 bytes，不猜語意。

// BattleStateSize 是一筆戰鬥狀態記錄的大小，來自 Reset(f, 1D5h)。
const BattleStateSize = 469

// 記錄內的偏移。名稱刻意中性（SideA/SideB），因為「哪一邊是攻方」未確認。
const (
	bsOffHeader   = 0   // 4 × u16
	bsOffSlotsA   = 8   // 10 B，空槽填 0xFF
	bsOffSlotsB   = 18  // 10 B
	bsOffValuesA  = 28  // 10 × u16
	bsOffValuesB  = 48  // 10 × u16
	bsOffUnitsA   = 68  // 200 B
	bsOffUnitsB   = 268 // 200 B
	bsOffTrailing = 468 // 1 B

	// BattleSlots 是每方的槽數。10 B 與 10 × u16 兩組都是這個數。
	BattleSlots = 10
	// BattleUnitArea 是每方的單位資料區大小。
	BattleUnitArea = 200
	// EmptySlot 是空槽標記。
	EmptySlot = 0xFF
)

// BattleState 是一個省的戰鬥狀態。
type BattleState struct {
	// Header 是開頭的 4 個 u16。語意未解。
	Header [4]uint16
	// SlotsA / SlotsB 各 10 個 byte，0xFF 表示空槽。
	SlotsA, SlotsB [BattleSlots]byte
	// ValuesA / ValuesB 各 10 個 u16，0 表示空。
	ValuesA, ValuesB [BattleSlots]uint16
	// UnitsA / UnitsB 是兩方各 200 B 的單位資料區。
	//
	// 初始狀態下全是 0。**欄位切法未知**——若每方 10 個單位就是 20 B/單位，
	// 但那是從槽數推的，沒有直接證據。原樣保留。
	UnitsA, UnitsB [BattleUnitArea]byte
	// Trailing 是最後一個 byte。語意未解。
	Trailing byte

	// Raw 是完整的 469 bytes，寫回時以它為基底。
	Raw [BattleStateSize]byte
}

// OccupiedA 回傳 SlotsA 裡非空的槽數。
func (b *BattleState) OccupiedA() int { return countOccupied(b.SlotsA) }

// OccupiedB 回傳 SlotsB 裡非空的槽數。
func (b *BattleState) OccupiedB() int { return countOccupied(b.SlotsB) }

func countOccupied(s [BattleSlots]byte) int {
	n := 0
	for _, v := range s {
		if v != EmptySlot {
			n++
		}
	}
	return n
}

// Engaged 回報這個省有沒有部隊在場。
//
// 判準是兩個 200 B 區不全為 0——初始檔（MEM_WAR.DAT）全部 39 省都是 0。
func (b *BattleState) Engaged() bool {
	for _, v := range b.UnitsA {
		if v != 0 {
			return true
		}
	}
	for _, v := range b.UnitsB {
		if v != 0 {
			return true
		}
	}
	return false
}

// ParseBattleState 解一筆 469 bytes 的戰鬥狀態。
func ParseBattleState(rec []byte) (BattleState, error) {
	var b BattleState
	if len(rec) < BattleStateSize {
		return b, fmt.Errorf("game: 戰鬥狀態需要 %d bytes，只有 %d",
			BattleStateSize, len(rec))
	}
	copy(b.Raw[:], rec)
	for i := range b.Header {
		b.Header[i] = binary.LittleEndian.Uint16(rec[bsOffHeader+i*2:])
	}
	copy(b.SlotsA[:], rec[bsOffSlotsA:])
	copy(b.SlotsB[:], rec[bsOffSlotsB:])
	for i := 0; i < BattleSlots; i++ {
		b.ValuesA[i] = binary.LittleEndian.Uint16(rec[bsOffValuesA+i*2:])
		b.ValuesB[i] = binary.LittleEndian.Uint16(rec[bsOffValuesB+i*2:])
	}
	copy(b.UnitsA[:], rec[bsOffUnitsA:])
	copy(b.UnitsB[:], rec[bsOffUnitsB:])
	b.Trailing = rec[bsOffTrailing]
	return b, nil
}

// Bytes 產生寫回用的 469 bytes：以原始 bytes 為基底，只蓋已切出的欄位。
//
// 未解區域一個 byte 都不動（CLAUDE.md §9）。
func (b *BattleState) Bytes() [BattleStateSize]byte {
	out := b.Raw
	for i, v := range b.Header {
		binary.LittleEndian.PutUint16(out[bsOffHeader+i*2:], v)
	}
	copy(out[bsOffSlotsA:], b.SlotsA[:])
	copy(out[bsOffSlotsB:], b.SlotsB[:])
	for i := 0; i < BattleSlots; i++ {
		binary.LittleEndian.PutUint16(out[bsOffValuesA+i*2:], b.ValuesA[i])
		binary.LittleEndian.PutUint16(out[bsOffValuesB+i*2:], b.ValuesB[i])
	}
	copy(out[bsOffUnitsA:], b.UnitsA[:])
	copy(out[bsOffUnitsB:], b.UnitsB[:])
	out[bsOffTrailing] = b.Trailing
	return out
}

// ParseBattleStates 解整個 MEM_WAR.DAT 或 SAVE(N).DT2：39 省各一筆。
func ParseBattleStates(data []byte) ([ProvinceCount]BattleState, error) {
	var out [ProvinceCount]BattleState
	if want := ProvinceCount * BattleStateSize; len(data) != want {
		return out, fmt.Errorf("game: 戰鬥狀態檔應為 %d bytes（%d 省 × %d），實得 %d",
			want, ProvinceCount, BattleStateSize, len(data))
	}
	for i := 0; i < ProvinceCount; i++ {
		b, err := ParseBattleState(data[i*BattleStateSize:])
		if err != nil {
			return out, fmt.Errorf("game: 第 %d 省: %w", i+1, err)
		}
		out[i] = b
	}
	return out, nil
}
