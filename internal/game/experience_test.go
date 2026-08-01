package game

import "testing"

// `+3` = 經驗（`docs/playtest/08` 的「經驗」列）。
//
// ⚠️ 這一格在劇本檔裡**登場的將領全是 0**——先前掃 `MAN(N).DAT`
// 找不到它，不是因為它不在記錄裡，是因為在錯的檔案裡查。
//
// ⛔ 初稿寫「274 筆全是 0」，**被這條測試當場打臉**：第 238–241 與 274 筆有值。
// 前四筆是兵種範本（見 `TestBranchTemplateRecords`），第 274 筆是
// 不可用的槽位。正確的敘述是「**有所屬省的將領全是 0**」。
func TestExperienceIsZeroForDeployedGenerals(t *testing.T) {
	gs, err := ParseGenerals(readGame(t, "MAN(1).DAT"), Stage1GeneralCount)
	if err != nil {
		t.Fatal(err)
	}
	for i := range gs {
		if gs[i].Province == 0 {
			continue // 沒有所屬省的槽位（範本與未登場）不算
		}
		if gs[i].Experience != 0 {
			t.Fatalf("第 %d 位將領的經驗 = %d，劇本檔裡登場的將領該一律 0",
				i+1, gs[i].Experience)
		}
	}
}

// 第 238–241 筆是**兵種範本**：三個能力值、經驗、戰技、武裝、士氣
// 全部 100，兵力正好是四個兵種的滿員數。
//
// ⭐ 這解釋了 `docs/spec/02` §4 的方法為什麼會**剛好**得到社群傳的四個數字
// ——那不是「最強的將領恰好有 20000 兵」，是檔案裡明擺著的範本記錄。
func TestBranchTemplateRecords(t *testing.T) {
	gs, err := ParseGenerals(readGame(t, "MAN(1).DAT"), Stage1GeneralCount)
	if err != nil {
		t.Fatal(err)
	}
	want := []struct {
		idx    int
		branch uint8
		force  uint16
	}{
		{238, BranchInfantry, 20000},
		{239, BranchCavalry, 10000},
		{240, BranchArmour, 200},
		{241, BranchArtiller, 2000},
	}
	for _, w := range want {
		g := gs[w.idx-1]
		if g.Branch != w.branch || g.Force != w.force {
			t.Errorf("第 %d 筆：兵種 %d 兵力 %d，預期 %d／%d",
				w.idx, g.Branch, g.Force, w.branch, w.force)
		}
		for _, v := range []uint8{g.AbilityA, g.AbilityB, g.AbilityC,
			g.Experience, g.F19, g.F20, g.F30} {
			if v != 100 {
				t.Errorf("第 %d 筆的範本欄位該全是 100，實得 %d", w.idx, v)
			}
		}
		if g.Province != 0 {
			t.Errorf("第 %d 筆是範本，不該有所屬省（實得 %d）", w.idx, g.Province)
		}
	}
}

// ⭐ 第二、三期只有 **191** 位將領（`word_6BC4A`），
// 檔案裡第 192 筆之後是垃圾——照 274 筆解會得到兵種 186、所屬省 232 這種東西。
func TestStageGeneralCountsRejectGarbageTail(t *testing.T) {
	valid := func(g General) bool {
		return (g.Branch == 1 || g.Branch == 4 || g.Branch == 5 || g.Branch == 6) &&
			g.Province <= 39
	}
	for _, c := range []struct {
		file  string
		stage uint8
	}{{"MAN(1).DAT", 1}, {"MAN(2).DAT", 2}, {"MAN(3).DAT", 3}} {
		sc, err := ScenarioByStage(c.stage)
		if err != nil {
			t.Fatal(err)
		}
		data := readGame(t, c.file)
		gs, err := ParseGenerals(data, sc.Generals)
		if err != nil {
			t.Fatal(err)
		}
		for i := range gs {
			if !valid(gs[i]) {
				t.Errorf("%s 第 %d 筆在有效範圍內卻不合法：兵種 %d 所屬省 %d",
					c.file, i+1, gs[i].Branch, gs[i].Province)
			}
		}
		// 尾端確實是垃圾（第一期沒有尾端）。
		if sc.Generals*GeneralRecordSize >= len(data) {
			continue
		}
		tail, err := ParseGenerals(data[sc.Generals*GeneralRecordSize:],
			len(data)/GeneralRecordSize-sc.Generals)
		if err != nil {
			t.Fatal(err)
		}
		bad := 0
		for i := range tail {
			if !valid(tail[i]) {
				bad++
			}
		}
		if bad == 0 {
			t.Errorf("%s 第 %d 筆之後該是垃圾，卻全部合法——將領數可能不是 %d",
				c.file, sc.Generals+1, sc.Generals)
		}
	}
}

func TestExperienceMatchesScreenInSave(t *testing.T) {
	// 實機「查閱將領」（`docs/playtest/08` §1）：
	// 吳佩孚 30，陳家謨／楊源溏／劉佐龍／孫建業 全 0。
	gs, err := ParseSaveGenerals(readGame(t, "SAVE(1).DT1"), 274)
	if err != nil {
		t.Fatal(err)
	}
	want := []uint8{30, 0, 0, 0, 0}
	for k, w := range want {
		if got := gs[57+k].Experience; got != w {
			t.Errorf("第 %d 位將領的經驗 = %d，實機畫面 %d", 58+k, got, w)
		}
	}
}

func TestSaveGeneralsOffsetAgreesWithHubeiBase(t *testing.T) {
	// 正對照：`general_test.go` 早就在用 7084 當吳佩孚（第 58 位）的位置。
	if got := SaveGeneralsOffset + 57*GeneralRecordSize; got != 7084 {
		t.Errorf("5203 + 57×33 = %d，該是 7084", got)
	}
}

func TestSaveGeneralsMatchScreenAttributes(t *testing.T) {
	// 存檔的士氣（`+30`）與畫面括號裡的存檔值逐格相同
	//（`docs/playtest/08` §4：存檔 65/59/36/47/25）。
	gs, err := ParseSaveGenerals(readGame(t, "SAVE(1).DT1"), 274)
	if err != nil {
		t.Fatal(err)
	}
	morale := []uint8{65, 59, 36, 47, 25}
	skill := []uint8{50, 44, 21, 32, 10} // 士兵戰技，畫面同值
	for k := range morale {
		if got := gs[57+k].F30; got != morale[k] {
			t.Errorf("第 %d 位將領的士氣 = %d，存檔值 %d", 58+k, got, morale[k])
		}
		if got := gs[57+k].F19; got != skill[k] {
			t.Errorf("第 %d 位將領的士兵戰技 = %d，畫面 %d", 58+k, got, skill[k])
		}
	}
}

func TestParseSaveGeneralsRejectsShortData(t *testing.T) {
	if _, err := ParseSaveGenerals(make([]byte, 100), 1); err == nil {
		t.Error("資料太短該回錯誤")
	}
}
