package game

import (
	"encoding/binary"
	"testing"
)

func TestParseDiplomacyLedgerUsesDT1Block8And9(t *testing.T) {
	save := readGame(t, "SAVE(1).DT1")
	l, err := ParseDiplomacyLedger(save)
	if err != nil {
		t.Fatal(err)
	}
	// `0x390E..0x3911` 的 01 01 01 00 是四個獨立 runtime byte；
	// 真正的信用度從 0x3912（14610）開始，外債則從 0x391C（14620）開始。
	for slot := 1; slot <= DiplomacyLedgerSlots; slot++ {
		if l.Debt[slot] != 0 {
			t.Fatalf("SAVE(1) 外債[%d] = %#x，預期目前樣本為 0", slot, l.Debt[slot])
		}
		if l.Credit[slot] != CreditMax {
			t.Fatalf("SAVE(1) 信用度[%d] = %d，預期目前樣本為 %d", slot, l.Credit[slot], CreditMax)
		}
	}

	save2 := readGame(t, "SAVE(2).DT1")
	l2, err := ParseDiplomacyLedger(save2)
	if err != nil {
		t.Fatal(err)
	}
	for slot := 1; slot <= DiplomacyLedgerSlots; slot++ {
		if l2.Debt[slot] != 0 || l2.Credit[slot] != CreditMax {
			t.Fatalf("SAVE(2) 帳本[%d] = debt %#x／credit %d，預期 0／%d",
				slot, l2.Debt[slot], l2.Credit[slot], CreditMax)
		}
	}
}

func TestWriteDiplomacyLedgerOnlyChangesProvenLedgerBytes(t *testing.T) {
	orig := readGame(t, "SAVE(1).DT1")
	l, err := ParseDiplomacyLedger(orig)
	if err != nil {
		t.Fatal(err)
	}
	l.Debt[1] = 0xAABBCCDD
	l.Credit[1] = 77
	out, err := WriteDiplomacyLedger(orig, l)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != len(orig) {
		t.Fatalf("寫回長度 = %d，預期 %d", len(out), len(orig))
	}
	for i := range orig {
		allowed := (i >= 14620 && i < 14660) || (i >= 14610 && i < 14620)
		if !allowed && out[i] != orig[i] {
			t.Fatalf("未授權 offset %d 被改寫：%02x -> %02x", i, orig[i], out[i])
		}
	}
	if got := binary.LittleEndian.Uint32(out[14620:14624]); got != 0xAABBCCDD {
		t.Fatalf("外債[1] bytes = %#x", got)
	}
	if out[14610] != 77 {
		t.Fatalf("信用度[1] = %d，預期 77", out[14610])
	}
	back, err := ParseDiplomacyLedger(out)
	if err != nil {
		t.Fatal(err)
	}
	if back.Debt[1] != l.Debt[1] || back.Credit[1] != l.Credit[1] {
		t.Fatalf("讀回帳本不一致：%#x/%d vs %#x/%d", back.Debt[1], back.Credit[1], l.Debt[1], l.Credit[1])
	}
}

func TestParseDiplomacyLedgerRejectsShortSave(t *testing.T) {
	debt, _, err := diplomacyLedgerBlocks()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := ParseDiplomacyLedger(make([]byte, debt.Offset+debt.Size-1)); err == nil {
		t.Fatal("短於外債區塊的存檔應拒絕")
	}
}

func TestNewDiplomacyLedgerMatchesRuntimeInitialisation(t *testing.T) {
	l := NewDiplomacyLedger()
	if l.Credit[0] != 0 || l.Debt[0] != 0 {
		t.Fatal("第 0 格必須是無效哨兵")
	}
	for i := 1; i <= DiplomacyLedgerSlots; i++ {
		if l.Credit[i] != CreditMax || l.Debt[i] != 0 {
			t.Fatalf("槽位 %d：信用度／外債 = %d／%d，預期 100／0",
				i, l.Credit[i], l.Debt[i])
		}
	}
}

func TestDiplomacyLedgerSlotBounds(t *testing.T) {
	l := NewDiplomacyLedger()
	for _, slot := range []int{0, -1, DiplomacyLedgerSlots + 1} {
		if _, err := l.CreditFor(slot); err == nil {
			t.Errorf("槽位 %d 不應讀到信用度", slot)
		}
		if _, err := l.DebtFor(slot); err == nil {
			t.Errorf("槽位 %d 不應讀到外債", slot)
		}
	}
}

func TestDiplomacyLedgerLoanAddsDebtAndDeductsCredit(t *testing.T) {
	w, p := supplyWorld(t, 100, 100, 100, 100)
	l := NewDiplomacyLedger()
	res, err := l.RequestLoan(w, 1, 1, 1500, NewRand(1))
	if err != nil {
		t.Fatal(err)
	}
	if !res.Approved || res.Amount != 1500 {
		t.Fatalf("小額貸款應核准：%+v", res)
	}
	if l.Credit[1] != 97 || l.Debt[1] != 1500 {
		t.Fatalf("帳本未同步：信用度／外債 = %d／%d，預期 97／1500",
			l.Credit[1], l.Debt[1])
	}
	if p.Gold != 1600 {
		t.Fatalf("省份黃金 = %d，預期 1600", p.Gold)
	}
}

func TestDiplomacyLedgerRejectedLoanLeavesLedgerUntouched(t *testing.T) {
	w, p := supplyWorld(t, 100, 100, 100, 100)
	l := NewDiplomacyLedger()
	res, err := l.RequestLoan(w, 1, 1, 10000, NewRand(1))
	if err != nil {
		t.Fatal(err)
	}
	if res.Approved || l.Credit[1] != CreditMax || l.Debt[1] != 0 || p.Gold != 100 {
		t.Fatalf("被拒貸款不應改帳本／黃金：res=%+v credit=%d debt=%d gold=%d",
			res, l.Credit[1], l.Debt[1], p.Gold)
	}
}

func TestDiplomacyLedgerCreditZeroBlocksLoan(t *testing.T) {
	w, p := supplyWorld(t, 100, 100, 100, 100)
	l := NewDiplomacyLedger()
	l.Credit[1] = 0
	rng := NewRand(17)
	seed := rng.Seed()

	res, err := l.RequestLoan(w, 1, 1, 1500, rng)
	if err != nil {
		t.Fatal(err)
	}
	if !res.CreditBlocked || res.Approved || res.CommandCompleted {
		t.Fatalf("帳本應保留信用 gate：%+v", res)
	}
	if l.Credit[1] != 0 || l.Debt[1] != 0 || p.Gold != 100 {
		t.Fatalf("信用 gate 不應改帳本／黃金：credit=%d debt=%d gold=%d",
			l.Credit[1], l.Debt[1], p.Gold)
	}
	if rng.Seed() != seed {
		t.Fatalf("信用 gate 不應消耗亂數：before=%d after=%d", seed, rng.Seed())
	}
}

func TestDiplomacyLedgerRepayConsumesDebtAndRestoresCredit(t *testing.T) {
	w, p := supplyWorld(t, 2000, 100, 100, 100)
	l := NewDiplomacyLedger()
	l.Credit[1] = 47
	l.Debt[1] = 1500
	res, err := l.RepayDebt(w, 1, 1, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if res.Units != 2 || res.CreditAfter != 49 || res.DebtAfter != 500 {
		t.Fatalf("償還結果 = %+v，預期 units=2 credit=49 debt=500", res)
	}
	if p.Gold != 1000 {
		t.Fatalf("償還後黃金 = %d，預期 1000", p.Gold)
	}
}

func TestDiplomacyLedgerRepayCannotExceedDebt(t *testing.T) {
	w, p := supplyWorld(t, 5000, 100, 100, 100)
	l := NewDiplomacyLedger()
	l.Debt[1] = 500
	if _, err := l.RepayDebt(w, 1, 1, 501); err == nil {
		t.Fatal("超過外債餘額應拒絕")
	}
	if l.Debt[1] != 500 || p.Gold != 5000 {
		t.Fatalf("拒絕超額償還仍改變狀態：debt=%d gold=%d", l.Debt[1], p.Gold)
	}
}

func TestAidResourceBasesAreRemainingCapacity(t *testing.T) {
	p := &Province{Gold: 100, Food: ResourceCap, Ammo: 60000, Fuel: 61000}
	got := AidResourceBases(p)
	want := [4]int{59900, 0, 0, 0}
	if got != want {
		t.Fatalf("外援基準 = %v，預期 %v", got, want)
	}
}

func TestAIWorldAidResourceBasesChecksProvince(t *testing.T) {
	w, _ := supplyWorld(t, 100, 200, 300, 400)
	got, err := w.AidResourceBases(1)
	if err != nil {
		t.Fatal(err)
	}
	if got != [4]int{59900, 59800, 59700, 59600} {
		t.Fatalf("世界外援基準 = %v", got)
	}
	if _, err := w.AidResourceBases(0); err == nil {
		t.Fatal("省份 0 應拒絕")
	}
}
