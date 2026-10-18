package layout

// BattleReturnButton is shared by drawing and pointer input on both surfaces.
func BattleReturnButton(high bool) Rect {
	if high {
		return Rect{1120, 6, 148, 30}
	}
	return Rect{8, 6, 118, 22}
}

const DeskWidth, DeskHeight, DeskToolbarHeight = 1280, 742, 58

var DeskHeader = Rect{0, 0, 1280, 76}
var DeskSidebar = Rect{0, 76, 180, 633}
var DeskInspector = Rect{1020, 76, 260, 633}
var DeskContent = Rect{180, 140, 840, 502}
var DeskBack = Rect{198, 87, 66, 42}
var DeskHome = Rect{884, 87, 118, 42}
var DeskMap = Rect{200, 241, 800, 378}
var DeskActions = Rect{180, 642, 840, 67}
var DeskFooter = Rect{0, 709, 1280, 33}
var DeskDialog = Rect{380, 230, 520, 250}
var DeskDialogNo = Rect{610, 410, 130, 42}
var DeskDialogYes = Rect{750, 410, 124, 42}

func DeskSidebarButton(i int) Rect   { return Rect{12, 134 + i*54, 155, 45} }
func DeskFormRow(i int) Rect         { return Rect{218, 274 + i*57, 764, 42} }
func DeskListRow(i, scroll int) Rect { return Rect{218, 405 + i*60 - scroll, 764, 54} }
func DeskToolbarField(i int) Rect {
	fields := [...]Rect{{46, 11, 146, 36}, {236, 11, 148, 36}, {423, 11, 127, 36}, {594, 11, 200, 36}, {837, 11, 183, 36}, {1120, 11, 144, 36}}
	if i < 0 || i >= len(fields) {
		return Rect{}
	}
	return fields[i]
}
func DeskToolbarMenu(field, rows, noteLines int) Rect {
	r := DeskToolbarField(field)
	if r.W == 0 || rows <= 0 {
		return Rect{}
	}
	r.Y = 56
	r.H = 10 + rows*ToolbarRow
	if field == 2 {
		r.W = 330
	}
	if field == 3 {
		r.W = 380
		r.X = min(r.X, DeskWidth-r.W-12)
	}
	if field == 4 {
		r.W = 318
		r.X = DeskWidth - r.W - 12
	}
	if field == 5 {
		r.W = 360
		r.X = DeskWidth - r.W - 12
	}
	if noteLines > 0 {
		r.H += 16 + noteLines*22
	}
	return r
}
func DeskToolbarHit(x, y, menu, rows, notes int) (int, int) {
	if r := DeskToolbarMenu(menu, rows, notes); r.Contains(x, y) {
		if y >= r.Y+5 && y < r.Y+5+rows*ToolbarRow {
			return menu, (y - r.Y - 5) / ToolbarRow
		}
		return menu, -2
	}
	for i := 0; i < ToolbarFields; i++ {
		if DeskToolbarField(i).Contains(x, y) {
			return i, -1
		}
	}
	return -1, -1
}
