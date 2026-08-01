package game

import "fmt"

// 勢力表：`SAVE(N).DT1` 的第 3 個 `$basg` 區塊（`byte_6EFAA`，1416 B）。
//
// 這一塊在 `docs/formats/07-dt1-layout.md` 曾經只知道 stride 是 59、語意未解。
// 2026-08-02 從實際位元組解開：**59 不是勢力數，是每個勢力槽的大小**，
// 24 才是槽數。
//
//	勢力槽 = 24 × 59 byte
//
//	+0        u8      領袖的將領 ID（1-based；0 = 空槽）
//	+1        u8      恆為 0（未解）
//	+2        u8      未解，劇本開場由 `sub_32FB2` 寫死
//	+3..+10   8×u8    全部 0FFh，24 槽都一樣
//	+11..+34  24×u8   對其餘 24 個勢力的關係值
//	+35..+58  24×u8   前 10 槽全零（未解）
//
// 11 + 24 + 24 = 59 ✔
//
// 判準不是「看起來像」：外交矩陣每一列的 0 精準落在對角線上（十列全中），
// 而且每一列在「對空槽」那一行也是 0——**兩個獨立的 0 分佈同時被解釋**。

const (
	// FactionSlotCount 是勢力槽數。
	//
	// ⚠️ 這一版劇本只用到前 10 槽，後 14 槽的 `+11` 之後是**未初始化的殘留**。
	// 讀進來的值不要當資料用——用 `HasFaction` 過濾。
	FactionSlotCount = 24
	// FactionSlotSize 是每個勢力槽的大小。
	FactionSlotSize = 59

	facOffLeader    = 0
	facOffUnknown1  = 1
	facOffUnknown2  = 2
	facOffSentinels = 3
	facOffRelation  = 11
	facOffTrailing  = 35
)

// FactionSentinel 是 `+3..+10` 那 8 個位元組的值。
//
// 24 個槽全都是這個值，所以那一段是**真的有初始化**（不像 `+11` 之後）。
// 用途未解；`0xFF` 在這個專案裡一貫是「沒有」的哨兵（`CLAUDE.md` §7 第 6.5 條）。
const FactionSentinel = 0xFF

// FactionRelationDefault 是開局時彼此的關係值。
//
// ⚠️ **不要叫它「敵對」**——開局所有勢力對彼此都是這個值，
// 而遊戲裡並不是所有人一開始就在打仗。定名要等外交指令的反組譯。
const FactionRelationDefault = 100

// FactionSlot 是一個勢力槽。
type FactionSlot struct {
	// Leader 是領袖的將領 ID（1-based），0 表示空槽。
	Leader GeneralID
	// Unknown2 是 `+2`。十個槽的值互不相同（1..14），而且**會隨時期改變**
	// （槽 1 閰錫山：時期 1 是 9，時期 2、3 是 3），所以不是顏色也不是固定編號。
	Unknown2 uint8
	// Relations 是對 24 個勢力的關係值（`+11..+34`）。自己對自己恆為 0。
	Relations [FactionSlotCount]uint8
	// Trailing 是 `+35..+58`。前 10 槽全零，形狀與 Relations 相同但語意未解。
	Trailing [FactionSlotCount]uint8
}

// ⛔ **不要用「領袖 ID 非零」判斷這個槽有沒有勢力。**
//
// 後 14 個殘留槽的 `+0` 也是非零（103、12、142…），拿去查名冊還會查到人，
// 看起來完全像資料。第一版的 `Active()` 就是這樣寫的，被測試當場抓到。
//
// 資料裡有更硬的判準：**對角線**。真正被初始化過的槽，
// `Relations[i][i]` 一定是 0（自己對自己）；殘留槽是隨機值。
// 所以判定分兩層，見 `Initialized` 與 `HasFaction`。

// FactionTable 是 24 個勢力槽。
type FactionTable [FactionSlotCount]FactionSlot

// ParseFactionTable 從勢力表區塊解出 24 個槽。
//
// `data` 是整份 `.DT1`；區塊位置取自 `SaveBlocks`。
func ParseFactionTable(data []byte) (FactionTable, error) {
	var t FactionTable
	blk, err := SaveBlockByGlobal("byte_6EFAA")
	if err != nil {
		return t, err
	}
	if len(data) < blk.Offset+blk.Size {
		return t, fmt.Errorf("game: .DT1 只有 %d bytes，放不下勢力表（需要 %d）",
			len(data), blk.Offset+blk.Size)
	}
	b := data[blk.Offset : blk.Offset+blk.Size]
	for i := range t {
		row := b[i*FactionSlotSize : (i+1)*FactionSlotSize]
		t[i].Leader = GeneralID(row[facOffLeader])
		t[i].Unknown2 = row[facOffUnknown2]
		copy(t[i].Relations[:], row[facOffRelation:facOffTrailing])
		copy(t[i].Trailing[:], row[facOffTrailing:])
	}
	return t, nil
}

// Initialized 回答這個槽有沒有被初始化過。
//
// 判準是對角線為 0。⚠️ 這是**從資料歸納**的判準，不是從程式碼驗的——
// 原版用什麼決定「有幾個勢力」還沒追到（候選：劇本表或某個全域計數）。
// 在追到之前，這是唯一擋得住殘留槽的判準。
func (t FactionTable) Initialized(i int) bool {
	return i >= 0 && i < FactionSlotCount && t[i].Relations[i] == 0
}

// HasFaction 回答這個槽有沒有勢力：初始化過**而且**領袖非零。
//
// 兩者的差別是空槽——北伐的槽 5 初始化過（整列 0 與 100 都對）但領袖是 0。
func (t FactionTable) HasFaction(i int) bool {
	return t.Initialized(i) && t[i].Leader != 0
}

// Relation 回傳 a 對 b 的關係值。
//
// ⚠️ 原版的矩陣**沒有保證對稱**——這裡照原樣回傳單向值，不做平均或取小。
// 開局時它是對稱的，但那是初始值的性質，不是資料結構的保證。
func (t FactionTable) Relation(a, b int) (uint8, error) {
	if a < 0 || a >= FactionSlotCount || b < 0 || b >= FactionSlotCount {
		return 0, fmt.Errorf("game: 勢力槽編號要在 0..%d，實得 %d／%d",
			FactionSlotCount-1, a, b)
	}
	return t[a].Relations[b], nil
}

// SlotOfLeader 找出某個將領是哪一槽的領袖，找不到回 -1。
func (t FactionTable) SlotOfLeader(id GeneralID) int {
	if id == 0 {
		return -1
	}
	for i, f := range t {
		if f.Leader == id && t.HasFaction(i) {
			return i
		}
	}
	return -1
}

// MajorPower 回答某個將領是不是勢力領袖。
//
// 這正是 `GameState.MajorPower` 需要的那個判定（原版 `sub_5C7FE`）——
// 在此之前那個欄位只能由呼叫端自己想辦法填，現在有資料來源了。
func (t FactionTable) MajorPower(id GeneralID) bool { return t.SlotOfLeader(id) >= 0 }

// ActiveFactions 回傳有勢力的槽編號。
func (t FactionTable) ActiveFactions() []int {
	var out []int
	for i := range t {
		if t.HasFaction(i) {
			out = append(out, i)
		}
	}
	return out
}

// InitializedSlots 回傳被初始化過的槽數——含領袖為 0 的空槽。
//
// 北伐劇本是 10（九個勢力 + 一個空槽）。
func (t FactionTable) InitializedSlots() int {
	n := 0
	for i := range t {
		if t.Initialized(i) {
			n++
		}
	}
	return n
}
