package i18n

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// WordingMode 是顯示用語軸；它不影響規則、候選、數值或存檔。
type WordingMode string

const (
	WordingOriginal WordingMode = "original"
	WordingPlain    WordingMode = "plain"
)

// ParseWordingMode 驗證命令列或偏好設定讀到的用語模式。
func ParseWordingMode(s string) (WordingMode, error) {
	m := WordingMode(s)
	if m != WordingOriginal && m != WordingPlain {
		return "", fmt.Errorf("i18n: 未知用語模式 %q（應為 original 或 plain）", s)
	}
	return m, nil
}

// WordingEntry 讓同一個穩定語意鍵同時擁有原典與現代白話文字。
type WordingEntry struct {
	Original string `json:"original"`
	Plain    string `json:"plain"`
}

// WordingCatalog 是不以原文字串作判斷條件的語意詞條表。
type WordingCatalog struct {
	Entries map[string]WordingEntry `json:"entries"`
}

// RequiredWordingKeys 是目前已有玩家入口的穩定語意鍵。語系檔少任何一項都
// 整份拒絕，避免某個深層畫面才突然退回原典或顯示空白。
var RequiredWordingKeys = []string{
	"common.confirm", "common.cancel",
	"settings.title", "settings.wording", "settings.wording.original", "settings.wording.plain",
	"other.save.confirm", "other.load.confirm", "other.message_time.prompt",
	"policy.title", "policy.autonomy", "policy.production", "policy.production.unavailable",
	"autonomy.title", "autonomy.normal", "autonomy.enabled", "autonomy.prompt",
	"production.title", "production.gold", "production.iron", "production.coal",
	"production.oil", "production.food", "production.select", "production.value",
	"recruit.action", "recruit.reorganize", "recruit.infantry", "recruit.armour",
	"recruit.artillery", "recruit.cavalry", "recruit.amount", "recruit.limit",
	"recruit.cost", "recruit.gold", "recruit.confirm", "recruit.remaining",
	"recruit.general", "recruit.force",
	"command.01", "command.02", "command.03", "command.04", "command.05",
	"command.06", "command.07", "command.08", "command.09", "command.10",
	"command.11", "command.12", "command.13", "command.14", "command.15",
	"trade.import", "trade.export", "trade.food", "trade.ammo", "trade.fuel",
	"trade.coal", "trade.iron", "trade.buy_amount", "trade.sell_amount",
	"supply.target", "supply.gold", "supply.food", "supply.ammo", "supply.fuel",
	"transfer.mode", "transfer.mode.partial", "transfer.mode.all",
	"transfer.target", "transfer.select.partial", "transfer.select.all",
	"transfer.selection.confirm", "transfer.resource.gold", "transfer.resource.food",
	"transfer.resource.ammo", "transfer.resource.fuel",
	"biography.unavailable", "biography.page",
}

// LoadWording 載入一個語系目錄的 wording.json，並採失敗即關閉：任何一套缺文
// 都拒絕載入，避免切換後安靜地退回另一套用語。
func LoadWording(dir string) (*WordingCatalog, error) {
	b, err := os.ReadFile(filepath.Join(dir, "wording.json"))
	if err != nil {
		return nil, fmt.Errorf("i18n: %w", err)
	}
	var c WordingCatalog
	if err := json.Unmarshal(b, &c); err != nil {
		return nil, fmt.Errorf("i18n: 解析 wording.json：%w", err)
	}
	if len(c.Entries) == 0 {
		return nil, fmt.Errorf("i18n: wording.json 沒有 entries")
	}
	for key, entry := range c.Entries {
		if key == "" || entry.Original == "" || entry.Plain == "" {
			return nil, fmt.Errorf("i18n: 用語鍵 %q 缺少 original 或 plain", key)
		}
	}
	for _, key := range RequiredWordingKeys {
		if _, ok := c.Entries[key]; !ok {
			return nil, fmt.Errorf("i18n: wording.json 缺少必要語意鍵 %q", key)
		}
	}
	return &c, nil
}

// Text 依模式取文字；缺鍵或錯誤模式明確回 false，不暗中 fallback。
func (c *WordingCatalog) Text(key string, mode WordingMode) (string, bool) {
	entry, ok := c.Entries[key]
	if !ok {
		return "", false
	}
	switch mode {
	case WordingOriginal:
		return entry.Original, true
	case WordingPlain:
		return entry.Plain, true
	default:
		return "", false
	}
}
