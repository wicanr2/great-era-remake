package game

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"sync"
)

//go:embed invasion-roster.json
var invasionRosterBytes []byte

type InvasionOfficer struct {
	ID  GeneralID
	Key string
}

type invasionServiceWindow struct {
	From uint32          `json:"from"`
	Exit *RosterExitDate `json:"exit"`
}

type invasionOfficerRecord struct {
	Key          string                  `json:"key"`
	ID           GeneralID               `json:"runtime_id"`
	Through      uint32                  `json:"verified_through"`
	Windows      []invasionServiceWindow `json:"windows"`
	SourceURL    string                  `json:"source_url"`
	SourceSHA256 string                  `json:"source_sha256"`
	Evidence     string                  `json:"evidence"`
}

var invasionRosterOnce sync.Once
var invasionRoster []invasionOfficerRecord
var invasionRosterError error

func loadInvasionRoster() {
	var catalog struct {
		Schema   int                     `json:"schema"`
		Officers []invasionOfficerRecord `json:"officers"`
	}
	d := json.NewDecoder(bytes.NewReader(invasionRosterBytes))
	d.DisallowUnknownFields()
	if err := d.Decode(&catalog); err != nil {
		invasionRosterError = err
		return
	}
	var extra any
	if err := d.Decode(&extra); err != io.EOF || catalog.Schema != 1 || len(catalog.Officers) != 19 {
		invasionRosterError = fmt.Errorf("game: invalid invasion roster catalog")
		return
	}
	seen := map[string]bool{}
	for i, p := range catalog.Officers {
		if p.ID != GeneralID(275+i) || p.Key == "" || seen[p.Key] || p.Through != 1945*12+8 ||
			len(p.SourceSHA256) != 64 || p.SourceURL == "" || p.Evidence == "" || len(p.Windows) == 0 {
			invasionRosterError = fmt.Errorf("game: invalid invasion officer %d", p.ID)
			return
		}
		seen[p.Key] = true
		var last uint32
		for _, w := range p.Windows {
			if w.From < 1926*12+1 || w.From <= last || w.From > p.Through {
				invasionRosterError = fmt.Errorf("game: invalid service coverage %d", p.ID)
				return
			}
			last = w.From
			if w.Exit != nil {
				probe, _ := InvasionDateFor(15, 1)
				if _, err := probe.BeforeRosterExit(*w.Exit); err != nil || uint32(w.Exit.Year)*12+uint32(w.Exit.Month) < w.From {
					invasionRosterError = fmt.Errorf("game: invalid exit coverage %d", p.ID)
					return
				}
				last = uint32(w.Exit.Year)*12 + uint32(w.Exit.Month)
			}
		}
	}
	invasionRoster = catalog.Officers
}

// InvasionActorKey is the single append-only storage identity registry.
func InvasionActorKey(id GeneralID) string {
	invasionRosterOnce.Do(loadInvasionRoster)
	if invasionRosterError != nil || id < 275 || int(id)-275 >= len(invasionRoster) {
		return ""
	}
	return invasionRoster[int(id)-275].Key
}

func invasionOfficerEligible(p invasionOfficerRecord, d InvasionDate) (bool, error) {
	reference := d.RosterYear*12 + uint32(d.RosterMonth)
	if reference > p.Through {
		return false, fmt.Errorf("game: unreviewed roster month")
	}
	for _, window := range p.Windows {
		if reference < window.From {
			continue
		}
		if window.Exit == nil {
			return true, nil
		}
		before, err := d.BeforeRosterExit(*window.Exit)
		if err != nil {
			return false, err
		}
		if before {
			return true, nil
		}
	}
	return false, nil
}

func InvasionOfficerEligible(id GeneralID, d InvasionDate) (bool, error) {
	invasionRosterOnce.Do(loadInvasionRoster)
	if invasionRosterError != nil {
		return false, invasionRosterError
	}
	if !canonicalInvasionDate(d) || InvasionActorKey(id) == "" {
		return false, fmt.Errorf("game: invalid cast date or officer")
	}
	return invasionOfficerEligible(invasionRoster[int(id)-275], d)
}

// SelectInvasionCast freezes ten distinct identities in catalog order. It never
// generates forces, consumes RNG, or derives historical rank from game stats.
func SelectInvasionCast(d InvasionDate) ([]InvasionOfficer, error) {
	invasionRosterOnce.Do(loadInvasionRoster)
	if invasionRosterError != nil {
		return nil, invasionRosterError
	}
	if !canonicalInvasionDate(d) {
		return nil, fmt.Errorf("game: invalid cast date")
	}
	var out []InvasionOfficer
	for _, p := range invasionRoster {
		ok, err := invasionOfficerEligible(p, d)
		if err != nil {
			return nil, err
		}
		if ok {
			out = append(out, InvasionOfficer{ID: p.ID, Key: p.Key})
		}
	}
	if len(out) < 10 {
		return nil, fmt.Errorf("game: fewer than ten reviewed officers")
	}
	return out[:10], nil
}
