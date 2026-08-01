package game

import (
	"fmt"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

// 一場戰鬥的完整流程，把已解出的規則接起來：
// 部署 → 回合初始化 → 移動 → 交戰 → 戰損 → 判定結束。
//
// **這一層不是新規則，是既有規則的組裝。** 每一步都指向它的出處；
// 沒有出處的環節（例如「攻擊時機由誰決定」）留成呼叫端的責任，
// 不在這裡發明。

// Combatant 是戰場上的一個單位，把規則層需要的欄位湊在一起。
type Combatant struct {
	CombatUnit
	Strength StrengthInput // 算戰力要用的欄位
}

// BattleSim 是一場戰鬥的執行期狀態。
//
// 對應原版的三組全域：`word_62A8`（佔用表）、`0x750`／`0x764`
// （兩方各 10 個部隊槽）、以及 `byte_6FFC4` 指到的當前省戰場。
type BattleSim struct {
	Field    *Battlefield // 戰場（地物 + 分區）
	Occ      Occupancy    // 每格的佔用者
	Attacker []*Combatant // 攻方，最多 UnitsPerSide
	Defender []*Combatant // 守方
	Opts     StrengthOpts // 難度與階段開關

	// From 是攻方的來源省，決定進場區（`.DT2` 的 +468）。
	From ProvinceID
	// At 是戰場所在的省（守方的省，`byte_6FFC4`）。
	At ProvinceID

	byID map[GeneralID]*Combatant
}

// NewBattleSim 建一場戰鬥並把攻方部署上場。
//
// 守方**不經過部署掃描**——原版的 `sub_41513` 只用來放攻方，
// 守方本來就在場上（`docs/re/07` §5）。所以守方的落點由呼叫端給定，
// 這裡只檢查它們合法。
func NewBattleSim(m *Map, at, from ProvinceID, attacker, defender []*Combatant, opt StrengthOpts) (*BattleSim, error) {
	bf, err := m.Battlefield(at)
	if err != nil {
		return nil, err
	}
	if len(attacker) > UnitsPerSide || len(defender) > UnitsPerSide {
		return nil, fmt.Errorf("game: 每方最多 %d 個部隊，得到 攻 %d／守 %d",
			UnitsPerSide, len(attacker), len(defender))
	}
	s := &BattleSim{
		Field: bf, Attacker: attacker, Defender: defender,
		Opts: opt, From: from, At: at,
		byID: make(map[GeneralID]*Combatant),
	}

	// 守方先就位——它們的 Cell 由呼叫端設好。
	for _, u := range defender {
		u.Attacking = false
		if !u.Cell.Valid() {
			return nil, fmt.Errorf("game: 守方單位 %d 沒有落點", u.General)
		}
		if s.Occ[u.Cell] != 0 {
			return nil, fmt.Errorf("game: 格 %d 被兩個守方單位佔住", u.Cell)
		}
		s.Occ[u.Cell] = u.General
		s.byID[u.General] = u
	}

	// 攻方走部署掃描：從 cell 195 往下，挑進場區裡旁邊沒敵人的空格。
	zone, err := m.DeployZone(at, from)
	if err != nil {
		return nil, err
	}
	for _, u := range attacker {
		u.Attacking = true
		cell, ok := s.Occ.Deploy(zone, u.General, s.isEnemyOf(true))
		if !ok {
			return nil, fmt.Errorf("game: 進場區放不下單位 %d", u.General)
		}
		u.Cell = cell
		s.byID[u.General] = u
	}
	return s, nil
}

// isEnemyOf 產生「對某一方而言誰是敵人」的判斷，給部署掃描用。
func (s *BattleSim) isEnemyOf(attacking bool) func(GeneralID) bool {
	return func(id GeneralID) bool {
		u := s.byID[id]
		return u != nil && u.Attacking != attacking
	}
}

// Unit 依將領 ID 取單位。
func (s *BattleSim) Unit(id GeneralID) *Combatant { return s.byID[id] }

// BeginTurn 套用回合開始的重置（`sub_5446D`／`sub_543C2`）：
// 機動力回滿、標記可行動。
func (s *BattleSim) BeginTurn() {
	for _, u := range s.all() {
		u.BeginTurn()
	}
}

// EndTurn 套用回合結束的衰減（`sub_54826`）。
func (s *BattleSim) EndTurn() {
	for _, u := range s.all() {
		u.EndTurn()
	}
}

func (s *BattleSim) all() []*Combatant {
	return append(append([]*Combatant{}, s.Attacker...), s.Defender...)
}

// Move 把單位往某方向移一格，扣機動力。
func (s *BattleSim) Move(id GeneralID, d HexDir) (CellIndex, error) {
	u := s.byID[id]
	if u == nil {
		return NoCell, fmt.Errorf("game: 沒有單位 %d", id)
	}
	return s.Occ.Move(s.Field, &u.CombatUnit, d)
}

// StrengthOf 算某個單位當下的戰力（`sub_5A0B9`）。
func (s *BattleSim) StrengthOf(u *Combatant) int {
	in := u.Strength
	in.Force = u.Force()
	in.Branch = u.Branch()
	in.Faction = u.Faction
	in.General = u.General
	return Strength(in, s.Opts)
}

// Engage 讓 a 與 b 交戰一次，回傳雙方的兵力損失。
//
// 流程照 `sub_51D68` → `sub_51B94`／`sub_51972`（`docs/re/08`）：
//
//  1. 兩邊各算戰力（`sub_5A0B9`）
//  2. 各自套地形係數，得到「施加於對方的攻擊值」
//     ——**攻方看自己站的格、守方看自己站的格**
//  3. 差距 4 倍以上走一面倒，否則走勢均力敵
//  4. 兩邊的兵力各扣一次（`sub_530B4`）
//
// ⚠️ **什麼時候該呼叫這支，還沒解出來。** 原版的攻擊時機在
// `sub_53DA9` 那條鏈上，而那支的入口條件（`+8` 的狀態機、七項選單）
// 尚未讀完。這裡只負責「一次交戰算多少」。
//
// ⚠️ **`a` 對應原版的 `arg_E`、`b` 對應 `arg_10`，但哪一個是攻方沒有確認。**
// `sub_530B4` 的呼叫端還沒讀。這個對應會影響兵種 4 的特例方向
// （`docs/playtest/04` §3）：實跑起來變成「兵種 4 打不動人」，
// 如果原版的 A 其實是守方，那條特例讀起來會合理得多。
// **在讀完呼叫端之前不要改**——訂正比新結論需要更硬的證據。
func (s *BattleSim) Engage(a, b *Combatant) (lossA, lossB int, err error) {
	if a == nil || b == nil {
		return 0, 0, fmt.Errorf("game: 交戰雙方不得為 nil")
	}
	if !Adjacent(a.Cell, b.Cell) {
		return 0, 0, fmt.Errorf("game: 格 %d 與 %d 不相鄰，打不到", a.Cell, b.Cell)
	}
	ta, tb := s.tileOf(a), s.tileOf(b)

	// 施加於對方的攻擊值。
	atkOnB := AttackValue(s.StrengthOf(a), ta, a.Branch(), tb, b.Branch())
	atkOnA := AttackValue(s.StrengthOf(b), tb, b.Branch(), ta, a.Branch())

	if Lopsided(atkOnA, atkOnB) {
		// 一面倒：損失由「打在自己身上的攻擊值」決定。
		lossA, lossB = CasualtiesRout(atkOnA, atkOnB,
			a.Force(), b.Force(), a.Branch(), b.Branch())
	} else {
		// 勢均力敵：pct 用「各自打出去的攻擊值」之比，
		// **不是打在身上的**——傳反會變成「打得越兇損失越大」。
		lossA, lossB = CasualtiesEven(atkOnB, atkOnA,
			a.Force(), b.Force(), a.Branch(), b.Branch())
	}
	a.applyLoss(lossA)
	b.applyLoss(lossB)
	return lossA, lossB, nil
}

func (s *BattleSim) tileOf(u *Combatant) assets.Tile {
	col, row := u.Cell.ColRow()
	return s.Field.Tiles[row][col]
}

// Force 是這個單位當下的兵力（執行期記錄的 `+17`）。
func (u *Combatant) Force() uint16 { return u.Strength.Force }

// Branch 是兵種（`+21`）。
func (u *Combatant) Branch() uint8 { return u.Strength.Branch }

// applyLoss 扣兵力，扣到 0 為止。原版是 `sub [di+7A8Eh], cx`，
// 沒有下界檢查——但兩支戰損函式都保證損失不超過兵力（`docs/re/08` §5b），
// 所以不會變成負的。這裡還是夾一次，讓不變式明顯。
func (u *Combatant) applyLoss(n int) {
	if n <= 0 {
		return
	}
	if n >= int(u.Strength.Force) {
		u.Strength.Force = 0
		return
	}
	u.Strength.Force -= uint16(n)
}

// Alive 回報單位還在不在場上。兵力歸零就是陣亡
// （`sub_54826` 把槽位設 0，`docs/mechanics/30-combat.md` §3g）。
func (u *Combatant) Alive() bool { return u.Strength.Force > 0 }

// Remove 把陣亡的單位從佔用表清掉。
func (s *BattleSim) Remove(id GeneralID) {
	u := s.byID[id]
	if u == nil {
		return
	}
	if u.Cell.Valid() && s.Occ[u.Cell] == id {
		s.Occ[u.Cell] = 0
	}
	u.Cell = NoCell
	delete(s.byID, id)
}

// Sweep 清掉所有兵力歸零的單位，回傳被清掉的 ID。
func (s *BattleSim) Sweep() []GeneralID {
	var dead []GeneralID
	for _, u := range s.all() {
		if !u.Alive() && u.Cell.Valid() {
			dead = append(dead, u.General)
		}
	}
	for _, id := range dead {
		s.Remove(id)
	}
	return dead
}

// Over 回報戰鬥是不是結束了：任一方全滅。
func (s *BattleSim) Over() (over bool, attackerWon bool) {
	a, d := 0, 0
	for _, u := range s.Attacker {
		if u.Alive() {
			a++
		}
	}
	for _, u := range s.Defender {
		if u.Alive() {
			d++
		}
	}
	switch {
	case d == 0 && a > 0:
		return true, true
	case a == 0:
		return true, false
	}
	return false, false
}
