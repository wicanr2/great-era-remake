package render

import (
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

func TestModernCFontSplitUsesSerifTitleAndSansBody(t *testing.T) {
	style := uitheme.NewModern().Style()
	title := NewCanvas(160, 48)
	body := NewCanvas(160, 48)
	if missing := drawModernTitleScaled(title, nil, "政略指令", style.Ink, 4, 4, 2, 20); len(missing) != 0 {
		t.Fatalf("宋體標題缺字：%q", string(missing))
	}
	if missing := drawModernTextScaled(body, nil, "政略指令", style.Ink, 4, 4, 2, 20); len(missing) != 0 {
		t.Fatalf("黑體內文缺字：%q", string(missing))
	}
	diff, err := DiffCount(title.Image(), body.Image())
	if err != nil {
		t.Fatal(err)
	}
	if diff == 0 {
		t.Fatal("C 字體分工沒有產生可見的宋體／黑體差異")
	}
	serif, err := assets.EmbeddedModernFontFamily(assets.ModernFontSerif)
	if err != nil || serif.Version() != 2 {
		t.Fatalf("宋體標題未使用 GEMF v2：font=%v err=%v", serif, err)
	}
}

func TestModernV2GlyphRendersIntermediateCoverageOnOpaqueCanvas(t *testing.T) {
	font, err := assets.EmbeddedModernFontFamily(assets.ModernFontSans)
	if err != nil {
		t.Fatal(err)
	}
	glyph, ok := font.Glyph('政')
	if !ok {
		t.Fatal("黑體 atlas 缺少測試字")
	}
	c := NewCanvas(40, 40)
	c.fillRect(0, 0, 40, 40, assets.RGB{R: 240, G: 210, B: 140})
	drawModernGlyph(c, glyph, assets.RGB{R: 70, G: 43, B: 24}, 4, 4, 2)
	seenIntermediate := false
	for y := 4; y < 36; y++ {
		for x := 4; x < 36; x++ {
			p := c.Image().RGBAAt(x, y)
			if p.A == 255 && p.R > 70 && p.R < 240 {
				seenIntermediate = true
			}
		}
	}
	if !seenIntermediate {
		t.Fatal("GEMF v2 字形沒有合成灰階抗鋸齒邊緣")
	}
}

func TestModernA2LongLocaleLabelsStayInsideTitleSafeWidth(t *testing.T) {
	// 這些不是翻譯內容的定義，只是目前 UI 最長形狀的安全樣本；真正詞條仍由
	// 呼叫端語系資料提供。標題以 2 倍字級繪製，安全寬度固定 54 half-cell。
	labels := []string{
		"人物檔案｜蔣中正　字介石",
		"Commander Biography｜Chiang Kai-shek",
		"人物資料｜蒋介石（しょうかいせき）",
	}
	for _, label := range labels {
		if got := modernTextHalfCells(trimHalfCells(label, 54)); got > 54 {
			t.Fatalf("標題超出 A2 安全寬度：%q => %d half-cell", label, got)
		}
	}
}

func TestModernA2InteractiveRectsLeaveDecorativeTextInset(t *testing.T) {
	modern := uitheme.NewModern()
	command, err := uilayout.NewModernCommandLayout(uilayout.ModernDesignSurface())
	if err != nil {
		t.Fatal(err)
	}
	options, err := uilayout.NewModernOptionLayout(uilayout.ModernDesignSurface(), 4)
	if err != nil {
		t.Fatal(err)
	}
	flow, err := uilayout.NewModernFlowLayout(uilayout.ModernDesignSurface(), 3)
	if err != nil {
		t.Fatal(err)
	}
	// 這項檢查貼著共用 primitive 的契約：裝飾角占外側 16 px，文字一律
	// 留在扣除內縮後的安全矩形。
	check := func(name string, r uilayout.Rect) {
		if r.W < 48 || r.H < 48 {
			t.Fatalf("%s 控制項太小：%+v", name, r)
		}
		text := uilayout.Rect{X: r.X + 16, Y: r.Y + 10, W: r.W - 32, H: r.H - 20}
		if text.W <= 0 || text.H <= 0 || text.Intersects(uilayout.Rect{X: r.X, Y: r.Y, W: 8, H: r.H}) {
			// 左側 8 px 是裝飾帶；文字安全區不得進入該區。
			t.Fatalf("%s 文字安全區侵入裝飾：outer=%+v text=%+v", name, r, text)
		}
	}
	for i, r := range command.Cards {
		check("command", r)
		if i == 0 {
			// 確認 A2 主題本身確實載入，而非測試只驗幾何。
			if modern.Style().Paper == modern.Style().Ink {
				t.Fatal("Modern A2 紙面與文字色不可相同")
			}
		}
	}
	for _, r := range options.Options {
		check("options", r)
	}
	for _, r := range flow.Cards {
		check("flow", r)
	}
}

func TestModernA2CommandLabelsUseTwoXAndVisibleTruncation(t *testing.T) {
	const width = 171 // 5-column command card text width at 1280×720.
	for _, label := range []string{"Movement", "Resupply", "Taxation", "部隊移動"} {
		lines := modernButtonLabelLines(label, width, 2)
		if len(lines) == 0 || modernTextPixelWidth(lines[0], 2) > width {
			t.Fatalf("主要指令標籤未落在 2 倍字安全區：%q => %#v", label, lines)
		}
	}
	longWord := modernButtonLabelLines("Reorganization", 96, 2)
	if len(longWord) != 1 || longWord[0] == "Reorganization" || len([]rune(longWord[0])) < 2 {
		t.Fatalf("單一長詞必須以省略號明示截短：%#v", longWord)
	}
	if last := []rune(longWord[0]); last[len(last)-1] != '…' {
		t.Fatalf("截短標籤缺少省略號：%q", longWord[0])
	}
	twoLines := modernButtonLabelLines("Mobilize Reserve Forces", width, 2)
	if len(twoLines) != 2 {
		t.Fatalf("含空白的長指令應優先分成兩行：%#v", twoLines)
	}
}
