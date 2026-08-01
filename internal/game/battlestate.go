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
	bsOffRosterA  = 28  // 10 × u16，攻方的參戰單位（將領 ID）
	bsOffRosterB  = 48  // 10 × u16，守方
	bsOffDetailA  = 68  // 200 B，攻方每單位 20 B
	bsOffDetailB  = 268 // 200 B，守方
	bsOffTrailing = 468 // 1 B

	// BattleSlots 是每方的**部隊**數。10 B 與 10 × u16 兩組都是這個數。
	BattleSlots = 10
	// BattleUnitArea 是每方單位區的 byte 數。
	BattleUnitArea = 200
	// BattleUnits 是每方的**單位**數：200 B ÷ 2 = 100 個 u16。
	//
	// 這是從程式碼確認的，不是從 byte 數推的——`sub_545B0` 與 `sub_5446D`
	// 遍歷 `word[0x6742 + i*2]` 與 `word[0x680A + i*2]` 時，
	// 迴圈上限是 `64h = 100`，而兩個基址相差 `0xC8 = 200 bytes = 100 word`。
	BattleUnits = 100
	// EmptySlot 是空槽標記。
	EmptySlot = 0xFF
)

// BattleState 是一個省的戰鬥狀態。
type BattleState struct {
	// Header 是開頭的 4 個 u16。語意未解。
	Header [4]uint16
	// SlotsA / SlotsB 各 10 個 byte，0xFF 表示空槽。
	SlotsA, SlotsB [BattleSlots]byte
	// RosterA 是**攻方**的參戰單位（將領 ID，0 表示空槽），
	// RosterB 是**守方**。每方最多 10 個。
	//
	// 證據：河南那筆的 RosterA 全是張作霖系的東北／河北將領
	// （所屬省 1/2/5/7/11），RosterB 全是河南本地（所屬省 19）
	// ——攻守分得乾乾淨淨。
	//
	// `sub_54CFD` 依 `byte_64901` 選用其中一組，也印證了這是成對的兩方。
	//
	// ⚠️ 這**不是**該省的將領名冊。湖北的兩組加起來只有 8 個，
	// 而該省有 15 位將領（docs/spec/02）——參戰的是其中一部分。
	RosterA, RosterB [BattleSlots]uint16
	// UnitsA / UnitsB 是兩方各 **100 個 u16** 的單位陣列。
	//
	// [訂正] 這裡原本切成 `[200]byte` 並假設「每方 10 個單位 × 20 B」。
	// **那個假設是錯的**——`sub_545B0`／`sub_5446D` 遍歷這兩塊時迴圈上限是
	// 100，而且是以 word 為單位（`shl di, 1`）。
	//
	// 所以戰鬥狀態有**兩層**：10 個部隊（Roster）與 100 個單位（這裡）。
	// 詞表也分得很清楚——`3.15` 的 48 是「單位數」、49 是「士兵數」。
	//
	// 初始狀態下全是 0，每一格裝什麼仍未知。
	UnitsA, UnitsB [BattleUnits]uint16
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
		b.RosterA[i] = binary.LittleEndian.Uint16(rec[bsOffRosterA+i*2:])
		b.RosterB[i] = binary.LittleEndian.Uint16(rec[bsOffRosterB+i*2:])
	}
	for i := 0; i < BattleUnits; i++ {
		b.UnitsA[i] = binary.LittleEndian.Uint16(rec[bsOffDetailA+i*2:])
		b.UnitsB[i] = binary.LittleEndian.Uint16(rec[bsOffDetailB+i*2:])
	}
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
		binary.LittleEndian.PutUint16(out[bsOffRosterA+i*2:], b.RosterA[i])
		binary.LittleEndian.PutUint16(out[bsOffRosterB+i*2:], b.RosterB[i])
	}
	for i := 0; i < BattleUnits; i++ {
		binary.LittleEndian.PutUint16(out[bsOffDetailA+i*2:], b.UnitsA[i])
		binary.LittleEndian.PutUint16(out[bsOffDetailB+i*2:], b.UnitsB[i])
	}
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

// Attackers 回傳攻方的參戰單位（將領 ID），已濾掉空槽。
func (b *BattleState) Attackers() []GeneralID { return roster(b.RosterA) }

// Defenders 回傳守方的參戰單位（將領 ID），已濾掉空槽。
func (b *BattleState) Defenders() []GeneralID { return roster(b.RosterB) }

func roster(r [BattleSlots]uint16) []GeneralID {
	var out []GeneralID
	for _, v := range r {
		if v != 0 {
			out = append(out, GeneralID(v))
		}
	}
	return out
}
