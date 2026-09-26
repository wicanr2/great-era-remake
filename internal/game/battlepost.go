package game

// execDefaultPost 是 WAR.EXE sub_3D411 的窄實作。
//
// sub_3D57B（值 13）在自己的命令分流結束後，依外部 gate 才會呼叫這一段；
// 它只處理命令 4／5 的單位。這裡把原版傳入的 byte[65BAh + 格] 抽象成
// reserved：呼叫端必須提供同一回合的預約表，不能在這一層偷偷重設未知狀態。
// AutoResolveByChain 目前只有在 BattleChainGates.DefaultPostStageOpen 明確成立
// 時才建立空表並呼叫，這是可回歸的 remake 邊界，不宣稱預約表生命週期已完全閉合。
//
// IDA Pro 9.4 證據：WAR.EXE sub_3D411 0x3D411–0x3D575，於 0x3D543 呼叫
// sub_3D261；sub_55CEC(mode=1) 的敵鄰檢查見 docs/re/31 §63。
func (s *BattleSim) execDefaultPost(units []*Combatant, cities []CellIndex,
	route func(to, from CellIndex) CellIndex, reserved []bool) BattleExecResult {
	if s == nil {
		return BattleExecResult{Note: "sub_3D411：nil 戰鬥狀態，保持 fail-closed"}
	}
	if len(reserved) < CellCount {
		return BattleExecResult{Note: "sub_3D411：預約表不足，保持 fail-closed"}
	}

	assigned, blocked, fallback := 0, 0, 0
	for _, u := range units {
		if u == nil || !u.Alive() || !u.Cell.Valid() {
			continue
		}
		if u.Command != BattleCmdCommitted && u.Command != BattleCmdUnknown5 {
			continue
		}

		// sub_55CEC(arg_0=1) 只看六格相鄰位置上的 +14 勢力；遇到
		// runtime 佔用表與單位表不一致時，負數代表未知，採跳過單位。
		if n := s.enemyAdjacentCount(u); n < 0 || n > 0 {
			blocked++
			continue
		}

		found := false
		for _, city := range cities {
			if !city.Valid() || s.Occ[city] != 0 || reserved[city] {
				continue
			}
			// 原版只寫 +12 與預約表，不清 +10，也不立 +13 bit 7。
			u.NextCell = city
			reserved[city] = true
			assigned++
			found = true
			break
		}
		if found {
			continue
		}

		// 找不到空城時，原版把命令改成 2，再於 0x3D543 呼叫
		// sub_3D261。後備本身的 +10／+12／+13 不對稱由窄入口保留。
		u.Command = BattleCmdStandby
		if s.assignCityFallback(u, cities, route) {
			assigned++
			fallback++
		}
	}

	note := "命令 4／5 城市後處理"
	if blocked > 0 {
		note += "，敵鄰／未知佔用跳過 " + itoa(blocked)
	}
	if fallback > 0 {
		note += "，後備 " + itoa(fallback)
	}
	return BattleExecResult{Assigned: assigned, Implemented: true, Note: note}
}

// enemyAdjacentCount 對應 sub_55CEC(arg_0=1)：回傳目前單位六格相鄰的
// 不同 +14 勢力單位數。-1 表示 Occupancy 有 ID、但 runtime 表找不到該 ID；
// 呼叫端應把這種狀態視為未知而跳過，不自行猜測敵我。
func (s *BattleSim) enemyAdjacentCount(u *Combatant) int {
	if s == nil || u == nil || !u.Cell.Valid() {
		return -1
	}
	count := 0
	for _, cell := range u.Cell.Neighbours() {
		id := s.Occ[cell]
		if id == 0 {
			continue
		}
		v := s.Unit(id)
		if v == nil {
			return -1
		}
		if v.Faction != u.Faction {
			count++
		}
	}
	return count
}
