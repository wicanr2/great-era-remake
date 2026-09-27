package theme

import "testing"

func TestDecodeOriginalOrnamentsIsAtomicAndOptional(t *testing.T) {
	// 1×1、四個 bit-plane 的最小合法 BGI；像素內容不重要，這裡只鎖定
	// 六個 runtime 原版遮罩必須同時成功，不能留下半套 theme。
	bgi := []byte{0, 0, 0, 0, 0, 0, 0, 0}
	read := func(name string) ([]byte, error) { return append([]byte(nil), bgi...), nil }
	o, err := DecodeOriginalOrnaments(read)
	if err != nil || !o.Valid() {
		t.Fatalf("合法裝飾組未載入：o=%+v err=%v", o, err)
	}
	m := NewModern()
	m.SetOriginalOrnaments(o)
	if m.Style().Ornaments == nil || !m.Style().Ornaments.Valid() {
		t.Fatal("A2 style 沒有取得原版 runtime 裝飾組")
	}
	if _, err := DecodeOriginalOrnaments(func(name string) ([]byte, error) { return []byte{0}, nil }); err == nil {
		t.Fatal("損壞 BGI 不應留下半套裝飾")
	}
}

func TestModernThemeHasCompleteTerrainAndRailIndexSets(t *testing.T) {
	m := NewModern()
	for i := 0; i < modernTileCount; i++ {
		b, err := m.Tile(i)
		if err != nil {
			t.Fatalf("地形 %d: %v", i, err)
		}
		if !b.Valid() || b.Image.W != modernTileW || b.Image.H != modernTileH {
			t.Fatalf("地形 %d bitmap=%+v", i, b)
		}
	}
	for i := 0; i < modernRailCount; i++ {
		b, err := m.Rail(i)
		if err != nil {
			t.Fatalf("鐵路 %d: %v", i, err)
		}
		if !b.Valid() || b.Image.W != 32 || b.Image.H != 24 {
			t.Fatalf("鐵路 %d bitmap=%+v", i, b)
		}
	}
	for i := 0; i < modernUnitCount; i++ {
		b, err := m.Unit(i)
		if err != nil {
			t.Fatalf("部隊圖示 %d: %v", i, err)
		}
		if !b.Valid() || b.Image.W != modernUnitW || b.Image.H != modernUnitH {
			t.Fatalf("部隊圖示 %d bitmap=%+v", i, b)
		}
	}
	if _, err := m.Tile(-1); err == nil {
		t.Fatal("負地形索引應拒絕")
	}
	if _, err := m.Rail(modernRailCount); err == nil {
		t.Fatal("超界鐵路索引應拒絕")
	}
	if _, err := m.Unit(modernUnitCount); err == nil {
		t.Fatal("超界部隊圖示索引應拒絕")
	}
	for i := 0; i < ResourceIconCount; i++ {
		b, err := m.ResourceIcon(i)
		if err != nil {
			t.Fatalf("資源圖示 %d: %v", i, err)
		}
		if !b.Valid() || b.Image.W != HUDIconW || b.Image.H != HUDIconH {
			t.Fatalf("資源圖示 %d bitmap=%+v", i, b)
		}
	}
	for i := 0; i < CommandIconCount; i++ {
		b, err := m.CommandIcon(i)
		if err != nil {
			t.Fatalf("指令圖示 %d: %v", i, err)
		}
		if !b.Valid() || b.Image.W != HUDIconW || b.Image.H != HUDIconH {
			t.Fatalf("指令圖示 %d bitmap=%+v", i, b)
		}
	}
	if _, err := m.ResourceIcon(-1); err == nil {
		t.Fatal("負資源圖示索引應拒絕")
	}
	if _, err := m.CommandIcon(CommandIconCount); err == nil {
		t.Fatal("超界指令圖示索引應拒絕")
	}
}

func TestModernHighResolutionProviderPreservesIndexesAndSizes(t *testing.T) {
	m := NewModern()
	var _ HighResolutionTheme = m
	var _ HighResolutionUnitProvider = m
	for i := 0; i < modernTileCount; i++ {
		im, err := m.HighTile(i, 48, 36)
		if err != nil || im.Bounds().Dx() != 48 || im.Bounds().Dy() != 36 {
			t.Fatalf("地形 %d 高解析尺寸：%v %v", i, im, err)
		}
	}
	for i := 0; i < modernRailCount; i++ {
		im, err := m.HighRail(i, 48, 36)
		if err != nil || im.Bounds().Dx() != 48 || im.Bounds().Dy() != 36 {
			t.Fatalf("鐵路 %d 高解析尺寸：%v %v", i, im, err)
		}
	}
	for i := 0; i < modernUnitCount; i++ {
		im, err := m.HighUnit(i, 48, 26)
		if err != nil || im.Bounds().Dx() != 48 || im.Bounds().Dy() != 26 {
			t.Fatalf("部隊 %d 高解析尺寸：%v %v", i, im, err)
		}
	}
}

func TestModernHighResolutionIsDeterministicAndKeepsRailAlpha(t *testing.T) {
	a, b := NewModern(), NewModern()
	aa, err := a.HighUnit(13, 47, 29)
	if err != nil {
		t.Fatal(err)
	}
	bb, err := b.HighUnit(13, 47, 29)
	if err != nil {
		t.Fatal(err)
	}
	if string(aa.Pix) != string(bb.Pix) {
		t.Fatal("高解析部隊圖示不 deterministic")
	}
	rail, err := a.HighRail(0, 48, 36)
	if err != nil {
		t.Fatal(err)
	}
	transparent := false
	for i := 3; i < len(rail.Pix); i += 4 {
		if rail.Pix[i] < 255 {
			transparent = true
			break
		}
	}
	if !transparent {
		t.Fatal("鐵路透明索引沒有保留 alpha")
	}
	if _, err := a.HighTile(-1, 48, 36); err == nil {
		t.Fatal("高解析地形負索引應拒絕")
	}
	if _, err := a.HighRail(modernRailCount, 48, 36); err == nil {
		t.Fatal("高解析鐵路超界索引應拒絕")
	}
	if _, err := a.HighUnit(modernUnitCount, 48, 26); err == nil {
		t.Fatal("高解析部隊超界索引應拒絕")
	}
}

func TestModernHighResolutionDoesNotReadLowResolutionBitmaps(t *testing.T) {
	m := NewModern()
	before, err := m.HighTile(0, 48, 36)
	if err != nil {
		t.Fatal(err)
	}
	for i := range m.tiles[0].Image.Pix {
		m.tiles[0].Image.Pix[i] = byte((i + 7) % 16)
	}
	after, err := m.HighTile(0, 48, 36)
	if err != nil {
		t.Fatal(err)
	}
	if string(before.Pix) != string(after.Pix) {
		t.Fatal("高解析地形不應依賴低解析 bitmap")
	}
	beforeUnit, err := m.HighUnit(0, 48, 26)
	if err != nil {
		t.Fatal(err)
	}
	for i := range m.units[0].Image.Pix {
		m.units[0].Image.Pix[i] = byte((i + 3) % 16)
	}
	afterUnit, err := m.HighUnit(0, 48, 26)
	if err != nil {
		t.Fatal(err)
	}
	if string(beforeUnit.Pix) != string(afterUnit.Pix) {
		t.Fatal("高解析部隊不應依賴低解析 bitmap")
	}
}

func TestModernHighWallVariantsAndUnknownFallbackAreDistinct(t *testing.T) {
	m := NewModern()
	seen := map[string]bool{}
	for kind := 11; kind <= 20; kind++ {
		im, err := m.HighTile(kind, 48, 36)
		if err != nil {
			t.Fatal(err)
		}
		key := string(im.Pix)
		if seen[key] {
			t.Fatalf("長城變體 %d 與前一變體完全相同", kind)
		}
		seen[key] = true
	}
	unknown, err := m.HighTile(21, 48, 36)
	if err != nil {
		t.Fatal(err)
	}
	// TileKind 22 is an unknown neutral card: its centre is paper, not a
	// wall line or gate, and it must differ from every formal wall variant.
	p := unknown.RGBAAt(24, 18)
	if p.R < 220 || p.G < 190 || p.B < 120 {
		t.Fatalf("unknown fallback 不是中性米黃：%+v", p)
	}
	if seen[string(unknown.Pix)] {
		t.Fatal("unknown fallback 不得重用長城圖形")
	}
}

func TestModernHUDIconsAreDeterministicAndDistinct(t *testing.T) {
	a, b := NewModern(), NewModern()
	for i := 0; i < ResourceIconCount; i++ {
		ai, err := a.ResourceIcon(i)
		if err != nil {
			t.Fatal(err)
		}
		bi, err := b.ResourceIcon(i)
		if err != nil {
			t.Fatal(err)
		}
		if string(ai.Image.Pix) != string(bi.Image.Pix) {
			t.Fatalf("資源圖示 %d 非 deterministic", i)
		}
	}
	seen := make(map[string]int, CommandIconCount)
	for i := 0; i < CommandIconCount; i++ {
		icon, err := a.CommandIcon(i)
		if err != nil {
			t.Fatal(err)
		}
		signature := string(icon.Image.Pix)
		if previous, ok := seen[signature]; ok {
			t.Fatalf("指令圖示 %d 與 %d 共用相同 primitive 結果", previous+1, i+1)
		}
		seen[signature] = i
	}
}

func TestModernHighCommandIconsAreAllUniqueAndDeterministic(t *testing.T) {
	a, b := NewModern(), NewModern()
	seen := make(map[string]int, CommandIconCount)
	for i := 0; i < CommandIconCount; i++ {
		ai, err := a.CommandIconRGBA(i, 48, 48)
		if err != nil {
			t.Fatal(err)
		}
		bi, err := b.CommandIconRGBA(i, 48, 48)
		if err != nil {
			t.Fatal(err)
		}
		signature := string(ai.Pix)
		if previous, ok := seen[signature]; ok {
			t.Fatalf("高解析指令圖示 %d 與 %d 共用相同 primitive 結果", previous+1, i+1)
		}
		seen[signature] = i
		if signature != string(bi.Pix) {
			t.Fatalf("高解析指令圖示 %d 非 deterministic", i+1)
		}
	}
}

func TestModernHighResolutionHUDIconsUseRequestedSizeAndAlphaEdges(t *testing.T) {
	m := NewModern()
	command, err := m.CommandIconRGBA(0, 44, 44)
	if err != nil {
		t.Fatal(err)
	}
	resource, err := m.ResourceIconRGBA(0, 20, 20)
	if err != nil {
		t.Fatal(err)
	}
	if command.Bounds().Dx() != 44 || command.Bounds().Dy() != 44 ||
		resource.Bounds().Dx() != 20 || resource.Bounds().Dy() != 20 {
		t.Fatalf("目的尺寸圖示不符：command=%v resource=%v", command.Bounds(), resource.Bounds())
	}
	seenIntermediate := false
	for y := 0; y < 44; y++ {
		for x := 0; x < 44; x++ {
			a := command.RGBAAt(x, y).A
			if a > 0 && a < 255 {
				seenIntermediate = true
			}
		}
	}
	if !seenIntermediate {
		t.Fatal("高解析 HUD 圖示缺少 supersample alpha 邊緣")
	}
}

func TestModernRailZeroKeepsOriginalVerticalGeometry(t *testing.T) {
	m := NewModern()
	b, err := m.Rail(0)
	if err != nil {
		t.Fatal(err)
	}
	for y := 0; y < b.Image.H; y++ {
		for x := 0; x < b.Image.W; x++ {
			v := b.Image.Pix[y*b.Image.W+x]
			want := x == b.Image.W/2 || x == b.Image.W/2-1
			if (v != 0) != want {
				t.Fatalf("rail 0 (%d,%d)=%d，應保留縱線接法", x, y, v)
			}
		}
	}
}

func TestModernRailKeepsOriginalEdgeConnections(t *testing.T) {
	// bit 0=上、1=右、2=下、3=左；成對的原版索引仍各自保留。
	want := []uint8{
		0b0101,
		0b1010, 0b1010,
		0b1001, 0b1001,
		0b0011, 0b0011,
		0b1100, 0b1100,
		0b0110, 0b0110,
		0b1011, 0b1011,
		0b1110, 0b1110,
		0b0111, 0b0111,
		0b1101, 0b1101,
		0b1111, 0b1111,
	}
	cx, cy := modernTileW/2, modernTileH/2
	for index, mask := range want {
		points := []struct {
			x, y int
			bit  uint8
		}{
			{cx - 1, 0, 1 << 0},
			{modernTileW - 1, cy - 1, 1 << 1},
			{cx - 1, modernTileH - 1, 1 << 2},
			{0, cy - 1, 1 << 3},
		}
		for _, p := range points {
			got := railMark(index, p.x, p.y)
			if got != (mask&p.bit != 0) {
				t.Fatalf("rail %d edge bit %#x = %v，預期 %v", index, p.bit, got, mask&p.bit != 0)
			}
		}
	}
}

func TestModernArtilleryFacingChangesPixels(t *testing.T) {
	m := NewModern()
	a, err := m.Unit(6) // 綠砲兵，朝向 1
	if err != nil {
		t.Fatal(err)
	}
	b, err := m.Unit(7) // 綠砲兵，朝向 2
	if err != nil {
		t.Fatal(err)
	}
	different := 0
	for i := range a.Image.Pix {
		if a.Image.Pix[i] != b.Image.Pix[i] {
			different++
		}
	}
	if different == 0 {
		t.Fatal("砲兵六朝向不應共用同一張像素圖")
	}
}

func TestModernTerrainHomageMotifs(t *testing.T) {
	m := NewModern()
	tiles := map[int]string{}
	for _, kind := range []int{0, 1, 2, 3, 4, 5, 6, 9, 10} {
		b, err := m.Tile(kind)
		if err != nil {
			t.Fatal(err)
		}
		key := string(b.Image.Pix)
		for other, otherKey := range tiles {
			if otherKey == key {
				t.Fatalf("地形 %d 與 %d 像素完全相同", kind, other)
			}
		}
		tiles[kind] = key
	}
	// 高山用赭紅山脊（調色盤 13），不是雪峰白。
	if base, accent := terrainColors(5); accent != 13 {
		t.Fatalf("高山 accent=%d，應為 13 赭紅", accent)
	} else {
		_ = base
	}
}

func TestModernHighUnitHomageSilhouettes(t *testing.T) {
	m := NewModern()
	for i := 0; i < 18; i++ {
		im, err := m.HighUnit(i, 48, 26)
		if err != nil {
			t.Fatal(err)
		}
		// 無框剪影：四角透明，中央不透明。
		for _, p := range [][2]int{{0, 0}, {47, 0}, {0, 25}, {47, 25}} {
			if got := im.RGBAAt(p[0], p[1]); got.A != 0 {
				t.Fatalf("高解析剪影 %d 角落 (%d,%d) alpha=%d 應透明", i, p[0], p[1], got.A)
			}
		}
		if got := im.RGBAAt(24, 13); got.A == 0 {
			t.Fatalf("高解析剪影 %d 中央透明，不應為空", i)
		}
	}
}

func TestModernUnitHomageSilhouettes(t *testing.T) {
	m := NewModern()
	countTeam := func(pix []byte, want byte) int {
		n := 0
		for _, v := range pix {
			if v == want {
				n++
			}
		}
		return n
	}
	for i := 0; i < 18; i++ {
		u, err := m.Unit(i)
		if err != nil {
			t.Fatal(err)
		}
		pix := u.Image.Pix
		at := func(x, y int) byte { return pix[y*32+x] }
		// 無框剪影：四角透明
		for _, p := range [][2]int{{0, 0}, {31, 0}, {0, 16}, {31, 16}} {
			if at(p[0], p[1]) != 0 {
				t.Fatalf("剪影 %d 角落 (%d,%d)=%d 應透明", i, p[0], p[1], at(p[0], p[1]))
			}
		}
		// 勢力色為體：0..5 奇數紅、砲兵 12..17 紅，其餘綠，且佔比足夠
		want := byte(5)
		if i == 1 || i == 3 || i == 5 || i >= 12 {
			want = 3
		}
		if n := countTeam(pix, want); n < 40 {
			t.Fatalf("剪影 %d 勢力色像素僅 %d，應 >= 40", i, n)
		}
		// 墨色陰影存在
		if n := countTeam(pix, 2); n < 10 {
			t.Fatalf("剪影 %d 墨色像素僅 %d，應 >= 10", i, n)
		}
	}
	// 四兵種剪影互不相同
	seen := map[string]int{}
	for _, i := range []int{0, 2, 4, 6} {
		u, err := m.Unit(i)
		if err != nil {
			t.Fatal(err)
		}
		key := string(u.Image.Pix)
		if prev, dup := seen[key]; dup {
			t.Fatalf("剪影 %d 與 %d 像素完全相同", i, prev)
		}
		seen[key] = i
	}
}

func TestModernCommandIconsShareCounterFrame(t *testing.T) {
	m := NewModern()
	for i := 0; i < 15; i++ {
		icon, err := m.CommandIcon(i)
		if err != nil {
			t.Fatal(err)
		}
		pix := icon.Image.Pix
		at := func(x, y int) byte { return pix[y*16+x] }
		// 四角透明，上下邊中點為暗紅框。
		for _, p := range [][2]int{{0, 0}, {15, 0}, {0, 15}, {15, 15}} {
			if at(p[0], p[1]) != 0 {
				t.Fatalf("指令 %d 角落 (%d,%d)=%d 應透明", i, p[0], p[1], at(p[0], p[1]))
			}
		}
		if at(8, 0) != 16 || at(8, 15) != 16 {
			t.Fatalf("指令 %d 上下框應為暗紅 16", i)
		}
	}
}
