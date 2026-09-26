package render

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

func TestStrategyPanelDrawsCommandCount(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(gameDir, name))
		if err != nil {
			t.Skipf("沒有原版素材 %s，跳過", name)
		}
		return b
	}
	fonts, err := LoadPanelFonts(read("1.15"), read("2.15"), read("3.15"), read("MAN115"))
	if err != nil {
		t.Fatal(err)
	}
	if w1Count >= len(fonts.W1.Glyphs) || (w2Command+1)*2 > len(fonts.W2.Glyphs) {
		t.Fatal("「指令數」詞條超出原版字模範圍")
	}

	p := &game.Province{Commander: 1, Governor: 1}
	c := NewBGICanvas()
	if err := c.DrawStrategyPanel(PanelData{
		ID: 1, Province: p, Commands: 4,
	}, fonts); err != nil {
		t.Fatal(err)
	}

	// 指令數是面板最後一行；右側數值區必須真的畫出「4」，不能只把值傳進資料。
	pixels := 0
	for y := 320; y < 340; y++ {
		for x := panelValue - 20; x <= panelValue; x++ {
			r, g, b, _ := c.Image().At(x, y).RGBA()
			if uint8(r>>8) == panelFG.R && uint8(g>>8) == panelFG.G && uint8(b>>8) == panelFG.B {
				pixels++
			}
		}
	}
	if pixels == 0 {
		t.Error("指令數的右側數值區沒有任何前景像素")
	}
}

func TestStrategyPanelModernHUDIconsStaySupplementalToText(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(gameDir, name))
		if err != nil {
			t.Skipf("沒有原版素材 %s，跳過", name)
		}
		return b
	}
	fonts, err := LoadPanelFonts(read("1.15"), read("2.15"), read("3.15"), read("MAN115"))
	if err != nil {
		t.Fatal(err)
	}
	c := NewBGICanvas()
	if err := c.DrawStrategyPanel(PanelData{
		ID: 1, Province: &game.Province{Commander: 1, Governor: 1,
			Gold: 100, Food: 200, Ammo: 300, Fuel: 400, Coal: 500, Iron: 600},
		Icons: uitheme.NewModern(),
	}, fonts); err != nil {
		t.Fatal(err)
	}
	// 資源圖示位於面板標籤左側（x=8..15）；這段不會被文字覆蓋。
	colored := 0
	for y := 70; y < 220; y++ {
		for x := panelX; x < panelLabel; x++ {
			r, g, b, _ := c.Image().At(x, y).RGBA()
			if uint8(r>>8) != panelBG.R || uint8(g>>8) != panelBG.G || uint8(b>>8) != panelBG.B {
				colored++
			}
		}
	}
	if colored == 0 {
		t.Fatal("modern 政略面板沒有畫出資源 HUD 圖示")
	}
}

func TestModernStrategyPanelUsesSemanticCardAndFactionTint(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	read := func(name string) []byte {
		b, err := os.ReadFile(filepath.Join(gameDir, name))
		if err != nil {
			t.Skipf("沒有原版素材 %s，跳過", name)
		}
		return b
	}
	fonts, err := LoadPanelFonts(read("1.15"), read("2.15"), read("3.15"), read("MAN115"))
	if err != nil {
		t.Fatal(err)
	}
	eten, err := assets.LoadEtenFonts(filepath.Join("../../../workplace", "eten"))
	if err != nil {
		t.Skipf("沒有完整倚天字庫，跳過：%v", err)
	}
	style := uitheme.NewModern().Style()
	c := NewBGICanvas()
	if err := c.DrawStrategyPanel(PanelData{
		ID: 1, Province: &game.Province{Commander: 1, Governor: 1,
			Gold: 100, Food: 200, Ammo: 300, Fuel: 400, Coal: 500, Iron: 600},
		Commands: 4, Icons: uitheme.NewModern(), Style: style, SemanticFonts: eten,
		ProvinceName: "湖北", CommanderName: "蔣中正", GovernorName: "蔣中正",
		Status: "正常", Faction: 2,
		Labels: PanelLabels{Status: "正常", Commander: "司令", Governor: "省長",
			Gold: "黃金", Food: "糧食", Ammo: "彈藥", Fuel: "燃料", Coal: "煤礦", Iron: "鐵礦",
			Land: "地價", Population: "人口", Cities: "城市", Arsenal: "兵工廠",
			Force: "兵力", Generals: "將領", People: "人民", Loyalty: "忠誠度",
			Commands: "指令", Count: "數"},
	}, fonts); err != nil {
		t.Fatal(err)
	}
	card := c.Image().RGBAAt(10, 10)
	if card.R != style.Panel.R || card.G != style.Panel.G || card.B != style.Panel.B {
		t.Fatalf("modern 面板標題卡片顏色=%v，預期=%v", card, style.Panel)
	}
	tint := c.Image().RGBAAt(5, 10)
	want := style.FactionTint[1]
	if tint.R != want.R || tint.G != want.G || tint.B != want.B {
		t.Fatalf("勢力色帶=%v，預期=%v", tint, want)
	}
}
