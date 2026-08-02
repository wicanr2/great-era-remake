package main

import (
	"fmt"

	"github.com/wicanr2/great-era-remake/internal/game"
)

// gameSession 是從一份存檔完整建立的執行期快照。
//
// 只有 buildSession 全部成功後才能交給 app 替換；解析中途不得觸碰
// 目前遊戲的任何指標。這是桌面載入與未來 Android 背景恢復共用的邊界。
type gameSession struct {
	tbl             *game.ProvinceTable
	generals        []game.General
	factions        game.FactionTable
	leaders         game.FactionLeaders
	factionOf       game.FactionOfGeneral
	world           *game.AIWorld
	cmdBudget       *game.CommandBudget
	origSave        []byte
	current         game.ProvinceID
	playerCommander game.GeneralID
	year            uint16
	month           uint8
}

// buildSession 從 .DT1 建立一份完整快照。省份、將領、勢力、領袖、
// 勢力反查與停火任一區塊解析失敗，整份載入就失敗。
//
// ⚠️ .DT1 沒有自述期別；呼叫端必須先以 SAVE(N).DT1 的 N 對應 sc.Stage。
// 不可單靠內容猜期別，因為第一期的前 191 筆也能被當成第二期解析。
func buildSession(save []byte, sc game.Scenario, current game.ProvinceID, player game.GeneralID) (*gameSession, error) {
	if len(save) == 0 {
		return nil, fmt.Errorf("session: 存檔是空的")
	}
	tbl, err := game.ParseSaveProvinces(save)
	if err != nil {
		return nil, fmt.Errorf("session: 省份區：%w", err)
	}
	gens, err := game.ParseSaveGenerals(save, sc.Generals)
	if err != nil {
		return nil, fmt.Errorf("session: 將領區：%w", err)
	}
	factions, err := game.ParseFactionTable(save)
	if err != nil {
		return nil, fmt.Errorf("session: 勢力表：%w", err)
	}
	leaders, err := game.ParseFactionLeaders(save)
	if err != nil {
		return nil, fmt.Errorf("session: 勢力領袖表：%w", err)
	}
	factionOf, err := game.ParseFactionOfGeneral(save)
	if err != nil {
		return nil, fmt.Errorf("session: 勢力反查表：%w", err)
	}
	ceasefire, err := game.ParseCeasefireStates(save)
	if err != nil {
		return nil, fmt.Errorf("session: 停火狀態表：%w", err)
	}

	// 區塊 3、6、7 是同一批勢力的三種索引；能各自解析還不夠，
	// 必須互相對得上，否則就是已損壞或不屬於這個期別的存檔。
	for i, leader := range leaders {
		if leader == 0 {
			continue
		}
		if int(leader) > sc.Generals {
			return nil, fmt.Errorf("session: 勢力 %d 領袖 %d 超出第 %d 期將領數 %d",
				i+1, leader, sc.Stage, sc.Generals)
		}
		if !factions.HasFaction(i) || factions[i].Leader != leader {
			return nil, fmt.Errorf("session: 勢力 %d 的領袖表與勢力表不一致", i+1)
		}
		if got := factionOf.SlotOf(leader, leaders); got != uint8(i+1) {
			return nil, fmt.Errorf("session: 勢力 %d 領袖 %d 的反查值是 %d",
				i+1, leader, got)
		}
	}

	if player == 0 {
		if p, e := tbl.At(current); e == nil {
			player = p.Commander
		}
	}
	if player == 0 {
		return nil, fmt.Errorf("session: 無法確定玩家司令")
	}
	chosen := game.ProvinceID(0)
	if p, e := tbl.At(current); e == nil && int(current) <= sc.Provinces && p.Commander == player {
		chosen = current
	}
	if chosen == 0 {
		for id := game.ProvinceID(1); int(id) <= sc.Provinces; id++ {
			if p, e := tbl.At(id); e == nil && p.Commander == player {
				chosen = id
				break
			}
		}
	}
	if chosen == 0 {
		return nil, fmt.Errorf("session: 存檔中找不到玩家司令 %d 的合法省份", player)
	}

	world := buildWorld(tbl, gens, &factions, sc.Stage)
	world.CeasefireState = ceasefire
	raw := append([]byte(nil), save...)
	s := &gameSession{
		tbl: tbl, generals: gens, factions: factions, leaders: leaders, factionOf: factionOf,
		world: world, cmdBudget: game.NewCommandBudget(world), origSave: raw,
		current: chosen, playerCommander: player,
	}
	if tbl.Date != nil {
		s.year, s.month = tbl.Date.Year, tbl.Date.Month
	}
	return s, nil
}

// applySession 只接受已完整建好的快照，並一次替換所有會交叉引用的核心指標。
func (a *app) applySession(s *gameSession) {
	a.tbl, a.generals, a.world = s.tbl, s.generals, s.world
	a.cmdBudget, a.origSave = s.cmdBudget, s.origSave
	a.current, a.playerCommander = s.current, s.playerCommander
	a.year, a.month = s.year, s.month
	// 舊 session 的選擇器都可能含將領索引、省份指標或未提交數值，
	// 核心快照替換時必須一併清掉，不可讓舊 UI 之後寫回新遊戲。
	a.battle = nil
	a.transferTargets, a.transferSession = nil, nil
	a.transferMode, a.transferInput, a.transferCursor = 0, 0, 0
	a.transferAmounts, a.transferGood = [4]int{}, 0
	a.tradeImport, a.tradeGood, a.tradeAmount = false, 0, 0
	a.supplyTargets, a.supplyTarget = nil, 0
	a.supplyAmounts, a.supplyGood, a.supplyInput = [4]int{}, 0, 0
	a.recruitBranch, a.recruitLimit, a.recruitAmount = 0, 0, 0
	a.reorganization, a.reorganizeID, a.reorganizeInput = nil, 0, 0
	a.viewGenerals, a.viewIndex, a.viewProvince, a.viewInput, a.viewPage = nil, 0, 0, 0, 0
	a.bioPage, a.bioPages = 0, 0
	a.covertAction, a.covertInput = 0, 0
	a.autonomyTargets, a.autonomyInput, a.autonomySpent = nil, 0, false
	a.productionItem, a.productionInput, a.productionSpent = 0, 0, false
}

// loadSessionBytes 先在舊 app 之外建完新快照；任何錯誤都不會改變舊狀態。
func (a *app) loadSessionBytes(save []byte) error {
	sc, err := game.ScenarioByStage(uint8(a.stage))
	if err != nil {
		return err
	}
	s, err := buildSession(save, sc, a.current, a.playerCommander)
	if err != nil {
		return err
	}
	a.applySession(s)
	return nil
}
