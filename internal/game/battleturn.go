package game

// 戰鬥的**回合上限**與**每回合補給消耗**（`sub_41D20` → `sub_4136B` →
// `sub_41209`，`docs/re/31` §49）。
//
// 這三支是回合結束時依序跑的：先扣補給、再加回合數、再檢查上限。
// 它們共用三個全域旗標：
//
//	byte_64900  回合數，開戰時 1，每回合末 +1
//	byte_64901  勝方：1 = 第一方、2 = 第二方；0 = 還沒分出來
//	byte_6A40A  戰鬥結束旗標

const (
	// BattleTurnLimit 是戰鬥的回合上限：**16**。
	//
	// `sub_41209`：`byte_64900 == 10h` 就設結束旗標。因為 `byte_64900`
	// 開戰時是 1、每回合末才 +1，所以實際打得到的是 15 個完整回合。
	BattleTurnLimit = 16
	// BattleTurnLimitAlt 是 `byte_6FE7E == 2` 那個模式的上限：**15**。
	// ⚠️ `byte_6FE7E` 的語意未解（疑似戰鬥種類，例如攻城 vs 野戰）。
	BattleTurnLimitAlt = 15

	// SupplyGoldDivisor / SupplyFoodDivisor 是每回合的補給消耗除數。
	//
	//	黃金 -= Round(兵力總和 / 1500)
	//	糧食 -= Round(兵力總和 / 150)
	//
	// 兩個都是組語裡的 Turbo Pascal Real 常數，用 `tools/tp_real.py` 解的。
	// **糧食的消耗率是黃金的 10 倍。**
	SupplyGoldDivisor = 1500
	SupplyFoodDivisor = 150
)

// BattleTurnCap 回傳這場戰鬥的回合上限。
//
// `alt` 對應原版的 `byte_6FE7E == 2`——語意未解，呼叫端目前一律給 false。
func BattleTurnCap(alt bool) int {
	if alt {
		return BattleTurnLimitAlt
	}
	return BattleTurnLimit
}

// Upkeep 扣掉這一方**一個回合**的補給，回傳補給是不是見底了。
//
// 對應 `sub_4136B` 的一半（另一半是對方，兩半完全對稱）。見底的條件有四個：
//
//	黃金扣完 <= 0    ／ 糧食扣完 <= 0    ／ 燃料 == 0    ／ 彈藥 == 0
//
// 燃料與彈藥在這裡**只被檢查、不被扣**——它們的消耗發生在別處
// （移動與射擊，未追）。
//
// ⚠️ **原版的怪癖照著做了**：見底時只把**變成負數的那一項**寫成 0，
// 沒變負的那一項**保持扣除前的原值**（`loc_4145F` 那一段只有兩個
// 條件式歸零，沒有無條件寫回）。所以敗方最後一回合的消耗會少扣一部分。
// 這在原版只影響戰後結算寫回省份的數字，但**是可觀察的差異**，
// 不要「順手修正」。
func (s *BattleSupply) Upkeep() bool {
	gold := s.Gold - roundDiv(s.Troops, SupplyGoldDivisor)
	food := s.Food - roundDiv(s.Troops, SupplyFoodDivisor)

	if gold > 0 && food > 0 && s.Fuel != 0 && s.Ammo != 0 {
		s.Gold, s.Food = gold, food
		return false
	}
	// 見底：只有負的歸零。
	if gold < 0 {
		s.Gold = 0
	}
	if food < 0 {
		s.Food = 0
	}
	return true
}

// BattleSide 是原版 `byte_64901` 的值域。
type BattleSide uint8

const (
	// BattleSideNone 表示還沒分出勝負。
	BattleSideNone BattleSide = 0
	// BattleSideFirst 是第一方（攻方）。
	BattleSideFirst BattleSide = 1
	// BattleSideSecond 是第二方（守方）。
	BattleSideSecond BattleSide = 2
)

// TurnUpkeep 是回合結束時的補給結算，回傳勝方（`byte_64901`）。
//
// 兩方各扣各的。原版先算第一方、再算第二方，**兩邊都會算**——
// 所以理論上可能同一回合雙方都見底，那時後算的第二方會覆蓋掉
// `byte_64901`（`sub_4136B` 的第二段無條件寫 1）。這裡照同樣的順序。
func TurnUpkeep(first, second *BattleSupply) BattleSide {
	out := BattleSideNone
	if first.Upkeep() {
		out = BattleSideSecond // 第一方見底 → 第二方勝
	}
	if second.Upkeep() {
		out = BattleSideFirst
	}
	return out
}

// TurnsOfFood 是「糧食還能撐幾回合」。
//
// ⭐ 這正是 `sub_3A57E` 算的那個值——它的分母 `兵力/150` 與每回合的
// 糧食消耗是**同一個式子**。所以 §48 那個比率門檻不是抽象的比率，
// 而是字面意義的**補給天數**，而 `< 15` 對上 16 回合的上限，
// 意思就是「糧食撐不完整場」。
func (s BattleSupply) TurnsOfFood() int {
	return SupplyPerTroop(s.Food, s.Troops)
}
