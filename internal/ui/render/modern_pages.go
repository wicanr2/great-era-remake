package render

import (
	"fmt"
	"image"
	"sort"
	"strconv"
	"strings"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
	"github.com/wicanr2/great-era-remake/internal/ui/textlayout"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

// ModernCommandSurfaceData 是高解析度 Modern 指令頁的純資料邊界。
type ModernCommandSurfaceData struct {
	Title  string
	Labels []string
	Fonts  *assets.EtenFonts
	Icons  uitheme.HUDIconProvider
	Style  uitheme.UIStyle
	Back   string
}

// ModernOptionSurfaceData 是設定／其他選項共用的卡片資料。
type ModernOptionSurfaceData struct {
	Title    string
	Options  []string
	Selected []bool
	Fonts    *assets.EtenFonts
	Style    uitheme.UIStyle
	Back     string
}

// ModernFlowSurfaceData 是高解析度政略流程頁的共用資料邊界。它涵蓋選項清單、
// 數字輸入與只讀詳情；Action／規則仍由 cmd/dsds 維護，renderer 只負責畫面。
type ModernFlowSurfaceData struct {
	Title     string
	Prompt    string
	Input     string
	Hint      string
	Options   []string
	Selected  int
	InputMode bool
	Fonts     *assets.EtenFonts
	Style     uitheme.UIStyle
	Back      string
	Previous  string
	Next      string
}

// DrawModernFlowSurface 把尚未需要專用插圖的政略／查閱流程統一放進
// 1280×720 Modern metrics。選項順序與 input 欄位由呼叫端依既有 screen state
// 提供；本函式不執行規則、不自行改寫語系或存檔。
func (c *Canvas) DrawModernFlowSurface(d ModernFlowSurfaceData) error {
	if err := validateModernPageCanvas(c, d.Style); err != nil {
		return err
	}
	if strings.TrimSpace(d.Title) == "" {
		return fmt.Errorf("Modern 流程頁缺少標題")
	}
	count := len(d.Options)
	if d.InputMode {
		count = 0
	}
	l, err := uilayout.NewModernFlowLayout(uilayout.ModernDesignSurface(), count)
	if err != nil {
		return err
	}
	drawModernPageFrame(c, l.Page, d.Style)
	missing := map[rune]bool{}
	drawModernTitleScaledInto(c, d.Fonts, trimHalfCells(d.Title, 34), d.Style.Ink,
		l.Page.Header.X+24, l.Page.Header.Y+8, 2, 34, missing)
	if d.Prompt != "" {
		drawModernTextScaledInto(c, d.Fonts, d.Prompt, d.Style.Ink, l.Page.Header.X+24, l.Page.Header.Y+34, 1, 72, missing)
	}
	if len(d.Options) > 0 {
		for i, card := range l.Cards {
			fill, ink, border := d.Style.Panel, d.Style.Ink, d.Style.Muted
			if i == d.Selected {
				fill, ink, border = d.Style.Accent, d.Style.Paper, d.Style.Accent
			}
			drawModernDecoratedControl(c, card, d.Style, fill, border)
			c.fillRect(card.X, card.Y, 6, card.H, d.Style.FactionTint[i%len(d.Style.FactionTint)])
			label := strconv.Itoa(i+1) + "  " + d.Options[i]
			maxHalf := (card.W - 44) / 8
			if maxHalf < 8 {
				maxHalf = 8
			}
			scale, y := 1, card.Y+(card.H-15)/2
			if modernTextHalfCells(label) <= (card.W-44)/16 {
				scale, y, maxHalf = 2, card.Y+(card.H-30)/2, (card.W-44)/16
			}
			drawModernTextScaledInto(c, d.Fonts, label, ink, card.X+22, y, scale, maxHalf, missing)
		}
	} else if d.InputMode {
		drawModernDecoratedPanel(c, l.Input, d.Style, d.Style.Panel)
		inputLabel := d.Prompt
		if inputLabel == "" {
			inputLabel = "輸入"
		}
		drawModernTextScaledInto(c, d.Fonts, inputLabel, d.Style.Muted, l.Input.X+24, l.Input.Y+14, 1, 48, missing)
		drawModernTextScaledInto(c, d.Fonts, d.Input, d.Style.Ink, l.Input.Right()-220, l.Input.Y+18, 2, 22, missing)
		keyLabels := []string{"1", "2", "3", "4", "5", "6", "7", "8", "9", "刪除", "0", "確認"}
		for i, key := range l.Keypad {
			drawModernDecoratedControl(c, key, d.Style, d.Style.Panel, d.Style.Accent)
			drawModernTextScaledInto(c, d.Fonts, keyLabels[i], d.Style.Ink,
				key.X+(key.W-16)/2, key.Y+(key.H-15)/2, 1, 20, missing)
		}
	} else {
		card := uilayout.Rect{X: l.Page.Body.X + 48, Y: l.Page.Body.Y + 40, W: l.Page.Body.W - 96, H: l.Page.Body.H - 80}
		drawModernDecoratedControl(c, card, d.Style, d.Style.Panel, d.Style.Muted)
		drawModernTextScaledInto(c, d.Fonts, d.Prompt, d.Style.Ink, card.X+32, card.Y+32, 1, 120, missing)
		drawModernTextScaledInto(c, d.Fonts, d.Input, d.Style.Accent, card.X+32, card.Y+78, 2, 60, missing)
	}
	if d.Hint != "" {
		drawModernTextScaledInto(c, d.Fonts, d.Hint, d.Style.Muted, l.Page.Footer.X+180, l.Page.Footer.Y+20, 1, 100, missing)
	}
	drawModernNav(c, l.Page, d.Fonts, d.Style, d.Back, d.Previous, d.Next)
	if len(missing) != 0 {
		chars := make([]rune, 0, len(missing))
		for r := range missing {
			chars = append(chars, r)
		}
		sort.Slice(chars, func(i, j int) bool { return chars[i] < chars[j] })
		return fmt.Errorf("Modern 流程頁缺字：%q", string(chars))
	}
	return nil
}

// DrawModernCommandSurface 以 H4 的 5×3 指令矩陣呈現既有 15 項政略指令。
// 卡片順序完全沿用 action Selection(1..15)，renderer 不複製規則或改變命令成本。
func (c *Canvas) DrawModernCommandSurface(d ModernCommandSurfaceData) error {
	if err := validateModernPageCanvas(c, d.Style); err != nil {
		return err
	}
	if len(d.Labels) != 15 {
		return fmt.Errorf("Modern 指令頁需要 15 個 label，得到 %d", len(d.Labels))
	}
	l, err := uilayout.NewModernCommandLayout(uilayout.ModernDesignSurface())
	if err != nil {
		return err
	}
	drawModernPageFrame(c, l.Page, d.Style)
	if missing := drawModernTitleScaled(c, d.Fonts, trimHalfCells(d.Title, 34), d.Style.Ink,
		l.Page.Header.X+24, l.Page.Header.Y+8, 2, 34); len(missing) != 0 {
		return fmt.Errorf("Modern 指令標題缺字：%q", string(missing))
	}
	for i, card := range l.Cards {
		drawModernDecoratedControl(c, card, d.Style, d.Style.Panel, d.Style.Muted)
		accent := d.Style.FactionTint[i%len(d.Style.FactionTint)]
		c.fillRect(card.X, card.Y, 5, card.H, accent)
		c.fillRect(card.X+20, card.Y+24, card.W-40, 1, d.Style.Muted)
		if missing := drawModernTextScaled(c, d.Fonts, strconv.Itoa(i+1), d.Style.Muted,
			card.Right()-32, card.Y+10, 1, 4); len(missing) != 0 {
			return fmt.Errorf("Modern 指令編號 %d 缺字：%q", i+1, string(missing))
		}
		if d.Icons != nil {
			if err := drawModernCommandIcon(c, d.Icons, i,
				uilayout.Rect{X: card.X + 24, Y: card.Y + 42, W: 44, H: 44}); err != nil {
				return err
			}
		}
		missing := map[rune]bool{}
		labelBox := uilayout.Rect{X: card.X + 24, Y: card.Y + 88, W: card.W - 48, H: card.H - 96}
		drawModernButtonLabel(c, d.Fonts, d.Labels[i], d.Style.Ink, labelBox, missing)
		if len(missing) != 0 {
			missingRunes := make([]rune, 0, len(missing))
			for r := range missing {
				missingRunes = append(missingRunes, r)
			}
			sort.Slice(missingRunes, func(a, b int) bool { return missingRunes[a] < missingRunes[b] })
			return fmt.Errorf("Modern 指令 %d 缺字：%q", i+1, string(missingRunes))
		}
		for signal := 0; signal < 3; signal++ {
			w := 16 + signal*8
			c.fillRect(card.X+24+signal*30, card.Bottom()-10, w, 3, accent)
		}
	}
	drawModernNav(c, l.Page, d.Fonts, d.Style, d.Back, "", "")
	return nil
}

// DrawModernOptionsSurface 畫設定類頁面。Options 與 Selected 由呼叫端依
// 現有偏好狀態建立；卡片本身只公告輸入位置，不直接修改 preferences。
func (c *Canvas) DrawModernOptionsSurface(d ModernOptionSurfaceData) error {
	if err := validateModernPageCanvas(c, d.Style); err != nil {
		return err
	}
	if len(d.Options) == 0 {
		return fmt.Errorf("Modern 選項頁不可為空")
	}
	l, err := uilayout.NewModernOptionLayout(uilayout.ModernDesignSurface(), len(d.Options))
	if err != nil {
		return err
	}
	drawModernPageFrame(c, l.Page, d.Style)
	if missing := drawModernTitleScaled(c, d.Fonts, trimHalfCells(d.Title, 34), d.Style.Ink,
		l.Page.Header.X+24, l.Page.Header.Y+8, 2, 34); len(missing) != 0 {
		return fmt.Errorf("Modern 選項標題缺字：%q", string(missing))
	}
	for i, card := range l.Options {
		selected := i < len(d.Selected) && d.Selected[i]
		fill, ink, border := d.Style.Panel, d.Style.Ink, d.Style.Muted
		if selected {
			fill, ink, border = d.Style.Accent, d.Style.Paper, d.Style.Accent
		}
		drawModernDecoratedControl(c, card, d.Style, fill, border)
		label := strconv.Itoa(i+1) + "  " + d.Options[i]
		missing := map[rune]bool{}
		drawModernButtonLabel(c, d.Fonts, label, ink,
			uilayout.Rect{X: card.X + 24, Y: card.Y + 16, W: card.W - 48, H: card.H - 32}, missing)
		if len(missing) != 0 {
			missingRunes := make([]rune, 0, len(missing))
			for r := range missing {
				missingRunes = append(missingRunes, r)
			}
			sort.Slice(missingRunes, func(a, b int) bool { return missingRunes[a] < missingRunes[b] })
			return fmt.Errorf("Modern 選項 %d 缺字：%q", i+1, string(missingRunes))
		}
	}
	drawModernNav(c, l.Page, d.Fonts, d.Style, d.Back, "", "")
	return nil
}

const (
	modernBiographyTextScale    = 1
	modernBiographyLineHeight   = 28
	modernBiographyTextTop      = 100
	modernBiographyTextBottom   = 24
	modernBiographyReadingWidth = 780
)

func modernBiographyAdvance(r rune) int {
	modernFont, _ := assets.EmbeddedModernFont()
	if modernFont != nil {
		if glyph, ok := modernFont.Glyph(r); ok && glyph.Advance() > 0 {
			return glyph.Advance()
		}
	}
	if textlayout.RuneHalfCells(r) == 1 {
		return 8
	}
	return 16
}

func modernBiographyMetrics(l uilayout.ModernPageLayout) (bodyWidth, bodyRows int, err error) {
	// 人物檔案右側固定保留已驗證肖像／檔案卡欄；有圖與無圖使用相同正文
	// 寬度，避免資料完整度改變分頁或造成閱讀欄跳動。
	textRight := l.Body.Right() - 32
	bodyWidth = (textRight - (l.Body.X + 64) - 8) / modernBiographyTextScale
	if bodyWidth > modernBiographyReadingWidth {
		bodyWidth = modernBiographyReadingWidth
	}
	bodyRows = (l.Body.H - modernBiographyTextTop - modernBiographyTextBottom) / modernBiographyLineHeight
	if bodyWidth <= 0 || bodyRows <= 0 {
		return 0, 0, fmt.Errorf("Modern 自傳正文區不足：%dx%d", bodyWidth, bodyRows)
	}
	return bodyWidth, bodyRows, nil
}

func modernBiographyDocumentForLayout(body string, l uilayout.ModernPageLayout) (textlayout.Document, int, error) {
	bodyWidth, bodyRows, err := modernBiographyMetrics(l)
	if err != nil {
		return textlayout.Document{}, 0, err
	}
	doc, err := textlayout.LayoutProportional(body, textlayout.ProportionalOptions{
		Width: bodyWidth, Rows: bodyRows, Advance: modernBiographyAdvance,
	})
	if err != nil {
		return textlayout.Document{}, 0, err
	}
	return doc, bodyWidth, nil
}

// ModernBiographyPageCount 是高解析人物檔案唯一的頁數計算入口。它和 renderer
// 使用同一份 H4 欄寬／字距，避免鍵盤、滑鼠與觸控公告的上一頁／下一頁與畫面
// 實際頁數分叉。
func ModernBiographyPageCount(body string) (int, error) {
	l, err := uilayout.NewModernPageLayout(uilayout.ModernDesignSurface())
	if err != nil {
		return 0, err
	}
	doc, _, err := modernBiographyDocumentForLayout(body, l)
	if err != nil {
		return 0, err
	}
	return len(doc.Pages), nil
}

// DrawModernBiographySurface 以 GEMF 的比例字距做高解析換行與分頁。頁數與
// 呼叫端的 bioPage 均使用 ModernBiographyPageCount 同一份度量，不再把 640×350
// 的 28 格固定寬度硬搬到 1280×720；Eten 只作玩家自備字庫 fallback。
func (c *Canvas) DrawModernBiographySurface(fonts *assets.EtenFonts, v BiographyView,
	fg, bg assets.RGB, style uitheme.UIStyle, portrait *i18n.Portrait, portraitUnavailable string,
	navigation ...string) (BiographyRenderResult, error) {
	if c == nil || c.img == nil || c.img.Bounds().Dx() != uilayout.ModernDesignWidth || c.img.Bounds().Dy() != uilayout.ModernDesignHeight {
		return BiographyRenderResult{}, fmt.Errorf("Modern 自傳需要 %dx%d Canvas", uilayout.ModernDesignWidth, uilayout.ModernDesignHeight)
	}
	if v.Person == nil {
		return BiographyRenderResult{}, fmt.Errorf("Modern 自傳缺少人物")
	}
	l, err := uilayout.NewModernPageLayout(uilayout.ModernDesignSurface())
	if err != nil {
		return BiographyRenderResult{}, err
	}
	body := v.Person.Biography
	if strings.TrimSpace(body) == "" {
		body = v.Unavailable
	}
	// H3 的 32 px 兩倍點陣字在 720p 仍像把舊畫面硬放大。H4 改用 16 px
	// 比例 glyph、28 px baseline 與可讀欄寬，讓人物檔案像現代情報閱讀面板，
	// 同時由既有 bioPage 正常承接分頁狀態。
	doc, bodyWidth, err := modernBiographyDocumentForLayout(body, l)
	if err != nil {
		return BiographyRenderResult{}, err
	}
	if v.Page < 0 || v.Page >= len(doc.Pages) {
		return BiographyRenderResult{}, fmt.Errorf("Modern 自傳頁 %d 超出 1..%d", v.Page+1, len(doc.Pages))
	}
	drawModernPageFrame(c, l, style)
	missing := map[rune]bool{}
	portraitCard := uilayout.Rect{X: l.Body.Right() - 244, Y: l.Body.Y + 16, W: 220, H: l.Body.H - 32}
	if err := c.drawModernPortraitCard(portraitCard, portrait, portraitUnavailable, fonts, style, missing); err != nil {
		return BiographyRenderResult{}, err
	}
	name := v.Person.NameInGame
	if v.Person.NameCommon != "" && v.Person.NameCommon != v.Person.NameInGame {
		name += "／" + v.Person.NameCommon
	}
	if v.Person.Courtesy != "" {
		name += "　字" + v.Person.Courtesy
	}
	// 標題層級固定使用 2 倍比例字；合併姓名前先截在 header 的文字安全寬度，
	// 避免長英／日姓名侵入右側狀態區或把字縮回舊版大小。
	drawModernTitleScaledInto(c, fonts, trimHalfCells(v.Title+"｜"+name, 54), style.Ink, l.Header.X+24, l.Header.Y+12, 2, 54, missing)
	meta := make([]string, 0, 3)
	if v.Person.Birth != nil || v.Person.Death != nil {
		birth, death := "？", "？"
		if v.Person.Birth != nil {
			birth = strconv.Itoa(*v.Person.Birth)
		}
		if v.Person.Death != nil {
			death = strconv.Itoa(*v.Person.Death)
		}
		meta = append(meta, birth+"—"+death)
	}
	for _, item := range []string{v.Person.Birthplace, v.Person.Faction, v.Person.HighestPost} {
		if item != "" {
			meta = append(meta, item)
		}
	}
	drawModernTextScaledInto(c, fonts, strings.Join(meta, "　"), style.Muted, l.Body.X+64, l.Body.Y+40, 1, 62, missing)
	c.strokeRect(l.Body.X+56, l.Body.Y+72, bodyWidth+16, 1, style.Muted)
	maxHalf := (bodyWidth + 7) / 8
	for i, line := range doc.Pages[v.Page].Lines {
		drawModernTextScaledInto(c, fonts, line.Text, fg, l.Body.X+64,
			l.Body.Y+modernBiographyTextTop+i*modernBiographyLineHeight, modernBiographyTextScale, maxHalf, missing)
	}
	footer := fmt.Sprintf("資料來源：%d筆　可靠度：%s", len(v.Person.Sources), biographyConfidenceLabel(v.Person.Confidence))
	if v.SourceNotice != "" {
		footer = v.SourceNotice + "　" + footer
	}
	drawModernTextScaledInto(c, fonts, footer, style.Muted, l.Footer.X+24, l.Footer.Y+12, 1, 100, missing)
	drawModernTextScaledInto(c, fonts, fmt.Sprintf("%d/%d", v.Page+1, len(doc.Pages)), style.Accent, l.Footer.Right()-120, l.Footer.Y+12, 1, 12, missing)
	back, previous, next := modernNavigationDefaults(navigation)
	if v.Page == 0 {
		previous = ""
	}
	if v.Page+1 >= len(doc.Pages) {
		next = ""
	}
	drawModernNav(c, l, fonts, style, back, previous, next)
	result := BiographyRenderResult{PageCount: len(doc.Pages)}
	for r := range missing {
		result.Missing = append(result.Missing, r)
	}
	sort.Slice(result.Missing, func(i, j int) bool { return result.Missing[i] < result.Missing[j] })
	return result, nil
}

// drawModernPortraitCard 畫人物檔案的固定肖像欄。圖片已由 i18n 登錄層查過
// 身分、權利與 SHA-256；這一層只依已確認的 crop 重取樣，從不自行選圖或連網。
func (c *Canvas) drawModernPortraitCard(card uilayout.Rect, portrait *i18n.Portrait,
	placeholder string, fonts *assets.EtenFonts, style uitheme.UIStyle, missing map[rune]bool) error {
	if card.W < 64 || card.H < 96 {
		return fmt.Errorf("Modern 人物肖像欄過小：%+v", card)
	}
	drawModernDecoratedPanel(c, card, style, style.Panel)
	c.fillRect(card.X, card.Y, 5, card.H, style.Accent)
	inner := uilayout.Rect{X: card.X + 14, Y: card.Y + 14, W: card.W - 28, H: card.H - 68}
	captionHalf := max(4, (card.W-28)/8)
	if portrait == nil || portrait.Image == nil {
		imageRect := inner
		c.fillRect(imageRect.X, imageRect.Y, imageRect.W, imageRect.H, style.Paper)
		// 檔案卡的幾何記號是可讀的 fallback，不是虛構人臉或通用剪影。
		insetX := max(8, imageRect.W/4)
		insetY := max(8, imageRect.H/6)
		fileW := max(20, imageRect.W-insetX*2)
		fileH := max(28, imageRect.H-insetY*2)
		c.strokeRect(imageRect.X+insetX, imageRect.Y+insetY, fileW, fileH, style.Muted)
		c.strokeRect(imageRect.X+insetX+fileW/4, imageRect.Y+insetY+fileH/5,
			max(8, fileW/2), max(8, fileH/3), style.Muted)
		c.fillRect(imageRect.X+insetX+fileW/4, imageRect.Y+insetY+fileH*3/4,
			max(8, fileW/2), 1, style.Muted)
		drawModernTextScaledInto(c, fonts, placeholder, style.Muted, card.X+14, card.Bottom()-34, 1, captionHalf, missing)
		return nil
	}
	bounds := portrait.Image.Bounds()
	cropX := bounds.Min.X + int(float64(bounds.Dx())*portrait.Crop.X)
	cropY := bounds.Min.Y + int(float64(bounds.Dy())*portrait.Crop.Y)
	cropW := int(float64(bounds.Dx()) * portrait.Crop.Width)
	cropH := int(float64(bounds.Dy()) * portrait.Crop.Height)
	if cropW <= 0 || cropH <= 0 {
		return fmt.Errorf("Modern 人物肖像裁切區無效：%s", portrait.Asset)
	}
	// 司令官欄很窄，不能把來源照片硬拉成固定正方形。依登錄 crop
	// 置中 fit 到內框，保留人物照片的寬高比，caption 仍固定在欄底。
	imageRect := inner
	if imageRect.W > 0 && imageRect.H > 0 {
		cropAspect := float64(cropW) / float64(cropH)
		innerAspect := float64(inner.W) / float64(inner.H)
		if innerAspect > cropAspect {
			imageRect.W = max(1, int(float64(inner.H)*cropAspect))
			imageRect.X = inner.X + (inner.W-imageRect.W)/2
		} else {
			imageRect.H = max(1, int(float64(inner.W)/cropAspect))
			imageRect.Y = inner.Y + (inner.H-imageRect.H)/2
		}
	}
	drawSampledImage(c.img, imageRect, portrait.Image,
		image.Rect(cropX, cropY, cropX+cropW, cropY+cropH))
	c.strokeRect(imageRect.X, imageRect.Y, imageRect.W, imageRect.H, style.Ink)
	// 民國戰棋檔案感：墨線外框內再加一圈銅章內線。
	if imageRect.W >= 24 && imageRect.H >= 24 {
		c.strokeRect(imageRect.X+3, imageRect.Y+3, imageRect.W-6, imageRect.H-6, style.Focus)
	}
	attribution := portrait.Attribution
	if card.W < 160 {
		if cut := strings.IndexRune(attribution, '（'); cut > 0 {
			attribution = strings.TrimSpace(attribution[:cut])
		}
	}
	drawModernTextScaledInto(c, fonts, attribution, style.Muted, card.X+14, card.Bottom()-34, 1, captionHalf, missing)
	return nil
}

// DrawModernNarrativeSurface 以四張可讀卡片呈現同一批 NEWSDATA 圖像；索引與
// catalog 狀態不變，未知翻譯仍使用 catalog.Fallback。
func (c *Canvas) DrawModernNarrativeSurface(fonts *assets.EtenFonts, catalog *i18n.NarrativeCatalog,
	images []*assets.Image, page int, title string, mode i18n.WordingMode, style uitheme.UIStyle,
	navigation ...string) error {
	if c == nil || c.img == nil || c.img.Bounds().Dx() != uilayout.ModernDesignWidth || c.img.Bounds().Dy() != uilayout.ModernDesignHeight {
		return fmt.Errorf("Modern 敘事需要 %dx%d Canvas", uilayout.ModernDesignWidth, uilayout.ModernDesignHeight)
	}
	if catalog == nil || len(images) == 0 {
		return fmt.Errorf("Modern 敘事缺少 catalog 或圖像")
	}
	pages := NarrativePageCount(images)
	if page < 0 || page >= pages {
		return fmt.Errorf("Modern 敘事頁 %d 超出 1..%d", page+1, pages)
	}
	l, err := uilayout.NewModernPageLayout(uilayout.ModernDesignSurface())
	if err != nil {
		return err
	}
	drawModernPageFrame(c, l, style)
	if missing := drawModernTitleScaled(c, fonts, trimHalfCells(title, 54), style.Ink, l.Header.X+24, l.Header.Y+12, 2, 54); len(missing) != 0 {
		return fmt.Errorf("Modern 敘事標題缺字：%q", string(missing))
	}
	start := page * narrativePerPage
	end := start + narrativePerPage
	if end > len(images) {
		end = len(images)
	}
	cardH := (l.Body.H - 3*12) / 4
	for i := start; i < end; i++ {
		row := i - start
		y := l.Body.Y + row*(cardH+12)
		drawModernDecoratedPanel(c, uilayout.Rect{X: l.Body.X, Y: y, W: l.Body.W, H: cardH}, style, style.Panel)
		if err := drawIndexedScaled(c, uitheme.Bitmap{Image: images[i], Palette: assets.EGADefaultPalette},
			uilayout.Rect{X: l.Body.X + 20, Y: y + 18, W: 430, H: 32}, false); err != nil {
			return err
		}
		text, status, ok := catalog.Text(i, mode)
		if !ok {
			text = catalog.Fallback
		}
		if missing := drawModernTextScaled(c, fonts, strconv.Itoa(i)+"  "+text, style.Ink,
			l.Body.X+480, y+30, 2, 70); len(missing) != 0 && status == i18n.NarrativeTranslated {
			return fmt.Errorf("Modern 敘事 #%d 缺字：%q", i, string(missing))
		}
	}
	drawModernTextScaled(c, fonts, fmt.Sprintf("%d/%d　%s", page+1, pages, catalog.Source), style.Muted,
		l.Footer.X+24, l.Footer.Y+12, 1, 100)
	back, previous, next := modernNavigationDefaults(navigation)
	drawModernNav(c, l, fonts, style, back, previous, next)
	return nil
}

func validateModernPageCanvas(c *Canvas, style uitheme.UIStyle) error {
	if c == nil || c.img == nil {
		return fmt.Errorf("Modern page Canvas 不可為 nil")
	}
	if c.img.Bounds().Dx() != uilayout.ModernDesignWidth || c.img.Bounds().Dy() != uilayout.ModernDesignHeight {
		return fmt.Errorf("Modern page 需要 %dx%d Canvas", uilayout.ModernDesignWidth, uilayout.ModernDesignHeight)
	}
	if style.Name != uitheme.ModeModern {
		return fmt.Errorf("Modern page 只接受 modern style，得到 %q", style.Name)
	}
	return nil
}

func drawModernPageFrame(c *Canvas, l uilayout.ModernPageLayout, style uitheme.UIStyle) {
	c.fillRect(0, 0, uilayout.ModernDesignWidth, uilayout.ModernDesignHeight, style.Paper)
	drawModernSatelliteChrome(c, l.Rail, l.Header, style)
	drawModernDecoratedPanel(c, l.Body, style, style.Paper)
	drawModernDecoratedPanel(c, l.Footer, style, style.Panel)
}

// drawModernSatelliteChrome 是 H4 共用的軍事衛星作戰室外殼。它只畫幾何狀態
// 與色彩層級：沒有硬編碼玩家文字、沒有可點 action，也不暗示不存在的遊戲情報。
func drawModernSatelliteChrome(c *Canvas, rail, header uilayout.Rect, style uitheme.UIStyle) {
	c.fillRect(rail.X, rail.Y, rail.W, rail.H, style.Panel)
	c.strokeRect(rail.X, rail.Y, rail.W, rail.H, style.Muted)
	c.fillRect(rail.Right()-4, rail.Y, 4, rail.H, style.AccentAlt)

	// 導覽軌上的方位／資料節點是非互動的視覺錨點。可操作的入口由各畫面
	// 明確再畫，避免把裝飾誤作按鈕。
	for i, y := range []int{rail.Y + 24, rail.Y + 312, rail.Y + 368, rail.Y + 424} {
		size := 16
		if i == 0 {
			size = 28
		}
		x := rail.X + (rail.W-size)/2
		c.strokeRect(x, y, size, size, style.Muted)
		if i == 0 {
			c.fillRect(x+6, y+6, size-12, size-12, style.Focus)
		} else {
			c.fillRect(x+5, y+5, size-10, size-10, style.AccentAlt)
		}
	}

	c.fillRect(header.X, header.Y, header.W, header.H, style.Panel)
	c.strokeRect(header.X, header.Y, header.W, header.H, style.Muted)
	c.fillRect(header.X, header.Y, header.W, 3, style.AccentAlt)
	c.fillRect(header.X, header.Y, 4, header.H, style.Focus)
	for x := header.Right() - 108; x < header.Right()-12; x += 16 {
		c.fillRect(x, header.Bottom()-12, 9, 2, style.Muted)
	}
}

// drawModernSatelliteGrid 為面板邊緣提供低密度掃描格線。它在地圖繪製前使用，
// 不會修改六角格資料或把網格誤標成新的地形／可移動區。
func drawModernSatelliteGrid(c *Canvas, r uilayout.Rect, style uitheme.UIStyle) {
	for x := r.X + 16; x < r.Right()-16; x += 32 {
		c.fillRect(x, r.Y+12, 1, r.H-24, style.Muted)
	}
	for y := r.Y + 16; y < r.Bottom()-16; y += 32 {
		c.fillRect(r.X+12, y, r.W-24, 1, style.Muted)
	}
	c.strokeRect(r.X+12, r.Y+12, r.W-24, r.H-24, style.AccentAlt)
}

// drawModernRailAction 畫既有 action 的第二個、圖像化入口。傳入 false 時只
// 顯示不可用狀態，pointer 也不得為它建立 target。
func drawModernRailAction(c *Canvas, r uilayout.Rect, style uitheme.UIStyle, available, primary bool) {
	border, fill := style.Muted, style.Paper
	if available {
		border = style.AccentAlt
		if primary {
			border, fill = style.Accent, style.Panel
		}
	}
	drawModernDecoratedControl(c, r, style, fill, border)
	if !available {
		c.fillRect(r.X+r.W/2-2, r.Y+r.H/2-2, 4, 4, style.Muted)
		return
	}
	if primary {
		// 三條訊令列：對應既有「開啟指令」入口，沒有額外語意或新 action。
		for y := r.Y + 12; y <= r.Bottom()-14; y += 9 {
			c.fillRect(r.X+11, y, r.W-22, 3, style.Accent)
		}
		return
	}
	// 四格資料頁：對應既有「新聞／史事」閱讀器。
	for _, p := range []struct{ x, y int }{{10, 10}, {25, 10}, {10, 25}, {25, 25}} {
		c.fillRect(r.X+p.x, r.Y+p.y, 10, 10, style.AccentAlt)
	}
}

func drawModernMapRailMarker(c *Canvas, r uilayout.Rect, style uitheme.UIStyle) {
	drawModernDecoratedControl(c, r, style, style.Panel, style.Focus)
	c.strokeRect(r.X+10, r.Y+10, r.W-20, r.H-20, style.AccentAlt)
	c.fillRect(r.X+r.W/2-2, r.Y+12, 4, r.H-24, style.Focus)
	c.fillRect(r.X+12, r.Y+r.H/2-2, r.W-24, 4, style.Focus)
}

func drawModernSignalStrip(c *Canvas, r uilayout.Rect, style uitheme.UIStyle) {
	c.fillRect(r.X, r.Y, r.W, r.H, style.Paper)
	c.strokeRect(r.X, r.Y, r.W, r.H, style.Muted)
	for i := 0; i < 12; i++ {
		w := 10 + (i%4)*8
		color := style.AccentAlt
		if i == 0 || i == 7 {
			color = style.Focus
		}
		c.fillRect(r.X+12+i*30, r.Y+r.H/2-2, w, 4, color)
	}
}

func drawModernNav(c *Canvas, l uilayout.ModernPageLayout, fonts *assets.EtenFonts,
	style uitheme.UIStyle, back, previous, next string) {
	for _, item := range []struct {
		r     uilayout.Rect
		label string
		fill  assets.RGB
	}{
		{l.Back, back, style.Panel}, {l.Previous, previous, style.Panel}, {l.Next, next, style.Panel},
	} {
		if item.label == "" {
			continue
		}
		drawModernDecoratedControl(c, item.r, style, item.fill, style.Accent)
		scale, y, maxHalf := 1, item.r.Y+14, 14
		if len([]rune(item.label)) <= 4 {
			scale, y, maxHalf = 2, item.r.Y+8, 7
		}
		_ = drawModernTextScaled(c, fonts, item.label, style.Ink, item.r.X+20, y, scale, maxHalf)
	}
}

func modernNavigationDefaults(values []string) (string, string, string) {
	back, previous, next := "返回", "上一頁", "下一頁"
	if len(values) > 0 && values[0] != "" {
		back = values[0]
	}
	if len(values) > 1 && values[1] != "" {
		previous = values[1]
	}
	if len(values) > 2 && values[2] != "" {
		next = values[2]
	}
	return back, previous, next
}

func drawModernTextScaled(c *Canvas, fonts *assets.EtenFonts, s string, fg assets.RGB,
	x, y, scale, maxHalf int) []rune {
	return drawModernTextScaledFamily(c, fonts, s, fg, x, y, scale, maxHalf, assets.ModernFontSans)
}

// drawModernTitleScaled 是使用者確認的 C 字體分工入口：Modern 主標題使用
// 宋體；按鈕、內文、數值與導覽仍由 drawModernTextScaled 使用黑體。
func drawModernTitleScaled(c *Canvas, fonts *assets.EtenFonts, s string, fg assets.RGB,
	x, y, scale, maxHalf int) []rune {
	return drawModernTextScaledFamily(c, fonts, s, fg, x, y, scale, maxHalf, assets.ModernFontSerif)
}

func drawModernTextScaledFamily(c *Canvas, fonts *assets.EtenFonts, s string, fg assets.RGB,
	x, y, scale, maxHalf int, family assets.ModernFontFamily) []rune {
	missing := map[rune]bool{}
	drawModernTextScaledIntoFamily(c, fonts, s, fg, x, y, scale, maxHalf, missing, family)
	out := make([]rune, 0, len(missing))
	for r := range missing {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func modernTextHalfCells(s string) int {
	total := 0
	for _, r := range s {
		total += textlayout.RuneHalfCells(r)
	}
	return total
}

// modernButtonLabelLines 是 Modern 主要按鈕的固定 2 倍字級排版入口。
// 先依空白拆成最多兩行；單一長詞或第二行仍超寬時，以明確的 Unicode
// 省略號截短，避免 drawModernTextScaledInto 的 maxHalf 造成無提示的斷字。
func modernButtonLabelLines(label string, maxPixels, scale int) []string {
	label = strings.TrimSpace(label)
	if label == "" || maxPixels <= 0 || scale <= 0 {
		return nil
	}
	if modernTextPixelWidth(label, scale) <= maxPixels {
		return []string{label}
	}
	words := strings.Fields(label)
	if len(words) <= 1 {
		return []string{modernEllipsize(label, maxPixels, scale)}
	}
	lines := make([]string, 0, 2)
	current := ""
	for i, word := range words {
		if current == "" {
			current = word
			continue
		}
		candidate := current + " " + word
		if modernTextPixelWidth(candidate, scale) <= maxPixels {
			current = candidate
			continue
		}
		lines = append(lines, modernEllipsize(current, maxPixels, scale))
		remaining := strings.Join(words[i:], " ")
		lines = append(lines, modernEllipsize(remaining, maxPixels, scale))
		return lines
	}
	if current != "" {
		lines = append(lines, modernEllipsize(current, maxPixels, scale))
	}
	return lines
}

func modernTextPixelWidth(s string, scale int) int {
	if scale <= 0 {
		return 0
	}
	modernFont, _ := assets.EmbeddedModernFont()
	width := 0
	for _, r := range s {
		advance := 8
		if r == ' ' {
			advance = 8
		} else if modernFont != nil {
			if glyph, ok := modernFont.Glyph(r); ok && glyph.Advance() > 0 {
				advance = glyph.Advance()
			} else if textlayout.RuneHalfCells(r) > 1 {
				advance = 16
			}
		} else if textlayout.RuneHalfCells(r) > 1 {
			advance = 16
		}
		width += advance * scale
	}
	return width
}

func modernEllipsize(s string, maxPixels, scale int) string {
	if modernTextPixelWidth(s, scale) <= maxPixels {
		return s
	}
	ellipsis := "…"
	ellipsisWidth := modernTextPixelWidth(ellipsis, scale)
	if ellipsisWidth >= maxPixels {
		return ellipsis
	}
	out := make([]rune, 0, len([]rune(s)))
	used := 0
	for _, r := range s {
		candidate := string(r)
		w := modernTextPixelWidth(candidate, scale)
		if used+w+ellipsisWidth > maxPixels {
			break
		}
		out = append(out, r)
		used += w
	}
	return strings.TrimRight(string(out), " ") + ellipsis
}

func drawModernButtonLabel(c *Canvas, fonts *assets.EtenFonts, label string, fg assets.RGB,
	box uilayout.Rect, missing map[rune]bool) {
	const scale = 2
	lines := modernButtonLabelLines(label, box.W, scale)
	if len(lines) == 0 {
		return
	}
	lineHeight := assets.GlyphH * scale
	totalHeight := len(lines) * lineHeight
	y := box.Y + (box.H-totalHeight)/2
	for _, line := range lines {
		drawModernTextScaledInto(c, fonts, line, fg, box.X, y, scale, 1000, missing)
		y += lineHeight
	}
}

func drawModernTextScaledInto(c *Canvas, fonts *assets.EtenFonts, s string, fg assets.RGB,
	x, y, scale, maxHalf int, missing map[rune]bool) {
	drawModernTextScaledIntoFamily(c, fonts, s, fg, x, y, scale, maxHalf, missing, assets.ModernFontSans)
}

func drawModernTitleScaledInto(c *Canvas, fonts *assets.EtenFonts, s string, fg assets.RGB,
	x, y, scale, maxHalf int, missing map[rune]bool) {
	drawModernTextScaledIntoFamily(c, fonts, s, fg, x, y, scale, maxHalf, missing, assets.ModernFontSerif)
}

func drawModernTextScaledIntoFamily(c *Canvas, fonts *assets.EtenFonts, s string, fg assets.RGB,
	x, y, scale, maxHalf int, missing map[rune]bool, family assets.ModernFontFamily) {
	if scale <= 0 || maxHalf <= 0 {
		return
	}
	// Modern 高解析畫面優先使用隨程式散布的 Unicode atlas：英文／日文
	// 不再依賴倚天字模，ASCII 也能使用比例字距。若 atlas 不存在或損壞，
	// 才退回 Eten（繁中母本仍保留原有字模外觀）。
	modernFont, _ := assets.EmbeddedModernFontFamily(family)
	used := 0
	for _, r := range s {
		width := textlayout.RuneHalfCells(r)
		if used+width > maxHalf {
			break
		}
		if modernFont != nil {
			if glyph, ok := modernFont.Glyph(r); ok {
				drawModernGlyph(c, glyph, fg, x, y, scale)
				advance := glyph.Advance()
				if advance <= 0 {
					advance = width * 8
				}
				x += advance * scale
				used += max(1, (advance+7)/8)
				continue
			}
		}
		if r == ' ' {
			x += 8 * scale
			used += width
			continue
		}
		if r >= 0x20 && r <= 0x7e {
			if fonts == nil {
				missing[r] = true
				drawScaledMissing(c, fg, x, y, 8, scale)
				x += 8 * scale
				used += width
				continue
			}
			glyph, ok := fonts.GlyphASCII(r)
			if !ok {
				missing[r] = true
				drawScaledMissing(c, fg, x, y, 8, scale)
			} else {
				for gy, row := range glyph {
					for gx := 0; gx < 8; gx++ {
						if row&(0x80>>gx) == 0 {
							continue
						}
						c.fillRect(x+gx*scale, y+gy*scale, scale, scale, fg)
					}
				}
			}
			x += 8 * scale
		} else {
			if fonts == nil {
				missing[r] = true
				drawScaledMissing(c, fg, x, y, 16, scale)
				x += 16 * scale
				used += width
				continue
			}
			glyph, ok := fonts.GlyphRune(r)
			if !ok {
				missing[r] = true
				drawScaledMissing(c, fg, x, y, 16, scale)
			} else {
				for gy := 0; gy < assets.GlyphH; gy++ {
					for gx := 0; gx < assets.GlyphW; gx++ {
						if glyph.At(gx, gy) {
							c.fillRect(x+gx*scale, y+gy*scale, scale, scale, fg)
						}
					}
				}
			}
			x += 16 * scale
		}
		used += width
	}
}

func drawModernGlyph(c *Canvas, glyph assets.ModernGlyph, fg assets.RGB, x, y, scale int) {
	if scale <= 0 {
		return
	}
	sourceW, sourceH := glyph.Size()
	targetW, targetH := assets.GlyphW*scale, assets.GlyphH*scale
	if sourceW <= 0 || sourceH <= 0 {
		return
	}
	for gy := 0; gy < targetH; gy++ {
		for gx := 0; gx < targetW; gx++ {
			coverage := modernGlyphCoverage(glyph, gx, gy, targetW, targetH, sourceW, sourceH)
			c.blendPixel(x+gx, y+gy, fg, coverage)
		}
	}
}

func modernGlyphCoverage(glyph assets.ModernGlyph, x, y, targetW, targetH, sourceW, sourceH int) uint8 {
	if targetW >= sourceW && targetH >= sourceH {
		return glyph.AlphaAt(x*sourceW/targetW, y*sourceH/targetH)
	}
	sx0, sy0 := x*sourceW/targetW, y*sourceH/targetH
	sx1 := ((x+1)*sourceW + targetW - 1) / targetW
	sy1 := ((y+1)*sourceH + targetH - 1) / targetH
	if sx1 > sourceW {
		sx1 = sourceW
	}
	if sy1 > sourceH {
		sy1 = sourceH
	}
	var sum, count int
	for sy := sy0; sy < sy1; sy++ {
		for sx := sx0; sx < sx1; sx++ {
			sum += int(glyph.AlphaAt(sx, sy))
			count++
		}
	}
	if count == 0 {
		return 0
	}
	return uint8((sum + count/2) / count)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func drawScaledMissing(c *Canvas, fg assets.RGB, x, y, width, scale int) {
	c.strokeRect(x, y, width*scale, assets.GlyphH*scale, fg)
	c.fillRect(x+(width*scale)/2, y+(assets.GlyphH*scale)/2, scale, scale, fg)
}
