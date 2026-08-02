package game

import "fmt"

// 存檔寫回。
//
// CLAUDE.md §9 的硬規則：**「改寫」不是「重建」**——從原始 bytes 出發，
// 只蓋已解欄位，未解區域一個 byte 都不動。驗收標準是 byte-for-byte round-trip。
//
// 已解的省份區與將領區可以寫回；其餘區塊仍原樣保留。

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

// WriteSave 把目前已解且會在遊戲中改變的 `.DT1` 狀態寫回副本：
// 日期／省份與將領。未解區塊及未解將領 byte 一律保留原樣。
//
// generals 可以只包含該劇本實際使用的筆數；函式只覆蓋傳入的範圍，
// 不碰後面的殘留槽位。
func WriteSave(orig []byte, t *ProvinceTable, generals []General) ([]byte, error) {
	out, err := WriteProvinces(orig, t)
	if err != nil {
		return nil, err
	}
	need := SaveGeneralsOffset + len(generals)*GeneralRecordSize
	if len(out) < need {
		return nil, fmt.Errorf("game: .DT1 需要 %d bytes 才能寫 %d 位將領，只有 %d",
			need, len(generals), len(out))
	}
	for i := range generals {
		rec := generals[i].Bytes()
		copy(out[SaveGeneralsOffset+i*GeneralRecordSize:], rec[:])
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
