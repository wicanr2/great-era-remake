package game

import "testing"

// 十個 `$basg` 區塊不重疊，且位移必須對上 IDA 的原始 record offset。
// 區塊 8／9／10 之間夾著七個由單 byte 指令讀寫的 runtime 欄位，不能假設
// 所有區塊首尾相接。
func TestSaveBlocksMatchRawOffsets(t *testing.T) {
	want := map[string]struct{ offset, size int }{
		"byte_6DFA0": {4, 1443},
		"byte_6F532": {1447, 2340},
		"byte_6EFAA": {3787, 1416},
		"byte_6BC4E": {5203, 9042},
		"byte_6FE56": {14245, 39},
		"byte_6EE68": {14284, 48},
		"byte_6EE98": {14332, 274},
		"byte_6FF8C": {14610, 10},
		"byte_6FF96": {14620, 40},
		"word_70026": {14662, 20},
	}
	var ranges []SaveBlock
	for _, b := range SaveBlocks {
		w, ok := want[b.Global]
		if !ok {
			t.Errorf("未預期的 SaveBlock %q", b.Global)
			continue
		}
		if b.Offset != w.offset || b.Size != w.size {
			t.Errorf("%s 位移／大小 = %d/%d，預期 %d/%d", b.Global,
				b.Offset, b.Size, w.offset, w.size)
		}
		if b.Size <= 0 {
			t.Errorf("%s 大小 %d 不合理", b.Global, b.Size)
		}
		ranges = append(ranges, b)
	}
	for i := 0; i < len(ranges); i++ {
		for j := i + 1; j < len(ranges); j++ {
			left, right := ranges[i], ranges[j]
			if left.Offset > right.Offset {
				left, right = right, left
			}
			if left.Offset+left.Size > right.Offset {
				t.Errorf("區塊 %s／%s 重疊", left.Global, right.Global)
			}
		}
	}
	if SaveTrailingBytes != 7 {
		t.Errorf("未映射 runtime bytes 該是 7，實得 %d", SaveTrailingBytes)
	}
}

// 三個獨立錨點：省份區、將領區起點、將領區長度。
func TestSaveLayoutMatchesIndependentAnchors(t *testing.T) {
	prov, err := SaveBlockByGlobal("byte_6DFA0")
	if err != nil {
		t.Fatal(err)
	}
	if prov.Size != ProvinceCount*ProvinceRecordSize {
		t.Errorf("省份區 %d，該是 %d × %d", prov.Size, ProvinceCount, ProvinceRecordSize)
	}
	gen, err := SaveBlockByGlobal("byte_6BC4E")
	if err != nil {
		t.Fatal(err)
	}
	if gen.Offset != SaveGeneralsOffset {
		t.Errorf("將領區從 %d 開始，指紋掃出來的是 %d", gen.Offset, SaveGeneralsOffset)
	}
	if gen.Size != Stage1GeneralCount*GeneralRecordSize {
		t.Errorf("將領區 %d，該是 %d × %d",
			gen.Size, Stage1GeneralCount, GeneralRecordSize)
	}
}

func TestSaveFileSizeMatchesRealFile(t *testing.T) {
	if got := len(readGame(t, "SAVE(1).DT1")); got != SaveFileSize {
		t.Errorf("SAVE(1).DT1 是 %d bytes，常數寫 %d", got, SaveFileSize)
	}
}

// ⭐ 區塊 2 的實測形狀。
//
// ⛔ 初稿寫「每一塊都只有前 4 個 byte 非零」並斷言「北伐只有 36 省可用，
// 所以第 37..39 塊為零」。**這條測試當場推翻了它**：第 37..39 塊不是空的，
// 是**滿的**（60 bytes 全非零）。正確的敘述在 `docs/formats/07` §3。
func TestBlock2ShapeAsMeasured(t *testing.T) {
	blk, err := SaveBlockByGlobal("byte_6F532")
	if err != nil {
		t.Fatal(err)
	}
	const per = 60
	if blk.Size != ProvinceCount*per {
		t.Fatalf("區塊 2 該是 %d × %d", ProvinceCount, per)
	}
	for _, f := range []string{"SAVE(1).DT1", "SAVE(2).DT1"} {
		d := readGame(t, f)
		for p := 0; p < 36; p++ {
			base := blk.Offset + p*per
			for j := 4; j < per; j++ {
				if d[base+j] != 0 {
					t.Errorf("%s 前 36 塊該只有前 4 個 byte 有值，"+
						"省 %d 的第 %d 個 byte = %d", f, p+1, j, d[base+j])
				}
			}
		}
		// 第 37..39 塊反而是滿的——**這正是「60 是不是真的 stride」的疑點**。
		full := 0
		for p := 36; p < ProvinceCount; p++ {
			base := blk.Offset + p*per
			nz := 0
			for j := 0; j < per; j++ {
				if d[base+j] != 0 {
					nz++
				}
			}
			if nz == per {
				full++
			}
		}
		if full != 3 {
			t.Errorf("%s 第 37..39 塊實測是滿的（%d/3）——"+
				"形狀若變了，`docs/formats/07` §3 要重寫", f, full)
		}
	}
}

func TestSaveWritableOnlyForKnownBlocks(t *testing.T) {
	// 省份區與將領區可以蓋。
	if !SaveWritable(4) || !SaveWritable(SaveGeneralsOffset) {
		t.Error("省份區與將領區該可寫")
	}
	// 中間那兩塊未解，一個 byte 都不能動。
	for _, off := range []int{1447, 3786, 3787, 5202} {
		if SaveWritable(off) {
			t.Errorf("offset %d 在未解區塊，不該可寫", off)
		}
	}
	// 檔頭與尾巴都不在任何區塊裡。
	for _, off := range []int{0, 3, 14606, 14609, 14660, 14661, 14682} {
		if _, ok := SaveBlockAt(off); ok {
			t.Errorf("offset %d 不該落在任何 $basg 區塊", off)
		}
		if SaveWritable(off) {
			t.Errorf("offset %d 不該可寫", off)
		}
	}
}

func TestSaveBlockByGlobalRejectsUnknown(t *testing.T) {
	if _, err := SaveBlockByGlobal("byte_DEADBEEF"); err == nil {
		t.Error("不存在的區塊該回錯誤")
	}
}
