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
