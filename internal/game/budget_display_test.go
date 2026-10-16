package game

import "testing"

func TestBudgetDisplayDoesNotInitializeOrSpend(t *testing.T) {
	limit := 3
	b := NewCommandBudgetFunc(func(ProvinceID) int { return limit })
	for i := 0; i < 3; i++ {
		if b.Peek(1) != 3 || len(b.Snapshot()) != 0 {
			t.Fatal("display mutated budget")
		}
	}
	if b.Remaining(1) != 3 {
		t.Fatal("first command lost capacity")
	}
	limit = 5
	if b.Peek(1) != 3 || !b.Spend(1) || b.Peek(1) != 2 {
		t.Fatal("display ignored latched or spent budget")
	}
}
