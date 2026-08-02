// Package layout 提供 renderer 與指標命中共用的版面幾何。
package layout

type Placement struct {
	X, Y       int
	HitX, HitY int
	HitW, HitH int
}

// Grid 依先往下、再往右的順序建立第 i 項。hitTop 是命中區相對
// 文字基準線的上移量；數值由 renderer 的真實版面常數傳入。
func Grid(i, originX, originY, perColumn, columnWidth, rowHeight, hitTop, hitWidth, hitHeight int) Placement {
	x := originX + (i/perColumn)*columnWidth
	y := originY + (i%perColumn)*rowHeight
	return Placement{X: x, Y: y, HitX: x, HitY: y - hitTop, HitW: hitWidth, HitH: hitHeight}
}

// OpenCommandButton 是地圖上供滑鼠／觸控叫出指令選單的重製版操作外殼。
// 90×48 邏輯像素同時供 renderer 與命中測試使用，避免兩邊各自硬編座標。
func OpenCommandButton(logicalWidth int) Placement {
	const width, height, margin = 90, 48, 8
	x, y := logicalWidth-width-margin, margin
	return Placement{X: x, Y: y, HitX: x, HitY: y, HitW: width, HitH: height}
}

// NavigationButton 從指定高度的右側往左排列 48×48 重製版導覽按鈕。
// slot=0 是最右側；renderer 與輸入命中共用這份幾何。
func NavigationButton(logicalWidth, y, slot int) Placement {
	const size, margin, gap = 48, 8, 8
	x := logicalWidth - margin - size - slot*(size+gap)
	return Placement{X: x, Y: y, HitX: x, HitY: y, HitW: size, HitH: size}
}

// NumericKeypadButton 是 6×2 觸控數字鍵盤；索引 0..11 依序為
// 1..9、0、刪除、送出。64×48 邏輯像素高於 48 px 操作基準。
func NumericKeypadButton(index int) Placement {
	const originX, originY = 202, 250
	const width, height, gap = 64, 48, 4
	col, row := index%6, index/6
	x, y := originX+col*(width+gap), originY+row*(height+gap)
	return Placement{X: x, Y: y, HitX: x, HitY: y, HitW: width, HitH: height}
}
