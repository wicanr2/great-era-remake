package game

// 戰鬥規則。依據：docs/mechanics/30-combat.md
//
// **這一層只放已經從程式碼讀出來的規則**，沒解出來的不猜。
// 目前只有一條完整的算式（每回合衰減）與回合初始化的欄位重置。

// 同一個欄位（執行期單位記錄的 +30）有**兩條不同的衰減規則**，
// 出現在戰鬥的不同階段。兩個除數都是從 Turbo Pascal 的 48-bit Real
// 常數解出來的（exponent bias 129，尾數最高 byte 給小數部分）。
const (
	// DecayThreshold 是每回合衰減的門檻：值大於它才會衰減。
	DecayThreshold = 40
	// DecayDivisor 是每回合衰減的除數。sub_54826 的常數
	// CX:SI:DI = 0x0083/0x0000/0x2000：2^(131-129) × 1.25 = 5.0。
	DecayDivisor = 5
	// CombatDecayDivisor 是戰鬥判定後的衰減除數。sub_53DA9 的常數
	// CX:SI:DI = 0x0084/0x0000/0x2000：2^(132-129) × 1.25 = 10.0。
	// **這一條沒有門檻**（該函式裡沒有與 40 的比較）。
	CombatDecayDivisor = 10
)

// Decay 套用每回合的衰減：值大於 40 時扣掉當前值的 20%。
//
// 原版：`v -= Round(Real(v) / 5.0)`。
//
// Turbo Pascal 的 `Round` 是四捨五入到最近的整數（.5 進到偶數），
// 但 `v / 5` 的小數只可能是 0／.2／.4／.6／.8，**碰不到 .5**，
// 所以整數版的 `(v + 2) / 5` 完全等價，不需要浮點數。
//
//	v=41 → 41 - (43/5=8)  = 33
//	v=43 → 43 - (45/5=9)  = 34
//	v=100 → 100 - (102/5=20) = 80
//
// ⚠️ **被衰減的欄位是什麼還不知道**（執行期單位記錄的 +30）。
// 候選是體力或士氣（`2.15` 詞條 80、81），沒有證據指定哪一個。
func Decay(v uint8) uint8 {
	if v <= DecayThreshold {
		return v
	}
	return v - uint8((uint16(v)+2)/DecayDivisor)
}

// CombatUnit 是戰場上一個單位的狀態。
//
// 原版的執行期單位記錄是 33 bytes（基址 0x7A7D），**與 MAN(N).DAT 的
// 33 bytes 是兩套不同的佈局**（docs/mechanics/30-combat.md §3c）。
// 這裡只放已經解出用途的欄位，其餘不猜。
type CombatUnit struct {
	// General 是這個單位的將領 ID，0 表示空槽。
	General GeneralID

	// Active 對應執行期記錄的 +16。回合初始化時設為 1，
	// `sub_15018` 會檢查它的 bit 0。
	Active bool

	// Cell 是所在的戰場格，對應 +5。NoCell（0xFF）表示不在場上。
	Cell CellIndex

	// Max / Current 對應 +6 與 +7 —— **機動力**上限與剩餘。
	// 回合初始化時 `+7 = +6`；移動函式 `sub_4A1C0` 每走一格做 `+7 -= 成本`，
	// 而呼叫端先檢查 `+7 >= 成本` 才准走（`docs/re/07` §7）。
	// 對上 `3.15` 詞條 51「機動力」。
	Max, Current uint8

	// Decaying 對應 +30：大於 40 時每回合衰減 20%。語意未知。
	Decaying uint8
}

// BeginTurn 套用回合開始的重置，順序照 sub_5446D。
//
//	+16 = 1        啟用
//	+7  = +6       當前值回到最大值
//	+13 = 0        歸零（語意未知，這裡沒有對應欄位）
//	+8  = 0        歸零（同上）
//
// **`+5 = 0xFF` 沒有在這裡重現**——那是執行期的空槽標記，
// remake 用 `General == 0` 表示空槽，不需要哨兵值。
func (u *CombatUnit) BeginTurn() {
	if u.General == 0 {
		return
	}
	u.Active = true
	u.Current = u.Max
}

// EndTurn 套用回合結束的衰減。
func (u *CombatUnit) EndTurn() {
	if u.General == 0 {
		return
	}
	u.Decaying = Decay(u.Decaying)
}

// UnitsPerSide 是每方的**部隊**上限。
//
// **三處獨立對上**：
//   - 戰鬥狀態記錄 469 B 裡每方 10 個槽（docs/re/05）
//   - `sub_54826` 的衰減迴圈跑 1..10
//   - 執行期的兩張清單 `0x750` 與 `0x764` 相差 0x14 = 20 bytes
//     = 10 個 word，是相鄰的兩張表
//
// ⚠️ 戰鬥狀態還有**另一層 100 格的單位陣列**（`0x6742`／`0x680A`，
// 對應 .DT2 的 200 B 區）。原版的詞表也分兩個量：`3.15` 的 48「單位數」
// 與 49「士兵數」。這裡的 10 是部隊層，不要與那 100 混用。
const UnitsPerSide = 10

// Side 是戰鬥的一方。原版用兩張相鄰的單位清單（`0x750` 與 `0x764`），
// 以及 `byte_64901`（取 1 或 2）決定誰先動。
type Side int

const (
	// SideA 對應單位清單 0x750。
	SideA Side = iota
	// SideB 對應單位清單 0x764。
	SideB
)

// Battle 是一場戰鬥的雙方單位。
//
// 哪一方是攻、哪一方是守由 `byte_64901` 決定（詞表有「攻方」「守方」，
// `2.15` 166/167），**但原版哪個值對應哪一方還沒驗**，
// 所以這裡用中性的 A／B 命名。
type Battle struct {
	Units [2][UnitsPerSide]CombatUnit
}

// BeginTurn 對雙方所有單位套用回合開始的重置（sub_543C2／sub_5446D）。
func (b *Battle) BeginTurn() {
	for s := range b.Units {
		for i := range b.Units[s] {
			b.Units[s][i].BeginTurn()
		}
	}
}

// EndTurn 對雙方所有單位套用回合結束的衰減（sub_54826）。
func (b *Battle) EndTurn() {
	for s := range b.Units {
		for i := range b.Units[s] {
			b.Units[s][i].EndTurn()
		}
	}
}

// CombatDecay 套用戰鬥判定後的衰減：無條件扣掉當前值的 10%。
//
// 原版：`v -= Round(Real(v) / 10.0)`（`sub_53DA9`）。
//
// # ⚠️ 捨入規則有一個未確認的地方
//
// `v / 10` 的小數**會碰到 .5**（v = 5, 15, 25, …），
// 而 `Decay` 那條的 `v / 5` 碰不到，所以那邊怎麼捨入都一樣、這邊不行。
//
// Turbo Pascal 的文件說 `Round` 在正好一半時取「絕對值較大的那個」
// （half-away-from-zero），所以這裡用 `(v + 5) / 10`。
// 但 Delphi 之後改成 banker's rounding（half-to-even），
// **而我們沒有實機驗證過這一版 RTL 的行為**。
//
// 兩者的差異只在 v ≡ 5 (mod 10) 的 26 個值上，各差 1：
//
//	half-away：Round(0.5)=1  Round(1.5)=2  Round(2.5)=3
//	half-even：Round(0.5)=0  Round(1.5)=2  Round(2.5)=2
//
// 要確認就實機打一場，記錄該欄位的變化序列。
func CombatDecay(v uint8) uint8 {
	return v - uint8((uint16(v)+CombatDecayDivisor/2)/CombatDecayDivisor)
}
