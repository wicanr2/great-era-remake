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
	// AIAttack 是出兵。
	//
	// ⛔ 這裡原本寫「政略決策鏈裡沒有任何一步會產生它」。**那是錯的**——
	// 決策鏈 **B**（`Decide`，`sub_15F3C`）確實六步全是調動，
	// 但決策鏈 **A**（`ChainA`，`sub_17ADA`）的步驟 1／4／5／6 都會出兵
	// （`docs/re/28`、`docs/re/29`，實機已驗 `docs/playtest/14`）。
	AIAttack

	// AIComfort 是慰勞軍民（`sub_3412B`），決策鏈 A 的步驟 0。
	AIComfort
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
	// AIRichestFrontGenerals / AIRichestRearGenerals 是 `sub_15925` 的前置
	// 門檻（`docs/re/12`）。**25 幾乎不可能達到**，所以前線省這一步等於關閉。
	AIRichestFrontGenerals = 25
	AIRichestRearGenerals  = 5
	// AITransferableFrontCap / AITransferableRearCap 是 `sub_1588C` 收目標時
	// 對「自己的省」的將領數上限。
	AITransferableFrontCap = 15
	AITransferableRearCap  = 2
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
// 三道守門（`sub_13F4E`／`sub_1401E`／`sub_36BC7`）已經實作，
// 見各自的方法。**含一個疑似原版 bug，照抄並標記**（`generalsLoyalTo`）。
func (w *AIWorld) Decide(p ProvinceID) AIAction {
	if w.Hostile(p) != 0 {
		return w.decideFrontline(p)
	}
	return w.decideRear(p)
}

// chain 回傳某省該走的決策鏈，順序就是原版的優先序。
//
// 前線 4 步、後方 6 步。**後方比前線長**——後方要做內政與分兵，
// 前線只管打。
func (w *AIWorld) chain(p ProvinceID) []func(ProvinceID) AIAction {
	var steps []func(ProvinceID) AIAction
	if w.Hostile(p) != 0 {
		steps = []func(ProvinceID) AIAction{
			w.pullBack,
			func(q ProvinceID) AIAction { return w.generalTransfer(q, 2) },
			func(q ProvinceID) AIAction { return w.richestTransfer(q, true) },
		}
	} else {
		steps = []func(ProvinceID) AIAction{
			w.rescueFront,
			w.spreadOut,
			func(q ProvinceID) AIAction { return w.generalTransfer(q, 1) },
			w.reinforceFront,
			func(q ProvinceID) AIAction { return w.richestTransfer(q, false) },
		}
	}
	if w.EnableExtra {
		steps = append(steps, w.supplyTransfer)
	}
	return steps
}

// decideFrontline 是前線省的鏈（4 步）。
func (w *AIWorld) decideFrontline(p ProvinceID) AIAction {
	return firstAction(w.chain(p), p)
}

// decideRear 是後方省的鏈（6 步）。
func (w *AIWorld) decideRear(p ProvinceID) AIAction {
	return firstAction(w.chain(p), p)
}

func firstAction(steps []func(ProvinceID) AIAction, p ProvinceID) AIAction {
	for _, f := range steps {
		if a := f(p); a.Kind != AINone {
			return a
		}
	}
	return AIAction{From: p}
}

// Step 走一次決策鏈並**實際執行**，回傳做成的那一件事。
//
// 與 `Decide` 的差別是這裡把執行結果回饋進鏈裡，語意照
// `sub_14F9A`（`docs/re/10` §4）：
//
//	call    sub_1408F
//	cmp     [bp+var_1], 0
//	jz      短路              ; ← 沒搬成就不立「已決定」旗標
//	...
//	mov     byte ptr ss:[di-7], 1
//
// **決策層與執行層的條件不一樣**：`sub_1541E` 只看「省內有缺員的部隊」，
// 而 `sub_1408F` 模式 4 要求「低於四分之一滿」。缺員但沒那麼缺的時候
// 決策說要調、執行挑不到人——這時旗標沒立，鏈**繼續往下走**。
//
// 少了這個回饋，電腦會有大量「決定了但什麼也沒發生」的空轉回合。
func (w *AIWorld) Step(p ProvinceID) (AIAction, TransferReport) {
	for _, f := range w.chain(p) {
		a := f(p)
		switch a.Kind {
		case AINone:
			continue
		case AIAttack:
			// 攻打不走 `sub_1408F`，直接算決定完成。
			return a, TransferReport{}
		}
		if rep := w.ApplyAction(a); len(rep.Moved) > 0 {
			return a, rep
		}
		// 一個人都沒搬動 → 旗標沒立 → 試下一步。
	}
	return AIAction{From: p}, TransferReport{}
}

// rescueFront 是 `sub_15018`（`docs/re/11` §1）：**後方省的第一優先**。
//
// 把滿員的部隊送去「我方的、前線的、戰力掉到 2,000 以下、沒在交戰」的省。
// 與其他五支調動最大的差別是**它掃全部 39 省**，不是只看鄰省——
// 所以目標可能很遠，要靠 `NextHop` 一步一步接力過去。
//
// 調動模式在原版是 1／5／7 三選一，由 `byte_6FFCA` bit 0 與一張
// 24 格表決定（§6）。那張表還沒解，所以這裡固定用 1，並標明差異。
func (w *AIWorld) rescueFront(p ProvinceID) AIAction {
	// sub_14761 bit 0：省內要有滿員的部隊。
	if w.ManpowerFlags(p)&1 == 0 {
		return AIAction{}
	}
	prov, err := w.Table.At(p)
	if err != nil || prov.Commander == 0 {
		return AIAction{}
	}
	for _, target := range w.RescueTargets(prov.Commander) {
		if target == p {
			continue // 自己不救自己
		}
		hop := w.NextHop(p, target, prov.Commander)
		if hop == 0 {
			continue
		}
		return AIAction{Kind: AITransfer, From: p, To: hop,
			TransferKind: int(TransferFullOnly), Step: "sub_15018 救援前線"}
	}
	return AIAction{}
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
//
// 六道篩選全部照原版，順序也一樣：
//
//	鄰省 != 0 且 != 0FFh
//	sub_5B7DC(鄰省) == 0        目標必須是後方
//	省份[鄰省].+32 & 40h == 0   沒在交戰
//	sub_13F4E(當前省, 鄰省)     那道守門（含照抄的 bug）
//	sub_5A881(鄰省) < sub_5A881(當前省)   目標的將領比自己少
//
// 候選依 `sub_5B983`（戰力總和）**升序**排序，挑最小的那個。
// 調動用**模式 3**（`mov al, 3`）——只看清單最後兩個、且未達半滿，
// 是八個模式裡最保守的一個。
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
		if !w.generalsLoyalTo(n, prov.Commander) {
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
	return AIAction{Kind: AITransfer, From: p, To: best,
		TransferKind: int(TransferUnderHalfTwo), Step: "sub_15667 分兵"}
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

// generalTransfer 是 `sub_150FC`（§6e）：一般調動 = **往無主省擴張**。
//
//	cmp     word ptr [di-6221h], 0   ; 省份[鄰省].+20（司令）
//	jnz     short loc_15219          ; ← 不等於 0 就跳過
//
// `jnz` 跳的是 skip，所以留下來的只有**司令 == 0** 的省。
// 配上 `sub_14F9A` 調動成功後把目標省的司令改成我方（`docs/re/10` §4），
// 整件事就是：**把部隊調進無人的省份，那個省就歸我了**。
//
// 接受條件是兩選一：
//
//	後方鏈（mode 1）：sub_1401E(鄰省, 當前省) 為真——目標比自己窮
//	前線鏈（mode 2）：sub_36BC7(鄰省) 為真——通向那五個特殊省
//
// 兩條路徑最後都要過「沒在交戰」。候選**不排序**，原版用覆蓋式賦值
// （`ss:[di-3]` 每次符合就蓋掉），所以挑到的是**最後一個**符合的鄰省。
//
// 調動模式：前線鏈用 4；後方鏈看目標省有沒有敵對鄰省，有用 1、沒有用 0。
func (w *AIWorld) generalTransfer(p ProvinceID, mode int) AIAction {
	prov, err := w.Table.At(p)
	if err != nil {
		return AIAction{}
	}
	best := ProvinceID(0)
	for _, n := range w.neighbours(prov) {
		np, err := w.Table.At(n)
		if err != nil || np.Commander != 0 {
			continue // 只往**無主**省調
		}
		// sub_13F4E：那道守門（含照抄的 bug）。
		if !w.generalsLoyalTo(n, prov.Commander) {
			continue
		}
		// 後方鏈走「比我窮」，前線鏈走「通向五省」。
		if !(mode == 1 && w.poorerThan(n, p)) {
			if !w.nearSpecial(n) || mode != 2 {
				continue
			}
		}
		if np.Flags&ProvinceFlagInBattle != 0 {
			continue
		}
		best = n // 覆蓋式：留最後一個
	}
	if best == 0 {
		return AIAction{}
	}
	kind := 0
	switch {
	case mode == 2:
		kind = int(TransferQuarterHalfRoster)
	case w.Hostile(best) != 0:
		kind = int(TransferFullOnly)
	}
	return AIAction{Kind: AITransfer, From: p, To: best, TransferKind: kind,
		Step: "sub_150FC 一般調動"}
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

// richestTransfer 是 `sub_15925`（`docs/re/12`）。
//
// ⛔ **這一支原本被當成「攻打」，是誤判。** 它最後呼叫的是
// `sub_14F9A(狀態, 2, 目標)`——調動，而且目標篩選 `sub_1588C`
// 只放行「無主省」與「自己的省」，一個敵省都不收。
//
// 一道很硬的前置門檻：
//
//	前線省（arg_2 = 1）：當前省將領數 >= 25（19h）
//	後方省（arg_2 = 0）：當前省將領數 >= 5
//
// 25 這個數字幾乎不會達到（`sub_1527A` 的增援上限才 15），
// 所以**前線省這一步實際上等於關閉**。
//
// 目標依省份記錄 `+0`（黃金）**降冪**排，挑**最富**的那個
// ——把部隊送去自己最有錢的省。
//
// ⛔ 2026-08-01 訂正：這裡原本寫「升序、挑最窮」。`sub_14EBA` 的交換條件是
//
//	cx = 黃金(候選[i])      ; i 是外圈
//	ax = 黃金(候選[j])      ; j 是內圈
//	cmp ax, cx
//	jbe 不換                ; 候選[j] ≤ 候選[i] 才跳過 → **大的往前換**
//
// 舊筆記把 `jbe` 讀成「大的不換」，方向整個反了。組語只有這一個指令決定
// 排序方向，重讀即可確認，沒有其他證據支持舊說法。
//
// 訂正後與 `sub_15A9A`「往補給充足的省調動」（§6k）方向一致——
// 兩支都是**往養得起兵的地方集中**，而不是把兵送去窮省。
func (w *AIWorld) richestTransfer(p ProvinceID, frontline bool) AIAction {
	need := AIRichestRearGenerals
	if frontline {
		need = AIRichestFrontGenerals
	}
	if w.GeneralCount(p) < need {
		return AIAction{}
	}
	prov, err := w.Table.At(p)
	if err != nil {
		return AIAction{}
	}
	best, bestGold := ProvinceID(0), 0
	for _, n := range w.neighbours(prov) {
		np, err := w.Table.At(n)
		if err != nil || !w.transferable(n, prov.Commander) {
			continue
		}
		if np.Flags&ProvinceFlagInBattle != 0 {
			continue
		}
		if best == 0 || int(np.Gold) > bestGold {
			best, bestGold = n, int(np.Gold)
		}
	}
	if best == 0 {
		return AIAction{}
	}
	return AIAction{Kind: AITransfer, From: p, To: best,
		TransferKind: int(TransferUnderHalfCapped), Step: "sub_15925 往富省集中"}
}

// transferable 是 `sub_1588C`：這個鄰省可不可以當目標。
//
//	無主省                     → 可以
//	別人的省                   → 不行（**一個敵省都不收**）
//	自己的前線省，將領數 <= 15 → 可以
//	自己的後方省，將領數 <= 2  → 可以
func (w *AIWorld) transferable(n ProvinceID, faction GeneralID) bool {
	np, err := w.Table.At(n)
	if err != nil {
		return false
	}
	if np.Commander == 0 {
		return true
	}
	if np.Commander != faction {
		return false
	}
	count := w.GeneralCount(n)
	if w.Hostile(n) != 0 {
		return count <= AITransferableFrontCap
	}
	return count <= AITransferableRearCap
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

// SpecialProvinces 是 `sub_36BC7` 硬編碼在程式裡的五個省。
//
// ⚠️ **這五個省與 `WARPOS.DAT` 裡帶哨兵 40 的五個省完全相同**
// （`70-ai.md` §6m）。一份是 1992 年的地圖資料、一份是程式常數，
// 兩處毫不相干卻一模一樣——那五個省有共同的特殊性質，**語意未定**。
var SpecialProvinces = [5]ProvinceID{7, 11, 20, 21, 36}

// nearSpecial 是 `sub_36BC7`（`70-ai.md` §6m）：目標省要與那五個特殊省
// 之一有關聯——是它本身、或與它相鄰且同勢力。
//
// ⚠️ 原版還有第三條分支：兩者不相鄰時做**省份層的路徑搜尋**
// （`sub_5BF74`，配置 39×39 的矩陣），有路徑就放行。那支還沒讀，
// 所以這裡比原版嚴格——會擋掉一些原版放行的遠距離情況。
func (w *AIWorld) nearSpecial(target ProvinceID) bool {
	tp, err := w.Table.At(target)
	if err != nil {
		return false
	}
	for _, sp := range SpecialProvinces {
		if sp == target {
			return true
		}
		// sub_5B76E：兩者相鄰嗎。
		if !w.adjacentProvinces(sp, target) {
			continue // 原版在這裡改走路徑搜尋，我們還沒實作
		}
		if p, err := w.Table.At(sp); err == nil && p.Commander == tp.Commander {
			return true
		}
	}
	return false
}

// adjacentProvinces 是 `sub_5B76E`：a 在不在 b 的鄰省表裡。
func (w *AIWorld) adjacentProvinces(a, b ProvinceID) bool {
	if !a.Valid() || !b.Valid() {
		return false
	}
	bp, err := w.Table.At(b)
	if err != nil {
		return false
	}
	for _, n := range w.neighbours(bp) {
		if n == a {
			return true
		}
	}
	return false
}

// poorerThan 是 `sub_1401E`（`70-ai.md` §6m）：
// 目標省的**黃金與糧食都比當前省少**才回 true。
func (w *AIWorld) poorerThan(target, here ProvinceID) bool {
	t, err1 := w.Table.At(target)
	h, err2 := w.Table.At(here)
	if err1 != nil || err2 != nil {
		return false
	}
	return t.Gold < h.Gold && t.Food < h.Food
}

// generalsLoyalTo 是 `sub_13F4E` 的核心判斷。
//
// # ⚠️ 這裡照抄了一個疑似原版 bug
//
// 原版拿**鄰省編號**去索引將領表（`mul 21h` = 33，將領記錄大小），
// 讀第 (鄰省編號) 個將領的效忠勢力。正確的寫法應該是拿省編號索引
// 省份表（`mul 25h` = 37）讀該省的司令——同樣做這件事的 `sub_5B7DC`
// 就是那樣寫的。
//
// 兩個常數同時不同，而型別剛好相容（兩邊都是勢力領袖 ID 的 u16），
// 所以不會崩，只是讀到不相干的將領。詳見 `70-ai.md` §6m。
//
// **照抄是刻意的**：`CLAUDE.md` §1 的定位是還原，「不得默默改動遊戲規則」
// ——bug 也是 1992 年玩家玩到的行為。要修是 remake 差異，由使用者決定。
func (w *AIWorld) generalsLoyalTo(neighbour ProvinceID, faction GeneralID) bool {
	i := int(neighbour) - 1 // 將領 ID 是 1-based
	if i < 0 || i >= len(w.Units) {
		return false
	}
	return w.Units[i].Faction == faction
}
