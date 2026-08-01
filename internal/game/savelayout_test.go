package game

import "testing"

// 十個區塊要首尾相接、不重疊、不留洞，而且加起來對得上檔案大小。
func TestSaveBlocksAreContiguous(t *testing.T) {
	want := SaveHeaderSize
	total := 0
	for _, b := range SaveBlocks {
		if b.Offset != want {
			t.Errorf("%s 該從 %d 開始，實得 %d", b.Global, want, b.Offset)
		}
		if b.Size <= 0 {
			t.Errorf("%s 大小 %d 不合理", b.Global, b.Size)
		}
		want = b.Offset + b.Size
		total += b.Size
	}
	if got := SaveHeaderSize + total + SaveTrailingBytes; got != SaveFileSize {
		t.Errorf("檔頭 %d + 區塊 %d + 尾巴 %d = %d，檔案是 %d",
			SaveHeaderSize, total, SaveTrailingBytes, got, SaveFileSize)
	}
	if SaveTrailingBytes != 7 {
		t.Errorf("尾巴該是 7 bytes，實得 %d", SaveTrailingBytes)
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
	for _, off := range []int{0, 3, 14676, SaveFileSize - 1} {
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
