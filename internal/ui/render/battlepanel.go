package render

import (
	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
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
}

// DrawBattlePanel 把戰鬥畫面的右側面板畫到畫布上。
//
// `f.Gen` 是該期的將領姓名字模（`MAN115` 等），用來畫攻守雙方的領袖名。
func (c *Canvas) DrawBattlePanel(d BattlePanelData, f PanelFonts) error {
	c.fillRect(BattlePanelX, 0, ModeBGIW-BattlePanelX, ModeBGIH, battlePanelPaper)

	// 第一行：省名（三字）+ 月日。編號與「月」「日」還沒接上詞條，
	// 目前只畫省名與兩個數字——**這是已知的缺口**，不假裝畫完了。
	if err := c.DrawEntry(f.W3, int(d.Province)-1, 3,
		battlePanelInk, battlePanelTitleX, battlePanelTitleY, true); err != nil {
		return err
	}
	c.DrawSmallNumber(uint32(d.Month), battlePanelInk, 596, battlePanelTitleY)
	c.DrawSmallNumber(uint32(d.Day), battlePanelInk, 632, battlePanelTitleY)

	// 第二行：攻方 攻擊 守方。
	if err := c.drawLeader(f, d.Attacker.Leader, battlePanelAtkX, battlePanelSideY); err != nil {
		return err
	}
	if err := c.DrawEntry(f.W2, w2AttackLabel, 2,
		battlePanelVs, battlePanelVsX, battlePanelSideY, true); err != nil {
		return err
	}
	if err := c.drawLeader(f, d.Defender.Leader, battlePanelDefX, battlePanelSideY); err != nil {
		return err
	}

	// 兩條分隔線。
	c.fillRect(BattlePanelX+7, battlePanelRuleY, ModeBGIW-BattlePanelX-14, 2, battlePanelLabel)
	c.fillRect(BattlePanelX+7, battlePanelRuleY2, ModeBGIW-BattlePanelX-14, 2, battlePanelLabel)

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
			battlePanelLabel, r.x, y, r.adv); err != nil {
			return err
		}
		c.DrawSmallNumber(r.atk, battlePanelInk, battlePanelAtkRight, y)
		c.DrawSmallNumber(r.def, battlePanelInk, battlePanelDefRight, y)
	}
	return nil
}

// drawLeader 畫一位領袖的姓名（三字）。0 表示沒有，什麼都不畫。
func (c *Canvas) drawLeader(f PanelFonts, id game.GeneralID, x, y int) error {
	if id == 0 || f.Gen == nil {
		return nil
	}
	return c.DrawEntry(f.Gen, int(id)-1, 3, battlePanelInk, x, y, true)
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
