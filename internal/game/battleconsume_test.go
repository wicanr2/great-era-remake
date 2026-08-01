package game

import "testing"

// §55 的兩張消耗表，用**滿員兵力**當基準——那是最容易對照的一組數字。
func TestAmmoAndFuelAtFullStrength(t *testing.T) {
	cases := []struct {
		name             string
		branch           uint8
		force            uint16
		ammo, fuel, rail int
	}{
		{"步兵", BranchInfantry, 20000, 100, 100, 33},
		{"砲兵", BranchArtiller, 2000, 500, 200, 66},
		{"裝甲", BranchArmour, 200, 400, 200, 66},
		{"騎兵", BranchCavalry, 10000, 50, 0, 0},
	}
	for _, c := range cases {
		if got := AmmoPerShot(c.branch, c.force); got != c.ammo {
			t.Errorf("%s 射擊耗彈 %d，預期 %d", c.name, got, c.ammo)
		}
		if got := FuelPerMove(c.branch, c.force, false); got != c.fuel {
			t.Errorf("%s 移動耗油 %d，預期 %d", c.name, got, c.fuel)
		}
		if got := FuelPerMove(c.branch, c.force, true); got != c.rail {
			t.Errorf("%s 走鐵路耗油 %d，預期 %d", c.name, got, c.rail)
		}
	}
}

func TestCavalryBurnsNoFuel(t *testing.T) {
	// ⭐ 馬不吃油。任何兵力都一樣。
	for _, f := range []uint16{1, 500, 10000, 65535} {
		if got := FuelPerMoveBase(BranchCavalry, f); got != 0 {
			t.Errorf("騎兵 %d 兵該耗油 0，實得 %d", f, got)
		}
	}
}

func TestArtilleryIsTheAmmoHog(t *testing.T) {
	// 砲兵一次齊射吃掉的彈藥是步兵的五倍（滿員對滿員）。
	art := AmmoPerShot(BranchArtiller, 2000)
	inf := AmmoPerShot(BranchInfantry, 20000)
	if art != inf*5 {
		t.Errorf("砲兵 %d 該是步兵 %d 的五倍", art, inf)
	}
}

func TestSpendClampsAtZero(t *testing.T) {
	// 不足就歸零，回傳實際扣掉的量。
	s := BattleSupply{Ammo: 30, Fuel: 10}
	if got := s.SpendAmmo(100); got != 30 || s.Ammo != 0 {
		t.Errorf("彈藥不足該扣光：扣了 %d，剩 %d", got, s.Ammo)
	}
	if got := s.SpendFuel(4); got != 4 || s.Fuel != 6 {
		t.Errorf("燃料夠該正常扣：扣了 %d，剩 %d", got, s.Fuel)
	}
}

// ⭐ 把 §49 與 §55 接起來：彈藥打光就輸。
func TestFiringUntilAmmoRunsOutEndsTheBattle(t *testing.T) {
	// 一個滿員砲兵旅（每次射擊 500 彈藥），帶 1200 彈藥。
	s := BattleSupply{Gold: 1 << 20, Food: 1 << 20, Ammo: 1200, Fuel: 1 << 20,
		Troops: 2000}
	shots := 0
	for s.Ammo > 0 {
		s.SpendAmmo(AmmoPerShot(BranchArtiller, 2000))
		shots++
		if shots > 10 {
			t.Fatal("扣不完")
		}
	}
	if shots != 3 {
		t.Errorf("1200 彈藥每次 500 該打 3 次（第 3 次扣光），實得 %d", shots)
	}
	// 回合結算此時判定補給見底（§49）。
	if !s.Upkeep() {
		t.Error("彈藥為 0 該在回合結算時判見底")
	}
}
