package game

import "fmt"

// 運補（政略指令 3）：把資源搬到別的省。
//
// 入口 `sub_2CB6B`（選目標省）→ `sub_3231A`（問四個數量）→
// 四支各 100 行、結構完全相同的搬運函式：
//
//	sub_3083A  黃金（省份 +0）
//	sub_30918  糧食（+2）
//	sub_309F6  彈藥（+4）
//	sub_30AD4  燃料（+8）
//
// **煤礦與鐵礦沒有對應的函式**——與「不可流動的原料」一致
// （`40-economy.md` §9、`trade.go` 的煤鐵只能賣不能買）。

// SupplyGood 是可運補的四種資源。
type SupplyGood int

const (
	SupplyGold SupplyGood = iota
	SupplyFood
	SupplyAmmo
	SupplyFuel
)

var supplyName = map[SupplyGood]string{
	SupplyGold: "黃金", SupplyFood: "糧食",
	SupplyAmmo: "彈藥", SupplyFuel: "燃料",
}

// SupplyGoodName 回傳資源的原版用詞。
func SupplyGoodName(g SupplyGood) string {
	if n, ok := supplyName[g]; ok {
		return n
	}
	return "未知"
}

func supplyField(p *Province, g SupplyGood) *uint16 {
	switch g {
	case SupplyGold:
		return &p.Gold
	case SupplyFood:
		return &p.Food
	case SupplyAmmo:
		return &p.Ammo
	case SupplyFuel:
		return &p.Fuel
	}
	return nil
}

// Supply 把 amount 單位的資源從 from 搬到 to，回傳**實際搬走的量**。
//
// 語意逐行照 `sub_3083A`：
//
//	dst = 目標省的量 + amount        ; 32-bit，不會中途溢位
//	來源省 −= amount
//	若 dst > 60000:
//	    溢出 = dst − 60000
//	    目標省 = 60000
//	    來源省 += 溢出               ; ← **退回來源省**
//	否則:
//	    目標省 = dst
//	回傳 來源省的減少量
//
// ⭐ **超過上限的部分會退回來源省，不會憑空消失。** 這比單純夾上限細緻得多——
// `AddResource`（`sub_5A467`）那種是直接截掉，兩者**不是同一個行為**，
// 別把這裡換成 `AddResource`。
//
// ⚠️ 原版用 32-bit 中間值算 `dst`，所以 `目標省 + amount` 不會先溢位再判斷。
// 這裡用 `uint32` 重現同一件事。
//
// **來源省的餘額夠不夠，原版在這一支裡沒有檢查**——`sub_3231A` 問數量時
// 應該有上限，但那段還沒讀。這裡擋下超額並回錯，標為 remake 行為。
func (w *AIWorld) Supply(from, to ProvinceID, g SupplyGood, amount int) (int, error) {
	if _, ok := supplyName[g]; !ok {
		return 0, fmt.Errorf("game: %d 不是可運補的資源", g)
	}
	if from == to {
		return 0, fmt.Errorf("game: 來源與目標是同一個省（%d）", from)
	}
	src, err := w.Table.At(from)
	if err != nil {
		return 0, err
	}
	dst, err := w.Table.At(to)
	if err != nil {
		return 0, err
	}
	if amount <= 0 {
		return 0, nil
	}

	sf, df := supplyField(src, g), supplyField(dst, g)
	if amount > int(*sf) {
		return 0, fmt.Errorf("game: %s 只有 %d，搬不了 %d",
			SupplyGoodName(g), *sf, amount)
	}

	before := *sf
	*sf -= uint16(amount)

	total := uint32(*df) + uint32(amount)
	if total > ResourceCap {
		overflow := total - ResourceCap
		*df = ResourceCap
		*sf += uint16(overflow) // 退回來源省
	} else {
		*df = uint16(total)
	}
	return int(before - *sf), nil
}

// SupplyAll 一次搬四種資源，回傳每一種實際搬走的量。
//
// 原版的 `sub_2CB6B` 依序問四次（黃金／糧食／彈藥／燃料）再一起執行，
// 所以這裡照同一個順序。任何一種出錯就整批停下——
// ⚠️ **原版遇到錯誤怎麼處理沒讀**，這是 remake 行為。
func (w *AIWorld) SupplyAll(from, to ProvinceID, amounts [4]int) ([4]int, error) {
	var moved [4]int
	for i, g := range []SupplyGood{SupplyGold, SupplyFood, SupplyAmmo, SupplyFuel} {
		n, err := w.Supply(from, to, g, amounts[i])
		if err != nil {
			return moved, err
		}
		moved[i] = n
	}
	return moved, nil
}
