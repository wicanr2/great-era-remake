package assets

import (
	"bufio"
	"encoding/hex"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
)

func TestEtenBig5IndexMatchesPythonOracleForEveryLegalPair(t *testing.T) {
	f, err := os.Open("testdata/eten-index-oracle.tsv")
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	s := bufio.NewScanner(f)
	count := 0
	for s.Scan() {
		line := s.Text()
		if strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 3 {
			t.Fatalf("壞 oracle 列：%q", line)
		}
		pair, err := hex.DecodeString(fields[0])
		if err != nil || len(pair) != 2 {
			t.Fatalf("壞 Big5：%q", fields[0])
		}
		kind, idx, ok := EtenBig5Index(pair[0], pair[1])
		if fields[1] == "-" {
			if ok {
				t.Errorf("%s：Go 得 %s/%d，Python 無字模", fields[0], kind, idx)
			}
		} else {
			wantIdx, _ := strconv.Atoi(fields[2])
			if !ok || kind.String() != fields[1] || idx != wantIdx {
				t.Errorf("%s：Go=%s/%d/%v，Python=%s/%d", fields[0], kind, idx, ok, fields[1], wantIdx)
			}
		}
		count++
	}
	if err := s.Err(); err != nil {
		t.Fatal(err)
	}
	if count != 13973 {
		t.Fatalf("oracle 組合數 = %d，預期 13973", count)
	}
}

func TestEtenBig5IndexRejectsInvalidBytesAndHoles(t *testing.T) {
	for _, pair := range [][2]byte{{0xa0, 0x40}, {0xfa, 0x40}, {0xa1, 0x3f},
		{0xa1, 0x7f}, {0xa1, 0xa0}, {0xa1, 0xff}, {0xa3, 0xc0}, {0xc6, 0xa1}} {
		if kind, idx, ok := EtenBig5Index(pair[0], pair[1]); ok {
			t.Errorf("%X 應拒絕，卻得到 %s/%d", pair, kind, idx)
		}
	}
}

func TestLoadEtenFontsAndSampleGlyphBytes(t *testing.T) {
	dir := filepath.Join("..", "..", "workplace", "eten")
	fonts, err := LoadEtenFonts(dir)
	if err != nil {
		if os.IsNotExist(rootCause(err)) {
			t.Skipf("沒有使用者倚天字庫：%v", err)
		}
		t.Fatal(err)
	}
	std, err := os.ReadFile(filepath.Join(dir, "STDFONT.15"))
	if err != nil {
		t.Fatal(err)
	}
	spc, err := os.ReadFile(filepath.Join(dir, "SPCFONT.15"))
	if err != nil {
		t.Fatal(err)
	}
	asc, err := os.ReadFile(filepath.Join(dir, "ASCFONT.15"))
	if err != nil {
		t.Fatal(err)
	}

	// 固定步距抽超過 100 個合法碼點；期望 byte 直接取原字庫對應 offset。
	checked := 0
	for hi := 0xa1; hi <= 0xf9 && checked < 120; hi++ {
		for lo := 0x40; lo <= 0xfe && checked < 120; lo += 17 {
			kind, idx, ok := EtenBig5Index(byte(hi), byte(lo))
			if !ok {
				continue
			}
			got, ok := fonts.GlyphBig5(byte(hi), byte(lo))
			if !ok {
				t.Fatalf("%02X%02X 有索引卻取不到字模", hi, lo)
			}
			buf := std
			if kind == EtenSymbol {
				buf = spc
			}
			want := buf[idx*GlyphStride : (idx+1)*GlyphStride]
			if string(got[:]) != string(want) {
				t.Fatalf("%02X%02X 的 30 byte 字模不符", hi, lo)
			}
			checked++
		}
	}
	if checked < 100 {
		t.Fatalf("只抽到 %d 個字模", checked)
	}
	for _, r := range []rune{'0', 'A', 'z', '?'} {
		got, ok := fonts.GlyphASCII(r)
		want := asc[int(r)*EtenASCIIStride : (int(r)+1)*EtenASCIIStride]
		if !ok || string(got[:]) != string(want) {
			t.Errorf("ASCII %q 的 15 byte 字模不符", r)
		}
	}
}

func TestLoadEtenFontsRejectsMissingAndWrongSizes(t *testing.T) {
	if _, err := LoadEtenFonts(t.TempDir()); err == nil {
		t.Fatal("缺字庫應回錯")
	}
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "STDFONT.15"), []byte{1}, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "SPCFONT.15"), make([]byte, EtenSymbolGlyphs*GlyphStride), 0o600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadEtenFonts(dir); err == nil || !strings.Contains(err.Error(), "大小 1") {
		t.Fatalf("錯誤尺寸應明確回報，得到 %v", err)
	}
}

func TestEtenGlyphAtBounds(t *testing.T) {
	f := &EtenFonts{standard: make([]Glyph, 1), symbol: make([]Glyph, 1),
		ascii: make([][EtenASCIIStride]byte, EtenASCIIGlyphs)}
	for _, tc := range []struct {
		kind EtenFontKind
		idx  int
	}{{0, 0}, {EtenStandard, -1}, {EtenStandard, 1}, {EtenSymbol, 1}} {
		if _, ok := f.GlyphAt(tc.kind, tc.idx); ok {
			t.Errorf("GlyphAt(%v,%d) 應拒絕", tc.kind, tc.idx)
		}
	}
	for _, r := range []rune{-1, '\n', 0x7f, '中'} {
		if _, ok := f.GlyphASCII(r); ok {
			t.Errorf("GlyphASCII(%q) 應拒絕", r)
		}
	}
}

func TestEtenGlyphRuneUsesBig5AndRejectsUnmapped(t *testing.T) {
	dir := filepath.Join("..", "..", "workplace", "eten")
	fonts, err := LoadEtenFonts(dir)
	if err != nil {
		t.Skipf("沒有使用者倚天字庫：%v", err)
	}
	for r, pair := range map[rune][2]byte{
		'中': {0xa4, 0xa4}, '，': {0xa1, 0x41},
		// WHATWG encoder 會選 FA66；倚天可用的傳統 Big5 碼是 B0B0。
		'偽': {0xb0, 0xb0},
	} {
		got, ok := fonts.GlyphRune(r)
		want, okBig5 := fonts.GlyphBig5(pair[0], pair[1])
		if !ok || !okBig5 || got != want {
			t.Errorf("%c 的 Unicode／Big5 字模未對上", r)
		}
	}
	for _, r := range []rune{'A', '1', '榘', '藁'} {
		if _, ok := fonts.GlyphRune(r); ok {
			t.Errorf("%q 不應由倚天全形字模安靜處理", r)
		}
	}
}

func rootCause(err error) error {
	for {
		u, ok := err.(interface{ Unwrap() error })
		if !ok || u.Unwrap() == nil {
			return err
		}
		err = u.Unwrap()
	}
}
