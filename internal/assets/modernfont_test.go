package assets

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/i18n"
)

func TestEmbeddedModernFontHasUnicodeCoverage(t *testing.T) {
	font, err := EmbeddedModernFont()
	if err != nil {
		t.Fatal(err)
	}
	if w, h := font.CellSize(); w != 32 || h != 32 {
		t.Fatalf("Modern glyph cell=%dx%d，預期 32×32", w, h)
	}
	if font.Version() != 2 {
		t.Fatalf("Modern atlas 版本=%d，預期 v2", font.Version())
	}
	if font.GlyphCount() < 2000 {
		t.Fatalf("Modern atlas glyph 數過少：%d", font.GlyphCount())
	}
	for _, r := range []rune{'A', '0', '中', '臺', 'あ', '日', '。', '？'} {
		glyph, ok := font.Glyph(r)
		if !ok || glyph.Advance() <= 0 {
			t.Fatalf("Modern atlas 缺少 U+%04X", r)
		}
		if glyph.PixelAdvance() > 32 || glyph.Advance() > 16 {
			t.Fatalf("Modern atlas 字距未正規化 U+%04X: base=%d pixel=%d", r, glyph.Advance(), glyph.PixelAdvance())
		}
	}
}

func TestEmbeddedModernFontFamiliesAreV2AndDistinct(t *testing.T) {
	sans, err := EmbeddedModernFontFamily(ModernFontSans)
	if err != nil {
		t.Fatal(err)
	}
	serif, err := EmbeddedModernFontFamily(ModernFontSerif)
	if err != nil {
		t.Fatal(err)
	}
	if sans.Version() != 2 || serif.Version() != 2 {
		t.Fatalf("字型家族未使用 GEMF v2：sans=%d serif=%d", sans.Version(), serif.Version())
	}
	if sans.GlyphCount() != serif.GlyphCount() {
		t.Fatalf("字型家族 coverage 不一致：sans=%d serif=%d", sans.GlyphCount(), serif.GlyphCount())
	}
	var different bool
	for _, r := range []rune{'A', '中', 'あ', '。'} {
		a, aok := sans.Glyph(r)
		b, bok := serif.Glyph(r)
		if !aok || !bok {
			t.Fatalf("字型家族缺 glyph U+%04X", r)
		}
		for y := 0; y < 32 && !different; y++ {
			for x := 0; x < 32; x++ {
				if a.AlphaAt(x, y) != b.AlphaAt(x, y) {
					different = true
					break
				}
			}
		}
	}
	if !different {
		t.Fatal("sans／serif atlas 不應完全相同")
	}
}

func TestParseModernFontRejectsBadOrderAndLength(t *testing.T) {
	valid := ModernFontBytesForTest('A', '中')
	if _, err := ParseModernFont(valid); err != nil {
		t.Fatal(err)
	}
	if _, err := ParseModernFont(valid[:len(valid)-1]); err == nil {
		t.Fatal("截斷 Modern atlas 應該失敗")
	}
	if _, err := ParseModernFont(ModernFontBytesForTest('中', 'A')); err == nil {
		t.Fatal("未排序 code point 應該失敗")
	}
}

func TestParseModernFontV2AlphaAndV1Compatibility(t *testing.T) {
	v2, err := ParseModernFont(ModernFontV2BytesForTest('A'))
	if err != nil {
		t.Fatal(err)
	}
	g, ok := v2.Glyph('A')
	if !ok || g.AlphaAt(0, 0) != 64 || g.AlphaAt(1, 0) != 128 || g.AlphaAt(2, 0) != 255 {
		t.Fatalf("v2 alpha fixture 解析錯誤：ok=%v alpha=%d/%d/%d", ok, g.AlphaAt(0, 0), g.AlphaAt(1, 0), g.AlphaAt(2, 0))
	}
	if g.Advance() != 16 || g.PixelAdvance() != 32 {
		t.Fatalf("v2 字距正規化錯誤：base=%d pixel=%d", g.Advance(), g.PixelAdvance())
	}
	v1, err := ParseModernFont(ModernFontBytesForTest('A'))
	if err != nil {
		t.Fatal(err)
	}
	if v1.Version() != 1 {
		t.Fatalf("v1 相容 parser 版本失敗：%d", v1.Version())
	}
	if w, h := v1.CellSize(); w != 16 || h != 16 {
		t.Fatalf("v1 相容 parser 尺寸失敗：%dx%d", w, h)
	}
}

func TestEmbeddedModernFontCoversTranslatedBiographies(t *testing.T) {
	font, err := EmbeddedModernFont()
	if err != nil {
		t.Fatal(err)
	}
	root := filepath.Join("..", "..")
	for _, locale := range []string{"en", "ja"} {
		t.Run(locale, func(t *testing.T) {
			db, err := i18n.LoadPeople(filepath.Join(root, "translations", locale),
				filepath.Join(root, "translations", "shared"))
			if err != nil {
				t.Fatal(err)
			}
			missing := map[rune]bool{}
			for id := 1; id <= 1000; id++ {
				person, ok := db.PersonByID(id)
				if !ok {
					continue
				}
				for _, r := range person.Biography {
					if _, ok := font.Glyph(r); !ok {
						missing[r] = true
					}
				}
			}
			if len(missing) != 0 {
				t.Fatalf("%s 人物 overlay 有 Modern atlas 缺字：%v", locale, missing)
			}
		})
	}
}
