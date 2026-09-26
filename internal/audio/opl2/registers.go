package opl2

import (
	"fmt"

	"github.com/wicanr2/great-era-remake/internal/audio/mus"
)

// RegisterWrite 是一筆可送到 YM3812/OPL2 的 index/data 寫入。
//
// 這個序列只描述 TIM 音色的靜態 operator 設定；SDFA 的實際寫入順序、
// 延遲、鍵位／頻率寄存器與 INT 66h 時序仍是未知，不能把它當作原版 trace。
type RegisterWrite struct {
	Address uint8
	Value   uint8
}

// opl2OperatorOffsets 是 OPL2 九個聲道的調變器 operator offset；carrier
// 在同一聲道固定位於 +3。這是 YM3812 公開 register layout 的強推論。
var opl2OperatorOffsets = [...]uint8{0, 1, 2, 8, 9, 10, 16, 17, 18}

// ProgramRegisterWrites 將一筆 TIM 音色的 28 個 word 打包成 OPL2 靜態
// 寄存器寫入。它保留合法 bitfield 的低位，忽略 TIM 中已知未初始化的
// carrier Feedback／Connection 殘值；呼叫端仍需自行寫入 F-number、Block、
// Key-On（0xA0/0xB0）與裝置延遲。
func ProgramRegisterWrites(channel int, instrument mus.Instrument) ([]RegisterWrite, error) {
	if channel < 0 || channel >= len(opl2OperatorOffsets) {
		return nil, fmt.Errorf("opl2: 聲道 %d 超出 0..%d", channel, len(opl2OperatorOffsets)-1)
	}
	modOffset := opl2OperatorOffsets[channel]
	carOffset := modOffset + 3
	out := make([]RegisterWrite, 0, 11)
	out = append(out, operatorRegisterWrites(modOffset, instrument.Modulator, instrument.ModWave)...)
	out = append(out, operatorRegisterWrites(carOffset, instrument.Carrier, instrument.CarWave)...)
	// Feedback／Connection 是聲道級欄位，資料模型以 modulator 半邊為準。
	out = append(out, RegisterWrite{
		Address: 0xC0 + uint8(channel),
		Value:   uint8((instrument.Modulator.Feedback&7)<<1 | (instrument.Modulator.Connection & 1)),
	})
	return out, nil
}

func operatorRegisterWrites(offset uint8, op mus.Operator, wave uint16) []RegisterWrite {
	return []RegisterWrite{
		{Address: 0x20 + offset, Value: uint8((op.AM&1)<<7 | (op.Vibrato&1)<<6 | (op.EG&1)<<5 | (op.KSR&1)<<4 | (op.Multiple & 0x0F))},
		{Address: 0x40 + offset, Value: uint8((op.KSL&3)<<6 | (op.Level & 0x3F))},
		{Address: 0x60 + offset, Value: uint8((op.Attack&0x0F)<<4 | (op.Decay & 0x0F))},
		{Address: 0x80 + offset, Value: uint8((op.Sustain&0x0F)<<4 | (op.Release & 0x0F))},
		{Address: 0xE0 + offset, Value: uint8(wave & 0x03)},
	}
}
