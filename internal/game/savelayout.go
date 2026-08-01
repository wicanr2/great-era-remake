package game

import "fmt"

// `SAVE(N).DT1` 的完整佈局（`sub_595D4`，`docs/formats/07-dt1-layout.md`）。
//
// 原版把整個檔案當一筆記錄讀進來（`Reset(f, 395Bh)`），再用 Turbo Pascal
// 的 `$basg` 逐段搬到全域變數——**每一段的大小就寫在參數裡**，不必猜。
//
// 這張表的用途是給「改寫而非重建」的存檔寫回劃出邊界（`CLAUDE.md` §9）：
// 哪些區域已解可以蓋、哪些一個 byte 都不能動。

// SaveBlock 是 `.DT1` 裡的一個 `$basg` 區塊。
type SaveBlock struct {
	// Global 是原版搬去的全域變數名，留著方便對照組語。
	Global string
	// Offset / Size 是它在檔案裡的位置。
	Offset, Size int
	// Shape 是已知的拆法（如「39 × 37」），空字串表示還沒拆開。
	Shape string
	// Known 表示這一塊的語意已解——只有已解的區域可以被寫回覆蓋。
	Known bool
	// Note 是一句話說明。
	Note string
}

// SaveHeaderSize 是 `.DT1` 的檔頭：年 u16、月 u8、未解 u8。
const SaveHeaderSize = 4

// SaveFileSize 是 `.DT1` 的大小（原版 `Reset` 的 recsize `395Bh`）。
const SaveFileSize = 14683

// SaveBlocks 是十個 `$basg` 區塊，照原版的搬運順序。
var SaveBlocks = []SaveBlock{
	{"byte_6DFA0", 4, 1443, "39 × 37", true, "省份記錄（docs/spec/03）"},
	{"byte_6F532", 1447, 2340, "39 × 60？", false,
		"前 36 塊只有前 4 個 byte 有值、後 3 塊全滿；連 stride 是不是 60 都未定"},
	{"byte_6EFAA", 3787, 1416, "24 × 59", false,
		"⭐ 勢力表：24 槽 × 59 B（internal/game/factiontable.go）。" +
			"領袖 ID 與 24×24 外交矩陣已解，+2 與 +35..+58 未解，" +
			"所以整塊仍不可整片寫回——要寫只能逐欄位寫"},
	{"byte_6BC4E", 5203, 9042, "274 × 33", true, "將領記錄（docs/spec/02）"},
	{"byte_6FE56", 14245, 39, "39 × 1", false, "每省 1 B，未解"},
	{"byte_6EE68", 14284, 48, "", false, "未解"},
	{"byte_6EE98", 14332, 274, "274 × 1", false, "每將領 1 B，未解"},
	{"byte_6FF96", 14606, 40, "", false, "未解"},
	{"byte_6FF8C", 14646, 10, "", false, "未解"},
	{"word_70026", 14656, 20, "10 × u16", false, "未解"},
}

// SaveTrailingBytes 是十個區塊之後**沒被涵蓋**的尾巴。
//
// 4 + 十個區塊 = 14,676，而檔案是 14,683。這 7 bytes 由別的方式讀寫，未追。
// ⚠️ 寫回時一個 byte 都不要動。
const SaveTrailingBytes = SaveFileSize - 14676

// SaveBlockByGlobal 依全域變數名取區塊。
func SaveBlockByGlobal(name string) (SaveBlock, error) {
	for _, b := range SaveBlocks {
		if b.Global == name {
			return b, nil
		}
	}
	return SaveBlock{}, fmt.Errorf("game: .DT1 沒有名為 %s 的區塊", name)
}

// SaveBlockAt 回傳涵蓋 offset 的區塊。
//
// 給寫回流程用：**改任何一個 byte 之前先問它落在哪一塊、那塊解了沒。**
func SaveBlockAt(offset int) (SaveBlock, bool) {
	for _, b := range SaveBlocks {
		if offset >= b.Offset && offset < b.Offset+b.Size {
			return b, true
		}
	}
	return SaveBlock{}, false
}

// SaveWritable 回答某個位置能不能被寫回覆蓋。
//
// `CLAUDE.md` §9：**存檔寫回是「改寫」不是「重建」**，未解區域一個 byte
// 都不動。檔頭與尾巴都算不可寫——檔頭的第 4 個 byte 仍未解。
func SaveWritable(offset int) bool {
	b, ok := SaveBlockAt(offset)
	return ok && b.Known
}
