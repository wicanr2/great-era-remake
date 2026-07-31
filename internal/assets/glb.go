package assets

import (
	"encoding/binary"
	"fmt"
)

// .GLB 256 色圖庫 + .GTB 目錄。
//
// GTB entry 固定 13 bytes：
//
//	+0  u8   flag       0 = 調色盤(768 B), 1 = 圖像
//	+1  u16  rawSize    解壓後大小（圖像 = w×h，每像素 1 byte）
//	+3  u16  compSize   壓縮後大小
//	+5  u32  offset     GLB 內位移
//	+9  u16  width
//	+11 u16  height
//
// GLB 的每一筆各自是獨立的 DCL implode 資料（header 00 06）。

// GTBEntrySize 是目錄每筆的固定大小。
const GTBEntrySize = 13

// GTBEntry 是圖庫目錄的一筆。
type GTBEntry struct {
	Index    int
	Flag     byte
	RawSize  int
	CompSize int
	Offset   int
	W, H     int
}

// IsPalette 回報這筆是不是調色盤。調色盤的 W/H 是 0、RawSize 是 768。
func (e GTBEntry) IsPalette() bool { return e.W == 0 || e.H == 0 }

// ParseGTB 解析 .GTB 目錄，並檢查自洽性。
func ParseGTB(data []byte) ([]GTBEntry, error) {
	if len(data)%GTBEntrySize != 0 {
		return nil, fmt.Errorf("assets: .GTB 大小 %d 不是 %d 的倍數", len(data), GTBEntrySize)
	}
	n := len(data) / GTBEntrySize
	out := make([]GTBEntry, 0, n)
	for k := 0; k < n; k++ {
		b := data[k*GTBEntrySize:]
		e := GTBEntry{
			Index:    k,
			Flag:     b[0],
			RawSize:  int(binary.LittleEndian.Uint16(b[1:])),
			CompSize: int(binary.LittleEndian.Uint16(b[3:])),
			Offset:   int(binary.LittleEndian.Uint32(b[5:])),
			W:        int(binary.LittleEndian.Uint16(b[9:])),
			H:        int(binary.LittleEndian.Uint16(b[11:])),
		}
		// 自洽性：圖像的 rawSize 必須等於 w×h，調色盤必須是 768。
		want := e.W * e.H
		if e.IsPalette() {
			want = 768
		}
		if e.RawSize != want {
			return nil, fmt.Errorf("assets: .GTB #%d rawSize=%d，預期 %d (%dx%d)",
				k, e.RawSize, want, e.W, e.H)
		}
		out = append(out, e)
	}
	return out, nil
}

// GLB 是一組目錄加上壓縮資料。
type GLB struct {
	Entries []GTBEntry
	Data    []byte
}

// OpenGLB 綁定 .GTB 與 .GLB，並驗證兩者對得起來。
func OpenGLB(gtb, glb []byte) (*GLB, error) {
	entries, err := ParseGTB(gtb)
	if err != nil {
		return nil, err
	}
	for i, e := range entries {
		if e.Offset+e.CompSize > len(glb) {
			return nil, fmt.Errorf("assets: .GTB #%d 指到 %d..%d，超出 .GLB 的 %d bytes",
				i, e.Offset, e.Offset+e.CompSize, len(glb))
		}
		// 每筆各自是完整的 DCL stream，起點必須是合法 header。
		h := glb[e.Offset : e.Offset+2]
		if h[0] > 1 || h[1] < 4 || h[1] > 6 {
			return nil, fmt.Errorf("assets: .GTB #%d 起點不是 DCL header (%02x %02x)", i, h[0], h[1])
		}
		if i > 0 {
			prev := entries[i-1]
			if prev.Offset+prev.CompSize != e.Offset {
				return nil, fmt.Errorf("assets: .GTB #%d 的位移 %d 接不上前一筆的結尾 %d",
					i, e.Offset, prev.Offset+prev.CompSize)
			}
		}
	}
	return &GLB{Entries: entries, Data: glb}, nil
}

// Raw 解出第 i 筆的原始資料。
func (g *GLB) Raw(i int) ([]byte, error) {
	if i < 0 || i >= len(g.Entries) {
		return nil, fmt.Errorf("assets: .GLB 索引 %d 超出範圍 0..%d", i, len(g.Entries)-1)
	}
	e := g.Entries[i]
	return Explode(g.Data[e.Offset:e.Offset+e.CompSize], e.RawSize)
}

// Palette 解出第 i 筆的調色盤。
func (g *GLB) Palette(i int) (Palette, error) {
	if !g.Entries[i].IsPalette() {
		return nil, fmt.Errorf("assets: .GLB #%d 不是調色盤", i)
	}
	raw, err := g.Raw(i)
	if err != nil {
		return nil, err
	}
	return ParseVGAPalette(raw)
}

// PaletteFor 回傳管轄第 i 筆圖像的調色盤索引。
//
// 調色盤與圖像在 .GLB 內交錯排列，一個調色盤管到下一個調色盤出現為止，
// 所以要往前找最近的一個。找不到回 -1。
func (g *GLB) PaletteFor(i int) int {
	for k := i; k >= 0; k-- {
		if g.Entries[k].IsPalette() {
			return k
		}
	}
	return -1
}

// Sprite 是 .GLB 裡的一張 256 色圖。
//
// 索引 0 是透明色：疊圖時不寫入，露出底圖。書法標題就是這樣疊上去的。
type Sprite struct {
	W, H int
	Pix  []byte // 每像素 1 byte，值 0-255
}

// TransparentIndex 是 .GLB 圖像的透明色索引。
const TransparentIndex = 0

// Sprite 解出第 i 筆圖像。
func (g *GLB) Sprite(i int) (*Sprite, error) {
	e := g.Entries[i]
	if e.IsPalette() {
		return nil, fmt.Errorf("assets: .GLB #%d 是調色盤不是圖像", i)
	}
	raw, err := g.Raw(i)
	if err != nil {
		return nil, err
	}
	return &Sprite{W: e.W, H: e.H, Pix: raw}, nil
}
