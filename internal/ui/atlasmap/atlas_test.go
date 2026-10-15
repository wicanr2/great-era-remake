package atlasmap

import (
	"math"
	"testing"
)

// These seams were visible through otherwise continuous province polygons.
// Check the actual unpainted surface so faction borders cannot hide a step.
func TestNationalSurfaceHasNoRectangularPaletteSeams(t *testing.T) {
	m, err := New()
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		a, b Point
	}{
		{"western longitude", Point{100.88, 46}, Point{101.12, 46}},
		{"southern latitude", Point{113, 31.88}, Point{113, 32.12}},
		{"northeastern longitude", Point{123.88, 46}, Point{124.12, 46}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			a := m.surfaceAt(tc.a[0], tc.a[1], false, &[39]int{})
			b := m.surfaceAt(tc.b[0], tc.b[1], false, &[39]int{})
			for channel := range a {
				if step := math.Abs(a[channel] - b[channel]); step > 6 {
					t.Fatalf("geographic palette seam: channel %d changes %.2f across adjacent land", channel, step)
				}
			}
		})
	}
}

func TestLiveCameraAndRenderBudget(t *testing.T) {
	m, err := New()
	if err != nil {
		t.Fatal(err)
	}
	if len(m.vertices) != 18193 || len(m.indices)/3 != 35840 {
		t.Fatal("mesh budget changed")
	}
	c := DefaultCamera()
	p := m.Project(Anchors[18], c, 1280, 580)
	if id := m.Hit(p[0], p[1], c, 1280, 580); id != 19 {
		t.Fatalf("province hit %d", id)
	}
	c.ZoomAt(1.4, p[0], p[1], 1280, 580)
	q := m.Project(Anchors[18], c, 1280, 580)
	if math.Hypot(p[0]-q[0], p[1]-q[1]) > .001 {
		t.Fatal("zoom lost cursor anchor")
	}
	c.ZoomAt(100, 640, 290, 1280, 580)
	if c.Zoom != 3.5 {
		t.Fatal("unbounded zoom")
	}
	c.ZoomAt(.0001, 640, 290, 1280, 580)
	if c.Zoom != .65 {
		t.Fatal("minimum zoom")
	}
	c = DefaultCamera()
	f := [39]int{}
	a := m.Render(c, 1280, 580, false, true, f)
	b := m.Render(c, 1280, 580, false, true, f)
	if a != b || m.Renders != 1 || len(a.Pix) > 1280*720*4 {
		t.Fatal("idle redraw or oversized texture")
	}
	f[18] = 2
	m.Render(c, 1280, 580, false, true, f)
	if m.Renders != 2 {
		t.Fatal("live faction change not rendered")
	}
	c.Pan(40, 25, 1280, 580)
	m.Render(c, 1280, 580, false, true, f)
	if m.Renders != 3 {
		t.Fatal("camera movement not rendered")
	}
}
