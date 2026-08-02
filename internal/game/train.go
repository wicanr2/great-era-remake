package game

import "fmt"

// TrainResult 是玩家執行指令 13「練兵」後的結果。
type TrainResult struct {
	Generals int
	Gain     int
}

// Train 依 `sub_1C916` 提升一省可行動將領的士兵戰技。
//
// 原版逐筆掃描將領表，只檢查所屬省與 `+16 == 1`；每人的增量是
// 帶兵能力 `+0` 除以 10（整數除法），結果夾到 100。
func (w *AIWorld) Train(p ProvinceID) (TrainResult, error) {
	if _, err := w.Table.At(p); err != nil {
		return TrainResult{}, err
	}
	var out TrainResult
	for i := range w.Units {
		u := &w.Units[i]
		if !u.Active || u.Province != p || i >= len(w.Strengths) {
			continue
		}
		s := &w.Strengths[i]
		before := int(s.F19)
		after := before + int(s.Ability)/10
		if after > 100 {
			after = 100
		}
		s.F19 = uint8(after)
		out.Generals++
		out.Gain += after - before
	}
	if out.Generals == 0 {
		return TrainResult{}, fmt.Errorf("game: 省 %d 沒有可練兵的將領", p)
	}
	return out, nil
}
