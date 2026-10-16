package layout

// CommandRoom uses the existing 1280×720 logical surface on every display.
// Rendering and hit testing consume these exact rectangles.
var CommandRoom = struct{ Header, Toolbar, Map, Sidebar, Right, Footer Rect }{
	Rect{0, 0, 1280, 46}, Rect{0, 46, 782, 50}, Rect{0, 96, 1280, 624},
	Rect{4, 104, 362, 594}, Rect{0, 0, 0, 0}, Rect{374, 700, 704, 20},
}

func RoomTab(i int) Rect { return Rect{10 + i*128, 52, 122, 36} }

const RoomVisibleCommands = 8

func RoomCommand(i int) Rect { return Rect{20, 322 + i*38, 320, 35} }

var RoomArt = Rect{1168, 6, 104, 34}
var RoomArmory = Rect{792, 52, 180, 36}
var RoomInterface = Rect{1036, 6, 124, 34}
var RoomPrevious = Rect{302, 122, 24, 30}
var RoomNext = Rect{330, 122, 24, 30}
var RoomCommander = Rect{124, 178, 224, 32}
var RoomMapTools = [5]Rect{{1086, 594, 82, 26}, {1086, 624, 38, 28}, {1128, 624, 40, 28}, {1086, 656, 82, 28}, {1174, 656, 94, 28}}
var RoomMiniMap = Rect{1174, 594, 94, 58}
var RoomCommandScrollUp = Rect{344, 322, 14, 22}
var RoomCommandScrollDown = Rect{344, 603, 14, 22}
var RoomBottomActions = [3]Rect{{18, 642, 106, 42}, {130, 642, 106, 42}, {242, 642, 108, 42}}
var RoomMapInput = Rect{374, 100, 900, 486}

// The lower right navigation panel and the sidebar never start a map drag.
func RoomMapContains(x, y int) bool {
	return x >= 374 && x < 1274 && y >= 100 && y < 698 && (y < 586 || x < 1078)
}
func RoomMarkerVisible(x, y int) bool {
	return x >= 408 && x < 1238 && y >= 128 && y < 652 && (y < 558 || x < 1042)
}

var RoomArchive = Rect{294, 124, 776, 556}
var RoomSearch = Rect{308, 178, 180, 36}
var RoomArchiveBack = Rect{944, 136, 110, 32}
var RoomArchivePrevious = Rect{514, 636, 100, 30}
var RoomArchiveNext = Rect{944, 636, 110, 30}

func RoomPerson(i int) Rect { return Rect{308, 224 + i*36, 180, 32} }

var RoomArchiveListPrevious = Rect{308, 636, 84, 30}
var RoomArchiveListNext = Rect{400, 636, 88, 30}

type InterfaceMode string

const (
	InterfaceModern   InterfaceMode = "modern"
	InterfaceCommand  InterfaceMode = "command"
	InterfaceOriginal InterfaceMode = "original"
)

func ValidInterface(mode string) bool {
	return mode == "" || mode == string(InterfaceCommand) || mode == string(InterfaceModern) || mode == string(InterfaceOriginal)
}
