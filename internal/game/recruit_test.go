package game

import "testing"

// 徵兵的成本與上限來自實機（`docs/playtest/10` §2）：
//
//	湖北 黃金 4,150 → 上限 (0-41,500)
//	河南 黃金 5,950 → 上限 (0-59,500)
//	輸入 1,000 兵   → 「共須黃金 100」
func TestRecruitLimitMatchesScreen(t *testing.T) {
	w := realWorld(t)
	cases := []struct {
		prov  ProvinceID
		gold  uint16
		limit int
	}{
		{26, 4150, 41500}, // 湖北
		{19, 5950, 59500}, // 河南
	}
	for _, c := range cases {
		prov, err := w.Table.At(c.prov)
		if err != nil {
			t.Fatal(err)
		}
		// 存檔裡的黃金與實機那一刻不同（實機跑了一個月），
		// 所以直接把畫面上的值塞進去再算。
		prov.Gold = c.gold
		if got := w.RecruitLimit(c.prov); got != c.limit {
			t.Errorf("省 %d 黃金 %d 的徵兵上限算出 %d，實機是 %d",
				c.prov, c.gold, got, c.limit)
		}
	}
	if got := RecruitCost(1000); got != 100 {
		t.Errorf("徵 1000 兵要 %d 金，實機顯示 100", got)
	}
}

// 徵兵補到滿員為止，而且只補同兵種。
func TestRecruitFillsToBranchCap(t *testing.T) {
	w := realWorld(t)
	p := ProvinceID(26)
	prov, _ := w.Table.At(p)
	prov.Gold = 60000 // 給足黃金，讓上限不擋路

	before := map[int]uint16{}
	for _, i := range w.RosterOf(p).ids {
		before[i] = w.Strengths[i].Force
	}
	added := w.RecruitToFull(p, BranchInfantry)
	if added == 0 {
		t.Fatal("一個兵都沒徵到")
	}
	full := BranchFullStrength(BranchInfantry)
	for _, i := range w.RosterOf(p).ids {
		s := &w.Strengths[i]
		if s.Branch != BranchInfantry {
			if s.Force != before[i] {
				t.Errorf("兵種 %d 的將領被補了兵", s.Branch)
			}
			continue
		}
		if s.Force > full {
			t.Errorf("將領 %d 補過頭：%d > 滿員 %d", i, s.Force, full)
		}
	}
	t.Logf("湖北補了 %d 人，花 %d 金", added, RecruitCost(added))
}

// 超過上限要擋下來，不能默默截斷。
func TestRecruitRejectsOverLimit(t *testing.T) {
	w := realWorld(t)
	p := ProvinceID(26)
	prov, _ := w.Table.At(p)
	prov.Gold = 10 // 上限 100
	if _, err := w.Recruit(p, BranchInfantry, 101); err == nil {
		t.Error("徵 101 人（上限 100）應該被擋下")
	}
	if _, err := w.Recruit(p, BranchInfantry, 100); err != nil {
		t.Errorf("徵 100 人（正好上限）不該失敗：%v", err)
	}
}
