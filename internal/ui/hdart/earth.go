package hdart

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
	"sync"
)

// Runtime uses bounded baked art only. Generated originals remain outside game/.
//
//go:embed earth-kit.png
var earthKitBytes []byte

//go:embed earth-units.png
var earthUnitBytes []byte
var earthOnce sync.Once
var earthKit, earthUnits *image.RGBA

func loadEarth() {
	decode := func(data []byte, w, h int) *image.RGBA {
		im, err := png.Decode(bytes.NewReader(data))
		if err != nil || im.Bounds().Dx() != w || im.Bounds().Dy() != h {
			return nil
		}
		out := image.NewRGBA(image.Rect(0, 0, w, h))
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				out.Set(x, y, im.At(x, y))
			}
		}
		return out
	}
	earthKit = decode(earthKitBytes, 512, 256)
	earthUnits = decode(earthUnitBytes, 384, 192)
}
func EarthPanel(kind int) *image.RGBA {
	earthOnce.Do(loadEarth)
	if earthKit == nil || kind < 0 || kind > 3 {
		return nil
	}
	return earthKit.SubImage(image.Rect(kind*128, 128, (kind+1)*128, 256)).(*image.RGBA)
}
func EarthUnit(index, w, h int) *image.RGBA {
	earthOnce.Do(loadEarth)
	if earthUnits == nil || index < 0 || index >= 18 || w < 1 || h < 1 || w > 512 || h > 512 {
		return nil
	}
	out := image.NewRGBA(image.Rect(0, 0, w, h))
	size := min(w, h)
	ox, oy := (w-size)/2, (h-size)/2
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			out.SetRGBA(ox+x, oy+y, earthUnits.RGBAAt(index%6*64+x*64/size, index/6*64+y*64/size))
		}
	}
	return out
}
func EarthBytes() int {
	earthOnce.Do(loadEarth)
	if earthKit == nil || earthUnits == nil {
		return 0
	}
	return len(earthKit.Pix) + len(earthUnits.Pix)
}
func earthSample(kind, x, y int, p color.RGBA) color.RGBA {
	earthOnce.Do(loadEarth)
	if earthKit != nil && kind <= Olive {
		q := earthKit.RGBAAt(32+((x%64+64)%64), 160+((y%64+64)%64))
		q.A = 255
		return q
	}
	// Texture luminance stays continuous, while the HD palette contains no green.
	bases := [16][3]int{{38, 31, 26}, {37, 31, 27}, {49, 40, 31}, {88, 58, 36}, {195, 169, 125}, {180, 151, 109}, {117, 91, 62}, {145, 125, 97}, {139, 105, 66}, {85, 70, 51}, {137, 128, 116}, {179, 148, 97}, {40, 65, 89}, {82, 102, 118}, {151, 114, 76}, {193, 185, 165}}
	base := bases[max(0, min(15, kind))]
	grain := (int(p.R)+int(p.G)+int(p.B))/3 - 110
	return color.RGBA{uint8(max(0, min(255, base[0]+grain/3))), uint8(max(0, min(255, base[1]+grain/3))), uint8(max(0, min(255, base[2]+grain/3))), 255}
}
