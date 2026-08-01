package game

import "fmt"

// 秘密行動（政略指令 11）的兩個子項，出自 `sub_1FC1C`（派遣游擊隊）與
// `sub_20447`（鼓動學潮）。
//
// 兩支的形狀相同：**先扣錢，再擲一次骰，成功才有效果**。
// 這是目前解出來最完整的一組「帶亂數的規則」，補上 `70-ai.md` 一直缺的
// 「亂數在哪裡介入、範圍多大」。

// 成功率。原版的判定是「`Random(10)` 落在某個區間」——
// 注意**不是取餘數**，`Random(n)` 是 `(seed × n) >> 32`（`docs/re/17`）。
//
//	游擊隊  Random(10)，0..3 成功  → 4/10
//	學潮    Random(10)，0..1 成功  → 2/10
const (
	// GuerrillaRollRange 是游擊隊判定的亂數範圍。
	GuerrillaRollRange = 10
	// GuerrillaSuccessMax 是成功的上界（含）。
	GuerrillaSuccessMax = 3
	// StudentRollRange / StudentSuccessMax 是學潮的對應值。
	StudentRollRange   = 10
	StudentSuccessMax  = 1
	// StudentProtestCost 是鼓動學潮的固定成本（`sub word ptr [di-6235h], 5DCh`）。
	//
	// **不論成功失敗都先扣**——那一行在成敗判定之前。
	StudentProtestCost = 1500
)

// 游擊隊成功後的破壞效果（`sub_1FC1C` 的 loc_20316 之後）。
const (
	// GuerrillaKeepNumer / GuerrillaKeepDenom 是資源被破壞後剩下的比例：
	// `資源 = 資源 ÷ 10 × 7`，也就是**剩 70%**。
	//
	// ⚠️ 整數除法先做，所以不是精確的 70%——1,000 會變 700，
	// 但 9 會變 0（9÷10=0，0×7=0）。**照抄這個順序，不要改寫成 ×7÷10。**
	GuerrillaKeepNumer = 7
	GuerrillaKeepDenom = 10
	// GuerrillaArsenalRollRange 是「炸掉一座兵工廠」的判定：
	// `Random(5) == 0` 而且該省還有兵工廠。
	GuerrillaArsenalRollRange = 5
)

// CovertResult 記錄一次秘密行動的結果。
type CovertResult struct {
	// Success 是這次行動成不成功。
	Success bool
	// Roll 是實際擲出的亂數，方便重現與對照。
	Roll int
	// Cost 是花掉的黃金。
	Cost int
	// ArsenalDestroyed 只有游擊隊會用：有沒有炸掉一座兵工廠。
	ArsenalDestroyed bool
}

// decay 是原版的「剩 70%」寫法：**先除後乘**。
func decay(v uint16) uint16 {
	return v / GuerrillaKeepDenom * GuerrillaKeepNumer
}

// SendGuerrillas 派遣游擊隊到目標省（`sub_1FC1C`）。
//
// `cost` 是這次行動的花費——原版在畫面上問「司令派遣要花 N，確定？」，
// **那個 N 怎麼算還沒讀**（`docs/mechanics/50-diplomacy.md` §2），
// 所以由呼叫端傳進來。
//
// 流程照原版：
//
//	來源省黃金 -= cost
//	Random(10) ≤ 3 → 成功
//	成功：目標省的黃金／糧食／彈藥各剩 70%
//	      再 Random(5)，== 0 且有兵工廠 → 拆掉一座
//
// ⚠️ **燃料、煤礦、鐵礦不受影響**——原版只動那三格。
func (w *AIWorld) SendGuerrillas(from, target ProvinceID, cost int, rng *Rand) (CovertResult, error) {
	src, err := w.Table.At(from)
	if err != nil {
		return CovertResult{}, err
	}
	dst, err := w.Table.At(target)
	if err != nil {
		return CovertResult{}, err
	}
	if cost < 0 {
		return CovertResult{}, fmt.Errorf("game: 成本不能為負（%d）", cost)
	}
	if int(src.Gold) < cost {
		return CovertResult{}, fmt.Errorf("game: 資金不足（有 %d，要 %d）", src.Gold, cost)
	}
	src.Gold -= uint16(cost)

	res := CovertResult{Cost: cost}
	res.Roll = rng.Int(GuerrillaRollRange)
	res.Success = res.Roll <= GuerrillaSuccessMax
	if !res.Success {
		return res, nil
	}

	dst.Gold = decay(dst.Gold)
	dst.Food = decay(dst.Food)
	dst.Ammo = decay(dst.Ammo)

	if rng.Int(GuerrillaArsenalRollRange) == 0 && dst.Arsenals > 0 {
		dst.Arsenals--
		res.ArsenalDestroyed = true
	}
	return res, nil
}

// IncitStudentProtest 鼓動學潮（`sub_20447`）。
//
// 固定花 1,500 黃金，**成敗都扣**。`Random(10) ≤ 1` 才成功（2/10）。
//
// ⚠️ 成功之後原版呼叫 `sub_5C441(0, 目標省, 10)`，那支 329 行**還沒讀**——
// 所以這裡只做到「判定成功」，**實際效果沒有實作**。
// 回傳的 `Success` 是可信的，但目標省不會有任何改變。
// 這是明確的缺口，不是「做完了」。
func (w *AIWorld) IncitStudentProtest(from ProvinceID, rng *Rand) (CovertResult, error) {
	src, err := w.Table.At(from)
	if err != nil {
		return CovertResult{}, err
	}
	if int(src.Gold) < StudentProtestCost {
		return CovertResult{}, fmt.Errorf("game: 資金不足（有 %d，要 %d）",
			src.Gold, StudentProtestCost)
	}
	src.Gold -= StudentProtestCost

	res := CovertResult{Cost: StudentProtestCost}
	res.Roll = rng.Int(StudentRollRange)
	res.Success = res.Roll <= StudentSuccessMax
	return res, nil
}
