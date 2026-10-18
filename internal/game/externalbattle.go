package game

import "fmt"

// These are deliberately separate from ordinary ProvinceID validity and the
// province neighbour table's SeaBorder sentinel (255). See SPEC-63.
const (
	InvasionEntryProvince  ProvinceID = 7
	ExternalInvasionSource ProvinceID = 40
)

func externalInvasionZone(m *Map, at ProvinceID) ([]CellIndex, error) {
	if m == nil || at != InvasionEntryProvince {
		return nil, fmt.Errorf("game: invalid external invasion battlefield %d", at)
	}
	bf, err := m.Battlefield(at)
	if err != nil {
		return nil, err
	}
	var zone []CellIndex
	for i := CellCount - 1; i >= 0; i-- {
		c := CellIndex(i)
		x, y := c.ColRow()
		if bf.Owner[y][x] == ExternalInvasionSource && !bf.Tiles[y][x].Kind.Blocks() {
			zone = append(zone, c)
		}
	}
	if len(zone) == 0 {
		return nil, fmt.Errorf("game: no external invasion deployment cells")
	}
	return zone, nil
}

// NewExternalBattleSim constructs on private copies. A deployment failure must
// not normalize, move or erase the player's units, even partway through a wave.
func NewExternalBattleSim(m *Map, at ProvinceID, attacker, defender []*Combatant, opt StrengthOpts) (*BattleSim, error) {
	zone, err := externalInvasionZone(m, at)
	if err != nil {
		return nil, err
	}
	if len(attacker) == 0 || len(attacker) > UnitsPerSide || len(defender) > UnitsPerSide {
		return nil, fmt.Errorf("game: invalid external invasion unit count")
	}
	bf, _ := m.Battlefield(at)
	seen := map[GeneralID]bool{}
	groups := [2][]*Combatant{}
	for side, list := range [][]*Combatant{attacker, defender} {
		for _, u := range list {
			if u == nil || u.General == 0 || seen[u.General] {
				return nil, fmt.Errorf("game: invalid or duplicate external invasion unit")
			}
			seen[u.General] = true
			if side == 1 {
				if !u.Cell.Valid() {
					return nil, fmt.Errorf("game: external invasion defender has no cell")
				}
				x, y := u.Cell.ColRow()
				if bf.Tiles[y][x].Kind.Blocks() {
					return nil, fmt.Errorf("game: external invasion defender cell is blocked")
				}
			}
			copyUnit := *u
			groups[side] = append(groups[side], &copyUnit)
		}
	}
	return newBattleSimInZone(m, at, ExternalInvasionSource, groups[0], groups[1], opt, zone)
}

// RestoreExternalBattleSim reconstructs a saved external battle, without
// deploying again. Only the explicit external entry may accept source 40.
func RestoreExternalBattleSim(m *Map, at ProvinceID, commander GeneralID, attacker, defender []*Combatant, opt StrengthOpts, generalCount int) (*BattleSim, error) {
	if _, err := externalInvasionZone(m, at); err != nil {
		return nil, err
	}
	sim, err := restoreBattleSim(m, at, ExternalInvasionSource, commander, attacker, defender, opt, generalCount, true)
	if err != nil {
		return nil, err
	}
	for _, list := range [][]*Combatant{sim.Attacker, sim.Defender} {
		for _, u := range list {
			if !u.Cell.Valid() {
				continue
			}
			x, y := u.Cell.ColRow()
			if sim.Field.Tiles[y][x].Kind.Blocks() {
				return nil, fmt.Errorf("game: external battle snapshot occupies blocked cell")
			}
		}
	}
	return sim, nil
}

// ApplyExternalBattleSettlement needs an explicit source faction, since 40 is
// not a province whose commander can be read. Ordinary settlement stays strict.
func ApplyExternalBattleSettlement(sim *BattleSim, table *ProvinceTable, generals []General, units []CombatUnit, faction GeneralID, settlement BattleSettlement) error {
	if sim == nil || sim.At != InvasionEntryProvince || sim.From != ExternalInvasionSource || !faction.Valid() || len(sim.Attacker) == 0 {
		return fmt.Errorf("game: invalid external battle settlement")
	}
	for _, u := range sim.Attacker {
		if u == nil || u.Faction != faction {
			return fmt.Errorf("game: external attackers do not belong to source faction")
		}
	}
	return applyBattleSettlementWinner(sim, table, generals, units, faction, settlement)
}
