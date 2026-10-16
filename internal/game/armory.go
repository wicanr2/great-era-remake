package game

// SPEC-55: optional remake expansion. No legacy record field is repurposed.
import (
	"errors"
	"fmt"
	"maps"
	"sort"
)

type EquipmentID string
type EquipmentClass uint8

const (
	EquipmentRifle EquipmentClass = iota
	EquipmentArmor
	EquipmentArtillery
	ArmoryStockLimit  uint32 = 1_000_000
	armoryLedgerLimit uint64 = 1_000_000_000_000
)

type ArmoryCost struct{ Gold, Iron, Coal, Fuel uint32 }
type EquipmentModel struct {
	ID                EquipmentID
	Class             EquipmentClass
	Months            uint16
	Prerequisite      EquipmentID
	Output            uint32
	Cost              ArmoryCost
	Power, RoughPower uint32
	Legacy            bool
}

var equipmentModels = []EquipmentModel{
	{ID: "rifle-old", Class: EquipmentRifle, Output: 30, Cost: ArmoryCost{Gold: 4, Iron: 1}, Power: 100, RoughPower: 100},
	{ID: "rifle-modern", Class: EquipmentRifle, Months: 2, Prerequisite: "rifle-old", Output: 24, Cost: ArmoryCost{Gold: 6, Iron: 1}, Power: 125, RoughPower: 125},
	{ID: "light-tank", Class: EquipmentArmor, Months: 3, Output: 1, Cost: ArmoryCost{Gold: 300, Iron: 15, Coal: 6, Fuel: 2}, Power: 125, RoughPower: 87},
	{ID: "armor-improved", Class: EquipmentArmor, Months: 4, Prerequisite: "light-tank", Output: 1, Cost: ArmoryCost{Gold: 450, Iron: 22, Coal: 9, Fuel: 3}, Power: 150, RoughPower: 105},
	{ID: "mountain-basic", Class: EquipmentArtillery, Output: 2, Cost: ArmoryCost{Gold: 65, Iron: 4, Coal: 2}, Power: 100, RoughPower: 100},
	{ID: "mountain-improved", Class: EquipmentArtillery, Months: 2, Prerequisite: "mountain-basic", Output: 2, Cost: ArmoryCost{Gold: 80, Iron: 5, Coal: 2}, Power: 125, RoughPower: 125},
	{ID: "field-howitzer", Class: EquipmentArtillery, Months: 3, Output: 1, Cost: ArmoryCost{Gold: 130, Iron: 8, Coal: 4}, Power: 140, RoughPower: 90},
	{ID: "antitank", Class: EquipmentArtillery, Months: 2, Output: 2, Cost: ArmoryCost{Gold: 90, Iron: 4, Coal: 2}, Power: 110, RoughPower: 110},
	{ID: "legacy-rifle", Class: EquipmentRifle, Power: 100, RoughPower: 100, Legacy: true},
	{ID: "legacy-armor", Class: EquipmentArmor, Power: 100, RoughPower: 100, Legacy: true},
	{ID: "legacy-artillery", Class: EquipmentArtillery, Power: 100, RoughPower: 100, Legacy: true},
}

func EquipmentCatalog() []EquipmentModel { return append([]EquipmentModel(nil), equipmentModels...) }
func EquipmentByID(id EquipmentID) (EquipmentModel, bool) {
	for _, m := range equipmentModels {
		if m.ID == id {
			return m, true
		}
	}
	return EquipmentModel{}, false
}
func EquipmentClassOf(branch uint8) (EquipmentClass, bool) {
	switch branch {
	case BranchInfantry, BranchCavalry:
		return EquipmentRifle, true
	case BranchArmour:
		return EquipmentArmor, true
	case BranchArtiller:
		return EquipmentArtillery, true
	}
	return 0, false
}
func EquipmentNeed(branch uint8, force uint16) uint32 {
	if _, ok := EquipmentClassOf(branch); !ok || force == 0 {
		return 0
	}
	n := uint32(100)
	if branch == BranchArmour {
		n = 10
	}
	return (uint32(force) + n - 1) / n
}
func legacyEquipment(c EquipmentClass) EquipmentID {
	return [...]EquipmentID{"legacy-rifle", "legacy-armor", "legacy-artillery"}[c]
}

type ArmoryProduction struct {
	Count    uint32
	Shortage bool
}
type NationArmory struct {
	Known    map[EquipmentID]bool
	Progress map[EquipmentID]uint16
	Active   EquipmentID
	Lines    map[EquipmentID]uint16
	Stock    map[EquipmentID]uint32
	Last     map[EquipmentID]ArmoryProduction
}
type UnitEquipment struct {
	Faction GeneralID
	Branch  uint8
	Force   uint16
	Items   map[EquipmentID]uint32
}
type Armory struct {
	Version                 uint8
	Provinces               int
	LastMonth               uint32
	Nations                 map[GeneralID]*NationArmory
	Units                   map[GeneralID]*UnitEquipment
	ReservePersonnel        map[GeneralID]uint32
	Initial, Produced, Lost map[EquipmentID]uint64
}

var (
	ErrArmoryState     = errors.New("game: invalid armory state")
	ErrArmoryLocked    = errors.New("game: equipment design is unavailable")
	ErrArmoryFactories = errors.New("game: factory allocation exceeds capacity")
	ErrArmoryUnit      = errors.New("game: unit cannot receive this equipment")
	ErrArmoryStock     = errors.New("game: no equipment available to issue")
	ErrArmoryMonth     = errors.New("game: armory month must advance exactly once")
)

func newNationArmory() *NationArmory {
	return &NationArmory{Known: map[EquipmentID]bool{"rifle-old": true, "mountain-basic": true}, Progress: map[EquipmentID]uint16{}, Lines: map[EquipmentID]uint16{}, Stock: map[EquipmentID]uint32{}, Last: map[EquipmentID]ArmoryProduction{}}
}
func NewArmory(w *AIWorld, provinces int, month uint32) (*Armory, error) {
	if w == nil || w.Table == nil || provinces < 1 || provinces > ProvinceCount || month == 0 || len(w.Units) != len(w.Strengths) {
		return nil, ErrArmoryState
	}
	a := &Armory{Version: 1, Provinces: provinces, LastMonth: month, Nations: map[GeneralID]*NationArmory{}, Units: map[GeneralID]*UnitEquipment{}, ReservePersonnel: map[GeneralID]uint32{}, Initial: map[EquipmentID]uint64{}, Produced: map[EquipmentID]uint64{}, Lost: map[EquipmentID]uint64{}}
	for _, p := range w.Table.Province[:provinces] {
		if p.Commander != 0 {
			a.Nations[p.Commander] = newNationArmory()
		}
	}
	for i, input := range w.Strengths {
		id := GeneralID(i + 1)
		if w.Units[i].General != id {
			return nil, ErrArmoryState
		}
		u := &UnitEquipment{Faction: w.Units[i].Faction, Branch: input.Branch, Force: input.Force, Items: map[EquipmentID]uint32{}}
		if class, ok := EquipmentClassOf(input.Branch); ok {
			m := legacyEquipment(class)
			n := EquipmentNeed(input.Branch, input.Force)
			u.Items[m] = n
			a.Initial[m] += uint64(n)
		}
		a.Units[id] = u
	}
	if err := a.Validate(w); err != nil {
		return nil, err
	}
	return a, nil
}
func (a *Armory) nation(w *AIWorld, faction GeneralID) (*NationArmory, error) {
	if a == nil || w == nil || w.Table == nil || faction == 0 || a.Provinces < 1 || a.Provinces > ProvinceCount {
		return nil, ErrArmoryState
	}
	for _, p := range w.Table.Province[:a.Provinces] {
		if p.Commander == faction {
			if n := a.Nations[faction]; n != nil {
				return n, nil
			}
		}
	}
	return nil, ErrArmoryState
}
func (a *Armory) Resources(w *AIWorld, faction GeneralID) ArmoryCost {
	var sum ArmoryCost
	if a == nil || w == nil || w.Table == nil {
		return sum
	}
	for _, p := range w.Table.Province[:a.Provinces] {
		if p.Commander == faction {
			sum.Gold += uint32(p.Gold)
			sum.Iron += uint32(p.Iron)
			sum.Coal += uint32(p.Coal)
			sum.Fuel += uint32(p.Fuel)
		}
	}
	return sum
}
func (a *Armory) Factories(w *AIWorld, faction GeneralID) uint32 {
	var sum uint32
	if a == nil || w == nil || w.Table == nil {
		return 0
	}
	for _, p := range w.Table.Province[:a.Provinces] {
		if p.Commander == faction {
			sum += uint32(p.Arsenals)
		}
	}
	return sum
}
func (n *NationArmory) Assigned() uint32 {
	var sum uint32
	if n != nil {
		for _, v := range n.Lines {
			sum += uint32(v)
		}
	}
	return sum
}
func (a *Armory) spend(w *AIWorld, faction GeneralID, c ArmoryCost) bool {
	r := a.Resources(w, faction)
	if r.Gold < c.Gold || r.Iron < c.Iron || r.Coal < c.Coal || r.Fuel < c.Fuel {
		return false
	}
	for i := 0; i < a.Provinces; i++ {
		p := &w.Table.Province[i]
		if p.Commander != faction {
			continue
		}
		for _, x := range []struct {
			have *uint16
			need *uint32
		}{{&p.Gold, &c.Gold}, {&p.Iron, &c.Iron}, {&p.Coal, &c.Coal}, {&p.Fuel, &c.Fuel}} {
			n := min(uint32(*x.have), *x.need)
			*x.have -= uint16(n)
			*x.need -= n
		}
	}
	return true
}
func (a *Armory) Research(w *AIWorld, faction GeneralID, id EquipmentID) error {
	n, err := a.nation(w, faction)
	if err != nil {
		return err
	}
	if id == "" {
		n.Active = ""
		return nil
	}
	m, ok := EquipmentByID(id)
	if !ok || m.Legacy || m.Months == 0 || n.Known[id] || (m.Prerequisite != "" && !n.Known[m.Prerequisite]) {
		return ErrArmoryLocked
	}
	n.Active = id
	return nil
}
func (a *Armory) Assign(w *AIWorld, faction GeneralID, id EquipmentID, delta int) error {
	n, err := a.nation(w, faction)
	if err != nil {
		return err
	}
	m, ok := EquipmentByID(id)
	if !ok || m.Legacy || !n.Known[id] {
		return ErrArmoryLocked
	}
	next := int(n.Lines[id]) + delta
	if next < 0 || next > ProvinceCount*255 || int64(n.Assigned())+int64(delta) > int64(a.Factories(w, faction)) && delta > 0 {
		return ErrArmoryFactories
	}
	n.Lines[id] = uint16(next)
	return nil
}
func (a *Armory) normalizeFactories(w *AIWorld, faction GeneralID, n *NationArmory) {
	left := a.Factories(w, faction)
	for _, m := range equipmentModels {
		if n.Lines[m.ID] == 0 {
			continue
		}
		keep := min(uint32(n.Lines[m.ID]), left)
		n.Lines[m.ID] = uint16(keep)
		left -= keep
	}
}
func affordable(r, c ArmoryCost, wanted uint32) uint32 {
	for _, x := range [][2]uint32{{r.Gold, c.Gold}, {r.Iron, c.Iron}, {r.Coal, c.Coal}, {r.Fuel, c.Fuel}} {
		if x[1] > 0 {
			wanted = min(wanted, x[0]/x[1])
		}
	}
	return wanted
}
func scaledCost(c ArmoryCost, n uint32) ArmoryCost {
	return ArmoryCost{c.Gold * n, c.Iron * n, c.Coal * n, c.Fuel * n}
}

// SettleMonth charges each completed item once. Research follows existing lines.
func (a *Armory) SettleMonth(w *AIWorld, month uint32) error {
	if a == nil || month != a.LastMonth+1 || month > 65536*12 {
		return ErrArmoryMonth
	}
	if err := a.Validate(w); err != nil {
		return err
	}
	ids := a.FactionIDs()
	for _, faction := range ids {
		n, err := a.nation(w, faction)
		if err != nil {
			continue
		}
		a.normalizeFactories(w, faction, n)
		n.Last = map[EquipmentID]ArmoryProduction{}
		for _, m := range equipmentModels {
			factories := uint32(n.Lines[m.ID])
			if m.Legacy || factories == 0 {
				continue
			}
			wanted := factories * m.Output
			count := affordable(a.Resources(w, faction), m.Cost, wanted)
			count = min(count, ArmoryStockLimit-n.Stock[m.ID])
			count = uint32(min(uint64(count), armoryLedgerLimit-a.Produced[m.ID]))
			if count > 0 && !a.spend(w, faction, scaledCost(m.Cost, count)) {
				return ErrArmoryState
			}
			n.Stock[m.ID] += count
			a.Produced[m.ID] += uint64(count)
			n.Last[m.ID] = ArmoryProduction{Count: count, Shortage: count < wanted}
		}
		if n.Active != "" && a.spend(w, faction, ArmoryCost{Gold: 180}) {
			m, _ := EquipmentByID(n.Active)
			n.Progress[m.ID]++
			if n.Progress[m.ID] >= m.Months {
				n.Known[m.ID] = true
				n.Active = ""
			}
		}
	}
	a.LastMonth = month
	return nil
}
func (a *Armory) FactionIDs() []GeneralID {
	ids := make([]GeneralID, 0, len(a.Nations))
	for id := range a.Nations {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	return ids
}
func (a *Armory) Issue(w *AIWorld, faction GeneralID, unit GeneralID, id EquipmentID) (uint32, error) {
	return a.issueLimit(w, faction, unit, id, ArmoryStockLimit)
}

type ArmoryConversion struct {
	Before, Force         uint16
	Groups, Reserve, Gold uint32
}

func (a *Armory) ownedInfantry(w *AIWorld, faction, id GeneralID) (*UnitEquipment, error) {
	if _, err := a.nation(w, faction); err != nil {
		return nil, err
	}
	i := int(id) - 1
	u := a.Units[id]
	if i < 0 || i >= len(w.Units) || i >= len(w.Strengths) || u == nil || u.Faction != faction || u.Branch != BranchInfantry || u.Force == 0 || w.Units[i].General != id || w.Units[i].Faction != faction || w.Strengths[i].Branch != u.Branch || w.Strengths[i].Force != u.Force {
		return nil, ErrArmoryUnit
	}
	p := w.Units[i].Province
	if !p.Valid() || int(p) > a.Provinces || w.Table.Province[int(p)-1].Commander != faction {
		return nil, ErrArmoryUnit
	}
	return u, nil
}

// ConversionPlan validates the entire transaction without spending or returning gear.
func (a *Armory) ConversionPlan(w *AIWorld, faction, id GeneralID, model EquipmentID) (ArmoryConversion, error) {
	u, err := a.ownedInfantry(w, faction, id)
	if err != nil {
		return ArmoryConversion{}, err
	}
	m, ok := EquipmentByID(model)
	n := a.Nations[faction]
	if !ok || m.Class != EquipmentArmor || m.Legacy || !n.Known[model] || u.Force < 100 {
		return ArmoryConversion{}, ErrArmoryLocked
	}
	p := ArmoryConversion{Before: u.Force, Force: min(uint16(200), u.Force/100), Gold: 1000}
	p.Groups = EquipmentNeed(BranchArmour, p.Force)
	p.Reserve = uint32(u.Force) - uint32(p.Force)*100
	if a.ReservePersonnel == nil || uint64(a.ReservePersonnel[faction])+uint64(p.Reserve) > uint64(ArmoryStockLimit) {
		return ArmoryConversion{}, ErrArmoryState
	}
	if n.Stock[model] < p.Groups || a.Resources(w, faction).Gold < p.Gold {
		return ArmoryConversion{}, ErrArmoryStock
	}
	for old, count := range u.Items {
		oldModel, known := EquipmentByID(old)
		if !known || count > 0 && oldModel.Class != EquipmentRifle || uint64(n.Stock[old])+uint64(count) > uint64(ArmoryStockLimit) {
			return ArmoryConversion{}, ErrArmoryStock
		}
	}
	return p, nil
}

func (a *Armory) ConvertToArmor(w *AIWorld, faction, id GeneralID, model EquipmentID) (ArmoryConversion, error) {
	p, err := a.ConversionPlan(w, faction, id, model)
	if err != nil {
		return ArmoryConversion{}, err
	}
	if !a.spend(w, faction, ArmoryCost{Gold: p.Gold}) {
		return ArmoryConversion{}, ErrArmoryStock
	}
	n, u := a.Nations[faction], a.Units[id]
	for old, count := range u.Items {
		if count > 0 {
			n.Stock[old] += count
		}
	}
	n.Stock[model] -= p.Groups
	u.Items = map[EquipmentID]uint32{model: p.Groups}
	u.Branch, u.Force = BranchArmour, p.Force
	w.Strengths[int(id)-1].Branch, w.Strengths[int(id)-1].Force = BranchArmour, p.Force
	a.ReservePersonnel[faction] += p.Reserve
	return p, nil
}

func (a *Armory) ReinforceReserve(w *AIWorld, faction, id GeneralID) (uint32, error) {
	u, err := a.ownedInfantry(w, faction, id)
	if err != nil {
		return 0, err
	}
	if u.Force >= 20000 || a.ReservePersonnel[faction] == 0 {
		return 0, ErrArmoryUnit
	}
	count := min(a.ReservePersonnel[faction], uint32(20000-u.Force))
	u.Force += uint16(count)
	w.Strengths[int(id)-1].Force = u.Force
	a.ReservePersonnel[faction] -= count
	return count, nil
}
func (a *Armory) issueLimit(w *AIWorld, faction GeneralID, unit GeneralID, id EquipmentID, limit uint32) (uint32, error) {
	n, err := a.nation(w, faction)
	if err != nil {
		return 0, err
	}
	u := a.Units[unit]
	m, ok := EquipmentByID(id)
	if u == nil || !ok || u.Faction != faction || u.Force == 0 || (!m.Legacy && !n.Known[id]) {
		return 0, ErrArmoryUnit
	}
	class, ok := EquipmentClassOf(u.Branch)
	if !ok || class != m.Class {
		return 0, ErrArmoryUnit
	}
	index := int(unit) - 1
	if index < 0 || index >= len(w.Units) || w.Units[index].Faction != faction || !w.Units[index].Province.Valid() || int(w.Units[index].Province) > a.Provinces {
		return 0, ErrArmoryUnit
	}
	p, _ := w.Table.At(w.Units[index].Province)
	if p.Commander != faction {
		return 0, ErrArmoryUnit
	}
	need := EquipmentNeed(u.Branch, u.Force)
	var equipped uint64
	for _, v := range u.Items {
		equipped += uint64(v)
	}
	if equipped > uint64(need) || u.Items[id] > need {
		return 0, ErrArmoryState
	}
	amount := min(limit, n.Stock[id], need-u.Items[id])
	if amount == 0 {
		return 0, ErrArmoryStock
	}
	replace := int64(amount) - int64(uint64(need)-equipped)
	if replace < 0 {
		replace = 0
	}
	models := EquipmentCatalog()
	sort.SliceStable(models, func(i, j int) bool { return models[i].Power < models[j].Power })
	returns := map[EquipmentID]uint32{}
	for _, old := range models {
		if old.ID == id {
			continue
		}
		v := min(uint32(replace), u.Items[old.ID])
		if uint64(n.Stock[old.ID])+uint64(v) > uint64(ArmoryStockLimit) {
			return 0, ErrArmoryStock
		}
		returns[old.ID] = v
		replace -= int64(v)
	}
	for old, v := range returns {
		u.Items[old] -= v
		n.Stock[old] += v
	}
	n.Stock[id] -= amount
	u.Items[id] += amount
	return amount, nil
}

// ApplyLoss is called at the actual casualty writer, never at save/load.
func (a *Armory) ApplyLoss(id GeneralID, before, after uint16) {
	if a == nil || before == 0 || after >= before {
		return
	}
	u := a.Units[id]
	if u == nil {
		return
	}
	for m, count := range u.Items {
		keep := uint32(uint64(count) * uint64(after) / uint64(before))
		u.Items[m] = keep
		a.Lost[m] += uint64(count - keep)
	}
	u.Force = after
}
func (a *Armory) Power(id GeneralID, branch uint8, force uint16, rough bool) uint32 {
	if a == nil {
		return 100
	}
	u := a.Units[id]
	need := EquipmentNeed(branch, force)
	if u == nil || need == 0 {
		return 100
	}
	class, ok := EquipmentClassOf(branch)
	if !ok {
		return 100
	}
	var total, power uint64
	for _, m := range equipmentModels {
		if m.Class != class {
			continue
		}
		n := min(uint64(u.Items[m.ID]), uint64(need)-min(total, uint64(need)))
		v := m.Power
		if rough {
			v = m.RoughPower
		}
		total += n
		power += n * uint64(v)
	}
	power += (uint64(need) - total) * 50
	return uint32((power + uint64(need)/2) / uint64(need))
}

// Sync handles noncombat reinforcement/reorganization. Excess is returned, not destroyed.
func (a *Armory) Sync(w *AIWorld) error {
	if a == nil {
		return ErrArmoryState
	}
	candidate := a.Clone()
	if err := candidate.sync(w); err != nil {
		return err
	}
	*a = *candidate
	return nil
}
func (a *Armory) Clone() *Armory {
	if a == nil {
		return nil
	}
	out := *a
	out.Nations = make(map[GeneralID]*NationArmory, len(a.Nations))
	out.Units = make(map[GeneralID]*UnitEquipment, len(a.Units))
	for id, n := range a.Nations {
		if n == nil {
			continue
		}
		copyN := *n
		copyN.Known = maps.Clone(n.Known)
		copyN.Progress = maps.Clone(n.Progress)
		copyN.Lines = maps.Clone(n.Lines)
		copyN.Stock = maps.Clone(n.Stock)
		copyN.Last = maps.Clone(n.Last)
		out.Nations[id] = &copyN
	}
	for id, u := range a.Units {
		if u == nil {
			continue
		}
		copyU := *u
		copyU.Items = maps.Clone(u.Items)
		out.Units[id] = &copyU
	}
	out.Initial = maps.Clone(a.Initial)
	out.Produced = maps.Clone(a.Produced)
	out.Lost = maps.Clone(a.Lost)
	out.ReservePersonnel = maps.Clone(a.ReservePersonnel)
	return &out
}
func (a *Armory) sync(w *AIWorld) error {
	if a == nil || w == nil || w.Table == nil || len(w.Units) != len(w.Strengths) || len(a.Units) != len(w.Units) {
		return ErrArmoryState
	}
	for i, u := range w.Units {
		if u.General != GeneralID(i+1) || a.Units[u.General] == nil {
			return ErrArmoryState
		}
	}
	for _, p := range w.Table.Province[:a.Provinces] {
		if p.Commander != 0 && a.Nations[p.Commander] == nil {
			a.Nations[p.Commander] = newNationArmory()
		}
	}
	for i, input := range w.Strengths {
		id := GeneralID(i + 1)
		unit := a.Units[id]
		faction := w.Units[i].Faction
		returnOwner := faction
		if returnOwner == 0 {
			returnOwner = unit.Faction
		}
		nation := a.Nations[returnOwner]
		class, valid := EquipmentClassOf(input.Branch)
		need := EquipmentNeed(input.Branch, input.Force)
		var total uint32
		for _, m := range equipmentModels {
			if (!valid || m.Class != class) && unit.Items[m.ID] > 0 {
				if nation != nil {
					if uint64(nation.Stock[m.ID])+uint64(unit.Items[m.ID]) > uint64(ArmoryStockLimit) {
						return ErrArmoryStock
					}
					nation.Stock[m.ID] += unit.Items[m.ID]
					unit.Items[m.ID] = 0
				}
			}
			total += unit.Items[m.ID]
		}
		if total > need && nation != nil {
			excess := total - need
			for _, m := range equipmentModels {
				take := min(excess, unit.Items[m.ID])
				if take == 0 {
					continue
				}
				if uint64(nation.Stock[m.ID])+uint64(take) > uint64(ArmoryStockLimit) {
					return ErrArmoryStock
				}
				unit.Items[m.ID] -= take
				nation.Stock[m.ID] += take
				excess -= take
			}
		}
		unit.Faction, unit.Branch, unit.Force = faction, input.Branch, input.Force
	}
	return nil
}

// PlanAI only adds expansion procurement; original political AI is unchanged.
func (a *Armory) PlanAI(w *AIWorld, player GeneralID) {
	if a == nil {
		return
	}
	for _, f := range a.FactionIDs() {
		if f == player {
			continue
		}
		n, err := a.nation(w, f)
		if err != nil {
			continue
		}
		present := map[EquipmentClass]bool{}
		for _, u := range a.Units {
			if u.Faction == f && u.Force > 0 {
				if c, ok := EquipmentClassOf(u.Branch); ok {
					present[c] = true
				}
			}
		}
		if n.Active == "" {
			for _, id := range []EquipmentID{"rifle-modern", "mountain-improved", "antitank", "light-tank", "field-howitzer", "armor-improved"} {
				m, _ := EquipmentByID(id)
				if present[m.Class] && !n.Known[id] && (m.Prerequisite == "" || n.Known[m.Prerequisite]) {
					_ = a.Research(w, f, id)
					break
				}
			}
		}
		for id := GeneralID(1); int(id) <= len(w.Units); id++ {
			u := a.Units[id]
			if u.Faction != f {
				continue
			}
			class, ok := EquipmentClassOf(u.Branch)
			if !ok {
				continue
			}
			models := EquipmentCatalog()
			sort.SliceStable(models, func(i, j int) bool { return models[i].Power > models[j].Power })
			for _, m := range models {
				if m.Class != class || n.Stock[m.ID] == 0 {
					continue
				}
				var total, lower uint32
				for old, count := range u.Items {
					oldModel, _ := EquipmentByID(old)
					total += count
					if oldModel.Power < m.Power {
						lower += count
					}
				}
				eligible := lower + EquipmentNeed(u.Branch, u.Force) - total
				if eligible > 0 {
					_, _ = a.issueLimit(w, f, id, m.ID, eligible)
				}
			}
		}
		for _, m := range equipmentModels {
			n.Lines[m.ID] = 0
		}
		capacity := a.Factories(w, f)
		for _, class := range []EquipmentClass{EquipmentRifle, EquipmentArtillery, EquipmentArmor} {
			if !present[class] || capacity == 0 {
				continue
			}
			var best EquipmentModel
			for _, m := range equipmentModels {
				if m.Class == class && !m.Legacy && n.Known[m.ID] && m.Power > best.Power {
					best = m
				}
			}
			if best.ID == "" {
				continue
			}
			var desired uint32
			for _, u := range a.Units {
				if u.Faction == f {
					c, ok := EquipmentClassOf(u.Branch)
					if ok && c == class {
						need := EquipmentNeed(u.Branch, u.Force)
						desired += need - u.Items[best.ID] + need/10
					}
				}
			}
			if desired > n.Stock[best.ID] {
				factories := min(capacity, (desired-n.Stock[best.ID]+best.Output-1)/best.Output)
				n.Lines[best.ID] = uint16(factories)
				capacity -= factories
			}
		}
	}
}

func (a *Armory) Validate(w *AIWorld) error {
	if a == nil || a.Version != 1 || a.Provinces < 1 || a.Provinces > ProvinceCount || a.LastMonth == 0 || a.LastMonth > 65536*12 || a.Nations == nil || a.Units == nil || a.Initial == nil || a.Produced == nil || a.Lost == nil || a.ReservePersonnel == nil || w == nil || w.Table == nil || len(a.Units) != len(w.Units) || len(w.Units) != len(w.Strengths) {
		return ErrArmoryState
	}
	for faction, count := range a.ReservePersonnel {
		if faction == 0 || a.Nations[faction] == nil || count > ArmoryStockLimit {
			return ErrArmoryState
		}
	}
	totals := map[EquipmentID]uint64{}
	for _, ledger := range []map[EquipmentID]uint64{a.Initial, a.Produced, a.Lost} {
		for id, n := range ledger {
			if _, ok := EquipmentByID(id); !ok || n > armoryLedgerLimit {
				return ErrArmoryState
			}
		}
	}
	for faction, n := range a.Nations {
		if faction == 0 || int(faction) > len(w.Units) || n == nil || n.Known == nil || n.Progress == nil || n.Lines == nil || n.Stock == nil || n.Last == nil || !n.Known["rifle-old"] || !n.Known["mountain-basic"] {
			return ErrArmoryState
		}
		for id, known := range n.Known {
			m, ok := EquipmentByID(id)
			if !ok || m.Legacy || !known || m.Prerequisite != "" && !n.Known[m.Prerequisite] || m.Months > 0 && n.Progress[id] != m.Months {
				return ErrArmoryState
			}
		}
		for id, v := range n.Progress {
			m, ok := EquipmentByID(id)
			if !ok || m.Legacy || m.Months == 0 || v > m.Months || v == m.Months && !n.Known[id] || m.Prerequisite != "" && !n.Known[m.Prerequisite] {
				return ErrArmoryState
			}
		}
		if n.Active != "" {
			m, ok := EquipmentByID(n.Active)
			if !ok || m.Legacy || m.Months == 0 || n.Known[m.ID] || m.Prerequisite != "" && !n.Known[m.Prerequisite] {
				return ErrArmoryState
			}
		}
		for id, v := range n.Lines {
			m, ok := EquipmentByID(id)
			if !ok || m.Legacy && v > 0 || v > ProvinceCount*255 || v > 0 && !m.Legacy && !n.Known[id] {
				return ErrArmoryState
			}
		}
		for id, v := range n.Stock {
			if _, ok := EquipmentByID(id); !ok || v > ArmoryStockLimit {
				return ErrArmoryState
			}
			totals[id] += uint64(v)
		}
		for id, v := range n.Last {
			m, ok := EquipmentByID(id)
			if !ok || m.Legacy || !n.Known[id] || v.Count > ProvinceCount*255*m.Output {
				return ErrArmoryState
			}
		}
	}
	for id, u := range a.Units {
		if id == 0 || int(id) > len(w.Units) || u == nil || u.Items == nil || int(u.Faction) > len(w.Units) || u.Faction != 0 && a.Nations[u.Faction] == nil {
			return ErrArmoryState
		}
		class, ok := EquipmentClassOf(u.Branch)
		var sum uint64
		for mid, n := range u.Items {
			m, exists := EquipmentByID(mid)
			if !exists || n > ArmoryStockLimit || n > 0 && (!ok || class != m.Class) {
				return ErrArmoryState
			}
			sum += uint64(n)
			totals[mid] += uint64(n)
		}
		if sum > uint64(EquipmentNeed(u.Branch, u.Force)) {
			return ErrArmoryState
		}
	}
	for _, m := range equipmentModels {
		if a.Initial[m.ID]+a.Produced[m.ID] != totals[m.ID]+a.Lost[m.ID] {
			return fmt.Errorf("%w: equipment ledger %s", ErrArmoryState, m.ID)
		}
	}
	return nil
}
