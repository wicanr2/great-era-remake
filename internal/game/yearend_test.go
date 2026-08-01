package game

import "testing"

// 人口成長：+10%、上限 100 萬、Trunc 捨去。
func TestPopulationGrowth(t *testing.T) {
	cases := []struct {
		name string
		in   uint32
		want uint32
	}{
		{"一般成長", 1_000_000 - 10, 999_990 + 99_999},
		{"整千", 500_000, 550_000},
		{"捨去", 19, 20},     // 19/10 = 1
		{"不足 10 不長", 9, 9}, // 9/10 = 0
		{"到頂不長", 1_000_000, 1_000_000},
		{"超過也不長", 2_000_000, 2_000_000},
		{"零", 0, 0},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := c.in
			if got < PopulationGrowthCap {
				got += got / PopulationGrowthDivisor
			}
			if got != c.want {
				t.Errorf("人口 %d → %d，預期 %d", c.in, got, c.want)
			}
		})
	}
}

// 跨年清旗標：清 bit 1 與 bit 7，保留 bit 2 與 bit 6。
func TestYearEndFlagMask(t *testing.T) {
	all := uint8(0xFF)
	got := all & YearEndFlagMask
	if got&ProvinceFlagTaxed != 0 {
		t.Error("徵稅旗標（bit 7）應該被清掉")
	}
	if got&0x02 != 0 {
		t.Error("bit 1 應該被清掉")
	}
	if got&ProvinceFlagActed == 0 {
		t.Error("bit 2（本回合已處理）不該被動")
	}
	if got&ProvinceFlagInBattle == 0 {
		t.Error("bit 6（交戰中）不該被動")
	}
}

// 年度結算的整體行為，跑在 SAVE(1) 的真實局面上。
func TestAdvanceYear(t *testing.T) {
	w := realWorld(t)
	gens, err := ParseGenerals(readGame(t, "MAN(1).DAT"), 274)
	if err != nil {
		t.Fatal(err)
	}

	// 先把所有省的旗標全開，才驗得出哪些被清掉。
	var popBefore []uint32
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		prov, err := w.Table.At(p)
		if err != nil {
			t.Fatal(err)
		}
		prov.Flags = 0xFF
		popBefore = append(popBefore, prov.Population)
	}

	st := &GameState{Stage: 1, Year: 16, Month: 1}
	rep := w.AdvanceYear(st, gens, 0, nil)

	if st.Year != 17 || rep.Year != 17 {
		t.Fatalf("年份 %d / report %d，預期都是 17", st.Year, rep.Year)
	}
	if rep.Grown == 0 {
		t.Error("沒有任何省的人口成長，與 `docs/re/25` §2 不符")
	}

	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		prov, _ := w.Table.At(p)
		before := popBefore[p-1]
		want := before
		if before < PopulationGrowthCap {
			want += before / PopulationGrowthDivisor
		}
		if prov.Population != want {
			t.Fatalf("省 %d 人口 %d → %d，預期 %d", p, before, prov.Population, want)
		}
		if prov.Taxed() {
			t.Fatalf("省 %d 的徵稅旗標沒被清掉", p)
		}
	}
}

// 民國 17 年的東北易幟：只動效忠 166 且在場的將領。
func TestManchurianFlagEvent(t *testing.T) {
	w := realWorld(t)
	gens, err := ParseGenerals(readGame(t, "MAN(1).DAT"), 274)
	if err != nil {
		t.Fatal(err)
	}
	// 造三筆：效忠 166 在場、效忠 166 不在場、效忠別人。
	w.Units = []CombatUnit{
		{Faction: ManchurianFlagEventLeader, Active: true},
		{Faction: ManchurianFlagEventLeader, Active: false},
		{Faction: 1, Active: true},
	}
	gens = gens[:3]
	for i := range gens {
		gens[i].FactionName = 99
	}

	st := &GameState{Stage: 1, Year: ManchurianFlagEventYear - 1}
	rep := w.AdvanceYear(st, gens, 0, nil)

	if rep.Reflagged != 1 {
		t.Fatalf("改到 %d 個將領，預期 1", rep.Reflagged)
	}
	if gens[0].FactionName != ManchurianFlagEventFaction {
		t.Errorf("在場的將領番號沒改：%d", gens[0].FactionName)
	}
	if gens[1].FactionName != 99 || gens[2].FactionName != 99 {
		t.Error("不該被改的將領被改了")
	}
}

// 歷史檢查沒過 → 玩家勢力全將領帶兵能力減半（不看是否在場）。
func TestHistoryCheckPenalty(t *testing.T) {
	w := realWorld(t)
	gens, err := ParseGenerals(readGame(t, "MAN(1).DAT"), 274)
	if err != nil {
		t.Fatal(err)
	}
	w.Units = []CombatUnit{
		{Faction: 7, Active: true},
		{Faction: 7, Active: false},
		{Faction: 8, Active: true},
	}
	gens = gens[:3]
	for i := range gens {
		gens[i].AbilityA = 80
	}

	st := &GameState{Stage: 2, Year: 20} // 20 < 28，門檻內
	rep := w.AdvanceYear(st, gens, 7, func(uint8) bool { return false })

	if rep.Penalised != 2 {
		t.Fatalf("砍了 %d 個將領，預期 2（在場與否都砍）", rep.Penalised)
	}
	if gens[0].AbilityA != 40 || gens[1].AbilityA != 40 {
		t.Errorf("帶兵能力沒減半：%d / %d", gens[0].AbilityA, gens[1].AbilityA)
	}
	if gens[2].AbilityA != 80 {
		t.Errorf("別的勢力不該被砍：%d", gens[2].AbilityA)
	}
}

// 歷史檢查過了、或年份已超過門檻，都不該懲罰。
func TestHistoryCheckSkipped(t *testing.T) {
	for _, c := range []struct {
		name   string
		year   uint8
		passed func(uint8) bool
	}{
		{"條件達成", 20, func(uint8) bool { return true }},
		{"年份已超過門檻", 30, func(uint8) bool { return false }},
		{"沒有述詞視同通過", 20, nil},
	} {
		t.Run(c.name, func(t *testing.T) {
			w := realWorld(t)
			gens, err := ParseGenerals(readGame(t, "MAN(1).DAT"), 274)
			if err != nil {
				t.Fatal(err)
			}
			w.Units = []CombatUnit{{Faction: 7, Active: true}}
			gens = gens[:1]
			gens[0].AbilityA = 80

			st := &GameState{Stage: 2, Year: c.year - 1}
			rep := w.AdvanceYear(st, gens, 7, c.passed)
			if rep.Penalised != 0 || gens[0].AbilityA != 80 {
				t.Errorf("不該懲罰，卻砍了 %d 個（能力 %d）", rep.Penalised, gens[0].AbilityA)
			}
		})
	}
}

// 挖金礦：保底 200、夾到 ResourceCap、體力扣 5。
func TestMineGold(t *testing.T) {
	w := realWorld(t)
	prov, err := w.Table.At(1)
	if err != nil {
		t.Fatal(err)
	}

	t.Run("體力不足", func(t *testing.T) {
		if _, _, err := w.MineGold(1, 5, 100, MineStaminaNeed-1, NewRand(1)); err == nil {
			t.Error("體力低於門檻應該報錯")
		}
	})

	t.Run("保底與扣體力", func(t *testing.T) {
		prov.Gold = 0
		got, stam, err := w.MineGold(1, 5, 100, 50, NewRand(1))
		if err != nil {
			t.Fatal(err)
		}
		if got < MineYieldBase {
			t.Errorf("產出 %d 低於保底 %d", got, MineYieldBase)
		}
		if (got-MineYieldBase)%MineYieldScale != 0 {
			t.Errorf("產出 %d 不是 200 + 50k", got)
		}
		if stam != 50-MineStaminaCost {
			t.Errorf("體力 %d，預期 %d", stam, 50-MineStaminaCost)
		}
	})

	t.Run("夾到上限", func(t *testing.T) {
		prov.Gold = ResourceCap - 10
		got, _, err := w.MineGold(1, 5, 100, 50, NewRand(1))
		if err != nil {
			t.Fatal(err)
		}
		if got != 10 || prov.Gold != ResourceCap {
			t.Errorf("入帳 %d、黃金 %d，預期 10 / %d", got, prov.Gold, ResourceCap)
		}
	})
}
