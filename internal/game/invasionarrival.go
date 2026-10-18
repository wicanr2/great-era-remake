package game

import (
	"errors"
	"fmt"
)

var ErrInvasionArrival = errors.New("game: invalid or duplicate invasion arrival")

const InvasionRuntimeLimit = 512

// The first scenario's 107 is a Chinese officer. 300 is a remake headquarters
// identity, never a MAN record or a historical person. Original faction tables
// stay unchanged; their ordinary readers cannot interpret this new identity.
func JapaneseInvasionFaction(stage uint8) (GeneralID, error) {
	if stage == 1 {
		return 300, nil
	}
	if stage == 2 || stage == 3 {
		return 107, nil
	}
	return 0, ErrInvasionArrival
}

// InvasionArrival records an authorized, finite import, separately from initial
// equipment and factory production. Equipment totals are derived from the plan.
type InvasionArrival struct {
	Date             InvasionDate
	ArrivedAt        uint32
	Faction          GeneralID
	Units            []GeneralID
	Delivered, Spent InvasionSupply
}

func canonicalInvasionDate(d InvasionDate) bool {
	if d.DeclaredAt == 0 {
		return false
	}
	year, month := (d.DeclaredAt-1)/12, uint8((d.DeclaredAt-1)%12+1)
	if year > 65535 {
		return false
	}
	expected, err := InvasionDateFor(uint16(year), month)
	return err == nil && d == expected
}

func invasionSupplyValues(s InvasionSupply) [4]uint32 {
	return [4]uint32{s.Gold, s.Food, s.Ammo, s.Fuel}
}

func invasionSupplyFrom(v [4]uint32) InvasionSupply {
	return InvasionSupply{Gold: v[0], Food: v[1], Ammo: v[2], Fuel: v[3]}
}

func (b InvasionArrival) remaining(number uint8) (InvasionSupply, error) {
	if number < 1 || number > 3 || !canonicalInvasionDate(b.Date) {
		return InvasionSupply{}, ErrInvasionArrival
	}
	plan, _ := FixedInvasionPlan(b.Date.Era)
	limit := invasionSupplyValues(plan.Waves[number-1].Supply)
	delivered, spent := invasionSupplyValues(b.Delivered), invasionSupplyValues(b.Spent)
	for i := range limit {
		used := uint64(delivered[i]) + uint64(spent[i])
		if used > uint64(limit[i]) {
			return InvasionSupply{}, ErrInvasionArrival
		}
		limit[i] -= uint32(used)
	}
	return invasionSupplyFrom(limit), nil
}

func (a *Armory) cloneInvasionBatches() map[uint8]InvasionArrival {
	if a.InvasionBatches == nil {
		return nil
	}
	out := make(map[uint8]InvasionArrival, len(a.InvasionBatches))
	for number, batch := range a.InvasionBatches {
		batch.Units = append([]GeneralID(nil), batch.Units...)
		out[number] = batch
	}
	return out
}

// Validate the receipt before it contributes to the equipment equation. A
// missing, duplicated, premature or foreign-event receipt cannot explain stock.
func (a *Armory) invasionEquipment(w *AIWorld) (map[EquipmentID]uint64, error) {
	out := map[EquipmentID]uint64{}
	if len(a.InvasionBatches) == 0 {
		return out, nil
	}
	faction, err := JapaneseInvasionFaction(w.Opts.Stage)
	if err != nil || len(a.InvasionBatches) > 3 {
		return nil, ErrInvasionArrival
	}
	first, found := a.InvasionBatches[1]
	if !found || !canonicalInvasionDate(first.Date) {
		return nil, ErrInvasionArrival
	}
	plan, _ := FixedInvasionPlan(first.Date.Era)
	seen := map[GeneralID]bool{}
	for number := uint8(1); int(number) <= len(a.InvasionBatches); number++ {
		batch, found := a.InvasionBatches[number]
		wave := plan.Waves[number-1]
		if !found || batch.Date != first.Date || batch.Faction != faction ||
			int(faction) > len(w.Units) || a.Nations[faction] == nil ||
			batch.ArrivedAt < batch.Date.DeclaredAt+uint32(wave.MonthOffset) || batch.ArrivedAt > a.LastMonth ||
			len(batch.Units) != len(wave.Formations) {
			return nil, ErrInvasionArrival
		}
		if _, err := batch.remaining(number); err != nil {
			return nil, err
		}
		for _, id := range batch.Units {
			if id < 275 || id == 300 || int(id) > InvasionRuntimeLimit || int(id) > len(w.Units) ||
				seen[id] || w.Units[id-1].General != id || a.Units[id] == nil {
				return nil, ErrInvasionArrival
			}
			seen[id] = true
		}
		for id, count := range wave.Equipment {
			out[id] += uint64(count)
		}
	}
	return out, nil
}

type InvasionAssembly struct {
	World    *AIWorld
	Generals []General
	Armory   *Armory
}

func emptyInvasionSlot(id GeneralID) (General, CombatUnit, StrengthInput) {
	return General{}, CombatUnit{General: id, Cell: NoCell, NextCell: NoCell}, StrengthInput{General: id}
}

// AssembleInvasionWave is transactional. Bindings come from the reviewed cast;
// this function cannot select historical people, take a province or start war.
func AssembleInvasionWave(w *AIWorld, army *Armory, generals []General, date InvasionDate, number uint8, ids []GeneralID, destination ProvinceID) (*InvasionAssembly, error) {
	if w == nil || army == nil || w.Table == nil || len(generals) != len(w.Units) ||
		number < 1 || number > 3 || !canonicalInvasionDate(date) || army.InvasionBatches[number].Units != nil {
		return nil, ErrInvasionArrival
	}
	if err := army.Validate(w); err != nil {
		return nil, err
	}
	if army.Landing != nil {
		last := army.Landing.Attempts[len(army.Landing.Attempts)-1]
		if last.Outcome == LandingActive || last.Outcome == LandingWithdrawn {
			return nil, ErrInvasionArrival
		}
	}
	faction, err := JapaneseInvasionFaction(w.Opts.Stage)
	if err != nil {
		return nil, err
	}
	plan, _ := FixedInvasionPlan(date.Era)
	wave := plan.Waves[number-1]
	if len(ids) != len(wave.Formations) || int(number) != len(army.InvasionBatches)+1 ||
		army.LastMonth < date.DeclaredAt+uint32(wave.MonthOffset) {
		return nil, ErrInvasionArrival
	}
	if first, ok := army.InvasionBatches[1]; ok && first.Date != date {
		return nil, ErrInvasionArrival
	}
	if destination != ExternalInvasionSource {
		province, err := w.Table.At(destination)
		if err != nil || destination != InvasionEntryProvince || province.Commander != faction || province.InBattle() {
			return nil, ErrInvasionArrival
		}
	}
	limit := max(len(w.Units), int(faction))
	seen := map[GeneralID]bool{}
	for _, id := range ids {
		if id < 275 || id == 300 || int(id) > InvasionRuntimeLimit || seen[id] {
			return nil, ErrInvasionArrival
		}
		seen[id] = true
		limit = max(limit, int(id))
		if int(id) <= len(w.Units) {
			g, u, in := emptyInvasionSlot(id)
			if generals[id-1] != g || w.Units[id-1] != u || w.Strengths[id-1] != in {
				return nil, ErrInvasionArrival
			}
		}
		for _, batch := range army.InvasionBatches {
			for _, previous := range batch.Units {
				if previous == id {
					return nil, ErrInvasionArrival
				}
			}
		}
	}
	if limit > InvasionRuntimeLimit {
		return nil, ErrInvasionArrival
	}
	world, next := *w, army.Clone()
	table := *w.Table
	world.Table = &table
	world.Units, world.Strengths = make([]CombatUnit, limit), make([]StrengthInput, limit)
	gs := make([]General, limit)
	copy(gs, generals)
	copy(world.Units, w.Units)
	copy(world.Strengths, w.Strengths)
	for i := len(w.Units); i < limit; i++ {
		id := GeneralID(i + 1)
		gs[i], world.Units[i], world.Strengths[i] = emptyInvasionSlot(id)
		next.Units[id] = &UnitEquipment{Items: map[EquipmentID]uint32{}}
	}
	if next.Nations[faction] == nil {
		next.Nations[faction] = newNationArmory()
	}
	var unlock func(EquipmentID)
	unlock = func(id EquipmentID) {
		model, _ := EquipmentByID(id)
		if model.Legacy {
			return
		}
		if model.Prerequisite != "" {
			unlock(model.Prerequisite)
		}
		n := next.Nations[faction]
		n.Known[id] = true
		if model.Months > 0 {
			n.Progress[id] = model.Months
		}
		if n.Active == id {
			n.Active = ""
		}
	}
	for i, formation := range wave.Formations {
		id := ids[i]
		gs[id-1] = General{AbilityA: formation.Ability, AbilityB: 100, AbilityC: formation.Ability,
			Force: formation.Force, Branch: formation.Branch, F19: formation.Skill, F20: formation.Arming,
			Stamina: formation.Stamina, F30: formation.Morale, Province: destination, Range: 1}
		world.Units[id-1] = CombatUnit{General: id, Faction: faction, Province: destination,
			Active: true, Cell: NoCell, NextCell: NoCell, Max: 12, Current: 12, Facing: 1, Decaying: formation.Morale}
		world.Strengths[id-1] = StrengthInput{General: id, Faction: faction, Ability: formation.Ability,
			Force: formation.Force, Branch: formation.Branch, F19: formation.Skill, F20: formation.Arming,
			F29: formation.Stamina, F30: formation.Morale}
		next.Units[id] = &UnitEquipment{Faction: faction, Branch: formation.Branch, Force: formation.Force,
			Items: map[EquipmentID]uint32{formation.Equipment: EquipmentNeed(formation.Branch, formation.Force)}}
		unlock(formation.Equipment)
	}
	if next.InvasionBatches == nil {
		next.InvasionBatches = map[uint8]InvasionArrival{}
	}
	next.InvasionBatches[number] = InvasionArrival{Date: date, ArrivedAt: next.LastMonth,
		Faction: faction, Units: append([]GeneralID(nil), ids...)}
	world.Opts.Armory = next
	if err := next.Validate(&world); err != nil {
		return nil, err
	}
	return &InvasionAssembly{World: &world, Generals: gs, Armory: next}, nil
}

func (a *Armory) InvasionRemainingSupply(number uint8) (InvasionSupply, error) {
	if a == nil {
		return InvasionSupply{}, ErrInvasionArrival
	}
	batch, found := a.InvasionBatches[number]
	if !found {
		return InvasionSupply{}, ErrInvasionArrival
	}
	return batch.remaining(number)
}

// SpendInvasionSupply records only an actual, already selected battle expense.
// No unit, equipment, RNG, province resource or command budget changes here.
func (a *Armory) SpendInvasionSupply(w *AIWorld, number uint8, expense InvasionSupply) error {
	if err := a.Validate(w); err != nil {
		return err
	}
	if a.Landing != nil && a.Landing.Attempts[len(a.Landing.Attempts)-1].Outcome == LandingWithdrawn {
		return ErrInvasionLanding
	}
	available, err := a.InvasionRemainingSupply(number)
	if err != nil {
		return err
	}
	left, take := invasionSupplyValues(available), invasionSupplyValues(expense)
	for i := range left {
		if take[i] > left[i] {
			return ErrInvasionArrival
		}
	}
	next := a.Clone()
	batch := next.InvasionBatches[number]
	spent := invasionSupplyValues(batch.Spent)
	for i := range spent {
		spent[i] += take[i]
	}
	batch.Spent = invasionSupplyFrom(spent)
	next.InvasionBatches[number] = batch
	if err := next.Validate(w); err != nil {
		return err
	}
	*a = *next
	return nil
}

// DeliverInvasionSupply credits the actual landing province once, up to each
// uint16 capacity. Surplus remains in the finite off-map ledger for later use.
func (a *Armory) DeliverInvasionSupply(w *AIWorld, number uint8) (InvasionSupply, error) {
	if err := a.Validate(w); err != nil {
		return InvasionSupply{}, err
	}
	if a.Landing != nil && a.Landing.Attempts[len(a.Landing.Attempts)-1].Outcome == LandingWithdrawn {
		return InvasionSupply{}, ErrInvasionLanding
	}
	available, err := a.InvasionRemainingSupply(number)
	if err != nil {
		return InvasionSupply{}, err
	}
	batch := a.InvasionBatches[number]
	province, err := w.Table.At(InvasionEntryProvince)
	if err != nil || province.Commander != batch.Faction || province.InBattle() {
		return InvasionSupply{}, ErrInvasionArrival
	}
	before := [4]uint16{province.Gold, province.Food, province.Ammo, province.Fuel}
	left := invasionSupplyValues(available)
	var credit [4]uint32
	for i := range credit {
		credit[i] = min(left[i], uint32(65535-before[i]))
	}
	next := a.Clone()
	delivered := invasionSupplyValues(batch.Delivered)
	for i := range delivered {
		delivered[i] += credit[i]
	}
	batch.Delivered = invasionSupplyFrom(delivered)
	next.InvasionBatches[number] = batch
	if err := next.Validate(w); err != nil {
		return InvasionSupply{}, fmt.Errorf("%w: %v", ErrInvasionArrival, err)
	}
	province.Gold += uint16(credit[0])
	province.Food += uint16(credit[1])
	province.Ammo += uint16(credit[2])
	province.Fuel += uint16(credit[3])
	*a = *next
	return invasionSupplyFrom(credit), nil
}
