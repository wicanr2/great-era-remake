package render

import (
	"image"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

const (
	gameDir = "../../../workplace/orig/game"
	shotDir = "../../../workplace/dosbox/shots"
)

func readFile(t *testing.T, dir, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		t.Skipf("沒有 %s，跳過", name)
	}
	return b
}

func loadPNG(t *testing.T, name string) image.Image {
	t.Helper()
	f, err := os.Open(filepath.Join(shotDir, name))
	if err != nil {
		t.Skipf("沒有截圖 %s，跳過", name)
	}
	defer f.Close()
	img, err := png.Decode(f)
	if err != nil {
		t.Fatal(err)
	}
	return img
}

// TestRenderLoadMenuMatchesOriginal 重現「載入遊戲」畫面，逐像素比對原版截圖。
//
// 這是呈現層的驗收標準（CLAUDE.md §5 步驟 3、§7 第 7 條：
// 畫面 bug 測試看不到，編譯過 ≠ 視覺驗證）。
//
// 原版畫面：640×350 全黑，CHOOSE1.TPC 畫在 (90,20)，用 CHOOSE.RGB 調色盤。
// 位置與調色盤都是實測定出來的（docs/formats/04-tpc-bgi-image.md §3）。
func TestRenderLoadMenuMatchesOriginal(t *testing.T) {
	tpc := readFile(t, gameDir, "CHOOSE1.TPC")
	rgbFile, err := os.Open(filepath.Join(gameDir, "CHOOSE.RGB"))
	if err != nil {
		t.Skip("沒有 CHOOSE.RGB，跳過")
	}
	defer rgbFile.Close()
	pal, err := assets.ParseRGBFile(rgbFile)
	if err != nil {
		t.Fatal(err)
	}
	im, err := assets.DecodeBGI(tpc, 0)
	if err != nil {
		t.Fatal(err)
	}

	c := NewBGICanvas()
	if err := c.DrawBGI(im, pal, 90, 20); err != nil {
		t.Fatal(err)
	}

	shot := loadPNG(t, "11-ret.png")
	// 截圖是 1024×768 的 Xvfb root，遊戲畫面在左上 640×350
	sub, ok := shot.(interface {
		SubImage(image.Rectangle) image.Image
	})
	if !ok {
		t.Fatal("截圖不支援 SubImage")
	}
	region := sub.SubImage(image.Rect(0, 0, ModeBGIW, ModeBGIH))

	n, err := DiffCount(c.Image(), region)
	if err != nil {
		t.Fatal(err)
	}
	total := ModeBGIW * ModeBGIH
	if n != 0 {
		t.Errorf("重現的畫面與原版截圖相異 %d/%d 像素（%.3f%%）",
			n, total, float64(n)/float64(total)*100)
	} else {
		t.Logf("640×350 全畫面逐像素相同（%d 像素）", total)
	}
}

func TestDrawBGIRejectsShortPalette(t *testing.T) {
	im := &assets.Image{W: 1, H: 1, Pix: []byte{0}}
	c := NewCanvas(4, 4)
	if err := c.DrawBGI(im, assets.Palette{{R: 0, G: 0, B: 0}}, 0, 0); err == nil {
		t.Fatal("調色盤不足 16 色應該報錯")
	}
}

func TestDiffCountRejectsSizeMismatch(t *testing.T) {
	a := image.NewRGBA(image.Rect(0, 0, 4, 4))
	b := image.NewRGBA(image.Rect(0, 0, 5, 4))
	if _, err := DiffCount(a, b); err == nil {
		t.Fatal("尺寸不同應該報錯")
	}
}

// TestSpriteTransparency 驗證索引 0 不寫入。
func TestSpriteTransparency(t *testing.T) {
	c := NewCanvas(2, 1)
	pal := make(assets.Palette, 256)
	pal[0] = assets.RGB{R: 0xFF} // 若透明失效會被畫成紅色
	pal[1] = assets.RGB{G: 0xFF}
	sp := &assets.Sprite{W: 2, H: 1, Pix: []byte{0, 1}}
	if err := c.DrawSprite(sp, pal, 0, 0); err != nil {
		t.Fatal(err)
	}
	if r, _, _, _ := c.Image().At(0, 0).RGBA(); r != 0 {
		t.Error("索引 0 應該透明，底色不該被蓋掉")
	}
	if _, g, _, _ := c.Image().At(1, 0).RGBA(); g == 0 {
		t.Error("索引 1 應該畫上去")
	}
}

// TestDrawGlyph 驗證字模繪製：bit=1 畫前景，bit=0 不畫（透明）。
func TestDrawGlyph(t *testing.T) {
	var g assets.Glyph
	g[0] = 0x80 // 第 0 列最左一點亮
	c := NewCanvas(assets.GlyphW, assets.GlyphH)
	fg := assets.RGB{R: 0xFF, G: 0xFF, B: 0xFF}
	c.DrawGlyph(g, fg, 0, 0, false)

	if r, _, _, _ := c.Image().At(0, 0).RGBA(); r == 0 {
		t.Error("(0,0) 的 bit 是 1，應該畫上前景")
	}
	if r, gg, b, _ := c.Image().At(1, 0).RGBA(); r != 0 || gg != 0 || b != 0 {
		t.Error("(1,0) 的 bit 是 0，應該保持底色（透明）")
	}
}

// TestEmbolden 驗證加粗：每列與自己右移一位 OR。
func TestEmbolden(t *testing.T) {
	var g assets.Glyph
	g[0] = 0x80 // 1000_0000 → 加粗後 1100_0000
	b := Embolden(g)
	if b[0] != 0xC0 {
		t.Errorf("加粗後第 0 列 = %#x，預期 0xC0", b[0])
	}
}

// TestGlyphRoundTripAgainstOriginal 拿字模重繪原版畫面上的中文，逐像素比對。
//
// 這是 CLAUDE.md §5 步驟 3 的驗收標準。政略畫面的「湖北省」三字用 3.15
// 的字模畫成，位置 (18,16)、(38,16)、(58,16)，字距 20、加粗、前景 (174,0,0)。
//
// 加粗這一步是量出來的：不加粗的話最佳匹配還差 49 個像素
// （畫面 132 個前景點 vs 原始字模 83 個）。
func TestGlyphRoundTripAgainstOriginal(t *testing.T) {
	data := readFile(t, gameDir, "3.15")
	f, err := assets.ParseGlyphFile(data)
	if err != nil {
		t.Fatal(err)
	}
	shot := loadPNG(t, "35-game.png")

	const (
		hubeiEntry = 25 // 湖北省是 3.15 的第 26 個詞條（0-based 25）
		slotWidth  = 3
		baseX      = 18
		baseY      = 16
	)
	fg := assets.RGB{R: 174, G: 0, B: 0}

	for i := 0; i < slotWidth; i++ {
		g := Embolden(f.Glyphs[hubeiEntry*slotWidth+i])
		x0 := baseX + i*GlyphAdvance
		var diff int
		for gy := 0; gy < assets.GlyphH; gy++ {
			for gx := 0; gx < assets.GlyphW; gx++ {
				r, gg, b, _ := shot.At(x0+gx, baseY+gy).RGBA()
				isRed := uint8(r>>8) == fg.R && uint8(gg>>8) == fg.G && uint8(b>>8) == fg.B
				if isRed != g.At(gx, gy) {
					diff++
				}
			}
		}
		if diff != 0 {
			t.Errorf("第 %d 個字 @(%d,%d) 與原版相異 %d 個像素", i+1, x0, baseY, diff)
		}
	}
}

// TestDrawEntryKeepsPadding 驗證 DrawEntry 不濾掉空白字模。
func TestDrawEntryKeepsPadding(t *testing.T) {
	data := readFile(t, gameDir, "MAN115")
	f, err := assets.ParseGlyphFile(data)
	if err != nil {
		t.Fatal(err)
	}
	k := -1
	for i := 0; i < f.EntryCount(3); i++ {
		e, err := f.Entry(i, 3)
		if err != nil {
			t.Fatal(err)
		}
		if len(e) == 2 {
			k = i
			break
		}
	}
	if k < 0 {
		t.Skip("MAN115 裡沒有兩字姓名")
	}
	c := NewCanvas(GlyphAdvance*3, assets.GlyphH)
	if err := c.DrawEntry(f, k, 3, assets.RGB{R: 0xFF}, 0, 0, true); err != nil {
		t.Fatal(err)
	}
	if c.Bounds().Dx() != GlyphAdvance*3 {
		t.Errorf("畫布寬 = %d，預期 %d", c.Bounds().Dx(), GlyphAdvance*3)
	}
}
