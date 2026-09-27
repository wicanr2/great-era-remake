package main

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
	uiresolution "github.com/wicanr2/great-era-remake/internal/ui/resolution"
	"github.com/wicanr2/great-era-remake/internal/ui/textlayout"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

func TestOpenBiographyJoinsRosterAndPaginates(t *testing.T) {
	localeDir := filepath.Join("..", "..", "translations", "zh-Hant")
	sharedDir := filepath.Join("..", "..", "translations", "shared")
	people, err := i18n.LoadPeople(localeDir, sharedDir)
	if err != nil {
		t.Fatal(err)
	}
	wording, err := i18n.LoadWording(localeDir)
	if err != nil {
		t.Fatal(err)
	}
	var slot int
	for candidate := 1; candidate <= 274; candidate++ {
		person, ok := people.PersonAt(1, candidate)
		if ok && person != nil && person.Biography != "" {
			slot = candidate
			break
		}
	}
	if slot == 0 {
		t.Fatal("第一期找不到可顯示自傳槽位")
	}
	a := &app{
		people: people, wording: wording, eten: &assets.EtenFonts{},
		wordingMode: i18n.WordingOriginal, stage: 1,
		viewGenerals: []game.GeneralID{game.GeneralID(slot)}, viewIndex: 0,
		screen: screenViewGeneral,
	}
	a.openBiography(screenViewGeneral)
	if a.screen != screenBiography || a.bioPages <= 0 || a.bioPage != 0 {
		t.Fatalf("自傳入口狀態錯誤：screen=%d page=%d/%d", a.screen, a.bioPage, a.bioPages)
	}
}

func TestOpenBiographyUsesH4PageMetricsWhenModernHighIsSelected(t *testing.T) {
	localeDir := filepath.Join("..", "..", "translations", "zh-Hant")
	sharedDir := filepath.Join("..", "..", "translations", "shared")
	people, err := i18n.LoadPeople(localeDir, sharedDir)
	if err != nil {
		t.Fatal(err)
	}
	wording, err := i18n.LoadWording(localeDir)
	if err != nil {
		t.Fatal(err)
	}
	// 尋找一筆 H4 比例欄寬確實和 640×350 原式分頁不同的既有資料；這能
	// 防止 app 又回頭以舊頁數公告不存在的 NextPage 命中區。
	slot, expectedPages := 0, 0
	for candidate := 1; candidate <= 274; candidate++ {
		person, ok := people.PersonAt(1, candidate)
		if !ok || person == nil || person.Biography == "" {
			continue
		}
		legacy, legacyErr := textlayout.Layout(person.Biography, textlayout.DefaultBiographyOptions)
		modernPages, modernErr := render.ModernBiographyPageCount(person.Biography)
		if legacyErr == nil && modernErr == nil && len(legacy.Pages) != modernPages {
			slot, expectedPages = candidate, modernPages
			break
		}
	}
	if slot == 0 {
		t.Fatal("找不到可驗證 H4 與 legacy 分頁差異的人物資料")
	}
	a := &app{
		people: people, wording: wording, eten: &assets.EtenFonts{}, wordingMode: i18n.WordingOriginal,
		stage: 1, viewGenerals: []game.GeneralID{game.GeneralID(slot)}, viewIndex: 0,
		screen: screenViewGeneral, resolution: uiresolution.ModeHigh, themeMode: uitheme.ModeModern,
	}
	a.openBiography(screenViewGeneral)
	if a.screen != screenBiography || a.bioPages != expectedPages || a.bioPage != 0 {
		t.Fatalf("H4 自傳頁數=%d/%d，預期 0/%d", a.bioPage, a.bioPages, expectedPages)
	}
	for _, target := range a.navigationTargets() {
		if target.Action == actions.NextPage && expectedPages <= 1 {
			t.Fatalf("H4 單頁自傳不得公告不存在的下一頁：%+v", target)
		}
	}
}

func TestBiographyRosterKeepsUnknownFallbackButExcludesPlaceholder(t *testing.T) {
	localeDir := filepath.Join("..", "..", "translations", "zh-Hant")
	sharedDir := filepath.Join("..", "..", "translations", "shared")
	people, err := i18n.LoadPeople(localeDir, sharedDir)
	if err != nil {
		t.Fatal(err)
	}
	wording, err := i18n.LoadWording(localeDir)
	if err != nil {
		t.Fatal(err)
	}
	common := func(slot int) *app {
		return &app{people: people, wording: wording, eten: &assets.EtenFonts{},
			wordingMode: i18n.WordingOriginal, stage: 1,
			viewGenerals: []game.GeneralID{game.GeneralID(slot)}, viewIndex: 0,
			screen: screenViewGeneral}
	}

	// #274 是名冊中的「無省長」排除槽，不應先畫入口再在下一頁報錯。
	placeholder := common(274)
	if _, ok := placeholder.currentBiography(); ok {
		t.Fatal("無省長排除槽不應提供人物資料")
	}
	for _, target := range placeholder.interactiveTargets() {
		if target.Action == actions.OpenBiography {
			t.Fatalf("無省長不應有自傳命中區：%+v", placeholder.interactiveTargets())
		}
	}
	placeholder.openBiography(screenViewGeneral)
	if placeholder.screen != screenViewGeneral || placeholder.bioPages != 0 {
		t.Fatalf("無省長不應進入自傳頁：screen=%d pages=%d", placeholder.screen, placeholder.bioPages)
	}

	// 找一位有槽位但沒有正文的人物；這條路徑仍可讀，頁面由 wording
	// fallback 呈現「查無可靠傳記記載」，不以杜撰內容填空。
	unknownSlot := 0
	for slot := 1; slot <= 274; slot++ {
		p, ok := people.PersonAt(1, slot)
		if ok && p != nil && p.Biography == "" {
			unknownSlot = slot
			break
		}
	}
	if unknownSlot == 0 {
		t.Fatal("第一期應有至少一位 unknown 人物供 fallback 驗收")
	}
	unknown := common(unknownSlot)
	if _, ok := unknown.currentBiography(); !ok {
		t.Fatalf("unknown 槽 %d 應保留人物入口", unknownSlot)
	}
	biographyTargets := 0
	for _, target := range unknown.interactiveTargets() {
		if target.Action == actions.OpenBiography {
			biographyTargets++
		}
	}
	if biographyTargets != 1 {
		t.Fatalf("unknown 槽 %d 應有一個自傳命中區，得到 %+v", unknownSlot, unknown.interactiveTargets())
	}
	unknown.openBiography(screenViewGeneral)
	if unknown.screen != screenBiography || unknown.bioPages != 1 {
		t.Fatalf("unknown fallback 應有一頁：screen=%d pages=%d", unknown.screen, unknown.bioPages)
	}
}

func TestBiographyFallbackBodyListsSourcedRecordOnly(t *testing.T) {
	p := &i18n.Person{Faction: "奉系", Periods: []string{"北伐時期"}}
	got := biographyFallbackBody(p, "查無可靠傳記記載", "登錄資料")
	want := "查無可靠傳記記載\n登錄資料：奉系／北伐時期"
	if got != want {
		t.Fatalf("檔案卡正文=%q，應為 %q", got, want)
	}
	empty := &i18n.Person{}
	got = biographyFallbackBody(empty, "查無可靠傳記記載", "登錄資料")
	want = "查無可靠傳記記載\n登錄資料：—／—"
	if got != want {
		t.Fatalf("空登錄檔案卡=%q，應為 %q", got, want)
	}
}

func TestBiographyAllRosterSlotsJoinOrExplicitlyExclude(t *testing.T) {
	people, err := i18n.LoadPeople(filepath.Join("..", "..", "translations", "zh-Hant"),
		filepath.Join("..", "..", "translations", "shared"))
	if err != nil {
		t.Fatal(err)
	}
	checked, excluded := 0, 0
	for period, limit := range map[int]int{1: 274, 2: 106, 3: 106} {
		for slot := 1; slot <= limit; slot++ {
			if _, isExcluded := people.ExclusionReason(period, slot); isExcluded {
				excluded++
				continue
			}
			if p, ok := people.PersonAt(period, slot); !ok || p == nil {
				t.Fatalf("期別 %d 槽 %d 沒有接合人物", period, slot)
			}
			checked++
		}
	}
	if checked != 485 || excluded != 1 {
		t.Fatalf("名冊槽位可查／排除 = %d/%d，預期 485/1", checked, excluded)
	}
}
