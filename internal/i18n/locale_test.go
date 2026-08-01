package i18n

import (
	"path/filepath"
	"testing"
)

func load(t *testing.T) *Locale {
	t.Helper()
	l, err := Load(filepath.Join("..", "..", "translations", "zh-Hant"))
	if err != nil {
		t.Fatal(err)
	}
	return l
}

// 省名要對得上原版：39 條，而且已知的幾個要正確。
func TestProvinceNames(t *testing.T) {
	l := load(t)
	cases := map[int]string{
		1: "嫩江省", 11: "河北省", 32: "西藏", 37: "臺灣省", 38: "海南島", 39: "緬甸",
	}
	for id, want := range cases {
		if got := l.Province(id); got != want {
			t.Errorf("省 %d 的名字是 %q，預期 %q", id, got, want)
		}
	}
	// 超出範圍要看得出來，不能回空字串。
	if got := l.Province(0); got != "省 0" {
		t.Errorf("省 0 回 %q，應該回可辨識的字串", got)
	}
	if got := l.Province(40); got != "省 40" {
		t.Errorf("省 40 回 %q，應該回可辨識的字串", got)
	}
}

// 原版排版要保留：「西藏」在畫面上是「西　藏」。
func TestRawKeepsOriginalLayout(t *testing.T) {
	l := load(t)
	raw, ok := l.RawEntry("3.15", 32)
	if !ok {
		t.Fatal("取不到第 32 條")
	}
	if raw != "西　藏" {
		t.Errorf("raw = %q，預期 %q（原版是三格分散排版）", raw, "西　藏")
	}
	if []rune(raw)[1] != '　' {
		t.Error("中間那格應該是全形空格")
	}
}

// 索引是 1-based，與原版的字模索引一致（`docs/re/24`）。
func TestEntryIsOneBased(t *testing.T) {
	l := load(t)
	if s, ok := l.Entry("2.15", 1); !ok || s != "調動" {
		t.Errorf("2.15 第 1 條 = %q（ok=%v），預期「調動」", s, ok)
	}
	if _, ok := l.Entry("2.15", 0); ok {
		t.Error("索引 0 不該取得到——那會讓 1-based 與 0-based 混用")
	}
	if _, ok := l.Entry("沒這個表", 1); ok {
		t.Error("不存在的表不該回 ok")
	}
}

// 詞條數要對得上 `docs/formats/01-glyph-text.md` §4 的表。
func TestTableSizes(t *testing.T) {
	l := load(t)
	want := map[string]int{
		"1.15": 79, "2.15": 231, "3.15": 57, "4.15": 51,
		"FAN(1).15": 47, "JAPAN1.15": 85,
		"MAN115": 274, "MAN215": 106, "MAN315": 106,
	}
	for table, n := range want {
		if got := l.Len(table); got != n {
			t.Errorf("%s 有 %d 條，文件說 %d 條", table, got, n)
		}
	}
	if l.Len("TN15.1") == 0 {
		t.Error("39 個省的地名表應該都在")
	}
}
