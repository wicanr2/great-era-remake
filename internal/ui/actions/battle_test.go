package actions

import "testing"

func TestBattleMoveRoundTrip(t *testing.T) {
	for d := 1; d <= 6; d++ {
		got, ok := BattleMoveDirection(BattleMove(d))
		if !ok || got != d {
			t.Fatalf("方向 %d 往返得到 (%d,%v)", d, got, ok)
		}
	}
	for _, action := range []Action{None, BattleMove(0), BattleMove(7), "battle.move.x", "battle.move.10"} {
		if _, ok := BattleMoveDirection(action); ok {
			t.Fatalf("非法動作 %q 不應被接受", action)
		}
	}
}
