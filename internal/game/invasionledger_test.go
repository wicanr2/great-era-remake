package game

import (
	"bytes"
	"testing"
)

func TestInvasionLedgerDelayedThreeWavesShareMonthlyLimit(t *testing.T) {
	for _, stage := range []uint8{1, 2, 3} {
		w, army, gs, date := arrivalFixture(t, stage)
		date, _ = InvasionDateFor(uint16((date.DeclaredAt-1)/12), 6)
		w.Table.Province[6].Commander = 1
		for army.LastMonth < date.DeclaredAt+6 {
			if err := army.PauseMonth(w, army.LastMonth+1); err != nil {
				t.Fatal(err)
			}
		}
		month := army.LastMonth
		plan, _ := FixedInvasionPlan(date.Era)
		id := GeneralID(275)
		for number := uint8(1); number <= 3; number++ {
			var ids []GeneralID
			for range plan.Waves[number-1].Formations {
				ids = append(ids, id)
				id++
			}
			out, err := AssembleInvasionWave(w, army, gs, date, number, ids, ExternalInvasionSource)
			if err != nil {
				t.Fatal(err)
			}
			w, army, gs = out.World, out.Armory, out.Generals
			if number == 3 {
				before := arrivalState(t, w, army, gs)
				if _, err := army.BeginInvasionLanding(w); err == nil || !bytes.Equal(before, arrivalState(t, w, army, gs)) {
					t.Fatal("delayed third wave bypassed the shared monthly limit")
				}
				if err := army.PauseMonth(w, army.LastMonth+1); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := army.BeginInvasionLanding(w); err != nil {
				t.Fatal(err)
			}
			settlement, _ := ClassifyBattleSettlement(16, 16, BattleSideNone)
			if err := army.FinishInvasionLanding(w, settlement); err != nil {
				t.Fatal(err)
			}
		}
		before := arrivalState(t, w, army, gs)
		used, err := army.InvasionMonthlyAttacks(month)
		if err != nil || used[275] != 2 || used[281] != 1 || used[283] != 0 {
			t.Fatal("wrong monthly basis", used, err)
		}
		current, err := army.InvasionMonthlyAttacks(month + 1)
		if err != nil || len(current) != 10 || current[275] != 1 {
			t.Fatal("new month inherited old counts", current, err)
		}
		used[275] = 0
		again, _ := army.InvasionMonthlyAttacks(month)
		if again[275] != 2 || !bytes.Equal(before, arrivalState(t, w, army, gs)) {
			t.Fatal("basis query changed persistent state")
		}
		bad := army.Clone()
		bad.Landing.Attempts[2].StartedAt, bad.Landing.Attempts[2].FinishedAt = month, month
		if bad.Validate(w) == nil {
			t.Fatal("same-month third deployment accepted in saved history")
		}
	}
}

func TestInvasionLedgerMonthBoundariesAndLegacy(t *testing.T) {
	var absent *Armory
	if used, err := absent.InvasionMonthlyAttacks(1); err != nil || used != nil {
		t.Fatal("nil legacy armory changed")
	}
	for _, month := range []uint32{0, 65536*12 + 1} {
		if _, err := absent.InvasionMonthlyAttacks(month); err == nil {
			t.Fatal("invalid month accepted")
		}
	}
}
