package theme

import (
	"bytes"
	"testing"
)

func TestCommandArtworkCompleteAndBounded(t *testing.T) {
	m := NewCommandModern()
	for i := 0; i < 22; i++ {
		b, e := m.Tile(i)
		if e != nil || !b.Valid() {
			t.Fatal(i, e)
		}
		im, e := m.HighTile(i, 96, 72)
		if e != nil || im.Bounds().Dx() != 96 {
			t.Fatal(i, e)
		}
	}
	for i := 0; i < 21; i++ {
		b, e := m.Rail(i)
		if e != nil || !b.Valid() {
			t.Fatal(i, e)
		}
	}
	faces := [][]byte{}
	for i := 0; i < 18; i++ {
		b, e := m.Unit(i)
		if e != nil || !b.Valid() {
			t.Fatal(i, e)
		}
		im, e := m.HighUnit(i, 96, 51)
		if e != nil {
			t.Fatal(i, e)
		}
		if im.RGBAAt(0, 0).A != 0 {
			t.Fatal("counter transparency")
		}
		if i >= 6 && i < 12 {
			faces = append(faces, im.Pix)
		}
	}
	for i := range faces {
		for j := 0; j < i; j++ {
			if bytes.Equal(faces[i], faces[j]) {
				t.Fatal("lost artillery facing")
			}
		}
	}
	for i := 0; i < 6; i++ {
		if _, e := m.ResourceIconRGBA(i, 24, 24); e != nil {
			t.Fatal(e)
		}
	}
	for i := 0; i < 15; i++ {
		if _, e := m.CommandIconRGBA(i, 24, 24); e != nil {
			t.Fatal(e)
		}
	}
	if _, e := m.HighUnit(18, 96, 51); e == nil {
		t.Fatal("invalid index accepted")
	}
	if _, e := m.HighTile(0, 8192, 8192); e == nil {
		t.Fatal("oversized asset accepted")
	}
	if m.cacheBytes > 16*1024*1024 {
		t.Fatal("unbounded artwork cache")
	}
}
