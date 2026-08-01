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
//	+0        u16     領袖的將領 ID（1-based；0 = 空槽）
//	                  ⛔ 初稿讀成 u8 + 「+1 恆為 0（未解）」。`sub_3512B` 用
//	                  `mov es:[di], ax` 一次寫兩個 byte——**+1 是高位元組**。
//	+2        u8      未解，劇本開場由 `sub_32FB2` 寫死
//	+3..+10   8×u8    全部 0FFh，24 槽都一樣
//	+11..+34  24×u8   對其餘 24 個勢力的關係值（1-based 索引）
//	+35..+58  24×u8   與關係值成對的第二個維度（1-based；開局全零）
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

	facOffLeader    = 0 // u16
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

// 關係值的兩個常數，來自 `sub_34A4E`——某個勢力做了某件事之後，
// **其他所有勢力對它的關係值**被這樣處理：
//
//	rel < 150  →  rel = 100      （拉回預設）
//	rel ≥ 150  →  rel -= 20      （原本友好的，扣 20）
//
// 所以 100 是預設值、150 是一道門檻，而且**關係值可以超過 100**。
// 同一支函式還把兩邊的 `+35..+58` 都清成 0。
//
// ⚠️ 「150 以上 = 同盟」「`+35..+58` = 盟約狀態」是**推論**，不是驗過的。
// 觸發這支函式的事件也還沒追（呼叫者 `sub_3523B`）。
const (
	// FactionRelationDefault 是開局值，也是 `sub_34A4E` 的拉回目標。
	FactionRelationDefault = 100
	// FactionRelationFriendly 是 `sub_34A4E` 的分界（`cmp ..., 96h`）。
	FactionRelationFriendly = 150
	// FactionRelationPenalty 是超過門檻時扣的量（`sub ..., 14h`）。
	FactionRelationPenalty = 20
)

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
// 後 14 個殘留槽的 `+0` 也是非零（當 u16 讀是 27239、1548、40590…；
// 當初誤讀成 u8 時更糟——103、12、142 拿去查名冊還會查到人）。
// 第一版的 `Active()` 就是這樣寫的，被測試當場抓到。
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
		t[i].Leader = GeneralID(row[facOffLeader]) | GeneralID(row[facOffLeader+1])<<8
		t[i].Unknown2 = row[facOffUnknown2]
		copy(t[i].Relations[:], row[facOffRelation:facOffTrailing])
		copy(t[i].Trailing[:], row[facOffTrailing:])
	}
	return t, nil
}

// Initialized 回答這個槽有沒有被初始化過。
//
// 判準是對角線為 0。這原本是**從資料歸納**的啟發式，現在有了獨立佐證：
// `FactionLeaders`（`.DT1` 區塊 6）是同一份領袖清單，而且**空槽是乾淨的 0**。
// 兩者在北伐存檔上完全一致。
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

// FactionLeaders 是 `.DT1` 區塊 6（`byte_6EE68`，48 B）——**24 × u16 的勢力領袖表**。
//
// 內容與勢力表的 `+0` 欄一模一樣，差別在**這一份的空槽是乾淨的 0**，
// 而勢力表的空槽帶著未初始化的殘留。所以要問「有幾個勢力」，問這一份。
//
// 原版以 `[di-534Ah]` + 勢力編號×2 存取（`sub_3512B`），1-based。
type FactionLeaders [FactionSlotCount]GeneralID

// ParseFactionLeaders 解出領袖表。
func ParseFactionLeaders(data []byte) (FactionLeaders, error) {
	var out FactionLeaders
	blk, err := SaveBlockByGlobal("byte_6EE68")
	if err != nil {
		return out, err
	}
	if len(data) < blk.Offset+blk.Size {
		return out, fmt.Errorf("game: .DT1 只有 %d bytes，放不下領袖表（需要 %d）",
			len(data), blk.Offset+blk.Size)
	}
	b := data[blk.Offset : blk.Offset+blk.Size]
	for i := range out {
		out[i] = GeneralID(b[i*2]) | GeneralID(b[i*2+1])<<8
	}
	return out, nil
}

// Count 是有勢力的槽數。
func (f FactionLeaders) Count() int {
	n := 0
	for _, id := range f {
		if id != 0 {
			n++
		}
	}
	return n
}
