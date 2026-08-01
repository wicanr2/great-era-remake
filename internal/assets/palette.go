package assets

import (
	"bufio"
	"fmt"
	"io"
	"strconv"
	"strings"
)

// RGB 是一個 8-bit 顏色。
type RGB struct{ R, G, B uint8 }

// Palette 是一組顏色。.RGB 檔是 16 色，.GLB 內嵌的調色盤是 256 色。
type Palette []RGB

// dac6to8 把 VGA 的 6-bit DAC 值展成 8-bit。
//
// [雷] 用位元複製，不是比例縮放。(v<<2)|(v>>4) 讓 63→255、51→207、16→65；
// 而 v*255/63 會得到 206 和 64。兩者只差 1，但逐像素比對原版畫面時，
// 這個 1 會讓整張圖判定為不同（實測相異率 99.99%）。
func dac6to8(v int) uint8 { return uint8(v<<2 | v>>4) }

// ParseRGBFile 讀 .RGB 調色盤：純文字，16 行，每行一組 RGB 百分比。
//
// [雷] 值是百分比（0-100），不是 6-bit DAC。八個 .RGB 檔的最大值都是 100。
// 轉 8-bit 一定要經過 6-bit DAC 這一步：
//
//	百分比 → DAC = round(pct×63/100) → 8-bit = (dac<<2)|(dac>>4)
//
// 直接 pct×255/100 會差 1-2（60% 算出 153，正確是 154）。
func ParseRGBFile(r io.Reader) (Palette, error) {
	var nums []int
	sc := bufio.NewScanner(r)
	for sc.Scan() {
		for _, f := range strings.Fields(sc.Text()) {
			n, err := strconv.Atoi(f)
			if err != nil {
				return nil, fmt.Errorf("assets: .RGB 含非數字 %q: %w", f, err)
			}
			if n < 0 || n > 100 {
				return nil, fmt.Errorf("assets: .RGB 值 %d 超出百分比範圍 0-100", n)
			}
			nums = append(nums, n)
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	if len(nums)%3 != 0 {
		return nil, fmt.Errorf("assets: .RGB 的數字個數 %d 不是 3 的倍數", len(nums))
	}

	pal := make(Palette, 0, len(nums)/3)
	for i := 0; i < len(nums); i += 3 {
		pal = append(pal, RGB{
			R: dac6to8(pctToDAC(nums[i])),
			G: dac6to8(pctToDAC(nums[i+1])),
			B: dac6to8(pctToDAC(nums[i+2])),
		})
	}
	return pal, nil
}

// pctToDAC 把 0-100 的百分比轉成 0-63 的 DAC 值（四捨五入）。
func pctToDAC(pct int) int { return (pct*63 + 50) / 100 }

// ParseVGAPalette 讀 .GLB 內嵌的 256 色調色盤：768 bytes 的 6-bit DAC 三元組。
func ParseVGAPalette(data []byte) (Palette, error) {
	if len(data) != 768 {
		return nil, fmt.Errorf("assets: VGA 調色盤應為 768 bytes，實為 %d", len(data))
	}
	pal := make(Palette, 256)
	for i := 0; i < 256; i++ {
		r, g, b := int(data[i*3]), int(data[i*3+1]), int(data[i*3+2])
		if r > 63 || g > 63 || b > 63 {
			return nil, fmt.Errorf("assets: 調色盤第 %d 色超出 6-bit 範圍 (%d,%d,%d)", i, r, g, b)
		}
		pal[i] = RGB{dac6to8(r), dac6to8(g), dac6to8(b)}
	}
	return pal, nil
}

// EGA16 是 BGI 的預設 16 色，沒有 .RGB 檔時的 fallback。
var EGA16 = Palette{
	{0, 0, 0}, {0, 0, 170}, {0, 170, 0}, {0, 170, 170},
	{170, 0, 0}, {170, 0, 170}, {170, 85, 0}, {170, 170, 170},
	{85, 85, 85}, {85, 85, 255}, {85, 255, 85}, {85, 255, 255},
	{255, 85, 85}, {255, 85, 255}, {255, 255, 85}, {255, 255, 255},
}

// EGADefaultPalette 是 EGA/VGA 的 16 色預設調色盤。
//
// `.TPC` 的 `palt` chunk 存的都是 EGA 預設暫存器值
// （`00 01 02 03 04 05 14 07 38 39 3a 3b 3c 3d 3e 3f`），對應的就是這 16 色。
//
// ⚠️ **這不是「原版畫面上的顏色」的最終答案**：`palt` 存的是 VGA DAC 索引，
// 真實顏色還要看當時的 DAC 被哪個 `.RGB` 設成什麼。戰場用哪一個 `.RGB`
// 還沒查出來（8 個檔名都不像戰場）。在查明之前用這組預設值，
// 渲染出來的草地／水／城牆顏色是合理的，但不保證與實機逐像素相同。
var EGADefaultPalette = Palette{
	{R: 0, G: 0, B: 0}, {R: 0, G: 0, B: 170}, {R: 0, G: 170, B: 0},
	{R: 0, G: 170, B: 170}, {R: 170, G: 0, B: 0}, {R: 170, G: 0, B: 170},
	{R: 170, G: 85, B: 0}, {R: 170, G: 170, B: 170},
	{R: 85, G: 85, B: 85}, {R: 85, G: 85, B: 255}, {R: 85, G: 255, B: 85},
	{R: 85, G: 255, B: 255}, {R: 255, G: 85, B: 85}, {R: 255, G: 85, B: 255},
	{R: 255, G: 255, B: 85}, {R: 255, G: 255, B: 255},
}
