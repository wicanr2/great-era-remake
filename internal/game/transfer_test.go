package game

import "testing"

// 調動的執行層測試（`docs/re/10`）。
//
// 這裡不驗「AI 決定往哪調」（那是 ai_test.go），只驗
// **搬動的規則本身**：誰被挑中、誰留下、狀態怎麼變。

// 滿員門檻與兵種無關——`兵力 × 人力權重 == 20000` 對四個兵種都成立。
// 這是 `docs/re/08` §4 那條公式在政略模組的第三個印證，
// 用真實將領表驗它確實對得上。
func TestForceRatioFullAtBranchCap(t *testing.T) {
	w := realWorld(t)
	full := map[uint8]int{
		BranchInfantry: 20000, BranchCavalry: 10000,
		BranchArtiller: 2000, BranchArmour: 200,
	}
	seen := map[uint8]bool{}
	for i := range w.Strengths {
		s := &w.Strengths[i]
		cap, ok := full[s.Branch]
		if !ok || int(s.Force) != cap {
			continue
		}
		seen[s.Branch] = true
		if got := w.forceRatio(i); got != transferFull {
			t.Errorf("兵種 %d 兵力 %d（滿員）算出 %d，應為 %d",
				s.Branch, s.Force, got, transferFull)
		}
	}
	if len(seen) < 2 {
		t.Fatalf("樣本不足：只有 %d 個兵種在 SAVE(1) 裡出現滿員的單位", len(seen))
	}
	t.Logf("驗過 %d 個兵種的滿員換算", len(seen))
}

// 模式 0 的迴圈跑到 count-1，**最後一個留下**。
// 這是原版寫死的行為（`docs/re/10` §3a），不是我們的選擇——
// 釘住它，免得哪天被「順手修好」。
func TestTransferAllLeavesLastBehind(t *testing.T) {
	w := realWorld(t)
	from := busiestProvince(w)
	before := w.RosterOf(from)
	n := before.Len()
	if n < 3 {
		t.Fatalf("挑到的省只有 %d 個將領，測不出留一個的行為", n)
	}
	last := before.Get(n)

	to := emptyNeighbour(t, w, from)
	rep := w.ApplyTransfer(from, TransferAll, to)

	if len(rep.Moved) != n-1 {
		t.Errorf("模式 0 從 %d 個將領裡搬了 %d 個，應為 %d",
			n, len(rep.Moved), n-1)
	}
	if w.Units[last].Province != from {
		t.Errorf("清單最後一個（索引 %d）被搬走了，原版會留下他", last)
	}
}

// 模式 6 只動第一個，不看兵力。
func TestTransferFirstMovesExactlyOne(t *testing.T) {
	w := realWorld(t)
	from := busiestProvince(w)
	first := w.RosterOf(from).Get(1)
	to := emptyNeighbour(t, w, from)

	rep := w.ApplyTransfer(from, TransferFirst, to)
	if len(rep.Moved) != 1 || rep.Moved[0] != first {
		t.Fatalf("模式 6 搬了 %v，應該只有 %d", rep.Moved, first)
	}
}

// 模式 1/5/7 只挑滿員的，其餘一個都不能動。
func TestTransferFullOnlySkipsUnderStrength(t *testing.T) {
	for _, mode := range []TransferMode{
		TransferFullOnly, TransferFullOnlyAll, TransferFullOnlyRoster,
	} {
		w := realWorld(t)
		from := busiestProvince(w)
		to := emptyNeighbour(t, w, from)
		ratios := map[int]int{}
		for _, i := range w.RosterOf(from).ids {
			ratios[i] = w.forceRatio(i)
		}

		rep := w.ApplyTransfer(from, mode, to)
		for _, i := range rep.Moved {
			if ratios[i] != transferFull {
				t.Errorf("模式 %d 搬走了非滿員的將領（比值 %d）", mode, ratios[i])
			}
		}
	}
}

// 模式 2 的額度上限是 5。
func TestTransferUnderHalfCapAtFive(t *testing.T) {
	w := realWorld(t)
	from := busiestProvince(w)
	to := emptyNeighbour(t, w, from)
	rep := w.ApplyTransfer(from, TransferUnderHalfCapped, to)
	if len(rep.Moved) > 5 {
		t.Fatalf("模式 2 搬了 %d 個，上限是 5", len(rep.Moved))
	}
	for _, i := range rep.Moved {
		if w.forceRatio(i) >= transferHalf {
			t.Errorf("模式 2 搬走了達半滿的將領（比值 %d）", w.forceRatio(i))
		}
	}
}

// 勢力領袖本人在模式 2/3 不會被調走。
func TestTransferKeepsCommanderInPlace(t *testing.T) {
	w := realWorld(t)
	// 找一個省：司令本人在省內，而且有不同勢力的鄰省可以調過去。
	var from, to ProvinceID
	for p := ProvinceID(1); p <= ProvinceCount && from == 0; p++ {
		prov, err := w.Table.At(p)
		if err != nil || prov.Commander == 0 {
			continue
		}
		has := false
		for _, i := range w.RosterOf(p).ids {
			has = has || w.Units[i].General == prov.Commander
		}
		if !has {
			continue
		}
		for _, n := range w.neighbours(prov) {
			if q, err := w.Table.At(n); err == nil && q.Commander != prov.Commander {
				from, to = p, n
				break
			}
		}
	}
	if from == 0 {
		t.Skip("SAVE(1) 裡沒有「司令在省內且有異勢力鄰省」的省")
	}
	prov, _ := w.Table.At(from)

	for _, mode := range []TransferMode{TransferUnderHalfCapped, TransferUnderHalfTwo} {
		rep := w.ApplyTransfer(from, mode, to)
		for _, i := range rep.Moved {
			if w.Units[i].General == prov.Commander {
				t.Errorf("模式 %d 把勢力領袖 %d 調走了", mode, prov.Commander)
			}
		}
	}
}

// 調動即佔領：`ApplyAction` 之後目標省的司令變成我方。
func TestApplyActionClaimsProvince(t *testing.T) {
	w := realWorld(t)
	from := busiestProvince(w)
	to := emptyNeighbour(t, w, from)
	src, _ := w.Table.At(from)
	want := src.Commander

	dst, _ := w.Table.At(to)
	if dst.Commander == want {
		t.Fatalf("目標省 %d 本來就是同一個勢力，測不出佔領", to)
	}
	rep := w.ApplyAction(AIAction{
		Kind: AITransfer, From: from, To: to, TransferKind: int(TransferFirst),
	})
	if len(rep.Moved) == 0 {
		t.Fatal("沒有將領被搬動，佔領無從發生")
	}
	if got, _ := w.Table.At(to); got.Commander != want {
		t.Errorf("省 %d 的司令是 %d，調動後應變成 %d", to, got.Commander, want)
	}
}

// Turbo Pascal 的 Round 是「.5 取偶」，不是 (n+1)/2。
// 模式 4 的掃描起點靠它，差一格就會挑到不同的人。
func TestRoundHalfBanker(t *testing.T) {
	want := map[int]int{
		1: 0, 2: 1, 3: 2, 4: 2, 5: 2, 6: 3, 7: 4, 8: 4, 9: 4, 10: 5, 11: 6,
	}
	for n, w := range want {
		if got := roundHalf(n); got != w {
			t.Errorf("roundHalf(%d) = %d，應為 %d", n, got, w)
		}
	}
}

// 清單的定義（`sub_5A881` 與 `sub_306CF` 的三條篩選）必須與
// `GeneralCount` 一致——原版那兩支用的是同一組條件。
func TestRosterMatchesGeneralCount(t *testing.T) {
	w := realWorld(t)
	checked := 0
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		if w.RosterOf(p).Len() != w.GeneralCount(p) {
			t.Errorf("省 %d：清單長度 %d 與 GeneralCount %d 不符",
				p, w.RosterOf(p).Len(), w.GeneralCount(p))
		}
		checked++
	}
	t.Logf("比對 %d 省的清單長度", checked)
}

// busiestProvince 找將領最多的省，讓迴圈邊界測得出來。
func busiestProvince(w *AIWorld) ProvinceID {
	best, n := ProvinceID(0), 0
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		if c := w.RosterOf(p).Len(); c > n {
			best, n = p, c
		}
	}
	return best
}

// emptyNeighbour 找一個不同勢力的鄰省當調動目標。
func emptyNeighbour(t *testing.T, w *AIWorld, from ProvinceID) ProvinceID {
	t.Helper()
	prov, err := w.Table.At(from)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range w.neighbours(prov) {
		if p, err := w.Table.At(n); err == nil && p.Commander != prov.Commander {
			return n
		}
	}
	t.Fatalf("省 %d 找不到不同勢力的鄰省", from)
	return 0
}

// 一般調動只往**無主**省。這條在第一版寫反了（寫成「只往有主省」），
// 症狀是兩個省每回合互相搬兵、局面永遠不變（`docs/playtest/05`）。
//
// 反組譯的依據是 `cmp word ptr [di-6221h], 0 / jnz skip`——
// jnz 跳的是 skip，留下來的只有司令 == 0 的省。
func TestGeneralTransferTargetsOnlyUnowned(t *testing.T) {
	w := realWorld(t)
	checked := 0
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		prov, err := w.Table.At(p)
		if err != nil || prov.Commander == 0 {
			continue
		}
		for _, mode := range []int{1, 2} {
			a := w.generalTransfer(p, mode)
			if a.Kind == AINone {
				continue
			}
			checked++
			to, err := w.Table.At(a.To)
			if err != nil {
				t.Fatal(err)
			}
			if to.Commander != 0 {
				t.Errorf("省 %d 模式 %d 調往省 %d，但那個省的司令是 %d（應為無主）",
					p, mode, a.To, to.Commander)
			}
		}
	}
	t.Logf("檢查 %d 個一般調動的目標", checked)
}

// 決策鏈對「調動失敗」有回饋：`sub_14F9A` 沒搬成就不立旗標，
// 鏈繼續往下走（`docs/re/10` §4）。
//
// 少了它會出現大量「決定了但什麼也沒發生」的空轉——`Step` 回報的
// 每一個調動動作都必須真的搬到人。
func TestStepNeverReportsEmptyTransfer(t *testing.T) {
	w := realWorld(t)
	acted := 0
	for turn := 0; turn < 3; turn++ {
		for p := ProvinceID(1); p <= ProvinceCount; p++ {
			prov, err := w.Table.At(p)
			if err != nil || prov.Commander == 0 {
				continue
			}
			a, rep := w.Step(p)
			if a.Kind != AITransfer {
				continue
			}
			acted++
			if len(rep.Moved) == 0 {
				t.Fatalf("省 %d 的 %s 回報調動卻沒搬動任何人", p, a.Step)
			}
		}
	}
	if acted == 0 {
		t.Skip("三回合內沒有任何調動成立")
	}
	t.Logf("驗過 %d 次實際成立的調動", acted)
}

// 省份層尋路：路徑必須連續（每一步都相鄰）且不重複。
//
// 原版用 `sub_5B76E`（鄰接檢查）驗同樣的事（`docs/re/11` §5），
// 這裡拿真實鄰省表把全國兩兩配對都跑一遍。
func TestProvincePathIsContiguous(t *testing.T) {
	w := realWorld(t)
	found, checked := 0, 0
	for a := ProvinceID(1); a <= ProvinceCount; a++ {
		pa, err := w.Table.At(a)
		if err != nil || pa.Commander == 0 {
			continue
		}
		for b := ProvinceID(1); b <= ProvinceCount; b++ {
			if a == b {
				continue
			}
			checked++
			path := w.ProvincePath(a, b, pa.Commander)
			if path == nil {
				continue
			}
			found++
			if path[0] != a || path[len(path)-1] != b {
				t.Fatalf("路徑 %v 的頭尾不是 %d → %d", path, a, b)
			}
			seen := map[ProvinceID]bool{}
			for i, p := range path {
				if seen[p] {
					t.Fatalf("路徑 %v 重複經過省 %d", path, p)
				}
				seen[p] = true
				if i > 0 && !w.adjacentProvinces(path[i-1], p) {
					t.Errorf("路徑 %v 裡省 %d 與 %d 不相鄰", path, path[i-1], p)
				}
			}
		}
	}
	if found == 0 {
		t.Fatal("一條路徑都找不到，尋路沒有真的在跑")
	}
	t.Logf("%d 組配對裡找到 %d 條路徑", checked, found)
}

// NextHop 回傳的必須是**與出發省相鄰**的那一格，不是目的地本身。
func TestNextHopIsAdjacent(t *testing.T) {
	w := realWorld(t)
	n := 0
	for a := ProvinceID(1); a <= ProvinceCount; a++ {
		pa, err := w.Table.At(a)
		if err != nil || pa.Commander == 0 {
			continue
		}
		for _, target := range w.RescueTargets(pa.Commander) {
			hop := w.NextHop(a, target, pa.Commander)
			if hop == 0 {
				continue
			}
			n++
			if !w.adjacentProvinces(hop, a) {
				t.Errorf("省 %d 往 %d 的第一步是 %d，但它與 %d 不相鄰", a, target, hop, a)
			}
		}
	}
	t.Logf("驗過 %d 個第一步", n)
}

// 救援候選的四道篩選都要成立（`sub_149C8`）。
func TestRescueTargetsMatchFilters(t *testing.T) {
	w := realWorld(t)
	total := 0
	for f := range w.Table.Factions() {
		for _, p := range w.RescueTargets(f) {
			total++
			prov, err := w.Table.At(p)
			if err != nil {
				t.Fatal(err)
			}
			if prov.Commander != f {
				t.Errorf("省 %d 的司令是 %d，不是 %d", p, prov.Commander, f)
			}
			if w.Hostile(p) == 0 {
				t.Errorf("省 %d 不是前線，不該進救援候選", p)
			}
			if pw := w.ProvincePower(p); pw >= RescueThreshold {
				t.Errorf("省 %d 的戰力 %d 沒有低於門檻 %d", p, pw, RescueThreshold)
			}
			if prov.Flags&ProvinceFlagInBattle != 0 {
				t.Errorf("省 %d 正在交戰，不該進救援候選", p)
			}
		}
	}
	t.Logf("SAVE(1) 的救援候選共 %d 個", total)
}

// 清單依攻擊力**降序**（`sub_13B44`）。這條先前標「未驗」，
// 現在是 confirmed，而且它決定了模式 0/2/3/4/6 挑到誰。
func TestRosterSortedByStrengthDesc(t *testing.T) {
	w := realWorld(t)
	checked := 0
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		r := w.RosterOf(p)
		if r.Len() < 2 {
			continue
		}
		checked++
		for i := 1; i < r.Len(); i++ {
			a, b := w.strengthAt(r.Get(i)), w.strengthAt(r.Get(i+1))
			if a < b {
				t.Errorf("省 %d 清單第 %d 個戰力 %d < 第 %d 個 %d，沒有降序",
					p, i, a, i+1, b)
			}
		}
	}
	if checked == 0 {
		t.Fatal("沒有省有兩個以上將領，測不出排序")
	}
	t.Logf("驗過 %d 個省的清單排序", checked)
}

// 排序的實際後果：模式 6 派出最強的，模式 0 把最弱的留下。
func TestTransferModesPickByStrength(t *testing.T) {
	w := realWorld(t)
	from := busiestProvince(w)
	r := w.RosterOf(from)
	if r.Len() < 3 {
		t.Fatalf("省 %d 只有 %d 個將領", from, r.Len())
	}
	strongest, weakest := r.Get(1), r.Get(r.Len())
	if w.strengthAt(strongest) < w.strengthAt(weakest) {
		t.Fatal("排序沒生效，後面的測試沒有意義")
	}
	to := emptyNeighbour(t, w, from)

	rep := w.ApplyTransfer(from, TransferFirst, to)
	if len(rep.Moved) != 1 || rep.Moved[0] != strongest {
		t.Errorf("模式 6 該派出最強的 %d，實際搬了 %v", strongest, rep.Moved)
	}

	w2 := realWorld(t)
	weakest2 := w2.RosterOf(from).Get(w2.RosterOf(from).Len())
	w2.ApplyTransfer(from, TransferAll, to)
	if w2.Units[weakest2].Province != from {
		t.Errorf("模式 0 該把最弱的 %d 留下，但他被搬走了", weakest2)
	}
}

// 命令數 = 將領數 ÷ 8 + 1（`sub_13D23`）。
//
// 這是社群那條「每個省份有命令數上限」的程式碼依據
// ——先前找錯格子（`docs/re/10` §1 推翻的那次），現在找到了。
func TestCommandsFormula(t *testing.T) {
	w := realWorld(t)
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		n := w.RosterOf(p).Len()
		if got, want := w.CommandsFor(p), n/8+1; got != want {
			t.Errorf("省 %d 有 %d 個將領，命令數算出 %d，應為 %d", p, n, got, want)
		}
	}
	// 沒有將領的省仍然有 1 個命令——原版是先除再加一，不是特例。
	empty := ProvinceID(0)
	for p := ProvinceID(1); p <= ProvinceCount; p++ {
		if w.RosterOf(p).Len() == 0 {
			empty = p
			break
		}
	}
	if empty != 0 && w.CommandsFor(empty) != 1 {
		t.Errorf("無將領的省 %d 命令數是 %d，公式說應為 1", empty, w.CommandsFor(empty))
	}
	t.Logf("最忙的省有 %d 個命令", w.CommandsFor(busiestProvince(w)))
}
