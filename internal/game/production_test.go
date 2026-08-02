package game

import "testing"

func TestProductionAllocationMappingAndLimit(t *testing.T) {
	var p Province
	p.Raw[33], p.Raw[34], p.Raw[35], p.Raw[36] = 25, 25, 25, 25
	a := p.ProductionAllocation()
	if a != (ProductionAllocation{Iron: 25, Coal: 25, Oil: 25, Food: 25}) || a.Gold() != 0 {
		t.Fatalf("allocation=%+v gold=%d", a, a.Gold())
	}
	if err := p.SetProductionAllocation(1, 20); err != nil {
		t.Fatal(err)
	}
	if got := p.ProductionAllocation().Gold(); got != 5 {
		t.Fatalf("gold=%d, want 5", got)
	}
	if err := p.SetProductionAllocation(2, 31); err == nil {
		t.Fatal("expected limit error")
	}
	if err := p.SetProductionAllocation(2, 30); err != nil {
		t.Fatal(err)
	}
	if got := p.Bytes()[35]; got != 30 {
		t.Fatalf("saved coal=%d", got)
	}
}
