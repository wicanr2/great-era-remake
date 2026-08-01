// battlesim 用原版資料跑一場戰鬥，把每一回合印出來。
//
// 這是規則層的實跑驗收工具——`CLAUDE.md` §10 第 4 條說「測試綠不是完成」，
// 這支讓戰鬥規則能被人眼看著跑一遍。
//
//	tools/go.sh run ./cmd/battlesim -at 19 -from 26
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/great-era-remake/internal/game"
)

func main() {
	dir := flag.String("game", "workplace/orig/game", "原版素材目錄（唯讀）")
	at := flag.Int("at", 19, "戰場所在的省（守方）")
	from := flag.Int("from", 26, "攻方的來源省")
	rounds := flag.Int("rounds", 40, "最多打幾回合")
	flag.Parse()

	m, err := loadMap(*dir)
	if err != nil {
		fail(err)
	}

	// 攻守各兩個部隊。屬性用固定值只是為了讓輸出可預期；
	// 真實資料的版本在 `cmd/dsds` 與 `cmd/screenshot -battle`
	// （`MAN(N).DAT` 的欄位已全部對上，`docs/spec/02` §3）。
	atk := []*game.Combatant{
		unit(101, 58, game.Branch1, 20000, 80),
		unit(102, 58, game.Branch4, 2000, 70),
	}
	def := []*game.Combatant{
		unit(201, 166, game.Branch1, 18000, 75),
		unit(202, 166, game.Branch1, 12000, 65),
	}

	bf, err := m.Battlefield(game.ProvinceID(*at))
	if err != nil {
		fail(err)
	}
	// 守方擺在腹地（WARPOS 值 0）且站得上去的格。
	placed := 0
	for i := 0; i < game.CellCount && placed < len(def); i++ {
		c := game.CellIndex(i)
		col, row := c.ColRow()
		if bf.Owner[row][col] != 0 || bf.Tiles[row][col].MoveCost() >= 255 {
			continue
		}
		def[placed].Cell = c
		placed++
	}
	if placed < len(def) {
		fail(fmt.Errorf("省 %d 的腹地放不下 %d 個守方單位", *at, len(def)))
	}

	sim, err := game.NewBattleSim(m, game.ProvinceID(*at), game.ProvinceID(*from),
		atk, def, game.StrengthOpts{Stage: 1})
	if err != nil {
		fail(err)
	}

	fmt.Printf("戰場：省 %d　攻方來自：省 %d\n", *at, *from)
	fmt.Println("進場後的佈署")
	for _, u := range atk {
		col, row := u.Cell.ColRow()
		fmt.Printf("  攻 %d　格 %3d (欄%2d,列%2d)　地形 %2d　兵力 %5d　戰力 %d\n",
			u.General, u.Cell, col, row,
			bf.Tiles[row][col].Kind, u.Force(), sim.StrengthOf(u))
	}
	for _, u := range def {
		col, row := u.Cell.ColRow()
		fmt.Printf("  守 %d　格 %3d (欄%2d,列%2d)　地形 %2d　兵力 %5d　戰力 %d\n",
			u.General, u.Cell, col, row,
			bf.Tiles[row][col].Kind, u.Force(), sim.StrengthOf(u))
	}

	// 讓攻方朝守方推進，相鄰就開打。
	fmt.Println("\n回合")
	for r := 1; r <= *rounds; r++ {
		if over, won := sim.Over(); over {
			fmt.Printf("\n結束：%s\n", result(won))
			return
		}
		fmt.Printf("  第 %2d 回合　%s\n", r, snapshot(atk, def))
		sim.BeginTurn()
		acted := advance(sim, atk, def)
		sim.EndTurn()
		if dead := sim.Sweep(); len(dead) > 0 {
			fmt.Printf("            陣亡 %v\n", dead)
		}
		if !acted {
			fmt.Println("            沒有可行動的單位，停止")
			break
		}
	}
	if over, won := sim.Over(); over {
		fmt.Printf("\n結束：%s\n", result(won))
	} else {
		fmt.Printf("\n打滿 %d 回合未分勝負　攻方剩 %d／守方剩 %d\n",
			*rounds, totalForce(atk), totalForce(def))
	}
}

// advance 讓每個攻方單位朝最近的守方推進；相鄰就交戰。
// **這不是原版的 AI**——原版的行動決策還沒解出來（`docs/mechanics/70-ai.md`
// §7）。這裡只是把規則跑起來用的最簡策略。
func advance(sim *game.BattleSim, atk, def []*game.Combatant) bool {
	acted := false
	for _, a := range atk {
		if !a.Alive() {
			continue
		}
		target := nearest(a, def)
		if target == nil {
			continue
		}
		if game.Adjacent(a.Cell, target.Cell) {
			la, ld, err := sim.Engage(a, target)
			if err == nil {
				fmt.Printf("            交戰 %d×%d　攻損 %d（剩 %d）　守損 %d（剩 %d）\n",
					a.General, target.General, la, a.Force(), ld, target.Force())
				acted = true
			}
			continue
		}
		// 六個方向裡挑離目標最近的走，一直走到機動力用完
		// ——原版一個回合本來就能連走好幾格。
		for step(sim, a, target) {
			acted = true
			if game.Adjacent(a.Cell, target.Cell) {
				break // 接觸了就停下，交戰留到下一回合
			}
		}
	}
	return acted
}

// snapshot 印出每個單位的位置與兵力，讓人眼看得出戰線在動。
func snapshot(atk, def []*game.Combatant) string {
	s := "攻["
	for i, u := range atk {
		if i > 0 {
			s += " "
		}
		s += fmt.Sprintf("%d@%d:%d", u.General, u.Cell, u.Force())
	}
	s += "] 守["
	for i, u := range def {
		if i > 0 {
			s += " "
		}
		s += fmt.Sprintf("%d@%d:%d", u.General, u.Cell, u.Force())
	}
	return s + "]"
}

func step(sim *game.BattleSim, a, target *game.Combatant) bool {
	best, bestDist := game.HexDir(0), 1<<30
	ax, ay := target.Cell.ScreenXY()
	for d := game.DirLowerLeft; d <= game.DirUpperRight; d++ {
		n, ok := a.Cell.Neighbour(d)
		if !ok {
			continue
		}
		nx, ny := n.ScreenXY()
		dist := abs(nx-ax) + abs(ny-ay)
		if dist < bestDist {
			best, bestDist = d, dist
		}
	}
	if best == 0 {
		return false
	}
	_, err := sim.Move(a.General, best)
	return err == nil
}

func nearest(a *game.Combatant, list []*game.Combatant) *game.Combatant {
	var best *game.Combatant
	bestDist := 1 << 30
	ax, ay := a.Cell.ScreenXY()
	for _, u := range list {
		if !u.Alive() || !u.Cell.Valid() {
			continue
		}
		x, y := u.Cell.ScreenXY()
		if d := abs(x-ax) + abs(y-ay); d < bestDist {
			best, bestDist = u, d
		}
	}
	return best
}

func unit(id, faction game.GeneralID, branch uint8, force uint16, ability uint8) *game.Combatant {
	return &game.Combatant{
		CombatUnit: game.CombatUnit{
			General: id, Faction: faction, Cell: game.NoCell,
			Max: 12, Current: 12, Active: true, Decaying: 80,
		},
		Strength: game.StrengthInput{
			Ability: ability, Force: force, F19: 60, F20: 60, F29: 64, F30: 80,
			Branch: branch, General: id, Faction: faction,
		},
	}
}

func loadMap(dir string) (*game.Map, error) {
	read := func(n string) []byte {
		b, err := os.ReadFile(filepath.Join(dir, n))
		if err != nil {
			fail(err)
		}
		return b
	}
	return game.LoadMap(read("WARPOS.DAT"), read("TERNAME.DAT"), read("NWMAP.DAT"))
}

func totalForce(us []*game.Combatant) int {
	n := 0
	for _, u := range us {
		n += int(u.Force())
	}
	return n
}

func result(attackerWon bool) string {
	if attackerWon {
		return "攻方獲勝"
	}
	return "守方獲勝"
}

func abs(v int) int {
	if v < 0 {
		return -v
	}
	return v
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "battlesim:", err)
	os.Exit(1)
}
