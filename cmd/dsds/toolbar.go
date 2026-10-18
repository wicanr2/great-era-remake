package main

import (
	"fmt"
	uiaudio "github.com/wicanr2/great-era-remake/internal/ui/audio"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	userprefs "github.com/wicanr2/great-era-remake/internal/prefs"
	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
	t "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

type toolbarPress struct {
	x, y, lastX, lastY, field, row, width int
	screen                                screen
	touch                                 ebiten.TouchID
	isTouch                               bool
}

type toolbarState struct {
	menu, hover, focus int
	press              *toolbarPress
	viewKey, frameKey  string
	view               render.ToolbarView
	notes              [l.ToolbarFields][]string
	frame              *ebiten.Image
}

func newToolbarState() *toolbarState { return &toolbarState{menu: -1, hover: -1, focus: -1} }

func (a *app) gameSurfaceSize() (int, int) {
	if a.armorySurface() {
		return 1280, 720
	}
	if a.opening != nil {
		return 1280, 720
	}
	if a.modernDeskVisible() {
		return l.DeskWidth, l.DeskHeight
	}
	if a.commandInterface() {
		return 1280, 720
	}
	return a.resolution.Size()
}

func (a *app) toolbarHeight() int {
	if a.toolbar == nil {
		return 0
	}
	if a.modernDeskVisible() {
		return l.DeskToolbarHeight
	}
	w, _ := a.gameSurfaceSize()
	return l.ToolbarHeightForWidth(w)
}

func (a *app) gameMouseEnabled() bool { return !a.preferences.KeyboardOnly }

func (a *app) gameSurfacePosition(x, y int) (int, int, bool) {
	w, h := a.gameSurfaceSize()
	y -= a.toolbarHeight()
	return x, y, x >= 0 && x < w && y >= 0 && y < h
}

func (a *app) gameCursorPosition() (int, int, bool) {
	x, y := ebiten.CursorPosition()
	x, y, ok := a.gameSurfacePosition(x, y)
	return x, y, ok && a.gameMouseEnabled()
}

func (a *app) cancelGamePointer() {
	a.desk.Press, a.desk.Slider = "", ""
	a.desk.Drag = false
	a.armoryView.Press = -1
	if a.battle != nil {
		a.battle.infoPress = ""
	}
	a.pointer.cancel()
	a.pointerAction = actions.None
	a.room.Drag = false
}

func (a *app) closeToolbar() {
	if a.toolbar != nil {
		a.toolbar.menu, a.toolbar.hover, a.toolbar.focus = -1, -1, -1
		a.toolbar.press = nil
	}
	a.cancelGamePointer()
}

func (a *app) toolbarView() render.ToolbarView {
	b := a.toolbar
	key := fmt.Sprintf("%s/%s/%t/%s/%s/%s/%s/%s", a.themeMode, a.room.Mode, a.preferences.KeyboardOnly, a.wordingMode, a.audioMode, a.preferences.MusicTrack, a.militaryStatus(), a.preferences.UIAppearance)
	if b.viewKey != key {
		word := func(key string) string {
			s, err := a.wordingText("toolbar." + key)
			if err != nil {
				return "[toolbar." + key + "]"
			}
			return s
		}
		v := render.ToolbarView{}
		v.Labels = [l.ToolbarFields]string{word("art"), word("interface"), word("ai"), word("music"), word("input"), a.militaryText("title")}
		v.Options = [l.ToolbarFields][]string{{word("art.original"), word("art.hd")}, {word("interface.original"), word("interface.command"), word("interface.light")}, {word("ai.current.long")}, {word("music.hd"), word("music.original"), word("music.off"), word("music.auto")}, {word("input.mouse"), word("input.keyboard")}, {a.militaryText("option.off"), a.militaryText("option.on")}}
		if a.militaryStatus() != "off" {
			v.Selected[5] = 1
		}
		for _, track := range musicTracks {
			v.Options[3] = append(v.Options[3], word("music."+string(track)))
		}
		if a.audioMode == uiaudio.ModeRetro {
			v.Selected[3] = 1
		}
		if a.audioMode == uiaudio.ModeOff {
			v.Selected[3] = 2
		}
		v.Checked[3] = make([]bool, len(v.Options[3]))
		v.Checked[3][v.Selected[3]] = true
		v.Checked[3][3] = a.preferences.MusicTrack == ""
		for i, track := range musicTracks {
			v.Checked[3][4+i] = a.preferences.MusicTrack == string(track)
		}
		if a.themeMode == t.ModeModern {
			v.Selected[0] = 1
		}
		if a.commandInterface() {
			v.Selected[1] = 1
			if a.lightInterface() {
				v.Selected[1] = 2
			}
		}
		if a.preferences.KeyboardOnly {
			v.Selected[4] = 1
		}
		for i := range v.Values {
			v.Values[i] = v.Options[i][v.Selected[i]]
		}
		v.Values[2] = word("ai.current")
		v.Values[5] = a.militaryText(a.militaryStatus())
		b.notes = [l.ToolbarFields][]string{}
		b.notes[2] = render.ToolbarNoteLines(word("ai.note.1")+"\n"+word("ai.note.2")+"\n"+word("ai.note.3"), 306)
		b.notes[4] = render.ToolbarNoteLines(word("input.note"), 318)
		b.notes[3] = render.ToolbarNoteLines(word("music.note"), 356)
		note := a.militaryText("note")
		if a.militaryStatus() == "locked" {
			note = a.militaryText("locked.note")
		}
		b.notes[5] = render.ToolbarNoteLines(note, 336)
		b.view, b.viewKey = v, key
	}
	v := b.view
	v.Desk = a.modernDeskVisible()
	v.Menu, v.Hover, v.Focus = b.menu, b.hover, b.focus
	if b.menu >= 0 {
		v.Notes = b.notes[b.menu]
	}
	return v
}

func (a *app) setKeyboardOnly(only bool) error {
	if a.preferences.KeyboardOnly == only {
		return nil
	}
	next := a.preferences
	next.KeyboardOnly = only
	if err := userprefs.Save(a.prefsPath, next); err != nil {
		return err
	}
	a.preferences = next
	a.cancelGamePointer()
	return nil
}

func (a *app) applyToolbarOption(field, row int) error {
	v := a.toolbarView()
	if field < 0 || field >= len(v.Options) || row < 0 || row >= len(v.Options[field]) {
		return fmt.Errorf("invalid toolbar option %d/%d", field, row)
	}
	switch field {
	case 0:
		if row == 0 {
			return a.setThemePreference(t.ModeRetro)
		}
		return a.setThemePreference(t.ModeModern)
	case 1:
		if row == 0 {
			return a.setInterfacePreference(l.InterfaceOriginal)
		}
		appearance := "dark"
		if row == 2 {
			appearance = "light"
		}
		return a.setInterfaceAppearance(l.InterfaceModern, appearance)
	case 3:
		mode, track := a.audioMode, uiaudio.Track(a.preferences.MusicTrack)
		if row < 3 {
			mode = []uiaudio.Mode{uiaudio.ModeModern, uiaudio.ModeRetro, uiaudio.ModeOff}[row]
		} else if row == 3 {
			track = ""
		} else {
			track = musicTracks[row-4]
		}
		return a.setMusic(mode, track, true)
	case 4:
		return a.setKeyboardOnly(row == 1)
	case 5:
		return a.setMilitaryEnabled(row == 1)
	}
	return nil // Current AI is informational; it never invokes a rule setter.
}

func (a *app) selectToolbarOption(field, row int) {
	if err := a.applyToolbarOption(field, row); err != nil {
		a.report(err.Error())
	}
	a.closeToolbar()
}

// updateToolbar captures a complete click, including outside-dismiss clicks.
// No pointer or keyboard event used here is dispatched to the game underneath.
func (a *app) updateToolbar() bool {
	b := a.toolbar
	if b == nil {
		return false
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyF6) {
		if b.focus < 0 {
			b.focus = 0
		} else {
			b.focus = -1
		}
		b.menu, b.hover = -1, -1
		a.cancelGamePointer()
		return true
	}
	if (b.menu >= 0 || b.focus >= 0) && inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		a.closeToolbar()
		return true
	}
	w, _ := a.gameSurfaceSize()
	x, y := ebiten.CursorPosition()
	pressed := inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	released := inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	isTouch, touch := false, ebiten.TouchID(0)
	if b.press != nil && b.press.isTouch {
		isTouch, touch = true, b.press.touch
		x, y = b.press.lastX*w/l.ToolbarWidth, b.press.lastY*w/l.ToolbarWidth
		for _, id := range ebiten.AppendTouchIDs(nil) {
			if id == touch {
				x, y = ebiten.TouchPosition(id)
				break
			}
		}
		for _, id := range inpututil.AppendJustReleasedTouchIDs(nil) {
			if id == touch {
				released = true
			}
		}
	} else if ids := inpututil.AppendJustPressedTouchIDs(nil); len(ids) > 0 {
		isTouch, touch, pressed = true, ids[0], true
		x, y = ebiten.TouchPosition(touch)
	}
	x, y = x*l.ToolbarWidth/w, y*l.ToolbarWidth/w
	v := a.toolbarView()
	field, row := v.Hit(x, y)
	insideBar := x >= 0 && x < l.ToolbarWidth && y >= 0 && y < a.toolbarHeight()*l.ToolbarWidth/w
	if b.menu >= 0 && b.focus < 0 {
		if field == b.menu && row >= 0 {
			b.hover = row
		} else {
			b.hover = -1
		}
	}
	if b.press != nil {
		p := b.press
		dx, dy := x-p.x, y-p.y
		if isTouch && !released {
			p.lastX, p.lastY = x, y
		}
		if released {
			b.press = nil
			if p.width == w && p.screen == a.screen && dx*dx+dy*dy <= 64 && field == p.field && row == p.row && field >= 0 {
				if row == -1 {
					if b.menu == field {
						b.menu = -1
					} else {
						b.menu = field
					}
					b.hover = -1
				} else if row >= 0 {
					a.selectToolbarOption(field, row)
				}
			}
		} else if !p.isTouch && !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
			b.press = nil
		}
		a.cancelGamePointer()
		return true
	}
	if pressed && (insideBar || b.menu >= 0 || b.focus >= 0 || field >= 0) {
		b.press = &toolbarPress{x: x, y: y, lastX: x, lastY: y, field: field, row: row, width: w, screen: a.screen, isTouch: isTouch, touch: touch}
		b.focus = -1
		if field < 0 {
			b.menu, b.hover = -1, -1
		}
		a.cancelGamePointer()
		return true
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) && (insideBar || b.menu >= 0 || b.focus >= 0) {
		a.closeToolbar()
		return true
	}
	if insideBar && (released || ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)) {
		a.cancelGamePointer()
		return true
	}
	if b.focus >= 0 {
		for _, delta := range []struct {
			key  ebiten.Key
			step int
		}{{ebiten.KeyLeft, -1}, {ebiten.KeyRight, 1}} {
			if inpututil.IsKeyJustPressed(delta.key) {
				b.focus = (b.focus + delta.step + l.ToolbarFields) % l.ToolbarFields
				b.menu, b.hover = -1, -1
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyDown) || inpututil.IsKeyJustPressed(ebiten.KeyUp) {
			if b.menu < 0 {
				b.menu = b.focus
				b.hover = v.Selected[b.focus]
			} else {
				delta := 1
				if inpututil.IsKeyJustPressed(ebiten.KeyUp) {
					delta = -1
				}
				b.hover = (b.hover + delta + len(v.Options[b.menu])) % len(v.Options[b.menu])
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeyKPEnter) {
			if b.menu < 0 {
				b.menu = b.focus
				b.hover = v.Selected[b.focus]
			} else {
				a.selectToolbarOption(b.menu, max(0, b.hover))
			}
		}
	}
	if b.menu >= 0 || b.focus >= 0 {
		a.cancelGamePointer()
		return true
	}
	return false
}

func (a *app) drawToolbar(dst *ebiten.Image) {
	b := a.toolbar
	if b == nil {
		return
	}
	v := a.toolbarView()
	key := fmt.Sprintf("%s/%d/%d/%d/%t", b.viewKey, b.menu, b.hover, b.focus, v.Desk)
	if b.frame == nil || b.frameKey != key {
		height := max(l.ToolbarHeight, v.MenuRect().Bottom())
		if v.Desk {
			height = max(l.DeskToolbarHeight, v.MenuRect().Bottom())
		}
		c := render.NewCanvas(l.ToolbarWidth, height)
		c.UseLightUI(a.lightInterface())
		if v.Desk {
			c.DrawDeskToolbar(a.eten, v)
		} else {
			c.DrawToolbar(a.eten, v)
		}
		if b.frame != nil {
			b.frame.Dispose()
		}
		b.frame = ebiten.NewImageFromImage(c.Image())
		b.frameKey = key
	}
	w, _ := a.gameSurfaceSize()
	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(float64(w)/l.ToolbarWidth, float64(w)/l.ToolbarWidth)
	dst.DrawImage(b.frame, op)
}
