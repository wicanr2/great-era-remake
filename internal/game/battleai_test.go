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
	in := BattleAIInput{SideStrength: 100, FoeStrength: 50, EnableLastSteps: true}
	if got := DecideBattleA(in); got.Action != ActADecapitateKeepOne {
		t.Errorf("我方領先且啟用後段，該走斬首，實際 %s", BattleActionName(got.Action))
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

func TestUndecidedStepsAreTracked(t *testing.T) {
	// 這份清單是「這條鏈目前只有部分行為」的憑證。
	// 補完一支就從清單移除——測試在這裡是為了讓縮水這件事被看見。
	if len(UndecidedBattleSteps) == 0 {
		t.Skip("清單空了：如果真的全部補完，這個測試該改成驗證決策鏈的完整行為")
	}
	if len(UndecidedBattleSteps) != 8 {
		t.Logf("未讀步驟從 8 變成 %d —— 若是補完了，記得同步更新 docs/re/31",
			len(UndecidedBattleSteps))
	}
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
