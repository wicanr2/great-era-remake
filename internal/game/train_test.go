package game

import "testing"

func TestTrainUsesAbilityDivTenAndCapsAtHundred(t *testing.T) {
	tbl := &ProvinceTable{}
	w := &AIWorld{
		Table: tbl,
		Units: []CombatUnit{
			{Province: 1, Active: true},
			{Province: 1, Active: true},
			{Province: 1, Active: false},
			{Province: 2, Active: true},
		},
		Strengths: []StrengthInput{
			{Ability: 87, F19: 50},
			{Ability: 100, F19: 96},
			{Ability: 90, F19: 20},
			{Ability: 90, F19: 20},
		},
	}
	got, err := w.Train(1)
	if err != nil {
		t.Fatal(err)
	}
	if got.Generals != 2 || got.Gain != 12 {
		t.Fatalf("結果 = %+v，預期 2 人、總提升 12", got)
	}
	if w.Strengths[0].F19 != 58 || w.Strengths[1].F19 != 100 {
		t.Fatalf("戰技 = %d/%d，預期 58/100", w.Strengths[0].F19, w.Strengths[1].F19)
	}
	if w.Strengths[2].F19 != 20 || w.Strengths[3].F19 != 20 {
		t.Fatal("不可行動或別省將領不應受影響")
	}
}

func TestTrainRejectsProvinceWithoutActiveGeneral(t *testing.T) {
	w := &AIWorld{Table: &ProvinceTable{}, Units: []CombatUnit{{Province: 1}}}
	if _, err := w.Train(1); err == nil {
		t.Fatal("沒有可行動將領時應失敗")
	}
}
