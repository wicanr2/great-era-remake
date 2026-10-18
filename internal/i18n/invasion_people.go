package i18n

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

func applyInvasionPeople(localeDir string, db *PeopleDB) error {
	data, err := os.ReadFile(filepath.Join(localeDir, "invasion-people.json"))
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	var file struct {
		Schema   string `json:"schema_version"`
		Language string `json:"language"`
		People   []struct {
			Key    string `json:"key"`
			Person Person `json:"person"`
		} `json:"people"`
	}
	if err := json.Unmarshal(data, &file); err != nil {
		return err
	}
	if file.Schema != "1" || file.Language != db.Language || len(file.People) != 19 {
		return fmt.Errorf("i18n: invalid invasion people file")
	}
	people := map[int]*Person{}
	keys := map[string]int{}
	for i, row := range file.People {
		p := row.Person
		if row.Key == "" || keys[row.Key] != 0 || p.ID != 10275+i || db.people[p.ID] != nil ||
			p.NameInGame == "" || strings.TrimSpace(p.Biography) == "" || p.BiographyLanguage != db.Language || len(p.Sources) == 0 ||
			p.BiographyStatus != "original" && p.BiographyStatus != "machine-draft" {
			return fmt.Errorf("i18n: invalid invasion person %s", row.Key)
		}
		for _, source := range p.Sources {
			if !validHTTPSURL(source) {
				return fmt.Errorf("i18n: invalid invasion person source")
			}
		}
		people[p.ID] = &p
		keys[row.Key] = p.ID
	}
	for id, p := range people {
		db.people[id] = p
		db.canonical[id] = id
	}
	db.invasion = keys
	return nil
}

func (db *PeopleDB) InvasionPerson(key string) (*Person, bool) {
	if db == nil {
		return nil, false
	}
	return db.PersonByID(db.invasion[key])
}

func (db *PeopleDB) AllPeople() []*Person {
	if db == nil {
		return nil
	}
	ids := make([]int, 0, len(db.people))
	for id := range db.people {
		ids = append(ids, id)
	}
	sort.Ints(ids)
	out := make([]*Person, 0, len(ids))
	for _, id := range ids {
		p, _ := db.PersonByID(id)
		out = append(out, p)
	}
	return out
}
