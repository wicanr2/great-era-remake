package game

import "testing"

// TestDecayMatchesOriginal 驗證整數版的衰減與原版的浮點算式等價。
//
// 原版：v -= Round(Real(v) / 5.0)。Turbo Pascal 的 Round 是四捨五入
// （.5 進偶數），但 v/5 的小數只可能是 0/.2/.4/.6/.8，碰不到 .5，
// 所以 (v+2)/5 完全等價。
func TestDecayMatchesOriginal(t *testing.T) {
	// 用浮點重算一遍當對照
	roundHalfEven := func(x float64) int {
		i := int(x)
		frac := x - float64(i)
		switch {
		case frac > 0.5:
			return i + 1
		case frac < 0.5:
			return i
		default:
			if i%2 == 0 {
				return i
			}
			return i + 1
		}
	}
	for v := 0; v <= 255; v++ {
		got := Decay(uint8(v))
		want := uint8(v)
		if v > DecayThreshold {
			want = uint8(v - roundHalfEven(float64(v)/5.0))
		}
		if got != want {
			t.Fatalf("Decay(%d) = %d，浮點算式給 %d", v, got, want)
		}
	}
}

// TestDecayThreshold 40 以下不衰減，41 開始衰減。
func TestDecayThreshold(t *testing.T) {
	for _, c := range []struct{ in, want uint8 }{
		{0, 0}, {40, 40}, // 門檻以下原樣
		{41, 33},  // 41 - Round(8.2)=8
		{43, 34},  // 43 - Round(8.6)=9
		{50, 40},  // 50 - 10
		{100, 80}, // 100 - 20
		{255, 204},
	} {
		if got := Decay(c.in); got != c.want {
			t.Errorf("Decay(%d) = %d，預期 %d", c.in, got, c.want)
		}
	}
}

// TestDecayConverges 反覆衰減會收斂到門檻，不會無限跑或反彈。
func TestDecayConverges(t *testing.T) {
	v := uint8(255)
	for i := 0; i < 100; i++ {
		next := Decay(v)
		if next > v {
			t.Fatalf("第 %d 次衰減後變大了：%d → %d", i, v, next)
		}
		if next == v {
			if v > DecayThreshold {
				t.Fatalf("在 %d 卡住了，但門檻是 %d", v, DecayThreshold)
			}
			t.Logf("%d 次之後收斂到 %d", i, v)
			return
		}
		v = next
	}
	t.Errorf("100 次之後還沒收斂，停在 %d", v)
}

// TestBeginTurnResets 回合開始時當前值回到最大值（原版的 +7 = +6）。
func TestBeginTurnResets(t *testing.T) {
	u := CombatUnit{General: 58, Max: 12, Current: 3, Decaying: 100}
	u.BeginTurn()
	if u.Current != 12 || !u.Active {
		t.Errorf("回合開始後 Current=%d Active=%v，預期 12 / true", u.Current, u.Active)
	}
	if u.Decaying != 100 {
		t.Errorf("BeginTurn 不該動到衰減欄位")
	}
	u.EndTurn()
	if u.Decaying != 80 {
		t.Errorf("回合結束後衰減欄位是 %d，預期 80", u.Decaying)
	}

	// 空槽不動
	empty := CombatUnit{Max: 12, Decaying: 100}
	empty.BeginTurn()
	empty.EndTurn()
	if empty.Active || empty.Current != 0 || empty.Decaying != 100 {
		t.Errorf("空槽不該被重置或衰減：%+v", empty)
	}
}

// TestCaptureFollowsOriginal 佔領結算：清除「正在打仗」旗標、司令改成勝方。
//
// 順序照 sub_54DAC：`+32 &= 0xBF` 然後 `+20 = 勝方`。
func TestCaptureFollowsOriginal(t *testing.T) {
	p := Province{Commander: 58, Flags: ProvinceFlagInBattle | ProvinceFlagTaxed}
	if !p.InBattle() {
		t.Fatal("前提不成立：應該在戰鬥中")
	}
	p.Capture(98) // 孫傳芳打下來
	if p.InBattle() {
		t.Error("戰後應該清除「正在打仗」旗標")
	}
	if p.Commander != 98 {
		t.Errorf("司令應改成 98，實得 %d", p.Commander)
	}
	if !p.Taxed() {
		t.Error("不該動到其他位元（已徵稅）")
	}

	// 勝方為 0 時只清旗標、不改司令
	q := Province{Commander: 58, Flags: ProvinceFlagInBattle}
	q.Capture(0)
	if q.InBattle() || q.Commander != 58 {
		t.Errorf("勝方為 0 時應只清旗標：flags=%#x commander=%d", q.Flags, q.Commander)
	}
}

// TestBattleTurnCycle 雙方各 10 個單位，一輪 BeginTurn/EndTurn 都要生效。
func TestBattleTurnCycle(t *testing.T) {
	var b Battle
	for s := range b.Units {
		for i := range b.Units[s] {
			b.Units[s][i] = CombatUnit{
				General: GeneralID(s*100 + i + 1),
				Max:     12, Current: 0, Decaying: 100,
			}
		}
	}
	b.BeginTurn()
	b.EndTurn()
	for s := range b.Units {
		for i, u := range b.Units[s] {
			if u.Current != 12 || !u.Active {
				t.Fatalf("side %d unit %d 沒有重置：%+v", s, i, u)
			}
			if u.Decaying != 80 {
				t.Fatalf("side %d unit %d 的衰減是 %d，預期 80", s, i, u.Decaying)
			}
		}
	}
}

// TestCombatDecay 戰鬥判定後的 10% 衰減，無門檻。
func TestCombatDecay(t *testing.T) {
	for _, c := range []struct{ in, want uint8 }{
		{0, 0},
		{5, 4},    // Round(0.5)=1（half-away-from-zero）
		{10, 9},   // Round(1.0)=1
		{15, 13},  // Round(1.5)=2
		{40, 36},  // 無門檻，40 也會扣
		{100, 90}, // Round(10.0)=10
		{255, 229}, // Round(25.5)=26
	} {
		if got := CombatDecay(c.in); got != c.want {
			t.Errorf("CombatDecay(%d) = %d，預期 %d", c.in, got, c.want)
		}
	}
}

// TestTwoDecayRulesDiffer 兩條衰減規則在同一個值上結果不同——
// 確認沒有把它們寫成同一件事。
func TestTwoDecayRulesDiffer(t *testing.T) {
	// 100：每回合衰減扣 20 → 80；戰鬥衰減扣 10 → 90
	if Decay(100) != 80 || CombatDecay(100) != 90 {
		t.Errorf("Decay(100)=%d CombatDecay(100)=%d，預期 80 / 90",
			Decay(100), CombatDecay(100))
	}
	// 門檻以下：每回合衰減不動，戰鬥衰減照扣
	if Decay(30) != 30 {
		t.Errorf("Decay(30) 應為 30（門檻以下），實得 %d", Decay(30))
	}
	if CombatDecay(30) != 27 {
		t.Errorf("CombatDecay(30) 應為 27（無門檻），實得 %d", CombatDecay(30))
	}
}
