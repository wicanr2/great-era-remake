package render

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	"github.com/wicanr2/great-era-remake/internal/ui/textlayout"
)

// BiographyView 是自傳頁所需的語系化顯示資料。Title 與 Unavailable 由
// wording catalog 提供，繪圖層不把某一套用語當成程式判斷條件。
type BiographyView struct {
	Person      *i18n.Person
	Page        int // 0-based
	Title       string
	Unavailable string
}

// BiographyRenderResult 回報實際頁數與無法由字庫呈現的字。Missing 永遠排序且
// 去重；呼叫端可把它寫入診斷訊息，不能把缺字安靜畫成空白。
type BiographyRenderResult struct {
	PageCount int
	Missing   []rune
}

const biographyHalfAdvance = 10

// DrawBiography 畫 640×350 的整頁人物自傳。這是 remake 新增畫面，座標不是
// 原版行為；文字仍沿用倚天 15 點字模與原版的水平加粗方式。
func (c *Canvas) DrawBiography(fonts *assets.EtenFonts, v BiographyView,
	fg, bg assets.RGB) (BiographyRenderResult, error) {
	if fonts == nil || v.Person == nil {
		return BiographyRenderResult{}, fmt.Errorf("render: 自傳缺少人物或字庫")
	}
	p := v.Person
	body := p.Biography
	if strings.TrimSpace(body) == "" {
		body = v.Unavailable
	}
	doc, err := textlayout.Layout(body, textlayout.DefaultBiographyOptions)
	if err != nil {
		return BiographyRenderResult{}, err
	}
	if v.Page < 0 || v.Page >= len(doc.Pages) {
		return BiographyRenderResult{}, fmt.Errorf("render: 自傳頁 %d 超出 1..%d", v.Page+1, len(doc.Pages))
	}

	c.fillRect(0, 0, ModeBGIW, ModeBGIH, bg)
	missing := map[rune]bool{}
	draw := func(s string, x, y, maxHalf int) {
		c.drawBiographyText(fonts, trimHalfCells(s, maxHalf), fg, x, y, missing)
	}

	name := p.NameInGame
	if p.NameCommon != "" && p.NameCommon != p.NameInGame {
		name += "／" + p.NameCommon
	}
	if p.Courtesy != "" {
		name += "　字" + p.Courtesy
	}
	draw(v.Title+"｜"+name, 20, 10, 60)

	var meta []string
	if p.Birth != nil || p.Death != nil {
		birth, death := "？", "？"
		if p.Birth != nil {
			birth = strconv.Itoa(*p.Birth)
		}
		if p.Death != nil {
			death = strconv.Itoa(*p.Death)
		}
		meta = append(meta, birth+"—"+death)
	}
	for _, value := range []string{p.Birthplace, p.Faction} {
		if value != "" {
			meta = append(meta, value)
		}
	}
	draw(strings.Join(meta, "　"), 20, 34, 60)
	draw(p.HighestPost, 20, 54, 60)
	c.strokeRect(20, 76, 600, 1, fg)

	for i, line := range doc.Pages[v.Page].Lines {
		c.drawBiographyText(fonts, line.Text, fg, 40, 84+i*18, missing)
	}
	c.strokeRect(20, 326, 600, 1, fg)
	confidence := biographyConfidenceLabel(p.Confidence)
	footer := fmt.Sprintf("資料來源：%d筆　可靠度：%s", len(p.Sources), confidence)
	draw(footer, 20, 332, 48)
	page := fmt.Sprintf("%d/%d", v.Page+1, len(doc.Pages))
	draw(page, 580, 332, 6)

	result := BiographyRenderResult{PageCount: len(doc.Pages)}
	for r := range missing {
		result.Missing = append(result.Missing, r)
	}
	sort.Slice(result.Missing, func(i, j int) bool { return result.Missing[i] < result.Missing[j] })
	return result, nil
}

func biographyConfidenceLabel(value string) string {
	if label := map[string]string{
		"confirmed": "確認", "partial": "部分資料", "unknown": "未知",
		// 保留研究層可能使用的細分等級，避免未來加入發行檔時漏出代碼。
		"high": "高", "medium": "中", "low": "低",
	}[value]; label != "" {
		return label
	}
	return "未標示"
}

func trimHalfCells(s string, max int) string {
	if max <= 0 {
		return ""
	}
	var out []rune
	used := 0
	for _, r := range s {
		w := textlayout.RuneHalfCells(r)
		if used+w > max {
			break
		}
		out, used = append(out, r), used+w
	}
	return string(out)
}

func (c *Canvas) drawBiographyText(fonts *assets.EtenFonts, s string, fg assets.RGB,
	x, y int, missing map[rune]bool) {
	for _, r := range s {
		advance := textlayout.RuneHalfCells(r) * biographyHalfAdvance
		switch r {
		case ' ':
		case '—':
			c.fillRect(x, y+7, 8, 1, fg)
		case '·':
			c.fillRect(x+3, y+7, 2, 2, fg)
		default:
			if r >= 0x20 && r <= 0x7e {
				if glyph, ok := fonts.GlyphASCII(r); ok {
					c.drawASCIIGlyph(glyph, fg, x, y)
				} else {
					missing[r] = true
					c.drawMissingGlyph(fg, x, y, 8)
				}
			} else if glyph, ok := fonts.GlyphRune(r); ok {
				c.DrawGlyph(glyph, fg, x, y, true)
			} else {
				missing[r] = true
				c.drawMissingGlyph(fg, x, y, 16)
			}
		}
		x += advance
	}
}

// DrawSemanticText 以自傳頁同一套全形 20 px／半形 10 px 網格繪製任意語系文字。
// 這是現代白話 UI 的 bitmap 路徑；回傳排序去重的缺字，呼叫端不得安靜忽略。
func (c *Canvas) DrawSemanticText(fonts *assets.EtenFonts, s string, fg assets.RGB,
	x, y int) []rune {
	missing := map[rune]bool{}
	c.drawBiographyText(fonts, s, fg, x, y, missing)
	out := make([]rune, 0, len(missing))
	for r := range missing {
		out = append(out, r)
	}
	sort.Slice(out, func(i, j int) bool { return out[i] < out[j] })
	return out
}

func (c *Canvas) drawASCIIGlyph(g [assets.EtenASCIIStride]byte, fg assets.RGB, x, y int) {
	for gy, row := range g {
		for gx := 0; gx < 8; gx++ {
			if row&(0x80>>gx) != 0 {
				c.setPixel(x+gx, y+gy, fg)
			}
		}
	}
}

func (c *Canvas) drawMissingGlyph(fg assets.RGB, x, y, width int) {
	c.strokeRect(x, y, width, assets.GlyphH, fg)
	c.setPixel(x+width/2, y+assets.GlyphH/2, fg)
}
