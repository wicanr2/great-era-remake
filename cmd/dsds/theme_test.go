package main

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/prefs"
	uiresolution "github.com/wicanr2/great-era-remake/internal/ui/resolution"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

type testThemeProvider struct{ name string }

func (t testThemeProvider) Name() string { return t.name }
func (t testThemeProvider) Tile(int) (uitheme.Bitmap, error) {
	return uitheme.Bitmap{}, nil
}
func (t testThemeProvider) Rail(int) (uitheme.Bitmap, error) {
	return uitheme.Bitmap{}, nil
}
func (t testThemeProvider) Unit(int) (uitheme.Bitmap, error) {
	return uitheme.Bitmap{}, nil
}
func (t testThemeProvider) ResourceIcon(int) (uitheme.Bitmap, error) {
	return uitheme.Bitmap{}, nil
}
func (t testThemeProvider) CommandIcon(int) (uitheme.Bitmap, error) {
	return uitheme.Bitmap{}, nil
}

func TestSetThemePreferenceSwapsTheWholeProviderAndPersists(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config", "prefs.json")
	retro := testThemeProvider{name: string(uitheme.ModeRetro)}
	modern := testThemeProvider{name: string(uitheme.ModeModern)}
	a := &app{
		retroTheme:       retro,
		modernTheme:      modern,
		themeMode:        uitheme.ModeRetro,
		battlefieldTheme: retro,
		unitTheme:        retro,
		preferences:      prefs.Default(),
		prefsPath:        path,
	}
	if err := a.setThemePreference(uitheme.ModeModern); err != nil {
		t.Fatal(err)
	}
	if a.themeMode != uitheme.ModeModern || a.battlefieldTheme != modern || a.unitTheme != modern || a.hudIcons != modern {
		t.Fatalf("主題 provider 未原子交換：mode=%s battlefield=%v unit=%v",
			a.themeMode, a.battlefieldTheme, a.unitTheme)
	}
	got, err := prefs.Load(path)
	if err != nil || got.Theme != string(uitheme.ModeModern) {
		t.Fatalf("主題偏好未持久化：%+v %v", got, err)
	}
}

type testThemeWithoutHUD struct{ name string }

func (t testThemeWithoutHUD) Name() string { return t.name }
func (t testThemeWithoutHUD) Tile(int) (uitheme.Bitmap, error) {
	return uitheme.Bitmap{}, nil
}
func (t testThemeWithoutHUD) Rail(int) (uitheme.Bitmap, error) {
	return uitheme.Bitmap{}, nil
}
func (t testThemeWithoutHUD) Unit(int) (uitheme.Bitmap, error) {
	return uitheme.Bitmap{}, nil
}

func TestSetThemePreferenceFailsClosedWithoutModernHUDProvider(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prefs.json")
	retro := testThemeProvider{name: string(uitheme.ModeRetro)}
	a := &app{
		retroTheme:  retro,
		modernTheme: testThemeWithoutHUD{name: string(uitheme.ModeModern)},
		themeMode:   uitheme.ModeRetro,
		preferences: prefs.Default(),
		prefsPath:   path,
	}
	if err := a.setThemePreference(uitheme.ModeModern); err == nil {
		t.Fatal("缺少 modern HUD provider 時應拒絕切換")
	}
	if a.themeMode != uitheme.ModeRetro || a.preferences.Theme != string(uitheme.ModeRetro) || a.hudIcons != nil {
		t.Fatalf("HUD provider 缺失時不應改變主題狀態：mode=%s prefs=%+v hud=%v",
			a.themeMode, a.preferences, a.hudIcons)
	}
}

type testThemeWithoutUnits struct{ name string }

func (t testThemeWithoutUnits) Name() string { return t.name }
func (t testThemeWithoutUnits) Tile(int) (uitheme.Bitmap, error) {
	return uitheme.Bitmap{}, nil
}
func (t testThemeWithoutUnits) Rail(int) (uitheme.Bitmap, error) {
	return uitheme.Bitmap{}, nil
}

func TestSetThemePreferenceFailsClosedWithoutUnitProvider(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prefs.json")
	retro := testThemeProvider{name: string(uitheme.ModeRetro)}
	a := &app{
		retroTheme:  retro,
		modernTheme: testThemeWithoutUnits{name: string(uitheme.ModeModern)},
		themeMode:   uitheme.ModeRetro,
		preferences: prefs.Default(),
		prefsPath:   path,
	}
	if err := a.setThemePreference(uitheme.ModeModern); err == nil {
		t.Fatal("缺少部隊圖示 provider 時應拒絕切換")
	}
	if a.themeMode != uitheme.ModeRetro || a.preferences.Theme != string(uitheme.ModeRetro) {
		t.Fatalf("失敗切換不應改變狀態：mode=%s prefs=%+v", a.themeMode, a.preferences)
	}
}

func TestSetResolutionPreferencePersistsAndTogglesWithoutWindow(t *testing.T) {
	path := filepath.Join(t.TempDir(), "prefs.json")
	a := &app{
		resolution:  uiresolution.ModeOriginal,
		preferences: prefs.Default(),
		prefsPath:   path,
	}
	if err := a.setResolutionPreference(uiresolution.ModeHigh); err != nil {
		t.Fatal(err)
	}
	if a.resolution != uiresolution.ModeHigh || a.preferences.Resolution != string(uiresolution.ModeHigh) {
		t.Fatalf("高解析偏好未套用：mode=%s prefs=%+v", a.resolution, a.preferences)
	}
	if err := a.toggleResolution(); err != nil {
		t.Fatal(err)
	}
	if a.resolution != uiresolution.ModeOriginal || a.preferences.Resolution != string(uiresolution.ModeOriginal) {
		t.Fatalf("解析度切換未回到原版：mode=%s prefs=%+v", a.resolution, a.preferences)
	}
	got, err := prefs.Load(path)
	if err != nil || got.Resolution != string(uiresolution.ModeOriginal) {
		t.Fatalf("解析度偏好未持久化：%+v %v", got, err)
	}
}
