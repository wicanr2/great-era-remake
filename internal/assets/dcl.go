package assets

import (
	"errors"
	"fmt"
)

// PKWARE Data Compression Library 的 "implode" 解壓。
//
// .GLB 圖庫的每一筆各自是一段獨立的 DCL stream，header 是 00 06
// （literal 不編碼、字典 4096）。演算法是公開的，這是 zlib contrib 的
// blast.c 移植，沒有從遊戲反組譯任何東西。
//
// 位元序是 LSB-first，且 Huffman 碼是 bit-inverted——見 decodeSym 的註解。

// 各表是 run-length 編碼的碼長：高 4 bits = 重複次數-1，低 4 bits = 碼長。
var (
	litLenRep = []byte{
		11, 124, 8, 7, 28, 7, 188, 13, 76, 4, 10, 8, 12, 10, 12, 10, 8, 23, 8,
		9, 7, 6, 7, 8, 7, 6, 55, 8, 23, 24, 12, 11, 7, 9, 11, 12, 6, 7, 22, 5,
		7, 24, 6, 11, 9, 6, 7, 22, 7, 11, 38, 7, 9, 8, 25, 11, 8, 11, 9, 12,
		8, 12, 5, 38, 5, 38, 5, 11, 7, 5, 6, 21, 6, 10, 53, 8, 7, 24, 10, 27,
		44, 253, 253, 253, 252, 252, 252, 13, 12, 45, 12, 45, 12, 61, 12, 45,
		44, 173,
	}
	lenLenRep  = []byte{2, 35, 36, 53, 38, 23}
	distLenRep = []byte{2, 20, 53, 230, 247, 151, 248}

	lengthBase  = []int{3, 2, 4, 5, 6, 7, 8, 9, 10, 12, 16, 24, 40, 72, 136, 264}
	lengthExtra = []int{0, 0, 0, 0, 0, 0, 0, 0, 1, 2, 3, 4, 5, 6, 7, 8}

	litCode  = buildHuffman(litLenRep)
	lenCode  = buildHuffman(lenLenRep)
	distCode = buildHuffman(distLenRep)
)

const endLength = 519 // 結束碼

// ErrCorrupt 表示碼流不合法。
var ErrCorrupt = errors.New("assets: DCL 碼流損壞")

type huffman struct {
	count  [16]int
	symbol []int
}

func buildHuffman(rep []byte) *huffman {
	var lengths []int
	h := &huffman{}
	for _, b := range rep {
		n := int(b>>4) + 1
		ln := int(b & 15)
		h.count[ln] += n
		for i := 0; i < n; i++ {
			lengths = append(lengths, ln)
		}
	}
	var offs [16]int
	for ln := 1; ln < 15; ln++ {
		offs[ln+1] = offs[ln] + h.count[ln]
	}
	h.symbol = make([]int, len(lengths))
	for sym, ln := range lengths {
		if ln != 0 {
			h.symbol[offs[ln]] = sym
			offs[ln]++
		}
	}
	return h
}

type bitReader struct {
	data []byte
	pos  int
	buf  uint32
	n    uint
}

func (r *bitReader) take(need uint) (int, error) {
	val := r.buf
	for r.n < need {
		if r.pos >= len(r.data) {
			return 0, fmt.Errorf("%w: 資料提前結束", ErrCorrupt)
		}
		val |= uint32(r.data[r.pos]) << r.n
		r.pos++
		r.n += 8
	}
	r.buf = val >> need
	r.n -= need
	return int(val & (1<<need - 1)), nil
}

func decodeSym(r *bitReader, h *huffman) (int, error) {
	code, first, index := 0, 0, 0
	for length := 1; length <= 13; length++ {
		b, err := r.take(1)
		if err != nil {
			return 0, err
		}
		// DCL 的 Huffman 碼是 bit-inverted（blast.c: code |= (bitbuf & 1) ^ 1）。
		// 漏掉這個反轉，literal 仍解得出來（走 raw 8 bits），但一遇到
		// length/distance 對就會算出離譜的距離，症狀是解到第 2-28 個 byte 就爆。
		code |= b ^ 1
		cnt := h.count[length]
		if code < first+cnt {
			return h.symbol[index+(code-first)], nil
		}
		index += cnt
		first = (first + cnt) << 1
		code <<= 1
	}
	return 0, fmt.Errorf("%w: 找不到有效碼", ErrCorrupt)
}

// Explode 解一段 DCL implode 資料。expected > 0 時順便驗證輸出長度。
func Explode(data []byte, expected int) ([]byte, error) {
	if len(data) < 3 {
		return nil, fmt.Errorf("%w: 資料太短 (%d bytes)", ErrCorrupt, len(data))
	}
	lit, dictBits := int(data[0]), uint(data[1])
	if lit > 1 {
		return nil, fmt.Errorf("%w: literal 模式非法 (%d)", ErrCorrupt, lit)
	}
	if dictBits < 4 || dictBits > 6 {
		return nil, fmt.Errorf("%w: 字典大小非法 (%d)", ErrCorrupt, dictBits)
	}

	r := &bitReader{data: data[2:]}
	out := make([]byte, 0, expected)
	for {
		isPair, err := r.take(1)
		if err != nil {
			return nil, err
		}
		if isPair == 0 {
			var b int
			if lit == 0 {
				b, err = r.take(8)
			} else {
				b, err = decodeSym(r, litCode)
			}
			if err != nil {
				return nil, err
			}
			out = append(out, byte(b))
			continue
		}

		sym, err := decodeSym(r, lenCode)
		if err != nil {
			return nil, err
		}
		extra, err := r.take(uint(lengthExtra[sym]))
		if err != nil {
			return nil, err
		}
		length := lengthBase[sym] + extra
		if length == endLength {
			break
		}
		shift := dictBits
		if length == 2 {
			shift = 2
		}
		dsym, err := decodeSym(r, distCode)
		if err != nil {
			return nil, err
		}
		dlow, err := r.take(shift)
		if err != nil {
			return nil, err
		}
		dist := dsym<<shift + dlow + 1
		if dist > len(out) {
			return nil, fmt.Errorf("%w: 距離 %d 超出已輸出的 %d bytes", ErrCorrupt, dist, len(out))
		}
		start := len(out) - dist
		for i := 0; i < length; i++ {
			out = append(out, out[start+i])
		}
	}

	if expected > 0 && len(out) != expected {
		return nil, fmt.Errorf("%w: 解出 %d bytes，預期 %d", ErrCorrupt, len(out), expected)
	}
	return out, nil
}
