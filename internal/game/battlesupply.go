package game

// 戰鬥決策鏈的**比率門檻**（`sub_3A57E`／`sub_3A672`／`sub_3A63C`／`sub_3A730`，
// `docs/re/31` §48）。
//
// 這一組是兩條鏈共用的局勢判斷，控制值 12（推倒重來）、值 19（全面接戰）
// 與值 2（佈防）。原本 §42 標「來源未解」，實際上答案早就在別的筆記裡：
//
//	word_64932/36/3A/3E = 第一方帶的黃金／糧食／彈藥／燃料（docs/re/05 §2）
//	word_64934/38/3C/40 = 第二方的同四項
//	sub_3A4CE(side)     = 該方 10 個單位的兵力（將領記錄 +17）總和
//
// 算出來的是**每 150 兵能分到多少資源**——補給還撐得住幾回合的意思。

// SupplyPerTroop 是 `sub_3A57E`／`sub_3A672` 的算式（兩支同型，只差分子）。
//
//	資源 == 0        → 0     （沒補給，一定吃緊）
//	兵力總和 == 0    → 15    （沒兵就沒有吃緊問題）
//	否則             → Round( 資源 / max(1, Round(兵力/150)) )
//
// `150` 是組語裡的 Turbo Pascal Real 常數（`CX:SI:DI = 0088/0000/1600`，
// 用 `tools/tp_real.py` 解出來）。
func SupplyPerTroop(resource, troops int) int {
	if resource <= 0 {
		return 0
	}
	if troops <= 0 {
		return AIBattleSupplyCap
	}
	d := roundDiv(troops, AIBattleSupplyPerTroop)
	if d == 0 {
		d = 1
	}
	return roundDiv(resource, d)
}

// roundDiv 是 Turbo Pascal 的 `Round(a/b)`。
//
// ⚠️ **.5 的取法是假說**：這裡照 8087 的預設「四捨六入五成雙」
// （banker's rounding）做，因為 `@Round$q4Real` 走的是 FPU／模擬器的
// round-to-nearest-even。**沒有實機對照過**。
//
// 影響範圍很小——只有剛好 .5 的輸入會差 1，而且門檻是 `< 15` 的整數比較，
// 差 1 只在邊界上才改變結果。真要驗，拿 DOSBox 跑一場資源湊到 .5 的戰鬥。
func roundDiv(a, b int) int {
	if b == 0 {
		return 0
	}
	q, r := a/b, a%b
	r2 := r * 2
	switch {
	case r2 > b:
		q++
	case r2 == b:
		if q%2 != 0 {
			q++
		}
	}
	return q
}

const (
	// AIBattleSupplyPerTroop 是分母的兵力單位：**150 兵**。
	AIBattleSupplyPerTroop = 150
	// AIBattleSupplyCap 是「沒兵」時直接回的值，同時也是門檻的比較基準。
	AIBattleSupplyCap = 15
)

// BattleSupply 是某一方帶進這場戰鬥的四項資源與兵力總和。
//
// 對應原版那四個成對的全域變數。名字照 `docs/spec/03` 的省份資源欄位取——
// 戰後結算（`sub_546D1`）就是把雙方這四項兩兩相加寫回省份的 `+0/+2/+4/+8`。
type BattleSupply struct {
	Gold, Food, Ammo, Fuel int
	// Troops 是該方 10 個單位的兵力（將領記錄 `+17`）總和，即 `sub_3A4CE`。
	Troops int
}

// SupplyTight 回答「這一項資源撐不撐得住」——原版 `sub_3A63C`／`sub_3A730`。
//
//	( 每 150 兵分到的量 + phase ) < 15
//
// ⭐ `turn` 是原版的 `byte_64900` = **回合數**（§49）。加上它之後
// 整條式子有了字面意義：
//
//	還能撐幾回合 + 已經打了幾回合 < 15
//	  = 這批補給撐不到第 15 回合（戰鬥上限 16 回合）
//
// 所以「吃緊」不是抽象的比率門檻，是**撐不到終局**。
func SupplyTight(resource, troops, turn int) bool {
	return SupplyPerTroop(resource, troops)+turn < AIBattleSupplyCap
}

// RatioGate 是決策鏈實際用的那個複合條件：**糧食夠、但黃金不夠**。
//
//	!sub_3A63C(side) && sub_3A730(side)
//	  = !糧食吃緊    && 黃金吃緊
//
// 兩條鏈都用它（§43）：兩方都成立 → 分支 A 選 12、分支 B 在彈藥為 0 時選 2；
// 只有第二方成立 → 分支 A 選 19。
//
// ⭐ 實際數量級：門檻等價於「資源 < 兵力 / 10」。省份的糧食通常遠高於這條線
// （實機湖北 18,050），黃金通常低於（4,200）——所以 `!糧食吃緊 && 黃金吃緊`
// 是**常見狀態**，那三個行動是主線而不是例外。
func (s BattleSupply) RatioGate(turn int) bool {
	return !SupplyTight(s.Food, s.Troops, turn) && SupplyTight(s.Gold, s.Troops, turn)
}

// TroopTotal 是 `sub_3A4CE`：一方 10 個槽位的兵力總和。
//
// 原版只檢查槽位有沒有單位（`id != 0`），**不檢查死活**——
// 陣亡的單位兵力本來就是 0，加進去不影響結果。這裡照同樣的寫法，
// 不多篩 `Alive()`，免得兩邊在「死了但槽位還在」的情況下分岔。
func TroopTotal(units []*Combatant) int {
	n := 0
	for _, u := range units {
		if u == nil {
			continue
		}
		n += int(u.Force())
	}
	return n
}
