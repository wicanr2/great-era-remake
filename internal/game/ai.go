package game

// 政略階段的電腦決策，照 `sub_15F3C` 的優先序鏈
// （`docs/mechanics/70-ai.md` §6d–§6k）。
//
// **這條鏈是由上而下的優先序，不是評分挑最高的**——一旦某一步做出決定，
// 後面全部跳過。原版用 stack 上的 `ss:[di-7]` 當「已決定」旗標，
// 這裡用回傳值代替。

// AIAction 是電腦決定要做的事。
type AIAction struct {
	// Kind 是動作種類。
	Kind AIActionKind
	// From 是出發省（電腦正在處理的省）。
	From ProvinceID
	// To 是目標省。
	To ProvinceID
	// TransferKind 是調動類型（`sub_14F9A` 的第二個參數）：
	// 前線省之間 4、往前線 1、往後方 0。
	TransferKind int
	// Step 記錄是鏈上哪一步做的決定，方便對照 `70-ai.md`。
	Step string
}

// AIActionKind 是動作種類。
type AIActionKind int

const (
	// AINone 表示這一輪沒有決定要做什麼。
	AINone AIActionKind = iota
	// AITransfer 是調動（`sub_14F9A`）。決策鏈六步裡有五步做這個。
	AITransfer
	// AIAttack 是攻打（`sub_15925`）。
	AIAttack
)

// AIWorld 是決策要用到的世界狀態。
//
// 用介面收斂依賴：規則層不該知道存檔怎麼讀、將領表怎麼來。
type AIWorld struct {
	Table *ProvinceTable
	Map   *Map
	// Units 是全部將領的執行期狀態。
	Units []CombatUnit
	// Strengths 與 Units 一一對應，用來算戰力。
	Strengths []StrengthInput
	Opts      StrengthOpts
	// EnableExtra 對應 `byte_6FFCA & 4`：關掉的話決策鏈的最後三步不執行
	// （`70-ai.md` §6d）。
	EnableExtra bool
}

// 各步驟的門檻。全部出自反組譯，不是調出來的。
const (
	// AIFrontlineGeneralCap 是增援前線時目標省的將領數上限
	// （`sub_1527A`，`70-ai.md` §6f）。**這是最具體的一條行為參數。**
	AIFrontlineGeneralCap = 15
	// AISpreadMinGenerals 是後方分兵的門檻：當前省至少要有這麼多將領
	// （`sub_15667`，§6j）。
	AISpreadMinGenerals = 3
	// AISupplyMinGenerals 是往補給省調動的門檻（`sub_15A9A`，§6k）。
	AISupplyMinGenerals = 4
)

// 補給門檻（`sub_15A9A`，§6k）。資源上限是 60,000，
// 所以後三項要求半滿，黃金只要六分之一。
const (
	AISupplyGold = 5000
	AISupplyRest = 30000
)

// Hostile 回報某省有沒有敵對的鄰省，語意照 `sub_5B7DC`（§6d）：
// 跳過填充、海洋境外、**無主省**與自己人，回傳第一個敵對的鄰省。
//
// 回傳 0 表示這是後方省。
func (w *AIWorld) Hostile(p ProvinceID) ProvinceID {
	prov, err := w.Table.At(p)
	if err != nil || prov.Commander == 0 {
		return 0
	}
	for k := 0; k < 7; k++ {
		b := prov.Raw[provOffNeighbour+k]
		if b == 0 || b == SeaBorder {
			continue
		}
		np, err := w.Table.At(ProvinceID(b))
		if err != nil || np.Commander == 0 {
			continue // 無主省不算敵對
		}
		if np.Commander == prov.Commander {
			continue
		}
		return ProvinceID(b)
	}
	return 0
}

// ManpowerFlags 回報某省有沒有滿員／缺員的部隊，語意照 `sub_14761`（§6g）。
//
//	bit 0 = 有兵力剛好滿員的將領
//	bit 1 = 有兵力不到半員的將領
//
// 司令本人不算。滿員 = 20000 ÷ 兵種權重，半員是它的一半。
func (w *AIWorld) ManpowerFlags(p ProvinceID) uint8 {
	prov, err := w.Table.At(p)
	if err != nil {
		return 0
	}
	var flags uint8
	for i := range w.Units {
		u := &w.Units[i]
		if !u.Active || u.Province != p || u.General == prov.Commander {
			continue
		}
		if i >= len(w.Strengths) {
			continue
		}
		full := int(BranchFullStrength(w.Strengths[i].Branch))
		if full == 0 {
			continue
		}
		switch f := int(w.Strengths[i].Force); {
		case f == full:
			flags |= 1
		case f < full/2:
			flags |= 2
		}
	}
	return flags
}

// ProvincePower 是某省可用將領的戰力總和，語意照 `sub_5B983`（§6h）。
//
// **電腦用它評估局勢，而它加總的就是攻擊力公式** ——
// remake 的 AI 必須與戰鬥用同一份實作。
func (w *AIWorld) ProvincePower(p ProvinceID) int {
	total := 0
	for i := range w.Units {
		u := &w.Units[i]
		if !u.Active || u.Province != p || i >= len(w.Strengths) {
			continue
		}
		total += Strength(w.Strengths[i], w.Opts)
	}
	return total
}

// GeneralCount 是某省效忠其司令且可用的將領數，語意照 `sub_5A881`
// （`docs/re/07` §9）。
func (w *AIWorld) GeneralCount(p ProvinceID) int {
	prov, err := w.Table.At(p)
	if err != nil {
		return 0
	}
	n := 0
	for i := range w.Units {
		u := &w.Units[i]
		if u.Active && u.Province == p && u.Faction == prov.Commander {
			n++
		}
	}
	return n
}

// Decide 跑一次政略決策，回傳電腦決定要做的事。
//
// 順序照 `sub_15F3C`（§6d）：先看這個省是不是前線，再依鏈往下試，
// **第一個做出決定的就是答案**。
//
// ⚠️ **已知與原版的差異。** 原版每一步還有幾道守門
// （`sub_13F4E`／`sub_1401E`／`sub_36BC7`）尚未讀出，這裡沒有實作。
// 後果很具體：`generalTransfer`（一般調動）的條件只剩「鄰省有主」，
// 幾乎總是成立，於是**它會吃掉所有決策，排在它後面的攻打永遠輪不到**。
//
// 拿 `SAVE(1)` 的局面跑一遍就看得出來：29 個有主的省裡，
// 18 個走一般調動、6 個撤回後方、5 個分兵，**攻打 0 次**。
// 原版當然會攻打——差在那三道守門。
//
// 在守門解出來之前，**不要把這支當成原版行為的替代品**。
func (w *AIWorld) Decide(p ProvinceID) AIAction {
	if w.Hostile(p) != 0 {
		return w.decideFrontline(p)
	}
	return w.decideRear(p)
}

// decideFrontline 是前線省的鏈（4 步）。
func (w *AIWorld) decideFrontline(p ProvinceID) AIAction {
	if a := w.pullBack(p); a.Kind != AINone {
		return a
	}
	if a := w.generalTransfer(p, 2); a.Kind != AINone {
		return a
	}
	if a := w.attack(p); a.Kind != AINone {
		return a
	}
	if w.EnableExtra {
		if a := w.supplyTransfer(p); a.Kind != AINone {
			return a
		}
	}
	return AIAction{From: p}
}

// decideRear 是後方省的鏈（6 步）。**比前線省長**——
// 後方要做內政與分兵，前線只管打。
func (w *AIWorld) decideRear(p ProvinceID) AIAction {
	if a := w.spreadOut(p); a.Kind != AINone {
		return a
	}
	if a := w.generalTransfer(p, 1); a.Kind != AINone {
		return a
	}
	if a := w.reinforceFront(p); a.Kind != AINone {
		return a
	}
	if a := w.attack(p); a.Kind != AINone {
		return a
	}
	if w.EnableExtra {
		if a := w.supplyTransfer(p); a.Kind != AINone {
			return a
		}
	}
	return AIAction{From: p}
}

// pullBack 是 `sub_1541E`（§6h）：前線省把缺員的部隊撤到後方。
//
// 目標必須是**自己人的後方省**、未交戰；候選裡挑**戰力總和最小**的。
//
// ⚠️ 原版還有一道 `x / y >= 10000` 的門檻（推測是平均兵力），
// 那兩個 local 的來源還沒追，所以這裡沒有實作。
func (w *AIWorld) pullBack(p ProvinceID) AIAction {
	if w.ManpowerFlags(p)&2 == 0 {
		return AIAction{}
	}
	prov, err := w.Table.At(p)
	if err != nil {
		return AIAction{}
	}
	best, bestPower := ProvinceID(0), 0
	for _, n := range w.neighbours(prov) {
		np, err := w.Table.At(n)
		if err != nil || np.Commander != prov.Commander {
			continue // 只找自己人
		}
		if w.Hostile(n) != 0 || np.Flags&ProvinceFlagInBattle != 0 {
			continue // 目標必須是安全的後方
		}
		if pw := w.ProvincePower(n); best == 0 || pw < bestPower {
			best, bestPower = n, pw
		}
	}
	if best == 0 {
		return AIAction{}
	}
	return AIAction{Kind: AITransfer, From: p, To: best, TransferKind: 4, Step: "sub_1541E 撤回後方"}
}

// spreadOut 是 `sub_15667`（§6j）：後方省往將領更少的後方省分兵。
func (w *AIWorld) spreadOut(p ProvinceID) AIAction {
	if w.ManpowerFlags(p)&2 == 0 {
		return AIAction{}
	}
	here := w.GeneralCount(p)
	if here < AISpreadMinGenerals {
		return AIAction{}
	}
	prov, err := w.Table.At(p)
	if err != nil {
		return AIAction{}
	}
	best, bestPower := ProvinceID(0), 0
	for _, n := range w.neighbours(prov) {
		np, err := w.Table.At(n)
		if err != nil {
			continue
		}
		if w.Hostile(n) != 0 || np.Flags&ProvinceFlagInBattle != 0 {
			continue
		}
		if w.GeneralCount(n) >= here {
			continue // 只往人比自己少的地方分
		}
		if pw := w.ProvincePower(n); best == 0 || pw < bestPower {
			best, bestPower = n, pw
		}
	}
	if best == 0 {
		return AIAction{}
	}
	return AIAction{Kind: AITransfer, From: p, To: best, TransferKind: 4, Step: "sub_15667 分兵"}
}

// reinforceFront 是 `sub_1527A`（§6f）：後方省把滿員部隊送去自己人的前線。
//
// **目標省的將領數 ≥ 15 就不去**——電腦不會把部隊全堆在一個前線省。
func (w *AIWorld) reinforceFront(p ProvinceID) AIAction {
	if w.ManpowerFlags(p)&1 == 0 {
		return AIAction{}
	}
	prov, err := w.Table.At(p)
	if err != nil {
		return AIAction{}
	}
	for _, n := range w.neighbours(prov) {
		np, err := w.Table.At(n)
		if err != nil || np.Commander != prov.Commander {
			continue
		}
		if w.Hostile(n) == 0 || np.Flags&ProvinceFlagInBattle != 0 {
			continue // 目標必須是前線
		}
		if w.GeneralCount(n) >= AIFrontlineGeneralCap {
			continue
		}
		return AIAction{Kind: AITransfer, From: p, To: n, TransferKind: 1, Step: "sub_1527A 增援前線"}
	}
	return AIAction{}
}

// generalTransfer 是 `sub_150FC`（§6e）：一般調動。
//
// 目標是任何**有主**的鄰省；模式 2 額外跳過正在打仗的省。
// 調動類型依目標省的處境定。
func (w *AIWorld) generalTransfer(p ProvinceID, mode int) AIAction {
	prov, err := w.Table.At(p)
	if err != nil {
		return AIAction{}
	}
	for _, n := range w.neighbours(prov) {
		np, err := w.Table.At(n)
		if err != nil || np.Commander == 0 {
			continue // 無主省不調
		}
		if mode == 2 && np.Flags&ProvinceFlagInBattle != 0 {
			continue
		}
		kind := 0
		switch {
		case mode == 2:
			kind = 4
		case w.Hostile(n) != 0:
			kind = 1
		}
		return AIAction{Kind: AITransfer, From: p, To: n, TransferKind: kind, Step: "sub_150FC 一般調動"}
	}
	return AIAction{}
}

// supplyTransfer 是 `sub_15A9A`（§6k）：往補給充足的省調動。
//
// 需要 `EnableExtra`（原版的 `byte_6FFCA & 4`）。
func (w *AIWorld) supplyTransfer(p ProvinceID) AIAction {
	if w.GeneralCount(p) < AISupplyMinGenerals {
		return AIAction{}
	}
	prov, err := w.Table.At(p)
	if err != nil {
		return AIAction{}
	}
	for _, n := range w.neighbours(prov) {
		np, err := w.Table.At(n)
		if err != nil || np.Commander == 0 {
			continue
		}
		if np.Gold < AISupplyGold || np.Food < AISupplyRest ||
			np.Ammo < AISupplyRest || np.Fuel < AISupplyRest {
			continue
		}
		return AIAction{Kind: AITransfer, From: p, To: n, TransferKind: 4, Step: "sub_15A9A 補給調動"}
	}
	return AIAction{}
}

// attack 是 `sub_15925`：挑一個可攻打的鄰省。
//
// 規則見 `SPEC-01` §2：鄰省、司令不同、且非無主。
func (w *AIWorld) attack(p ProvinceID) AIAction {
	if t := w.Table.FirstAttackable(p); t != 0 {
		return AIAction{Kind: AIAttack, From: p, To: t, Step: "sub_15925 攻打"}
	}
	return AIAction{}
}

// neighbours 回傳鄰省表的前 7 格（原版只掃 7 格），濾掉填充與海洋境外。
func (w *AIWorld) neighbours(prov *Province) []ProvinceID {
	out := make([]ProvinceID, 0, 7)
	for k := 0; k < 7; k++ {
		b := prov.Raw[provOffNeighbour+k]
		if b == 0 || b == SeaBorder {
			continue
		}
		out = append(out, ProvinceID(b))
	}
	return out
}
