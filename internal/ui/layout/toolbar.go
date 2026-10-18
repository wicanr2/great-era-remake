package layout

const (
	ToolbarWidth  = 1280
	ToolbarHeight = 48
	ToolbarRow    = 36
	ToolbarFields = 6
)

var toolbarFields = [ToolbarFields]Rect{
	{X: 10, Y: 6, W: 205, H: 34}, {X: 223, Y: 6, W: 221, H: 34},
	{X: 452, Y: 6, W: 144, H: 34}, {X: 604, Y: 6, W: 180, H: 34},
	{X: 792, Y: 6, W: 236, H: 34}, {X: 1036, Y: 6, W: 234, H: 34},
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
	if field == 4 {
		r.W = 318
	}
	if field == 5 {
		r.W = 360
		r.X = ToolbarWidth - r.W - 10
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
