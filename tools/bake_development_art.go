//go:build ignore

// Run inside the project Docker toolchain; never ships full generated originals.
package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
)

func main() {
	if len(os.Args) != 3 {
		panic("usage: bake_development source-directory output.png")
	}
	out := image.NewRGBA(image.Rect(0, 0, 384, 480))
	for i, kind := range []string{"reclaim", "arsenal", "goldmine"} {
		file, err := os.Open(filepath.Join(os.Args[1], kind+"-source.png"))
		if err != nil {
			panic(err)
		}
		im, err := png.Decode(file)
		file.Close()
		if err != nil {
			panic(err)
		}
		// Technical reduction only, using 4x4 premultiplied box samples.
		bounds := im.Bounds()
		for y := 0; y < 160; y++ {
			for x := 0; x < 384; x++ {
				var r, g, b, a uint32
				for sy := 0; sy < 4; sy++ {
					for sx := 0; sx < 4; sx++ {
						px := bounds.Min.X + (x*8+sx*2+1)*bounds.Dx()/(384*8)
						py := bounds.Min.Y + (y*8+sy*2+1)*bounds.Dy()/(160*8)
						pr, pg, pb, pa := im.At(px, py).RGBA()
						r += pr
						g += pg
						b += pb
						a += pa
					}
				}
				out.SetRGBA(x, i*160+y, color.RGBA{uint8(r / 16 >> 8), uint8(g / 16 >> 8), uint8(b / 16 >> 8), uint8(a / 16 >> 8)})
			}
		}
	}
	f, err := os.OpenFile(os.Args[2], os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0644)
	if err != nil {
		panic(err)
	}
	encoder := png.Encoder{CompressionLevel: png.BestCompression}
	if err := encoder.Encode(f, out); err != nil {
		panic(err)
	}
	if err := f.Close(); err != nil {
		panic(err)
	}
	fmt.Println("Baked 384x480 atlas, three 384x160 scenes; decoded RGBA bytes:", len(out.Pix))
}
