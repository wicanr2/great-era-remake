package game

import (
	"encoding/binary"
	"fmt"
)

// DiplomacyLedgerSlots 是原版外交帳本的勢力索引數。
//
// `sub_215F0` 由省份司令反查 `word_70026`，得到 1..10 的勢力索引；
// `sub_391E1` 再用同一個索引初始化信用度與外債。因此這裡保留第 0 格
// 作為規則層哨兵，和省份／停火等 1-based 表一致。
const DiplomacyLedgerSlots = 10

// DiplomacyLedger 是 WAR.EXE 執行期的外交帳本。
//
// IDA 證據（WAR.EXE SHA-256 =
// `11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`，
// IDA Pro 9.4.0.260610，IDA linear address；原始運算元仍保留）：
//
//   - `[di-4225h]`：信用度 byte 表，索引 1..10，ds:BDDB 起；
//     index 1 的線性位址是 6FF8Ch。
//   - `[di-421Eh]`／`[di-421Ch]`：外債的 32-bit little-endian 表，
//     索引 1..10，ds:BDE2 起；index 1 的線性位址是 6FF96h。
//   - `sub_391E1` 對索引 1..10 寫入信用度 100 與外債 0。
//   - `sub_2164A` 核准貸款時扣信用度、增加省份黃金，並以 32-bit 加法
//     增加外債；`sub_223ED` 償還時先驗證外債足夠，再 32-bit 減債、扣
//     省份黃金、按每 500 黃金回補信用度並夾到 100。
//
// `sub_595D4`（IDA 595D4..598C0）把 runtime 表逐段複製到 `.DT1` 暫存記錄：
// `byte_6FF96` 以 28h bytes 寫到檔案 offset `391Ch`（14620），
// `byte_6FF8C` 以 0Ah bytes 寫到 offset `3912h`（14610）；
// `sub_59CBF` 在載入時以相同大小反向複製。因此這兩段不是未知空位，
// 而是可安全逐欄讀寫的 `.DT1` 區塊。`0x390E..0x3911` 是四個獨立的
// runtime byte，不能把它們誤當成外債第 1 槽。
type DiplomacyLedger struct {
	// Credit[1..10] 對應 `[di-4225h]`；Credit[0] 是無效哨兵。
	Credit [DiplomacyLedgerSlots + 1]uint8
	// Debt[1..10] 對應 `[di-421Eh]`／`[di-421Ch]`；Debt[0] 是無效哨兵。
	Debt [DiplomacyLedgerSlots + 1]uint32
}

// ParseDiplomacyLedger 從 `SAVE(N).DT1` 讀出區塊 8／9。
//
// 證據：WAR.EXE SHA-256 `11dbfcf24686ab7765f788b38514cefd2039d0f60b6bd517d89fb5a84c068015`；
// IDA Pro 9.4.0.260610、IDA linear address `6FF96`／`6FF8C`；
// `sub_59CBF` 的 IDA 位址 `59FB0..59FC3` 讀 0Ah／28h bytes，
// `sub_595D4` 的 `596FA..59725` 寫相同大小。未解區域不會被讀取或重建。
func ParseDiplomacyLedger(data []byte) (DiplomacyLedger, error) {
	debtBlock, creditBlock, err := diplomacyLedgerBlocks()
	if err != nil {
		return DiplomacyLedger{}, err
	}
	need := debtBlock.Offset + debtBlock.Size
	if end := creditBlock.Offset + creditBlock.Size; end > need {
		need = end
	}
	if len(data) < need {
		return DiplomacyLedger{}, fmt.Errorf("game: .DT1 太短，外交帳本需要至少 %d bytes，只有 %d", need, len(data))
	}
	var l DiplomacyLedger
	for slot := 1; slot <= DiplomacyLedgerSlots; slot++ {
		debtOffset := debtBlock.Offset + (slot-1)*4
		l.Debt[slot] = binary.LittleEndian.Uint32(data[debtOffset : debtOffset+4])
		l.Credit[slot] = data[creditBlock.Offset+slot-1]
	}
	return l, nil
}

// WriteDiplomacyLedger 只覆蓋 `.DT1` 已證實的區塊 8／9，其他 bytes 原樣保留。
// 第 0 格是 Go 的無效哨兵，不會寫入檔案；所有原始 1-based slot 都會寫回。
func WriteDiplomacyLedger(orig []byte, l DiplomacyLedger) ([]byte, error) {
	debtBlock, creditBlock, err := diplomacyLedgerBlocks()
	if err != nil {
		return nil, err
	}
	need := debtBlock.Offset + debtBlock.Size
	if end := creditBlock.Offset + creditBlock.Size; end > need {
		need = end
	}
	if len(orig) < need {
		return nil, fmt.Errorf("game: .DT1 太短，外交帳本需要至少 %d bytes，只有 %d", need, len(orig))
	}
	out := append([]byte(nil), orig...)
	for slot := 1; slot <= DiplomacyLedgerSlots; slot++ {
		debtOffset := debtBlock.Offset + (slot-1)*4
		binary.LittleEndian.PutUint32(out[debtOffset:debtOffset+4], l.Debt[slot])
		out[creditBlock.Offset+slot-1] = l.Credit[slot]
	}
	return out, nil
}

// diplomacyLedgerBlocks 由受版控的 SaveBlocks 取得兩段實際檔案位移。
// 不在解析器裡複製 magic number，避免把區塊 8／9 的讀寫順序誤當成檔案順序。
func diplomacyLedgerBlocks() (debt, credit SaveBlock, err error) {
	debt, err = SaveBlockByGlobal("byte_6FF96")
	if err != nil {
		return SaveBlock{}, SaveBlock{}, err
	}
	credit, err = SaveBlockByGlobal("byte_6FF8C")
	if err != nil {
		return SaveBlock{}, SaveBlock{}, err
	}
	if debt.Size != DiplomacyLedgerSlots*4 || credit.Size != DiplomacyLedgerSlots {
		return SaveBlock{}, SaveBlock{}, fmt.Errorf(
			"game: 外交帳本區塊形狀不符（外債 %d、信用度 %d）", debt.Size, credit.Size)
	}
	return debt, credit, nil
}

// NewDiplomacyLedger 是 `sub_391E1` 的可重現初始化：信用度 100、外債 0。
func NewDiplomacyLedger() DiplomacyLedger {
	var l DiplomacyLedger
	for i := 1; i <= DiplomacyLedgerSlots; i++ {
		l.Credit[i] = CreditMax
	}
	return l
}

func checkDiplomacySlot(slot int) error {
	if slot < 1 || slot > DiplomacyLedgerSlots {
		return fmt.Errorf("game: 外交勢力索引要在 1..%d（%d）", DiplomacyLedgerSlots, slot)
	}
	return nil
}

// CreditFor 讀取一個 1-based 勢力索引的信用度。
func (l *DiplomacyLedger) CreditFor(slot int) (uint8, error) {
	if err := checkDiplomacySlot(slot); err != nil {
		return 0, err
	}
	return l.Credit[slot], nil
}

// DebtFor 讀取一個 1-based 勢力索引的外債餘額。
func (l *DiplomacyLedger) DebtFor(slot int) (uint32, error) {
	if err := checkDiplomacySlot(slot); err != nil {
		return 0, err
	}
	return l.Debt[slot], nil
}

// RequestLoan 以帳本中的信用度執行一次貸款；核准後同時更新信用度與
// 32-bit 外債。這是規則層窄接入；`cmd/dsds` 的外交貸款 M0／償還 M1
// 會呼叫同一入口，帳本再由 `autosave` 以 `WriteDiplomacyLedger` 持久化到
// `.DT1` 區塊 8／9。
func (l *DiplomacyLedger) RequestLoan(w *AIWorld, p ProvinceID, slot, amount int,
	rng *Rand) (LoanResult, error) {
	if err := checkDiplomacySlot(slot); err != nil {
		return LoanResult{}, err
	}
	if w == nil || w.Table == nil {
		return LoanResult{}, fmt.Errorf("game: 貸款需要有效世界")
	}
	res, credit, err := w.RequestLoan(p, amount, l.Credit[slot], rng)
	if err != nil {
		return LoanResult{}, err
	}
	if !res.Approved {
		return res, nil
	}
	// 原版是 32-bit `add/adc`；Go 的 uint32 溢位同樣採 modulo-2^32，
	// 但一般遊戲額度遠低於此上限。
	l.Credit[slot] = credit
	l.Debt[slot] += uint32(res.Amount)
	return res, nil
}

// RepayResult 記錄帳本償還後的狀態。
type RepayResult struct {
	Units       int
	Amount      int
	CreditAfter uint8
	DebtAfter   uint32
}

// RepayDebt 以帳本中的外債餘額執行償還。原版輸入流程不允許超過目前
// 外債；因此先做餘額檢查，再呼叫既有的黃金／信用度規則。
func (l *DiplomacyLedger) RepayDebt(w *AIWorld, p ProvinceID, slot, amount int) (RepayResult, error) {
	if err := checkDiplomacySlot(slot); err != nil {
		return RepayResult{}, err
	}
	if w == nil || w.Table == nil {
		return RepayResult{}, fmt.Errorf("game: 償還外債需要有效世界")
	}
	if amount <= 0 {
		return RepayResult{}, fmt.Errorf("game: 償還額要為正（%d）", amount)
	}
	if uint64(amount) > uint64(l.Debt[slot]) {
		return RepayResult{}, fmt.Errorf("game: 償還額 %d 超過外債 %d", amount, l.Debt[slot])
	}
	units, credit, err := w.RepayDebt(p, amount, l.Credit[slot])
	if err != nil {
		return RepayResult{}, err
	}
	l.Debt[slot] -= uint32(amount)
	l.Credit[slot] = credit
	return RepayResult{Units: units, Amount: amount, CreditAfter: credit,
		DebtAfter: l.Debt[slot]}, nil
}
