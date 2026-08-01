package game

import "testing"

// 兵力總和的篩選條件與 `GeneralCount`（`sub_5A881`）逐字相同。
func TestProvinceForceTotal(t *testing.T) {
	w := realWorld(t)
	prov, err := w.Table.At(1)
	if err != nil {
		t.Fatal(err)
	}
	prov.Commander = 7

	w.Units = []CombatUnit{
		{Province: 1, Faction: 7, Active: true},  // 算
		{Province: 1, Faction: 7, Active: false}, // 不可行動
		{Province: 1, Faction: 8, Active: true},  // 別的勢力
		{Province: 2, Faction: 7, Active: true},  // 別的省
		{Province: 1, Faction: 7, Active: true},  // 算
	}
	gens := []General{
		{Force: 20000}, {Force: 9000}, {Force: 5000}, {Force: 3000}, {Force: 15000},
	}

	if got := w.ProvinceForceTotal(1, gens); got != 35000 {
		t.Errorf("兵力總和 %d，預期 35000（只有第 1、5 筆符合）", got)
	}
}

// 原版特地用 add/adc 湊成 32-bit，總和超過 65,535 不能溢位。
func TestProvinceForceTotalOver16Bit(t *testing.T) {
	w := realWorld(t)
	prov, err := w.Table.At(1)
	if err != nil {
		t.Fatal(err)
	}
	prov.Commander = 7

	var gens []General
	w.Units = nil
	for i := 0; i < 5; i++ {
		w.Units = append(w.Units, CombatUnit{Province: 1, Faction: 7, Active: true})
		gens = append(gens, General{Force: 20000})
	}
	if got := w.ProvinceForceTotal(1, gens); got != 100000 {
		t.Errorf("兵力總和 %d，預期 100000（五個滿員步兵師，不可截成 16-bit）", got)
	}
}

// 缺糧判定是「或」：黃金與糧食任一不足就成立。
func TestAISupplyStarving(t *testing.T) {
	// 兵力總和固定 100,000 → 黃金需求 1,000、糧食需求 10,000。
	setup := func(t *testing.T, gold, food uint16) (*AIWorld, []General) {
		t.Helper()
		w := realWorld(t)
		prov, err := w.Table.At(1)
		if err != nil {
			t.Fatal(err)
		}
		prov.Commander = 7
		prov.Gold, prov.Food = gold, food
		var gens []General
		w.Units = nil
		for i := 0; i < 5; i++ {
			w.Units = append(w.Units, CombatUnit{Province: 1, Faction: 7, Active: true})
			gens = append(gens, General{Force: 20000})
		}
		return w, gens
	}

	t.Run("兩者都夠不算缺", func(t *testing.T) {
		w, g := setup(t, 1000, 10000)
		if w.AISupplyStarving(1, g) {
			t.Error("黃金 1000 ≥ 1000、糧食 10000 ≥ 10000，不該算缺")
		}
	})
	t.Run("黃金差一就算缺", func(t *testing.T) {
		w, g := setup(t, 999, 60000)
		if !w.AISupplyStarving(1, g) {
			t.Error("黃金 999 < 1000 就該算缺，糧食再多也一樣")
		}
	})
	t.Run("糧食差一就算缺", func(t *testing.T) {
		w, g := setup(t, 60000, 9999)
		if !w.AISupplyStarving(1, g) {
			t.Error("糧食 9999 < 10000 就該算缺，黃金再多也一樣")
		}
	})
	t.Run("沒有部隊時需求是零", func(t *testing.T) {
		w, _ := setup(t, 0, 0)
		w.Units = nil
		if w.AISupplyStarving(1, nil) {
			t.Error("兵力 0 → 需求 0，空省不該算缺糧")
		}
	})
}

// 除數比是 1:10，與 `AISupplyOK` 進階模式同一套比例。
func TestAIDesperateDivisorRatio(t *testing.T) {
	if AIDesperateGoldDiv != AIDesperateFoodDiv*10 {
		t.Errorf("黃金除數 %d 應該是糧食除數 %d 的十倍（1:10 的資源比）",
			AIDesperateGoldDiv, AIDesperateFoodDiv)
	}
	// 這一關要比 `sub_16B44` 的補給門檻寬鬆：除數更大 → 需求更小。
	if AIDesperateGoldDiv <= AISupplyUpkeepDivEasy {
		t.Errorf("撐不住門檻的除數 %d 應該大於補給門檻的 %d（更寬鬆）",
			AIDesperateGoldDiv, AISupplyUpkeepDivEasy)
	}
}

// `sub_17019` 是「缺糧 **且** 被包圍」，兩者缺一不可。
func TestAIDesperate(t *testing.T) {
	setup := func(t *testing.T, starving, encircled bool) (*AIWorld, []General) {
		t.Helper()
		w := realWorld(t)
		prov, err := w.Table.At(1)
		if err != nil {
			t.Fatal(err)
		}
		prov.Commander = 7
		var gens []General
		w.Units = nil
		for i := 0; i < 5; i++ {
			w.Units = append(w.Units, CombatUnit{Province: 1, Faction: 7, Active: true})
			gens = append(gens, General{Force: 20000})
		}
		if starving {
			prov.Gold, prov.Food = 0, 0
		} else {
			prov.Gold, prov.Food = 60000, 60000
		}
		for _, n := range prov.Neighbours {
			np, err := w.Table.At(n)
			if err != nil {
				continue
			}
			if encircled {
				np.Commander = 8 // 全是敵人
			} else {
				np.Commander = 7 // 自己人
			}
		}
		return w, gens
	}

	t.Run("缺糧且被包圍", func(t *testing.T) {
		w, g := setup(t, true, true)
		if !w.AIDesperate(1, g) {
			t.Error("缺糧又被包圍，應該回 true")
		}
	})
	t.Run("只缺糧不算", func(t *testing.T) {
		w, g := setup(t, true, false)
		if w.AIDesperate(1, g) {
			t.Error("沒被包圍就不算 desperate")
		}
	})
	t.Run("只被包圍不算", func(t *testing.T) {
		w, g := setup(t, false, true)
		if w.AIDesperate(1, g) {
			t.Error("補給充足就不算 desperate，即使被包圍")
		}
	})
}
