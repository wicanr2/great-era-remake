package game

import "fmt"

// 存檔寫回。
//
// CLAUDE.md §9 的硬規則：**「改寫」不是「重建」**——從原始 bytes 出發，
// 只蓋已解欄位，未解區域一個 byte 都不動。驗收標準是 byte-for-byte round-trip。
//
// 這一層只處理 `.DT1` 的前段（4 B 檔頭 + 39 省 × 37 B）。
// 1447 之後的 13,236 bytes 語意未解，原樣保留。

// WriteProvinces 把省份表寫回一份 .DT1 的副本。
//
// 傳入的 orig 是原始檔案內容，回傳的是新的 bytes——**不修改 orig**。
// 檔案的其餘部分（檔頭之後的 1447 以降）原樣複製。
func WriteProvinces(orig []byte, t *ProvinceTable) ([]byte, error) {
	need := SaveArrayOffset + ProvinceCount*ProvinceRecordSize
	if len(orig) < need {
		return nil, fmt.Errorf("game: .DT1 至少需要 %d bytes，只有 %d", need, len(orig))
	}
	out := make([]byte, len(orig))
	copy(out, orig)

	if t.Date != nil {
		out[0] = byte(t.Date.Year)
		out[1] = byte(t.Date.Year >> 8)
		out[2] = t.Date.Month
		out[3] = t.Date.Unknown3
	}
	for i := range t.Province {
		rec := t.Province[i].Bytes()
		copy(out[SaveArrayOffset+i*ProvinceRecordSize:], rec[:])
	}
	return out, nil
}

// WriteTownFile 把省份表寫回一份 TOWN(N).DAT 的副本。
//
// TOWN 檔沒有檔頭，陣列從 offset 0 開始（docs/spec/03 §1 的相位差）。
func WriteTownFile(orig []byte, t *ProvinceTable) ([]byte, error) {
	if want := ProvinceCount * ProvinceRecordSize; len(orig) != want {
		return nil, fmt.Errorf("game: TOWN 檔應為 %d bytes，實得 %d", want, len(orig))
	}
	out := make([]byte, len(orig))
	copy(out, orig)
	for i := range t.Province {
		rec := t.Province[i].Bytes()
		copy(out[TownArrayOffset+i*ProvinceRecordSize:], rec[:])
	}
	return out, nil
}

// DiffBytes 回傳兩份 bytes 有哪些位置不同，最多列 limit 個。
//
// 用來驗證「只改了該改的地方」——存檔寫回的驗收不是「能讀回來」，
// 而是「沒動到不該動的 byte」。
func DiffBytes(a, b []byte, limit int) []int {
	var out []int
	n := len(a)
	if len(b) < n {
		n = len(b)
	}
	for i := 0; i < n; i++ {
		if a[i] != b[i] {
			out = append(out, i)
			if limit > 0 && len(out) >= limit {
				break
			}
		}
	}
	return out
}
