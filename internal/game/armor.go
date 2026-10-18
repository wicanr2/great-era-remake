package game

// SPEC-65: remake design coefficients, not historical millimetres.
const EquipmentCoefficientScale uint32 = 1000

type EquipmentProtection struct {
	Armor, Piercing uint32 // thousandths of a coefficient
}

// Protection derives the current loadout. Research and warehouse stock confer
// no protection; missing equipment contributes zero to the required total.
func (a *Armory) Protection(id GeneralID, branch uint8, force uint16) EquipmentProtection {
	var out EquipmentProtection
	if a == nil || a.Units[id] == nil {
		return out
	}
	need := EquipmentNeed(branch, force)
	class, ok := EquipmentClassOf(branch)
	if !ok || need == 0 {
		return out
	}
	var count, armor, piercing uint64
	for _, model := range equipmentModels {
		if model.Class != class {
			continue
		}
		n := min(uint64(a.Units[id].Items[model.ID]), uint64(need)-count)
		count += n
		armor += n * uint64(model.Armor) * uint64(EquipmentCoefficientScale)
		piercing += n * uint64(model.Piercing) * uint64(EquipmentCoefficientScale)
	}
	out.Armor = uint32((armor + uint64(need)/2) / uint64(need))
	out.Piercing = uint32((piercing + uint64(need)/2) / uint64(need))
	return out
}

// ArmorDamage applies the adopted 50% cap, with one final half-up rounding.
// Coefficients use the same scale. The bounded integer inputs cannot overflow.
func ArmorDamage(loss uint16, piercing, armor uint32) uint16 {
	const limit = 100 * EquipmentCoefficientScale
	armor, piercing = min(armor, limit), min(piercing, limit)
	if armor == 0 || piercing >= armor {
		return loss
	}
	numerator := uint64(loss) * (uint64(armor) + uint64(piercing))
	return uint16((numerator + uint64(armor)) / (2 * uint64(armor)))
}

// protectedLosses reads both pre-loss loadouts before either ledger is changed.
func (s *BattleSim) protectedLosses(attacker, target *Combatant, lossAttacker, lossTarget int) (int, int) {
	if s.Opts.Armory == nil {
		return lossAttacker, lossTarget
	}
	a := s.Opts.Armory.Protection(attacker.General, attacker.Branch(), attacker.Force())
	t := s.Opts.Armory.Protection(target.General, target.Branch(), target.Force())
	return int(ArmorDamage(uint16(clampLoss(lossAttacker, attacker.Force())), t.Piercing, a.Armor)),
		int(ArmorDamage(uint16(clampLoss(lossTarget, target.Force())), a.Piercing, t.Armor))
}
