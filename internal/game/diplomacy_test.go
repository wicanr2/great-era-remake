package game

import "testing"

// 外援核准率 30%：0..6 會走原版拒絕文字，7..9 才進入資源寫入。
func TestAidApprovalRate(t *testing.T) {
	rng := NewRand(555)
	const n = 20000
	ok := 0
	for i := 0; i < n; i++ {
		if rng.Int(AidRollRange) >= AidApprovalMin {
			ok++
		}
	}
	if rate := float64(ok) / n; rate < 0.27 || rate > 0.33 {
		t.Errorf("外援核准率 %.3f，原版是 3/10", rate)
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

func TestAidDonorCode(t *testing.T) {
	for _, tc := range []struct {
		stage, faction, roll int
		want                 int
	}{
		{1, 10, 0, 100}, {1, 10, 1, 99}, {1, 10, 6, 100}, {1, 10, 7, 99},
		{1, 1, 0, 142}, {1, 5, 9, 100},
		{1, 4, 5, 99}, {1, 4, 6, 100}, {1, 4, 7, 146},
		{1, 3, 0, 101}, {1, 2, 9, 101},
		{1, 0, 5, 5}, {1, 0, 6, 99}, {1, 0, 7, 100},
		{1, 0, 8, 101}, {1, 0, 9, 141},
		{2, 1, 0, 99}, {3, 10, 9, 99},
	} {
		if got := AidDonorCode(uint8(tc.stage), tc.faction, tc.roll); got != tc.want {
			t.Errorf("stage=%d faction=%d roll=%d：代碼 %d，預期 %d",
				tc.stage, tc.faction, tc.roll, got, tc.want)
		}
	}
}

func TestRequestAidForFactionUsesSameRoll(t *testing.T) {
	findSeed := func(want int) uint32 {
		for seed := uint32(1); seed < 1000; seed++ {
			if NewRand(seed).Int(AidRollRange) == want {
				return seed
			}
		}
		t.Fatalf("找不到 roll=%d 的固定種子", want)
		return 0
	}

	w := realWorld(t)
	prov, err := w.Table.At(1)
	if err != nil {
		t.Fatal(err)
	}
	prov.Commander = 1 // 避免命中日期／司令特殊核准
	base := [4]int{}
	approved, err := w.RequestAidForFaction(1, GameState{Stage: 1, Year: 15, Month: 8}, 0,
		base, NewRand(findSeed(7)))
	if err != nil {
		t.Fatal(err)
	}
	if approved.Roll != 7 || !approved.Approved || approved.Donor != 100 || approved.Divisor != 3 {
		t.Fatalf("roll=7 的 default 第一期待援助代碼錯誤：%+v", approved)
	}

	refused, err := w.RequestAidForFaction(1, GameState{Stage: 1, Year: 15, Month: 8}, 0,
		base, NewRand(findSeed(6)))
	if err != nil {
		t.Fatal(err)
	}
	if refused.Roll != 6 || refused.Approved || refused.Donor != 99 || refused.Divisor != 1 {
		t.Fatalf("roll=6 的 default 第一期待援助代碼錯誤：%+v", refused)
	}
}

// 張作霖在民國 17 年 2–6 月命中特殊核准分支；其他日期仍依原版拒絕骰。
func TestAidSpecialOverride(t *testing.T) {
	setup := func(t *testing.T) (*AIWorld, *Province) {
		w := realWorld(t)
		prov, err := w.Table.At(1)
		if err != nil {
			t.Fatal(err)
		}
		prov.Commander = AidSpecialLeader
		prov.Gold = 0
		return w, prov
	}
	base := [4]int{1000, 1000, 1000, 1000}

	for _, month := range []uint8{2, 4, 6} {
		found := false
		for seed := uint32(1); seed < 30; seed++ {
			w, _ := setup(t)
			st := GameState{Year: AidSpecialYear, Month: month}
			res, err := w.RequestAid(1, st, 99, base, NewRand(seed))
			if err != nil {
				t.Fatal(err)
			}
			if res.Roll > AidRefusalMax {
				continue // 這顆種子本來就會核准，不算特殊分支
			}
			found = true
			if !res.SpecialOverride || !res.Approved {
				t.Errorf("民國 17 年 %d 月：張作霖特殊分支應核准，override=%v approved=%v",
					month, res.SpecialOverride, res.Approved)
			}
			if !res.CommandCompleted {
				t.Errorf("民國 17 年 %d 月：特殊分支仍應完成援助指令", month)
			}
			break
		}
		if !found {
			t.Errorf("%d 月：30 顆種子都未落在拒絕骰，測不到特殊分支", month)
		}
	}

	// 17 年 1 月與 7 月不在特殊日期區間。
	for _, month := range []uint8{1, 7} {
		for seed := uint32(1); seed < 30; seed++ {
			w, _ := setup(t)
			st := GameState{Year: AidSpecialYear, Month: month}
			res, err := w.RequestAid(1, st, 99, base, NewRand(seed))
			if err != nil {
				t.Fatal(err)
			}
			if res.Roll > AidRefusalMax {
				continue
			}
			if res.SpecialOverride || res.Approved {
				t.Errorf("%d 月不該套用特殊核准，override=%v approved=%v",
					month, res.SpecialOverride, res.Approved)
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
		if res.Roll > AidRefusalMax {
			continue
		}
		if res.SpecialOverride || res.Approved {
			t.Error("民國 18 年不該套用特殊核准")
		}
		break
	}
}

// 原版在第一次援助亂數前就立完成旗標；隨機拒絕與特殊核准都會消耗指令，
// 但進入規則函式前的驗證錯誤不會產生 AidResult。
func TestAidCommandCompletedOnRefusal(t *testing.T) {
	foundRefusal := false
	for seed := uint32(1); seed < 40; seed++ {
		w := realWorld(t)
		prov, err := w.Table.At(1)
		if err != nil {
			t.Fatal(err)
		}
		prov.Commander = 1
		beforeResources := [4]uint16{prov.Gold, prov.Food, prov.Ammo, prov.Fuel}
		res, err := w.RequestAid(1, GameState{Year: 15, Month: 8}, AidDonorGenerous,
			[4]int{}, NewRand(seed))
		if err != nil {
			t.Fatal(err)
		}
		if !res.CommandCompleted {
			t.Errorf("種子 %d：援助結果未標記指令完成", seed)
		}
		if res.Roll <= AidRefusalMax {
			foundRefusal = true
			if res.Approved || res.SpecialOverride {
				t.Errorf("種子 %d：隨機拒絕卻 approved=%v override=%v", seed,
					res.Approved, res.SpecialOverride)
			}
			afterResources := [4]uint16{prov.Gold, prov.Food, prov.Ammo, prov.Fuel}
			if afterResources != beforeResources {
				t.Errorf("種子 %d：援助拒絕不應改動省份", seed)
			}
			break
		}
	}
	if !foundRefusal {
		t.Fatal("40 顆固定種子都未找到援助隨機拒絕")
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
