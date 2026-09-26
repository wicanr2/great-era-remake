// Package opl2 提供不依賴裝置的、純 Go OPL2 風格離線合成。
//
// 這不是 SDFA.EXE 的逐暫存器重現：`An` 音量、鼓組與 envelope 仍是明示的
// remake approximation。保留這個邊界，才能在取得 DOSBox oracle 後替換合成器，
// 而不牽動 MUS/TIM parser 或遊戲規則層。
package opl2

import (
	"encoding/binary"
	"errors"
	"fmt"
	"math"

	"github.com/wicanr2/great-era-remake/internal/audio/mus"
)

const (
	// DefaultSampleRate 是後續 Ebiten audio adapter 的預設 PCM rate。
	DefaultSampleRate = 48000
	// MaxChannels 是 OPL2 的九個旋律聲道。
	MaxChannels = 9
	// PercussionFirstChannel..PercussionLastChannel 是 MUS/AdLib 慣用的五個
	// OPL2 rhythm channels（6..10）。它們走純 Go 的打擊音色，不依賴 SDFA。
	PercussionFirstChannel = 6
	PercussionChannels     = 5
)

// RenderOptions 控制離線 PCM 產量。零值使用 48 kHz，並以歌曲實際 tick 結尾。
type RenderOptions struct {
	SampleRate int
	MaxFrames  int
}

// VoiceEvent 是套用到 Chip 的已正規化事件，供測試與未來串流 adapter 追查。
type VoiceEvent struct {
	Tick     uint32
	Channel  uint8
	Kind     VoiceEventKind
	Value1   uint8
	Value2   uint8
	Program  int
	TempoMul float64
}

// VoiceEventKind 是 MUS 事件的音源層語意；未知事件不會被假造為這裡的種類。
type VoiceEventKind uint8

const (
	EventProgram VoiceEventKind = iota + 1
	EventNoteOn
	EventNoteOff
	EventVolume
	EventPitchBend
	EventTempo
)

// OperatorParams 是由 TIM 13-word operator 欄位擷取的合成參數。
type OperatorParams struct {
	Multiple uint16
	Attack   uint16
	Decay    uint16
	Sustain  uint16
	Release  uint16
	Level    uint16
}

type voice struct {
	active       bool
	note         uint8
	velocity     float64
	volume       float64
	bend         float64
	modPhase     float64
	carPhase     float64
	modEnv       float64
	carEnv       float64
	release      float64
	instrument   mus.Instrument
	attack       float64
	decay        float64
	sustain      float64
	releaseRate  float64
	modLevel     float64
	carrierLevel float64
}

// drumVoice 是 OPL2 rhythm mode 的可聽 fallback。它刻意只保留播放需要的
// phase／noise／envelope，不宣稱與原版 SDFA 的 operator/register 逐值等價。
type drumVoice struct {
	active    bool
	note      uint8
	velocity  float64
	volume    float64
	phase     float64
	envelope  float64
	decay     float64
	noiseSeed uint32
}

// Chip 是九聲道、兩 operator 的 deterministic 軟體合成器。
type Chip struct {
	sampleRate int
	voices     [MaxChannels]voice
	drums      [PercussionChannels]drumVoice
	programs   []mus.Instrument
}

// NewChip 建立一枚純 Go Chip。bank 只讀取並複製 instrument slice。
func NewChip(sampleRate int, bank mus.TimbreBank) (*Chip, error) {
	if sampleRate <= 0 || sampleRate > 192000 {
		return nil, fmt.Errorf("opl2: sample rate %d 不在 1..192000", sampleRate)
	}
	if len(bank.Instruments) == 0 {
		return nil, errors.New("opl2: TIM 沒有任何音色")
	}
	programs := append([]mus.Instrument(nil), bank.Instruments...)
	chip := &Chip{sampleRate: sampleRate, programs: programs}
	for i := range chip.drums {
		chip.drums[i].volume = 1
	}
	return chip, nil
}

// SampleRate 回傳 Chip 的固定取樣率。
func (c *Chip) SampleRate() int { return c.sampleRate }

// SetProgram 套用一個 TIM 音色到旋律聲道。
func (c *Chip) SetProgram(channel, program int) error {
	if c == nil {
		return errors.New("opl2: nil chip")
	}
	if channel < 0 || channel >= MaxChannels {
		return fmt.Errorf("opl2: 聲道 %d 超出 0..%d", channel, MaxChannels-1)
	}
	if program < 0 || program >= len(c.programs) {
		return fmt.Errorf("opl2: 音色 %d 超出 0..%d", program, len(c.programs)-1)
	}
	c.voices[channel].instrument = c.programs[program]
	setEnvelope(&c.voices[channel])
	return nil
}

// NoteOn 啟動一個旋律聲道。velocity 0 依 MUS 慣例等同 NoteOff。
func (c *Chip) NoteOn(channel, note, velocity int) error {
	if err := c.validChannel(channel); err != nil {
		return err
	}
	if note < 0 || note > 127 || velocity < 0 || velocity > 127 {
		return fmt.Errorf("opl2: note/velocity=%d/%d 超出 MIDI 範圍", note, velocity)
	}
	v := &c.voices[channel]
	if velocity == 0 {
		v.active = false
		v.release = 1
		return nil
	}
	v.active = true
	v.note = uint8(note)
	v.velocity = float64(velocity) / 127
	v.release = 0
	v.modPhase, v.carPhase = 0, 0
	v.modEnv, v.carEnv = 0, 0
	setEnvelope(v)
	return nil
}

// NoteOff 釋放一個旋律聲道。
func (c *Chip) NoteOff(channel int) error {
	if err := c.validChannel(channel); err != nil {
		return err
	}
	c.voices[channel].active = false
	c.voices[channel].release = 1
	return nil
}

// SetVolume 套用 MUS An 的 remake 線性映射；原版對數映射仍未證實。
func (c *Chip) SetVolume(channel, value int) error {
	if err := c.validChannel(channel); err != nil {
		return err
	}
	if value < 0 || value > 127 {
		return fmt.Errorf("opl2: volume %d 超出 0..127", value)
	}
	c.voices[channel].volume = float64(value) / 127
	return nil
}

// SetPitchBend 套用 pitch bend（value 是 14-bit、中心 8192）。
func (c *Chip) SetPitchBend(channel, value, semitones int) error {
	if err := c.validChannel(channel); err != nil {
		return err
	}
	if value < 0 || value > 16383 || semitones < 0 || semitones > 24 {
		return fmt.Errorf("opl2: pitch bend=%d/%d 超出範圍", value, semitones)
	}
	c.voices[channel].bend = (float64(value) - 8192) / 8192 * float64(semitones)
	return nil
}

// DrumOn 啟動 OPL2 rhythm mode 的一個打擊聲道（MUS channel 6..10）。
// 這是直接播放路徑的明確 API；其聲音是穩定的純 Go 合成，不是 SDFA 暫存器
// 模擬。velocity 0 依 MUS 慣例等同 DrumOff。
func (c *Chip) DrumOn(channel, note, velocity int) error {
	if c == nil {
		return errors.New("opl2: nil chip")
	}
	index := percussionIndex(channel)
	if index < 0 {
		return fmt.Errorf("opl2: 鼓組聲道 %d 不在 %d..%d", channel,
			PercussionFirstChannel, PercussionFirstChannel+PercussionChannels-1)
	}
	if note < 0 || note > 127 || velocity < 0 || velocity > 127 {
		return fmt.Errorf("opl2: 鼓組 note/velocity=%d/%d 超出 MIDI 範圍", note, velocity)
	}
	d := &c.drums[index]
	if velocity == 0 {
		return c.DrumOff(channel)
	}
	d.active = true
	d.note = uint8(note)
	d.velocity = float64(velocity) / 127
	d.phase = 0
	d.envelope = 1
	d.noiseSeed = 0x9e3779b9 ^ uint32(index+1)*0x45d9f3b
	// kick／snare／hat／tom／cymbal 的尾音長度不同；數值只控制
	// 聽感與 bounded render，不對應任何未解的 SDFA 欄位。
	d.decay = []float64{0.00065, 0.0018, 0.0045, 0.0012, 0.0028}[index]
	return nil
}

// DrumOff 釋放一個 OPL2 rhythm 聲道，保留短尾音避免 click。
func (c *Chip) DrumOff(channel int) error {
	if c == nil {
		return errors.New("opl2: nil chip")
	}
	index := percussionIndex(channel)
	if index < 0 {
		return fmt.Errorf("opl2: 鼓組聲道 %d 不在 %d..%d", channel,
			PercussionFirstChannel, PercussionFirstChannel+PercussionChannels-1)
	}
	c.drums[index].active = false
	return nil
}

// SetDrumVolume 套用單一鼓組聲道的線性音量（0..127）。
func (c *Chip) SetDrumVolume(channel, value int) error {
	if c == nil {
		return errors.New("opl2: nil chip")
	}
	index := percussionIndex(channel)
	if index < 0 {
		return fmt.Errorf("opl2: 鼓組聲道 %d 不在 %d..%d", channel,
			PercussionFirstChannel, PercussionFirstChannel+PercussionChannels-1)
	}
	if value < 0 || value > 127 {
		return fmt.Errorf("opl2: 鼓組音量 %d 超出 0..127", value)
	}
	c.drums[index].volume = float64(value) / 127
	return nil
}

// Render 產生 frames 個雙聲道 signed 16-bit little-endian samples。
func (c *Chip) Render(frames int) []byte {
	if c == nil || frames <= 0 {
		return nil
	}
	out := make([]byte, frames*4)
	for frame := 0; frame < frames; frame++ {
		var mixed float64
		for i := range c.voices {
			mixed += c.renderVoice(&c.voices[i])
		}
		for i := range c.drums {
			mixed += c.renderDrum(i, &c.drums[i])
		}
		if mixed > 1 {
			mixed = 1
		} else if mixed < -1 {
			mixed = -1
		}
		value := int16(math.Round(mixed * 30000))
		binary.LittleEndian.PutUint16(out[frame*4:], uint16(value))
		binary.LittleEndian.PutUint16(out[frame*4+2:], uint16(value))
	}
	return out
}

func (c *Chip) validChannel(channel int) error {
	if c == nil {
		return errors.New("opl2: nil chip")
	}
	if channel < 0 || channel >= MaxChannels {
		return fmt.Errorf("opl2: 聲道 %d 超出 0..%d", channel, MaxChannels-1)
	}
	return nil
}

func (c *Chip) renderVoice(v *voice) float64 {
	if !v.active && v.release <= 0 {
		return 0
	}
	modFreq, carFreq := noteFrequency(v.note, v.bend, v.instrument.Modulator.Multiple),
		noteFrequency(v.note, v.bend, v.instrument.Carrier.Multiple)
	modStep := 2 * math.Pi * modFreq / float64(c.sampleRate)
	carStep := 2 * math.Pi * carFreq / float64(c.sampleRate)
	v.modPhase = math.Mod(v.modPhase+modStep, 2*math.Pi)
	v.carPhase = math.Mod(v.carPhase+carStep, 2*math.Pi)
	if v.active {
		v.modEnv += v.attack
		v.carEnv += v.attack
		if v.modEnv > 1 {
			v.modEnv = 1
		}
		if v.carEnv > 1 {
			v.carEnv = 1
		}
	} else {
		v.release -= v.releaseRate
		v.modEnv -= v.releaseRate
		v.carEnv -= v.releaseRate
		if v.release <= 0 || v.carEnv <= 0 {
			v.release, v.modEnv, v.carEnv = 0, 0, 0
			return 0
		}
	}
	mod := math.Sin(v.modPhase) * v.modLevel * v.modEnv
	carrier := math.Sin(v.carPhase+mod) * v.carrierLevel * v.carEnv
	if v.instrument.Modulator.Connection != 0 {
		carrier += math.Sin(v.modPhase) * v.modLevel * v.modEnv * 0.5
	}
	return carrier * v.velocity * v.volume
}

func (c *Chip) renderDrum(index int, d *drumVoice) float64 {
	if d == nil || d.envelope <= 0 || index < 0 || index >= PercussionChannels {
		return 0
	}
	freq := 80.0 + float64(d.note)*2
	if index == 0 {
		freq = 55 + float64(d.note%12)*2
	} else if index == 2 {
		freq = 3200 + float64(d.note%8)*80
	} else if index == 4 {
		freq = 1800 + float64(d.note%16)*35
	}
	d.phase = math.Mod(d.phase+2*math.Pi*freq/float64(c.sampleRate), 2*math.Pi)
	d.noiseSeed = d.noiseSeed*1664525 + 1013904223
	noise := (float64((d.noiseSeed>>9)&0x7fffff)/4194303.5 - 1)
	var sample float64
	switch index {
	case 0: // kick：低頻率正弦加一點 transient
		sample = math.Sin(d.phase) * 0.9 * (0.7 + 0.3*d.envelope)
		sample += noise * 0.08
	case 1: // snare：noise 為主，帶短促的 tone
		sample = noise*0.78 + math.Sin(d.phase)*0.22
	case 2: // hi-hat：高頻 noise
		sample = noise * 0.9
	case 3: // tom：中低頻 tone
		sample = math.Sin(d.phase)*0.82 + noise*0.12
	case 4: // cymbal：長一點的 noise
		sample = noise*0.82 + math.Sin(d.phase)*0.12
	}
	d.envelope -= d.decay
	if d.envelope < 0 {
		d.envelope = 0
		d.active = false
	}
	return sample * d.velocity * d.volume * d.envelope * 0.32
}

func setEnvelope(v *voice) {
	// 這些係數是穩定的近似，不是 SDFA 的 envelope 參數重建。
	v.attack = 1 / float64(4800+int(v.instrument.Modulator.Attack)*600)
	v.releaseRate = 1 / float64(7200+int(v.instrument.Carrier.Release)*900)
	v.sustain = 0.5 + float64(v.instrument.Carrier.Sustain&15)/30
	v.modLevel = 0.25 + (63-float64(v.instrument.Modulator.Level&63))/84
	v.carrierLevel = 0.2 + (63-float64(v.instrument.Carrier.Level&63))/70
}

func noteFrequency(note uint8, bend float64, multiple uint16) float64 {
	base := 440 * math.Pow(2, (float64(note)-69+bend)/12)
	m := float64(multiple & 15)
	if m == 0 {
		m = 0.5
	}
	return base * m
}

// NormalizeEvents 只把 confirmed MUS channel events 轉成 typed voice events。
// 未知 SysEx 與 F8 filler 會被忽略，避免猜測驅動程式語意。
func NormalizeEvents(song mus.Song) ([]VoiceEvent, error) {
	if song.Header.TickRate() <= 0 {
		return nil, errors.New("opl2: MUS tick rate 必須為正數")
	}
	out := make([]VoiceEvent, 0, len(song.Events))
	for _, event := range song.Events {
		status := event.Status
		switch {
		case status == 0xfc || status == 0xf8:
			continue
		case status == 0xf0:
			mul, ok := parseTempo(event.Data)
			if ok {
				out = append(out, VoiceEvent{Tick: event.Tick, Kind: EventTempo, TempoMul: mul})
			}
			continue
		case status >= 0x80 && status <= 0xef:
			channel := status & 0x0f
			kind := status >> 4
			if len(event.Data) == 0 {
				return nil, fmt.Errorf("opl2: tick %d status 0x%02x 缺少資料", event.Tick, status)
			}
			ve := VoiceEvent{Tick: event.Tick, Channel: channel, Value1: event.Data[0]}
			switch kind {
			case 0x8:
				ve.Kind = EventNoteOff
			case 0x9:
				ve.Kind = EventNoteOn
				if len(event.Data) < 2 {
					return nil, fmt.Errorf("opl2: tick %d note-on 截斷", event.Tick)
				}
				ve.Value2 = event.Data[1]
				if ve.Value2 == 0 {
					ve.Kind = EventNoteOff
				}
			case 0xa:
				ve.Kind = EventVolume
			case 0xc:
				ve.Kind = EventProgram
				ve.Program = int(ve.Value1)
			case 0xe:
				if len(event.Data) < 2 {
					return nil, fmt.Errorf("opl2: tick %d pitch bend 截斷", event.Tick)
				}
				ve.Kind = EventPitchBend
				ve.Value2 = event.Data[1]
			default:
				continue
			}
			out = append(out, ve)
		}
	}
	return out, nil
}

// RenderSong 將一首 MUS/TIM 渲染成 bounded、deterministic 的 PCM。
func RenderSong(song mus.Song, bank mus.TimbreBank, options RenderOptions) ([]byte, error) {
	sampleRate := options.SampleRate
	if sampleRate == 0 {
		sampleRate = DefaultSampleRate
	}
	if sampleRate <= 0 || sampleRate > 192000 {
		return nil, fmt.Errorf("opl2: sample rate %d 不在 1..192000", sampleRate)
	}
	events, err := NormalizeEvents(song)
	if err != nil {
		return nil, err
	}
	chip, err := NewChip(sampleRate, bank)
	if err != nil {
		return nil, err
	}
	for ch := 0; ch < MaxChannels; ch++ {
		if err := chip.SetProgram(ch, 0); err != nil {
			return nil, err
		}
		_ = chip.SetVolume(ch, 100)
	}
	if options.MaxFrames < 0 {
		return nil, errors.New("opl2: MaxFrames 不可為負數")
	}
	maxFrames := options.MaxFrames
	if maxFrames == 0 {
		seconds := float64(song.ActualTick) / song.Header.TickRate()
		maxFrames = int(math.Ceil(seconds*float64(sampleRate))) + sampleRate/20
	}
	if maxFrames <= 0 {
		return nil, errors.New("opl2: 計算出的 frame 數不是正數")
	}
	if maxFrames > 192000*60*10 {
		return nil, errors.New("opl2: 渲染上限超過 10 分鐘")
	}

	out := make([]byte, 0, maxFrames*4)
	currentFrame := 0
	lastTick := uint32(0)
	tickRate := song.Header.TickRate()
	for _, event := range events {
		if event.Tick < lastTick {
			return nil, errors.New("opl2: MUS event tick 非單調")
		}
		frames := int(math.Round(float64(event.Tick-lastTick) / tickRate * float64(sampleRate)))
		if frames > 0 {
			if currentFrame+frames > maxFrames {
				frames = maxFrames - currentFrame
			}
			out = append(out, chip.Render(frames)...)
			currentFrame += frames
			if currentFrame >= maxFrames {
				return out, nil
			}
		}
		if err := chip.applyEvent(event); err != nil {
			return nil, err
		}
		if event.Kind == EventTempo && event.TempoMul > 0 {
			tickRate = song.Header.TickRate() * event.TempoMul
		}
		lastTick = event.Tick
	}
	if currentFrame < maxFrames {
		out = append(out, chip.Render(maxFrames-currentFrame)...)
	}
	return out, nil
}

// RenderAdLib 直接把 MUS/TIM 渲染成可交給 Ebiten audio.Player 的 PCM。
//
// AdLib 與 OPL2 在本專案的播放邊界是同一條純 Go software path：解析 MUS/TIM、
// 套用音色、渲染九個旋律聲道與五個 rhythm 聲道。它不載入、不反組 SDFA.EXE，
// 也不承諾原版寄存器時序等價；交付目標只有「音樂與音效可播放」。
func RenderAdLib(musData, timData []byte, options RenderOptions) ([]byte, error) {
	song, err := mus.ParseSong(musData)
	if err != nil {
		return nil, fmt.Errorf("opl2: 解析 MUS：%w", err)
	}
	bank, err := mus.ParseTimbreBank(timData)
	if err != nil {
		return nil, fmt.Errorf("opl2: 解析 TIM：%w", err)
	}
	return RenderSong(song, bank, options)
}

func (c *Chip) applyEvent(event VoiceEvent) error {
	channel := int(event.Channel)
	if index := percussionIndex(channel); index >= 0 {
		switch event.Kind {
		case EventProgram, EventPitchBend, EventTempo:
			return nil
		case EventNoteOn:
			return c.DrumOn(channel, int(event.Value1), int(event.Value2))
		case EventNoteOff:
			return c.DrumOff(channel)
		case EventVolume:
			return c.SetDrumVolume(channel, int(event.Value1))
		default:
			return nil
		}
	}
	if channel >= MaxChannels {
		return nil
	}
	switch event.Kind {
	case EventProgram:
		return c.SetProgram(channel, event.Program)
	case EventNoteOn:
		return c.NoteOn(channel, int(event.Value1), int(event.Value2))
	case EventNoteOff:
		return c.NoteOff(channel)
	case EventVolume:
		return c.SetVolume(channel, int(event.Value1))
	case EventPitchBend:
		return c.SetPitchBend(channel, int(event.Value1)|int(event.Value2)<<7, 1)
	case EventTempo:
		return nil
	default:
		return nil
	}
}

func percussionIndex(channel int) int {
	if channel < PercussionFirstChannel || channel >= PercussionFirstChannel+PercussionChannels {
		return -1
	}
	return channel - PercussionFirstChannel
}

func parseTempo(data []byte) (float64, bool) {
	if len(data) != 4 || data[0] != 0x7f || data[1] != 0x00 {
		return 0, false
	}
	mul := float64(data[2]) + float64(data[3])/128
	return mul, mul > 0
}
