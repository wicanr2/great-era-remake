package game

import "testing"

// 授權自治是 toggle：同一個指令開也關。
func TestToggleAutonomy(t *testing.T) {
	w := realWorld(t)
	prov, err := w.Table.At(1)
	if err != nil {
		t.Fatal(err)
	}
	prov.Commander = 7
	prov.Flags = 0

	on, err := w.ToggleAutonomy(1)
	if err != nil {
		t.Fatal(err)
	}
	if !on || !prov.Autonomous() {
		t.Error("第一次切換應該變成自治")
	}
	if prov.Flags != ProvinceFlagAutonomous {
		t.Errorf("旗標 %#x，預期 %#x", prov.Flags, ProvinceFlagAutonomous)
	}

	on, err = w.ToggleAutonomy(1)
	if err != nil {
		t.Fatal(err)
	}
	if on || prov.Autonomous() {
		t.Error("第二次切換應該取消自治")
	}
}

// 切換不該動到其他位元。
func TestToggleAutonomyKeepsOtherFlags(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	prov.Commander = 7
	prov.Flags = ProvinceFlagActed | ProvinceFlagInBattle | ProvinceFlagTaxed

	if _, err := w.ToggleAutonomy(1); err != nil {
		t.Fatal(err)
	}
	for name, bit := range map[string]uint8{
		"已處理": ProvinceFlagActed, "交戰中": ProvinceFlagInBattle,
		"已徵稅": ProvinceFlagTaxed,
	} {
		if prov.Flags&bit == 0 {
			t.Errorf("%s 的位元被切換自治動到了", name)
		}
	}
	if !prov.Autonomous() {
		t.Error("自治位元沒設起來")
	}
}

// 無主省談不上自治。
func TestToggleAutonomyNeedsCommander(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	prov.Commander = 0
	if _, err := w.ToggleAutonomy(1); err == nil {
		t.Error("無主省應該報錯")
	}
}

// 自治的省一個月只能下一個指令——下完剩餘數歸零。
func TestAutonomousCommandCap(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	prov.Commander = 7

	prov.Flags = 0
	if got := w.CommandsRemainingAfter(1, 3); got != 2 {
		t.Errorf("非自治省：3 個指令用掉 1 剩 %d，預期 2", got)
	}

	prov.Flags = ProvinceFlagAutonomous
	if got := w.CommandsRemainingAfter(1, 3); got != 0 {
		t.Errorf("自治省：不論剩幾個都該歸零，得到 %d", got)
	}
	if AutonomousCommandCap != 1 {
		t.Errorf("自治省的指令上限是 %d，原版是 1", AutonomousCommandCap)
	}
}
