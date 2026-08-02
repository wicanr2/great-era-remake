package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

// Person 是發行語系檔中的一筆人物資料；玩家畫面只讀，不回寫研究資料。
type Person struct {
	ID          int      `json:"id"`
	NameInGame  string   `json:"name_ingame"`
	NameCommon  string   `json:"name_common"`
	Aliases     []string `json:"aliases"`
	Courtesy    string   `json:"courtesy"`
	Birth       *int     `json:"birth"`
	Death       *int     `json:"death"`
	Birthplace  string   `json:"birthplace"`
	Faction     string   `json:"faction"`
	HighestPost string   `json:"highest_post"`
	Biography   string   `json:"bio_zh"`
	Periods     []string `json:"periods"`
	Sources     []string `json:"sources"`
	Confidence  string   `json:"confidence"`
}

type peopleFile struct {
	Language string   `json:"language"`
	People   []Person `json:"people"`
}

type slotLink struct {
	Slot, Person int
}

type excludedSlot struct {
	Period, Slot, Person int
	Reason               string
}

type slotsFile struct {
	Periods    map[string][]slotLink `json:"periods"`
	Excluded   []excludedSlot        `json:"excluded"`
	SamePerson [][]int               `json:"same_person"`
}

type personSlot struct{ period, slot int }

// PeopleDB 接合語系人物資料與不隨語系改變的期別槽位表。
type PeopleDB struct {
	Language  string
	people    map[int]*Person
	slots     map[personSlot]int
	excluded  map[personSlot]string
	canonical map[int]int
}

// LoadPeople 載入 localeDir/people.json 與 sharedDir/roster-slots.json。
func LoadPeople(localeDir, sharedDir string) (*PeopleDB, error) {
	readJSON := func(path string, dst any) error {
		b, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("i18n: %w", err)
		}
		if err := json.Unmarshal(b, dst); err != nil {
			return fmt.Errorf("i18n: 解析 %s：%w", filepath.Base(path), err)
		}
		return nil
	}
	var pf peopleFile
	if err := readJSON(filepath.Join(localeDir, "people.json"), &pf); err != nil {
		return nil, err
	}
	var sf slotsFile
	if err := readJSON(filepath.Join(sharedDir, "roster-slots.json"), &sf); err != nil {
		return nil, err
	}
	db := &PeopleDB{Language: pf.Language, people: map[int]*Person{},
		slots: map[personSlot]int{}, excluded: map[personSlot]string{}, canonical: map[int]int{}}
	for i := range pf.People {
		p := &pf.People[i]
		if p.ID <= 0 || p.NameInGame == "" {
			return nil, fmt.Errorf("i18n: 人物 id=%d 缺少有效姓名", p.ID)
		}
		if _, exists := db.people[p.ID]; exists {
			return nil, fmt.Errorf("i18n: 人物 id=%d 重複", p.ID)
		}
		db.people[p.ID], db.canonical[p.ID] = p, p.ID
	}
	for rawPeriod, links := range sf.Periods {
		period, err := strconv.Atoi(rawPeriod)
		if err != nil || period < 1 || period > 3 {
			return nil, fmt.Errorf("i18n: 無效期別 %q", rawPeriod)
		}
		for _, link := range links {
			key := personSlot{period, link.Slot}
			if _, exists := db.slots[key]; exists {
				return nil, fmt.Errorf("i18n: 期別 %d 槽 %d 重複", period, link.Slot)
			}
			if db.people[link.Person] == nil {
				return nil, fmt.Errorf("i18n: 期別 %d 槽 %d 指向不存在人物 %d", period, link.Slot, link.Person)
			}
			db.slots[key] = link.Person
		}
	}
	for _, ex := range sf.Excluded {
		key := personSlot{ex.Period, ex.Slot}
		if db.slots[key] != ex.Person || ex.Reason == "" {
			return nil, fmt.Errorf("i18n: 排除項 %d/%d 與槽位表不一致", ex.Period, ex.Slot)
		}
		db.excluded[key] = ex.Reason
	}
	for _, group := range sf.SamePerson {
		if len(group) < 2 {
			return nil, fmt.Errorf("i18n: same_person 群組至少要兩人")
		}
		canonical := group[0]
		if db.people[canonical] == nil {
			return nil, fmt.Errorf("i18n: same_person 指向不存在人物 %d", canonical)
		}
		for _, id := range group {
			if db.people[id] == nil {
				return nil, fmt.Errorf("i18n: same_person 指向不存在人物 %d", id)
			}
			db.canonical[id] = canonical
		}
	}
	return db, nil
}

// PersonAt 以遊戲期別與期內將領槽位查人物。排除項與不存在槽位都回 false。
func (db *PeopleDB) PersonAt(period, slot int) (*Person, bool) {
	key := personSlot{period, slot}
	if _, excluded := db.excluded[key]; excluded {
		return nil, false
	}
	id, ok := db.slots[key]
	if !ok {
		return nil, false
	}
	p := db.people[id]
	return p, p != nil
}

// ExclusionReason 回傳為何某槽位不提供自傳入口。
func (db *PeopleDB) ExclusionReason(period, slot int) (string, bool) {
	reason, ok := db.excluded[personSlot{period, slot}]
	return reason, ok
}

// CanonicalPersonID 讓不同名冊寫法可識別為同一人物；顯示仍使用各槽自己的 Person。
func (db *PeopleDB) CanonicalPersonID(id int) (int, bool) {
	canonical, ok := db.canonical[id]
	return canonical, ok
}

func (db *PeopleDB) PersonCount() int { return len(db.people) }
func (db *PeopleDB) SlotCount() int   { return len(db.slots) }
