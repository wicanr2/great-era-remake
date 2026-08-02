package game

import "testing"

func reorganizationWorld(t *testing.T, branch uint8, forces []uint16, arms []uint8) *AIWorld {
	t.Helper()
	tbl := &ProvinceTable{}
	p, _ := tbl.At(1)
	p.Commander = 1
	w := &AIWorld{Table: tbl}
	for i, force := range forces {
		w.Units = append(w.Units, CombatUnit{
			General: GeneralID(i + 1), Faction: 1, Province: 1, Active: true,
		})
		w.Strengths = append(w.Strengths, StrengthInput{
			General: GeneralID(i + 1), Faction: 1, Branch: branch,
			Force: force, F20: arms[i],
		})
	}
	return w
}

func TestReorganizationConservesForce(t *testing.T) {
	w := reorganizationWorld(t, BranchInfantry,
		[]uint16{12000, 3000}, []uint8{40, 80})
	r, err := w.BeginReorganization(1, BranchInfantry)
	if err != nil {
		t.Fatal(err)
	}
	if got := r.Targets(); len(got) != 2 || got[0] != 1 || got[1] != 2 {
		t.Fatalf("候選順序 = %v，預期 [1 2]", got)
	}
	if err := r.Assign(1, 5000); err != nil {
		t.Fatal(err)
	}
	if !r.Changed() {
		t.Fatal("兵力已改但 Changed 仍為 false")
	}
	if r.Remaining() != 7000 || r.CanFinish() {
		t.Fatalf("第一次後兵力池 = %d，應為 7000 且不可結束", r.Remaining())
	}
	if got := r.Limit(2); got != 10000 {
		t.Fatalf("第二將可配 %d，預期 10000", got)
	}
	if err := r.Assign(2, 10000); err != nil {
		t.Fatal(err)
	}
	if !r.CanFinish() || r.Remaining() != 0 {
		t.Fatalf("配完後仍剩 %d", r.Remaining())
	}
	if got := int(w.Strengths[0].Force) + int(w.Strengths[1].Force); got != 15000 {
		t.Fatalf("總兵力變成 %d，原本是 15000", got)
	}
}

func TestReorganizationBlendsArmsForInfantry(t *testing.T) {
	w := reorganizationWorld(t, BranchInfantry,
		[]uint16{100, 100}, []uint8{20, 80})
	r, _ := w.BeginReorganization(1, BranchInfantry)
	if err := r.Assign(1, 0); err != nil {
		t.Fatal(err)
	}
	if err := r.Assign(2, 100); err != nil {
		t.Fatal(err)
	}
	// 100×20 與 100×80 混合成 200 人、平均武裝 50；先配回 100 人。
	if w.Strengths[1].F20 != 50 || r.Remaining() != 100 {
		t.Fatalf("混合後 F20=%d、剩餘=%d，預期 50、100",
			w.Strengths[1].F20, r.Remaining())
	}
	if err := r.Assign(1, 100); err != nil {
		t.Fatal(err)
	}
	if w.Strengths[0].F20 != 50 || !r.CanFinish() {
		t.Fatalf("第二份 F20=%d、剩餘=%d", w.Strengths[0].F20, r.Remaining())
	}
}

func TestReorganizationHeavyBranchesKeepArms(t *testing.T) {
	for _, branch := range []uint8{BranchArtiller, BranchArmour} {
		w := reorganizationWorld(t, branch, []uint16{100, 50}, []uint8{20, 80})
		r, err := w.BeginReorganization(1, branch)
		if err != nil {
			t.Fatal(err)
		}
		if err := r.Assign(1, 0); err != nil {
			t.Fatal(err)
		}
		if err := r.Assign(2, 150); err != nil {
			t.Fatal(err)
		}
		if w.Strengths[0].F20 != 20 || w.Strengths[1].F20 != 80 {
			t.Errorf("兵種 %d 不該改武裝程度：%d/%d", branch,
				w.Strengths[0].F20, w.Strengths[1].F20)
		}
	}
}

func TestReorganizationRejectsOverBranchCap(t *testing.T) {
	w := reorganizationWorld(t, BranchArmour, []uint16{200, 200}, []uint8{50, 50})
	r, _ := w.BeginReorganization(1, BranchArmour)
	if err := r.Assign(1, 0); err != nil {
		t.Fatal(err)
	}
	if r.Limit(2) != 200 {
		t.Fatalf("裝甲兵上限 = %d，預期 200", r.Limit(2))
	}
	if err := r.Assign(2, 201); err == nil {
		t.Fatal("超過裝甲兵滿員上限應失敗")
	}
}
