// Package prefs 管理不屬於遊戲存檔的玩家顯示偏好。
package prefs

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/great-era-remake/internal/i18n"
)

// Preferences 預留 DESIGN-10 的四條呈現軸；目前 Wording 與
// MessageTime 已接玩家功能。
// 空字串表示沿用 preset／內建預設，不把尚未實作的軸假裝成可用。
type Preferences struct {
	Theme       string `json:"theme,omitempty"`
	Icons       string `json:"icons,omitempty"`
	Typography  string `json:"typography,omitempty"`
	Layout      string `json:"layout,omitempty"`
	Wording     string `json:"wording,omitempty"`
	Scale       int    `json:"scale,omitempty"`
	MessageTime int    `json:"message_time,omitempty"`
}

func Default() Preferences { return Preferences{Theme: "retro", Scale: 2, MessageTime: 5} }

// DefaultPath 遵循 XDG；os.UserConfigDir 在 Linux 會優先使用 XDG_CONFIG_HOME。
func DefaultPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("prefs: 找不到使用者設定目錄：%w", err)
	}
	return filepath.Join(dir, "dsds", "prefs.json"), nil
}

// Load 讀取偏好。不存在回內建預設且不算錯；壞檔或未知 wording 回預設與錯誤，
// 由程式明示警告後繼續，不把壞值部分套用。
func Load(path string) (Preferences, error) {
	p := Default()
	b, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return p, nil
	}
	if err != nil {
		return p, fmt.Errorf("prefs: 讀取 %s：%w", path, err)
	}
	var loaded Preferences
	if err := json.Unmarshal(b, &loaded); err != nil {
		return p, fmt.Errorf("prefs: 解析 %s：%w", path, err)
	}
	if loaded.Wording != "" {
		if _, err := i18n.ParseWordingMode(loaded.Wording); err != nil {
			return p, fmt.Errorf("prefs: %w", err)
		}
	}
	if loaded.MessageTime == 0 { // 相容舊版尚無此欄位的 prefs.json。
		loaded.MessageTime = 5
	}
	if loaded.MessageTime < 1 || loaded.MessageTime > 10 {
		return p, fmt.Errorf("prefs: 訊息時間要在 1..10，實得 %d", loaded.MessageTime)
	}
	return loaded, nil
}

// Save 以同目錄暫存檔 + rename 原子取代，避免程式中止留下半份 JSON。
func Save(path string, p Preferences) (err error) {
	if p.Wording != "" {
		if _, err := i18n.ParseWordingMode(p.Wording); err != nil {
			return fmt.Errorf("prefs: %w", err)
		}
	}
	if p.MessageTime < 1 || p.MessageTime > 10 {
		return fmt.Errorf("prefs: 訊息時間要在 1..10，實得 %d", p.MessageTime)
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("prefs: 建立設定目錄：%w", err)
	}
	f, err := os.CreateTemp(dir, ".prefs-*.json")
	if err != nil {
		return fmt.Errorf("prefs: 建立暫存檔：%w", err)
	}
	tmp := f.Name()
	defer func() { _ = f.Close(); _ = os.Remove(tmp) }()
	if err := f.Chmod(0o600); err != nil {
		return fmt.Errorf("prefs: 設定權限：%w", err)
	}
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(p); err != nil {
		return fmt.Errorf("prefs: 寫入 JSON：%w", err)
	}
	if err := f.Sync(); err != nil {
		return fmt.Errorf("prefs: 同步暫存檔：%w", err)
	}
	if err := f.Close(); err != nil {
		return fmt.Errorf("prefs: 關閉暫存檔：%w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		return fmt.Errorf("prefs: 原子取代：%w", err)
	}
	return nil
}
