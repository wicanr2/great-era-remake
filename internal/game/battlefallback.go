package game

// assignCityFallback 是 WAR.EXE `sub_3D261` 的可回查窄實作。
//
// 這不是一般化的「找最近城市」：原版先取城市清單裡第一個相鄰城市，
// 只有該城市沒有可直接處理的駐軍時，才依原清單順序掃描同勢力的佔用城市。
// `+10`、`+12`、`+13` 的不對稱寫入也刻意保留；不能用 CombatUnit.AssignTo
// 代替，因為原版有些成功分支不立 `+13` bit 7。
//
// IDA Pro 9.4 證據：WAR.EXE `sub_3D261` 0x3D261–0x3D40B，
// 輸入雜湊與位址空間見 `docs/re/31-battle-ai-chain.md` §62。
func (s *BattleSim) assignCityFallback(u *Combatant, cities []CellIndex,
	route func(to, from CellIndex) CellIndex) bool {
	if s == nil || u == nil {
		return false
	}

	// `sub_560D7(1, currentCell)` 只回第一個相鄰城市；這裡不能在城市
	// 空著時改挑第二個相鄰城市，否則就把原版的清單順序改掉了。
	for _, c := range cities {
		if !Adjacent(c, u.Cell) {
			continue
		}
		id := s.Occ[c]
		if id == 0 {
			break
		}
		v := s.Unit(id)
		if v == nil {
			// 原版不可能有脫離 runtime 表的佔用者；remake 遇到這種
			// 不一致時停止這個直接分支，避免猜測未知單位的 +8。
			break
		}
		u.NextCell = u.Cell // `+12 = current +5`，不是城市格
		if v.Attacking {
			// +8 != 0：寫 +10，**不**立 +13 bit 7。
			u.TargetUnit = id
		} else {
			// +8 == 0：只寫 +12 並立 bit 7；原版不清舊 +10，照抄。
			u.Flags13 |= UnitAssignedBit
		}
		return true
	}

	// 直接鄰城分支沒有命中時，原版先清 +10，再依複製的城市清單順序
	// 找「佔用者效忠勢力相同」的城市。尋路失敗仍會留下最後一個候選
	// 的 +10 與 `+12 = 0xFF`，所以寫入順序不能延後到成功之後。
	u.TargetUnit = 0
	for _, c := range cities {
		if !c.Valid() {
			continue
		}
		id := s.Occ[c]
		if id == 0 {
			continue
		}
		v := s.Unit(id)
		if v == nil || v.Faction != u.Faction {
			continue
		}
		u.TargetUnit = id
		u.NextCell = NoCell
		if route == nil {
			continue
		}
		next := route(c, u.Cell)
		u.NextCell = next
		if next != NoCell {
			return true
		}
	}
	return false
}
