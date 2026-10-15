package main

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	userprefs "github.com/wicanr2/great-era-remake/internal/prefs"
	uiaudio "github.com/wicanr2/great-era-remake/internal/ui/audio"
)

func TestMusicPreferenceFailureKeepsActiveManagerAndSelection(t *testing.T) {
	d := t.TempDir()
	blocked := filepath.Join(d, "file")
	os.WriteFile(blocked, []byte("x"), 0600)
	old := uiaudio.NewOff()
	a := &app{audio: old, audioMode: uiaudio.ModeModern, preferences: userprefs.Default(), prefsPath: filepath.Join(blocked, "prefs.json")}
	if err := a.setMusic(uiaudio.ModeOff, "", true); err == nil || a.audio != old || a.preferences.Audio != "modern" {
		t.Fatal("failed save changed active audio")
	}
	a.music.create = func(uiaudio.Mode, uiaudio.Track) (*uiaudio.Manager, error) { return nil, errors.New("device error") }
	if err := a.setMusic(uiaudio.ModeRetro, "", true); err == nil || a.audio != old || a.preferences.Audio != "modern" {
		t.Fatal("failed creation changed active audio")
	}
	a.music.create = nil
	a.prefsPath = filepath.Join(d, "prefs.json")
	if err := a.setMusic(uiaudio.ModeOff, uiaudio.TrackBattle2, true); err != nil {
		t.Fatal(err)
	}
	p, err := userprefs.Load(a.prefsPath)
	if err != nil || p.Audio != "off" || p.MusicTrack != "battle-2" {
		t.Fatalf("prefs=%+v %v", p, err)
	}
}

func TestOldPreferencesDefaultToHDMusic(t *testing.T) {
	p := filepath.Join(t.TempDir(), "prefs.json")
	os.WriteFile(p, []byte(`{"theme":"retro","layout":"original","message_time":5}`), 0600)
	got, err := userprefs.Load(p)
	if err != nil || got.Audio != "modern" || got.Theme != "retro" || got.Layout != "original" {
		t.Fatalf("prefs=%+v %v", got, err)
	}
}
