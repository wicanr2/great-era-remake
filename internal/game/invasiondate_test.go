package game

import "testing"

func TestInvasionDateApprovedBoundaries(t *testing.T) {
	for _, tt := range []struct {
		year        uint16
		month       uint8
		era         InvasionEra
		rosterYear  uint32
		rosterMonth uint8
		alternate   bool
	}{
		{15, 7, InvasionEarly, 1926, 7, false},
		{25, 12, InvasionEarly, 1936, 12, false},
		{26, 1, InvasionMid, 1937, 1, false},
		{29, 12, InvasionMid, 1940, 12, false},
		{30, 1, InvasionLate, 1941, 1, false},
		{34, 8, InvasionLate, 1945, 8, false},
		{34, 9, InvasionLate, 1945, 8, true},
		{34, 12, InvasionLate, 1945, 8, true},
		{65535, 12, InvasionLate, 1945, 8, true},
	} {
		d, err := InvasionDateFor(tt.year, tt.month)
		if err != nil {
			t.Fatal(err)
		}
		if d.DeclaredAt != uint32(tt.year)*12+uint32(tt.month) || d.GregorianYear != uint32(tt.year)+1911 || d.Month != tt.month || d.Era != tt.era || d.RosterYear != tt.rosterYear || d.RosterMonth != tt.rosterMonth || d.AlternateHistory != tt.alternate {
			t.Fatalf("%+v: %+v", tt, d)
		}
		if _, err := FixedInvasionPlan(d.Era); err != nil {
			t.Fatal(err)
		}
	}
}

func TestInvasionDateRejectsInvalidInput(t *testing.T) {
	for _, tt := range []struct {
		y uint16
		m uint8
	}{{0, 1}, {14, 12}, {15, 0}, {15, 13}, {65535, 255}} {
		if _, err := InvasionDateFor(tt.y, tt.m); err == nil {
			t.Fatalf("accepted %+v", tt)
		}
	}
}
