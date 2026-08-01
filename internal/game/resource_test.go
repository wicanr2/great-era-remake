package game

import "testing"

// TestAddResourceClamps 累加要夾到 60000，而且中間不能用 u16 溢位。
func TestAddResourceClamps(t *testing.T) {
	for _, c := range []struct{ cur, delta, want uint16 }{
		{0, 0, 0},
		{100, 200, 300},
		{59000, 500, 59500},
		{59000, 2000, ResourceCap},   // 剛好超過
		{ResourceCap, 1, ResourceCap}, // 已經在上限
		{50000, 20000, ResourceCap},   // u16 會溢位成 4464，原版給 60000
		{65535, 1, ResourceCap},
	} {
		if got := AddResource(c.cur, c.delta); got != c.want {
			t.Errorf("AddResource(%d, %d) = %d，預期 %d", c.cur, c.delta, got, c.want)
		}
	}
}

// TestMergeResourcesAfterBattle 戰後資源合併，礦藏不動。
func TestMergeResourcesAfterBattle(t *testing.T) {
	p := Province{
		Gold: 4200, Food: 18050, Ammo: 8787,
		Coal: 13000, Fuel: 12048, Iron: 14031,
	}
	a := BattleResources{Gold: 1000, Food: 2000, Ammo: 3000, Fuel: 4000}
	b := BattleResources{Gold: 500, Food: 600, Ammo: 700, Fuel: 800}
	p.MergeResourcesAfterBattle(a, b)

	if p.Gold != 1500 || p.Food != 2600 || p.Ammo != 3700 || p.Fuel != 4800 {
		t.Errorf("合併後 金%d 糧%d 彈%d 燃%d，預期 1500/2600/3700/4800",
			p.Gold, p.Food, p.Ammo, p.Fuel)
	}
	// 礦藏維持原值——原版沒有歸零 +6 與 +10
	if p.Coal != 13000 || p.Iron != 14031 {
		t.Errorf("煤礦/鐵礦被動到了：%d / %d，預期 13000 / 14031", p.Coal, p.Iron)
	}
}

// TestMergeClampsToCapS 合併時也要夾上限。
func TestMergeClampsToCap(t *testing.T) {
	var p Province
	a := BattleResources{Gold: 40000, Food: 40000}
	b := BattleResources{Gold: 40000, Food: 30000}
	p.MergeResourcesAfterBattle(a, b)
	if p.Gold != ResourceCap || p.Food != ResourceCap {
		t.Errorf("合併後應夾到 %d，實得 金%d 糧%d", ResourceCap, p.Gold, p.Food)
	}
}
