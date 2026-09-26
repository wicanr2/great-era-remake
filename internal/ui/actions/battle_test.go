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

func TestBattleCommandRoundTrip(t *testing.T) {
	for n := 1; n <= 5; n++ {
		got, ok := BattleCommandNumber(BattleCommand(n))
		if !ok || got != n {
			t.Fatalf("命令 %d 往返得到 (%d,%v)", n, got, ok)
		}
	}
	for _, action := range []Action{None, BattleCommand(0), BattleCommand(6), "battle.command.x"} {
		if _, ok := BattleCommandNumber(action); ok {
			t.Fatalf("非法命令動作 %q 不應被接受", action)
		}
	}
}

// 原版第二層選單的第 6 鍵沒有 handler；remake 的第 6 個攻擊動作若出現，
// 只代表「第 6 個目標序號」，不能被誤當成第 6 種命名攻擊。
func TestApproximateKey6KeepsCommandAndTargetNamespacesSeparate(t *testing.T) {
	if _, ok := BattleCommandNumber(BattleCommand(6)); ok {
		t.Fatal("第二層第 6 鍵不應被包成可執行的 battle command")
	}
	if got, ok := BattleAttackTargetNumber(BattleAttackTarget(6)); !ok || got != 6 {
		t.Fatalf("第 6 個攻擊目標序號應保留在目標 namespace，得到 %d/%v", got, ok)
	}
}

func TestBattleAttackTargetRoundTrip(t *testing.T) {
	for n := 1; n <= 6; n++ {
		got, ok := BattleAttackTargetNumber(BattleAttackTarget(n))
		if !ok || got != n {
			t.Fatalf("攻擊目標 %d 往返得到 (%d,%v)", n, got, ok)
		}
	}
	for _, action := range []Action{None, BattleAttackTarget(0), BattleAttackTarget(7),
		"battle.attack-target.x", "battle.attack-target.10"} {
		if _, ok := BattleAttackTargetNumber(action); ok {
			t.Fatalf("非法攻擊目標動作 %q 不應被接受", action)
		}
	}
}
