package main

import (
	"testing"

	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	"github.com/wicanr2/great-era-remake/internal/ui/actions"
)

func TestPointerTargetsMapVisibleOtherOptionsAndDisplayRows(t *testing.T) {
	mapApp := &app{screen: screenMap}
	mapTargets := mapApp.pointerTargets()
	if len(mapTargets) != 1 || mapTargets[0].Rect.W != 90 || mapTargets[0].Rect.H != 48 {
		t.Fatalf("地圖指令入口=%+v", mapTargets)
	}
	mr := mapTargets[0].Rect
	if got := actions.Hit(mapTargets, mr.X+mr.W/2, mr.Y+mr.H/2); got != actions.OpenCommands {
		t.Fatalf("地圖指令入口中央=%q", got)
	}

	a := &app{screen: screenOtherOptions}
	targets := a.pointerTargets()
	if len(targets) != 10 {
		t.Fatalf("其他選項 targets=%d", len(targets))
	}
	for i := 0; i < 9; i++ {
		r := targets[i].Rect
		if got := actions.Hit(targets, r.X+r.W/2, r.Y+r.H/2); got != actions.Selection(i+1) {
			t.Fatalf("第 %d 項中央=%q", i+1, got)
		}
	}

	a.screen = screenDisplayOptions
	targets = a.pointerTargets()
	if actions.Hit(targets, 400, 120) != actions.Select1 ||
		actions.Hit(targets, 400, 172) != actions.Select2 ||
		actions.Hit(targets, 230, 325) != actions.Back {
		t.Fatal("顯示設定命中區與 renderer 版面不符")
	}
}

func TestCommandFifteenTargetWorksInBothWordingLayouts(t *testing.T) {
	for _, mode := range []i18n.WordingMode{i18n.WordingOriginal, i18n.WordingPlain} {
		a := &app{screen: screenCommand, wordingMode: mode}
		target := a.pointerTargets()[14]
		if got := actions.Hit(a.pointerTargets(), target.Rect.X+10, target.Rect.Y+target.Rect.H/2); got != actions.Select15 {
			t.Fatalf("mode=%s got=%q", mode, got)
		}
	}
}

func TestPointerTargetsCoverSimpleM0Menus(t *testing.T) {
	tests := []struct {
		name   string
		screen screen
		count  int
	}{
		{"發展", screenDevelop, 3},
		{"調動方式", screenTransferMode, 2},
		{"商業方式", screenTradeMode, 2},
		{"徵兵或整編", screenRecruitAction, 2},
		{"徵兵兵種", screenRecruitBranch, 4},
		{"整編兵種", screenReorganizeBranch, 4},
		{"秘密行動", screenCovertAction, 2},
		{"查閱", screenViewMenu, 4},
		{"他省查閱", screenViewProvinceChoice, 2},
		{"政策", screenPolicy, 2},
	}
	for _, mode := range []i18n.WordingMode{i18n.WordingOriginal, i18n.WordingPlain} {
		for _, tt := range tests {
			t.Run(tt.name+"/"+string(mode), func(t *testing.T) {
				a := &app{screen: tt.screen, wordingMode: mode}
				targets := a.pointerTargets()
				if len(targets) != tt.count {
					t.Fatalf("targets=%d want=%d", len(targets), tt.count)
				}
				for i, target := range targets {
					got := actions.Hit(targets, target.Rect.X+target.Rect.W/2,
						target.Rect.Y+target.Rect.H/2)
					if got != actions.Selection(i+1) {
						t.Fatalf("第 %d 項中央=%q", i+1, got)
					}
				}
			})
		}
	}
}

func TestTradeGoodTargetsFollowImportExportCount(t *testing.T) {
	for _, tc := range []struct {
		importing bool
		want      int
	}{{true, 3}, {false, 5}} {
		a := &app{screen: screenTradeGood, tradeImport: tc.importing}
		if got := len(a.pointerTargets()); got != tc.want {
			t.Fatalf("import=%v targets=%d want=%d", tc.importing, got, tc.want)
		}
	}
}

func TestPointerConfirmTargetsCoverGameplayConfirmations(t *testing.T) {
	for _, s := range []screen{screenTransferConfirm, screenRecruitConfirm, screenTrainConfirm} {
		a := &app{screen: s}
		targets := a.pointerTargets()
		if len(targets) != 2 || targets[0].Action != actions.Confirm || targets[1].Action != actions.Cancel {
			t.Fatalf("screen=%d targets=%+v", s, targets)
		}
	}
}

func TestNavigationTargetsAreVisibleAndTakePriority(t *testing.T) {
	a := &app{screen: screenCommand}
	nav := a.navigationTargets()
	if len(nav) != 1 || nav[0].Action != actions.Back || nav[0].Rect.W != 48 || nav[0].Rect.H != 48 {
		t.Fatalf("command nav=%+v", nav)
	}
	if nav[0].Rect.Y != 294 {
		t.Fatalf("政略返回鈕 y=%d，應避開右欄選項", nav[0].Rect.Y)
	}
	r := nav[0].Rect
	if got := actions.Hit(a.interactiveTargets(), r.X+r.W/2, r.Y+r.H/2); got != actions.Back {
		t.Fatalf("可見返回鈕=%q", got)
	}

	a.screen, a.viewPage = screenViewProvinceNames, 1
	nav = a.navigationTargets()
	if len(nav) != 2 || nav[0].Action != actions.Back || nav[1].Action != actions.NextPage {
		t.Fatalf("省名翻頁 nav=%+v", nav)
	}
}

func TestGeneralListTargetsUseAbsoluteSelectionNumbers(t *testing.T) {
	ids := make([]game.GeneralID, 25)
	a := &app{screen: screenViewGenerals, viewGenerals: ids, viewIndex: 20}
	targets := a.pointerTargets()
	if len(targets) != 5 {
		t.Fatalf("第二頁 targets=%d", len(targets))
	}
	for i, target := range targets {
		n, ok := actions.SelectionNumber(target.Action)
		if !ok || n != 21+i {
			t.Fatalf("第 %d 個 action=%q parsed=%d", i, target.Action, n)
		}
	}
}

func TestNumericKeypadMappingAndGeometry(t *testing.T) {
	for _, s := range []screen{screenTransferTarget, screenTransferAmount, screenTradeAmount,
		screenSupplyTarget, screenSupplyAmount, screenRecruitAmount,
		screenReorganizeAmount, screenCovertTarget, screenMessageTime} {
		a := &app{screen: s}
		targets := a.numericKeypadTargets()
		if len(targets) != 12 {
			t.Fatalf("screen=%d targets=%d", s, len(targets))
		}
		want := []actions.Action{actions.Digit1, actions.Digit2, actions.Digit3,
			actions.Digit4, actions.Digit5, actions.Digit6, actions.Digit7,
			actions.Digit8, actions.Digit9, actions.Digit0, actions.DeleteDigit, actions.Submit}
		for i, target := range targets {
			if target.Action != want[i] || target.Rect.W != 64 || target.Rect.H != 48 {
				t.Fatalf("screen=%d key=%d target=%+v want=%q", s, i, target, want[i])
			}
			if got := actions.Hit(targets, target.Rect.X+32, target.Rect.Y+24); got != want[i] {
				t.Fatalf("screen=%d key=%d center=%q", s, i, got)
			}
		}
	}

	a := &app{screen: screenViewProvinceSelect}
	if a.numericKeypadVisible() || len(a.numericKeypadTargets()) != 0 {
		t.Fatal("39 省清單應直接點列，不得以鍵盤遮住資料")
	}
}

func TestDirectProvinceAndTargetListsKeepStableSelection(t *testing.T) {
	tests := []struct {
		name string
		app  app
		want []int
	}{
		{"調動目標", app{screen: screenTransferTarget,
			transferTargets: []game.ProvinceID{25, 27}}, []int{1, 2}},
		{"運補目標", app{screen: screenSupplyTarget,
			supplyTargets: []game.ProvinceID{19, 25, 27}}, []int{1, 2, 3}},
		{"查閱三十九省", app{screen: screenViewProvinceSelect}, []int{1, 20, 39}},
		{"自治候選", app{screen: screenAutonomy,
			autonomyTargets: []game.ProvinceID{19, 25, 26, 27}}, []int{1, 2, 3, 4}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			targets := tt.app.pointerTargets()
			for _, n := range tt.want {
				if n > len(targets) {
					t.Fatalf("targets=%d，缺第 %d 項", len(targets), n)
				}
				target := targets[n-1]
				got, ok := actions.SelectionNumber(actions.Hit(targets,
					target.Rect.X+target.Rect.W/2, target.Rect.Y+target.Rect.H/2))
				if !ok || got != n {
					t.Fatalf("第 %d 項中央 parsed=(%d,%v)", n, got, ok)
				}
			}
		})
	}
}

func TestTransferSelectionShowsSubmitOnlyAfterSelection(t *testing.T) {
	// nil 工作階段只能返回；實際工作階段的勾選／送出由規則層測試與 GUI 驗收覆蓋。
	a := &app{screen: screenTransferSelection}
	buttons := a.navigationActions()
	if len(buttons) != 1 || buttons[0] != actions.Back {
		t.Fatalf("未選人時 buttons=%v", buttons)
	}
}

func TestProductionRowsThenConditionalKeypad(t *testing.T) {
	a := &app{screen: screenProduction}
	rows := a.pointerTargets()
	if len(rows) != 4 || a.numericKeypadVisible() {
		t.Fatalf("未選項目 rows=%d keypad=%v", len(rows), a.numericKeypadVisible())
	}
	for i, target := range rows {
		got := actions.Hit(rows, target.Rect.X+target.Rect.W/2, target.Rect.Y+target.Rect.H/2)
		if got != actions.Selection(i+1) {
			t.Fatalf("產能第 %d 列=%q", i+1, got)
		}
	}

	a.productionItem = 3
	if !a.numericKeypadVisible() || len(a.numericKeypadTargets()) != 12 {
		t.Fatal("選定產能項目後應顯示共用數字鍵盤")
	}
	if got := len(a.pointerTargets()); got != 0 {
		t.Fatalf("輸入比例時不得同時保留列選取 target=%d", got)
	}
}

func TestPointerGestureRejectsDragAndScreenChange(t *testing.T) {
	start := pointerPress{x: 100, y: 100, screen: screenOtherOptions}
	if !withinClick(start, 106, 105, screenOtherOptions) {
		t.Fatal("小移動應仍是點擊")
	}
	if withinClick(start, 109, 100, screenOtherOptions) {
		t.Fatal("超過門檻的拖曳不得點擊")
	}
	if withinClick(start, 100, 100, screenDisplayOptions) {
		t.Fatal("畫面改變後的放開不得派送舊動作")
	}
}
