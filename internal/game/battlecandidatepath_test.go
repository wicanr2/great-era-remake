package game

import (
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

func TestSelectOriginalBattleCandidateUsesTargetNeighboursAndConfirmedFilters(t *testing.T) {
	s := newBattlePathTestSim()
	target, _ := CellAt(6, 6)
	current, _ := CellAt(2, 6)
	setBattlePathTile(s, target, assets.TilePlain, assets.NoRail)
	setBattlePathTile(s, current, assets.TilePlain, assets.NoRail)
	s.Occ[current] = 1
	s.byID[1] = &Combatant{
		CombatUnit: CombatUnit{General: 1, Cell: current},
		Strength:   StrengthInput{Branch: Branch1, Force: 100},
	}

	neighbours := target.Neighbours()
	if len(neighbours) != 6 {
		t.Fatalf("測試目標格應有 6 個鄰格，得到 %d", len(neighbours))
	}
	setBattlePathTile(s, neighbours[0], assets.TileCity, assets.NoRail) // 防禦值 5，應勝出
	setBattlePathTile(s, neighbours[1], assets.TilePlain, assets.NoRail)
	setBattlePathTile(s, neighbours[2], assets.TileDesert, assets.NoRail)
	setBattlePathTile(s, neighbours[3], assets.TileMountain, assets.NoRail) // 成本 255，排除
	setBattlePathTile(s, neighbours[4], assets.TileCity, assets.NoRail)     // 由預約排除
	setBattlePathTile(s, neighbours[5], assets.TileHill, assets.NoRail)

	reserved := neighbours[4]
	got := s.SelectOriginalBattleCandidate(target, current, BattleCandidateOptions{
		Reserved: func(c CellIndex) bool { return c == reserved },
	})
	if !got.Complete {
		t.Fatalf("基礎候選分支應完整，理由：%s", got.Reason)
	}
	if got.Cell != neighbours[0] {
		t.Fatalf("應從目標格鄰格選城市防禦值 5，得到 %d，目標鄰格 %v", got.Cell, neighbours)
	}
	if got.Cell == current || !Adjacent(got.Cell, target) {
		t.Fatalf("候選應是目標格鄰格，不應誤回目前格／非鄰格：%d", got.Cell)
	}
}

func TestSelectOriginalBattleCandidateFailsClosedForUnresolvedTail(t *testing.T) {
	s := newBattlePathTestSim()
	target, _ := CellAt(6, 6)
	current, _ := CellAt(2, 6)
	setBattlePathTile(s, target, assets.TilePlain, assets.NoRail)
	setBattlePathTile(s, current, assets.TilePlain, assets.NoRail)
	s.Occ[current] = 1
	s.byID[1] = &Combatant{
		CombatUnit: CombatUnit{General: 1, Cell: current},
		Strength:   StrengthInput{Branch: Branch1, Force: 100},
	}
	for _, c := range target.Neighbours() {
		setBattlePathTile(s, c, assets.TilePlain, assets.NoRail)
	}

	got := s.SelectOriginalBattleCandidate(target, current, BattleCandidateOptions{
		Mode:            1,
		EnableLastSteps: true,
	})
	if got.Complete || got.Cell != NoCell {
		t.Fatalf("未解尾端必須 fail-closed，得到 %+v", got)
	}
	if got.Reason == "" {
		t.Fatal("fail-closed 結果應帶理由")
	}
}

func TestSelectOriginalBattleCandidateKeepsCurrentOnEqualAdjacentDefence(t *testing.T) {
	s := newBattlePathTestSim()
	target, _ := CellAt(6, 6)
	current, ok := target.Neighbour(DirUp)
	if !ok {
		t.Fatal("測試目標格應有上方鄰格")
	}
	setBattlePathTile(s, target, assets.TilePlain, assets.NoRail)
	setBattlePathTile(s, current, assets.TilePlain, assets.NoRail)
	s.Occ[current] = 1
	s.byID[1] = &Combatant{
		CombatUnit: CombatUnit{General: 1, Cell: current},
		Strength:   StrengthInput{Branch: Branch1, Force: 100},
	}
	for _, c := range target.Neighbours() {
		setBattlePathTile(s, c, assets.TilePlain, assets.NoRail)
	}

	got := s.SelectOriginalBattleCandidate(target, current, BattleCandidateOptions{})
	if !got.Complete || got.Cell != current {
		t.Fatalf("同防禦值且目前格鄰接目標時應保留目前格，得到 %+v", got)
	}
}

func TestSelectOriginalBattleCandidateTreatsCurrentAsTemporarilyEmpty(t *testing.T) {
	s := newBattlePathTestSim()
	target, _ := CellAt(6, 6)
	current, ok := target.Neighbour(DirUp)
	if !ok {
		t.Fatal("測試目標格應有上方鄰格")
	}
	setBattlePathTile(s, target, assets.TilePlain, assets.NoRail)
	setBattlePathTile(s, current, assets.TileCity, assets.NoRail)
	s.Occ[current] = 1
	s.byID[1] = &Combatant{
		CombatUnit: CombatUnit{General: 1, Cell: current},
		Strength:   StrengthInput{Branch: Branch1, Force: 100},
	}
	for _, c := range target.Neighbours() {
		if c != current {
			setBattlePathTile(s, c, assets.TilePlain, assets.NoRail)
		}
	}

	got := s.SelectOriginalBattleCandidate(target, current, BattleCandidateOptions{})
	if !got.Complete || got.Cell != current {
		t.Fatalf("清空目前格佔用後，高防禦目前格應可成為候選，得到 %+v", got)
	}
}

func TestOriginalBattleCandidateCostLimit13ForFourUnrailedWaterNeighbours(t *testing.T) {
	s := newBattlePathTestSim()
	current, _ := CellAt(6, 6)
	setBattlePathTile(s, current, assets.TilePlain, assets.NoRail)
	water := 0
	for _, c := range current.Neighbours() {
		if water < 4 {
			setBattlePathTile(s, c, assets.TileWater, assets.NoRail)
			water++
		} else {
			setBattlePathTile(s, c, assets.TilePlain, assets.NoRail)
		}
	}
	if got := originalBattleCandidateCostLimit(s, current, Branch1); got != 13 {
		t.Fatalf("兵種 1 周圍四個無鐵路水域時成本上限 = %d，應為 13", got)
	}
	if got := originalBattleCandidateCostLimit(s, current, Branch4); got != 12 {
		t.Fatalf("兵種 4 不應套用 13 上限，得到 %d", got)
	}
}
