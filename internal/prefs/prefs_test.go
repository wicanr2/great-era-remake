package prefs

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestMissingPreferencesUseDefaults(t *testing.T) {
	p, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil || p.Theme != "retro" || p.Scale != 2 || p.Wording != "" || p.MessageTime != 5 {
		t.Fatalf("預設偏好不符：%+v, %v", p, err)
	}
}

func TestSaveLoadRoundTripAndPermissions(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "prefs.json")
	want := Preferences{Theme: "retro", Wording: "plain", Scale: 2, MessageTime: 7}
	if err := Save(path, want); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil || got != want {
		t.Fatalf("round-trip = %+v, %v，預期 %+v", got, err, want)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o600 {
		t.Fatalf("prefs 權限 = %o，預期 600", info.Mode().Perm())
	}
	if matches, _ := filepath.Glob(filepath.Join(filepath.Dir(path), ".prefs-*.json")); len(matches) != 0 {
		t.Fatalf("遺留暫存檔：%v", matches)
	}
}

func TestBrokenOrUnknownPreferencesFailAsWholeFile(t *testing.T) {
	for name, body := range map[string]string{
		"broken":  `{`,
		"unknown": `{"theme":"modern","wording":"future"}`,
		"message": `{"message_time":11}`,
	} {
		t.Run(name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "prefs.json")
			if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
				t.Fatal(err)
			}
			p, err := Load(path)
			if err == nil || p != Default() {
				t.Fatalf("壞檔應整份回預設：%+v, %v", p, err)
			}
		})
	}
}

func TestThemeValidationFailsClosed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prefs.json")
	if err := Save(path, Preferences{Theme: "neon", Scale: 2, MessageTime: 5}); err == nil {
		t.Fatal("未知主題不得寫入偏好")
	}
	if err := os.WriteFile(path, []byte(`{"theme":"neon"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := Load(path); err == nil || got.Theme != "retro" {
		t.Fatalf("未知主題應回到預設：got=%+v err=%v", got, err)
	}
}

func TestResolutionValidationFailsClosed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prefs.json")
	if err := Save(path, Preferences{Resolution: "4k", MessageTime: 5}); err == nil {
		t.Fatal("未知解析度不得寫入偏好")
	}
	if err := os.WriteFile(path, []byte(`{"resolution":"4k"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	if got, err := Load(path); err == nil || got != Default() {
		t.Fatalf("未知解析度應回到預設：got=%+v err=%v", got, err)
	}
}

func TestDefaultPathUsesXDGConfigHome(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)
	path, err := DefaultPath()
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(path, dir+string(os.PathSeparator)) || filepath.Base(path) != "prefs.json" {
		t.Fatalf("XDG path = %q", path)
	}
}
