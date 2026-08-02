package render

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
)

// TestStrategyCommandsShape 指令表要是 1..15 連續，且每個詞條索引在字模檔範圍內。
func TestStrategyCommandsShape(t *testing.T) {
	if len(StrategyCommands) != 15 {
		t.Fatalf("政略指令應為 15 個，實得 %d", len(StrategyCommands))
	}
	for i, c := range StrategyCommands {
		if c.Num != i+1 {
			t.Errorf("第 %d 個指令的編號是 %d，預期 %d", i, c.Num, i+1)
		}
		if len(c.Parts) == 0 || len(c.Parts) > 2 {
			t.Errorf("指令 %d 由 %d 個詞條組成，預期 1 或 2", c.Num, len(c.Parts))
		}
	}
}

// TestCommandWordsResolve 每個詞條都要能從原版字模檔取到，不越界。
//
// 這一條同時保證「指令名是用原版詞表拼的」——如果哪天有人改成寫死字串，
// 這個測試就沒有東西可驗了。
func TestCommandWordsResolve(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	read := func(name string) *assets.GlyphFile {
		b, err := os.ReadFile(filepath.Join(gameDir, name))
		if err != nil {
			t.Skipf("沒有原版素材 %s，跳過", name)
		}
		gf, err := assets.ParseGlyphFile(b)
		if err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		return gf
	}
	f := CommandFonts{W2: read("2.15"), W4: read("4.15")}

	for _, c := range StrategyCommands {
		for _, p := range c.Parts {
			gf, w := f.W2, 2
			if p.FromW4 {
				gf, w = f.W4, 4
			}
			if end := (p.Index + 1) * w; end > len(gf.Glyphs) {
				t.Errorf("指令 %d 的詞條 %d（槽寬 %d）超出字模檔範圍（共 %d 條）",
					c.Num, p.Index, w, len(gf.Glyphs)/w)
			}
		}
	}
}

// TestCommandMenuDraws 畫得出來且不報錯，選單區域確實有前景像素。
func TestCommandMenuDraws(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	load := func(name string) *assets.GlyphFile {
		b, err := os.ReadFile(filepath.Join(gameDir, name))
		if err != nil {
			t.Skipf("沒有原版素材 %s，跳過", name)
		}
		gf, _ := assets.ParseGlyphFile(b)
		return gf
	}
	f := CommandFonts{W2: load("2.15"), W4: load("4.15")}

	c := NewBGICanvas()
	fg := assets.RGB{R: 0xAE}
	if err := c.DrawCommandMenu(f, fg, 20, 20); err != nil {
		t.Fatal(err)
	}
	n := 0
	b := c.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			r, g, bl, _ := c.Image().At(x, y).RGBA()
			if uint8(r>>8) == fg.R && uint8(g>>8) == fg.G && uint8(bl>>8) == fg.B {
				n++
			}
		}
	}
	if n < 1000 {
		t.Errorf("選單只畫出 %d 個前景像素，看起來沒畫成功", n)
	}
	t.Logf("十五個指令畫出 %d 個前景像素", n)
}

func TestTradeScreensUseOriginalGlyphs(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	b, err := os.ReadFile(filepath.Join(gameDir, "2.15"))
	if err != nil {
		t.Skipf("沒有原版素材 2.15，跳過")
	}
	w2, err := assets.ParseGlyphFile(b)
	if err != nil {
		t.Fatal(err)
	}
	f := CommandFonts{W2: w2}
	fg, bg := assets.RGB{R: 0xAE}, assets.RGB{R: 0xFF, G: 0xFF, B: 0xA2}

	for _, draw := range []func(*Canvas) error{
		func(c *Canvas) error { return c.DrawTradeMenu(f, fg, bg, 190, 0, 450, 350, nil) },
		func(c *Canvas) error {
			v := true
			return c.DrawTradeMenu(f, fg, bg, 190, 0, 450, 350, &v)
		},
		func(c *Canvas) error {
			v := false
			return c.DrawTradeMenu(f, fg, bg, 190, 0, 450, 350, &v)
		},
		func(c *Canvas) error {
			return c.DrawTradeAmount(f, fg, bg, 190, 0, 450, 350, true, 0, 300)
		},
	} {
		c := NewBGICanvas()
		if err := draw(c); err != nil {
			t.Fatal(err)
		}
		pixels := 0
		for y := 20; y < 180; y++ {
			for x := 200; x < 620; x++ {
				r, _, _, _ := c.Image().At(x, y).RGBA()
				if uint8(r>>8) == fg.R {
					pixels++
				}
			}
		}
		if pixels == 0 {
			t.Error("商業活動畫面沒有畫出原版字模或數量")
		}
	}
}

func TestCovertScreensUseOriginalGlyphs(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	load := func(name string) *assets.GlyphFile {
		b, err := os.ReadFile(filepath.Join(gameDir, name))
		if err != nil {
			t.Skipf("沒有原版素材 %s，跳過", name)
		}
		gf, err := assets.ParseGlyphFile(b)
		if err != nil {
			t.Fatal(err)
		}
		return gf
	}
	f := CommandFonts{W2: load("2.15"), W4: load("4.15")}
	w3 := load("3.15")
	fg, bg := assets.RGB{R: 0xAE}, assets.RGB{R: 0xFF, G: 0xFF, B: 0xA2}
	for _, draw := range []func(*Canvas) error{
		func(c *Canvas) error { return c.DrawCovertAction(f, w3, fg, bg, 190, 0, 450, 350) },
		func(c *Canvas) error { return c.DrawCovertTarget(f, fg, bg, 190, 0, 450, 350, 2, 19) },
	} {
		c := NewBGICanvas()
		if err := draw(c); err != nil {
			t.Fatal(err)
		}
		pixels := 0
		for y := 10; y < 150; y++ {
			for x := 195; x < 635; x++ {
				r, _, _, _ := c.Image().At(x, y).RGBA()
				if uint8(r>>8) == fg.R {
					pixels++
				}
			}
		}
		if pixels == 0 {
			t.Fatal("秘密行動畫面沒有畫出原版字模或數值")
		}
	}
}

func TestSupplyScreensUseOriginalGlyphs(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	load := func(name string) *assets.GlyphFile {
		b, err := os.ReadFile(filepath.Join(gameDir, name))
		if err != nil {
			t.Skipf("沒有原版素材 %s，跳過", name)
		}
		gf, err := assets.ParseGlyphFile(b)
		if err != nil {
			t.Fatal(err)
		}
		return gf
	}
	f := CommandFonts{W2: load("2.15")}
	w3 := load("3.15")
	fg, bg := assets.RGB{R: 0xAE}, assets.RGB{R: 0xFF, G: 0xFF, B: 0xA2}

	for _, draw := range []func(*Canvas) error{
		func(c *Canvas) error {
			return c.DrawSupplyTarget(f, w3, fg, bg, 190, 0, 450, 350,
				[]int{19, 26}, 19)
		},
		func(c *Canvas) error {
			return c.DrawSupplyAmount(f, fg, bg, 190, 0, 450, 350, 2, 500)
		},
	} {
		c := NewBGICanvas()
		if err := draw(c); err != nil {
			t.Fatal(err)
		}
		pixels := 0
		for y := 15; y < 190; y++ {
			for x := 200; x < 630; x++ {
				r, _, _, _ := c.Image().At(x, y).RGBA()
				if uint8(r>>8) == fg.R {
					pixels++
				}
			}
		}
		if pixels == 0 {
			t.Error("運補畫面沒有畫出原版字模、省名或數量")
		}
	}
}

func TestPlayerTransferScreensUseOriginalGlyphs(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	load := func(name string) *assets.GlyphFile {
		b, err := os.ReadFile(filepath.Join(gameDir, name))
		if err != nil {
			t.Skipf("沒有原版素材 %s，跳過", name)
		}
		gf, err := assets.ParseGlyphFile(b)
		if err != nil {
			t.Fatal(err)
		}
		return gf
	}
	f := CommandFonts{W2: load("2.15"), W4: load("4.15")}
	gn := load("MAN115")
	fg, bg := assets.RGB{R: 0xAE}, assets.RGB{R: 0xFF, G: 0xFF, B: 0xA2}
	for _, draw := range []func(*Canvas) error{
		func(c *Canvas) error {
			return c.DrawPlayerTransferMode(f, fg, bg, 190, 0, 450, 350)
		},
		func(c *Canvas) error {
			return c.DrawPlayerTransferTarget(f, load("3.15"), fg, bg, 190, 0, 450, 350,
				[]int{19, 26}, 19)
		},
		func(c *Canvas) error {
			return c.DrawPlayerTransferSelection(f, gn, fg, bg, 190, 0, 450, 350,
				[]int{58, 59, 60}, []bool{true, false, true}, 1, 0, game.PlayerTransferAll)
		},
	} {
		c := NewBGICanvas()
		if err := draw(c); err != nil {
			t.Fatal(err)
		}
		pixels := 0
		for y := 5; y < 345; y++ {
			for x := 195; x < 635; x++ {
				r, _, _, _ := c.Image().At(x, y).RGBA()
				if uint8(r>>8) == fg.R {
					pixels++
				}
			}
		}
		if pixels == 0 {
			t.Fatal("玩家調動畫面沒有畫出原版字模、候選或游標")
		}
	}
}

func TestRecruitScreensUseOriginalGlyphs(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	load := func(name string) *assets.GlyphFile {
		b, err := os.ReadFile(filepath.Join(gameDir, name))
		if err != nil {
			t.Skipf("沒有原版素材 %s，跳過", name)
		}
		gf, err := assets.ParseGlyphFile(b)
		if err != nil {
			t.Fatal(err)
		}
		return gf
	}
	f := CommandFonts{W2: load("2.15"), W4: load("4.15")}
	fg, bg := assets.RGB{R: 0xAE}, assets.RGB{R: 0xFF, G: 0xFF, B: 0xA2}

	for _, draw := range []func(*Canvas) error{
		func(c *Canvas) error { return c.DrawRecruitAction(f, fg, bg, 190, 0, 450, 350) },
		func(c *Canvas) error { return c.DrawRecruitBranch(f, fg, bg, 190, 0, 450, 350) },
		func(c *Canvas) error {
			return c.DrawRecruitAmount(f, fg, bg, 190, 0, 450, 350, 5, 200, 12)
		},
		func(c *Canvas) error {
			return c.DrawRecruitConfirm(f, fg, bg, 190, 0, 450, 350, 120)
		},
	} {
		c := NewBGICanvas()
		if err := draw(c); err != nil {
			t.Fatal(err)
		}
		pixels := 0
		for y := 15; y < 210; y++ {
			for x := 200; x < 630; x++ {
				r, _, _, _ := c.Image().At(x, y).RGBA()
				if uint8(r>>8) == fg.R {
					pixels++
				}
			}
		}
		if pixels == 0 {
			t.Error("徵兵畫面沒有畫出原版字模或數量")
		}
	}
}

func TestReorganizationScreensUseOriginalGlyphs(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	load := func(name string) *assets.GlyphFile {
		b, err := os.ReadFile(filepath.Join(gameDir, name))
		if err != nil {
			t.Skipf("沒有原版素材 %s，跳過", name)
		}
		gf, err := assets.ParseGlyphFile(b)
		if err != nil {
			t.Fatal(err)
		}
		return gf
	}
	f := CommandFonts{W2: load("2.15"), W4: load("4.15")}
	man := load("MAN115")
	fg, bg := assets.RGB{R: 0xAE}, assets.RGB{R: 0xFF, G: 0xFF, B: 0xA2}

	for _, draw := range []func(*Canvas) error{
		func(c *Canvas) error {
			return c.DrawReorganizationTarget(f, man, fg, bg, 190, 0, 450, 350,
				[]int{1, 2, 3}, 7500, 2)
		},
		func(c *Canvas) error {
			return c.DrawReorganizationAmount(f, man, fg, bg, 190, 0, 450, 350,
				1, 7500, 20000, 5000)
		},
	} {
		c := NewBGICanvas()
		if err := draw(c); err != nil {
			t.Fatal(err)
		}
		pixels := 0
		for y := 15; y < 310; y++ {
			for x := 200; x < 630; x++ {
				r, _, _, _ := c.Image().At(x, y).RGBA()
				if uint8(r>>8) == fg.R {
					pixels++
				}
			}
		}
		if pixels == 0 {
			t.Error("重新整編畫面沒有畫出原版字模、將領名或數量")
		}
	}
}

func TestTrainConfirmUsesOriginalGlyphs(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	load := func(name string) *assets.GlyphFile {
		b, err := os.ReadFile(filepath.Join(gameDir, name))
		if err != nil {
			t.Skipf("沒有原版素材 %s，跳過", name)
		}
		gf, err := assets.ParseGlyphFile(b)
		if err != nil {
			t.Fatal(err)
		}
		return gf
	}
	f := CommandFonts{W2: load("2.15"), W4: load("4.15")}
	fg, bg := assets.RGB{R: 0xAE}, assets.RGB{R: 0xFF, G: 0xFF, B: 0xA2}
	c := NewBGICanvas()
	if err := c.DrawTrainConfirm(f, fg, bg, 190, 0, 450, 350); err != nil {
		t.Fatal(err)
	}
	pixels := 0
	for y := 30; y < 210; y++ {
		for x := 200; x < 630; x++ {
			r, _, _, _ := c.Image().At(x, y).RGBA()
			if uint8(r>>8) == fg.R {
				pixels++
			}
		}
	}
	if pixels == 0 {
		t.Fatal("練兵確認畫面沒有畫出原版字模")
	}
}

func TestDevelopPageUsesThreeOriginalGlyphEntries(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	load := func(name string) *assets.GlyphFile {
		b, err := os.ReadFile(filepath.Join(gameDir, name))
		if err != nil {
			t.Skipf("沒有原版素材 %s，跳過", name)
		}
		gf, err := assets.ParseGlyphFile(b)
		if err != nil {
			t.Fatal(err)
		}
		return gf
	}
	f := PanelFonts{W1: load("1.15"), W2: load("2.15"), W3: load("3.15")}
	fg, bg := assets.RGB{R: 0xAE}, assets.RGB{R: 0xFF, G: 0xFF, B: 0xA2}
	c := NewBGICanvas()
	if err := c.DrawDevelopPage(f, fg, bg, 190, 0, 450, 350); err != nil {
		t.Fatal(err)
	}
	for i, top := range []int{26, 64, 102} {
		pixels := 0
		for y := top; y < top+23; y++ {
			for x := 210; x < 380; x++ {
				r, g, b, _ := c.Image().At(x, y).RGBA()
				if uint8(r>>8) == fg.R && uint8(g>>8) == fg.G && uint8(b>>8) == fg.B {
					pixels++
				}
			}
		}
		if pixels == 0 {
			t.Errorf("發展選單第 %d 項沒有畫出編號與原版字模", i+1)
		}
	}
}

func TestGeneralViewScreensUseOriginalGlyphs(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	load := func(name string) *assets.GlyphFile {
		b, err := os.ReadFile(filepath.Join(gameDir, name))
		if err != nil {
			t.Skipf("沒有原版素材 %s，跳過", name)
		}
		gf, err := assets.ParseGlyphFile(b)
		if err != nil {
			t.Fatal(err)
		}
		return gf
	}
	w1, w2, w3, man := load("1.15"), load("2.15"), load("3.15"), load("MAN115")
	f := CommandFonts{W2: w2, W4: load("4.15")}
	pf := PanelFonts{W1: w1, W2: w2, W3: w3, Gen: man}
	fg, bg := assets.RGB{R: 0xAE}, assets.RGB{R: 0xFF, G: 0xFF, B: 0xA2}
	g := game.General{AbilityA: 90, AbilityB: 100, Experience: 30, Branch: game.BranchInfantry,
		Force: 20000, F19: 50, F20: 10, Stamina: 55, F30: 74,
		FactionName: 4, TitlePrefix: 10, TitleNumber: 25, TitleSuffix: 18}
	fan := load("FAN(1).15")
	for _, draw := range []func(*Canvas) error{
		func(c *Canvas) error {
			return c.DrawGeneralList(f, man, fg, bg, 190, 0, 450, 350,
				[]int{58, 59, 60}, []uint16{20000, 6000, 4500}, 0)
		},
		func(c *Canvas) error {
			ids, forces := make([]int, 25), make([]uint16, 25)
			for i := range ids {
				ids[i], forces[i] = i+1, uint16(1000+i)
			}
			return c.DrawGeneralList(f, man, fg, bg, 190, 0, 450, 350, ids, forces, 24)
		},
		func(c *Canvas) error {
			return c.DrawGeneralDetail(f, pf, fan, fg, bg, 190, 0, 450, 350, 58, g, 33)
		},
	} {
		c := NewBGICanvas()
		if err := draw(c); err != nil {
			t.Fatal(err)
		}
		pixels := 0
		for y := 10; y < 340; y++ {
			for x := 195; x < 635; x++ {
				r, _, _, _ := c.Image().At(x, y).RGBA()
				if uint8(r>>8) == fg.R {
					pixels++
				}
			}
		}
		if pixels == 0 {
			t.Fatal("查閱將領畫面沒有畫出原版字模或數值")
		}
	}
}

func TestProvinceViewScreensUseOriginalGlyphs(t *testing.T) {
	const gameDir = "../../../workplace/orig/game"
	load := func(name string) *assets.GlyphFile {
		b, err := os.ReadFile(filepath.Join(gameDir, name))
		if err != nil {
			t.Skipf("沒有原版素材 %s，跳過", name)
		}
		gf, err := assets.ParseGlyphFile(b)
		if err != nil {
			t.Fatal(err)
		}
		return gf
	}
	w1, w2, w3, man := load("1.15"), load("2.15"), load("3.15"), load("MAN115")
	f := CommandFonts{W2: w2, W4: load("4.15")}
	pf := PanelFonts{W1: w1, W2: w2, W3: w3, Gen: man}
	fg, bg := assets.RGB{R: 0xAE}, assets.RGB{R: 0xFF, G: 0xFF, B: 0xA2}
	p := &game.Province{Governor: 58, Gold: 4200, Food: 18050, Ammo: 8787,
		Fuel: 12048, Coal: 13000, Iron: 14031, LandValue: 22,
		Population: 18250000, Cities: 5, Arsenals: 3, Loyalty: 79}
	data := PanelData{ID: 26, Province: p, Force: 97500, Generals: 15}
	for _, draw := range []func(*Canvas) error{
		func(c *Canvas) error { return c.DrawViewMenu(f, fg, bg, 190, 0, 450, 350) },
		func(c *Canvas) error { return c.DrawProvinceSelect(f, w3, fg, bg, 190, 0, 450, 350, 26) },
		func(c *Canvas) error { return c.DrawProvinceChoice(f, w3, fg, bg, 190, 0, 450, 350, 19) },
		func(c *Canvas) error { return c.DrawProvinceNames(f, w3, fg, bg, 190, 0, 450, 350, 1, 36) },
		func(c *Canvas) error { return c.DrawProvinceNames(f, w3, fg, bg, 190, 0, 450, 350, 2, 36) },
		func(c *Canvas) error { return c.DrawProvinceDetail(f, pf, data, fg, bg, 190, 0, 450, 350) },
		func(c *Canvas) error {
			return c.DrawOwnedProvinceOverview(f, pf, fg, bg, 190, 0, 450, 350,
				[]int{19, 26}, []uint32{140500, 97500})
		},
	} {
		c := NewBGICanvas()
		if err := draw(c); err != nil {
			t.Fatal(err)
		}
		pixels := 0
		for y := 5; y < 345; y++ {
			for x := 195; x < 635; x++ {
				r, _, _, _ := c.Image().At(x, y).RGBA()
				if uint8(r>>8) == fg.R {
					pixels++
				}
			}
		}
		if pixels == 0 {
			t.Fatal("省份查閱畫面沒有畫出原版字模或數值")
		}
	}
}
