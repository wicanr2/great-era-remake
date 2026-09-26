package render

import (
	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
	uitheme "github.com/wicanr2/great-era-remake/internal/ui/theme"
)

// 戰鬥畫面的右側面板。
//
// 版面全部從實機截圖量到（`docs/playtest/14`，蔣中正攻打孫傳芳的江西那一場）：
//
//	江西省 [25]    8 月 1 日          ← 標題（黑字）
//	蔣中正  攻擊  孫傳芳              ← 攻方／守方（「攻擊」是暗紅）
//	────────────────────
//	     3    單位數      9
//	 33472    士兵數  28200
//	   335    黃　金   3655
//	  3348    糧　食  19089
//	 15300    彈　藥  16170
//	 23988    燃　料  12047
//	────────────────────
//
// **六個資料列的四種資源正好是 `sub_174C9` 從省份複製進參戰部隊表的那四格**
// （`docs/re/29` §2）——黃金／糧食／彈藥／燃料，攻守成對。
// 煤礦與鐵礦不在這裡，與「不可流動的原料」一致（`40-economy.md` §9）。

// 戰鬥面板的版面常數，單位是 BGI 640×350 的像素。
//
// ⚠️ 這些數字是**量出來的不是設計的**。改動前先回去看截圖。
const (
	// BattlePanelX 是面板左邊界（米黃底色的第一欄）。
	BattlePanelX = 451
	// battlePanelTitleY / battlePanelSideY 是前兩行的頂端。
	battlePanelTitleY = 8
	battlePanelSideY  = 26
	// battlePanelRuleY / battlePanelRuleY2 是上下兩條分隔線。
	battlePanelRuleY  = 45
	battlePanelRuleY2 = 146
	// battlePanelRowY 是第一個資料列的頂端，battlePanelRowH 是行高。
	// 實機六列的頂端是 49／65／81／97／113／129，間隔一律 16。
	battlePanelRowY = 49
	battlePanelRowH = 16
	// 中央標籤的排版。**三字與二字用不同的起點與字距**，兩者都是量到的：
	//	三字（單位數／士兵數）  x=520／538／556  → 起點 520、字距 18
	//	二字（黃金／糧食…）      x=530／550       → 起點 530、字距 20
	// 二字**不是把三字置中**——原版是「黃　金」那種分散排版，
	// 兩個字各佔一格、中間空一格，所以字距反而比三字大。
	battlePanelLabelX  = 520
	battlePanelLabelW  = 18
	battlePanelLabel2X = 530
	battlePanelLabel2W = 20
	// battlePanelAtkRight / battlePanelDefRight 是兩欄數值的右邊界。
	battlePanelAtkRight = 508
	battlePanelDefRight = 627
	// battlePanelNameX 是標題與攻守名字的起點，三字、字距 20。
	battlePanelTitleX = 458
	battlePanelAtkX   = 468
	battlePanelVsX    = 530
	battlePanelDefX   = 568
)

// 戰鬥面板的配色，取自實機截圖的實際像素值。
var (
	battlePanelInk   = assets.RGB{R: 0x00, G: 0x00, B: 0x00} // 標題與數值的黑
	battlePanelLabel = assets.RGB{R: 0x00, G: 0x00, B: 0xAA} // 標籤與分隔線的藍
	battlePanelVs    = assets.RGB{R: 0xAE, G: 0x00, B: 0x00} // 「攻擊」的暗紅
	battlePanelPaper = assets.RGB{R: 0xFF, G: 0xFF, B: 0xA2} // 底色米黃
)

// 標籤用的詞條索引。
//
// ⚠️ **這裡是 0-based**，與 `panel.go` 的 `w2Gold` 那組一致
// （`DrawEntry` 收的就是 0-based）。`docs/re/24` §4 講的 1-based
// 是**原版執行檔裡的立即數**，兩者差 1，換算時很容易錯。
const (
	w3UnitCount   = 48  // 3.15 詞 49「單位數」
	w3SoldierCnt  = 49  // 3.15 詞 50「士兵數」
	w2AttackLabel = 159 // 2.15 詞 160「攻擊」
	w2MoveLabel   = 168 // 2.15 詞 169「移動」
	w2Retreat     = 169 // 2.15 詞 170「撤退」
	w2Garrison    = 170 // 2.15 詞 171「駐軍」
	w2Inspect     = 5   // 2.15 詞 6「查閱」
	w2HowProvince = 39  // 2.15 詞 40「何省」
)

// BattleMenuMode 是戰鬥五項選單目前顯示的輸入層。
//
// 數值刻意與 `cmd/dsds` 的 battleMode 保持一致，但 render 不依賴
// Ebiten 或命令列程式；若未來換入口，只要遵守這個顯示契約即可。
type BattleMenuMode uint8

const (
	BattleMenuCommand BattleMenuMode = iota
	BattleMenuMove
	BattleMenuAttack
	BattleMenuRetreat
	BattleMenuGarrison
	BattleMenuInspect
)

// BattleSide 是交戰一方在面板上要顯示的六個數字。
type BattleSide struct {
	Units    uint32 // 單位數
	Soldiers uint32 // 士兵數
	Gold     uint32
	Food     uint32
	Ammo     uint32
	Fuel     uint32
	// Leader 是這一方的勢力領袖，用來取姓名字模（`MAN{期}15` 的槽位）。
	Leader game.GeneralID
}

// BattlePanelData 是戰鬥面板要顯示的全部內容。
type BattlePanelData struct {
	// Province 是交戰的省，用來取省名（`3.15` 前 39 條）與顯示編號。
	Province game.ProvinceID
	Month    uint8
	Day      uint8
	Attacker BattleSide
	Defender BattleSide

	// AIAction／AIMoves／AIFights 是守方 AI 上一回合的行動編號與動作次數。
	//
	// ⚠️ **這是 remake 新增的**，原版面板沒有這一區（`docs/re/31` 沒有
	// 任何「把決策值畫出來」的痕跡）。加它的理由是玩家看不到 AI 在做什麼
	// 就無從判斷戰鬥是否正常——而中文訊息要等 M6 的完整字型才畫得出來
	// （原版字模是場景子集，畫不出自由組合的字）。
	//
	// `AIAction == 0` 表示還沒有 AI 行動，不畫。
	AIAction, AIMoves, AIFights int

	// ShowBattleMenu 是 remake 戰鬥畫面才使用的五項指令區；一般面板測試
	// 預設為 false，因此不會改動既有原版欄位的逐像素基準。
	ShowBattleMenu bool
	BattleMenuMode BattleMenuMode
	// ShowBattleControls 是 M3 remake 外殼的三個大按鍵；預設 false，
	// 讓既有原版面板逐像素測試不被新增控制圖形污染。
	ShowBattleControls bool
	// RetreatActive 只在已取得正常玩家撤退候選清單的切片顯示輸入面板。
	// 候選清單由 cmd/dsds 注入，renderer 不推導省份關係。
	RetreatActive  bool
	RetreatInput   uint32
	RetreatTargets []game.ProvinceID
	// Style 只改 battle HUD 的色彩外殼；零值仍是原版米黃／藍紅。
	Style uitheme.UIStyle
}

// DrawBattlePanel 把戰鬥畫面的右側面板畫到畫布上。
//
// `f.Gen` 是該期的將領姓名字模（`MAN115` 等），用來畫攻守雙方的領袖名。
func (c *Canvas) DrawBattlePanel(d BattlePanelData, f PanelFonts) error {
	ink, label, vs, paper := battlePalette(d.Style)
	c.fillRect(BattlePanelX, 0, ModeBGIW-BattlePanelX, ModeBGIH, paper)

	// 第一行：省名（三字）+ 月日。編號與「月」「日」還沒接上詞條，
	// 目前只畫省名與兩個數字——**這是已知的缺口**，不假裝畫完了。
	if err := c.DrawEntry(f.W3, int(d.Province)-1, 3,
		ink, battlePanelTitleX, battlePanelTitleY, true); err != nil {
		return err
	}
	c.DrawSmallNumber(uint32(d.Month), ink, 596, battlePanelTitleY)
	c.DrawSmallNumber(uint32(d.Day), ink, 632, battlePanelTitleY)

	// ⚠️ remake 新增：守方 AI 的行動編號與動作次數（見 BattlePanelData）。
	// 畫在面板最下緣，避開所有對實機驗證過的欄位。
	if d.AIAction != 0 {
		const aiY = ModeBGIH - 14
		c.DrawSmallNumber(uint32(d.AIAction), ink, BattlePanelX+8, aiY)
		c.DrawSmallNumber(uint32(d.AIMoves), ink, BattlePanelX+40, aiY)
		c.DrawSmallNumber(uint32(d.AIFights), ink, BattlePanelX+72, aiY)
	}

	// 第二行：攻方 攻擊 守方。
	if err := c.drawLeaderWithColor(f, d.Attacker.Leader, battlePanelAtkX, battlePanelSideY, ink); err != nil {
		return err
	}
	if err := c.DrawEntry(f.W2, w2AttackLabel, 2,
		vs, battlePanelVsX, battlePanelSideY, true); err != nil {
		return err
	}
	if err := c.drawLeaderWithColor(f, d.Defender.Leader, battlePanelDefX, battlePanelSideY, ink); err != nil {
		return err
	}

	// 兩條分隔線。
	c.fillRect(BattlePanelX+7, battlePanelRuleY, ModeBGIW-BattlePanelX-14, 2, label)
	c.fillRect(BattlePanelX+7, battlePanelRuleY2, ModeBGIW-BattlePanelX-14, 2, label)

	// 六個資料列。前兩列的標籤是三字（`3.15`），後四列是二字（`2.15`）——
	// 二字的畫在中間一格，與原版「黃　金」那種分散排版一致。
	rows := []struct {
		font     *assets.GlyphFile
		entry    int
		width    int
		atk, def uint32
		x, adv   int
	}{
		{f.W3, w3UnitCount, 3, d.Attacker.Units, d.Defender.Units,
			battlePanelLabelX, battlePanelLabelW},
		{f.W3, w3SoldierCnt, 3, d.Attacker.Soldiers, d.Defender.Soldiers,
			battlePanelLabelX, battlePanelLabelW},
		{f.W2, w2Gold, 2, d.Attacker.Gold, d.Defender.Gold,
			battlePanelLabel2X, battlePanelLabel2W},
		{f.W2, w2Food, 2, d.Attacker.Food, d.Defender.Food,
			battlePanelLabel2X, battlePanelLabel2W},
		{f.W2, w2Ammo, 2, d.Attacker.Ammo, d.Defender.Ammo,
			battlePanelLabel2X, battlePanelLabel2W},
		{f.W2, w2Fuel, 2, d.Attacker.Fuel, d.Defender.Fuel,
			battlePanelLabel2X, battlePanelLabel2W},
	}
	for i, r := range rows {
		y := battlePanelRowY + i*battlePanelRowH
		if err := c.drawSpacedEntry(r.font, r.entry, r.width,
			label, r.x, y, r.adv); err != nil {
			return err
		}
		c.DrawSmallNumber(r.atk, ink, battlePanelAtkRight, y)
		c.DrawSmallNumber(r.def, ink, battlePanelDefRight, y)
	}
	if d.RetreatActive {
		return c.drawBattleRetreat(f, d, ink, label, vs, paper)
	}
	if d.ShowBattleMenu {
		if err := c.drawBattleCommandMenu(f, d.BattleMenuMode, label, vs); err != nil {
			return err
		}
	}
	if d.ShowBattleControls {
		if err := c.drawBattleControls(f, label, vs, paper); err != nil {
			return err
		}
	}
	return nil
}

func battlePalette(style uitheme.UIStyle) (ink, label, vs, paper assets.RGB) {
	ink, label, vs, paper = battlePanelInk, battlePanelLabel, battlePanelVs, battlePanelPaper
	if style.Name == uitheme.ModeModern {
		ink, label, vs, paper = style.Ink, style.AccentAlt, style.Accent, style.Panel
	}
	return ink, label, vs, paper
}

// drawBattleRetreat 畫目前證據閉合的撤退輸入外殼。它覆蓋右側的戰鬥
// 選單／控制鍵，但保留上方省名與左側戰場；數字鍵盤與指標命中共用
// BattleRetreatKeypadButton，不增加另一套觸控規則。
func (c *Canvas) drawBattleRetreat(f PanelFonts, d BattlePanelData,
	ink, label, vs, paper assets.RGB) error {
	const (
		overlayY = battlePanelSideY - 2
		overlayH = ModeBGIH - overlayY
		textX    = BattlePanelX + 7
		inputR   = BattlePanelX + 176
	)
	c.fillRect(BattlePanelX+1, overlayY, ModeBGIW-BattlePanelX-2, overlayH, paper)
	if err := c.DrawEntry(f.W2, w2Retreat, 2, vs, textX, overlayY+4, true); err != nil {
		return err
	}
	if err := c.DrawEntry(f.W2, w2HowProvince, 2, label,
		textX+58, overlayY+4, true); err != nil {
		return err
	}
	c.DrawNumber(d.RetreatInput, ink, inputR, overlayY+4)

	// 只顯示注入的已證實候選；未知場景不會由 renderer 自行補出鄰省。
	for i, id := range d.RetreatTargets {
		col, row := i%4, i/4
		x := textX + col*42 + 24
		y := overlayY + 38 + row*22
		c.DrawNumber(uint32(id), label, x, y)
	}

	for i := 0; i < 12; i++ {
		p := uilayout.BattleRetreatKeypadButton(i)
		c.fillRect(p.X, p.Y, p.HitW, p.HitH, paper)
		c.strokeRect(p.X, p.Y, p.HitW, p.HitH, label)
		switch {
		case i < 9:
			c.DrawNumber(uint32(i+1), label, p.X+36, p.Y+15)
		case i == 9:
			c.DrawNumber(0, label, p.X+36, p.Y+15)
		case i == 10:
			// 刪除：向左箭頭加尾端叉記，與一般數字鍵盤相同。
			c.fillRect(p.X+11, p.Y+22, 28, 4, label)
			for j := 0; j < 4; j++ {
				c.fillRect(p.X+12-j*2, p.Y+22-j*2, 4, 4, label)
				c.fillRect(p.X+12-j*2, p.Y+22+j*2, 4, 4, label)
			}
			c.fillRect(p.X+42, p.Y+17, 3, 14, label)
		case i == 11:
			c.drawCheckmark(label, p.X-4, p.Y)
		}
	}
	return nil
}

const (
	battleMenuY      = 164
	battleMenuRowH   = 24
	battleMenuNumX   = BattlePanelX + 10
	battleMenuLabelX = BattlePanelX + 30
)

// drawBattleCommandMenu 畫出已由 DOSBox／IDA 證實的五項戰鬥主選單。
//
// 這裡只顯示選單本身與目前輸入層；攻擊方式、駐軍、查閱的原版後續欄位
// 尚未閉合，但可玩的 remake 外殼會在 cmd/dsds 以同一命中區接線。
func (c *Canvas) drawBattleCommandMenu(f PanelFonts, mode BattleMenuMode,
	label, active assets.RGB) error {
	entries := [...]int{w2MoveLabel, w2AttackLabel, w2Retreat, w2Garrison, w2Inspect}
	for i, entry := range entries {
		y := battleMenuY + i*battleMenuRowH
		fg := label
		if mode == BattleMenuMode(i+1) {
			fg = active
		}
		c.DrawSmallDigit(i+1, fg, battleMenuNumX, y)
		if err := c.DrawEntry(f.W2, entry, 2, fg, battleMenuLabelX, y, true); err != nil {
			return err
		}
	}
	return nil
}

// drawBattleControls 畫三個不依賴新增字模的 remake 控制鍵：攻擊沿用原版
// 「攻擊」詞條，換部隊與結束回合使用程式圖形。三者的命中區由
// `internal/ui/layout.BattleControlButton` 共用，輸入端不可另造座標。
func (c *Canvas) drawBattleControls(f PanelFonts,
	label, active, paper assets.RGB) error {
	for i := 0; i < 3; i++ {
		p := uilayout.BattleControlButton(i)
		c.fillRect(p.X, p.Y, p.HitW, p.HitH, paper)
		c.strokeRect(p.X, p.Y, p.HitW, p.HitH, label)
		switch i {
		case 0:
			// 攻擊按鍵沿用已驗證的 2.15 詞條；原版第二層六鍵選單尚未完整映射，
			// 實際派送的是既有近身攻擊 remake 差異。
			if err := c.DrawEntry(f.W2, w2AttackLabel, 2, active,
				p.X+12, p.Y+18, true); err != nil {
				return err
			}
		case 1:
			// 兩個向右三角代表換到下一個攻方單位（Tab 等價）。
			for row := 0; row < 5; row++ {
				c.fillRect(p.X+16+row*2, p.Y+17+row*3, 4, 4, label)
				c.fillRect(p.X+28+row*2, p.Y+17+row*3, 4, 4, label)
			}
		case 2:
			// 方框加向下箭頭代表結束回合（Space 等價）。
			c.fillRect(p.X+18, p.Y+14, 20, 4, label)
			c.fillRect(p.X+18, p.Y+30, 20, 4, label)
			for row := 0; row < 4; row++ {
				c.fillRect(p.X+22+row*2, p.Y+20+row*2, 4, 4, label)
			}
		}
	}
	return nil
}

// drawLeader 畫一位領袖的姓名（三字）。0 表示沒有，什麼都不畫。
func (c *Canvas) drawLeader(f PanelFonts, id game.GeneralID, x, y int) error {
	return c.drawLeaderWithColor(f, id, x, y, battlePanelInk)
}

func (c *Canvas) drawLeaderWithColor(f PanelFonts, id game.GeneralID, x, y int, fg assets.RGB) error {
	if id == 0 || f.Gen == nil {
		return nil
	}
	return c.DrawEntry(f.Gen, int(id)-1, 3, fg, x, y, true)
}

// drawSpacedEntry 與 DrawEntry 相同，但字距可以指定。
//
// 戰鬥面板的標籤字距是 **18**（實機「單位數」三字的左邊界 520／538／556），
// 與政略面板的 20（`GlyphAdvance`）不同——所以不能直接用 DrawEntry。
func (c *Canvas) drawSpacedEntry(f *assets.GlyphFile, k, slotWidth int,
	fg assets.RGB, x, y, advance int) error {
	if f == nil {
		return nil
	}
	for i := 0; i < slotWidth; i++ {
		idx := k*slotWidth + i
		if idx < 0 || idx >= len(f.Glyphs) {
			continue
		}
		c.DrawGlyph(f.Glyphs[idx], fg, x+i*advance, y, true)
	}
	return nil
}
