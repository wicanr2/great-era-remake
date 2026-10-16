package game

import "github.com/wicanr2/great-era-remake/internal/assets"

// BattleInformation is a detached, read-only view of the actual participants.
// Force retains the original branch-dependent scale. It is not a vehicle count.
type BattleInformation struct{ Sides [2]BattleSideInformation }
type BattleSideInformation struct {
	Units, Force uint32
	Branches     map[uint8]BattleBranchInformation
	Roster       []BattleUnitInformation
}
type BattleBranchInformation struct{ Units, Force uint32 }
type BattleEquipmentInformation struct {
	ID    EquipmentID
	Count uint32
}
type BattleUnitInformation struct {
	General                           GeneralID
	Strength                          StrengthInput
	Cell                              CellIndex
	Mobility, MaxMobility, Experience uint8
	Tracked                           bool
	Need, Equipped, Power             uint32
	Equipment                         []BattleEquipmentInformation
}

func InspectBattle(s *BattleSim) BattleInformation {
	var out BattleInformation
	if s == nil {
		return out
	}
	for side, group := range [2][]*Combatant{s.Attacker, s.Defender} {
		d := &out.Sides[side]
		d.Branches = map[uint8]BattleBranchInformation{}
		for _, u := range group {
			if u == nil {
				continue
			}
			row := BattleUnitInformation{General: u.General, Strength: u.Strength, Cell: u.Cell, Mobility: u.Current, MaxMobility: u.Max, Experience: u.Experience}
			if u.Alive() {
				d.Units++
				d.Force += uint32(u.Force())
				b := d.Branches[u.Branch()]
				b.Units++
				b.Force += uint32(u.Force())
				d.Branches[u.Branch()] = b
			}
			if army := s.Opts.Armory; army != nil && army.Units[u.General] != nil {
				row.Tracked = true
				row.Need = EquipmentNeed(u.Branch(), u.Force())
				rough := false
				if s.Field != nil && u.Cell.Valid() {
					col, r := u.Cell.ColRow()
					kind := s.Field.Tiles[r][col].Kind
					rough = kind == assets.TileHill || kind == assets.TileMountain
				}
				row.Power = army.Power(u.General, u.Branch(), u.Force(), rough)
				for _, m := range EquipmentCatalog() {
					if count := army.Units[u.General].Items[m.ID]; count > 0 {
						row.Equipment = append(row.Equipment, BattleEquipmentInformation{ID: m.ID, Count: count})
						row.Equipped += count
					}
				}
			}
			d.Roster = append(d.Roster, row)
		}
	}
	return out
}
