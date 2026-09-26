package layout

import "fmt"

// ModernDesignWidth／ModernDesignHeight 是 H1 的設計座標，不是裝置實體像素。
// A 方案固定 1280×720；Android 與桌面都先在這個座標系排版，再由平台做等比
// letterbox／安全區轉換。
const (
	ModernDesignWidth  = 1280
	ModernDesignHeight = 720

	// H4「軍事衛星作戰室」保留常駐導覽／狀態軌，再以中央地圖窗、右側軍情
	// 檢視器與底部操作列分開既有資訊。這些只是 Modern 呈現 metrics，不增加
	// 規則或 action。
	modernRailW      = 72
	modernRailGap    = 16
	modernGap        = 16
	modernCardInset  = 20
	modernHeaderH    = 58
	modernMapCardW   = 704
	modernInfoH      = 360
	modernButtonH    = 64
	modernSecondaryH = 48

	modernMapBaseW = 448 // game.BattlefieldW；避免 layout package 依賴 game。
	modernMapBaseH = 348 // game.BattlefieldH。
	modernScaleNum = 3
	modernScaleDen = 2
)

// Rect 是 renderer 與滑鼠／觸控 adapter 共用的設計座標矩形。
// 它不攜帶 Ebiten 或規則層型別，方便無頭測試。
type Rect struct {
	X, Y, W, H int
}

func (r Rect) Right() int  { return r.X + r.W }
func (r Rect) Bottom() int { return r.Y + r.H }

func (r Rect) Contains(x, y int) bool {
	return x >= r.X && x < r.Right() && y >= r.Y && y < r.Bottom()
}

func (r Rect) Intersects(other Rect) bool {
	return r.X < other.Right() && other.X < r.Right() &&
		r.Y < other.Bottom() && other.Y < r.Bottom()
}

// SafeInsets 是設計座標中的安全區。裝置像素與瀏海／系統列由平台 adapter
// 轉換，不讓 renderer 直接猜 Android density。
type SafeInsets struct {
	Top, Right, Bottom, Left int
}

// ModernSurface 是一個固定設計畫布的描述。
type ModernSurface struct {
	Width, Height int
	Safe          SafeInsets
}

// ModernDesignSurface 回傳目前已確認的 A 方案。
func ModernDesignSurface() ModernSurface {
	return ModernSurface{
		Width: ModernDesignWidth, Height: ModernDesignHeight,
		Safe: SafeInsets{Top: 24, Right: 24, Bottom: 24, Left: 24},
	}
}

func (s ModernSurface) Bounds() Rect { return Rect{W: s.Width, H: s.Height} }

func (s ModernSurface) ContentBounds() (Rect, error) {
	if s.Width <= 0 || s.Height <= 0 {
		return Rect{}, fmt.Errorf("modern surface 尺寸必須為正：%dx%d", s.Width, s.Height)
	}
	if s.Safe.Left < 0 || s.Safe.Right < 0 || s.Safe.Top < 0 || s.Safe.Bottom < 0 {
		return Rect{}, fmt.Errorf("modern surface 安全區不可為負：%+v", s.Safe)
	}
	r := Rect{
		X: s.Safe.Left, Y: s.Safe.Top,
		W: s.Width - s.Safe.Left - s.Safe.Right,
		H: s.Height - s.Safe.Top - s.Safe.Bottom,
	}
	if r.W <= 0 || r.H <= 0 {
		return Rect{}, fmt.Errorf("modern surface 安全區吃掉畫布：%+v", s.Safe)
	}
	return r, nil
}

// ModernMapLayout 是 H4 的全域指揮室切片：導覽／狀態軌、衛星地圖窗、軍情
// 檢視器與操作／事件列共用同一份矩形。這些位置是 remake 的高解析排版，
// 不改 14×14 六角規則。
type ModernMapLayout struct {
	Surface             ModernSurface
	Rail                Rect
	MapRailMarker       Rect
	CommandRailButton   Rect
	NarrativeRailButton Rect
	Header              Rect
	InfoPanel           Rect
	CommanderPortrait   Rect
	EventStrip          Rect
	SignalStrip         Rect
	MapCard             Rect
	Map                 Rect
	MetricCards         []Rect
	CommandButton       Rect
	NarrativeButton     Rect
}

// ModernPageLayout 是高解析度文件／選單頁共用的頁首、內容與導覽 metrics。
// 所有按鈕都大於 48×48 設計像素，Android 的單指觸控不需另外猜座標。
type ModernPageLayout struct {
	Surface  ModernSurface
	Rail     Rect
	Header   Rect
	Body     Rect
	Footer   Rect
	Back     Rect
	Previous Rect
	Next     Rect
}

// NewModernPageLayout 建立一般 Modern 頁。它只描述幾何，不知道 screen 或
// action；renderer 與 pointer adapter 各自把同一矩形映射成視覺／動作。
func NewModernPageLayout(surface ModernSurface) (ModernPageLayout, error) {
	if surface.Width != ModernDesignWidth || surface.Height != ModernDesignHeight {
		return ModernPageLayout{}, fmt.Errorf("Modern 頁目前只接受 %dx%d 設計畫布，得到 %dx%d",
			ModernDesignWidth, ModernDesignHeight, surface.Width, surface.Height)
	}
	content, err := surface.ContentBounds()
	if err != nil {
		return ModernPageLayout{}, err
	}
	rail := Rect{X: content.X, Y: content.Y, W: modernRailW, H: content.H}
	workspace := Rect{X: rail.Right() + modernRailGap, Y: content.Y,
		W: content.Right() - (rail.Right() + modernRailGap), H: content.H}
	if workspace.W <= 0 {
		return ModernPageLayout{}, fmt.Errorf("Modern 頁工作區不足：rail=%+v content=%+v", rail, content)
	}
	const headerH, footerH, gap = 58, 64, 16
	header := Rect{X: workspace.X, Y: workspace.Y, W: workspace.W, H: headerH}
	footer := Rect{X: workspace.X, Y: workspace.Bottom() - footerH, W: workspace.W, H: footerH}
	body := Rect{X: content.X, Y: header.Bottom() + gap, W: content.W,
		H: footer.Y - (header.Bottom() + gap) - gap}
	body.X, body.W = workspace.X, workspace.W
	if body.W <= 0 || body.H <= 0 {
		return ModernPageLayout{}, fmt.Errorf("Modern 頁內容區不足：%+v", content)
	}
	const navW, navH = 144, 52
	back := Rect{X: workspace.X, Y: footer.Y + (footer.H-navH)/2, W: navW, H: navH}
	previous := Rect{X: workspace.Right() - navW*2 - 16, Y: back.Y, W: navW, H: navH}
	next := Rect{X: workspace.Right() - navW, Y: back.Y, W: navW, H: navH}
	return ModernPageLayout{Surface: surface, Rail: rail, Header: header, Body: body, Footer: footer,
		Back: back, Previous: previous, Next: next}, nil
}

// ModernCommandLayout 是 15 項政略指令的高解析度矩陣排列。H4 改為 5×3，
// 讓策略玩家可以掃視所有既有指令，而不是把舊式垂直清單放大；編號與 action
// 順序仍是 row-major 的 1..15。
type ModernCommandLayout struct {
	Page  ModernPageLayout
	Cards []Rect
}

func NewModernCommandLayout(surface ModernSurface) (ModernCommandLayout, error) {
	page, err := NewModernPageLayout(surface)
	if err != nil {
		return ModernCommandLayout{}, err
	}
	const columns, rows, gap = 5, 3, 12
	cardW := (page.Body.W - gap*(columns-1)) / columns
	cardH := (page.Body.H - gap*(rows-1)) / rows
	if cardW < 48 || cardH < 48 {
		return ModernCommandLayout{}, fmt.Errorf("Modern 指令卡片過小：%dx%d", cardW, cardH)
	}
	cards := make([]Rect, 0, columns*rows)
	for row := 0; row < rows; row++ {
		for col := 0; col < columns; col++ {
			cards = append(cards, Rect{X: page.Body.X + col*(cardW+gap),
				Y: page.Body.Y + row*(cardH+gap), W: cardW, H: cardH})
		}
	}
	return ModernCommandLayout{Page: page, Cards: cards}, nil
}

// ModernOptionLayout 是設定／其他選項的雙欄卡片排列。最後一個選項若是奇數
// 會保留原寬度，不偷偷拉伸到另一欄以外。
type ModernOptionLayout struct {
	Page    ModernPageLayout
	Options []Rect
}

func NewModernOptionLayout(surface ModernSurface, count int) (ModernOptionLayout, error) {
	page, err := NewModernPageLayout(surface)
	if err != nil {
		return ModernOptionLayout{}, err
	}
	if count <= 0 {
		return ModernOptionLayout{}, fmt.Errorf("Modern 選項數必須為正：%d", count)
	}
	const columns, gap = 2, 16
	rows := (count + columns - 1) / columns
	cardW := (page.Body.W - gap) / columns
	cardH := (page.Body.H - gap*(rows-1)) / rows
	if cardW < 48 || cardH < 48 {
		return ModernOptionLayout{}, fmt.Errorf("Modern 選項卡片過小：%dx%d", cardW, cardH)
	}
	options := make([]Rect, 0, count)
	for i := 0; i < count; i++ {
		row, col := i/columns, i%columns
		options = append(options, Rect{X: page.Body.X + col*(cardW+gap),
			Y: page.Body.Y + row*(cardH+gap), W: cardW, H: cardH})
	}
	return ModernOptionLayout{Page: page, Options: options}, nil
}

// ModernFlowLayout 是尚未需要專用插圖的政略流程頁共用 metrics。它把原本
// 640×350 的垂直清單／數字輸入頁搬到同一張 1280×720 設計畫布；renderer
// 與 pointer 只共享矩形，不在 layout 層知道 screen 或規則。
type ModernFlowLayout struct {
	Page   ModernPageLayout
	Cards  []Rect
	Input  Rect
	Keypad []Rect
}

// NewModernFlowLayout 產生 0..99 個流程選項卡。選項很多時增加欄數，維持
// 至少 48×48 設計像素，讓同一份命中區可以直接映射 Android 觸控 adapter。
// count=0 代表數字輸入頁，會另外產生輸入框與 3×4 鍵盤。
func NewModernFlowLayout(surface ModernSurface, count int) (ModernFlowLayout, error) {
	page, err := NewModernPageLayout(surface)
	if err != nil {
		return ModernFlowLayout{}, err
	}
	if count < 0 || count > 99 {
		return ModernFlowLayout{}, fmt.Errorf("Modern 流程選項數超出 0..99：%d", count)
	}
	if count == 0 {
		input := Rect{X: page.Body.X + 48, Y: page.Body.Y + 24, W: page.Body.W - 96, H: 76}
		const keyW, keyH, gap, columns = 112, 52, 10, 3
		rows := 4
		keyStartX := page.Body.X + (page.Body.W-(keyW*columns+gap*(columns-1)))/2
		keyStartY := input.Bottom() + 28
		keys := make([]Rect, 0, rows*columns)
		for i := 0; i < rows*columns; i++ {
			col, row := i%columns, i/columns
			keys = append(keys, Rect{X: keyStartX + col*(keyW+gap), Y: keyStartY + row*(keyH+gap), W: keyW, H: keyH})
		}
		for _, key := range keys {
			if !page.Body.Contains(key.X, key.Y) || key.Right() > page.Body.Right() || key.Bottom() > page.Body.Bottom() {
				return ModernFlowLayout{}, fmt.Errorf("Modern 數字鍵盤超出內容區：%+v", key)
			}
		}
		return ModernFlowLayout{Page: page, Input: input, Keypad: keys}, nil
	}

	columns := 2
	switch {
	case count > 10 && count <= 24:
		columns = 3
	case count > 24 && count <= 40:
		columns = 5
	case count > 40 && count <= 60:
		columns = 6
	case count > 60 && count <= 80:
		columns = 8
	case count > 80:
		columns = 11
	}
	const gap = 10
	rows := (count + columns - 1) / columns
	cardW := (page.Body.W - gap*(columns-1)) / columns
	cardH := (page.Body.H - gap*(rows-1)) / rows
	if cardW < 48 || cardH < 48 {
		return ModernFlowLayout{}, fmt.Errorf("Modern 流程卡片過小：%dx%d（count=%d columns=%d）", cardW, cardH, count, columns)
	}
	cards := make([]Rect, 0, count)
	for i := 0; i < count; i++ {
		row, col := i/columns, i%columns
		cards = append(cards, Rect{X: page.Body.X + col*(cardW+gap), Y: page.Body.Y + row*(cardH+gap), W: cardW, H: cardH})
	}
	return ModernFlowLayout{Page: page, Cards: cards}, nil
}

// ModernBattleLayout 是 H1-c 戰鬥 HUD 的第一份安全 metrics：左側保留 3/2
// 戰場，右側把資源、五項命令與三個大控制鍵分層。撤退時同一側改顯示 3×4
// 數字鍵盤；規則與既有 BattleCommand／Digit action 不在 layout 層複製。
type ModernBattleLayout struct {
	Surface        ModernSurface
	Rail           Rect
	Header         Rect
	MapCard        Rect
	Map            Rect
	Panel          Rect
	Log            Rect
	CommandButtons []Rect
	ControlButtons []Rect
	RetreatButtons []Rect
}

func NewModernBattleLayout(surface ModernSurface) (ModernBattleLayout, error) {
	if surface.Width != ModernDesignWidth || surface.Height != ModernDesignHeight {
		return ModernBattleLayout{}, fmt.Errorf("Modern 戰鬥目前只接受 %dx%d 設計畫布，得到 %dx%d",
			ModernDesignWidth, ModernDesignHeight, surface.Width, surface.Height)
	}
	content, err := surface.ContentBounds()
	if err != nil {
		return ModernBattleLayout{}, err
	}
	rail := Rect{X: content.X, Y: content.Y, W: modernRailW, H: content.H}
	workspace := Rect{X: rail.Right() + modernRailGap, Y: content.Y,
		W: content.Right() - (rail.Right() + modernRailGap), H: content.H}
	if workspace.W <= 0 {
		return ModernBattleLayout{}, fmt.Errorf("Modern 戰鬥工作區不足：rail=%+v content=%+v", rail, content)
	}
	const headerH, gap = 56, 16
	header := Rect{X: workspace.X, Y: workspace.Y, W: workspace.W, H: headerH}
	mapRect := Rect{X: workspace.X, Y: header.Bottom() + gap, W: scaleModern(modernMapBaseW), H: scaleModern(modernMapBaseH)}
	mapCard := Rect{X: mapRect.X, Y: mapRect.Y, W: mapRect.W, H: mapRect.H}
	panel := Rect{X: mapCard.Right() + gap, Y: mapCard.Y, W: workspace.Right() - (mapCard.Right() + gap), H: mapCard.H}
	if panel.W < 300 || panel.H < 480 {
		return ModernBattleLayout{}, fmt.Errorf("Modern 戰鬥面板不足：%+v", panel)
	}
	// 資源列、戰況訊息、五項命令與控制鍵各保留自己的帶狀區；舊版的
	// 6 列資源曾與命令、log 疊在一起，雖能繪製卻不是可讀的策略介面。
	log := Rect{X: panel.X + 16, Y: panel.Y + 154, W: panel.W - 32, H: 28}
	const commandH, commandGap = 48, 4
	commands := make([]Rect, 0, 5)
	commandY := log.Bottom() + 12
	for i := 0; i < 5; i++ {
		commands = append(commands, Rect{X: panel.X + 16, Y: commandY + i*(commandH+commandGap), W: panel.W - 32, H: commandH})
	}
	const controlH, controlGap = 48, 12
	controlW := (panel.W - 32 - controlGap*2) / 3
	controls := make([]Rect, 0, 3)
	controlY := panel.Bottom() - controlH - 16
	controlStart := panel.X + (panel.W-(controlW*3+controlGap*2))/2
	for i := 0; i < 3; i++ {
		controls = append(controls, Rect{X: controlStart + i*(controlW+controlGap), Y: controlY, W: controlW, H: controlH})
	}
	const keyW, keyH, keyGap, keyCols = 112, 48, 8, 3
	keys := make([]Rect, 0, 12)
	keyStartX := panel.X + (panel.W-(keyW*keyCols+keyGap*(keyCols-1)))/2
	keyStartY := panel.Y + 222
	for i := 0; i < 12; i++ {
		col, row := i%keyCols, i/keyCols
		keys = append(keys, Rect{X: keyStartX + col*(keyW+keyGap), Y: keyStartY + row*(keyH+keyGap), W: keyW, H: keyH})
	}
	return ModernBattleLayout{Surface: surface, Rail: rail, Header: header, MapCard: mapCard, Map: mapRect,
		Panel: panel, Log: log, CommandButtons: commands, ControlButtons: controls, RetreatButtons: keys}, nil
}

func (l ModernBattleLayout) MapCellRect(col, row int) (Rect, error) {
	if col < 0 || col >= 14 || row < 0 || row >= 14 {
		return Rect{}, fmt.Errorf("Modern 戰鬥格 (%d,%d) 超出 14×14", col, row)
	}
	x := l.Map.X + scaleModern(col*32)
	y := l.Map.Y + scaleModern(row*24)
	if col%2 == 1 {
		y += scaleModern(12)
	}
	return Rect{X: x, Y: y, W: scaleModern(32), H: scaleModern(24)}, nil
}

// NewModernMapLayout 建立 1280×720 H4 軍事衛星作戰室的 metrics。
func NewModernMapLayout(surface ModernSurface) (ModernMapLayout, error) {
	if surface.Width != ModernDesignWidth || surface.Height != ModernDesignHeight {
		return ModernMapLayout{}, fmt.Errorf("Modern H1 目前只接受 %dx%d 設計畫布，得到 %dx%d",
			ModernDesignWidth, ModernDesignHeight, surface.Width, surface.Height)
	}
	content, err := surface.ContentBounds()
	if err != nil {
		return ModernMapLayout{}, err
	}
	rail := Rect{X: content.X, Y: content.Y, W: modernRailW, H: content.H}
	workspace := Rect{X: rail.Right() + modernRailGap, Y: content.Y,
		W: content.Right() - (rail.Right() + modernRailGap), H: content.H}
	if workspace.W <= modernMapCardW+modernGap || workspace.H <= 0 {
		return ModernMapLayout{}, fmt.Errorf("Modern H4 工作區不足：workspace=%+v", workspace)
	}

	header := Rect{X: workspace.X, Y: workspace.Y, W: workspace.W, H: modernHeaderH}
	bodyY := header.Bottom() + 16
	bodyH := content.Bottom() - bodyY
	if bodyH <= 0 {
		return ModernMapLayout{}, fmt.Errorf("Modern H4 內容區不足：%+v", content)
	}
	card := Rect{X: workspace.X, Y: bodyY, W: modernMapCardW, H: bodyH}
	panel := Rect{X: card.Right() + modernGap, Y: bodyY, W: workspace.Right() - (card.Right() + modernGap), H: modernInfoH}
	events := Rect{X: panel.X, Y: panel.Bottom() + modernGap, W: panel.W, H: content.Bottom() - (panel.Bottom() + modernGap)}
	if card.W <= 0 || panel.W <= 0 || panel.H <= 0 || events.H <= 0 {
		return ModernMapLayout{}, fmt.Errorf("Modern H4 欄位不足：card=%+v panel=%+v events=%+v", card, panel, events)
	}
	mapW := scaleModern(modernMapBaseW)
	mapH := scaleModern(modernMapBaseH)
	// 地圖窗依 SPEC-41 固定為 704 px 寬，左右各留 16 px 的掃描框邊界；
	// 軍情／操作欄的卡片仍使用 20 px 內距。
	const mapInset = 16
	if mapW > card.W-mapInset*2 || mapH > card.H-mapInset*2 {
		return ModernMapLayout{}, fmt.Errorf("Modern 地圖無法放入指揮台：map=%dx%d card=%+v", mapW, mapH, card)
	}
	mapRect := Rect{
		X: card.X + (card.W-mapW)/2,
		Y: card.Y + (card.H-mapH)/2,
		W: mapW, H: mapH,
	}
	// 司令肖像只顯示通過人物身分與權利 gate 的離線登錄；未登錄人物使用
	// 同尺寸檔案卡。文字區不因有無照片改變，六張軍情摘要卡仍共用固定幾何。
	const metricGap, metricH, metricCols, metricRows = 10, 54, 3, 2
	metricW := (panel.W - modernCardInset*2 - metricGap*(metricCols-1)) / metricCols
	portrait := Rect{X: panel.Right() - modernCardInset - 108, Y: panel.Y + 88, W: 108, H: 148}
	metricY := panel.Y + 242
	metrics := make([]Rect, 0, metricCols*metricRows)
	for row := 0; row < metricRows; row++ {
		for col := 0; col < metricCols; col++ {
			metrics = append(metrics, Rect{X: panel.X + modernCardInset + col*(metricW+metricGap),
				Y: metricY + row*(metricH+metricGap), W: metricW, H: metricH})
		}
	}
	signal := Rect{X: events.X + modernCardInset, Y: events.Y + 16, W: events.W - modernCardInset*2, H: 28}
	command := Rect{X: events.X + modernCardInset, Y: events.Y + 86,
		W: events.W - modernCardInset*2, H: modernButtonH}
	narrative := Rect{X: command.X, Y: events.Bottom() - modernCardInset - modernSecondaryH,
		W: command.W, H: modernSecondaryH}
	for _, r := range metrics {
		if !panel.Contains(r.X, r.Y) || r.Right() > panel.Right() || r.Bottom() > panel.Bottom() {
			return ModernMapLayout{}, fmt.Errorf("Modern H4 軍情卡超出檢視器：%+v panel=%+v", r, panel)
		}
	}
	if !panel.Contains(portrait.X, portrait.Y) || portrait.Right() > panel.Right() || portrait.Bottom() > panel.Bottom() {
		return ModernMapLayout{}, fmt.Errorf("Modern H4 司令肖像超出檢視器：%+v panel=%+v", portrait, panel)
	}
	for _, r := range []Rect{signal, command, narrative} {
		if !events.Contains(r.X, r.Y) || r.Right() > events.Right() || r.Bottom() > events.Bottom() {
			return ModernMapLayout{}, fmt.Errorf("Modern H4 操作列元件超出：%+v events=%+v", r, events)
		}
	}
	marker := Rect{X: rail.X + 12, Y: rail.Y + 56, W: 48, H: 48}
	commandRail := Rect{X: rail.X + 12, Y: marker.Bottom() + 32, W: 48, H: 48}
	narrativeRail := Rect{X: rail.X + 12, Y: commandRail.Bottom() + 16, W: 48, H: 48}
	return ModernMapLayout{
		Surface: surface, Rail: rail, MapRailMarker: marker, CommandRailButton: commandRail,
		NarrativeRailButton: narrativeRail, Header: header, InfoPanel: panel, CommanderPortrait: portrait, EventStrip: events,
		SignalStrip: signal, MapCard: card, Map: mapRect, MetricCards: metrics,
		CommandButton: command, NarrativeButton: narrative,
	}, nil
}

// MapCellRect 將原版 14×14 六角格幾何映射到 Modern 設計座標。
// 奇數欄下移半格的原版契約仍保留，只改顯示尺寸。
func (l ModernMapLayout) MapCellRect(col, row int) (Rect, error) {
	if col < 0 || col >= 14 || row < 0 || row >= 14 {
		return Rect{}, fmt.Errorf("Modern 地圖格 (%d,%d) 超出 14×14", col, row)
	}
	x := l.Map.X + scaleModern(col*32)
	y := l.Map.Y + scaleModern(row*24)
	if col%2 == 1 {
		y += scaleModern(12)
	}
	return Rect{X: x, Y: y, W: scaleModern(32), H: scaleModern(24)}, nil
}

func scaleModern(v int) int {
	if v <= 0 {
		return 0
	}
	return (v*modernScaleNum + modernScaleDen/2) / modernScaleDen
}
