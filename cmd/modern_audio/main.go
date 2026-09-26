// modern_audio 產生 Modern 純 Go cue／效果音的 WAV 技術預覽與可重生 JSON QA 報告。
//
// 輸出只供本機聽審與 waveform QA，不是正式 Ogg、發行資產或原版音樂轉檔。
package main

import (
	"encoding/binary"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	uiaudio "github.com/wicanr2/great-era-remake/internal/ui/audio"
)

var tracks = []uiaudio.Track{
	uiaudio.TrackScene, uiaudio.TrackStrategy, uiaudio.TrackMainTheme,
	uiaudio.TrackBattle1, uiaudio.TrackBattle2, uiaudio.TrackBattleAlt,
	uiaudio.TrackWall, uiaudio.TrackFinal,
}

var effects = []uiaudio.Effect{
	uiaudio.EffectConfirm, uiaudio.EffectCancel, uiaudio.EffectKey, uiaudio.EffectCommand,
	uiaudio.EffectBattleMove, uiaudio.EffectBattleAttack, uiaudio.EffectBattleHit,
	uiaudio.EffectBattleTurn,
}

func main() {
	out := flag.String("out", "workplace/promo/modern_audio_pcm", "WAV 技術預覽輸出目錄")
	sampleRate := flag.Int("sample-rate", 48000, "取樣率（8000..192000）")
	only := flag.String("track", "", "只輸出指定 track；空字串輸出八條 cue")
	withEffects := flag.Bool("effects", true, "是否同時輸出八類效果音")
	report := flag.String("report", "", "另輸出 JSON 技術 QA 報告（空字串不寫檔）")
	flag.Parse()

	if err := runWithReport(*out, *sampleRate, *only, *withEffects, *report); err != nil {
		fmt.Fprintln(os.Stderr, "錯誤：", err)
		os.Exit(1)
	}
}

func run(out string, sampleRate int, only string, withEffects bool) error {
	return runWithReport(out, sampleRate, only, withEffects, "")
}

type previewReport struct {
	Schema     int            `json:"schema"`
	SampleRate int            `json:"sample_rate"`
	Tracks     []previewEntry `json:"tracks"`
	Effects    []previewEntry `json:"effects"`
}

type previewEntry struct {
	Name     string              `json:"name"`
	Analysis uiaudio.PCMAnalysis `json:"analysis"`
}

func runWithReport(out string, sampleRate int, only string, withEffects bool, reportPath string) error {
	if err := os.MkdirAll(out, 0o755); err != nil {
		return fmt.Errorf("建立輸出目錄：%w", err)
	}
	report := previewReport{Schema: 1, SampleRate: sampleRate,
		Tracks: make([]previewEntry, 0), Effects: make([]previewEntry, 0)}
	want := tracks
	if only != "" {
		want = nil
		for _, track := range tracks {
			if string(track) == only {
				want = []uiaudio.Track{track}
				break
			}
		}
		if len(want) == 0 {
			return fmt.Errorf("未知 Modern track %q", only)
		}
	}
	for _, track := range want {
		pcm, err := uiaudio.RenderModernCue(track, sampleRate)
		if err != nil {
			return fmt.Errorf("產生 %s：%w", track, err)
		}
		path := filepath.Join(out, string(track)+".wav")
		if err := writeWAV(path, pcm, sampleRate); err != nil {
			return err
		}
		analysis, err := uiaudio.AnalyzeStereoPCM(pcm, sampleRate)
		if err != nil {
			return fmt.Errorf("分析 %s：%w", track, err)
		}
		report.Tracks = append(report.Tracks, previewEntry{Name: string(track), Analysis: analysis})
		fmt.Printf("寫出 %s（%.2f 秒）\n", path, uiaudio.ModernCueSeconds(track))
	}
	if withEffects {
		for _, effect := range effects {
			pcm, err := uiaudio.RenderModernEffect(effect, sampleRate)
			if err != nil {
				return fmt.Errorf("產生效果音 %s：%w", effect, err)
			}
			path := filepath.Join(out, "fx-"+string(effect)+".wav")
			if err := writeWAV(path, pcm, sampleRate); err != nil {
				return err
			}
			analysis, err := uiaudio.AnalyzeStereoPCM(pcm, sampleRate)
			if err != nil {
				return fmt.Errorf("分析效果音 %s：%w", effect, err)
			}
			report.Effects = append(report.Effects, previewEntry{Name: string(effect), Analysis: analysis})
			fmt.Printf("寫出 %s\n", path)
		}
	}
	if reportPath != "" {
		b, err := json.MarshalIndent(report, "", "  ")
		if err != nil {
			return fmt.Errorf("編碼 QA 報告：%w", err)
		}
		if err := os.MkdirAll(filepath.Dir(reportPath), 0o755); err != nil {
			return fmt.Errorf("建立 QA 報告目錄：%w", err)
		}
		if err := os.WriteFile(reportPath, append(b, '\n'), 0o644); err != nil {
			return fmt.Errorf("寫入 QA 報告：%w", err)
		}
		fmt.Printf("寫出 %s\n", reportPath)
	}
	return nil
}

func writeWAV(path string, pcm []byte, sampleRate int) error {
	if sampleRate < 8000 || sampleRate > 192000 || len(pcm) == 0 || len(pcm)%4 != 0 {
		return fmt.Errorf("WAV 參數無效：rate=%d bytes=%d", sampleRate, len(pcm))
	}
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("建立 %s：%w", path, err)
	}
	defer func() { _ = f.Close() }()
	write := func(v interface{}) error { return binary.Write(f, binary.LittleEndian, v) }
	if _, err := f.Write([]byte("RIFF")); err != nil {
		return err
	}
	if err := write(uint32(36 + len(pcm))); err != nil {
		return err
	}
	if _, err := f.Write([]byte("WAVEfmt ")); err != nil {
		return err
	}
	if err := write(uint32(16)); err != nil { // PCM fmt chunk
		return err
	}
	if err := write(uint16(1)); err != nil { // PCM
		return err
	}
	if err := write(uint16(2)); err != nil { // stereo
		return err
	}
	if err := write(uint32(sampleRate)); err != nil {
		return err
	}
	if err := write(uint32(sampleRate * 4)); err != nil { // byte rate
		return err
	}
	if err := write(uint16(4)); err != nil { // block align
		return err
	}
	if err := write(uint16(16)); err != nil { // bits per sample
		return err
	}
	if _, err := f.Write([]byte("data")); err != nil {
		return err
	}
	if err := write(uint32(len(pcm))); err != nil {
		return err
	}
	if _, err := f.Write(pcm); err != nil {
		return err
	}
	return nil
}
