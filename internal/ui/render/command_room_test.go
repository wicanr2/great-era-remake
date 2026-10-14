package render

import (
	"bytes"
	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
	"strings"
	"testing"
	"unicode"
)

func TestRoomTextKeepsEveryResourceDigit(t *testing.T) {
	a, b := NewCanvas(1280, 720), NewCanvas(1280, 720)
	ink := assets.RGB{R: 220, G: 220, B: 220}
	r := l.Rect{254, 29, 126, 32}
	roomText(a, nil, "18050", r, ink, 2)
	drawModernTextScaled(b, nil, "18050", ink, r.X, r.Y+1, 2, 100000)
	if !bytes.Equal(a.Image().Pix, b.Image().Pix) {
		t.Fatal("resource number was truncated")
	}
}

func TestTranslatedRoomControlsKeepTheirFullLabels(t *testing.T) {
	f, err := assets.EmbeddedModernFont()
	if err != nil {
		t.Fatal(err)
	}
	for _, sample := range []struct {
		text string
		box  l.Rect
		size int
	}{
		{"Personnel archive", l.Rect{0, 0, 64, 42}, 15},
		{"Save game", l.Rect{0, 0, 64, 42}, 15},
		{"UI: Command room", l.Rect{0, 0, 108, 34}, 18},
		{"Diplomacy", l.Rect{0, 0, 76, 36}, 18},
		{"美術：オリジナル", l.Rect{0, 0, 88, 34}, 14},
		{"人物檔案", l.Rect{0, 0, 64, 42}, 15},
	} {
		lines, size := roomLabelPlan(f, sample.text, sample.box, sample.size)
		if strings.Contains(strings.Join(lines, ""), "...") || strings.ReplaceAll(strings.Join(lines, ""), " ", "") != strings.ReplaceAll(sample.text, " ", "") {
			t.Fatalf("truncated control %q: %v", sample.text, lines)
		}
		for _, line := range lines {
			if roomLabelWidth(f, line, size) > sample.box.W {
				t.Fatalf("overflow: %q", line)
			}
		}
		if len(lines)*(size+2)-2 > sample.box.H {
			t.Fatalf("vertical overflow: %q", sample.text)
		}
	}
}

func TestRoomArchivePreservesAllBiographyText(t *testing.T) {
	for _, language := range []string{"zh-Hant", "en", "ja"} {
		t.Run(language, func(t *testing.T) { testRoomArchivePreservesAllBiographyText(t, language) })
	}
}

func testRoomArchivePreservesAllBiographyText(t *testing.T, language string) {
	db, err := i18n.LoadPeople("../../../translations/"+language, "../../../translations/shared")
	if err != nil {
		t.Fatal(err)
	}
	clean := func(s string) string {
		return strings.Map(func(r rune) rune {
			if unicode.IsSpace(r) {
				return -1
			}
			return r
		}, s)
	}
	full, unknown := 0, 0
	for id := 1; id <= 417; id++ {
		if id == 274 {
			continue
		}
		p, ok := db.PersonByID(id)
		if !ok {
			t.Fatalf("missing person %d", id)
		}
		doc, err := RoomBiographyDocument(p, "待考", "資料來源")
		if err != nil {
			t.Fatal(err)
		}
		var lines []string
		for _, page := range doc.Pages {
			if len(page.Lines) > 15 {
				t.Fatal("overflow")
			}
			for _, line := range page.Lines {
				if line.PixelWidth > 508 || modernEllipsize(line.Text, 512, 1) != line.Text {
					t.Fatalf("人物 %d 的正文被繪製截短：%q，寬 %d", id, line.Text, line.PixelWidth)
				}
				lines = append(lines, line.Text)
			}
		}
		text := clean(strings.Join(lines, ""))
		if strings.TrimSpace(p.Biography) == "" {
			unknown++
			if !strings.Contains(text, "待考") {
				t.Fatalf("unknown not identified %d", id)
			}
		} else {
			full++
			if !strings.HasPrefix(text, clean(p.Biography)) {
				t.Fatalf("truncated biography %d", id)
			}
		}
		for _, src := range p.Sources {
			if !strings.Contains(text, clean(src)) {
				t.Fatalf("missing source %d", id)
			}
		}
	}
	if full != 387 || unknown != 29 {
		t.Fatalf("coverage %d full / %d unknown", full, unknown)
	}
}

func TestRoomArchiveKeepsPhotoAttributionAndMarkerHitVisibility(t *testing.T) {
	p := &i18n.Person{Biography: strings.Repeat("歷史正文。", 200), Sources: []string{"原有來源"}}
	photo := &i18n.Portrait{Attribution: "照片作者與年代", SourcePage: "https://example.com/photo", LicenseURL: "https://example.com/rights"}
	doc, err := RoomArchiveDocument(&RoomArchiveData{Selected: p, Portrait: photo, Sources: "資料來源"})
	if err != nil {
		t.Fatal(err)
	}
	text := ""
	for _, page := range doc.Pages {
		for _, line := range page.Lines {
			text += line.Text
		}
	}
	for _, want := range []string{p.Biography, "原有來源", photo.Attribution, photo.SourcePage, photo.LicenseURL} {
		if !strings.Contains(text, want) {
			t.Fatalf("lost text: %s", want)
		}
	}
	if len(p.Sources) != 1 {
		t.Fatal("archive modified source data")
	}
	markers := VisibleRoomMarkers([]RoomMarker{{ID: 1, X: 800, Y: 300}, {ID: 2, X: 810, Y: 305, Selected: true}, {ID: 3, X: 1100, Y: 620}, {ID: 4, X: 950, Y: 400}})
	if len(markers) != 2 || markers[0].ID != 2 || markers[1].ID != 4 {
		t.Fatalf("overlap/occlusion mismatch: %+v", markers)
	}
}
