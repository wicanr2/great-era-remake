package game

import "testing"

// SupportingNeighbours 只是把 `sub_534FF` 轉給 `ReinforcementSources`
// （那支的行為由 `province_test.go` 的 `TestReinforcementSources` 顧著），
// 所以這裡只驗**轉接層**：兩者回一樣的東西，以及決策鏈那兩個守門條件。
func TestSupportingNeighboursDelegates(t *testing.T) {
	tbl, err := ParseTownFile(readGame(t, "TOWN(1).DAT"))
	if err != nil {
		t.Fatal(err)
	}
	const at = ProvinceID(26) // 湖北
	p, err := tbl.At(at)
	if err != nil {
		t.Fatal(err)
	}

	want, err := tbl.ReinforcementSources(at, p.Commander, nil)
	if err != nil {
		t.Fatal(err)
	}
	got := SupportingNeighbours(tbl, at, p.Commander, nil)
	if len(got) != len(want) {
		t.Fatalf("該與 ReinforcementSources 一致：%v vs %v", got, want)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("該與 ReinforcementSources 一致：%v vs %v", got, want)
		}
	}
	if HasBattleSupport(tbl, at, p.Commander, nil) != (len(want) > 0) {
		t.Error("HasBattleSupport 該等於「清單非空」")
	}
}

func TestSupportingNeighboursGuards(t *testing.T) {
	tbl, err := ParseTownFile(readGame(t, "TOWN(1).DAT"))
	if err != nil {
		t.Fatal(err)
	}
	// leader 為 0 一律沒有支援——原版的 `+14` 是 1-based，0 代表沒有效忠對象。
	if HasBattleSupport(tbl, 26, 0, nil) {
		t.Error("leader 為 0 該回 false")
	}
	if SupportingNeighbours(nil, 26, 58, nil) != nil {
		t.Error("表為 nil 該回 nil")
	}
	// 省編號無效時 ReinforcementSources 會回錯誤，這裡吞掉回 nil。
	if SupportingNeighbours(tbl, 0, 58, nil) != nil {
		t.Error("省編號無效該回 nil")
	}
}
