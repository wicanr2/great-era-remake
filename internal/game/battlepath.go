package game

import "github.com/wicanr2/great-era-remake/internal/assets"

// 戰場 AI 尋路。
//
// 原版先由 sub_4FCCC 建立 196×196 的鄰接成本矩陣，再由
// sub_5778B／sub_5770F 做逐點鬆弛。矩陣的權重不是玩家實際移動成本：
// 鐵路 1、河海 30、沙漠 100；一般地形沿用 byte_9E2，佔用格是 80，
// 不可通行格是 0xFF（`docs/mechanics/70-ai.md` §6b、`docs/re/31` §58）。
//
// 這裡把已確認的矩陣規則落成確定性的 Dijkstra。它供 remake 的命令派工
// 回傳 `+12` 對應的下一跳；實際的 sub_567B9 還有目標周圍候選排序與
// 特定模式分支，尚未用原版 oracle 逐格核對，因此不能把這支稱為完全等價。

const (
	// AIPathRailWeight、AIPathWaterWeight、AIPathDesertWeight 出自
	// sub_4FCCC 建矩陣前暫時寫入 byte_9E2 的三組值。
	AIPathRailWeight     = 1
	AIPathWaterWeight    = 30
	AIPathDesertWeight   = 100
	AIPathOccupiedWeight = OccupiedCost // sub_4FCCC 的 50h
)

// battlePathWeight 回傳 AI 矩陣在「走進這一格」時使用的權重。
//
// 鐵路必須先判斷，與 sub_4FCCC／sub_506B0 的覆蓋順序一致；因此就算
// 底層地形是高山，只要資料真的標成鐵路，權重仍是 1。高山無鐵路由呼叫端
// 以 0xFF 視為不可達，不在這裡把它轉成另一個語意名稱。
func battlePathWeight(t assets.Tile) int {
	if t.HasRail() {
		return AIPathRailWeight
	}
	switch t.Kind {
	case assets.TileWater:
		return AIPathWaterWeight
	case assets.TileDesert:
		return AIPathDesertWeight
	default:
		return t.MoveCost()
	}
}

// RouteNextCell 回傳從 from 朝 to 的第一個下一跳；無法抵達時回 NoCell。
//
// 尋路時把目前單位的起點視為已清空（原版 sub_567B9 進入時也會暫存並
// 清掉 word_62A8[current]），目標若有單位則允許以 80 的矩陣成本抵達，
// 但中間的佔用格不列入候選，避免回傳給 Move 後必然撞上佔用 gate。
func (s *BattleSim) RouteNextCell(to, from CellIndex) CellIndex {
	if s == nil || s.Field == nil || !to.Valid() || !from.Valid() || to == from {
		return NoCell
	}

	// 玩家／AI 實際移動仍受長城 gate 約束；由起點佔用者取得已解出的
	// CanCross 性質。若呼叫端只提供裸格，保守沿用原版矩陣的可通行觀點。
	canCross := true
	if id := s.Occ[from]; id != 0 {
		if u := s.Unit(id); u != nil {
			canCross = u.CanCross
		}
	}

	path, _ := s.weightedBattlePath(to, from, canCross)
	if len(path) == 0 {
		return NoCell
	}
	return path[0]
}

// weightedBattlePath 回傳完整的下一跳序列與 AI 權重總和。
// 這個較低階的回傳值只供規則測試與除錯使用，玩家層仍只需要
// RouteNextCell 的第一格。
func (s *BattleSim) weightedBattlePath(to, from CellIndex, canCross bool) ([]CellIndex, int) {
	if s == nil || s.Field == nil || !to.Valid() || !from.Valid() || to == from {
		return nil, 0
	}
	const inf = int(^uint(0) >> 1)

	dist := [CellCount]int{}
	prev := [CellCount]CellIndex{}
	seen := [CellCount]bool{}
	for i := 0; i < CellCount; i++ {
		dist[i] = inf
		prev[i] = NoCell
	}
	dist[from] = 0

	for iter := 0; iter < CellCount; iter++ {
		best := NoCell
		bestDist := inf
		// sub_5770F 的選點是嚴格小於；索引由 0 往上掃，故相同距離
		// 保留較小格號，讓重製版的 tie-break 可重現。
		for i := 0; i < CellCount; i++ {
			c := CellIndex(i)
			if !seen[c] && dist[c] < bestDist {
				best, bestDist = c, dist[c]
			}
		}
		if best == NoCell {
			break
		}
		seen[best] = true
		if best == to {
			break
		}

		for _, next := range best.Neighbours() {
			if seen[next] {
				continue
			}
			// 矩陣確實把佔用格標成 80；但 sub_567B9 的候選下一跳
			// 會先排除佔用格。保留目標格的 80，避免把敵方目標當成
			// 不存在；其餘佔用格不加入，確保下一跳可以實際執行。
			occupied := s.Occ[next] != 0
			if occupied && next != to {
				continue
			}

			col, row := next.ColRow()
			tile := s.Field.Tiles[row][col]
			if !canCross && tile.Kind.Blocks() && next != to {
				continue
			}
			weight := battlePathWeight(tile)
			if occupied {
				weight = AIPathOccupiedWeight
			} else if weight >= assets.MoveCostImpassable {
				continue
			}

			alt := dist[best] + weight
			if alt < dist[next] {
				dist[next] = alt
				prev[next] = best
			}
		}
	}

	if dist[to] == inf {
		return nil, 0
	}
	rev := make([]CellIndex, 0, CellCount)
	for cur := to; cur != from; {
		p := prev[cur]
		if p == NoCell {
			return nil, 0
		}
		rev = append(rev, cur)
		cur = p
	}
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	return rev, dist[to]
}
