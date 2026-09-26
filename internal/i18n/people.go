package i18n

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
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
	// BiographyLanguage 記錄 Biography 實際撰寫語言。繁中母本沿用 bio_zh；
	// 英／日語系在尚無可追溯譯稿時會明示 zh-Hant，避免把來源文字冒充翻譯。
	BiographyLanguage string `json:"bio_language,omitempty"`
	// BiographyStatus 是發行資料的狀態標記；目前允許 original、translated、source-fallback。
	BiographyStatus string   `json:"bio_status,omitempty"`
	Periods         []string `json:"periods"`
	Sources         []string `json:"sources"`
	Confidence      string   `json:"confidence"`
}

type peopleFile struct {
	Language string   `json:"language"`
	People   []Person `json:"people"`
}

type authoredPeopleFile struct {
	SchemaVersion string           `json:"schema_version"`
	Language      string           `json:"language"`
	OverlayScope  string           `json:"overlay_scope"`
	GeneratedFrom string           `json:"generated_from"`
	People        []authoredPerson `json:"people"`
}

type authoredPerson struct {
	ID         int                `json:"id"`
	NameInGame string             `json:"name_ingame"`
	Biography  string             `json:"bio_zh"`
	Confidence string             `json:"confidence"`
	Provenance authoredProvenance `json:"provenance"`
}

type authoredProvenance struct {
	Facts        string `json:"facts"`
	FactsSHA     string `json:"facts_sha256"`
	Bios         string `json:"bios"`
	BiosSHA      string `json:"bios_sha256"`
	BiographySHA string `json:"bio_sha256"`
}

// translatedPeopleFile 是英／日人物自傳的正式 overlay 外殼。
//
// 它與 zh-Hant 的研究 overlay 分開：研究正文仍是來源，翻譯稿只覆蓋
// Biography 與語系狀態；姓名、年代、派系與來源欄位一律沿用 locale base。
// machine-draft 可以供預覽載入，但是否能發行由 tools/locale_bio_gate.py
// 的 --release gate 決定。
type translatedPeopleFile struct {
	SchemaVersion     string                   `json:"schema_version"`
	Language          string                   `json:"language"`
	OverlayScope      string                   `json:"overlay_scope"`
	TranslationStatus string                   `json:"translation_status"`
	GeneratedFrom     []translatedSource       `json:"generated_from"`
	People            []translatedBiographyRow `json:"people"`
}

type translatedSource struct {
	Facts    string `json:"facts"`
	FactsSHA string `json:"facts_sha256"`
	Bios     string `json:"bios"`
	BiosSHA  string `json:"bios_sha256"`
}

type translatedBiographyRow struct {
	ID                     int              `json:"id"`
	NameInGame             string           `json:"name_ingame"`
	Biography              string           `json:"bio"`
	BiographyLanguage      string           `json:"bio_language"`
	BiographyStatus        string           `json:"bio_status"`
	SourceBiographySHA     string           `json:"source_bio_sha256"`
	TranslatedBiographySHA string           `json:"translated_bio_sha256"`
	Review                 translatedReview `json:"review"`
}

type translatedReview struct {
	Status     string `json:"status"`
	Reviewer   string `json:"reviewer"`
	ReviewedAt string `json:"reviewed_at"`
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

const authoredOverlayFile = "people-authored.json"

// SPEC-11 固定的增量產物大小。這不是把 387 篇硬編進規則，而是防止產生器
// 輸出範圍漂移後，載入器意外覆蓋既有產品正文。
const authoredOverlayPeople = 61

const translatedOverlayPeople = 387

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
	if !SupportedLanguages[pf.Language] {
		return nil, fmt.Errorf("i18n: 人物資料不支援語系 %q", pf.Language)
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
	if err := applyAuthoredOverlay(filepath.Join(localeDir, authoredOverlayFile), pf.Language, db); err != nil {
		return nil, err
	}
	if err := applyTranslatedBiographyOverlay(
		filepath.Join(localeDir, "people-biography-overlay.json"), pf.Language, db,
	); err != nil {
		return nil, err
	}
	return db, nil
}

func applyAuthoredOverlay(path, language string, db *PeopleDB) error {
	data, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("i18n: 讀取人物自撰 overlay：%w", err)
	}
	var overlay authoredPeopleFile
	if err := json.Unmarshal(data, &overlay); err != nil {
		return fmt.Errorf("i18n: 解析人物自撰 overlay：%w", err)
	}
	if overlay.SchemaVersion != "1" || overlay.Language == "" || overlay.Language != language ||
		overlay.OverlayScope != "new-biographies-only" || strings.TrimSpace(overlay.GeneratedFrom) == "" {
		return fmt.Errorf("i18n: 人物自撰 overlay 的 schema／語系／範圍無效")
	}
	if len(overlay.People) != authoredOverlayPeople {
		return fmt.Errorf("i18n: 人物自撰 overlay 應有 %d 筆，實際 %d", authoredOverlayPeople, len(overlay.People))
	}
	seen := make(map[int]bool, len(overlay.People))
	previousID := 0
	for _, row := range overlay.People {
		if row.ID <= previousID || row.ID <= 0 || seen[row.ID] {
			return fmt.Errorf("i18n: 人物自撰 overlay 有無效或重複 id=%d", row.ID)
		}
		seen[row.ID] = true
		previousID = row.ID
		if row.ID == 274 {
			return fmt.Errorf("i18n: 人物自撰 overlay 不得包含 #274 無省長")
		}
		p := db.people[row.ID]
		if p == nil {
			return fmt.Errorf("i18n: 人物自撰 overlay 指向不存在人物 %d", row.ID)
		}
		if row.NameInGame != p.NameInGame {
			return fmt.Errorf("i18n: 人物自撰 overlay #%d 姓名與 base 不一致", row.ID)
		}
		if strings.TrimSpace(row.Biography) == "" {
			return fmt.Errorf("i18n: 人物自撰 overlay #%d 正文為空", row.ID)
		}
		if p.Biography != "" {
			return fmt.Errorf("i18n: 人物自撰 overlay #%d 試圖覆蓋既有正文", row.ID)
		}
		if row.Confidence != "high" && row.Confidence != "medium" && row.Confidence != "low" {
			return fmt.Errorf("i18n: 人物自撰 overlay #%d 信心度無效：%q", row.ID, row.Confidence)
		}
		if err := validateAuthoredProvenance(row.Provenance); err != nil {
			return fmt.Errorf("i18n: 人物自撰 overlay #%d provenance：%w", row.ID, err)
		}
		bodySHA := sha256.Sum256([]byte(row.Biography))
		if !strings.EqualFold(hex.EncodeToString(bodySHA[:]), row.Provenance.BiographySHA) {
			return fmt.Errorf("i18n: 人物自撰 overlay #%d 正文雜湊不一致", row.ID)
		}
		p.Biography = row.Biography
		p.Confidence = row.Confidence
	}
	return nil
}

func validateAuthoredProvenance(p authoredProvenance) error {
	if strings.TrimSpace(p.Facts) == "" || strings.TrimSpace(p.Bios) == "" {
		return fmt.Errorf("缺少研究檔路徑")
	}
	for name, value := range map[string]string{
		"facts_sha256": p.FactsSHA, "bios_sha256": p.BiosSHA, "bio_sha256": p.BiographySHA,
	} {
		if len(value) != 64 {
			return fmt.Errorf("%s 不是 64 位十六進位雜湊", name)
		}
		if _, err := hex.DecodeString(value); err != nil {
			return fmt.Errorf("%s 不是十六進位雜湊", name)
		}
	}
	return nil
}

// applyTranslatedBiographyOverlay 以 fail-closed 方式套用英／日 387 篇自傳。
//
// 沒有 overlay 時維持 source-fallback，讓目前的英／日包仍可啟動；一旦檔案
// 存在，則整份驗證完才修改 db，避免半份翻譯悄悄進入遊戲。這裡不把
// machine-draft 升格成正式譯文，也不驗證外部來源檔內容（來源檔雜湊由
// locale_bio_gate.py 在產生／發行流程驗證）。
func applyTranslatedBiographyOverlay(path, language string, db *PeopleDB) error {
	if language == "zh-Hant" {
		return nil
	}
	b, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("i18n: 讀取人物翻譯 overlay：%w", err)
	}
	var overlay translatedPeopleFile
	if err := json.Unmarshal(b, &overlay); err != nil {
		return fmt.Errorf("i18n: 解析人物翻譯 overlay：%w", err)
	}
	if overlay.SchemaVersion != "1" || overlay.Language != language ||
		overlay.OverlayScope != "translated-biographies" ||
		(overlay.TranslationStatus != "machine-draft" && overlay.TranslationStatus != "human-reviewed") {
		return fmt.Errorf("i18n: 人物翻譯 overlay 的 schema／語系／狀態無效")
	}
	if len(overlay.GeneratedFrom) == 0 {
		return fmt.Errorf("i18n: 人物翻譯 overlay 缺少 generated_from")
	}
	for _, source := range overlay.GeneratedFrom {
		if strings.TrimSpace(source.Facts) == "" || strings.TrimSpace(source.Bios) == "" ||
			!validSHA256(source.FactsSHA) || !validSHA256(source.BiosSHA) {
			return fmt.Errorf("i18n: 人物翻譯 overlay 的 generated_from 無效")
		}
	}
	if len(overlay.People) != translatedOverlayPeople {
		return fmt.Errorf("i18n: 人物翻譯 overlay 應有 %d 筆，實際 %d", translatedOverlayPeople, len(overlay.People))
	}

	// 先完整驗證並建立暫存 map；任何一列錯誤都不污染 db。
	updates := make(map[int]translatedBiographyRow, len(overlay.People))
	previousID := 0
	for _, row := range overlay.People {
		if row.ID <= previousID || row.ID == 274 || row.ID <= 0 {
			return fmt.Errorf("i18n: 人物翻譯 overlay id 未遞增或含排除槽：%d", row.ID)
		}
		previousID = row.ID
		if _, exists := updates[row.ID]; exists {
			return fmt.Errorf("i18n: 人物翻譯 overlay id=%d 重複", row.ID)
		}
		person := db.people[row.ID]
		if person == nil {
			return fmt.Errorf("i18n: 人物翻譯 overlay 指向不存在人物 %d", row.ID)
		}
		if row.NameInGame != person.NameInGame {
			return fmt.Errorf("i18n: 人物翻譯 overlay #%d 姓名與 base 不一致", row.ID)
		}
		if strings.TrimSpace(row.Biography) == "" || row.BiographyLanguage != language ||
			row.BiographyStatus != overlay.TranslationStatus {
			return fmt.Errorf("i18n: 人物翻譯 overlay #%d 正文／語系／狀態無效", row.ID)
		}
		if row.SourceBiographySHA != sha256Text(person.Biography) ||
			row.TranslatedBiographySHA != sha256Text(row.Biography) {
			return fmt.Errorf("i18n: 人物翻譯 overlay #%d 雜湊不一致", row.ID)
		}
		if overlay.TranslationStatus == "machine-draft" {
			if row.Review.Status != "unreviewed" && row.Review.Status != "in-review" {
				return fmt.Errorf("i18n: 人物翻譯 overlay #%d machine-draft review 無效", row.ID)
			}
		} else if row.Review.Status != "human-reviewed" ||
			strings.TrimSpace(row.Review.Reviewer) == "" || strings.TrimSpace(row.Review.ReviewedAt) == "" {
			return fmt.Errorf("i18n: 人物翻譯 overlay #%d 缺少人工審閱紀錄", row.ID)
		}
		updates[row.ID] = row
	}
	if len(updates) != translatedOverlayPeople {
		return fmt.Errorf("i18n: 人物翻譯 overlay 有重複或遺漏人物")
	}
	for id, row := range updates {
		person := db.people[id]
		person.Biography = row.Biography
		person.BiographyLanguage = row.BiographyLanguage
		person.BiographyStatus = row.BiographyStatus
	}
	return nil
}

func sha256Text(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func validSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
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

// PersonByID 供 modern UI 以穩定人物 id 取本地化姓名。它只回傳資料副本，
// 呼叫端不能藉此修改 PeopleDB 的 overlay 狀態；找不到人物時明確回 false。
func (db *PeopleDB) PersonByID(id int) (*Person, bool) {
	if db == nil {
		return nil, false
	}
	p, ok := db.people[id]
	if !ok || p == nil {
		return nil, false
	}
	copy := *p
	copy.Aliases = append([]string(nil), p.Aliases...)
	copy.Periods = append([]string(nil), p.Periods...)
	copy.Sources = append([]string(nil), p.Sources...)
	return &copy, true
}
