package game

import "fmt"

// 把政略層與戰鬥層接起來：電腦決定攻打之後，真的打一場。
//
// 原版的流程是政略階段選「攻打」→ 進戰鬥模組（`docs/re/06` 的三個 gate）
// → 打完回到政略。這裡做同一件事，差別是戰場上的行動決策用
// `AutoResolve` 的簡化策略（見那份檔案的警告）。

// ResolveAttack 讓 from 省的部隊打 at 省，跑完整場戰鬥並把結果寫回世界。
//
// 三件事會改變：
//
//  1. 雙方將領的兵力（`+17`）依戰損扣減，歸零的標記為不可用
//  2. 勝方佔領該省（`Province.Capture`，`+20` 改成勝方勢力）
//  3. 攻方存活的將領移防到 at 省——**這是 remake 的補完**，
//     原版打完之後部隊怎麼安置還沒解（`docs/re/06` 只解到進戰鬥）
func (w *AIWorld) ResolveAttack(from, at ProvinceID) (BattleOutcome, error) {
	if w.Map == nil {
		return BattleOutcome{}, fmt.Errorf("game: ResolveAttack 需要 Map")
	}
	src, err := w.Table.At(from)
	if err != nil {
		return BattleOutcome{}, err
	}
	dst, err := w.Table.At(at)
	if err != nil {
		return BattleOutcome{}, err
	}

	atk, atkIdx := w.combatants(from, src.Commander)
	def, defIdx := w.combatants(at, dst.Commander)
	if len(atk) == 0 {
		return BattleOutcome{}, fmt.Errorf("game: 省 %d 沒有可出兵的將領", from)
	}
	if len(def) == 0 {
		// 沒有守軍：直接易主，不用打。原版有沒有這條捷徑未解，
		// 但戰鬥模組要求雙方都有部隊（`NewBattleSim` 會擋），
		// 所以這裡必須有個出口。**標為 remake 差異。**
		//
		// ⚠️ 佔領之後**攻方一定要有人移防過去**，否則這個省變成
		// 「有主但沒有將領」，下一回合又被別人白拿——連續跑時
		// 會看到「拿下 151 省」但有主省只有 31 個（`docs/playtest/06` §2）。
		dst.Capture(src.Commander)
		w.Units[atkIdx[0]].Province = at
		return BattleOutcome{AttackerWon: true, Decided: true}, nil
	}

	bf, err := w.Map.Battlefield(at)
	if err != nil {
		return BattleOutcome{}, err
	}
	// 守方擺在腹地（`WARPOS` 值 0）且站得上去的格。
	// 原版守方的落點從哪來還沒解（`sub_41513` 只放攻方），標為差異。
	placed := 0
	for i := 0; i < CellCount && placed < len(def); i++ {
		c := CellIndex(i)
		col, row := c.ColRow()
		if bf.Owner[row][col] != 0 || bf.Tiles[row][col].Kind.Blocks() {
			continue
		}
		def[placed].Cell = c
		placed++
	}
	if placed == 0 {
		return BattleOutcome{}, fmt.Errorf("game: 省 %d 的腹地放不下守軍", at)
	}
	def, defIdx = def[:placed], defIdx[:placed]

	sim, err := NewBattleSim(w.Map, at, from, atk, def, w.Opts)
	if err != nil {
		return BattleOutcome{}, err
	}
	out := sim.AutoResolve(AutoBattleTurnCap)

	// 兵力寫回世界。
	w.writeBack(atk, atkIdx)
	w.writeBack(def, defIdx)

	if out.Decided && out.AttackerWon {
		dst.Capture(src.Commander)
		// 活下來的攻方移防過去。
		for k, u := range atk {
			if u.Alive() {
				w.Units[atkIdx[k]].Province = at
			}
		}
	} else {
		// 守住了：清掉「正在打仗」旗標，司令不變。
		dst.Capture(0)
	}
	return out, nil
}

// combatants 把某省的將領換成戰場單位，並回傳它們在 `Units` 裡的索引。
//
// 清單的定義照 `sub_5A881`（`docs/re/10` §6），最多取 `UnitsPerSide` 個
// ——原版每方 10 個部隊槽（`docs/re/05`）。
func (w *AIWorld) combatants(p ProvinceID, faction GeneralID) ([]*Combatant, []int) {
	var out []*Combatant
	var idx []int
	for _, i := range w.RosterOf(p).ids {
		if len(out) >= UnitsPerSide {
			break
		}
		if i >= len(w.Strengths) {
			continue
		}
		u := w.Units[i]
		u.Cell = NoCell
		u.Max, u.Current = 12, 12
		u.Active, u.Decaying = true, 80
		u.Faction = faction
		out = append(out, &Combatant{CombatUnit: u, Strength: w.Strengths[i]})
		idx = append(idx, i)
	}
	return out, idx
}

// writeBack 把戰損寫回世界狀態。
//
// ⛔ 這裡原本多做一件事：「兵力歸零的將領標記為不可用」。
// **那條規則是我們發明的，原版沒有依據**——`sub_28259`（徵兵）
// 檢查的是 `+16 == 1`，跟兵力無關，所以兵力 0 的將領照樣能被補兵。
//
// 那條規則的代價：`docs/playtest/09` 的長跑會枯竭，而且補上徵兵之後
// **仍然**枯竭——因為被標成不可用的將領再也回不來（`docs/playtest/11`）。
//
// 原版兵力歸零的將領會怎樣仍未解（`+16` 的 `{0, 1, 32}` 三個值、
// 畫面上的「狀態：任用」都還沒對上），**所以這裡什麼都不做**。
// `CLAUDE.md` §9：不准為了讓行為看起來合理而編規則。
func (w *AIWorld) writeBack(us []*Combatant, idx []int) {
	for k, u := range us {
		w.Strengths[idx[k]].Force = u.Strength.Force
	}
}
