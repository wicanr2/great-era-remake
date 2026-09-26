package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/game"
)

func TestPersistBattleSnapshotWritesBothCopiesNonDestructively(t *testing.T) {
	orig := make([]byte, game.ProvinceCount*game.BattleStateSize)
	// 用未解尾端的非零 sentinel 證明 writer 不是重建零值記錄。
	orig[19*game.BattleStateSize+8] = 0xA5
	states, err := game.ParseBattleStates(orig)
	if err != nil {
		t.Fatal(err)
	}
	memStates, err := game.ParseBattleStates(orig)
	if err != nil {
		t.Fatal(err)
	}
	atk := &game.Combatant{CombatUnit: game.CombatUnit{General: 7, Cell: game.NoCell}}
	def := &game.Combatant{CombatUnit: game.CombatUnit{General: 12, Cell: game.NoCell}}
	dir := t.TempDir()
	a := &app{
		battle: &battleState{
			sim:      &game.BattleSim{At: 20, From: 19, Attacker: []*game.Combatant{atk}, Defender: []*game.Combatant{def}},
			finished: true,
			supAtk:   game.BattleSupply{Gold: 100, Food: 200, Ammo: 300, Fuel: 400},
		},
		battleDT2: orig, battleMemWar: append([]byte(nil), orig...),
		battleDT2States: states, battleMemStates: memStates,
		battleDT2Path:    filepath.Join(dir, "SAVE(1).DT2"),
		battleMemWarPath: filepath.Join(dir, "MEM_WAR.DAT"),
	}
	if err := a.persistBattleSnapshot(); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{a.battleDT2Path, a.battleMemWarPath} {
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		got, err := game.ParseBattleStates(data)
		if err != nil {
			t.Fatal(err)
		}
		state := got[19]
		if state.Header != [4]uint16{100, 200, 300, 400} || state.RosterA[0] != 7 ||
			state.RosterB[0] != 12 || state.UnitsA[0] != 7 || state.UnitsB[0] != 12 ||
			state.Trailing != 19 || data[19*game.BattleStateSize+8] != 0xA5 {
			t.Fatalf("%s 寫回欄位／未知 byte 錯誤：header=%v roster=%v/%v trailing=%d sentinel=%#x",
				path, state.Header, state.RosterA, state.RosterB, state.Trailing, data[19*game.BattleStateSize+8])
		}
	}
}

func TestBattleSourceNameUsesSaveStem(t *testing.T) {
	if got := battleSourceName("/tmp/custom/SAVE(2).DT1"); got != "SAVE(2).DT2" {
		t.Fatalf("custom stem = %q", got)
	}
	if got := battleSourceName("/tmp/custom/save.bin"); got != "SAVE(1).DT2" {
		t.Fatalf("fallback stem = %q", got)
	}
}

func TestLoadBattleRecordPrefersExistingOutput(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "SAVE(1).DT2")
	output := make([]byte, game.ProvinceCount*game.BattleStateSize)
	output[0] = 0xA7
	if err := os.WriteFile(path, output, 0o644); err != nil {
		t.Fatal(err)
	}
	data, _, source, err := loadBattleRecord(func(name string) ([]byte, error) {
		fallback := make([]byte, game.ProvinceCount*game.BattleStateSize)
		fallback[0] = 0x11
		return fallback, nil
	}, path, "SAVE(1).DT2", "")
	if err != nil {
		t.Fatal(err)
	}
	if source != path || data[0] != 0xA7 {
		t.Fatalf("未優先讀既有輸出：source=%q byte=%#x", source, data[0])
	}
}
