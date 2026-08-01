package game

import (
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

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
			F19: g.F19, F20: g.F20, F29: g.Stamina, F30: g.F30,
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
		// 決策鏈實際會產生的調動類型：
		//   0/1  一般調動（sub_150FC）
		//   2    補窮省（sub_15925）
		//   3    分兵（sub_15667）
		//   4    撤回後方（sub_1541E）、增援前線、補給調動
		switch a.TransferKind {
		case 0, 1, 2, 3, 4:
		default:
			if a.Kind == AITransfer {
				t.Errorf("省 %d 的調動類型 %d 不在 0..4", p, a.TransferKind)
			}
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

// 守門的作用是「讓一般調動不要吃掉所有決策」。
//
// ⛔ 這條測試原本盯的是「攻打次數 > 0」，前提錯了——
// **政略決策鏈六步全是調動，沒有一步是攻打**（`docs/re/12`）。
// 現在盯的是真正該盯的事：一般調動之後的步驟仍然有機會執行。
//
// 沒有守門時一般調動 18 次、把後面全吃掉；有守門之後它剩個位數。
// 改壞了會靜靜消音，所以要有測試盯著。
func TestGatesLeaveLaterStepsReachable(t *testing.T) {
	w := realWorld(t)
	steps := map[string]int{}
	acted := 0
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		prov, err := w.Table.At(p)
		if err != nil || prov.Commander == 0 {
			continue
		}
		a := w.Decide(p)
		if a.Kind == AINone {
			continue
		}
		acted++
		steps[a.Step]++
	}
	if acted == 0 {
		t.Fatal("沒有任何省做出決定")
	}
	if n := steps["sub_150FC 一般調動"]; n > acted/2 {
		t.Errorf("一般調動佔了 %d/%d 個決定，守門可能失效了", n, acted)
	}
	if len(steps) < 2 {
		t.Errorf("只有 %d 種決策被觸發，鏈上其他步驟都不可達", len(steps))
	}
	for k, n := range steps {
		t.Logf("  %-24s %d 次", k, n)
	}
}

// 那個疑似原版 bug 要照抄：省編號當將領 ID 索引。
// 這條測試把行為釘住，免得有人「順手修好」而不自知改了遊戲規則。
func TestPortsOriginalIndexingBug(t *testing.T) {
	w := &AIWorld{
		Units: []CombatUnit{
			{General: 1, Faction: 100}, // 將領 1
			{General: 2, Faction: 200}, // 將領 2
			{General: 3, Faction: 300}, // 將領 3
		},
	}
	// 傳「省 2」，原版讀的是**將領 2** 的效忠勢力（200），
	// 而不是省 2 的司令。
	if !w.generalsLoyalTo(2, 200) {
		t.Error("應該讀到將領 2 的效忠勢力 200——原版就是這樣索引的")
	}
	if w.generalsLoyalTo(2, 100) {
		t.Error("不該讀到將領 1")
	}
	// 超出將領表就回 false，不 panic。
	if w.generalsLoyalTo(99, 100) {
		t.Error("省編號超出將領表時應回 false")
	}
}

// sub_5B76E 的鄰接檢查：與地圖導出的鄰接表比對。
//
// 兩邊是不同的資料路徑（一個讀省份記錄 +22 的鄰省表、一個從 WARPOS 導），
// **已知有三筆差異**，全是隔海／隔境那一段（`SPEC-03` §4）：
//
//	福建(24) 的鄰省表缺 臺灣(37)
//	雲南(34) 缺 緬甸(39)
//	廣東(36) 缺 海南島(38)
//
// 把它們釘成明確的例外而不是放寬條件——差異變多或變少都要被抓到。
func TestAdjacentProvincesMatchesMap(t *testing.T) {
	known := map[[2]ProvinceID]bool{
		{24, 37}: true, // 福建 → 臺灣
		{34, 39}: true, // 雲南 → 緬甸
		{36, 38}: true, // 廣東 → 海南島
	}
	seen := map[[2]ProvinceID]bool{}

	w := realWorld(t)
	m := loadTestMap(t)
	checked := 0
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		ns, err := m.Neighbours(p)
		if err != nil {
			t.Fatal(err)
		}
		for _, n := range ns {
			if !w.adjacentProvinces(n, p) {
				k := [2]ProvinceID{p, n}
				if !known[k] {
					t.Errorf("地圖說 %d 與 %d 相鄰，鄰省表卻說不是（不在已知例外裡）", p, n)
				}
				seen[k] = true
			}
			checked++
		}
		// 不相鄰的也要對：隨便挑一個不在鄰省表裡的省。
		for q := ProvinceID(1); q <= ProvinceCount; q++ {
			inList := false
			for _, n := range ns {
				if n == q {
					inList = true
				}
			}
			if !inList && q != p && w.adjacentProvinces(q, p) {
				t.Errorf("鄰省表說 %d 與 %d 相鄰，地圖卻說不是", p, q)
			}
		}
	}
	for k := range known {
		if !seen[k] {
			t.Errorf("已知例外 %d→%d 消失了——資料或解讀變了，重看 SPEC-03 §4", k[0], k[1])
		}
	}
	t.Logf("比對 %d 組鄰接，已知例外 %d 筆全部出現", checked, len(known))
}

// 五個特殊省的清單不能悄悄變動——它與 WARPOS 的哨兵 40 完全相同，
// 那個巧合本身就是證據（70-ai.md §6m）。
func TestSpecialProvincesMatchWarposSentinel(t *testing.T) {
	want := []ProvinceID{7, 11, 20, 21, 36}
	if len(SpecialProvinces) != len(want) {
		t.Fatalf("特殊省清單長度 = %d，應為 %d", len(SpecialProvinces), len(want))
	}
	for i, p := range want {
		if SpecialProvinces[i] != p {
			t.Errorf("特殊省[%d] = %d，應為 %d", i, SpecialProvinces[i], p)
		}
	}

	// 與 WARPOS 的哨兵 40 對照——這是清單的出處旁證。
	grids, err := assetsProvinceGrids(t)
	if err != nil {
		t.Fatal(err)
	}
	var withSentinel []ProvinceID
	for k := 0; k < ProvinceCount; k++ {
		for _, v := range grids[k] {
			if v == 40 {
				withSentinel = append(withSentinel, ProvinceID(k+1))
				break
			}
		}
	}
	if len(withSentinel) != len(want) {
		t.Fatalf("WARPOS 帶哨兵 40 的省有 %d 個，應為 %d：%v",
			len(withSentinel), len(want), withSentinel)
	}
	for i, p := range want {
		if withSentinel[i] != p {
			t.Errorf("哨兵 40 的省[%d] = %d，應為 %d", i, withSentinel[i], p)
		}
	}
}

// assetsProvinceGrids 讀 WARPOS.DAT 的原始格子圖。
func assetsProvinceGrids(t *testing.T) (*assets.ProvinceGrids, error) {
	t.Helper()
	return assets.ParseProvinceGrids(readGame(t, "WARPOS.DAT"))
}
