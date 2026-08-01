package game

import "github.com/wicanr2/great-era-remake/internal/assets"

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

	// Faction 對應執行期記錄的 `+14`（word）：**這個單位效忠的勢力領袖**，
	// 值是將領 ID。
	//
	// 證據是 `sub_534FF`：它拿 `+14` 直接與**省份記錄的司令欄位 `+20`**
	// 比較（`docs/spec/03` §2），三個偏移（`+20` 司令、`+22` 鄰省表、
	// `+32` 旗標）在同一段程式裡全部自洽。
	//
	// `sub_55CEC` 判敵我就是比這個欄位——不同勢力領袖就是敵人。
	Faction GeneralID

	// Province 對應 `+4`：這個將領所屬的省，**與 `MAN(N).DAT` 的 `+4`
	// 是同一個欄位**（`docs/spec/02`）。執行期記錄不是另起爐灶的佈局，
	// 至少這一格是照抄的。
	Province ProvinceID

	// Attacking 對應執行期記錄的 `+8`：**攻方 1、守方 0**。
	//
	// 兩條獨立證據（`docs/re/07` §8）：
	//   - 全程式只有 0 與 1 兩個常數寫入，而 `sub_41513`（把攻方部隊
	//     放上戰場）寫的是 1
	//   - `sub_53DA9` 用它決定去問 `byte_6B969` 的 bit 1 還是 bit 3
	//     ——那正是「顯示攻方動畫」與「顯示守方動畫」兩個獨立開關
	//
	// ⚠️ 這與 `+14`（`sub_55CEC` 用來判敵我的 word 欄位）**不是同一個東西**，
	// `+14` 的語意仍未證實。
	Attacking bool

	// Facing 對應 `+31`：砲兵的射擊朝向（1..6）。
	//
	// `NEWICON.TPC` 給砲兵準備了**六個朝向**的圖示
	// （`docs/formats/05` §3），配上六角格的六個方向，
	// 支持「`+31` 是朝向而不是射程」。第一期 274 筆全是 1。
	// ⚠️ `sub_58854`（產生格子清單的那支，494 行）還沒讀完，未證實。
	Facing uint8

	// CanCross 對應執行期記錄 `+21 == 1`：這種單位無視長城的阻隔
	// （`sub_4A583`）。`+21` 的值域含 1／4／5／6，**強證據是兵種編號**，
	// 但哪個號碼是哪個兵種沒有證據，所以這裡只放已確認的那條性質。
	// 對上原版詞表的「穿越」與「阻隔」（`2.15`）。
	CanCross bool

	// Decaying 對應 +30：大於 40 時每回合衰減 20%。語意未知。
	Decaying uint8

	// Command 對應 `+9`：**這個單位這一輪要做什麼**（值域 1–6）。
	// 常數與語意見 `battleforce.go` 的 `BattleCmd*`（`docs/re/31` §15）。
	Command uint8

	// NextCell 對應 `+12`：**朝目標走的下一格**。`NoCell`（0xFF）表示沒有。
	//
	// `sub_47EAA` 設命令 1 時一併寫它（值來自 `sub_47B6D`）；
	// `sub_3DED9` 發現命令 3 卻沒有它就把命令降回 2；
	// `sub_3B492` 拿它比對城市格來剔除「已經有人去了」的城市。
	NextCell CellIndex

	// TargetUnit 對應 `+10`：**目標單位的 ID**。0 表示沒有目標單位
	// ——`sub_47EAA` 設命令 1（前往城市）時就明確寫 0，因為那時
	// 目標是「格」不是「單位」。
	TargetUnit GeneralID

	// Flags13 是 `+13` 的完整 byte。**存整個 byte 而不是拆成具名 bool**，
	// 因為只有 bit 7 解出語意，其餘位元要原封不動帶回存檔
	// （`CLAUDE.md` §9：只蓋已解欄位，未解區域一個 byte 都不動）。
	//
	// 位元地圖見 `docs/re/31` §31。bit 2／3／6 由玩家互動層操作，語意未解。
	Flags13 uint8
}

// UnitAssignedBit 是 `+13` 的 bit 7：**這個單位這一輪已經被指派了目標**。
//
// 三條獨立證據（`docs/re/31` §31）：分支 A 的五支行動函式決定完就設它；
// `sub_3E81F` 把它和 `+10`／`+12` **一起清**（同一組狀態）；
// `sub_4732C` 讀它並「已設就跳過」。
//
// 與省份記錄 `+32` bit 2「本回合已處理過」（`docs/re/14`）是同一個手法，
// 只是一個管省份層一個管單位層。
const UnitAssignedBit uint8 = 0x80

// Assigned 回報這個單位這一輪是否已經被指派了目標。
func (u *CombatUnit) Assigned() bool { return u.Flags13&UnitAssignedBit != 0 }

// AssignTo 把目標與下一跳寫進去並立起「已指派」——三個欄位一起動，
// 對應分支 A 行動函式的收尾。
func (u *CombatUnit) AssignTo(target GeneralID, next CellIndex) {
	u.TargetUnit = target
	u.NextCell = next
	u.Flags13 |= UnitAssignedBit
}

// ClearAssignment 是 `sub_3E81F`：把 `+10`／`+12`／`+13` bit 7 一起清掉。
//
// ⚠️ 只清 bit 7，其餘位元不動——那些是別層在用的。
func (u *CombatUnit) ClearAssignment() {
	u.TargetUnit = 0
	u.NextCell = NoCell
	u.Flags13 &^= UnitAssignedBit
}

// ResetToStandby 是 `sub_3B15E`（`docs/re/31` §33）：把單位打回待命。
//
//	+9  = 2（待命）    +12 = 0xFF（沒有下一跳）    +10 = 0（沒有目標）
//
// ⭐ 與 `ClearAssignment` 是一對：兩者都清 `+10`／`+12`，但
// `ClearAssignment`（`sub_3E81F`）清 `+13` bit 7 而**不動命令**，
// 這一支改設命令為待命而**不動 `+13`**。原版就是這樣分工的，別合併。
func (u *CombatUnit) ResetToStandby() {
	u.Command = BattleCmdStandby
	u.NextCell = NoCell
	u.TargetUnit = 0
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

// 地形對戰力的加成。倍率出自 `sub_51D68`（`docs/re/08` §3a），
// **這是社群說法「地形影響防禦」的第一份程式碼證據**。
//
// 原版的寫法是自我相加：`v += v` 就是 ×2，城市那條先算 `v×2` 再加回去
// 所以是 ×3。
const (
	// TerrainBonusRail 是有鐵路的格。鐵路先問，問到就不看地形。
	TerrainBonusRail = 2
	// TerrainBonusHillPass 是丘陵與關口。
	TerrainBonusHillPass = 2
	// TerrainBonusCity 是城市——全遊戲最高的加成。
	TerrainBonusCity = 3
)

// TerrainBonus 回傳站在這一格的戰力倍率。
//
// 依 `sub_51D68` 的順序：先問鐵路，再看地形，其餘一律 1 倍。
//
// ⚠️ 這個加成有一個開關（`sub_51D68` 的 `arg_A`）。
// `sub_5301B` 那條呼叫路徑傳 0，**不套用**；另一個呼叫端
// （`sub_483A5`）還沒讀。所以「什麼時候算地形」尚未確定。
func TerrainBonus(t assets.Tile) int {
	if t.HasRail() {
		return TerrainBonusRail
	}
	switch t.Kind {
	case assets.TileHill, assets.TilePass:
		return TerrainBonusHillPass
	case assets.TileCity:
		return TerrainBonusCity
	}
	return 1
}

// RoutThreshold 是「一面倒」的門檻倍率。
//
// `sub_51D68` 算完雙方戰力後分兩條路：任一方的戰力 ×4 還不到對方，
// 就走 `sub_51972`；雙方都在 4 倍之內才走 `sub_51B94`。
// 兩支的戰損算法不同，**都還沒讀**。
//
// 詞表的「崩潰」「敗逃」「打敗」（`2.15`）很可能對應一面倒那條，
// 但沒有證據。
const RoutThreshold = 4

// Lopsided 回報兩個戰力值是不是差到「一面倒」的程度，語意照 `sub_51D68`。
func Lopsided(a, b int) bool {
	return a*RoutThreshold <= b || b*RoutThreshold <= a
}
