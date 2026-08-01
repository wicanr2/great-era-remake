// Package i18n 是 remake 的語系層：把玩家看得到的文字從程式碼裡抽出來。
//
// `CLAUDE.md` §6 的兩條要求：
//
//	所有玩家看得到的文字都在語系資料檔，不進 Go 原始碼
//	原版的字模索引機制到 remake 就結束，重寫版用真正的字串表
//
// 所以這一層**執行期載入**外部 JSON，不用 `go:embed`——
// 英文與日文版只要換一個目錄就行，不必重編。
//
// ⚠️ 這個 package **不依賴 Ebiten**，才能在無頭環境測試
// （`CLAUDE.md` §11 對 `internal/ui/textlayout` 的同一條理由）。
package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// Entry 是一個詞條。
type Entry struct {
	// Text 是去掉排版空白的文字，remake 用這個。
	Text string `json:"text"`
	// Raw 保留原版的排版空白（全形空格），用來還原原版畫面。
	//
	// 例：省名「西藏」在原版是「西　藏」——三個字模槽，中間留白。
	Raw string `json:"raw"`
}

// File 是一個原版字模檔對應的詞條表。
type File struct {
	// SlotWidth 是原版的槽寬（每個詞條佔幾個字模格）。
	SlotWidth int     `json:"slot_width"`
	Entries   []Entry `json:"entries"`
}

// Locale 是一份語系資料。
type Locale struct {
	Language string           `json:"language"`
	Files    map[string]*File `json:"files"`
}

// ProvinceCount 是省份數。省名在 `3.15` 的前 39 條
// （`docs/formats/01-glyph-text.md` §4）。
const ProvinceCount = 39

// provinceTable 是省名所在的檔案。
const provinceTable = "3.15"

// Load 從目錄讀語系資料（該目錄要有 `glyphtext.json`）。
func Load(dir string) (*Locale, error) {
	b, err := os.ReadFile(filepath.Join(dir, "glyphtext.json"))
	if err != nil {
		return nil, fmt.Errorf("i18n: %w", err)
	}
	var l Locale
	if err := json.Unmarshal(b, &l); err != nil {
		return nil, fmt.Errorf("i18n: 解析 glyphtext.json：%w", err)
	}
	if l.Files == nil {
		return nil, fmt.Errorf("i18n: %s 沒有 files", dir)
	}
	return &l, nil
}

// Province 回傳省名（1-based）。超出範圍回 "省 N" 而不是空字串——
// **看得出是哪一省，也看得出語系表沒涵蓋它**。
func (l *Locale) Province(id int) string {
	if s, ok := l.Entry(provinceTable, id); ok && id >= 1 && id <= ProvinceCount {
		return s
	}
	return fmt.Sprintf("省 %d", id)
}

// Entry 取某個表的第 n 條（**1-based**，與原版的字模索引一致，
// `docs/re/24`）。第二個回傳值表示有沒有取到。
func (l *Locale) Entry(table string, n int) (string, bool) {
	f, ok := l.Files[table]
	if !ok || n < 1 || n > len(f.Entries) {
		return "", false
	}
	return f.Entries[n-1].Text, true
}

// RawEntry 同 Entry，但保留原版排版空白。
func (l *Locale) RawEntry(table string, n int) (string, bool) {
	f, ok := l.Files[table]
	if !ok || n < 1 || n > len(f.Entries) {
		return "", false
	}
	return f.Entries[n-1].Raw, true
}

// Len 回傳某個表有幾條詞條，0 表示沒有這個表。
func (l *Locale) Len(table string) int {
	if f, ok := l.Files[table]; ok {
		return len(f.Entries)
	}
	return 0
}
