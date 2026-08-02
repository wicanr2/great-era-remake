package render

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"unicode"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/i18n"
)

func loadBiographyTestFonts(t *testing.T) *assets.EtenFonts {
	t.Helper()
	f, err := assets.LoadEtenFonts(filepath.Join("..", "..", "..", "workplace", "eten"))
	if err != nil {
		t.Skipf("沒有使用者倚天字庫：%v", err)
	}
	return f
}

func TestAllBiographyDisplayTextHasOnlyLedgeredMissingRunes(t *testing.T) {
	fonts := loadBiographyTestFonts(t)
	var file struct {
		People []i18n.Person `json:"people"`
	}
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "translations", "zh-Hant", "people.json"))
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(b, &file); err != nil {
		t.Fatal(err)
	}
	missing := map[rune]bool{}
	for _, p := range file.People {
		// 自傳頁尾只顯示來源筆數，不直接把 URL（可能含簡體路徑）畫上畫面。
		fields := []string{p.NameInGame, p.NameCommon, p.Courtesy, p.Birthplace,
			p.Faction, p.HighestPost, p.Biography, p.Confidence}
		for _, r := range strings.Join(fields, " ") {
			switch {
			case unicode.IsSpace(r) || r == '—' || r == '·':
			case r >= 0x20 && r <= 0x7e:
				if _, ok := fonts.GlyphASCII(r); !ok {
					missing[r] = true
				}
			default:
				if _, ok := fonts.GlyphRune(r); !ok {
					missing[r] = true
				}
			}
		}
	}
	var got []rune
	for r := range missing {
		got = append(got, r)
	}
	sort.Slice(got, func(i, j int) bool { return got[i] < got[j] })
	if want := []rune{'榘', '藁'}; !reflect.DeepEqual(got, want) {
		t.Fatalf("發行人物資料的缺字集合 = %q，預期帳本中的 %q", string(got), string(want))
	}
}

func TestDrawBiographyConfirmedAndPagination(t *testing.T) {
	fonts := loadBiographyTestFonts(t)
	birth, death := 1887, 1975
	p := &i18n.Person{NameInGame: "測試者", NameCommon: "測試者", Birth: &birth, Death: &death,
		Birthplace: "浙江", Faction: "中央軍", HighestPost: "委員長", Confidence: "confirmed",
		Sources: []string{"one"}, Biography: repeatBiography("民國人物生平。", 100)}
	c := NewBGICanvas()
	res, err := c.DrawBiography(fonts, BiographyView{Person: p, Title: "人物自傳",
		Unavailable: "查無可靠傳記記載"}, assets.RGB{R: 174}, assets.RGB{R: 255, G: 255, B: 162})
	if err != nil {
		t.Fatal(err)
	}
	if res.PageCount < 2 {
		t.Fatalf("長文應分頁，得到 %d 頁", res.PageCount)
	}
	if len(res.Missing) != 0 {
		t.Fatalf("一般中文與 ASCII 不應缺字：%q", string(res.Missing))
	}
}

func TestDrawBiographyUnavailableAndVisibleMissingGlyph(t *testing.T) {
	fonts := loadBiographyTestFonts(t)
	p := &i18n.Person{NameInGame: "韓復榘", Confidence: "unknown"}
	c := NewBGICanvas()
	res, err := c.DrawBiography(fonts, BiographyView{Person: p, Title: "人物自傳",
		Unavailable: "查無可靠傳記記載"}, assets.RGB{R: 174}, assets.RGB{R: 255, G: 255, B: 162})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(res.Missing, []rune{'榘'}) {
		t.Fatalf("缺字帳不符：%q", string(res.Missing))
	}
	// 缺字框位於標題列姓名的末字位置；至少必須有前景像素，不能安靜留白。
	if c.Image().RGBAAt(160, 10).R == 255 {
		t.Fatal("缺字位置應畫出可見方框")
	}
}

func TestDrawBiographyRejectsBadPage(t *testing.T) {
	fonts := loadBiographyTestFonts(t)
	c := NewBGICanvas()
	_, err := c.DrawBiography(fonts, BiographyView{Person: &i18n.Person{NameInGame: "甲"},
		Page: 1, Title: "人物自傳", Unavailable: "查無記載"}, assets.RGB{}, assets.RGB{})
	if err == nil {
		t.Fatal("超出頁數應回錯")
	}
}

func TestBiographyConfidenceLabelsArePlayerFacingChinese(t *testing.T) {
	want := map[string]string{
		"confirmed": "確認", "partial": "部分資料", "unknown": "未知",
		"high": "高", "medium": "中", "low": "低", "future": "未標示",
	}
	for input, expected := range want {
		if got := biographyConfidenceLabel(input); got != expected {
			t.Errorf("%q = %q，預期 %q", input, got, expected)
		}
	}
}

func repeatBiography(s string, n int) string {
	out := ""
	for i := 0; i < n; i++ {
		out += s
	}
	return out
}
