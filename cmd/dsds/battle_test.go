package main

import (
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
)

func TestCombatantsPreserveMANSlotAndFaction(t *testing.T) {
	gs := []game.General{
		{Province: 2, Force: 111}, // 槽位 1，不屬於本場
		{Province: 7, Force: 222}, // 槽位 2，第一個入場單位
		{Province: 3, Force: 333}, // 槽位 3，不屬於本場
		{Province: 7, Force: 444}, // 槽位 4，第二個入場單位
	}

	got := combatants(gs, 7, 58)
	if len(got) != 2 {
		t.Fatalf("省 7 應有 2 個戰鬥單位，實際 %d", len(got))
	}
	if got[0].General != 2 || got[1].General != 4 {
		t.Errorf("應保留 MAN 槽位 ID [2 4]，實際 [%d %d]",
			got[0].General, got[1].General)
	}
	for i, u := range got {
		if u.Faction != 58 || u.Strength.Faction != 58 {
			t.Errorf("第 %d 個單位的 +14 效忠勢力應為 58，實際 CombatUnit=%d Strength=%d",
				i, u.Faction, u.Strength.Faction)
		}
	}
}

func TestBattleMenuCommandsEnterTheConfirmedModes(t *testing.T) {
	want := map[int]battleMode{
		1: battleModeMove,
		2: battleModeAttack,
		3: battleModeRetreat,
		4: battleModeGarrison,
		5: battleModeInspect,
	}
	for n, expected := range want {
		got, ok := battleModeForCommand(n)
		if !ok || got != expected {
			t.Errorf("戰鬥主選單 %d 應進入 mode %d，得到 mode=%d ok=%v",
				n, expected, got, ok)
		}
	}
	for _, n := range []int{0, 6, 99, -1} {
		if _, ok := battleModeForCommand(n); ok {
			t.Errorf("未定義的戰鬥主選單鍵 %d 不應被接受", n)
		}
	}
}

func TestConfirmedRetreatTargetsDoNotGeneralizeUnknownBattles(t *testing.T) {
	want := []game.ProvinceID{19, 26, 14, 17}
	got := confirmedRetreatTargets(18, 19)
	if len(got) != len(want) {
		t.Fatalf("陝西←河南撤退候選=%v，預期=%v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("撤退候選順序[%d]=%d，預期=%d", i, got[i], want[i])
		}
	}
	for _, tc := range [][2]game.ProvinceID{{18, 16}, {25, 19}, {1, 2}} {
		if got := confirmedRetreatTargets(tc[0], tc[1]); len(got) != 0 {
			t.Fatalf("未閉合交戰 %d←%d 不應補出候選：%v", tc[0], tc[1], got)
		}
	}
}

func TestProvinceInputIsBoundedAndRetreatDestinationIsFailClosed(t *testing.T) {
	var input uint32
	for _, digit := range []int{1, 9} {
		var ok bool
		input, ok = appendProvinceInput(input, digit)
		if !ok {
			t.Fatalf("追加省份數字 %d 失敗", digit)
		}
	}
	if input != 19 {
		t.Fatalf("省份輸入=%d，預期 19", input)
	}
	if next, ok := appendProvinceInput(3, 9); !ok || next != 39 {
		t.Fatalf("39 應是最後合法省份輸入：next=%d ok=%v", next, ok)
	}
	if next, ok := appendProvinceInput(4, 0); ok || next != 4 {
		t.Fatalf("40 應失敗即保持原值：next=%d ok=%v", next, ok)
	}
	if next, ok := appendProvinceInput(39, 9); ok || next != 39 {
		t.Fatalf("超界追加應失敗即保持原值：next=%d ok=%v", next, ok)
	}

	targets := []game.ProvinceID{19, 26, 14, 17}
	if got, ok := retreatDestination(19, targets); !ok || got != 19 {
		t.Fatalf("合法撤退目標解析錯誤：got=%d ok=%v", got, ok)
	}
	for _, input := range []uint32{0, 18, 39} {
		if got, ok := retreatDestination(input, targets); ok || got != game.ProvinceID(input) {
			t.Fatalf("非法撤退目標未失敗即關閉：input=%d got=%d ok=%v", input, got, ok)
		}
	}
}

func TestBattleFinishHelpersUseTheSharedSettlementClassification(t *testing.T) {
	b := &battleState{turn: game.BattleTurnLimit}
	b.finishWithWinner(game.BattleSideFirst)
	if !b.finished || b.settlement.Kind != game.BattleSettlementDecisive ||
		b.settlement.Winner != game.BattleSideFirst || b.settlement.RequiresDT2Writeback {
		t.Fatalf("分出勝負沒有留下正確分類：finished=%v settlement=%+v", b.finished, b.settlement)
	}

	b = &battleState{turn: game.BattleTurnLimit}
	b.finishByTurnLimit()
	if !b.finished || b.settlement.Kind != game.BattleSettlementTurnLimitDraw ||
		b.settlement.Winner != game.BattleSideNone || !b.settlement.RequiresDT2Writeback {
		t.Fatalf("回合上限沒有留下平局／寫回旗標：finished=%v settlement=%+v", b.finished, b.settlement)
	}

	b = &battleState{turn: game.BattleTurnLimitAlt, turnCap: game.BattleTurnLimitAlt}
	b.finishByTurnLimit()
	if !b.finished || b.settlement.TurnCap != game.BattleTurnLimitAlt ||
		b.settlement.Kind != game.BattleSettlementTurnLimitDraw {
		t.Fatalf("二月回合上限沒有使用 15：finished=%v settlement=%+v", b.finished, b.settlement)
	}
}

func TestBattleSettlementProjectsWorldAndIsIdempotent(t *testing.T) {
	table := &game.ProvinceTable{}
	table.Province[17] = game.Province{Commander: 2, Flags: game.ProvinceFlagInBattle}
	table.Province[18] = game.Province{Commander: 1}
	attacker := &game.Combatant{
		CombatUnit: game.CombatUnit{General: 1, Province: 19, Cell: 100},
		Strength:   game.StrengthInput{Force: 1000},
	}
	defender := &game.Combatant{
		CombatUnit: game.CombatUnit{General: 2, Province: 18, Cell: 101},
		Strength:   game.StrengthInput{Force: 1000},
	}
	a := &app{
		tbl:      table,
		generals: []game.General{{Province: 19, Force: 1000}, {Province: 18, Force: 1000}},
		world:    &game.AIWorld{Units: []game.CombatUnit{{General: 1, Province: 19}, {General: 2, Province: 18}}},
		current:  19,
		battle: &battleState{
			sim:        &game.BattleSim{From: 19, At: 18, Attacker: []*game.Combatant{attacker}, Defender: []*game.Combatant{defender}},
			settlement: game.BattleSettlement{Kind: game.BattleSettlementDecisive, Winner: game.BattleSideFirst},
		},
	}
	if err := a.applyBattleSettlement(); err != nil {
		t.Fatal(err)
	}
	target, _ := table.At(18)
	if target.Commander != 1 || target.InBattle() || a.current != 18 || a.generals[0].Province != 18 ||
		a.world.Units[0].Province != 18 || !a.battle.settled {
		t.Fatalf("攻方勝世界投影錯誤：target=%+v current=%d general=%+v unit=%+v settled=%v",
			target, a.current, a.generals[0], a.world.Units[0], a.battle.settled)
	}
	// 第二次呼叫不應再用已改寫的世界狀態重跑任何副作用。
	if err := a.applyBattleSettlement(); err != nil {
		t.Fatal(err)
	}
	if target.Commander != 1 || a.generals[0].Province != 18 {
		t.Fatal("重複套用結算改變了已完成的世界狀態")
	}
}

func TestSyncBattleForcesUpdatesSaveAndAIWorldMirrors(t *testing.T) {
	attacker := &game.Combatant{
		CombatUnit: game.CombatUnit{General: 1}, Strength: game.StrengthInput{Force: 1234},
	}
	defender := &game.Combatant{
		CombatUnit: game.CombatUnit{General: 2}, Strength: game.StrengthInput{Force: 5678},
	}
	a := &app{
		generals: []game.General{{Force: 10}, {Force: 20}},
		world:    &game.AIWorld{Strengths: []game.StrengthInput{{Force: 30}, {Force: 40}}},
		battle: &battleState{
			sim:        &game.BattleSim{Attacker: []*game.Combatant{attacker}, Defender: []*game.Combatant{defender}},
			forceDirty: true,
		},
	}
	if err := a.syncBattleForces(); err != nil {
		t.Fatal(err)
	}
	if a.generals[0].Force != 1234 || a.generals[1].Force != 5678 ||
		a.world.Strengths[0].Force != 1234 || a.world.Strengths[1].Force != 5678 || a.battle.forceDirty {
		t.Fatalf("兵力鏡像未同步：generals=%+v strengths=%+v dirty=%v",
			a.generals, a.world.Strengths, a.battle.forceDirty)
	}
}

func TestBattleImmediateRetreatStoresNoWritebackSettlementBeforeLeaving(t *testing.T) {
	b := &battleState{}
	b.settlement = game.BattleSettlementImmediateRetreat()
	if b.settlement.Kind != game.BattleSettlementRetreat || b.settlement.RequiresDT2Writeback {
		t.Fatalf("立即撤退分類錯誤：%+v", b.settlement)
	}
}

func TestBattleAttackTargetsAddsConfirmedRemoteCandidatesForBranchFour(t *testing.T) {
	field := &game.Battlefield{}
	for y := range field.Tiles {
		for x := range field.Tiles[y] {
			field.Tiles[y][x] = assets.Tile{Kind: assets.TilePlain, Rail: assets.NoRail}
		}
	}
	source := game.CellIndex(100)
	var remote game.CellIndex
	for _, cell := range game.RangedTargetCells(field, source, 1) {
		if !game.Adjacent(source, cell) {
			remote = cell
			break
		}
	}
	if !remote.Valid() {
		t.Fatal("測試需要一個已閉合的非相鄰遠程候選")
	}
	attacker := &game.Combatant{
		CombatUnit: game.CombatUnit{General: 1, Cell: source, Active: true, Attacking: true, Max: 12, Current: 12, Facing: 1},
		Strength:   game.StrengthInput{General: 1, Faction: 1, Branch: game.BranchArtiller, Force: 20000},
	}
	target := &game.Combatant{
		CombatUnit: game.CombatUnit{General: 2, Cell: remote, Active: true, Max: 12, Current: 12},
		Strength:   game.StrengthInput{General: 2, Faction: 2, Branch: game.Branch1, Force: 20000},
	}
	b := &battleState{sim: &game.BattleSim{Field: field, Attacker: []*game.Combatant{attacker}, Defender: []*game.Combatant{target}}}
	got := b.battleAttackTargets(attacker)
	if len(got) != 1 || got[0] != target {
		t.Fatalf("兵種 4 應公告非相鄰遠程目標：%v", got)
	}
}
