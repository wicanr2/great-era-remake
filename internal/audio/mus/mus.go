// Package mus 解碼《大時代的故事》使用的 AdLib MUS／TIM 資料。
//
// 這一層只處理 bytes → typed records；不依賴 Ebiten、音效裝置或原版檔案路徑。
// OPL2 暫存器排程與合成器是後續切片，不能在 parser 中偷塞一套未驗證的語意。
package mus

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

const (
	// SongHeaderSize 是 MUS 固定檔頭大小。
	SongHeaderSize = 0x46
	// TimbreNameSize 是 TIM 每個音色名稱槽位大小。
	TimbreNameSize = 9
	// TimbreWords 是每筆 TIM 音色的 little-endian u16 數量。
	TimbreWords = 28
	// TimbreRecordSize 是每筆 TIM 音色的 bytes 大小。
	TimbreRecordSize = TimbreWords * 2
)

// Header 是 MUS 檔頭。保留檔案的有號欄位型別，並在 parser 入口拒絕負值。
type Header struct {
	Major          uint8
	Minor          uint8
	TuneID         int32
	TuneName       string
	TickBeat       uint8
	BeatMeasure    uint8
	TotalTick      int32
	DataSize       int32
	CommandCount   int32
	SoundMode      uint8
	PitchBendRange uint8
	BasicTempo     uint16
}

// TickRate 回傳每秒 MUS tick 數；tempo 或 tickBeat 為零時回傳零。
func (h Header) TickRate() float64 {
	if h.TickBeat == 0 || h.BasicTempo == 0 {
		return 0
	}
	return float64(h.BasicTempo) * float64(h.TickBeat) / 60
}

// Event 是一筆已轉成絕對 tick 的 MUS 事件。Data 是獨立副本。
type Event struct {
	Tick   uint32
	Status byte
	Data   []byte
}

// Song 是完整 MUS 解碼結果。
type Song struct {
	Header            Header
	Events            []Event
	EndOffset         int
	ActualTick        uint32
	HeaderTickMatches bool
}

// Operator 是 TIM 的 13-word operator 區段。
// Feedback／Connection 對 carrier 在 OPL2 上不使用，但仍保留原始值。
type Operator struct {
	KSL        uint16
	Multiple   uint16
	Feedback   uint16
	Attack     uint16
	Sustain    uint16
	EG         uint16
	Decay      uint16
	Release    uint16
	Level      uint16
	AM         uint16
	Vibrato    uint16
	KSR        uint16
	Connection uint16
}

// Instrument 是 TIM 的一筆 28-word 音色。
type Instrument struct {
	Modulator Operator
	Carrier   Operator
	ModWave   uint16
	CarWave   uint16
}

// TimbreBank 是完整 TIM 音色庫。
type TimbreBank struct {
	Major           uint8
	Minor           uint8
	Count           uint16
	ParameterOffset uint16
	Names           []string
	Instruments     []Instrument
}

// ParseSong 解析一份 MUS。結構長度與事件數不一致會 fail-closed；header total tick
// 的原版 metadata 差異則保留並透過 ActualTick／HeaderTickMatches 暴露。
func ParseSong(data []byte) (Song, error) {
	var out Song
	if len(data) < SongHeaderSize {
		return out, fmt.Errorf("mus: 檔案只有 %d bytes，少於固定檔頭 %d", len(data), SongHeaderSize)
	}
	h := Header{
		Major:          data[0],
		Minor:          data[1],
		TuneID:         int32(binary.LittleEndian.Uint32(data[2:6])),
		TuneName:       cString(data[6:0x24]),
		TickBeat:       data[0x24],
		BeatMeasure:    data[0x25],
		TotalTick:      int32(binary.LittleEndian.Uint32(data[0x26:0x2a])),
		DataSize:       int32(binary.LittleEndian.Uint32(data[0x2a:0x2e])),
		CommandCount:   int32(binary.LittleEndian.Uint32(data[0x2e:0x32])),
		SoundMode:      data[0x3a],
		PitchBendRange: data[0x3b],
		BasicTempo:     binary.LittleEndian.Uint16(data[0x3c:0x3e]),
	}
	if h.DataSize < 0 || int64(h.DataSize) != int64(len(data)-SongHeaderSize) {
		return out, fmt.Errorf("mus: dataSize=%d，實際事件區=%d", h.DataSize, len(data)-SongHeaderSize)
	}
	if h.TotalTick < 0 || h.CommandCount < 0 {
		return out, fmt.Errorf("mus: header 含負值 totalTick=%d commandCount=%d", h.TotalTick, h.CommandCount)
	}

	events := make([]Event, 0, h.CommandCount)
	pos := SongHeaderSize
	var tick uint64
	var running byte
	for pos < len(data) {
		// F8 是 real-time filler：不吃 delta、不推進 tick、不清 running status。
		if data[pos] == 0xf8 {
			events = append(events, Event{Tick: uint32(tick), Status: 0xf8})
			pos++
			continue
		}
		delta := int(data[pos])
		pos++
		if pos >= len(data) {
			return out, fmt.Errorf("mus: offset 0x%x 的 delta 沒有 status", pos-1)
		}
		tick += uint64(delta)
		if tick > uint64(^uint32(0)) {
			return out, fmt.Errorf("mus: tick 超過 u32：%d", tick)
		}

		status := data[pos]
		if status&0x80 != 0 {
			pos++
		} else {
			if running == 0 {
				return out, fmt.Errorf("mus: offset 0x%x 的資料 byte 沒有 running status", pos)
			}
			status = running
		}

		switch {
		case status == 0xf0:
			end := bytes.IndexByte(data[pos:], 0xf7)
			if end < 0 {
				return out, fmt.Errorf("mus: offset 0x%x 的 SysEx 沒有 F7", pos-1)
			}
			payload := cloneBytes(data[pos : pos+end])
			events = append(events, Event{Tick: uint32(tick), Status: status, Data: payload})
			pos += end + 1
			running = 0
		case status >= 0xf1:
			events = append(events, Event{Tick: uint32(tick), Status: status})
			running = 0
			if status == 0xfc {
				if pos != len(data) {
					return out, fmt.Errorf("mus: FC 後仍有 %d bytes", len(data)-pos)
				}
				return finishSong(h, events, pos, tick)
			}
		default:
			n, ok := channelDataSize(status)
			if !ok {
				return out, fmt.Errorf("mus: offset 0x%x 的未知 status 0x%02x", pos-1, status)
			}
			if len(data)-pos < n {
				return out, fmt.Errorf("mus: status 0x%02x 在 offset 0x%x 截斷（需要 %d bytes）", status, pos-1, n)
			}
			events = append(events, Event{
				Tick:   uint32(tick),
				Status: status,
				Data:   cloneBytes(data[pos : pos+n]),
			})
			pos += n
			running = status
		}
	}
	return out, fmt.Errorf("mus: 事件區到 EOF 前沒有 FC 結束事件")
}

func finishSong(h Header, events []Event, endOffset int, tick uint64) (Song, error) {
	if int64(len(events)) != int64(h.CommandCount) {
		return Song{}, fmt.Errorf("mus: commandCount=%d，實際事件=%d", h.CommandCount, len(events))
	}
	// 原版八首中 MAINTHEM／STRATEGY 的 header totalTick 與事件累計值不一致；
	// 這是輸入檔的可觀察 metadata anomaly，不應讓解碼器偷偷修正或拒絕整首曲子。
	// 呼叫端可用 ActualTick／HeaderTickMatches 明確看見差異，後續播放時序另行裁決。
	actual := uint32(tick)
	return Song{
		Header: h, Events: events, EndOffset: endOffset,
		ActualTick: actual, HeaderTickMatches: actual == uint32(h.TotalTick),
	}, nil
}

func channelDataSize(status byte) (int, bool) {
	switch status >> 4 {
	case 0x8, 0x9, 0xe:
		return 2, true
	case 0xa, 0xb, 0xc, 0xd:
		return 1, true
	default:
		return 0, false
	}
}

// ParseTimbreBank 解析一份 TIM，保留每一個 28-word 音色欄位的原始值。
func ParseTimbreBank(data []byte) (TimbreBank, error) {
	var out TimbreBank
	if len(data) < 6 {
		return out, fmt.Errorf("tim: 檔案只有 %d bytes，少於固定檔頭 6", len(data))
	}
	major, minor := data[0], data[1]
	count := binary.LittleEndian.Uint16(data[2:4])
	paramOffset := binary.LittleEndian.Uint16(data[4:6])
	expectedOffset := 6 + int(count)*TimbreNameSize
	expectedSize := expectedOffset + int(count)*TimbreRecordSize
	if int(paramOffset) != expectedOffset {
		return out, fmt.Errorf("tim: parameterOffset=%d，預期 %d", paramOffset, expectedOffset)
	}
	if len(data) != expectedSize {
		return out, fmt.Errorf("tim: 檔案大小=%d，預期 %d", len(data), expectedSize)
	}

	names := make([]string, count)
	for i := range names {
		start := 6 + i*TimbreNameSize
		names[i] = cString(data[start : start+TimbreNameSize])
	}
	instruments := make([]Instrument, count)
	for i := range instruments {
		start := expectedOffset + i*TimbreRecordSize
		var words [TimbreWords]uint16
		for j := range words {
			words[j] = binary.LittleEndian.Uint16(data[start+j*2 : start+j*2+2])
		}
		instruments[i] = instrumentFromWords(words)
	}
	return TimbreBank{
		Major: major, Minor: minor, Count: count, ParameterOffset: paramOffset,
		Names: names, Instruments: instruments,
	}, nil
}

func instrumentFromWords(w [TimbreWords]uint16) Instrument {
	return Instrument{
		Modulator: operatorFromWords(w[0:13]),
		Carrier:   operatorFromWords(w[13:26]),
		ModWave:   w[26], CarWave: w[27],
	}
}

func operatorFromWords(w []uint16) Operator {
	return Operator{
		KSL: w[0], Multiple: w[1], Feedback: w[2], Attack: w[3],
		Sustain: w[4], EG: w[5], Decay: w[6], Release: w[7],
		Level: w[8], AM: w[9], Vibrato: w[10], KSR: w[11], Connection: w[12],
	}
}

func cString(data []byte) string {
	if i := bytes.IndexByte(data, 0); i >= 0 {
		data = data[:i]
	}
	return string(data)
}

func cloneBytes(data []byte) []byte {
	if len(data) == 0 {
		return nil
	}
	return append([]byte(nil), data...)
}
