package game

import "testing"

func TestSupportingNeighboursFiltersCorrectly(t *testing.T) {
	// §47：只收「屬於同一勢力、+32 bit 6 沒立起來」的鄰省。
	tbl, err := ParseTownFile(readGame(t, "TOWN(1).DAT"))
	if err != nil {
		t.Fatal(err)
	}

	const at = ProvinceID(19) // 河南
	p, err := tbl.At(at)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Neighbours) < 2 {
		t.Skipf("省 %d 只有 %d 個鄰省，測不出篩選", at, len(p.Neighbours))
	}

	// 先把**所有**鄰省歸零，再指定第一個屬於 leader。
	//
	// ⚠️ 不歸零會踩到坑：原始檔裡剛好有別的鄰省司令就是 leader，
	// 測試會多收一個而失敗——那不是實作錯，是測試沒控制輸入。
	const leader = GeneralID(58)
	for _, n := range p.Neighbours {
		np, err := tbl.At(n)
		if err != nil {
			t.Fatal(err)
		}
		np.Commander, np.Flags = 0, 0
	}
	n0, _ := tbl.At(p.Neighbours[0])
	n1, _ := tbl.At(p.Neighbours[1])
	n0.Commander = leader
	n1.Commander = leader + 1

	got := SupportingNeighbours(tbl, at, leader)
	if len(got) != 1 || got[0] != p.Neighbours[0] {
		t.Errorf("只有第一個鄰省屬於 leader，實際收到 %v", got)
	}
	if !HasBattleSupport(tbl, at, leader) {
		t.Error("有一個支援省，HasBattleSupport 該回 true")
	}

	// bit 6 立起來就被跳過。
	n0.Flags |= ProvinceFlagNoSupport
	if got := SupportingNeighbours(tbl, at, leader); len(got) != 0 {
		t.Errorf("+32 bit 6 立起來該被跳過，實際 %v", got)
	}
	if HasBattleSupport(tbl, at, leader) {
		t.Error("沒有支援省了，HasBattleSupport 該回 false")
	}

	// 無主省（司令 0）不算支援。
	n0.Flags = 0
	n0.Commander = 0
	if HasBattleSupport(tbl, at, leader) {
		t.Error("無主省不該算支援")
	}
	// leader 為 0 一律沒有支援。
	n0.Commander = leader
	if HasBattleSupport(tbl, at, 0) {
		t.Error("leader 為 0 該回 false")
	}
	if SupportingNeighbours(nil, at, leader) != nil {
		t.Error("表為 nil 該回 nil")
	}
}
