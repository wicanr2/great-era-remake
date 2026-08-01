package game

// 調動的執行層：`sub_1408F`（`docs/re/10-troop-transfer.md`）。
//
// `internal/game/ai.go` 的 `Decide` 只決定「往哪一省調」，
// 真正搬動兵力的是這裡。兩者在原版是分開的兩支
// （`sub_15F3C` 決策鏈 → `sub_14F9A` → `sub_1408F`），這裡保持一樣的切分。
//
// **調動 = 把將領記錄的 `+4`（所屬省）改寫成目標省。** 兵力不會被拆分，
// 是整個將領連人帶兵過去。原版只有這一種寫法（`[將領×33 + 0x7A81] = 目標省`），
// 沒有「分一半兵力」那種操作。

// TransferMode 是 `sub_1408F` 的 `arg_6`：調動要挑哪些將領。
//
// 八個值在原版是四段互斥的 `cmp al, N`，語意差在**挑選條件**，
// 搬動的動作完全一樣。
type TransferMode int

const (
	// TransferAll 是模式 0：清單裡除了最後一個以外全部調走，不看兵力。
	//
	// ⚠️ 迴圈跑 `1..count-1`，最後一個留下。這是原版寫死的
	// （`docs/re/10` §3a），不是我們的選擇。
	TransferAll TransferMode = 0

	// TransferFullOnly 是模式 1：只調**滿員**的將領，且目標省兵力
	// 到 40 萬就停手。
	TransferFullOnly TransferMode = 1

	// TransferUnderHalfCapped 是模式 2：從清單尾端往前挑**未達半滿**的，
	// 最多調 5 個。
	TransferUnderHalfCapped TransferMode = 2

	// TransferUnderHalfTwo 是模式 3：同上但只看清單最後兩個。
	TransferUnderHalfTwo TransferMode = 3

	// TransferQuarterHalfRoster 是模式 4：清單後半段裡**低於四分之一滿**的。
	TransferQuarterHalfRoster TransferMode = 4

	// TransferFullOnlyAll 是模式 5：與模式 1 同樣只調滿員的，
	// 但掃完整份清單、也不看目標省兵力。
	TransferFullOnlyAll TransferMode = 5

	// TransferFirst 是模式 6：只調清單第一個。
	TransferFirst TransferMode = 6

	// TransferFullOnlyRoster 是模式 7：只調滿員的，掃到 `count-1`。
	TransferFullOnlyRoster TransferMode = 7
)

// 調動的三道數值門檻，全部出自 `sub_1408F`。
const (
	// TransferTargetGeneralCap 是目標省將領數的硬上限。原版每調一個就
	// 把計數加一，碰到就停（`cmp [var_6], 64h`）。
	TransferTargetGeneralCap = 100

	// TransferModeOneForceCap 是模式 1 的煞車：目標省兵力總和到這個數
	// 就不再送人過去（`cmp dx, 6 / cmp ax, 1A80h` = 0x61A80）。
	TransferModeOneForceCap = 400000

	// TransferAlertBefore / TransferAlertAfter 是「大軍集結」提示的兩道門檻
	// （`docs/re/10` §5）。調動前低於前者、調動後高於後者才提示。
	TransferAlertBefore = 250000
	TransferAlertAfter  = 300000
)

// 兵力門檻是滿員數的倍率，全部寫成「×兵種人力權重後與常數比」，
// 與原版一致——原版比的是 `兵力 ÷ (1/權重)`，也就是 `兵力 × 權重`。
//
//	模式 1/5/7：== 20000  → 滿員
//	模式 2/3：  <  10000  → 未達半滿
//	模式 4：    <   5000  → 未達四分之一滿
const (
	transferFull    = 20000
	transferHalf    = 10000
	transferQuarter = 5000
)

// Roster 是一個省的將領清單，對應原版 `狀態[-22Eh]` 那個
// Pascal 風格陣列：`[0]` 是元素數，`[1..n]` 是將領。
//
// ⚠️ 這裡用 slice，`Len()` 對應 `[0]`。原版所有迴圈的索引都從 1 起跳，
// 為了讓門檻（`count-1`、`count/2`）讀起來與反組譯一致，
// `Get(i)` 也採 1-based。
type Roster struct {
	ids []int // Units 的索引
}

// Len 是清單長度，也就是原版的 `狀態[-22Eh]`。
func (r Roster) Len() int { return len(r.ids) }

// Get 取第 i 個（1-based）。超出範圍回 -1。
func (r Roster) Get(i int) int {
	if i < 1 || i > len(r.ids) {
		return -1
	}
	return r.ids[i-1]
}

// RosterOf 建某省的將領清單，語意照 `sub_13B44`（`docs/re/13` §1）。
//
// 篩選三條與 `sub_5A881`／`sub_306CF` 一致：所屬省相符、可用、
// 且效忠該省司令。
//
// **順序是攻擊力（`sub_5A0B9`）由高到低。** 這一條先前標「未驗」，
// 現在 confirmed——`sub_13B44` 建完清單就跑一輪降序排序。
//
// 順序決定模式 0/2/3/4/6 挑到誰，語意因此變得很清楚：
//
//	模式 0  調走 1..count-1     → **最弱的那個留守**
//	模式 6  只調第一個          → 派出**最強的**
//	模式 2/3 從尾端往前         → 從**最弱的**開始挑
//	模式 4  後半段              → **較弱的那一半**
//
// 強的先派出去、弱的留下——與電腦「分散不集中」的整體傾向一致。
func (w *AIWorld) RosterOf(p ProvinceID) Roster {
	prov, err := w.Table.At(p)
	if err != nil {
		return Roster{}
	}
	var r Roster
	for i := range w.Units {
		u := &w.Units[i]
		// ⚠️ 原版比的是 `+16 == 1`，不是 `+16 & 1`。目前觀察到的
		// `+16` 只有 {0, 1, 32} 三個值，兩種寫法結果相同
		// （32 & 1 == 0），但若出現別的值就會分歧。
		if u.Active && u.Province == p && u.Faction == prov.Commander {
			r.ids = append(r.ids, i)
		}
	}
	// 依攻擊力降序。原版是「外層 i、內層 j=i+1、比較後交換」，
	// **不是穩定排序**——同分時的順序取決於交換過程，所以照抄寫法。
	for i := 0; i < len(r.ids); i++ {
		for j := i + 1; j < len(r.ids); j++ {
			if w.strengthAt(r.ids[j]) > w.strengthAt(r.ids[i]) {
				r.ids[i], r.ids[j] = r.ids[j], r.ids[i]
			}
		}
	}
	return r
}

// strengthAt 是 `sub_5A0B9`（攻擊力公式）套在某個將領上。
func (w *AIWorld) strengthAt(i int) int {
	if i < 0 || i >= len(w.Strengths) {
		return 0
	}
	return Strength(w.Strengths[i], w.Opts)
}

// CommandsFor 是一個省這回合能下幾個命令，出自 `sub_13D23`：
//
//	mov     ax, ss:[di-22Eh]     ; 將領清單長度
//	cwd
//	mov     cx, 8
//	idiv    cx
//	inc     ax                   ; 將領數 ÷ 8 + 1
//	mov     ss:[di-232h], ax
//
// **這就是社群說的「每個省份有命令數上限」**（`CLAUDE.md` §1.5），
// 而且是算出來的：將領越多，能下的命令越多。
//
// `sub_1ACCC` 的主迴圈跑到這個數字歸零才換下一個省，
// `sub_1398D`（遷都）與 `sub_174C9` 各消耗一次（`docs/re/13` §2）。
func (w *AIWorld) CommandsFor(p ProvinceID) int {
	return w.RosterOf(p).Len()/8 + 1
}

// TransferReport 是一次調動的結果。
type TransferReport struct {
	// Moved 是實際被搬走的將領（`Units` 的索引）。
	Moved []int
	// Alert 為 true 表示觸發了「大軍集結」提示（`docs/re/10` §5）。
	Alert bool
}

// forceRatio 回傳 `兵力 × 兵種人力權重`，也就是原版
// `Round(兵力 ÷ sub_349C1(將領))` 算出來的東西。
//
// 滿員時它正好等於 20000，與兵種無關——`sub_349C1` 回的是
// 1/人力權重，而滿員數 = 20000 ÷ 權重（`docs/re/08` §4）。
func (w *AIWorld) forceRatio(idx int) int {
	if idx < 0 || idx >= len(w.Strengths) {
		return 0
	}
	s := &w.Strengths[idx]
	wt, ok := routWeight[s.Branch]
	if !ok {
		return 0
	}
	return int(s.Force) * wt
}

// ForceTotal 是某省效忠其司令、可用的將領兵力總和，語意照 `sub_306CF`。
//
// 它同時是 `sub_1408F` 兩道煞車與「大軍集結」提示的判斷依據。
func (w *AIWorld) ForceTotal(p ProvinceID) int {
	total := 0
	for _, i := range w.RosterOf(p).ids {
		if i < len(w.Strengths) {
			total += int(w.Strengths[i].Force)
		}
	}
	return total
}

// ApplyTransfer 執行一次調動，回傳搬了誰。
//
// 對應 `sub_1408F(狀態, &旗標, 模式, 目標省)`。**會改寫 `w.Units`**
// ——被調走的將領 `Province` 變成 `to`。
//
// 這裡不做 `sub_14F9A` 的收尾（把目標省的司令改成我方勢力、重畫地圖），
// 那是上一層的事，見 `ApplyAction`。
func (w *AIWorld) ApplyTransfer(from ProvinceID, mode TransferMode, to ProvinceID) TransferReport {
	var rep TransferReport
	r := w.RosterOf(from)
	count := r.Len()
	if count == 0 {
		return rep
	}
	before := w.ForceTotal(to)

	move := func(i int) {
		idx := r.Get(i)
		if idx < 0 {
			return
		}
		w.Units[idx].Province = to
		rep.Moved = append(rep.Moved, idx)
	}

	switch mode {
	case TransferAll:
		// 模式 0 與 6 是唯二無條件的：不看兵力，直接搬。
		if count == 1 {
			move(1)
			break
		}
		for i := 1; i <= count-1; i++ {
			move(i)
		}

	case TransferFirst:
		move(1)

	case TransferFullOnly, TransferFullOnlyAll, TransferFullOnlyRoster:
		w.transferFullOnly(r, mode, to, move)

	case TransferUnderHalfCapped, TransferUnderHalfTwo:
		// 模式 2 最多調 5 個或 count-1 個（取小）；模式 3 只看最後兩個。
		limit := count - 1
		low := 2
		if mode == TransferUnderHalfCapped {
			if limit > 5 {
				limit = 5
			}
			if count < 2 {
				break
			}
		} else {
			if count <= 2 {
				break
			}
			low = count - 1
			limit = count // 模式 3 不設額度
		}
		for i := count; i >= low; i-- {
			if limit <= 0 {
				break
			}
			idx := r.Get(i)
			// 司令本人不調動（原版比的是 `arr[i] == 狀態[-2]`）。
			if idx < 0 || w.isCommander(from, idx) {
				continue
			}
			if w.forceRatio(idx) < transferHalf {
				move(i)
				limit--
			}
		}

	case TransferQuarterHalfRoster:
		if count == 1 {
			move(1)
			break
		}
		// 原版用 48-bit Real 除以 2 再 Round，也就是四捨五入到偶數半途；
		// count 是整數，Round(count/2) 對 .5 取偶（Turbo Pascal 的 Round）。
		low := roundHalf(count)
		for i := count; i >= low; i-- {
			idx := r.Get(i)
			if idx < 0 {
				continue
			}
			if w.forceRatio(idx) < transferQuarter {
				move(i)
			}
		}
	}

	// 「大軍集結」提示：調動前不到 25 萬、調動後超過 30 萬。
	// 兩道門檻不重疊，所以**只有這次調動本身讓兵力跨過去**才會觸發。
	if before < TransferAlertBefore && w.ForceTotal(to) >= TransferAlertAfter {
		rep.Alert = true
	}
	return rep
}

// transferFullOnly 是模式 1/5/7 共用的那條迴圈：只挑滿員的將領。
//
// 三個模式差在**掃到哪裡**與**要不要看目標省兵力**：
//
//	1  掃到 count-1，目標省兵力到 40 萬就停
//	5  掃完整份清單
//	7  掃到 count-1
func (w *AIWorld) transferFullOnly(r Roster, mode TransferMode,
	to ProvinceID, move func(int)) {
	count := r.Len()
	// 進入條件：模式 5 隨時可進；模式 1/7 要求清單長度 > 1。
	if mode != TransferFullOnlyAll && count <= 1 {
		return
	}
	// 目標省將領數的硬上限。
	n := w.GeneralCount(to)
	if n >= TransferTargetGeneralCap {
		return
	}
	for i := 1; i <= count; i++ {
		idx := r.Get(i)
		if idx >= 0 && w.forceRatio(idx) == transferFull {
			move(i)
			n++
		}
		switch mode {
		case TransferFullOnly:
			if i == count-1 {
				return
			}
			if w.ForceTotal(to) >= TransferModeOneForceCap {
				return
			}
		case TransferFullOnlyAll:
			if i == count {
				return
			}
		case TransferFullOnlyRoster:
			if i == count-1 {
				return
			}
		}
		if n >= TransferTargetGeneralCap {
			return
		}
	}
}

// isCommander 判斷某個將領是不是該省所屬勢力的領袖。
//
// 原版比的是 `清單[i] == 狀態[-2]`，而 `狀態[-2]` 在 `sub_14F9A`
// 被寫進目標省的司令欄位（`docs/re/10` §4），所以它就是**我方勢力領袖**。
func (w *AIWorld) isCommander(p ProvinceID, idx int) bool {
	prov, err := w.Table.At(p)
	if err != nil || idx < 0 || idx >= len(w.Units) {
		return false
	}
	return w.Units[idx].General == prov.Commander
}

// roundHalf 是 `Round(n / 2)`，照 Turbo Pascal 的「.5 取偶」。
//
// n 是整數，所以只有奇數會落在 .5 上：3/2 = 1.5 → 2、5/2 = 2.5 → 2、
// 7/2 = 3.5 → 4。**這不是 (n+1)/2**，差別在 5 與 9 這種數字上。
func roundHalf(n int) int {
	q, rem := n/2, n%2
	if rem == 0 {
		return q
	}
	// 落在 .5：取最接近的偶數。
	if q%2 == 0 {
		return q
	}
	return q + 1
}

// ApplyAction 把 `Decide` 的結果實際執行，對應 `sub_14F9A`。
//
// 調動成功後原版做三件事：
//
//  1. 目標省的司令欄位改成我方勢力（`省份[目標].+20 = 狀態[-2]`）
//     ——**這就是佔領無主省的機制**
//  2. 立起「已決定」旗標，讓決策鏈後面的步驟全部跳過
//  3. 重畫出發省與目標省
//
// 這裡做 1 與 2；重畫是呈現層的事。
func (w *AIWorld) ApplyAction(a AIAction) TransferReport {
	if a.Kind != AITransfer || a.To == 0 {
		return TransferReport{}
	}
	rep := w.ApplyTransfer(a.From, TransferMode(a.TransferKind), a.To)
	if len(rep.Moved) == 0 {
		return rep
	}
	from, err := w.Table.At(a.From)
	if err != nil {
		return rep
	}
	to, err := w.Table.At(a.To)
	if err != nil {
		return rep
	}
	// `At` 回的是表裡那筆記錄的指標，改它就是改狀態。
	to.Commander = from.Commander
	return rep
}
