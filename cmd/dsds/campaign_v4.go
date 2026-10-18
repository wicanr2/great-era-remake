package main

import (
	"fmt"
	"sort"

	"github.com/wicanr2/great-era-remake/internal/game"
)

const campaignActorFirstID game.GeneralID = 275
const campaignDeclarationID = "post-unification-japan-v1"

type campaignActor struct {
	Key      string
	ID       game.GeneralID
	General  game.General
	Unit     game.CombatUnit
	Strength game.StrengthInput
}

type campaignDeclaration struct {
	ID               string
	Date             game.InvasionDate
	AnnouncementRead bool
	Imported         bool
	Controller       uint8            `json:",omitempty"`
	Cast             []game.GeneralID `json:",omitempty"`
	StartedWithPort  bool             `json:",omitempty"`
	Outcome          string           `json:",omitempty"`
	OutcomeRead      bool             `json:",omitempty"`
}

// The Base keeps exactly the original scenario prefix. Projection holes are
// runtime-only and must never be serialized as original people.
type campaignV4 struct {
	Base        campaignSnapshot
	Rules       game.MilitaryRules
	Armory      *game.Armory
	View        armoryViewSnapshot
	Actors      []campaignActor
	Declaration *campaignDeclaration
}

func campaignActorKey(id game.GeneralID) string {
	return game.InvasionActorKey(id)
}

func importedCampaignDeclaration(rules game.MilitaryRules) (*campaignDeclaration, error) {
	if !rules.Locked() {
		return nil, nil
	}
	year, month := (rules.DeclaredAt-1)/12, uint8((rules.DeclaredAt-1)%12+1)
	if year > 65535 {
		return nil, fmt.Errorf("invalid declaration year")
	}
	date, err := game.InvasionDateFor(uint16(year), month)
	if err != nil {
		return nil, err
	}
	return &campaignDeclaration{ID: campaignDeclarationID, Date: date, AnnouncementRead: true, Imported: true}, nil
}

func validCampaignDeclaration(d *campaignDeclaration, rules game.MilitaryRules, year uint16, month uint8) bool {
	if !rules.Valid(uint32(year)*12+uint32(month)) || rules.Locked() != (d != nil) {
		return false
	}
	if d == nil {
		return true
	}
	expected, err := importedCampaignDeclaration(rules)
	if err != nil || d.ID != campaignDeclarationID || d.Date != expected.Date || d.Imported && !d.AnnouncementRead {
		return false
	}
	if d.Controller == 0 {
		return len(d.Cast) == 0 && !d.StartedWithPort && d.Outcome == "" && !d.OutcomeRead
	}
	if d.Controller != 1 || d.Imported || len(d.Cast) != 10 ||
		d.Outcome != "" && d.Outcome != "victory" && d.Outcome != "defeat" || d.OutcomeRead && d.Outcome == "" {
		return false
	}
	cast, err := game.SelectInvasionCast(d.Date)
	if err != nil {
		return false
	}
	for i, person := range cast {
		if d.Cast[i] != person.ID {
			return false
		}
	}
	return true
}

func campaignHole(id game.GeneralID) (game.General, game.CombatUnit, game.StrengthInput) {
	return game.General{}, game.CombatUnit{General: id, Cell: game.NoCell, NextCell: game.NoCell}, game.StrengthInput{General: id}
}

func validCampaignArrivals(v *campaignV4, keys map[game.GeneralID]string) bool {
	if v.Armory == nil || len(v.Armory.InvasionBatches) == 0 {
		return true
	}
	if v.Declaration == nil || v.Declaration.Imported || !v.Rules.Enabled || !v.Rules.Locked() {
		return false
	}
	seen := map[game.GeneralID]bool{}
	for _, batch := range v.Armory.InvasionBatches {
		if batch.Date != v.Declaration.Date {
			return false
		}
		for _, id := range batch.Units {
			if keys[id] == "" || seen[id] {
				return false
			}
			seen[id] = true
		}
	}
	if v.Declaration.Controller == 1 {
		for number, batch := range v.Armory.InvasionBatches {
			start, end := 0, 6
			if number == 2 {
				start, end = 6, 8
			}
			if number == 3 {
				start, end = 8, 10
			}
			if number < 1 || number > 3 || len(batch.Units) != end-start {
				return false
			}
			for i, id := range batch.Units {
				if id != v.Declaration.Cast[start+i] {
					return false
				}
			}
		}
	}
	return len(seen) == len(keys)
}

// projectCampaignActors validates private data before producing any live arrays.
// Original v1-v3 validation does not call this function or accept external IDs.
func projectCampaignActors(v *campaignV4, sc game.Scenario) ([]game.General, []game.CombatUnit, []game.StrengthInput, map[game.GeneralID]string, error) {
	bad := func() ([]game.General, []game.CombatUnit, []game.StrengthInput, map[game.GeneralID]string, error) {
		return nil, nil, nil, nil, fmt.Errorf("invalid v4 actor or declaration state")
	}
	firstStageInvasion := sc.Stage == 1 && v.Declaration != nil && v.Declaration.Controller == 1
	if !validCampaignDeclaration(v.Declaration, v.Rules, v.Base.Year, v.Base.Month) || len(v.Actors) > 10 ||
		len(v.Actors) > 0 && (sc.Stage < 2 && !firstStageInvasion || sc.Stage > 1 && v.Base.Identity.Player == 107 || !v.Rules.Locked() || v.Armory == nil) {
		return bad()
	}
	if len(v.Base.State.Generals) != sc.Generals || len(v.Base.Units) != sc.Generals || len(v.Base.Strengths) != sc.Generals {
		return bad()
	}
	limit := sc.Generals
	if firstStageInvasion && len(v.Actors) > 0 {
		limit = 300
	}
	keys := map[game.GeneralID]string{}
	for _, actor := range v.Actors {
		if actor.Key == "" || campaignActorKey(actor.ID) != actor.Key || keys[actor.ID] != "" {
			return bad()
		}
		keys[actor.ID] = actor.Key
		limit = max(limit, int(actor.ID))
	}
	if !validCampaignArrivals(v, keys) {
		return bad()
	}
	validID := func(id game.GeneralID) bool { return id > 0 && (int(id) <= sc.Generals || keys[id] != "") }
	faction, err := game.JapaneseInvasionFaction(sc.Stage)
	if err != nil {
		return bad()
	}
	for _, actor := range v.Actors {
		g, u, in := actor.General, actor.Unit, actor.Strength
		offshore := g.Province == game.ExternalInvasionSource && v.Armory != nil && v.Armory.Landing != nil
		if g.Raw != ([game.GeneralRecordSize]byte{}) || g.Province != u.Province || int(g.Province) > sc.Provinces && !offshore ||
			g.Force > 0 && !g.Province.Valid() && !offshore || u.General != actor.ID || in.General != actor.ID ||
			u.Faction != faction || in.Faction != faction || g.Force != in.Force || g.Branch != in.Branch ||
			game.BranchFullStrength(g.Branch) == 0 || g.Force > game.BranchFullStrength(g.Branch) ||
			g.AbilityA != in.Ability || g.F19 != in.F19 || g.F20 != in.F20 || g.Stamina != in.F29 || g.F30 != in.F30 ||
			g.Experience != u.Experience || g.AbilityA > 100 || g.AbilityB > 100 || g.AbilityC > 100 ||
			g.F19 > 100 || g.F20 > 100 || g.Stamina > 100 || g.F30 > 100 ||
			!u.Cell.Valid() && u.Cell != game.NoCell || !u.NextCell.Valid() && u.NextCell != game.NoCell ||
			u.TargetUnit != 0 && !validID(u.TargetUnit) {
			return bad()
		}
	}
	gens := make([]game.General, limit)
	units := make([]game.CombatUnit, limit)
	strengths := make([]game.StrengthInput, limit)
	copy(gens, v.Base.State.Generals)
	copy(units, v.Base.Units)
	copy(strengths, v.Base.Strengths)
	for i := sc.Generals; i < limit; i++ {
		gens[i], units[i], strengths[i] = campaignHole(game.GeneralID(i + 1))
	}
	for _, actor := range v.Actors {
		i := int(actor.ID) - 1
		gens[i], units[i], strengths[i] = actor.General, actor.Unit, actor.Strength
	}
	return gens, units, strengths, keys, nil
}

func (a *app) campaignV4State(base campaignSnapshot) (*campaignV4, error) {
	sc, err := game.ScenarioByStage(base.Identity.Stage)
	if err != nil {
		return nil, err
	}
	v := &campaignV4{Base: base, Armory: a.armory.Clone(), Declaration: a.declaration}
	if a.armory != nil {
		v.View = a.armoryViewSnapshot()
	}
	if a.militaryRules != nil {
		v.Rules = *a.militaryRules
	} else {
		v.Rules.Enabled = a.armory != nil
	}
	if v.Declaration == nil {
		v.Declaration, err = importedCampaignDeclaration(v.Rules)
		if err != nil {
			return nil, err
		}
	}
	limit := sc.Generals
	if sc.Stage == 1 && v.Declaration != nil && v.Declaration.Controller == 1 && len(a.campaignActors) > 0 {
		limit = 300
	}
	ids := make([]game.GeneralID, 0, len(a.campaignActors))
	for id := range a.campaignActors {
		ids = append(ids, id)
		limit = max(limit, int(id))
	}
	if len(base.State.Generals) != limit || len(base.Units) != limit || len(base.Strengths) != limit {
		return nil, fmt.Errorf("unregistered runtime actor or missing original prefix")
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		if campaignActorKey(id) != a.campaignActors[id] || campaignActorKey(id) == "" {
			return nil, fmt.Errorf("invalid runtime actor registration")
		}
		i := int(id) - 1
		v.Actors = append(v.Actors, campaignActor{Key: a.campaignActors[id], ID: id, General: base.State.Generals[i], Unit: base.Units[i], Strength: base.Strengths[i]})
	}
	for i := sc.Generals; i < limit; i++ {
		if a.campaignActors[game.GeneralID(i+1)] != "" {
			continue
		}
		g, u, in := campaignHole(game.GeneralID(i + 1))
		if base.State.Generals[i] != g || base.Units[i] != u || base.Strengths[i] != in {
			return nil, fmt.Errorf("runtime projection hole contains game state")
		}
	}
	v.Base.State.Generals = base.State.Generals[:sc.Generals]
	v.Base.Units = base.Units[:sc.Generals]
	v.Base.Strengths = base.Strengths[:sc.Generals]
	return v, nil
}
