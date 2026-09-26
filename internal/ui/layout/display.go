package layout

// DisplayWordingOption 回傳顯示設定頁「用語」選項的繪製與命中幾何。
// renderer 與滑鼠／觸控輸入共用這份位置，避免新增選項時兩邊漂移。
func DisplayWordingOption(index, x, y, width int) Placement {
	return Grid(index, x, y+112, 2, 0, 52, 15, width, 48)
}

// DisplayThemeOption 回傳顯示設定頁「圖形主題」選項的繪製與命中幾何。
// 兩個主題並排，保留原版顯示設定頁底部的返回提示與可點擊空間。
func DisplayThemeOption(index, x, y, width int) Placement {
	const gap = 8
	choiceWidth := (width - gap) / 2
	return Grid(index, x, y+226, 1, choiceWidth+gap, 0, 15, choiceWidth, 48)
}

// NarrativeButton 是地圖右上角的史事入口，與 OpenCommandButton 並排。
func NarrativeButton(logicalWidth int) Placement {
	const width, height, margin, gap = 90, 48, 8, 8
	x := logicalWidth - margin - width - gap - width
	return Placement{X: x, Y: margin, HitX: x, HitY: margin, HitW: width, HitH: height}
}
