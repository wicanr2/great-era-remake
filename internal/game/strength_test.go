package game

import (
	"math"
	"testing"
)

// Turbo Pascal 的 48-bit Real：byte 0 是 biased exponent（bias 129），
// bytes 1..5 是尾數，最高位元是符號，尾數的最高有效位隱含為 1。
//
// 這個函式是**驗證用的**，不是遊戲邏輯——它把反組譯裡看到的暫存器值
// 還原成數字，證明我們沒有解錯。
func realFromRegs(cx, si, di uint16) float64 {
	exp := int(cx & 0xFF)
	// 尾數 bytes 由低到高：CH, SI 低, SI 高, DI 低, DI 高
	m := []byte{
		byte(cx >> 8), byte(si), byte(si >> 8), byte(di), byte(di >> 8),
	}
	if exp == 0 {
		return 0
	}
	// 最高 byte 的 bit7 是符號，其餘 7 bits 是尾數的最高 7 位。
	frac := 1.0
	bit := 0.5
	top := m[4]
	for i := 6; i >= 0; i-- {
		if top&(1<<uint(i)) != 0 {
			frac += bit
		}
		bit /= 2
	}
	for k := 3; k >= 0; k-- {
		for i := 7; i >= 0; i-- {
			if m[k]&(1<<uint(i)) != 0 {
				frac += bit
			}
			bit /= 2
		}
	}
	v := frac * math.Pow(2, float64(exp-129))
	if top&0x80 != 0 {
		v = -v
	}
	return v
}

// 七個 Real 常數全部解成整齊的整數——這本身就是「沒解錯」的證明。
// 如果格式理解錯了，不可能同時算出 9、10、80、85、20000 這種數字。
func TestRealConstantsDecodeToWholeNumbers(t *testing.T) {
	cases := []struct {
		cx, si, di uint16
		want       float64
		where      string
	}{
		{0x0084, 0, 0x2000, 10, "sub_5A0B9 第一步的分母"},
		{0x008F, 0, 0x1C40, 20000, "sub_5A0B9 第一步的總分母 ← 20000 藏在這裡"},
		{0x0084, 0, 0x1000, 9, "sqrt 那一步的分母"},
		{0x0087, 0, 0x2A00, 85, "F30 那一步的分母"},
		{0x0087, 0, 0x2000, 80, "Ability 那一步的分母"},
		{0x0081, 0, 0x4000, 1.5, "勢力加成的倍率"},
		{0x0083, 0, 0x2000, 5, "sub_54826 每回合衰減的除數"},
		{0x0084, 0, 0x2000, 10, "sub_53DA9 戰鬥後衰減的除數"},
		{0x007F, 0, 0, 0.25, "sub_50655 的斜率"},
		{0x0080, 0, 0, 0.5, "sub_50600 的係數"},
		{0x0081, 0, 0, 1, "sub_50655 的截距"},
	}
	for _, c := range cases {
		got := realFromRegs(c.cx, c.si, c.di)
		if math.Abs(got-c.want) > 1e-9 {
			t.Errorf("Real(%#04x,%#04x,%#04x) = %v，應為 %v（%s）",
				c.cx, c.si, c.di, got, c.want, c.where)
		}
	}
}

// 20000 值得單獨一條：CLAUDE.md §1.5 記過「grep 不到是正常的」。
// 它不是整數常數，是 48-bit Real 的尾數，所以 grep 20000 或 4E20 都不會中。
func TestTwentyThousandIsHiddenInAReal(t *testing.T) {
	if got := realFromRegs(0x008F, 0, 0x1C40); got != 20000 {
		t.Fatalf("解出 %v，應為 20000", got)
	}
	// 這個 Real 的 6 個 bytes 是 8F 00 00 00 40 1C——
	// 裡面沒有 20000 的小端表示 20 4E，也沒有十進位字串。
	raw := []byte{0x8F, 0x00, 0x00, 0x00, 0x40, 0x1C}
	for i := 0; i+1 < len(raw); i++ {
		if raw[i] == 0x20 && raw[i+1] == 0x4E {
			t.Error("居然找得到 4E20 的小端表示，那 grep 就會中了")
		}
	}
}

// 兵種權重：值域只有 1/4/5/6，其餘回 0。
func TestBranchWeight(t *testing.T) {
	want := map[uint8]int{1: 1, 4: 10, 5: 100, 6: 4}
	for b, w := range want {
		if got := BranchWeight(b); got != w {
			t.Errorf("兵種 %d 的權重 = %d，應為 %d", b, got, w)
		}
	}
	for _, b := range []uint8{0, 2, 3, 7, 255} {
		if got := BranchWeight(b); got != 0 {
			t.Errorf("兵種 %d 不在值域，應回 0，得到 %d", b, got)
		}
	}
}

// 戰力隨兵力線性成長，隨能力值階梯式跳升。
func TestStrengthMonotonic(t *testing.T) {
	base := StrengthInput{
		Ability: 80, Force: 10000, F19: 50, F20: 50, F29: 81, F30: 80,
		Branch: 1, General: 58, Faction: 58,
	}
	var opt StrengthOpts

	prev := -1
	for f := uint16(2000); f <= 20000; f += 2000 {
		in := base
		in.Force = f
		got := Strength(in, opt)
		if got <= prev {
			t.Errorf("兵力 %d 的戰力 = %d，沒有比前一級（%d）高", f, got, prev)
		}
		prev = got
	}

	// 兵力 0 → 戰力 0。
	zero := base
	zero.Force = 0
	if got := Strength(zero, opt); got != 0 {
		t.Errorf("兵力 0 的戰力 = %d，應為 0", got)
	}
}

// 名將加成的五個門檻：85／90／95／98／100，各自的倍率不同。
func TestStrengthAbilityTiers(t *testing.T) {
	mk := func(a uint8) StrengthInput {
		return StrengthInput{
			Ability: a, Force: 20000, F19: 60, F20: 60, F29: 100, F30: 100,
			Branch: 1, General: 58, Faction: 58,
		}
	}
	var opt StrengthOpts

	// 84 沒有加成、85 有——門檻是 >=。
	if Strength(mk(85), opt) <= Strength(mk(84), opt) {
		t.Error("能力 85 應該開始有加成")
	}
	// 各級單調遞增。
	tiers := []uint8{84, 85, 90, 95, 98, 100}
	prev := -1
	for _, a := range tiers {
		got := Strength(mk(a), opt)
		if got <= prev {
			t.Errorf("能力 %d 的戰力 = %d，沒有比前一級（%d）高", a, got, prev)
		}
		prev = got
	}

	// 滿分 100 是 ×5/3，比 98 的 ×1.5 高。
	if a100, a98 := Strength(mk(100), opt), Strength(mk(98), opt); a100 <= a98 {
		t.Errorf("能力 100（×5/3）= %d 應高於 98（×1.5）= %d", a100, a98)
	}
}

// 兵種 5 走兩條加成：權重 100（最高）+ 最後再 ×2。
func TestStrengthBranchFiveIsDoubledTwice(t *testing.T) {
	mk := func(b uint8) StrengthInput {
		return StrengthInput{
			Ability: 70, Force: 20000, F19: 50, F20: 50, F29: 64, F30: 70,
			Branch: b, General: 58, Faction: 58,
		}
	}
	var opt StrengthOpts
	b1, b4, b6, b5 := Strength(mk(1), opt), Strength(mk(4), opt),
		Strength(mk(6), opt), Strength(mk(5), opt)
	// 權重序：1 < 4(=權重4) < 4(=權重10) < 5(=權重100，再 ×2)
	if !(b1 < b6 && b6 < b4 && b4 < b5) {
		t.Errorf("兵種強弱序不對：1=%d 6=%d 4=%d 5=%d", b1, b6, b4, b5)
	}
	// 兵種 5 的最後 ×2 讓它比「權重 100 但不翻倍」還高一倍。
	// 直接比比例：權重 100 vs 10 是 10 倍，再 ×2 → 約 20 倍。
	// 中間有四次 Round，所以只驗量級不驗精確值。
	if ratio := float64(b5) / float64(b4); ratio < 15 || ratio > 25 {
		t.Errorf("兵種 5 / 兵種 4 = %.1f，應在 20 倍上下（權重 10 倍再 ×2）", ratio)
	}
	// 未知兵種回 0。
	if got := Strength(mk(3), opt); got != 0 {
		t.Errorf("未知兵種的戰力 = %d，應為 0", got)
	}
}

// 勢力／難度加成的三種設定。
func TestStrengthFactionBonus(t *testing.T) {
	in := StrengthInput{
		Ability: 70, Force: 10000, F19: 50, F20: 50, F29: 64, F30: 70,
		Branch: 1, General: 58, Faction: 58,
	}
	major := func(GeneralID) bool { return true }
	plain := Strength(in, StrengthOpts{MajorPower: major})

	if got := Strength(in, StrengthOpts{FactionBonus: BonusDouble, MajorPower: major}); got != plain*2 {
		t.Errorf("×2 設定下 = %d，應為 %d", got, plain*2)
	}
	if got := Strength(in, StrengthOpts{FactionBonus: BonusHalfMore, MajorPower: major}); got != round(float64(plain)*1.5) {
		t.Errorf("×1.5 設定下 = %d，應為 %d", got, round(float64(plain)*1.5))
	}
	// BonusLeaderOnly 只對蔣中正（ID 1）生效。
	if got := Strength(in, StrengthOpts{FactionBonus: BonusLeaderOnly, MajorPower: major}); got != plain {
		t.Errorf("吳佩孚在 LeaderOnly 設定下不該加成，得到 %d（基準 %d）", got, plain)
	}
	chiang := in
	chiang.Faction = LeaderChiang
	if got := Strength(chiang, StrengthOpts{FactionBonus: BonusLeaderOnly, MajorPower: major}); got != round(float64(plain)*1.5) {
		t.Errorf("蔣中正在 LeaderOnly 設定下 = %d，應為 %d", got, round(float64(plain)*1.5))
	}

	// 不是十大勢力時走另一條分支：只有蔣中正 + 第一期才 ×1.5。
	none := func(GeneralID) bool { return false }
	if got := Strength(chiang, StrengthOpts{MajorPower: none, Stage: 1}); got != round(float64(plain)*1.5) {
		t.Errorf("非十大勢力的蔣中正（第一期）= %d，應為 %d", got, round(float64(plain)*1.5))
	}
	if got := Strength(chiang, StrengthOpts{MajorPower: none, Stage: 2}); got != plain {
		t.Errorf("第二期不該套用那條加成，得到 %d（基準 %d）", got, plain)
	}
}

// 一面倒用的兵種權重是**另一張表**：兵種 6 是 2，不是攻擊力那張的 4。
func TestRoutWeightDiffersFromStrengthWeight(t *testing.T) {
	same := []uint8{1, 4, 5}
	for _, b := range same {
		if BranchWeight(b) != RoutWeight(b) {
			t.Errorf("兵種 %d 兩張表應該相同：%d vs %d", b, BranchWeight(b), RoutWeight(b))
		}
	}
	if BranchWeight(6) != 4 || RoutWeight(6) != 2 {
		t.Errorf("兵種 6 應該是 4（攻擊力）與 2（一面倒），得到 %d 與 %d",
			BranchWeight(6), RoutWeight(6))
	}
}

// 勢均力敵：雙方損失比例合計恆為 1/4，戰力相等時各 1/8。
func TestCasualtiesEvenSplitsAQuarter(t *testing.T) {
	const f = 20000
	// 戰力相等 → pct = 100 → 各損失 1/8 = 2500
	lossE, lossF := CasualtiesEven(1000, 1000, f, f, 1, 1)
	if lossE != f/8 || lossF != f/8 {
		t.Errorf("戰力相等時應各損失 %d，得到 %d / %d", f/8, lossE, lossF)
	}

	// E 是 F 的 3 倍強 → pct = 300 → F 損 3/4×1/4、E 損 1/4×1/4
	lossE, lossF = CasualtiesEven(3000, 1000, f, f, 1, 1)
	if wantF := f * 3 / 16; lossF != wantF {
		t.Errorf("F 的損失 = %d，應為 %d", lossF, wantF)
	}
	if wantE := f / 16; lossE != wantE {
		t.Errorf("E 的損失 = %d，應為 %d", lossE, wantE)
	}
	// 強的一方損失比較少——這是這條公式的核心性質。
	if lossE >= lossF {
		t.Errorf("E 比較強卻損失比較多：E=%d F=%d", lossE, lossF)
	}

	// 損失永遠不超過兵力。
	lossE, lossF = CasualtiesEven(1000000, 1, 100, 100, 1, 1)
	if lossE > 100 || lossF > 100 {
		t.Errorf("損失超過兵力：E=%d F=%d", lossE, lossF)
	}
}

// 兵種 4 兩條特例：pct 減半、而且任一方是 4 時 F 完全不受損失。
func TestCasualtiesEvenBranchFourExceptions(t *testing.T) {
	const f = 20000
	// E 是兵種 4 → F 不受損失
	lossE, lossF := CasualtiesEven(1000, 1000, f, f, 4, 1)
	if lossF != 0 {
		t.Errorf("E 是兵種 4 時 F 不該有損失，得到 %d", lossF)
	}
	if lossE == 0 {
		t.Error("E 自己還是要受損失")
	}
	// F 是兵種 4 → 一樣，F 不受損失
	_, lossF = CasualtiesEven(1000, 1000, f, f, 1, 4)
	if lossF != 0 {
		t.Errorf("F 是兵種 4 時也不該有損失，得到 %d", lossF)
	}
	// E 是兵種 4 時 pct 減半 → E 自己的損失變大（100/(pct+100) 變大）
	normalE, _ := CasualtiesEven(1000, 1000, f, f, 1, 1)
	halvedE, _ := CasualtiesEven(1000, 1000, f, f, 4, 1)
	if halvedE <= normalE {
		t.Errorf("兵種 4 的 pct 減半應讓自己損失變大：%d vs %d", halvedE, normalE)
	}
}

// 一面倒：門檻與全滅是自洽的，損失不會超過兵力。
func TestCasualtiesRoutAnnihilationThreshold(t *testing.T) {
	const force = 10000
	// 兵種 1（權重 1）：人力值 = 10000。攻擊 1000 × 10 = 10000 >= 10000 → 全滅
	lossE, _ := CasualtiesRout(1000, 0, force, force, 1, 1)
	if lossE != force {
		t.Errorf("剛好達門檻應全滅，得到 %d", lossE)
	}
	// 差一點就不全滅
	lossE, _ = CasualtiesRout(999, 0, force, force, 1, 1)
	if lossE >= force {
		t.Errorf("沒到門檻不該全滅，得到 %d", lossE)
	}
	if lossE != 9990 {
		t.Errorf("損失 = %d，應為 attack×10/權重 = 9990", lossE)
	}

	// 兵種 5（權重 100）撐得住得多：人力值 = 1,000,000
	lossE, _ = CasualtiesRout(1000, 0, force, force, 5, 1)
	if lossE != 100 {
		t.Errorf("兵種 5 的損失 = %d，應為 1000×10/100 = 100", lossE)
	}

	// 全部兵種掃一遍：損失永遠落在 [0, 兵力]
	for _, b := range []uint8{1, 4, 5, 6} {
		for _, atk := range []int{0, 1, 100, 10000, 1000000} {
			l := routSide(atk, force, b)
			if l < 0 || l > force {
				t.Errorf("兵種 %d 攻擊 %d 的損失 %d 超出 [0, %d]", b, atk, l, force)
			}
		}
	}
	// 未知兵種回 0。
	if got := routSide(99999, force, 3); got != 0 {
		t.Errorf("未知兵種的損失 = %d，應為 0", got)
	}
}
