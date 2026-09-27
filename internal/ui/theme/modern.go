package theme

import (
	"fmt"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

const (
	modernTileCount = int(assets.TileKindMax)
	modernRailCount = assets.RailTileCount
	modernTileW     = 32
	modernTileH     = 24
	modernUnitCount = 18
	modernUnitW     = 32
	modernUnitH     = 17
)

// Modern 是 P1 的 deterministic modern terrain theme。
//
// 它保留原版 32×24 圖塊尺寸、22 個地物索引與 21 個鐵路索引，只改變
// 顏色、紋理與線條呈現。資產以純 Go 生成，沒有第三方字型／原版衍生檔，
// 因而可以在所有平台重現；未來有經審核的 SVG/PNG 時可替換這個 provider，
// 不需改 renderer 或規則層。
type Modern struct {
	tiles     []Bitmap
	rails     []Bitmap
	units     []Bitmap
	resources []Bitmap
	commands  []Bitmap
	ornaments *OriginalOrnaments
}

// SetOriginalOrnaments 原子掛入玩家原版資料的執行期裝飾遮罩。無效組不取代
// 程式化 fallback；它不改變地圖、規則或存檔資料。
func (m *Modern) SetOriginalOrnaments(o *OriginalOrnaments) {
	if m != nil && o.Valid() {
		m.ornaments = o
	}
}

// NewModern 建立完整的 P1 主題資產組。所有圖塊在建構時一次生成，切換時
// 只換 Theme 指標，不在畫面更新期間重新解碼或部分載入。
func NewModern() *Modern {
	pal := modernPalette()
	m := &Modern{
		tiles:     make([]Bitmap, modernTileCount),
		rails:     make([]Bitmap, modernRailCount),
		units:     make([]Bitmap, modernUnitCount),
		resources: make([]Bitmap, ResourceIconCount),
		commands:  make([]Bitmap, CommandIconCount),
	}
	for i := range m.tiles {
		m.tiles[i] = Bitmap{Image: modernTerrain(i), Palette: pal}
	}
	for i := range m.rails {
		m.rails[i] = Bitmap{Image: modernRail(i), Palette: pal}
	}
	for i := range m.units {
		m.units[i] = Bitmap{Image: modernUnit(i), Palette: pal}
	}
	for i := range m.resources {
		m.resources[i] = Bitmap{Image: modernResourceIcon(i), Palette: pal}
	}
	for i := range m.commands {
		m.commands[i] = Bitmap{Image: modernCommandIcon(i), Palette: pal}
	}
	return m
}

func (m *Modern) Name() string { return string(ModeModern) }

// Style 回傳與 modern 圖像同一批次建立的語意外殼。原版致敬（SPEC-47）
// 直接採用 RetroStyle 的米黃紙／暗紅墨／寶藍焦點為基準；戰棋土黃與銅章退場。
// 十個勢力色只作資訊色帶／焦點標記，不改變原版勢力編號、外交規則或存檔內容。
func (m *Modern) Style() UIStyle {
	return UIStyle{
		Name:      ModeModern,
		Ink:       assets.RGB{R: 0x2B, G: 0x26, B: 0x20},
		Paper:     assets.RGB{R: 0xE8, G: 0xE2, B: 0xD4},
		Panel:     assets.RGB{R: 0xD6, G: 0xCD, B: 0xB4},
		Muted:     assets.RGB{R: 0x5A, G: 0x44, B: 0x32},
		Accent:    assets.RGB{R: 0x1F, G: 0x3A, B: 0x5F},
		AccentAlt: assets.RGB{R: 0xB0, G: 0x2A, B: 0x20},
		Focus:     assets.RGB{R: 0x1F, G: 0x3A, B: 0x5F},
		FactionTint: [10]assets.RGB{
			{R: 176, G: 42, B: 32}, {R: 31, G: 58, B: 95},
			{R: 110, G: 107, B: 88}, {R: 200, G: 154, B: 46},
			{R: 122, G: 138, B: 153}, {R: 184, G: 154, B: 106},
			{R: 160, G: 154, B: 134}, {R: 207, G: 199, B: 178},
			{R: 90, G: 68, B: 50}, {R: 58, G: 53, B: 46},
		},
		Ornaments: m.ornaments,
	}
}

func (m *Modern) Tile(index int) (Bitmap, error) {
	if index < 0 || index >= len(m.tiles) {
		return Bitmap{}, fmt.Errorf("theme: modern 地形索引 %d 超出 0..%d", index, len(m.tiles)-1)
	}
	return m.tiles[index], nil
}

func (m *Modern) Rail(index int) (Bitmap, error) {
	if index < 0 || index >= len(m.rails) {
		return Bitmap{}, fmt.Errorf("theme: modern 鐵路索引 %d 超出 0..%d", index, len(m.rails)-1)
	}
	return m.rails[index], nil
}

// Unit 回傳與 NEWICON.TPC 相同的 0-based 圖示索引。索引對照與攻守／朝向
// 語意仍由 render.BranchIcon 保留，modern provider 不重新發明兵種規則。
func (m *Modern) Unit(index int) (Bitmap, error) {
	if index < 0 || index >= len(m.units) {
		return Bitmap{}, fmt.Errorf("theme: modern 部隊圖示索引 %d 超出 0..%d", index, len(m.units)-1)
	}
	return m.units[index], nil
}

// ResourceIcon 回傳固定順序的六種資源輔助圖示。
func (m *Modern) ResourceIcon(index int) (Bitmap, error) {
	if index < 0 || index >= len(m.resources) {
		return Bitmap{}, fmt.Errorf("theme: modern 資源圖示索引 %d 超出 0..%d", index, len(m.resources)-1)
	}
	return m.resources[index], nil
}

// CommandIcon 回傳政略指令 1..15 的輔助圖示；index 仍是 0-based。
func (m *Modern) CommandIcon(index int) (Bitmap, error) {
	if index < 0 || index >= len(m.commands) {
		return Bitmap{}, fmt.Errorf("theme: modern 指令圖示索引 %d 超出 0..%d", index, len(m.commands)-1)
	}
	return m.commands[index], nil
}

// modernPalette 是 A2 明亮土黃介面的程式化地圖色組，並保留第 0 格作鐵路
// 透明索引。所有顏色是可再散布的程式常數，不是原版調色盤或衛星照片。
func modernPalette() assets.Palette {
	// 老照片套印 18 色（SPEC-49 §1，設計師定稿）：銀鹽暖灰為底、
	// 朱紅為攻、藏青為守。索引角色沿用舊盤，色值整套換。
	return assets.Palette{
		{R: 0, G: 0, B: 0},          // 0：透明（鐵路）
		{R: 0xE8, G: 0xE2, B: 0xD4}, // 1：銀鹽紙
		{R: 0x2B, G: 0x26, B: 0x20}, // 2：暖墨
		{R: 0xB0, G: 0x2A, B: 0x20}, // 3：朱紅／攻方
		{R: 0x1F, G: 0x3A, B: 0x5F}, // 4：藏青／守方
		{R: 0x6E, G: 0x6B, B: 0x58}, // 5：灰苔
		{R: 0xC8, G: 0x9A, B: 0x2E}, // 6：赭黃
		{R: 0xD6, G: 0xCD, B: 0xB4}, // 7：淺灰地
		{R: 0x7A, G: 0x8A, B: 0x99}, // 8：銀灰藍
		{R: 0x8A, G: 0x81, B: 0x78}, // 9：中山灰
		{R: 0xB8, G: 0x9A, B: 0x6A}, // 10：sepia 赭
		{R: 0xA0, G: 0x9A, B: 0x86}, // 11：丘陵灰
		{R: 0xCF, G: 0xC7, B: 0xB2}, // 12：壁灰
		{R: 0x5A, G: 0x44, B: 0x32}, // 13：焦赭
		{R: 0xEF, G: 0xE9, B: 0xDB}, // 14：雪灰白
		{R: 0x3A, G: 0x35, B: 0x2E}, // 15：鐵灰
		{R: 0x4A, G: 0x42, B: 0x38}, // 16：深框
		{R: 0x7D, G: 0x74, B: 0x66}, // 17：細線灰
	}
}

func modernTerrain(kind int) *assets.Image {
	base, accent := terrainColors(kind)
	pix := make([]byte, modernTileW*modernTileH)
	for y := 0; y < modernTileH; y++ {
		for x := 0; x < modernTileW; x++ {
			// 四邊維持同一底色，讓相同地物相鄰時不出現程式化接縫。
			v := base
			if x > 1 && x < modernTileW-2 && y > 1 && y < modernTileH-2 {
				if modernPattern(kind, x, y) {
					v = accent
				}
				// 銀鹽顆粒（SPEC-49 M2）：確定性雜訊，中央留白，種類參與雜湊。
				if photoGrain(kind, x, y) {
					if v == base {
						v = 17 // 細線灰顆粒
					} else {
						v = base
					}
				}
			}
			pix[y*modernTileW+x] = v
		}
	}
	return &assets.Image{W: modernTileW, H: modernTileH, Pix: pix}
}

// modernUnit 以純 Go 產生 P2 的 18 張新圖示。尺寸、透明像素與索引順序
// 沿用 NEWICON.TPC；線條是新作，不是原版像素的轉存。前六張是步兵／
// 裝甲／騎兵的綠紅對，後十二張是綠紅砲兵的六個朝向。
func modernUnit(index int) *assets.Image {
	pix := make([]byte, modernUnitW*modernUnitH)
	red := false
	kind := 0
	facing := 1
	switch {
	case index < 2:
		kind = 0 // infantry
		red = index == 1
	case index < 4:
		kind = 1 // armour
		red = index == 3
	case index < 6:
		kind = 2 // cavalry
		red = index == 5
	default:
		kind = 3 // artillery
		red = index >= 12
		facing = (index % 6) + 1
	}
	team := byte(3) // 朱紅：攻方（SPEC-49 §1）
	if red {
		team = 4 // 藏青：守方（SPEC-49 §1）
	}
	ink := byte(2)
	// 原版致敬剪影（SPEC-47 M2）：照 NEWICON.TPC 的無框平面剪影重繪——
	// 步兵盔帽半身、裝甲側視戰車、騎兵馬頭、砲兵炮身，勢力色為體、墨色為影。
	// 32×17 尺寸、透明索引 0、18 張索引順序、砲兵六朝向一律保留。
	switch kind {
	case 0:
		// 步兵：盔帽半身像（滿框）。盔體勢力色、帽帶與衣褶墨線。
		iconRect(pix, 4, 2, 27, 7, team)
		iconLine(pix, 4, 5, 27, 5, ink)
		iconRect(pix, 2, 8, 29, 9, team)
		iconRect(pix, 3, 11, 9, 14, team)
		iconRect(pix, 22, 11, 28, 14, team)
		iconRect(pix, 8, 10, 23, 16, team)
		iconLine(pix, 14, 10, 16, 12, ink)
		iconLine(pix, 18, 10, 16, 12, ink)
		iconLine(pix, 12, 11, 12, 15, ink)
		iconLine(pix, 19, 11, 19, 15, ink)
	case 1:
		// 裝甲：側視戰車（滿框）。車體勢力色、履帶墨色鑲勢力色輪。
		iconRect(pix, 20, 5, 30, 6, team)
		iconRect(pix, 10, 4, 20, 8, team)
		iconLine(pix, 10, 4, 20, 4, ink)
		iconRect(pix, 2, 8, 29, 12, team)
		iconLine(pix, 2, 9, 29, 9, ink)
		iconRect(pix, 3, 13, 28, 15, ink)
		for _, x := range []int{6, 11, 16, 21, 26} {
			iconCircle(pix, x, 14, 1, team)
		}
	case 2:
		// 騎兵：馬頭剪影（滿框）。頭頸勢力色、鬃溝與眼口墨線。
		iconRect(pix, 6, 7, 18, 16, team)
		iconRect(pix, 3, 12, 10, 16, team)
		iconRect(pix, 15, 1, 27, 8, team)
		iconRect(pix, 15, 0, 18, 3, team)
		iconRect(pix, 24, 6, 29, 10, team)
		iconLine(pix, 9, 9, 9, 15, ink)
		iconLine(pix, 12, 9, 12, 15, ink)
		iconLine(pix, 16, 8, 24, 8, ink)
		iconPixel(pix, 22, 4, ink)
		iconLine(pix, 25, 9, 28, 9, ink)
	case 3:
		// 砲兵：炮身＋雙輪側視（滿框），方向依原版 +31 的六值保留。
		iconRect(pix, 6, 8, 25, 10, team)
		iconCircle(pix, 8, 12, 3, team)
		iconCircle(pix, 23, 12, 3, team)
		iconCircle(pix, 8, 12, 1, ink)
		iconCircle(pix, 23, 12, 1, ink)
		iconPixel(pix, 8, 12, team)
		iconPixel(pix, 23, 12, team)
		iconLine(pix, 6, 11, 25, 11, ink)
		dirs := [...]struct{ dx, dy int }{
			{0, -1}, {1, -1}, {1, 0}, {0, 1}, {-1, 1}, {-1, 0},
		}
		d := dirs[facing-1]
		ex, ey := 16+d.dx*12, 8+d.dy*6
		iconLine(pix, 16, 8, ex, ey, team)
		iconLine(pix, 16, 9, ex, ey+1, team)
	}
	return &assets.Image{W: modernUnitW, H: modernUnitH, Pix: pix}
}

// photoGrain 是地形銀鹽顆粒的確定性雜訊（SPEC-49 M2）：約 1/23 的內部
// 像素染成細線灰（底色區）或打回底色（紋理區），模擬照片顆粒；
// 中央 4×4 留白，種類參與雜湊使各地形顆粒分佈不同。
func photoGrain(kind, x, y int) bool {
	if x >= modernTileW/2-2 && x < modernTileW/2+2 && y >= modernTileH/2-2 && y < modernTileH/2+2 {
		return false
	}
	return (x*31+y*17+kind*7)%23 == 0
}

func iconRect(pix []byte, x0, y0, x1, y1 int, value byte) {
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			iconPixel(pix, x, y, value)
		}
	}
}

func iconCircle(pix []byte, cx, cy, radius int, value byte) {
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy <= radius*radius {
				iconPixel(pix, x, y, value)
			}
		}
	}
}

func iconLine(pix []byte, x0, y0, x1, y1 int, value byte) {
	dx, sx := absInt(x1-x0), 1
	if x0 > x1 {
		sx = -1
	}
	dy, sy := -absInt(y1-y0), 1
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		iconPixel(pix, x0, y0, value)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func iconPixel(pix []byte, x, y int, value byte) {
	if x < 0 || x >= modernUnitW || y < 0 || y >= modernUnitH {
		return
	}
	pix[y*modernUnitW+x] = value
}

func absInt(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

// modernResourceIcon 以 16×16 的簡單幾何保留資源類別差異。圖示是 HUD 輔助，
// 不承擔資源數值或規則語意；索引 0 永遠透明，讓畫布底色透出。
func modernResourceIcon(index int) *assets.Image {
	pix := make([]byte, HUDIconW*HUDIconH)
	ink, accent, light := byte(2), byte(3+index%8), byte(1)
	switch index {
	case 0: // 黃金：圓幣
		hudCircle(pix, 8, 8, 6, accent)
		hudCircle(pix, 8, 8, 3, light)
		hudLine(pix, 8, 5, 8, 11, ink)
	case 1: // 糧食：穀穗
		hudLine(pix, 8, 13, 8, 3, ink)
		for y := 5; y <= 11; y += 2 {
			hudLine(pix, 8, y, 4, y-2, accent)
			hudLine(pix, 8, y+1, 12, y-1, accent)
		}
	case 2: // 彈藥：彈匣／子彈
		hudRect(pix, 5, 3, 10, 13, accent)
		hudRect(pix, 7, 5, 8, 11, light)
		hudLine(pix, 11, 4, 11, 13, ink)
	case 3: // 燃料：油滴
		hudLine(pix, 8, 2, 3, 9, accent)
		hudCircle(pix, 8, 10, 5, accent)
		hudCircle(pix, 8, 10, 2, light)
	case 4: // 煤礦：不規則礦塊
		hudPolygon(pix, [][2]int{{3, 11}, {5, 4}, {11, 3}, {14, 8}, {12, 14}, {5, 14}}, accent)
		hudLine(pix, 5, 9, 12, 6, light)
		hudLine(pix, 7, 12, 11, 10, ink)
	case 5: // 鐵礦：錠塊
		hudPolygon(pix, [][2]int{{3, 6}, {12, 6}, {14, 10}, {11, 14}, {4, 13}}, accent)
		hudLine(pix, 4, 8, 13, 10, light)
		hudLine(pix, 6, 12, 11, 12, ink)
	default:
		return &assets.Image{W: HUDIconW, H: HUDIconH, Pix: pix}
	}
	return &assets.Image{W: HUDIconW, H: HUDIconH, Pix: pix}
}

// modernCommandIcon 用十五個穩定但不含文字的幾何符號提示指令類別。
// 民國戰棋收斂：十五張共用圓角銅框算子底（框線後畫，輪廓統一），
// 內部符號沿用既有十五組幾何，不重繪。
func modernCommandIcon(index int) *assets.Image {
	pix := make([]byte, HUDIconW*HUDIconH)
	ink, accent, light := byte(2), commandAccentIndex(index), byte(1)
	hudCircle(pix, 8, 8, 6, accent)
	hudCircle(pix, 8, 8, 4, ink)
	drawCommandMarkLow(pix, index, light, accent)
	hudCommandCounterFrame(pix)
	return &assets.Image{W: HUDIconW, H: HUDIconH, Pix: pix}
}

// hudCommandCounterFrame 在 16×16 指令圖上以後畫方式罩上圓角銅框。
// 四角留透明，框線統一 15 張的剪影；被罩掉的符號邊緣像素至多 1 px。
func hudCommandCounterFrame(pix []byte) {
	copper, bronze := byte(16), byte(17)
	hudLine(pix, 1, 0, 14, 0, copper)
	hudLine(pix, 1, 15, 14, 15, copper)
	hudLine(pix, 0, 1, 0, 14, copper)
	hudLine(pix, 15, 1, 15, 14, copper)
	hudLine(pix, 2, 1, 13, 1, bronze)
	hudLine(pix, 2, 14, 13, 14, bronze)
	hudLine(pix, 1, 2, 1, 13, bronze)
	hudLine(pix, 14, 2, 14, 13, bronze)
}

func commandAccentIndex(index int) byte {
	switch index {
	case 0, 3, 6, 11:
		return 6 // 赭金：調動／稅務／開發／商業
	case 1, 4, 9, 12:
		return 3 // 朱紅：軍務／徵兵／停火／練兵
	case 2, 13:
		return 5 // 橄欖綠：運補／慰勞
	default:
		return 15 // 黛青：情報／政務／外交
	}
}

func drawCommandMarkLow(pix []byte, index int, light, accent byte) {
	switch index {
	case 0: // 調動：兩省定位點與雙向轉移箭頭
		hudCircle(pix, 4, 8, 2, light)
		hudCircle(pix, 12, 8, 2, light)
		hudLine(pix, 6, 7, 10, 7, light)
		hudLine(pix, 10, 9, 6, 9, light)
		hudLine(pix, 10, 7, 9, 6, light)
		hudLine(pix, 10, 7, 9, 8, light)
		hudLine(pix, 6, 9, 7, 8, light)
		hudLine(pix, 6, 9, 7, 10, light)
	case 1: // 軍事：軍旗與交叉軍刀
		hudLine(pix, 5, 3, 5, 13, light)
		hudPolygon(pix, [][2]int{{5, 3}, {12, 4}, {9, 7}}, light)
		hudLine(pix, 6, 5, 12, 12, accent)
		hudLine(pix, 12, 5, 6, 12, accent)
	case 2: // 運補：箱體、道路、方向箭頭與封印
		hudRect(pix, 4, 7, 9, 11, light)
		hudLine(pix, 5, 6, 8, 6, light)
		hudLine(pix, 9, 9, 13, 9, light)
		hudLine(pix, 11, 7, 13, 9, light)
		hudLine(pix, 11, 11, 13, 9, light)
		hudCircle(pix, 5, 12, 1, accent)
	case 3: // 徵稅：帳冊、錢幣與收訖印
		hudRect(pix, 4, 4, 9, 12, light)
		hudLine(pix, 6, 6, 8, 6, accent)
		hudLine(pix, 6, 8, 8, 8, accent)
		hudCircle(pix, 11, 8, 3, light)
		hudLine(pix, 11, 6, 11, 10, accent)
		hudLine(pix, 10, 8, 12, 8, accent)
		hudCircle(pix, 12, 12, 2, accent)
	case 4: // 徵兵：鋼盔、人形與招募加號
		hudPolygon(pix, [][2]int{{5, 6}, {6, 4}, {10, 4}, {11, 6}}, light)
		hudRect(pix, 6, 7, 10, 12, light)
		hudLine(pix, 4, 10, 12, 10, accent)
		hudLine(pix, 13, 8, 13, 12, light)
		hudLine(pix, 11, 10, 15, 10, light)
	case 5: // 查閱：摺頁地圖與放大鏡
		hudPolygon(pix, [][2]int{{3, 5}, {6, 4}, {10, 5}, {13, 4}, {13, 11}, {10, 12}, {6, 11}, {3, 12}}, light)
		hudLine(pix, 6, 4, 6, 11, accent)
		hudLine(pix, 10, 5, 10, 12, accent)
		hudCircle(pix, 11, 9, 2, light)
		hudLine(pix, 13, 11, 15, 13, light)
	case 6: // 開發：藍圖、磚塊與十字鎬
		hudRect(pix, 3, 4, 10, 11, light)
		hudLine(pix, 5, 6, 9, 6, accent)
		hudLine(pix, 5, 8, 9, 8, accent)
		hudRect(pix, 4, 12, 6, 14, accent)
		hudRect(pix, 8, 12, 10, 14, accent)
		hudLine(pix, 11, 12, 14, 5, light)
		hudLine(pix, 10, 7, 14, 5, light)
	case 7: // 政策：蓋章公文
		hudRect(pix, 4, 4, 11, 12, light)
		hudLine(pix, 6, 6, 9, 6, accent)
		hudLine(pix, 6, 8, 10, 8, accent)
		hudLine(pix, 6, 10, 9, 10, accent)
		hudCircle(pix, 11, 11, 2, accent)
		hudLine(pix, 10, 11, 12, 11, light)
		hudLine(pix, 11, 10, 11, 12, light)
	case 8: // 外交：兩旗與握手
		hudLine(pix, 4, 3, 4, 12, light)
		hudLine(pix, 12, 3, 12, 12, light)
		hudPolygon(pix, [][2]int{{4, 3}, {8, 5}, {4, 7}}, light)
		hudPolygon(pix, [][2]int{{12, 3}, {8, 5}, {12, 7}}, light)
		hudLine(pix, 5, 9, 8, 8, accent)
		hudLine(pix, 11, 9, 8, 8, accent)
		hudLine(pix, 7, 9, 9, 9, accent)
	case 9: // 談判停火：交叉軍刀被停火條覆蓋
		hudLine(pix, 5, 4, 11, 12, light)
		hudLine(pix, 11, 4, 5, 12, light)
		hudRect(pix, 3, 7, 13, 9, accent)
		hudLine(pix, 2, 8, 4, 8, light)
		hudLine(pix, 12, 8, 14, 8, light)
	case 10: // 秘密行動：眼睛、虛線路徑與遮蔽角
		hudPolygon(pix, [][2]int{{3, 8}, {6, 5}, {10, 5}, {13, 8}, {10, 11}, {6, 11}}, light)
		hudCircle(pix, 8, 8, 2, accent)
		for _, x := range []int{3, 6, 9, 12} {
			hudLine(pix, x, 13, x+1, 13, light)
		}
		hudPolygon(pix, [][2]int{{11, 11}, {14, 11}, {14, 14}}, accent)
	case 11: // 商業：商店櫃台與雙向交換箭頭
		hudPolygon(pix, [][2]int{{3, 6}, {8, 3}, {13, 6}}, light)
		hudRect(pix, 4, 7, 12, 12, light)
		hudRect(pix, 6, 9, 8, 11, accent)
		hudLine(pix, 3, 13, 13, 13, light)
		hudLine(pix, 5, 5, 11, 5, accent)
		hudLine(pix, 10, 4, 12, 5, accent)
	case 12: // 練兵：三列隊形、旗標與靶心
		for _, x := range []int{4, 8, 12} {
			hudCircle(pix, x, 6, 1, light)
			hudLine(pix, x, 8, x, 11, light)
			hudLine(pix, x-1, 9, x+1, 9, light)
		}
		hudLine(pix, 3, 3, 3, 13, accent)
		hudPolygon(pix, [][2]int{{3, 3}, {7, 4}, {3, 6}}, accent)
		hudCircle(pix, 11, 12, 2, light)
		hudCircle(pix, 11, 12, 1, accent)
	case 13: // 慰勞軍民：慰問桌、兩個人形與心形
		hudCircle(pix, 5, 5, 1, light)
		hudCircle(pix, 11, 4, 1, light)
		hudLine(pix, 5, 7, 5, 10, light)
		hudLine(pix, 11, 6, 11, 10, light)
		hudLine(pix, 3, 11, 13, 11, light)
		hudLine(pix, 6, 12, 6, 14, accent)
		hudLine(pix, 10, 12, 10, 14, accent)
		hudPolygon(pix, [][2]int{{8, 6}, {9, 5}, {10, 6}, {8, 9}, {6, 6}, {7, 5}}, accent)
	case 14: // 其他：三列設定方格與右下三點
		for _, y := range []int{5, 8, 11} {
			hudRect(pix, 4, y, 11, y+1, light)
			hudCircle(pix, 7+(y%3), y, 1, accent)
		}
		hudCircle(pix, 12, 13, 1, light)
		hudCircle(pix, 14, 13, 1, light)
		hudCircle(pix, 10, 13, 1, light)
	}
}

func hudPixel(pix []byte, x, y int, value byte) {
	if x < 0 || x >= HUDIconW || y < 0 || y >= HUDIconH {
		return
	}
	pix[y*HUDIconW+x] = value
}

func hudRect(pix []byte, x0, y0, x1, y1 int, value byte) {
	for y := y0; y <= y1; y++ {
		for x := x0; x <= x1; x++ {
			hudPixel(pix, x, y, value)
		}
	}
}

func hudCircle(pix []byte, cx, cy, radius int, value byte) {
	for y := cy - radius; y <= cy+radius; y++ {
		for x := cx - radius; x <= cx+radius; x++ {
			dx, dy := x-cx, y-cy
			if dx*dx+dy*dy <= radius*radius {
				hudPixel(pix, x, y, value)
			}
		}
	}
}

func hudLine(pix []byte, x0, y0, x1, y1 int, value byte) {
	dx, sx := absInt(x1-x0), 1
	if x0 > x1 {
		sx = -1
	}
	dy, sy := -absInt(y1-y0), 1
	if y0 > y1 {
		sy = -1
	}
	err := dx + dy
	for {
		hudPixel(pix, x0, y0, value)
		if x0 == x1 && y0 == y1 {
			return
		}
		e2 := 2 * err
		if e2 >= dy {
			err += dy
			x0 += sx
		}
		if e2 <= dx {
			err += dx
			y0 += sy
		}
	}
}

func hudPolygon(pix []byte, points [][2]int, value byte) {
	if len(points) < 2 {
		return
	}
	for i := range points {
		next := points[(i+1)%len(points)]
		hudLine(pix, points[i][0], points[i][1], next[0], next[1], value)
	}
}

func terrainColors(kind int) (base, accent byte) {
	switch {
	case kind == 2: // 河海
		return 8, 4
	case kind == 3: // 森林
		return 5, 2
	case kind == 5: // 高山
		return 9, 13
	case kind == 6: // 沙漠
		return 10, 6
	case kind == 9: // 高原
		return 12, 11
	case kind == 10: // 關口
		return 13, 2
	case kind >= 11: // 長城各段
		return 12, 13
	case kind == 1: // 丘陵
		return 11, 5
	case kind == 4: // 城市
		return 7, 12
	default: // 平原與兩種橋面
		return 7, 1
	}
}

func modernPattern(kind, x, y int) bool {
	// 原版致敬（SPEC-47 M2）：照 MAP1.TPC 的濃密 BGI 肌理重繪——水波紋、
	// 樹簇、雪峰、沙丘弧皆採原版式密集重複，取代測繪期的低雜訊留白。
	switch kind {
	case 2: // 水波：原版式雙排波紋。
		return (y%6 == 2 && x%8 >= 1 && x%8 <= 5) || (y%6 == 5 && x%8 >= 3 && x%8 <= 7)
	case 3: // 森林：三簇樹冠＋樹幹點。
		px, py := x%11, y%9
		return (px >= 1 && px <= 3 && py >= 1 && py <= 6) ||
			(px >= 5 && px <= 7 && py >= 2 && py <= 7) ||
			(px >= 8 && px <= 10 && py >= 0 && py <= 4) ||
			(py == 8 && (px == 2 || px == 6 || px == 9))
	case 5: // 高山：赭紅山脊雪峰輪廓。
		px, py := x%16, y%12
		return py >= 2 && py <= 9 && absInt(px-8) <= py/2
	case 9: // 高原：密脊線。
		return (y%6 == 3 && x%12 >= 1 && x%12 <= 10) || (y%6 == 4 && x%2 == 0)
	case 10: // 關口：窄門與兩翼岩壁。
		return (x >= 13 && x <= 18) || (y >= 8 && y <= 10 && (x < 8 || x > 23))
	case 6: // 沙漠：密沙丘弧。
		return (y%6 == 4 && x%10 >= 2 && x%10 <= 8) || (y%6 == 5 && x%2 == 0)
	case 4: // 城市：方整街區，與地圖資料的城市語意一致。
		return (x%10 >= 3 && x%10 <= 6 && y%8 >= 2 && y%8 <= 5) || (x%10 == 7 && y%8 == 6) ||
			(x%10 <= 1 && y%8 <= 1)
	case 7: // 縱橋
		return x >= 14 && x <= 17
	case 8: // 橫橋
		return y >= 10 && y <= 13
	default: // 長城／平原：原版式有序抖點。
		return (x*11+y*7+kind*5)%23 == 0
	}
}

func modernRail(index int) *assets.Image {
	pix := make([]byte, assets.RailW*assets.RailH)
	ink := byte(15)
	for y := 0; y < assets.RailH; y++ {
		for x := 0; x < assets.RailW; x++ {
			if railMark(index, x, y) {
				pix[y*assets.RailW+x] = ink
			}
		}
	}
	return &assets.Image{W: assets.RailW, H: assets.RailH, Pix: pix}
}

func railMark(index, x, y int) bool {
	// 這個表是依 RAIL.TPC 的 21 張原始圖塊逐張檢查邊緣接線所得。
	// index 0 是有效的縱線；透明語意只存在於圖塊內的像素索引 0。
	mask := [...]uint8{
		0b0101,         // 0：上、下
		0b1010, 0b1010, // 1..2：左、右
		0b1001, 0b1001, // 3..4：左、下
		0b0011, 0b0011, // 5..6：右、下
		0b1100, 0b1100, // 7..8：左、上
		0b0110, 0b0110, // 9..10：右、上
		0b1011, 0b1011, // 11..12：左、右、下
		0b1110, 0b1110, // 13..14：左、右、上
		0b0111, 0b0111, // 15..16：右、上、下
		0b1101, 0b1101, // 17..18：左、上、下
		0b1111, 0b1111, // 19..20：十字
	}
	if index < 0 || index >= len(mask) {
		return false
	}
	cx, cy := assets.RailW/2, assets.RailH/2
	const (
		top = 1 << iota
		right
		bottom
		left
	)
	m := mask[index]
	// 保持原版的四向接線，modern 只把筆畫換成清楚的雙像素線。
	if m&top != 0 && x >= cx-1 && x <= cx && y <= cy {
		return true
	}
	if m&bottom != 0 && x >= cx-1 && x <= cx && y >= cy {
		return true
	}
	if m&left != 0 && y >= cy-1 && y <= cy && x <= cx {
		return true
	}
	if m&right != 0 && y >= cy-1 && y <= cy && x >= cx {
		return true
	}
	return false
}
