package assets

import "fmt"

// RAIL.TPC — 鐵路圖塊，21 張 32×24。
//
// **這個檔不是 BGI 格式**，雖然副檔名是 `.TPC`。檔頭讀出來是 1×1，
// 用 BGI 的 plane 佈局解會得到雜訊。它是 **4-bit chunky**：
// 每個 byte 存兩個像素（高 nibble 在左），一列 16 bytes = 32 px。
//
// 判斷依據：整個檔案只有 7 種 byte 值，nibble 只有 {0, 7, 8}——
// 也就是透明、淺灰、深灰三色的鐵軌。索引 0 是透明，疊在地形圖塊上。
//
//	8,064 = 21 張 × (32 × 24 ÷ 2)
//
// 對照：`NEWTERR.TPC` 是正規的 BGI 圖集（含 `palt`），兩者格式不同。
const (
	// RailW / RailH 是一張鐵路圖塊的尺寸，與地形圖塊相同。
	RailW, RailH = 32, 24
	railBytes    = RailW * RailH / 2
)

// RailTiles 是解好的 21 張鐵路圖塊，每張是 32×24 的索引像素。
type RailTiles [][]byte

// ParseRailTiles 解 RAIL.TPC。
func ParseRailTiles(data []byte) (RailTiles, error) {
	if len(data) != RailTileCount*railBytes {
		return nil, fmt.Errorf("assets: RAIL.TPC 應為 %d bytes（%d 張 × %d），實得 %d",
			RailTileCount*railBytes, RailTileCount, railBytes, len(data))
	}
	out := make(RailTiles, RailTileCount)
	for i := range out {
		pix := make([]byte, RailW*RailH)
		src := data[i*railBytes:]
		for y := 0; y < RailH; y++ {
			for x := 0; x < RailW; x++ {
				b := src[y*(RailW/2)+x/2]
				if x%2 == 0 {
					pix[y*RailW+x] = b >> 4
				} else {
					pix[y*RailW+x] = b & 0x0F
				}
			}
		}
		out[i] = pix
	}
	return out, nil
}
