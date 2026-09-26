package game

import "testing"

func TestClassifyBattleSettlementDecisiveDoesNotRequestDT2Writeback(t *testing.T) {
	for _, winner := range []BattleSide{BattleSideFirst, BattleSideSecond} {
		got, err := ClassifyBattleSettlement(3, BattleTurnLimit, winner)
		if err != nil {
			t.Fatalf("勝方 %d 分類失敗：%v", winner, err)
		}
		if got.Kind != BattleSettlementDecisive || got.Winner != winner || got.RequiresDT2Writeback {
			t.Fatalf("勝方 %d 分類=%+v，預期 decisive 且不寫回 DT2", winner, got)
		}
	}
}

func TestClassifyBattleSettlementTurnLimitUses15Or16AndRequestsDT2Writeback(t *testing.T) {
	for _, tc := range []struct {
		name string
		turn int
		cap  int
	}{
		{name: "一般月份", turn: BattleTurnLimit, cap: BattleTurnLimit},
		{name: "二月", turn: BattleTurnLimitAlt, cap: BattleTurnLimitAlt},
		{name: "超過上限仍是同一平局", turn: BattleTurnLimit + 2, cap: BattleTurnLimit},
	} {
		got, err := ClassifyBattleSettlement(tc.turn, tc.cap, BattleSideNone)
		if err != nil {
			t.Fatalf("%s 分類失敗：%v", tc.name, err)
		}
		if got.Kind != BattleSettlementTurnLimitDraw || got.Winner != BattleSideNone || !got.RequiresDT2Writeback {
			t.Fatalf("%s 分類=%+v，預期回合上限平局且需要 DT2 寫回", tc.name, got)
		}
		if got.Turn != tc.turn || got.TurnCap != tc.cap {
			t.Fatalf("%s 沒保留回合資訊：%+v", tc.name, got)
		}
	}
}

func TestClassifyBattleSettlementRejectsUnfinishedAndInvalidValues(t *testing.T) {
	for _, tc := range []struct {
		name string
		turn int
		cap  int
		win  BattleSide
	}{
		{name: "尚未到上限", turn: 1, cap: BattleTurnLimit, win: BattleSideNone},
		{name: "零上限", turn: 1, cap: 0, win: BattleSideNone},
		{name: "未知勝方", turn: BattleTurnLimit, cap: BattleTurnLimit, win: BattleSide(3)},
	} {
		if got, err := ClassifyBattleSettlement(tc.turn, tc.cap, tc.win); err == nil || got.Kind != BattleSettlementUnknown {
			t.Errorf("%s 應 fail-closed：got=%+v err=%v", tc.name, got, err)
		}
	}
}

func TestBattleSettlementImmediateRetreatDoesNotWriteBack(t *testing.T) {
	got := BattleSettlementImmediateRetreat()
	if got.Kind != BattleSettlementRetreat || got.Winner != BattleSideNone || got.RequiresDT2Writeback {
		t.Fatalf("立即撤退分類=%+v，預期不寫回 DT2", got)
	}
}

func TestBattleSettlementKindString(t *testing.T) {
	for _, tc := range []struct {
		kind BattleSettlementKind
		want string
	}{
		{BattleSettlementUnknown, "unknown"},
		{BattleSettlementRetreat, "immediate-retreat"},
		{BattleSettlementDecisive, "decisive"},
		{BattleSettlementTurnLimitDraw, "turn-limit-draw"},
		{BattleSettlementKind(99), "unknown"},
	} {
		if got := tc.kind.String(); got != tc.want {
			t.Errorf("kind %d String()=%q，預期 %q", tc.kind, got, tc.want)
		}
	}
}
