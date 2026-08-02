package render

import (
	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
)

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
	w2Action        = 16  // 行動
	w2Activity      = 17  // 活動
	w2Option        = 19  // 選項
	w2Ceasefire     = 149 // 停火
	w2Commander     = 21  // 司令（panel.go 的同詞條沿用歷史拼音名稱 w2Sikrei）
	w2HowMany       = 41  // 多少
	w2Buy           = 84  // 購買
	w2Sell          = 85  // 拋售
	w2Import        = 86  // 進口
	w2Export        = 87  // 出口
	w2Infantry      = 127 // 步兵
	w2Artillery     = 128 // 砲兵
	w2Armour        = 129 // 裝甲
	w2Cavalry       = 130 // 騎兵
	w2Lead          = 70  // 帶兵
	w2Ability       = 71  // 能力
	w2Experience    = 74  // 經驗
	w2Branch        = 75  // 軍種
	w2UnitTitle     = 76  // 番號
	w2Soldier       = 77  // 士兵
	w2Armed         = 78  // 武裝
	w2Degree        = 79  // 程度
	w2Stamina       = 80  // 體力
	w2Morale        = 81  // 士氣
	w2Skill         = 161 // 戰技
	w2Overview      = 83  // 概況
	w2AllProvince   = 84  // 各省
	w2Belong        = 85  // 所屬
	w2OtherProvince = 86  // 他省
	w2ProvinceName  = 67  // 省名
	w2Encourage     = 134 // 鼓動
	w2Student       = 135 // 學潮
	w2All           = 47  // 全部
	w2Partial       = 48  // 部份（原版用字）
	w2StayBehind    = 91  // 留守
	w2Confirm       = 195 // 確認
)

const (
	w3Attack   = 42 // 攻擊力
	w3Soldiers = 49 // 士兵數
)

// StrategyCommands 是十五個政略指令，順序照實機選單。
var StrategyCommands = []Command{
	{1, []CommandWord{{false, 0}, {false, w2Action}}},     // 調動行動
	{2, []CommandWord{{false, 1}, {false, w2Action}}},     // 軍事行動
	{3, []CommandWord{{false, 2}}},                        // 運補
	{4, []CommandWord{{false, 3}}},                        // 徵稅
	{5, []CommandWord{{false, 4}}},                        // 徵兵
	{6, []CommandWord{{false, 5}}},                        // 查閱
	{7, []CommandWord{{false, 6}}},                        // 開發
	{8, []CommandWord{{false, 7}}},                        // 政策
	{9, []CommandWord{{false, 8}}},                        // 外交
	{10, []CommandWord{{false, 9}, {false, w2Ceasefire}}}, // 談判停火
	{11, []CommandWord{{false, 10}, {false, w2Action}}},   // 秘密行動
	{12, []CommandWord{{false, 11}, {false, w2Activity}}}, // 商業活動
	{13, []CommandWord{{false, 12}}},                      // 練兵
	{14, []CommandWord{{true, 35}}},                       // 慰勞軍民（4.15 四字）
	{15, []CommandWord{{false, 15}, {false, w2Option}}},   // 其他選項
}

// DrawCovertAction 畫秘密行動的兩個原版子項。
func (c *Canvas) DrawCovertAction(f CommandFonts, w3 *assets.GlyphFile,
	fg, bg assets.RGB, x, y, w, h int) error {
	c.fillRect(x, y, w, h, bg)
	c.DrawNumber(1, fg, x+36, y+34)
	if err := c.DrawEntry(w3, 45, 3, fg, x+72, y+34, true); err != nil {
		return err
	}
	c.DrawNumber(2, fg, x+36, y+66)
	if err := c.DrawEntry(f.W2, w2Encourage, 2, fg, x+72, y+66, true); err != nil {
		return err
	}
	return c.DrawEntry(f.W2, w2Student, 2, fg, x+112, y+66, true)
}

// DrawCovertTarget 畫「司令／鼓動／何省／學潮」與省編號輸入。
func (c *Canvas) DrawCovertTarget(f CommandFonts, fg, bg assets.RGB,
	x, y, w, h, action int, input uint32) error {
	c.fillRect(x, y, w, h, bg)
	entries := []int{w2Commander, w2Encourage, 39, w2Student}
	if action != 2 {
		entries = []int{w2Commander, 10, w2Action, 39}
	}
	wx := x + 24
	for _, entry := range entries {
		if err := c.DrawEntry(f.W2, entry, 2, fg, wx, y+34, true); err != nil {
			return err
		}
		wx += 2 * GlyphAdvance
	}
	c.DrawNumber(input, fg, x+36, y+78)
	return nil
}

// CommandFonts 是畫指令選單需要的兩個字模檔。
type CommandFonts struct {
	W2 *assets.GlyphFile // 2.15，二字詞
	W4 *assets.GlyphFile // 4.15，四字詞
}

// DrawOpenCommandButton 畫地圖上的可見指令入口。文字沿用原版 2.15
// 「指令」字模；按鈕本身是為滑鼠與觸控新增的重製版操作外殼。
func (c *Canvas) DrawOpenCommandButton(f CommandFonts, fg, bg assets.RGB, logicalWidth int) error {
	p := uilayout.OpenCommandButton(logicalWidth)
	c.fillRect(p.X, p.Y, p.HitW, p.HitH, bg)
	c.strokeRect(p.X, p.Y, p.HitW, p.HitH, fg)
	return c.DrawEntry(f.W2, w2Command, 2, fg, p.X+25, p.Y+15, true)
}

// DrawNavigationButton 以程式圖形畫 48×48 導覽鍵，不依賴授權字庫。
// back/previous 畫向左箭頭，next 畫向右箭頭。
func (c *Canvas) DrawNavigationButton(fg, bg assets.RGB, logicalWidth, y, slot int, next bool) {
	p := uilayout.NavigationButton(logicalWidth, y, slot)
	c.fillRect(p.X, p.Y, p.HitW, p.HitH, bg)
	c.strokeRect(p.X, p.Y, p.HitW, p.HitH, fg)
	// 粗像素箭身。
	c.fillRect(p.X+13, p.Y+22, 22, 4, fg)
	if next {
		for i := 0; i < 4; i++ {
			c.fillRect(p.X+30+i*2, p.Y+14+i*2, 4, 4, fg)
			c.fillRect(p.X+30+i*2, p.Y+30-i*2, 4, 4, fg)
		}
		return
	}
	for i := 0; i < 4; i++ {
		c.fillRect(p.X+14-i*2, p.Y+22-i*2, 4, 4, fg)
		c.fillRect(p.X+14-i*2, p.Y+22+i*2, 4, 4, fg)
	}
}

// DrawCommandPage 先鋪底色再畫選單，用於整頁的指令清單。
func (c *Canvas) DrawCommandPage(f CommandFonts, fg, bg assets.RGB, x, y, w, h int) error {
	c.fillRect(x, y, w, h, bg)
	return c.DrawCommandMenu(f, fg, x+14, y+16)
}

// DrawDevelopPage 畫原版「開發」的三個子項。詞條與索引取自
// IDA `sub_24D59` 與 `docs/re/27-screen-map.md`。前三段可直接對到
// 2.15 #50「墾地」、1.15 #34「建」、3.15 #41「兵工廠」（IDA 為
// 1-based）；「挖金礦」則取受版控槽表確認的 3.15 0-based #52。
// `sub_24D59` 同一位置的立即數 #10 來自另一個尚未追回來源檔的緩衝區，
// 不可誤套到 3.15（會畫成「熱河省」）。
func (c *Canvas) DrawDevelopPage(f PanelFonts, fg, bg assets.RGB, x, y, w, h int) error {
	c.fillRect(x, y, w, h, bg)
	for i := 0; i < 3; i++ {
		c.DrawNumber(uint32(i+1), fg, x+36, y+30+i*38)
	}
	if err := c.DrawEntry(f.W2, 49, 2, fg, x+72, y+30, true); err != nil {
		return err
	}
	if err := c.DrawEntry(f.W1, 33, 1, fg, x+72, y+68, true); err != nil {
		return err
	}
	if err := c.DrawEntry(f.W3, 40, 3, fg, x+92, y+68, true); err != nil {
		return err
	}
	return c.DrawEntry(f.W3, 52, 3, fg, x+72, y+106, true)
}

// DrawSemanticConfirm 畫完整字庫的確認提示；Y/N 與鍵盤契約保持可見。
func (c *Canvas) DrawSemanticConfirm(fonts *assets.EtenFonts, fg, bg assets.RGB,
	x, y, w, h int, prompt string) []rune {
	c.fillRect(x, y, w, h, bg)
	missing := c.DrawSemanticText(fonts, prompt, fg, x+34, y+52)
	missing = append(missing, c.DrawSemanticText(fonts, "Y：確定　N：取消", fg, x+80, y+105)...)
	return uniqueRunes(missing)
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
		p := uilayout.Grid(i, x, y, perCol, colW, rowH, 0, colW-5, rowH)
		cx, cy := p.X, p.Y

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

// DrawSemanticCommandPage 以完整字庫畫十五項語意名稱。
// 座標、編號與選項數量與原典路徑相同。
func (c *Canvas) DrawSemanticCommandPage(fonts *assets.EtenFonts, fg, bg assets.RGB,
	x, y, w, h int, labels []string) []rune {
	c.fillRect(x, y, w, h, bg)
	const perCol, colW, rowH = 8, 210, 38
	missing := []rune{}
	for i, label := range labels {
		p := uilayout.Grid(i, x+14, y+30, perCol, colW, rowH, 0, colW-5, rowH)
		cx, cy := p.X, p.Y
		c.DrawNumber(uint32(i+1), fg, cx+DigitAdvance*2, cy)
		missing = append(missing, c.DrawSemanticText(fonts, label, fg, cx+DigitAdvance*2+14, cy)...)
	}
	return uniqueRunes(missing)
}

// DrawTradeMenu 畫商業活動的進口／出口或品項選單。
//
// 詞條全部取自原版 `2.15`。版面是 remake 的兩欄選單排法；選項順序則照
// `sub_1D703`（進口）與 `sub_1ED95`（出口）。
func (c *Canvas) DrawTradeMenu(f CommandFonts, fg, bg assets.RGB, x, y, w, h int,
	importing *bool) error {
	c.fillRect(x, y, w, h, bg)
	entries := []int{w2Import, w2Export}
	if importing != nil {
		if *importing {
			entries = []int{w2Food, w2Ammo, w2Fuel}
		} else {
			entries = []int{w2Food, w2Ammo, w2Coal, w2Iron, w2Fuel}
		}
	}
	for i, entry := range entries {
		cy := y + 30 + i*28
		c.DrawNumber(uint32(i+1), fg, x+36, cy)
		if err := c.DrawEntry(f.W2, entry, 2, fg, x+72, cy, true); err != nil {
			return err
		}
	}
	return nil
}

// DrawTradeAmount 畫「司令／購買或拋售／多少／品項」與目前輸入的數量。
func (c *Canvas) DrawTradeAmount(f CommandFonts, fg, bg assets.RGB,
	x, y, w, h int, importing bool, good int, amount uint32) error {
	c.fillRect(x, y, w, h, bg)
	verb := w2Sell
	if importing {
		verb = w2Buy
	}
	goodEntry := []int{w2Food, w2Ammo, w2Fuel, w2Coal, w2Iron}
	if good < 0 || good >= len(goodEntry) {
		good = 0
	}
	wx := x + 24
	for _, entry := range []int{w2Commander, verb, w2HowMany, goodEntry[good]} {
		if err := c.DrawEntry(f.W2, entry, 2, fg, wx, y+34, true); err != nil {
			return err
		}
		wx += 2 * GlyphAdvance
	}
	c.DrawNumber(amount, fg, x+36, y+78)
	return nil
}

// DrawSupplyTarget 畫運補目標提示與原版順序的候選省編號／省名。
func (c *Canvas) DrawSupplyTarget(f CommandFonts, provinces *assets.GlyphFile,
	fg, bg assets.RGB, x, y, w, h int, candidates []int, input uint32) error {
	c.fillRect(x, y, w, h, bg)
	wx := x + 24
	for _, entry := range []int{w2Commander, 2, 39} { // 司令／運補／何省
		if err := c.DrawEntry(f.W2, entry, 2, fg, wx, y+24, true); err != nil {
			return err
		}
		wx += 2 * GlyphAdvance
	}
	for i, id := range candidates {
		cy := y + 64 + i*24
		c.DrawNumber(uint32(id), fg, x+44, cy)
		if id >= 1 {
			if err := c.DrawEntry(provinces, id-1, 3, fg, x+100, cy, true); err != nil {
				return err
			}
		}
	}
	c.DrawNumber(input, fg, x+300, y+24)
	return nil
}

// DrawPlayerTransferTarget 畫「司令／調動／何省」與合法鄰省。
func (c *Canvas) DrawPlayerTransferTarget(f CommandFonts, provinces *assets.GlyphFile,
	fg, bg assets.RGB, x, y, w, h int, candidates []int, input uint32) error {
	c.fillRect(x, y, w, h, bg)
	wx := x + 24
	for _, entry := range []int{w2Commander, 0, 39} {
		if err := c.DrawEntry(f.W2, entry, 2, fg, wx, y+24, true); err != nil {
			return err
		}
		wx += 2 * GlyphAdvance
	}
	for i, id := range candidates {
		cy := y + 64 + i*24
		c.DrawNumber(uint32(id), fg, x+44, cy)
		if id >= 1 {
			if err := c.DrawEntry(provinces, id-1, 3, fg, x+100, cy, true); err != nil {
				return err
			}
		}
	}
	c.DrawNumber(input, fg, x+300, y+24)
	return nil
}

// DrawSupplyAmount 畫目前正在輸入的運補資源與數量。
func (c *Canvas) DrawSupplyAmount(f CommandFonts, fg, bg assets.RGB,
	x, y, w, h int, good int, amount uint32) error {
	c.fillRect(x, y, w, h, bg)
	goods := []int{w2Gold, w2Food, w2Ammo, w2Fuel}
	if good < 0 || good >= len(goods) {
		good = 0
	}
	wx := x + 24
	for _, entry := range []int{w2Commander, 0, w2HowMany, goods[good]} { // 司令／調動／多少／資源
		if err := c.DrawEntry(f.W2, entry, 2, fg, wx, y+34, true); err != nil {
			return err
		}
		wx += 2 * GlyphAdvance
	}
	c.DrawNumber(amount, fg, x+36, y+78)
	return nil
}

// DrawPlayerTransferMode 畫原版指令 1 的「部份調動／全部調動」。
func (c *Canvas) DrawPlayerTransferMode(f CommandFonts, fg, bg assets.RGB,
	x, y, w, h int) error {
	c.fillRect(x, y, w, h, bg)
	for i, entry := range []int{w2Partial, w2All} {
		cy := y + 34 + i*34
		c.DrawNumber(uint32(i+1), fg, x+36, cy)
		if err := c.DrawEntry(f.W2, entry, 2, fg, x+72, cy, true); err != nil {
			return err
		}
		if err := c.DrawEntry(f.W2, 0, 2, fg, x+112, cy, true); err != nil {
			return err
		}
	}
	return nil
}

// DrawPlayerTransferSelection 畫最多一頁 20 位候選。
// 方框是游標，星號以實心小方塊表示；Enter 的確認行是現代操作修正。
func (c *Canvas) DrawPlayerTransferSelection(f CommandFonts, generals *assets.GlyphFile,
	fg, bg assets.RGB, x, y, w, h int, ids []int, selected []bool, cursor, page int,
	mode game.PlayerTransferMode) error {
	c.fillRect(x, y, w, h, bg)
	prompt := 0 // 調動
	if mode == game.PlayerTransferAll {
		prompt = w2StayBehind
	}
	if err := c.DrawEntry(f.W2, prompt, 2, fg, x+24, y+14, true); err != nil {
		return err
	}
	start := page * 20
	end := start + 20
	if end > len(ids) {
		end = len(ids)
	}
	for i := start; i < end; i++ {
		row := i - start
		cx := x + 26 + (row/10)*205
		cy := y + 50 + (row%10)*24
		if i == cursor {
			c.strokeRect(cx-8, cy-3, 190, 20, fg)
		}
		if i < len(selected) && selected[i] {
			c.fillRect(cx, cy+4, 7, 7, fg)
		}
		c.DrawNumber(uint32(i+1), fg, cx+34, cy)
		if ids[i] > 0 {
			if err := c.DrawEntry(generals, ids[i]-1, 3, fg, cx+48, cy, true); err != nil {
				return err
			}
		}
	}
	if err := c.DrawEntry(f.W2, w2Confirm, 2, fg, x+24, y+h-30, true); err != nil {
		return err
	}
	return nil
}

// DrawRecruitAction 畫指令 5 的第一層：徵兵／重新整編。
func (c *Canvas) DrawRecruitAction(f CommandFonts, fg, bg assets.RGB,
	x, y, w, h int) error {
	c.fillRect(x, y, w, h, bg)
	c.DrawNumber(1, fg, x+36, y+34)
	if err := c.DrawEntry(f.W2, 4, 2, fg, x+72, y+34, true); err != nil {
		return err
	}
	c.DrawNumber(2, fg, x+36, y+66)
	return c.DrawEntry(f.W4, 22, 4, fg, x+72, y+66, true)
}

// DrawRecruitBranch 畫四兵種子選單。順序照原版 `sub_29B46`：
// 步兵、裝甲兵、砲兵、騎兵。
func (c *Canvas) DrawRecruitBranch(f CommandFonts, fg, bg assets.RGB,
	x, y, w, h int) error {
	c.fillRect(x, y, w, h, bg)
	entries := []int{w2Infantry, w2Armour, w2Artillery, w2Cavalry}
	for i, entry := range entries {
		cy := y + 30 + i*28
		c.DrawNumber(uint32(i+1), fg, x+36, cy)
		if err := c.DrawEntry(f.W2, entry, 2, fg, x+72, cy, true); err != nil {
			return err
		}
		if entry == w2Armour {
			// 「裝甲」詞條後接原版「兵」字；2.15 字元 9 是「兵」。
			c.DrawGlyph(f.W2.Glyphs[9], fg, x+112, cy, true)
		}
	}
	return nil
}

func recruitBranchEntry(branch uint8) (entry int, armour bool) {
	switch branch {
	case 1:
		return w2Infantry, false
	case 4:
		return w2Artillery, false
	case 5:
		return w2Armour, true
	case 6:
		return w2Cavalry, false
	default:
		return w2Infantry, false
	}
}

// DrawRecruitAmount 畫兵種、可徵上限與目前輸入值。
func (c *Canvas) DrawRecruitAmount(f CommandFonts, fg, bg assets.RGB,
	x, y, w, h int, branch uint8, limit int, amount uint32) error {
	c.fillRect(x, y, w, h, bg)
	entry, armour := recruitBranchEntry(branch)
	if err := c.DrawEntry(f.W2, w2Commander, 2, fg, x+24, y+30, true); err != nil {
		return err
	}
	if err := c.DrawEntry(f.W2, w2HowMany, 2, fg, x+104, y+30, true); err != nil {
		return err
	}
	if err := c.DrawEntry(f.W2, entry, 2, fg, x+184, y+30, true); err != nil {
		return err
	}
	if armour {
		c.DrawGlyph(f.W2.Glyphs[9], fg, x+224, y+30, true)
	}
	c.DrawNumber(uint32(limit), fg, x+36, y+72)
	c.DrawNumber(amount, fg, x+236, y+72)
	return nil
}

// DrawRecruitConfirm 畫徵兵成本與確認提示。
func (c *Canvas) DrawRecruitConfirm(f CommandFonts, fg, bg assets.RGB,
	x, y, w, h int, cost uint32) error {
	c.fillRect(x, y, w, h, bg)
	if err := c.DrawEntry(f.W2, w2Gold, 2, fg, x+44, y+42, true); err != nil {
		return err
	}
	c.DrawNumber(cost, fg, x+132, y+42)
	c.DrawConfirmBox(f.W4, fg, bg, x+80, y+105)
	return nil
}

// DrawTrainConfirm 畫指令 13「司令欲練兵嗎？」的確認畫面。
func (c *Canvas) DrawTrainConfirm(f CommandFonts, fg, bg assets.RGB,
	x, y, w, h int) error {
	c.fillRect(x, y, w, h, bg)
	if err := c.DrawEntry(f.W2, w2Commander, 2, fg, x+54, y+52, true); err != nil {
		return err
	}
	if err := c.DrawEntry(f.W2, 12, 2, fg, x+134, y+52, true); err != nil {
		return err
	}
	c.DrawConfirmBox(f.W4, fg, bg, x+80, y+105)
	return nil
}

// DrawGeneralList 畫目前省份可查閱的將領清單。原版每頁五人；remake 在同一張
// 640×350 畫布排成兩欄，但保留原始 ID 順序、姓名字模與兵力數值。
func (c *Canvas) DrawGeneralList(f CommandFonts, generals *assets.GlyphFile,
	fg, bg assets.RGB, x, y, w, h int, ids []int, forces []uint16, selected int) error {
	c.fillRect(x, y, w, h, bg)
	if err := c.DrawEntry(f.W2, w2General, 2, fg, x+24, y+18, true); err != nil {
		return err
	}
	if err := c.DrawEntry(f.W2, w2Force, 2, fg, x+w-148, y+18, true); err != nil {
		return err
	}
	start := (selected / 20) * 20
	if start < 0 {
		start = 0
	}
	end := start + 20
	if end > len(ids) {
		end = len(ids)
	}
	for i := start; i < end; i++ {
		id := ids[i]
		visible := i - start
		cx := x + 24 + (visible/10)*215
		cy := y + 52 + (visible%10)*27
		if i == selected {
			c.strokeRect(cx-8, cy-3, 198, 22, fg)
		}
		c.DrawNumber(uint32(i+1), fg, cx+22, cy)
		if id > 0 {
			if err := c.DrawEntry(generals, id-1, 3, fg, cx+38, cy, true); err != nil {
				return err
			}
		}
		if i < len(forces) {
			c.DrawNumber(uint32(forces[i]), fg, cx+186, cy)
		}
	}
	if len(ids) > 20 {
		c.DrawNumber(uint32(start/20+1), fg, x+w-132, y+18)
	}
	return nil
}

// DrawViewMenu 畫原版 `sub_2C351` 的四項查閱選單。
func (c *Canvas) DrawViewMenu(f CommandFonts, fg, bg assets.RGB, x, y, w, h int) error {
	c.fillRect(x, y, w, h, bg)
	rows := [][]int{
		{5, w2OtherProvince},         // 查閱他省
		{5, w2Belong, w2AllProvince}, // 查閱所屬各省
		{5, w2General},               // 查閱將領
		{5, w2ProvinceName},          // 查閱省名
	}
	for i, entries := range rows {
		cy := y + 48 + i*52
		c.DrawNumber(uint32(i+1), fg, x+42, cy)
		wx := x + 62
		for _, entry := range entries {
			if err := c.DrawEntry(f.W2, entry, 2, fg, wx, cy, true); err != nil {
				return err
			}
			wx += 2 * GlyphAdvance
		}
	}
	return nil
}

// DrawProvinceSelect 畫「查閱他省」的省編號／省名清單與輸入值。
func (c *Canvas) DrawProvinceSelect(f CommandFonts, provinces *assets.GlyphFile,
	fg, bg assets.RGB, x, y, w, h int, input uint32) error {
	c.fillRect(x, y, w, h, bg)
	if err := c.DrawEntry(f.W2, 5, 2, fg, x+22, y+14, true); err != nil {
		return err
	}
	if err := c.DrawEntry(f.W2, w2OtherProvince, 2, fg, x+62, y+14, true); err != nil {
		return err
	}
	c.DrawNumber(input, fg, x+w-132, y+14)
	for i := 0; i < game.ProvinceCount; i++ {
		col, line := i/13, i%13
		cx, cy := x+18+col*145, y+48+line*22
		c.DrawNumber(uint32(i+1), fg, cx+28, cy)
		if err := c.DrawEntry(provinces, i, 3, fg, cx+38, cy, true); err != nil {
			return err
		}
	}
	return nil
}

// DrawProvinceChoice 畫選定他省後的兩條原版子路徑：省份概況／查閱將領。
func (c *Canvas) DrawProvinceChoice(f CommandFonts, provinces *assets.GlyphFile,
	fg, bg assets.RGB, x, y, w, h, province int) error {
	c.fillRect(x, y, w, h, bg)
	if province > 0 {
		if err := c.DrawEntry(provinces, province-1, 3, fg, x+28, y+18, true); err != nil {
			return err
		}
	}
	c.DrawNumber(1, fg, x+54, y+76)
	if err := c.DrawEntry(f.W2, w2Overview, 2, fg, x+76, y+76, true); err != nil {
		return err
	}
	c.DrawNumber(2, fg, x+54, y+132)
	if err := c.DrawEntry(f.W2, 5, 2, fg, x+76, y+132, true); err != nil {
		return err
	}
	return c.DrawEntry(f.W2, w2General, 2, fg, x+116, y+132, true)
}

// DrawProvinceNames 畫 `sub_2C12F` 的兩頁省名：第一頁固定 1..20，第二頁
// 21..當期可用省份數（北伐 36、抗戰兩期 39）。
func (c *Canvas) DrawProvinceNames(f CommandFonts, provinces *assets.GlyphFile,
	fg, bg assets.RGB, x, y, w, h, page, limit int) error {
	c.fillRect(x, y, w, h, bg)
	if err := c.DrawEntry(f.W2, 5, 2, fg, x+22, y+14, true); err != nil {
		return err
	}
	if err := c.DrawEntry(f.W2, w2ProvinceName, 2, fg, x+62, y+14, true); err != nil {
		return err
	}
	start, end := 1, 20
	if page == 2 {
		start, end = 21, limit
	}
	if limit < end {
		end = limit
	}
	for id := start; id <= end; id++ {
		i := id - start
		col, line := i/10, i%10
		cx, cy := x+30+col*205, y+58+line*27
		c.DrawNumber(uint32(id), fg, cx+30, cy)
		if err := c.DrawEntry(provinces, id-1, 3, fg, cx+42, cy, true); err != nil {
			return err
		}
	}
	c.DrawNumber(uint32(page), fg, x+w-132, y+14)
	return nil
}

// DrawProvinceDetail 以原版字模呈現 `sub_2A941` 的省份欄位。
func (c *Canvas) DrawProvinceDetail(f CommandFonts, pf PanelFonts, d PanelData,
	fg, bg assets.RGB, x, y, w, h int) error {
	c.fillRect(x, y, w, h, bg)
	if err := c.DrawEntry(pf.W3, int(d.ID)-1, 3, fg, x+22, y+14, true); err != nil {
		return err
	}
	if d.Province.Governor.Valid() {
		if err := c.DrawEntry(pf.Gen, int(d.Province.Governor)-1, 3, fg, x+130, y+14, true); err != nil {
			return err
		}
	}
	type row struct {
		file         *assets.GlyphFile
		index, width int
		value        uint32
	}
	p := d.Province
	rows := []row{
		{f.W2, w2Gold, 2, uint32(p.Gold)}, {f.W2, w2Food, 2, uint32(p.Food)},
		{f.W2, w2Ammo, 2, uint32(p.Ammo)}, {f.W2, w2Fuel, 2, uint32(p.Fuel)},
		{f.W2, w2Coal, 2, uint32(p.Coal)}, {f.W2, w2Iron, 2, uint32(p.Iron)},
		{f.W2, w2Land, 2, uint32(p.LandValue)}, {f.W2, w2Pop, 2, p.PopulationWan()},
		{f.W2, w2City, 2, uint32(p.Cities)}, {pf.W3, w3Arsenal, 3, uint32(p.Arsenals)},
		{f.W2, w2Force, 2, d.Force}, {f.W2, w2General, 2, uint32(d.Generals)},
		{pf.W3, w3Loyalty, 3, uint32(p.Loyalty)},
	}
	for i, r := range rows {
		col, line := i/7, i%7
		cx, cy := x+22+col*220, y+58+line*39
		if err := c.DrawEntry(r.file, r.index, r.width, fg, cx, cy, true); err != nil {
			return err
		}
		c.DrawNumber(r.value, fg, cx+194, cy)
	}
	return nil
}

// DrawOwnedProvinceOverview 畫「所屬各省」清單；排序維持省編號順序。
func (c *Canvas) DrawOwnedProvinceOverview(f CommandFonts, pf PanelFonts,
	fg, bg assets.RGB, x, y, w, h int, ids []int, forces []uint32) error {
	c.fillRect(x, y, w, h, bg)
	for _, entry := range []int{w2Belong, w2AllProvince, w2Overview} {
		if err := c.DrawEntry(f.W2, entry, 2, fg, x+22, y+14, true); err != nil {
			return err
		}
		x += 2 * GlyphAdvance
	}
	for i, id := range ids {
		col, line := i/10, i%10
		cx, cy := x-100+col*205, y+52+line*27
		c.DrawNumber(uint32(id), fg, cx+28, cy)
		if id > 0 {
			if err := c.DrawEntry(pf.W3, id-1, 3, fg, cx+38, cy, true); err != nil {
				return err
			}
		}
		if i < len(forces) {
			c.DrawNumber(forces[i], fg, cx+194, cy)
		}
	}
	return nil
}

// DrawGeneralDetail 畫原版查閱畫面的十一項核心數值。番號需要 FAN 字模，
// 目前先保留標籤；其餘欄位均用原版字模與已證實的記錄欄位／戰力公式。
func (c *Canvas) DrawGeneralDetail(f CommandFonts, pf PanelFonts, fan *assets.GlyphFile, fg, bg assets.RGB,
	x, y, w, h int, id int, g game.General, attack int) error {
	c.fillRect(x, y, w, h, bg)
	if err := c.DrawEntry(f.W2, w2General, 2, fg, x+22, y+16, true); err != nil {
		return err
	}
	if id > 0 {
		if err := c.DrawEntry(pf.Gen, id-1, 3, fg, x+86, y+16, true); err != nil {
			return err
		}
	}
	type labelPart struct {
		file         *assets.GlyphFile
		index, width int
	}
	rows := []struct {
		parts []labelPart
		value uint32
	}{
		{[]labelPart{{f.W2, w2Lead, 2}, {f.W2, w2Ability, 2}}, uint32(g.AbilityA)},
		{[]labelPart{{pf.W3, w3Loyalty, 3}}, uint32(g.AbilityB)},
		{[]labelPart{{f.W2, w2Experience, 2}}, uint32(g.Experience)},
		{[]labelPart{{f.W2, w2Branch, 2}}, uint32(g.Branch)},
		{[]labelPart{{f.W2, w2UnitTitle, 2}}, 0},
		{[]labelPart{{pf.W3, w3Soldiers, 3}}, uint32(g.Force)},
		{[]labelPart{{f.W2, w2Soldier, 2}, {pf.W3, w3Attack, 3}}, uint32(attack)},
		{[]labelPart{{f.W2, w2Armed, 2}, {f.W2, w2Degree, 2}}, uint32(g.F20)},
		{[]labelPart{{f.W2, w2Soldier, 2}, {f.W2, w2Skill, 2}}, uint32(g.F19)},
		{[]labelPart{{f.W2, w2Stamina, 2}}, uint32(g.Stamina)},
		{[]labelPart{{f.W2, w2Morale, 2}}, uint32(g.F30)},
	}
	for i, row := range rows {
		col, line := i/6, i%6
		cx, cy := x+22+col*220, y+54+line*42
		wx := cx
		for _, p := range row.parts {
			if err := c.DrawEntry(p.file, p.index, p.width, fg, wx, cy, true); err != nil {
				return err
			}
			wx += p.width * GlyphAdvance
		}
		if i == 4 {
			wx := cx + 72
			for _, part := range []uint8{g.FactionName, g.TitlePrefix} {
				if part == 0 || fan == nil {
					continue
				}
				if err := c.DrawEntry(fan, int(part)-1, 5, fg, wx, cy, true); err != nil {
					return err
				}
				wx += 5 * GlyphAdvance
			}
			if g.TitleNumber != 0 {
				c.DrawNumber(uint32(g.TitleNumber), fg, wx+42, cy)
				wx += 52
			}
			if g.TitleSuffix != 0 && fan != nil {
				if err := c.DrawEntry(fan, int(g.TitleSuffix)-1, 5, fg, wx, cy, true); err != nil {
					return err
				}
			}
		} else {
			c.DrawNumber(row.value, fg, cx+194, cy)
		}
	}
	return nil
}

// DrawReorganizationTarget 畫同兵種候選將領與目前剩餘兵力。
func (c *Canvas) DrawReorganizationTarget(f CommandFonts, generals *assets.GlyphFile,
	fg, bg assets.RGB, x, y, w, h int, targets []int, remaining int, input uint32) error {
	c.fillRect(x, y, w, h, bg)
	if err := c.DrawEntry(f.W4, 23, 4, fg, x+24, y+20, true); err != nil {
		return err
	}
	c.DrawNumber(uint32(remaining), fg, x+190, y+20)
	for i, id := range targets {
		cx := x + 30 + (i/10)*205
		cy := y + 56 + (i%10)*24
		c.DrawNumber(uint32(i+1), fg, cx, cy)
		if id > 0 {
			if err := c.DrawEntry(generals, id-1, 3, fg, cx+38, cy, true); err != nil {
				return err
			}
		}
	}
	c.DrawNumber(input, fg, x+w-132, y+20)
	return nil
}

// DrawReorganizationAmount 畫選中將領、可指派上限與目前輸入值。
func (c *Canvas) DrawReorganizationAmount(f CommandFonts, generals *assets.GlyphFile,
	fg, bg assets.RGB, x, y, w, h int, generalID, remaining, limit int, amount uint32) error {
	c.fillRect(x, y, w, h, bg)
	if err := c.DrawEntry(f.W2, w2Commander, 2, fg, x+24, y+28, true); err != nil {
		return err
	}
	if generalID > 0 {
		if err := c.DrawEntry(generals, generalID-1, 3, fg, x+88, y+28, true); err != nil {
			return err
		}
	}
	if err := c.DrawEntry(f.W2, w2HowMany, 2, fg, x+172, y+28, true); err != nil {
		return err
	}
	if err := c.DrawEntry(f.W2, w2Force, 2, fg, x+236, y+28, true); err != nil {
		return err
	}
	if err := c.DrawEntry(f.W4, 23, 4, fg, x+24, y+72, true); err != nil {
		return err
	}
	c.DrawNumber(uint32(remaining), fg, x+190, y+72)
	// DrawNumber 的座標是右邊界，不是起點；預留完整數字寬度，避免伸回左側面板。
	c.DrawNumber(uint32(limit), fg, x+170, y+118)
	c.DrawNumber(amount, fg, x+w-34, y+118)
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
