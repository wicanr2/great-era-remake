package game

import "github.com/wicanr2/great-era-remake/internal/assets"

// BattleCandidateOptions 是 `sub_567B9` 候選目的地切片所需的外部狀態。
//
// Reserved 對應原版的 `byte[65BAh+格]` 預約表。這裡刻意用 callback 注入，
// 不把尚未解出的預約表生命週期偷偷塞進 BattleSim；nil 表示呼叫端明確提供
// 「本次沒有既有預約」的空表。
type BattleCandidateOptions struct {
	// Mode 是 `sub_567B9` 的原始 mode byte。它只影響尚未閉合的尾端分支，
	// 不在已證實的六格前置排序中另造語意名稱。
	Mode uint8
	// EnableLastSteps 對應 `byte_6FFCA & 4`。開啟且不在特殊省份時，
	// 原版還會呼叫尚未解出的 `sub_562BF`／`sub_56729`，本 API 會 fail-closed。
	EnableLastSteps bool
	// Reserved 回傳候選格是否已被 `byte[65BAh+格]` 預約。
	Reserved func(CellIndex) bool
}

// BattleCandidateResult 是可回查證據邊界的候選選擇結果。
//
// Complete=false 代表結果不能當成原版目的地使用；呼叫端不得把 Cell
// 當成有效下一跳。NoCell 且 Complete=true 則是已證實的「沒有候選」分支。
type BattleCandidateResult struct {
	Cell     CellIndex
	Complete bool
	Reason   string
}

// SelectOriginalBattleCandidate 落地 `sub_567B9` 已閉合的前置與排序資料流。
//
// 重要的參數方向是：候選來自 target 的六個鄰格，current 只用來計算
// `sub_5619C` 成本上限、暫時清空佔用語意與最後的同防禦值目前格回退。這不是
// `RouteNextCell` 的替代品，也不會改寫 Occ、NextCell 或預約表；它是可獨立測試
// 的純選擇切片。`sub_562BF`、`sub_56729` 與預約表生命週期未解時，結果會
// 明確標成 incomplete。
func (s *BattleSim) SelectOriginalBattleCandidate(target, current CellIndex,
	opt BattleCandidateOptions) BattleCandidateResult {
	if s == nil || s.Field == nil || !target.Valid() || !current.Valid() {
		return BattleCandidateResult{Cell: NoCell, Reason: "戰場或格編號無效"}
	}

	unitID := s.Occ[current]
	if unitID == 0 {
		return BattleCandidateResult{Cell: NoCell, Reason: "目前格沒有可查詢的 runtime 單位"}
	}
	unit := s.Unit(unitID)
	if unit == nil {
		return BattleCandidateResult{Cell: NoCell, Reason: "目前格的 runtime 單位未建立"}
	}
	branch := unit.Branch()
	if _, ok := DefenceFactor(s.battleCandidateTile(current), branch); !ok {
		return BattleCandidateResult{Cell: NoCell, Reason: "目前格的兵種／地形沒有已證實防禦值"}
	}

	costLimit := originalBattleCandidateCostLimit(s, current, branch)
	defence := make([]CellIndex, 0, 6)
	for _, candidate := range target.Neighbours() {
		// 原版在掃描前先暫時清掉 current 的 `word_62A8` 佔用；因此
		// current 若恰好也是 target 鄰格，必須仍可進入候選清單。
		if s.Occ[candidate] != 0 && candidate != current {
			continue
		}
		if opt.Reserved != nil && opt.Reserved(candidate) {
			continue
		}
		tile, ok := s.battleCandidateTileOK(candidate)
		if !ok || tile.MoveCost() > costLimit {
			continue
		}
		defence = append(defence, candidate)
	}
	if len(defence) == 0 {
		return BattleCandidateResult{Cell: NoCell, Complete: true}
	}

	// `sub_56461` 與 `sub_56548` 使用兩份獨立 stack 清單；保留兩份副本，
	// 不把其中一支的選擇排序誤當成一般全序。
	distance := append([]CellIndex(nil), defence...)
	score := func(c CellIndex) int {
		return mustBattleCandidateDefence(s.battleCandidateTile(c), branch)
	}
	SortBattleCandidatesByDefense(defence, score)
	SortBattleCandidatesByTargetDistance(distance, target, func(c CellIndex) bool {
		return IsUnrailedWaterTile(s.battleCandidateTile(c))
	})

	special := originalBattleCandidateSpecialProvince(s.At)
	if opt.EnableLastSteps && !special {
		return BattleCandidateResult{
			Cell:     NoCell,
			Complete: false,
			Reason:   "sub_562BF／sub_56729 與預約生命週期尚未閉合",
		}
	}

	selected := SelectBattleCandidate(defence[0], distance[0], special, score)
	currentScore := score(current)
	if currentScore == score(selected) && Adjacent(current, target) {
		// `sub_567B9` 的最後回退：目前格與所選格防禦值相同，且目前格
		// 確實鄰接目標時，保留目前格而不是強迫改派。
		selected = current
	}
	return BattleCandidateResult{Cell: selected, Complete: true}
}

// originalBattleCandidateCostLimit 對照 `sub_5619C` 的 12／13 門檻。
// 分支 1／5 的名稱仍只用原始兵種代號；不在這裡替它們命名。
func originalBattleCandidateCostLimit(s *BattleSim, current CellIndex, branch uint8) int {
	water := 0
	for _, neighbour := range current.Neighbours() {
		tile, ok := s.battleCandidateTileOK(neighbour)
		if ok && IsUnrailedWaterTile(tile) {
			water++
		}
	}
	if water >= 4 && (branch == Branch1 || branch == Branch5) {
		return 13
	}
	return 12
}

func originalBattleCandidateSpecialProvince(p ProvinceID) bool {
	switch p {
	case 15, 20, 26, 29, 33, 34:
		return true
	default:
		return false
	}
}

func (s *BattleSim) battleCandidateTile(c CellIndex) assets.Tile {
	tile, _ := s.battleCandidateTileOK(c)
	return tile
}

func (s *BattleSim) battleCandidateTileOK(c CellIndex) (assets.Tile, bool) {
	if s == nil || s.Field == nil || !c.Valid() {
		return assets.Tile{}, false
	}
	col, row := c.ColRow()
	return s.Field.Tiles[row][col], true
}

func mustBattleCandidateDefence(tile assets.Tile, branch uint8) int {
	v, ok := DefenceFactor(tile, branch)
	if !ok {
		// 呼叫端先驗證目前格，候選又已通過可走成本門檻；這個分支只
		// 可能由不一致的外部資料觸發，保留一個低優先級值而不 panic。
		return -1
	}
	return v
}
