package opl2

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/audio/mus"
)

func testBank() mus.TimbreBank {
	return mus.TimbreBank{Instruments: []mus.Instrument{{
		Modulator: mus.Operator{Multiple: 1, Attack: 1, Level: 12, Connection: 0},
		Carrier:   mus.Operator{Multiple: 1, Attack: 1, Level: 8, Release: 2},
	}}}
}

func TestChipDeterministicStereoPCM(t *testing.T) {
	chip, err := NewChip(48000, testBank())
	if err != nil {
		t.Fatal(err)
	}
	if err := chip.SetProgram(0, 0); err != nil {
		t.Fatal(err)
	}
	if err := chip.SetVolume(0, 100); err != nil {
		t.Fatal(err)
	}
	if err := chip.NoteOn(0, 60, 100); err != nil {
		t.Fatal(err)
	}
	got := chip.Render(256)
	if len(got) != 256*4 {
		t.Fatalf("PCM bytes=%d，預期 %d", len(got), 256*4)
	}
	for i := 0; i < len(got); i += 4 {
		if string(got[i:i+2]) != string(got[i+2:i+4]) {
			t.Fatalf("frame %d 不是 mono folded stereo", i/4)
		}
	}
	h := sha256.Sum256(got)
	if want := "d87fa84950725e9e76716aa5bbaf212f24e72da280ffe1736863faf0b5ea7338"; hex.EncodeToString(h[:]) != want {
		t.Fatalf("deterministic PCM digest=%x，預期 %s", h, want)
	}
}

func TestNormalizeEventsKeepsConfirmedMUSSemantics(t *testing.T) {
	song := mus.Song{Header: mus.Header{TickBeat: 240, BasicTempo: 120}, Events: []mus.Event{
		{Tick: 0, Status: 0xc0, Data: []byte{0}},
		{Tick: 2, Status: 0x90, Data: []byte{60, 100}},
		{Tick: 5, Status: 0xa0, Data: []byte{90}},
		{Tick: 7, Status: 0xe0, Data: []byte{0, 64}},
		{Tick: 9, Status: 0x90, Data: []byte{60, 0}},
		{Tick: 10, Status: 0xf0, Data: []byte{0x7f, 0x00, 0x01, 0x00}},
		{Tick: 11, Status: 0xfc},
	}}
	got, err := NormalizeEvents(song)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 6 || got[0].Kind != EventProgram || got[1].Kind != EventNoteOn ||
		got[4].Kind != EventNoteOff || got[5].Kind != EventTempo || got[5].TempoMul != 1 {
		t.Fatalf("typed events=%+v", got)
	}
}

func TestNormalizeEventsRejectsZeroTickRate(t *testing.T) {
	if _, err := NormalizeEvents(mus.Song{}); err == nil {
		t.Fatal("零 tick rate 應 fail-closed")
	}
}

func TestRenderSongHonorsMaxFramesAndBounds(t *testing.T) {
	song := mus.Song{Header: mus.Header{TickBeat: 240, BasicTempo: 120}, ActualTick: 480,
		Events: []mus.Event{{Tick: 0, Status: 0xc0, Data: []byte{0}}, {Tick: 1, Status: 0x90, Data: []byte{60, 100}}}}
	got, err := RenderSong(song, testBank(), RenderOptions{SampleRate: 48000, MaxFrames: 37})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 37*4 {
		t.Fatalf("bounded PCM bytes=%d", len(got))
	}
	if _, err := RenderSong(song, testBank(), RenderOptions{MaxFrames: -1}); err == nil {
		t.Fatal("負 MaxFrames 應 fail-closed")
	}
}

func TestRenderSongPercussionChannelsProducePCM(t *testing.T) {
	song := mus.Song{Header: mus.Header{TickBeat: 240, BasicTempo: 120}, ActualTick: 480,
		Events: []mus.Event{
			{Tick: 0, Status: 0x96, Data: []byte{36, 110}}, // channel 6：kick
			{Tick: 0, Status: 0x97, Data: []byte{38, 100}}, // channel 7：snare
			{Tick: 0, Status: 0x98, Data: []byte{42, 90}},  // channel 8：hat
		}}
	pcm, err := RenderSong(song, testBank(), RenderOptions{SampleRate: 8000, MaxFrames: 256})
	if err != nil {
		t.Fatal(err)
	}
	nonZero := 0
	for i := 0; i+1 < len(pcm); i += 2 {
		if pcm[i] != 0 || pcm[i+1] != 0 {
			nonZero++
		}
	}
	if nonZero == 0 {
		t.Fatal("OPL2 rhythm channels 6..8 產出的 PCM 全為零")
	}
}

func TestRenderAdLibRealSceneProducesPCM(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	musData, err := os.ReadFile(filepath.Join(gameDir, "SCENE.MUS"))
	if err != nil {
		t.Fatal(err)
	}
	timData, err := os.ReadFile(filepath.Join(gameDir, "SCENE.TIM"))
	if err != nil {
		t.Fatal(err)
	}
	pcm, err := RenderAdLib(musData, timData, RenderOptions{SampleRate: 8000, MaxFrames: 1024})
	if err != nil {
		t.Fatal(err)
	}
	if len(pcm) != 1024*4 {
		t.Fatalf("直接 AdLib/OPL2 renderer 回傳 %d bytes，預期 %d", len(pcm), 1024*4)
	}
	nonZero := 0
	for i := 0; i < len(pcm); i += 2 {
		if pcm[i] != 0 || pcm[i+1] != 0 {
			nonZero++
		}
	}
	if nonZero == 0 {
		t.Fatal("SCENE.MUS/TIM 的直接 AdLib/OPL2 路徑沒有可聽 PCM")
	}
}

func TestProgramRegisterWritesPacksTIMFields(t *testing.T) {
	in := mus.Instrument{
		Modulator: mus.Operator{
			KSL: 3, Multiple: 0xA, Feedback: 7, Attack: 0xB, Sustain: 0xC,
			EG: 1, Decay: 0xD, Release: 0xE, Level: 0x2F, AM: 1,
			Vibrato: 1, KSR: 1, Connection: 1,
		},
		Carrier: mus.Operator{
			KSL: 2, Multiple: 0x9, Feedback: 0xFFFF, Attack: 1, Sustain: 2,
			EG: 1, Decay: 3, Release: 4, Level: 0x11, AM: 1,
			Vibrato: 1, KSR: 1, Connection: 0xFFFF,
		},
		ModWave: 2, CarWave: 3,
	}
	got, err := ProgramRegisterWrites(1, in)
	if err != nil {
		t.Fatal(err)
	}
	want := []RegisterWrite{
		{0x21, 0xFA}, {0x41, 0xEF}, {0x61, 0xBD}, {0x81, 0xCE}, {0xE1, 0x02},
		{0x24, 0xF9}, {0x44, 0x91}, {0x64, 0x13}, {0x84, 0x24}, {0xE4, 0x03},
		{0xC1, 0x0F},
	}
	if len(got) != len(want) {
		t.Fatalf("寄存器寫入數=%d，預期 %d：%+v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("第 %d 筆寄存器=%+v，預期 %+v；全部=%+v", i, got[i], want[i], got)
		}
	}
}

func TestProgramRegisterWritesRejectsInvalidChannel(t *testing.T) {
	for _, channel := range []int{-1, MaxChannels} {
		if _, err := ProgramRegisterWrites(channel, mus.Instrument{}); err == nil {
			t.Fatalf("聲道 %d 應拒絕", channel)
		}
	}
}
