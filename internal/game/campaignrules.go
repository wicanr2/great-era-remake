package game

import (
	"errors"
	"fmt"
)

var ErrMilitaryLocked = errors.New("game: military rules are locked after declaration")

// MilitaryRules is a remake option, independent of art and interface choices.
// DeclaredAt uses the game's Republican year*12+month, not a Gregorian year.
// The invasion event must set it only after its own transaction succeeds.
type MilitaryRules struct {
	Enabled    bool
	DeclaredAt uint32
}

func (r MilitaryRules) Locked() bool { return r.DeclaredAt != 0 }

func (r MilitaryRules) Valid(month uint32) bool {
	return month > 0 && month <= 65535*12+12 && (!r.Locked() || r.Enabled && r.DeclaredAt <= month)
}

func (r *MilitaryRules) SetEnabled(enabled bool) error {
	if r.Locked() && !enabled {
		return ErrMilitaryLocked
	}
	r.Enabled = enabled
	return nil
}

// JapaneseOffensiveAllowed leaves human Japan and ordinary campaigns alone.
// In the enabled Chinese campaign, Japan may defend but cannot expand before
// declaration. The caller must apply the gate to attacks AND neutral entry.
func (r MilitaryRules) JapaneseOffensiveAllowed(stage uint8, player, faction GeneralID) bool {
	return !r.Enabled || r.Locked() || stage == 1 || player == 107 || faction != 107
}

// UnificationStatus is derived from current ownership, never saved. It does not
// commit a declaration, change ownership or decide whether a UI transaction is safe.
type UnificationStatus struct {
	PlayerProvinces, JapaneseProvinces, UnownedProvinces, OtherChineseProvinces int
	JapanesePlayer, Unified, CanDeclare                                         bool
}

// AssessUnification implements the SPEC-63 user decision of 2026-10-18:
// unowned provinces are excluded, with no change to their data or ownership.
// The event caller must separately validate its safe point and commit all event
// state atomically; this predicate alone does not implement an invasion.
func (r MilitaryRules) AssessUnification(stage uint8, player GeneralID, table *ProvinceTable) (UnificationStatus, error) {
	sc, err := ScenarioByStage(stage)
	if err != nil {
		return UnificationStatus{}, err
	}
	if table == nil || player == 0 || int(player) > sc.Generals {
		return UnificationStatus{}, fmt.Errorf("game: invalid unification world or player")
	}
	s := UnificationStatus{JapanesePlayer: stage > 1 && player == 107}
	for i, p := range table.Province[:sc.Provinces] {
		if int(p.Commander) > sc.Generals {
			return UnificationStatus{}, fmt.Errorf("game: province %d has invalid commander %d", i+1, p.Commander)
		}
		switch {
		case p.Commander == 0:
			s.UnownedProvinces++
		case stage > 1 && p.Commander == 107:
			s.JapaneseProvinces++
		case p.Commander == player:
			s.PlayerProvinces++
		default:
			s.OtherChineseProvinces++
		}
	}
	s.Unified = !s.JapanesePlayer && s.PlayerProvinces > 0 && s.OtherChineseProvinces == 0
	s.CanDeclare = s.Unified && r.Enabled && !r.Locked()
	return s, nil
}
