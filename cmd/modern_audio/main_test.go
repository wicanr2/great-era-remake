package main

import (
	"encoding/binary"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestWriteWAVWritesStereoPCMHeader(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "cue.wav")
	pcm := make([]byte, 16)
	if err := writeWAV(path, pcm, 48000); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(b[:4]) != "RIFF" || string(b[8:16]) != "WAVEfmt " || string(b[36:40]) != "data" {
		t.Fatalf("WAV header=%q", b[:40])
	}
	if got := binary.LittleEndian.Uint32(b[24:28]); got != 48000 {
		t.Fatalf("sample rate=%d", got)
	}
	if got := binary.LittleEndian.Uint32(b[40:44]); got != uint32(len(pcm)) {
		t.Fatalf("data size=%d", got)
	}
}

func TestRunWithReportReproducesCueAndEffectAnalysis(t *testing.T) {
	dir := t.TempDir()
	reportPath := filepath.Join(dir, "qa", "report.json")
	if err := runWithReport(dir, 8000, "main-theme", true, reportPath); err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(reportPath)
	if err != nil {
		t.Fatal(err)
	}
	var report previewReport
	if err := json.Unmarshal(b, &report); err != nil {
		t.Fatal(err)
	}
	if report.Schema != 1 || report.SampleRate != 8000 || len(report.Tracks) != 1 || len(report.Effects) != len(effects) {
		t.Fatalf("QA 報告摘要=%+v", report)
	}
	if report.Tracks[0].Name != "main-theme" || report.Tracks[0].Analysis.Frames == 0 || report.Tracks[0].Analysis.SHA256 == "" {
		t.Fatalf("cue QA 資料=%+v", report.Tracks[0])
	}
	for _, entry := range report.Effects {
		if entry.Analysis.Frames == 0 || entry.Analysis.Clipped || !entry.Analysis.NonSilent {
			t.Fatalf("效果音 QA 不合格：%+v", entry)
		}
	}
}
