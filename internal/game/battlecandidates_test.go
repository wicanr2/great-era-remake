package game

import (
	"reflect"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/assets"
)

func TestIsUnrailedWaterTileKeepsRailBridgeDistinct(t *testing.T) {
	if !IsUnrailedWaterTile(assets.Tile{Kind: assets.TileWater, Rail: assets.NoRail}) {
		t.Fatal("沒有鐵路的河海應被 sub_58FD9 判為真")
	}
	if IsUnrailedWaterTile(assets.Tile{Kind: assets.TileWater, Rail: 0}) {
		t.Fatal("鐵橋不應被 sub_58FD9 判為無鐵路河海")
	}
	if IsUnrailedWaterTile(assets.Tile{Kind: assets.TilePlain, Rail: assets.NoRail}) {
		t.Fatal("平原不應被判為河海")
	}
}

func TestSortBattleCandidatesByDefenseDescending(t *testing.T) {
	candidates := []CellIndex{1, 2, 3, NoCell}
	score := map[CellIndex]int{1: 1, 2: 5, 3: 4}
	SortBattleCandidatesByDefense(candidates, func(c CellIndex) int { return score[c] })
	want := []CellIndex{2, 3, 1, NoCell}
	if !reflect.DeepEqual(candidates, want) {
		t.Fatalf("防禦排序 = %v，want %v", candidates, want)
	}
}

func TestSortBattleCandidatesByTargetDistanceKeepsUnrailedWaterFromSwapping(t *testing.T) {
	// 以 0 為目標，三格距離依序為 10、7、4。
	candidates := []CellIndex{10, 20, 30}
	water := map[CellIndex]bool{20: true}
	SortBattleCandidatesByTargetDistance(candidates, 0,
		func(c CellIndex) bool { return water[c] })
	// 20 雖較近但被 sub_58FD9 擋住；30 可前移，且原版的巢狀交換
	// 會留下 [30, 20, 10]。
	want := []CellIndex{30, 20, 10}
	if !reflect.DeepEqual(candidates, want) {
		t.Fatalf("目標距離排序 = %v，want %v", candidates, want)
	}

	candidates = []CellIndex{10, 20}
	SortBattleCandidatesByTargetDistance(candidates, 0,
		func(CellIndex) bool { return false })
	if want = []CellIndex{20, 10}; !reflect.DeepEqual(candidates, want) {
		t.Fatalf("一般河海／非河海候選應可交換，得到 %v，want %v", candidates, want)
	}
}

func TestSelectBattleCandidateMatchesSub566B6Branches(t *testing.T) {
	score := map[CellIndex]int{10: 5, 20: 3, 30: 5}
	lookup := func(c CellIndex) int { return score[c] }
	if got := SelectBattleCandidate(10, 20, true, lookup); got != 20 {
		t.Fatalf("特殊旗標應選 alternate，得到 %d", got)
	}
	if got := SelectBattleCandidate(10, 20, false, lookup); got != 10 {
		t.Fatalf("防禦值不同應選 primary，得到 %d", got)
	}
	if got := SelectBattleCandidate(10, 30, false, lookup); got != 30 {
		t.Fatalf("防禦值相等應選 alternate，得到 %d", got)
	}
	if got := SelectBattleCandidate(10, 10, true, lookup); got != 10 {
		t.Fatalf("兩格相同應直接回原格，得到 %d", got)
	}
}
