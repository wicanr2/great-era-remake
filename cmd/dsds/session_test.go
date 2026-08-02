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
