package game

import "fmt"

// These are the three accepted SPEC-63 design groups, not a historic order of
// battle. Date selection and officer identities belong to the invasion event.
type InvasionEra uint8

const (
	InvasionEarly InvasionEra = iota + 1
	InvasionMid
	InvasionLate
)

type InvasionSupply struct{ Gold, Food, Ammo, Fuel uint32 }

// Formation is a stable plan ordinal, never a legacy GeneralID or MAN slot.
type InvasionFormation struct {
	Formation              uint8
	Branch                 uint8
	Force                  uint16
	Ability, Skill, Arming uint8
	Stamina, Morale        uint8
	Equipment              EquipmentID
}

type InvasionWavePlan struct {
	MonthOffset uint16
	Formations  []InvasionFormation
	Equipment   map[EquipmentID]uint32
	Supply      InvasionSupply
}

type InvasionPlan struct {
	Era   InvasionEra
	Waves [3]InvasionWavePlan
}

// FixedInvasionPlan returns fresh, finite data only. It cannot spawn an army or
// credit supplies. No player-strength input exists in the approved fixed policy.
func FixedInvasionPlan(era InvasionEra) (InvasionPlan, error) {
	var models [3]EquipmentID
	switch era {
	case InvasionEarly:
		models = [3]EquipmentID{"rifle-old", "mountain-basic", "legacy-armor"}
	case InvasionMid:
		models = [3]EquipmentID{"rifle-modern", "mountain-improved", "light-tank"}
	case InvasionLate:
		models = [3]EquipmentID{"rifle-modern", "field-howitzer", "armor-improved"}
	default:
		return InvasionPlan{}, fmt.Errorf("game: invalid invasion era %d", era)
	}
	branches := [3][]uint8{{BranchInfantry, BranchInfantry, BranchInfantry, BranchInfantry, BranchArtiller, BranchArmour}, {BranchInfantry, BranchArtiller}, {BranchInfantry, BranchArmour}}
	plan := InvasionPlan{Era: era}
	var formation uint8
	for i, list := range branches {
		wave := InvasionWavePlan{MonthOffset: uint16(i * 3), Equipment: map[EquipmentID]uint32{}}
		// The accepted total is apportioned 6:2:2; every batch stays bounded.
		n := uint32(len(list))
		wave.Supply = InvasionSupply{Gold: 2400 * n, Food: 3200 * n, Ammo: 2400 * n, Fuel: 1200 * n}
		for _, branch := range list {
			formation++
			force, model := uint16(12000), models[0]
			if branch == BranchArtiller {
				force, model = 1000, models[1]
			}
			if branch == BranchArmour {
				force, model = 120, models[2]
			}
			unit := InvasionFormation{Formation: formation, Branch: branch, Force: force, Ability: 85, Skill: 80, Arming: 80, Stamina: 90, Morale: 90, Equipment: model}
			wave.Formations = append(wave.Formations, unit)
			wave.Equipment[model] += EquipmentNeed(branch, force)
		}
		plan.Waves[i] = wave
	}
	return plan, nil
}
