package game

// 電腦決策鏈 A 的第一步：**要不要慰勞軍民**（`sub_16E77`，148 行）。
//
// 這是**第一條解出來的「電腦為什麼選這個指令」**。先前 `docs/re/28` 只解出
// 六個步驟的骨架與呼叫關係，每一步的判斷函式都還是黑盒；這是第一個打開的。

// 慰勞的士氣／體力門檻，**依難度分兩檔**（`byte_6FFCA & 1`）。
//
//	高難度：體力 < 80 或 士氣 < 70  → 慰勞
//	低難度：體力 < 60 或 士氣 < 50  → 慰勞
//
// 士氣的上限是 80（`MoraleMax`），所以高難度的士氣門檻 70 相當接近滿值——
// **電腦在高難度下會更勤於維持部隊狀態**。體力上限 100，門檻 80／60 同理。
const (
	AIComfortStaminaHard = 80
	AIComfortMoraleHard  = 70
	AIComfortStaminaEasy = 60
	AIComfortMoraleEasy  = 50
)

// AIComfortForceThreshold 是「值得慰勞」的兵力門檻。
//
// 原版算的是 `Round(兵力 ÷ sub_349C1(將領)) ≥ 10000`，而 `sub_349C1`
// 回的是 **1 / 人力權重**（`docs/re/10` §3），所以等價於
// **`兵力 × 人力權重 ≥ 10000`**——也就是**半滿以上**：
//
//	步兵   權重 1    滿員 20,000 → 門檻 10,000
//	騎兵   權重 2    滿員 10,000 → 門檻  5,000
//	砲兵   權重 10   滿員  2,000 → 門檻  1,000
//	裝甲兵 權重 100  滿員    200 → 門檻    100
//
// 四個兵種都正好是滿員的一半。**電腦不慰勞殘兵**——兵力掉太多的部隊
// 它會走別的步驟（撤回後方之類），不浪費指令數在上面。
//
// 這與 `docs/re/10` 的「兵力門檻一律用兵力 × 人力權重表達」是同一套寫法，
// 是第三個獨立出現的例子。
const AIComfortForceThreshold = 10000

// AIComfortGate 是「電腦要不要對這個省下慰勞軍民」的判斷（`sub_16E77`）。
//
// 掃該省的將領，只要有**一個**滿足全部條件就回 true：
//
//	所屬省 == 目標省
//	`+16` == 1                    ← 完全等於 1，不是 bit test
//	效忠勢力 == 該省的司令
//	兵力 × 人力權重 ≥ 10,000      ← 半滿以上
//	且（體力 < 門檻 或 士氣 < 門檻）
//
// [雷] `+16` 的檢查是 `cmp byte ptr [di+7A8Dh], 1` —— **值要正好是 1**，
// 不是 bit test。那一格的值域是 `{0, 1, 32}`（`docs/spec/02` §3）。
//
// ⚠️ **remake 差異**：`CombatUnit` 只存了 `Active`（bit 0），沒有保留
// `+16` 的原始值，所以這裡用 `Active` 近似。在值域 `{0, 1, 32}` 下
// 兩者對 1 與 0 的判斷相同，**但 32 會被 `Active` 判成 false**
// （32 的 bit 0 是 0），所以目前恰好等價。若之後發現 `+16` 有別的值
// （例如 33），就必須改成保留原始值。
//
// `gens` 與 `w.Units` 一一對應。`hard` 來自 `byte_6FFCA & 1`。
func (w *AIWorld) AIComfortGate(p ProvinceID, gens []General, hard bool) bool {
	prov, err := w.Table.At(p)
	if err != nil {
		return false
	}
	staminaMin, moraleMin := AIComfortStaminaEasy, AIComfortMoraleEasy
	if hard {
		staminaMin, moraleMin = AIComfortStaminaHard, AIComfortMoraleHard
	}

	for i := range w.Units {
		u := &w.Units[i]
		if u.Province != p || i >= len(gens) {
			continue
		}
		g := &gens[i]
		// `+16` 要正好是 1（見上方對 remake 差異的說明）。
		if !u.Active {
			continue
		}
		if u.Faction != prov.Commander {
			continue
		}
		if !AIWorthComforting(g.Force, g.Branch) {
			continue
		}
		if int(g.Stamina) < staminaMin || int(g.F30) < moraleMin {
			return true
		}
	}
	return false
}

// AIWorthComforting 回報某個部隊的兵力有沒有到「值得慰勞」的半滿門檻。
//
// 用 `兵力 × 人力權重 ≥ 10000` 表達，與原版的除法等價而且不會有除零問題。
//
// 人力權重就是 `BranchWeight`（1／4／10／100）。⚠️ 兵種 4 的權重是 4，
// 推出來的半滿門檻是 2,500 而不是社群說的騎兵 5,000——
// `strength.go` 已經記過這個對不上的地方，**不因為這裡而改動它**。
func AIWorthComforting(force uint16, branch uint8) bool {
	wgt := BranchWeight(branch)
	if wgt == 0 {
		return false
	}
	return int(force)*wgt >= AIComfortForceThreshold
}
