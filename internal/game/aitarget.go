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

// ── 步驟 6：被包圍時的突圍目標（`sub_17236`）────────────────────────────

// AIBreakoutMinProvinces 是突圍的前置門檻：**我方省份數要 > 2**。
//
//	cmp     [bp+var_4], 2
//	ja      繼續              ; > 2，也就是 **≥ 3**
//	否則直接回 0
//
// 剩兩個省以下就不突圍了——沒有可以接應的後方，打出去也無處可去。
const AIBreakoutMinProvinces = 3

// FactionProvinceCount 數某勢力有幾個省（`sub_17236` 的第一段迴圈）。
//
// 原版掃 `1 .. word_7001E`，`word_7001E` 就是省份總數。
func (w *AIWorld) FactionProvinceCount(faction GeneralID) int {
	n := 0
	for i := ProvinceID(1); ; i++ {
		p, err := w.Table.At(i)
		if err != nil {
			break
		}
		if p.Commander == faction {
			n++
		}
	}
	return n
}

// BreakoutTarget 挑被包圍時要打哪個鄰省（`sub_17236`）。
//
// 兩層鄰省搜尋：
//
//	for nb1 in 當前省的鄰省:
//	    for nb2 in nb1 的鄰省:
//	        if 省[nb2].司令 == 我方 且 nb2 != 當前省:
//	            回 nb1                          ← 回的是**第一層**
//
// 讀作：**突圍的方向要朝著能連回自己領地的那一邊**。
// 打下 `nb1` 之後，我方就從 `nb2` 那一側接上了，不會變成一塊飛地。
//
// 回 0 表示挑不出來（呼叫端就不出兵）。
//
// ⚠️ 函式**不檢查 `nb1` 是不是敵方**。步驟 6 的前提是 `Encircled` 為真，
// 那時每個鄰省都是敵人，所以不需要再查——但這條前提是呼叫端給的，
// 單獨用這個函式時要自己保證。照抄原版，不補防禦。
//
// ⚠️ **第一個找到就回**，不比較好壞。候選順序來自鄰省表，
// 與 `MostConnectedTarget` 一樣受排列影響。
func (w *AIWorld) BreakoutTarget(current ProvinceID, faction GeneralID) ProvinceID {
	if w.FactionProvinceCount(faction) < AIBreakoutMinProvinces {
		return 0
	}
	prov, err := w.Table.At(current)
	if err != nil {
		return 0
	}
	for _, nb1 := range prov.Neighbours {
		p1, err := w.Table.At(nb1)
		if err != nil {
			continue
		}
		for _, nb2 := range p1.Neighbours {
			if nb2 == current {
				continue
			}
			p2, err := w.Table.At(nb2)
			if err != nil || p2.Commander != faction {
				continue
			}
			return nb1
		}
	}
	return 0
}

// AIChainASkippedProvinces 是**開頭就跳過整條決策鏈 A** 的省
// （`sub_17ADA+E`..`+36`）：37 臺灣省、38 海南島、39 緬甸。
//
// 條件是「當前省在這三個裡面 **且** `byte_6FFCA & 2`」——bit 1 沒設就照跑。
//
// 這三個省各只有**一個**鄰省（臺灣→24 福建、海南島→36 廣東、緬甸→34 雲南），
// 而且初始都無主。整條決策鏈的核心（挑鄰省、三層優先序、包圍判定）
// 對只有一個鄰省的省沒有意義，`Encircled` 更會因為「唯一的鄰省是敵人」
// 而永遠成立，誤觸發突圍。
//
// ⛔ `docs/re/28` §2 原本把這一段記成「那三個**勢力**不跑決策鏈 A」。
// `ss:[arg_0+0Ah]` 是**當前省編號**不是勢力——它被原封不動傳給
// `sub_16E77(狀態, 省)` 與 `sub_5A9F6(省)`，兩支的參數都是省。
var AIChainASkippedProvinces = [...]ProvinceID{37, 38, 39}

// AIBreakoutSkippedProvinces 是**步驟 6 單獨排除**的省（`sub_17ADA+525`..`+533`）：
// 只有 37 臺灣省與 38 海南島，**不含 39 緬甸**。
//
// 兩份清單不一樣是有道理的：緬甸與雲南陸路相連，臺灣與海南島是離島。
// 突圍要打出去，離島打不出去。
var AIBreakoutSkippedProvinces = [...]ProvinceID{37, 38}

// AIChainASkipsProvince 回報某省會不會在決策鏈 A 開頭就被跳過。
//
// `bit1` 來自 `byte_6FFCA & 2`。
func AIChainASkipsProvince(p ProvinceID, bit1 bool) bool {
	if !bit1 {
		return false
	}
	for _, s := range AIChainASkippedProvinces {
		if s == p {
			return true
		}
	}
	return false
}

// AIBreakoutSkipsProvince 回報某省會不會被步驟 6 排除（與 `byte_6FFCA` 無關）。
func AIBreakoutSkipsProvince(p ProvinceID) bool {
	for _, s := range AIBreakoutSkippedProvinces {
		if s == p {
			return true
		}
	}
	return false
}

// ── 步驟 5：依戰力排序候選（`sub_16047`）────────────────────────────────

// SortProvincesByStrength 依**省份戰力總和**（`sub_5B983`）就地排序候選清單
// （`sub_16047`）。
//
//	mode 1（ascending = true）  → 升序，**戰力最弱的排最前面**
//	mode 2（ascending = false） → 降序，**戰力最強的排最前面**
//
// 呼叫端都只取排序後的第一個，所以實際語意是「挑最弱的／挑最強的」：
//
//	決策鏈 A 步驟 5（`sub_17ADA+4E2`）：mode 1 → **挑最弱的鄰省打**
//	                                    → `sub_174C9(狀態, 3)` 出兵模式 3
//	`sub_1655C+45`（未解）：            mode 2 → 取**最強的敵對鄰省**做比較
//
// 步驟 5 的前置門檻很高（`byte_6FFCA & 2`、`[-234h] > 3`、
// `[-236h] ≥ 63,392`），所以它是「**兵多將廣時的擴張**」——
// 而擴張的對象是**柿子挑軟的捏**。
//
// ⚠️ 這是**就地排序**，會改動呼叫端的陣列，不是回傳一個省。
// `docs/re/28` 的欄位掃描把 `es:[di-1]` 記成「回傳值」，
// 實際上那是寫回候選陣列本身。
//
// ⚠️ 原版是雙迴圈選擇排序，相等時不交換（`jbe`／`jnb` 都跳過）。
// 這種寫法**不是穩定排序**——交換會打亂中間元素的相對順序。
// 這裡照抄同樣的雙迴圈而不用 `sort.SliceStable`，
// 因為呼叫端可能依賴「第一名以外的排列」，兩者在並列時結果不同。
func SortProvincesByStrength(list []ProvinceID, ascending bool,
	strength func(ProvinceID) int) {
	for i := 0; i+1 < len(list); i++ {
		for j := i + 1; j < len(list); j++ {
			a, b := strength(list[i]), strength(list[j])
			swap := a > b
			if !ascending {
				swap = a < b
			}
			if swap {
				list[i], list[j] = list[j], list[i]
			}
		}
	}
}

// WeakestNeighbourTarget 是決策鏈 A 步驟 5 挑目標的完整語意：
// **候選鄰省裡戰力總和最小的那一個**。
//
// 候選的收集條件在呼叫端（`sub_17ADA` `loc_17ED1`）：
// 有效鄰省、**跳過交戰中的**（省份 `+32` bit 6）。
//
// 回 0 表示沒有候選。
func (w *AIWorld) WeakestNeighbourTarget(cands []ProvinceID) ProvinceID {
	if len(cands) == 0 {
		return 0
	}
	sorted := append([]ProvinceID(nil), cands...)
	SortProvincesByStrength(sorted, true, func(p ProvinceID) int {
		return ProvinceStrength(p, w.Units, w.Strengths, w.Opts)
	})
	return sorted[0]
}
