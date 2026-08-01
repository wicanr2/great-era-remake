package game

import "testing"

// 停止門檻是 Round(目標省戰力 × 2.5)。
func TestAISortieStopThreshold(t *testing.T) {
	cases := []struct{ in, want int }{
		{0, 0},
		{2, 5},
		{4, 10},
		{1, 3},   // 2.5 → Round 3（半數進位）
		{3, 8},   // 7.5 → Round 8
		{100, 250},
	}
	for _, c := range cases {
		if got := AISortieStopThreshold(c.in); got != c.want {
			t.Errorf("目標戰力 %d → 門檻 %d，預期 %d", c.in, got, c.want)
		}
	}
}

// sortieWorld 造一份可控的將領清單：戰力都相同，兵種可指定。
func sortieRoster(n int, str int, branch map[GeneralID]uint8) (
	[]GeneralID, []StrengthInput, StrengthOpts, func(GeneralID) uint8) {
	var roster []GeneralID
	var in []StrengthInput
	for i := 1; i <= n; i++ {
		roster = append(roster, GeneralID(i))
		// 參數要能算出遠高於 AISortieMinStrength(10) 的戰力，
		// 否則整份清單會被單兵下限過濾光——第一版就是這樣紅的。
		in = append(in, StrengthInput{
			Ability: 80, Force: uint16(str), F19: 10, F20: 10,
			F29: 100, F30: 80, Branch: 1, General: GeneralID(i),
		})
	}
	return roster, in, StrengthOpts{Stage: 1}, func(g GeneralID) uint8 {
		if b, ok := branch[g]; ok {
			return b
		}
		return 1
	}
}

// 至少留一個守家：清單長度 − 1 是上限。
func TestPlanSortieLeavesOneBehind(t *testing.T) {
	roster, in, opt, branch := sortieRoster(4, 20000, nil)
	// 目標戰力給很大，讓「湊夠了」永遠不成立，只剩人數上限生效。
	plan := PlanSortie(roster, 100000, 1_000_000, in, opt, branch)
	if len(plan.Generals) != len(roster)-1 {
		t.Errorf("派出 %d 人，清單有 %d 人，應該留一個守家",
			len(plan.Generals), len(roster))
	}
}

// 硬上限 10：清單再長也只帶 10 個。
func TestPlanSortieHardCap(t *testing.T) {
	roster, in, opt, branch := sortieRoster(30, 20000, nil)
	plan := PlanSortie(roster, 100000, 1_000_000, in, opt, branch)
	if len(plan.Generals) != AISortieMaxGenerals {
		t.Errorf("派出 %d 人，硬上限是 %d", len(plan.Generals), AISortieMaxGenerals)
	}
}

// 兵種 4 一個都不收（原版 inc 排在比較之前造成的）。
func TestPlanSortieExcludesBranch4(t *testing.T) {
	branch := map[GeneralID]uint8{2: 4, 3: 4}
	roster, in, opt, branchOf := sortieRoster(5, 20000, branch)
	plan := PlanSortie(roster, 100000, 1_000_000, in, opt, branchOf)
	for _, g := range plan.Generals {
		if branchOf(g) == AISortieBranch4Excluded {
			t.Errorf("名單裡出現兵種 4 的將領 %d，原版一個都不收", g)
		}
	}
	if len(plan.Generals) == 0 {
		t.Error("其他兵種應該還是收得到")
	}
}

// 戰力低於 10 的不收。
func TestPlanSortieSkipsWeak(t *testing.T) {
	roster, in, opt, branch := sortieRoster(5, 20000, nil)
	// 把第 2 個打成殘兵。
	in[1].Force = 0
	in[1].Ability = 0
	if s := Strength(in[1], opt); s >= AISortieMinStrength {
		t.Skipf("造不出戰力 < %d 的部隊（實際 %d），跳過",
			AISortieMinStrength, s)
	}
	plan := PlanSortie(roster, 100000, 1_000_000, in, opt, branch)
	for _, g := range plan.Generals {
		if g == 2 {
			t.Error("戰力低於門檻的將領 2 不該被收進名單")
		}
	}
}

// 湊夠了就停：門檻設低一點，只會收到剛好超過的人數。
func TestPlanSortieStopsWhenEnough(t *testing.T) {
	roster, in, opt, branch := sortieRoster(10, 20000, nil)
	one := Strength(in[0], opt)
	if one <= 0 {
		t.Skip("造不出正的戰力")
	}
	// 目標戰力設成「一個人就超過門檻」的程度：
	// Round(target × 2.5) ≤ one  ⟹  target ≤ one/2.5
	target := one / 3
	plan := PlanSortie(roster, 100000, target, in, opt, branch)
	if len(plan.Generals) != 1 {
		t.Errorf("派出 %d 人（每人戰力 %d，門檻 %d），一個就該夠了",
			len(plan.Generals), one, AISortieStopThreshold(target))
	}
	if plan.Strength != one {
		t.Errorf("名單戰力 %d，預期 %d", plan.Strength, one)
	}
	if plan.Remaining != 100000-one {
		t.Errorf("留守戰力 %d，預期 %d", plan.Remaining, 100000-one)
	}
}

// 空清單不能 panic。
func TestPlanSortieEmpty(t *testing.T) {
	plan := PlanSortie(nil, 500, 100, nil, StrengthOpts{}, func(GeneralID) uint8 { return 1 })
	if len(plan.Generals) != 0 || plan.Remaining != 500 {
		t.Errorf("空清單的結果不對：%+v", plan)
	}
}

// 出兵閘門的分支。
//
// ⛔ 這個測試原本有兩個子測試在驗 `Field234` 的三向分支（< 3 否決、
// > 3 放行、== 3 才看 `Field236`）。那三向分支是**把一個 32-bit 比較
// 讀成兩個獨立欄位**造成的（`docs/re/13` §3），訂正後只剩一條
// 「兵力總和 ≥ 260,000」。舊的子測試連同那個誤解一起刪掉。
func TestSortieGate(t *testing.T) {
	base := SortieGateInput{
		Approved: true, Count: 3,
		PlanStrength: 100, TargetStrength: 100, // 100 < 200 → 第一關不過
		TotalForce: 0,                          // 家底也不夠
	}
	if SortieGate(base) {
		t.Error("三條放行條件都不成立，不該出兵")
	}

	t.Run("旗標沒放行就不出兵", func(t *testing.T) {
		in := base
		in.Approved = false
		in.PlanStrength = 100000 // 戰力再高也沒用
		if SortieGate(in) {
			t.Error("Approved=false 應該直接否決")
		}
	})
	t.Run("人數不足就不出兵", func(t *testing.T) {
		in := base
		in.Count = AISortieMinGenerals - 1
		in.PlanStrength = 100000
		if SortieGate(in) {
			t.Errorf("人數 %d < %d 應該直接否決", in.Count, AISortieMinGenerals)
		}
	})
	t.Run("戰力達兩倍就出兵", func(t *testing.T) {
		in := base
		in.PlanStrength = in.TargetStrength * AISortieGateRatio
		if !SortieGate(in) {
			t.Error("名單戰力剛好兩倍就該放行（原版相等時 fallthrough 到放行）")
		}
	})
	t.Run("目標是軟柿子就出兵", func(t *testing.T) {
		in := base
		in.TargetDesperate = true
		if !SortieGate(in) {
			t.Error("目標缺糧又被包圍，即使沒湊到兩倍也該打")
		}
	})
	t.Run("家底夠厚就出兵", func(t *testing.T) {
		in := base
		in.TotalForce = AISortieForceThreshold
		if !SortieGate(in) {
			t.Errorf("兵力總和 %d 達門檻應該放行", in.TotalForce)
		}
		in.TotalForce = AISortieForceThreshold - 1
		if SortieGate(in) {
			t.Error("差一就不該放行")
		}
	})
}

// 門檻是 32-bit 比較拆成兩半：(3 << 16) | 0xF7A0。
func TestSortieForceThresholdIsSplit32Bit(t *testing.T) {
	if want := (3 << 16) | 0xF7A0; AISortieForceThreshold != want {
		t.Errorf("門檻 %d，原版兩個 cmp 合起來是 %d（0x%X）",
			AISortieForceThreshold, want, want)
	}
	// 260,000 剛好是 13 個滿員步兵師——這是它像個刻意選的值的理由。
	if AISortieForceThreshold%20000 != 0 {
		t.Errorf("門檻 %d 不是滿員步兵師（20000）的整數倍，"+
			"與「13 個師」的讀法對不上", AISortieForceThreshold)
	}
}
