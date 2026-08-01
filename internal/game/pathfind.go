package game

// 省份層的路徑搜尋，對應 `sub_5BF74`（`docs/re/11` §5）。
//
// 原版配置一張 39 × 39 的矩陣（`GetMem(5F1h)` = 1,521），輸出
// `byte_70013`（路徑長度）與 `[-41C5h + i]`（路徑節點）。
//
// **權重模式 11 已經對齊**（`docs/re/15`）：`sub_5BD7B` 填的成本矩陣
// 裡，通得過的邊**成本一律 1**，所以原版就是無權最短路。
// 「只走同勢力的省」那條硬條件我們也照做了。
//
// ⚠️ 兩處仍是差異：
//
//  1. `sub_5BA56`（權重模式 10，344 行）還沒讀，可能是帶權版本。
//     目前沒看到誰用模式 10。
//  2. **同分時挑哪一條**：原版依省編號由小到大掃（`for j = 1..39`），
//     這裡依鄰省表順序。有多條等長路徑時會分歧。

// ProvincePath 找 from 到 to 的省份路徑，含頭尾。
//
// 回傳的順序是 `[from, ..., to]`。找不到路徑回 nil。
// 只走**司令相同**的省——原版的兩個權重模式（10 / 11）都帶勢力參數，
// 顯然路徑要經過自己的地盤。
func (w *AIWorld) ProvincePath(from, to ProvinceID, faction GeneralID) []ProvinceID {
	if from == to {
		return []ProvinceID{from}
	}
	prev := make(map[ProvinceID]ProvinceID, ProvinceCount)
	seen := map[ProvinceID]bool{from: true}
	queue := []ProvinceID{from}

	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		prov, err := w.Table.At(cur)
		if err != nil {
			continue
		}
		for _, n := range w.neighbours(prov) {
			if seen[n] {
				continue
			}
			np, err := w.Table.At(n)
			if err != nil {
				continue
			}
			// 終點本身不必是自己的省（救援目標一定是，但保持一般性）。
			if n != to && np.Commander != faction {
				continue
			}
			seen[n] = true
			prev[n] = cur
			if n == to {
				return buildPath(prev, from, to)
			}
			queue = append(queue, n)
		}
	}
	return nil
}

func buildPath(prev map[ProvinceID]ProvinceID, from, to ProvinceID) []ProvinceID {
	var rev []ProvinceID
	for p := to; ; p = prev[p] {
		rev = append(rev, p)
		if p == from {
			break
		}
	}
	for i, j := 0, len(rev)-1; i < j; i, j = i+1, j-1 {
		rev[i], rev[j] = rev[j], rev[i]
	}
	return rev
}

// NextHop 回傳從 from 往 to 的**第一步**，也就是 `sub_14CCF` 真正要的東西。
//
// 原版挑的不是目的地而是路徑上與當前省相鄰的那一格
// （`docs/re/11` §5）——調動一次只能走一格，遠處的目標要接力。
//
// 三道檢查照原版：
//
//	路徑長 > 1（單格路徑要額外驗相鄰）
//	第一步 != 當前省
//	第一步與當前省相鄰
func (w *AIWorld) NextHop(from, to ProvinceID, faction GeneralID) ProvinceID {
	path := w.ProvincePath(from, to, faction)
	if len(path) < 2 {
		return 0
	}
	hop := path[1]
	if hop == from || !w.adjacentProvinces(hop, from) {
		return 0
	}
	return hop
}

// RescueThreshold 是「快撐不住」的門檻：戰力總和低於它才算候選
// （`sub_149C8` 的 `cmp ax, 7D0h`）。
const RescueThreshold = 2000

// rescueCandidate 是一個待救援的省與它的兩個排序指標。
type rescueCandidate struct {
	prov ProvinceID
	// hostileNeighbours 是敵方鄰省數（`sub_14860` 存的第二個 byte）。
	hostileNeighbours int
	// worstRelation 是鄰省裡敵方勢力的最大關係值（第一個 byte）。
	//
	// ⚠️ 原版查的是一張 59×59 的矩陣（`docs/re/11` §3），
	// 那張表的**內容還沒解**。這裡先用「敵方鄰省數」當替代排序鍵，
	// 並標為 remake 差異——不是猜一個公式填進去。
	worstRelation int
}

// RescueTargets 回傳排序後的救援候選，語意照 `sub_149C8` + 兩輪排序。
//
// 篩選（四條全部照原版）：
//
//	省份[p].+20 == 我方勢力
//	sub_5B7DC(p) != 0        必須是前線
//	sub_5B983(p) < 2000      戰力掉到門檻以下
//	省份[p].+32 & 40h == 0   沒在交戰
//
// **掃全部 39 省，不是只看鄰省**——這是它與其他調動最大的差別。
func (w *AIWorld) RescueTargets(faction GeneralID) []ProvinceID {
	var cands []rescueCandidate
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		prov, err := w.Table.At(p)
		if err != nil || prov.Commander != faction {
			continue
		}
		if w.Hostile(p) == 0 {
			continue
		}
		if w.ProvincePower(p) >= RescueThreshold {
			continue
		}
		if prov.Flags&ProvinceFlagInBattle != 0 {
			continue
		}
		c := rescueCandidate{prov: p}
		for _, n := range w.neighbours(prov) {
			np, err := w.Table.At(n)
			if err != nil || np.Commander == 0 || np.Commander == faction {
				continue
			}
			c.hostileNeighbours++
		}
		c.worstRelation = c.hostileNeighbours
		cands = append(cands, c)
	}
	// 兩輪都是降序，第二輪（關係值）是最終主鍵。
	sortLikeOriginal(cands, func(a, b rescueCandidate) bool {
		return a.hostileNeighbours > b.hostileNeighbours
	})
	sortLikeOriginal(cands, func(a, b rescueCandidate) bool {
		return a.worstRelation > b.worstRelation
	})
	out := make([]ProvinceID, 0, len(cands))
	for _, c := range cands {
		out = append(out, c.prov)
	}
	return out
}

// sortLikeOriginal 照 `sub_14B0D`／`sub_14BEE` 的寫法：外層 i、內層
// j = i+1，比較後直接交換。**這不是穩定排序**，同分時挑到誰取決於
// 交換的順序——所以不能換成 `sort.SliceStable`，那會挑到不同的省。
func sortLikeOriginal(s []rescueCandidate, less func(a, b rescueCandidate) bool) {
	for i := 0; i < len(s); i++ {
		for j := i + 1; j < len(s); j++ {
			if less(s[j], s[i]) {
				s[i], s[j] = s[j], s[i]
			}
		}
	}
}
