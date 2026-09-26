package render

import (
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

func syntheticThemeTileSet() *TileSet {
	pal := make(assets.Palette, 16)
	for i := range pal {
		pal[i] = assets.RGB{R: byte(i), G: byte(i * 3), B: byte(i * 5)}
	}
	tiles := make([]*assets.Image, int(assets.TileKindMax))
	for i := range tiles {
		pix := make([]byte, TileW*TileH)
		for j := range pix {
			pix[j] = 2
		}
		tiles[i] = &assets.Image{W: TileW, H: TileH, Pix: pix}
	}
	rails := make(assets.RailTiles, assets.RailTileCount)
	for i := range rails {
		pix := make([]byte, assets.RailW*assets.RailH)
		for y := 0; y < assets.RailH; y++ {
			for x := 0; x < assets.RailW; x++ {
				if (x+y+i)%9 == 0 {
					pix[y*assets.RailW+x] = byte((i % 15) + 1)
				}
			}
		}
		rails[i] = pix
	}
	return &TileSet{Tiles: tiles, Rails: rails, Pal: pal}
}

func syntheticBattlefield() *game.Battlefield {
	bf := &game.Battlefield{}
	for y := 0; y < assets.GridH; y++ {
		for x := 0; x < assets.GridW; x++ {
			bf.Tiles[y][x] = assets.Tile{Kind: assets.TileKind((x+y)%int(assets.TileKindMax) + 1), Rail: assets.NoRail}
		}
	}
	bf.Tiles[0][0].Rail = 0 // rail tile 0 is a valid vertical tile, not an empty tile.
	return bf
}

func TestRetroAdapterKeepsLegacyBattlefieldPixels(t *testing.T) {
	ts := syntheticThemeTileSet()
	bf := syntheticBattlefield()
	w, h := assets.GridW*TileW, assets.GridH*TileH+TileH/2
	legacy := NewCanvas(w, h)
	modernized := NewCanvas(w, h)
	if err := legacy.DrawTiledBattlefield(bf, ts, 0, 0); err != nil {
		t.Fatal(err)
	}
	if err := modernized.DrawThemedBattlefield(bf, NewRetroTheme(ts), 0, 0); err != nil {
		t.Fatal(err)
	}
	n, err := DiffCount(legacy.Image(), modernized.Image())
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("retro adapter changed %d pixels", n)
	}
}

func TestThemedBattlefieldTreatsRailZeroAsValid(t *testing.T) {
	ts := syntheticThemeTileSet()
	bf := syntheticBattlefield()
	c := NewCanvas(assets.GridW*TileW, assets.GridH*TileH+TileH/2)
	if err := c.DrawThemedBattlefield(bf, NewRetroTheme(ts), 0, 0); err != nil {
		t.Fatal(err)
	}
	// Cell (0,0) uses rail 0. Its synthetic rail has a non-zero pixel at (0,0)
	// only when the deterministic pattern hits; scan the cell to prove overlay.
	var railColour bool
	for y := 0; y < TileH; y++ {
		for x := 0; x < TileW; x++ {
			if got := c.Image().RGBAAt(x, y); got.R == ts.Pal[1].R &&
				got.G == ts.Pal[1].G && got.B == ts.Pal[1].B {
				railColour = true
			}
		}
	}
	if !railColour {
		t.Fatal("rail 0 沒有疊上非透明像素")
	}
}

func TestRetroUnitProviderKeepsOriginalPixels(t *testing.T) {
	ts := syntheticThemeTileSet()
	icons := make([]*assets.Image, IconCount)
	for i := range icons {
		pix := make([]byte, IconW*IconH)
		for j := range pix {
			if (i+j)%7 == 0 {
				pix[j] = byte((i % 15) + 1)
			}
		}
		icons[i] = &assets.Image{W: IconW, H: IconH, Pix: pix}
	}
	pal := ts.Pal
	provider := NewRetroThemeWithIcons(ts, icons, pal)
	units, ok := provider.(uitheme.UnitProvider)
	if !ok {
		t.Fatal("retro 主題未提供部隊圖示介面")
	}
	b, err := units.Unit(7)
	if err != nil {
		t.Fatal(err)
	}
	cell := game.CellIndex(15)
	legacy := NewCanvas(100, 100)
	themed := NewCanvas(100, 100)
	if err := legacy.DrawUnitAtCell(icons, 7, pal, 0, 0, cell); err != nil {
		t.Fatal(err)
	}
	if err := themed.DrawThemedUnitAtCell(b, 0, 0, cell); err != nil {
		t.Fatal(err)
	}
	n, err := DiffCount(legacy.Image(), themed.Image())
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatalf("retro unit provider changed %d pixels", n)
	}
}

func TestThemedHUDIconRejectsInvalidPaletteIndex(t *testing.T) {
	pix := make([]byte, uitheme.HUDIconW*uitheme.HUDIconH)
	pix[0] = 16 // 調色盤只有索引 0..15，必須 fail-closed。
	palette := make(assets.Palette, 16)
	icon := uitheme.Bitmap{
		Image:   &assets.Image{W: uitheme.HUDIconW, H: uitheme.HUDIconH, Pix: pix},
		Palette: palette,
	}
	if err := NewBGICanvas().DrawThemedHUDIcon(icon, 0, 0); err == nil {
		t.Fatal("HUD 圖示像素索引超出調色盤時應拒絕")
	}
}
