package render

import (
	"image"
	"image/color"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
	theme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

func lightTestPicture(w, h int) *image.RGBA {
	im := image.NewRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			im.SetRGBA(x, y, color.RGBA{byte(x % 256), byte(y % 256), byte((x + y) % 256), 255})
		}
	}
	return im
}

func TestLightDeskKeepsMapPixelsAndDarkDefault(t *testing.T) {
	d := DeskData{Map: lightTestPicture(l.DeskMap.W, l.DeskMap.H)}
	dark := DrawDesk(d)
	d.LightUI = true
	light := DrawDesk(d)
	for y := 1; y < l.DeskMap.H-1; y++ {
		for x := 1; x < l.DeskMap.W-1; x++ {
			a, b := dark.img.RGBAAt(l.DeskMap.X+x, l.DeskMap.Y+y), light.img.RGBAAt(l.DeskMap.X+x, l.DeskMap.Y+y)
			if a != b || a != d.Map.RGBAAt(x, y) {
				t.Fatalf("map pixel changed at %d,%d", x, y)
			}
		}
	}
	if dark.img.RGBAAt(190, 150) == light.img.RGBAAt(190, 150) {
		t.Fatal("light panel was not applied")
	}
	d.LightUI = false
	if diff, err := DiffCount(dark.Image(), DrawDesk(d).Image()); err != nil || diff != 0 {
		t.Fatal("light palette leaked to a later dark canvas", diff, err)
	}
}

func TestLightPortraitKeepsEveryPhotoPixel(t *testing.T) {
	photo := lightTestPicture(40, 40)
	p := &i18n.Portrait{Image: photo}
	p.Crop.Width, p.Crop.Height = 1, 1
	baseline := NewCanvas(46, 46)
	roomPortrait(baseline, p, "", l.Rect{W: 46, H: 46}, theme.UIStyle{})
	for _, light := range []bool{false, true} {
		c := NewCanvas(46, 46)
		c.UseLightUI(light)
		roomPortrait(c, p, "", l.Rect{W: 46, H: 46}, theme.UIStyle{})
		for y := 0; y < 40; y++ {
			for x := 0; x < 40; x++ {
				// The existing renderer samples pixel centers bilinearly, even
				// at 1:1. Compare displayed photo pixels with the dark baseline.
				if c.img.RGBAAt(x+3, y+3) != baseline.img.RGBAAt(x+3, y+3) {
					t.Fatalf("photo changed at %d,%d; light=%v", x, y, light)
				}
			}
		}
	}
}

func TestLightUILeavesLegacyImageDrawingUntouched(t *testing.T) {
	pal := make(assets.Palette, 16)
	for i := range pal {
		pal[i] = assets.RGB{byte(i * 10), byte(i * 8), byte(i * 6)}
	}
	im := &assets.Image{W: 4, H: 4, Pix: []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15}}
	a, b := NewCanvas(4, 4), NewCanvas(4, 4)
	b.UseLightUI(true)
	if err := a.DrawBGI(im, pal, 0, 0); err != nil {
		t.Fatal(err)
	}
	if err := b.DrawBGI(im, pal, 0, 0); err != nil {
		t.Fatal(err)
	}
	if diff, err := DiffCount(a.Image(), b.Image()); err != nil || diff != 0 {
		t.Fatal("light palette recolored asset pixels", diff, err)
	}
}
