package assets

import "fmt"

// .15 字模檔：16×15 點陣，每字 30 bytes，無檔頭。
//
// 文字不經過索引表——字模序列本身就是文本。每 w 個字模是一個詞條，
// 不足的格子填全零字模。詳見 docs/formats/01-glyph-text.md。

const (
	GlyphW      = 16
	GlyphH      = 15
	GlyphRowB   = (GlyphW + 7) / 8 // 2
	GlyphStride = GlyphRowB * GlyphH
)

// Glyph 是一個 30 bytes 的字模。
type Glyph [GlyphStride]byte

// IsBlank 回報這是不是全零字模。
//
// 全零字模是排版填充或執行期的數字佔位，不是缺字——原版 51 個字模檔裡
// 有 1,374 個（22%），全部是這兩種用途。
func (g Glyph) IsBlank() bool {
	for _, b := range g {
		if b != 0 {
			return false
		}
	}
	return true
}

// At 回報 (x, y) 這個點是不是亮的。每列 MSB-first。
func (g Glyph) At(x, y int) bool {
	if x < 0 || x >= GlyphW || y < 0 || y >= GlyphH {
		return false
	}
	return g[y*GlyphRowB+x>>3]&(0x80>>(x&7)) != 0
}

// GlyphFile 是一個 .15 字模檔。
type GlyphFile struct {
	Glyphs []Glyph
}

// ParseGlyphFile 解析 .15 字模檔。
func ParseGlyphFile(data []byte) (*GlyphFile, error) {
	if len(data) == 0 || len(data)%GlyphStride != 0 {
		return nil, fmt.Errorf("assets: .15 大小 %d 不是 %d 的倍數", len(data), GlyphStride)
	}
	n := len(data) / GlyphStride
	f := &GlyphFile{Glyphs: make([]Glyph, n)}
	for i := 0; i < n; i++ {
		copy(f.Glyphs[i][:], data[i*GlyphStride:])
	}
	return f, nil
}

// Entry 取出第 k 個詞條的字模（每 slotWidth 個字模一條），並去掉填充的空白。
func (f *GlyphFile) Entry(k, slotWidth int) ([]Glyph, error) {
	if slotWidth <= 0 {
		return nil, fmt.Errorf("assets: 槽寬必須為正，得到 %d", slotWidth)
	}
	start := k * slotWidth
	if start < 0 || start+slotWidth > len(f.Glyphs) {
		return nil, fmt.Errorf("assets: 詞條 %d 超出範圍（共 %d 條）",
			k, len(f.Glyphs)/slotWidth)
	}
	out := make([]Glyph, 0, slotWidth)
	for _, g := range f.Glyphs[start : start+slotWidth] {
		if !g.IsBlank() {
			out = append(out, g)
		}
	}
	return out, nil
}

// EntryCount 回報依 slotWidth 切出來有幾個詞條。
func (f *GlyphFile) EntryCount(slotWidth int) int {
	if slotWidth <= 0 {
		return 0
	}
	return len(f.Glyphs) / slotWidth
}
