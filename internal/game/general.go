package game

import (
	"encoding/binary"
	"fmt"
)

// 依據：docs/spec/02-generals.md（READY）

// GeneralRecordSize 是一筆將領記錄的大小。
//
// MAN(N).DAT 與存檔 .DT1 的將領區用同一個結構。
// MAN(1).DAT 的 9,042 ÷ 274 人 = 33 整除。
const GeneralRecordSize = 33

// 已定名欄位在記錄內的位移。其餘 30 個 byte 語意未解。
const (
	genOffAbilityA = 0
	genOffAbilityB = 1
	genOffAbilityC = 2
	genOffForce    = 17 // u16 little-endian
)

// General 是一位將領。
//
// 33 bytes 裡目前只解出四個欄位，**其餘 30 個 byte 原樣保留在 Raw**——
// 存檔寫回是「改寫」不是「重建」，未解區域一個 byte 都不能動
// （CLAUDE.md §9）。
type General struct {
	// AbilityA/B/C 是三個 0-100 的能力值。
	//
	// **不知道哪個是統率、武力、政治或忠誠**（docs/spec/02 §1），
	// 所以刻意不做語意化命名。蔣中正三項全 100 可當 sanity check。
	AbilityA, AbilityB, AbilityC uint8

	// Force 是兵力。實測上限 20000（步兵師滿員），
	// 但騎兵師 10000／炮兵旅 2000／裝甲 200 這三個社群數字尚未驗證。
	Force uint16

	// Raw 是完整的 33 bytes，寫回時以它為基底只蓋已解欄位。
	Raw [GeneralRecordSize]byte
}

// ParseGeneral 解一筆 33 bytes 的將領記錄。
func ParseGeneral(rec []byte) (General, error) {
	var g General
	if len(rec) < GeneralRecordSize {
		return g, fmt.Errorf("game: 將領記錄需要 %d bytes，只有 %d",
			GeneralRecordSize, len(rec))
	}
	copy(g.Raw[:], rec)
	g.AbilityA = rec[genOffAbilityA]
	g.AbilityB = rec[genOffAbilityB]
	g.AbilityC = rec[genOffAbilityC]
	g.Force = binary.LittleEndian.Uint16(rec[genOffForce:])
	return g, nil
}

// Bytes 產生寫回用的 33 bytes：以原始 bytes 為基底，只蓋已解欄位。
//
// 這是 CLAUDE.md §9 要求的「改寫而非重建」——未解的 30 個 byte 原樣帶過去，
// 驗收標準是 byte-for-byte round-trip。
func (g *General) Bytes() [GeneralRecordSize]byte {
	out := g.Raw
	out[genOffAbilityA] = g.AbilityA
	out[genOffAbilityB] = g.AbilityB
	out[genOffAbilityC] = g.AbilityC
	binary.LittleEndian.PutUint16(out[genOffForce:], g.Force)
	return out
}

// ParseGenerals 解一整段連續的將領記錄。
func ParseGenerals(data []byte, count int) ([]General, error) {
	if count < 0 {
		return nil, fmt.Errorf("game: 將領數不能為負（%d）", count)
	}
	need := count * GeneralRecordSize
	if len(data) < need {
		return nil, fmt.Errorf("game: %d 位將領需要 %d bytes，只有 %d",
			count, need, len(data))
	}
	out := make([]General, count)
	for i := 0; i < count; i++ {
		g, err := ParseGeneral(data[i*GeneralRecordSize:])
		if err != nil {
			return nil, err
		}
		out[i] = g
	}
	return out, nil
}

// TotalForce 是一組將領的兵力總和。
//
// 政略畫面的「兵力」就是這個值，**不是存檔裡的單一欄位**——
// 這也解釋了它為什麼會超過 u16（湖北 97,500）。實機驗證見 docs/spec/02 §2。
//
// 回傳 uint32，因為總和會超過 u16。
func TotalForce(gs []General) uint32 {
	var sum uint32
	for i := range gs {
		sum += uint32(gs[i].Force)
	}
	return sum
}
