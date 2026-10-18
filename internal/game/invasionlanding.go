package game

import "errors"

var ErrInvasionLanding = errors.New("game: invalid invasion landing or no new reinforcement")

type InvasionLandingOutcome string

const (
	LandingActive    InvasionLandingOutcome = "active"
	LandingRepulsed  InvasionLandingOutcome = "repulsed"
	LandingLanded    InvasionLandingOutcome = "landed"
	LandingWithdrawn InvasionLandingOutcome = "withdrawn"
)

type InvasionAttempt struct {
	Through               uint8
	StartedAt, FinishedAt uint32
	Outcome               InvasionLandingOutcome
	Units                 []GeneralID
}

// This optional ledger does not start an event when an old save is loaded.
type InvasionLanding struct {
	Attempts []InvasionAttempt
}

func (l *InvasionLanding) Clone() *InvasionLanding {
	if l == nil {
		return nil
	}
	out := &InvasionLanding{Attempts: append([]InvasionAttempt(nil), l.Attempts...)}
	for i := range out.Attempts {
		out.Attempts[i].Units = append([]GeneralID(nil), out.Attempts[i].Units...)
	}
	return out
}

func (a *Armory) validInvasionLanding() bool {
	if a.Landing == nil {
		return true
	}
	if len(a.Landing.Attempts) < 1 || len(a.Landing.Attempts) > 3 {
		return false
	}
	var through uint8
	var ended uint32
	for i, attempt := range a.Landing.Attempts {
		if attempt.Through <= through || int(attempt.Through) > len(a.InvasionBatches) ||
			attempt.StartedAt < ended || attempt.StartedAt > a.LastMonth ||
			len(attempt.Units) < 1 || len(attempt.Units) > UnitsPerSide {
			return false
		}
		permitted := map[GeneralID]bool{}
		for number := uint8(1); number <= attempt.Through; number++ {
			batch := a.InvasionBatches[number]
			if attempt.StartedAt < batch.ArrivedAt {
				return false
			}
			for _, id := range batch.Units {
				permitted[id] = true
			}
		}
		for _, id := range attempt.Units {
			if !permitted[id] {
				return false
			}
			delete(permitted, id)
		}
		last := i == len(a.Landing.Attempts)-1
		switch attempt.Outcome {
		case LandingActive:
			if !last || attempt.FinishedAt != 0 || int(attempt.Through) != len(a.InvasionBatches) {
				return false
			}
		case LandingRepulsed, LandingLanded, LandingWithdrawn:
			if attempt.FinishedAt < attempt.StartedAt || attempt.FinishedAt > a.LastMonth ||
				attempt.Outcome == LandingRepulsed && attempt.Through == 3 ||
				attempt.Outcome == LandingWithdrawn && attempt.Through != 3 ||
				attempt.Outcome != LandingRepulsed && !last {
				return false
			}
		default:
			return false
		}
		through, ended = attempt.Through, attempt.FinishedAt
	}
	for _, attempt := range a.Landing.Attempts {
		for _, count := range a.Landing.monthlyAttacks(attempt.StartedAt) {
			if count > JapaneseAttackLimit {
				return false
			}
		}
	}
	return true
}

func (l *InvasionLanding) monthlyAttacks(month uint32) map[GeneralID]uint8 {
	var used map[GeneralID]uint8
	if l == nil {
		return nil
	}
	for _, attempt := range l.Attempts {
		if attempt.StartedAt != month {
			continue
		}
		if used == nil {
			used = map[GeneralID]uint8{}
		}
		for _, id := range attempt.Units {
			used[id]++
		}
	}
	return used
}

// InvasionMonthlyAttacks derives a fresh, read-only basis for the shared blitz
// ledger. No separate persistent counter can drift from actual deployments.
func (a *Armory) InvasionMonthlyAttacks(month uint32) (map[GeneralID]uint8, error) {
	if month == 0 || month > 65536*12 {
		return nil, ErrInvasionLanding
	}
	if a == nil {
		return nil, nil
	}
	if !a.validInvasionLanding() {
		return nil, ErrInvasionLanding
	}
	return a.Landing.monthlyAttacks(month), nil
}

// BeginInvasionLanding uses survivors as they are. Only a new imported batch
// permits a second/third attempt; advancing months alone cannot reset this gate.
func (a *Armory) BeginInvasionLanding(w *AIWorld) ([]GeneralID, error) {
	if err := a.Validate(w); err != nil {
		return nil, err
	}
	through := uint8(len(a.InvasionBatches))
	if through == 0 {
		return nil, ErrInvasionLanding
	}
	if a.Landing != nil {
		last := a.Landing.Attempts[len(a.Landing.Attempts)-1]
		if last.Outcome != LandingRepulsed || through <= last.Through {
			return nil, ErrInvasionLanding
		}
	}
	p, err := w.Table.At(InvasionEntryProvince)
	faction := a.InvasionBatches[1].Faction
	if err != nil || p.Commander == faction || p.InBattle() {
		return nil, ErrInvasionLanding
	}
	var ids []GeneralID
	used, err := a.InvasionMonthlyAttacks(a.LastMonth)
	if err != nil {
		return nil, err
	}
	for number := uint8(1); number <= through; number++ {
		for _, id := range a.InvasionBatches[number].Units {
			u, in := w.Units[id-1], w.Strengths[id-1]
			if in.Force == 0 {
				continue
			}
			if !u.Active || u.Faction != faction || in.Faction != faction || u.Province != ExternalInvasionSource || used[id] >= JapaneseAttackLimit {
				return nil, ErrInvasionLanding
			}
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		return nil, ErrInvasionLanding
	}
	next := a.Clone()
	if next.Landing == nil {
		next.Landing = &InvasionLanding{}
	}
	next.Landing.Attempts = append(next.Landing.Attempts, InvasionAttempt{Through: through,
		StartedAt: a.LastMonth, Outcome: LandingActive, Units: append([]GeneralID(nil), ids...)})
	if err := next.Validate(w); err != nil {
		return nil, err
	}
	*a = *next
	return ids, nil
}

// FinishInvasionLanding records the real battle classification. The caller
// settles ownership and supply on private copies before committing the world.
func (a *Armory) FinishInvasionLanding(w *AIWorld, settlement BattleSettlement) error {
	if err := a.Validate(w); err != nil {
		return err
	}
	if a.Landing == nil {
		return ErrInvasionLanding
	}
	i := len(a.Landing.Attempts) - 1
	last := a.Landing.Attempts[i]
	if last.Outcome != LandingActive {
		return ErrInvasionLanding
	}
	switch settlement.Kind {
	case BattleSettlementDecisive:
		if settlement.Winner != BattleSideFirst && settlement.Winner != BattleSideSecond {
			return ErrInvasionLanding
		}
	case BattleSettlementTurnLimitDraw:
		if settlement.Winner != BattleSideNone || settlement.TurnCap < 1 || settlement.Turn < settlement.TurnCap {
			return ErrInvasionLanding
		}
	default:
		return ErrInvasionLanding
	}
	last.Outcome = LandingRepulsed
	if settlement.Kind == BattleSettlementDecisive && settlement.Winner == BattleSideFirst {
		last.Outcome = LandingLanded
	} else if last.Through == 3 {
		last.Outcome = LandingWithdrawn
	}
	last.FinishedAt = a.LastMonth
	next := a.Clone()
	next.Landing.Attempts[i] = last
	if err := next.Validate(w); err != nil {
		return err
	}
	*a = *next
	if last.Outcome == LandingWithdrawn {
		for _, id := range last.Units {
			w.Units[id-1].Active = false
		}
	}
	return nil
}

func (a *Armory) InvasionOffshoreSupply() (InvasionSupply, error) {
	var total [4]uint32
	if a == nil || len(a.InvasionBatches) == 0 {
		return InvasionSupply{}, ErrInvasionLanding
	}
	for number := uint8(1); int(number) <= len(a.InvasionBatches); number++ {
		s, err := a.InvasionRemainingSupply(number)
		if err != nil {
			return InvasionSupply{}, err
		}
		for i, v := range invasionSupplyValues(s) {
			total[i] += v
		}
	}
	return invasionSupplyFrom(total), nil
}
