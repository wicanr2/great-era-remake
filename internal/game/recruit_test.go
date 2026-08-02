package game

import "testing"

// 徵兵的成本與上限來自實機（`docs/playtest/10` §2）：
//
//	湖北 黃金 4,150 → 上限 (0-41,500)
//	河南 黃金 5,950 → 上限 (0-59,500)
//	輸入 1,000 兵   → 「共須黃金 100」
func TestRecruitAffordableLimitMatchesOriginal(t *testing.T) {
	cases := []struct {
		branch uint8
		gold   int
		limit  int
	}{
		{BranchInfantry, 4150, 41500},
		{BranchInfantry, 5950, 59500},
		{BranchCavalry, 100, 500},
		{BranchArtiller, 100, 100},
		{BranchArmour, 100, 10},
	}
	for _, c := range cases {
		if got := RecruitAffordableLimit(c.branch, c.gold); got != c.limit {
			t.Errorf("兵種 %d、黃金 %d 的負擔上限算出 %d，原版是 %d",
				c.branch, c.gold, got, c.limit)
		}
	}
	costs := []struct {
		branch  uint8
		n, cost int
	}{
		{BranchInfantry, 1000, 100},
		{BranchCavalry, 1000, 200},
		{BranchArtiller, 1000, 1000},
		{BranchArmour, 100, 1000},
	}
	for _, c := range costs {
		if got := RecruitCost(c.branch, c.n); got != c.cost {
			t.Errorf("兵種 %d 徵 %d 人成本算出 %d，原版是 %d",
				c.branch, c.n, got, c.cost)
		}
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
	t.Logf("湖北補了 %d 人，花 %d 金", added, RecruitCost(BranchInfantry, added))
}

// 超過上限要擋下來，不能默默截斷。
func TestRecruitRejectsOverLimit(t *testing.T) {
	w := realWorld(t)
	p := ProvinceID(26)
	prov, _ := w.Table.At(p)
	prov.Gold = 10 // 上限 100
	limit := w.RecruitLimit(p, BranchInfantry)
	if _, err := w.Recruit(p, BranchInfantry, limit+1); err == nil {
		t.Error("徵 101 人（上限 100）應該被擋下")
	}
	if _, err := w.Recruit(p, BranchInfantry, limit); err != nil {
		t.Errorf("徵 100 人（正好上限）不該失敗：%v", err)
	}
}
