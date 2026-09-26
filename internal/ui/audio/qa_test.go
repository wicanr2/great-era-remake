package audio

import (
	"encoding/binary"
	"testing"
)

func TestAnalyzeStereoPCMReportsLoopAndMonoMetrics(t *testing.T) {
	pcm := make([]byte, 4*4)
	values := [][2]int16{{1000, 500}, {2000, 1000}, {-2000, -1000}, {1000, 500}}
	for i, pair := range values {
		binary.LittleEndian.PutUint16(pcm[i*4:i*4+2], uint16(pair[0]))
		binary.LittleEndian.PutUint16(pcm[i*4+2:i*4+4], uint16(pair[1]))
	}
	got, err := AnalyzeStereoPCM(pcm, 8000)
	if err != nil {
		t.Fatal(err)
	}
	if got.Frames != 4 || got.Channels != 2 || got.SampleRate != 8000 {
		t.Fatalf("PCM metadata=%+v", got)
	}
	if got.Peak != 2000 || got.FirstLastDelta != 0 || !got.NonSilent || got.Clipped {
		t.Fatalf("PCM 邊界／峰值=%+v", got)
	}
	if got.RMS <= 0 || got.MonoRMS <= 0 || got.SHA256 == "" {
		t.Fatalf("PCM 能量／雜湊未產生=%+v", got)
	}
}

func TestAnalyzeStereoPCMMarksHardClipAndRejectsMalformedInput(t *testing.T) {
	pcm := make([]byte, 4)
	binary.LittleEndian.PutUint16(pcm[:2], 0x8000)
	got, err := AnalyzeStereoPCM(pcm, 48000)
	if err != nil {
		t.Fatal(err)
	}
	if !got.Clipped || got.Peak != 32768 {
		t.Fatalf("應標記 -32768 clipping：%+v", got)
	}
	for _, tc := range []struct {
		name string
		pcm  []byte
		rate int
	}{
		{"empty", nil, 48000},
		{"partial-frame", []byte{1, 2, 3}, 48000},
		{"low-rate", make([]byte, 4), 7999},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := AnalyzeStereoPCM(tc.pcm, tc.rate); err == nil {
				t.Fatal("非法 PCM 應拒絕")
			}
		})
	}
}
