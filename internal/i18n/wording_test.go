package i18n

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestWordingCatalogCoversOriginalAndPlain(t *testing.T) {
	c, err := LoadWording(filepath.Join("..", "..", "translations", "zh-Hant"))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string][2]string{
		"transfer.select.partial": {"司令，欲調動何將？", "要調動哪位將領？"},
		"transfer.select.all":     {"司令，欲何將留守？", "哪些將領要留下？"},
		"transfer.target":         {"司令，欲調動至何省？", "要調往哪個省？"},
	}
	for key, pair := range want {
		if got, ok := c.Text(key, WordingOriginal); !ok || got != pair[0] {
			t.Errorf("%s original = %q, %v", key, got, ok)
		}
		if got, ok := c.Text(key, WordingPlain); !ok || got != pair[1] {
			t.Errorf("%s plain = %q, %v", key, got, ok)
		}
	}
	if _, ok := c.Text("不存在", WordingPlain); ok {
		t.Fatal("缺鍵不可暗中 fallback")
	}
}

func TestWordingRejectsIncompleteEntry(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "wording.json"),
		[]byte(`{"entries":{"broken":{"original":"原文","plain":""}}}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadWording(dir); err == nil {
		t.Fatal("缺白話文字應失敗即關閉")
	}
}

func TestWordingRejectsMissingRequiredKey(t *testing.T) {
	c, err := LoadWording(filepath.Join("..", "..", "translations", "zh-Hant"))
	if err != nil {
		t.Fatal(err)
	}
	delete(c.Entries, "transfer.resource.fuel")
	b, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "wording.json"), b, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadWording(dir); err == nil || !strings.Contains(err.Error(), "transfer.resource.fuel") {
		t.Fatalf("缺必要鍵應明確失敗，得到 %v", err)
	}
}

func TestParseWordingMode(t *testing.T) {
	for _, s := range []string{"original", "plain"} {
		if _, err := ParseWordingMode(s); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := ParseWordingMode("modern"); err == nil {
		t.Fatal("未知模式應回錯")
	}
}
