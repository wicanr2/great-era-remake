package game

// 出兵時**派誰、派多少**（`sub_1629E` 挑人 + `sub_163BE` 判斷夠了沒）。
//
// 這兩支是 `sub_1655C` 湊兵迴圈的內容（`docs/re/30` §4）：
//
//	loc_165F6:
//	    idx++
//	    sub_1629E(父)          ← 試著把第 idx 個將領收進名單
//	    if !sub_163BE(父): 回 loc_165F6
//
// 高難度才走這條路；低難度走 `sub_16A9E`（未讀）。

// 挑人的兩道篩選（`sub_1629E`）。
const (
	// AISortieMinStrength 是**單兵戰力下限**：`sub_5A0B9(將領) < 10` 就不收。
	//
	// 攻擊力公式算出來低於 10 的部隊，電腦不帶出門——
	// 殘兵留在省裡，不拖累出征部隊。與 `AIComfortForceThreshold`
	// 的「不慰勞殘兵」（`aicomfort.go`）是同一個傾向。
	AISortieMinStrength = 10

	// AISortieBranch4Excluded 記的是：**兵種 4 一個都不會被收進出兵名單**。
	//
	//	cmp     byte ptr [di+7A92h], 4     ; 將領 +21 = 兵種
	//	jnz     loc_162D4
	//	inc     byte ptr ss:[di-0Bh]       ; ← **計數先加**
	//	loc_162D4:
	//	cmp     byte ptr [di+7A92h], 4
	//	jnz     loc_16306
	//	cmp     byte ptr ss:[di-0Bh], 1
	//	jb      loc_16306                  ; < 1 才繼續收
	//	jmp     loc_163BA                  ; ≥ 1 → 不收
	//
	// ⚠️ **疑似原版 bug**。`inc` 排在比較之前，所以碰到第一個兵種 4 時
	// 計數已經是 1，`1 < 1` 不成立 → 當場就被攔下。看起來本意是
	// 「最多帶一個」，實際效果是「**一個都不帶**」。
	//
	// 兵種 4 的人力權重是 4（`strength.go`），滿員 5,000。
	// 照抄這個行為——`CLAUDE.md` §9 要求原版行為就是規格，
	// 不在 remake 補上「本意」。
	AISortieBranch4Excluded = 4

	// AISortieMaxGenerals 是名單的硬上限：收到 10 個就停（`sub_163BE`）。
	//
	// 出兵名單的陣列本身有 100 格（`sub_161EB` 清 100 個 word），
	// 但這道檢查讓它永遠用不到第 11 格。
	AISortieMaxGenerals = 10

	// AISortieTailMargin 是尾端那道追加判斷的邊界（`sub_163BE` 末段）：
	//
	//	if sub_13710(當前省) 且 已收 + 5 ≥ 清單長度 且 byte_6FE88 != 1: 停
	//
	// ⚠️ `sub_13710` 與 `byte_6FE88` 都未解，所以這一條**沒有實作**，
	// 只把常數記在這裡。
	AISortieTailMargin = 5
)

// AISortieStopRatioNum / AISortieStopRatioDen 是模式 2／3 的**停止倍率 2.5**。
//
// 原版用 Turbo Pascal 的 48-bit Real 常數（`CX=82h SI=0 DI=2000h`）：
//
//	exp 82h → 2^(82h-81h) = 2
//	20h → 補隱含位 A0h → A0h/80h = 1.25   → 1.25 × 2 = 2.5
//
// 條件是 `Round(目標省戰力 × 2.5) ≤ 已派出的戰力` → 停止湊兵。
//
// 用整數的分子分母表示，避免浮點誤差；原版是 `Round` 不是 `Trunc`
// （與學潮的 0.8 同一個處理，`covert.go`）。
const (
	AISortieStopRatioNum = 5
	AISortieStopRatioDen = 2
)

// AISortieStopThreshold 回傳「湊到多少就收手」的門檻，
// 即 `Round(目標省戰力 × 2.5)`。
func AISortieStopThreshold(targetStrength int) int {
	// Round(x × 5/2)：×5 之後除以 2 四捨五入。
	n := targetStrength * AISortieStopRatioNum
	q, r := n/AISortieStopRatioDen, n%AISortieStopRatioDen
	if r*2 >= AISortieStopRatioDen {
		q++
	}
	return q
}

// SortiePlan 是湊兵的結果。
type SortiePlan struct {
	// Generals 是出兵名單，順序就是收進來的順序。
	Generals []GeneralID
	// Strength 是名單的攻擊力總和（原版的 `[-10h:-0Eh]`）。
	Strength int
	// Remaining 是**留守**的戰力（原版的 `[-18h:-16h]`）：
	// 本省戰力減去已派出的。
	Remaining int
}

// PlanSortie 跑一次湊兵迴圈，回傳要派出去的名單（`sub_1629E` + `sub_163BE`）。
//
//	roster        當前省的將領清單，順序照原版 `ss:[X-22Eh+i×2]`（i 從 1 起）
//	homeStrength  本省戰力總和（`sub_5B983(當前省)`）
//	targetStr     目標省戰力總和（`sub_5B983(目標省)`）
//
// 停止條件（任一成立就停，`sub_163BE`）：
//
//	已收人數 == 清單長度 − 1              ← **至少留一個守家**
//	已收人數 == 10                        ← 硬上限
//	掃完整份清單
//	Round(目標省戰力 × 2.5) ≤ 已派出戰力  ← 湊夠了
//
// ⚠️ 「留一個守家」是 `sub_163BE` 開頭那段算出來的：`ss:[X+0Eh]` 恆為 10
// （`sub_1655C` 的呼叫端寫死），第一個分支恆成立 → 上限 = 清單長度 − 1。
// 另外三個分支要 `sub_18509` 那個呼叫端傳別的值才走得到，**未驗**。
func PlanSortie(roster []GeneralID, homeStrength, targetStr int,
	in []StrengthInput, opt StrengthOpts, branchOf func(GeneralID) uint8) SortiePlan {
	plan := SortiePlan{Remaining: homeStrength}
	if len(roster) == 0 {
		return plan
	}
	// 不用 `cap` 當變數名——那會遮蔽 Go 的內建函式。
	limit := len(roster) - 1 // 至少留一個守家
	if limit > AISortieMaxGenerals {
		limit = AISortieMaxGenerals
	}
	stop := AISortieStopThreshold(targetStr)

	for _, g := range roster {
		if len(plan.Generals) >= limit {
			break
		}
		if branchOf(g) == AISortieBranch4Excluded {
			// 見 `AISortieBranch4Excluded`：原版的計數在比較之前遞增，
			// 所以第一個就被攔下。等效於整個兵種被排除。
			continue
		}
		idx := int(g) - 1
		if idx < 0 || idx >= len(in) {
			continue
		}
		s := Strength(in[idx], opt)
		if s < AISortieMinStrength {
			continue
		}
		plan.Generals = append(plan.Generals, g)
		plan.Strength += s
		plan.Remaining -= s
		if plan.Strength >= stop {
			break
		}
	}
	return plan
}

// ── 湊完兵之後的出兵閘門（`sub_174C9` `loc_1754F`..`loc_175B9`）──────────

// AISortieMinGenerals 是**出兵的最低人數：3**。
//
// 兩種難度都有這道門檻，只是位置不同：
//
//	高難度：`sub_174C9` 的 `cmp byte ptr ss:[di-5], 3 / jb 不出兵`
//	低難度：`sub_16A9E` 收尾時 `已收 < 3 → *arg_C = 0`
//
// 湊不到三個人就不打。
const AISortieMinGenerals = 3

// AISortieGateRatio 是出兵閘門的倍率：**名單戰力 ≥ 目標省戰力 × 2**
// （`sub_17437`）。
//
// ⚠️ 與湊兵的停止倍率 **2.5** 不同（`AISortieStopRatioNum`）。
// 湊兵可能因為別的條件先停（人數上限、清單掃完），那時戰力不到 2.5 倍；
// 這道閘門用比較寬鬆的 2 倍再確認一次。
const AISortieGateRatio = 2

// SortieGateInput 是出兵閘門要看的東西。
type SortieGateInput struct {
	// Approved 是湊兵函式寫回的放行旗標（原版 `狀態[-12h]`）。
	Approved bool
	// Count 是名單人數（原版 `狀態[-5]`）。
	Count int
	// PlanStrength 是名單的攻擊力總和。
	PlanStrength int
	// TargetStrength 是目標省的戰力總和。
	TargetStrength int
	// TargetDesperate 是 `sub_17019(目標省)`：目標是不是軟柿子
	// （缺糧又被包圍，`aidesperate.go`）。
	TargetDesperate bool
	// TotalForce 是該省的**兵力總和**（`sub_306CF`）。
	//
	// ⛔ 這裡原本是 `Field234` 與 `Field236` 兩個「語意未解」的欄位。
	// 那是誤讀：`sub_13D23` 做的是
	//
	//	call    sub_306CF               ; 32-bit 回傳
	//	mov     ss:[di-236h], ax        ; ← 低位字
	//	mov     ss:[di-234h], dx        ; ← 高位字
	//
	// **兩格是同一個 32-bit 值**（`docs/re/13` §3，早就 confirmed），
	// 而那三條 `cmp`／`jg`／`jge` 是標準的 32-bit 比較拆成兩半。
	TotalForce int
}

// AISortieForceThreshold 是「家底夠不夠厚」的門檻：**兵力總和 ≥ 260,000**。
//
//	cmp     word ptr ss:[di-234h], 3        ; 高位
//	jg      放行
//	jge     再看低位
//	jmp     否決
//	cmp     word ptr ss:[di-236h], 0F7A0h   ; 低位
//
//	(3 << 16) | 0xF7A0 = 0x3F7A0 = 260,000
//
// 剛好是 **13 個滿員步兵師**（20,000）。決策鏈 A 步驟 5 的前置用的是
// 同一個門檻——`docs/re/28` §2 原本把它記成「`[-234h] > 3` 且
// `[-236h] ≥ F7A0h`」兩個條件，其實是一個。
const AISortieForceThreshold = 260000

// SortieGate 決定湊好的兵到底出不出（`sub_174C9` `loc_17568`）。
//
//	if !Approved 或 Count < 3:            不出兵
//	if 名單戰力 ≥ 目標戰力 × 2:            出兵    ← sub_17437
//	if 目標是軟柿子:                       出兵    ← sub_17019
//	if 兵力總和 ≥ 260,000:                 出兵
//	否則                                   不出兵
//
// 讀作：**湊夠了就打；沒湊夠但對方快垮了也打；兩者都不成立就看家底夠不夠厚。**
//
// ⚠️ `[-28h]` 那 10 格（`sub_17437` 掃的）是出兵名單的副本，
// 由 `sub_173B5` 從 `[-0EEh]` 複製過來——但 `sub_173B5` **只在高難度呼叫**。
// 低難度大概是由 `sub_16905`（未讀）做同樣的事，**未驗**（`docs/re/30` §5d）。
func SortieGate(in SortieGateInput) bool {
	if !in.Approved || in.Count < AISortieMinGenerals {
		return false
	}
	if in.PlanStrength >= in.TargetStrength*AISortieGateRatio {
		return true
	}
	if in.TargetDesperate {
		return true
	}
	return in.TotalForce >= AISortieForceThreshold
}
