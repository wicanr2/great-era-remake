package render

import (
	"image"
	"image/color"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

func TestCanvasBlendPixelCoverage(t *testing.T) {
	c := NewCanvas(1, 1)
	c.setPixel(0, 0, assets.RGB{R: 20, G: 40, B: 60})
	c.blendPixel(0, 0, assets.RGB{R: 220, G: 140, B: 60}, 0)
	if got := c.Image().RGBAAt(0, 0); got != (color.RGBA{R: 20, G: 40, B: 60, A: 255}) {
		t.Fatalf("coverage=0 不得改寫背景：%+v", got)
	}
	c.blendPixel(0, 0, assets.RGB{R: 220, G: 140, B: 60}, 128)
	if got := c.Image().RGBAAt(0, 0); got != (color.RGBA{R: 120, G: 90, B: 60, A: 255}) {
		t.Fatalf("coverage=128 混色錯誤：%+v", got)
	}
	c.blendPixel(0, 0, assets.RGB{R: 1, G: 2, B: 3}, 255)
	if got := c.Image().RGBAAt(0, 0); got != (color.RGBA{R: 1, G: 2, B: 3, A: 255}) {
		t.Fatalf("coverage=255 應完全覆蓋：%+v", got)
	}
}

func TestDrawModernPortraitCardRendersVerifiedImageAndFallback(t *testing.T) {
	style := uitheme.NewModern().Style()
	card := uilayout.Rect{X: 800, Y: 120, W: 232, H: 340}
	portrait := &i18n.Portrait{
		PersonID: 1,
		Image:    image.NewUniform(color.RGBA{R: 0x31, G: 0x62, B: 0x93, A: 0xFF}),
		Asset:    "portraits/test.jpg",
		Crop:     i18n.PortraitCrop{Width: 1, Height: 1},
	}
	withImage := NewCanvas(1280, 720)
	missing := map[rune]bool{}
	if err := withImage.drawModernPortraitCard(card, portrait, "", nil, style, missing); err != nil {
		t.Fatal(err)
	}
	imageY := card.Y + 14 + ((card.H-68)-(card.W-28))/2 + 20
	if got := withImage.Image().RGBAAt(card.X+20, imageY); got.R != 0x31 || got.G != 0x62 || got.B != 0x93 {
		t.Fatalf("肖像像素沒有進入固定欄位：%+v", got)
	}

	fallback := NewCanvas(1280, 720)
	if err := fallback.drawModernPortraitCard(card, nil, "", nil, style, map[rune]bool{}); err != nil {
		t.Fatal(err)
	}
	if got := fallback.Image().RGBAAt(card.X+20, card.Y+20); got.R == 0x31 && got.G == 0x62 && got.B == 0x93 {
		t.Fatalf("缺登錄時不得保留或猜測肖像：%+v", got)
	}
}

func TestPortraitSamplingUsesAreaReductionAndBilinearEnlargement(t *testing.T) {
	src := image.NewRGBA(image.Rect(0, 0, 2, 2))
	src.SetRGBA(0, 0, color.RGBA{R: 255, A: 255})
	src.SetRGBA(1, 0, color.RGBA{G: 255, A: 255})
	src.SetRGBA(0, 1, color.RGBA{B: 255, A: 255})
	src.SetRGBA(1, 1, color.RGBA{R: 255, G: 255, B: 255, A: 255})
	down := image.NewRGBA(image.Rect(0, 0, 1, 1))
	drawSampledImage(down, uilayout.Rect{W: 1, H: 1}, src, src.Bounds())
	if got := down.RGBAAt(0, 0); got.R < 120 || got.G < 120 || got.B < 120 {
		t.Fatalf("縮小沒有做面積平均：%+v", got)
	}
	up := image.NewRGBA(image.Rect(0, 0, 4, 4))
	drawSampledImage(up, uilayout.Rect{W: 4, H: 4}, src, src.Bounds())
	got := up.RGBAAt(1, 1)
	if got.R == 0 || got.G == 0 || got.B == 0 || got.A != 255 {
		t.Fatalf("放大沒有做雙線性混色：%+v", got)
	}
}

func TestModernHUDIconFilteringProducesIntermediateEdge(t *testing.T) {
	modern := uitheme.NewModern()
	icon, err := modern.CommandIcon(0)
	if err != nil {
		t.Fatal(err)
	}
	c := NewCanvas(64, 64)
	if err := drawFilteredIndexedIcon(c, icon, uilayout.Rect{X: 4, Y: 4, W: 44, H: 44}); err != nil {
		t.Fatal(err)
	}
	seenIntermediate := false
	background := color.RGBA{A: 0xff}
	for y := 4; y < 48; y++ {
		for x := 4; x < 48; x++ {
			p := c.Image().RGBAAt(x, y)
			// Canvas 維持不透明；抗鋸齒證據是邊緣已和黑底合成出
			// 中間色，而不是把半透明像素留給後段不確定地處理。
			if p.A == 255 && p != background && p.R != 255 && p.G != 255 && p.B != 255 {
				seenIntermediate = true
			}
		}
	}
	if !seenIntermediate {
		t.Fatal("指令圖示仍只有最近鄰硬邊，沒有 RGBA 邊緣混色")
	}
}

func TestDrawModernPortraitCardPreservesPortraitAspectRatio(t *testing.T) {
	style := uitheme.NewModern().Style()
	card := uilayout.Rect{X: 800, Y: 120, W: 108, H: 148}
	source := image.NewRGBA(image.Rect(0, 0, 2, 4))
	for y := 0; y < 4; y++ {
		fill := color.RGBA{R: 0xD0, A: 0xFF}
		if y >= 2 {
			fill = color.RGBA{B: 0xD0, A: 0xFF}
		}
		for x := 0; x < 2; x++ {
			source.SetRGBA(x, y, fill)
		}
	}
	portrait := &i18n.Portrait{PersonID: 1, Image: source, Asset: "portraits/tall.jpg",
		Crop: i18n.PortraitCrop{Width: 1, Height: 1}}
	c := NewCanvas(1280, 720)
	if err := c.drawModernPortraitCard(card, portrait, "", nil, style, map[rune]bool{}); err != nil {
		t.Fatal(err)
	}
	// 2:4 來源必須依比例置中在約 80×80 的內框；中央欄仍應保持
	// 上紅下藍的次序，證明沒有被強制拉成正方形。
	if got := c.Image().RGBAAt(card.X+card.W/2, card.Y+18); got.R != 0xD0 || got.B != 0 {
		t.Fatalf("肖像上半部色帶不在固定槽：%+v", got)
	}
	if got := c.Image().RGBAAt(card.X+card.W/2, card.Bottom()-58); got.B != 0xD0 || got.R != 0 {
		t.Fatalf("肖像下半部色帶不在固定槽：%+v", got)
	}
}
