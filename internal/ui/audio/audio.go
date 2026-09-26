// Package audio 把純 Go PCM／Ogg 音源接到 Ebiten audio.Player。
//
// 這是 UI／平台適配層；MUS/TIM 解析與 OPL2 風格合成仍在 internal/audio，
// 因而可以在沒有音效裝置的 Docker 裡獨立測試。
package audio

import (
	"bytes"
	"errors"
	"fmt"
	"io"

	ebitenaudio "github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/wicanr2/great-era-remake/internal/audio/opl2"
)

// Mode 是音訊軸的目前窄切片。
type Mode string

const (
	ModeOff    Mode = "off"
	ModeRetro  Mode = "retro"
	ModeModern Mode = "modern"
)

// Track 是遊戲情境音樂的裝置無關名稱。檔名映射留在 cmd/dsds，
// 管理器只接收已讀入的 MUS/TIM bytes。
type Track string

const (
	TrackScene     Track = "scene"
	TrackStrategy  Track = "strategy"
	TrackMainTheme Track = "main-theme"
	TrackBattle1   Track = "battle-1"
	TrackBattle2   Track = "battle-2"
	TrackBattleAlt Track = "battle-alt"
	TrackWall      Track = "wall"
	TrackFinal     Track = "final"
)

// Effect 是 Modern 外殼的短效果音。它不是原版音效的逐樣本仿製；每一個
// 名稱只代表玩家已知的 UI／戰鬥事件，實際波形由純 Go 合成器產生。
type Effect string

const (
	EffectConfirm      Effect = "confirm"
	EffectCancel       Effect = "cancel"
	EffectKey          Effect = "key"
	EffectCommand      Effect = "command"
	EffectBattleMove   Effect = "battle-move"
	EffectBattleAttack Effect = "battle-attack"
	EffectBattleHit    Effect = "battle-hit"
	EffectBattleTurn   Effect = "battle-turn"
)

// TrackSource 是一首玩家自備曲目的原始 MUS/TIM 配對。
type TrackSource struct {
	MUS []byte
	TIM []byte
}

// ParseMode 只接受本規格已接通的三個值；空字串不偷偷選擇音效。
func ParseMode(value string) (Mode, error) {
	switch Mode(value) {
	case ModeOff, ModeRetro, ModeModern:
		return Mode(value), nil
	default:
		return "", fmt.Errorf("audio: 未知模式 %q（只接受 off、retro 或 modern）", value)
	}
}

// Player 是 Ebiten player 的最小邊界，測試可注入 fake 而不建立 audio.Context。
type Player interface {
	Play()
	Pause()
	SetVolume(float64)
	Close() error
}

type context interface {
	NewPlayer(io.Reader) (Player, error)
}

type ebitenContext struct{ ctx *ebitenaudio.Context }

func (c ebitenContext) NewPlayer(src io.Reader) (Player, error) {
	return c.ctx.NewPlayer(src)
}

// Manager 管理目前 BGM 與 bounded crossfade 期間的 outgoing player。retro 只保留
// 已產生的 PCM 副本；modern 的 Ogg bytes 由 manifest loader 複製後延遲解碼。
type Manager struct {
	mode               Mode
	ctx                context
	pcm                []byte
	sources            map[Track]TrackSource
	modernSources      map[Track]ModernTrackSource
	pcmByTrack         map[Track][]byte // 僅供測試注入／已渲染的窄快取
	proceduralFallback bool             // Ogg 缺少 cue 時使用原創純 Go loop；不讀原版 MUS/TIM
	current            Track
	sampleRateValue    int
	player             Player
	outgoing           Player   // 交叉淡出期間的舊 player
	effects            []Player // 有界的一次性短效果音 player
	fadeRemaining      int
	fadeTotal          int
	fadeFromVolume     float64
	volume             float64
	closed             bool
}

// NewOff 建立不初始化 Ebiten audio.Context 的管理器。
func NewOff() *Manager { return &Manager{mode: ModeOff} }

// NewRetro 解析一首玩家自備的 MUS/TIM，產生 PCM 並延後到此處才建立 audio.Context。
// sampleRate 為零時沿用 opl2.DefaultSampleRate。
func NewRetro(musData, timData []byte, sampleRate int) (*Manager, error) {
	return NewRetroTracks(map[Track]TrackSource{
		TrackScene: {MUS: musData, TIM: timData},
	}, TrackScene, sampleRate)
}

// NewRetroTracks 建立可依情境切換的 retro 管理器。只先渲染 initial，
// 其他曲目在 PlayTrack 時延遲解析／產生 PCM，避免把整套長曲目一次塞進記憶體。
func NewRetroTracks(sources map[Track]TrackSource, initial Track, sampleRate int) (*Manager, error) {
	if len(sources) == 0 {
		return nil, errors.New("audio: 至少需要一首 retro 曲目")
	}
	if initial == "" {
		return nil, errors.New("audio: 初始曲目不可為空")
	}
	source, ok := sources[initial]
	if !ok {
		return nil, fmt.Errorf("audio: 缺少初始曲目 %q", initial)
	}
	pcm, err := renderTrack(source, sampleRate)
	if err != nil {
		return nil, fmt.Errorf("audio: 產生曲目 %q：%w", initial, err)
	}
	if sampleRate == 0 {
		sampleRate = opl2.DefaultSampleRate
	}
	ctx := ebitenaudio.NewContext(sampleRate)
	copySources := make(map[Track]TrackSource, len(sources))
	for track, value := range sources {
		copySources[track] = TrackSource{MUS: append([]byte(nil), value.MUS...), TIM: append([]byte(nil), value.TIM...)}
	}
	rendered := map[Track][]byte{initial: append([]byte(nil), pcm...)}
	return &Manager{mode: ModeRetro, ctx: ebitenContext{ctx: ctx}, pcm: append([]byte(nil), pcm...),
		sources: copySources, pcmByTrack: rendered, current: initial,
		sampleRateValue: sampleRate, volume: 0.7}, nil
}

func renderTrack(source TrackSource, sampleRate int) ([]byte, error) {
	pcm, err := opl2.RenderAdLib(source.MUS, source.TIM, opl2.RenderOptions{SampleRate: sampleRate})
	if err != nil {
		return nil, fmt.Errorf("產生 PCM：%w", err)
	}
	return pcm, nil
}

// Mode 回傳目前音訊模式。
func (m *Manager) Mode() Mode {
	if m == nil {
		return ModeOff
	}
	return m.mode
}

// CurrentTrack 回傳目前已選的情境曲目；尚未選曲時回空字串。
func (m *Manager) CurrentTrack() Track {
	if m == nil {
		return ""
	}
	return m.current
}

// HasTrack 回報曲目是否已載入來源或測試用 PCM。它不解析、不建立 player，
// 讓 UI 可以在切換前採用缺檔 fallback。
func (m *Manager) HasTrack(track Track) bool {
	if m == nil || track == "" {
		return false
	}
	if _, ok := m.sources[track]; ok {
		return true
	}
	if _, ok := m.modernSources[track]; ok {
		return true
	}
	if m.mode == ModeModern && m.proceduralFallback && isProceduralTrack(track) {
		return true
	}
	return len(m.pcmByTrack[track]) > 0
}

// PlayTrack 解析並播放指定情境曲目。新曲目必須先成功解碼／建立 player，才會
// 停止舊 player；因此壞曲目不會讓目前音訊無聲。
func (m *Manager) PlayTrack(track Track, volume float64) error {
	return m.PlayTrackWithFade(track, volume, 0)
}

// PlayTrackWithFade 切換情境曲目；fadeFrames>0 時以 Update 驅動有限長度的
// 交叉淡入淡出。渲染／建立新 player 失敗時，舊曲目與 CurrentTrack 都保持不變。
// frame 數是 UI tick，不是音訊取樣數，避免把 OPL2 近似器的時序誤當成原版 oracle。
func (m *Manager) PlayTrackWithFade(track Track, volume float64, fadeFrames int) error {
	if m == nil || m.mode == ModeOff {
		return nil
	}
	if m.closed {
		return errors.New("audio: manager 已關閉")
	}
	if m.mode == ModeModern {
		if source, ok := m.modernSources[track]; ok {
			stream, err := decodeModernStream(source.OGG, m.sampleRate())
			if err != nil {
				return fmt.Errorf("audio: 解碼 modern 曲目 %q：%w", track, err)
			}
			reader, err := newLoopReader(stream, source)
			if err != nil {
				return fmt.Errorf("audio: modern 曲目 %q 的循環點無效：%w", track, err)
			}
			return m.startReader(track, reader, volume, fadeFrames, nil)
		}
		if m.proceduralFallback && isProceduralTrack(track) {
			pcm := m.pcmByTrack[track]
			if len(pcm) == 0 {
				var err error
				pcm, err = renderModernCue(track, m.sampleRate())
				if err != nil {
					return fmt.Errorf("audio: 產生 modern 曲目 %q：%w", track, err)
				}
				if m.pcmByTrack == nil {
					m.pcmByTrack = make(map[Track][]byte)
				}
				m.pcmByTrack[track] = append([]byte(nil), pcm...)
			}
			return m.startReader(track, newPCMLoopReader(pcm), volume, fadeFrames, pcm)
		}
		return fmt.Errorf("audio: 找不到 modern 曲目 %q", track)
	}
	var pcm []byte
	if m.pcmByTrack != nil {
		pcm = m.pcmByTrack[track]
	}
	if len(pcm) == 0 {
		source, ok := m.sources[track]
		if !ok {
			return fmt.Errorf("audio: 找不到曲目 %q", track)
		}
		var err error
		pcm, err = renderTrack(source, m.sampleRate())
		if err != nil {
			return fmt.Errorf("audio: 產生曲目 %q：%w", track, err)
		}
		if m.pcmByTrack == nil {
			m.pcmByTrack = make(map[Track][]byte)
		}
		m.pcmByTrack[track] = append([]byte(nil), pcm...)
	}
	if len(pcm) == 0 {
		return fmt.Errorf("audio: 找不到曲目 %q", track)
	}
	return m.startPCM(track, pcm, volume, fadeFrames)
}

func (m *Manager) sampleRate() int {
	if m == nil || m.sampleRateValue <= 0 {
		return opl2.DefaultSampleRate
	}
	return m.sampleRateValue
}

// Start 開始（或重新開始）目前 PCM。off 模式是 no-op；volume 必須在 0..1。
func (m *Manager) Start(volume float64) error {
	if m == nil || m.mode == ModeOff {
		return nil
	}
	if m.closed {
		return errors.New("audio: manager 已關閉")
	}
	if m.mode == ModeModern {
		return m.PlayTrackWithFade(m.current, volume, 0)
	}
	if volume < 0 || volume > 1 {
		return fmt.Errorf("audio: volume %.3f 不在 0..1", volume)
	}
	if m.ctx == nil || len(m.pcm) == 0 {
		return errors.New("audio: retro PCM 或 context 為空")
	}
	return m.startPCM(m.current, m.pcm, volume, 0)
}

func (m *Manager) startPCM(track Track, pcm []byte, volume float64, fadeFrames int) error {
	if volume < 0 || volume > 1 {
		return fmt.Errorf("audio: volume %.3f 不在 0..1", volume)
	}
	if len(pcm) == 0 {
		return errors.New("audio: retro PCM 為空")
	}
	return m.startReader(track, bytes.NewReader(pcm), volume, fadeFrames, pcm)
}

// PlayEffect 播放一個有界的短效果音。它只建立新的 player，不中斷目前
// BGM；超過上限時先關閉最舊效果，避免長時間點擊造成 player 無限累積。
func (m *Manager) PlayEffect(effect Effect, volume float64) error {
	if m == nil || m.mode == ModeOff {
		return nil
	}
	if m.closed {
		return errors.New("audio: manager 已關閉")
	}
	if volume < 0 || volume > 1 {
		return fmt.Errorf("audio: effect volume %.3f 不在 0..1", volume)
	}
	pcm, err := renderModernEffect(effect, m.sampleRate())
	if err != nil {
		return err
	}
	if m.ctx == nil {
		return errors.New("audio: effect audio context 為空")
	}
	p, err := m.ctx.NewPlayer(bytes.NewReader(pcm))
	if err != nil {
		return fmt.Errorf("audio: 建立 effect player：%w", err)
	}
	p.SetVolume(volume)
	p.Play()
	m.effects = append(m.effects, p)
	const maxEffects = 12
	if len(m.effects) > maxEffects {
		old := m.effects[0]
		old.Pause()
		_ = old.Close()
		m.effects = append([]Player(nil), m.effects[1:]...)
	}
	return nil
}

// startReader 建立並交換一個 player。reader 可是有限 PCM，也可是 modern
// Ogg 解碼器提供的可循環串流；建立失敗時保留舊曲目。
func (m *Manager) startReader(track Track, reader io.Reader, volume float64, fadeFrames int, pcm []byte) error {
	if volume < 0 || volume > 1 {
		return fmt.Errorf("audio: volume %.3f 不在 0..1", volume)
	}
	if m.ctx == nil || reader == nil {
		return errors.New("audio: audio context 或來源為空")
	}
	p, err := m.ctx.NewPlayer(reader)
	if err != nil {
		return fmt.Errorf("audio: 建立 player：%w", err)
	}
	// 建立新 player 成功後才交換狀態；這是音訊切換的 fail-closed 邊界。
	if m.outgoing != nil {
		m.outgoing.Pause()
		_ = m.outgoing.Close()
		m.outgoing = nil
	}
	old := m.player
	if pcm != nil {
		m.pcm = append(m.pcm[:0], pcm...)
	}
	m.current = track
	if fadeFrames > 0 && old != nil {
		p.SetVolume(0)
		p.Play()
		m.outgoing = old
		m.fadeRemaining = fadeFrames
		m.fadeTotal = fadeFrames
		m.fadeFromVolume = m.volume
		m.player = p
		m.volume = volume
		return nil
	}
	p.SetVolume(volume)
	p.Play()
	m.player, m.volume = p, volume
	if old != nil {
		old.Pause()
		if err := old.Close(); err != nil {
			return err
		}
	}
	return nil
}

// Update 推進一次有限長度的交叉淡入淡出；每幀最多調整一次兩個 player。
// 無 transition 時是 no-op，因而可安全放在 Update 的固定入口。
func (m *Manager) Update() {
	if m == nil || m.outgoing == nil || m.fadeTotal <= 0 {
		return
	}
	step := m.fadeTotal - m.fadeRemaining + 1
	progress := float64(step) / float64(m.fadeTotal)
	if progress > 1 {
		progress = 1
	}
	m.outgoing.SetVolume(m.fadeFromVolume * (1 - progress))
	m.player.SetVolume(m.volume * progress)
	m.fadeRemaining--
	if m.fadeRemaining <= 0 {
		m.outgoing.Pause()
		_ = m.outgoing.Close()
		m.outgoing = nil
		m.fadeTotal = 0
		m.fadeFromVolume = 0
		m.player.SetVolume(m.volume)
	}
}

// TransitionActive 回報是否仍有兩個情境 player 在交叉淡化。
func (m *Manager) TransitionActive() bool {
	return m != nil && m.outgoing != nil
}

// Stop 暫停並關閉目前 player；重複呼叫安全。
func (m *Manager) Stop() error {
	if m == nil {
		return nil
	}
	var firstErr error
	for _, p := range []Player{m.outgoing, m.player} {
		if p == nil {
			continue
		}
		p.Pause()
		if err := p.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	for _, p := range m.effects {
		if p == nil {
			continue
		}
		p.Pause()
		if err := p.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	m.effects = nil
	m.outgoing, m.player = nil, nil
	m.fadeRemaining, m.fadeTotal, m.fadeFromVolume = 0, 0, 0
	return firstErr
}

// Close 釋放 player。Ebiten Context 的全域生命週期由 Ebiten 管理，不在此重置。
func (m *Manager) Close() error {
	if m == nil {
		return nil
	}
	err := m.Stop()
	m.closed = true
	return err
}

// newWithContext 僅供本 package 測試注入 fake context；實際程式使用 NewOff/NewRetro。
func newWithContext(mode Mode, pcm []byte, ctx context) (*Manager, error) {
	parsed, err := ParseMode(string(mode))
	if err != nil {
		return nil, err
	}
	if parsed == ModeRetro && len(pcm) == 0 {
		return nil, errors.New("audio: retro PCM 不可為空")
	}
	return &Manager{mode: parsed, pcm: append([]byte(nil), pcm...), ctx: ctx, sampleRateValue: opl2.DefaultSampleRate, volume: 0.7}, nil
}

// newWithTrackPCMContext 僅供 package 測試注入已產生 PCM；正式路徑一律
// 由 NewRetroTracks 解析 MUS/TIM，避免測試為了 audio.Context 依賴顯示器。
func newWithTrackPCMContext(tracks map[Track][]byte, initial Track, ctx context) (*Manager, error) {
	if len(tracks) == 0 || initial == "" || len(tracks[initial]) == 0 {
		return nil, errors.New("audio: 測試曲目或初始 PCM 為空")
	}
	copyTracks := make(map[Track][]byte, len(tracks))
	for track, pcm := range tracks {
		copyTracks[track] = append([]byte(nil), pcm...)
	}
	return &Manager{mode: ModeRetro, ctx: ctx, pcm: append([]byte(nil), copyTracks[initial]...),
		pcmByTrack: copyTracks, current: initial, sampleRateValue: opl2.DefaultSampleRate, volume: 0.7}, nil
}
