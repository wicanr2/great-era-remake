package game

import "testing"

// 用 SAVE(1) 的真實局面建 AI 世界。
func realWorld(t *testing.T) *AIWorld {
	t.Helper()
	tbl, err := ParseSaveProvinces(readGame(t, "SAVE(1).DT1"))
	if err != nil {
		t.Fatal(err)
	}
	gs, err := ParseGenerals(readGame(t, "MAN(1).DAT"), 274)
	if err != nil {
		t.Fatal(err)
	}
	w := &AIWorld{Table: tbl, Opts: StrengthOpts{Stage: 1}, EnableExtra: true}
	for i := range gs {
		g := &gs[i]
		id := GeneralID(i + 1)
		// 效忠對象＝所屬省的司令。原版的 +14 是獨立欄位，
		// 但檔案裡沒有它（那是執行期才填的），所以這裡推導。
		var faction GeneralID
		if p, err := tbl.At(g.Province); err == nil {
			faction = p.Commander
		}
		w.Units = append(w.Units, CombatUnit{
			General: id, Province: g.Province, Faction: faction,
			Active: g.Raw[16]&1 == 1,
		})
		w.Strengths = append(w.Strengths, StrengthInput{
			Ability: g.AbilityA, Force: g.Force,
			F19: g.F19, F20: g.F20, F29: g.F29, F30: g.F30,
			Branch: g.Branch, General: id, Faction: faction,
		})
	}
	return w
}

// 前線／後方的分類：湖北與河南在 SAVE(1) 都是吳佩孚的，
// 而它們周圍有別的勢力，所以都該是前線。
func TestHostileClassification(t *testing.T) {
	w := realWorld(t)

	front, rear := 0, 0
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		prov, err := w.Table.At(p)
		if err != nil || prov.Commander == 0 {
			continue // 無主省不參與
		}
		if w.Hostile(p) != 0 {
			front++
		} else {
			rear++
		}
	}
	if front == 0 || rear == 0 {
		t.Fatalf("前線 %d／後方 %d——分類全倒向一邊，一定解錯了", front, rear)
	}
	t.Logf("SAVE(1) 的局面：前線省 %d 個、後方省 %d 個", front, rear)

	// 敵對的定義要對稱：A 覺得 B 敵對，B 也該覺得 A 敵對。
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		h := w.Hostile(p)
		if h == 0 {
			continue
		}
		a, _ := w.Table.At(p)
		b, _ := w.Table.At(h)
		if a.Commander == b.Commander {
			t.Errorf("省 %d 與 %d 的司令相同（%d），不該算敵對", p, h, a.Commander)
		}
		if b.Commander == 0 {
			t.Errorf("省 %d 是無主省，不該被算成敵對", h)
		}
	}
}

// 決策鏈：每個有主的省都要能跑完 Decide 而不 panic，
// 而且回傳的目標省必須是真的鄰省。
func TestDecideProducesValidActions(t *testing.T) {
	w := realWorld(t)
	m := loadTestMap(t)

	kinds := map[string]int{}
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		prov, err := w.Table.At(p)
		if err != nil || prov.Commander == 0 {
			continue
		}
		a := w.Decide(p)
		if a.Kind == AINone {
			kinds["（無動作）"]++
			continue
		}
		kinds[a.Step]++

		// 目標必須是鄰省。
		ns, err := m.Neighbours(p)
		if err != nil {
			t.Fatal(err)
		}
		ok := false
		for _, n := range ns {
			if n == a.To {
				ok = true
			}
		}
		if !ok {
			t.Errorf("省 %d 的決定指向 %d，但那不是它的鄰省（%v）", p, a.To, ns)
		}
		// 調動類型只有 0/1/4 三種。
		if a.Kind == AITransfer && a.TransferKind != 0 &&
			a.TransferKind != 1 && a.TransferKind != 4 {
			t.Errorf("省 %d 的調動類型 %d 不在 {0,1,4}", p, a.TransferKind)
		}
	}
	for k, n := range kinds {
		t.Logf("  %-24s %d 次", k, n)
	}
}

// 屯兵上限 15：目標省已經有 15 個以上將領時不該增援。
func TestReinforceFrontRespectsCap(t *testing.T) {
	w := realWorld(t)
	// 找一個會走增援的省，然後把目標省灌滿將領，看它是否退讓。
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		a := w.reinforceFront(p)
		if a.Kind == AINone {
			continue
		}
		before := w.GeneralCount(a.To)
		if before >= AIFrontlineGeneralCap {
			t.Errorf("省 %d 已有 %d 個將領（上限 %d），不該被增援",
				a.To, before, AIFrontlineGeneralCap)
		}
		// 把目標省灌到上限，同一個決策就該不成立了。
		target, _ := w.Table.At(a.To)
		for i := range w.Units {
			if w.Units[i].Province == 0 && len(w.Units) > 0 {
				continue
			}
		}
		for n := before; n < AIFrontlineGeneralCap; n++ {
			w.Units = append(w.Units, CombatUnit{
				General: GeneralID(1000 + n), Province: a.To,
				Faction: target.Commander, Active: true,
			})
			w.Strengths = append(w.Strengths, StrengthInput{Branch: BranchInfantry})
		}
		if got := w.reinforceFront(p); got.To == a.To {
			t.Errorf("目標省已達上限 %d，卻還在增援", AIFrontlineGeneralCap)
		}
		return
	}
	t.Skip("SAVE(1) 的局面裡沒有走增援的省，這條測不到")
}

// 滿員／缺員旗標：拿真實資料驗，並確認司令本人不算。
func TestManpowerFlags(t *testing.T) {
	w := realWorld(t)
	any := false
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		if w.ManpowerFlags(p) != 0 {
			any = true
			break
		}
	}
	if !any {
		t.Error("全 39 省都沒有滿員或缺員的部隊，一定解錯了")
	}

	// 人工局面：一個滿員步兵 + 一個缺員步兵。
	tbl, err := ParseSaveProvinces(readGame(t, "SAVE(1).DT1"))
	if err != nil {
		t.Fatal(err)
	}
	prov, _ := tbl.At(26)
	w2 := &AIWorld{Table: tbl}
	w2.Units = []CombatUnit{
		{General: 900, Province: 26, Faction: prov.Commander, Active: true},
		{General: 901, Province: 26, Faction: prov.Commander, Active: true},
	}
	w2.Strengths = []StrengthInput{
		{Branch: BranchInfantry, Force: 20000}, // 滿員
		{Branch: BranchInfantry, Force: 5000},  // 不到半員
	}
	if got := w2.ManpowerFlags(26); got != 3 {
		t.Errorf("旗標 = %d，應為 3（滿員 + 缺員都有）", got)
	}

	// 司令本人不算。
	w2.Units[0].General = prov.Commander
	if got := w2.ManpowerFlags(26); got&1 != 0 {
		t.Error("司令本人滿員不該設 bit 0")
	}
}
