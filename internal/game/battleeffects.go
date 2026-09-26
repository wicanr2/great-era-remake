package game

import (
	"fmt"
)

// BattleEffectKind 是戰鬥 helper 可觀察到的副作用類型。
//
// 名稱只描述 remake 規則層的事件，不取代原版函式名，也不宣稱畫面／音效等價。
type BattleEffectKind string

const (
	BattleEffectStandard      BattleEffectKind = "standard"
	BattleEffectCoordinated   BattleEffectKind = "coordinated"
	BattleEffectCavalryCharge BattleEffectKind = "cavalry-charge"
	BattleEffectRanged        BattleEffectKind = "ranged"
	BattleEffectSpecialVisual BattleEffectKind = "special-visual-only"
)

// BattleCellMotion 記錄衝鋒 helper 的暫時格位搬動畫面。
//
// `sub_58449` 會把目標格暫時寫成攻擊者格，再於畫面 frame 2 還原；目前沒有
// 證據表示這是永久佔位變更，所以只暴露事件，不直接改 `Combatant.Cell`。
type BattleCellMotion struct {
	Unit GeneralID
	From CellIndex
	To   CellIndex
}

// BattleUnitLoss 是協同攻擊中支援單位的個別損失。
type BattleUnitLoss struct {
	General GeneralID
	Loss    int
}

// BattleEffect 彙總一次 handler 的已證實副作用。
type BattleEffect struct {
	Kind BattleEffectKind

	ExperienceGainAttacker  int
	ExperienceGainTarget    int
	StaminaLossAttacker     int
	StaminaLossTarget       int
	MoraleLossAttacker      int
	MoraleLossTarget        int
	MoraleForceLossAttacker int
	MoraleForceLossTarget   int

	Supporters            []GeneralID
	Died                  []GeneralID
	SkipSupportAfterDeath bool
	CellMotions           []BattleCellMotion
	VisualOnly            bool
}

// BattleAttackResult 是可供 UI、AI、存檔同步共用的攻擊結果。
type BattleAttackResult struct {
	LossAttacker  int
	LossTarget    int
	SupportLosses []BattleUnitLoss
	Effect        BattleEffect
}

func (s *BattleSim) validatePair(attacker, target *Combatant) error {
	if s == nil || attacker == nil || target == nil {
		return fmt.Errorf("game: 交戰雙方不得為 nil")
	}
	if s.Field == nil {
		return fmt.Errorf("game: 戰鬥缺少戰場")
	}
	if !attacker.Cell.Valid() || !target.Cell.Valid() || !Adjacent(attacker.Cell, target.Cell) {
		return fmt.Errorf("game: 格 %d 與 %d 不相鄰，打不到", attacker.Cell, target.Cell)
	}
	return nil
}

// addDied 只把同一個單位加入一次死亡清單。
func addDied(effect *BattleEffect, unit *Combatant) {
	if effect == nil || unit == nil || unit.Alive() || unit.General == 0 {
		return
	}
	for _, id := range effect.Died {
		if id == unit.General {
			return
		}
	}
	effect.Died = append(effect.Died, unit.General)
}

func applyStaminaLoss(unit *Combatant, amount int) int {
	if unit == nil || amount <= 0 || unit.General == 0 {
		return 0
	}
	before := int(unit.Strength.F29)
	if before <= 0 {
		return 0
	}
	if amount > before {
		amount = before
	}
	unit.Strength.F29 = uint8(before - amount)
	return amount
}

// applyMoraleLoss mirrors sub_51F19. The extra force loss is deliberately returned
// separately because it is a confirmed side effect, not part of sub_51D68 output.
func applyMoraleLoss(unit *Combatant, amount int) (moraleLoss, forceLoss int) {
	if unit == nil || amount <= 0 || unit.General == 0 {
		return 0, 0
	}
	before := int(unit.Strength.F30)
	if before <= 0 {
		return 0, 0
	}
	if amount > before {
		amount = before
	}
	unit.Strength.F30 = uint8(before - amount)
	if unit.Strength.F30 <= 10 && unit.Force() > 0 {
		forceLoss = roundDiv(int(unit.Force()), 3)
		unit.applyLoss(forceLoss)
	}
	return amount, forceLoss
}

// applyExperienceGain mirrors sub_5A3B2: +3 reaches 100, increments +0 once,
// then resets; an already maxed Ability keeps the experience sentinel at 100.
func applyExperienceGain(unit *Combatant, amount int) int {
	if unit == nil || amount <= 0 || unit.General == 0 {
		return 0
	}
	current := int(unit.Experience) + amount
	if current < 100 {
		unit.Experience = uint8(current)
		return amount
	}
	if unit.Strength.Ability < 100 {
		unit.Strength.Ability++
		unit.Experience = 0
	} else {
		unit.Experience = 100
	}
	return amount
}

func preparePair(attacker, target *Combatant, effect *BattleEffect,
	experience, stamina, morale int) (extraAttacker, extraTarget int) {
	if effect != nil {
		effect.ExperienceGainAttacker += applyExperienceGain(attacker, experience)
		effect.ExperienceGainTarget += applyExperienceGain(target, experience)
		effect.StaminaLossAttacker += applyStaminaLoss(attacker, stamina)
		effect.StaminaLossTarget += applyStaminaLoss(target, stamina)
		var m, f int
		m, f = applyMoraleLoss(attacker, morale)
		effect.MoraleLossAttacker += m
		effect.MoraleForceLossAttacker += f
		extraAttacker += f
		m, f = applyMoraleLoss(target, morale)
		effect.MoraleLossTarget += m
		effect.MoraleForceLossTarget += f
		extraTarget += f
		return extraAttacker, extraTarget
	}
	applyExperienceGain(attacker, experience)
	applyExperienceGain(target, experience)
	applyStaminaLoss(attacker, stamina)
	applyStaminaLoss(target, stamina)
	_, extraAttacker = applyMoraleLoss(attacker, morale)
	_, extraTarget = applyMoraleLoss(target, morale)
	return extraAttacker, extraTarget
}

func (s *BattleSim) pairLosses(attacker, target *Combatant) (lossAttacker, lossTarget int) {
	ta, tb := s.tileOf(attacker), s.tileOf(target)
	atkOnTarget := AttackValue(s.StrengthOf(attacker), ta, attacker.Branch(), tb, target.Branch())
	atkOnAttacker := AttackValue(s.StrengthOf(target), tb, target.Branch(), ta, attacker.Branch())
	return battleLosses(attacker, target, atkOnTarget, atkOnAttacker)
}

// EngageWithEffects 執行 `sub_5301B`／`sub_530B4` 的已閉合單次正規切片。
func (s *BattleSim) EngageWithEffects(attacker, target *Combatant) (BattleAttackResult, error) {
	if err := s.validatePair(attacker, target); err != nil {
		return BattleAttackResult{}, err
	}
	result := BattleAttackResult{Effect: BattleEffect{Kind: BattleEffectStandard}}
	if !attacker.Alive() || !target.Alive() {
		return result, nil
	}
	extraA, extraT := preparePair(attacker, target, &result.Effect, 5, 1, 1)
	rawA, rawT := s.pairLosses(attacker, target)
	attacker.applyLoss(rawA)
	target.applyLoss(rawT)
	result.LossAttacker = rawA + extraA
	result.LossTarget = rawT + extraT
	addDied(&result.Effect, attacker)
	addDied(&result.Effect, target)
	result.Effect.SkipSupportAfterDeath = len(result.Effect.Died) > 0
	return result, nil
}

// EngageCoordinated 接上 `sub_53111` 的六鄰支援與損失分攤窄切片。
func (s *BattleSim) EngageCoordinated(attacker, target *Combatant) (BattleAttackResult, error) {
	if err := s.validatePair(attacker, target); err != nil {
		return BattleAttackResult{}, err
	}
	result := BattleAttackResult{Effect: BattleEffect{Kind: BattleEffectCoordinated}}
	if !attacker.Alive() || !target.Alive() {
		return result, nil
	}
	supporters := s.coordinationSupporters(attacker, target)
	for _, support := range supporters {
		result.Effect.Supporters = append(result.Effect.Supporters, support.General)
	}
	supportCount := len(supporters)
	extraA, extraT := preparePair(attacker, target, &result.Effect, 5, 1, 1)
	rawA, rawT := s.pairLosses(attacker, target)
	if supportCount > 0 {
		rawT = roundDiv(rawT, supportCount)
	}
	attacker.applyLoss(rawA)
	target.applyLoss(rawT)
	result.LossAttacker = rawA + extraA
	result.LossTarget = rawT + extraT
	addDied(&result.Effect, attacker)
	addDied(&result.Effect, target)
	if len(result.Effect.Died) > 0 {
		result.Effect.SkipSupportAfterDeath = true
		return result, nil
	}

	for _, support := range supporters {
		if !attacker.Alive() || !support.Alive() {
			continue
		}
		// The helper is called as (support, attacker), therefore the support
		// unit receives the first loss and the attacker the second.
		extraSupport, extraAttacker := preparePair(support, attacker, &result.Effect, 5, 1, 1)
		rawSupport, rawAttacker := s.pairLosses(support, attacker)
		rawSupport = roundDiv(rawSupport, supportCount)
		support.applyLoss(rawSupport)
		attacker.applyLoss(rawAttacker)
		result.SupportLosses = append(result.SupportLosses, BattleUnitLoss{
			General: support.General,
			Loss:    rawSupport + extraSupport,
		})
		result.LossAttacker += rawAttacker + extraAttacker
		addDied(&result.Effect, support)
		addDied(&result.Effect, attacker)
		if len(result.Effect.Died) > 0 {
			result.Effect.SkipSupportAfterDeath = true
			break
		}
	}
	return result, nil
}

func (s *BattleSim) coordinationSupporters(attacker, target *Combatant) []*Combatant {
	if s == nil || attacker == nil || target == nil || !attacker.Cell.Valid() {
		return nil
	}
	byCell := make(map[CellIndex]*Combatant, len(s.Attacker)+len(s.Defender))
	for _, unit := range s.all() {
		if unit != nil && unit.Cell.Valid() && unit.Alive() {
			byCell[unit.Cell] = unit
		}
	}
	var out []*Combatant
	for direction := DirLowerLeft; direction <= DirUpperRight; direction++ {
		cell, ok := attacker.Cell.Neighbour(direction)
		if !ok {
			continue
		}
		unit := byCell[cell]
		if unit == nil || unit == target || unit.General == 0 || unit.Branch() == BranchArtiller {
			continue
		}
		if unit.Attacking != target.Attacking || unit.Faction != target.Faction {
			continue
		}
		out = append(out, unit)
		if len(out) == 6 {
			break
		}
	}
	return out
}

// EngageCavalrySpecial 接上 `sub_58449`／`sub_58613` 的衝鋒窄切片。
func (s *BattleSim) EngageCavalrySpecial(attacker, target *Combatant) (BattleAttackResult, error) {
	if err := s.validatePair(attacker, target); err != nil {
		return BattleAttackResult{}, err
	}
	if attacker.Branch() != BranchCavalry || (target.Branch() != BranchInfantry && target.Branch() != BranchCavalry) {
		return BattleAttackResult{}, fmt.Errorf("game: 目前雙方兵種不符合衝鋒 gate")
	}
	if !attacker.Alive() || !target.Alive() {
		return BattleAttackResult{Effect: BattleEffect{Kind: BattleEffectCavalryCharge}}, nil
	}
	targetTile := s.tileOf(target)
	if targetTile.Kind.Blocks() {
		return BattleAttackResult{}, fmt.Errorf("game: 長城段不允許衝鋒")
	}
	result := BattleAttackResult{Effect: BattleEffect{Kind: BattleEffectCavalryCharge}}
	passes := int(target.Strength.Ability)/30 + int(target.Strength.F30)/40 + 1
	for pass := 0; pass < passes; pass++ {
		if !attacker.Alive() || !target.Alive() {
			break
		}
		rawA, rawT := s.pairLosses(attacker, target)
		chargeTarget := roundDiv(rawT*2, 3)
		attacker.applyLoss(rawA)
		target.applyLoss(chargeTarget)
		result.LossAttacker += rawA
		result.LossTarget += chargeTarget
		result.Effect.CellMotions = append(result.Effect.CellMotions, BattleCellMotion{
			Unit: target.General, From: target.Cell, To: attacker.Cell,
		})
		addDied(&result.Effect, attacker)
		addDied(&result.Effect, target)
		if attacker.Force() <= 3 || target.Force() <= 3 {
			break
		}
	}
	// sub_58613 applies these after the repeated charge passes.
	result.Effect.StaminaLossTarget += applyStaminaLoss(target, 2)
	result.Effect.StaminaLossAttacker += applyStaminaLoss(attacker, 5)
	m, f := applyMoraleLoss(target, 3)
	result.Effect.MoraleLossTarget += m
	result.Effect.MoraleForceLossTarget += f
	result.LossTarget += f
	m, f = applyMoraleLoss(attacker, 10)
	result.Effect.MoraleLossAttacker += m
	result.Effect.MoraleForceLossAttacker += f
	result.LossAttacker += f
	result.Effect.ExperienceGainTarget += applyExperienceGain(target, 15)
	result.Effect.ExperienceGainAttacker += applyExperienceGain(attacker, 15)
	addDied(&result.Effect, attacker)
	addDied(&result.Effect, target)
	return result, nil
}

// ResolveBattleAttack 是玩家與 AI 共用的 handler 分派入口。
// 目標兵種 4／5 的 `sub_4B854`／`sub_4BCBE` 分支目前只有已證實的畫面／音效
// 控制流，因此以零損失事件表示，避免把未知副作用誤寫成兵力規則。
func (s *BattleSim) ResolveBattleAttack(attacker, target *Combatant) (BattleAttackResult, error) {
	if s == nil || attacker == nil || target == nil {
		return BattleAttackResult{}, fmt.Errorf("game: 交戰雙方不得為 nil")
	}
	if target.Branch() == BranchArtiller || target.Branch() == BranchArmour {
		// `sub_4B854`／`sub_4BCBE` 的窄證據只閉合了非砲兵攻擊者的
		// 畫面／音效分支；不能讓兵種 4 的遠程路徑被這個零損失事件吃掉。
		if attacker.Branch() == BranchArtiller {
			// 繼續走下方的遠程／一般判斷。
		} else {
			if err := s.validatePair(attacker, target); err != nil {
				return BattleAttackResult{}, err
			}
			return BattleAttackResult{Effect: BattleEffect{
				Kind: BattleEffectSpecialVisual, VisualOnly: true,
			}}, nil
		}
	}
	if attacker.Branch() == BranchArtiller && !Adjacent(attacker.Cell, target.Cell) {
		return s.EngageRangedWithEffects(attacker, target)
	}
	if attacker.Branch() == BranchCavalry &&
		(target.Branch() == BranchInfantry || target.Branch() == BranchCavalry) {
		if err := s.validatePair(attacker, target); err != nil {
			return BattleAttackResult{}, err
		}
		if !s.tileOf(target).Kind.Blocks() {
			return s.EngageCavalrySpecial(attacker, target)
		}
	}
	// 呼叫端目前沒有另一個可見的「協同」按鍵；只要已證實的六鄰
	// 支援候選存在，就共用原版 `sub_53111` 窄切片。這是 remake 的
	// 分派補完，不能當成原版 UI 選單已閉合。
	if len(s.coordinationSupporters(attacker, target)) > 0 {
		return s.EngageCoordinated(attacker, target)
	}
	return s.EngageWithEffects(attacker, target)
}

// RetreatCandidateSource 分辨原版固定樣本與 remake 泛化候選。
type RetreatCandidateSource string

const (
	RetreatCandidateOriginalSample    RetreatCandidateSource = "original-sample"
	RetreatCandidateNeighbourFallback RetreatCandidateSource = "province-neighbour-fallback"
)

// RetreatCandidatesResult 是完整輸入外殼需要的撤退候選與證據標籤。
type RetreatCandidatesResult struct {
	Candidates []ProvinceID
	Source     RetreatCandidateSource
}

// RetreatCandidates 保留省份表的鄰接順序，提供其他交戰組合可玩的 remake fallback。
// 目前只有 18←19 的順序有原版 DOSBox oracle；其餘值不是 parity 宣稱。
func RetreatCandidates(table *ProvinceTable, at, from ProvinceID) RetreatCandidatesResult {
	if table == nil || !at.Valid() || !from.Valid() {
		return RetreatCandidatesResult{}
	}
	prov, err := table.At(at)
	if err != nil {
		return RetreatCandidatesResult{}
	}
	var result []ProvinceID
	for _, candidate := range prov.Neighbours {
		if candidate == 0 || candidate == SeaBorder || !candidate.Valid() {
			continue
		}
		if candidate == from {
			result = append(result, candidate)
			continue
		}
		if _, err := table.At(candidate); err == nil {
			result = append(result, candidate)
		}
	}
	if len(result) == 0 {
		return RetreatCandidatesResult{}
	}
	if at == 18 && from == 19 && len(result) == 4 &&
		result[0] == 19 && result[1] == 26 && result[2] == 14 && result[3] == 17 {
		return RetreatCandidatesResult{Candidates: result, Source: RetreatCandidateOriginalSample}
	}
	return RetreatCandidatesResult{Candidates: result, Source: RetreatCandidateNeighbourFallback}
}
