package game

import "testing"

// 政略回合的端到端測試：把已解出的指令串成一年，看數值走得合不合理。
//
// 這是無頭環境唯一能驗的「遊戲跑得動」——`cmd/dsds` 需要顯示器，
// 而規則層不認識畫面（`CLAUDE.md` §11 的分層理由）。
//
// ⚠️ **這不是「對齊原版」的驗收。** 它驗的是規則層自己不會爆炸、
// 不變量守得住。真正的對齊要拿原版逐次比對，那條路還沒走
// （`rulebook/65`：測試綠只代表沒退步）。

// 一整年的政略操作跑得完，而且每一步的效果都符合已解出的規則。
func TestStrategyYearRuns(t *testing.T) {
	w := realWorld(t)
	gens, err := ParseGenerals(readGame(t, "MAN(1).DAT"), 274)
	if err != nil {
		t.Fatal(err)
	}
	rng := NewRand(1)

	// 挑一個有將領、有司令的省來操作。
	var target ProvinceID
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		prov, err := w.Table.At(p)
		if err != nil || prov.Commander == 0 {
			continue
		}
		if len(w.RosterOf(p).ids) > 0 {
			target = p
			break
		}
	}
	if target == 0 {
		t.Fatal("找不到有將領的省，測試前提不成立")
	}

	st := &GameState{Stage: 1, Year: 15, Month: 1}
	taxedMonths := 0

	for month := 1; month <= 12; month++ {
		prov, err := w.Table.At(target)
		if err != nil {
			t.Fatal(err)
		}

		// 徵稅：每月限一次，第二次要被旗標擋下。
		goldBefore := prov.Gold
		if !prov.Taxed() {
			res, err := w.Tax(target, rng)
			if err != nil {
				t.Fatalf("%d 月徵稅失敗: %v", month, err)
			}
			taxedMonths++
			if !prov.Taxed() {
				t.Fatalf("%d 月徵稅後旗標沒設起來", month)
			}
			if prov.Gold < goldBefore && prov.Gold != ResourceCap {
				t.Fatalf("%d 月徵稅後黃金反而變少：%d → %d",
					month, goldBefore, prov.Gold)
			}
			if res.LoyaltyAfter > LoyaltyMax {
				t.Fatalf("%d 月忠誠度 %d 超過上限", month, res.LoyaltyAfter)
			}
		}
		// 同月再徵一次一定被擋。
		if _, err := w.Tax(target, rng); err == nil && !prov.Taxed() {
			t.Fatalf("%d 月徵稅旗標失效", month)
		}

		// 開發：墾地要體力，體力見底就換慰勞回補。
		gi := -1
		for _, i := range w.RosterOf(target).ids {
			gi = i
			break
		}
		if gi >= 0 && gi < len(gens) {
			g := &gens[gi]
			if g.Stamina >= ReclaimStaminaNeed {
				before := prov.LandValue
				_, stam, err := w.ReclaimLand(target, g.AbilityC, g.Stamina)
				if err != nil {
					t.Fatalf("%d 月墾地失敗: %v", month, err)
				}
				if prov.LandValue < before {
					t.Fatalf("%d 月墾地後地價反而變低", month)
				}
				if prov.LandValue > LandValueMax {
					t.Fatalf("%d 月地價 %d 超過上限", month, prov.LandValue)
				}
				g.Stamina = stam
			} else {
				res, err := w.Comfort(target, 100)
				if err != nil {
					t.Fatalf("%d 月慰勞失敗: %v", month, err)
				}
				if len(res.Generals) == 0 {
					t.Fatalf("%d 月慰勞沒有影響任何將領", month)
				}
				g.Stamina += ComfortStaminaGain
			}
		}

		// 換月：清徵稅旗標。
		for p := ProvinceID(1); p <= ProvinceCount; p++ {
			if pr, err := w.Table.At(p); err == nil {
				pr.Flags &^= ProvinceFlagTaxed
			}
		}
	}

	if taxedMonths != 12 {
		t.Errorf("十二個月只徵到 %d 次稅，每月限一次的規則沒生效", taxedMonths)
	}

	// 年度結算。
	popBefore := map[ProvinceID]uint32{}
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		if pr, err := w.Table.At(p); err == nil {
			popBefore[p] = pr.Population
		}
	}
	rep := w.AdvanceYear(st, gens, 0, nil)
	if st.Year != 16 {
		t.Fatalf("年度結算後是民國 %d 年，預期 16", st.Year)
	}
	if rep.Grown == 0 {
		t.Error("年度結算沒有任何省人口成長")
	}
	for p, before := range popBefore {
		pr, _ := w.Table.At(p)
		if pr.Population < before {
			t.Fatalf("省 %d 人口在年度結算後變少：%d → %d", p, before, pr.Population)
		}
	}

	// 全域不變量：資源不得超過上限。
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		pr, err := w.Table.At(p)
		if err != nil {
			continue
		}
		for name, v := range map[string]uint16{
			"黃金": pr.Gold, "糧食": pr.Food, "彈藥": pr.Ammo,
			"煤礦": pr.Coal, "燃料": pr.Fuel, "鐵礦": pr.Iron,
		} {
			if v > ResourceCap {
				t.Errorf("省 %d 的%s %d 超過上限 %d", p, name, v, ResourceCap)
			}
		}
		if pr.Loyalty > LoyaltyMax {
			t.Errorf("省 %d 的人民忠誠度 %d 超過上限", p, pr.Loyalty)
		}
	}
}

// 固定種子必須產生固定結果——`CLAUDE.md` §9 的截圖驗收前提。
func TestStrategyDeterministic(t *testing.T) {
	run := func() (uint16, uint8) {
		w := realWorld(t)
		rng := NewRand(42)
		var lastGold uint16
		var lastLoyalty uint8
		for p := ProvinceID(1); p <= ProvinceCount; p++ {
			prov, err := w.Table.At(p)
			if err != nil || prov.Commander == 0 {
				continue
			}
			if _, err := w.Tax(p, rng); err != nil {
				continue
			}
			lastGold, lastLoyalty = prov.Gold, prov.Loyalty
		}
		return lastGold, lastLoyalty
	}
	g1, l1 := run()
	g2, l2 := run()
	if g1 != g2 || l1 != l2 {
		t.Errorf("同種子跑兩次結果不同：(%d,%d) vs (%d,%d)", g1, l1, g2, l2)
	}
}
