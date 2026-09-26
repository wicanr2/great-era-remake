package render

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/i18n"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

func TestModernBiographyLongOverlaysRenderWithEmbeddedFontOnly(t *testing.T) {
	style := uitheme.NewModern().Style()
	root := filepath.Join("..", "..", "..")
	for _, locale := range []string{"en", "ja"} {
		t.Run(locale, func(t *testing.T) {
			db, err := i18n.LoadPeople(filepath.Join(root, "translations", locale),
				filepath.Join(root, "translations", "shared"))
			if err != nil {
				t.Fatal(err)
			}
			var longest *i18n.Person
			for id := 1; id <= 1000; id++ {
				person, ok := db.PersonByID(id)
				if !ok || len([]rune(person.Biography)) <= lenRunes(longest) {
					continue
				}
				longest = person
			}
			if longest == nil {
				t.Fatal("找不到 overlay 自傳")
			}
			if len([]rune(longest.Biography)) < 200 {
				t.Fatalf("最長自傳過短：%d rune", len([]rune(longest.Biography)))
			}
			first, err := NewCanvas(1280, 720).DrawModernBiographySurface(nil, BiographyView{
				Person: longest, Page: 0, Title: "人物自傳", Unavailable: "查無可靠傳記記載",
			}, style.Ink, style.Paper, style, nil, "尚未找到可驗證照片", "返回", "上一頁", "下一頁")
			if err != nil {
				t.Fatal(err)
			}
			// 無頭像版使用完整文件寬度，過去必須分成兩頁的稿件可能
			// 合併為一頁；這裡只鎖定有有效頁數且所有字都可顯示。
			if first.PageCount < 1 || len(first.Missing) != 0 {
				t.Fatalf("最長 %s 自傳結果=%+v，未閉合比例長文／缺字 gate", locale, first)
			}
			last, err := NewCanvas(1280, 720).DrawModernBiographySurface(nil, BiographyView{
				Person: longest, Page: first.PageCount - 1, Title: "人物自傳", Unavailable: "查無可靠傳記記載",
			}, style.Ink, style.Paper, style, nil, "尚未找到可驗證照片", "返回", "上一頁", "下一頁")
			if err != nil {
				t.Fatal(err)
			}
			if last.PageCount != first.PageCount || len(last.Missing) != 0 {
				t.Fatalf("最長 %s 自傳末頁結果=%+v", locale, last)
			}
		})
	}
}

func lenRunes(person *i18n.Person) int {
	if person == nil {
		return 0
	}
	return len([]rune(person.Biography))
}
