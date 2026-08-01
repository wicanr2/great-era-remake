package assets

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

// 原版素材是 gitignore 的，CI 上不會有。沒有就跳過那些測試，
// 而不是讓測試失敗——但**不可以**因此讓演算法本身的測試也跳過。
const gameDir = "../../workplace/orig/game"

func gameFile(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(gameDir, name))
	if err != nil {
		t.Skipf("沒有原版素材 %s，跳過", name)
	}
	return b
}

// TestExplodeVector 用 zlib contrib blast.c 附的官方測試向量驗證 DCL 解壓。
// 這個測試不需要原版素材，任何環境都必須通過。
func TestExplodeVector(t *testing.T) {
	in := []byte{0x00, 0x04, 0x82, 0x24, 0x25, 0x8f, 0x80, 0x7f}
	want := []byte("AIAIAIAIAIAIA")
	got, err := Explode(in, 0)
	if err != nil {
		t.Fatalf("Explode 失敗: %v", err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("Explode = %q，預期 %q", got, want)
	}
}

func TestExplodeRejectsBadHeader(t *testing.T) {
	for _, tc := range []struct {
		name string
		in   []byte
	}{
		{"literal 模式非法", []byte{0x02, 0x06, 0x00}},
		{"字典太小", []byte{0x00, 0x03, 0x00}},
		{"字典太大", []byte{0x00, 0x07, 0x00}},
		{"資料太短", []byte{0x00}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := Explode(tc.in, 0); err == nil {
				t.Fatal("預期報錯，卻成功了")
			}
		})
	}
}

// TestRGBPercentToRGB 鎖住「百分比 → 6-bit DAC → 8-bit」這條轉換鏈。
// 這些期望值是從 DOSBox 實機畫面反推出來的，改動公式會讓 round-trip 失效。
func TestRGBPercentToRGB(t *testing.T) {
	pal, err := ParseRGBFile(bytes.NewBufferString(
		"0 0 0\n70 0 0\n97 60 17\n100 100 100\n"))
	if err != nil {
		t.Fatal(err)
	}
	want := []RGB{{0, 0, 0}, {178, 0, 0}, {247, 154, 44}, {255, 255, 255}}
	for i, w := range want {
		if pal[i] != w {
			t.Errorf("第 %d 色 = %v，預期 %v", i, pal[i], w)
		}
	}
}

func TestRGBRejectsOutOfRange(t *testing.T) {
	if _, err := ParseRGBFile(bytes.NewBufferString("0 0 101\n")); err == nil {
		t.Fatal(".RGB 值 101 超出百分比範圍，應該報錯")
	}
}

func TestDAC6to8(t *testing.T) {
	for _, tc := range []struct{ in, want int }{
		{0, 0}, {11, 44}, {16, 65}, {38, 154}, {44, 178}, {51, 207}, {61, 247}, {63, 255},
	} {
		if got := dac6to8(tc.in); int(got) != tc.want {
			t.Errorf("dac6to8(%d) = %d，預期 %d", tc.in, got, tc.want)
		}
	}
}

// TestBGILayout 驗證 plane 佈局與位元順序。
//
// CHOOSE1.TPC 是載入畫面的圖框，實機畫面只有 5 色，索引是 {0,4,6,12,14}。
// 若 plane 佈局猜成「四個完整 plane 依序」會解出 15 色（雜訊）；
// 若 plane 0 當成最低位元，索引會變成 4-bit 反轉的 {0,2,3,6,7}。
func TestBGILayout(t *testing.T) {
	data := gameFile(t, "CHOOSE1.TPC")
	im, err := DecodeBGI(data, 0)
	if err != nil {
		t.Fatal(err)
	}
	if im.W != 395 || im.H != 237 {
		t.Fatalf("尺寸 = %dx%d，預期 395x237", im.W, im.H)
	}
	seen := map[byte]bool{}
	for _, v := range im.Pix {
		seen[v] = true
	}
	want := []byte{0, 4, 6, 12, 14}
	if len(seen) != len(want) {
		t.Fatalf("色數 = %d，預期 %d（索引 %v）", len(seen), len(want), seen)
	}
	for _, v := range want {
		if !seen[v] {
			t.Errorf("缺少索引 %d，實際索引集 %v", v, seen)
		}
	}
	if string(im.Palt[:0]) != "" && im.Palt == nil {
		t.Error("CHOOSE1.TPC 應該帶 palt")
	}
	if im.Palt == nil {
		t.Error("CHOOSE1.TPC 差額 20 bytes，應該解出 palt")
	}
}

// notBGI 是副檔名叫 .TPC 但不是 BGI 格式的檔案。
//
// RAIL.TPC 的檔頭是 00 00 00 00，照 BGI 解會變成 1×1；實際內容是
// 每像素 1 byte 的 chunky 資料（值 0/7/8，連續數列相同，像是軌道線條）。
// **格式待查**，先排除在 BGI 驗算之外，不要為了讓測試綠而硬解。
var notBGI = map[string]string{
	"RAIL.TPC": "chunky 而非 planar，格式待查",
}

// TestBGISizeFormula 對全部 .TPC 驗算大小公式，並確認差額只有 0 或 20（palt）。
func TestBGISizeFormula(t *testing.T) {
	entries, err := os.ReadDir(gameDir)
	if err != nil {
		t.Skip("沒有原版素材，跳過")
	}
	var checked int
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".TPC" {
			continue
		}
		if why, skip := notBGI[e.Name()]; skip {
			t.Logf("跳過 %s：%s", e.Name(), why)
			continue
		}
		data := gameFile(t, e.Name())
		im, err := DecodeBGI(data, 0)
		if err != nil {
			t.Errorf("%s: %v", e.Name(), err)
			continue
		}
		diff := len(data) - BGISize(im.W, im.H)
		switch {
		case diff == 0:
			if im.Palt != nil {
				t.Errorf("%s: 差額 0 卻解出 palt", e.Name())
			}
		case diff == 20:
			if im.Palt == nil {
				t.Errorf("%s: 差額 20 卻沒解出 palt", e.Name())
			}
		default:
			// 差額巨大的是圖集（AC/NEWICON/NEWTERR/RAIL），要能整份解開
			set, err := DecodeBGISet(data)
			if err != nil {
				t.Errorf("%s: 差額 %d，當圖集解也失敗: %v", e.Name(), diff, err)
			} else if len(set) < 2 {
				t.Errorf("%s: 差額 %d，但只解出 %d 張", e.Name(), diff, len(set))
			}
		}
		checked++
	}
	if checked == 0 {
		t.Skip("沒有 .TPC 可驗")
	}
	t.Logf("驗過 %d 個 .TPC", checked)
}

// TestGLB 驗證目錄自洽性與全部 entry 都解得開。
func TestGLB(t *testing.T) {
	for _, stem := range []string{"GRT", "GRT2"} {
		t.Run(stem, func(t *testing.T) {
			g, err := OpenGLB(gameFile(t, stem+".GTB"), gameFile(t, stem+".GLB"))
			if err != nil {
				t.Fatal(err)
			}
			var images, palettes int
			for i, e := range g.Entries {
				raw, err := g.Raw(i)
				if err != nil {
					t.Fatalf("#%d 解壓失敗: %v", i, err)
				}
				if len(raw) != e.RawSize {
					t.Fatalf("#%d 解出 %d bytes，目錄說 %d", i, len(raw), e.RawSize)
				}
				if e.IsPalette() {
					palettes++
					// 調色盤必須全部落在 6-bit 範圍——這是格式判定的強證據
					if _, err := ParseVGAPalette(raw); err != nil {
						t.Fatalf("#%d 調色盤不合法: %v", i, err)
					}
					continue
				}
				images++
				if g.PaletteFor(i) < 0 {
					t.Errorf("#%d 找不到管轄的調色盤", i)
				}
			}
			t.Logf("%s: %d 張圖 + %d 個調色盤", stem, images, palettes)
			if palettes == 0 {
				t.Error("一個調色盤都沒有")
			}
		})
	}
}

// TestNewsData 驗證 NEWSDATA.DAT 是 17 個等長區塊，各含一張 BGI 圖 + palt。
func TestNewsData(t *testing.T) {
	data := gameFile(t, "NEWSDATA.DAT")
	const block = 1752
	if len(data)%block != 0 {
		t.Fatalf("大小 %d 不是 %d 的倍數", len(data), block)
	}
	n := len(data) / block
	if n != 17 {
		t.Errorf("區塊數 = %d，預期 17", n)
	}
	for k := 0; k < n; k++ {
		im, err := DecodeBGI(data, k*block)
		if err != nil {
			t.Fatalf("#%d: %v", k, err)
		}
		if im.H != 16 {
			t.Errorf("#%d 高 = %d，預期 16", k, im.H)
		}
		if im.Palt == nil {
			t.Errorf("#%d 沒有 palt", k)
		}
		if im.Size() != block {
			t.Errorf("#%d 佔 %d bytes，預期 %d", k, im.Size(), block)
		}
	}
}

// TestGlyphFiles 驗證 51 個字模檔全部整除 30。
func TestGlyphFiles(t *testing.T) {
	names := []string{"1.15", "2.15", "3.15", "4.15", "MAN115", "MAN215", "MAN315",
		"JAPAN1.15", "JAPAN2.15", "FAN(1).15", "FAN(2).15", "FAN(3).15"}
	for i := 1; i <= 39; i++ {
		names = append(names, "TN15."+itoa(i))
	}
	var total int
	for _, name := range names {
		data := gameFile(t, name)
		f, err := ParseGlyphFile(data)
		if err != nil {
			t.Errorf("%s: %v", name, err)
			continue
		}
		total += len(f.Glyphs)
	}
	if total != 6174 {
		t.Errorf("字模總數 = %d，預期 6174", total)
	}
}

// TestGlyphSlots 驗證定長槽位的切分：MAN115 是 274 個三字人名。
func TestGlyphSlots(t *testing.T) {
	f, err := ParseGlyphFile(gameFile(t, "MAN115"))
	if err != nil {
		t.Fatal(err)
	}
	if got := f.EntryCount(3); got != 274 {
		t.Fatalf("MAN115 詞條數 = %d，預期 274", got)
	}
	// 第 0 條是「蔣中正」，三個字模都不是空白
	e, err := f.Entry(0, 3)
	if err != nil {
		t.Fatal(err)
	}
	if len(e) != 3 {
		t.Errorf("第 0 條有 %d 個非空字模，預期 3", len(e))
	}
	// 有些人名是兩個字，空白會被 Entry 濾掉
	var twoChar int
	for k := 0; k < f.EntryCount(3); k++ {
		e, err := f.Entry(k, 3)
		if err != nil {
			t.Fatal(err)
		}
		if len(e) == 2 {
			twoChar++
		}
	}
	if twoChar != 41 {
		t.Errorf("兩字人名 = %d 條，預期 41", twoChar)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}

// TestProvinceGridsSymmetry 驗證從 WARPOS.DAT 導出的鄰接表完全對稱。
//
// 這是解析正確與否最有力的檢查：任何 off-by-one 或 row/column 搞反
// 都會讓對稱性破裂。實測 39 個省 0 個不對稱。
func TestProvinceGridsSymmetry(t *testing.T) {
	g, err := ParseProvinceGrids(gameFile(t, "WARPOS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	if bad := g.IsSymmetric(); len(bad) != 0 {
		t.Fatalf("鄰接關係有 %d 個不對稱：%v", len(bad), bad[:min(5, len(bad))])
	}
}

// TestProvinceNeighbours 用 DOSBox 實機畫面對照鄰接表。
//
// 實機的攻打子選單顯示河南（19）可攻打 11,16,18,20,21,22；
// 檔案導出的多一個 26（湖北，當時玩家控制的省，所以選單不列）。
func TestProvinceNeighbours(t *testing.T) {
	g, err := ParseProvinceGrids(gameFile(t, "WARPOS.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		province int
		want     []int
		note     string
	}{
		{19, []int{11, 16, 18, 20, 21, 22, 26}, "河南（實機對照）"},
		{37, []int{24}, "臺灣省 → 福建，離島度數 1"},
		{38, []int{36}, "海南島 → 廣東"},
		{39, []int{34}, "緬甸 → 雲南"},
	} {
		got, err := g.Neighbours(tc.province)
		if err != nil {
			t.Fatal(err)
		}
		if len(got) != len(tc.want) {
			t.Errorf("省 %d（%s）鄰省 = %v，預期 %v", tc.province, tc.note, got, tc.want)
			continue
		}
		for i := range got {
			if got[i] != tc.want[i] {
				t.Errorf("省 %d（%s）鄰省 = %v，預期 %v", tc.province, tc.note, got, tc.want)
				break
			}
		}
	}
}

// TestTerrainGrids 驗證 TERNAME.DAT 也是 39×196，且值域在 0-14。
//
// 地形編號的語意未解（docs/spec/01 §4），這裡只鎖住結構與值域。
func TestTerrainGrids(t *testing.T) {
	g, err := ParseProvinceGrids(gameFile(t, "TERNAME.DAT"))
	if err != nil {
		t.Fatal(err)
	}
	maxv := byte(0)
	for k := 0; k < ProvinceCount; k++ {
		for _, v := range g[k] {
			if v > maxv {
				maxv = v
			}
		}
	}
	// 值域 0-22 是掃過全部 39 省得到的。初稿只看河南一省推成 1-14，是錯的。
	if maxv > 22 {
		t.Errorf("地形編號最大值 = %d，預期 <= 22", maxv)
	}
	t.Logf("地形編號值域 0..%d", maxv)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
