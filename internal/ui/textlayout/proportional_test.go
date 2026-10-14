package textlayout

import (
	"strings"
	"testing"
)

func TestLayoutProportionalUsesMeasuredAdvanceAndPages(t *testing.T) {
	advance := func(r rune) int {
		if r <= 0x7f {
			return 4
		}
		return 16
	}
	doc, err := LayoutProportional("AB中CDEFGHIJKLM", ProportionalOptions{Width: 24, Rows: 2, Advance: advance})
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Pages) != 2 || len(doc.Pages[0].Lines) != 2 || doc.Pages[0].Lines[0].Text != "AB中" {
		t.Fatalf("比例分頁=%#v，未依像素字距斷行", doc)
	}
	if got := doc.Pages[0].Lines[0].PixelWidth; got != 24 {
		t.Fatalf("PixelWidth=%d，預期 24", got)
	}
}

func TestLayoutProportionalPreservesPunctuationRules(t *testing.T) {
	doc, err := LayoutProportional("甲乙，丙", ProportionalOptions{
		Width: 32, Rows: 4, Advance: func(r rune) int {
			if r == '，' {
				return 16
			}
			return 16
		},
	})
	if err != nil {
		t.Fatal(err)
	}
	if got := doc.Pages[0].Lines; len(got) != 3 || got[0].Text != "甲" || got[1].Text != "乙，" || got[2].Text != "丙" {
		t.Fatalf("禁則分行=%#v", got)
	}
}

func TestJapaneseProportionalBoundariesAndFullText(t *testing.T) {
	for _, text := range []string{"甲乙きゃくー。丙丁", "甲乙（『かな』）丙丁", "甲乙カッパ々。丙", "甲乙……丙丁", "甲乙\n\n（かな）。"} {
		doc, err := LayoutProportional(text, ProportionalOptions{Width: 64, Rows: 2, Advance: func(rune) int { return 16 }})
		if err != nil {
			t.Fatalf("%q: %v", text, err)
		}
		var recovered string
		for _, page := range doc.Pages {
			for _, line := range page.Lines {
				v := []rune(line.Text)
				if line.PixelWidth > 64 {
					t.Fatalf("overflow: %+v", line)
				}
				if len(v) > 0 && (strings.ContainsRune(proportionalNoStart, v[0]) || strings.ContainsRune(proportionalNoEnd, v[len(v)-1])) {
					t.Fatalf("illegal boundary: %q", line.Text)
				}
				recovered += line.Text
			}
		}
		if recovered != strings.ReplaceAll(text, "\n", "") {
			t.Fatalf("lost text: %q", recovered)
		}
	}
}

func TestEnglishWordWrapAndLongURL(t *testing.T) {
	doc, err := LayoutProportional("One three four https://example.com/averylongword", ProportionalOptions{Width: 10, Rows: 20, Advance: func(rune) int { return 1 }})
	if err != nil {
		t.Fatal(err)
	}
	var recovered string
	for _, line := range doc.Pages[0].Lines {
		if line.PixelWidth > 10 {
			t.Fatal(line)
		}
		recovered += line.Text
	}
	if strings.TrimSpace(doc.Pages[0].Lines[0].Text) != "One three" || !strings.HasPrefix(doc.Pages[0].Lines[1].Text, "four ") {
		t.Fatalf("split a fitting word: %#v", doc)
	}
	if recovered != "One three four https://example.com/averylongword" {
		t.Fatal(recovered)
	}
}

func TestProportionalDoesNotAbsorbExplicitNewline(t *testing.T) {
	doc, err := LayoutProportional("甲\n（乙）", ProportionalOptions{Width: 48, Rows: 4, Advance: func(rune) int { return 16 }})
	if err != nil {
		t.Fatal(err)
	}
	if doc.Pages[0].Lines[0].Text != "甲" || doc.Pages[0].Lines[1].Text != "（乙）" {
		t.Fatal(doc)
	}
}

func TestLayoutProportionalRejectsMissingMetrics(t *testing.T) {
	if _, err := LayoutProportional("甲", ProportionalOptions{Width: 16, Rows: 1}); err == nil {
		t.Fatal("缺少 Advance 應失敗")
	}
	if _, err := LayoutProportional("甲", ProportionalOptions{Width: 16, Rows: 1, Advance: func(rune) int { return 0 }}); err == nil {
		t.Fatal("零字距應失敗")
	}
}
