package hdart

import (
	"bytes"
	_ "embed"
	"image"
	"image/draw"
	"image/png"
	"sync"
)

//go:embed development.png
var developmentBytes []byte

type DevelopmentIllustration int

const (
	DevelopmentReclaim DevelopmentIllustration = iota
	DevelopmentArsenal
	DevelopmentGoldMine
)

var developmentOnce sync.Once
var developmentAtlas *image.RGBA

func decodeDevelopment(data []byte) *image.RGBA {
	config, err := png.DecodeConfig(bytes.NewReader(data))
	if err != nil || config.Width != 384 || config.Height != 480 {
		return nil
	}
	im, err := png.Decode(bytes.NewReader(data))
	if err != nil {
		return nil
	}
	out := image.NewRGBA(image.Rect(0, 0, 384, 480))
	draw.Draw(out, out.Bounds(), im, im.Bounds().Min, draw.Src)
	return out
}

func DevelopmentScene(kind DevelopmentIllustration) *image.RGBA {
	if kind < DevelopmentReclaim || kind > DevelopmentGoldMine {
		return nil
	}
	developmentOnce.Do(func() { developmentAtlas = decodeDevelopment(developmentBytes) })
	if developmentAtlas == nil {
		return nil
	}
	y := int(kind) * 160
	return developmentAtlas.SubImage(image.Rect(0, y, 384, y+160)).(*image.RGBA)
}

func DevelopmentBytes() int {
	developmentOnce.Do(func() { developmentAtlas = decodeDevelopment(developmentBytes) })
	if developmentAtlas == nil {
		return 0
	}
	return len(developmentAtlas.Pix)
}
