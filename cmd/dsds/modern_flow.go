package main

import (
	"fmt"
	"strconv"

	"github.com/wicanr2/great-era-remake/internal/game"
	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	uilayout "github.com/wicanr2/great-era-remake/internal/ui/layout"
	"github.com/wicanr2/great-era-remake/internal/ui/render"
)

// strategyCommand* 是十五項實際政略卡的選單序號。Modern 流程頁只能引用
// 這組名稱來取得標題，避免把畫面 state 的歷史分類順序錯當成選單序號。
const (
	strategyCommandTransfer  = 1
	strategyCommandSupply    = 3
	strategyCommandRecruit   = 5
	strategyCommandView      = 6
	strategyCommandDevelop   = 7
	strategyCommandDiplomacy = 9
	strategyCommandCeasefire = 10
	strategyCommandCovert    = 11
	strategyCommandTrade     = 12
	strategyCommandTrain     = 13
)

func (a *app) modernNavigationLabels() (string, string, string, error) {
	back, err := a.wordingText("common.back")
	if err != nil {
		return "", "", "", err
	}
	previous, err := a.wordingText("common.previous")
	if err != nil {
		return "", "", "", err
	}
	next, err := a.wordingText("common.next")
	if err != nil {
		return "", "", "", err
	}
	return back, previous, next, nil
}

// modernFlowData 把既有流程 screen state 整理成高解析 Modern 的純資料模型。
// 它只讀 app state；選項仍以 Selection／Confirm／Digit Action 回到原有 Update
// 分支，避免為了換版面複製任何規則。
func (a *app) modernFlowData() (render.ModernFlowSurfaceData, error) {
	if a == nil || a.wording == nil {
		return render.ModernFlowSurfaceData{}, fmt.Errorf("Modern 流程頁缺少 wording catalog")
	}
	d := render.ModernFlowSurfaceData{
		Fonts: a.eten, Style: a.uiStyle(), Selected: -1,
		Back: "", Previous: "", Next: "",
	}
	var err error
	if d.Back, err = a.wordingText("common.back"); err != nil {
		return d, err
	}
	if a.screen == screenViewGeneral {
		if _, ok := a.currentBiography(); ok {
			if d.Next, err = a.wordingText("biography.page"); err != nil {
				return d, err
			}
		}
	}
	if a.messages != nil && a.messages.Active() {
		d.Hint = a.messages.Current()
	}
	text := func(key string) string {
		value, textErr := a.wordingText(key)
		if textErr != nil && err == nil {
			err = textErr
		}
		return value
	}
	commandTitle := func(number int) string {
		return text(fmt.Sprintf("command.%02d", number))
	}
	provinceOptions := func(ids []game.ProvinceID) []string {
		out := make([]string, 0, len(ids))
		for _, id := range ids {
			out = append(out, fmt.Sprintf("%02d　%s", id, a.provinceName(id)))
		}
		return out
	}
	generalOptions := func(ids []game.GeneralID) []string {
		out := make([]string, 0, len(ids))
		for _, id := range ids {
			out = append(out, fmt.Sprintf("%02d　%s", id, a.generalDisplayName(id)))
		}
		return out
	}
	input := func(title, prompt, value string) {
		d.Title, d.Prompt, d.Input, d.InputMode = title, prompt, value, true
	}
	confirm := func(title, prompt string) {
		d.Title, d.Prompt, d.Options = title, prompt, []string{text("common.confirm"), text("common.cancel")}
	}

	switch a.screen {
	case screenDevelop:
		d.Title, d.Options = commandTitle(strategyCommandDevelop), []string{text("develop.reclaim"), text("develop.arsenal"), text("develop.mine")}
	case screenTransferMode:
		d.Title, d.Options = commandTitle(strategyCommandTransfer), []string{text("transfer.mode.partial"), text("transfer.mode.all")}
	case screenTransferTarget:
		d.Title, d.Prompt, d.Options = commandTitle(strategyCommandTransfer), text("transfer.target"), provinceOptions(a.transferTargets)
	case screenTransferSelection:
		d.Title, d.Prompt = commandTitle(strategyCommandTransfer), text("transfer.selection.confirm")
		if a.transferSession != nil {
			candidates := a.transferSession.Candidates()
			start := (a.transferCursor / 20) * 20
			end := start + 20
			if end > len(candidates) {
				end = len(candidates)
			}
			d.Options = generalOptions(candidates[start:end])
		}
	case screenTransferAmount:
		keys := []string{"transfer.resource.gold", "transfer.resource.food", "transfer.resource.ammo", "transfer.resource.fuel"}
		good := a.transferGood
		if good < 0 || good >= len(keys) {
			good = 0
		}
		input(commandTitle(strategyCommandTransfer), text(keys[good]), strconv.Itoa(int(a.transferInput)))
	case screenTradeMode:
		d.Title, d.Options = commandTitle(strategyCommandTrade), []string{text("trade.import"), text("trade.export")}
	case screenTradeGood:
		d.Title = commandTitle(strategyCommandTrade)
		goods := []string{"trade.food", "trade.ammo", "trade.fuel", "trade.coal", "trade.iron"}
		count := 5
		if a.tradeImport {
			count = 3
		}
		for _, key := range goods[:count] {
			d.Options = append(d.Options, text(key))
		}
	case screenTradeAmount:
		key := "trade.sell_amount"
		if a.tradeImport {
			key = "trade.buy_amount"
		}
		input(commandTitle(strategyCommandTrade), text(key), strconv.Itoa(int(a.tradeAmount)))
	case screenSupplyTarget:
		d.Title, d.Prompt, d.Options = commandTitle(strategyCommandSupply), text("supply.target"), provinceOptions(a.supplyTargets)
	case screenSupplyAmount:
		keys := []string{"supply.gold", "supply.food", "supply.ammo", "supply.fuel"}
		good := a.supplyGood
		if good < 0 || good >= len(keys) {
			good = 0
		}
		input(commandTitle(strategyCommandSupply), text(keys[good]), strconv.Itoa(int(a.supplyInput)))
	case screenRecruitAction:
		d.Title, d.Options = commandTitle(strategyCommandRecruit), []string{text("recruit.action"), text("recruit.reorganize")}
	case screenRecruitBranch, screenReorganizeBranch:
		d.Title = commandTitle(strategyCommandRecruit)
		for _, key := range []string{"recruit.infantry", "recruit.armour", "recruit.artillery", "recruit.cavalry"} {
			d.Options = append(d.Options, text(key))
		}
	case screenRecruitAmount:
		d.Hint = fmt.Sprintf("%s：%d", text("recruit.limit"), a.recruitLimit)
		input(commandTitle(strategyCommandRecruit), text("recruit.amount"), strconv.Itoa(int(a.recruitAmount)))
	case screenRecruitConfirm:
		confirm(commandTitle(strategyCommandRecruit), text("recruit.confirm"))
	case screenReorganizeTarget:
		d.Title, d.Prompt = commandTitle(strategyCommandRecruit), text("recruit.general")
		if a.reorganization != nil {
			d.Options = generalOptions(a.reorganization.Targets())
		}
	case screenReorganizeAmount:
		limit := 0
		if a.reorganization != nil {
			limit = a.reorganization.Limit(a.reorganizeID)
		}
		d.Hint = fmt.Sprintf("%s：%d", text("recruit.remaining"), limit)
		input(commandTitle(strategyCommandRecruit), text("recruit.force"), strconv.Itoa(int(a.reorganizeInput)))
	case screenTrainConfirm:
		confirm(commandTitle(strategyCommandTrain), text("train.confirm"))
	case screenCovertAction:
		d.Title, d.Options = commandTitle(strategyCommandCovert), []string{text("covert.guerrilla"), text("covert.student")}
	case screenCovertTarget:
		input(commandTitle(strategyCommandCovert), text("covert.student.target"), strconv.Itoa(int(a.covertInput)))
	case screenDiplomacy:
		d.Title, d.Options = commandTitle(strategyCommandDiplomacy), []string{text("diplomacy.loan"), text("diplomacy.aid"), text("diplomacy.repay")}
	case screenCeasefireTarget:
		input(commandTitle(strategyCommandCeasefire), text("ceasefire.prompt"), strconv.Itoa(int(a.ceasefireInput)))
	case screenLoanAmount:
		credit := 0
		if a.diplomacySlot >= 0 && a.diplomacySlot < len(a.ledger.Credit) {
			credit = int(a.ledger.Credit[a.diplomacySlot])
		}
		d.Hint = fmt.Sprintf("%s：%d", text("diplomacy.loan.credit"), credit)
		input(commandTitle(strategyCommandDiplomacy), text("diplomacy.loan.prompt"), strconv.Itoa(int(a.diplomacyInput)))
	case screenRepayAmount:
		debt, gold := 0, 0
		if a.diplomacySlot >= 0 && a.diplomacySlot < len(a.ledger.Debt) {
			debt = int(a.ledger.Debt[a.diplomacySlot])
		}
		if a.tbl != nil {
			if p, tableErr := a.tbl.At(a.current); tableErr == nil {
				gold = int(p.Gold)
			}
		}
		if debtText, formatErr := a.wordingFormat("diplomacy.repay.debt", debt, gold); formatErr == nil {
			d.Hint = debtText
		} else if err == nil {
			err = formatErr
		}
		input(commandTitle(strategyCommandDiplomacy), text("diplomacy.repay.prompt"), strconv.Itoa(int(a.diplomacyInput)))
	case screenViewMenu:
		d.Title, d.Options = commandTitle(strategyCommandView), []string{text("view.other"), text("view.owned"), text("view.generals"), text("view.province_names")}
	case screenViewProvinceSelect:
		d.Title, d.Prompt = commandTitle(strategyCommandView), text("view.select_prompt")
		ids := make([]game.ProvinceID, game.ProvinceCount)
		for i := range ids {
			ids[i] = game.ProvinceID(i + 1)
		}
		d.Options = provinceOptions(ids)
	case screenViewProvinceChoice:
		d.Title, d.Options = commandTitle(strategyCommandView), []string{text("view.choice.overview"), text("view.choice.generals")}
	case screenViewGenerals:
		d.Title, d.Prompt, d.Options = commandTitle(strategyCommandView), text("view.generals"), generalOptions(a.viewGenerals)
	case screenViewGeneral:
		if a.viewIndex < 0 || a.viewIndex >= len(a.viewGenerals) {
			return render.ModernFlowSurfaceData{}, fmt.Errorf("Modern 將領詳情索引失效：%d/%d", a.viewIndex, len(a.viewGenerals))
		}
		id := a.viewGenerals[a.viewIndex]
		gi := int(id) - 1
		if gi < 0 || gi >= len(a.generals) {
			return render.ModernFlowSurfaceData{}, fmt.Errorf("Modern 將領編號失效：%d", id)
		}
		g := a.generals[gi]
		attack := 0
		if a.world != nil && gi < len(a.world.Strengths) {
			attack = game.Strength(a.world.Strengths[gi], a.world.Opts)
		}
		d.Title = text("view.generals")
		d.Prompt = fmt.Sprintf("%s　%s", a.generalDisplayName(id), a.provinceName(g.Province))
		d.Options = []string{
			fmt.Sprintf("%s：%d", text("view.general.lead"), g.AbilityA),
			fmt.Sprintf("%s：%d", text("view.general.loyalty"), g.AbilityB),
			fmt.Sprintf("%s：%d", text("view.general.politics"), g.AbilityC),
			fmt.Sprintf("%s：%d", text("view.general.experience"), g.Experience),
			fmt.Sprintf("%s：%d", text("view.general.branch"), g.Branch),
			fmt.Sprintf("%s：%d", text("view.general.force"), g.Force),
			fmt.Sprintf("%s：%d", text("view.general.attack"), attack),
			fmt.Sprintf("%s：%d", text("view.general.armed"), g.F20),
			fmt.Sprintf("%s：%d", text("view.general.skill"), g.F19),
			fmt.Sprintf("%s：%d", text("view.general.stamina"), g.Stamina),
			fmt.Sprintf("%s：%d", text("view.general.morale"), g.F30),
		}
	case screenViewProvince:
		d.Title = text("view.choice.overview")
		if a.tbl != nil {
			if p, tableErr := a.tbl.At(a.viewProvince); tableErr == nil {
				d.Prompt = a.provinceName(a.viewProvince)
				d.Options = []string{
					fmt.Sprintf("%s：%d", text("panel.gold"), p.Gold),
					fmt.Sprintf("%s：%d", text("panel.food"), p.Food),
					fmt.Sprintf("%s：%d", text("panel.ammo"), p.Ammo),
					fmt.Sprintf("%s：%d", text("panel.fuel"), p.Fuel),
					fmt.Sprintf("%s：%d", text("panel.population"), p.PopulationWan()),
				}
			}
		}
	case screenViewOverview:
		d.Title = text("view.owned.title")
		if a.world != nil {
			d.Options = provinceOptions(a.world.OwnedProvinces(a.current))
		}
	case screenViewProvinceNames:
		d.Title = text("view.names.title")
		ids := make([]game.ProvinceID, game.ProvinceCount)
		for i := range ids {
			ids[i] = game.ProvinceID(i + 1)
		}
		d.Options = provinceOptions(ids)
	case screenPolicy:
		d.Title, d.Options = text("policy.title"), []string{text("policy.autonomy"), text("policy.production")}
	case screenAutonomy:
		d.Title, d.Prompt, d.Options = text("autonomy.title"), text("autonomy.prompt"), provinceOptions(a.autonomyTargets)
	case screenProduction:
		if a.productionItem == 0 {
			d.Title, d.Prompt = text("production.title"), text("production.select")
			for _, key := range []string{"production.gold", "production.iron", "production.coal", "production.food"} {
				d.Options = append(d.Options, text(key))
			}
		} else {
			d.Title, d.Prompt = text("production.title"), text("production.value")
			d.Input, d.InputMode = strconv.Itoa(int(a.productionInput)), true
		}
	case screenTransferConfirm, screenSaveConfirm, screenLoadConfirm, screenQuit:
		prompt := text("common.confirm")
		switch a.screen {
		case screenSaveConfirm:
			prompt = text("other.save.confirm")
		case screenLoadConfirm:
			prompt = text("other.load.confirm")
		}
		confirm(text("other.title"), prompt)
	case screenMessageTime:
		input(text("other.title"), text("other.message_time.prompt"), strconv.Itoa(int(a.messageTimeInput)))
	default:
		d.Title = text("other.title")
		if a.msg != "" {
			d.Prompt = a.msg
		}
	}
	if err != nil {
		return render.ModernFlowSurfaceData{}, err
	}
	return d, nil
}

// modernFlowTargets 與 DrawModernFlowSurface 共用 ModernFlowLayout。清單卡一律
// 保持既有 Selection(n)；確認頁與數字鍵盤才換成明確的 Confirm／Digit Action。
func (a *app) modernFlowTargets() []actions.Target {
	data, err := a.modernFlowData()
	if err != nil {
		return nil
	}
	count := len(data.Options)
	if data.InputMode {
		count = 0
	}
	l, err := uilayout.NewModernFlowLayout(uilayout.ModernDesignSurface(), count)
	if err != nil {
		return nil
	}
	if data.InputMode {
		out := make([]actions.Target, 0, len(l.Keypad))
		for i, key := range l.Keypad {
			action := actions.DeleteDigit
			switch {
			case i < 9:
				action = actions.Digit(i + 1)
			case i == 10:
				action = actions.Digit0
			case i == 11:
				action = actions.Submit
			}
			out = append(out, actions.Target{Action: action, Rect: actions.Rect{X: key.X, Y: key.Y, W: key.W, H: key.H}})
		}
		return out
	}
	if a.screen == screenTransferConfirm || a.screen == screenRecruitConfirm || a.screen == screenTrainConfirm ||
		a.screen == screenSaveConfirm || a.screen == screenLoadConfirm || a.screen == screenQuit {
		if len(l.Cards) < 2 {
			return nil
		}
		return []actions.Target{
			{Action: actions.Confirm, Rect: actions.Rect{X: l.Cards[0].X, Y: l.Cards[0].Y, W: l.Cards[0].W, H: l.Cards[0].H}},
			{Action: actions.Cancel, Rect: actions.Rect{X: l.Cards[1].X, Y: l.Cards[1].Y, W: l.Cards[1].W, H: l.Cards[1].H}},
		}
	}
	// 這些頁面的卡片是只讀資料，不公告假的 Selection；真正可操作的返回鍵
	// 由 interactiveTargets 另外從 navigationTargets 提供。
	if a.screen == screenViewGeneral || a.screen == screenViewProvince ||
		a.screen == screenViewOverview || a.screen == screenViewProvinceNames {
		return nil
	}
	out := make([]actions.Target, 0, len(l.Cards))
	for i, card := range l.Cards {
		out = append(out, actions.Target{Action: actions.Selection(i + 1), Rect: actions.Rect{X: card.X, Y: card.Y, W: card.W, H: card.H}})
	}
	return out
}
