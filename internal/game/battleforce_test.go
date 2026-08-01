package game

import "testing"

func TestAdvantageLevel(t *testing.T) {
	cases := []struct {
		leader, foe, want int
		note              string
	}{
		{200, 100, AIAdvantageOverwhelming, "剛好兩倍算壓倒（≤ 不是 <）"},
		{199, 100, AIAdvantageAhead, "差一點就只是優勢"},
		{100, 100, AIAdvantageAhead, "打平算優勢"},
		{99, 100, AIAdvantageBehind, "差一點就是劣勢"},
		{0, 0, AIAdvantageOverwhelming, "雙方都是 0：0×2 ≤ 0 成立"},
	}
	for _, c := range cases {
		if got := AdvantageLevel(c.leader, c.foe); got != c.want {
			t.Errorf("AdvantageLevel(%d, %d) = %d，要 %d（%s）",
				c.leader, c.foe, got, c.want, c.note)
		}
	}
}

func TestForceRatioLE(t *testing.T) {
	// sub_3C26A 的 0.5：回 true ⟺ 敵方 ≥ 我方的一半。
	if !ForceRatioLE(100, 50, AIBattleRatioAheadNum, AIBattleRatioAheadDen) {
		t.Error("我方 100、敵方 50，100×0.5 ≤ 50 應成立（邊界）")
	}
	if ForceRatioLE(100, 49, AIBattleRatioAheadNum, AIBattleRatioAheadDen) {
		t.Error("敵方 49 時不該成立——那正是「我方大幅領先」")
	}
	// 5.0：我方被壓到五分之一以下。
	if !ForceRatioLE(20, 100, AIBattleRatioCollapseNum, AIBattleRatioCollapseDen) {
		t.Error("我方 20、敵方 100 應成立")
	}
	if ForceRatioLE(21, 100, AIBattleRatioCollapseNum, AIBattleRatioCollapseDen) {
		t.Error("我方 21 就還沒到五分之一")
	}
	// 0.67 是唯一與原版 Real 有偏差的倍率，這裡只驗整數語意。
	if !ForceRatioLE(100, 67, AIBattleRatioEvenNum, AIBattleRatioEvenDen) {
		t.Error("100×0.67 = 67 ≤ 67 應成立")
	}
}

func TestAIReserveKeepAndActivate(t *testing.T) {
	// 六個單位待命（index 2..7），index 1 放一個待命當「不該被動到」的對照。
	mk := func() []uint8 {
		c := make([]uint8, 11)
		for i := 1; i <= 7; i++ {
			c[i] = BattleCmdStandby
		}
		return c
	}
	count := func(c []uint8, v uint8) int {
		n := 0
		for i := 2; i <= 10; i++ {
			if c[i] == v {
				n++
			}
		}
		return n
	}

	for _, tc := range []struct{ adv, keep int }{
		{AIAdvantageBehind, 2},
		{AIAdvantageAhead, 1},
		{AIAdvantageOverwhelming, 0},
	} {
		c := mk()
		ActivateReserves(c, tc.adv)
		if got := count(c, BattleCmdStandby); got != tc.keep {
			t.Errorf("優勢等級 %d：留下 %d 個待命，要 %d", tc.adv, got, tc.keep)
		}
		if got := count(c, BattleCmdCommitted); got != 6-tc.keep {
			t.Errorf("優勢等級 %d：投入 %d 個，要 %d", tc.adv, got, 6-tc.keep)
		}
		// index 1 永遠不碰——原版兩個迴圈的下界都是 2。
		if c[1] != BattleCmdStandby {
			t.Errorf("優勢等級 %d：index 1 被動到了（%d）", tc.adv, c[1])
		}
	}
}

func TestActivateReservesKeepsHighIndexLast(t *testing.T) {
	// 倒著掃：先被轉的是 index 大的那些，留下的是 index 小的。
	c := make([]uint8, 11)
	for i := 2; i <= 5; i++ {
		c[i] = BattleCmdStandby
	}
	ActivateReserves(c, AIAdvantageBehind) // 留 2
	if c[2] != BattleCmdStandby || c[3] != BattleCmdStandby {
		t.Errorf("留下的應該是 index 最小的兩個，實際 c[2]=%d c[3]=%d", c[2], c[3])
	}
	if c[4] != BattleCmdCommitted || c[5] != BattleCmdCommitted {
		t.Errorf("index 大的先被投入，實際 c[4]=%d c[5]=%d", c[4], c[5])
	}
}

func TestActivateReservesIgnoresNonStandby(t *testing.T) {
	// 只有命令 2 會被轉，其他命令原封不動。
	c := make([]uint8, 11)
	c[2] = BattleCmdGarrison
	c[3] = BattleCmdSeekTarget
	c[4] = BattleCmdStandby
	c[5] = BattleCmdGarrisoned
	ActivateReserves(c, AIAdvantageOverwhelming) // 留 0
	if c[2] != BattleCmdGarrison || c[3] != BattleCmdSeekTarget ||
		c[5] != BattleCmdGarrisoned {
		t.Errorf("非待命的命令被改了：%v", c[1:6])
	}
	if c[4] != BattleCmdCommitted {
		t.Errorf("唯一的待命單位沒被投入：c[4]=%d", c[4])
	}
}
