package game

import "testing"

// 四個兵種的匯率，以及「徵滿一個部隊要多少金」。
func TestAIRecruitRates(t *testing.T) {
	cases := []struct {
		branch        uint8
		name          string
		goldForFull   int
		soldiersPer1k int // 1,000 金能徵幾個兵
	}{
		{BranchInfantry, "步兵", 2000, 10000},
		{BranchCavalry, "騎兵", 5000, 2000},
		{BranchArtiller, "砲兵", 10000, 200},
		{BranchArmour, "裝甲", 2000, 100},
	}
	for _, c := range cases {
		full := int(BranchFullStrength(c.branch))
		if got := AIRecruitGoldFor(c.branch, full); got != c.goldForFull {
			t.Errorf("%s 徵滿 %d 要 %d 金，預期 %d",
				c.name, full, got, c.goldForFull)
		}
		if got := AIRecruitSoldiers(c.branch, 1000); got != c.soldiersPer1k {
			t.Errorf("%s 花 1000 金得 %d 兵，預期 %d",
				c.name, got, c.soldiersPer1k)
		}
	}
	// 未知兵種什麼也不做——原版四個 cmp 都不中就跳出口。
	if AIRecruitSoldiers(99, 1000) != 0 || AIRecruitGoldFor(99, 100) != 0 {
		t.Error("未知兵種應該回 0")
	}
}

// 步兵的匯率要與實機觀察一致（`recruit.go` 的 RecruitSoldiersPerGold）。
//
// 那個常數是從畫面「黃金 4,150 → 上限 41,500」推的，而實機只測過徵步兵。
// 兩個獨立來源對得上，這一項可以互相佐證。
func TestAIRecruitInfantryMatchesObservedRate(t *testing.T) {
	if got := AIRecruitSoldiers(BranchInfantry, 4150); got != 41500 {
		t.Errorf("黃金 4,150 徵得 %d 步兵，實機畫面是 41,500", got)
	}
	if got := AIRecruitSoldiers(BranchInfantry, 1); got != RecruitSoldiersPerGold {
		t.Errorf("1 金 %d 兵，與 recruit.go 的 %d 對不上",
			got, RecruitSoldiersPerGold)
	}
}

// 單次上限 2,000 金，而且夾到滿員之後只扣實際用掉的錢。
func TestAIRecruitInto(t *testing.T) {
	t.Run("單次最多動用 2000 金", func(t *testing.T) {
		// 空的步兵師，預算給很多 → 只會用 2,000 金補 20,000 兵。
		added, spent := AIRecruitInto(BranchInfantry, 0, 100000)
		if added != 20000 {
			t.Errorf("補了 %d 兵，2,000 金應該補滿 20,000", added)
		}
		if spent != AIRecruitPerCall {
			t.Errorf("花了 %d 金，預期 %d", spent, AIRecruitPerCall)
		}
	})
	t.Run("夾到滿員只扣實際花費", func(t *testing.T) {
		// 已經有 19,000 兵，只差 1,000 → 只該花 100 金。
		added, spent := AIRecruitInto(BranchInfantry, 19000, 2000)
		if added != 1000 {
			t.Errorf("補了 %d 兵，只差 1,000", added)
		}
		if spent != 100 {
			t.Errorf("花了 %d 金，補 1,000 步兵只要 100", spent)
		}
	})
	t.Run("已經滿員就不動", func(t *testing.T) {
		added, spent := AIRecruitInto(BranchInfantry, 20000, 2000)
		if added != 0 || spent != 0 {
			t.Errorf("滿員的部隊補了 %d 兵花 %d 金，兩個都該是 0", added, spent)
		}
	})
	t.Run("騎兵一次補不滿", func(t *testing.T) {
		// 騎兵 1 金 2 兵，2,000 金只能補 4,000，離滿員 10,000 還很遠。
		added, spent := AIRecruitInto(BranchCavalry, 0, 100000)
		if added != 4000 {
			t.Errorf("補了 %d 騎兵，2,000 金應該只補 4,000", added)
		}
		if spent != AIRecruitPerCall {
			t.Errorf("花了 %d 金，預期 %d", spent, AIRecruitPerCall)
		}
	})
	t.Run("未知兵種不動", func(t *testing.T) {
		if a, s := AIRecruitInto(99, 0, 5000); a != 0 || s != 0 {
			t.Errorf("未知兵種補了 %d 花 %d", a, s)
		}
	})
}

// 徵新兵會拉低素質，砲兵與裝甲例外。
func TestAIRecruitQualityCost(t *testing.T) {
	t.Run("步兵掉戰技與 F20", func(t *testing.T) {
		skill, f20 := AIRecruitQualityCost(BranchInfantry, 100, 100)
		if skill != 80 {
			t.Errorf("戰技 100 → %d，預期 80（−20%%）", skill)
		}
		if f20 != 75 {
			t.Errorf("F20 100 → %d，預期 75（−25%%）", f20)
		}
	})
	t.Run("F20 低於門檻就不扣", func(t *testing.T) {
		_, f20 := AIRecruitQualityCost(BranchInfantry, 100, AIRecruitF20Min-1)
		if f20 != AIRecruitF20Min-1 {
			t.Errorf("F20 %d 低於門檻 %d 不該扣，實際變成 %d",
				AIRecruitF20Min-1, AIRecruitF20Min, f20)
		}
	})
	t.Run("砲兵與裝甲不掉素質", func(t *testing.T) {
		for _, b := range []uint8{BranchArtiller, BranchArmour} {
			skill, f20 := AIRecruitQualityCost(b, 100, 100)
			if skill != 100 || f20 != 100 {
				t.Errorf("兵種 %d 掉了素質（%d/%d），原版排除這兩個兵種",
					b, skill, f20)
			}
		}
	})
	t.Run("騎兵與步兵一樣會掉", func(t *testing.T) {
		skill, f20 := AIRecruitQualityCost(BranchCavalry, 100, 100)
		if skill == 100 || f20 == 100 {
			t.Error("騎兵不在排除名單裡，應該跟步兵一樣掉素質")
		}
	})
}
