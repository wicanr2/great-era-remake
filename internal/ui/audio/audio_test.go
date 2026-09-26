package audio

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

type fakePlayer struct {
	played   int
	paused   int
	closed   int
	volume   float64
	closeErr error
}

func (p *fakePlayer) Play()               { p.played++ }
func (p *fakePlayer) Pause()              { p.paused++ }
func (p *fakePlayer) SetVolume(v float64) { p.volume = v }
func (p *fakePlayer) Close() error        { p.closed++; return p.closeErr }

type fakeContext struct {
	players []*fakePlayer
	err     error
}

func (c *fakeContext) NewPlayer(src io.Reader) (Player, error) {
	if c.err != nil {
		return nil, c.err
	}
	if _, err := io.ReadAll(src); err != nil {
		return nil, err
	}
	p := &fakePlayer{}
	c.players = append(c.players, p)
	return p, nil
}

func TestParseModeAndOffAreFailClosed(t *testing.T) {
	if got, err := ParseMode("off"); err != nil || got != ModeOff {
		t.Fatalf("off = %q, %v", got, err)
	}
	if got, err := ParseMode("retro"); err != nil || got != ModeRetro {
		t.Fatalf("retro = %q, %v", got, err)
	}
	if got, err := ParseMode("modern"); err != nil || got != ModeModern {
		t.Fatalf("modern = %q, %v", got, err)
	}
	if _, err := ParseMode(""); err == nil {
		t.Fatal("空模式應拒絕")
	}
	off := NewOff()
	if err := off.Start(2); err != nil {
		t.Fatalf("off 不應碰 volume／context：%v", err)
	}
	if err := off.Close(); err != nil {
		t.Fatal(err)
	}
}

func TestRetroPlayerLifecycleKeepsOnePlayer(t *testing.T) {
	ctx := &fakeContext{}
	m, err := newWithContext(ModeRetro, []byte{1, 2, 3, 4}, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Start(0.65); err != nil {
		t.Fatal(err)
	}
	if len(ctx.players) != 1 || ctx.players[0].played != 1 || ctx.players[0].volume != 0.65 {
		t.Fatalf("第一個 player = %+v", ctx.players)
	}
	first := ctx.players[0]
	if err := m.Start(0.4); err != nil {
		t.Fatal(err)
	}
	if first.paused != 1 || first.closed != 1 || len(ctx.players) != 2 || ctx.players[1].played != 1 {
		t.Fatalf("重播未關閉舊 player：first=%+v all=%+v", first, ctx.players)
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	if ctx.players[1].paused != 1 || ctx.players[1].closed != 1 {
		t.Fatalf("Close 未釋放第二個 player：%+v", ctx.players[1])
	}
	if err := m.Start(0.5); err == nil {
		t.Fatal("關閉後 Start 應拒絕")
	}
}

func TestRetroVolumeAndPlayerErrors(t *testing.T) {
	ctx := &fakeContext{}
	m, err := newWithContext(ModeRetro, []byte{1, 2}, ctx)
	if err != nil {
		t.Fatal(err)
	}
	for _, volume := range []float64{-0.01, 1.01} {
		if err := m.Start(volume); err == nil {
			t.Fatalf("volume %.2f 應拒絕", volume)
		}
	}
	ctx.err = errors.New("fake player failure")
	if err := m.Start(0.5); err == nil {
		t.Fatal("player 建立錯誤應回傳")
	}
}

func TestProceduralModernCueIsDeterministicAndBounded(t *testing.T) {
	a, err := renderModernCue(TrackScene, 8000)
	if err != nil {
		t.Fatal(err)
	}
	b, err := renderModernCue(TrackScene, 8000)
	if err != nil {
		t.Fatal(err)
	}
	if len(a) == 0 || string(a) != string(b) {
		t.Fatal("procedural cue 應為非空且 deterministic")
	}
	expectedFrames := int(math.Round(modernCueSeconds(TrackScene) * 8000))
	if len(a) != expectedFrames*4 {
		t.Fatalf("cue bytes=%d，預期 %d 小節 stereo int16", len(a), expectedFrames*4)
	}
	other, err := renderModernCue(TrackBattle1, 8000)
	if err != nil {
		t.Fatal(err)
	}
	if string(a) == string(other) {
		t.Fatal("不同情境不應共用完全相同波形")
	}
	for i := 0; i+1 < len(a); i += 2 {
		v := int16(binary.LittleEndian.Uint16(a[i : i+2]))
		if v > 32767 || v < -32767 {
			t.Fatalf("sample %d clipping: %d", i/2, v)
		}
	}
}

func TestProceduralCampaignCueKeepsVoidBeforeClimax(t *testing.T) {
	const sampleRate = 8000
	data, err := renderModernCue(TrackMainTheme, sampleRate)
	if err != nil {
		t.Fatal(err)
	}
	frames := len(data) / 4
	barFrames := frames / modernCueBars
	energy := func(firstBar, lastBar int) float64 {
		start, end := firstBar*barFrames, lastBar*barFrames
		if start < 0 {
			start = 0
		}
		if end > frames {
			end = frames
		}
		var sum float64
		for frame := start; frame < end; frame++ {
			value := int16(binary.LittleEndian.Uint16(data[frame*4 : frame*4+2]))
			sum += math.Abs(float64(value))
		}
		return sum / float64(end-start)
	}
	void := energy(12, 14)
	climax := energy(14, 20)
	if void <= 0 || climax <= void*1.5 {
		t.Fatalf("campaign cue 應有真空後高潮：void=%.1f climax=%.1f", void, climax)
	}
}

func TestAllProceduralModernCuesAreDeterministicAndUnclipped(t *testing.T) {
	const sampleRate = 8000
	seen := make(map[[32]byte]Track, len(proceduralTracks))
	for _, track := range proceduralTracks {
		first, err := renderModernCue(track, sampleRate)
		if err != nil {
			t.Fatalf("%s 產生失敗：%v", track, err)
		}
		second, err := renderModernCue(track, sampleRate)
		if err != nil {
			t.Fatalf("%s 第二次產生失敗：%v", track, err)
		}
		if len(first) == 0 || string(first) != string(second) {
			t.Fatalf("%s 應為非空且 deterministic", track)
		}
		wantFrames := int(math.Round(modernCueSeconds(track) * sampleRate))
		if len(first) != wantFrames*4 {
			t.Fatalf("%s bytes=%d，預期 %d", track, len(first), wantFrames*4)
		}
		var peak int
		var energy uint64
		for i := 0; i+1 < len(first); i += 2 {
			value := int(int16(binary.LittleEndian.Uint16(first[i : i+2])))
			magnitude := value
			if magnitude < 0 {
				magnitude = -magnitude
			}
			if magnitude > peak {
				peak = magnitude
			}
			if magnitude > 32767 {
				t.Fatalf("%s sample %d clipping：%d", track, i/2, value)
			}
			energy += uint64(magnitude)
		}
		if peak == 0 || energy == 0 {
			t.Fatalf("%s 不應為靜音", track)
		}
		digest := sha256.Sum256(first)
		if previous, exists := seen[digest]; exists {
			t.Fatalf("%s 與 %s 產生相同波形，情境 cue 未區分", track, previous)
		}
		seen[digest] = track
	}
}

func TestModernEffectsAreShortAndDistinct(t *testing.T) {
	confirm, err := renderModernEffect(EffectConfirm, 8000)
	if err != nil {
		t.Fatal(err)
	}
	attack, err := renderModernEffect(EffectBattleAttack, 8000)
	if err != nil {
		t.Fatal(err)
	}
	if len(confirm) == 0 || len(attack) == 0 || len(confirm) >= 8000*4 || len(attack) >= 8000*4 {
		t.Fatalf("effect 不應超過一秒：%d %d", len(confirm), len(attack))
	}
	if string(confirm) == string(attack) {
		t.Fatal("不同 effect 不應共用完全相同波形")
	}
}

func TestAllModernEffectsAreBoundedNonSilentAndUnclipped(t *testing.T) {
	const sampleRate = 8000
	for _, effect := range []Effect{
		EffectConfirm, EffectCancel, EffectKey, EffectCommand,
		EffectBattleMove, EffectBattleAttack, EffectBattleHit, EffectBattleTurn,
	} {
		data, err := renderModernEffect(effect, sampleRate)
		if err != nil {
			t.Fatalf("%s 產生失敗：%v", effect, err)
		}
		if len(data) == 0 || len(data) >= sampleRate*4 {
			t.Fatalf("%s 應為 0..1 秒內的非空 stereo PCM：%d bytes", effect, len(data))
		}
		var peak int
		for i := 0; i+1 < len(data); i += 2 {
			value := int(int16(binary.LittleEndian.Uint16(data[i : i+2])))
			magnitude := value
			if magnitude < 0 {
				magnitude = -magnitude
			}
			if magnitude > peak {
				peak = magnitude
			}
			if magnitude > 32767 {
				t.Fatalf("%s sample %d clipping：%d", effect, i/2, value)
			}
		}
		if peak == 0 {
			t.Fatalf("%s 不應為靜音", effect)
		}
	}
}

func TestEffectPlayerIsBoundedAndClosed(t *testing.T) {
	ctx := &fakeContext{}
	m, err := newWithContext(ModeModern, nil, ctx)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 20; i++ {
		if err := m.PlayEffect(EffectKey, 0.4); err != nil {
			t.Fatal(err)
		}
	}
	if len(m.effects) > 12 || len(ctx.players) != 20 {
		t.Fatalf("effect player 未受界限控制：active=%d created=%d", len(m.effects), len(ctx.players))
	}
	if err := m.Close(); err != nil {
		t.Fatal(err)
	}
	for i, p := range ctx.players {
		if p.closed == 0 {
			t.Fatalf("player %d 未關閉：%+v", i, p)
		}
	}
}

func TestRetroTrackSwitchUsesOneActivePlayer(t *testing.T) {
	ctx := &fakeContext{}
	m, err := newWithTrackPCMContext(map[Track][]byte{
		TrackScene:    {1, 2, 3, 4},
		TrackStrategy: {5, 6, 7, 8},
	}, TrackScene, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if m.CurrentTrack() != TrackScene {
		t.Fatalf("初始曲目 = %q", m.CurrentTrack())
	}
	if len(m.pcmByTrack) != 2 {
		t.Fatalf("測試曲目應在管理器建立時可重用快取：%d", len(m.pcmByTrack))
	}
	if err := m.Start(0.7); err != nil {
		t.Fatal(err)
	}
	first := ctx.players[0]
	if err := m.PlayTrack(TrackStrategy, 0.45); err != nil {
		t.Fatal(err)
	}
	if m.CurrentTrack() != TrackStrategy || first.paused != 1 || first.closed != 1 || len(ctx.players) != 2 {
		t.Fatalf("切曲未釋放舊 player：track=%q first=%+v players=%d", m.CurrentTrack(), first, len(ctx.players))
	}
	if got := string(m.pcmByTrack[TrackStrategy]); got != string([]byte{5, 6, 7, 8}) {
		t.Fatalf("切曲後不應丟失 PCM 快取：%x", m.pcmByTrack[TrackStrategy])
	}
	if err := m.PlayTrack("missing", 0.5); err == nil {
		t.Fatal("不存在曲目應拒絕")
	}
}

func TestRetroTrackSwitchFadeClosesOutgoingAfterBoundedFrames(t *testing.T) {
	ctx := &fakeContext{}
	m, err := newWithTrackPCMContext(map[Track][]byte{
		TrackScene:    {1, 2, 3, 4},
		TrackStrategy: {5, 6, 7, 8},
	}, TrackScene, ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := m.Start(0.7); err != nil {
		t.Fatal(err)
	}
	old := ctx.players[0]
	if err := m.PlayTrackWithFade(TrackStrategy, 0.6, 3); err != nil {
		t.Fatal(err)
	}
	if !m.TransitionActive() || len(ctx.players) != 2 || old.closed != 0 {
		t.Fatalf("切曲後應保留有限 transition：active=%v players=%d old=%+v", m.TransitionActive(), len(ctx.players), old)
	}
	for i := 0; i < 2; i++ {
		m.Update()
		if !m.TransitionActive() {
			t.Fatalf("第 %d 幀不應提前結束 transition", i+1)
		}
	}
	m.Update()
	if m.TransitionActive() || old.closed != 1 || old.paused != 1 {
		t.Fatalf("transition 未在固定幀數收束：active=%v old=%+v", m.TransitionActive(), old)
	}
	if got := ctx.players[1].volume; got != 0.6 {
		t.Fatalf("新曲目最終音量 = %.3f，預期 0.6", got)
	}
}

func TestNewRetroRejectsMalformedInputBeforeAudioContext(t *testing.T) {
	if _, err := NewRetro([]byte{1, 2}, nil, 48000); err == nil {
		t.Fatal("截斷 MUS 應拒絕")
	}
	if _, err := newWithContext(ModeRetro, nil, &fakeContext{}); err == nil {
		t.Fatal("空 PCM 應拒絕")
	}
}

func TestLoadModernTracksRequiresManifestProvenanceAndAllowsOptionalCue(t *testing.T) {
	dir := t.TempDir()
	data := []byte("not an ogg yet")
	if err := os.WriteFile(filepath.Join(dir, "scene.ogg"), data, 0o600); err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(data)
	manifest := ModernManifest{
		Schema:  1,
		Version: "test-1",
		Tracks: map[string]ModernManifestTrack{
			"modern_scene": {
				File: "scene.ogg", Author: "test", License: "CC0",
				SHA256: hex.EncodeToString(digest[:]),
			},
			"modern_story": {
				File: "story.ogg", Author: "test", License: "CC0",
				SHA256: strings.Repeat("0", 64),
			},
		},
	}
	encoded, err := json.Marshal(manifest)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "manifest.json"), encoded, 0o600); err != nil {
		t.Fatal(err)
	}
	set, err := LoadModernTracks(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(set.Sources) != 1 || len(set.Warnings) != 1 || !bytes.Equal(set.Sources[TrackScene].OGG, data) {
		t.Fatalf("modern manifest 結果 = sources=%d warnings=%v", len(set.Sources), set.Warnings)
	}
}

func TestLoadModernTracksRejectsHashAndPathErrors(t *testing.T) {
	dir := t.TempDir()
	for name, meta := range map[string]ModernManifestTrack{
		"hash": {File: "scene.ogg", Author: "test", License: "CC0", SHA256: strings.Repeat("1", 64)},
		"path": {File: "../scene.ogg", Author: "test", License: "CC0", SHA256: strings.Repeat("0", 64)},
	} {
		encoded, err := json.Marshal(ModernManifest{Schema: 1, Version: name, Tracks: map[string]ModernManifestTrack{"modern_scene": meta}})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "manifest.json"), encoded, 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadModernTracks(dir); err == nil {
			t.Fatalf("%s 應拒絕", name)
		}
	}
}

func TestNewModernRejectsMalformedOggBeforeContext(t *testing.T) {
	if _, err := NewModernTracks(map[Track]ModernTrackSource{
		TrackScene: {OGG: []byte("not ogg")},
	}, TrackScene); err == nil {
		t.Fatal("壞 Ogg 應在建立 audio.Context 前拒絕")
	}
}

func TestLoopReaderRepeatsDeclaredRange(t *testing.T) {
	src := bytes.NewReader([]byte("abcdef"))
	if _, err := src.Seek(2, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	r := &loopReader{src: src, start: 2, end: 5, pos: 2}
	buf := make([]byte, 8)
	n, err := r.Read(buf)
	if err != nil {
		t.Fatal(err)
	}
	if got := string(buf[:n]); got != "cdecdecd" {
		t.Fatalf("循環資料 = %q", got)
	}
}
