package main

import (
	"fmt"
	"os"

	userprefs "github.com/wicanr2/great-era-remake/internal/prefs"
	uiaudio "github.com/wicanr2/great-era-remake/internal/ui/audio"
)

var musicTracks = []uiaudio.Track{uiaudio.TrackScene, uiaudio.TrackStrategy, uiaudio.TrackMainTheme,
	uiaudio.TrackBattle1, uiaudio.TrackBattle2, uiaudio.TrackBattleAlt, uiaudio.TrackWall, uiaudio.TrackFinal,
	uiaudio.TrackHuanWoHeShan, uiaudio.TrackNationalRevolution, uiaudio.TrackEightHundredHeroes}

type musicState struct {
	read   func(string) ([]byte, error)
	dir    string
	create func(uiaudio.Mode, uiaudio.Track) (*uiaudio.Manager, error)
}

// A saved local-only choice must not prevent an engine-only package from opening.
// Keep the preference for later, and start an available scene cue this session.
func (a *app) startMusic(mode uiaudio.Mode, track uiaudio.Track) error {
	err := a.setMusic(mode, track, false)
	if err == nil || !uiaudio.HDOnlyTrack(track) {
		return err
	}
	fmt.Fprintf(os.Stderr, "Saved military music unavailable; playing scene music: %v\n", err)
	if err := a.setMusic(mode, "", false); err != nil {
		return err
	}
	a.preferences.MusicTrack = string(track)
	return nil
}

func (a *app) createMusic(mode uiaudio.Mode, track uiaudio.Track) (*uiaudio.Manager, error) {
	if a.music.create != nil {
		return a.music.create(mode, track)
	}
	if mode == uiaudio.ModeOff {
		return uiaudio.NewOff(), nil
	}
	if mode == uiaudio.ModeRetro {
		stems := []string{"SCENE", "STRATEGY", "MAINTHEM", "BATTLE1", "BATTLE2", "BT02", "WALL", "FINAL"}
		sources := map[uiaudio.Track]uiaudio.TrackSource{}
		for i, stem := range stems {
			mus, e1 := a.music.read(stem + ".MUS")
			tim, e2 := a.music.read(stem + ".TIM")
			if e1 == nil && e2 == nil {
				sources[musicTracks[i]] = uiaudio.TrackSource{MUS: mus, TIM: tim}
			}
		}
		if _, ok := sources[track]; !ok {
			s, _ := a.wordingText("toolbar.music.unavailable")
			return nil, fmt.Errorf("%s", s)
		}
		return uiaudio.NewRetroTracks(sources, track, 0)
	}
	set, err := uiaudio.LoadModernTracks(a.music.dir)
	if err == nil {
		if uiaudio.HDOnlyTrack(track) && len(set.Sources[track].OGG) == 0 {
			s, _ := a.wordingText("toolbar.music.unavailable")
			return nil, fmt.Errorf("%s: %s", s, track)
		}
		manager, e := uiaudio.NewModernTracks(set.Sources, track)
		if e == nil {
			return manager, nil
		}
		err = e
	}
	if uiaudio.HDOnlyTrack(track) {
		return nil, fmt.Errorf("HD music %s: %w", track, err)
	}
	fmt.Fprintf(os.Stderr, "HD music fallback: %v\n", err)
	return uiaudio.NewModernProceduralTracks(0)
}

// Prepare at zero volume, persist, then exchange. Failure leaves active audio intact.
func (a *app) setMusic(mode uiaudio.Mode, track uiaudio.Track, persist bool) error {
	if _, err := uiaudio.ParseMode(string(mode)); err != nil {
		return err
	}
	want := a.musicPlaybackTrack(mode, track)
	next, err := a.createMusic(mode, want)
	if err != nil {
		return err
	}
	if err = next.PlayTrack(want, 0); err != nil {
		_ = next.Close()
		return err
	}
	p := a.preferences
	p.Audio = string(mode)
	p.MusicTrack = string(track)
	if persist {
		if err = userprefs.Save(a.prefsPath, p); err != nil {
			_ = next.Close()
			return err
		}
	}
	old := a.audio
	a.audio, a.audioMode, a.audioTrack = next, mode, next.CurrentTrack()
	a.preferences = p
	a.audioWarned = nil
	if old != nil {
		_ = old.Close()
	}
	next.SetVolume(0.7)
	return nil
}

// A military selection survives Retro/Off; Retro plays the scene's DOS cue.
func (a *app) musicPlaybackTrack(mode uiaudio.Mode, selected uiaudio.Track) uiaudio.Track {
	if selected != "" && !(mode == uiaudio.ModeRetro && uiaudio.HDOnlyTrack(selected)) {
		return selected
	}
	if track := a.desiredAudioTrack(); track != "" {
		return track
	}
	return uiaudio.TrackScene
}
