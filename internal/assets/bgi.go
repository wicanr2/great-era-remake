package assets

import (
	"encoding/binary"
	"fmt"
	"image"
	"image/color"
)

// Borland BGI getimage 緩衝區（.TPC 與 NEWSDATA.DAT 用的格式）。
//
//	+0  u16  width  - 1
//	+2  u16  height - 1
//	+4  像素資料（4 個 bit-plane）
//	[+20 選配：'palt' + 16 bytes EGA 調色盤暫存器值]

// PaltTag 是內嵌調色盤的標記。帶它的檔案比公式算出來的大小多 20 bytes。
const PaltTag = "palt"

// Image 是解碼後的 16 色索引圖。不依賴 Ebiten——這一層不認識畫面。
type Image struct {
	W, H int
	Pix  []byte // 每像素 1 byte，值 0-15
	Palt []byte // 內嵌的 16 色 EGA 調色盤暫存器值；沒有就是 nil
}

// BGISize 回傳 w×h 的 BGI 圖像佔多少 bytes（含 4 bytes 檔頭）。
func BGISize(w, h int) int { return 4 + ((w+7)/8)*h*4 }

// DecodeBGI 從 data[off:] 解一張 BGI 圖像。
//
// [雷] plane 是逐列交錯，不是四個完整 plane 依序：
//
//	列0: plane0 plane1 plane2 plane3   每段 rowBytes = ceil(w/8)
//	列1: plane0 plane1 plane2 plane3
//
// 兩種佈局算出的檔案大小完全一樣，所以大小驗算過了不代表佈局對。
//
// [雷] plane 0 是最高位元：pix = plane0<<3 | plane1<<2 | plane2<<1 | plane3。
// 用相反的順序，圖形完全正確、色數也對，只有索引被 4-bit 反轉，
// 肉眼是「圖對但配色錯」。
func DecodeBGI(data []byte, off int) (*Image, error) {
	if off+4 > len(data) {
		return nil, fmt.Errorf("assets: BGI 檔頭超出範圍 (off=%d, len=%d)", off, len(data))
	}
	w := int(binary.LittleEndian.Uint16(data[off:])) + 1
	h := int(binary.LittleEndian.Uint16(data[off+2:])) + 1
	size := BGISize(w, h)
	if off+size > len(data) {
		return nil, fmt.Errorf("assets: BGI 圖像 %dx%d 需要 %d bytes，只剩 %d",
			w, h, size, len(data)-off)
	}

	rowBytes := (w + 7) / 8
	base := off + 4
	pix := make([]byte, w*h)
	for y := 0; y < h; y++ {
		ry := y * w
		for p := 0; p < 4; p++ {
			rb := base + (y*4+p)*rowBytes
			bit := byte(1) << (3 - p) // plane 0 是最高位
			for x := 0; x < w; x++ {
				if data[rb+x>>3]&(0x80>>(x&7)) != 0 {
					pix[ry+x] |= bit
				}
			}
		}
	}

	img := &Image{W: w, H: h, Pix: pix}
	if off+size+20 <= len(data) && string(data[off+size:off+size+4]) == PaltTag {
		img.Palt = append([]byte(nil), data[off+size+4:off+size+20]...)
	}
	return img, nil
}

// Size 回傳這張圖在原始檔案裡佔的 bytes，含 palt。解圖集時用來跳到下一張。
func (im *Image) Size() int {
	n := BGISize(im.W, im.H)
	if im.Palt != nil {
		n += 20
	}
	return n
}

// DecodeBGISet 解一個圖集（多張 BGI 圖串接），例如 AC.TPC、RAIL.TPC。
func DecodeBGISet(data []byte) ([]*Image, error) {
	var out []*Image
	for off := 0; off+4 <= len(data); {
		im, err := DecodeBGI(data, off)
		if err != nil {
			if len(out) == 0 {
				return nil, err
			}
			break // 尾端殘料，已經解到的照回
		}
		out = append(out, im)
		off += im.Size()
	}
	return out, nil
}

// ToRGBA 用給定的 16 色調色盤轉成標準 image.RGBA。
func (im *Image) ToRGBA(pal Palette) *image.RGBA {
	dst := image.NewRGBA(image.Rect(0, 0, im.W, im.H))
	for i, v := range im.Pix {
		c := color.RGBA{A: 0xFF}
		if int(v) < len(pal) {
			c.R, c.G, c.B = pal[v].R, pal[v].G, pal[v].B
		}
		dst.Set(i%im.W, i/im.W, c)
	}
	return dst
}
