package layout

const (
	ToolbarWidth  = 1280
	ToolbarHeight = 48
	ToolbarRow    = 36
	ToolbarFields = 5
)

var toolbarFields = [ToolbarFields]Rect{
	{X: 10, Y: 6, W: 298, H: 34}, {X: 316, Y: 6, W: 298, H: 34},
	{X: 622, Y: 6, W: 142, H: 34}, {X: 772, Y: 6, W: 148, H: 34},
	{X: 928, Y: 6, W: 342, H: 34},
}

func ToolbarField(field int) Rect {
	if field < 0 || field >= ToolbarFields {
		return Rect{}
	}
	return toolbarFields[field]
}

func ToolbarHeightForWidth(width int) int { return width * ToolbarHeight / ToolbarWidth }

// ToolbarMenu includes the note; ToolbarHit never turns note pixels into options.
func ToolbarMenu(field, rows, noteLines int) Rect {
	r := ToolbarField(field)
	if r.W == 0 || rows <= 0 {
		return Rect{}
	}
	r.Y, r.H = 46, 10+rows*ToolbarRow
	if field == 2 {
		r.W = 330
	}
	if field == 3 {
		r.W = 380
	}
	if noteLines > 0 {
		r.H += 16 + noteLines*22
	}
	return r
}

// row=-1 is a field, row=-2 the inert menu background or help text.
func ToolbarHit(x, y, menu, rows, noteLines int) (field, row int) {
	if r := ToolbarMenu(menu, rows, noteLines); r.Contains(x, y) {
		if y >= r.Y+5 && y < r.Y+5+rows*ToolbarRow {
			return menu, (y - r.Y - 5) / ToolbarRow
		}
		return menu, -2
	}
	for i, r := range toolbarFields {
		if r.Contains(x, y) {
			return i, -1
		}
	}
	return -1, -1
}

// ToolbarToGame retains the complete game surface below the permanent strip.
func ToolbarToGame(x, y, width, height int, enabled bool) (int, int, bool) {
	if enabled {
		y -= ToolbarHeightForWidth(width)
	}
	return x, y, x >= 0 && x < width && y >= 0 && y < height
}
