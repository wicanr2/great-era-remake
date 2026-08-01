package game

import "testing"

// 月份表：2/4/5/9/10 月沒有內政。
func TestAIInteriorMonths(t *testing.T) {
	for _, m := range []uint8{2, 4, 5, 9, 10} {
		if AIInteriorMonths[m] {
			t.Errorf("%d 月不該有內政（docs/re/32 §1）", m)
		}
	}
	for _, m := range []uint8{1, 3, 6, 7, 8, 11, 12} {
		if !AIInteriorMonths[m] {
			t.Errorf("%d 月應該有內政", m)
		}
	}
	// 徵兵只在 1/3/8/11。
	for _, m := range []uint8{1, 3, 8, 11} {
		if _, ok := AIRecruitMonths[m]; !ok {
			t.Errorf("%d 月應該徵兵", m)
		}
	}
	for _, m := range []uint8{6, 7, 12} {
		if _, ok := AIRecruitMonths[m]; ok {
			t.Errorf("%d 月不該徵兵", m)
		}
	}
	// 3 月與 11 月有黃金門檻，1 月與 8 月無條件。
	if AIRecruitMonths[3] != 5000 || AIRecruitMonths[11] != 3000 {
		t.Error("3 月門檻應是 5,000、11 月是 3,000")
	}
	if AIRecruitMonths[1] != 0 || AIRecruitMonths[8] != 0 {
		t.Error("1 月與 8 月是無條件徵兵")
	}
}

// 預算公式：min(黃金 × 2/3, 缺額 ÷ 10)，夾難度上限。
func TestAIRecruitBudget(t *testing.T) {
	// 缺額大 → 受黃金的 2/3 限制。
	if got := AIRecruitBudget(3000, 1_000_000, 0); got != 2000 {
		t.Errorf("黃金 3,000 的 2/3 = %d，預期 2000", got)
	}
	// 黃金多 → 受缺額限制。
	if got := AIRecruitBudget(1_000_000, 20000, 0); got != 2000 {
		t.Errorf("缺額 20,000 ÷ 10 = %d，預期 2000", got)
	}
	// 難度上限：bit 4 → 20,000。
	if got := AIRecruitBudget(1_000_000, 10_000_000, 0x10); got != AIRecruitBudgetHard {
		t.Errorf("bit4 的上限 %d，預期 %d", got, AIRecruitBudgetHard)
	}
	// bit 5 → 40,000。
	if got := AIRecruitBudget(1_000_000, 10_000_000, 0x20); got != AIRecruitBudgetEasy {
		t.Errorf("bit5 的上限 %d，預期 %d", got, AIRecruitBudgetEasy)
	}
	// 都沒設 → 只受絕對上限 100,000。
	if got := AIRecruitBudget(10_000_000, 100_000_000, 0); got != AIRecruitBudgetCap {
		t.Errorf("絕對上限 %d，預期 %d", got, AIRecruitBudgetCap)
	}
}

// 在真實存檔上跑一次內政，驗證它真的會補兵。
func TestRunInteriorOnRealSave(t *testing.T) {
	w := realWorld(t)
	gens, err := ParseGenerals(readGame(t, "MAN(1).DAT"), 274)
	if err != nil {
		t.Fatal(err)
	}

	var recruited, gold, provinces int
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		prov, err := w.Table.At(p)
		if err != nil || prov.Commander == 0 {
			continue
		}
		r := w.RunInterior(p, 1, gens, 0) // 1 月：無條件徵兵
		if r.Recruited > 0 {
			provinces++
			recruited += r.Recruited
			gold += r.GoldSpent
		}
	}
	t.Logf("1 月：%d 省徵兵，共補 %d 兵、花 %d 金", provinces, recruited, gold)
	if provinces == 0 {
		t.Error("39 省跑一次 1 月內政，一個省都沒徵到兵")
	}
	// 匯率自洽：步兵 10 兵 = 1 金，所以補的兵不會少於花的金。
	if recruited < gold {
		t.Errorf("補了 %d 兵卻花 %d 金——最貴的砲兵也是 5 金 1 兵，不該倒過來",
			recruited, gold)
	}
}

// 2 月不徵兵。
func TestRunInteriorSkipsQuietMonths(t *testing.T) {
	w := realWorld(t)
	gens, err := ParseGenerals(readGame(t, "MAN(1).DAT"), 274)
	if err != nil {
		t.Fatal(err)
	}
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		if prov, err := w.Table.At(p); err != nil || prov.Commander == 0 {
			continue
		}
		if r := w.RunInterior(p, 2, gens, 0); r.Recruited != 0 || r.LoyaltyGained {
			t.Fatalf("省 %d 在 2 月做了內政：%+v", p, r)
		}
	}
}

// 6 月補人民忠誠度，夾 100。
func TestRunInteriorLoyalty(t *testing.T) {
	w := realWorld(t)
	gens, err := ParseGenerals(readGame(t, "MAN(1).DAT"), 274)
	if err != nil {
		t.Fatal(err)
	}
	prov, err := w.Table.At(1)
	if err != nil {
		t.Fatal(err)
	}
	prov.Commander = 7
	prov.Loyalty = 50
	if r := w.RunInterior(1, AIComfortMonth, gens, 0); !r.LoyaltyGained {
		t.Error("6 月應該補忠誠度")
	}
	if prov.Loyalty != 70 {
		t.Errorf("忠誠度 50 + 20 = %d，預期 70", prov.Loyalty)
	}
	prov.Loyalty = 95
	w.RunInterior(1, AIComfortMonth, gens, 0)
	if prov.Loyalty != 100 {
		t.Errorf("95 + 20 應該夾到 100，實際 %d", prov.Loyalty)
	}
}
