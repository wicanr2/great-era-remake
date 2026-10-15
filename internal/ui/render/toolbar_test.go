package render

import (
	"strings"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/i18n"
	l "github.com/wicanr2/great-era-remake/internal/ui/layout"
)

func TestToolbarLocalesHaveEmbeddedGlyphsAndCompleteNotes(t *testing.T) {
	f, err := assets.EmbeddedModernFontFamily(assets.ModernFontSans)
	if err != nil {
		t.Fatal(err)
	}
	for _, lang := range []string{"zh-Hant", "en", "ja"} {
		catalog, err := i18n.LoadWording("../../../translations/" + lang)
		if err != nil {
			t.Fatal(err)
		}
		for key, e := range catalog.Entries {
			if !strings.HasPrefix(key, "toolbar.") {
				continue
			}
			for _, r := range e.Original {
				if r == '\n' {
					continue
				}
				if _, ok := f.Glyph(r); !ok {
					t.Errorf("%s %s missing glyph %q", lang, key, r)
				}
			}
		}
		for i, key := range []string{"toolbar.art", "toolbar.interface", "toolbar.ai", "toolbar.music", "toolbar.input"} {
			label := catalog.Entries[key].Original
			name, value := toolbarTextRects(l.ToolbarField(i), label)
			lines, size := roomLabelPlan(f, label, name, 13)
			if len(lines) != 1 || lines[0] != label || size != 13 || value.W < 50 {
				t.Fatalf("%s toolbar label truncated or wrapped: %q %v size=%d value=%+v", lang, label, lines, size, value)
			}
		}
		note := catalog.Entries["toolbar.ai.note.1"].Original + "\n" + catalog.Entries["toolbar.ai.note.2"].Original + "\n" + catalog.Entries["toolbar.ai.note.3"].Original
		lines := ToolbarNoteLines(note, 306)
		if len(lines) < 3 {
			t.Fatalf("note lost paragraphs: %s %v", lang, lines)
		}
	}
}

func TestToolbarMenuKeepsGamePixelsTransparent(t *testing.T) {
	v := ToolbarView{Menu: 2, Hover: -1, Focus: -1, Options: [l.ToolbarFields][]string{{"A", "B"}, {"A", "B"}, {"Current AI"}, {"Mouse", "Keyboard"}}, Notes: []string{"AI help"}}
	c := NewCanvas(1280, 250)
	c.DrawToolbar(nil, v)
	if c.Image().RGBAAt(20, 100).A != 0 {
		t.Fatal("menu covered unrelated game pixels")
	}
	if c.Image().RGBAAt(l.ToolbarField(0).X+2, 10).A != 255 {
		t.Fatal("permanent bar disappeared")
	}
	r := v.MenuRect()
	field, row := v.Hit(r.X+20, r.Bottom()-5)
	if field != 2 || row != -2 {
		t.Fatal("help text became a game option")
	}
}
