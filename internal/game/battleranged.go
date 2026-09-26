package game

import (
	"fmt"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

// rangedBaseOffsets 是 WAR.EXE 的 sub_58854 使用的六個方向基準位移。
//
// 這不是用畫面座標重新推的六角範圍：IDA linear 0x588FC..0x5894D
// 逐一把 +0Ch、+1Ch、+10h、-10h、-1Ch、-0Ch 寫入暫存格，再交給
// sub_55632 產生候選格。格編號仍是 row-major 14×14；超出 0..195 的
// 位移會在候選收集時 fail-closed。
var rangedBaseOffsets = [...]int{0, 12, 28, 16, -16, -28, -12}

const maxRangedTargets = 10

// RangedTargetCells 產生一個兵種 4 單位的遠程攻擊候選格。
//
// 已證實的部分（WAR.EXE sub_42BC3、sub_58854、sub_55632）：
//   - +31 會分派到 1..6 的幾何分支（其在所有呼叫端的高階語意仍是強推論）；
//   - 來源格是森林時不產生候選；
//   - 射擊基準周圍的高山格不列入；
//   - 清單最多 10 格，並且以方向基準周圍的六角鄰接順序填入。
//
// sub_58854 的邊界分流與 sub_57B15 的畫面／音效副作用不屬於這個
// 裝置無關函式；這裡只回傳可供規則層驗證的格位，未知欄位不自行補值。
func RangedTargetCells(field *Battlefield, source CellIndex, facing uint8) []CellIndex {
	if field == nil || !source.Valid() || facing < 1 || facing > 6 {
		return nil
	}
	col, row := source.ColRow()
	if field.Tiles[row][col].Kind == assets.TileForest {
		return nil
	}

	result := make([]CellIndex, 0, maxRangedTargets)
	seen := make(map[CellIndex]struct{}, maxRangedTargets)
	appendNeighbours := func(anchor CellIndex) {
		if !anchor.Valid() {
			return
		}
		for direction := DirLowerLeft; direction <= DirUpperRight && len(result) < maxRangedTargets; direction++ {
			cell, ok := anchor.Neighbour(direction)
			if !ok || cell == source {
				continue
			}
			if _, ok := seen[cell]; ok {
				continue
			}
			cc, rr := cell.ColRow()
			if field.Tiles[rr][cc].Kind == assets.TileMountain {
				continue
			}
			seen[cell] = struct{}{}
			result = append(result, cell)
		}
	}

	base := int(source) + rangedBaseOffsets[facing]
	if base >= 0 && base < CellCount {
		appendNeighbours(CellIndex(base))
		// sub_58854 在邊界條件允許時會沿同一方向再展開一次，
		// sub_587B4 再把新格壓回同一張最多 10 格的清單。
		if len(result) < maxRangedTargets {
			if next, ok := CellIndex(base).Neighbour(HexDir(facing)); ok {
				appendNeighbours(next)
			}
		}
	}
	return result
}

// RangedTargets 回傳目前單位可命中的存活敵方單位，順序固定為
// RangedTargetCells 的候選格順序。它不改戰鬥狀態，讓鍵盤、滑鼠與觸控共用
// 同一份目標公告。
func (s *BattleSim) RangedTargets(attacker *Combatant) []*Combatant {
	if s == nil || attacker == nil || attacker.Branch() != BranchArtiller ||
		!attacker.Alive() || !attacker.Cell.Valid() {
		return nil
	}
	cells := RangedTargetCells(s.Field, attacker.Cell, attacker.Facing)
	if len(cells) == 0 {
		return nil
	}
	byCell := make(map[CellIndex]*Combatant, len(s.Defender)+len(s.Attacker))
	for _, unit := range s.all() {
		if unit == nil || !unit.Alive() || !unit.Cell.Valid() || unit.Attacking == attacker.Attacking {
			continue
		}
		byCell[unit.Cell] = unit
	}
	targets := make([]*Combatant, 0, len(cells))
	for _, cell := range cells {
		if unit := byCell[cell]; unit != nil {
			targets = append(targets, unit)
		}
	}
	return targets
}

// EngageRanged 執行 sub_57B15 的已讀出戰損切片：
// `sub_51D68` 連續呼叫三次，遠程攻擊者（F）不承受反擊，三次的目標損失
// 在最後一次才一次扣回目標（E）。
//
// 這個入口只接受已通過 RangedTargetCells 的兵種 4 攻擊；沒有把未知的
// 畫面、音效或彈藥消耗偷偷混進規則層。回傳值保持 Engage 的順序：
// (攻擊者損失, 目標損失)。
func (s *BattleSim) EngageRanged(attacker, target *Combatant) (lossAttacker, lossTarget int, err error) {
	result, err := s.EngageRangedWithEffects(attacker, target)
	if err != nil {
		return 0, 0, err
	}
	return result.LossAttacker, result.LossTarget, nil
}

// EngageRangedWithEffects 是 `sub_57B15` 的可觀察結果版本。
//
// 三次戰損仍以同一份戰前戰力計算，最後一次性扣目標兵力；結束後再套用
// 已證實的攻方體力 -5、士氣按當前值十分之一衰減，以及雙方經驗 +5。
func (s *BattleSim) EngageRangedWithEffects(attacker, target *Combatant) (BattleAttackResult, error) {
	result := BattleAttackResult{Effect: BattleEffect{Kind: BattleEffectRanged}}
	if s == nil || attacker == nil || target == nil {
		return BattleAttackResult{}, fmt.Errorf("game: 遠程交戰雙方不得為 nil")
	}
	if attacker.Branch() != BranchArtiller {
		return BattleAttackResult{}, fmt.Errorf("game: 將領 %d 不是兵種 4，不能遠程攻擊", attacker.General)
	}
	if !attacker.Alive() || !target.Alive() {
		return result, nil
	}
	if !attacker.Cell.Valid() || !target.Cell.Valid() {
		return BattleAttackResult{}, fmt.Errorf("game: 遠程交戰需要雙方都有戰場格")
	}
	if !containsCell(RangedTargetCells(s.Field, attacker.Cell, attacker.Facing), target.Cell) {
		return BattleAttackResult{}, fmt.Errorf("game: 格 %d 不在兵種 4 的射擊候選內", target.Cell)
	}

	attackerTile, targetTile := s.tileOf(attacker), s.tileOf(target)
	atkOnTarget := AttackValue(s.StrengthOf(attacker), attackerTile, attacker.Branch(), targetTile, target.Branch())
	// sub_51D68 的 arg_A=1 會把目標對攻擊者的值套上防禦地形加成。
	// 遠程攻擊者不吃反擊，但這個值仍決定「勢均力敵／一面倒」分支。
	atkOnAttacker := AttackValue(s.StrengthOf(target), targetTile, target.Branch(), attackerTile, attacker.Branch())
	atkOnAttacker *= rangedDefenceMultiplier(targetTile)

	_, oneShotTarget := rangedLossOnce(attacker, target, atkOnTarget, atkOnAttacker)
	// sub_57B15 的三次 sub_51D68 看到的是同一份兵力，最後才扣回，
	// 因此每次 lossE 相同；原版 16-bit 寫回可能過量，remake 以兵力上限
	// 夾住，避免 Go 的整數下溢把死亡單位變成大兵力。
	lossTarget := clampLoss(oneShotTarget*3, target.Force())
	target.applyLoss(lossTarget)
	result.LossTarget = lossTarget
	result.Effect.StaminaLossAttacker += applyStaminaLoss(attacker, 5)
	result.Effect.MoraleLossAttacker += applyMoraleLossRanged(attacker, &result.Effect)
	result.Effect.ExperienceGainAttacker += applyExperienceGain(attacker, 5)
	result.Effect.ExperienceGainTarget += applyExperienceGain(target, 5)
	addDied(&result.Effect, attacker)
	addDied(&result.Effect, target)
	return result, nil
}

// applyMoraleLossRanged 是 `sub_57B15` 的動態衰減：扣當前士氣的
// Round(value/10)，而不是固定 1。回傳實際扣除量並將可能的低士氣兵損記入效果。
func applyMoraleLossRanged(unit *Combatant, effect *BattleEffect) int {
	if unit == nil || effect == nil || unit.General == 0 {
		return 0
	}
	amount := roundDiv(int(unit.Strength.F30), 10)
	morale, force := applyMoraleLoss(unit, amount)
	effect.MoraleForceLossAttacker += force
	return morale
}

func rangedLossOnce(attacker, target *Combatant, atkOnTarget, atkOnAttacker int) (lossAttacker, lossTarget int) {
	if Lopsided(atkOnAttacker, atkOnTarget) {
		lossTarget, lossAttacker = CasualtiesRout(atkOnTarget, atkOnAttacker,
			target.Force(), attacker.Force(), target.Branch(), attacker.Branch())
	} else {
		lossTarget, lossAttacker = CasualtiesEven(atkOnAttacker, atkOnTarget,
			target.Force(), attacker.Force(), target.Branch(), attacker.Branch())
	}
	// sub_51B94／sub_51972 在任一方是兵種 4 時不寫 F 的損失；
	// 這裡再明示一次，避免未來替換 casualty helper 時破壞遠程契約。
	return 0, lossTarget
}

func rangedDefenceMultiplier(tile assets.Tile) int {
	if tile.HasRail() {
		return 2
	}
	switch tile.Kind {
	case assets.TileHill, assets.TilePass:
		return 2
	case assets.TileCity:
		return 3
	default:
		return 1
	}
}

func containsCell(cells []CellIndex, wanted CellIndex) bool {
	for _, cell := range cells {
		if cell == wanted {
			return true
		}
	}
	return false
}
