package game

import "testing"

func TestCellManhattan(t *testing.T) {
	// 格編號 = row×14 + col。
	at := func(col, row int) CellIndex {
		c, err := CellAt(col, row)
		if err != nil {
			t.Fatalf("CellAt(%d,%d): %v", col, row, err)
		}
		return c
	}
	cases := []struct {
		a, b CellIndex
		want int
	}{
		{at(0, 0), at(0, 0), 0},
		{at(0, 0), at(3, 0), 3},
		{at(0, 0), at(0, 3), 3},
		{at(2, 5), at(6, 9), 8},
		{at(6, 9), at(2, 5), 8}, // 對稱
		{at(13, 13), at(0, 0), 26},
	}
	for _, c := range cases {
		if got := CellManhattan(c.a, c.b); got != c.want {
			t.Errorf("CellManhattan(%d, %d) = %d，要 %d", c.a, c.b, got, c.want)
		}
	}
}

// rankFixture 造一個小場面：起點在格 0，三個城市分別在距離 2、5、8 處。
func rankFixture(t *testing.T) (cities []CellIndex, occ *Occupancy,
	units map[GeneralID]*CombatUnit, lookup func(GeneralID) *CombatUnit) {
	t.Helper()
	at := func(col, row int) CellIndex {
		c, err := CellAt(col, row)
		if err != nil {
			t.Fatalf("CellAt: %v", err)
		}
		return c
	}
	cities = []CellIndex{at(2, 0), at(5, 0), at(8, 0)} // 距離 2 / 5 / 8
	occ = &Occupancy{}
	units = map[GeneralID]*CombatUnit{}
	lookup = func(id GeneralID) *CombatUnit { return units[id] }
	return
}

func TestRankCityTargetsPicksNearestDefended(t *testing.T) {
	cities, occ, units, lookup := rankFixture(t)
	// 三個城市都被守方佔著。
	for i, c := range cities {
		id := GeneralID(i + 1)
		occ[c] = id
		units[id] = &CombatUnit{General: id, Cell: c, Attacking: false}
	}
	scores := RankCityTargets(cities, occ, 0, lookup)
	if got := FirstPositiveCity(scores); got != cities[0] {
		t.Errorf("該挑最近的城市 %d，實際 %d", cities[0], got)
	}
	// 升序：2 < 5 < 8。
	want := []int32{2, 5, 8}
	for i, w := range want {
		if scores[i].Score != w {
			t.Errorf("排序後第 %d 名分數 %d，要 %d", i, scores[i].Score, w)
		}
	}
}

func TestRankCityTargetsIgnoresEmptyAndAttackerHeld(t *testing.T) {
	cities, occ, units, lookup := rankFixture(t)
	// 最近的空著、中間的被攻方佔著、最遠的才有守方。
	occ[cities[1]] = 1
	units[1] = &CombatUnit{General: 1, Cell: cities[1], Attacking: true,
		NextCell: cities[1], Command: BattleCmdStandby} // 不符剔除條件
	occ[cities[2]] = 2
	units[2] = &CombatUnit{General: 2, Cell: cities[2], Attacking: false}

	scores := RankCityTargets(cities, occ, 0, lookup)
	if got := FirstPositiveCity(scores); got != cities[2] {
		t.Errorf("空格與攻方佔的城市都是 0 分，該挑 %d，實際 %d", cities[2], got)
	}
}

func TestRankCityTargetsDropsCityAlreadyTargeted(t *testing.T) {
	// 命令 5 + 下一跳正好是這個城市 → 整格剔除。
	cities, occ, units, lookup := rankFixture(t)
	occ[cities[0]] = 1
	units[1] = &CombatUnit{General: 1, Cell: cities[0], Attacking: true,
		NextCell: cities[0], Command: BattleCmdUnknown5}
	scores := RankCityTargets(cities, occ, 0, lookup)
	if scores[0].Cell != NoCell {
		t.Errorf("命令 5 且下一跳就是這格，該被剔除，實際 Cell=%d", scores[0].Cell)
	}

	// 換成別的命令就不剔除。
	units[1].Command = BattleCmdSeekTarget
	scores = RankCityTargets(cities, occ, 0, lookup)
	if scores[0].Cell == NoCell {
		t.Error("命令不是 5 就不該剔除——這是命令 5 唯一的比較點")
	}
}

func TestRankCityTargetsDropsCityWhenNoNextCell(t *testing.T) {
	// 第二條剔除條件：攻方站著但沒有下一跳。
	cities, occ, units, lookup := rankFixture(t)
	occ[cities[1]] = 1
	units[1] = &CombatUnit{General: 1, Cell: cities[1], Attacking: true,
		NextCell: NoCell, Command: BattleCmdStandby}
	scores := RankCityTargets(cities, occ, 0, lookup)
	for _, s := range scores {
		if s.Cell == cities[1] {
			t.Errorf("沒有下一跳的攻方佔位，該剔除 %d", cities[1])
		}
	}
}

func TestFirstPositiveCitySkipsZeroDistance(t *testing.T) {
	// 距離 0 = 自己就站在那格，被 `> 0` 濾掉。
	at, err := CellAt(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	far, err := CellAt(4, 0)
	if err != nil {
		t.Fatal(err)
	}
	scores := []CityScore{{Cell: at, Score: 0}, {Cell: far, Score: 4}}
	if got := FirstPositiveCity(scores); got != far {
		t.Errorf("距離 0 該被跳過，要 %d，實際 %d", far, got)
	}
	if got := FirstPositiveCity(nil); got != NoCell {
		t.Errorf("沒有城市時要回 NoCell，實際 %d", got)
	}
}

func TestWithinTwoSteps(t *testing.T) {
	mid, err := CellAt(6, 6)
	if err != nil {
		t.Fatal(err)
	}
	// ⭐ 同一格**成立**：原版的兩層掃描會走出去再走回來
	// （n1 = b 的鄰格，n2 = n1 的鄰格，其中一個就是 b 本身）。
	// 第一版測試斷言成 false，是斷言寫錯不是實作錯。
	if !WithinTwoSteps(mid, mid) {
		t.Error("同一格該成立——走出去再走回來，原版也是這樣")
	}
	ns := mid.Neighbours()
	if len(ns) == 0 {
		t.Fatal("中央格應該有六個鄰格")
	}
	for _, n := range ns {
		if !WithinTwoSteps(n, mid) {
			t.Errorf("一步的鄰格 %d 該成立", n)
		}
		for _, n2 := range n.Neighbours() {
			if n2 == mid {
				continue
			}
			if !WithinTwoSteps(n2, mid) {
				t.Errorf("兩步的格 %d 該成立", n2)
			}
		}
	}
	// 隔很遠的不成立。
	far, err := CellAt(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if WithinTwoSteps(far, mid) {
		t.Errorf("格 %d 離 %d 超過兩步，不該成立", far, mid)
	}
	if WithinTwoSteps(NoCell, mid) || WithinTwoSteps(mid, NoCell) {
		t.Error("NoCell 一律不成立")
	}
}
