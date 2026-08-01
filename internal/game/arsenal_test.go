package game

import "testing"

// 建兵工廠：每座 5,000 金，上限 5 座（`docs/re/20`）。
func TestArsenalCostAndCap(t *testing.T) {
	w := realWorld(t)
	p := ProvinceID(19)
	prov, _ := w.Table.At(p)

	cases := []struct {
		gold    uint16
		have    uint8
		wantMax int
	}{
		{4999, 0, 0}, // 不足一座
		{5000, 0, 1},
		{25000, 0, 5}, // 買得起 5 座，正好上限
		{60000, 0, 5}, // 買得起 12 座，被上限擋住
		{60000, 3, 2}, // 已有 3 座，只能再蓋 2
		{60000, 5, 0}, // 已滿
	}
	for _, c := range cases {
		prov.Gold, prov.Arsenals = c.gold, c.have
		if got := w.ArsenalAffordable(p); got != c.wantMax {
			t.Errorf("黃金 %d、已有 %d 座：算出最多 %d，應為 %d",
				c.gold, c.have, got, c.wantMax)
		}
	}
}

// 蓋成之後黃金與座數都要對。
func TestBuildArsenalApplies(t *testing.T) {
	w := realWorld(t)
	p := ProvinceID(19)
	prov, _ := w.Table.At(p)
	prov.Gold, prov.Arsenals = 20000, 1

	n, err := w.BuildArsenal(p, 3)
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Errorf("蓋了 %d 座，應為 3", n)
	}
	if prov.Gold != 20000-3*ArsenalCost {
		t.Errorf("黃金剩 %d，應為 %d", prov.Gold, 20000-3*ArsenalCost)
	}
	if prov.Arsenals != 4 {
		t.Errorf("兵工廠 %d 座，應為 4", prov.Arsenals)
	}
}

// 超出上限或買不起要擋下來，不能默默截斷。
func TestBuildArsenalRejects(t *testing.T) {
	w := realWorld(t)
	p := ProvinceID(19)
	prov, _ := w.Table.At(p)

	prov.Gold, prov.Arsenals = 20000, 4
	if _, err := w.BuildArsenal(p, 2); err == nil {
		t.Error("已有 4 座還要蓋 2 座（上限 5）應該被擋")
	}
	prov.Gold, prov.Arsenals = 4999, 0
	if _, err := w.BuildArsenal(p, 1); err == nil {
		t.Error("黃金 4999 蓋不起一座，應該被擋")
	}
}

// 五個資料檔裡的兵工廠數都不該超過上限——這是對「上限 5」的獨立佐證。
func TestArsenalCapHoldsInData(t *testing.T) {
	for _, name := range []string{"TOWN(1).DAT", "TOWN(2).DAT", "TOWN(3).DAT"} {
		tbl, err := ParseTownFile(readGame(t, name))
		if err != nil {
			t.Fatal(err)
		}
		for i := range tbl.Province {
			if n := tbl.Province[i].Arsenals; n > ArsenalMax {
				t.Errorf("%s 第 %d 省有 %d 座兵工廠，超過上限 %d",
					name, i+1, n, ArsenalMax)
			}
		}
	}
}
