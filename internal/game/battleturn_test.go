package game

import "testing"

func TestUpkeepConsumesGoldAndFood(t *testing.T) {
	// 三個滿員步兵師：黃金 −Round(60000/1500)=−40、糧食 −Round(60000/150)=−400。
	s := BattleSupply{Gold: 4200, Food: 18050, Ammo: 500, Fuel: 500, Troops: 60000}
	if s.Upkeep() {
		t.Fatal("補給還很多，不該見底")
	}
	if s.Gold != 4160 || s.Food != 17650 {
		t.Errorf("扣完該是 4160／17650，實得 %d／%d", s.Gold, s.Food)
	}
}

func TestUpkeepEndsOnEmptyAmmoOrFuel(t *testing.T) {
	// 彈藥與燃料在這裡只被檢查、不被扣——為 0 就直接見底。
	for _, c := range []struct {
		name       string
		ammo, fuel int
	}{
		{"彈藥為 0", 0, 500},
		{"燃料為 0", 500, 0},
	} {
		s := BattleSupply{Gold: 9999, Food: 9999, Ammo: c.ammo, Fuel: c.fuel, Troops: 1000}
		if !s.Upkeep() {
			t.Errorf("%s 該判見底", c.name)
		}
	}
}

func TestUpkeepQuirkOnlyClampsNegatives(t *testing.T) {
	// ⚠️ 原版怪癖：見底時只有**變成負數**的那一項會被寫成 0，
	// 沒變負的那一項保持**扣除前**的原值。
	s := BattleSupply{Gold: 10, Food: 9999, Ammo: 500, Fuel: 500, Troops: 60000}
	if !s.Upkeep() {
		t.Fatal("黃金 10 扣 40 該見底")
	}
	if s.Gold != 0 {
		t.Errorf("變負的黃金該歸零，實得 %d", s.Gold)
	}
	if s.Food != 9999 {
		t.Errorf("沒變負的糧食該保持扣除前的 9999，實得 %d", s.Food)
	}
}

func TestTurnUpkeepPicksWinner(t *testing.T) {
	rich := func() BattleSupply {
		return BattleSupply{Gold: 9999, Food: 9999, Ammo: 1, Fuel: 1, Troops: 1000}
	}
	// 第一方見底 → 第二方勝。
	a, b := rich(), rich()
	a.Food = 1
	if got := TurnUpkeep(&a, &b); got != BattleSideSecond {
		t.Errorf("第一方見底該由第二方勝，實得 %d", got)
	}
	// 第二方見底 → 第一方勝。
	a, b = rich(), rich()
	b.Gold = 0
	if got := TurnUpkeep(&a, &b); got != BattleSideFirst {
		t.Errorf("第二方見底該由第一方勝，實得 %d", got)
	}
	// 都沒見底。
	a, b = rich(), rich()
	if got := TurnUpkeep(&a, &b); got != BattleSideNone {
		t.Errorf("都沒見底該回 0，實得 %d", got)
	}
	// 同一回合雙方都見底 → 後算的第二方覆蓋，照原版順序。
	a, b = rich(), rich()
	a.Ammo, b.Ammo = 0, 0
	if got := TurnUpkeep(&a, &b); got != BattleSideFirst {
		t.Errorf("雙方同時見底時第二方後算會覆蓋，該回 1，實得 %d", got)
	}
}

func TestTurnsOfFoodMatchesUpkeepRate(t *testing.T) {
	// ⭐ §49 的關鍵對照：`sub_3A57E` 的分母就是每回合的糧食消耗，
	// 所以那個比率**字面上**是「還能撐幾回合」。
	s := BattleSupply{Food: 4000, Troops: 60000}
	if got, want := s.TurnsOfFood(), 10; got != want {
		t.Fatalf("TurnsOfFood = %d，預期 %d", got, want)
	}
	// 實際扣扣看：10 回合後歸零。
	s.Gold, s.Ammo, s.Fuel = 1<<20, 1, 1
	turns := 0
	for !s.Upkeep() {
		turns++
		if turns > 50 {
			t.Fatal("扣不完")
		}
	}
	if turns != 9 {
		t.Errorf("撐 %d 個完整回合，第 %d 回合見底；預期撐 9 個", turns, turns+1)
	}
}

func TestBattleTurnCap(t *testing.T) {
	if BattleTurnCap(false) != 16 || BattleTurnCap(true) != 15 {
		t.Error("回合上限該是 16／15")
	}
	if AutoBattleTurnCap != BattleTurnLimit {
		t.Error("AutoResolve 的上限該對齊原版")
	}
}
