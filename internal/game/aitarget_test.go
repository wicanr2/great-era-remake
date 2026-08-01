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
