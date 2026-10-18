package game

import (
	"fmt"
	"sort"
)

// JapaneseAttackLimit is an adopted remake rule, not an original AI constant.
const JapaneseAttackLimit uint8 = 2

type BlitzPlan struct {
	Action AIAction
	Units  []GeneralID
}

// PlanJapaneseBlitz reads the current world without changing it or using RNG.
// The caller controls the declared-war gate and the monthly command budget.
func (w *AIWorld) PlanJapaneseBlitz(from ProvinceID, faction GeneralID, used map[GeneralID]uint8, limit int) BlitzPlan {
	if w == nil || w.Table == nil || limit < 1 || limit > ProvinceCount || int(from) > limit {
		return BlitzPlan{}
	}
	src, err := w.Table.At(from)
	if err != nil || src.Commander != faction || faction == 0 || src.InBattle() {
		return BlitzPlan{}
	}
	available := w.blitzRoster(from, used)
	if len(available) == 0 {
		return BlitzPlan{}
	}
	attack := available[:min(len(available), UnitsPerSide)]
	power, troops := 0, 0
	for _, id := range attack {
		power += w.strengthAt(int(id) - 1)
		troops += int(w.Strengths[int(id)-1].Force)
	}
	target, enemyPower := w.blitzWeakPoint(from, faction, limit)
	if target != 0 && power > 0 && power >= enemyPower && src.Ammo > 0 && src.Fuel > 0 &&
		int(src.Gold) > (BattleTurnLimit-1)*roundDiv(troops, SupplyGoldDivisor) &&
		int(src.Food) > (BattleTurnLimit-1)*roundDiv(troops, SupplyFoodDivisor) {
		return BlitzPlan{AIAction{Kind: AIAttack, From: from, To: target}, append([]GeneralID(nil), attack...)}
	}
	// A frontline holds while adjacent rear areas concentrate their armies.
	if target != 0 || len(w.blitzRoster(from, nil)) < 2 {
		return BlitzPlan{}
	}
	bestFront, bestHop, bestPower, bestDistance := ProvinceID(0), ProvinceID(0), 0, 0
	for p := ProvinceID(1); int(p) <= limit; p++ {
		front, _ := w.Table.At(p)
		if front.Commander != faction || front.InBattle() || len(w.blitzRoster(p, used)) >= UnitsPerSide {
			continue
		}
		weak, defence := w.blitzWeakPoint(p, faction, limit)
		if weak == 0 {
			continue
		}
		hop, distance := w.blitzFriendlyHop(from, p, faction, limit)
		if hop == 0 || hop == from || len(w.blitzRoster(hop, used)) >= UnitsPerSide {
			continue
		}
		if bestFront == 0 || defence < bestPower || defence == bestPower && (distance < bestDistance || distance == bestDistance && p < bestFront) {
			bestFront, bestHop, bestPower, bestDistance = p, hop, defence, distance
		}
	}
	if bestHop == 0 {
		return BlitzPlan{}
	}
	// Leave the weakest living army at the source, even if it is exhausted.
	all := w.blitzRoster(from, nil)
	keep := all[len(all)-1]
	var moving []GeneralID
	capacity := UnitsPerSide - len(w.blitzRoster(bestHop, used))
	for _, id := range available {
		if id != keep && len(moving) < capacity {
			moving = append(moving, id)
		}
	}
	if len(moving) == 0 {
		return BlitzPlan{}
	}
	return BlitzPlan{AIAction{Kind: AITransfer, From: from, To: bestHop}, moving}
}

func (w *AIWorld) blitzRoster(p ProvinceID, used map[GeneralID]uint8) []GeneralID {
	prov, err := w.Table.At(p)
	if err != nil {
		return nil
	}
	var ids []GeneralID
	for i, u := range w.Units {
		if i >= len(w.Strengths) || u.General != GeneralID(i+1) || !u.Active || u.Province != p ||
			u.Faction != prov.Commander || w.Strengths[i].Force == 0 || used[u.General] >= JapaneseAttackLimit {
			continue
		}
		ids = append(ids, u.General)
	}
	sort.Slice(ids, func(i, j int) bool {
		a, b := w.strengthAt(int(ids[i])-1), w.strengthAt(int(ids[j])-1)
		return a > b || a == b && ids[i] < ids[j]
	})
	return ids
}

func (w *AIWorld) blitzWeakPoint(from ProvinceID, faction GeneralID, limit int) (ProvinceID, int) {
	src, err := w.Table.At(from)
	if err != nil {
		return 0, 0
	}
	best, bestPower, bestFood := ProvinceID(0), 0, uint16(0)
	for _, id := range src.Neighbours {
		if int(id) > limit {
			continue
		}
		p, err := w.Table.At(id)
		if err != nil || p.Commander == 0 || p.Commander == faction || p.InBattle() {
			continue
		}
		power := 0
		for _, general := range w.blitzRoster(id, nil) {
			power += w.strengthAt(int(general) - 1)
		}
		if best == 0 || power < bestPower || power == bestPower && (p.Food < bestFood || p.Food == bestFood && id < best) {
			best, bestPower, bestFood = id, power, p.Food
		}
	}
	return best, bestPower
}

func (w *AIWorld) blitzFriendlyHop(from, to ProvinceID, faction GeneralID, limit int) (ProvinceID, int) {
	type node struct {
		id, hop  ProvinceID
		distance int
	}
	queue := []node{{id: from}}
	seen := map[ProvinceID]bool{from: true}
	for len(queue) > 0 {
		current := queue[0]
		queue = queue[1:]
		if current.id == to {
			return current.hop, current.distance
		}
		p, _ := w.Table.At(current.id)
		for _, id := range p.Neighbours {
			if seen[id] || int(id) > limit {
				continue
			}
			n, err := w.Table.At(id)
			if err != nil || n.Commander != faction || n.InBattle() || n.Flags&provinceFlagTransferBlocked != 0 {
				continue
			}
			seen[id] = true
			hop := current.hop
			if hop == 0 {
				hop = id
			}
			queue = append(queue, node{id, hop, current.distance + 1})
		}
	}
	return 0, 0
}

// SelectedCombatants validates the entire selection before returning copies.
func (w *AIWorld) SelectedCombatants(from ProvinceID, faction GeneralID, ids []GeneralID) ([]*Combatant, []int, error) {
	if w == nil || len(ids) == 0 || len(ids) > UnitsPerSide {
		return nil, nil, fmt.Errorf("game: invalid selected army size")
	}
	seen := map[GeneralID]bool{}
	var out []*Combatant
	var indices []int
	for _, id := range ids {
		i := int(id) - 1
		if seen[id] || i < 0 || i >= len(w.Units) || i >= len(w.Strengths) {
			return nil, nil, fmt.Errorf("game: invalid selected general %d", id)
		}
		u := w.Units[i]
		if u.General != id || u.Faction != faction || u.Province != from || !u.Active || w.Strengths[i].Force == 0 {
			return nil, nil, fmt.Errorf("game: selected army %d is unavailable", id)
		}
		seen[id] = true
		u.Cell, u.Max, u.Current, u.Decaying = NoCell, 12, 12, 80
		out = append(out, &Combatant{CombatUnit: u, Strength: w.Strengths[i]})
		indices = append(indices, i)
	}
	return out, indices, nil
}

// ApplyBlitzTransfer moves only the validated plan, one friendly adjacent hop.
func (w *AIWorld) ApplyBlitzTransfer(plan BlitzPlan, faction GeneralID) error {
	src, err := w.Table.At(plan.Action.From)
	if err != nil {
		return err
	}
	dst, err := w.Table.At(plan.Action.To)
	if err != nil {
		return err
	}
	adjacent := false
	for _, id := range src.Neighbours {
		adjacent = adjacent || id == plan.Action.To
	}
	if plan.Action.Kind != AITransfer || !adjacent || src.InBattle() || dst.InBattle() || src.Commander != faction || dst.Commander != faction {
		return fmt.Errorf("game: invalid blitz concentration route")
	}
	legal, err := w.PlayerTransferTargets(plan.Action.From)
	if err != nil {
		return err
	}
	allowed := false
	for _, id := range legal {
		allowed = allowed || id == plan.Action.To
	}
	if !allowed {
		return fmt.Errorf("game: concentration destination is blocked")
	}
	_, indices, err := w.SelectedCombatants(plan.Action.From, faction, plan.Units)
	if err != nil {
		return err
	}
	if len(indices) >= len(w.blitzRoster(plan.Action.From, nil)) {
		return fmt.Errorf("game: concentration must leave a garrison")
	}
	for _, i := range indices {
		w.Units[i].Province = plan.Action.To
	}
	return nil
}
