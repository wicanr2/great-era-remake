package layout

// BattleCommandButton 是戰鬥右側五項原版命令的命中區。
// 文字仍維持原版 2.15 的緊湊排版；命中區只作裝置輸入外殼，
// 不改變規則或鍵盤契約。
func BattleCommandButton(index int) Placement {
	const (
		panelX = 451
		panelW = 189
		top    = 160
		rowH   = 24
		margin = 2
	)
	if index < 0 || index >= 5 {
		return Placement{}
	}
	return Placement{
		X: panelX + margin, Y: top + index*rowH,
		HitX: panelX + margin, HitY: top + index*rowH,
		HitW: panelW - margin*2, HitH: rowH,
	}
}

// BattleControlButton 是戰鬥面板底部三個 remake 控制鍵的共用幾何。
// 這三個按鍵以 48×48 邏輯像素提供較大的觸控目標：攻擊沿用既有
// 「相鄰敵軍」捷徑，換部隊等同 Tab，結束回合等同 Space。
// 它們是現代化外殼，不宣稱是原版畫面或原版輸入。
func BattleControlButton(index int) Placement {
	const (
		panelX = 451
		top    = 286
		width  = 56
		height = 48
		gap    = 4
	)
	if index < 0 || index >= 3 {
		return Placement{}
	}
	return Placement{
		X: panelX + index*(width+gap), Y: top,
		HitX: panelX + index*(width+gap), HitY: top,
		HitW: width, HitH: height,
	}
}

// BattleRetreatKeypadButton 是戰鬥撤退輸入用的 3×4 數字鍵盤；索引 0..11
// 依序為 1..9、0、刪除、送出。每個命中區維持 48×48 邏輯像素，讓桌面
// 滑鼠與 Android 單指觸控共用同一份幾何。鍵盤會覆蓋右側的 remake 選單，
// 不碰左側戰場與已量測的戰鬥資源欄位。
func BattleRetreatKeypadButton(index int) Placement {
	const (
		originX, originY = 458, 146
		width, height    = 56, 48
		gap              = 4
		columns          = 3
	)
	if index < 0 || index >= 12 {
		return Placement{}
	}
	col, row := index%columns, index/columns
	x := originX + col*(width+gap)
	y := originY + row*(height+gap)
	return Placement{X: x, Y: y, HitX: x, HitY: y, HitW: width, HitH: height}
}
