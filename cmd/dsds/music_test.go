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

func TestMilitarySelectionSurvivesModeChangesAndRestart(t *testing.T) {
	a := &app{audio: uiaudio.NewOff(), preferences: userprefs.Default(), prefsPath: filepath.Join(t.TempDir(), "prefs.json")}
	var played uiaudio.Track
	a.music.create = func(mode uiaudio.Mode, track uiaudio.Track) (*uiaudio.Manager, error) {
		played = track
		return uiaudio.NewOff(), nil
	}
	for _, track := range []uiaudio.Track{uiaudio.TrackHuanWoHeShan, uiaudio.TrackNationalRevolution, uiaudio.TrackEightHundredHeroes} {
		for _, mode := range []uiaudio.Mode{uiaudio.ModeModern, uiaudio.ModeRetro, uiaudio.ModeOff, uiaudio.ModeModern} {
			if err := a.setMusic(mode, track, true); err != nil {
				t.Fatal(err)
			}
			if mode == uiaudio.ModeRetro && uiaudio.HDOnlyTrack(played) {
				t.Fatal("retro requested an invented DOS military cue")
			}
			if mode == uiaudio.ModeModern && played != track {
				t.Fatalf("HD restored %s instead of %s", played, track)
			}
			p, err := userprefs.Load(a.prefsPath)
			if err != nil || p.MusicTrack != string(track) || p.Audio != string(mode) {
				t.Fatalf("restart preference %+v: %v", p, err)
			}
		}
	}
}

func TestMissingMilitaryTrackDoesNotBecomeProceduralMusic(t *testing.T) {
	old := uiaudio.NewOff()
	a := &app{audio: old, audioMode: uiaudio.ModeModern, preferences: userprefs.Default(), prefsPath: filepath.Join(t.TempDir(), "prefs.json")}
	a.music.dir = "../../assets/music/modern"
	before := a.preferences
	for _, track := range []uiaudio.Track{uiaudio.TrackHuanWoHeShan, uiaudio.TrackNationalRevolution, uiaudio.TrackEightHundredHeroes} {
		if err := a.setMusic(uiaudio.ModeModern, track, true); err == nil {
			t.Fatal("missing military arrangement accepted")
		}
		if a.audio != old || a.preferences != before {
			t.Fatal("missing track replaced active audio or preferences")
		}
	}
}

func TestUnavailableSavedMilitaryMusicDoesNotBlockStartup(t *testing.T) {
	a := &app{audio: uiaudio.NewOff(), preferences: userprefs.Default()}
	a.music.create = func(mode uiaudio.Mode, track uiaudio.Track) (*uiaudio.Manager, error) {
		if uiaudio.HDOnlyTrack(track) {
			return nil, errors.New("private arrangement absent")
		}
		return uiaudio.NewOff(), nil
	}
	if err := a.startMusic(uiaudio.ModeModern, uiaudio.TrackHuanWoHeShan); err != nil {
		t.Fatal(err)
	}
	if a.preferences.MusicTrack != string(uiaudio.TrackHuanWoHeShan) {
		t.Fatal("startup discarded the saved local choice")
	}
}
