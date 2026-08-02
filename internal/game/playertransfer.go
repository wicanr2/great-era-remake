package game

import "fmt"

// PlayerTransferMode 是玩家指令 1 的兩個畫面模式。
// 它與 transfer.go 的 AI TransferMode 0..7 是兩套不同的介面。
type PlayerTransferMode uint8

const (
	PlayerTransferPartial PlayerTransferMode = 1
	PlayerTransferAll     PlayerTransferMode = 2

	// provinceFlagTransferBlocked 是省份 +32 bit 4。原版 sub_2E2F3
	// 明確排除設了此位元的調動目標，但語意尚未定名；保留原始位元語意，
	// 不把推測寫成公開名稱。
	provinceFlagTransferBlocked = 0x10

	PlayerTransferStaminaMinimum = 30
	PlayerTransferStaminaCost    = 10
)

// PlayerTransferSelection 保存玩家在確認前的選將狀態。
// 建立與切換都不改遊戲狀態；取消畫面不需要復原任何資料。
type PlayerTransferSelection struct {
	from, to ProvinceID
	mode     PlayerTransferMode
	capacity int
	cands    []GeneralID
	eligible map[GeneralID]bool
	selected map[GeneralID]bool
}

// PlayerTransferReport 是玩家確認調動後實際寫入的結果。
type PlayerTransferReport struct {
	Moved    []GeneralID
	Supplies [4]int
}

func (s *PlayerTransferSelection) From() ProvinceID         { return s.from }
func (s *PlayerTransferSelection) To() ProvinceID           { return s.to }
func (s *PlayerTransferSelection) Mode() PlayerTransferMode { return s.mode }

// Candidates 回傳原版掃描順序的來源省將領。
func (s *PlayerTransferSelection) Candidates() []GeneralID {
	return append([]GeneralID(nil), s.cands...)
}

// Selected 回傳目前會被搬走的將領，順序與 Candidates 相同。
func (s *PlayerTransferSelection) Selected() []GeneralID {
	out := make([]GeneralID, 0, len(s.cands))
	for _, id := range s.cands {
		if s.selected[id] {
			out = append(out, id)
		}
	}
	return out
}

// Toggle 切換一位將領是否搬遷。體力不足 30 的候選不能被勾選。
func (s *PlayerTransferSelection) Toggle(id GeneralID) error {
	if _, ok := s.eligible[id]; !ok {
		return fmt.Errorf("game: 將領 %d 不在玩家調動候選中", id)
	}
	if !s.eligible[id] {
		return fmt.Errorf("game: 將領 %d 體力不足 %d", id, PlayerTransferStaminaMinimum)
	}
	if !s.selected[id] && len(s.Selected()) >= s.capacity {
		return fmt.Errorf("game: 目標省只能再容納 %d 位將領", s.capacity)
	}
	s.selected[id] = !s.selected[id]
	return nil
}

// PlayerTransferTargets 回傳玩家指令 1 可輸入的鄰省，保留原始鄰接表順序。
func (w *AIWorld) PlayerTransferTargets(from ProvinceID) ([]ProvinceID, error) {
	src, err := w.Table.At(from)
	if err != nil {
		return nil, err
	}
	if src.Commander == 0 {
		return nil, nil
	}
	var out []ProvinceID
	for _, id := range src.Neighbours {
		dst, err := w.Table.At(id)
		if err != nil || dst.InBattle() || dst.Flags&provinceFlagTransferBlocked != 0 {
			continue
		}
		if dst.Commander == 0 || dst.Commander == src.Commander || dst.Commander == GeneralID(0xffff) {
			out = append(out, id)
		}
	}
	return out, nil
}

// BeginPlayerTransfer 建立不改狀態的選將工作階段。
// 部分調動從空集合開始；全部調動預選所有體力至少 30 的候選。
func (w *AIWorld) BeginPlayerTransfer(from, to ProvinceID, mode PlayerTransferMode,
	gens []General) (*PlayerTransferSelection, error) {
	if mode != PlayerTransferPartial && mode != PlayerTransferAll {
		return nil, fmt.Errorf("game: 未知玩家調動模式 %d", mode)
	}
	targets, err := w.PlayerTransferTargets(from)
	if err != nil {
		return nil, err
	}
	valid := false
	for _, id := range targets {
		valid = valid || id == to
	}
	if !valid {
		return nil, fmt.Errorf("game: 省 %d 不是省 %d 的合法調動目標", to, from)
	}
	src, _ := w.Table.At(from)
	targetCount := 0
	for _, u := range w.Units {
		if u.Active && u.Province == to && u.Faction == src.Commander {
			targetCount++
		}
	}
	capacity := TransferTargetGeneralCap - targetCount
	if capacity <= 0 {
		return nil, fmt.Errorf("game: 省 %d 已有 %d 位將領", to, targetCount)
	}
	s := &PlayerTransferSelection{from: from, to: to, mode: mode, capacity: capacity,
		eligible: make(map[GeneralID]bool), selected: make(map[GeneralID]bool)}
	for _, u := range w.Units {
		if !u.Active || u.Province != from || u.Faction != src.Commander {
			continue
		}
		s.cands = append(s.cands, u.General)
		i := int(u.General) - 1
		ok := i >= 0 && i < len(gens) && gens[i].Stamina >= PlayerTransferStaminaMinimum
		s.eligible[u.General] = ok
		if mode == PlayerTransferAll && ok {
			s.selected[u.General] = true
		}
	}
	if len(s.cands) == 0 {
		return nil, fmt.Errorf("game: 省 %d 沒有可調動的將領", from)
	}
	if mode == PlayerTransferAll && len(s.cands) > capacity {
		return nil, fmt.Errorf("game: 省 %d 無法容納全部 %d 位將領", to, len(s.cands))
	}
	return s, nil
}

// ApplyPlayerTransfer 在玩家按下確認後一次套用選將與四種物資。
// 驗證失敗時完全不改狀態；選取畫面本身仍由 BeginPlayerTransfer 保持唯讀。
func (w *AIWorld) ApplyPlayerTransfer(s *PlayerTransferSelection, gens []General,
	amounts [4]int) (PlayerTransferReport, error) {
	var rep PlayerTransferReport
	if s == nil {
		return rep, fmt.Errorf("game: 玩家調動工作階段不可為 nil")
	}
	selected := s.Selected()
	if len(selected) == 0 {
		return rep, fmt.Errorf("game: 玩家調動至少要選一位將領")
	}
	if len(selected) > s.capacity {
		return rep, fmt.Errorf("game: 選了 %d 位將領，但目標只容納 %d 位", len(selected), s.capacity)
	}
	src, err := w.Table.At(s.from)
	if err != nil {
		return rep, err
	}
	dst, err := w.Table.At(s.to)
	if err != nil {
		return rep, err
	}
	leader := src.Commander
	if leader == 0 {
		return rep, fmt.Errorf("game: 來源省 %d 已無司令", s.from)
	}
	legalTargets, err := w.PlayerTransferTargets(s.from)
	if err != nil || !containsProvince(legalTargets, s.to) {
		return rep, fmt.Errorf("game: 目標省 %d 已不符合調動條件", s.to)
	}
	currentTargetCount := 0
	for _, u := range w.Units {
		if u.Active && u.Province == s.to && u.Faction == leader {
			currentTargetCount++
		}
	}
	if currentTargetCount+len(selected) > TransferTargetGeneralCap {
		return rep, fmt.Errorf("game: 目標省調動後會有 %d 位將領，超過上限 %d",
			currentTargetCount+len(selected), TransferTargetGeneralCap)
	}

	// 先驗證所有會改動的欄位，使錯誤不會留下半套用狀態。
	unitByGeneral := make(map[GeneralID]int, len(w.Units))
	for i := range w.Units {
		unitByGeneral[w.Units[i].General] = i
	}
	for _, id := range selected {
		i, ok := unitByGeneral[id]
		gi := int(id) - 1
		if !ok || !w.Units[i].Active || w.Units[i].Province != s.from ||
			w.Units[i].Faction != leader || gi < 0 || gi >= len(gens) ||
			gens[gi].Stamina < PlayerTransferStaminaMinimum {
			return rep, fmt.Errorf("game: 將領 %d 已不符合調動條件", id)
		}
	}
	for i, good := range []SupplyGood{SupplyGold, SupplyFood, SupplyAmmo, SupplyFuel} {
		if amounts[i] < 0 {
			return rep, fmt.Errorf("game: %s調動量不可為負數", SupplyGoodName(good))
		}
		if f := supplyField(src, good); amounts[i] > int(*f) {
			return rep, fmt.Errorf("game: %s只有 %d，搬不了 %d", SupplyGoodName(good), *f, amounts[i])
		}
	}

	// 原版順序：先物資，再搬將領，最後修正兩省的司令／省長。
	rep.Supplies, err = w.SupplyAll(s.from, s.to, amounts)
	if err != nil { // 上面的完整預檢應使此分支不可達。
		return PlayerTransferReport{}, err
	}
	for _, id := range selected {
		i := unitByGeneral[id]
		gi := int(id) - 1
		w.Units[i].Province = s.to
		gens[gi].Province = s.to
		gens[gi].Stamina -= PlayerTransferStaminaCost
		if i < len(w.Strengths) {
			w.Strengths[i].F29 = gens[gi].Stamina
		}
		rep.Moved = append(rep.Moved, id)
	}
	if containsGeneral(selected, leader) {
		dst.Flags &^= ProvinceFlagAutonomous
	}
	dst.Commander = leader
	w.repairPlayerTransferProvince(s.to, leader, gens)
	w.repairPlayerTransferProvince(s.from, leader, gens)
	return rep, nil
}

func containsGeneral(ids []GeneralID, want GeneralID) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

func containsProvince(ids []ProvinceID, want ProvinceID) bool {
	for _, id := range ids {
		if id == want {
			return true
		}
	}
	return false
}

// repairPlayerTransferProvince 重現 sub_5A541 在調動後的領導欄位修正。
// 原版先依 +2（政治手腕）降序，再依 +1（忠誠度）降序；第二輪是主要鍵。
func (w *AIWorld) repairPlayerTransferProvince(p ProvinceID, leader GeneralID, gens []General) {
	prov, err := w.Table.At(p)
	if err != nil {
		return
	}
	var candidates []GeneralID
	for _, u := range w.Units {
		if !u.Active || u.Province != p || u.Faction != leader {
			continue
		}
		if u.General == leader {
			prov.Commander, prov.Governor = leader, leader
			return
		}
		candidates = append(candidates, u.General)
	}
	if len(candidates) == 0 {
		prov.Commander, prov.Governor = 0, 0
		prov.Flags &^= ProvinceFlagAutonomous
		return
	}
	less := func(a, b GeneralID, field func(General) uint8) bool {
		ai, bi := int(a)-1, int(b)-1
		if ai < 0 || ai >= len(gens) || bi < 0 || bi >= len(gens) {
			return false
		}
		return field(gens[ai]) < field(gens[bi])
	}
	selectionSort := func(field func(General) uint8) {
		for i := 0; i < len(candidates); i++ {
			for j := i + 1; j < len(candidates); j++ {
				if less(candidates[i], candidates[j], field) {
					candidates[i], candidates[j] = candidates[j], candidates[i]
				}
			}
		}
	}
	selectionSort(func(g General) uint8 { return g.AbilityC })
	selectionSort(func(g General) uint8 { return g.AbilityB })
	prov.Commander, prov.Governor = leader, candidates[0]
}
