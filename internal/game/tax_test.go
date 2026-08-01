package game

import "testing"

// 徵稅的三個結構特徵（`docs/re/18`）：黃金 clamp 到 [1500, 6500]、
// 糧食沒有上限只有底、忠誠度固定扣 30。
func TestTaxStructure(t *testing.T) {
	w := realWorld(t)
	p := ProvinceID(19) // 河南
	prov, _ := w.Table.At(p)
	prov.Gold, prov.Food, prov.Loyalty = 0, 0, 41

	// 跑一百次不同種子，收入必須全部落在範圍內。
	for seed := uint32(1); seed <= 100; seed++ {
		w2 := realWorld(t)
		q, _ := w2.Table.At(p)
		q.Gold, q.Food, q.Loyalty = 0, 0, 41
		res, err := w2.Tax(p, NewRand(seed))
		if err != nil {
			t.Fatal(err)
		}
		if res.Gold < TaxGoldMin || res.Gold > TaxGoldMax {
			t.Fatalf("種子 %d：黃金 %d 超出 [%d, %d]",
				seed, res.Gold, TaxGoldMin, TaxGoldMax)
		}
		if res.Food < TaxFoodBase {
			t.Fatalf("種子 %d：糧食 %d 低於底 %d", seed, res.Food, TaxFoodBase)
		}
		if res.LoyaltyAfter != 41-TaxLoyaltyDrop {
			t.Fatalf("種子 %d：忠誠度變成 %d，應為 %d",
				seed, res.LoyaltyAfter, 41-TaxLoyaltyDrop)
		}
	}
}

// 忠誠度不足 30 就歸 0，不會變成負數（u8 會環繞）。
func TestTaxLoyaltyFloor(t *testing.T) {
	w := realWorld(t)
	p := ProvinceID(19)
	prov, _ := w.Table.At(p)
	prov.Loyalty = 12
	res, err := w.Tax(p, NewRand(1))
	if err != nil {
		t.Fatal(err)
	}
	if res.LoyaltyAfter != 0 {
		t.Errorf("忠誠度 12 徵稅後是 %d，應為 0", res.LoyaltyAfter)
	}
}

// 資源上限 60,000（`sub_5A467`）。
func TestTaxResourceCap(t *testing.T) {
	w := realWorld(t)
	p := ProvinceID(19)
	prov, _ := w.Table.At(p)
	prov.Gold, prov.Food = 59000, 59000
	if _, err := w.Tax(p, NewRand(3)); err != nil {
		t.Fatal(err)
	}
	if prov.Gold != ResourceCap || prov.Food != ResourceCap {
		t.Errorf("黃金 %d／糧食 %d，都應該被夾到 %d",
			prov.Gold, prov.Food, ResourceCap)
	}
}

// 同種子同結果——徵稅是可重現的（`docs/re/17`）。
func TestTaxReproducible(t *testing.T) {
	a, b := realWorld(t), realWorld(t)
	ra, _ := a.Tax(19, NewRand(777))
	rb, _ := b.Tax(19, NewRand(777))
	if ra != rb {
		t.Errorf("同種子結果不同：%+v vs %+v", ra, rb)
	}
}

// 下限 1500 會被打到——這解釋了實機三個樣本裡兩次「+1500」
// （`docs/playtest/12` §3）。
func TestTaxHitsGoldFloor(t *testing.T) {
	p := ProvinceID(19)
	floor := 0
	for seed := uint32(1); seed <= 200; seed++ {
		w2 := realWorld(t)
		q, _ := w2.Table.At(p)
		q.Gold, q.Loyalty = 0, 41
		if res, _ := w2.Tax(p, NewRand(seed)); res.Gold == TaxGoldMin {
			floor++
		}
	}
	t.Logf("200 個種子裡有 %d 次打到黃金下限 %d", floor, TaxGoldMin)
}
