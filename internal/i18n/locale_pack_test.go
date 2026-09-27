package i18n

import (
	"path/filepath"
	"regexp"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

func TestEnglishAndJapanesePacksCoverRuntimeSurface(t *testing.T) {
	baseDir := filepath.Join("..", "..", "translations")
	base, err := LoadWording(filepath.Join(baseDir, "zh-Hant"))
	if err != nil {
		t.Fatal(err)
	}
	formatTokens := func(s string) []string { return regexp.MustCompile(`%[a-zA-Z]`).FindAllString(s, -1) }
	for _, language := range []string{"en", "ja"} {
		t.Run(language, func(t *testing.T) {
			dir := filepath.Join(baseDir, language)
			loc, err := Load(dir)
			if err != nil {
				t.Fatal(err)
			}
			if loc.Language != language {
				t.Fatalf("語系標記 = %q，預期 %q", loc.Language, language)
			}
			if loc.Len("3.15") < ProvinceCount {
				t.Fatalf("省名表只有 %d 筆", loc.Len("3.15"))
			}
			catalog, err := LoadWording(dir)
			if err != nil {
				t.Fatal(err)
			}
			narrative, err := LoadNarrative(dir)
			if err != nil {
				t.Fatal(err)
			}
			if narrative.Language != language || narrative.Source != "NEWSDATA.DAT" {
				t.Fatalf("敘事語系／來源 = %q/%q", narrative.Language, narrative.Source)
			}
			if text, status, ok := narrative.Text(0, WordingPlain); !ok || status != NarrativeTranslated || text == "" {
				t.Fatalf("已知敘事模板未載入：%q/%q/%v", text, status, ok)
			}
			if _, status, ok := narrative.Text(1, WordingPlain); ok || status != NarrativeSourceImage {
				t.Fatalf("未知敘事不應被冒充翻譯：%q/%v", status, ok)
			}
			for key, source := range base.Entries {
				translated, ok := catalog.Entries[key]
				if !ok {
					t.Fatalf("缺少語意鍵 %q", key)
				}
				for mode, pair := range map[string][2]string{
					"original": {translated.Original, source.Original},
					"plain":    {translated.Plain, source.Plain},
				} {
					got, want := pair[0], pair[1]
					if gotTokens, wantTokens := formatTokens(got), formatTokens(want); len(gotTokens) != len(wantTokens) {
						t.Fatalf("%s/%s 的格式參數數量 %d != %d", key, mode, len(gotTokens), len(wantTokens))
					}
				}
			}
			people, err := LoadPeople(dir, filepath.Join(baseDir, "shared"))
			if err != nil {
				t.Fatal(err)
			}
			if people.Language != language || people.PersonCount() != 417 {
				t.Fatalf("人物語系／數量 = %q/%d", people.Language, people.PersonCount())
			}
			biographies := 0
			for _, person := range people.people {
				if person.Biography != "" {
					biographies++
				}
			}
			if biographies != translatedOverlayPeople {
				t.Fatalf("%s source-fallback 正文數 = %d，預期 %d", language, biographies, translatedOverlayPeople)
			}
			p, ok := people.PersonAt(1, 1)
			if !ok || p.Biography == "" || p.BiographyLanguage != language || p.BiographyStatus != "wiki-sourced" {
				t.Fatalf("人物自傳 wiki-sourced 標記未明示：%+v", p)
			}
			if _, ok := people.PersonAt(1, 274); ok {
				t.Fatal("無省長排除槽不應因語系包複製而重新出現")
			}
		})
	}
}

func TestLocalizedWordingUsesAvailableEtenGlyphs(t *testing.T) {
	fonts, err := assets.LoadEtenFonts(filepath.Join("..", "..", "workplace", "eten"))
	if err != nil {
		t.Fatal(err)
	}
	for _, language := range []string{"en", "ja"} {
		catalog, err := LoadWording(filepath.Join("..", "..", "translations", language))
		if err != nil {
			t.Fatal(err)
		}
		missing := map[rune]bool{}
		for _, entry := range catalog.Entries {
			for _, text := range []string{entry.Original, entry.Plain} {
				for _, r := range text {
					if r >= 0x20 && r <= 0x7e {
						if _, ok := fonts.GlyphASCII(r); !ok {
							missing[r] = true
						}
					} else if _, ok := fonts.GlyphRune(r); !ok {
						missing[r] = true
					}
				}
			}
		}
		if len(missing) != 0 {
			t.Fatalf("%s wording 有倚天字庫未覆蓋字元：%q", language, string(sortedRunes(missing)))
		}
		narrative, err := LoadNarrative(filepath.Join("..", "..", "translations", language))
		if err != nil {
			t.Fatal(err)
		}
		missingNarrative := map[rune]bool{}
		for i := 0; i < NarrativeCount; i++ {
			for _, mode := range []WordingMode{WordingOriginal, WordingPlain} {
				if text, _, ok := narrative.Text(i, mode); ok {
					for _, r := range text {
						if r >= 0x20 && r <= 0x7e {
							if _, ok := fonts.GlyphASCII(r); !ok {
								missingNarrative[r] = true
							}
						} else if _, ok := fonts.GlyphRune(r); !ok {
							missingNarrative[r] = true
						}
					}
				}
			}
		}
		if len(missingNarrative) != 0 {
			t.Fatalf("%s narrative 有倚天字庫未覆蓋字元：%q", language, string(sortedRunes(missingNarrative)))
		}
	}
}

func sortedRunes(set map[rune]bool) []rune {
	out := make([]rune, 0, len(set))
	for r := range set {
		out = append(out, r)
	}
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}
