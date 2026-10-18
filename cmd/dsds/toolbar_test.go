package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	userprefs "github.com/wicanr2/great-era-remake/internal/prefs"
	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
	uiresolution "github.com/wicanr2/great-era-remake/internal/ui/resolution"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

func TestLightInterfacePreferenceIsAtomicAndIndependent(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "file")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	a := &app{toolbar: newToolbarState(), eten: &assets.EtenFonts{}, wording: &i18n.WordingCatalog{}, room: commandRoomState{Mode: l.InterfaceModern}, themeMode: uitheme.ModeModern, preferences: userprefs.Default(), prefsPath: filepath.Join(blocker, "prefs.json")}
	before := a.preferences
	if err := a.setInterfaceAppearance(l.InterfaceModern, "light"); err == nil || a.preferences != before || a.lightInterface() {
		t.Fatal("failed preference write changed active appearance")
	}
	a.prefsPath = filepath.Join(dir, "prefs.json")
	if err := a.setInterfaceAppearance(l.InterfaceModern, "light"); err != nil || !a.lightInterface() {
		t.Fatal("light interface not selected", err)
	}
	want := before
	want.UIAppearance = "light"
	if a.preferences != want {
		t.Fatal("light setting changed unrelated preferences")
	}
	if err := a.setInterfacePreference(l.InterfaceOriginal); err != nil || a.lightInterface() || a.preferences.UIAppearance != "light" {
		t.Fatal("original interface forgot the light selection", err)
	}
	if err := a.toggleInterface(); err != nil || !a.lightInterface() {
		t.Fatal("returning to Modern lost appearance", err)
	}
	loaded, err := userprefs.Load(a.prefsPath)
	if err != nil || loaded.UIAppearance != "light" {
		t.Fatal("restart lost appearance", err)
	}
	if err := a.applyToolbarOption(1, 1); err != nil || a.lightInterface() {
		t.Fatal("explicit Modern dark did not restore dark", err)
	}
}

func TestPermanentToolbarLayoutAndPointerOffsets(tst *testing.T) {
	for _, tc := range []struct {
		mode                     l.InterfaceMode
		resolution               uiresolution.Mode
		w, h, x, y, wantX, wantY int
	}{
		{l.InterfaceCommand, uiresolution.ModeHigh, 1280, 768, 640, 408, 640, 360},
		{l.InterfaceOriginal, uiresolution.ModeHigh, 1280, 768, 640, 408, 320, 175},
		{l.InterfaceOriginal, uiresolution.ModeOriginal, 640, 374, 320, 199, 320, 175},
	} {
		a := &app{toolbar: newToolbarState(), room: commandRoomState{Mode: tc.mode}, resolution: tc.resolution, themeMode: uitheme.ModeModern}
		w, h := a.Layout(0, 0)
		if w != tc.w || h != tc.h {
			tst.Fatalf("layout=%dx%d", w, h)
		}
		x, y, ok := a.basePointerPosition(tc.x, tc.y)
		if !ok || x != tc.wantX || y != tc.wantY {
			tst.Fatalf("pointer=%d,%d,%v want=%d,%d", x, y, ok, tc.wantX, tc.wantY)
		}
		if _, _, ok := a.basePointerPosition(30, a.toolbarHeight()-1); ok {
			tst.Fatal("toolbar reached game")
		}
	}
}

func TestToolbarInputPreferenceIsAtomicAndCanRestoreMouse(t *testing.T) {
	dir := t.TempDir()
	blocker := filepath.Join(dir, "file")
	if err := os.WriteFile(blocker, []byte("not a directory"), 0600); err != nil {
		t.Fatal(err)
	}
	a := &app{preferences: userprefs.Default(), prefsPath: filepath.Join(blocker, "prefs.json"), pointer: pointerTracker{mouse: &pointerPress{}}, room: commandRoomState{Drag: true}}
	if err := a.setKeyboardOnly(true); err == nil || !a.gameMouseEnabled() {
		t.Fatal("failed save changed active input")
	}
	a.prefsPath = filepath.Join(dir, "prefs.json")
	if err := a.setKeyboardOnly(true); err != nil {
		t.Fatal(err)
	}
	if a.gameMouseEnabled() || a.pointer.mouse != nil || a.room.Drag {
		t.Fatal("keyboard mode retained game pointer")
	}
	loaded, err := userprefs.Load(a.prefsPath)
	if err != nil || !loaded.KeyboardOnly {
		t.Fatalf("saved input=%+v %v", loaded, err)
	}
	if err := a.setKeyboardOnly(false); err != nil || !a.gameMouseEnabled() {
		t.Fatalf("restore mouse: %v", err)
	}
}

func TestToolbarOffersOnlyCurrentAIAndDoesNotWritePreferences(t *testing.T) {
	wording, err := i18n.LoadWording("../../translations/zh-Hant")
	if err != nil {
		t.Fatal(err)
	}
	a := &app{toolbar: newToolbarState(), wording: wording, themeMode: uitheme.ModeModern, preferences: userprefs.Default(), prefsPath: filepath.Join(t.TempDir(), "prefs.json")}
	v := a.toolbarView()
	if len(v.Options[2]) != 1 || v.Selected[2] != 0 {
		t.Fatal("unapproved AI modes")
	}
	before := a.preferences
	if err := a.applyToolbarOption(2, 0); err != nil || a.preferences != before {
		t.Fatal("AI description modified preferences")
	}
	if _, err := os.Stat(a.prefsPath); !os.IsNotExist(err) {
		t.Fatal("AI description wrote a preference file")
	}
	if err := a.applyToolbarOption(2, 1); err == nil {
		t.Fatal("nonexistent AI option accepted")
	}
}

func TestRightClickUsesCancelForConfirmationsAndBackForBattle(t *testing.T) {
	for _, s := range []screen{screenQuit, screenTransferConfirm, screenRecruitConfirm, screenTrainConfirm, screenSaveConfirm, screenLoadConfirm} {
		if got := (&app{screen: s}).pointerBackAction(); got != actions.Cancel {
			t.Fatalf("screen %v=%v", s, got)
		}
	}
	for _, s := range []screen{screenBiography, screenBattle, screenCommand} {
		if got := (&app{screen: s}).pointerBackAction(); got != actions.Back {
			t.Fatalf("screen %v=%v", s, got)
		}
	}
}
