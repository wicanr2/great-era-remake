package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/game"
)

func readStage1Save(t *testing.T) []byte {
	t.Helper()
	b, err := os.ReadFile("../../workplace/orig/game/SAVE(1).DT1")
	if err != nil {
		t.Fatalf("讀測試存檔: %v", err)
	}
	return b
}

func stage1Player(t *testing.T, save []byte) (game.ProvinceID, game.GeneralID) {
	t.Helper()
	tbl, err := game.ParseSaveProvinces(save)
	if err != nil {
		t.Fatal(err)
	}
	for id := game.ProvinceID(1); id <= 36; id++ {
		p, err := tbl.At(id)
		if err == nil && p.Commander != 0 {
			return id, p.Commander
		}
	}
	t.Fatal("測試存檔沒有可用司令")
	return 0, 0
}

func TestBuildSessionParsesCompleteSave(t *testing.T) {
	save := readStage1Save(t)
	current, player := stage1Player(t, save)
	sc, _ := game.ScenarioByStage(1)
	s, err := buildSession(save, sc, current, player)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.generals) != sc.Generals || s.world == nil || s.cmdBudget == nil {
		t.Fatalf("快照不完整：generals=%d world=%v budget=%v",
			len(s.generals), s.world != nil, s.cmdBudget != nil)
	}
	if s.current != current || s.playerCommander != player {
		t.Fatalf("玩家定位變了：current=%d player=%d", s.current, s.playerCommander)
	}
	if s.year == 0 || s.month == 0 {
		t.Fatalf("存檔日期沒有進入快照：%d/%d", s.year, s.month)
	}
	// session 必須擁有自己的原始 bytes，不能被呼叫者之後改寫。
	want := s.origSave[0]
	save[0] ^= 0xff
	if s.origSave[0] != want {
		t.Fatal("session 沒有複製輸入存檔")
	}
}

func TestBuildSessionRelocatesToPlayerProvince(t *testing.T) {
	save := readStage1Save(t)
	own, player := stage1Player(t, save)
	tbl, _ := game.ParseSaveProvinces(save)
	wrong := game.ProvinceID(0)
	for id := game.ProvinceID(1); id <= 36; id++ {
		p, _ := tbl.At(id)
		if p.Commander != player {
			wrong = id
			break
		}
	}
	if wrong == 0 {
		t.Fatal("測試存檔沒有其他勢力省份")
	}
	sc, _ := game.ScenarioByStage(1)
	s, err := buildSession(save, sc, wrong, player)
	if err != nil {
		t.Fatal(err)
	}
	p, _ := s.tbl.At(s.current)
	if p.Commander != player || s.current == wrong {
		t.Fatalf("沒有重定位到玩家省：own=%d got=%d commander=%d", own, s.current, p.Commander)
	}
}

func TestLoadSessionFailureLeavesOldStateUntouched(t *testing.T) {
	save := readStage1Save(t)
	current, player := stage1Player(t, save)
	oldTable := &game.ProvinceTable{}
	oldWorld := &game.AIWorld{Table: oldTable}
	oldBudget := game.NewCommandBudget(oldWorld)
	oldRaw := []byte{1, 2, 3}
	a := &app{
		tbl: oldTable, world: oldWorld, cmdBudget: oldBudget, origSave: oldRaw,
		generals: []game.General{{}}, stage: 1, current: current,
		playerCommander: player, year: 99, month: 9,
	}
	if err := a.loadSessionBytes(save[:100]); err == nil {
		t.Fatal("過短存檔應該被拒絕")
	}
	if a.tbl != oldTable || a.world != oldWorld || a.cmdBudget != oldBudget ||
		&a.origSave[0] != &oldRaw[0] || a.year != 99 || a.month != 9 || len(a.generals) != 1 {
		t.Fatal("載入失敗後舊狀態被部分替換")
	}
}

func TestBuildSessionRejectsCrossBlockFactionMismatch(t *testing.T) {
	save := append([]byte(nil), readStage1Save(t)...)
	current, player := stage1Player(t, save)
	blk, err := game.SaveBlockByGlobal("byte_6EFAA")
	if err != nil {
		t.Fatal(err)
	}
	// 只改勢力表第一槽的領袖，領袖表與反查表保持原樣。
	save[blk.Offset] ^= 1
	sc, _ := game.ScenarioByStage(1)
	if _, err := buildSession(save, sc, current, player); err == nil {
		t.Fatal("三份勢力索引不一致時應該拒絕載入")
	}
}

func TestAutosaveUsesAtomicOutputAndRefreshesBase(t *testing.T) {
	save := readStage1Save(t)
	current, player := stage1Player(t, save)
	sc, _ := game.ScenarioByStage(1)
	s, err := buildSession(save, sc, current, player)
	if err != nil {
		t.Fatal(err)
	}
	a := &app{savePath: filepath.Join(t.TempDir(), "SAVE(1).DT1")}
	a.applySession(s)
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	if diff := game.DiffBytes(save, got, 1); len(diff) != 0 {
		t.Fatalf("無操作儲存改了 byte %d", diff[0])
	}
	if len(a.origSave) != len(got) || &a.origSave[0] == &save[0] {
		t.Fatal("儲存後沒有用獨立輸出更新 session 基底")
	}
	if matches, _ := filepath.Glob(filepath.Join(filepath.Dir(a.savePath), ".dsds-save-*")); len(matches) != 0 {
		t.Fatalf("原子儲存留下暫存檔：%v", matches)
	}
}

func TestAutosavePersistsCeasefireStateWithoutTouchingOtherBytes(t *testing.T) {
	save := readStage1Save(t)
	current, player := stage1Player(t, save)
	sc, _ := game.ScenarioByStage(1)
	s, err := buildSession(save, sc, current, player)
	if err != nil {
		t.Fatal(err)
	}
	state, err := game.ParseCeasefireStates(save)
	if err != nil {
		t.Fatal(err)
	}
	province := game.ProvinceID(1)
	state[province]++
	s.world.CeasefireState = state

	a := &app{savePath: filepath.Join(t.TempDir(), "SAVE(1).DT1")}
	a.applySession(s)
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	blk, err := game.SaveBlockByGlobal("byte_6FE56")
	if err != nil {
		t.Fatal(err)
	}
	wantOffset := blk.Offset + int(province) - 1
	diff := game.DiffBytes(save, got, 0)
	if len(diff) != 1 || diff[0] != wantOffset {
		t.Fatalf("停火寫回應只改 offset %d，實際差分=%v", wantOffset, diff)
	}
	back, err := game.ParseCeasefireStates(got)
	if err != nil {
		t.Fatal(err)
	}
	if back[province] != state[province] {
		t.Fatalf("停火狀態未持久化：want=%d got=%d", state[province], back[province])
	}
}

func TestAutosavePersistsDiplomacyLedgerWithoutTouchingOtherBytes(t *testing.T) {
	save := readStage1Save(t)
	current, player := stage1Player(t, save)
	sc, _ := game.ScenarioByStage(1)
	s, err := buildSession(save, sc, current, player)
	if err != nil {
		t.Fatal(err)
	}
	s.ledger.Debt[1] = 0xAABBCCDD
	s.ledger.Credit[1] = 77

	a := &app{savePath: filepath.Join(t.TempDir(), "SAVE(1).DT1")}
	a.applySession(s)
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	back, err := game.ParseDiplomacyLedger(got)
	if err != nil {
		t.Fatal(err)
	}
	if back.Debt[1] != 0xAABBCCDD || back.Credit[1] != 77 {
		t.Fatalf("外交帳本未寫回：debt=%#x credit=%d", back.Debt[1], back.Credit[1])
	}
	diff := game.DiffBytes(save, got, 0)
	for _, off := range diff {
		if !((off >= 14620 && off < 14624) || off == 14610) {
			t.Fatalf("外交帳本寫回改到未授權 offset %d（全部差分=%v）", off, diff)
		}
	}
	if len(diff) != 5 {
		t.Fatalf("外債 4 bytes + 信用度 1 byte 應有 5 個差分，實際=%v", diff)
	}
}

func TestAutosaveSyncsBattleForceWithoutTouchingUnknownBytes(t *testing.T) {
	save := readStage1Save(t)
	current, player := stage1Player(t, save)
	sc, _ := game.ScenarioByStage(1)
	s, err := buildSession(save, sc, current, player)
	if err != nil {
		t.Fatal(err)
	}
	// 翻動高低 byte，讓差分護欄確實驗到 little-endian u16 的兩個已解位置。
	wantForce := s.generals[0].Force ^ 0x0101
	sim := &game.BattleSim{
		Attacker: []*game.Combatant{{
			CombatUnit: game.CombatUnit{General: 1},
			Strength:   game.StrengthInput{Force: wantForce},
		}},
	}
	a := &app{savePath: filepath.Join(t.TempDir(), "SAVE(1).DT1")}
	a.applySession(s)
	a.battle = &battleState{sim: sim, forceDirty: true}
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	back, err := game.ParseSaveGenerals(got, sc.Generals)
	if err != nil {
		t.Fatal(err)
	}
	if back[0].Force != wantForce {
		t.Fatalf("戰鬥兵力沒有透過 autosave 寫回：want=%d got=%d", wantForce, back[0].Force)
	}
	// Force 是已解欄位；其餘將領區與整份存檔都不得因同步被重建。
	diff := game.DiffBytes(save, got, 0)
	if len(diff) != 2 {
		t.Fatalf("只改 Force 應有 2 個 byte 差分，實際=%v", diff)
	}
	for _, off := range diff {
		if off < game.SaveGeneralsOffset || off >= game.SaveGeneralsOffset+game.GeneralRecordSize {
			t.Fatalf("戰損同步改到將領區外 offset %d", off)
		}
	}
}
