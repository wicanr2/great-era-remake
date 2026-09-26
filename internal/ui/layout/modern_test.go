package layout

import "testing"

func TestModernMapLayoutUsesConfirmedDesignSurface(t *testing.T) {
	l, err := NewModernMapLayout(ModernDesignSurface())
	if err != nil {
		t.Fatal(err)
	}
	if l.Surface.Width != 1280 || l.Surface.Height != 720 {
		t.Fatalf("設計畫布錯誤：%+v", l.Surface)
	}
	if l.Rail.Intersects(l.Header) || l.Rail.Intersects(l.MapCard) || l.Rail.Intersects(l.InfoPanel) ||
		l.Rail.Intersects(l.EventStrip) || l.Header.Intersects(l.MapCard) || l.Header.Intersects(l.InfoPanel) ||
		l.InfoPanel.Intersects(l.MapCard) || l.InfoPanel.Intersects(l.EventStrip) || l.EventStrip.Intersects(l.MapCard) {
		t.Fatalf("H4 導覽軌／頂列／地圖窗／軍情／操作列重疊：rail=%+v header=%+v panel=%+v events=%+v map=%+v",
			l.Rail, l.Header, l.InfoPanel, l.EventStrip, l.MapCard)
	}
	if !l.MapCard.Contains(l.Map.X, l.Map.Y) || l.Map.Right() > l.MapCard.Right() || l.Map.Bottom() > l.MapCard.Bottom() {
		t.Fatalf("地圖未完整落在地圖卡：map=%+v card=%+v", l.Map, l.MapCard)
	}
	for name, button := range map[string]Rect{
		"command":   l.CommandButton,
		"narrative": l.NarrativeButton,
	} {
		if !l.EventStrip.Contains(button.X, button.Y) ||
			!l.EventStrip.Contains(button.Right()-1, button.Bottom()-1) {
			t.Fatalf("%s 按鈕不在操作／事件列內：%+v events=%+v", name, button, l.EventStrip)
		}
		if button.W < 48 || button.H < 48 {
			t.Fatalf("%s 未達 48 dp 設計下限：%+v", name, button)
		}
	}
	for name, button := range map[string]Rect{
		"map-marker":     l.MapRailMarker,
		"command-rail":   l.CommandRailButton,
		"narrative-rail": l.NarrativeRailButton,
	} {
		if !l.Rail.Contains(button.X, button.Y) || !l.Rail.Contains(button.Right()-1, button.Bottom()-1) {
			t.Fatalf("%s 不在導覽軌內：%+v rail=%+v", name, button, l.Rail)
		}
	}
	if l.CommandRailButton.Intersects(l.NarrativeRailButton) || l.CommandRailButton.Intersects(l.MapRailMarker) ||
		l.NarrativeRailButton.Intersects(l.MapRailMarker) {
		t.Fatalf("H4 導覽軌標記重疊：map=%+v command=%+v narrative=%+v",
			l.MapRailMarker, l.CommandRailButton, l.NarrativeRailButton)
	}
	if !l.EventStrip.Contains(l.SignalStrip.X, l.SignalStrip.Y) ||
		l.SignalStrip.Right() > l.EventStrip.Right() || l.SignalStrip.Bottom() > l.EventStrip.Bottom() {
		t.Fatalf("訊號列不在操作／事件列內：signal=%+v events=%+v", l.SignalStrip, l.EventStrip)
	}
	if len(l.MetricCards) != 6 {
		t.Fatalf("摘要卡數=%d，預期 6", len(l.MetricCards))
	}
	if !l.InfoPanel.Contains(l.CommanderPortrait.X, l.CommanderPortrait.Y) ||
		l.CommanderPortrait.Right() > l.InfoPanel.Right() || l.CommanderPortrait.Bottom() > l.InfoPanel.Bottom() {
		t.Fatalf("司令肖像不在軍情面板：portrait=%+v panel=%+v", l.CommanderPortrait, l.InfoPanel)
	}
	for i, card := range l.MetricCards {
		if card.W < 48 || card.H < 48 || !l.InfoPanel.Contains(card.X, card.Y) ||
			card.Right() > l.InfoPanel.Right() || card.Bottom() > l.InfoPanel.Bottom() {
			t.Fatalf("摘要卡 %d 超出策略儀表板：card=%+v panel=%+v", i, card, l.InfoPanel)
		}
		for j := 0; j < i; j++ {
			if card.Intersects(l.MetricCards[j]) {
				t.Fatalf("摘要卡 %d／%d 重疊", i, j)
			}
		}
		if card.Intersects(l.CommanderPortrait) {
			t.Fatalf("摘要卡 %d 與司令肖像重疊：card=%+v portrait=%+v", i, card, l.CommanderPortrait)
		}
	}
	if l.Map.W != 672 || l.Map.H != 522 {
		t.Fatalf("地圖高解析尺寸錯誤：%+v", l.Map)
	}
}

func TestModernMapLayoutPreservesStaggeredHexGeometry(t *testing.T) {
	l, err := NewModernMapLayout(ModernDesignSurface())
	if err != nil {
		t.Fatal(err)
	}
	even, err := l.MapCellRect(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	odd, err := l.MapCellRect(1, 0)
	if err != nil {
		t.Fatal(err)
	}
	if even.W != 48 || even.H != 36 || odd.W != 48 || odd.H != 36 {
		t.Fatalf("格子縮放錯誤：even=%+v odd=%+v", even, odd)
	}
	if odd.X-even.X != 48 || odd.Y-even.Y != 18 {
		t.Fatalf("奇數欄半格位移錯誤：even=%+v odd=%+v", even, odd)
	}
	if _, err := l.MapCellRect(14, 0); err == nil {
		t.Fatal("越界地圖格應拒絕")
	}
}

func TestModernMapLayoutRejectsNonDesignCanvas(t *testing.T) {
	if _, err := NewModernMapLayout(ModernSurface{Width: 1280, Height: 800}); err == nil {
		t.Fatal("尚未支援的 1280×800 不應靜默套用 1280×720 metrics")
	}
}

func TestModernPageAndCommandLayoutHaveNonOverlappingTouchCards(t *testing.T) {
	surface := ModernDesignSurface()
	page, err := NewModernPageLayout(surface)
	if err != nil {
		t.Fatal(err)
	}
	if page.Header.Intersects(page.Body) || page.Body.Intersects(page.Footer) {
		t.Fatal("頁首／內容／頁尾不應重疊")
	}
	command, err := NewModernCommandLayout(surface)
	if err != nil {
		t.Fatal(err)
	}
	if len(command.Cards) != 15 {
		t.Fatalf("指令卡數=%d，預期 15", len(command.Cards))
	}
	if command.Cards[0].W != 219 || command.Cards[0].H != 164 ||
		command.Cards[4].Y != command.Cards[0].Y || command.Cards[5].Y <= command.Cards[0].Y {
		t.Fatalf("H4 指令矩陣應為 5×3／219×164：first=%+v fifth=%+v sixth=%+v",
			command.Cards[0], command.Cards[4], command.Cards[5])
	}
	for i, card := range command.Cards {
		if card.W < 48 || card.H < 48 {
			t.Fatalf("指令卡 %d 過小：%+v", i, card)
		}
		for j := 0; j < i; j++ {
			if card.Intersects(command.Cards[j]) {
				t.Fatalf("指令卡 %d 與 %d 重疊", i, j)
			}
		}
	}
}

func TestModernOptionLayoutSupportsSettingsAndOtherCounts(t *testing.T) {
	for _, count := range []int{2, 7, 10} {
		options, err := NewModernOptionLayout(ModernDesignSurface(), count)
		if err != nil {
			t.Fatalf("count=%d: %v", count, err)
		}
		if len(options.Options) != count {
			t.Fatalf("count=%d got=%d", count, len(options.Options))
		}
		for i, option := range options.Options {
			if option.W < 48 || option.H < 48 {
				t.Fatalf("option %d 過小：%+v", i, option)
			}
			if !options.Page.Body.Contains(option.X+option.W/2, option.Y+option.H/2) {
				t.Fatalf("option %d 不在 body：%+v body=%+v", i, option, options.Page.Body)
			}
		}
	}
}

func TestModernBattleLayoutKeepsMapPanelAndControlsSeparate(t *testing.T) {
	l, err := NewModernBattleLayout(ModernDesignSurface())
	if err != nil {
		t.Fatal(err)
	}
	if l.MapCard.Intersects(l.Panel) || len(l.CommandButtons) != 5 || len(l.ControlButtons) != 3 || len(l.RetreatButtons) != 12 {
		t.Fatalf("Modern 戰鬥 layout=%+v", l)
	}
	if l.Log.W < 48 || l.Log.H < 24 || !l.Panel.Contains(l.Log.X, l.Log.Y) ||
		l.Log.Right() > l.Panel.Right() || l.Log.Bottom() > l.Panel.Bottom() {
		t.Fatalf("戰況訊息區不在面板：log=%+v panel=%+v", l.Log, l.Panel)
	}
	for i, card := range l.CommandButtons {
		if card.W < 48 || card.H < 48 {
			t.Fatalf("戰鬥命令 %d 過小：%+v", i, card)
		}
		for j := 0; j < i; j++ {
			if card.Intersects(l.CommandButtons[j]) {
				t.Fatalf("戰鬥命令 %d／%d 重疊", i, j)
			}
		}
		if card.Intersects(l.Log) {
			t.Fatalf("戰鬥命令 %d 與戰況訊息重疊：command=%+v log=%+v", i, card, l.Log)
		}
	}
	for i, card := range l.ControlButtons {
		if card.W < 48 || card.H < 48 || !l.Panel.Contains(card.X+card.W/2, card.Y+card.H/2) {
			t.Fatalf("戰鬥控制 %d 不在面板或過小：%+v", i, card)
		}
		if card.Intersects(l.Log) {
			t.Fatalf("戰鬥控制 %d 與戰況訊息重疊：control=%+v log=%+v", i, card, l.Log)
		}
	}
	odd, err := l.MapCellRect(1, 0)
	if err != nil {
		t.Fatal(err)
	}
	even, err := l.MapCellRect(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if odd.Y-even.Y != 18 || odd.X-even.X != 48 {
		t.Fatalf("戰鬥六角位移錯誤：even=%+v odd=%+v", even, odd)
	}
}

func TestModernFlowLayoutKeepsLargeListsAndKeypadInsideBody(t *testing.T) {
	surface := ModernDesignSurface()
	for _, count := range []int{2, 15, 39, 99} {
		layout, err := NewModernFlowLayout(surface, count)
		if err != nil {
			t.Fatalf("count=%d: %v", count, err)
		}
		for i, card := range layout.Cards {
			if card.W < 48 || card.H < 48 {
				t.Fatalf("count=%d card=%d too small: %+v", count, i, card)
			}
			if !layout.Page.Body.Contains(card.X, card.Y) || card.Right() > layout.Page.Body.Right() || card.Bottom() > layout.Page.Body.Bottom() {
				t.Fatalf("count=%d card=%d outside body: %+v body=%+v", count, i, card, layout.Page.Body)
			}
		}
	}
	layout, err := NewModernFlowLayout(surface, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(layout.Keypad) != 12 || layout.Input.W < 48 || layout.Input.H < 48 {
		t.Fatalf("numeric flow metrics = input=%+v keypad=%d", layout.Input, len(layout.Keypad))
	}
	for i, key := range layout.Keypad {
		if !layout.Page.Body.Contains(key.X, key.Y) || key.Right() > layout.Page.Body.Right() || key.Bottom() > layout.Page.Body.Bottom() {
			t.Fatalf("keypad %d outside body: %+v", i, key)
		}
	}
}
