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
// 已證實的欄位語意會以附加方法提供；尚未解出的 bytes 仍保留在 Raw，不能
// 因為 Go 欄位命名就把整份記錄當成已完全理解。

// BattleStateSize 是一筆戰鬥狀態記錄的大小，來自 Reset(f, 1D5h)。
const BattleStateSize = 469

// 記錄內的偏移。名稱刻意中性（SideA/SideB），因為「哪一邊是攻方」未確認。
const (
	bsOffHeader   = 0   // 4 × u16：第一方帶進場資源
	bsOffSlotsA   = 8   // 10 B；形狀／0xFF 空槽已知，元素語意仍未知
	bsOffSlotsB   = 18  // 10 B；形狀／0xFF 空槽已知，元素語意仍未知
	bsOffRosterA  = 28  // 10 × u16，攻方的參戰單位（將領 ID）
	bsOffRosterB  = 48  // 10 × u16，守方
	bsOffDetailA  = 68  // 200 B = 100 × u16，攻方單位陣列
	bsOffDetailB  = 268 // 200 B = 100 × u16，守方單位陣列
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
	// Header 是第一方的四項戰鬥資源快照：黃金、糧食、彈藥、燃料。
	//
	// 證據：IDA 的 sub_4F468 將記錄 +0、+2、+4、+6 載入
	// word_64932、word_64936、word_6493A、word_6493E；sub_3964E
	// 以相同順序寫回。第二方的 word_64934／64938／6493C／64940
	// 在該寫回分支直接更新省份記錄，沒有再放進這 469 bytes 的開頭。
	// A/B 仍維持中性命名；哪一方在所有戰鬥分支都等同攻方尚未證實。
	Header [4]uint16
	// SlotsA / SlotsB 各 10 個 byte。0xFF 是目前樣本中的空槽標記；
	// 兩組陣列的元素是什麼（部隊索引、旗標或其他狀態）仍未知，
	// 因此 parser 只保留原始 byte，不替它們加上玩法語意。
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
	// UnitsA / UnitsB 是兩方各 **100 個 u16** 的 runtime 將領清單。
	//
	// [訂正] 這裡原本切成 `[200]byte` 並假設「每方 10 個單位 × 20 B」。
	// **那個假設是錯的**——`sub_545B0`／`sub_5446D` 遍歷這兩塊時迴圈上限是
	// 100，而且是以 word 為單位（`shl di, 1`）。
	//
	// 所以戰鬥狀態有**兩層**：10 個部隊（Roster）與 100 個 runtime 將領槽（這裡）。
	// `sub_4F6FF` 會把非零的省份參戰資料壓入 runtime 位址 `0x6742`／`0x680A`；
	// `sub_545B0`／`sub_5446D` 再以每個 u16 × 0x21 加上 runtime 記錄基址 0x7A7D，
	// 直接索引 33-byte 將領記錄。這證實非零值是 1-based 將領 ID，0 是空槽。
	//
	// 初始檔仍可能含未初始化殘料；parser 不做值域清洗，保留原始 u16。
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
// 未解區域一個 byte 都不動（AGENTS.md §8）。
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

// WriteBattleStates 把整份 `.DT2`／`MEM_WAR.DAT` 寫回副本。
//
// 每個 BattleState 都應由 ParseBattleStates 取得；Bytes 會以該狀態保留的
// Raw bytes 為基底，只覆蓋目前已切出的欄位。這讓尚未解出的區域不會因為
// 重建結構而被清成 Go 零值（AGENTS.md §8）。orig 不會被修改。
func WriteBattleStates(orig []byte, states [ProvinceCount]BattleState) ([]byte, error) {
	want := ProvinceCount * BattleStateSize
	if len(orig) != want {
		return nil, fmt.Errorf("game: 戰鬥狀態檔應為 %d bytes（%d 省 × %d），實得 %d",
			want, ProvinceCount, BattleStateSize, len(orig))
	}
	out := make([]byte, len(orig))
	copy(out, orig)
	for i := range states {
		rec := states[i].Bytes()
		copy(out[i*BattleStateSize:], rec[:])
	}
	return out, nil
}

// Attackers 回傳攻方的參戰單位（將領 ID），已濾掉空槽。
func (b *BattleState) Attackers() []GeneralID { return roster(b.RosterA) }

// Defenders 回傳守方的參戰單位（將領 ID），已濾掉空槽。
func (b *BattleState) Defenders() []GeneralID { return roster(b.RosterB) }

// AttackerUnitIDs 回傳檔案 +68 區的非零值，作為 runtime 將領 ID 清單的原始投影，
// 並保留原始槽位順序。IDA 已確認 runtime 陣列 `0x6742` 的元素定位；檔案全域緩衝
// 與 runtime 陣列的同步時機仍待正常戰鬥快照補證。
//
// 這不是把殘料判成合法將領：原版的未初始化存檔可能含超出劇本人數的 u16，
// 因此這個投影只移除明確的 0 哨兵，不做值域驗證。要驗證是否能對上目前劇本，
// 呼叫端必須另以將領表檢查。
func (b *BattleState) AttackerUnitIDs() []GeneralID { return unitIDs(b.UnitsA) }

// DefenderUnitIDs 回傳檔案 +268 區的非零值，作為 runtime 將領 ID 清單的原始投影，
// 並保留原始槽位順序。規則與 AttackerUnitIDs 相同：0 是空槽，其他值原樣呈現。
func (b *BattleState) DefenderUnitIDs() []GeneralID { return unitIDs(b.UnitsB) }

// ApplyRemakeSnapshot 把 remake 的一場戰鬥投影到目前已解出的戰鬥狀態欄位。
//
// 這不是對原版 sub_3964E 的時機宣稱，而是玩家副本的明確寫回契約：
// +0..+7 保存第一方的四項資源、+28/+48 保存兩方參戰將領、+68/+268
// 保存兩方 runtime 將領 ID 清單，+468 保存攻方來源省。SlotsA/B 的元素語意
// 尚未解出，故刻意不碰；Bytes 會從 Raw 帶回所有未解 bytes。
//
// runtime 清單是 remake 的快照投影：呼叫端提供的順序會填入前 100 槽，
// 其餘清為零。不要把這個方法當成原版每一回合的同步證據。
func (b *BattleState) ApplyRemakeSnapshot(from ProvinceID, resources BattleResources,
	attackers, defenders []GeneralID) error {
	if b == nil {
		return fmt.Errorf("game: nil 戰鬥狀態不能寫回")
	}
	if !from.Valid() {
		return fmt.Errorf("game: 來源省 %d 超出 1..%d", from, ProvinceCount)
	}
	if err := validateBattleSnapshotIDs("攻方", attackers); err != nil {
		return err
	}
	if err := validateBattleSnapshotIDs("守方", defenders); err != nil {
		return err
	}
	seen := make(map[GeneralID]struct{}, len(attackers)+len(defenders))
	for _, ids := range [][]GeneralID{attackers, defenders} {
		for _, id := range ids {
			if _, exists := seen[id]; exists {
				return fmt.Errorf("game: 兩方參戰將領 %d 重複", id)
			}
			seen[id] = struct{}{}
		}
	}
	b.Header = [4]uint16{resources.Gold, resources.Food, resources.Ammo, resources.Fuel}
	b.RosterA = [BattleSlots]uint16{}
	b.RosterB = [BattleSlots]uint16{}
	for i, id := range attackers {
		b.RosterA[i] = uint16(id)
	}
	for i, id := range defenders {
		b.RosterB[i] = uint16(id)
	}
	b.UnitsA = [BattleUnits]uint16{}
	b.UnitsB = [BattleUnits]uint16{}
	for i, id := range attackers {
		b.UnitsA[i] = uint16(id)
	}
	for i, id := range defenders {
		b.UnitsB[i] = uint16(id)
	}
	b.Trailing = byte(from)
	return nil
}

func validateBattleSnapshotIDs(side string, ids []GeneralID) error {
	if len(ids) > BattleSlots {
		return fmt.Errorf("game: %s參戰將領最多 %d 個，得到 %d", side, BattleSlots, len(ids))
	}
	seen := make(map[GeneralID]struct{}, len(ids))
	for i, id := range ids {
		if id == 0 {
			return fmt.Errorf("game: %s參戰將領第 %d 筆是 0 哨兵", side, i+1)
		}
		if _, ok := seen[id]; ok {
			return fmt.Errorf("game: %s參戰將領 %d 重複", side, id)
		}
		seen[id] = struct{}{}
	}
	return nil
}

func roster(r [BattleSlots]uint16) []GeneralID {
	var out []GeneralID
	for _, v := range r {
		if v != 0 {
			out = append(out, GeneralID(v))
		}
	}
	return out
}

func unitIDs(r [BattleUnits]uint16) []GeneralID {
	var out []GeneralID
	for _, v := range r {
		if v != 0 {
			out = append(out, GeneralID(v))
		}
	}
	return out
}
