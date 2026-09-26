package game

import (
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

func repeatedAttackField() *Battlefield {
	f := &Battlefield{}
	for y := range f.Tiles {
		for x := range f.Tiles[y] {
			f.Tiles[y][x] = assets.Tile{Kind: assets.TilePlain, Rail: assets.NoRail}
		}
	}
	return f
}

func repeatedAttackUnit(id GeneralID, cell CellIndex, attacking bool) *Combatant {
	return &Combatant{
		CombatUnit: CombatUnit{General: id, Cell: cell, Attacking: attacking, Active: true},
		Strength: StrengthInput{
			Ability: 90, Force: 20000, F19: 60, F20: 60, F29: 64, F30: 80,
			Branch: BranchInfantry, General: id, Faction: id,
		},
	}
}

func TestEngageRepeatedUsesFivePassesAndKeepsLossBounded(t *testing.T) {
	field := repeatedAttackField()
	attacker := repeatedAttackUnit(1, 100, true)
	target := repeatedAttackUnit(2, 101, false)
	sim := &BattleSim{
		Field: field, Attacker: []*Combatant{attacker}, Defender: []*Combatant{target},
		Opts: StrengthOpts{Stage: 1}, byID: map[GeneralID]*Combatant{1: attacker, 2: target},
	}
	initialA, initialT := attacker.Force(), target.Force()
	la, lt, err := sim.EngageRepeated(attacker, target)
	if err != nil {
		t.Fatal(err)
	}
	if RepeatedAttackPasses != 5 {
		t.Fatalf("重複正規攻擊 pass=%d，證據契約應為 5", RepeatedAttackPasses)
	}
	if la < 0 || lt < 0 || la > int(initialA) || lt > int(initialT) {
		t.Fatalf("累計損失超出兵力界線：攻=%d 守=%d 初始=%d/%d", la, lt, initialA, initialT)
	}
	if attacker.Force() != initialA-uint16(la) || target.Force() != initialT-uint16(lt) {
		t.Fatalf("五次戰損未回寫到單位：攻=%d 守=%d 損失=%d/%d", attacker.Force(), target.Force(), la, lt)
	}
	if la == 0 && lt == 0 {
		t.Fatal("有效相鄰的五次正規交戰不應完全沒有戰損")
	}
}

func TestEngageRepeatedFailsClosedForNonAdjacentUnits(t *testing.T) {
	field := repeatedAttackField()
	attacker := repeatedAttackUnit(1, 0, true)
	target := repeatedAttackUnit(2, 195, false)
	sim := &BattleSim{Field: field, Attacker: []*Combatant{attacker}, Defender: []*Combatant{target}, Opts: StrengthOpts{Stage: 1}}
	if _, _, err := sim.EngageRepeated(attacker, target); err == nil {
		t.Fatal("不相鄰的重複交戰應失敗即關閉")
	}
}
