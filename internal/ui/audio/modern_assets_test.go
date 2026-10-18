package audio

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"
)

// Actual locally supplied assets exercise Vorbis seek at the real loop boundary.
// Public source checkouts may lack these ignored audio files.
func TestModernLocalAssetsRepeatThreeLoops(t *testing.T) {
	for _, dir := range []string{"../../../assets/music/modern", "../../../audio-preview"} {
		t.Run(filepath.Base(dir), func(t *testing.T) {
			if _, err := os.Stat(filepath.Join(dir, "manifest.json")); os.IsNotExist(err) {
				t.Skip("local audio preview not supplied")
			}
			set, err := LoadModernTracks(dir)
			if err != nil {
				t.Fatal(err)
			}
			want := 6
			if filepath.Base(dir) == "modern" {
				want = 8
				for _, track := range []Track{TrackMainTheme, TrackBattleAlt} {
					if len(set.Sources[track].OGG) == 0 {
						t.Fatalf("release cue %s must have real Ogg bytes instead of procedural fallback", track)
					}
				}
			}
			if len(set.Sources) != want || len(set.Warnings) != 0 {
				t.Fatalf("need %d actual cues without warnings: %d, %v", want, len(set.Sources), set.Warnings)
			}
			for track, source := range set.Sources {
				t.Run(string(track), func(t *testing.T) {
					stream, err := decodeModernStream(source.OGG, 0)
					if err != nil {
						t.Fatal(err)
					}
					pcm, err := io.ReadAll(stream)
					if err != nil || len(pcm) == 0 || int64(len(pcm)) != stream.Length() {
						t.Fatalf("decoded size %d, declared %d: %v", len(pcm), stream.Length(), err)
					}
					loop, err := newLoopReader(stream, source)
					if err != nil {
						t.Fatal(err)
					}
					for repeat := 0; repeat < 3; repeat++ {
						got := make([]byte, len(pcm))
						if _, err := io.ReadFull(loop, got); err != nil {
							t.Fatal(err)
						}
						if !bytes.Equal(got, pcm) {
							t.Fatalf("loop %d differs after real Vorbis seek", repeat+1)
						}
					}
				})
			}
		})
	}
}

func TestModernLoopMetadataResamplesToSharedOutputRate(t *testing.T) {
	dir := "../../../assets/music/modern"
	if _, err := os.Stat(filepath.Join(dir, "manifest.json")); os.IsNotExist(err) {
		t.Skip("local music unavailable")
	}
	set, err := LoadModernTracks(dir)
	if err != nil {
		t.Fatal(err)
	}
	source := set.Sources[TrackScene]
	native, err := decodeModernStream(source.OGG, 0)
	if err != nil {
		t.Fatal(err)
	}
	source.LoopStart = int64(native.SampleRate())
	source.LoopLength = int64(native.SampleRate() * 2)
	stream, scaled, err := decodeModernSource(source, 48000)
	if err != nil || scaled.LoopStart != 48000 || scaled.LoopLength != 96000 || stream.SampleRate() != 48000 {
		t.Fatalf("resampled=%+v %v", scaled, err)
	}
	loop, err := newLoopReader(stream, scaled)
	if err != nil {
		t.Fatal(err)
	}
	buf := make([]byte, 96000*4*3)
	if _, err := io.ReadFull(loop, buf); err != nil {
		t.Fatal(err)
	}
}

func TestLocalOriginalHDArrangementsLoopAtDeviceRates(t *testing.T) {
	dir := os.Getenv("DSDS_LOCAL_HD_MUSIC")
	if dir == "" {
		t.Skip("private full arrangements not supplied")
	}
	set, err := LoadModernTracks(dir)
	if err != nil || len(set.Sources) != 11 || len(set.Warnings) != 0 {
		t.Fatalf("full HD catalog: %d, %v, %v", len(set.Sources), set.Warnings, err)
	}
	for track, source := range set.Sources {
		t.Run(string(track), func(t *testing.T) {
			for _, rate := range []int{44100, 48000} {
				stream, scaled, err := decodeModernSource(source, rate)
				if err != nil {
					t.Fatalf("%s @ %d: %v", track, rate, err)
				}
				pcm, err := io.ReadAll(stream)
				if err != nil || len(pcm) == 0 {
					t.Fatalf("decode %s: %v", track, err)
				}
				loop, err := newLoopReader(stream, scaled)
				if err != nil {
					t.Fatalf("loop %s @ %d: %v", track, rate, err)
				}
				n := int(scaled.LoopLength * 4)
				if n <= 8192 || n > len(pcm) {
					t.Fatalf("loop range %d/%d", n, len(pcm))
				}
				for repeat := 0; repeat < 3; repeat++ {
					got := make([]byte, n)
					if _, err := io.ReadFull(loop, got); err != nil {
						t.Fatal(err)
					}
					if !bytes.Equal(got, pcm[:n]) {
						t.Fatalf("%s @ %d loop %d differs", track, rate, repeat)
					}
				}
			}
		})
	}
}
