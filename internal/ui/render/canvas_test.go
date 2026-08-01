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
