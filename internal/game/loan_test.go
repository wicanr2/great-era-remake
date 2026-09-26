package game

import "testing"

// 核貸門檻的三個區間：必過、機率、必拒。
func TestLoanThresholds(t *testing.T) {
	cases := []struct {
		amount        int
		always, never bool
		name          string
	}{
		{500, true, false, "借 500"},
		{1999, true, false, "借 1999（units=3，9+3=12 剛好不超過）"},
		{2000, false, false, "借 2000（units=4，骰 9 就爆）"},
		{6000, false, false, "借 6000（units=12，只有骰 0 才過）"},
		{6500, false, true, "借 6500（units=13，必拒）"},
		{20000, false, true, "借 20000"},
	}
	for _, c := range cases {
		if got := LoanAlwaysApproved(c.amount); got != c.always {
			t.Errorf("%s：必過 = %v，預期 %v", c.name, got, c.always)
		}
		if got := LoanAlwaysRejected(c.amount); got != c.never {
			t.Errorf("%s：必拒 = %v，預期 %v", c.name, got, c.never)
		}
	}
}

// 小額貸款：不論種子都會過，黃金入帳、信用度扣 units。
func TestLoanSmallAlwaysApproved(t *testing.T) {
	for seed := uint32(1); seed <= 20; seed++ {
		w := realWorld(t)
		prov, _ := w.Table.At(1)
		prov.Gold = 0

		res, credit, err := w.RequestLoan(1, 1500, 50, NewRand(seed))
		if err != nil {
			t.Fatal(err)
		}
		if !res.Approved {
			t.Fatalf("種子 %d：借 1500 被拒（骰 %d + units %d）",
				seed, res.Roll, res.Units)
		}
		if !res.CommandCompleted {
			t.Fatalf("種子 %d：核准貸款應標記指令已完成", seed)
		}
		if prov.Gold != 1500 {
			t.Errorf("種子 %d：黃金 %d，預期 1500", seed, prov.Gold)
		}
		if res.Units != 3 || res.CreditCost != 3 {
			t.Errorf("種子 %d：units %d、扣信用 %d，預期 3／3",
				seed, res.Units, res.CreditCost)
		}
		if credit != 47 {
			t.Errorf("種子 %d：信用度 %d，預期 47", seed, credit)
		}
	}
}

// 大額貸款：不論種子都被拒，而且什麼都不動。
func TestLoanLargeAlwaysRejected(t *testing.T) {
	for seed := uint32(1); seed <= 20; seed++ {
		w := realWorld(t)
		prov, _ := w.Table.At(1)
		prov.Gold = 100

		res, credit, err := w.RequestLoan(1, 10000, 50, NewRand(seed))
		if err != nil {
			t.Fatal(err)
		}
		if res.Approved {
			t.Fatalf("種子 %d：借 10000 竟然過了", seed)
		}
		if !res.CommandCompleted {
			t.Fatalf("種子 %d：隨機拒絕仍應標記指令已完成", seed)
		}
		if prov.Gold != 100 || credit != 50 {
			t.Errorf("種子 %d：被拒卻動到黃金(%d)或信用度(%d)",
				seed, prov.Gold, credit)
		}
	}
}

// 借 6000（units=12）的核准率應該接近 1/10——只有骰到 0 才過。
func TestLoanBorderlineRate(t *testing.T) {
	approved := 0
	const n = 2000
	rng := NewRand(4242)
	for i := 0; i < n; i++ {
		w := realWorld(t)
		res, _, err := w.RequestLoan(1, 6000, 100, rng)
		if err != nil {
			t.Fatal(err)
		}
		if res.Approved {
			approved++
		}
	}
	rate := float64(approved) / n
	if rate < 0.06 || rate > 0.14 {
		t.Errorf("借 6000 的核准率 %.3f，預期接近 0.1", rate)
	}
}

// 信用度不夠扣時歸零，不會下溢。
func TestLoanCreditFloor(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	prov.Gold = 0

	res, credit, err := w.RequestLoan(1, 1500, 2, NewRand(7))
	if err != nil {
		t.Fatal(err)
	}
	if !res.Approved {
		t.Skip("這顆種子被拒，跳過")
	}
	if credit != 0 {
		t.Errorf("信用度 2 扣 3 之後是 %d，應該夾到 0", credit)
	}
}

func TestLoanCreditZeroBlocksBeforeRandomAndMutation(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	prov.Gold = 123
	rng := NewRand(4242)
	seed := rng.Seed()

	res, credit, err := w.RequestLoan(1, 1500, 0, rng)
	if err != nil {
		t.Fatal(err)
	}
	if !res.CreditBlocked || res.Approved || res.CommandCompleted || res.Roll != 0 {
		t.Fatalf("信用度為 0 應走早期 gate：%+v", res)
	}
	if credit != 0 || prov.Gold != 123 {
		t.Fatalf("信用 gate 不應改信用／黃金：credit=%d gold=%d", credit, prov.Gold)
	}
	if rng.Seed() != seed {
		t.Fatalf("信用 gate 不應消耗貸款亂數：before=%d after=%d", seed, rng.Seed())
	}
}

// 黃金入帳夾在資源上限內。
func TestLoanGoldCap(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	prov.Gold = ResourceCap - 100

	res, _, err := w.RequestLoan(1, 1000, 50, NewRand(3))
	if err != nil {
		t.Fatal(err)
	}
	if !res.Approved {
		t.Skip("這顆種子被拒，跳過")
	}
	if prov.Gold != ResourceCap || res.Amount != 100 {
		t.Errorf("黃金 %d、入帳 %d，預期 %d／100", prov.Gold, res.Amount, ResourceCap)
	}
}

func TestLoanRejectsNonPositive(t *testing.T) {
	w := realWorld(t)
	if _, _, err := w.RequestLoan(1, 0, 50, NewRand(1)); err == nil {
		t.Error("借 0 應該報錯")
	}
}
