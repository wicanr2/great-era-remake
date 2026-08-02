package textlayout

import (
	"encoding/json"
	"os"
	"testing"
)

func TestMixedWidthMeasurement(t *testing.T) {
	if got := MeasureHalfCells("1926年"); got != 6 {
		t.Fatalf("1926年 = %d 半格，預期 6（3 全形格）", got)
	}
}

func TestClosingPunctuationHangsInsteadOfStartingLine(t *testing.T) {
	doc, err := Layout("甲乙，丙", Options{Columns: 2, Rows: 4})
	if err != nil {
		t.Fatal(err)
	}
	lines := doc.Pages[0].Lines
	if len(lines) != 2 || lines[0].Text != "甲乙，" || lines[1].Text != "丙" {
		t.Fatalf("斷行 = %#v，預期 [甲乙，][丙]", lines)
	}
	if lines[0].HalfCells != 6 {
		t.Fatalf("懸掛標點行寬 = %d，預期可超出成 6 半格", lines[0].HalfCells)
	}
}

func TestOpeningPunctuationDoesNotEndLine(t *testing.T) {
	doc, err := Layout("甲「乙", Options{Columns: 2, Rows: 4})
	if err != nil {
		t.Fatal(err)
	}
	lines := doc.Pages[0].Lines
	if len(lines) != 2 || lines[0].Text != "甲" || lines[1].Text != "「乙" {
		t.Fatalf("斷行 = %#v，預期 [甲][「乙]", lines)
	}
}

func TestFixedRowsCreatePages(t *testing.T) {
	doc, err := Layout("一二三四五六七", Options{Columns: 1, Rows: 3})
	if err != nil {
		t.Fatal(err)
	}
	if len(doc.Pages) != 3 || len(doc.Pages[0].Lines) != 3 || len(doc.Pages[2].Lines) != 1 {
		t.Fatalf("分頁形狀 = %#v", doc.Pages)
	}
}

func TestExplicitBlankLineIsPreserved(t *testing.T) {
	doc, err := Layout("甲\n\n乙", Options{Columns: 4, Rows: 4})
	if err != nil {
		t.Fatal(err)
	}
	lines := doc.Pages[0].Lines
	if len(lines) != 3 || lines[1].Text != "" {
		t.Fatalf("空白行未保留：%#v", lines)
	}
}

func TestRejectsInvalidGeometry(t *testing.T) {
	if _, err := Layout("甲", Options{}); err == nil {
		t.Fatal("零尺寸應回錯")
	}
}

func TestAllBiographyBodiesPaginate(t *testing.T) {
	type person struct {
		ID    int    `json:"id"`
		Name  string `json:"name_ingame"`
		BioZH string `json:"bio_zh"`
	}
	b, err := os.ReadFile("../../../docs/reference/people/people.json")
	if err != nil {
		t.Fatal(err)
	}
	var people []person
	if err := json.Unmarshal(b, &people); err != nil {
		t.Fatal(err)
	}
	pageCounts := map[int]int{}
	var twoPages []string
	withBio := 0
	for _, p := range people {
		if p.BioZH == "" {
			continue
		}
		withBio++
		doc, err := Layout(p.BioZH, DefaultBiographyOptions)
		if err != nil {
			t.Fatalf("人物 %d %s：%v", p.ID, p.Name, err)
		}
		pageCounts[len(doc.Pages)]++
		if len(doc.Pages) == 2 {
			twoPages = append(twoPages, p.Name)
		}
		if len(doc.Pages) > 2 {
			t.Errorf("人物 %d %s 需要 %d 頁，超出設計資料分布", p.ID, p.Name, len(doc.Pages))
		}
	}
	if withBio != 326 {
		t.Fatalf("有傳人物 = %d，預期目前 people.json 的 326", withBio)
	}
	t.Logf("326 篇自傳頁數分布：%v；兩頁人物：%v", pageCounts, twoPages)
}
