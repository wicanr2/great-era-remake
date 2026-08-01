package game

import "testing"

// 外援成功率 70%。
func TestAidSuccessRate(t *testing.T) {
	rng := NewRand(555)
	const n = 20000
	ok := 0
	for i := 0; i < n; i++ {
		if rng.Int(AidRollRange) <= AidSuccessMax {
			ok++
		}
	}
	if rate := float64(ok) / n; rate < 0.67 || rate > 0.73 {
		t.Errorf("外援成功率 %.3f，原版是 7/10", rate)
	}
}

// 援助國決定除數：美國(99)／142 最慷慨，英國(100) 次之，其餘 5。
func TestAidDivisor(t *testing.T) {
	cases := map[int]int{99: 1, 142: 1, 100: 3, 101: 5, 141: 5, 0: 5}
	for donor, want := range cases {
		if got := AidDivisor(donor); got != want {
			t.Errorf("援助國 %d 的除數是 %d，預期 %d", donor, got, want)
		}
	}
}

// 張作霖在民國 17 年 2–6 月一律拿不到援助，但其他時候可以。
func TestAidHistoricalEmbargo(t *testing.T) {
	setup := func(t *testing.T) (*AIWorld, *Province) {
		w := realWorld(t)
		prov, err := w.Table.At(1)
		if err != nil {
			t.Fatal(err)
		}
		prov.Commander = AidEmbargoLeader
		prov.Gold = 0
		return w, prov
	}
	base := [4]int{1000, 1000, 1000, 1000}

	for _, month := range []uint8{2, 4, 6} {
		found := false
		for seed := uint32(1); seed < 30; seed++ {
			w, prov := setup(t)
			st := GameState{Year: AidEmbargoYear, Month: month}
			res, err := w.RequestAid(1, st, 99, base, NewRand(seed))
			if err != nil {
				t.Fatal(err)
			}
			if res.Roll > AidSuccessMax {
				continue // 這顆種子本來就被拒，不算數
			}
			found = true
			if !res.Embargoed || res.Approved {
				t.Errorf("民國 17 年 %d 月：張作霖應被禁運，卻 approved=%v",
					month, res.Approved)
			}
			if prov.Gold != 0 {
				t.Errorf("被禁運卻拿到 %d 黃金", prov.Gold)
			}
			break
		}
		if !found {
			t.Errorf("%d 月：30 顆種子都擲出拒絕，測不到禁運", month)
		}
	}

	// 17 年 1 月與 7 月不在禁運區間。
	for _, month := range []uint8{1, 7} {
		for seed := uint32(1); seed < 30; seed++ {
			w, _ := setup(t)
			st := GameState{Year: AidEmbargoYear, Month: month}
			res, err := w.RequestAid(1, st, 99, base, NewRand(seed))
			if err != nil {
				t.Fatal(err)
			}
			if res.Roll > AidSuccessMax {
				continue
			}
			if res.Embargoed {
				t.Errorf("%d 月不該禁運", month)
			}
			break
		}
	}

	// 別的年份不受影響。
	for seed := uint32(1); seed < 30; seed++ {
		w, _ := setup(t)
		st := GameState{Year: 18, Month: 4}
		res, err := w.RequestAid(1, st, 99, base, NewRand(seed))
		if err != nil {
			t.Fatal(err)
		}
		if res.Roll > AidSuccessMax {
			continue
		}
		if res.Embargoed {
			t.Error("民國 18 年不該禁運")
		}
		break
	}
}

// 黃金夾在 6000，其他三種不夾。
func TestAidGoldCap(t *testing.T) {
	for seed := uint32(1); seed < 40; seed++ {
		w := realWorld(t)
		prov, _ := w.Table.At(1)
		prov.Commander = 1
		prov.Gold, prov.Food = 0, 0

		st := GameState{Year: 15, Month: 8}
		// 基準開很大，除數 1，逼出上限。
		res, err := w.RequestAid(1, st, AidDonorGenerous,
			[4]int{60000, 60000, 0, 0}, NewRand(seed))
		if err != nil {
			t.Fatal(err)
		}
		if !res.Approved {
			continue
		}
		if res.Gold > AidGoldCap {
			t.Errorf("黃金 %d 超過上限 %d", res.Gold, AidGoldCap)
		}
		// 糧食沒有上限檢查，可能超過 6000。
		return
	}
	t.Skip("40 顆種子都被拒，跳過")
}

// 償還外債與貸款對稱：每 500 金 +1 信用度，夾 100。
func TestRepayDebtSymmetry(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	prov.Gold = 5000

	units, credit, err := w.RepayDebt(1, 1500, 50)
	if err != nil {
		t.Fatal(err)
	}
	if units != 3 || credit != 53 {
		t.Errorf("還 1500：units %d、信用度 %d，預期 3／53", units, credit)
	}
	if prov.Gold != 3500 {
		t.Errorf("黃金 %d，預期 3500", prov.Gold)
	}

	// 借同樣的錢會扣回來——兩邊用同一個 LoanUnit。
	w2 := realWorld(t)
	p2, _ := w2.Table.At(1)
	p2.Gold = 0
	res, back, err := w2.RequestLoan(1, 1500, 53, NewRand(1))
	if err != nil {
		t.Fatal(err)
	}
	if res.Approved && back != 50 {
		t.Errorf("借 1500 之後信用度 %d，預期回到 50", back)
	}
}

// 信用度加到上限就停。
func TestRepayDebtCreditCap(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	prov.Gold = 60000

	_, credit, err := w.RepayDebt(1, 50000, 90)
	if err != nil {
		t.Fatal(err)
	}
	if credit != CreditMax {
		t.Errorf("信用度 %d，應該夾到 %d", credit, CreditMax)
	}
}

// 黃金不足時擋下來且不扣錢。
func TestRepayDebtInsufficient(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	prov.Gold = 100
	if _, _, err := w.RepayDebt(1, 500, 50); err == nil {
		t.Error("黃金 100 還 500 應該報錯")
	}
	if prov.Gold != 100 {
		t.Error("報錯後不該扣錢")
	}
}
