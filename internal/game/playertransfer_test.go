package game

import "testing"

func playerTransferWorld(t *testing.T) (*AIWorld, []General) {
	t.Helper()
	tbl := &ProvinceTable{}
	src, _ := tbl.At(1)
	dst, _ := tbl.At(2)
	enemy, _ := tbl.At(3)
	src.Commander, src.Neighbours = 1, []ProvinceID{2, 3, 4, 5}
	dst.Commander = 1
	enemy.Commander = 9
	w := &AIWorld{Table: tbl, Units: []CombatUnit{
		{General: 1, Active: true, Province: 1, Faction: 1},
		{General: 2, Active: true, Province: 1, Faction: 1},
		{General: 3, Active: true, Province: 1, Faction: 1},
	}, Strengths: make([]StrengthInput, 3)}
	gens := []General{{Province: 1, Stamina: 30}, {Province: 1, Stamina: 29}, {Province: 1, Stamina: 80}}
	return w, gens
}

func TestApplyPlayerTransferIsAtomicAndSynchronizesState(t *testing.T) {
	w, gens := playerTransferWorld(t)
	src, _ := w.Table.At(1)
	dst, _ := w.Table.At(2)
	src.Gold, src.Food, src.Ammo, src.Fuel = 100, 200, 300, 400
	dst.Gold, dst.Food, dst.Ammo, dst.Fuel = 59990, 20, 30, 40
	dst.Flags = ProvinceFlagAutonomous
	s, err := w.BeginPlayerTransfer(1, 2, PlayerTransferAll, gens)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.ApplyPlayerTransfer(s, gens, [4]int{101, 0, 0, 0}); err == nil {
		t.Fatal("物資超額應拒絕")
	}
	if src.Gold != 100 || dst.Gold != 59990 || gens[0].Province != 1 || gens[0].Stamina != 30 {
		t.Fatal("驗證失敗後不應留下部分變更")
	}
	rep, err := w.ApplyPlayerTransfer(s, gens, [4]int{25, 50, 75, 100})
	if err != nil {
		t.Fatal(err)
	}
	if len(rep.Moved) != 2 || rep.Moved[0] != 1 || rep.Moved[1] != 3 {
		t.Fatalf("搬遷將領 = %v，預期 [1 3]", rep.Moved)
	}
	if rep.Supplies != [4]int{10, 50, 75, 100} {
		t.Fatalf("實際物資 = %v，預期黃金上限退回後 [10 50 75 100]", rep.Supplies)
	}
	if gens[0].Province != 2 || gens[0].Stamina != 20 || gens[2].Province != 2 || gens[2].Stamina != 70 {
		t.Fatalf("將領狀態未同步：%+v / %+v", gens[0], gens[2])
	}
	if w.Units[0].Province != 2 || w.Units[2].Province != 2 || w.Strengths[0].F29 != 20 || w.Strengths[2].F29 != 70 {
		t.Fatal("執行期單位／戰力輸入未同步")
	}
	if src.Commander != 1 || src.Governor != 2 {
		t.Fatalf("來源領導 = %d/%d，預期司令 1、省長 2", src.Commander, src.Governor)
	}
	if dst.Commander != 1 || dst.Governor != 1 || dst.Autonomous() {
		t.Fatalf("目標領導或自治旗標錯誤：%d/%d flags=%02x", dst.Commander, dst.Governor, dst.Flags)
	}
}

func TestApplyPlayerTransferClearsEmptySource(t *testing.T) {
	w, gens := playerTransferWorld(t)
	// 讓三人都達門檻，以「全部調動」搬空來源。
	gens[1].Stamina = 30
	src, _ := w.Table.At(1)
	src.Governor, src.Flags = 3, ProvinceFlagAutonomous
	s, err := w.BeginPlayerTransfer(1, 2, PlayerTransferAll, gens)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := w.ApplyPlayerTransfer(s, gens, [4]int{}); err != nil {
		t.Fatal(err)
	}
	if src.Commander != 0 || src.Governor != 0 || src.Autonomous() {
		t.Fatalf("搬空來源後 = 司令%d 省長%d flags=%02x", src.Commander, src.Governor, src.Flags)
	}
}

func TestPlayerTransferGovernorUsesOriginalTwoPassOrder(t *testing.T) {
	w, gens := playerTransferWorld(t)
	// 領袖不在來源；忠誠度是第二輪排序的主要鍵，政治手腕只決定同分順序。
	w.Units[0].Active = false
	gens[1].AbilityB, gens[1].AbilityC = 80, 90
	gens[2].AbilityB, gens[2].AbilityC = 80, 50
	w.repairPlayerTransferProvince(1, 1, gens)
	p, _ := w.Table.At(1)
	if p.Commander != 1 || p.Governor != 2 {
		t.Fatalf("司令／省長 = %d/%d，預期 1/2", p.Commander, p.Governor)
	}
}

func TestPlayerTransferSurvivesSaveReload(t *testing.T) {
	w, gens := playerTransferWorld(t)
	s, err := w.BeginPlayerTransfer(1, 2, PlayerTransferPartial, gens)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Toggle(3); err != nil {
		t.Fatal(err)
	}
	if _, err := w.ApplyPlayerTransfer(s, gens, [4]int{}); err != nil {
		t.Fatal(err)
	}
	orig := make([]byte, SaveGeneralsOffset+len(gens)*GeneralRecordSize)
	out, err := WriteSave(orig, w.Table, gens)
	if err != nil {
		t.Fatal(err)
	}
	provinces, err := ParseSaveProvinces(out)
	if err != nil {
		t.Fatal(err)
	}
	reloaded, err := ParseSaveGenerals(out, len(gens))
	if err != nil {
		t.Fatal(err)
	}
	src, _ := provinces.At(1)
	dst, _ := provinces.At(2)
	if reloaded[2].Province != 2 || reloaded[2].Stamina != 70 {
		t.Fatalf("重載後將領 3 = 省%d／體力%d", reloaded[2].Province, reloaded[2].Stamina)
	}
	if src.Commander != 1 || src.Governor != 1 || dst.Commander != 1 || dst.Governor != 3 {
		t.Fatalf("重載後來源 %d/%d、目標 %d/%d", src.Commander, src.Governor, dst.Commander, dst.Governor)
	}
}

func TestApplyPlayerTransferRechecksTargetCapacity(t *testing.T) {
	w, gens := playerTransferWorld(t)
	// 建立工作階段時尚有空位，確認前目標被其他狀態更新塞滿。
	for i := 0; i < TransferTargetGeneralCap-1; i++ {
		w.Units = append(w.Units, CombatUnit{General: GeneralID(1000 + i), Active: true, Province: 2, Faction: 1})
	}
	s, err := w.BeginPlayerTransfer(1, 2, PlayerTransferPartial, gens)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Toggle(1); err != nil {
		t.Fatal(err)
	}
	w.Units = append(w.Units, CombatUnit{General: 2000, Active: true, Province: 2, Faction: 1})
	if _, err := w.ApplyPlayerTransfer(s, gens, [4]int{}); err == nil {
		t.Fatal("確認前目標已滿，應拒絕套用")
	}
	if gens[0].Province != 1 || w.Units[0].Province != 1 {
		t.Fatal("容量競態失敗後不應搬動將領")
	}
}

func TestPlayerTransferTargetsUseOriginalFilters(t *testing.T) {
	w, _ := playerTransferWorld(t)
	p4, _ := w.Table.At(4)
	p5, _ := w.Table.At(5)
	p4.Commander = 0
	p5.Commander, p5.Flags = 1, ProvinceFlagInBattle
	got, err := w.PlayerTransferTargets(1)
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 || got[0] != 2 || got[1] != 4 {
		t.Fatalf("合法目標 = %v，預期 [2 4]", got)
	}
	p4.Flags = provinceFlagTransferBlocked
	got, _ = w.PlayerTransferTargets(1)
	if len(got) != 1 || got[0] != 2 {
		t.Fatalf("bit 4 阻擋後目標 = %v，預期 [2]", got)
	}
}

func TestPlayerTransferPartialStartsEmptyAndChecksStamina(t *testing.T) {
	w, gens := playerTransferWorld(t)
	s, err := w.BeginPlayerTransfer(1, 2, PlayerTransferPartial, gens)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Selected(); len(got) != 0 {
		t.Fatalf("部分調動初始選取 = %v，預期空集合", got)
	}
	if err := s.Toggle(2); err == nil {
		t.Fatal("體力 29 的將領不應能勾選")
	}
	if err := s.Toggle(1); err != nil {
		t.Fatal(err)
	}
	if got := s.Selected(); len(got) != 1 || got[0] != 1 {
		t.Fatalf("選取 = %v，預期 [1]", got)
	}
}

func TestPlayerTransferAllPreselectsEligibleAndCancelsStayBehind(t *testing.T) {
	w, gens := playerTransferWorld(t)
	s, err := w.BeginPlayerTransfer(1, 2, PlayerTransferAll, gens)
	if err != nil {
		t.Fatal(err)
	}
	if got := s.Selected(); len(got) != 2 || got[0] != 1 || got[1] != 3 {
		t.Fatalf("全部調動預選 = %v，預期 [1 3]", got)
	}
	if err := s.Toggle(1); err != nil {
		t.Fatal(err)
	}
	if got := s.Selected(); len(got) != 1 || got[0] != 3 {
		t.Fatalf("指定 1 留守後搬遷 = %v，預期 [3]", got)
	}
}
