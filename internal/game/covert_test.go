package game

import "testing"

// 「剩 70%」是先除後乘，不是先乘後除——順序會改變結果。
func TestGuerrillaDecayOrder(t *testing.T) {
	cases := []struct{ in, want uint16 }{
		{1000, 700},
		{9, 0},   // 9÷10 = 0，×7 還是 0
		{19, 7},  // 19÷10 = 1，×7 = 7（先乘後除會是 13）
		{55, 35}, // 55÷10 = 5，×7 = 35（先乘後除會是 38）
		{0, 0},
	}
	for _, c := range cases {
		if got := decay(c.in); got != c.want {
			t.Errorf("decay(%d) = %d，原版是 %d（先除後乘）", c.in, got, c.want)
		}
	}
}

// 游擊隊的成功率是 4/10，用固定種子掃一段序列驗分佈。
func TestGuerrillaSuccessRate(t *testing.T) {
	rng := NewRand(12345)
	const n = 20000
	success := 0
	for i := 0; i < n; i++ {
		if rng.Int(GuerrillaRollRange) <= GuerrillaSuccessMax {
			success++
		}
	}
	rate := float64(success) / n
	if rate < 0.37 || rate > 0.43 {
		t.Errorf("游擊隊成功率 %.3f，原版是 4/10", rate)
	}
}

// 學潮的成功率是 2/10。
func TestStudentProtestRate(t *testing.T) {
	rng := NewRand(999)
	const n = 20000
	success := 0
	for i := 0; i < n; i++ {
		if rng.Int(StudentRollRange) <= StudentSuccessMax {
			success++
		}
	}
	rate := float64(success) / n
	if rate < 0.17 || rate > 0.23 {
		t.Errorf("學潮成功率 %.3f，原版是 2/10", rate)
	}
}

// 游擊隊成功時：黃金／糧食／彈藥剩 70%，燃料煤鐵不動。
func TestGuerrillaEffect(t *testing.T) {
	w := realWorld(t)
	src, _ := w.Table.At(1)
	dst, _ := w.Table.At(2)
	src.Gold = 5000
	dst.Gold, dst.Food, dst.Ammo = 1000, 2000, 3000
	dst.Fuel, dst.Coal, dst.Iron = 4000, 5000, 6000
	dst.Arsenals = 3

	// 找一個會成功的種子。
	var res CovertResult
	var err error
	for seed := uint32(1); seed < 50; seed++ {
		w2 := realWorld(t)
		p1, _ := w2.Table.At(1)
		p1.Gold = 5000
		rng := NewRand(seed)
		if r, e := w2.SendGuerrillas(1, 2, 100, rng); e == nil && r.Success {
			// 用同一顆種子在真正的世界跑一次。
			res, err = w.SendGuerrillas(1, 2, 100, NewRand(seed))
			break
		}
	}
	if err != nil {
		t.Fatal(err)
	}
	if !res.Success {
		t.Skip("50 顆種子裡沒找到成功的，跳過")
	}

	if dst.Gold != 700 || dst.Food != 1400 || dst.Ammo != 2100 {
		t.Errorf("破壞後黃金/糧食/彈藥 = %d/%d/%d，預期 700/1400/2100",
			dst.Gold, dst.Food, dst.Ammo)
	}
	if dst.Fuel != 4000 || dst.Coal != 5000 || dst.Iron != 6000 {
		t.Errorf("燃料/煤礦/鐵礦被動到了：%d/%d/%d，原版不碰這三種",
			dst.Fuel, dst.Coal, dst.Iron)
	}
	if src.Gold != 4900 {
		t.Errorf("來源省黃金 %d，預期 4900（扣 100）", src.Gold)
	}
}

// 失敗時目標省完全不動，但錢照扣。
func TestGuerrillaFailureStillCosts(t *testing.T) {
	w := realWorld(t)
	src, _ := w.Table.At(1)
	dst, _ := w.Table.At(2)
	src.Gold = 5000
	dst.Gold, dst.Food, dst.Ammo = 1000, 2000, 3000

	for seed := uint32(1); seed < 50; seed++ {
		w2 := realWorld(t)
		p1, _ := w2.Table.At(1)
		p1.Gold = 5000
		if r, e := w2.SendGuerrillas(1, 2, 100, NewRand(seed)); e == nil && !r.Success {
			res, err := w.SendGuerrillas(1, 2, 100, NewRand(seed))
			if err != nil {
				t.Fatal(err)
			}
			if res.Success {
				t.Fatal("同種子跑出不同結果")
			}
			if dst.Gold != 1000 || dst.Food != 2000 || dst.Ammo != 3000 {
				t.Error("失敗卻動到目標省的資源")
			}
			if src.Gold != 4900 {
				t.Errorf("來源省黃金 %d，預期 4900——失敗也要付錢", src.Gold)
			}
			return
		}
	}
	t.Skip("50 顆種子裡沒找到失敗的，跳過")
}

// 資金不足時擋下來，而且不扣錢。
func TestCovertInsufficientFunds(t *testing.T) {
	w := realWorld(t)
	src, _ := w.Table.At(1)
	src.Gold = 100

	if _, err := w.SendGuerrillas(1, 2, 500, NewRand(1)); err == nil {
		t.Error("黃金 100 花 500 應該報錯")
	}
	if src.Gold != 100 {
		t.Errorf("報錯後黃金變成 %d，不該扣", src.Gold)
	}
	if _, err := w.IncitStudentProtest(1, NewRand(1)); err == nil {
		t.Error("黃金 100 鼓動學潮（要 1500）應該報錯")
	}
}

// 學潮成敗都扣 1500。
func TestStudentProtestAlwaysCosts(t *testing.T) {
	for seed := uint32(1); seed < 6; seed++ {
		w := realWorld(t)
		src, _ := w.Table.At(1)
		src.Gold = 5000
		res, err := w.IncitStudentProtest(1, NewRand(seed))
		if err != nil {
			t.Fatal(err)
		}
		if src.Gold != 5000-StudentProtestCost {
			t.Errorf("種子 %d：黃金 %d，預期 %d（成敗都扣）",
				seed, src.Gold, 5000-StudentProtestCost)
		}
		if res.Cost != StudentProtestCost {
			t.Errorf("回報的成本 %d，預期 %d", res.Cost, StudentProtestCost)
		}
	}
}
