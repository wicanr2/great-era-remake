package game

import "testing"

// 一般搬運：來源減、目標加，數量相符。
func TestSupplyBasic(t *testing.T) {
	w := realWorld(t)
	src, _ := w.Table.At(1)
	dst, _ := w.Table.At(2)
	src.Gold, dst.Gold = 5000, 1000

	moved, err := w.Supply(1, 2, SupplyGold, 3000)
	if err != nil {
		t.Fatal(err)
	}
	if moved != 3000 || src.Gold != 2000 || dst.Gold != 4000 {
		t.Errorf("搬 3000：實搬 %d、來源 %d、目標 %d，預期 3000／2000／4000",
			moved, src.Gold, dst.Gold)
	}
}

// ⭐ 超過上限的部分退回來源省，不會憑空消失。
//
// 這是 `sub_3083A` 與 `AddResource`（截斷）最大的差別。
func TestSupplyOverflowReturnsToSource(t *testing.T) {
	w := realWorld(t)
	src, _ := w.Table.At(1)
	dst, _ := w.Table.At(2)
	src.Gold = 20000
	dst.Gold = ResourceCap - 5000 // 55000，還能收 5000

	moved, err := w.Supply(1, 2, SupplyGold, 20000)
	if err != nil {
		t.Fatal(err)
	}
	if dst.Gold != ResourceCap {
		t.Errorf("目標省 %d，預期填到 %d", dst.Gold, ResourceCap)
	}
	// 只有 5000 真的搬過去，其餘 15000 退回來源。
	if moved != 5000 {
		t.Errorf("實搬 %d，預期 5000", moved)
	}
	if src.Gold != 15000 {
		t.Errorf("來源省 %d，預期 15000（20000 − 5000）", src.Gold)
	}
	// 總量守恆——這是這條規則的重點。
	if int(src.Gold)+int(dst.Gold) != 15000+ResourceCap {
		t.Errorf("總量不守恆：%d + %d", src.Gold, dst.Gold)
	}
}

// 目標省已經滿了 → 一分都搬不走，全部退回。
func TestSupplyToFullProvince(t *testing.T) {
	w := realWorld(t)
	src, _ := w.Table.At(1)
	dst, _ := w.Table.At(2)
	src.Food, dst.Food = 10000, ResourceCap

	moved, err := w.Supply(1, 2, SupplyFood, 10000)
	if err != nil {
		t.Fatal(err)
	}
	if moved != 0 || src.Food != 10000 || dst.Food != ResourceCap {
		t.Errorf("目標已滿：實搬 %d、來源 %d、目標 %d，預期 0／10000／%d",
			moved, src.Food, dst.Food, ResourceCap)
	}
}

// 四種資源都能搬，煤鐵不行。
func TestSupplyGoods(t *testing.T) {
	w := realWorld(t)
	src, _ := w.Table.At(1)
	dst, _ := w.Table.At(2)
	src.Gold, src.Food, src.Ammo, src.Fuel = 1000, 1000, 1000, 1000
	dst.Gold, dst.Food, dst.Ammo, dst.Fuel = 0, 0, 0, 0

	for _, g := range []SupplyGood{SupplyGold, SupplyFood, SupplyAmmo, SupplyFuel} {
		if _, err := w.Supply(1, 2, g, 100); err != nil {
			t.Errorf("搬 %s 失敗：%v", SupplyGoodName(g), err)
		}
	}
	if dst.Gold != 100 || dst.Food != 100 || dst.Ammo != 100 || dst.Fuel != 100 {
		t.Errorf("四種都該各搬 100，得到 %d/%d/%d/%d",
			dst.Gold, dst.Food, dst.Ammo, dst.Fuel)
	}
	// 煤礦與鐵礦沒有對應的常數，用越界值驗證會被擋。
	if _, err := w.Supply(1, 2, SupplyGood(99), 100); err == nil {
		t.Error("未知資源應該報錯")
	}
}

// 餘額不足擋下來（remake 行為），且不動任何欄位。
func TestSupplyInsufficient(t *testing.T) {
	w := realWorld(t)
	src, _ := w.Table.At(1)
	dst, _ := w.Table.At(2)
	src.Ammo, dst.Ammo = 500, 0

	if _, err := w.Supply(1, 2, SupplyAmmo, 1000); err == nil {
		t.Error("只有 500 卻搬 1000 應該報錯")
	}
	if src.Ammo != 500 || dst.Ammo != 0 {
		t.Error("報錯後不該動欄位")
	}
}

// 同省搬運沒有意義，擋下來。
func TestSupplySameProvince(t *testing.T) {
	w := realWorld(t)
	if _, err := w.Supply(1, 1, SupplyGold, 100); err == nil {
		t.Error("來源與目標相同應該報錯")
	}
}

// SupplyAll 依原版順序搬四種。
func TestSupplyAll(t *testing.T) {
	w := realWorld(t)
	src, _ := w.Table.At(1)
	dst, _ := w.Table.At(2)
	src.Gold, src.Food, src.Ammo, src.Fuel = 900, 800, 700, 600
	dst.Gold, dst.Food, dst.Ammo, dst.Fuel = 0, 0, 0, 0

	moved, err := w.SupplyAll(1, 2, [4]int{100, 200, 300, 400})
	if err != nil {
		t.Fatal(err)
	}
	if moved != [4]int{100, 200, 300, 400} {
		t.Errorf("實搬 %v，預期 [100 200 300 400]", moved)
	}
	if dst.Gold != 100 || dst.Food != 200 || dst.Ammo != 300 || dst.Fuel != 400 {
		t.Errorf("目標省 %d/%d/%d/%d", dst.Gold, dst.Food, dst.Ammo, dst.Fuel)
	}
	if src.Gold != 800 || src.Fuel != 200 {
		t.Errorf("來源省黃金 %d、燃料 %d，預期 800／200", src.Gold, src.Fuel)
	}
}

// 搬 0 或負數什麼都不做。
func TestSupplyZero(t *testing.T) {
	w := realWorld(t)
	src, _ := w.Table.At(1)
	src.Gold = 1000
	for _, n := range []int{0, -5} {
		moved, err := w.Supply(1, 2, SupplyGold, n)
		if err != nil || moved != 0 {
			t.Errorf("搬 %d：moved=%d err=%v，預期 0／nil", n, moved, err)
		}
	}
	if src.Gold != 1000 {
		t.Error("搬 0 不該動欄位")
	}
}
