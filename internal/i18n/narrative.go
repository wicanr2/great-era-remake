package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// NarrativeCount 是 NEWSDATA.DAT 目前已證實的新聞橫幅數量。
const NarrativeCount = 17

// NarrativeStatus 說明文字是否已從原版點陣新聞反查成可翻譯字串。
type NarrativeStatus string

const (
	NarrativeTranslated  NarrativeStatus = "translated"
	NarrativeSourceImage NarrativeStatus = "source-image"
)

// NarrativeEntry 是一條不影響規則的敘事／新聞模板。source_image 永遠保留，
// 讓未知字型或填字機制仍能回到原版圖像，不把猜測當成翻譯。
type NarrativeEntry struct {
	ID          int             `json:"id"`
	Status      NarrativeStatus `json:"status"`
	SourceText  string          `json:"source_text,omitempty"`
	Original    string          `json:"original,omitempty"`
	Plain       string          `json:"plain,omitempty"`
	SourceImage string          `json:"source_image"`
}

type narrativeFile struct {
	SchemaVersion string           `json:"schema_version"`
	Language      string           `json:"language"`
	Source        string           `json:"source"`
	Fallback      string           `json:"fallback"`
	Entries       []NarrativeEntry `json:"entries"`
}

// NarrativeCatalog 是語系目錄中的 17 條新聞索引。
type NarrativeCatalog struct {
	Language string
	Source   string
	Fallback string
	entries  map[int]NarrativeEntry
}

// LoadNarrative 以 fail-closed 方式載入 narrative.json；未知條目可以是
// source-image，但不能少 id、重複 id 或缺少可回查的原版資產索引。
func LoadNarrative(dir string) (*NarrativeCatalog, error) {
	b, err := os.ReadFile(filepath.Join(dir, "narrative.json"))
	if err != nil {
		return nil, fmt.Errorf("i18n: %w", err)
	}
	var f narrativeFile
	if err := json.Unmarshal(b, &f); err != nil {
		return nil, fmt.Errorf("i18n: 解析 narrative.json：%w", err)
	}
	if f.SchemaVersion != "1" || !SupportedLanguages[f.Language] || f.Source != "NEWSDATA.DAT" ||
		f.Fallback == "" || len(f.Entries) != NarrativeCount {
		return nil, fmt.Errorf("i18n: narrative.json schema／語系／來源／筆數無效")
	}
	c := &NarrativeCatalog{Language: f.Language, Source: f.Source, Fallback: f.Fallback,
		entries: make(map[int]NarrativeEntry, NarrativeCount)}
	for i, entry := range f.Entries {
		if entry.ID != i || entry.SourceImage == "" ||
			(entry.Status != NarrativeTranslated && entry.Status != NarrativeSourceImage) {
			return nil, fmt.Errorf("i18n: narrative entry %d 的 id／來源／狀態無效", i)
		}
		if entry.Status == NarrativeTranslated && (entry.Original == "" || entry.Plain == "") {
			return nil, fmt.Errorf("i18n: narrative entry %d 缺少 original／plain", i)
		}
		if _, exists := c.entries[entry.ID]; exists {
			return nil, fmt.Errorf("i18n: narrative entry %d 重複", entry.ID)
		}
		c.entries[entry.ID] = entry
	}
	return c, nil
}

// Entry 取一條新聞的證據狀態；id 超界明確回 false。
func (c *NarrativeCatalog) Entry(id int) (NarrativeEntry, bool) {
	if c == nil {
		return NarrativeEntry{}, false
	}
	e, ok := c.entries[id]
	return e, ok
}

// Text 回傳可翻譯文字。source-image 只回傳 false，呼叫端應顯示原版圖像與
// Catalog.Fallback，而不是自行 OCR 或補寫未知句子。
func (c *NarrativeCatalog) Text(id int, mode WordingMode) (string, NarrativeStatus, bool) {
	e, ok := c.Entry(id)
	if !ok || e.Status != NarrativeTranslated {
		return "", e.Status, false
	}
	switch mode {
	case WordingOriginal:
		return e.Original, e.Status, true
	case WordingPlain:
		return e.Plain, e.Status, true
	default:
		return "", e.Status, false
	}
}
