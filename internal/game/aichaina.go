package game

// 決策鏈 A（`sub_17ADA`，667 行）的六個步驟。
//
// 這是**電腦唯一會出兵的地方**（`docs/re/28`／`docs/re/29`）。
// 決策鏈 B（`sub_15F3C`，`ai.go` 的 `Decide`）六步全是調動。
//
// 每一步成功就收工——**一次決策只做一件事**，與決策鏈 B 的優先序結構相同。

// ChainAOpts 是決策鏈 A 要看的旗標與外部狀態。
type ChainAOpts struct {
	// Hard 是 `byte_6FFCA & 1`：高難度。
	Hard bool
	// EnableStep5 是 `byte_6FFCA & 2`（bit 1）。它同時也讓離島三省
	// 跳過整條鏈（`AIChainASkipsProvince`）。
	EnableStep5 bool
	// EnableLastThree 是 `byte_6FFCA & 4`（bit 2）：啟用步驟 4 的中間層
	// 與步驟 6。
	EnableLastThree bool
	// FirstStage 是「第一期」（步驟 2 的前提，步驟 1 的排除條件）。
	FirstStage bool
	// Spectating 是觀戰模式，步驟 2 會跳過。
	Spectating bool
	// TotalForce 是 `[-236h]`，AISupplyOK 與步驟 5 都用它。
	//
	// ⚠️ 語意是**假說**（疑似兵力總和），由呼叫端傳入。
	TotalForce int
	// Field234 是 `[-234h]`，步驟 5 與出兵閘門用。語意未解。
	Field234 int
}

// ChainAResult 是決策鏈 A 的結果。
type ChainAResult struct {
	// Action 是決定要做的事。`Kind == AINone` 表示這一輪什麼也沒決定。
	Action AIAction
	// SortieMode 是出兵模式（1／2／3），只有 `Kind == AIAttack` 時有意義。
	SortieMode int
}

// ChainA 跑一次決策鏈 A。
//
// `gens` 與 `w.Units` 一一對應（`AIComfortGate` 需要）。
//
// ⚠️ **步驟 1／2 的下一跳用 `NextHop` 近似**。原版走的是
// `sub_14E85(狀態, ?, 3, 省)`，而 `arg_4 = 3` 會跳過 `sub_14B0D` 與
// `sub_14BEE` 兩支（各 100 行以上，未讀）；`NextHop` 對應的是
// `arg_4 = 1` 那條路徑（`sub_14CCF`，`docs/re/11` §5）。
// **兩者的差別未驗**，這是目前最大的已知偏差（`docs/re/28` §2b）。
func (w *AIWorld) ChainA(p ProvinceID, gens []General, opt ChainAOpts) ChainAResult {
	// 離島三省在 bit 1 設起來時整條跳過。
	if AIChainASkipsProvince(p, opt.EnableStep5) {
		return ChainAResult{}
	}
	prov, err := w.Table.At(p)
	if err != nil || prov.Commander == 0 {
		return ChainAResult{}
	}

	// ── 步驟 0：慰勞軍民 ──────────────────────────────────────────
	if w.AIComfortGate(p, gens, opt.Hard) {
		return ChainAResult{Action: AIAction{
			Kind: AIComfort, From: p, To: p, Step: "步驟 0 慰勞軍民"}}
	}

	supplyOK := w.AISupplyOK(p, opt.TotalForce, opt.EnableLastThree, opt.Hard)

	// ── 步驟 1：打通去救援的路 ─────────────────────────────────────
	//
	// ⚠️ **三道前置缺一不可**（`docs/re/28` §2，`loc_17BAC`）：
	//
	//	有敵對鄰省（`sub_5B7DC`）  ← 第一版漏了這道，見下
	//	非第一期
	//	司令 != 1
	//
	// 終點是「我方的、前線的、戰力 < 2000、未交戰」的省（`sub_149C8`），
	// 回傳的是**下一跳**。那一跳是敵省就打下來，是自己人或無主就調動。
	//
	// ⛔ 第一版沒檢查「有敵對鄰省」，於是**每一個省都走這一步**——
	// 後方省也去找救援目標、也調動，把步驟 3 之後整個擋住。
	// 症狀是「39 省跑一輪零出兵」，而單元測試全綠看不出來
	// （`TestChainAOnRealSave` 的對照子測試就是為了抓它）。
	if w.Hostile(p) != 0 && !opt.FirstStage && prov.Commander != 1 {
		if r, ok := w.rescueStep(p, prov, supplyOK, opt, "步驟 1"); ok {
			return r
		}
	}

	// ── 步驟 2：第一期只調動，不出兵 ───────────────────────────────
	//
	// 前置是「第一期、非觀戰、`sub_5B76E` 鄰接」。
	// ⚠️ `sub_5B76E` 未讀，這裡用「有敵對鄰省」近似——**標為差異**。
	if opt.FirstStage && !opt.Spectating && w.Hostile(p) != 0 {
		if hop := w.rescueHop(p, prov.Commander); hop != 0 {
			if hp, err := w.Table.At(hop); err == nil &&
				(hp.Commander == 0 || hp.Commander == prov.Commander) {
				return ChainAResult{Action: AIAction{
					Kind: AITransfer, From: p, To: hop, TransferKind: 1,
					Step: "步驟 2 第一期調動"}}
			}
		}
	}

	// ── 步驟 3：收候選（敵省，好不好打）────────────────────────────
	//
	// 這一步**只收候選不動作**，結果給步驟 4 用。
	var cands []ProvinceID
	if w.GeneralCount(p) >= AIStep3MinGenerals && supplyOK {
		cands = w.attackCandidates(p, prov.Commander, gens, opt.Hard)
	}

	// ── 步驟 4：三層優先序挑一個打 ─────────────────────────────────
	if len(cands) >= 1 {
		if t := w.ChainATarget(cands, prov.Commander, opt.EnableLastThree); t != 0 {
			return ChainAResult{
				Action:     AIAction{Kind: AIAttack, From: p, To: t, Step: "步驟 4 三層優先序"},
				SortieMode: 2,
			}
		}
	}

	// ── 步驟 5：兵多將廣時挑最弱的鄰省擴張 ─────────────────────────
	if opt.EnableStep5 && opt.Field234 > 3 && opt.TotalForce >= AISortieField236Threshold &&
		w.Hostile(p) != 0 && supplyOK {
		// 候選：有效鄰省，跳過交戰中的。⚠️ 與步驟 3 的五道篩選不同，
		// 這裡**不排除自己人與無主省**（`docs/re/28` §2）。
		var c []ProvinceID
		for _, n := range prov.Neighbours {
			np, err := w.Table.At(n)
			if err != nil || np.Flags&ProvinceFlagInBattle != 0 {
				continue
			}
			c = append(c, n)
		}
		if t := w.WeakestNeighbourTarget(c); t != 0 {
			return ChainAResult{
				Action:     AIAction{Kind: AIAttack, From: p, To: t, Step: "步驟 5 挑最弱的擴張"},
				SortieMode: 3,
			}
		}
	}

	// ── 步驟 6：被包圍就突圍 ───────────────────────────────────────
	if opt.EnableLastThree && w.Encircled(p) && supplyOK && !AIBreakoutSkipsProvince(p) {
		if t := w.BreakoutTarget(p, prov.Commander); t != 0 {
			return ChainAResult{
				Action:     AIAction{Kind: AIAttack, From: p, To: t, Step: "步驟 6 突圍"},
				SortieMode: 1,
			}
		}
	}

	return ChainAResult{Action: AIAction{From: p}}
}

// rescueStep 是步驟 1 的動作部分：拿到下一跳之後決定打還是調。
//
// 回 `ok == false` 表示這一步沒做出決定，鏈要繼續往下走。
func (w *AIWorld) rescueStep(p ProvinceID, prov *Province, supplyOK bool,
	opt ChainAOpts, label string) (ChainAResult, bool) {
	hop := w.rescueHop(p, prov.Commander)
	if hop == 0 {
		return ChainAResult{}, false
	}
	hp, err := w.Table.At(hop)
	if err != nil {
		return ChainAResult{}, false
	}
	if hp.Commander != 0 && hp.Commander != prov.Commander {
		// 擋路的敵省，條件齊了就打下來。
		if w.AIHasAdvantage(p, hop, opt.Hard) &&
			w.GeneralCount(p) >= AIStep1MinGenerals && supplyOK {
			return ChainAResult{
				Action:     AIAction{Kind: AIAttack, From: p, To: hop, Step: label + " 打通救援的路"},
				SortieMode: 2,
			}, true
		}
		// 條件不齊 → 這一步不做決定，鏈繼續往下（原版 `loc_17C6C` 的分流）。
		return ChainAResult{}, false
	}
	// 自己人或無主省，直接調動過去。
	return ChainAResult{Action: AIAction{
		Kind: AITransfer, From: p, To: hop, TransferKind: 1,
		Step: label + " 往救援目標調動"}}, true
}

// 兩道將領數門檻。
const (
	// AIStep1MinGenerals 是步驟 1 打下擋路敵省的門檻（`cmp ss:[di-22Eh], 3 / jl`）。
	AIStep1MinGenerals = 3
	// AIStep3MinGenerals 是步驟 3 收候選的門檻（`cmp ss:[di-22Eh], 4 / jge`）。
	AIStep3MinGenerals = 4
)

// rescueHop 回傳「往我方快垮的前線省」的下一跳（`sub_14E85` → `sub_14CCF`）。
//
// 終點的篩選在 `sub_149C8`，與 `RescueTargets`（`ai.go`）是同一組條件。
func (w *AIWorld) rescueHop(p ProvinceID, faction GeneralID) ProvinceID {
	for _, target := range w.RescueTargets(faction) {
		if target == p {
			continue // 自己不救自己
		}
		if hop := w.NextHop(p, target, faction); hop != 0 {
			return hop
		}
	}
	return 0
}

// attackCandidates 是步驟 3 的五道篩選（`docs/re/28` §2）。
//
//	鄰省 != 0 且 != 255
//	未交戰（`+32` bit 6）
//	司令 != 我方          ← 排除自己人
//	司令 != 0             ← **排除無主省**
//	`sub_16F89`（我方壓倒它）**或** `sub_17019`（它自己撐不住）
//
// ⚠️ 排除無主省這一條與調動時的 `transferable`（無主省是首選）正好相反：
// 出兵打的是有主的省，無主的省用調動去拿。
func (w *AIWorld) attackCandidates(p ProvinceID, faction GeneralID,
	gens []General, hard bool) []ProvinceID {
	prov, err := w.Table.At(p)
	if err != nil {
		return nil
	}
	var out []ProvinceID
	for _, n := range prov.Neighbours {
		np, err := w.Table.At(n)
		if err != nil {
			continue
		}
		if np.Flags&ProvinceFlagInBattle != 0 {
			continue
		}
		if np.Commander == faction || np.Commander == 0 {
			continue
		}
		if w.AIHasAdvantage(p, n, hard) || w.AIDesperate(n, gens) {
			out = append(out, n)
		}
	}
	return out
}
