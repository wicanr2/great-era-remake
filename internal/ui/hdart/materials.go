// Package hdart shares bounded, newly generated materials between the map and UI.
package hdart

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/png"
	"sync"
)

//go:embed materials.png
var source []byte

const TileSize = 128
const (
	Steel = iota
	Gunmetal
	Olive
	Leather
	Paper
	AgedPaper
	Canvas
	Limestone
	Plain
	Forest
	Mountain
	Desert
	Ocean
	Coastal
	Grass
	Snow
)

var once sync.Once
var atlas *image.RGBA

func load() {
	im, err := png.Decode(bytes.NewReader(source))
	if err != nil || im.Bounds().Dx() != 512 || im.Bounds().Dy() != 512 {
		return
	}
	atlas = image.NewRGBA(image.Rect(0, 0, 512, 512))
	for y := 0; y < 512; y++ {
		for x := 0; x < 512; x++ {
			atlas.Set(x, y, im.At(x, y))
		}
	}
}

// Sample tiles within their own bounds; the generated original is never loaded.
func Sample(kind, x, y int) color.RGBA {
	once.Do(load)
	if atlas == nil {
		return color.RGBA{40, 42, 39, 255}
	}
	kind = max(0, min(15, kind))
	x = ((x % TileSize) + TileSize) % TileSize
	y = ((y % TileSize) + TileSize) % TileSize
	i := ((kind/4)*TileSize+y)*atlas.Stride + ((kind%4)*TileSize+x)*4
	return earthSample(kind, x, y, color.RGBA{atlas.Pix[i], atlas.Pix[i+1], atlas.Pix[i+2], 255})
}

func Bytes() int {
	once.Do(load)
	if atlas == nil {
		return 0
	}
	return len(atlas.Pix)
}
