package hdart

import (
	"bytes"
	_ "embed"
	"image"
	"image/draw"
	"image/jpeg"
	"sync"
)

//go:embed declaration.jpg
var declarationBytes []byte

var declarationOnce sync.Once
var declarationScene *image.RGBA

// DeclarationScene keeps the corrected project illustration below 1.6 MiB
// decoded. Source, conversion and rights are recorded in generated-hd-assets.
func DeclarationScene() *image.RGBA {
	declarationOnce.Do(func() {
		config, err := jpeg.DecodeConfig(bytes.NewReader(declarationBytes))
		if err != nil || config.Width != 768 || config.Height != 512 {
			return
		}
		im, err := jpeg.Decode(bytes.NewReader(declarationBytes))
		if err != nil {
			return
		}
		declarationScene = image.NewRGBA(image.Rect(0, 0, 768, 512))
		draw.Draw(declarationScene, declarationScene.Bounds(), im, im.Bounds().Min, draw.Src)
	})
	return declarationScene
}
