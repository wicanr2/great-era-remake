package render

import (
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

func TestDrawDisplayOptionsHasBothChoicesAndNoMissingGlyphs(t *testing.T) {
	fonts, err := assets.LoadEtenFonts(filepath.Join("..", "..", "..", "workplace", "eten"))
	if err != nil {
		t.Skipf("沒有使用者倚天字庫：%v", err)
	}
	c := NewBGICanvas()
	missing := c.DrawDisplayOptions(fonts, assets.RGB{R: 174},
		assets.RGB{R: 255, G: 255, B: 162}, 190, 0, 450, 350,
		"顯示設定", "顯示用語", "原典用語", "現代白話", 1)
	if len(missing) != 0 {
		t.Fatalf("設定頁缺字：%q", string(missing))
	}
	// selected=1 的框左上角固定在第二列。
	if got := c.Image().RGBAAt(218, 156); got.R != 174 {
		t.Fatalf("現代白話選取框未畫出：%v", got)
	}
}

func TestDrawOtherOptionsHasNineEntriesAndNoMissingGlyphs(t *testing.T) {
	fonts, err := assets.LoadEtenFonts(filepath.Join("..", "..", "..", "workplace", "eten"))
	if err != nil {
		t.Skipf("沒有使用者倚天字庫：%v", err)
	}
	c := NewBGICanvas()
	labels := []string{"儲存進度", "載入進度", "音效", "指令插圖", "背景音樂",
		"訊息顯示時間", "觀看電腦戰鬥", "儲存並離開", "顯示設定"}
	available := []bool{false, false, false, false, false, false, false, true, true}
	missing := c.DrawOtherOptions(fonts, assets.RGB{R: 174},
		assets.RGB{R: 255, G: 255, B: 162}, 190, 0, 450, 350,
		labels, available, "尚未完成")
	if len(missing) != 0 {
		t.Fatalf("其他選項缺字：%q", string(missing))
	}
	// 第 1 與第 9 項分處兩欄；兩處編號區都必須有前景像素。
	for _, p := range [][2]int{{214, 30}, {429, 195}} {
		found := false
		for y := p[1]; y < p[1]+18; y++ {
			for x := p[0]; x < p[0]+30; x++ {
				if c.Image().RGBAAt(x, y).R == 174 {
					found = true
				}
			}
		}
		if !found {
			t.Fatalf("其他選項編號區 %v 沒有前景像素", p)
		}
	}
}
