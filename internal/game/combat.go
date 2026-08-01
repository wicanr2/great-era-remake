package game

// 戰鬥規則。依據：docs/mechanics/30-combat.md
//
// **這一層只放已經從程式碼讀出來的規則**，沒解出來的不猜。
// 目前只有一條完整的算式（每回合衰減）與回合初始化的欄位重置。

// 每回合衰減的兩個常數，來自 WAR.EXE 的 sub_54826。
const (
	// DecayThreshold 是衰減的門檻：值大於它才會衰減。
	DecayThreshold = 40
	// DecayDivisor 是衰減的除數。原版存的是 Turbo Pascal 的 48-bit Real
	// 常數 CX:SI:DI = 0x0083/0x0000/0x2000：
	// exponent 0x83 → 2^(131-129) = 4，尾數最高 byte 0x20 → 1.25，
	// 相乘得 5.0。
	DecayDivisor = 5
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

	// Max / Current 對應 +6 與 +7。回合初始化時 `+7 = +6`——
	// 典型的「當前值 = 最大值」重置。**是什麼量未知**，
	// 機動力是最可能的候選（`3.15` 詞條 51）但沒有證據。
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

// UnitsPerSide 是每方的單位上限。
//
// 兩處獨立對上：戰鬥狀態記錄 469 B 裡每方 10 個槽（docs/re/05），
// 以及 sub_54826 的衰減迴圈跑 1..10。
const UnitsPerSide = 10
