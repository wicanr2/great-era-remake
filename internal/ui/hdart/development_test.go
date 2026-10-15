package hdart

import (
	"bytes"
	"image"
	"image/png"
	"testing"
)

func TestDevelopmentArtIsBoundedAndCached(t *testing.T) {
	if got := DevelopmentBytes(); got != 384*480*4 {
		t.Fatalf("development RGBA bytes: %d", got)
	}
	if got := Bytes() + EarthBytes() + DevelopmentBytes(); got > 5*1024*1024/2 {
		t.Fatalf("all static HD atlases exceed 2.5 MiB: %d", got)
	}
	for i := DevelopmentReclaim; i <= DevelopmentGoldMine; i++ {
		a, b := DevelopmentScene(i), DevelopmentScene(i)
		if a == nil || a.Bounds().Dx() != 384 || a.Bounds().Dy() != 160 {
			t.Fatalf("scene %d is missing or oversized", i)
		}
		if &a.Pix[0] != &b.Pix[0] {
			t.Fatal("repeated scene access allocated a second atlas")
		}
	}
	if DevelopmentScene(-1) != nil || DevelopmentScene(3) != nil {
		t.Fatal("invalid artwork index accepted")
	}
}

func TestDevelopmentDecodeRejectsCorruptOrOversizedInput(t *testing.T) {
	if decodeDevelopment([]byte("bad PNG")) != nil {
		t.Fatal("corrupt art must fall back")
	}
	var input bytes.Buffer
	if err := png.Encode(&input, image.NewRGBA(image.Rect(0, 0, 385, 480))); err != nil {
		t.Fatal(err)
	}
	if decodeDevelopment(input.Bytes()) != nil {
		t.Fatal("unexpected dimensions must be rejected before pixel decode")
	}
}
