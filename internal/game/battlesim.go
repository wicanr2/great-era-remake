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
	// AtCommander 是戰場所在省的司令（省份記錄 `+20`，`word_64944`）。
	//
	// 原版值 4（`sub_3CA09`）傳 `mode=1` 給 `sub_3BCED`；除了中心兩圈的
	// 守方候選，還會把這個司令無條件追加進候選，只要司令在場上。建立
	// `BattleSim` 的呼叫端負責從省份表填入；0 表示尚未提供這項脈絡。
	AtCommander GeneralID

	byID map[GeneralID]*Combatant
}

// NewBattleSim 建一場戰鬥並把攻方部署上場。
//
// 守方**不經過部署掃描**——原版的 `sub_41513` 只用來放攻方，
// 守方本來就在場上（`docs/re/07` §5）。所以守方的落點由呼叫端給定，
// 這裡只檢查它們合法。
func NewBattleSim(m *Map, at, from ProvinceID, attacker, defender []*Combatant, opt StrengthOpts) (*BattleSim, error) {
	// ⚠️ **Go 的零值對不上原版的欄位語意，在唯一入口一次修好。**
	//
	//	`+12` NextCell  原版哨兵是 0xFF，而 Go 零值 0 是**合法格編號**
	//	`+9`  Command   原版值域是 1–6，Go 零值 0 **不是有效值**
	//
	// 兩者的症狀不同但同源：前者讓「還沒指派」被讀成「要走到格 0」，
	// 後者讓單位被所有「只派命令 N」的行動跳過，整場空轉。
	// 都不會報錯，只會安靜地錯（2026-08-02 各被測試抓到一次）。
	//
	// ⚠️ 命令補成 2（待命）是 **remake 的補充**——原版單位進場時的命令
	// 由哪裡設定還沒解，2 是值域裡最中性的那個（不需要目標格，§15）。
	for _, u := range append(append([]*Combatant(nil), attacker...), defender...) {
		if u == nil {
			continue
		}
		if u.NextCell == 0 && u.TargetUnit == 0 {
			u.NextCell = NoCell
		}
		if u.Command == 0 {
			u.Command = BattleCmdStandby
		}
	}

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
// **參數順序：`attacker` 是發動攻擊的一方。** 原版的對應是
// `attacker → arg_10 (F)`、`target → arg_E (E)`——與直覺相反，
// 但有證據（`docs/re/09` §2）：`sub_42C8F` 讓兵種 4 對射程內的目標
// 發動遠程攻擊時，兵種 4 那個單位是 `arg_4`，一路傳下去就是 F。
//
// 這也讓「任一方是兵種 4 時 F 不受損失」變得合理：
// **遠程攻擊者不吃反擊**。
func (s *BattleSim) Engage(attacker, target *Combatant) (lossAttacker, lossTarget int, err error) {
	result, err := s.EngageWithEffects(attacker, target)
	if err != nil {
		return 0, 0, err
	}
	return result.LossAttacker, result.LossTarget, nil
}

// battleLosses 將已算出的「F 打 E」與「E 打 F」攻擊值送進原版兩條
// 戰損分支。回傳順序是 (F 損失, E 損失)，與 Engage 的公開契約一致。
func battleLosses(attacker, target *Combatant, atkOnTarget, atkOnAttacker int) (lossAttacker, lossTarget int) {
	if Lopsided(atkOnAttacker, atkOnTarget) {
		lossTarget, lossAttacker = CasualtiesRout(atkOnTarget, atkOnAttacker,
			target.Force(), attacker.Force(), target.Branch(), attacker.Branch())
		return lossAttacker, lossTarget
	}
	lossTarget, lossAttacker = CasualtiesEven(atkOnAttacker, atkOnTarget,
		target.Force(), attacker.Force(), target.Branch(), attacker.Branch())
	return lossAttacker, lossTarget
}

func (s *BattleSim) tileOf(u *Combatant) assets.Tile {
	col, row := u.Cell.ColRow()
	return s.Field.Tiles[row][col]
}

// Force 是這個單位當下的兵力（執行期記錄的 `+17`）。
func (u *Combatant) Force() uint16 { return u.Strength.Force }

// SyncForcesToGenerals 把戰場副本裡已確認的兵力（執行期 +17）同步回
// 將領記錄的 Force 欄位。
//
// 這是「戰損結果回到 remake 世界狀態」的窄入口：只寫已證實的兵力欄位，
// 不把戰場格、命令、補給或 `.DT2` 的未知區域順手當成已解。呼叫端傳入的
// slice 會原地更新，後續以 `General.Bytes`／`WriteSave` 寫回時仍會保留
// 每筆 Raw 的未解 bytes。
//
// GeneralID 是完整 MAN 槽位的 1-based 編號；無法對應或同一槽位重複出現在
// 兩方時回錯，避免把錯誤資料靜默寫進存檔。
func (s *BattleSim) SyncForcesToGenerals(generals []General) (int, error) {
	if s == nil {
		return 0, fmt.Errorf("game: nil 戰鬥不能同步兵力")
	}
	seen := make(map[GeneralID]struct{}, len(s.Attacker)+len(s.Defender))
	updated := 0
	for _, u := range s.all() {
		if u == nil || u.General == 0 {
			continue
		}
		if _, ok := seen[u.General]; ok {
			return updated, fmt.Errorf("game: 將領 %d 同時出現在戰鬥兩方", u.General)
		}
		seen[u.General] = struct{}{}
		i := int(u.General) - 1
		if i < 0 || i >= len(generals) {
			return updated, fmt.Errorf("game: 戰鬥將領 %d 超出 %d 個將領槽位", u.General, len(generals))
		}
		force := u.Force()
		if generals[i].Force == force {
			continue
		}
		generals[i].Force = force
		updated++
	}
	return updated, nil
}

// SyncBattleStatsToGenerals 把戰場副本中已閉合的將領欄位同步回策略層。
//
// 除了 Force 之外，協同／衝鋒／遠程 handler 已證實會改動經驗、體力與士氣；
// 這支只覆寫對應的已解欄位，Raw 及 `.DT2` 其他 bytes 仍交給既有 writer 保留。
func (s *BattleSim) SyncBattleStatsToGenerals(generals []General) (int, error) {
	if s == nil {
		return 0, fmt.Errorf("game: nil 戰鬥不能同步戰鬥欄位")
	}
	seen := make(map[GeneralID]struct{}, len(s.Attacker)+len(s.Defender))
	updated := 0
	for _, u := range s.all() {
		if u == nil || u.General == 0 {
			continue
		}
		if _, ok := seen[u.General]; ok {
			return updated, fmt.Errorf("game: 將領 %d 同時出現在戰鬥兩方", u.General)
		}
		seen[u.General] = struct{}{}
		i := int(u.General) - 1
		if i < 0 || i >= len(generals) {
			return updated, fmt.Errorf("game: 戰鬥將領 %d 超出 %d 個將領槽位", u.General, len(generals))
		}
		g := &generals[i]
		before := *g
		g.Force = u.Strength.Force
		// 舊呼叫端／測試 fixture 可能只建立 Force。只有規則層已
		// 完整帶入自身 General ID 時，才把其餘戰鬥欄位視為可同步，
		// 避免 Go 零值覆蓋存檔裡的已解資料。
		if u.Strength.General == u.General && u.General != 0 {
			g.AbilityA = u.Strength.Ability
			g.F19, g.F20, g.Stamina, g.F30 = u.Strength.F19, u.Strength.F20, u.Strength.F29, u.Strength.F30
			g.Experience = u.Experience
		}
		if *g != before {
			updated++
		}
	}
	return updated, nil
}

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
