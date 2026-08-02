package game

import "fmt"

// ProductionAllocation 是政策指令的本省產能百分比。
// 黃金沒有獨立欄位，而是 100 減去其餘四項。
type ProductionAllocation struct {
	Iron, Coal, Oil, Food uint8
}

func (p *Province) ProductionAllocation() ProductionAllocation {
	return ProductionAllocation{Iron: p.Raw[34], Coal: p.Raw[35], Oil: p.Raw[36], Food: p.Raw[33]}
}

func (a ProductionAllocation) Gold() uint8 {
	s := int(a.Iron) + int(a.Coal) + int(a.Oil) + int(a.Food)
	if s >= 100 {
		return 0
	}
	return uint8(100 - s)
}

func (a ProductionAllocation) Value(resource int) uint8 {
	switch resource {
	case 1:
		return a.Iron
	case 2:
		return a.Coal
	case 3:
		return a.Oil
	case 4:
		return a.Food
	}
	return 0
}

// SetProductionAllocation 對齊 sub_23511：只能分配當前黃金剩餘
// 加上該項舊值，不會自動擠壓其他產能。
func (p *Province) SetProductionAllocation(resource int, value uint8) error {
	a := p.ProductionAllocation()
	limit := int(a.Gold()) + int(a.Value(resource))
	if resource < 1 || resource > 4 {
		return fmt.Errorf("game: 產能項目 %d 超出 1..4", resource)
	}
	if int(value) > limit {
		return fmt.Errorf("game: 產能最多可調整為 %d%%", limit)
	}
	off := [...]int{0, 34, 35, 36, 33}[resource]
	p.Raw[off] = value
	return nil
}
