package render

import (
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

func TestDrawPolicyAndAutonomyHaveNoMissingGlyphs(t *testing.T) {
	fonts := loadBiographyTestFonts(t)
	fg, bg := assets.RGB{R: 174}, assets.RGB{R: 255, G: 255, B: 162}
	c := NewBGICanvas()
	missing := c.DrawPolicyMenu(fonts, fg, bg, 190, 0, 450, 350,
		"政策", "授權自治", "產能分配", "產能分配尚未開放")
	if len(missing) != 0 {
		t.Fatalf("政策頁缺字：%q", string(missing))
	}

	c = NewBGICanvas()
	missing = c.DrawAutonomyPolicy(fonts, fg, bg, 190, 0, 450, 350,
		"授權自治", "正常", "自治", "輸入省份編號；ESC 完成",
		[]AutonomyRow{{Province: 19, Name: "河南省"},
			{Province: 26, Name: "湖北省", Autonomous: true}}, 19)
	if len(missing) != 0 {
		t.Fatalf("自治頁缺字：%q", string(missing))
	}
	drawn := false
	for y := 60; y < 75; y++ {
		for x := 200; x < 234; x++ {
			got := c.Image().RGBAAt(x, y)
			if got.R != bg.R || got.G != bg.G || got.B != bg.B {
				drawn = true
			}
		}
	}
	if !drawn {
		t.Fatal("省份編號未畫在導覽保留區下方")
	}
	for y := 45; y < 58; y++ {
		for x := 200; x < 234; x++ {
			got := c.Image().RGBAAt(x, y)
			if got.R != bg.R || got.G != bg.G || got.B != bg.B {
				t.Fatal("舊列位置仍有像素，會與可見導覽區重疊")
			}
		}
	}
}

func TestSelectedProductionMovesPromptAboveKeypad(t *testing.T) {
	fonts := loadBiographyTestFonts(t)
	fg, bg := assets.RGB{R: 174}, assets.RGB{R: 255, G: 255, B: 162}
	c := NewBGICanvas()
	missing := c.DrawProductionPolicy(fonts, fg, bg, 190, 0, 450, 350,
		"產能分配", [5]string{"黃金", "鐵礦", "煤礦", "石油", "糧食"},
		[5]uint8{20, 20, 20, 20, 20}, 1, "輸入新比例", 25)
	if len(missing) != 0 {
		t.Fatalf("產能頁缺字：%q", string(missing))
	}
	// 選取框應包住第一個可調欄位，輸入值則在上方導覽保留列。
	if got := c.Image().RGBAAt(208, 100); got.R != fg.R {
		t.Fatalf("鐵礦選取框左上=%v", got)
	}
	drawn := false
	for y := 42; y < 58; y++ {
		for x := 480; x < 520; x++ {
			got := c.Image().RGBAAt(x, y)
			if got.R == fg.R && got.G == fg.G && got.B == fg.B {
				drawn = true
			}
		}
	}
	if !drawn {
		t.Fatal("選定後的新比例輸入值未移到鍵盤上方")
	}
}
