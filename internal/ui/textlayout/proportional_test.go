package textlayout

import "testing"

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
	if got := doc.Pages[0].Lines; len(got) != 2 || got[0].Text != "甲乙，" || got[1].Text != "丙" {
		t.Fatalf("禁則分行=%#v", got)
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
