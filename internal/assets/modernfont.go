package assets

import (
	"bytes"
	"embed"
	"encoding/binary"
	"fmt"
	"sync"
)

//go:embed data/modern-font*.bin
var embeddedModernFontData embed.FS

const (
	modernFontMagic       = "GEMF"
	modernFontVersion1    = 1
	modernFontVersion2    = 2
	modernFontCellW1      = 16
	modernFontCellH1      = 16
	modernFontRowBytes1   = (modernFontCellW1 + 7) / 8
	modernFontCellW2      = 32
	modernFontCellH2      = 32
	modernFontRowBytes2   = modernFontCellW2
	modernFontHeaderBytes = 20
	modernFontMetaBytes   = 8
)

// ModernFontFamily 是 Modern 版面可選的字型家族。標題預設使用 Serif，
// 按鈕／內文由呼叫端選 Sans；兩者都是同一個 GEMF 資料契約。
type ModernFontFamily string

const (
	ModernFontSans  ModernFontFamily = "sans"
	ModernFontSerif ModernFontFamily = "serif"
)

// ModernGlyph 可同時表示 GEMF v1 的 1-bit glyph 與 GEMF v2 的 8-bit alpha
// glyph。bitmap 保留 v1 相容資料；v2 使用 alpha。這層不依賴 cgo、FreeType
// 或 Ebiten text renderer。
type ModernGlyph struct {
	advance uint16
	cellW   uint8
	cellH   uint8
	bitmap  []byte
	alpha   []byte
}

// Advance 是 renderer 的 16-unit 正規化比例字距。v2 的 32 px atlas 會除以
// 2，維持既有 Modern scale=1≈16 px、scale=2≈32 px 的排版語意。
func (g ModernGlyph) Advance() int {
	if g.cellW == modernFontCellW2 {
		return maxInt(1, int(g.advance)/2)
	}
	return int(g.advance)
}

// PixelAdvance 回傳 atlas 原生像素字距，供未來 alpha renderer 直接取樣。
func (g ModernGlyph) PixelAdvance() int { return int(g.advance) }

// At 保留既有 1-bit API；v2 alpha 以 128 為可見門檻，讓未更新的 renderer
// 仍能安全使用新 atlas。
func (g ModernGlyph) At(x, y int) bool { return g.AlphaAt(x, y) >= 128 }

// AlphaAt 回傳 glyph 像素 alpha；v1 glyph 只會回傳 0 或 255。
func (g ModernGlyph) AlphaAt(x, y int) uint8 {
	if x < 0 || y < 0 || int(g.cellW) <= x || int(g.cellH) <= y {
		return 0
	}
	if len(g.alpha) != 0 {
		return g.alpha[y*int(g.cellW)+x]
	}
	if len(g.bitmap) == 0 {
		return 0
	}
	if g.bitmap[y*((int(g.cellW)+7)/8)+x/8]&(0x80>>uint(x%8)) != 0 {
		return 255
	}
	return 0
}

// Size 回傳 glyph raster cell 尺寸。
func (g ModernGlyph) Size() (int, int) { return int(g.cellW), int(g.cellH) }

// ModernFont 是一份 keyed Unicode atlas。Version 回傳輸入 GEMF 版本，讓
// 發行診斷能明確知道是否使用 alpha atlas。
type ModernFont struct {
	cellW   int
	cellH   int
	version uint16
	glyphs  map[rune]ModernGlyph
}

func (f *ModernFont) CellSize() (int, int) {
	if f == nil {
		return 0, 0
	}
	return f.cellW, f.cellH
}

func (f *ModernFont) Version() int {
	if f == nil {
		return 0
	}
	return int(f.version)
}

func (f *ModernFont) Glyph(r rune) (ModernGlyph, bool) {
	if f == nil {
		return ModernGlyph{}, false
	}
	g, ok := f.glyphs[r]
	return g, ok
}

func (f *ModernFont) GlyphCount() int {
	if f == nil {
		return 0
	}
	return len(f.glyphs)
}

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// ParseModernFont 解析 GEMF v1／v2；長度、排序與 alpha 邊界皆 fail-closed。
func ParseModernFont(data []byte) (*ModernFont, error) {
	if len(data) < modernFontHeaderBytes {
		return nil, fmt.Errorf("assets: Modern font header 太短：%d", len(data))
	}
	if string(data[:4]) != modernFontMagic {
		return nil, fmt.Errorf("assets: Modern font magic 不正確")
	}
	version := binary.LittleEndian.Uint16(data[4:6])
	cellW := int(binary.LittleEndian.Uint16(data[6:8]))
	cellH := int(binary.LittleEndian.Uint16(data[8:10]))
	rowBytes := int(binary.LittleEndian.Uint16(data[10:12]))
	count := binary.LittleEndian.Uint32(data[16:20])
	var payload int
	switch version {
	case modernFontVersion1:
		if cellW != modernFontCellW1 || cellH != modernFontCellH1 || rowBytes != modernFontRowBytes1 {
			return nil, fmt.Errorf("assets: Modern v1 cell／row 不支援：%dx%d row=%d", cellW, cellH, rowBytes)
		}
		payload = modernFontMetaBytes + modernFontRowBytes1*modernFontCellH1
	case modernFontVersion2:
		if cellW != modernFontCellW2 || cellH != modernFontCellH2 || rowBytes != modernFontRowBytes2 {
			return nil, fmt.Errorf("assets: Modern v2 cell／row 不支援：%dx%d row=%d", cellW, cellH, rowBytes)
		}
		payload = modernFontMetaBytes + modernFontCellW2*modernFontCellH2
	default:
		return nil, fmt.Errorf("assets: Modern font 版本 %d，不支援", version)
	}
	if count == 0 {
		return nil, fmt.Errorf("assets: Modern font 沒有 glyph")
	}
	want := modernFontHeaderBytes + int(count)*payload
	if want != len(data) {
		return nil, fmt.Errorf("assets: Modern font 大小 %d，預期 %d", len(data), want)
	}
	f := &ModernFont{cellW: cellW, cellH: cellH, version: version, glyphs: make(map[rune]ModernGlyph, count)}
	offset := modernFontHeaderBytes
	var previous rune
	for i := uint32(0); i < count; i++ {
		codepoint := rune(binary.LittleEndian.Uint32(data[offset : offset+4]))
		advance := binary.LittleEndian.Uint16(data[offset+4 : offset+6])
		if codepoint == 0 || (i > 0 && codepoint <= previous) {
			return nil, fmt.Errorf("assets: Modern font code point 未嚴格遞增：U+%04X", codepoint)
		}
		if advance == 0 || int(advance) > cellW {
			return nil, fmt.Errorf("assets: Modern font U+%04X 字距無效：%d", codepoint, advance)
		}
		g := ModernGlyph{advance: advance, cellW: uint8(cellW), cellH: uint8(cellH)}
		body := data[offset+modernFontMetaBytes : offset+payload]
		if version == modernFontVersion1 {
			g.bitmap = append([]byte(nil), body...)
		} else {
			g.alpha = append([]byte(nil), body...)
		}
		f.glyphs[codepoint] = g
		previous = codepoint
		offset += payload
	}
	return f, nil
}

var (
	embeddedModernFontOnce sync.Once
	embeddedModernFonts    map[ModernFontFamily]*ModernFont
	embeddedModernFontErr  error
)

// EmbeddedModernFont 保留既有 API，回傳 Sans atlas。
func EmbeddedModernFont() (*ModernFont, error) {
	return EmbeddedModernFontFamily(ModernFontSans)
}

// EmbeddedModernFontFamily 載入隨程式散布的 sans／serif GEMF v2 atlas。
func EmbeddedModernFontFamily(family ModernFontFamily) (*ModernFont, error) {
	embeddedModernFontOnce.Do(func() {
		embeddedModernFonts = make(map[ModernFontFamily]*ModernFont, 2)
		var legacy *ModernFont
		loadLegacy := func() (*ModernFont, error) {
			if legacy != nil {
				return legacy, nil
			}
			data, err := embeddedModernFontData.ReadFile("data/modern-font.bin")
			if err != nil {
				return nil, err
			}
			legacy, err = ParseModernFont(data)
			return legacy, err
		}
		for _, item := range []struct {
			family ModernFontFamily
			name   string
		}{
			{ModernFontSans, "data/modern-font-sans.bin"},
			{ModernFontSerif, "data/modern-font-serif.bin"},
		} {
			data, err := embeddedModernFontData.ReadFile(item.name)
			var font *ModernFont
			if err == nil {
				font, err = ParseModernFont(data)
			}
			if err != nil {
				// v2 缺少或損壞時明確退回舊 16×16、1-bit atlas；這只
				// 維持可讀性，不能作為高解析驗收證據。
				font, err = loadLegacy()
				if err != nil {
					embeddedModernFontErr = fmt.Errorf("%s v2 與 v1 fallback 均失敗: %w", item.family, err)
					return
				}
			}
			embeddedModernFonts[item.family] = font
		}
	})
	if embeddedModernFontErr != nil {
		return nil, embeddedModernFontErr
	}
	font, ok := embeddedModernFonts[family]
	if !ok {
		return nil, fmt.Errorf("assets: 未知 Modern 字型家族 %q", family)
	}
	return font, nil
}

// ModernFontBytesForTest 產生 GEMF v1 最小 atlas，保留舊測試與相容性。
func ModernFontBytesForTest(codepoints ...rune) []byte {
	return modernFontBytesForTest(modernFontVersion1, codepoints...)
}

// ModernFontV2BytesForTest 產生含中間 alpha 的 GEMF v2 fixture。
func ModernFontV2BytesForTest(codepoints ...rune) []byte {
	return modernFontBytesForTest(modernFontVersion2, codepoints...)
}

func modernFontBytesForTest(version uint16, codepoints ...rune) []byte {
	if len(codepoints) == 0 {
		return nil
	}
	cellW, cellH, rowBytes := modernFontCellW1, modernFontCellH1, modernFontRowBytes1
	payloadBytes := modernFontRowBytes1 * modernFontCellH1
	if version == modernFontVersion2 {
		cellW, cellH, rowBytes = modernFontCellW2, modernFontCellH2, modernFontRowBytes2
		payloadBytes = cellW * cellH
	}
	buf := bytes.NewBuffer(nil)
	_ = binary.Write(buf, binary.LittleEndian, [4]byte{'G', 'E', 'M', 'F'})
	_ = binary.Write(buf, binary.LittleEndian, version)
	_ = binary.Write(buf, binary.LittleEndian, uint16(cellW))
	_ = binary.Write(buf, binary.LittleEndian, uint16(cellH))
	_ = binary.Write(buf, binary.LittleEndian, uint16(rowBytes))
	_ = binary.Write(buf, binary.LittleEndian, uint16(0))
	_ = binary.Write(buf, binary.LittleEndian, uint16(0))
	_ = binary.Write(buf, binary.LittleEndian, uint32(len(codepoints)))
	for _, r := range codepoints {
		_ = binary.Write(buf, binary.LittleEndian, uint32(r))
		_ = binary.Write(buf, binary.LittleEndian, uint16(cellW))
		_ = binary.Write(buf, binary.LittleEndian, uint16(0))
		body := make([]byte, payloadBytes)
		if version == modernFontVersion2 {
			body[0], body[1], body[2] = 64, 128, 255
		}
		buf.Write(body)
	}
	return buf.Bytes()
}
