package game

import (
	"reflect"
	"testing"
)

func TestInvasionCastEveryReviewedMonthHasTenDistinctOfficers(t *testing.T) {
	for year := uint16(15); year <= 34; year++ {
		for month := uint8(1); month <= 12; month++ {
			if year == 34 && month > 8 {
				break
			}
			d, _ := InvasionDateFor(year, month)
			cast, err := SelectInvasionCast(d)
			if err != nil || len(cast) != 10 {
				t.Fatalf("%d/%d: %v, %v", year, month, cast, err)
			}
			seen := map[GeneralID]bool{}
			for _, p := range cast {
				ok, err := InvasionOfficerEligible(p.ID, d)
				if seen[p.ID] || !ok || err != nil || InvasionActorKey(p.ID) != p.Key {
					t.Fatalf("%d/%d: invalid distinct cast %+v", year, month, p)
				}
				seen[p.ID] = true
			}
		}
	}
}

func TestInvasionCastHistoricalExitsRecallAndFrozenLateRoster(t *testing.T) {
	for _, tt := range []struct {
		id       GeneralID
		year     uint16
		month    uint8
		eligible bool
	}{
		{275, 22, 7, true}, {275, 22, 8, false}, // JACAR Muto death, 1933-07-28.
		{281, 30, 2, true}, {281, 30, 3, false}, // NDL month-only Ishiwara retirement.
		{288, 34, 8, true}, {288, 65535, 12, true}, // NDL Anami, frozen August reference.
		{293, 30, 1, true}, {293, 30, 2, false}, {293, 30, 11, false}, {293, 30, 12, true}, // Ando recall.
		{287, 33, 6, true}, {287, 33, 7, false}, // Primary month-only Tojo retirement.
	} {
		d, _ := InvasionDateFor(tt.year, tt.month)
		got, err := InvasionOfficerEligible(tt.id, d)
		if err != nil || got != tt.eligible {
			t.Fatalf("%+v: %v, %v", tt, got, err)
		}
	}
	d, _ := InvasionDateFor(34, 8)
	cast, err := SelectInvasionCast(d)
	if err != nil {
		t.Fatal(err)
	}
	var ids []GeneralID
	for _, p := range cast {
		ids = append(ids, p.ID)
	}
	want := []GeneralID{279, 280, 282, 284, 285, 288, 290, 291, 292, 293}
	if !reflect.DeepEqual(ids, want) {
		t.Fatalf("late cast %v, want %v", ids, want)
	}
	late, _ := InvasionDateFor(35, 1)
	lateCast, err := SelectInvasionCast(late)
	if err != nil || !reflect.DeepEqual(cast, lateCast) {
		t.Fatal("alternate history reselected the frozen roster")
	}
	cast[0].Key = "changed by caller"
	again, _ := SelectInvasionCast(d)
	if again[0].Key == cast[0].Key {
		t.Fatal("caller changed shared catalog")
	}
}

func TestInvasionCastRejectsForgedReferenceAndUnknownIdentity(t *testing.T) {
	d, _ := InvasionDateFor(34, 9)
	d.RosterMonth = 9
	if cast, err := SelectInvasionCast(d); err == nil || cast != nil {
		t.Fatal("accepted forged reference")
	}
	good, _ := InvasionDateFor(26, 7)
	for _, id := range []GeneralID{0, 107, 274, 294, 300, 65535} {
		if InvasionActorKey(id) != "" {
			t.Fatalf("registered unknown identity %d", id)
		}
		if ok, err := InvasionOfficerEligible(id, good); err == nil || ok {
			t.Fatalf("accepted unknown identity %d", id)
		}
	}
	if InvasionActorKey(275) != "jp-muto-nobuyoshi" || InvasionActorKey(290) != "jp-yamada-otozo" ||
		InvasionActorKey(291) != "jp-okamura-yasuji" || InvasionActorKey(293) != "jp-ando-rikichi" {
		t.Fatal("append-only IDs moved")
	}
}
