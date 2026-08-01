package game

// 勢力領袖遷都（`sub_1398D`，`docs/re/13` §4）。
//
// 電腦回合的固定一步，排在兩條決策鏈**之前**：
//
//	內政 → sub_19B89 → **遷都** → 決策鏈 A → 決策鏈 B
//
// 它也是**第一個被確認會消耗命令數**的行為。

// RelocateResult 是一次遷都的結果。
type RelocateResult struct {
	// Moved 為真表示真的搬了。
	Moved bool
	// From / To 是搬遷前後的省。
	From, To ProvinceID
}

// Relocate 讓某勢力的領袖從前線省搬到後方的鄰省（`sub_1398D`）。
//
//	領袖不在這個省          → 不搬
//	這個省不是前線          → 不搬（沒有威脅就不用動）
//	鄰省要是**我方的、非交戰中的、後方的**
//	找不到符合的鄰省        → 不搬
//
// 找到之後做三件事：
//
//	將領[領袖].+4  = 目標省      ← 領袖搬過去
//	省份[目標].+20 = 領袖        ← 司令
//	省份[目標].+30 = 領袖        ← **省長也換成領袖本人**
//
// **領袖所在的省變成前線就往後方搬**——這是「大本營」的概念。
//
// ⚠️ `gens` 會被就地修改（`+4` 所屬省）。呼叫端要確保它與 `w.Units`
// 對應的那一份是同一個切片，否則兩邊會不一致。
func (w *AIWorld) Relocate(p ProvinceID, leader GeneralID, gens []General) RelocateResult {
	i := int(leader) - 1
	if i < 0 || i >= len(gens) {
		return RelocateResult{}
	}
	if gens[i].Province != p {
		return RelocateResult{} // 領袖不在這裡
	}
	if w.Hostile(p) == 0 {
		return RelocateResult{} // 這裡不是前線，不用搬
	}
	prov, err := w.Table.At(p)
	if err != nil {
		return RelocateResult{}
	}

	for _, n := range prov.Neighbours {
		np, err := w.Table.At(n)
		if err != nil {
			continue
		}
		if np.Commander != leader {
			continue // 不是我方的省
		}
		if np.Flags&ProvinceFlagInBattle != 0 {
			continue // 交戰中
		}
		if w.Hostile(n) != 0 {
			continue // 目標必須是後方
		}
		gens[i].Province = n
		if k := int(leader) - 1; k >= 0 && k < len(w.Units) {
			w.Units[k].Province = n
		}
		np.Commander = leader
		np.Governor = leader
		return RelocateResult{Moved: true, From: p, To: n}
	}
	return RelocateResult{}
}
