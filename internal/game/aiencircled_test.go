package game

import "testing"

// 無主鄰省不算敵人——這與攻打候選的規則正好相反。
func TestEnemyNeighboursIgnoresNeutral(t *testing.T) {
	w := realWorld(t)
	me, _ := w.Table.At(1)
	me.Commander = 7
	if len(me.Neighbours) < 2 {
		t.Skip("省 1 的鄰省太少，換一個測不了")
	}
	n1, _ := w.Table.At(me.Neighbours[0])
	n2, _ := w.Table.At(me.Neighbours[1])
	n1.Commander = 0 // 無主
	n2.Commander = 8 // 敵人
	for _, n := range me.Neighbours[2:] {
		p, _ := w.Table.At(n)
		p.Commander = 7 // 自己人
	}

	got := w.EnemyNeighbours(1, 7)
	if len(got) != 1 || got[0] != me.Neighbours[1] {
		t.Errorf("敵對鄰省 %v，預期只有 %d（無主與自己人都不算）",
			got, me.Neighbours[1])
	}
}

// 全部鄰省都是敵人 → 被包圍。
func TestEncircled(t *testing.T) {
	w := realWorld(t)
	me, _ := w.Table.At(1)
	me.Commander = 7
	for _, n := range me.Neighbours {
		p, _ := w.Table.At(n)
		p.Commander = 8
	}
	if !w.Encircled(1) {
		t.Error("全部鄰省都是敵人，應該判為被包圍")
	}

	// 有一個鄰省變成自己人 → 不算被包圍。
	first, _ := w.Table.At(me.Neighbours[0])
	first.Commander = 7
	if w.Encircled(1) {
		t.Error("有自己人的鄰省就不該算被包圍")
	}

	// ⚠️ 無主鄰省也會讓它不成立——無主不算敵人。
	first.Commander = 0
	if w.Encircled(1) {
		t.Error("無主鄰省不算敵人，所以不該算被包圍")
	}
}

// 無主省本身算不出結果，回 false。
func TestEncircledNeedsCommander(t *testing.T) {
	w := realWorld(t)
	me, _ := w.Table.At(1)
	me.Commander = 0
	if w.Encircled(1) {
		t.Error("無主省不該判為被包圍")
	}
}

// 在鄰省裡找第一個被包圍的省。
func TestFirstEncircledNeighbour(t *testing.T) {
	w := realWorld(t)
	me, _ := w.Table.At(1)
	me.Commander = 7
	if len(me.Neighbours) == 0 {
		t.Skip("省 1 沒有鄰省")
	}

	// 先讓所有鄰省都不被包圍。
	for _, n := range me.Neighbours {
		p, _ := w.Table.At(n)
		p.Commander = 7
		for _, nn := range p.Neighbours {
			q, _ := w.Table.At(nn)
			q.Commander = 7
		}
	}
	if got := w.FirstEncircledNeighbour(1); got != 0 {
		t.Errorf("沒有被包圍的鄰省，卻回了 %d", got)
	}

	// 把第一個鄰省的所有鄰省變成敵人。
	target := me.Neighbours[0]
	tp, _ := w.Table.At(target)
	tp.Commander = 7
	for _, nn := range tp.Neighbours {
		q, _ := w.Table.At(nn)
		q.Commander = 9
	}
	if got := w.FirstEncircledNeighbour(1); got != target {
		t.Errorf("回了 %d，預期 %d", got, target)
	}
}
