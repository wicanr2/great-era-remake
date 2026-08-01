package game

// 決策鏈 A 步驟 4 的**目標挑選三層優先序**。
//
// 步驟 4 收集完候選鄰省之後，並不是隨便挑一個，而是依序試三種挑法
// （`sub_17ADA+33B`..`+38A`）：
//
//	1. sub_170C0  ← 候選裡有「被敵人完全包圍」的省嗎        （aiencircled.go）
//	2. sub_17135  ← 挑「與我方接壤最多」的省，至少要 2 個    （本檔）
//	3. sub_14EBA  ← 都不行就挑黃金最多的                     （ai.go richestTransfer）
//
// 第 2 層**只在 `byte_6FFCA & 4` 設起來時才試**，否則直接跳到第 3 層。
// 這是 bit 2 的**第三個用途**（前兩個：啟用決策鏈最後三步、切換補給判定模式）。

// AIFriendlyNeighbourMin 是「與我方接壤最多」這一層的門檻。
//
// `sub_17135` 算出最高的我方鄰省數之後：
//
//	cmp     [bp+var_D], 1
//	ja      回傳該省          ; > 1，也就是 **≥ 2**
//	否則回 0
//
// 只接壤一個我方省不算數——那種省一旦被切斷就孤立了。
// 要求兩個以上，等於要求**目標省至少有兩條補給線**。
const AIFriendlyNeighbourMin = 2

// FriendlyNeighbourCount 數某省有幾個鄰省是我方的（`sub_17135` 的內圈）。
//
// 過濾條件只有三個，比 `EnemyNeighbours` 少一層：
//
//	鄰省 != 0 且 != 255（SeaBorder）
//	省[鄰省].司令 == faction
//
// ⚠️ 這裡的 `faction` 是**呼叫者的勢力**（狀態結構的 `[-2]`），
// 不是目標省自己的司令——與 `Encircled` 的「用該省自己的司令」不同。
func (w *AIWorld) FriendlyNeighbourCount(p ProvinceID, faction GeneralID) int {
	prov, err := w.Table.At(p)
	if err != nil {
		return 0
	}
	n := 0
	for _, nb := range prov.Neighbours {
		np, err := w.Table.At(nb)
		if err != nil {
			continue
		}
		if np.Commander == faction {
			n++
		}
	}
	return n
}

// MostConnectedTarget 在候選省裡挑**與我方接壤最多**的（`sub_17135`）。
//
// 回 0 表示沒有任何候選達到 `AIFriendlyNeighbourMin`——呼叫端這時退回
// `richestTransfer` 那一層（挑黃金最多的）。
//
// ⚠️ 更新條件是**嚴格大於**（`jbe` 跳過相等），所以並列時保留**較前面**的候選。
// 這與 `sub_14EBA` 的排序一致，都是「相等不動」。候選的順序來自鄰省表，
// 所以結果**與鄰省表的排列有關**——remake 不可以改鄰省順序，否則行為會漂。
func (w *AIWorld) MostConnectedTarget(cands []ProvinceID, faction GeneralID) ProvinceID {
	best, bestCount := ProvinceID(0), 0
	for _, c := range cands {
		n := w.FriendlyNeighbourCount(c, faction)
		if n > bestCount {
			best, bestCount = c, n
		}
	}
	if bestCount < AIFriendlyNeighbourMin {
		return 0
	}
	return best
}

// ChainATarget 跑完步驟 4 的三層優先序，回傳挑中的目標省（0 = 挑不出來）。
//
// `advanced` 來自 `byte_6FFCA & 4`：關掉的話**中間那一層整層跳過**，
// 電腦只會「有被包圍的就去救，否則挑最富的」。
//
// 讀作：**先救急、再求穩、最後求利**。
func (w *AIWorld) ChainATarget(cands []ProvinceID, faction GeneralID, advanced bool) ProvinceID {
	// 第 1 層：候選裡有沒有被敵人完全包圍的。
	for _, c := range cands {
		if w.Encircled(c) {
			return c
		}
	}
	// 第 2 層：只在進階模式試。
	if advanced {
		if t := w.MostConnectedTarget(cands, faction); t != 0 {
			return t
		}
	}
	// 第 3 層：黃金最多的。
	best, bestGold := ProvinceID(0), 0
	for _, c := range cands {
		np, err := w.Table.At(c)
		if err != nil {
			continue
		}
		if best == 0 || int(np.Gold) > bestGold {
			best, bestGold = c, int(np.Gold)
		}
	}
	return best
}
