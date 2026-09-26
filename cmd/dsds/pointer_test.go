package main

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
	uiresolution "github.com/wicanr2/great-era-remake/internal/ui/resolution"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
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

	highMap := &app{screen: screenMap, resolution: uiresolution.ModeHigh, themeMode: uitheme.ModeModern}
	highTargets := highMap.pointerTargets()
	highLayout, err := uilayout.NewModernMapLayout(uilayout.ModernDesignSurface())
	if err != nil {
		t.Fatal(err)
	}
	if len(highTargets) != 2 {
		t.Fatalf("H4 地圖指令入口數=%d targets=%+v", len(highTargets), highTargets)
	}
	wantCommandTargets := []actions.Rect{
		{X: highLayout.CommandButton.X, Y: highLayout.CommandButton.Y, W: highLayout.CommandButton.W, H: highLayout.CommandButton.H},
		{X: highLayout.CommandRailButton.X, Y: highLayout.CommandRailButton.Y, W: highLayout.CommandRailButton.W, H: highLayout.CommandRailButton.H},
	}
	for i, want := range wantCommandTargets {
		if highTargets[i].Action != actions.OpenCommands || highTargets[i].Rect != want {
			t.Fatalf("H4 指令入口 %d=%+v，預期 rect=%+v", i, highTargets[i], want)
		}
		if got := actions.Hit(highTargets, want.X+want.W/2, want.Y+want.H/2); got != actions.OpenCommands {
			t.Fatalf("H4 指令入口 %d 中央=%q", i, got)
		}
	}
	hr := highTargets[0].Rect
	x, y, ok := highMap.basePointerPosition(hr.X+hr.W/2, hr.Y+hr.H/2)
	if !ok || x != hr.X+hr.W/2 || y != hr.Y+hr.H/2 {
		t.Fatalf("H1 地圖輸入不應反算回 640×350：%d,%d,%v", x, y, ok)
	}

	highNarrative := &app{screen: screenMap, resolution: uiresolution.ModeHigh, themeMode: uitheme.ModeModern,
		narrative: &i18n.NarrativeCatalog{}, narrativeImages: []*assets.Image{{W: 1, H: 1, Pix: []byte{0}}}}
	narrativeTargets := highNarrative.pointerTargets()
	if len(narrativeTargets) != 4 {
		t.Fatalf("H4 新聞入口數=%d targets=%+v", len(narrativeTargets), narrativeTargets)
	}
	for _, target := range narrativeTargets[2:] {
		if target.Action != actions.OpenNarrative || actions.Hit(narrativeTargets,
			target.Rect.X+target.Rect.W/2, target.Rect.Y+target.Rect.H/2) != actions.OpenNarrative {
			t.Fatalf("H4 新聞入口未派送既有 action：%+v", target)
		}
	}

	highCommand := &app{screen: screenCommand, resolution: uiresolution.ModeHigh, themeMode: uitheme.ModeModern}
	commandTargets := highCommand.pointerTargets()
	if len(commandTargets) != 15 || commandTargets[14].Rect.W < 48 || commandTargets[14].Rect.H < 48 {
		t.Fatalf("H1 指令卡片=%+v", commandTargets)
	}
	last := commandTargets[14].Rect
	if got := actions.Hit(commandTargets, last.X+last.W/2, last.Y+last.H/2); got != actions.Select15 {
		t.Fatalf("H1 第 15 項中央=%q", got)
	}
	if x, y, ok := highCommand.basePointerPosition(last.X+last.W/2, last.Y+last.H/2); !ok || x != last.X+last.W/2 || y != last.Y+last.H/2 {
		t.Fatalf("H1 指令頁輸入不應反算：%d,%d,%v", x, y, ok)
	}
	highCommand.screen = screenResolutionOptions
	resolutionTargets := highCommand.pointerTargets()
	if len(resolutionTargets) != 2 {
		t.Fatalf("H1 解析度選項=%+v", resolutionTargets)
	}
	for i, target := range resolutionTargets {
		if got := actions.Hit(resolutionTargets, target.Rect.X+target.Rect.W/2, target.Rect.Y+target.Rect.H/2); got != actions.Selection(i+1) {
			t.Fatalf("H1 解析度第 %d 項=%q", i+1, got)
		}
	}

	a := &app{screen: screenOtherOptions}
	targets := a.pointerTargets()
	if len(targets) != 11 {
		t.Fatalf("其他選項 targets=%d", len(targets))
	}
	for i := 0; i < 10; i++ {
		r := targets[i].Rect
		if got := actions.Hit(targets, r.X+r.W/2, r.Y+r.H/2); got != actions.Selection(i+1) {
			t.Fatalf("第 %d 項中央=%q", i+1, got)
		}
	}
	if targets[9].Action != actions.Select10 {
		t.Fatalf("第 10 項應開啟解析度設定：%q", targets[9].Action)
	}

	a.screen = screenDisplayOptions
	targets = a.pointerTargets()
	if actions.Hit(targets, 400, 120) != actions.Select1 ||
		actions.Hit(targets, 400, 172) != actions.Select2 ||
		actions.Hit(targets, 230, 235) != actions.Select3 ||
		actions.Hit(targets, 500, 235) != actions.Select4 ||
		actions.Hit(targets, 230, 325) != actions.Back {
		t.Fatal("顯示設定命中區與 renderer 版面不符")
	}
}

func TestBattlePointerTargetsExposeCommandsAndLargeControls(t *testing.T) {
	a := &app{
		screen: screenBattle,
		battle: &battleState{mode: battleModeCommand, sim: &game.BattleSim{}},
	}
	targets := a.battlefieldTargets()
	if len(targets) != 8 {
		t.Fatalf("戰鬥命中區=%d，應有五項命令加三個控制鍵", len(targets))
	}
	for i := 0; i < 5; i++ {
		r := targets[i].Rect
		if targets[i].Action != actions.BattleCommand(i+1) ||
			r.W != 185 || r.H != 24 {
			t.Fatalf("命令 %d target=%+v", i+1, targets[i])
		}
		if got := actions.Hit(targets, r.X+r.W/2, r.Y+r.H/2); got != actions.BattleCommand(i+1) {
			t.Fatalf("命令 %d 中央命中=%q", i+1, got)
		}
	}
	wantControls := []actions.Action{actions.BattleAttack, actions.BattleNextUnit, actions.BattleEndTurn}
	for i, want := range wantControls {
		r := targets[5+i].Rect
		if targets[5+i].Action != want || r.W != 56 || r.H != 48 {
			t.Fatalf("控制 %d target=%+v", i, targets[5+i])
		}
		if got := actions.Hit(targets, r.X+r.W/2, r.Y+r.H/2); got != want {
			t.Fatalf("控制 %d 中央命中=%q", i, got)
		}
	}
}

func TestHighModernBattlePointerTargetsUseDesignMetrics(t *testing.T) {
	a := &app{screen: screenBattle, resolution: uiresolution.ModeHigh, themeMode: uitheme.ModeModern,
		battle: &battleState{mode: battleModeCommand, sim: &game.BattleSim{}}}
	targets := a.battlefieldTargets()
	if len(targets) != 8 {
		t.Fatalf("H1 戰鬥命中區=%d，應有五命令加三控制", len(targets))
	}
	if targets[0].Rect.W < 300 || targets[0].Rect.H < 48 {
		t.Fatalf("H1 戰鬥命令過小：%+v", targets[0])
	}
	for i, target := range targets {
		if got := actions.Hit(targets, target.Rect.X+target.Rect.W/2, target.Rect.Y+target.Rect.H/2); got != target.Action {
			t.Fatalf("H1 戰鬥 target %d 中央=%q 預期 %q", i, got, target.Action)
		}
	}
	r := targets[7].Rect
	if x, y, ok := a.basePointerPosition(r.X+r.W/2, r.Y+r.H/2); !ok || x != r.X+r.W/2 || y != r.Y+r.H/2 {
		t.Fatalf("H1 戰鬥輸入不應反算：%d,%d,%v", x, y, ok)
	}
}

func TestBattleAttackPointerTargetsExposeAdjacentNumberedEnemy(t *testing.T) {
	attacker := &game.Combatant{
		CombatUnit: game.CombatUnit{General: 1, Cell: 100},
		Strength:   game.StrengthInput{Branch: game.Branch1, Force: 1000},
	}
	neighbor, ok := attacker.Cell.Neighbour(game.DirDown)
	if !ok {
		t.Fatal("測試格沒有合法鄰格")
	}
	defender := &game.Combatant{
		CombatUnit: game.CombatUnit{General: 2, Cell: neighbor},
		Strength:   game.StrengthInput{Branch: game.Branch1, Force: 1000},
	}
	a := &app{screen: screenBattle, battle: &battleState{
		sim:  &game.BattleSim{Attacker: []*game.Combatant{attacker}, Defender: []*game.Combatant{defender}},
		sel:  0,
		mode: battleModeAttack,
	}}
	targets := a.battlefieldTargets()
	want := actions.BattleAttackTarget(1)
	var found *actions.Target
	for i := range targets {
		if targets[i].Action == want {
			found = &targets[i]
			break
		}
	}
	if found == nil {
		t.Fatalf("攻擊子狀態遺漏相鄰目標：%+v", targets)
	}
	if got := actions.Hit(targets, found.Rect.X+found.Rect.W/2, found.Rect.Y+found.Rect.H/2); got != want {
		t.Fatalf("相鄰攻擊目標中央=%q，預期 %q", got, want)
	}
}

func TestBattleRetreatPointerTargetsShareNumericActions(t *testing.T) {
	a := &app{
		screen: screenBattle,
		battle: &battleState{mode: battleModeRetreat, sim: &game.BattleSim{}},
	}
	targets := a.battlefieldTargets()
	if len(targets) != 12 {
		t.Fatalf("撤退鍵盤命中區=%d，應有 1..9、0、刪除、送出", len(targets))
	}
	want := []actions.Action{actions.Digit1, actions.Digit2, actions.Digit3,
		actions.Digit4, actions.Digit5, actions.Digit6, actions.Digit7,
		actions.Digit8, actions.Digit9, actions.Digit0, actions.DeleteDigit,
		actions.Submit}
	for i, target := range targets {
		if target.Action != want[i] || target.Rect.W != 56 || target.Rect.H != 48 {
			t.Fatalf("撤退按鍵 %d target=%+v want=%q", i, target, want[i])
		}
		if got := actions.Hit(targets, target.Rect.X+target.Rect.W/2,
			target.Rect.Y+target.Rect.H/2); got != want[i] {
			t.Fatalf("撤退按鍵 %d 中央命中=%q want=%q", i, got, want[i])
		}
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

func TestCeasefireCommandAndNumericTargetUseSharedActions(t *testing.T) {
	a := &app{screen: screenCommand, wordingMode: i18n.WordingOriginal}
	targets := a.pointerTargets()
	if got := actions.Hit(targets, targets[9].Rect.X+10, targets[9].Rect.Y+targets[9].Rect.H/2); got != actions.Select10 {
		t.Fatalf("政略第 10 項 pointer=%q，預期 %q", got, actions.Select10)
	}
	a.screen = screenCeasefireTarget
	keypad := a.numericKeypadTargets()
	if len(keypad) != 12 || keypad[0].Action != actions.Digit1 || keypad[11].Action != actions.Submit {
		t.Fatalf("停火數字鍵盤=%+v", keypad)
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
		{"外交", screenDiplomacy, 3},
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

func TestGeneralDetailExposesVisibleBiographyPointerTarget(t *testing.T) {
	people, err := i18n.LoadPeople(filepath.Join("..", "..", "translations", "zh-Hant"),
		filepath.Join("..", "..", "translations", "shared"))
	if err != nil {
		t.Fatal(err)
	}
	wording := &i18n.WordingCatalog{Entries: map[string]i18n.WordingEntry{
		"biography.page": {Original: "人物自傳", Plain: "人物生平"},
	}}
	var slot int
	for candidate := 1; candidate <= 274; candidate++ {
		if _, ok := people.PersonAt(1, candidate); ok {
			slot = candidate
			break
		}
	}
	if slot == 0 {
		t.Fatal("找不到可供自傳入口測試的名冊槽位")
	}
	a := &app{screen: screenViewGeneral, eten: &assets.EtenFonts{}, wording: wording,
		people: people, stage: 1, viewGenerals: []game.GeneralID{game.GeneralID(slot)}}
	targets := a.pointerTargets()
	if len(targets) != 1 || targets[0].Action != actions.OpenBiography {
		t.Fatalf("人物詳細自傳 target=%+v", targets)
	}
	r := targets[0].Rect
	if r.W != 144 || r.H != 48 || r.Y != 294 {
		t.Fatalf("人物自傳命中區=%+v", r)
	}
	if got := actions.Hit(a.interactiveTargets(), r.X+r.W/2, r.Y+r.H/2); got != actions.OpenBiography {
		t.Fatalf("人物自傳中央命中=%q", got)
	}
}

func TestNumericKeypadMappingAndGeometry(t *testing.T) {
	for _, s := range []screen{screenTransferTarget, screenTransferAmount, screenTradeAmount,
		screenSupplyTarget, screenSupplyAmount, screenRecruitAmount,
		screenReorganizeAmount, screenCovertTarget, screenLoanAmount,
		screenRepayAmount, screenCeasefireTarget, screenMessageTime} {
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
