package game

import "testing"

// 五種資源的匯率，逐一對照反組譯讀到的除數。
func TestTradeRates(t *testing.T) {
	cases := []struct {
		good TradeGood
		rate int
		name string
	}{
		{GoodFood, 30, "糧食"},
		{GoodAmmo, 10, "彈藥"},
		{GoodFuel, 30, "燃料"},
		{GoodCoal, 20, "煤礦"},
		{GoodIron, 20, "鐵礦"},
	}
	for _, c := range cases {
		// 剛好一單位價格的量，換算出來要正好是 1。
		if got := TradePrice(c.good, c.rate); got != 1 {
			t.Errorf("%s：%d 單位換到 %d 黃金，預期 1", c.name, c.rate, got)
		}
		if got := TradePrice(c.good, c.rate*7); got != 7 {
			t.Errorf("%s：%d 單位換到 %d 黃金，預期 7", c.name, c.rate*7, got)
		}
		if TradeGoodName(c.good) != c.name {
			t.Errorf("%s 的名字是 %q", c.name, TradeGoodName(c.good))
		}
	}
}

// 保底 1：不足一個匯率單位也要算 1 塊。
func TestTradePriceFloor(t *testing.T) {
	for _, n := range []int{1, 2, 29} {
		if got := TradePrice(GoodFood, n); got != 1 {
			t.Errorf("買 %d 糧食算 %d 黃金，原版保底是 1", n, got)
		}
	}
	if got := TradePrice(GoodFood, 0); got != 0 {
		t.Errorf("買 0 應該是 0，得到 %d", got)
	}
}

// 煤礦與鐵礦只能賣不能買。
func TestCoalIronExportOnly(t *testing.T) {
	w := realWorld(t)
	for _, g := range []TradeGood{GoodCoal, GoodIron} {
		if Importable(g) {
			t.Errorf("%s 不該是可進口的", TradeGoodName(g))
		}
		if _, err := w.Import(1, g, 100); err == nil {
			t.Errorf("進口%s 應該報錯", TradeGoodName(g))
		}
	}
	for _, g := range []TradeGood{GoodFood, GoodAmmo, GoodFuel} {
		if !Importable(g) {
			t.Errorf("%s 應該可以進口", TradeGoodName(g))
		}
	}
}

// 進口：資源增加、黃金減少，數字對得上匯率。
func TestImport(t *testing.T) {
	w := realWorld(t)
	prov, err := w.Table.At(1)
	if err != nil {
		t.Fatal(err)
	}
	prov.Gold, prov.Food = 1000, 0

	res, err := w.Import(1, GoodFood, 300)
	if err != nil {
		t.Fatal(err)
	}
	if res.Amount != 300 || res.Gold != 10 {
		t.Errorf("買 300 糧食：入帳 %d、花 %d，預期 300／10", res.Amount, res.Gold)
	}
	if prov.Food != 300 || prov.Gold != 990 {
		t.Errorf("糧食 %d、黃金 %d，預期 300／990", prov.Food, prov.Gold)
	}
}

// 黃金不足擋下來（remake 行為，原版的處理未讀）。
func TestImportInsufficientGold(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	prov.Gold = 5
	if _, err := w.Import(1, GoodFood, 3000); err == nil {
		t.Error("黃金 5 買 3000 糧食（要 100）應該報錯")
	}
}

// 出口：資源減少、黃金增加；資源不足時只賣手上有的。
func TestExport(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	prov.Gold, prov.Coal = 0, 500

	res, err := w.Export(1, GoodCoal, 400)
	if err != nil {
		t.Fatal(err)
	}
	if res.Amount != 400 || res.Gold != 20 {
		t.Errorf("賣 400 煤礦：出貨 %d、得 %d，預期 400／20", res.Amount, res.Gold)
	}
	if prov.Coal != 100 || prov.Gold != 20 {
		t.Errorf("煤礦 %d、黃金 %d，預期 100／20", prov.Coal, prov.Gold)
	}

	// 只剩 100，要賣 999 只能賣 100。
	res, err = w.Export(1, GoodCoal, 999)
	if err != nil {
		t.Fatal(err)
	}
	if res.Amount != 100 || prov.Coal != 0 {
		t.Errorf("超賣：出貨 %d、剩 %d，預期 100／0", res.Amount, prov.Coal)
	}

	if _, err := w.Export(1, GoodCoal, 10); err == nil {
		t.Error("沒有煤礦可賣時應該報錯")
	}
}

// 買賣同價——原版沒有價差，來回換不會虧（除了保底與整除）。
func TestTradeNoSpread(t *testing.T) {
	for _, g := range []TradeGood{GoodFood, GoodAmmo, GoodFuel} {
		n := 600
		buy := TradePrice(g, n)
		sell := TradePrice(g, n)
		if buy != sell {
			t.Errorf("%s 買 %d 花 %d、賣回只得 %d——原版買賣同價",
				TradeGoodName(g), n, buy, sell)
		}
	}
}

// 黃金夾在資源上限內。
func TestExportGoldCap(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	prov.Gold = ResourceCap - 5
	prov.Ammo = 1000

	res, err := w.Export(1, GoodAmmo, 1000)
	if err != nil {
		t.Fatal(err)
	}
	if prov.Gold != ResourceCap || res.Gold != 5 {
		t.Errorf("黃金 %d、入帳 %d，預期 %d／5", prov.Gold, res.Gold, ResourceCap)
	}
}
