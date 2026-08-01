package game

import "testing"

func TestDecideBattleBDecisive(t *testing.T) {
	// 第一方被壓到第二方的五分之一以下 → 必勝結算（對方贏）。
	d := DecideBattleB(BattleAIInput{SideStrength: 20, FoeStrength: 100})
	if d.Action != ActBDecisive {
		t.Errorf("戰力差五倍該走必勝結算，實際 %s", BattleActionName(d.Action))
	}
	// 差一點就不成立。
	d = DecideBattleB(BattleAIInput{SideStrength: 21, FoeStrength: 100})
	if d.Action != ActBTakeCity {
		t.Errorf("還沒到五分之一，該落到預設「打城市」，實際 %s", BattleActionName(d.Action))
	}
}

func TestDecideBattleADecisive(t *testing.T) {
	d := DecideBattleA(BattleAIInput{SideStrength: 20, FoeStrength: 100})
	if d.Action != ActADecisive {
		t.Errorf("該走必勝結算，實際 %s", BattleActionName(d.Action))
	}
}

func TestDecideBattleAEvenTriggersDecapitate(t *testing.T) {
	// sub_3A8F7 要 byte_6FFCA & 4，且條件**不成立**才出手
	//（不成立 = 敵方 < 我方的 2/3 = 我方領先）。
	//
	// ⭐ §53：這一步是三向分流，不是單一行動。
	in := BattleAIInput{SideStrength: 100, FoeStrength: 50, EnableLastSteps: true}

	// 有攻方單位逼近城市 → 18（挑最弱的圍城者），優先度最高。
	near := in
	near.AttackerNearCity = true
	near.FoeFewerThanCities = true // 就算這個也成立，18 仍然優先
	if got := DecideBattleA(near); got.Action != ActAWeakest {
		t.Errorf("有敵軍逼近城市該挑最弱的，實際 %s", BattleActionName(got.Action))
	}

	// 沒人逼近、而且攻方單位比城市少 → 14（只留一個守城，大膽出擊）。
	few := in
	few.FoeFewerThanCities = true
	if got := DecideBattleA(few); got.Action != ActADecapitateKeepOne {
		t.Errorf("敵軍比城市少該只留一個守城，實際 %s", BattleActionName(got.Action))
	}

	// 兩個都不成立 → 15（駐守的都留，保守）。
	if got := DecideBattleA(in); got.Action != ActADecapitateKeepAll {
		t.Errorf("敵軍不比城市少該把駐守的都留下，實際 %s", BattleActionName(got.Action))
	}
	// 旗標沒開就跳過這一步。
	in.EnableLastSteps = false
	if got := DecideBattleA(in); got.Action != ActADefault {
		t.Errorf("旗標沒開該落到預設，實際 %s", BattleActionName(got.Action))
	}
	// 敵方 ≥ 我方 2/3（勢均或落後）→ 條件成立 → 不走這一步。
	in = BattleAIInput{SideStrength: 100, FoeStrength: 67, EnableLastSteps: true}
	if got := DecideBattleA(in); got.Action != ActADefault {
		t.Errorf("敵方有我方 2/3 時不該斬首，實際 %s", BattleActionName(got.Action))
	}
}

func TestBattleActionValuesMatchOriginal(t *testing.T) {
	// 兩條鏈的值域不重疊，是原版共用同一張分派表的前提。
	a := []BattleAction{ActADecisive, ActAReset, ActADefault, ActADecapitateKeepOne,
		ActADecapitateKeepAll, ActAStandbyOnly, ActARecompute, ActAWeakest, ActAEngageAll}
	b := []BattleAction{ActBDecisive, ActBDeploy, ActBTakeCity, ActBStrikeForce}
	for _, x := range a {
		if x < 11 || x > 19 {
			t.Errorf("分支 A 的值要落在 11–19，實際 %d", x)
		}
		for _, y := range b {
			if x == y {
				t.Errorf("兩條鏈的值不得重疊：%d", x)
			}
		}
	}
	for _, y := range b {
		if y < 1 || y > 4 {
			t.Errorf("分支 B 的值要落在 1–4，實際 %d", y)
		}
	}
	// 13 種行動每一種都要有名字，不能落到「未知行動」。
	for _, x := range append(a, b...) {
		if BattleActionName(x) == "未知行動" {
			t.Errorf("行動 %d 沒有名字", x)
		}
	}
}

// ⭐ 這條測試換過一次寫法。舊版只在數 `UndecidedBattleSteps` 的長度，
// 並且自己留了話：「清單空了：如果真的全部補完，這個測試該改成驗證
// 決策鏈的完整行為」。2026-08-02 `docs/re/31` §56 補完最後一筆，照做。
//
// 現在它驗兩件事：清單真的空了，而且**每一個行動都走得到**。
func TestEveryChainStepIsReachable(t *testing.T) {
	if n := len(UndecidedBattleSteps); n != 0 {
		t.Fatalf("未讀步驟又出現 %d 筆：%v —— 決策鏈退化了，先補 docs/re/31",
			n, UndecidedBattleSteps)
	}

	// 分支 A 的六個步驟，各給一組讓它走到的輸入。
	weak := BattleAIInput{SideStrength: 20, FoeStrength: 100} // 我方被壓到 1/5
	lead := BattleAIInput{SideStrength: 100, FoeStrength: 50, EnableLastSteps: true}
	for _, c := range []struct {
		name string
		in   BattleAIInput
		want BattleAction
	}{
		{"11 必勝結算", weak, ActADecisive},
		{"12 推倒重來", BattleAIInput{SideStrength: 100, FoeStrength: 100,
			RatioGateSelf: true, RatioGateFoe: true}, ActAReset},
		{"19 全面接戰", BattleAIInput{SideStrength: 100, FoeStrength: 100,
			RatioGateSelf: true}, ActAEngageAll},
		{"18 挑最弱", withA(lead, func(i *BattleAIInput) { i.AttackerNearCity = true }),
			ActAWeakest},
		{"14 斬首只留一個", withA(lead, func(i *BattleAIInput) { i.FoeFewerThanCities = true }),
			ActADecapitateKeepOne},
		{"15 斬首都留", lead, ActADecapitateKeepAll},
		{"17 重算全軍", BattleAIInput{SideStrength: 100, FoeStrength: 100,
			EnableLastSteps: true, FoeLeaderOnField: true, Sub53619: true}, ActARecompute},
		{"16 只處理待命", BattleAIInput{SideStrength: 100, FoeStrength: 100,
			EnableLastSteps: true, FoeLeaderOnField: true}, ActAStandbyOnly},
		{"13 預設分流", BattleAIInput{SideStrength: 100, FoeStrength: 100}, ActADefault},
	} {
		if got := DecideBattleA(c.in); got.Action != c.want {
			t.Errorf("分支 A %s：實得 %s（步驟 %q）",
				c.name, BattleActionName(got.Action), got.Step)
		}
	}

	// 分支 B 的四個行動。
	for _, c := range []struct {
		name string
		in   BattleAIInput
		want BattleAction
	}{
		{"2 佈防", BattleAIInput{SideStrength: 100, FoeStrength: 100,
			RatioGateSelf: true, RatioGateFoe: true, DeployGateOpen: true}, ActBDeploy},
		{"1 必勝結算", BattleAIInput{SideStrength: 20, FoeStrength: 100}, ActBDecisive},
		{"3 打城市（沒後援）", BattleAIInput{SideStrength: 20, FoeStrength: 100,
			Sub53619: true}, ActBTakeCity},
		{"4 打敵方主力周邊", BattleAIInput{SideStrength: 100, FoeStrength: 100,
			EnableLastSteps: true, FoeLeaderOnField: true}, ActBStrikeForce},
	} {
		if got := DecideBattleB(c.in); got.Action != c.want {
			t.Errorf("分支 B %s：實得 %s（步驟 %q）",
				c.name, BattleActionName(got.Action), got.Step)
		}
	}
}

func withA(in BattleAIInput, f func(*BattleAIInput)) BattleAIInput {
	f(&in)
	return in
}

func TestDecideBattleARatioGates(t *testing.T) {
	// §42：sub_3A885 與 sub_3A8C8 共用同一組比率門檻，只差看幾方。
	base := BattleAIInput{SideStrength: 100, FoeStrength: 90} // 不觸發必勝門檻

	// 兩方都成立 → 12（推倒重來）
	in := base
	in.RatioGateSelf, in.RatioGateFoe = true, true
	if got := DecideBattleA(in); got.Action != ActAReset {
		t.Errorf("兩方都成立該走推倒重來，實際 %s", BattleActionName(got.Action))
	}

	// 只有我方成立 → 19（全面接戰）
	in = base
	in.RatioGateSelf = true
	if got := DecideBattleA(in); got.Action != ActAEngageAll {
		t.Errorf("只有我方成立該走全面接戰，實際 %s", BattleActionName(got.Action))
	}

	// 只有對方成立 → 不是這兩步的事，落到後面
	in = base
	in.RatioGateFoe = true
	if got := DecideBattleA(in); got.Action == ActAReset || got.Action == ActAEngageAll {
		t.Errorf("只有對方成立不該走這兩步，實際 %s", BattleActionName(got.Action))
	}

	// ⚠️ 順序要對：必勝門檻排在比率門檻前面，成立時先攔下。
	in = BattleAIInput{SideStrength: 20, FoeStrength: 100,
		RatioGateSelf: true, RatioGateFoe: true}
	if got := DecideBattleA(in); got.Action != ActADecisive {
		t.Errorf("必勝門檻該先攔下，實際 %s", BattleActionName(got.Action))
	}
}

func TestDecideBattleBDeployAndStrike(t *testing.T) {
	base := BattleAIInput{SideStrength: 100, FoeStrength: 90} // 不觸發必勝門檻

	// §43：值 2（佈防）與分支 A 值 12 條件相同，只多 DeployGateOpen。
	in := base
	in.RatioGateSelf, in.RatioGateFoe, in.DeployGateOpen = true, true, true
	if got := DecideBattleB(in); got.Action != ActBDeploy {
		t.Errorf("兩方比率門檻 + 閘門開 該走佈防，實際 %s", BattleActionName(got.Action))
	}
	// 閘門關就跳過這一步。
	in.DeployGateOpen = false
	if got := DecideBattleB(in); got.Action == ActBDeploy {
		t.Error("word_6493A 閘門關著時不該佈防")
	}

	// §42：值 4 要 EnableLastSteps 且 sub_56D49。
	in = base
	in.EnableLastSteps, in.FoeLeaderOnField = true, true
	if got := DecideBattleB(in); got.Action != ActBStrikeForce {
		t.Errorf("該走打敵方主力周邊，實際 %s", BattleActionName(got.Action))
	}
	in.EnableLastSteps = false
	if got := DecideBattleB(in); got.Action != ActBTakeCity {
		t.Errorf("旗標沒開該落到預設打城市，實際 %s", BattleActionName(got.Action))
	}
}

func TestDecideBattleAStep5(t *testing.T) {
	// §43：sub_3A94E 用 sub_56D49 當前置、sub_53619 決定 16 還是 17。
	// 要避開第四步（sub_3A8F7），所以讓「敵方 ≥ 我方 2/3」成立。
	base := BattleAIInput{SideStrength: 100, FoeStrength: 80, EnableLastSteps: true}

	in := base
	in.FoeLeaderOnField, in.Sub53619 = true, true
	if got := DecideBattleA(in); got.Action != ActARecompute {
		t.Errorf("sub_53619 為真該走重算全軍，實際 %s", BattleActionName(got.Action))
	}
	in.Sub53619 = false
	if got := DecideBattleA(in); got.Action != ActAStandbyOnly {
		t.Errorf("sub_53619 為假該走只處理待命，實際 %s", BattleActionName(got.Action))
	}
	// 前置不成立就落到預設。
	in.FoeLeaderOnField = false
	if got := DecideBattleA(in); got.Action != ActADefault {
		t.Errorf("前置不成立該落到預設，實際 %s", BattleActionName(got.Action))
	}
}

func TestSub53619GatesDecisiveResolution(t *testing.T) {
	// §45：戰力差五倍**還不夠**，還要 sub_53619 == 0 才准直接判勝負。
	crushed := BattleAIInput{SideStrength: 20, FoeStrength: 100}

	// 分支 B：Sub53619 為真 → 改打城市，不是必勝結算。
	in := crushed
	in.Sub53619 = true
	if got := DecideBattleB(in); got.Action != ActBTakeCity {
		t.Errorf("sub_53619 非 0 時該改打城市，實際 %s", BattleActionName(got.Action))
	}
	in.Sub53619 = false
	if got := DecideBattleB(in); got.Action != ActBDecisive {
		t.Errorf("sub_53619 為 0 才必勝結算，實際 %s", BattleActionName(got.Action))
	}

	// 分支 A：Sub53619 為真時看守方領袖在不在場。
	in = crushed
	in.Sub53619, in.FoeLeaderOnField = true, false
	if got := DecideBattleA(in); got.Action != ActAReset {
		t.Errorf("守方領袖不在場該推倒重來，實際 %s", BattleActionName(got.Action))
	}
	in.FoeLeaderOnField = true
	if got := DecideBattleA(in); got.Action != ActARecompute {
		t.Errorf("領袖在場恆走重算全軍（值 16 是死碼），實際 %s", BattleActionName(got.Action))
	}
	in.Sub53619 = false
	if got := DecideBattleA(in); got.Action != ActADecisive {
		t.Errorf("sub_53619 為 0 才必勝結算，實際 %s", BattleActionName(got.Action))
	}
}

func TestValue16IsDeadCodeInStep1(t *testing.T) {
	// ⛔ §45：sub_3A817 裡的值 16 走不到——外層已確定 sub_53619 非 0，
	// 內層又問一次同一個無副作用的函式，結果必然相同。
	// 這個測試釘住「照抄死碼」這個決定：只要第一步還會回值 16，就是實作偏離原版。
	for _, leader := range []bool{true, false} {
		in := BattleAIInput{SideStrength: 20, FoeStrength: 100,
			Sub53619: true, FoeLeaderOnField: leader}
		if got := DecideBattleA(in); got.Action == ActAStandbyOnly {
			t.Errorf("第一步不該回值 16（原版那條路是死碼），FoeLeaderOnField=%v", leader)
		}
	}
	// 但值 16 本身不是死碼——第五步走得到。
	in := BattleAIInput{SideStrength: 100, FoeStrength: 80,
		EnableLastSteps: true, FoeLeaderOnField: true, Sub53619: false}
	if got := DecideBattleA(in); got.Action != ActAStandbyOnly {
		t.Errorf("第五步該走得到值 16，實際 %s", BattleActionName(got.Action))
	}
}
