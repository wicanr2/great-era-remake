package game

import "testing"

// 半滿門檻：四個兵種都是滿員的一半。
func TestAIWorthComforting(t *testing.T) {
	cases := []struct {
		branch uint8
		half   int
		name   string
	}{
		{1, 10000, "步兵（權重 1，滿員 20000）"},
		{10, 1000, "權重 10（滿員 2000）"},
		{100, 100, "權重 100（滿員 200）"},
	}
	for _, c := range cases {
		if BranchWeight(c.branch) == 0 {
			continue // 不是有效兵種代碼，跳過
		}
		if !AIWorthComforting(uint16(c.half), c.branch) {
			t.Errorf("%s：兵力 %d 應該達到門檻", c.name, c.half)
		}
		if AIWorthComforting(uint16(c.half-1), c.branch) {
			t.Errorf("%s：兵力 %d 不該達到門檻", c.name, c.half-1)
		}
	}
	// 未知兵種一律不慰勞，不能因為權重 0 就除零或全通過。
	if AIWorthComforting(60000, 99) {
		t.Error("未知兵種不該通過門檻")
	}
}

// 難度分兩檔門檻。
func TestAIComfortThresholds(t *testing.T) {
	if AIComfortStaminaHard <= AIComfortStaminaEasy ||
		AIComfortMoraleHard <= AIComfortMoraleEasy {
		t.Error("高難度的門檻應該比低難度嚴格（數字更大）")
	}
	// 士氣上限是 80，高難度門檻 70 相當接近滿值。
	if AIComfortMoraleHard >= MoraleMax {
		t.Errorf("高難度士氣門檻 %d 不該達到或超過上限 %d",
			AIComfortMoraleHard, MoraleMax)
	}
}

// 觸發條件：同勢力、可行動、半滿、且體力或士氣偏低。
func TestAIComfortGate(t *testing.T) {
	setup := func(t *testing.T, stamina, morale uint8, force uint16,
		faction GeneralID, active bool) (*AIWorld, []General) {
		w := realWorld(t)
		prov, err := w.Table.At(1)
		if err != nil {
			t.Fatal(err)
		}
		prov.Commander = 7
		w.Units = []CombatUnit{{Province: 1, Faction: faction, Active: active}}
		gens := []General{{Force: force, Branch: 1, Stamina: stamina, F30: morale}}
		return w, gens
	}

	t.Run("體力偏低就慰勞", func(t *testing.T) {
		w, g := setup(t, AIComfortStaminaHard-1, 80, 20000, 7, true)
		if !w.AIComfortGate(1, g, true) {
			t.Error("體力 79 < 80 應該觸發")
		}
	})
	t.Run("士氣偏低就慰勞", func(t *testing.T) {
		w, g := setup(t, 100, AIComfortMoraleHard-1, 20000, 7, true)
		if !w.AIComfortGate(1, g, true) {
			t.Error("士氣 69 < 70 應該觸發")
		}
	})
	t.Run("兩者都夠就不慰勞", func(t *testing.T) {
		w, g := setup(t, 100, 80, 20000, 7, true)
		if w.AIComfortGate(1, g, true) {
			t.Error("狀態都好不該觸發")
		}
	})
	t.Run("低難度門檻較寬", func(t *testing.T) {
		w, g := setup(t, 70, 60, 20000, 7, true)
		if w.AIComfortGate(1, g, false) {
			t.Error("體力 70／士氣 60 在低難度下不該觸發")
		}
		if !w.AIComfortGate(1, g, true) {
			t.Error("同樣的狀態在高難度下該觸發")
		}
	})
	t.Run("殘兵不慰勞", func(t *testing.T) {
		w, g := setup(t, 10, 10, 9999, 7, true)
		if w.AIComfortGate(1, g, true) {
			t.Error("兵力 9999 < 半滿 10000，不該慰勞")
		}
	})
	t.Run("別的勢力不慰勞", func(t *testing.T) {
		w, g := setup(t, 10, 10, 20000, 8, true)
		if w.AIComfortGate(1, g, true) {
			t.Error("效忠別人的將領不該觸發")
		}
	})
	t.Run("不可行動的不慰勞", func(t *testing.T) {
		w, g := setup(t, 10, 10, 20000, 7, false)
		if w.AIComfortGate(1, g, true) {
			t.Error("Active=false 不該觸發")
		}
	})
}
