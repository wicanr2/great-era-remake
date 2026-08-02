package game

import "testing"

func fixedBudget(n int) *CommandBudget {
	return NewCommandBudgetFunc(func(ProvinceID) int { return n })
}

// 消耗到零就擋下來。
func TestCommandBudgetSpend(t *testing.T) {
	b := fixedBudget(2)
	if got := b.Remaining(1); got != 2 {
		t.Fatalf("剩餘 %d，預期 2", got)
	}
	if !b.Spend(1) || !b.Spend(1) {
		t.Fatal("有 2 個應該扣得動兩次")
	}
	if b.Spend(1) {
		t.Error("歸零之後不該扣得動")
	}
	if got := b.Remaining(1); got != 0 {
		t.Errorf("用完之後剩 %d，預期 0", got)
	}
}

// 各省的帳互不影響。
func TestCommandBudgetPerProvince(t *testing.T) {
	b := fixedBudget(1)
	if !b.Spend(1) {
		t.Fatal("省 1 應該扣得動")
	}
	if got := b.Remaining(2); got != 1 {
		t.Errorf("省 2 剩 %d，不該被省 1 影響", got)
	}
}

// 換月重新算。
func TestCommandBudgetReset(t *testing.T) {
	b := fixedBudget(3)
	b.Spend(1)
	b.Spend(1)
	if got := b.Remaining(1); got != 1 {
		t.Fatalf("扣兩次之後剩 %d", got)
	}
	b.Reset()
	if got := b.Remaining(1); got != 3 {
		t.Errorf("換月之後剩 %d，應該重新算成 3", got)
	}
}

func TestCommandBudgetAutonomousProvinceStopsAfterOneCommand(t *testing.T) {
	w := realWorld(t)
	p, _ := w.Table.At(1)
	p.Flags |= ProvinceFlagAutonomous
	b := NewCommandBudget(w)
	if before := b.Remaining(1); before <= 0 {
		t.Fatalf("自治省開始時應至少有一個指令，得到 %d", before)
	}
	if !b.Spend(1) {
		t.Fatal("自治省的第一個指令應可執行")
	}
	if got := b.Remaining(1); got != 0 {
		t.Fatalf("自治省執行一項後應歸零，得到 %d", got)
	}
}

// 接上真實世界時，上限就是「將領數 ÷ 8 + 1」。
func TestCommandBudgetUsesRealFormula(t *testing.T) {
	w := realWorld(t)
	b := NewCommandBudget(w)
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		if got, want := b.Remaining(p), w.CommandsFor(p); got != want {
			t.Fatalf("省 %d 的指令數 %d，公式算出來是 %d", p, got, want)
		}
	}
}
