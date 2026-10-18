package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
)

func TestInvasionPeopleIndependentIdentitiesAndGlyphs(t *testing.T) {
	font, err := assets.EmbeddedModernFont()
	if err != nil {
		t.Fatal(err)
	}
	for _, language := range []string{"zh-Hant", "en", "ja"} {
		db, err := LoadPeople(filepath.Join("../../translations", language), "../../translations/shared")
		if err != nil {
			t.Fatal(err)
		}
		for id := game.GeneralID(275); id <= 293; id++ {
			p, ok := db.InvasionPerson(game.InvasionActorKey(id))
			if !ok || p.ID != 10000+int(id) || len(p.Sources) == 0 || p.Biography == "" {
				t.Fatal("external biography missing", language, id, p)
			}
			if language != "zh-Hant" && p.BiographyStatus != "machine-draft" {
				t.Fatal("translation lacks review status")
			}
			for _, r := range p.NameInGame + p.NameCommon + p.Biography {
				if r == '\n' || r == '\r' || r == '\t' {
					continue
				}
				if _, ok := font.Glyph(r); !ok {
					t.Fatal("new biography missing glyph", language, id, string(r))
				}
			}
		}
		if _, ok := db.PersonAt(1, 275); ok {
			t.Fatal("external person overwrote original slot model")
		}
		if len(db.AllPeople()) != 436 {
			t.Fatal("archive omitted external people")
		}
	}
}

func TestInvasionPeopleMalformedFileCannotPartiallyInstall(t *testing.T) {
	source, err := os.ReadFile("../../translations/zh-Hant/invasion-people.json")
	if err != nil {
		t.Fatal(err)
	}
	var data map[string]any
	if err := json.Unmarshal(source, &data); err != nil {
		t.Fatal(err)
	}
	rows := data["people"].([]any)
	last := rows[len(rows)-1].(map[string]any)
	last["key"] = rows[0].(map[string]any)["key"]
	raw, _ := json.Marshal(data)
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "invasion-people.json"), raw, 0600); err != nil {
		t.Fatal(err)
	}
	db := &PeopleDB{Language: "zh-Hant", people: map[int]*Person{}, canonical: map[int]int{}}
	if applyInvasionPeople(dir, db) == nil || len(db.people) != 0 || len(db.canonical) != 0 || db.invasion != nil {
		t.Fatal("invalid roster partly installed")
	}
}
