package main

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/i18n"
	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	uiaudio "github.com/wicanr2/great-era-remake/internal/ui/audio"
	uiresolution "github.com/wicanr2/great-era-remake/internal/ui/resolution"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

func loadTraditionalWordingForModernFlowTest(t *testing.T) *i18n.WordingCatalog {
	t.Helper()
	dir := filepath.Join("..", "..", "translations", "zh-Hant")
	catalog, err := i18n.LoadWording(dir)
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}

func TestHighModernFlowTargetsUseDesignCardsAndKeypad(t *testing.T) {
	wording := loadTraditionalWordingForModernFlowTest(t)
	a := &app{resolution: uiresolution.ModeHigh, themeMode: uitheme.ModeModern,
		wording: wording, wordingMode: i18n.WordingOriginal, screen: screenDevelop}
	targets := a.pointerTargets()
	if len(targets) != 3 {
		t.Fatalf("高解析發展選單 targets=%d", len(targets))
	}
	for i, target := range targets {
		if target.Rect.W < 48 || target.Rect.H < 48 {
			t.Fatalf("流程卡片 %d 過小：%+v", i, target.Rect)
		}
		if got := actions.Hit(targets, target.Rect.X+target.Rect.W/2, target.Rect.Y+target.Rect.H/2); got != actions.Selection(i+1) {
			t.Fatalf("流程卡片 %d action=%q", i, got)
		}
	}
	if x, y, ok := a.basePointerPosition(targets[0].Rect.X+10, targets[0].Rect.Y+10); !ok || x != targets[0].Rect.X+10 || y != targets[0].Rect.Y+10 {
		t.Fatalf("高解析流程座標不應反算：%d,%d,%v", x, y, ok)
	}

	a.screen = screenTradeAmount
	keypad := a.pointerTargets()
	if len(keypad) != 12 || keypad[0].Action != actions.Digit1 || keypad[9].Action != actions.DeleteDigit || keypad[11].Action != actions.Submit {
		t.Fatalf("高解析流程數字鍵盤=%+v", keypad)
	}
	nav := a.navigationTargets()
	if len(nav) != 1 || nav[0].Action != actions.Back || nav[0].Rect.W < 48 {
		t.Fatalf("高解析流程返回鍵=%+v", nav)
	}
}

func TestModernFlowUsesActualStrategyCommandTitles(t *testing.T) {
	wording := loadTraditionalWordingForModernFlowTest(t)
	a := &app{wording: wording, wordingMode: i18n.WordingOriginal}
	for _, tc := range []struct {
		name   string
		screen screen
		want   string
	}{
		{name: "運補", screen: screenSupplyTarget, want: "運補"},
		{name: "徵兵", screen: screenRecruitAction, want: "徵兵"},
		{name: "查閱", screen: screenViewMenu, want: "查閱"},
		{name: "開發", screen: screenDevelop, want: "開發"},
		{name: "秘密行動", screen: screenCovertAction, want: "秘密行動"},
		{name: "商業活動", screen: screenTradeMode, want: "商業活動"},
		{name: "練兵", screen: screenTrainConfirm, want: "練兵"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a.screen = tc.screen
			data, err := a.modernFlowData()
			if err != nil {
				t.Fatal(err)
			}
			if data.Title != tc.want {
				t.Fatalf("screen=%d 標題=%q，預期實際政略卡 %q", tc.screen, data.Title, tc.want)
			}
		})
	}
}

func TestModernAudioUsesCampaignAndSceneCuesByScreen(t *testing.T) {
	audio, err := uiaudio.NewModernProceduralTracks(8000)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = audio.Close() })
	a := &app{audio: audio, audioMode: uiaudio.ModeModern}
	for _, test := range []struct {
		screen screen
		want   uiaudio.Track
	}{
		{screen: screenMap, want: uiaudio.TrackMainTheme},
		{screen: screenCommand, want: uiaudio.TrackMainTheme},
		{screen: screenBiography, want: uiaudio.TrackWall},
		{screen: screenViewGeneral, want: uiaudio.TrackWall},
		{screen: screenViewGenerals, want: uiaudio.TrackWall},
		{screen: screenNarrative, want: uiaudio.TrackScene},
		{screen: screenDevelop, want: uiaudio.TrackStrategy},
		{screen: screenBattle, want: uiaudio.TrackBattle1},
		{screen: screenQuit, want: uiaudio.TrackFinal},
	} {
		a.screen = test.screen
		if got := a.desiredAudioTrack(); got != test.want {
			t.Fatalf("screen=%d cue=%q，預期 %q", test.screen, got, test.want)
		}
	}
}
