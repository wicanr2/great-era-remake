package game

import "testing"

// 墾地：地價 += 政治手腕 ÷ 10，體力 −5，地價夾在 200（`docs/re/21`）。
func TestReclaimLand(t *testing.T) {
	cases := []struct {
		politics, stamina uint8
		before            uint8
		wantGain          int
		wantAfter         uint8
		wantStamina       uint8
	}{
		{94, 100, 19, 9, 28, 95},   // 吳佩孚：94÷10 = 9
		{41, 50, 19, 4, 23, 45},    // 41÷10 = 4
		{9, 50, 19, 0, 19, 45},     // 9÷10 = 0，白跑一趟還是扣體力
		{100, 30, 195, 5, 200, 25}, // 夾到上限
		{100, 30, 200, 0, 200, 25}, // 已滿
	}
	for _, c := range cases {
		w := realWorld(t)
		prov, _ := w.Table.At(19)
		prov.LandValue = c.before
		gain, st, err := w.ReclaimLand(19, c.politics, c.stamina)
		if err != nil {
			t.Fatal(err)
		}
		if gain != c.wantGain || prov.LandValue != c.wantAfter || st != c.wantStamina {
			t.Errorf("政治 %d／體力 %d／地價 %d → 增 %d、地價 %d、體力 %d；"+
				"應為 增 %d、地價 %d、體力 %d",
				c.politics, c.stamina, c.before, gain, prov.LandValue, st,
				c.wantGain, c.wantAfter, c.wantStamina)
		}
	}
}

// 體力不足 5 不能墾地。
func TestReclaimNeedsStamina(t *testing.T) {
	w := realWorld(t)
	if _, _, err := w.ReclaimLand(19, 90, 4); err == nil {
		t.Error("體力 4 不該能墾地（門檻 5）")
	}
	if _, _, err := w.ReclaimLand(19, 90, 5); err != nil {
		t.Errorf("體力 5 正好夠，不該失敗：%v", err)
	}
}

// 五個資料檔的地價都不該超過上限——對「上限 200」的獨立佐證。
func TestLandValueCapHoldsInData(t *testing.T) {
	for _, name := range []string{"TOWN(1).DAT", "TOWN(2).DAT", "TOWN(3).DAT"} {
		tbl, err := ParseTownFile(readGame(t, name))
		if err != nil {
			t.Fatal(err)
		}
		for i := range tbl.Province {
			if v := tbl.Province[i].LandValue; v > LandValueMax {
				t.Errorf("%s 第 %d 省地價 %d 超過上限 %d", name, i+1, v, LandValueMax)
			}
		}
	}
}
