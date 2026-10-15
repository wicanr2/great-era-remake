package main

import (
	"fmt"
	"os"

	userprefs "github.com/wicanr2/great-era-remake/internal/prefs"
	uiaudio "github.com/wicanr2/great-era-remake/internal/ui/audio"
)

var musicTracks = []uiaudio.Track{uiaudio.TrackScene, uiaudio.TrackStrategy, uiaudio.TrackMainTheme,
	uiaudio.TrackBattle1, uiaudio.TrackBattle2, uiaudio.TrackBattleAlt, uiaudio.TrackWall, uiaudio.TrackFinal}

type musicState struct {
	read   func(string) ([]byte, error)
	dir    string
	create func(uiaudio.Mode, uiaudio.Track) (*uiaudio.Manager, error)
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
		manager, e := uiaudio.NewModernTracks(set.Sources, track)
		if e == nil {
			return manager, nil
		}
		err = e
	}
	fmt.Fprintf(os.Stderr, "HD music fallback: %v\n", err)
	return uiaudio.NewModernProceduralTracks(0)
}

// Prepare at zero volume, persist, then exchange. Failure leaves active audio intact.
func (a *app) setMusic(mode uiaudio.Mode, track uiaudio.Track, persist bool) error {
	if _, err := uiaudio.ParseMode(string(mode)); err != nil {
		return err
	}
	want := track
	if want == "" {
		want = a.desiredAudioTrack()
		if want == "" {
			want = uiaudio.TrackScene
		}
	}
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
