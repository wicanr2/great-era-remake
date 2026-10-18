package game

import "testing"

func TestInvasionRosterMonthApprovedHistoricalBoundaries(t *testing.T) {
	// These are sourced exit facts, not full active-service intervals.
	// Source locators and raw SHA-256 records are in SPEC-63 and roster-month-*.
	for _, tt := range []struct {
		name   string
		year   uint16
		month  uint8
		exit   RosterExitDate
		before bool
	}{
		{"muto-june", 22, 6, RosterExitDate{1933, 7, 28}, true},
		{"muto-july-first", 22, 7, RosterExitDate{1933, 7, 28}, true},
		{"muto-august-first", 22, 8, RosterExitDate{1933, 7, 28}, false},
		{"ishiwara-february", 30, 2, RosterExitDate{1941, 3, 0}, true},
		{"ishiwara-march-excluded", 30, 3, RosterExitDate{1941, 3, 0}, false},
		{"ishiwara-april-excluded", 30, 4, RosterExitDate{1941, 3, 0}, false},
		{"anami-august-first", 34, 8, RosterExitDate{1945, 8, 15}, true},
		{"anami-frozen-august-reference", 34, 9, RosterExitDate{1945, 8, 15}, true},
		{"anami-late-alternate-reference", 65535, 12, RosterExitDate{1945, 8, 15}, true},
		{"exact-first-excluded", 22, 7, RosterExitDate{1933, 7, 1}, false},
		{"exact-second-included", 22, 7, RosterExitDate{1933, 7, 2}, true},
		{"month-precision-excluded", 22, 7, RosterExitDate{1933, 7, 0}, false},
		{"year-before-exit", 22, 12, RosterExitDate{1934, 1, 0}, true},
		{"year-exit-month", 23, 1, RosterExitDate{1934, 1, 0}, false},
		{"leap-day-included", 25, 2, RosterExitDate{1936, 2, 29}, true},
		{"leap-day-next-month", 25, 3, RosterExitDate{1936, 2, 29}, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			d, err := InvasionDateFor(tt.year, tt.month)
			if err != nil {
				t.Fatal(err)
			}
			original := d
			got, err := d.BeforeRosterExit(tt.exit)
			if err != nil || got != tt.before || d != original {
				t.Fatalf("got %v, %v, date %+v; want %v and unchanged %+v", got, err, d, tt.before, original)
			}
		})
	}
}

func TestInvasionRosterMonthRejectsUnknownAndForgedDates(t *testing.T) {
	d, _ := InvasionDateFor(34, 9)
	for _, exit := range []RosterExitDate{
		{}, {Year: 1945}, {1945, 0, 15}, {0, 8, 15}, {1945, 13, 0},
		{1945, 2, 29}, {1936, 2, 30}, {1945, 4, 31}, {1945, 8, 32},
	} {
		if before, err := d.BeforeRosterExit(exit); err == nil || before {
			t.Fatalf("accepted unknown/invalid exit %+v: %v, %v", exit, before, err)
		}
	}
	for _, mutate := range []func(*InvasionDate){
		func(v *InvasionDate) { v.RosterMonth = 9 },
		func(v *InvasionDate) { v.RosterYear = 1946 },
		func(v *InvasionDate) { v.Month = 8 },
		func(v *InvasionDate) { v.DeclaredAt = 0 },
		func(v *InvasionDate) { v.Era = InvasionEarly },
		func(v *InvasionDate) { v.AlternateHistory = false },
	} {
		bad := d
		mutate(&bad)
		if before, err := bad.BeforeRosterExit(RosterExitDate{1945, 8, 15}); err == nil || before {
			t.Fatalf("accepted forged reference %+v: %v, %v", bad, before, err)
		}
	}
}
