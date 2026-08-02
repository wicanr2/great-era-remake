package i18n

import (
	"path/filepath"
	"testing"
)

func loadPeople(t *testing.T) *PeopleDB {
	t.Helper()
	db, err := LoadPeople(
		filepath.Join("..", "..", "translations", "zh-Hant"),
		filepath.Join("..", "..", "translations", "shared"))
	if err != nil {
		t.Fatal(err)
	}
	return db
}

func TestPeopleSlotsCoverAllThreePeriods(t *testing.T) {
	db := loadPeople(t)
	if db.PersonCount() != 417 || db.SlotCount() != 486 {
		t.Fatalf("人物／槽位 = %d/%d，預期 417/486", db.PersonCount(), db.SlotCount())
	}
	for period, count := range map[int]int{1: 274, 2: 106, 3: 106} {
		for slot := 1; slot <= count; slot++ {
			if period == 1 && slot == 274 {
				continue
			}
			if _, ok := db.PersonAt(period, slot); !ok {
				t.Errorf("期別 %d 槽 %d 沒有人物", period, slot)
			}
		}
	}
}

func TestNoGovernorPlaceholderIsExcluded(t *testing.T) {
	db := loadPeople(t)
	if p, ok := db.PersonAt(1, 274); ok || p != nil {
		t.Fatalf("無省長不應成為人物：%+v", p)
	}
	if reason, ok := db.ExclusionReason(1, 274); !ok || reason != "placeholder-no-governor" {
		t.Fatalf("排除理由 = %q, %v", reason, ok)
	}
}

func TestTianZhennanOccupiesTwoSlots(t *testing.T) {
	db := loadPeople(t)
	a, okA := db.PersonAt(2, 81)
	b, okB := db.PersonAt(2, 105)
	if !okA || !okB || a.ID != 335 || b.ID != 335 || a != b {
		t.Fatalf("田鎮南雙槽 = %+v / %+v", a, b)
	}
}

func TestSamePersonRelationsPreserveInGameSpelling(t *testing.T) {
	db := loadPeople(t)
	cases := [][4]int{{1, 4, 2, 24}, {1, 50, 2, 23}, {1, 99, 3, 74}}
	for _, c := range cases {
		a, okA := db.PersonAt(c[0], c[1])
		b, okB := db.PersonAt(c[2], c[3])
		if !okA || !okB {
			t.Fatalf("槽位查詢失敗：%v", c)
		}
		ca, _ := db.CanonicalPersonID(a.ID)
		cb, _ := db.CanonicalPersonID(b.ID)
		if ca != cb || a.ID == b.ID || a.NameInGame == b.NameInGame {
			t.Errorf("同一人物關聯／原版寫法未保留：%+v / %+v，canonical %d/%d", a, b, ca, cb)
		}
	}
}

func TestPeopleOutputAppliedAuditedNormalization(t *testing.T) {
	db := loadPeople(t)
	p, ok := db.PersonAt(1, 58) // 吳佩孚傳記含民國年份與異體檢驗樣本的常見字。
	if !ok || p.Biography == "" {
		t.Fatal("吳佩孚應有自傳")
	}
	for _, forbidden := range []string{"鲁", "献", "専", "継", "荣", "钟", "衞", "啓", "羣", "〇"} {
		for _, person := range db.people {
			if containsRune(person.NameInGame+person.NameCommon+person.Biography, []rune(forbidden)[0]) {
				t.Fatalf("發行語系資料仍含未正規化字 %q（人物 %d）", forbidden, person.ID)
			}
		}
	}
}

func containsRune(s string, want rune) bool {
	for _, r := range s {
		if r == want {
			return true
		}
	}
	return false
}
