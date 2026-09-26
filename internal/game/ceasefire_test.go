package game

import (
	"encoding/binary"
	"testing"
)

func TestParseCeasefireStatesIsOneBasedAndChecksLength(t *testing.T) {
	b, err := SaveBlockByGlobal("byte_6FE56")
	if err != nil {
		t.Fatal(err)
	}
	save := make([]byte, b.Offset+b.Size)
	save[b.Offset], save[b.Offset+b.Size-1] = 1, 9
	states, err := ParseCeasefireStates(save)
	if err != nil {
		t.Fatal(err)
	}
	if states[0] != 0 || states[1] != 1 || states[ProvinceCount] != 9 {
		t.Fatalf("停火表 1-based 映射錯誤：0=%d 1=%d 39=%d",
			states[0], states[1], states[ProvinceCount])
	}
	if _, err := ParseCeasefireStates(save[:len(save)-1]); err == nil {
		t.Fatal("過短存檔應該報錯")
	}
}

func TestWriteCeasefireStatesChangesOnlyTheRequestedTable(t *testing.T) {
	orig := readGame(t, "SAVE(1).DT1")
	states, err := ParseCeasefireStates(orig)
	if err != nil {
		t.Fatal(err)
	}
	const province = ProvinceID(7)
	states[province] = 23
	out, err := WriteCeasefireStates(orig, states)
	if err != nil {
		t.Fatal(err)
	}
	b, err := SaveBlockByGlobal("byte_6FE56")
	if err != nil {
		t.Fatal(err)
	}
	want := b.Offset + int(province) - 1
	if d := DiffBytes(orig, out, 8); len(d) != 1 || d[0] != want {
		t.Fatalf("只改省 %d 的停火 byte，實際差分：%v（預期 offset %d）", province, d, want)
	}
	back, err := ParseCeasefireStates(out)
	if err != nil {
		t.Fatal(err)
	}
	if back[province] != 23 || back[0] != 0 {
		t.Fatalf("停火表寫回錯誤：省 %d=%d、哨兵=%d", province, back[province], back[0])
	}
}

func TestParseWarRecordsKeepsOneBasedLeadersAndRawBytes(t *testing.T) {
	blk, err := SaveBlockByGlobal("byte_6F532")
	if err != nil {
		t.Fatal(err)
	}
	save := make([]byte, blk.Offset+blk.Size)
	const province = ProvinceID(7)
	base := blk.Offset + int(province-1)*WarRecordSize
	// sub_3964E 的 +0／+2 是兩個勢力領袖 ID；其餘 bytes 應照原樣保留。
	save[base], save[base+1] = 0x34, 0x12
	save[base+2], save[base+3] = 0x78, 0x56
	binary.LittleEndian.PutUint16(save[base+0x04:], 0x1101)
	binary.LittleEndian.PutUint16(save[base+0x06:], 0x1102)
	binary.LittleEndian.PutUint16(save[base+0x0C:], 0x1103)
	binary.LittleEndian.PutUint16(save[base+0x10:], 0x1104)
	binary.LittleEndian.PutUint16(save[base+0x08:], 0x2201)
	binary.LittleEndian.PutUint16(save[base+0x0A:], 0x2202)
	binary.LittleEndian.PutUint16(save[base+0x0E:], 0x2203)
	binary.LittleEndian.PutUint16(save[base+WarRecordGeneralList18Offset:], 0x0102)
	binary.LittleEndian.PutUint16(save[base+WarRecordGeneralList18Offset+18:], 0x0304)
	binary.LittleEndian.PutUint16(save[base+WarRecordGeneralList38Offset:], 0x0506)
	binary.LittleEndian.PutUint16(save[base+WarRecordGeneralList38Offset+18:], 0x0708)
	save[base+59] = 0xA5
	records, err := ParseWarRecords(save)
	if err != nil {
		t.Fatal(err)
	}
	if records[0].SideALeader != 0 || records[0].SideBLeader != 0 {
		t.Error("index 0 應保留為無效哨兵")
	}
	got := records[province]
	if got.SideALeader != 0x1234 || got.SideBLeader != 0x5678 {
		t.Fatalf("省 %d 的 +0/+2 解析錯誤：%d/%d", province,
			got.SideALeader, got.SideBLeader)
	}
	if got.Branch4.GeneralIDs[0] != 0x0102 || got.Branch4.GeneralIDs[9] != 0x0304 ||
		got.Branch8.GeneralIDs[0] != 0x0506 || got.Branch8.GeneralIDs[9] != 0x0708 {
		t.Fatalf("省 %d 的分支將領清單解析錯誤：4=%#v 8=%#v",
			province, got.Branch4.GeneralIDs, got.Branch8.GeneralIDs)
	}
	if got.Branch8.Resources.Fuel != 0x0102 {
		t.Fatalf("省 %d 的 +18 union 未保留原始值：branch4=%+v branch8=%+v",
			province, got.Branch4.Resources, got.Branch8.Resources)
	}
	if got.Branch4.Resources != (WarRecordResources{0x1101, 0x1102, 0x1103, 0x1104}) ||
		got.Branch8.Resources.Gold != 0x2201 || got.Branch8.Resources.Food != 0x2202 ||
		got.Branch8.Resources.Ammunition != 0x2203 {
		t.Fatalf("省 %d 的資源分支解析錯誤：4=%+v 8=%+v",
			province, got.Branch4.Resources, got.Branch8.Resources)
	}
	if got.Raw[59] != 0xA5 {
		t.Fatalf("未解尾端 byte 被改寫：%#x", got.Raw[59])
	}
	if _, err := ParseWarRecords(save[:len(save)-1]); err == nil {
		t.Fatal("過短存檔應該報錯")
	}
}

func TestWriteWarRecordsPreservesUnknownBytes(t *testing.T) {
	orig := readGame(t, "SAVE(1).DT1")
	records, err := ParseWarRecords(orig)
	if err != nil {
		t.Fatal(err)
	}
	const province = ProvinceID(7)
	base, err := SaveBlockByGlobal("byte_6F532")
	if err != nil {
		t.Fatal(err)
	}
	offset := base.Offset + int(province-1)*WarRecordSize
	oldA, oldB := records[province].SideALeader, records[province].SideBLeader
	records[province].SideALeader = oldA ^ 0x55AA
	records[province].SideBLeader = oldB ^ 0x0F0F
	records[province].Branch4.GeneralIDs[0] ^= 0x1111
	records[province].Branch8.GeneralIDs[9] ^= 0x2222
	// 即使呼叫端誤改 Raw，writer 也必須以 orig 為基底，不能把未知尾端寫掉。
	records[province].Raw[WarRecordSize-1] ^= 0xFF
	out, err := WriteWarRecords(orig, records)
	if err != nil {
		t.Fatal(err)
	}
	d := DiffBytes(orig, out, 16)
	want := []int{offset, offset + 1, offset + 2, offset + 3}
	if len(d) != len(want) {
		t.Fatalf("戰爭記錄只應改兩個 u16，未知 Raw 不得變動；實際差分：%v", d)
	}
	for i, off := range want {
		if d[i] != off {
			t.Fatalf("第 %d 個差分 offset=%d，預期 %d；全部=%v", i, d[i], off, d)
		}
	}
	back, err := ParseWarRecords(out)
	if err != nil {
		t.Fatal(err)
	}
	if back[province].SideALeader != records[province].SideALeader ||
		back[province].SideBLeader != records[province].SideBLeader ||
		back[province].Raw[WarRecordSize-1] != orig[offset+WarRecordSize-1] {
		t.Fatalf("戰爭記錄寫回後讀值不一致：%+v", back[province])
	}
}

func TestWriteWarRecordsRoundTripsUntouchedSave(t *testing.T) {
	orig := readGame(t, "SAVE(1).DT1")
	records, err := ParseWarRecords(orig)
	if err != nil {
		t.Fatal(err)
	}
	out, err := WriteWarRecords(orig, records)
	if err != nil {
		t.Fatal(err)
	}
	if d := DiffBytes(orig, out, 8); len(d) != 0 {
		t.Fatalf("未改戰爭記錄卻有差分：%v", d)
	}
}

func TestWriteWarRecordBranchWritesOnlySelectedLayout(t *testing.T) {
	blk, err := SaveBlockByGlobal("byte_6F532")
	if err != nil {
		t.Fatal(err)
	}
	const province = ProvinceID(7)
	orig := make([]byte, blk.Offset+blk.Size)
	base := blk.Offset + int(province-1)*WarRecordSize
	for i := 0; i < WarRecordSize; i++ {
		orig[base+i] = byte(0x80 + i)
	}
	out, err := WriteWarRecordBranch(orig, province, WarRecordBranch4,
		WarRecordResources{0x1101, 0x1102, 0x1103, 0x1104},
		[]GeneralID{0x0102, 0x0304})
	if err != nil {
		t.Fatal(err)
	}
	d := DiffBytes(orig, out, 32)
	want := []int{
		base + WarRecordBranch4GoldOffset, base + WarRecordBranch4GoldOffset + 1,
		base + WarRecordBranch4FoodOffset, base + WarRecordBranch4FoodOffset + 1,
		base + WarRecordBranch4AmmunitionOffset, base + WarRecordBranch4AmmunitionOffset + 1,
		base + WarRecordBranch4FuelOffset, base + WarRecordBranch4FuelOffset + 1,
		base + WarRecordGeneralList18Offset, base + WarRecordGeneralList18Offset + 1,
		base + WarRecordGeneralList18Offset + 2, base + WarRecordGeneralList18Offset + 3,
	}
	if len(d) != len(want) {
		t.Fatalf("Branch4 應只改已選資源與 ID 前綴：%v", d)
	}
	for i, off := range want {
		if d[i] != off {
			t.Fatalf("Branch4 第 %d 個差分 offset=%d，預期 %d；全部=%v", i, d[i], off, d)
		}
	}
	// +0x12 是 union：Branch4 第一個 ID 會覆蓋該處原本的 Branch8 Fuel，
	// 但未提供的第二個 ID 之後仍保留原始 bytes。
	if got := binary.LittleEndian.Uint16(out[base+WarRecordGeneralList18Offset:]); got != 0x0102 {
		t.Fatalf("Branch4 第一個 ID 未寫入 union：%#x", got)
	}
	if out[base+WarRecordGeneralList18Offset+4] != orig[base+WarRecordGeneralList18Offset+4] {
		t.Fatal("未提供的 Branch4 ID 槽位不應被清零")
	}

	out, err = WriteWarRecordBranch(orig, province, WarRecordBranch8,
		WarRecordResources{0x2201, 0x2202, 0x2203, 0x2204},
		[]GeneralID{0x0506})
	if err != nil {
		t.Fatal(err)
	}
	row := out[base : base+WarRecordSize]
	if got := binary.LittleEndian.Uint16(row[WarRecordBranch8FuelOffset:]); got != 0x2204 {
		t.Fatalf("Branch8 Fuel 未寫入 union：%#x", got)
	}
	if got := binary.LittleEndian.Uint16(row[WarRecordGeneralList38Offset:]); got != 0x0506 {
		t.Fatalf("Branch8 第一個 ID 未寫入：%#x", got)
	}
	if row[WarRecordGeneralList38Offset+2] != orig[base+WarRecordGeneralList38Offset+2] {
		t.Fatal("未提供的 Branch8 ID 槽位不應被清零")
	}
}

func TestWriteWarRecordBranchRejectsAmbiguousInput(t *testing.T) {
	blk, err := SaveBlockByGlobal("byte_6F532")
	if err != nil {
		t.Fatal(err)
	}
	orig := make([]byte, blk.Offset+blk.Size)
	for _, tc := range []struct {
		name   string
		prov   ProvinceID
		branch WarRecordBranchKind
		ids    []GeneralID
	}{
		{name: "省份 0", prov: 0, branch: WarRecordBranch4},
		{name: "未知分支", prov: 1, branch: 9},
		{name: "零 ID", prov: 1, branch: WarRecordBranch4, ids: []GeneralID{0}},
		{name: "重複 ID", prov: 1, branch: WarRecordBranch4, ids: []GeneralID{7, 7}},
		{name: "超過槽位", prov: 1, branch: WarRecordBranch4,
			ids: []GeneralID{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if _, err := WriteWarRecordBranch(orig, tc.prov, tc.branch,
				WarRecordResources{}, tc.ids); err == nil {
				t.Fatal("應拒絕不明確或非法的 DT1 分支輸入")
			}
		})
	}
}

// 兩檔門檻的機率：佔上風 70%、劣勢 20%。
func TestCeasefireRates(t *testing.T) {
	cases := []struct {
		name string
		min  int
		want float64
	}{
		{"佔上風（roll ≥ 3）", CeasefireStrongMin, 0.7},
		{"居劣勢（roll ≥ 8）", CeasefireWeakMin, 0.2},
	}
	for _, c := range cases {
		rng := NewRand(2024)
		const n = 20000
		ok := 0
		for i := 0; i < n; i++ {
			if rng.Int(CeasefireRollRange) >= c.min {
				ok++
			}
		}
		rate := float64(ok) / n
		if rate < c.want-0.03 || rate > c.want+0.03 {
			t.Errorf("%s 的同意率 %.3f，預期 %.1f", c.name, rate, c.want)
		}
	}
}

// 佔上風的判定：守方請求時比守方戰力，攻方請求時比攻方戰力。
func TestCeasefireStrongerSide(t *testing.T) {
	w := realWorld(t)
	prov, err := w.Table.At(1)
	if err != nil {
		t.Fatal(err)
	}
	defender := GeneralID(7)
	attacker := GeneralID(8)
	prov.Commander = defender

	// 造兩個在場單位：守方強、攻方弱。
	w.Units = []CombatUnit{
		{General: 1, Province: 1, Cell: 0, Faction: defender, Active: true},
		{General: 2, Province: 1, Cell: 1, Faction: attacker, Active: true},
	}
	w.Strengths = []StrengthInput{
		{Ability: 90, Force: 20000, Branch: BranchInfantry, F19: 60, F20: 60, F29: 80, F30: 80},
		{Ability: 30, Force: 2000, Branch: BranchInfantry, F19: 30, F20: 30, F29: 40, F30: 40},
	}

	atk, def := w.BattleForces(1)
	if def <= atk {
		t.Fatalf("測試前提不成立：守方 %d 沒有比攻方 %d 強", def, atk)
	}

	// 守方（省司令）來談 → 佔上風。
	res, err := w.NegotiateCeasefire(1, defender, NewRand(1))
	if err != nil {
		t.Fatal(err)
	}
	if !res.RequesterStronger {
		t.Error("守方較強時，守方請求應該算佔上風")
	}

	// 攻方來談 → 劣勢。
	res, err = w.NegotiateCeasefire(1, attacker, NewRand(1))
	if err != nil {
		t.Fatal(err)
	}
	if res.RequesterStronger {
		t.Error("守方較強時，攻方請求應該算劣勢")
	}
}

func TestCeasefireAgreementIncrementsRawState(t *testing.T) {
	w := realWorld(t)
	prov, err := w.Table.At(1)
	if err != nil {
		t.Fatal(err)
	}
	requester := prov.Commander
	if requester == 0 {
		t.Fatal("測試存檔的省 1 不應該沒有司令")
	}
	w.CeasefireState[1] = 7
	for seed := uint32(1); seed <= 100; seed++ {
		res, err := w.NegotiateCeasefire(1, requester, NewRand(seed))
		if err != nil {
			t.Fatal(err)
		}
		if !res.Agreed {
			continue
		}
		if w.CeasefireState[1] != 8 {
			t.Fatalf("停火成功後應依原版 inc：實得 %d", w.CeasefireState[1])
		}
		return
	}
	t.Fatal("100 個固定 seed 都未取得一次佔上風停火成功，測試前提失效")
}

func TestCeasefireStateIncrementWrapsLikeByte(t *testing.T) {
	w := realWorld(t)
	prov, err := w.Table.At(1)
	if err != nil {
		t.Fatal(err)
	}
	if prov.Commander == 0 {
		t.Fatal("測試存檔的省 1 不應該沒有司令")
	}
	w.CeasefireState[1] = 0xFF
	for seed := uint32(1); seed <= 100; seed++ {
		res, err := w.NegotiateCeasefire(1, prov.Commander, NewRand(seed))
		if err != nil {
			t.Fatal(err)
		}
		if res.Agreed {
			if w.CeasefireState[1] != 0 {
				t.Fatalf("byte inc 溢位應回 0，實得 %d", w.CeasefireState[1])
			}
			return
		}
	}
	t.Fatal("100 個固定 seed 都未取得一次停火成功，測試前提失效")
}

// 同一顆種子下，佔上風的一方成功率必須不低於劣勢方。
func TestCeasefireStrongerIsEasier(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	defender, attacker := GeneralID(7), GeneralID(8)
	prov.Commander = defender
	w.Units = []CombatUnit{
		{General: 1, Province: 1, Cell: 0, Faction: defender, Active: true},
		{General: 2, Province: 1, Cell: 1, Faction: attacker, Active: true},
	}
	w.Strengths = []StrengthInput{
		{Ability: 90, Force: 20000, Branch: BranchInfantry, F19: 60, F20: 60, F29: 80, F30: 80},
		{Ability: 30, Force: 2000, Branch: BranchInfantry, F19: 30, F20: 30, F29: 40, F30: 40},
	}

	strongOK, weakOK := 0, 0
	const n = 300
	for seed := uint32(1); seed <= n; seed++ {
		if r, _ := w.NegotiateCeasefire(1, defender, NewRand(seed)); r.Agreed {
			strongOK++
		}
		if r, _ := w.NegotiateCeasefire(1, attacker, NewRand(seed)); r.Agreed {
			weakOK++
		}
	}
	if strongOK <= weakOK {
		t.Errorf("佔上風談成 %d 次、劣勢 %d 次——強勢方應該比較容易談成",
			strongOK, weakOK)
	}
}

// 不在場的單位（Cell 無效）不計入戰力。
func TestBattleForcesIgnoresOffField(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	prov.Commander = 7
	w.Units = []CombatUnit{
		{General: 1, Province: 1, Cell: NoCell, Faction: 7, Active: true},
		{General: 2, Province: 1, Cell: NoCell, Faction: 8, Active: true},
	}
	w.Strengths = []StrengthInput{
		{Ability: 90, Force: 20000, Branch: BranchInfantry, F19: 60, F20: 60, F29: 80, F30: 80},
		{Ability: 90, Force: 20000, Branch: BranchInfantry, F19: 60, F20: 60, F29: 80, F30: 80},
	}
	if atk, def := w.BattleForces(1); atk != 0 || def != 0 {
		t.Errorf("不在場的單位被計入了：攻 %d／守 %d", atk, def)
	}
}

// 無主省談不了停火。
func TestCeasefireNeedsCommander(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	prov.Commander = 0
	if _, err := w.NegotiateCeasefire(1, 7, NewRand(1)); err == nil {
		t.Error("無主省應該報錯")
	}
}

func TestCeasefirePlayerGatesMatchOriginalCaller(t *testing.T) {
	if got := CeasefireProvinceLimit(1); got != 36 {
		t.Fatalf("第 1 期停火省份上限=%d，預期 36", got)
	}
	if got := CeasefireProvinceLimit(2); got != 0 {
		t.Fatalf("第 2 期應在輸入前停火，實得上限 %d", got)
	}

	w := realWorld(t)
	current := ProvinceID(1)
	prov, err := w.Table.At(current)
	if err != nil || prov.Commander == 0 {
		t.Fatalf("測試存檔省 1 應有司令：%v", err)
	}
	leader := prov.Commander
	if got, err := w.CeasefireRequester(1, current); err != nil || got != leader {
		t.Fatalf("司令駐在目前省時 gate 失敗：%d／%v", got, err)
	}
	if _, err := w.CeasefireRequester(2, current); err == nil {
		t.Fatal("第 2 期不應進入停火輸入")
	}
	for i := range w.Units {
		if w.Units[i].General == leader {
			w.Units[i].Province = current + 1
			break
		}
	}
	if _, err := w.CeasefireRequester(1, current); err == nil {
		t.Fatal("司令離開目前省時應拒絕")
	}
}

func TestValidateCeasefireTargetRequiresCommanderAndBattle(t *testing.T) {
	w := realWorld(t)
	current := ProvinceID(1)
	cur, _ := w.Table.At(current)
	if cur.Commander == 0 {
		t.Fatal("測試存檔省 1 應有司令")
	}
	target := ProvinceID(2)
	p, err := w.Table.At(target)
	if err != nil || p.Commander == 0 {
		t.Fatal("測試需要一個有司令的目標省")
	}
	if _, err := w.ValidateCeasefireTarget(1, current, target); err == nil {
		t.Fatal("未交戰目標不應通過停火前置檢查")
	}
	before := w.CeasefireState[target]
	p.Flags |= ProvinceFlagInBattle
	requester, err := w.ValidateCeasefireTarget(1, current, target)
	if err != nil || requester != cur.Commander {
		t.Fatalf("交戰且有司令的目標應通過：%d／%v", requester, err)
	}
	if w.CeasefireState[target] != before {
		t.Fatal("前置檢查不應改停火狀態")
	}
	if _, err := w.ValidateCeasefireTarget(1, current, ProvinceID(37)); err == nil {
		t.Fatal("第 1 期的第 37 省不應成為停火輸入")
	}
}

func TestBattleForcesCountsDeployedSaveUnitWithoutCell(t *testing.T) {
	w := realWorld(t)
	prov, _ := w.Table.At(1)
	prov.Commander = 7
	w.Units = []CombatUnit{{General: 1, Province: 1, Cell: NoCell, Deployed: true,
		Faction: 7, Active: true}}
	w.Strengths = []StrengthInput{{Ability: 90, Force: 20000, Branch: BranchInfantry,
		F19: 60, F20: 60, F29: 80, F30: 80}}
	if atk, def := w.BattleForces(1); atk != 0 || def == 0 {
		t.Fatalf("已部署但尚未建 Cell 的存檔單位應計入守方：攻 %d／守 %d", atk, def)
	}
}
