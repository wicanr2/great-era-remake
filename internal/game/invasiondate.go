package game

import (
	"fmt"
	"time"
)

// InvasionDate separates the actual declaration from the historical reference.
// After August 1945 the user-approved campaign is explicitly alternate history.
type InvasionDate struct {
	DeclaredAt       uint32 // Republican year*12+month, same as MilitaryRules.
	GregorianYear    uint32
	Month            uint8
	RosterYear       uint32
	RosterMonth      uint8
	Era              InvasionEra
	AlternateHistory bool
}

func InvasionDateFor(year uint16, month uint8) (InvasionDate, error) {
	if year < 15 || month < 1 || month > 12 {
		return InvasionDate{}, fmt.Errorf("game: invalid invasion date %d/%d", year, month)
	}
	y := uint32(year) + 1911
	d := InvasionDate{DeclaredAt: uint32(year)*12 + uint32(month), GregorianYear: y, Month: month, RosterYear: y, RosterMonth: month, Era: InvasionEarly}
	if y >= 1937 {
		d.Era = InvasionMid
	}
	if y >= 1941 {
		d.Era = InvasionLate
	}
	if y > 1945 || y == 1945 && month >= 9 {
		d.RosterYear, d.RosterMonth, d.AlternateHistory = 1945, 8, true
	}
	return d, nil
}

// RosterExitDate preserves the source precision of a retirement or death.
// Day == 0 means that only the month is known, never a guessed first day.
type RosterExitDate struct {
	Year  uint16 // Gregorian year.
	Month uint8
	Day   uint8
}

// BeforeRosterExit applies SPEC-63's approved first-of-month roster policy.
// True only passes this exclusion: it does not establish active service, rank,
// theatre, or eligibility for a production cast. Unknown dates return errors.
// The frozen reference is used, so an ongoing war never reselects its cast.
func (d InvasionDate) BeforeRosterExit(exit RosterExitDate) (bool, error) {
	if !canonicalInvasionDate(d) || exit.Year == 0 || exit.Month < 1 || exit.Month > 12 {
		return false, fmt.Errorf("game: invalid roster reference or unknown exit month")
	}
	if exit.Day > 0 {
		day := time.Date(int(exit.Year), time.Month(exit.Month), int(exit.Day), 0, 0, 0, 0, time.UTC)
		if day.Year() != int(exit.Year) || day.Month() != time.Month(exit.Month) || day.Day() != int(exit.Day) {
			return false, fmt.Errorf("game: invalid roster exit day")
		}
	}
	referenceMonth := d.RosterYear*12 + uint32(d.RosterMonth)
	exitMonth := uint32(exit.Year)*12 + uint32(exit.Month)
	if referenceMonth != exitMonth {
		return referenceMonth < exitMonth, nil
	}
	// The first is already outside service. A later exact day is still ahead
	// of the monthly snapshot; month-only exits conservatively exclude it.
	return exit.Day > 1, nil
}
