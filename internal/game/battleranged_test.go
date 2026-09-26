package game

import (
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

func rangedTestField() *Battlefield {
	f := &Battlefield{}
	for y := range f.Tiles {
		for x := range f.Tiles[y] {
			f.Tiles[y][x] = assets.Tile{Kind: assets.TilePlain, Rail: assets.NoRail}
		}
	}
	return f
}

func rangedTestUnit(id GeneralID, branch uint8, cell CellIndex, attacking bool) *Combatant {
	return &Combatant{
		CombatUnit: CombatUnit{
			General: id, Cell: cell, Attacking: attacking, Facing: 1,
			Max: 12, Current: 12, Active: true, Decaying: 80,
		},
		Strength: StrengthInput{
			Ability: 90, Force: 20000, F19: 60, F20: 60, F29: 64, F30: 80,
			Branch: branch, General: id, Faction: id,
		},
	}
}

func TestRangedTargetCellsHonorsForestAndMountainGates(t *testing.T) {
	f := rangedTestField()
	source := CellIndex(100)
	if got := RangedTargetCells(f, source, 1); len(got) == 0 {
		t.Fatal("平原上的有效射擊方向不應沒有候選")
	}
	col, row := source.ColRow()
	f.Tiles[row][col].Kind = assets.TileForest
	if got := RangedTargetCells(f, source, 1); len(got) != 0 {
		t.Fatalf("森林來源不應產生遠程候選：%v", got)
	}

	f = rangedTestField()
	cells := RangedTargetCells(f, source, 1)
	if len(cells) == 0 {
		t.Fatal("重建測試候選失敗")
	}
	for _, cell := range cells {
		cc, rr := cell.ColRow()
		f.Tiles[rr][cc].Kind = assets.TileMountain
	}
	if got := RangedTargetCells(f, source, 1); len(got) != 0 {
		t.Fatalf("候選高山全部排除後應為空：%v", got)
	}
}

func TestEngageRangedAppliesThreeTargetPassesWithoutRetaliation(t *testing.T) {
	f := rangedTestField()
	source := CellIndex(100)
	cells := RangedTargetCells(f, source, 1)
	if len(cells) == 0 {
		t.Fatal("找不到遠程候選")
	}
	var targetCell CellIndex
	for _, cell := range cells {
		if !Adjacent(source, cell) {
			targetCell = cell
			break
		}
	}
	if !targetCell.Valid() {
		t.Fatal("測試需要非相鄰遠程候選")
	}
	attacker := rangedTestUnit(1, BranchArtiller, source, true)
	target := rangedTestUnit(2, Branch1, targetCell, false)
	s := &BattleSim{
		Field: f, Attacker: []*Combatant{attacker}, Defender: []*Combatant{target},
		Opts: StrengthOpts{Stage: 1}, byID: map[GeneralID]*Combatant{1: attacker, 2: target},
	}
	tc, tr := targetCell.ColRow()
	sc, sr := source.ColRow()
	attackerTile := f.Tiles[sr][sc]
	targetTile := f.Tiles[tr][tc]
	atkOnTarget := AttackValue(s.StrengthOf(attacker), attackerTile, attacker.Branch(), targetTile, target.Branch())
	atkOnAttacker := AttackValue(s.StrengthOf(target), targetTile, target.Branch(), attackerTile, attacker.Branch())
	_, onePass := rangedLossOnce(attacker, target, atkOnTarget, atkOnAttacker)
	la, lt, err := s.EngageRanged(attacker, target)
	if err != nil {
		t.Fatal(err)
	}
	if la != 0 {
		t.Fatalf("遠程攻擊者不應吃反擊，損失=%d", la)
	}
	if want := clampLoss(onePass*3, 20000); lt != want {
		t.Fatalf("遠程三次戰損=%d，預期=%d（單次=%d）", lt, want, onePass)
	}
	if attacker.Force() != 20000 || target.Force() != uint16(20000-lt) {
		t.Fatalf("遠程寫回兵力錯誤：攻=%d 守=%d 損失=%d", attacker.Force(), target.Force(), lt)
	}
}

func TestRangedTargetsOnlyReturnsOppositeLiveUnits(t *testing.T) {
	f := rangedTestField()
	source := CellIndex(100)
	cells := RangedTargetCells(f, source, 1)
	if len(cells) == 0 {
		t.Fatal("找不到遠程候選")
	}
	target := rangedTestUnit(2, Branch1, cells[0], false)
	friendly := rangedTestUnit(3, Branch1, cells[len(cells)-1], true)
	attacker := rangedTestUnit(1, BranchArtiller, source, true)
	s := &BattleSim{Field: f, Attacker: []*Combatant{attacker, friendly}, Defender: []*Combatant{target}}
	got := s.RangedTargets(attacker)
	if len(got) != 1 || got[0] != target {
		t.Fatalf("遠程目標應只公告對方存活單位：%v", got)
	}
	friendly.Strength.Force = 0
	if got := s.RangedTargets(attacker); len(got) != 1 || got[0] != target {
		t.Fatalf("友軍死亡後仍應只公告目標：%v", got)
	}
}
