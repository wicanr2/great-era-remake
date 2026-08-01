package game

import "testing"

func TestScenarioTableMatchesOriginal(t *testing.T) {
	// §50 的表：三個劇本的起始參數。
	want := []struct {
		stage     uint8
		year, mon uint8
		provinces int
	}{
		{1, 15, 7, 36},
		{2, 26, 7, 39},
		{3, 30, 12, 39},
	}
	for i, w := range want {
		got := Scenarios[i]
		if got.Stage != w.stage || got.Year != w.year || got.Month != w.mon ||
			got.Provinces != w.provinces {
			t.Errorf("第 %d 個劇本 = %+v，預期 %+v", i+1, got, w)
		}
	}
}

func TestScenario1EndsWhereScenario2Begins(t *testing.T) {
	// ⭐ 第一期的時限（`gameover.go`）就是第二期的開局日期——
	// 北伐打到七七事變就強制結束接抗戰。兩邊是獨立解出來的，對得上。
	s2, err := ScenarioByStage(2)
	if err != nil {
		t.Fatal(err)
	}
	if int(s2.Year) != Stage1DeadlineYear || int(s2.Month) != Stage1DeadlineMonth {
		t.Errorf("第二期開局 民國%d年%d月，第一期時限 民國%d年%d月，該一致",
			s2.Year, s2.Month, Stage1DeadlineYear, Stage1DeadlineMonth)
	}
}

func TestScenarioLookupRejectsUnknownStage(t *testing.T) {
	if _, err := ScenarioByStage(0); err == nil {
		t.Error("第 0 期不存在，該回錯誤")
	}
	if _, err := ScenarioByStage(4); err == nil {
		t.Error("第 4 期不存在，該回錯誤")
	}
}

func TestNorthernExpeditionHasFewerProvinces(t *testing.T) {
	// 北伐時期的隨機事件挑不到最後三個省。
	s1, _ := ScenarioByStage(1)
	if s1.RandomProvinceRange() != 36 || ProvinceCount != 39 {
		t.Errorf("北伐 %d 省、檔案 %d 筆記錄", s1.RandomProvinceRange(), ProvinceCount)
	}
}
