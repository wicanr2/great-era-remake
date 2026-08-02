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

func TestAutonomyTargetsMatchLeaderLocationOwnershipAndCeasefire(t *testing.T) {
	w := realWorld(t)
	current := ProvinceID(1)
	p1, _ := w.Table.At(current)
	p1.Commander = 7
	p2, _ := w.Table.At(2)
	p2.Commander = 7
	p3, _ := w.Table.At(3)
	p3.Commander = 7
	p4, _ := w.Table.At(4)
	p4.Commander = 8
	w.Units = []CombatUnit{{General: 7, Province: current, Active: true}}
	w.CeasefireState[3] = 1

	targets, err := w.AutonomyTargets(current)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) != 1 || targets[0] != 2 {
		t.Fatalf("自治候選 = %v，預期只有省 2", targets)
	}
}

func TestAutonomyTargetsRequireLeaderInCurrentProvince(t *testing.T) {
	w := realWorld(t)
	p1, _ := w.Table.At(1)
	p2, _ := w.Table.At(2)
	p1.Commander, p2.Commander = 7, 7
	w.Units = []CombatUnit{{General: 7, Province: 2, Active: true}}
	if _, err := w.AutonomyTargets(1); err == nil {
		t.Fatal("司令不在本省時不應列出自治候選")
	}
}

func TestAutonomyTargetsRequireAnotherOwnedProvince(t *testing.T) {
	w := realWorld(t)
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		q, _ := w.Table.At(p)
		q.Commander = 8
	}
	p1, _ := w.Table.At(1)
	p1.Commander = 7
	w.Units = []CombatUnit{{General: 7, Province: 1, Active: true}}
	if _, err := w.AutonomyTargets(1); err == nil {
		t.Fatal("只有一省時不應允許授權自治")
	}
}

func TestTogglePlayerAutonomyRejectsNonCandidate(t *testing.T) {
	w := realWorld(t)
	p1, _ := w.Table.At(1)
	p2, _ := w.Table.At(2)
	p3, _ := w.Table.At(3)
	p1.Commander, p2.Commander, p3.Commander = 7, 7, 8
	w.Units = []CombatUnit{{General: 7, Province: 1, Active: true}}
	if _, err := w.TogglePlayerAutonomy(1, 3); err == nil {
		t.Fatal("不同司令的省份不應能切換自治")
	}
	if p3.Autonomous() {
		t.Fatal("被拒絕的省份狀態被改動")
	}
}

func TestRealSaveAutonomyTargetsForHubei(t *testing.T) {
	save := readGame(t, "SAVE(1).DT1")
	w := realWorld(t)
	gs, err := ParseSaveGenerals(save, 274)
	if err != nil {
		t.Fatal(err)
	}
	for i := range w.Units {
		w.Units[i].Province = gs[i].Province
	}
	w.CeasefireState, err = ParseCeasefireStates(save)
	if err != nil {
		t.Fatal(err)
	}
	targets, err := w.AutonomyTargets(26)
	if err != nil {
		t.Fatal(err)
	}
	if len(targets) == 0 {
		t.Fatal("SAVE(1) 的湖北應有可授權自治的其他省份")
	}
	t.Logf("SAVE(1) 湖北的自治候選：%v", targets)
}
