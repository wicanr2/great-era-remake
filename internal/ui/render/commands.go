package render

import "github.com/wicanr2/great-era-remake/internal/assets"

// 政略指令選單。
//
// 十五個指令的名稱**不是重打的字串**，是用原版詞表的詞條拼出來的：
// `2.15` 的 0–15 是指令名的前半（調動、軍事、運補…），
// 16–19 是後綴（行動、活動、指令、選項），畫面上的「調動行動」「商業活動」
// 就是這樣組出來的。「慰勞軍民」是完整的四字詞，在 `4.15`。
//
// 對照實機選單（`docs/playtest/02` §5）逐項相符。
//
// ⚠️ `2.15` 的 13「休息」與 14「特殊」**沒有出現在選單裡**——
// 詞表有但十五個指令用不到，用途未知。

// CommandWord 是一個詞條的來源與編號。
type CommandWord struct {
	FromW4 bool // true 取 4.15（四字），false 取 2.15（二字）
	Index  int
}

// Command 是一個政略指令：由一到兩個詞條組成。
type Command struct {
	Num   int // 選單編號 1–15
	Parts []CommandWord
}

// 2.15 的後綴詞條。
const (
	w2Action   = 16 // 行動
	w2Activity = 17 // 活動
	w2Option   = 19 // 選項
	w2Ceasefire = 149 // 停火
)

// StrategyCommands 是十五個政略指令，順序照實機選單。
var StrategyCommands = []Command{
	{1, []CommandWord{{false, 0}, {false, w2Action}}},   // 調動行動
	{2, []CommandWord{{false, 1}, {false, w2Action}}},   // 軍事行動
	{3, []CommandWord{{false, 2}}},                      // 運補
	{4, []CommandWord{{false, 3}}},                      // 徵稅
	{5, []CommandWord{{false, 4}}},                      // 徵兵
	{6, []CommandWord{{false, 5}}},                      // 查閱
	{7, []CommandWord{{false, 6}}},                      // 開發
	{8, []CommandWord{{false, 7}}},                      // 政策
	{9, []CommandWord{{false, 8}}},                      // 外交
	{10, []CommandWord{{false, 9}, {false, w2Ceasefire}}}, // 談判停火
	{11, []CommandWord{{false, 10}, {false, w2Action}}}, // 秘密行動
	{12, []CommandWord{{false, 11}, {false, w2Activity}}}, // 商業活動
	{13, []CommandWord{{false, 12}}},                    // 練兵
	{14, []CommandWord{{true, 35}}},                     // 慰勞軍民（4.15 四字）
	{15, []CommandWord{{false, 15}, {false, w2Option}}}, // 其他選項
}

// CommandFonts 是畫指令選單需要的兩個字模檔。
type CommandFonts struct {
	W2 *assets.GlyphFile // 2.15，二字詞
	W4 *assets.GlyphFile // 4.15，四字詞
}

// DrawCommandPage 先鋪底色再畫選單，用於整頁的指令清單。
func (c *Canvas) DrawCommandPage(f CommandFonts, fg, bg assets.RGB, x, y, w, h int) error {
	c.fillRect(x, y, w, h, bg)
	return c.DrawCommandMenu(f, fg, x+14, y+16)
}

// DrawCommandMenu 把十五個指令畫成兩欄。
//
// 原版是按鍵之後才列出完整清單（`docs/playtest/02` 截圖 A2），
// 這裡的版面（兩欄、欄寬、行距）是 remake 的排版選擇。
func (c *Canvas) DrawCommandMenu(f CommandFonts, fg assets.RGB, x, y int) error {
	const (
		perCol = 8
		colW   = 210
		rowH   = 20
	)
	for i, cmd := range StrategyCommands {
		cx := x + (i/perCol)*colW
		cy := y + (i%perCol)*rowH

		c.DrawNumber(uint32(cmd.Num), fg, cx+DigitAdvance*2, cy)
		wx := cx + DigitAdvance*2 + 8
		for _, p := range cmd.Parts {
			gf, width := f.W2, 2
			if p.FromW4 {
				gf, width = f.W4, 4
			}
			if err := c.DrawEntry(gf, p.Index, width, fg, wx, cy, true); err != nil {
				return err
			}
			wx += width * GlyphAdvance
		}
	}
	return nil
}

// w4Confirm 是 `4.15` 的詞條 0：「您確定嗎」。
const w4Confirm = 0

// DrawConfirmBox 畫一個離開確認框。
//
// 用原版詞表的「您確定嗎」（`4.15` 詞條 0），不是重打的字串。
// 依 CLAUDE.md §9：ESC 只取消／退回，F10 才離開，離開前要確認並自動存檔。
func (c *Canvas) DrawConfirmBox(w4 *assets.GlyphFile, fg, bg assets.RGB, x, y int) {
	const w, h = 4*GlyphAdvance + 40, 60
	c.fillRect(x, y, w, h, bg)
	c.strokeRect(x, y, w, h, fg)
	_ = c.DrawEntry(w4, w4Confirm, 4, fg, x+20, y+12, true)
	// Y／N 用數字字型畫不出來，這裡用兩個框標示按鍵位置（remake 差異）。
	c.strokeRect(x+24, y+36, 18, 16, fg)
	c.strokeRect(x+64, y+36, 18, 16, fg)
}
