// aisim 讓電腦連續跑政略回合，看局面怎麼演變。
//
//	tools/go.sh run ./cmd/aisim -game workplace/orig/game -turns 20
//
// 用途是**實跑驗收**：單元測試只驗單一規則，串起來會不會收斂
// （還是所有兵擠成一堆、還是誰都不動）只有跑才看得出來。
// 對照 docs/playtest/04 在戰鬥層做的同一件事。
//
// ⚠️ 這不是與原版對照。DOSBox 的輸入問題（docs/playtest/03）沒解之前
// 拿不到實機的逐回合狀態。
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/wicanr2/great-era-remake/internal/game"
)

func main() {
	dir := flag.String("game", "workplace/orig/game", "原版素材目錄（唯讀）")
	turns := flag.Int("turns", 20, "跑幾個回合")
	verbose := flag.Bool("v", false, "逐省印出決策")
	flag.Parse()

	if err := run(*dir, *turns, *verbose); err != nil {
		fmt.Fprintln(os.Stderr, "錯誤:", err)
		os.Exit(1)
	}
}

func run(dir string, turns int, verbose bool) error {
	read := func(name string) ([]byte, error) {
		return os.ReadFile(filepath.Join(dir, name))
	}
	tblRaw, err := read("SAVE(1).DT1")
	if err != nil {
		return err
	}
	tbl, err := game.ParseSaveProvinces(tblRaw)
	if err != nil {
		return err
	}
	manRaw, err := read("MAN(1).DAT")
	if err != nil {
		return err
	}
	gs, err := game.ParseGenerals(manRaw, len(manRaw)/game.GeneralRecordSize)
	if err != nil {
		return err
	}

	w := &game.AIWorld{Table: tbl, Opts: game.StrengthOpts{Stage: 1}, EnableExtra: true}
	for i := range gs {
		g := &gs[i]
		id := game.GeneralID(i + 1)
		var faction game.GeneralID
		if p, err := tbl.At(g.Province); err == nil {
			faction = p.Commander
		}
		w.Units = append(w.Units, game.CombatUnit{
			General: id, Province: g.Province, Faction: faction,
			Active: g.Raw[16]&1 == 1,
		})
		w.Strengths = append(w.Strengths, game.StrengthInput{
			Ability: g.AbilityA, Force: g.Force,
			F19: g.F19, F20: g.F20, F29: g.F29, F30: g.F30,
			Branch: g.Branch, General: id, Faction: faction,
		})
	}

	fmt.Println("開局：", summary(w))

	stats := map[string]int{}
	moved := 0
	for t := 1; t <= turns; t++ {
		for p := game.ProvinceID(1); p <= game.ProvinceCount; p++ {
			prov, err := tbl.At(p)
			if err != nil || prov.Commander == 0 {
				continue
			}
			a, rep := w.Step(p)
			if a.Kind == game.AINone {
				continue
			}
			stats[a.Step]++
			moved += len(rep.Moved)
			if verbose && len(rep.Moved) > 0 {
				fmt.Printf("  T%02d 省 %2d → %2d　%s　模式 %d　搬 %d 人%s\n",
					t, a.From, a.To, a.Step, a.TransferKind, len(rep.Moved), alertMark(rep))
			}
		}
		if t%5 == 0 || t == turns {
			fmt.Printf("第 %2d 回合後：%s\n", t, summary(w))
		}
	}

	fmt.Println("\n決策分佈")
	keys := make([]string, 0, len(stats))
	for k := range stats {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Printf("  %-28s %4d\n", k, stats[k])
	}
	fmt.Printf("\n累計搬動 %d 人次\n", moved)
	return nil
}

func alertMark(r game.TransferReport) string {
	if r.Alert {
		return "　⚠ 大軍集結提示"
	}
	return ""
}

// summary 印一行局面：幾個勢力、佔了幾省、將領最集中的省有幾人。
//
// **看狀態不是看畫面**（`CLAUDE.md` §7 第 6 條）——兵力全擠到一省
// 這種失控只有數字看得出來。
func summary(w *game.AIWorld) string {
	factions := w.Table.Factions()
	owned := 0
	for _, ps := range factions {
		owned += len(ps)
	}
	top, topN := game.ProvinceID(0), 0
	empty := 0
	for p := game.ProvinceID(1); p <= game.ProvinceCount; p++ {
		n := w.RosterOf(p).Len()
		if n > topN {
			top, topN = p, n
		}
		if n == 0 {
			empty++
		}
	}
	return fmt.Sprintf("勢力 %d／有主省 %d／無將領的省 %d／最集中：省 %d 有 %d 人",
		len(factions), owned, empty, top, topN)
}
