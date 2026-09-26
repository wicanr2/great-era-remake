package mus

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
)

const gameDir = "../../../workplace/orig/game"

var songGolden = map[string]struct {
	totalTick  int32
	actualTick uint32
	dataSize   int32
	commands   int32
	mode       uint8
	tempo      uint16
	digest     string
}{
	"MAINTHEM": {62880, 60800, 31157, 7967, 1, 115, "f60284ebe40e81a713added16093720b654991b13d086770176eb7f315af6d6d"},
	"BATTLE1":  {31680, 31680, 13273, 3413, 1, 110, "419aafb4151b7a6875780c4e13cdd919350974d3705adbb8e60443d9d70453b5"},
	"BATTLE2":  {38400, 38400, 11350, 2941, 1, 85, "327819e61cc192e73759162a9e8d719db6964fb0824a0fd9ad876b661253b01c"},
	"SCENE":    {26880, 26880, 12920, 3262, 0, 94, "d8a57b9b157a4833e3071c7caa6fbb2763a5f33cf04d56f3d889d02c13026dd7"},
	"FINAL":    {17760, 17760, 11455, 2918, 1, 85, "9e4326e08c12796c264903ebe7037cc742108666380a3b28c00f952edee2d66d"},
	"BT02":     {19200, 19200, 7062, 1799, 1, 120, "b0c33838aaa6749b908321682da89c606f4b978ca898b32b64a80b7d4c823cdd"},
	"WALL":     {19200, 19200, 5627, 1456, 0, 76, "cb865b75f909c25ac71629e4f8f63d14f6a8efc066926bb32b3e94fa79c80ea0"},
	"STRATEGY": {15540, 15600, 4497, 1213, 1, 40, "2a95ca2a5ba6a9d03fc557530386e88ff8ecdfc3fadf5b9969d58fad9e319721"},
}

var timGolden = map[string]struct {
	count  uint16
	digest string
}{
	"MAINTHEM": {18, "963f8356cd04f721a272123007dee1e837f7c8fbc4f2913a29840a810e9837cd"},
	"BATTLE1":  {18, "f1859c7a9aca03c7c44b9d64fccd697bf9490d491faa70db2e174ebcbc7ed612"},
	"BATTLE2":  {22, "c4160e15ba9d18a1f61cc130250d9c1b4efcdbae6364e8fa20091b8509250a39"},
	"SCENE":    {22, "c4160e15ba9d18a1f61cc130250d9c1b4efcdbae6364e8fa20091b8509250a39"},
	"FINAL":    {14, "65f8dc13496e82ddc3bb2d02123ba336051d2a41bcdd2cf92ca40dc960bb83d9"},
	"BT02":     {9, "211b14ac8a0f1db3738c680166d969e64bf900c3730a5741d6bfaa8158ae3c82"},
	"WALL":     {9, "33cd2bf0a145d2970ae4693ef89afa69ba894df998206363d8c0040ea4524321"},
	"STRATEGY": {11, "eaa71359a691ae36fd29bba1cd49f2830003e804825befccdf1ad5ac0ce48855"},
}

func gameFile(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(gameDir, name))
	if err != nil {
		t.Skipf("沒有原版素材 %s，跳過 real-file golden：%v", name, err)
	}
	return b
}

func TestParseRealSongsMatchesPythonGolden(t *testing.T) {
	for name, want := range songGolden {
		t.Run(name, func(t *testing.T) {
			data := gameFile(t, name+".MUS")
			got, err := ParseSong(data)
			if err != nil {
				t.Fatal(err)
			}
			if got.Header.TotalTick != want.totalTick || got.ActualTick != want.actualTick ||
				got.Header.DataSize != want.dataSize ||
				got.Header.CommandCount != want.commands || got.Header.SoundMode != want.mode ||
				got.Header.BasicTempo != want.tempo {
				t.Fatalf("header／actual tick = %+v／%d，預期 total=%d actual=%d data=%d commands=%d mode=%d tempo=%d",
					got.Header, got.ActualTick, want.totalTick, want.actualTick, want.dataSize, want.commands, want.mode, want.tempo)
			}
			if got.HeaderTickMatches != (want.actualTick == uint32(want.totalTick)) {
				t.Fatalf("HeaderTickMatches=%v，預期 %v", got.HeaderTickMatches, want.actualTick == uint32(want.totalTick))
			}
			if len(got.Events) != int(want.commands) || got.EndOffset != len(data) {
				t.Fatalf("事件數／終點 = %d／0x%x，預期 %d／0x%x", len(got.Events), got.EndOffset, want.commands, len(data))
			}
			if got.Events[len(got.Events)-1].Status != 0xfc {
				t.Fatalf("最後事件 status = %#x，預期 FC", got.Events[len(got.Events)-1].Status)
			}
			if digestEvents(got.Events) != want.digest {
				t.Fatalf("逐事件 golden digest = %s，預期 %s", digestEvents(got.Events), want.digest)
			}
		})
	}
}

func TestParseRealTimbresMatchesPythonGolden(t *testing.T) {
	for name, want := range timGolden {
		t.Run(name, func(t *testing.T) {
			data := gameFile(t, name+".TIM")
			got, err := ParseTimbreBank(data)
			if err != nil {
				t.Fatal(err)
			}
			if got.Major != 1 || got.Minor != 0 || got.Count != want.count ||
				got.ParameterOffset != uint16(6+int(want.count)*TimbreNameSize) {
				t.Fatalf("TIM header = %+v，預期 count=%d", got, want.count)
			}
			if len(got.Names) != int(want.count) || len(got.Instruments) != int(want.count) {
				t.Fatalf("名稱／音色筆數 = %d／%d，預期 %d", len(got.Names), len(got.Instruments), want.count)
			}
			if digestTimbres(got) != want.digest {
				t.Fatalf("逐音色 golden digest = %s，預期 %s", digestTimbres(got), want.digest)
			}
		})
	}
}

func TestParseSongSupportsRunningStatusFillerAndSysEx(t *testing.T) {
	data := minimalMUS([]byte{
		0, 0xc0, 7, // program change
		0, 7, // running status, another program value
		0xf8, // filler: no delta
		0, 0xf0, 0x7f, 0x00, 0x01, 0x00, 0xf7,
		0, 0xfc,
	}, 5, 0)
	got, err := ParseSong(data)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Events) != 5 || got.Events[1].Status != 0xc0 || got.Events[1].Data[0] != 7 ||
		got.Events[2].Status != 0xf8 || got.Events[3].Status != 0xf0 {
		t.Fatalf("事件 = %+v", got.Events)
	}
	if string(got.Events[3].Data) != string([]byte{0x7f, 0, 1, 0}) {
		t.Fatalf("SysEx payload = %x", got.Events[3].Data)
	}
}

func TestParseSongRejectsMalformedInput(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"short header", []byte{1, 0}},
		{"missing FC", minimalMUS([]byte{0, 0xc0, 1}, 1, 0)},
		{"bad running status", minimalMUS([]byte{0, 1, 0, 0xfc}, 2, 0)},
		{"unknown status", minimalMUS([]byte{0, 0x70, 0, 0xfc}, 2, 0)},
		{"truncated channel data", minimalMUS([]byte{0, 0x90, 60}, 1, 0)},
		{"unterminated sysex", minimalMUS([]byte{0, 0xf0, 1, 2}, 1, 0)},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := ParseSong(tc.data); err == nil {
				t.Fatal("畸形 MUS 應拒絕")
			}
		})
	}
}

func TestParseTimbreBankPreservesCarrierResidue(t *testing.T) {
	data := make([]byte, 6+TimbreNameSize+TimbreRecordSize)
	data[0], data[1] = 1, 0
	binary.LittleEndian.PutUint16(data[2:4], 1)
	binary.LittleEndian.PutUint16(data[4:6], 15)
	copy(data[6:15], []byte("hihat1\x00\x00"))
	for i := 0; i < TimbreWords; i++ {
		binary.LittleEndian.PutUint16(data[15+i*2:17+i*2], uint16(i+1))
	}
	got, err := ParseTimbreBank(data)
	if err != nil {
		t.Fatal(err)
	}
	if got.Names[0] != "hihat1" || got.Instruments[0].Carrier.Feedback != 16 ||
		got.Instruments[0].Carrier.Connection != 26 {
		t.Fatalf("TIM 欄位被正規化：%+v", got.Instruments[0])
	}
}

func TestParseTimbreBankRejectsWrongLayout(t *testing.T) {
	data := make([]byte, 6+TimbreNameSize+TimbreRecordSize)
	binary.LittleEndian.PutUint16(data[2:4], 1)
	binary.LittleEndian.PutUint16(data[4:6], 16)
	if _, err := ParseTimbreBank(data); err == nil {
		t.Fatal("錯誤 parameterOffset 應拒絕")
	}
}

func minimalMUS(events []byte, count int, totalTick int32) []byte {
	data := make([]byte, SongHeaderSize+len(events))
	data[0], data[1] = 1, 0
	data[0x24], data[0x25] = 240, 4
	binary.LittleEndian.PutUint32(data[0x26:0x2a], uint32(totalTick))
	binary.LittleEndian.PutUint32(data[0x2a:0x2e], uint32(len(events)))
	binary.LittleEndian.PutUint32(data[0x2e:0x32], uint32(count))
	data[0x3a], data[0x3b] = 0, 1
	binary.LittleEndian.PutUint16(data[0x3c:0x3e], 100)
	copy(data[SongHeaderSize:], events)
	return data
}

func digestEvents(events []Event) string {
	h := sha256.New()
	var buf [6]byte
	for _, event := range events {
		binary.LittleEndian.PutUint32(buf[:4], event.Tick)
		buf[4], buf[5] = event.Status, byte(len(event.Data))
		h.Write(buf[:])
		h.Write(event.Data)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func digestTimbres(bank TimbreBank) string {
	h := sha256.New()
	var buf [2]byte
	for i, name := range bank.Names {
		h.Write([]byte{byte(len(name))})
		h.Write([]byte(name))
		words := instrumentWords(bank.Instruments[i])
		for _, word := range words {
			binary.LittleEndian.PutUint16(buf[:], word)
			h.Write(buf[:])
		}
	}
	return hex.EncodeToString(h.Sum(nil))
}

func instrumentWords(in Instrument) [TimbreWords]uint16 {
	return [TimbreWords]uint16{
		in.Modulator.KSL, in.Modulator.Multiple, in.Modulator.Feedback, in.Modulator.Attack,
		in.Modulator.Sustain, in.Modulator.EG, in.Modulator.Decay, in.Modulator.Release,
		in.Modulator.Level, in.Modulator.AM, in.Modulator.Vibrato, in.Modulator.KSR,
		in.Modulator.Connection,
		in.Carrier.KSL, in.Carrier.Multiple, in.Carrier.Feedback, in.Carrier.Attack,
		in.Carrier.Sustain, in.Carrier.EG, in.Carrier.Decay, in.Carrier.Release,
		in.Carrier.Level, in.Carrier.AM, in.Carrier.Vibrato, in.Carrier.KSR,
		in.Carrier.Connection, in.ModWave, in.CarWave,
	}
}
