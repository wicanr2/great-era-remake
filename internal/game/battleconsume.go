package game

// 戰鬥中的**彈藥與燃料消耗**（`sub_507C7`／`sub_50905`，`docs/re/31` §55）。
//
// §49 解出回合結算只**檢查**彈藥與燃料是否為 0，不扣它們——扣在這裡：
//
//	射擊 → sub_507C7 扣彈藥
//	移動 → sub_50905 扣燃料
//
// 兩支的形狀相同：依兵種與兵力算出消耗量，再依 `+8`（攻方 1／守方 0）
// 從自己那一方的存量扣掉，不足就歸零。

// AmmoPerShot 是一次射擊的彈藥消耗（`sub_507C7` 前半）。
//
//	1 步兵  兵力 / 200
//	4 砲兵  兵力 / 4        ← 最耗
//	5 裝甲  兵力 × 2
//	6 騎兵  兵力 / 200
//
// ⭐ 換成滿員兵力來看差距很大：**砲兵 500、裝甲 400、步兵 100、騎兵 50**。
// 砲兵一次齊射吃掉的彈藥是步兵的五倍。
func AmmoPerShot(branch uint8, force uint16) int {
	f := int(force)
	switch branch {
	case BranchInfantry, BranchCavalry:
		return f / 200
	case BranchArtiller:
		return f / 4
	case BranchArmour:
		return f * 2
	}
	// ⚠️ 原版對 1/4/5/6 以外的兵種**不設消耗變數**（局部變數未初始化就用），
	// 這裡回 0。檔案裡的兵種只有這四個值，所以踩不到。
	return 0
}

// FuelPerMoveBase 是移動的燃料基本消耗（`sub_5088C`）。
//
//	1 步兵  兵力 / 200
//	4 砲兵  兵力 / 10       ← 最耗
//	5 裝甲  兵力（原值）
//	6 騎兵  **0**
//
// ⭐ **騎兵不耗燃料**——馬不吃油。滿員時：步兵 100、砲兵 200、裝甲 200、騎兵 0。
func FuelPerMoveBase(branch uint8, force uint16) int {
	f := int(force)
	switch branch {
	case BranchInfantry:
		return f / 200
	case BranchArtiller:
		return f / 10
	case BranchArmour:
		return f
	case BranchCavalry:
		return 0
	}
	return 0
}

// FuelRailDivisor 是走鐵路的折扣：**除以 3**（`sub_50905`）。
//
// `sub_4FEF0` 判斷該格有沒有鐵路（`22 < 地圖值 & 0FFFh < 550`，
// `docs/re/07` §442 已解）。
const FuelRailDivisor = 3

// FuelPerMove 是實際扣的燃料：基本消耗，走鐵路的話除以 3。
//
// ⭐ 這條規則對這個時代的題材很貼切——1920–40 年代的中國，
// 鐵路就是戰略骨幹。原版把它做進了燃料帳裡。
func FuelPerMove(branch uint8, force uint16, onRail bool) int {
	n := FuelPerMoveBase(branch, force)
	if onRail {
		n /= FuelRailDivisor
	}
	return n
}

// SpendAmmo 依 `sub_507C7` 後半扣彈藥：不足就歸零，回傳實際扣了多少。
func (s *BattleSupply) SpendAmmo(cost int) int {
	if cost > s.Ammo {
		spent := s.Ammo
		s.Ammo = 0
		return spent
	}
	s.Ammo -= cost
	return cost
}

// SpendFuel 依 `sub_50905` 後半扣燃料：不足就歸零，回傳實際扣了多少。
func (s *BattleSupply) SpendFuel(cost int) int {
	if cost > s.Fuel {
		spent := s.Fuel
		s.Fuel = 0
		return spent
	}
	s.Fuel -= cost
	return cost
}
