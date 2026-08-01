package game

import "testing"

// 十大勢力（`docs/spec/03` §3）。SAVE(1) 的司令欄位裡就是這些人。
func majorSet(t *testing.T, w *AIWorld) map[GeneralID]bool {
	t.Helper()
	m := map[GeneralID]bool{}
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		if prov, err := w.Table.At(p); err == nil && prov.Commander != 0 {
			m[prov.Commander] = true
		}
	}
	return m
}

func testState(t *testing.T, w *AIWorld) GameState {
	major := majorSet(t, w)
	return GameState{
		Stage: 1, Year: 15, Month: 8,
		MajorPower: func(id GeneralID) bool { return major[id] },
		LeaderAvailable: func(id GeneralID) bool {
			for i := range w.Units {
				if w.Units[i].General == id {
					return w.Units[i].Active
				}
			}
			return false
		},
	}
}

// 開局不該結束。
func TestGameNotOverAtStart(t *testing.T) {
	w := realWorld(t)
	if over, code := w.CheckGameOver(testState(t, w)); over {
		t.Errorf("開局就判結束了，結局代碼 %d", code)
	}
}

// 第一期的時限：民國 26 年 7 月。
//
// 民國 26 年 = 1937 年，7 月 = 盧溝橋事變。
func TestStage1Deadline(t *testing.T) {
	w := realWorld(t)
	cases := []struct {
		year, month uint8
		wantOver    bool
	}{
		{26, 6, false}, // 差一個月
		{26, 7, true},  // 正好
		{26, 8, true},  // 之後
		{25, 12, false},
	}
	for _, c := range cases {
		st := testState(t, w)
		st.Year, st.Month = c.year, c.month
		over, code := w.CheckGameOver(st)
		if over != c.wantOver {
			t.Errorf("民國 %d 年 %d 月：over=%v，應為 %v", c.year, c.month, over, c.wantOver)
		}
		if over && code != EndingDefeatStage1 {
			t.Errorf("民國 %d 年 %d 月的結局代碼是 %d，應為 %d",
				c.year, c.month, code, EndingDefeatStage1)
		}
	}
	// 第二期沒有這道時限。
	st := testState(t, w)
	st.Stage, st.Year, st.Month = 2, 26, 7
	if over, _ := w.CheckGameOver(st); over {
		t.Error("第二期不該套用第一期的時限")
	}
}

// 十大勢力全滅 = 覆滅。
func TestDefeatWhenNoMajorLeaderLeft(t *testing.T) {
	w := realWorld(t)
	st := testState(t, w)
	st.LeaderAvailable = func(GeneralID) bool { return false }
	over, code := w.CheckGameOver(st)
	if !over || code != EndingDefeatStage1 {
		t.Errorf("十大勢力全滅應判覆滅，得到 over=%v code=%d", over, code)
	}
	st.Stage = 3
	if _, code := w.CheckGameOver(st); code != EndingDefeatLater {
		t.Errorf("第三期覆滅的代碼是 %d，應為 %d", code, EndingDefeatLater)
	}
}

// 觀戰模式只看統一，而且不寫結局代碼。
func TestWatchingModeUnificationOnly(t *testing.T) {
	w := realWorld(t)
	st := testState(t, w)
	st.Watching = true
	if over, _ := w.CheckGameOver(st); over {
		t.Error("開局有九個勢力，觀戰模式不該判結束")
	}
	// 把所有有主省改成同一個司令。
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		if prov, err := w.Table.At(p); err == nil && prov.Commander != 0 {
			prov.Commander = 58
		}
	}
	over, code := w.CheckGameOver(st)
	if !over {
		t.Error("天下統一了，觀戰模式應該判結束")
	}
	if code != EndingNone {
		t.Errorf("觀戰模式不寫結局代碼，得到 %d", code)
	}
}
