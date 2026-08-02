package assets

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/text/encoding/traditionalchinese"
)

// 倚天中文系統 3.53 的 16×15 字庫。公式逐行移植自 tools/eten.py。
const (
	EtenStandardGlyphs = 13094
	EtenSymbolGlyphs   = 408
	EtenASCIIGlyphs    = 256
	EtenCommonGlyphs   = 5401
	EtenASCIIStride    = 15 // ASCFONT.15：8×15，每列一 byte
)

type EtenFontKind uint8

const (
	EtenSymbol EtenFontKind = iota + 1
	EtenStandard
)

func (k EtenFontKind) String() string {
	if k == EtenSymbol {
		return "SPCFONT"
	}
	if k == EtenStandard {
		return "STDFONT"
	}
	return "UNKNOWN"
}

func etenRaw(hi, lo byte) int {
	low := int(lo) - 0x40
	if lo >= 0x7f {
		low = int(lo) - 0x62
	}
	return (int(hi)-0xa1)*157 + low
}

var (
	etenLastSymbol = etenRaw(0xa3, 0xbf)
	etenBaseA440   = etenRaw(0xa4, 0x40)
	etenLastCommon = etenRaw(0xc6, 0x7e)
	etenBaseC940   = etenRaw(0xc9, 0x40)
)

// EtenBig5Index 將合法 Big5 雙位元組轉成字庫種類與字庫內 0-based 索引。
func EtenBig5Index(hi, lo byte) (EtenFontKind, int, bool) {
	if hi < 0xa1 || hi > 0xf9 || !((lo >= 0x40 && lo <= 0x7e) || (lo >= 0xa1 && lo <= 0xfe)) {
		return 0, 0, false
	}
	raw := etenRaw(hi, lo)
	switch {
	case raw <= etenLastSymbol:
		return EtenSymbol, raw, true
	case raw < etenBaseA440: // A3C0–A3FE 造字／控制區
		return 0, 0, false
	case raw <= etenLastCommon:
		return EtenStandard, raw - etenBaseA440, true
	case raw < etenBaseC940: // C6A1–C8FE 造字區
		return 0, 0, false
	default:
		idx := EtenCommonGlyphs + raw - etenBaseC940
		if idx >= EtenStandardGlyphs {
			return 0, 0, false
		}
		return EtenStandard, idx, true
	}
}

// EtenFonts 是唯讀載入的完整倚天漢字與全形符號字庫。
type EtenFonts struct {
	standard []Glyph
	symbol   []Glyph
	ascii    [][EtenASCIIStride]byte
	runes    map[rune]etenRuneRef
}

type etenRuneRef struct {
	kind EtenFontKind
	idx  int
}

// LoadEtenFonts 從使用者提供的目錄載入字庫；檔案不嵌入執行檔。
func LoadEtenFonts(dir string) (*EtenFonts, error) {
	load := func(name string, count int) ([]Glyph, error) {
		path := filepath.Join(dir, name)
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("assets: 讀取 %s：%w", name, err)
		}
		want := count * GlyphStride
		if len(b) != want {
			return nil, fmt.Errorf("assets: %s 大小 %d，預期 %d", name, len(b), want)
		}
		glyphs := make([]Glyph, count)
		for i := range glyphs {
			copy(glyphs[i][:], b[i*GlyphStride:(i+1)*GlyphStride])
		}
		return glyphs, nil
	}
	standard, err := load("STDFONT.15", EtenStandardGlyphs)
	if err != nil {
		return nil, err
	}
	symbol, err := load("SPCFONT.15", EtenSymbolGlyphs)
	if err != nil {
		return nil, err
	}
	asciiData, err := os.ReadFile(filepath.Join(dir, "ASCFONT.15"))
	if err != nil {
		return nil, fmt.Errorf("assets: 讀取 ASCFONT.15：%w", err)
	}
	if want := EtenASCIIGlyphs * EtenASCIIStride; len(asciiData) != want {
		return nil, fmt.Errorf("assets: ASCFONT.15 大小 %d，預期 %d", len(asciiData), want)
	}
	ascii := make([][EtenASCIIStride]byte, EtenASCIIGlyphs)
	for i := range ascii {
		copy(ascii[i][:], asciiData[i*EtenASCIIStride:(i+1)*EtenASCIIStride])
	}
	fonts := &EtenFonts{standard: standard, symbol: symbol, ascii: ascii,
		runes: make(map[rune]etenRuneRef)}
	decoder := traditionalchinese.Big5.NewDecoder()
	for hi := 0xa1; hi <= 0xf9; hi++ {
		for _, bounds := range [][2]int{{0x40, 0x7e}, {0xa1, 0xfe}} {
			for lo := bounds[0]; lo <= bounds[1]; lo++ {
				kind, idx, ok := EtenBig5Index(byte(hi), byte(lo))
				if !ok {
					continue
				}
				decoded, err := decoder.Bytes([]byte{byte(hi), byte(lo)})
				if err != nil {
					continue
				}
				rs := []rune(string(decoded))
				if len(rs) != 1 {
					continue
				}
				if _, exists := fonts.runes[rs[0]]; !exists {
					fonts.runes[rs[0]] = etenRuneRef{kind: kind, idx: idx}
				}
			}
		}
	}
	return fonts, nil
}

func (f *EtenFonts) GlyphAt(kind EtenFontKind, idx int) (Glyph, bool) {
	var glyphs []Glyph
	switch kind {
	case EtenStandard:
		glyphs = f.standard
	case EtenSymbol:
		glyphs = f.symbol
	default:
		return Glyph{}, false
	}
	if idx < 0 || idx >= len(glyphs) {
		return Glyph{}, false
	}
	return glyphs[idx], true
}

func (f *EtenFonts) GlyphBig5(hi, lo byte) (Glyph, bool) {
	kind, idx, ok := EtenBig5Index(hi, lo)
	if !ok {
		return Glyph{}, false
	}
	return f.GlyphAt(kind, idx)
}

// GlyphRune 由載入時反建的「倚天實際碼位 → Unicode」索引取字模。不能直接
// 使用 WHATWG Big5 encoder：像「偽」有重複碼，encoder 會偏好倚天範圍外的
// FA66，但字庫內真正存在的傳統 Big5 碼是 B0B0。
// ASCII／半形字元由 ASCFONT 處理；找不到時回 false，不安靜替代。
func (f *EtenFonts) GlyphRune(r rune) (Glyph, bool) {
	ref, ok := f.runes[r]
	if !ok {
		return Glyph{}, false
	}
	return f.GlyphAt(ref.kind, ref.idx)
}

// GlyphASCII 取得 ASCFONT.15 的 8×15 半形字模。只接受 ASCII；控制字元
// 由排版層處理，不在這裡偷偷換成空白。
func (f *EtenFonts) GlyphASCII(r rune) ([EtenASCIIStride]byte, bool) {
	if r < 0x20 || r > 0x7e || int(r) >= len(f.ascii) {
		return [EtenASCIIStride]byte{}, false
	}
	return f.ascii[int(r)], true
}
