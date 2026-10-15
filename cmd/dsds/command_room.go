package main

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	userprefs "github.com/wicanr2/great-era-remake/internal/prefs"
	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	"github.com/wicanr2/great-era-remake/internal/ui/atlasmap"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
	"github.com/wicanr2/great-era-remake/internal/ui/textlayout"
	t "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

type commandRoomState struct {
	Mode                        l.InterfaceMode
	Map                         *atlasmap.Map
	Camera                      atlasmap.Camera
	Political                   bool
	Tab                         int
	CommandScroll               int
	Archive, Searching          bool
	Query                       string
	ListPage, BioPage, PersonID int
	Drag                        bool
	DragX, DragY                int
}

func interfaceFromPreference(s string) l.InterfaceMode {
	if s == "original" {
		return l.InterfaceOriginal
	}
	return l.InterfaceCommand
}
func (a *app) commandInterface() bool { return a != nil && a.room.Mode == l.InterfaceCommand }
func (a *app) highInterface() bool {
	if a == nil || a.room.Mode == l.InterfaceOriginal {
		return false
	}
	return a.commandInterface() || (a.resolution == "high" && a.themeMode == t.ModeModern)
}
func (a *app) roomVisible() bool {
	return a.commandInterface() && (a.screen == screenMap || a.screen == screenCommand)
}
func (a *app) setInterfacePreference(mode l.InterfaceMode) error {
	if !l.ValidInterface(string(mode)) || mode == "" {
		return fmt.Errorf("未知介面 %q", mode)
	}
	if mode == l.InterfaceCommand && (a.eten == nil || a.wording == nil) {
		return fmt.Errorf("新介面需要完整字庫與語系資料")
	}
	next := a.preferences
	next.Layout = string(mode)
	if err := userprefs.Save(a.prefsPath, next); err != nil {
		return err
	}
	a.preferences = next
	a.room.Mode = mode
	a.room.Archive = false
	a.room.Drag = false
	a.pointer.cancel()
	a.dirty = true
	a.refreshBiographyPages()
	if a.frame != nil {
		a.applyWindowSize()
	}
	return nil
}
func (a *app) toggleInterface() error {
	mode := l.InterfaceCommand
	if a.commandInterface() {
		mode = l.InterfaceOriginal
	}
	return a.setInterfacePreference(mode)
}
func (a *app) refreshBiographyPages() {
	if a.screen != screenBiography {
		return
	}
	p, ok := a.currentBiography()
	if !ok || p == nil {
		return
	}
	text := p.Biography
	if strings.TrimSpace(text) == "" {
		text = a.biographyUnavailable()
	}
	if a.highInterface() {
		if n, err := render.ModernBiographyPageCount(text); err == nil {
			a.bioPages = n
		}
	} else {
		doc, err := textlayout.Layout(text, textlayout.DefaultBiographyOptions)
		if err == nil {
			a.bioPages = len(doc.Pages)
		}
	}
	a.bioPage = max(0, min(a.bioPage, a.bioPages-1))
}
func (a *app) biographyUnavailable() string {
	v, _ := a.wording.Text("biography.unavailable", a.wordingMode)
	return v
}
func (a *app) roomText(key string) string {
	v, err := a.wordingText("room." + key)
	if err != nil {
		return key
	}
	return v
}
func (a *app) roomPeople() []*i18n.Person {
	out := []*i18n.Person{}
	if a.people == nil {
		return out
	}
	query := strings.ToLower(strings.TrimSpace(a.room.Query))
	for id := 1; id <= 417; id++ {
		if id == 274 {
			continue
		}
		p, ok := a.people.PersonByID(id)
		if !ok {
			continue
		}
		text := strings.Join([]string{p.NameInGame, p.NameCommon, strings.Join(p.Aliases, " "), p.Faction, p.Biography}, " ")
		if query == "" || strings.Contains(strings.ToLower(text), query) {
			out = append(out, p)
		}
	}
	return out
}
func (a *app) roomArchiveData() *render.RoomArchiveData {
	people := a.roomPeople()
	a.room.ListPage = max(0, min(a.room.ListPage, max(0, (len(people)-1)/11)))
	selected, _ := a.people.PersonByID(a.room.PersonID)
	if selected == nil && len(people) > 0 {
		selected = people[0]
		a.room.PersonID = selected.ID
	}
	portrait, placeholder := a.portraitForPerson(selected)
	return &render.RoomArchiveData{People: people, Selected: selected, Portrait: portrait, PortraitUnavailable: placeholder, Query: a.room.Query, ListPage: a.room.ListPage, Page: a.room.BioPage, Searching: a.room.Searching, Unavailable: a.biographyUnavailable(), Sources: a.roomText("sources"), Title: a.roomText("people"), Search: a.roomText("search")}
}
func (a *app) openRoomArchive() {
	if a.people == nil {
		return
	}
	a.room.Archive = true
	a.room.Searching = false
	if a.room.PersonID == 0 {
		a.room.PersonID = 1
	}
	for i, p := range a.roomPeople() {
		if p.ID == a.room.PersonID {
			a.room.ListPage = i / 11
			break
		}
	}
	a.dirty = true
}
func (a *app) roomMarkers() []render.RoomMarker {
	out := make([]render.RoomMarker, 0, 39)
	if a.room.Map == nil {
		return out
	}
	for i, p := range atlasmap.Anchors {
		id := game.ProvinceID(i + 1)
		province, err := a.tbl.At(id)
		if err != nil {
			continue
		}
		v := a.room.Map.Project(p, a.room.Camera, l.CommandRoom.Map.W, l.CommandRoom.Map.H)
		out = append(out, render.RoomMarker{ID: i + 1, X: int(v[0]), Y: int(v[1]) + l.CommandRoom.Map.Y, Faction: max(0, a.currentFaction(province.Commander)), Force: int(game.ForceOf(a.generals, id)), Name: a.provinceName(id), Selected: a.current == id})
	}
	return out
}
func (a *app) ensureRoomMap() error {
	if a.room.Map != nil {
		return nil
	}
	m, err := atlasmap.New()
	if err != nil {
		return err
	}
	a.room.Map = m
	a.room.Camera = atlasmap.DefaultCamera()
	a.room.Political = true
	return nil
}
func (a *app) composeCommandRoom() error {
	if err := a.ensureRoomMap(); err != nil {
		return err
	}
	p, err := a.tbl.At(a.current)
	if err != nil {
		return err
	}
	labels, err := a.panelLabels()
	if err != nil {
		return err
	}
	factions := [39]int{}
	for i := range factions {
		p, _ := a.tbl.At(game.ProvinceID(i + 1))
		factions[i] = max(0, a.currentFaction(p.Commander))
	}
	style := a.uiStyle()
	data := render.CommandRoomData{Map: a.room.Map.Render(a.room.Camera, l.CommandRoom.Map.W, l.CommandRoom.Map.H, a.themeMode == t.ModeRetro, a.room.Political, factions), Fonts: a.eten, Style: style, Title: a.roomText("title"), Date: fmt.Sprintf("%d / %02d", a.year, a.month), Interface: a.roomText("interface." + string(a.room.Mode)), Art: a.roomText("art." + string(a.themeMode)), Tab: a.room.Tab, CommandScroll: a.room.CommandScroll, Message: a.msg, Labels: map[string]string{}, Markers: a.roomMarkers()}
	data.Portrait, data.PortraitUnavailable = a.portraitForGeneral(p.Commander)
	data.Panel = render.PanelData{ID: a.current, Province: p, Force: game.ForceOf(a.generals, a.current), Generals: game.CountOf(a.generals, a.current), Commands: a.cmdBudget.Remaining(a.current), ProvinceName: a.provinceName(a.current), CommanderName: a.generalDisplayName(p.Commander), Labels: labels, Style: style}
	if a.themeMode == t.ModeRetro {
		data.OriginalUnits = a.unitTheme
	}
	for _, key := range []string{"previous_province", "next_province", "people", "narrative", "save", "end_turn", "camera_hint", "shortcuts", "close"} {
		data.Labels[key] = a.roomText(key)
	}
	for i := 0; i < 6; i++ {
		data.Tabs = append(data.Tabs, a.roomText("tab."+strconv.Itoa(i)))
	}
	for i := 1; i <= 15; i++ {
		v, err := a.wordingText(fmt.Sprintf("command.%02d", i))
		if err != nil {
			return err
		}
		data.Commands = append(data.Commands, v)
	}
	mode := "political"
	if !a.room.Political {
		mode = "terrain"
	}
	for _, key := range []string{mode, "zoom_out", "zoom_in", "pose", "reset"} {
		data.Tools = append(data.Tools, a.roomText(key))
	}
	if a.room.Archive && a.people != nil {
		data.Archive = a.roomArchiveData()
	}
	c := render.NewCanvas(1280, 720)
	if err := c.DrawCommandRoom(data); err != nil {
		return err
	}
	if a.messages != nil && a.messages.Active() {
		c.DrawCommandMessage(a.eten, style, a.messages.Current())
	}
	if a.frame != nil {
		a.frame.Dispose()
	}
	a.frame = ebiten.NewImageFromImage(c.Image())
	a.dirty = false
	return nil
}
func roomAction(key string) actions.Action { return actions.Action("room." + key) }
func (a *app) commandRoomTargets() []actions.Target {
	out := []actions.Target{}
	add := func(r l.Rect, key string) {
		out = append(out, actions.Target{Action: roomAction(key), Rect: actions.Rect{X: r.X, Y: r.Y, W: r.W, H: r.H}})
	}
	if a.room.Archive {
		add(l.RoomArchiveBack, "close")
		add(l.RoomSearch, "search")
		add(l.RoomArchivePrevious, "bio.previous")
		add(l.RoomArchiveNext, "bio.next")
		add(l.RoomArchiveListPrevious, "list.previous")
		add(l.RoomArchiveListNext, "list.next")
		people := a.roomPeople()
		for i := 0; i < 11; i++ {
			n := a.room.ListPage*11 + i
			if n < len(people) {
				add(l.RoomPerson(i), "person."+strconv.Itoa(people[n].ID))
			}
		}
		return out
	}
	add(l.RoomInterface, "interface")
	add(l.RoomArt, "art")
	add(l.RoomPrevious, "province.previous")
	add(l.RoomNext, "province.next")
	add(l.RoomCommander, "commander")
	for i := 0; i < 6; i++ {
		add(l.RoomTab(i), "tab."+strconv.Itoa(i))
	}
	for row := 0; row < l.RoomVisibleCommands; row++ {
		i := a.room.CommandScroll + row
		if i >= 15 {
			break
		}
		r := l.RoomCommand(row)
		out = append(out, actions.Target{Action: actions.Selection(i + 1), Rect: actions.Rect{X: r.X, Y: r.Y, W: r.W, H: r.H}})
	}
	add(l.RoomCommandScrollUp, "commands.up")
	add(l.RoomCommandScrollDown, "commands.down")
	for i, key := range []string{"people", "save", "end_turn"} {
		add(l.RoomBottomActions[i], key)
	}
	for i, key := range []string{"mapmode", "zoomout", "zoomin", "pose", "reset"} {
		add(l.RoomMapTools[i], key)
	}
	for _, m := range render.VisibleRoomMarkers(a.roomMarkers()) {
		if l.RoomMarkerVisible(m.X, m.Y) {
			add(l.Rect{m.X - 40, m.Y - 23, 80, 56}, "province."+strconv.Itoa(m.ID))
		}
	}
	return out
}
func (a *app) roomPage(delta int) {
	d := a.roomArchiveData()
	if d.Selected == nil {
		return
	}
	doc, err := render.RoomArchiveDocument(d)
	if err == nil {
		a.room.BioPage = max(0, min(a.room.BioPage+delta, len(doc.Pages)-1))
		a.dirty = true
	}
}
func (a *app) updateCommandRoom() bool {
	if !a.roomVisible() {
		return false
	}
	action := string(a.pointerAction)
	key := strings.TrimPrefix(action, "room.")
	if a.room.Archive {
		if key == "close" || a.actionPressed(actions.Back, ebiten.KeyEscape) {
			a.room.Archive = false
			a.room.Searching = false
			a.dirty = true
			return true
		}
		if key == "search" {
			a.room.Searching = true
			a.dirty = true
		}
		if a.room.Searching {
			chars := ebiten.AppendInputChars(nil)
			if len(chars) > 0 {
				q := append([]rune(a.room.Query), chars...)
				if len(q) > 64 {
					q = q[:64]
				}
				a.room.Query = string(q)
				a.room.ListPage = 0
				a.room.PersonID = 0
				a.room.BioPage = 0
				a.dirty = true
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) {
				q := []rune(a.room.Query)
				if len(q) > 0 {
					a.room.Query = string(q[:len(q)-1])
					a.room.ListPage = 0
					a.room.PersonID = 0
					a.room.BioPage = 0
					a.dirty = true
				}
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
				a.room.Searching = false
				a.dirty = true
			}
		}
		switch key {
		case "bio.previous":
			a.roomPage(-1)
		case "bio.next":
			a.roomPage(1)
		case "list.previous":
			a.room.ListPage = max(0, a.room.ListPage-1)
			a.dirty = true
		case "list.next":
			a.room.ListPage++
			a.dirty = true
		}
		if strings.HasPrefix(key, "person.") {
			id, _ := strconv.Atoi(strings.TrimPrefix(key, "person."))
			a.room.PersonID = id
			a.room.BioPage = 0
			a.room.Searching = false
			a.dirty = true
		}
		if !a.room.Searching {
			if inpututil.IsKeyJustPressed(ebiten.KeyPageDown) {
				a.roomPage(1)
			}
			if inpututil.IsKeyJustPressed(ebiten.KeyPageUp) {
				a.roomPage(-1)
			}
		}
		_, wheel := ebiten.Wheel()
		if _, _, ok := a.gameCursorPosition(); !ok {
			wheel = 0
		}
		if wheel != 0 {
			if wheel > 0 {
				a.roomPage(-1)
			} else {
				a.roomPage(1)
			}
		}
		return true
	}
	if strings.HasPrefix(action, "select.") {
		a.screen = screenCommand
		return false
	}
	switch key {
	case "commands.up":
		a.scrollRoomCommands(-1)
	case "commands.down":
		a.scrollRoomCommands(1)
	case "interface":
		if err := a.toggleInterface(); err != nil {
			a.report(err.Error())
		}
		return true
	case "art":
		if err := a.toggleTheme(); err != nil {
			a.report(err.Error())
		}
		return true
	case "province.previous":
		a.current--
		if !a.current.Valid() {
			a.current = 39
		}
		a.dirty = true
	case "province.next":
		a.current++
		if !a.current.Valid() {
			a.current = 1
		}
		a.dirty = true
	case "people":
		a.openRoomArchive()
	case "commander":
		if a.people == nil {
			return true
		}
		a.room.Query = ""
		p, _ := a.tbl.At(a.current)
		if person, ok := a.people.PersonAt(a.stage, int(p.Commander)); ok {
			a.room.PersonID = person.ID
			a.room.BioPage = 0
		}
		a.openRoomArchive()
	case "narrative":
		a.narrativePage = 0
		a.screen = screenNarrative
		a.dirty = true
	case "save":
		if a.origSave == nil {
			a.report(a.roomText("save_unavailable"))
		} else if err := a.autosave(); err != nil {
			a.report(err.Error())
		} else {
			a.report(a.roomText("saved"))
		}
	case "end_turn":
		a.report(a.endTurn())
	case "mapmode":
		a.room.Political = !a.room.Political
		a.dirty = true
	case "zoomout":
		a.room.Camera.ZoomAt(.85, 822, 290, l.CommandRoom.Map.W, l.CommandRoom.Map.H)
		a.dirty = true
	case "zoomin":
		a.room.Camera.ZoomAt(1.18, 822, 290, l.CommandRoom.Map.W, l.CommandRoom.Map.H)
		a.dirty = true
	case "pose":
		if a.room.Camera.Tilt > 1.35 {
			a.room.Camera.Tilt = 1.08
		} else {
			a.room.Camera.Tilt = 1.56
		}
		a.dirty = true
	case "reset":
		a.room.Camera = atlasmap.DefaultCamera()
		a.dirty = true
	}
	if strings.HasPrefix(key, "province.") && key != "province.previous" && key != "province.next" {
		id, _ := strconv.Atoi(strings.TrimPrefix(key, "province."))
		if game.ProvinceID(id).Valid() {
			a.current = game.ProvinceID(id)
			a.dirty = true
		}
	}
	if strings.HasPrefix(key, "tab.") {
		n, _ := strconv.Atoi(strings.TrimPrefix(key, "tab."))
		a.room.Tab = n
		a.dirty = true
		switch n {
		case 1:
			// Selecting a military tab must not execute an attack.
		case 2:
			a.pointerAction = actions.Select7
			a.screen = screenCommand
			return false
		case 3:
			a.pointerAction = actions.Select9
			a.screen = screenCommand
			return false
		case 4:
			a.openRoomArchive()
		case 5:
			a.screen = screenNarrative
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyB) {
		a.openRoomArchive()
		return true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyLeft) {
		a.current--
		if !a.current.Valid() {
			a.current = 39
		}
		a.dirty = true
		return true
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyRight) {
		a.current++
		if !a.current.Valid() {
			a.current = 1
		}
		a.dirty = true
		return true
	}
	// Pointer coordinates are the logical 1280×720 surface, independent of window size.
	x, y, mouseEnabled := a.gameCursorPosition()
	inside := mouseEnabled && l.RoomMapContains(x, y)
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && inside {
		a.room.Drag = true
		a.room.DragX = x
		a.room.DragY = y
	}
	if !mouseEnabled || !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		a.room.Drag = false
	}
	if a.room.Drag {
		dx, dy := x-a.room.DragX, y-a.room.DragY
		if dx != 0 || dy != 0 {
			a.room.Camera.Pan(float64(dx), float64(dy), l.CommandRoom.Map.W, l.CommandRoom.Map.H)
			a.room.DragX = x
			a.room.DragY = y
			a.dirty = true
		}
	}
	_, wheel := ebiten.Wheel()
	if !mouseEnabled {
		wheel = 0
	}
	if x >= 18 && x <= 358 && y >= 320 && y <= 628 && wheel != 0 {
		if wheel > 0 {
			a.scrollRoomCommands(-1)
		} else {
			a.scrollRoomCommands(1)
		}
	}
	if inside && wheel != 0 {
		a.room.Camera.ZoomAt(math.Exp(wheel*.12), float64(x), float64(y-l.CommandRoom.Map.Y), l.CommandRoom.Map.W, l.CommandRoom.Map.H)
		a.dirty = true
	}
	return strings.HasPrefix(action, "room.")
}

func (a *app) scrollRoomCommands(delta int) {
	a.room.CommandScroll = max(0, min(15-l.RoomVisibleCommands, a.room.CommandScroll+delta))
	a.dirty = true
}
