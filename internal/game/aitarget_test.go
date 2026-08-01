package game

import "testing"

// 我方鄰省計數：只認司令相同的，海洋與填充跳過。
func TestFriendlyNeighbourCount(t *testing.T) {
	w := realWorld(t)
	me, err := w.Table.At(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(me.Neighbours) < 3 {
		t.Skip("省 1 的鄰省太少，測不了")
	}
	for i, n := range me.Neighbours {
		np, _ := w.Table.At(n)
		switch i {
		case 0, 1:
			np.Commander = 7 // 我方
		default:
			np.Commander = 8 // 敵方
		}
	}
	if got := w.FriendlyNeighbourCount(1, 7); got != 2 {
		t.Errorf("我方鄰省數 %d，預期 2", got)
	}
	// 換一個勢力去問，答案要跟著變。
	if got := w.FriendlyNeighbourCount(1, 8); got != len(me.Neighbours)-2 {
		t.Errorf("勢力 8 的鄰省數 %d，預期 %d", got, len(me.Neighbours)-2)
	}
}

// 門檻是 ≥ 2：只接壤一個我方省的候選不算數。
//
// 候選固定用**省 2 與省 4**：它們互不相鄰（省 2 的鄰省 [1 5 8 9]、
// 省 4 的 [1 3 6]，只共用省 1），所以設定其中一個的鄰省不會污染另一個的計數。
func TestMostConnectedTargetThreshold(t *testing.T) {
	w, a, b := hostileWorld(t)

	// 兩個候選各只有一個我方鄰省 → 達不到門檻，回 0。
	setCommander(t, w, 8, 7) // 省 2 的鄰省
	setCommander(t, w, 3, 7) // 省 4 的鄰省
	if got := w.MostConnectedTarget([]ProvinceID{a, b}, 7); got != 0 {
		t.Errorf("最高只有 1 個我方鄰省，應該回 0，實際回 %d", got)
	}

	// 讓**排在後面的** b 有兩個 → 挑 b。
	// 用後面那個才證明它是挑最多的，而不是挑第一個。
	setCommander(t, w, 6, 7) // 省 4 的另一個鄰省
	if got := w.MostConnectedTarget([]ProvinceID{a, b}, 7); got != b {
		t.Errorf("挑了 %d，應該挑我方鄰省最多的 %d", got, b)
	}
}

// 並列時保留較前面的候選（原版用 `jbe` 跳過相等）。
func TestMostConnectedTargetKeepsFirstOnTie(t *testing.T) {
	w, a, b := hostileWorld(t)
	for _, n := range []ProvinceID{8, 9} { // 省 2 的鄰省
		setCommander(t, w, n, 7)
	}
	for _, n := range []ProvinceID{3, 6} { // 省 4 的鄰省
		setCommander(t, w, n, 7)
	}
	if got := w.MostConnectedTarget([]ProvinceID{a, b}, 7); got != a {
		t.Errorf("並列時挑了 %d，應該保留較前面的 %d", got, a)
	}
	// 順序反過來答案也要反過來——證明它看的是候選順序，不是省編號。
	if got := w.MostConnectedTarget([]ProvinceID{b, a}, 7); got != b {
		t.Errorf("候選順序反轉後挑了 %d，應該是 %d", got, b)
	}
}

// hostileWorld 造一個「全部省份都屬於敵方 8、黃金歸零」的世界，
// 並回傳兩個互不相鄰的候選省（2 與 4）。
func hostileWorld(t *testing.T) (*AIWorld, ProvinceID, ProvinceID) {
	t.Helper()
	w := realWorld(t)
	for i := ProvinceID(1); ; i++ {
		p, err := w.Table.At(i)
		if err != nil {
			break
		}
		p.Commander = 8
		p.Gold = 0
	}
	a, b := ProvinceID(2), ProvinceID(4)
	pa, err := w.Table.At(a)
	if err != nil {
		t.Fatal(err)
	}
	pb, err := w.Table.At(b)
	if err != nil {
		t.Fatal(err)
	}
	// 前提檢查：兩個候選必須互不相鄰，否則整組測試的隔離假設不成立。
	for _, n := range pa.Neighbours {
		if n == b {
			t.Fatalf("省 %d 與省 %d 相鄰，測試前提不成立", a, b)
		}
	}
	_ = pb
	return w, a, b
}

func setCommander(t *testing.T, w *AIWorld, p ProvinceID, c GeneralID) {
	t.Helper()
	prov, err := w.Table.At(p)
	if err != nil {
		t.Fatal(err)
	}
	prov.Commander = c
}

// 三層優先序：被包圍 > 我方接壤最多 > 最富。
//
// 一樣用 `hostileWorld` 的省 2／省 4，避開兩個候選互為鄰省時
// 「設定 A 的鄰省連帶改到 B」的污染。
func TestChainATargetPriority(t *testing.T) {
	t.Run("被包圍的優先於最富的", func(t *testing.T) {
		w, a, b := hostileWorld(t)
		// a 換一個勢力 → 它的鄰省（全是 8）就全都成了敵人 → 被包圍。
		setCommander(t, w, a, 9)
		// b 留在勢力 8，鄰省也全是 8（同勢力，不算敵人）→ 不被包圍，但很有錢。
		pb, err := w.Table.At(b)
		if err != nil {
			t.Fatal(err)
		}
		pb.Gold = 60000

		if !w.Encircled(a) {
			t.Fatal("測試前提不成立：a 應該被判為被包圍")
		}
		if w.Encircled(b) {
			t.Fatal("測試前提不成立：b 不該被判為被包圍")
		}
		// b 排在前面而且更有錢，仍應該挑被包圍的 a。
		if got := w.ChainATarget([]ProvinceID{b, a}, 7, true); got != a {
			t.Errorf("挑了 %d，被包圍的 %d 應該優先", got, a)
		}
	})

	t.Run("進階模式關掉就跳過中間那層", func(t *testing.T) {
		w, a, b := hostileWorld(t)
		// a 有兩個我方鄰省（第 2 層挑它），b 比較有錢（第 3 層挑它）。
		setCommander(t, w, a, 7)
		setCommander(t, w, 8, 7)
		setCommander(t, w, 9, 7)
		pa, err := w.Table.At(a)
		if err != nil {
			t.Fatal(err)
		}
		pb, err := w.Table.At(b)
		if err != nil {
			t.Fatal(err)
		}
		pa.Gold, pb.Gold = 100, 60000

		if w.Encircled(a) || w.Encircled(b) {
			t.Fatal("測試前提不成立：兩個候選都不該被判為被包圍")
		}
		if got := w.ChainATarget([]ProvinceID{a, b}, 7, true); got != a {
			t.Errorf("進階模式挑了 %d，應該挑我方接壤最多的 %d", got, a)
		}
		if got := w.ChainATarget([]ProvinceID{a, b}, 7, false); got != b {
			t.Errorf("關掉進階模式挑了 %d，應該退回挑最富的 %d", got, b)
		}
	})
}

// 突圍門檻：我方省份數 ≥ 3。
func TestBreakoutNeedsThreeProvinces(t *testing.T) {
	w, _, _ := hostileWorld(t) // 全部屬於勢力 8
	// 勢力 7 一個省都沒有 → 一定挑不出來。
	if got := w.BreakoutTarget(1, 7); got != 0 {
		t.Errorf("我方 0 個省還挑得出 %d，應該回 0", got)
	}
	// 給勢力 7 兩個省，仍不到門檻。
	setCommander(t, w, 1, 7)
	setCommander(t, w, 2, 7)
	if w.FactionProvinceCount(7) != 2 {
		t.Fatalf("前提不成立：勢力 7 應該有 2 個省，實際 %d", w.FactionProvinceCount(7))
	}
	if got := w.BreakoutTarget(1, 7); got != 0 {
		t.Errorf("我方 2 個省還挑得出 %d，門檻是 %d", got, AIBreakoutMinProvinces)
	}
}

// 突圍目標挑「它自己也接壤到另一個我方省」的鄰省。
//
// 用真實地圖：省 1 鄰省 [2 3 4 5 6 8]，省 3 鄰省 [1 4 8]。
// 讓省 8 屬於我方，則從省 1 出發時省 3 合格（省 3 接壤省 8，而省 8 != 省 1）。
func TestBreakoutTargetPicksConnectedNeighbour(t *testing.T) {
	w, _, _ := hostileWorld(t)
	// 我方（勢力 7）要有 ≥ 3 個省才過門檻。
	for _, p := range []ProvinceID{1, 8, 20, 21} {
		setCommander(t, w, p, 7)
	}
	got := w.BreakoutTarget(1, 7)
	if got == 0 {
		t.Fatal("應該挑得出突圍目標")
	}
	// 驗證挑出來的確實滿足條件，而不是碰巧。
	p1, err := w.Table.At(got)
	if err != nil {
		t.Fatal(err)
	}
	ok := false
	for _, nb2 := range p1.Neighbours {
		p2, err := w.Table.At(nb2)
		if err != nil {
			continue
		}
		if nb2 != 1 && p2.Commander == 7 {
			ok = true
		}
	}
	if !ok {
		t.Errorf("挑了省 %d，但它沒有接壤到當前省以外的我方省（鄰省 %v）",
			got, p1.Neighbours)
	}
	// 我方只剩當前省一個鄰居可連時（把 8 拿掉），答案要變。
	setCommander(t, w, 8, 8)
	if again := w.BreakoutTarget(1, 7); again == got {
		t.Errorf("拿掉省 8 之後仍挑 %d，代表判斷沒真的看第二層鄰省", again)
	}
}

// 兩份跳過清單不同：決策鏈 A 三省、步驟 6 只有兩省（緬甸不在內）。
func TestSkippedProvinceLists(t *testing.T) {
	// bit 1 沒設就不跳過任何省。
	for _, p := range AIChainASkippedProvinces {
		if AIChainASkipsProvince(p, false) {
			t.Errorf("byte_6FFCA bit 1 未設時不該跳過省 %d", p)
		}
		if !AIChainASkipsProvince(p, true) {
			t.Errorf("bit 1 設起來時應該跳過省 %d", p)
		}
	}
	// 39 緬甸在決策鏈 A 的清單裡，但不在步驟 6 的清單裡。
	if !AIChainASkipsProvince(39, true) {
		t.Error("緬甸（39）應該在決策鏈 A 的跳過清單裡")
	}
	if AIBreakoutSkipsProvince(39) {
		t.Error("緬甸（39）不該在步驟 6 的排除清單裡——它與雲南陸路相連")
	}
	for _, p := range []ProvinceID{37, 38} {
		if !AIBreakoutSkipsProvince(p) {
			t.Errorf("離島省 %d 應該被步驟 6 排除", p)
		}
	}
	if AIBreakoutSkipsProvince(1) {
		t.Error("省 1 不該被排除")
	}
}

// 排序方向：mode 1 升序（最弱在前）、mode 2 降序（最強在前）。
func TestSortProvincesByStrength(t *testing.T) {
	str := map[ProvinceID]int{1: 300, 2: 100, 3: 200}
	get := func(p ProvinceID) int { return str[p] }

	asc := []ProvinceID{1, 2, 3}
	SortProvincesByStrength(asc, true, get)
	if asc[0] != 2 || asc[len(asc)-1] != 1 {
		t.Errorf("升序結果 %v，最弱的省 2 應該在最前、最強的省 1 在最後", asc)
	}

	desc := []ProvinceID{1, 2, 3}
	SortProvincesByStrength(desc, false, get)
	if desc[0] != 1 || desc[len(desc)-1] != 2 {
		t.Errorf("降序結果 %v，最強的省 1 應該在最前", desc)
	}

	// 邊界：0 個與 1 個元素不能 panic。
	SortProvincesByStrength(nil, true, get)
	one := []ProvinceID{5}
	SortProvincesByStrength(one, true, get)
	if len(one) != 1 || one[0] != 5 {
		t.Errorf("單元素被動到了：%v", one)
	}
}

// 相等時不交換（原版 `jbe`／`jnb` 都跳過）。
func TestSortProvincesByStrengthTieKeepsOrder(t *testing.T) {
	get := func(ProvinceID) int { return 42 } // 全部一樣
	for _, asc := range []bool{true, false} {
		list := []ProvinceID{3, 1, 2}
		SortProvincesByStrength(list, asc, get)
		if list[0] != 3 || list[1] != 1 || list[2] != 2 {
			t.Errorf("ascending=%v 全部並列時順序被改成 %v，應該保持 [3 1 2]",
				asc, list)
		}
	}
}

// 步驟 5 挑的是戰力最小的候選。
func TestWeakestNeighbourTarget(t *testing.T) {
	w := realWorld(t)
	w.Units = nil
	w.Strengths = nil
	// 省 5 放一個部隊，省 3、省 4 都空 → 省 5 戰力最高。
	// `ProvinceStrength` 要求 `General != 0`（空槽不算），不能省。
	w.Units = append(w.Units, CombatUnit{General: 1, Province: 5, Active: true})
	w.Strengths = append(w.Strengths, StrengthInput{
		Ability: 80, Force: 20000, F19: 10, F20: 10, F29: 100, F30: 80, Branch: 1,
	})
	if got := w.WeakestNeighbourTarget([]ProvinceID{5, 3, 4}); got == 5 {
		t.Errorf("挑了戰力最高的省 5，應該挑空的省 3 或 4")
	}
	if got := w.WeakestNeighbourTarget(nil); got != 0 {
		t.Errorf("沒有候選時應該回 0，實際回 %d", got)
	}
	// 不可以改動呼叫端的切片。
	cands := []ProvinceID{5, 3, 4}
	w.WeakestNeighbourTarget(cands)
	if cands[0] != 5 {
		t.Errorf("呼叫端的候選切片被就地改動了：%v", cands)
	}
}
