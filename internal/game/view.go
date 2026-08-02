package game

// ActiveGeneralsAt 回傳原版查閱將領畫面可見的人員，順序是將領 ID 由小到大。
// `sub_2BADC` 使用既有將領表順序，並只收同省且 `+16 == 1` 的記錄。
func (w *AIWorld) ActiveGeneralsAt(p ProvinceID) []GeneralID {
	out := make([]GeneralID, 0)
	for i := range w.Units {
		u := &w.Units[i]
		if u.Active && u.Province == p {
			out = append(out, u.General)
		}
	}
	return out
}

// OwnedProvinces 回傳與指定省同司令的省份，保持原版 1..39 掃描順序。
// 查閱選單第二項的原版文字是「查閱所屬各省」。
func (w *AIWorld) OwnedProvinces(p ProvinceID) []ProvinceID {
	prov, err := w.Table.At(p)
	if err != nil || prov.Commander == 0 {
		return nil
	}
	out := make([]ProvinceID, 0)
	for id := ProvinceID(1); id <= ProvinceCount; id++ {
		q, err := w.Table.At(id)
		if err == nil && q.Commander == prov.Commander {
			out = append(out, id)
		}
	}
	return out
}
