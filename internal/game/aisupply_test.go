package game

import "testing"

func supplyWorld(t *testing.T, gold, food, ammo, fuel uint16) (*AIWorld, *Province) {
	t.Helper()
	w := realWorld(t)
	prov, err := w.Table.At(1)
	if err != nil {
		t.Fatal(err)
	}
	prov.Gold, prov.Food, prov.Ammo, prov.Fuel = gold, food, ammo, fuel
	return w, prov
}

// 簡單模式：黃金 ≥ 1000 且 糧食 ≥ 10000，再加彈藥燃料。
func TestAISupplySimpleMode(t *testing.T) {
	cases := []struct {
		gold, food, ammo, fuel uint16
		want                   bool
		name                   string
	}{
		{1000, 10000, 15000, 10000, true, "剛好全部達標"},
		{999, 10000, 15000, 10000, false, "黃金差 1"},
		{1000, 9999, 15000, 10000, false, "糧食差 1"},
		{1000, 10000, 14999, 10000, false, "彈藥差 1"},
		{1000, 10000, 15000, 9999, false, "燃料差 1"},
		{60000, 60000, 14999, 60000, false, "錢糧再多，彈藥不夠一樣不行"},
	}
	for _, c := range cases {
		w, _ := supplyWorld(t, c.gold, c.food, c.ammo, c.fuel)
		if got := w.AISupplyOK(1, 0, false, false); got != c.want {
			t.Errorf("%s：得到 %v，預期 %v", c.name, got, c.want)
		}
	}
}

// 進階模式：需求 = 兵力 ÷ 除數，高難度的除數較小 → 需求較大。
func TestAISupplyUpkeep(t *testing.T) {
	const force = 70000
	if AIUpkeepNeed(force, true) != force/AISupplyUpkeepDivHard {
		t.Error("高難度的需求算錯")
	}
	if AIUpkeepNeed(force, false) != force/AISupplyUpkeepDivEasy {
		t.Error("低難度的需求算錯")
	}
	if AIUpkeepNeed(force, true) <= AIUpkeepNeed(force, false) {
		t.Error("高難度的需求應該比低難度大")
	}
}

// 同樣的局面，高難度可能不過而低難度過——這是難度差異的具體表現。
func TestAISupplyDifficultyDiffers(t *testing.T) {
	// 兵力 70000：低難度需求 1000、高難度需求 1400。
	// 黃金 1200 夾在兩者之間，糧食給足以免被第二條件干擾。
	w, _ := supplyWorld(t, 1200, 60000, 15000, 10000)
	if !w.AISupplyOK(1, 70000, true, false) {
		t.Error("低難度（需求 1000 ≤ 黃金 1200）應該過")
	}
	// 高難度需求 1400 > 黃金 1200，但糧食 60000 ≥ 25000 且黃金 < 4000，
	// 所以第二條捷徑也不成立 → 應該不過。
	if w.AISupplyOK(1, 70000, true, true) {
		t.Error("高難度（需求 1400 > 黃金 1200）不該過")
	}
}

// 進階模式的捷徑：黃金 ≥ 4000 且 糧食 ≥ 25000，不看兵力。
func TestAISupplyRichShortcut(t *testing.T) {
	w, _ := supplyWorld(t, AISupplyRichGold, AISupplyRichFood, 15000, 10000)
	// 兵力開到極大，需求遠超黃金，但捷徑成立。
	if !w.AISupplyOK(1, 60000*50, true, true) {
		t.Error("黃金 4000 + 糧食 25000 的捷徑應該讓它通過")
	}
	// 黃金差 1 就不成立。
	w2, _ := supplyWorld(t, AISupplyRichGold-1, AISupplyRichFood, 15000, 10000)
	if w2.AISupplyOK(1, 60000*50, true, true) {
		t.Error("黃金 3999 不該走捷徑")
	}
}

// 彈藥燃料是最後一關，兩套模式都要過。
func TestAISupplyAmmoFuelGateBothModes(t *testing.T) {
	for _, advanced := range []bool{false, true} {
		w, _ := supplyWorld(t, 60000, 60000, AISupplyAmmo-1, 60000)
		if w.AISupplyOK(1, 0, advanced, false) {
			t.Errorf("advanced=%v：彈藥不足仍該擋下", advanced)
		}
		w2, _ := supplyWorld(t, 60000, 60000, 60000, AISupplyFuel-1)
		if w2.AISupplyOK(1, 0, advanced, false) {
			t.Errorf("advanced=%v：燃料不足仍該擋下", advanced)
		}
	}
}
