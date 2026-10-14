package main

import (
	"strings"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
)

func TestRoomScrollKeepsAllCommandsReachable(t *testing.T) {
	a := &app{}
	seen := map[actions.Action]bool{}
	for scroll := 0; scroll < 15; scroll++ {
		count := 0
		for _, target := range a.commandRoomTargets() {
			if !strings.HasPrefix(string(target.Action), "select.") {
				continue
			}
			seen[target.Action] = true
			count++
			r := target.Rect
			if r.Y < 322 || r.Y+r.H > 626 || r.X+r.W > 344 {
				t.Fatalf("command outside scroll viewport: %+v", r)
			}
		}
		if count != l.RoomVisibleCommands {
			t.Fatalf("visible commands: %d", count)
		}
		a.scrollRoomCommands(1)
	}
	for i := 1; i <= 15; i++ {
		if !seen[actions.Selection(i)] {
			t.Fatalf("command %d is unreachable", i)
		}
	}
	a.scrollRoomCommands(-100)
	if a.room.CommandScroll != 0 {
		t.Fatal("scroll underflow")
	}
	if l.RoomMapContains(180, 453) || l.RoomMapContains(1105, 638) {
		t.Fatal("UI controls start a map drag")
	}
}
