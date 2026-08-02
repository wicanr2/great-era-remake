package game

import "testing"

func TestActiveGeneralsAtKeepsOriginalOrder(t *testing.T) {
	w := &AIWorld{Units: []CombatUnit{
		{General: 1, Province: 2, Active: true},
		{General: 2, Province: 1, Active: true},
		{General: 3, Province: 1, Active: false},
		{General: 4, Province: 1, Active: true},
	}}
	got := w.ActiveGeneralsAt(1)
	if len(got) != 2 || got[0] != 2 || got[1] != 4 {
		t.Fatalf("候選 = %v，預期 [2 4]", got)
	}
}

func TestOwnedProvincesUsesCommanderAndIDOrder(t *testing.T) {
	tbl := &ProvinceTable{}
	for id, commander := range map[ProvinceID]GeneralID{1: 7, 2: 8, 3: 7, 5: 7} {
		p, _ := tbl.At(id)
		p.Commander = commander
	}
	w := &AIWorld{Table: tbl}
	got := w.OwnedProvinces(3)
	if len(got) != 3 || got[0] != 1 || got[1] != 3 || got[2] != 5 {
		t.Fatalf("所屬各省 = %v，預期 [1 3 5]", got)
	}
}
