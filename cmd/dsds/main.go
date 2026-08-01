// dsds 是《大時代的故事》remake 的執行檔。
//
// 目前做到的：載入原版資料，用**原版的地形圖塊**畫出各省的 14×14 戰場，
// 側欄顯示省名，可切換省份。
// 遊戲邏輯（戰鬥、政略指令、存檔）尚未實作——那些的規格還沒解出來。
//
//	tools/go.sh run ./cmd/dsds -game workplace/orig/game
//
// 操作：← → 切換省份，F10 離開（CLAUDE.md §9：ESC 只取消／退回，F10 才離開）。
//
// 需要顯示器。無頭環境請跑 internal/ui/render 的測試，
// 那一層不依賴 Ebiten，會逐像素比對原版截圖。
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
)

// scale 是視窗放大倍率。
//
// remake 差異（CLAUDE.md §1「外殼允許現代化」）：只放大視窗，
// 邏輯解析度仍是原版的 640×350。
const scale = 2

// 版面。戰場是 14×14 格 × 32×24 = 448×336，放進 640×350 之後
// 右側剩 192 寬給側欄——這個比例與原版政略畫面（左側面板 + 右側地圖）一致，
// 但**座標本身是 remake 的排版選擇，不是還原原版數值**。
const (
	fieldX, fieldY = 0, 14
	sideX          = 452
	nameY          = 16
)

type app struct {
	m        *game.Map
	provName *assets.GlyphFile // 3.15：前 39 個三字詞條是省名
	tiles    *render.TileSet   // NEWTERR.TPC 的 22 張地形圖塊
	current  game.ProvinceID
	dirty    bool
	frame    *ebiten.Image
}

func (a *app) Update() error {
	if ebiten.IsKeyPressed(ebiten.KeyF10) {
		return ebiten.Termination
	}
	switch {
	case inpututil.IsKeyJustPressed(ebiten.KeyRight):
		a.current++
		if a.current > game.ProvinceCount {
			a.current = 1
		}
		a.dirty = true
	case inpututil.IsKeyJustPressed(ebiten.KeyLeft):
		a.current--
		if a.current < 1 {
			a.current = game.ProvinceCount
		}
		a.dirty = true
	}
	return nil
}

func (a *app) compose() error {
	c := render.NewBGICanvas()

	bf, err := a.m.Battlefield(a.current)
	if err != nil {
		return err
	}
	// 用原版的 NEWTERR 圖塊畫戰場：NWMAP 的地物編號減 1 就是圖塊索引。
	if err := c.DrawTiledBattlefield(bf, a.tiles, fieldX, fieldY); err != nil {
		return err
	}

	// 側欄：省名（3.15 的第 current-1 個三字詞條）
	if err := c.DrawEntry(a.provName, int(a.current)-1, 3,
		assets.RGB{R: 0xFF, G: 0xFF, B: 0xA2}, sideX, nameY, true); err != nil {
		return err
	}

	a.frame = ebiten.NewImageFromImage(c.Image())
	a.dirty = false
	return nil
}

func (a *app) Draw(dst *ebiten.Image) {
	if a.dirty || a.frame == nil {
		if err := a.compose(); err != nil {
			fmt.Fprintln(os.Stderr, "合成失敗:", err)
			return
		}
	}
	dst.DrawImage(a.frame, nil)
}

func (a *app) Layout(_, _ int) (int, int) {
	return render.ModeBGIW, render.ModeBGIH
}

func main() {
	gameDir := flag.String("game", "workplace/orig/game", "原版素材目錄（唯讀）")
	start := flag.Int("province", 26, "起始省編號（1-39），預設 26 = 湖北省")
	flag.Parse()

	if err := run(*gameDir, game.ProvinceID(*start)); err != nil {
		fmt.Fprintln(os.Stderr, "錯誤:", err)
		os.Exit(1)
	}
}

func run(dir string, start game.ProvinceID) error {
	if !start.Valid() {
		return fmt.Errorf("省編號 %d 超出 1..%d", start, game.ProvinceCount)
	}
	read := func(name string) ([]byte, error) {
		b, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("讀不到 %s: %w", name, err)
		}
		return b, nil
	}

	warpos, err := read("WARPOS.DAT")
	if err != nil {
		return err
	}
	tername, err := read("TERNAME.DAT")
	if err != nil {
		return err
	}
	nwmap, err := read("NWMAP.DAT")
	if err != nil {
		return err
	}
	m, err := game.LoadMap(warpos, tername, nwmap)
	if err != nil {
		return err
	}

	names, err := read("3.15")
	if err != nil {
		return err
	}
	gf, err := assets.ParseGlyphFile(names)
	if err != nil {
		return err
	}

	newterr, err := read("NEWTERR.TPC")
	if err != nil {
		return err
	}
	rail, err := read("RAIL.TPC")
	if err != nil {
		return err
	}
	ts, err := render.LoadTileSet(newterr, rail, assets.EGADefaultPalette)
	if err != nil {
		return err
	}

	ebiten.SetWindowSize(render.ModeBGIW*scale, render.ModeBGIH*scale)
	ebiten.SetWindowTitle("大時代的故事")
	return ebiten.RunGame(&app{
		m: m, provName: gf, tiles: ts, current: start, dirty: true,
	})
}
