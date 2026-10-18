package game

import "fmt"

// DefenderRetreatTargets is a SPEC-63 remake rule, separate from the original
// attacker-only retreat candidates. Preserve the province's adjacency order.
func DefenderRetreatTargets(table *ProvinceTable, at ProvinceID, faction GeneralID) []ProvinceID {
	if table == nil || !faction.Valid() {
		return nil
	}
	p, err := table.At(at)
	if err != nil {
		return nil
	}
	var out []ProvinceID
	for _, id := range p.Neighbours {
		if id == at {
			continue
		}
		q, err := table.At(id)
		if err != nil || q.Commander != faction || q.InBattle() {
			continue
		}
		duplicate := false
		for _, old := range out {
			duplicate = duplicate || old == id
		}
		if !duplicate {
			out = append(out, id)
		}
	}
	return out
}

// ApplyDefenderWithdrawal moves the entire surviving loyal garrison, including
// reserves outside the ten tactical slots. Commit only after both the ordinary
// or external attacker settlement and every defender index validate on copies.
func ApplyDefenderWithdrawal(sim *BattleSim, table *ProvinceTable, generals []General, units []CombatUnit, to ProvinceID) error {
	if sim == nil || table == nil || len(generals) == 0 || len(generals) != len(units) || len(sim.Attacker) == 0 || !sim.AtCommander.Valid() {
		return fmt.Errorf("game: invalid defender withdrawal world")
	}
	target, err := table.At(sim.At)
	if err != nil || target.Commander != sim.AtCommander || !target.InBattle() {
		return fmt.Errorf("game: defender withdrawal ownership changed")
	}
	allowed := false
	for _, candidate := range DefenderRetreatTargets(table, sim.At, sim.AtCommander) {
		allowed = allowed || candidate == to
	}
	if !allowed {
		return fmt.Errorf("game: defender withdrawal destination is not a safe friendly neighbour")
	}
	var attackerFaction GeneralID
	living := false
	for _, u := range sim.Attacker {
		if u == nil || !u.Faction.Valid() || u.Faction == sim.AtCommander || u.General == 0 || int(u.General) > len(units) {
			return fmt.Errorf("game: invalid withdrawal attacker")
		}
		i := int(u.General) - 1
		if units[i].General != u.General || units[i].Faction != u.Faction || generals[i].Force != u.Force() {
			return fmt.Errorf("game: withdrawal attacker identity or force is not synchronized")
		}
		if sim.From != ExternalInvasionSource && (units[i].Province != sim.From || generals[i].Province != sim.From) {
			return fmt.Errorf("game: withdrawal attacker location changed")
		}
		if attackerFaction == 0 {
			attackerFaction = u.Faction
		}
		if u.Faction != attackerFaction {
			return fmt.Errorf("game: mixed withdrawal attacker factions")
		}
		living = living || u.Alive()
	}
	if !living {
		return fmt.Errorf("game: defeated attackers cannot capture by withdrawal")
	}
	if sim.From != ExternalInvasionSource {
		source, err := table.At(sim.From)
		if err != nil || source.Commander != attackerFaction {
			return fmt.Errorf("game: withdrawal attacker source changed")
		}
	}
	for _, u := range sim.Defender {
		if u == nil || u.General == 0 || int(u.General) > len(generals) || u.Faction != sim.AtCommander {
			return fmt.Errorf("game: invalid withdrawal defender")
		}
		i := int(u.General) - 1
		if units[i].General != u.General || units[i].Faction != sim.AtCommander || units[i].Province != sim.At || generals[i].Province != sim.At || generals[i].Force != u.Force() {
			return fmt.Errorf("game: withdrawal defender is not synchronized")
		}
	}
	copyTable := *table
	copyGenerals := append([]General(nil), generals...)
	copyUnits := append([]CombatUnit(nil), units...)
	for i, u := range copyUnits {
		if u.General != GeneralID(i+1) {
			return fmt.Errorf("game: invalid withdrawal unit index")
		}
		if u.Faction == sim.AtCommander && u.Province == sim.At && copyGenerals[i].Force > 0 {
			if copyGenerals[i].Province != sim.At {
				return fmt.Errorf("game: withdrawal garrison location mismatch")
			}
			copyGenerals[i].Province, copyUnits[i].Province = to, to
		}
	}
	settlement := BattleSettlement{Kind: BattleSettlementDecisive, Winner: BattleSideFirst}
	if sim.From == ExternalInvasionSource {
		err = ApplyExternalBattleSettlement(sim, &copyTable, copyGenerals, copyUnits, attackerFaction, settlement)
	} else {
		err = ApplyBattleSettlement(sim, &copyTable, copyGenerals, copyUnits, settlement)
	}
	if err != nil {
		return err
	}
	*table = copyTable
	copy(generals, copyGenerals)
	copy(units, copyUnits)
	return nil
}
