// Package textlayout 提供不依賴 Ebiten 的繁體中文斷行與分頁。
package textlayout

import "fmt"

// Options 使用半格作為內部寬度單位；Columns 則仍以全形格表示。
type Options struct {
	Columns int
	Rows    int
}

// DefaultBiographyOptions 是 DESIGN-20 的 640×350 自傳本文區：28 格 × 13 行。
var DefaultBiographyOptions = Options{Columns: 28, Rows: 13}

// Line 是一行排版結果。HalfCells 可用來精確定位全形／半形混排。
type Line struct {
	Text      string
	HalfCells int
}

// Page 是固定行數的一頁。
type Page struct {
	Lines []Line
}

// Document 是完整的分頁結果。
type Document struct {
	Pages []Page
}

var forbiddenLineStart = map[rune]bool{
	'，': true, '。': true, '、': true, '；': true, '：': true,
	'」': true, '』': true, '）': true, '》': true, '】': true,
	'！': true, '？': true,
}

var forbiddenLineEnd = map[rune]bool{
	'「': true, '『': true, '（': true, '《': true, '【': true,
}

// RuneHalfCells 回傳字元佔用的半格數。數字、拉丁字母、ASCII 符號、間隔點與
// 破折號佔半格；其餘文字與中文標點佔一個全形格。
func RuneHalfCells(r rune) int {
	if r <= 0x7f || r == '·' || r == '—' {
		return 1
	}
	return 2
}

// MeasureHalfCells 量一段文字的半格寬度；換行符不佔寬。
func MeasureHalfCells(s string) int {
	n := 0
	for _, r := range s {
		if r != '\n' && r != '\r' {
			n += RuneHalfCells(r)
		}
	}
	return n
}

// Layout 依固定全形格寬與固定行數斷行、分頁。
// 行首禁則採懸掛標點：若閉標點原本會落到下一行行首，將它附在上一行末端，
// 因此上一行最多可比名義寬度多一個全形標點。
func Layout(text string, opt Options) (Document, error) {
	if opt.Columns <= 0 || opt.Rows <= 0 {
		return Document{}, fmt.Errorf("textlayout: Columns 與 Rows 必須大於 0")
	}
	max := opt.Columns * 2
	var lines []Line
	var current []rune
	width := 0

	flush := func(force bool) {
		if len(current) == 0 && !force {
			return
		}
		lines = append(lines, Line{Text: string(current), HalfCells: width})
		current, width = nil, 0
	}
	for _, r := range text {
		if r == '\r' {
			continue
		}
		if r == '\n' {
			flush(true)
			continue
		}
		rw := RuneHalfCells(r)
		// 開標點不可成為滿行最後一字，否則先換行。
		if forbiddenLineEnd[r] && len(current) > 0 && width+rw >= max {
			flush(false)
		}
		if width+rw > max {
			flush(false)
		}
		// 閉標點不得置於行首；回吸到上一行形成懸掛標點。
		if forbiddenLineStart[r] && len(current) == 0 && len(lines) > 0 {
			lines[len(lines)-1].Text += string(r)
			lines[len(lines)-1].HalfCells += rw
			continue
		}
		current = append(current, r)
		width += rw
	}
	flush(false)
	if len(lines) == 0 {
		lines = append(lines, Line{})
	}

	doc := Document{Pages: make([]Page, 0, (len(lines)+opt.Rows-1)/opt.Rows)}
	for start := 0; start < len(lines); start += opt.Rows {
		end := start + opt.Rows
		if end > len(lines) {
			end = len(lines)
		}
		pageLines := append([]Line(nil), lines[start:end]...)
		doc.Pages = append(doc.Pages, Page{Lines: pageLines})
	}
	return doc, nil
}
