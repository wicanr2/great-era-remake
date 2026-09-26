package game

import "github.com/wicanr2/great-era-remake/internal/assets"

// IsUnrailedWaterTile 是 `sub_58FD9` 的逐步對照。
//
// 原版先以 `sub_50151` 判斷地物碼是否為 3，再以 `sub_4FEF0` 判斷
// 鐵路；只有「河海且沒有鐵路」才回傳 1。這不是把河海一律視為不可用，
// 鐵橋（河海＋鐵路）必須保留為另一種情形。
func IsUnrailedWaterTile(t assets.Tile) bool {
	return t.Kind == assets.TileWater && !t.HasRail()
}

// SortBattleCandidatesByDefense 對照 `sub_56461` 的候選清單排序。
//
// 原版是選擇排序：從前往後固定一格，後面的候選若
// `sub_503BB(候選格, 目前單位)` 較大就交換，因此防禦值降冪；相等時不換。
// score 封裝目前單位與地形資料的查詢，避免在這個純排序原語裡假定尚未
// 證實的單位欄位語意。
//
// `sub_567B9` 建立的清單只會把通過前置 gate 的格寫入；NoCell 防護是重製版
// 的 fail-closed 邊界，讓未完整初始化的清單不會被拿去查表。
func SortBattleCandidatesByDefense(candidates []CellIndex, score func(CellIndex) int) {
	if len(candidates) < 2 || score == nil {
		return
	}
	for i := 0; i+1 < len(candidates); i++ {
		if candidates[i] == NoCell {
			continue
		}
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j] == NoCell {
				continue
			}
			if score(candidates[j]) <= score(candidates[i]) {
				continue
			}
			candidates[i], candidates[j] = candidates[j], candidates[i]
		}
	}
}

// SortBattleCandidatesByTargetDistance 對照 `sub_56548` 的第二份排序清單。
//
// 距離是 `col = cell mod 14`、`row = cell div 14` 的矩形曼哈頓距離，
// 不是六角距離。較近的後項只有在不是「無鐵路河海」時才能前移；因此
// 不能改寫成一般 sort.Slice 的全序比較器，必須保留原版的巢狀選擇排序。
// isUnrailedWater 對應 `sub_58FD9`，由呼叫端以格資料提供。
func SortBattleCandidatesByTargetDistance(candidates []CellIndex, target CellIndex,
	isUnrailedWater func(CellIndex) bool) {
	if len(candidates) < 2 || !target.Valid() {
		return
	}
	for i := 0; i+1 < len(candidates); i++ {
		current := candidates[i]
		if current == NoCell || !current.Valid() {
			continue
		}
		for j := i + 1; j < len(candidates); j++ {
			later := candidates[j]
			if later == NoCell || !later.Valid() {
				continue
			}
			if CellManhattan(later, target) >= CellManhattan(current, target) {
				continue
			}
			if isUnrailedWater != nil && isUnrailedWater(later) {
				continue
			}
			candidates[i], candidates[j] = candidates[j], candidates[i]
		}
	}
}

// SelectBattleCandidate 對照 `sub_566B6` 的兩份排序結果選擇器。
//
// primary 與 alternate 是呼叫端暫存區的兩個第一候選；這裡刻意不用
// 「防禦排序／距離排序」命名，因為 `sub_567B9` 的區域 copy 尚未取得足夠
// 證據可將兩個 stack slot 與高階名稱一一對上。原版規則是：特殊旗標為真
// 時選 alternate；否則兩格的 `sub_503BB` 值不同選 primary，相等選
// alternate。相同格則直接回 primary。
func SelectBattleCandidate(primary, alternate CellIndex, specialFlag bool,
	defenseScore func(CellIndex) int) CellIndex {
	if primary == alternate {
		return primary
	}
	if specialFlag {
		return alternate
	}
	if defenseScore == nil {
		return primary
	}
	if defenseScore(primary) == defenseScore(alternate) {
		return alternate
	}
	return primary
}
