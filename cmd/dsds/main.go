// dsds 是《大時代的故事》remake 的執行檔。
//
// 目前只做到「把原版畫面重現出來」——載入 .TPC 與 .RGB，合成後用 Ebiten 顯示。
// 遊戲邏輯（M3/M4）尚未接上。
//
//	tools/go.sh run ./cmd/dsds -game workplace/orig/game
//
// 需要顯示器。無頭環境請改跑 internal/ui/render 的測試，
// 那一層不依賴 Ebiten，會逐像素比對原版截圖。
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
)

// scale 是視窗放大倍率。原版是 640×350，現代螢幕上太小。
//
// 這是 remake 差異（CLAUDE.md §1「外殼允許現代化」）：只放大視窗，
// 不改變邏輯解析度，畫面內容仍是原版的 640×350。
const scale = 2

type app struct {
	screen *ebiten.Image
}

func (a *app) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyF10) {
		return ebiten.Termination
	}
	return nil
}

func (a *app) Draw(dst *ebiten.Image) {
	dst.DrawImage(a.screen, nil)
}

func (a *app) Layout(_, _ int) (int, int) {
	return render.ModeBGIW, render.ModeBGIH
}

func main() {
	game := flag.String("game", "workplace/orig/game", "原版素材目錄（唯讀）")
	tpc := flag.String("tpc", "CHOOSE1.TPC", "要顯示的 .TPC")
	rgb := flag.String("rgb", "CHOOSE.RGB", "調色盤")
	x := flag.Int("x", 90, "貼圖的 x 座標")
	y := flag.Int("y", 20, "貼圖的 y 座標")
	flag.Parse()

	if err := run(*game, *tpc, *rgb, *x, *y); err != nil {
		fmt.Fprintln(os.Stderr, "錯誤:", err)
		os.Exit(1)
	}
}

func run(gameDir, tpcName, rgbName string, x, y int) error {
	data, err := os.ReadFile(filepath.Join(gameDir, tpcName))
	if err != nil {
		return fmt.Errorf("讀不到 %s: %w", tpcName, err)
	}
	im, err := assets.DecodeBGI(data, 0)
	if err != nil {
		return err
	}

	f, err := os.Open(filepath.Join(gameDir, rgbName))
	if err != nil {
		return fmt.Errorf("讀不到 %s: %w", rgbName, err)
	}
	defer f.Close()
	pal, err := assets.ParseRGBFile(f)
	if err != nil {
		return err
	}

	c := render.NewBGICanvas()
	if err := c.DrawBGI(im, pal, x, y); err != nil {
		return err
	}

	ebiten.SetWindowSize(render.ModeBGIW*scale, render.ModeBGIH*scale)
	ebiten.SetWindowTitle("大時代的故事")
	// F10 離開（CLAUDE.md §9：ESC 只取消／退回，F10 才離開）
	return ebiten.RunGame(&app{screen: ebiten.NewImageFromImage(c.Image())})
}
