package game

import (
	"bytes"
	"reflect"
	"testing"
)

func TestInvasionLandingWaitsForNewWavesAndFinallyWithdraws(t *testing.T) {
	for _, stage := range []uint8{1, 2, 3} {
		w, army, gs, date := arrivalFixture(t, stage)
		w.Table.Province[6].Commander = 1 // Explicit alternate-state rule fixture.
		first := GeneralID(275)
		var survivors []GeneralID
		for number := uint8(1); number <= 3; number++ {
			plan, _ := FixedInvasionPlan(date.Era)
			for army.LastMonth < date.DeclaredAt+uint32(plan.Waves[number-1].MonthOffset) {
				if err := army.PauseMonth(w, army.LastMonth+1); err != nil {
					t.Fatal(err)
				}
				before := arrivalState(t, w, army, gs)
				if _, err := army.BeginInvasionLanding(w); err == nil || !bytes.Equal(before, arrivalState(t, w, army, gs)) {
					t.Fatal("month advance alone allowed a retry or changed state")
				}
			}
			var ids []GeneralID
			for range plan.Waves[number-1].Formations {
				ids = append(ids, first)
				first++
			}
			out, err := AssembleInvasionWave(w, army, gs, date, number, ids, ExternalInvasionSource)
			if err != nil {
				t.Fatal(err)
			}
			w, army, gs = out.World, out.Armory, out.Generals
			survivors = append(survivors, ids...)
			got, err := army.BeginInvasionLanding(w)
			if err != nil || !reflect.DeepEqual(got, survivors) {
				t.Fatal("survivors and reinforcement did not merge", got, err)
			}
			if _, err := army.BeginInvasionLanding(w); err == nil {
				t.Fatal("active attempt repeated")
			}
			if number == 1 {
				// The real loss account preserves damage and excludes dead units.
				army.ApplyLoss(275, w.Strengths[274].Force, 9000)
				gs[274].Force, w.Strengths[274].Force = 9000, 9000
				army.ApplyLoss(276, w.Strengths[275].Force, 0)
				gs[275].Force, w.Strengths[275].Force = 0, 0
				survivors = append(survivors[:1], survivors[2:]...)
			}
			settlement, _ := ClassifyBattleSettlement(16, 16, BattleSideNone)
			if err := army.FinishInvasionLanding(w, settlement); err != nil {
				t.Fatal(err)
			}
			want := LandingRepulsed
			if number == 3 {
				want = LandingWithdrawn
				for _, id := range survivors {
					if w.Units[id-1].Active {
						t.Fatal("final survivors remain active")
					}
				}
				w.Table.Province[6].Commander = army.InvasionBatches[1].Faction
				if _, err := army.DeliverInvasionSupply(w, 1); err == nil || army.SpendInvasionSupply(w, 1, InvasionSupply{}) == nil {
					t.Fatal("withdrawn external budget became available again")
				}
			}
			if army.Landing.Attempts[number-1].Outcome != want || gs[274].Force != 9000 || gs[275].Force != 0 {
				t.Fatal("outcome or losses were reset")
			}
			before := arrivalState(t, w, army, gs)
			if _, err := army.BeginInvasionLanding(w); err == nil || army.FinishInvasionLanding(w, settlement) == nil ||
				!bytes.Equal(before, arrivalState(t, w, army, gs)) {
				t.Fatal("settlement repeated or retry without fresh wave succeeded")
			}
		}
	}
}

func TestInvasionLandingRejectsDamagedHistoryAndCloneIsIndependent(t *testing.T) {
	w, army, gs, date := arrivalFixture(t, 2)
	w.Table.Province[6].Commander = 1
	out, err := AssembleInvasionWave(w, army, gs, date, 1, []GeneralID{275, 276, 277, 278, 279, 280}, ExternalInvasionSource)
	if err != nil {
		t.Fatal(err)
	}
	w, army = out.World, out.Armory
	if _, err := army.BeginInvasionLanding(w); err != nil {
		t.Fatal(err)
	}
	for _, change := range []string{"empty", "duplicate-id", "unknown-id", "future", "active-finished", "bad-outcome", "withdraw-first", "duplicate-attempt"} {
		x := army.Clone()
		switch change {
		case "empty":
			x.Landing.Attempts = nil
		case "duplicate-id":
			x.Landing.Attempts[0].Units[1] = 275
		case "unknown-id":
			x.Landing.Attempts[0].Units[0] = 274
		case "future":
			x.Landing.Attempts[0].StartedAt++
		case "active-finished":
			x.Landing.Attempts[0].FinishedAt = x.LastMonth
		case "bad-outcome":
			x.Landing.Attempts[0].Outcome = "unknown"
		case "withdraw-first":
			x.Landing.Attempts[0].Outcome, x.Landing.Attempts[0].FinishedAt = LandingWithdrawn, x.LastMonth
		case "duplicate-attempt":
			x.Landing.Attempts = append(x.Landing.Attempts, x.Landing.Attempts[0])
		}
		if x.Validate(w) == nil || army.Validate(w) != nil {
			t.Fatal("corruption accepted or clone changed source", change)
		}
	}
	settlement := BattleSettlement{Kind: BattleSettlementDecisive, Winner: BattleSideFirst}
	if err := army.FinishInvasionLanding(w, settlement); err != nil {
		t.Fatal(err)
	}
	if _, err := army.BeginInvasionLanding(w); err == nil {
		t.Fatal("successful landing restarted offshore")
	}
}
