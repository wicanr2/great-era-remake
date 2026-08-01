package game

import "math"

// 攻擊力（戰力）公式。整條算式出自 `sub_5A0B9`（`docs/re/08` §4），
// 包含七個 Turbo Pascal 48-bit Real 常數，全部解出來都是整齊的整數。
//
// **`20000` 這個常數就在這裡。** `CLAUDE.md` §1.5 記過「20000／10000／
// 2000／200 是很好的 grep 起點，但搜不到是正常的」——它確實不是整數常數，
// 而是編碼成 48-bit Real 的尾數（`0x8F, 0x0000, 0x1C40`），
// grep `20000` 或 `4E20` 都不會中。

// 兵種權重。`sub_5A0B9` 開頭依 `+21`（兵種）取值。
//
// **這也確認了 `+21` 的值域就是 1／4／5／6**——`sub_5A0B9` 只認這四個，
// 其餘落到未初始化的 stack 值。
var branchWeight = map[uint8]int{
	1: 1,
	4: 10,
	5: 100,
	6: 4,
}

// BranchWeight 回傳兵種的戰力權重，未知兵種回 0。
//
// ⚠️ 哪個號碼是哪個兵種**沒有證據**。一個觀察（不是結論）：
// 權重 1／10／100 與社群傳的滿員數 20000／2000／200 成反比，
// 三項吻合；但 4 對應的 5000 與社群說的騎兵 10000 對不上。
// 社群數值不一定出自 1992 原版（`CLAUDE.md` §4），**不拿來定名**。
func BranchWeight(branch uint8) int { return branchWeight[branch] }

// StrengthInput 是算戰力要用到的欄位，偏移標在旁邊。
//
// 這些是**執行期單位記錄**的偏移（基址 0x7A7D）。`+0` 與 `+17` 同時也是
// `MAN(N).DAT` 的欄位（`docs/spec/02` 的 AbilityA 與 Force），
// `+19`／`+20`／`+29`／`+30` 的語意**未解**——公式用得到它們，
// 但不知道畫面上叫什麼。
type StrengthInput struct {
	Ability uint8  // +0，0..100。名將加成看的就是它
	Force   uint16 // +17，兵力
	F19     uint8  // +19，語意未解
	F20     uint8  // +20，語意未解
	F29     uint8  // +29，語意未解。開平方根
	F30     uint8  // +30，語意未解。就是每回合衰減 20% 的那一格
	Branch  uint8  // +21，兵種
	General GeneralID
	Faction GeneralID // +14，效忠的勢力領袖
}

// 難度／勢力加成的設定值，取自 `byte_6FFCA & 0x70`。
const (
	// BonusMaskFaction 是取出這三個位元的遮罩。
	BonusMaskFaction = 0x70
	// BonusDouble：十大勢力的單位戰力 ×2。
	BonusDouble = 0x10
	// BonusHalfMore：十大勢力的單位戰力 ×1.5。
	BonusHalfMore = 0x20
	// BonusLeaderOnly：只有效忠蔣中正（ID 1）的單位 ×1.5。
	BonusLeaderOnly = 0x40
)

// StrengthOpts 是公式裡幾個由全域狀態決定的開關。
type StrengthOpts struct {
	// FactionBonus 是 `byte_6FFCA & 0x70`。
	FactionBonus uint8
	// MajorPower 回報某個勢力領袖是不是十大勢力之一（`sub_5C7FE`，
	// `docs/re/06`）。
	MajorPower func(GeneralID) bool
	// Stage 是 `byte_6FE88`。公式有兩處看它，但它的語意仍未確認
	// （CONTEXT worklist #4：值 1/4 才載字模，推測是幕別或階段）。
	Stage uint8
}

// LeaderChiang 是蔣中正的將領 ID。公式裡有兩處專門檢查 `+14 == 1`。
//
// ID 1 = 蔣中正是 `docs/spec/03` §3 從 `TOWN(1)` 的司令欄位解出來的
// ——那張表排出來就是 1926 年的割據圖。
const LeaderChiang GeneralID = 1

// Strength 算出一個單位的戰力，逐步照 `sub_5A0B9`。
//
// 每一步都有 Round，所以**順序不能重排**——原版是一連串
// `Round(Real ...)`，中途的截斷會累積。這裡用整數運算重現：
// 除數都是正整數、被除數非負，所以 `(x + d/2) / d` 等價於
// Turbo Pascal 的 half-away-from-zero `Round`。
//
//	w  = 兵種權重
//	S  = (F30 + Ability + 3×F19 + 5×F20) / 10 × Force × w / 20000
//	S  = S × sqrt(F29) / 9
//	S  = S × F30 / 85
//	S  = S × Ability / 80
//	名將加成（依 Ability 分五級）
//	後期將領加成
//	勢力／難度加成
//	兵種 5 最後再 ×2
func Strength(in StrengthInput, opt StrengthOpts) int {
	w := BranchWeight(in.Branch)
	if w == 0 {
		return 0 // 原版在這裡讀到未初始化的 stack 值，我們回 0
	}

	// 第一步。用 float64 重現原版的 Real 運算——中間值會超過 int32
	// （20000 兵力 × 權重 100 × 分子）而且原版本來就是浮點。
	base := float64(int(in.F30)+int(in.Ability)+3*int(in.F19)+5*int(in.F20)) / 10
	s := round(base * float64(in.Force) * float64(w) / 20000)

	s = round(float64(s) * sqrt(float64(in.F29)) / 9)
	s = round(float64(s) * float64(in.F30) / 85)
	s = round(float64(s) * float64(in.Ability) / 80)

	// 名將加成。原版用整數移位與除法，不是浮點，所以照抄整數版。
	switch a := in.Ability; {
	case a == 100:
		s += (s / 3) * 2 // ×5/3
	case a >= 98:
		s += s / 2 // ×1.5
	case a >= 95:
		s += s / 4 // ×1.25
	case a >= 90:
		s += s / 6
	case a >= 85:
		s += s / 8
	}

	// 後期將領。`byte_6FE88 != 1` 且將領 ID >= 107 時 +25%。
	//
	// ⚠️ 這一條**在原版資料上可能永遠不成立**：第二、三期的名表只有
	// 106 位將領，ID 107 以上不存在。照抄是為了忠實，不是因為它會生效。
	if opt.Stage != 1 && in.General >= 107 {
		s += s / 4
	}

	// 勢力／難度加成。
	major := opt.MajorPower != nil && opt.MajorPower(in.Faction)
	if major {
		switch opt.FactionBonus & BonusMaskFaction {
		case BonusDouble:
			s *= 2
		case BonusHalfMore:
			s = round(float64(s) * 1.5)
		case BonusLeaderOnly:
			if in.Faction == LeaderChiang {
				s = round(float64(s) * 1.5)
			}
		}
	} else if in.Faction == LeaderChiang && opt.Stage == 1 {
		s = round(float64(s) * 1.5)
	}

	// 兵種 5 最後再翻倍——它的基礎權重已經是最高的 100。
	if in.Branch == 5 {
		s *= 2
	}
	return s
}

// round 重現 Turbo Pascal 的 Round：正好一半時取絕對值較大的那個
// （half-away-from-zero）。戰力全程非負，所以只處理正的一側。
//
// ⚠️ 與 `CombatDecay` 同一個未確認點：Delphi 之後的 RTL 改成 banker's
// rounding。這一版是 Turbo Pascal 5.5/6.0，文件說 half-away-from-zero，
// **但沒有實機驗證過**（`docs/mechanics/30-combat.md` §3h）。
func round(x float64) int {
	if x < 0 {
		return -int(-x + 0.5)
	}
	return int(x + 0.5)
}

// sqrt 是 math.Sqrt 的別名，獨立出來只是為了讓上面的公式讀起來
// 貼近原版的 `Sqrt(Real)`。
func sqrt(x float64) float64 { return math.Sqrt(x) }

// ---------------------------------------------------------------------------
// 戰損。`sub_51D68` 依「雙方戰力差是否超過 4 倍」分兩條路，
// 兩支的算法完全不同（`docs/re/08` §5）。
// ---------------------------------------------------------------------------

// routWeight 是**第二張**兵種權重表（`sub_5190E`）。
//
// ⚠️ **與 `branchWeight` 不完全相同**：兵種 6 在這裡是 2，在攻擊力公式裡是 4。
// 其餘三個一樣。這不是抄錯——原版就是兩張獨立的表，用途不同。
var routWeight = map[uint8]int{
	1: 1,
	4: 10,
	5: 100,
	6: 2, // ← 攻擊力那張表是 4
}

// RoutWeight 回傳一面倒結算用的兵種權重（`sub_5190E`）。
func RoutWeight(branch uint8) int { return routWeight[branch] }

// CasualtiesEven 算勢均力敵時雙方的兵力損失（`sub_51B94`）。
//
//	pct = Round(atkByE / atkByF × 100)      ← E 是兵種 4 時再除以 2
//	損失F = Round( 兵力F / 4 / (pct+100) × pct )
//	損失E = Round( 兵力E / 4 / (pct+100) × 100 )
//
// ⚠️ **前兩個參數是「各自打出去的攻擊值」，不是「打在身上的」。**
// 原版的 `pct = var_8 / var_4`，而 `var_8` 是用 E 的強度算的
// ——也就是 E 施加於 F 的值。傳反了會變成「打得越兇損失越大」。
//
// 兩個係數 `pct/(pct+100)` 與 `100/(pct+100)` 相加恆為 1，
// 所以**雙方的損失比例合計恆為各自兵力的 1/4**。戰力相等時（pct = 100）
// 各損失自己兵力的 1/8。
//
// ⚠️ **任一方是兵種 4 時，F 完全不受損失**——原版在寫回 F 的損失之前
// 檢查兩邊的 `+21`，只要有一邊是 4 就跳過那次寫入。
//
// 這條規則的**方向未確認**：E／F 與攻守的對應要等 `sub_530B4` 的
// 呼叫端讀完才知道（`docs/playtest/04` §3）。實跑起來的效果是
// 「兵種 4 打不動人、自己還會被磨掉」，讀起來怪；若 E 其實是守方，
// 就變成「兵種 4 防守時攻擊者打不動它」，合理得多。
//
// `atkByE`／`atkByF` 是 `sub_51D68` 算出的兩個攻擊值（`Strength` 的產物
// 再經地形修正係數）。E 是 `sub_51D68` 的 `arg_E`，F 是 `arg_10`。
func CasualtiesEven(atkByE, atkByF int, forceE, forceF uint16, branchE, branchF uint8) (lossE, lossF int) {
	if atkByF <= 0 {
		return 0, 0
	}
	pct := round(float64(atkByE) / float64(atkByF) * 100)
	if branchE == 4 {
		pct = round(float64(atkByE) / float64(atkByF) * 100 / 2)
	}
	denom := float64(pct + 100)
	lossF = round(float64(forceF) / 4 / denom * float64(pct))
	lossE = round(float64(forceE) / 4 / denom * 100)

	// 任一方是兵種 4 → F 不受損失。
	if branchE == 4 || branchF == 4 {
		lossF = 0
	}
	return clampLoss(lossE, forceE), clampLoss(lossF, forceF)
}

// CasualtiesRout 算一面倒時的損失（`sub_51972`）。
//
//	人力值(u) = 兵種權重(sub_5190E) × 兵力
//	if 對方攻擊值 × 10 >= 人力值(u):  u 全滅
//	else:                            損失 = Round(對方攻擊值 × 10 / 兵種權重(u))
//
// 兩條分支是自洽的：`攻擊×10/W < 兵力` 恰好等價於 `攻擊×10 < W×兵力`，
// 所以損失永遠不會超過兵力——**全滅是門檻，不是溢位的補救**。
//
// `atkOnE` 是 `sub_51D68` 的 var_4（用 F 的強度算出來的、施加於 E 的值），
// `atkOnF` 是 var_8。
func CasualtiesRout(atkOnE, atkOnF int, forceE, forceF uint16, branchE, branchF uint8) (lossE, lossF int) {
	lossE = routSide(atkOnE, forceE, branchE)
	lossF = routSide(atkOnF, forceF, branchF)
	return lossE, lossF
}

func routSide(atk int, force uint16, branch uint8) int {
	w := RoutWeight(branch)
	if w == 0 {
		return 0
	}
	if atk*10 >= w*int(force) {
		return int(force) // 全滅
	}
	return clampLoss(round(float64(atk*10)/float64(w)), force)
}

func clampLoss(loss int, force uint16) int {
	if loss < 0 {
		return 0
	}
	if loss > int(force) {
		return int(force)
	}
	return loss
}
