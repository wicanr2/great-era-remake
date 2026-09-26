package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
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

func TestAuthoredOverlayCompletesRuntimeBiographies(t *testing.T) {
	db := loadPeople(t)
	withBiography := 0
	for _, person := range db.people {
		if person.Biography != "" {
			withBiography++
		}
	}
	if withBiography != 387 {
		t.Fatalf("overlay 後正文數 = %d，預期 387", withBiography)
	}
	for _, id := range []int{13, 19, 417} {
		if db.people[id].Biography == "" {
			t.Fatalf("新增人物 #%d 沒有接入自傳", id)
		}
	}
	if db.people[274].Biography != "" {
		t.Fatal("#274 無省長不應被 overlay 成人物自傳")
	}
}

func TestAuthoredOverlayMalformedPayloadFailsClosed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "people-authored.json")
	if err := os.WriteFile(path, []byte(`{"schema_version":"1","language":"zh-Hant","overlay_scope":"new-biographies-only","generated_from":"test","people":[]}`), 0o600); err != nil {
		t.Fatal(err)
	}
	db := &PeopleDB{people: map[int]*Person{}}
	if err := applyAuthoredOverlay(path, "zh-Hant", db); err == nil {
		t.Fatal("壞的 overlay 不應被靜默接受")
	}
}

func TestAuthoredOverlayMissingIsCompatible(t *testing.T) {
	db := &PeopleDB{people: map[int]*Person{}}
	missing := filepath.Join(t.TempDir(), "does-not-exist.json")
	if err := applyAuthoredOverlay(missing, "zh-Hant", db); err != nil {
		t.Fatalf("沒有 overlay 應維持 base-only 相容性：%v", err)
	}
}

func TestTranslatedBiographyOverlayAppliesAfterFullValidation(t *testing.T) {
	db, err := LoadPeople(
		filepath.Join("..", "..", "translations", "en"),
		filepath.Join("..", "..", "translations", "shared"),
	)
	if err != nil {
		t.Fatal(err)
	}
	rows := make([]translatedBiographyRow, 0, translatedOverlayPeople)
	ids := make([]int, 0, len(db.people))
	for id, p := range db.people {
		if id != 274 && p.Biography != "" {
			ids = append(ids, id)
		}
	}
	sort.Ints(ids)
	if len(ids) != translatedOverlayPeople {
		t.Fatalf("英語 source-fallback 正文數 = %d", len(ids))
	}
	for _, id := range ids {
		p := db.people[id]
		bio := p.Biography + " [draft]"
		rows = append(rows, translatedBiographyRow{
			ID:                     id,
			NameInGame:             p.NameInGame,
			Biography:              bio,
			BiographyLanguage:      "en",
			BiographyStatus:        "machine-draft",
			SourceBiographySHA:     sha256Text(p.Biography),
			TranslatedBiographySHA: sha256Text(bio),
			Review:                 translatedReview{Status: "unreviewed"},
		})
	}
	overlay := translatedPeopleFile{
		SchemaVersion:     "1",
		Language:          "en",
		OverlayScope:      "translated-biographies",
		TranslationStatus: "machine-draft",
		GeneratedFrom: []translatedSource{{
			Facts: "facts.json", FactsSHA: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			Bios: "bios.md", BiosSHA: "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		}},
		People: rows,
	}
	b, err := json.Marshal(overlay)
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), "people-biography-overlay.json")
	if err := os.WriteFile(path, b, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := applyTranslatedBiographyOverlay(path, "en", db); err != nil {
		t.Fatal(err)
	}
	if db.people[1].BiographyStatus != "machine-draft" || db.people[1].BiographyLanguage != "en" {
		t.Fatalf("翻譯 overlay 狀態未接上：%+v", db.people[1])
	}
}

func TestTranslatedBiographyOverlayMissingIsCompatible(t *testing.T) {
	db := &PeopleDB{people: map[int]*Person{}}
	missing := filepath.Join(t.TempDir(), "does-not-exist.json")
	if err := applyTranslatedBiographyOverlay(missing, "en", db); err != nil {
		t.Fatalf("沒有翻譯 overlay 應維持 source-fallback：%v", err)
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
