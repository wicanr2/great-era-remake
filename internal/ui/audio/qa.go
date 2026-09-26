package audio

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"math"
)

// PCMAnalysis 是離線技術 QA 的結果。它只描述傳入的 stereo signed-16
// little-endian PCM，不代表人耳聽感、作者授權或正式發行核准。
type PCMAnalysis struct {
	SampleRate      int     `json:"sample_rate"`
	Channels        int     `json:"channels"`
	Frames          int     `json:"frames"`
	DurationSeconds float64 `json:"duration_seconds"`
	Peak            int     `json:"peak_abs_sample"`
	PeakNormalized  float64 `json:"peak_normalized"`
	RMS             float64 `json:"rms"`
	MonoRMS         float64 `json:"mono_rms"`
	FirstLastDelta  int     `json:"first_last_delta_abs"`
	Clipped         bool    `json:"clipped"`
	NonSilent       bool    `json:"non_silent"`
	SHA256          string  `json:"sha256"`
}

// AnalyzeStereoPCM 驗證並分析 stereo signed-16 little-endian PCM。它刻意不
// 依賴音訊裝置、Vorbis encoder 或外部命令，因此可在 Docker／CI 重生同一份
// Modern cue 與效果音的技術報告。
func AnalyzeStereoPCM(pcm []byte, sampleRate int) (PCMAnalysis, error) {
	if sampleRate < 8000 || sampleRate > 192000 {
		return PCMAnalysis{}, fmt.Errorf("audio: QA sample rate 必須在 8000..192000：%d", sampleRate)
	}
	if len(pcm) == 0 || len(pcm)%4 != 0 {
		return PCMAnalysis{}, fmt.Errorf("audio: QA 需要非空 stereo PCM：%d bytes", len(pcm))
	}
	frames := len(pcm) / 4
	var firstL, firstR, lastL, lastR int16
	var sum, monoSum float64
	peak := 0
	clipped := false
	for frame := 0; frame < frames; frame++ {
		off := frame * 4
		left := int16(binary.LittleEndian.Uint16(pcm[off : off+2]))
		right := int16(binary.LittleEndian.Uint16(pcm[off+2 : off+4]))
		if frame == 0 {
			firstL, firstR = left, right
		}
		if frame == frames-1 {
			lastL, lastR = left, right
		}
		for _, sample := range []int16{left, right} {
			magnitude := int(sample)
			if magnitude < 0 {
				magnitude = -magnitude
			}
			if magnitude > peak {
				peak = magnitude
			}
			if sample == int16(-32768) {
				clipped = true
			}
			value := float64(sample) / 32768.0
			sum += value * value
		}
		mono := (float64(left) + float64(right)) / (2 * 32768.0)
		monoSum += mono * mono
	}
	delta := func(a, b int16) int {
		d := int(a) - int(b)
		if d < 0 {
			d = -d
		}
		return d
	}
	edgeDelta := delta(firstL, lastL)
	if d := delta(firstR, lastR); d > edgeDelta {
		edgeDelta = d
	}
	digest := sha256.Sum256(pcm)
	return PCMAnalysis{
		SampleRate:      sampleRate,
		Channels:        2,
		Frames:          frames,
		DurationSeconds: float64(frames) / float64(sampleRate),
		Peak:            peak,
		PeakNormalized:  float64(peak) / 32768.0,
		RMS:             math.Sqrt(sum / float64(frames*2)),
		MonoRMS:         math.Sqrt(monoSum / float64(frames)),
		FirstLastDelta:  edgeDelta,
		Clipped:         clipped,
		NonSilent:       peak > 0,
		SHA256:          hex.EncodeToString(digest[:]),
	}, nil
}
