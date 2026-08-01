package game

import "testing"

// relocateWorld 造一個「領袖在省 1（前線），省 1 有個後方鄰省」的世界。
func relocateWorld(t *testing.T) (*AIWorld, []General, ProvinceID, ProvinceID) {
	t.Helper()
	w := realWorld(t)
	const leader = GeneralID(7)

	// 全部設成無主，再逐一擺。
	for i := ProvinceID(1); ; i++ {
		p, err := w.Table.At(i)
		if err != nil {
			break
		}
		p.Commander, p.Governor = 0, 0
		p.Flags &^= ProvinceFlagInBattle
	}
	front := ProvinceID(1)
	fp, err := w.Table.At(front)
	if err != nil {
		t.Fatal(err)
	}
	if len(fp.Neighbours) < 2 {
		t.Skip("省 1 的鄰省太少")
	}
	fp.Commander = leader

	// 第一個鄰省當敵方（讓 front 成為前線），第二個當我方後方。
	enemy, _ := w.Table.At(fp.Neighbours[0])
	enemy.Commander = 8
	rear := fp.Neighbours[1]
	rp, _ := w.Table.At(rear)
	rp.Commander = leader

	// rear 不能是前線——它的鄰省不能有敵人。
	rpp, _ := w.Table.At(rear)
	for _, n := range rpp.Neighbours {
		if n == front {
			continue
		}
		np, err := w.Table.At(n)
		if err != nil {
			continue
		}
		np.Commander = leader
	}

	gens := make([]General, 274)
	gens[leader-1].Province = front
	w.Units = make([]CombatUnit, 274)
	w.Units[leader-1] = CombatUnit{General: leader, Province: front, Faction: leader, Active: true}

	if w.Hostile(front) == 0 {
		t.Skip("造不出前線省")
	}
	if w.Hostile(rear) != 0 {
		t.Skip("造不出後方省")
	}
	return w, gens, front, rear
}

// 領袖在前線省就往後方鄰省搬，並同時當上司令與省長。
func TestRelocateMovesLeaderToRear(t *testing.T) {
	w, gens, front, rear := relocateWorld(t)
	const leader = GeneralID(7)

	got := w.Relocate(front, leader, gens)
	if !got.Moved {
		t.Fatal("領袖在前線省又有後方鄰省，應該要搬")
	}
	if got.To != rear {
		t.Errorf("搬到省 %d，預期 %d", got.To, rear)
	}
	if gens[leader-1].Province != rear {
		t.Errorf("將領記錄的所屬省是 %d，應該跟著改成 %d",
			gens[leader-1].Province, rear)
	}
	if w.Units[leader-1].Province != rear {
		t.Error("執行期單位記錄沒有跟著搬——兩邊會不一致")
	}
	rp, _ := w.Table.At(rear)
	if rp.Commander != leader {
		t.Errorf("目標省的司令是 %d，應該是領袖 %d", rp.Commander, leader)
	}
	if rp.Governor != leader {
		t.Errorf("目標省的省長是 %d，應該也是領袖 %d（原版 +20 與 +30 寫同一個值）",
			rp.Governor, leader)
	}
}

// 三道不搬的條件。
func TestRelocateSkips(t *testing.T) {
	const leader = GeneralID(7)

	t.Run("領袖不在這個省就不搬", func(t *testing.T) {
		w, gens, front, _ := relocateWorld(t)
		gens[leader-1].Province = 99
		if w.Relocate(front, leader, gens).Moved {
			t.Error("領袖不在這裡，不該搬")
		}
	})
	t.Run("這個省不是前線就不搬", func(t *testing.T) {
		w, gens, front, _ := relocateWorld(t)
		// 把敵人拿掉 → front 不再是前線。
		fp, _ := w.Table.At(front)
		for _, n := range fp.Neighbours {
			np, _ := w.Table.At(n)
			np.Commander = leader
		}
		if w.Relocate(front, leader, gens).Moved {
			t.Error("沒有威脅就不用搬")
		}
	})
	t.Run("鄰省交戰中就不選它", func(t *testing.T) {
		w, gens, front, rear := relocateWorld(t)
		rp, _ := w.Table.At(rear)
		rp.Flags |= ProvinceFlagInBattle
		if got := w.Relocate(front, leader, gens); got.Moved && got.To == rear {
			t.Error("交戰中的省不該當遷都目標")
		}
	})
}
