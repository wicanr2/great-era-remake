package audio

// 這個檔案是 Modern 音訊的安全 fallback：以短、原創的合成片段提供可播放的
// 旋律／氛圍與效果音。它不讀取、取樣或轉錄 workplace/orig 的 MUS、TIM、SDFA，
// 因而不會把「技術參考」誤變成原版旋律的發行資產。若玩家提供通過 manifest
// 驗證的 Ogg，Manager 仍優先播放 Ogg；只有缺 cue 時才走這條路徑。

import (
	"encoding/binary"
	"errors"
	"io"
	"math"

	ebitenaudio "github.com/hajimehoshi/ebiten/v2/audio"
)

const defaultModernSampleRate = 48000

var proceduralTracks = [...]Track{
	TrackScene, TrackStrategy, TrackMainTheme, TrackBattle1,
	TrackBattle2, TrackBattleAlt, TrackWall, TrackFinal,
}

// NewModernProceduralTracks 建立不依賴外部音檔的 Modern 音訊管理器。它是
// Android／展示包可直接播放的原創 24 小節 composition；日後換成 Ogg 只需把
// manifest 傳給 NewModernTracks，不必改畫面或事件接線。
func NewModernProceduralTracks(sampleRate int) (*Manager, error) {
	if sampleRate <= 0 {
		sampleRate = defaultModernSampleRate
	}
	if sampleRate < 8000 || sampleRate > 192000 {
		return nil, errors.New("audio: modern sample rate 必須在 8000..192000")
	}
	ctx := ebitenaudio.NewContext(sampleRate)
	return &Manager{
		mode:               ModeModern,
		ctx:                ebitenContext{ctx: ctx},
		pcmByTrack:         make(map[Track][]byte),
		proceduralFallback: true,
		current:            TrackScene,
		sampleRateValue:    sampleRate,
		volume:             0.7,
	}, nil
}

func isProceduralTrack(track Track) bool {
	for _, candidate := range proceduralTracks {
		if candidate == track {
			return true
		}
	}
	return false
}

const (
	modernCueBars          = 24
	modernCueBeatsPerBar   = 4
	modernCueDefaultBPM    = 112.0
	modernCueSecondsPerBar = modernCueBeatsPerBar * 60 / modernCueDefaultBPM
)

// modernCueSpec 是 runtime 內建的原創配器描述。它把同一個六音動機依玩家
// 情境轉成稀疏敘事、政略脈衝、戰鬥推進與結局回望；不讀取原版 MUS/TIM，
// 也不把原版音符宣稱成等價還原。
type modernCueSpec struct {
	root        float64
	bpm         float64
	motif       []int
	durations   []float64
	progression []int
	leadGain    float64
	pulseGain   float64
	brassGain   float64
	leadEvery   int
	voidBars    bool
	response    bool
}

var modernCampaignMotif = []int{0, 5, 3, 8, 7, 12} // D4 G4 F4 B♭4 A4 D5
var modernSceneMotif = []int{0, 5, 4, 7, 2}        // D4 G4 F♯4 A4 E4
var modernMotifDurations = []float64{0.5, 0.5, 1, 0.5, 0.5, 1}

func modernCueSpecFor(track Track) (modernCueSpec, bool) {
	progression := []int{0, 5, -2, 3, 0, 5, -2, 3, 0, 5, -2, 3, 0, 0, 0, 5, -2, 3, 0, 5, -2, 3, 0, 3}
	spec := modernCueSpec{
		root:        293.6648, // D4：新作動機的基準音，不是原版旋律轉錄。
		bpm:         modernCueDefaultBPM,
		motif:       modernCampaignMotif,
		durations:   modernMotifDurations,
		progression: progression,
		leadGain:    0.24,
		pulseGain:   0.04,
		brassGain:   0.08,
		leadEvery:   2,
		voidBars:    true,
		response:    true,
	}
	switch track {
	case TrackScene:
		spec.root, spec.bpm, spec.motif = 293.6648, 92, modernSceneMotif
		spec.durations = []float64{0.5, 0.5, 1, 0.5, 1.5}
		spec.progression = []int{0, 5, 10, 7, 0, 5, 3, 7}
		spec.leadGain, spec.pulseGain, spec.brassGain = 0.20, 0.008, 0
		spec.leadEvery, spec.voidBars, spec.response = 2, false, false
	case TrackStrategy:
		spec.leadGain, spec.pulseGain, spec.brassGain = 0.25, 0.07, 0.14
	case TrackMainTheme:
		spec.leadGain, spec.pulseGain, spec.brassGain = 0.29, 0.10, 0.22
		spec.leadEvery, spec.response = 1, false
	case TrackBattle1:
		spec.root, spec.leadGain, spec.pulseGain, spec.brassGain = 277.1826, 0.28, 0.17, 0.20
		spec.leadEvery, spec.voidBars, spec.response = 1, false, false
	case TrackBattle2:
		spec.root, spec.leadGain, spec.pulseGain, spec.brassGain = 329.6276, 0.27, 0.20, 0.24
		spec.leadEvery, spec.voidBars, spec.response = 1, false, true
	case TrackBattleAlt:
		spec.root, spec.leadGain, spec.pulseGain, spec.brassGain = 246.9417, 0.24, 0.16, 0.16
		spec.leadEvery, spec.response = 1, false
	case TrackWall:
		spec.root, spec.bpm, spec.leadGain, spec.pulseGain, spec.brassGain = 246.9417, 82, 0.16, 0.012, 0.02
		spec.leadEvery, spec.voidBars, spec.response = 4, false, false
	case TrackFinal:
		spec.root, spec.bpm, spec.leadGain, spec.pulseGain, spec.brassGain = 293.6648, 104, 0.31, 0.18, 0.28
		spec.leadEvery, spec.voidBars, spec.response = 1, true, true
	default:
		return modernCueSpec{}, false
	}
	return spec, true
}

func modernCueSeconds(track Track) float64 {
	spec, ok := modernCueSpecFor(track)
	if !ok || spec.bpm <= 0 {
		return modernCueBars * modernCueSecondsPerBar
	}
	return float64(modernCueBars*modernCueBeatsPerBar) * 60 / spec.bpm
}

// ModernCueSeconds 回傳程序音樂 cue 的名義長度，供離線預覽／QA 工具使用。
// 它不是原版音樂的時序證據。
func ModernCueSeconds(track Track) float64 { return modernCueSeconds(track) }

// RenderModernCue 以目前 runtime composition 產生 stereo signed-16 PCM，供
// 遊戲內播放與離線聽審工具共用同一份來源。輸出不是正式 Ogg 發行資產。
func RenderModernCue(track Track, sampleRate int) ([]byte, error) {
	return renderModernCue(track, sampleRate)
}

// renderModernCue 產生 24 小節、立體聲 16-bit PCM。每首曲目都保留同一個
// 原創六音動機，但以不同的速度、低音進行、脈衝密度與銅管進場點區分情境。
// 這是可直接播放的 runtime composition；若日後有正式清權 Ogg，Manager 仍優先
// 採用 Ogg，這個 fallback 不會把未審核二進位音源帶入發行包。
func renderModernCue(track Track, sampleRate int) ([]byte, error) {
	spec, ok := modernCueSpecFor(track)
	if !ok || !isProceduralTrack(track) {
		return nil, errors.New("audio: 未知 procedural modern cue")
	}
	if sampleRate <= 0 {
		sampleRate = defaultModernSampleRate
	}
	seconds := modernCueSeconds(track)
	frames := int(math.Round(seconds * float64(sampleRate)))
	if frames <= 0 || len(spec.progression) == 0 || len(spec.motif) == 0 || len(spec.durations) != len(spec.motif) {
		return nil, errors.New("audio: procedural cue 參數不完整")
	}
	pcm := make([]byte, frames*4) // stereo int16 little-endian
	for frame := 0; frame < frames; frame++ {
		t := float64(frame) / float64(sampleRate)
		beat := t * spec.bpm / 60
		bar := int(math.Floor(beat / modernCueBeatsPerBar))
		if bar >= modernCueBars {
			bar = modernCueBars - 1
		}
		barBeat := beat - float64(bar*modernCueBeatsPerBar)
		section := modernSectionLevel(bar, spec.voidBars)
		root := spec.root * math.Pow(2, float64(spec.progression[bar%len(spec.progression)])/12)

		// 木桌／紙張般的低密度基底：低音與五度保持清楚，讓 UI／旁白有空間。
		bass := (0.13*math.Sin(2*math.Pi*(root/4)*t) +
			0.035*math.Sin(2*math.Pi*(root/2)*1.003*t)) * section
		pad := (0.075*math.Sin(2*math.Pi*(root/2)*t) +
			0.035*math.Sin(2*math.Pi*(root*1.498)*t)) * (0.72 + section*0.28)

		lead := 0.0
		if spec.leadEvery > 0 && bar%spec.leadEvery == 0 && !(spec.voidBars && (bar == 12 || bar == 13)) {
			lead = modernMotifVoice(t, barBeat, spec.root, spec.motif, spec.durations, spec.leadGain*section,
				spec.response && bar%4 == 2, false)
		}
		brass := 0.0
		if bar >= 8 && !(spec.voidBars && (bar == 12 || bar == 13)) {
			brass = modernMotifVoice(t, barBeat, spec.root, spec.motif, spec.durations, spec.brassGain*section,
				bar%4 == 0, true)
		}
		pulse := modernPulse(t, beat, bar, spec.pulseGain*section)
		if spec.voidBars && (bar == 12 || bar == 13) {
			// 真空段只留一個低 D 心跳，讓推廣片旁白和遊戲效果音有位置。
			pulse *= 0.16
			bass *= 0.30
			pad *= 0.22
		}

		// 0.12 秒淡入／淡出只處理循環邊界的瞬態，不改動節拍或動機。
		env := 1.0
		fade := 0.12
		if t < fade {
			env *= t / fade
		}
		if t > seconds-fade {
			env *= (seconds - t) / fade
		}
		left := (bass + pad + lead + brass + pulse) * env
		right := (bass*0.96 + pad*0.90 + lead*0.86 + brass*0.78 + pulse*0.82) * env
		writeStereoFrame(pcm[frame*4:], left, right)
	}
	return pcm, nil
}

func modernSectionLevel(bar int, voidBars bool) float64 {
	switch {
	case voidBars && (bar == 12 || bar == 13):
		return 0.08
	case bar < 4:
		return 0.34
	case bar < 8:
		return 0.58
	case bar < 12:
		return 0.78
	case bar < 14:
		return 0.08
	case bar < 20:
		return 1.0
	default:
		return 0.68
	}
}

func modernMotifVoice(t, localBeat, root float64, motif []int, durations []float64, gain float64, reverse, brass bool) float64 {
	if localBeat < 0 || localBeat >= modernCueBeatsPerBar || len(motif) == 0 {
		return 0
	}
	cursor := 0.0
	index := -1
	for i, duration := range durations {
		if localBeat < cursor+duration {
			index = i
			break
		}
		cursor += duration
	}
	if index < 0 || index >= len(motif) {
		return 0
	}
	if reverse {
		index = len(motif) - 1 - index
	}
	semitones := motif[index]
	if reverse {
		semitones = -semitones / 2
	}
	if brass {
		semitones += 12
	}
	freq := root * math.Pow(2, float64(semitones)/12)
	if index >= len(durations) {
		return 0
	}
	duration := durations[index]
	inside := localBeat - cursor
	env := math.Min(1, inside/0.08)
	env = math.Min(env, math.Max(0, (duration-inside)/0.18))
	if env <= 0 {
		return 0
	}
	phase := 2 * math.Pi * freq * t
	if brass {
		// 銅管只用有限泛音，避免把純 Go fallback 做成持續白噪聲牆。
		return gain * env * (0.62*math.Sin(phase) + 0.22*math.Sin(phase*2.01) + 0.08*math.Sin(phase*3.02))
	}
	return gain * env * (0.78*math.Sin(phase) + 0.18*math.Sin(phase*2.01) + 0.06*math.Sin(phase*3.01))
}

func modernPulse(t, beat float64, bar int, gain float64) float64 {
	if gain <= 0 {
		return 0
	}
	beatIndex := int(math.Floor(beat))
	phase := beat - math.Floor(beat)
	if phase >= 0.14 {
		return 0
	}
	local := phase / 0.14
	if beatIndex%4 == 0 {
		freq := 74 - 30*local
		return gain * 0.92 * math.Sin(2*math.Pi*freq*t) * math.Exp(-local*5)
	}
	if beatIndex%4 == 2 {
		noise := math.Sin(float64((beatIndex+bar*7))*91.731 + t*137.0)
		return gain * (0.34*noise + 0.16*math.Sin(2*math.Pi*180*t)) * math.Exp(-local*8)
	}
	return gain * 0.10 * math.Sin(2*math.Pi*420*t) * math.Exp(-local*9)
}

// renderModernEffect 產生短促的一次性效果音。數值事件使用不同的頻率／
// 掃頻／雜訊比例，讓 UI、移動、攻擊與回合結束在聽感上可區分。
func renderModernEffect(effect Effect, sampleRate int) ([]byte, error) {
	if sampleRate <= 0 {
		sampleRate = defaultModernSampleRate
	}
	if sampleRate < 8000 || sampleRate > 192000 {
		return nil, errors.New("audio: effect sample rate 必須在 8000..192000")
	}
	base, span, seconds := 440.0, 0.0, 0.10
	switch effect {
	case EffectConfirm:
		base, span, seconds = 520, 170, 0.13
	case EffectCancel:
		base, span, seconds = 330, -130, 0.14
	case EffectKey:
		base, span, seconds = 760, -80, 0.055
	case EffectCommand:
		base, span, seconds = 260, 180, 0.18
	case EffectBattleMove:
		base, span, seconds = 180, 80, 0.11
	case EffectBattleAttack:
		base, span, seconds = 120, 520, 0.28
	case EffectBattleHit:
		base, span, seconds = 92, -48, 0.20
	case EffectBattleTurn:
		base, span, seconds = 300, 260, 0.24
	default:
		return nil, errors.New("audio: 未知 modern effect")
	}
	frames := int(float64(sampleRate) * seconds)
	pcm := make([]byte, frames*4)
	for frame := 0; frame < frames; frame++ {
		t := float64(frame) / float64(sampleRate)
		norm := t / seconds
		freq := base + span*norm
		env := math.Sin(math.Pi * norm)
		if effect == EffectBattleHit {
			// 低比例 deterministic noise 只作撞擊質感，不依賴系統亂數。
			env *= 0.75 + 0.25*math.Sin(float64(frame)*12.9898)
		}
		value := (math.Sin(2*math.Pi*freq*t)*0.30 +
			math.Sin(2*math.Pi*(freq*2.0)*t)*0.08) * env
		writeStereoFrame(pcm[frame*4:], value, value*0.86)
	}
	return pcm, nil
}

// RenderModernEffect 暴露同一份純 Go 效果音給離線 QA 工具；遊戲執行期仍由
// Manager.PlayEffect 管理 bounded player。
func RenderModernEffect(effect Effect, sampleRate int) ([]byte, error) {
	return renderModernEffect(effect, sampleRate)
}

func writeStereoFrame(dst []byte, left, right float64) {
	if len(dst) < 4 {
		return
	}
	left = math.Max(-0.85, math.Min(0.85, left))
	right = math.Max(-0.85, math.Min(0.85, right))
	binary.LittleEndian.PutUint16(dst[0:2], uint16(int16(left*32767)))
	binary.LittleEndian.PutUint16(dst[2:4], uint16(int16(right*32767)))
}

type pcmLoopReader struct {
	data []byte
	pos  int
}

func newPCMLoopReader(data []byte) io.Reader {
	return &pcmLoopReader{data: data}
}

func (r *pcmLoopReader) Read(p []byte) (int, error) {
	if r == nil || len(r.data) == 0 {
		return 0, errors.New("audio: procedural PCM 為空")
	}
	if len(p) == 0 {
		return 0, nil
	}
	for i := range p {
		p[i] = r.data[r.pos]
		r.pos++
		if r.pos >= len(r.data) {
			r.pos = 0
		}
	}
	return len(p), nil
}
