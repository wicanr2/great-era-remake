package game

import (
	"encoding/binary"
	"fmt"
)

// 談判停火（政略指令 10），出自 `sub_211D5`（440 行）與它底下的
// `sub_20CF0`（同意判定）、`sub_21168`（取對手）、`sub_20E05`（套用結果）。
//
// 畫面上的話：「司令不在本省／欲在何省談判停火？／並無戰事／同意／拒絕」。

// 停火判定的兩檔門檻（`sub_20CF0`）。
//
//	roll = Random(10)
//	請求方佔上風 → roll ≥ 3 才同意   → 7/10
//	請求方居劣勢 → roll ≥ 8 才同意   → 2/10
//
// **你強對方才肯停火**——這條規則讀起來完全合理，
// 也是目前解出來最「有設計感」的一段 AI 判斷。
const (
	// CeasefireStage 是原版 sub_211D5 的可用期別。其它兩期會先顯示
	// 「無法使用」並返回，不進入省份輸入。
	CeasefireStage = uint8(1)

	CeasefireRollRange = 10
	// CeasefireStrongMin 是佔上風時的最低骰值（含）。
	CeasefireStrongMin = 3
	// CeasefireWeakMin 是居劣勢時的最低骰值（含）。
	CeasefireWeakMin = 8
)

// 兩張與戰爭狀態有關的表，位址從 `sub_20E05`／`sub_21168` 讀到。
//
// ⚠️ **兩張都還沒完全解**，記位址是為了以後接得上。
const (
	// WarRecordSize 是 `.DT1` 區塊 2 的每省記錄大小。
	WarRecordSize = 60
	// WarRecordGeneralSlots 是玩家派將／戰鬥記錄中每側可寫入的將領槽位數。
	// `sub_2DD1F` 逐項複製最多 10 個選中 ID，`sub_2D812` 再寫入
	// `ds:B346h` 記錄的 +12h 或 +26h；這裡保留槽位數，不替兩側命名成攻守。
	WarRecordGeneralSlots = 10
	// WarRecordResourceSlots 是每個原始分支的四種可搬運資源欄位數。
	WarRecordResourceSlots = 4
	// WarRecordGeneralList18Offset 與 WarRecordGeneralList38Offset 是兩個
	// 1-based 將領 ID 清單在 60-byte 記錄中的原始 byte offset。最後兩 bytes
	//（+58..+59）不屬於這兩組清單，仍保持未知。
	WarRecordGeneralList18Offset = 0x12
	WarRecordGeneralList38Offset = 0x26
	// 四種資源欄位的原始 offset。+0x12 是 union：Branch4 的第一個
	// 將領 ID 與 Branch8 的 Fuel 共用同一個 u16，不能同時寫入兩種語意。
	WarRecordBranch4GoldOffset       = 0x04
	WarRecordBranch4FoodOffset       = 0x06
	WarRecordBranch4AmmunitionOffset = 0x0C
	WarRecordBranch4FuelOffset       = 0x10
	WarRecordBranch8GoldOffset       = 0x08
	WarRecordBranch8FoodOffset       = 0x0A
	WarRecordBranch8AmmunitionOffset = 0x0E
	WarRecordBranch8FuelOffset       = 0x12

	// WarRecordAddr 是每省的戰爭記錄表：`ds:B346h` 起、**60 B/筆**，
	// 以省編號索引（`mul 3Ch`）。`+0`／`+2` 的寫入端是
	// `sub_3964E`：分別寫 `word_64942`／`word_64944`，也就是兩方勢力領袖
	// ID（`docs/re/31` §37）；其餘欄位仍不可猜。
	//
	// 39 省 × 60 = 2,340 B，正好落在 `CeasefireStateAddr` 之前。
	WarRecordAddr = 0xB346
	// CeasefireStateAddr 是停火狀態表：`ds:BCA5h` 起、每省 1 byte。
	//
	//	值 > 0  → 停火中
	//	同意時  `inc byte ptr [di-435Bh]`
	//
	// 遞增而不是設成固定值，形狀像**剩餘停火月數**，但沒有證據。
	CeasefireStateAddr = 0xBCA5
)

// WarRecordResources 是 `sub_3231A` 的四個輸出：黃金、糧食、彈藥、燃料。
// 這些欄位在 `.DT1` 記錄內是 u16；不把它們解釋成當前省份的最終資源。
type WarRecordResources struct {
	Gold       uint16
	Food       uint16
	Ammunition uint16
	Fuel       uint16
}

// WarRecordBranch 是同一筆 60-byte 戰爭記錄的兩種原始 layout 之一。
// `Branch4` 使用 +4/+6/+12/+16 資源並把將領清單寫到 +18；
// `Branch8` 使用 +8/+10/+14/+18 資源並把將領清單寫到 +38。
// 這兩個分支不是攻方／守方名稱；在未知分支的存檔快照中，另一組值可能只是
// 未初始化殘料，呼叫端必須依已知執行期狀態選用其中一組。
type WarRecordBranch struct {
	Resources  WarRecordResources
	GeneralIDs [WarRecordGeneralSlots]GeneralID
}

// WarRecordBranchKind 是寫回端明確選擇的原始 layout 分支。
// 數值保留反組譯中的分支標記（4／8），不把它們命名成攻方／守方。
type WarRecordBranchKind uint8

const (
	WarRecordBranch4 WarRecordBranchKind = 4
	WarRecordBranch8 WarRecordBranchKind = 8
)

// WarRecord 是 `.DT1` 區塊 2 的一省戰爭記錄。
//
// `SideALeader`／`SideBLeader` 保留原版兩個中性側別，不把它們硬命名成
// 攻方／守方；在寫回端的資料流中，Side A 對應 `word_64942`（首位單位的
// 效忠勢力領袖），Side B 對應 `word_64944`（當前交戰省司令）。
// Raw 保留其餘尚未解出的 56 bytes，避免未來接寫回時遺失資料。
type WarRecord struct {
	SideALeader GeneralID
	SideBLeader GeneralID
	// Branch4／Branch8 是原始 offset 分支，不是固定側別。兩者的
	// `Resources`／`GeneralIDs` 在檔案中分別落於不同區域；Branch4 的清單
	// 與 Branch8 的 Fuel 都使用 +18，因此不能同時視為一筆記錄的兩組有效值。
	Branch4 WarRecordBranch
	Branch8 WarRecordBranch
	Raw     [WarRecordSize]byte
}

// ParseWarRecords 從 `.DT1` 解析 39 筆 1-based 戰爭記錄。
//
// 回傳陣列的 index 0 保留為無效哨兵，省 1..39 對應原版的省編號。
// parser 另外解出 `sub_2D812`／`sub_2DD1F` 明確形成的兩個原始分支：
// 各自包含四個資源 word 與最多 10 個將領 ID。由於兩分支在 +18 有 union，
// 其餘 bytes 原樣保存；哪一分支在某次存檔有效，仍由上層 oracle 裁決。
func ParseWarRecords(save []byte) ([ProvinceCount + 1]WarRecord, error) {
	var records [ProvinceCount + 1]WarRecord
	blk, err := SaveBlockByGlobal("byte_6F532")
	if err != nil {
		return records, err
	}
	if blk.Size != ProvinceCount*WarRecordSize {
		return records, fmt.Errorf("game: 戰爭記錄區大小 %d，不是 %d × %d",
			blk.Size, ProvinceCount, WarRecordSize)
	}
	if len(save) < blk.Offset+blk.Size {
		return records, fmt.Errorf("game: .DT1 只有 %d bytes，放不下戰爭記錄（需要 %d）",
			len(save), blk.Offset+blk.Size)
	}
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		row := save[blk.Offset+int(p-1)*WarRecordSize:]
		copy(records[p].Raw[:], row[:WarRecordSize])
		records[p].SideALeader = GeneralID(binary.LittleEndian.Uint16(row[0:]))
		records[p].SideBLeader = GeneralID(binary.LittleEndian.Uint16(row[2:]))
		for i := 0; i < WarRecordGeneralSlots; i++ {
			off18 := WarRecordGeneralList18Offset + i*2
			off38 := WarRecordGeneralList38Offset + i*2
			records[p].Branch4.GeneralIDs[i] = GeneralID(binary.LittleEndian.Uint16(row[off18 : off18+2]))
			records[p].Branch8.GeneralIDs[i] = GeneralID(binary.LittleEndian.Uint16(row[off38 : off38+2]))
		}
		records[p].Branch4.Resources = WarRecordResources{
			Gold:       binary.LittleEndian.Uint16(row[0x04:0x06]),
			Food:       binary.LittleEndian.Uint16(row[0x06:0x08]),
			Ammunition: binary.LittleEndian.Uint16(row[0x0C:0x0E]),
			Fuel:       binary.LittleEndian.Uint16(row[0x10:0x12]),
		}
		records[p].Branch8.Resources = WarRecordResources{
			Gold:       binary.LittleEndian.Uint16(row[0x08:0x0A]),
			Food:       binary.LittleEndian.Uint16(row[0x0A:0x0C]),
			Ammunition: binary.LittleEndian.Uint16(row[0x0E:0x10]),
			Fuel:       binary.LittleEndian.Uint16(row[0x12:0x14]),
		}
	}
	return records, nil
}

// WriteWarRecords 把每省戰爭記錄的兩個已證實欄位寫回 `.DT1` 副本。
//
// 每列先以 `orig` 的原始 bytes 為基底，再只覆蓋 `+0`／`+2`；因此 `+4..+59`
// 的未知 byte（包含原版未初始化殘留）即使呼叫端誤改 `WarRecord.Raw` 也不會
// 被清掉。index 0 是規則層哨兵，不會寫進檔案。
func WriteWarRecords(orig []byte, records [ProvinceCount + 1]WarRecord) ([]byte, error) {
	blk, err := SaveBlockByGlobal("byte_6F532")
	if err != nil {
		return nil, err
	}
	if blk.Size != ProvinceCount*WarRecordSize {
		return nil, fmt.Errorf("game: 戰爭記錄區大小 %d，不是 %d × %d",
			blk.Size, ProvinceCount, WarRecordSize)
	}
	if len(orig) < blk.Offset+blk.Size {
		return nil, fmt.Errorf("game: .DT1 戰爭記錄需要 %d bytes，只有 %d",
			blk.Offset+blk.Size, len(orig))
	}
	out := append([]byte(nil), orig...)
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		base := blk.Offset + int(p-1)*WarRecordSize
		row := out[base : base+WarRecordSize]
		binary.LittleEndian.PutUint16(row[0:], uint16(records[p].SideALeader))
		binary.LittleEndian.PutUint16(row[2:], uint16(records[p].SideBLeader))
	}
	return out, nil
}

// WriteWarRecordBranch 將一筆已選定 layout 的 `.DT1` 戰爭記錄寫回副本。
//
// 這是比 WriteWarRecords 更窄的明確 API：呼叫端必須指出 Branch4 或
// Branch8，因為 +0x12 是兩者的 union。它只覆蓋該分支的四個資源 word
// 與傳入的將領 ID 前綴；未提供的清單槽位、另一分支欄位及 +0x3A..+0x3B
// 一律保留原始 bytes。這保留了 `sub_2D812` 的「依既有計數逐項寫入」證據，
// 不把未解的清除時機誤當成規則。
func WriteWarRecordBranch(orig []byte, province ProvinceID, branch WarRecordBranchKind,
	resources WarRecordResources, ids []GeneralID) ([]byte, error) {
	blk, err := SaveBlockByGlobal("byte_6F532")
	if err != nil {
		return nil, err
	}
	if blk.Size != ProvinceCount*WarRecordSize {
		return nil, fmt.Errorf("game: 戰爭記錄區大小 %d，不是 %d × %d",
			blk.Size, ProvinceCount, WarRecordSize)
	}
	if len(orig) < blk.Offset+blk.Size {
		return nil, fmt.Errorf("game: .DT1 戰爭記錄需要 %d bytes，只有 %d",
			blk.Offset+blk.Size, len(orig))
	}
	if !province.Valid() {
		return nil, fmt.Errorf("game: 戰爭記錄省份 %d 超出 1..%d", province, ProvinceCount)
	}
	if branch != WarRecordBranch4 && branch != WarRecordBranch8 {
		return nil, fmt.Errorf("game: 未知戰爭記錄 layout 分支 %d", branch)
	}
	if len(ids) > WarRecordGeneralSlots {
		return nil, fmt.Errorf("game: 戰爭記錄將領最多 %d 個，得到 %d",
			WarRecordGeneralSlots, len(ids))
	}
	seen := make(map[GeneralID]struct{}, len(ids))
	for i, id := range ids {
		if id == 0 {
			return nil, fmt.Errorf("game: 戰爭記錄將領第 %d 筆是 0 哨兵", i+1)
		}
		if _, ok := seen[id]; ok {
			return nil, fmt.Errorf("game: 戰爭記錄將領 %d 重複", id)
		}
		seen[id] = struct{}{}
	}

	var gold, food, ammunition, fuel, generals int
	switch branch {
	case WarRecordBranch4:
		gold, food, ammunition, fuel, generals = WarRecordBranch4GoldOffset,
			WarRecordBranch4FoodOffset, WarRecordBranch4AmmunitionOffset,
			WarRecordBranch4FuelOffset, WarRecordGeneralList18Offset
	case WarRecordBranch8:
		gold, food, ammunition, fuel, generals = WarRecordBranch8GoldOffset,
			WarRecordBranch8FoodOffset, WarRecordBranch8AmmunitionOffset,
			WarRecordBranch8FuelOffset, WarRecordGeneralList38Offset
	}
	out := append([]byte(nil), orig...)
	row := out[blk.Offset+int(province-1)*WarRecordSize:]
	binary.LittleEndian.PutUint16(row[gold:], resources.Gold)
	binary.LittleEndian.PutUint16(row[food:], resources.Food)
	binary.LittleEndian.PutUint16(row[ammunition:], resources.Ammunition)
	binary.LittleEndian.PutUint16(row[fuel:], resources.Fuel)
	for i, id := range ids {
		binary.LittleEndian.PutUint16(row[generals+i*2:], uint16(id))
	}
	return out, nil
}

// ParseCeasefireStates 從 `.DT1` 讀出 `ds:BCA5h` 的 39-byte 逐省表。
// 存檔區塊從省 1 開始，規則層保留第 0 格當無效哨兵。
func ParseCeasefireStates(save []byte) ([ProvinceCount + 1]uint8, error) {
	var states [ProvinceCount + 1]uint8
	b, err := SaveBlockByGlobal("byte_6FE56")
	if err != nil {
		return states, err
	}
	if len(save) < b.Offset+b.Size {
		return states, fmt.Errorf("game: .DT1 停火狀態表需要 %d bytes，只有 %d",
			b.Offset+b.Size, len(save))
	}
	copy(states[1:], save[b.Offset:b.Offset+b.Size])
	return states, nil
}

// WriteCeasefireStates 把已解析的停火狀態表寫回一份 `.DT1` 副本。
//
// 這一塊的形狀與「每省一 byte」已由 `sub_21168`／`sub_20E05` 證實；
// 值的長期語意仍保持中性（只把原始 byte 寫回）。index 0 是規則層的
// 無效哨兵，不會寫進檔案。其餘區域與傳入的 `orig` 完全保留。
func WriteCeasefireStates(orig []byte, states [ProvinceCount + 1]uint8) ([]byte, error) {
	b, err := SaveBlockByGlobal("byte_6FE56")
	if err != nil {
		return nil, err
	}
	if b.Size != ProvinceCount {
		return nil, fmt.Errorf("game: 停火狀態表大小 %d，不是 %d", b.Size, ProvinceCount)
	}
	if len(orig) < b.Offset+b.Size {
		return nil, fmt.Errorf("game: .DT1 停火狀態表需要 %d bytes，只有 %d",
			b.Offset+b.Size, len(orig))
	}
	out := append([]byte(nil), orig...)
	copy(out[b.Offset:b.Offset+b.Size], states[1:])
	return out, nil
}

// CeasefireResult 記錄一次停火談判的結果。
type CeasefireResult struct {
	Agreed bool
	Roll   int
	// RequesterStronger 是請求方在該省是不是佔上風。
	RequesterStronger bool
	// AttackForce / DefendForce 是雙方已部署單位的攻擊力總和。
	AttackForce, DefendForce int
}

// CeasefireProvinceLimit 回傳原版談判停火輸入的省份上限。
// sub_211D5 在第 1 期使用 (1-36)；第二、三期雖然劇本有 39 省，
// 仍會在輸入前因期別 gate 返回，所以這裡對不可用期別回傳 0。
func CeasefireProvinceLimit(stage uint8) int {
	if stage != CeasefireStage {
		return 0
	}
	s, err := ScenarioByStage(stage)
	if err != nil {
		return 0
	}
	return s.Provinces
}

// CeasefireRequester 做停火玩家路徑在輸入前的兩道原版檢查：
// 只有第 1 期可用，而且目前省的勢力司令本人必須駐在目前省。
// 回傳值是 sub_211D5 傳給 sub_20CF0 的請求方司令 ID。
func (w *AIWorld) CeasefireRequester(stage uint8, current ProvinceID) (GeneralID, error) {
	if stage != CeasefireStage {
		return 0, fmt.Errorf("game: 談判停火只在第 %d 期可用", CeasefireStage)
	}
	if w == nil || w.Table == nil {
		return 0, fmt.Errorf("game: 談判停火需要有效的遊戲世界")
	}
	prov, err := w.Table.At(current)
	if err != nil {
		return 0, err
	}
	if prov.Commander == 0 {
		return 0, fmt.Errorf("game: 目前省沒有司令")
	}
	for i := range w.Units {
		if w.Units[i].General == prov.Commander {
			if w.Units[i].Province != current {
				return 0, fmt.Errorf("game: 司令不在本省")
			}
			return prov.Commander, nil
		}
	}
	return 0, fmt.Errorf("game: 司令不在本省")
}

// ValidateCeasefireTarget 套用 sub_211D5 的玩家輸入前置條件。
// 它不消耗亂數，也不改停火表；同意／拒絕的成本由呼叫端在
// NegotiateCeasefire 返回後處理。
func (w *AIWorld) ValidateCeasefireTarget(stage uint8, current, target ProvinceID) (GeneralID, error) {
	requester, err := w.CeasefireRequester(stage, current)
	if err != nil {
		return 0, err
	}
	limit := CeasefireProvinceLimit(stage)
	if int(target) < 1 || int(target) > limit {
		return 0, fmt.Errorf("game: 停火目標省 %d 超出 1..%d", target, limit)
	}
	prov, err := w.Table.At(target)
	if err != nil {
		return 0, err
	}
	if prov.Commander == 0 {
		return 0, fmt.Errorf("game: 省 %d 沒有司令", target)
	}
	if !prov.InBattle() {
		return 0, fmt.Errorf("game: 省 %d 並無戰事", target)
	}
	return requester, nil
}

// BattleForces 累加某省在場單位的攻守雙方戰力（`sub_20CF0` 的前半）。
//
// 原版的篩選條件三個，全部對得上 `docs/spec/02` 的欄位：
//
//	將領 +4  == 目標省          （所屬省）
//	將領 +16 bit 2 == 1         （已部署）
//	將領 +8  == 0 → 守方，否則攻方
//
// ⚠️ **remake 差異**：`CombatUnit` 沒有建模 `+8`（攻守旗標），
// 但已建模 `Deployed` 對應 `+16` bit 2。守方仍以效忠對象判斷——
//
//	在場   → `Cell.Valid()`（`+5` != 0xFF）或 `Deployed`
//	守方   → 效忠對象 == 該省司令，其餘算攻方
//
// 在「某省正在交戰」這個前提下兩者等價（`docs/re/06` 的增援邏輯
// 就是拿 `+14` 與省份 `+20` 直接比）。要完全照抄得先把 `+8` 建模進來。
//
// 用的是攻擊力公式 `sub_5A0B9`（`strength.go` 的 `Strength`）——
// **AI 的評估函式與戰鬥用的是同一支**，`70-ai.md` §6e 已經記過，
// 這裡是第二個獨立的例子。
func (w *AIWorld) BattleForces(p ProvinceID) (attack, defend int) {
	prov, err := w.Table.At(p)
	if err != nil {
		return 0, 0
	}
	for i := range w.Units {
		u := &w.Units[i]
		if u.Province != p || (!u.Cell.Valid() && !u.Deployed) || i >= len(w.Strengths) {
			continue
		}
		f := Strength(w.Strengths[i], w.Opts)
		if u.Faction == prov.Commander {
			defend += f
		} else {
			attack += f
		}
	}
	return
}

// NegotiateCeasefire 在某省談停火（`sub_211D5` → `sub_20CF0`）。
//
// `requester` 是提出的一方（勢力領袖 ID）。判定照原版：
//
//	佔上風 = 請求方是守方且守方戰力 ≥ 攻方，或請求方是攻方且攻方 ≥ 守方
//	佔上風 → Random(10) ≥ 3 同意
//	劣勢   → Random(10) ≥ 8 同意
//
// ⚠️ 同意之後原版做的是 `inc` 停火狀態表（`ds:BCA5h`），
// **那張表的長期語意仍未定名**（見 `CeasefireStateAddr`），但原版在
// 談成時確實執行 `inc byte ptr [di-435Bh]`；因此這裡保留原始 byte 的
// 遞增行為，並把「剩餘幾個月」之類的解釋留給上層文件。
func (w *AIWorld) NegotiateCeasefire(p ProvinceID, requester GeneralID,
	rng *Rand) (CeasefireResult, error) {
	if w == nil || w.Table == nil {
		return CeasefireResult{}, fmt.Errorf("game: 談判停火需要有效的遊戲世界")
	}
	if rng == nil {
		return CeasefireResult{}, fmt.Errorf("game: 談判停火需要亂數來源")
	}
	prov, err := w.Table.At(p)
	if err != nil {
		return CeasefireResult{}, err
	}
	if prov.Commander == 0 {
		return CeasefireResult{}, fmt.Errorf("game: 省 %d 無主，談不了停火", p)
	}

	res := CeasefireResult{}
	res.Roll = rng.Int(CeasefireRollRange)
	res.AttackForce, res.DefendForce = w.BattleForces(p)

	// 請求方是這個省的司令 → 他是守方，比守方戰力；否則他是攻方。
	if prov.Commander == requester {
		res.RequesterStronger = res.DefendForce >= res.AttackForce
	} else {
		res.RequesterStronger = res.AttackForce >= res.DefendForce
	}

	min := CeasefireWeakMin
	if res.RequesterStronger {
		min = CeasefireStrongMin
	}
	res.Agreed = res.Roll >= min
	if res.Agreed {
		// 原版 `sub_20E05` 對該省停火表做 inc；uint8 的溢位行為
		// 正好對應 8086 byte inc 的 modulo-256 語意。
		w.CeasefireState[p]++
	}
	return res, nil
}
