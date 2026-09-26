package render

import (
	"image"
	"image/color"
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

func modernHighFixture() (game.Battlefield, PanelData) {
	var bf game.Battlefield
	for y := range bf.Tiles {
		for x := range bf.Tiles[y] {
			bf.Tiles[y][x] = assets.Tile{Kind: assets.TilePlain, Rail: assets.NoRail}
		}
	}
	bf.Tiles[0][0].Rail = 0
	bf.Owner[1][1] = 2
	p := &game.Province{Gold: 12, Food: 34, Ammo: 56, Fuel: 78, Coal: 90, Iron: 12,
		Population: 123456, Cities: 2, LandValue: 3, Arsenals: 4, Loyalty: 66}
	return bf, PanelData{ID: 26, Province: p, Force: 9876, Generals: 3}
}

func TestDrawModernMapSurfaceUsesHighCanvasAndSafeThemeAssets(t *testing.T) {
	bf, panel := modernHighFixture()
	modern := uitheme.NewModern()
	c := NewCanvas(1280, 720)
	if err := c.DrawModernMapSurface(ModernMapSurfaceData{
		Battlefield: &bf,
		Theme:       modern,
		Style:       modern.Style(),
		Panel:       panel,
	}); err != nil {
		t.Fatal(err)
	}
	if got := c.Image().At(0, 0); got == (color.RGBA{}) {
		t.Fatal("高解析畫布沒有背景像素")
	}
	if got := c.Image().At(384, 96); got == (color.RGBA{}) {
		t.Fatal("高解析地圖卡沒有繪製")
	}
}

func TestDrawModernMapSurfaceUsesSatelliteRailAndEventStrip(t *testing.T) {
	bf, panel := modernHighFixture()
	modern := uitheme.NewModern()
	style := modern.Style()
	l, err := uilayout.NewModernMapLayout(uilayout.ModernDesignSurface())
	if err != nil {
		t.Fatal(err)
	}
	c := NewCanvas(1280, 720)
	if err := c.DrawModernMapSurface(ModernMapSurfaceData{
		Battlefield: &bf, Theme: modern, Style: style, Panel: panel, Command: "指令", Narrative: "新聞",
	}); err != nil {
		t.Fatal(err)
	}
	if got := c.Image().RGBAAt(l.Rail.Right()-2, l.Rail.Y+300); got.R != style.AccentAlt.R || got.G != style.AccentAlt.G || got.B != style.AccentAlt.B {
		t.Fatalf("H4 導覽軌沒有青藍連線：got=%+v style=%+v", got, style.AccentAlt)
	}
	if got := c.Image().RGBAAt(l.EventStrip.X+1, l.EventStrip.Y+100); got.R != style.Accent.R || got.G != style.Accent.G || got.B != style.Accent.B {
		t.Fatalf("H4 操作／事件列沒有主指令色帶：got=%+v style=%+v", got, style.Accent)
	}
}

func TestDrawModernMapSurfaceRendersVerifiedPortraitSlot(t *testing.T) {
	bf, panel := modernHighFixture()
	panel.ProvinceName = "廣東"
	panel.CommanderName = "蔣中正"
	modern := uitheme.NewModern()
	c := NewCanvas(1280, 720)
	fonts := loadBiographyTestFonts(t)
	portrait := &i18n.Portrait{PersonID: 1,
		Image: image.NewUniform(color.RGBA{R: 0x31, G: 0x62, B: 0x93, A: 0xff}),
		Asset: "portraits/test.jpg", Crop: i18n.PortraitCrop{Width: 1, Height: 1}}
	if err := c.DrawModernMapSurface(ModernMapSurfaceData{Battlefield: &bf, Theme: modern,
		Style: modern.Style(), Panel: panel, Fonts: fonts, Portrait: portrait,
		PortraitUnavailable: "尚未找到可驗證照片"}); err != nil {
		t.Fatal(err)
	}
	l, err := uilayout.NewModernMapLayout(uilayout.ModernDesignSurface())
	if err != nil {
		t.Fatal(err)
	}
	if got := c.Image().RGBAAt(l.CommanderPortrait.X+l.CommanderPortrait.W/2,
		l.CommanderPortrait.Y+24); got.R != 0x31 || got.G != 0x62 || got.B != 0x93 {
		t.Fatalf("已驗證司令肖像沒有畫入固定欄：%+v", got)
	}
}

func TestDrawModernMapSurfaceFailsClosedForWrongCanvasOrStyle(t *testing.T) {
	bf, panel := modernHighFixture()
	modern := uitheme.NewModern()
	if err := NewCanvas(640, 350).DrawModernMapSurface(ModernMapSurfaceData{
		Battlefield: &bf, Theme: modern, Style: modern.Style(), Panel: panel,
	}); err == nil {
		t.Fatal("640×350 不應靜默當成 H1 高解析畫布")
	}
	if err := NewCanvas(1280, 720).DrawModernMapSurface(ModernMapSurfaceData{
		Battlefield: &bf, Theme: modern, Style: uitheme.RetroStyle(), Panel: panel,
	}); err == nil {
		t.Fatal("retro style 不應走 Modern H1 renderer")
	}
}

func TestDrawModernBattleSurfaceUsesHighBattleLayout(t *testing.T) {
	bf, _ := modernHighFixture()
	modern := uitheme.NewModern()
	c := NewCanvas(1280, 720)
	if err := c.DrawModernBattleSurface(ModernBattleSurfaceData{
		Battlefield: &bf,
		Theme:       modern,
		Units:       modern,
		Panel: BattlePanelData{Province: 26, Month: 8, Day: 1,
			Attacker: BattleSide{Units: 2, Soldiers: 1200, Gold: 3},
			Defender: BattleSide{Units: 3, Soldiers: 1800, Food: 4}},
		Style: modern.Style(),
	}); err != nil {
		t.Fatal(err)
	}
	if got := c.Image().At(24, 24); got == (color.RGBA{}) {
		t.Fatal("高解析戰鬥頁首沒有繪製")
	}
	if got := c.Image().At(728, 100); got == (color.RGBA{}) {
		t.Fatal("高解析戰鬥面板沒有繪製")
	}
}

func TestDrawModernBattleSurfaceFailsClosedForMissingProvider(t *testing.T) {
	bf, _ := modernHighFixture()
	modern := uitheme.NewModern()
	if err := NewCanvas(1280, 720).DrawModernBattleSurface(ModernBattleSurfaceData{
		Battlefield: &bf, Theme: modern, Style: modern.Style(),
	}); err == nil {
		t.Fatal("缺少 unit provider 不應靜默繪製戰鬥")
	}
}

func TestDrawModernFlowSurfaceRendersOptionsAndNumericInput(t *testing.T) {
	modern := uitheme.NewModern()
	options := []string{"開墾土地", "建造兵工廠", "開採金礦"}
	c := NewCanvas(1280, 720)
	if err := c.DrawModernFlowSurface(ModernFlowSurfaceData{
		Title: "發展", Options: options, Selected: 1, Style: modern.Style(),
	}); err != nil {
		t.Fatal(err)
	}
	if got := c.Image().At(24, 24); got == (color.RGBA{}) {
		t.Fatal("Modern flow 標題頁沒有背景")
	}
	if err := NewCanvas(1280, 720).DrawModernFlowSurface(ModernFlowSurfaceData{
		Title: "輸入", Prompt: "輸入數量", Input: "123", InputMode: true, Style: modern.Style(),
	}); err != nil {
		t.Fatal(err)
	}
}

func TestDrawModernFlowSurfaceFailsClosedForWrongCanvas(t *testing.T) {
	modern := uitheme.NewModern()
	if err := NewCanvas(640, 350).DrawModernFlowSurface(ModernFlowSurfaceData{
		Title: "流程", Options: []string{"一"}, Style: modern.Style(),
	}); err == nil {
		t.Fatal("流程頁不應靜默接受 640×350")
	}
}

func TestDrawModernCommandAndOptionsSurfacesUseHighCanvas(t *testing.T) {
	modern := uitheme.NewModern()
	fonts := loadBiographyTestFonts(t)
	labels := make([]string, 15)
	for i := range labels {
		labels[i] = "指令"
	}
	if err := NewCanvas(1280, 720).DrawModernCommandSurface(ModernCommandSurfaceData{
		Title: "政略指令", Labels: labels, Fonts: fonts, Icons: modern, Style: modern.Style(), Back: "返回",
	}); err != nil {
		t.Fatal(err)
	}
	if err := NewCanvas(1280, 720).DrawModernOptionsSurface(ModernOptionSurfaceData{
		Title: "顯示設定", Options: []string{"原典用語", "現代白話", "復古圖形", "Modern 圖形"},
		Selected: []bool{true, false, false, true}, Fonts: fonts, Style: modern.Style(), Back: "返回",
	}); err != nil {
		t.Fatal(err)
	}
}

func TestDrawModernBiographyAndNarrativeSurfacesUseHighCanvas(t *testing.T) {
	fonts := loadBiographyTestFonts(t)
	modern := uitheme.NewModern()
	birth, death := 1887, 1975
	person := &i18n.Person{
		NameInGame: "測試者", NameCommon: "測試者", Birth: &birth, Death: &death,
		Birthplace: "浙江", Faction: "中央軍", HighestPost: "委員長", Confidence: "confirmed",
		Sources: []string{"fixture"}, Biography: "民國人物生平。",
	}
	result, err := NewCanvas(1280, 720).DrawModernBiographySurface(fonts, BiographyView{
		Person: person, Title: "人物自傳", Unavailable: "查無可靠傳記記載", Page: 0,
	}, modern.Style().Ink, modern.Style().Paper, modern.Style(), nil, "尚未找到可驗證照片", "返回", "上一頁", "下一頁")
	if err != nil {
		t.Fatal(err)
	}
	if result.PageCount != 1 || len(result.Missing) != 0 {
		t.Fatalf("Modern 自傳結果=%+v", result)
	}

	catalog, err := i18n.LoadNarrative(filepath.Join("..", "..", "..", "translations", "zh-Hant"))
	if err != nil {
		t.Fatal(err)
	}
	images := make([]*assets.Image, 4)
	for i := range images {
		images[i] = &assets.Image{W: 215, H: 16, Pix: make([]byte, 215*16)}
		images[i].Pix[(i+1)%len(images[i].Pix)] = byte(i + 1)
	}
	if err := NewCanvas(1280, 720).DrawModernNarrativeSurface(fonts, catalog, images, 0,
		"新聞／史事", i18n.WordingOriginal, modern.Style(), "返回", "上一頁", "下一頁"); err != nil {
		t.Fatal(err)
	}
}
