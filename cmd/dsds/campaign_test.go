package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/wicanr2/great-era-remake/internal/game"
)

func campaignTestRead(name string) ([]byte, error) {
	return os.ReadFile(filepath.Join("../../workplace/orig/game", name))
}

func campaignTestMap(t *testing.T) *game.Map {
	t.Helper()
	w, e := campaignTestRead("WARPOS.DAT")
	if e != nil {
		t.Fatal(e)
	}
	r, e := campaignTestRead("TERNAME.DAT")
	if e != nil {
		t.Fatal(e)
	}
	n, e := campaignTestRead("NWMAP.DAT")
	if e != nil {
		t.Fatal(e)
	}
	m, e := game.LoadMap(w, r, n)
	if e != nil {
		t.Fatal(e)
	}
	return m
}

func TestCampaignAllFactionsCommandsAIRestart(t *testing.T) {
	m := campaignTestMap(t)
	count := 0
	for _, sc := range game.Scenarios {
		b, _ := campaignTestRead(fmt.Sprintf("TOWN(%d).DAT", sc.Stage))
		town, err := game.ParseTownFile(b)
		if err != nil {
			t.Fatal(err)
		}
		b, _ = campaignTestRead(fmt.Sprintf("MAN(%d).DAT", sc.Stage))
		gens, err := game.ParseGenerals(b, sc.Generals)
		if err != nil {
			t.Fatal(err)
		}
		choices, err := game.ScenarioFactionChoices(sc.Stage, town, gens)
		if err != nil {
			t.Fatal(err)
		}
		for _, f := range choices {
			t.Run(fmt.Sprintf("stage%d-leader%d", sc.Stage, f.Leader), func(t *testing.T) {
				read := func(name string) ([]byte, error) {
					if name != "TOWN(1).DAT" && name != "TOWN(2).DAT" && name != "TOWN(3).DAT" && name != "MAN(1).DAT" && name != "MAN(2).DAT" && name != "MAN(3).DAT" && name != "MAN115" && name != "MAN215" && name != "MAN315" && name != "FAN(1).15" && name != "FAN(2).15" && name != "FAN(3).15" {
						t.Fatalf("新局借用了非劇本來源 %s", name)
					}
					return campaignTestRead(name)
				}
				a := &app{m: m, dataRead: read, initialSeed: 1, savePath: filepath.Join(t.TempDir(), "campaign.gesave")}
				if err := a.startCampaign(sc.Stage, f.Leader); err != nil {
					t.Fatal(err)
				}
				if a.current != f.Home || a.playerCommander != f.Leader || a.origSave != nil || a.year != uint16(sc.Year) || a.month != sc.Month {
					t.Fatal("新局身份、日期或來源錯誤")
				}
				if !reflect.DeepEqual(a.tbl.Province, town.Province) || !reflect.DeepEqual(a.generals, gens) {
					t.Fatal("新局改動原始世界資料")
				}
				if a.factionLeaders.Count() != len(choices) {
					t.Fatal("勢力遺失")
				}
				for p := game.ProvinceID(1); int(p) <= sc.Provinces; p++ {
					prov, _ := a.tbl.At(p)
					if a.ownsProvince(p) != (prov.Commander == f.Leader) {
						t.Fatal("敵省取得控制權")
					}
				}
				before := a.cmdBudget.Remaining(a.current)
				prov, _ := a.tbl.At(a.current)
				if !prov.Taxed() {
					a.withBudget(a.current, a.execTax)
					if a.cmdBudget.Remaining(a.current) != before-1 {
						t.Fatal("合法指令未扣預算")
					}
				}
				if err := a.autosave(); err != nil {
					t.Fatal(err)
				}
				data, err := os.ReadFile(a.savePath)
				if err != nil {
					t.Fatal(err)
				}
				bapp := &app{m: m, dataRead: read, savePath: a.savePath}
				if err := bapp.loadCampaign(data); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(a.campaignState(), bapp.campaignState()) {
					t.Fatal("存讀遺失世界、亂數、身份或剩餘指令")
				}
				// 同一保存點的一個月：AI、跨年期別、亂數及結果須可重播。
				a.endTurn()
				bapp.endTurn()
				if !reflect.DeepEqual(a.campaignState(), bapp.campaignState()) {
					t.Fatal("重啟後 AI 回合結果不同")
				}
				if err := bapp.autosave(); err != nil {
					t.Fatal(err)
				}
				count++
			})
		}
	}
	if count != 14 {
		t.Fatalf("未測全部勢力：%d", count)
	}
}

func TestCampaignBattleSaveDoesNotRedeploy(t *testing.T) {
	a := &app{m: campaignTestMap(t), dataRead: campaignTestRead, initialSeed: 1, savePath: filepath.Join(t.TempDir(), "battle.gesave")}
	if err := a.startCampaign(1, 58); err != nil {
		t.Fatal(err)
	}
	target := a.tbl.FirstAttackable(a.current)
	if target == 0 {
		t.Fatal("實際劇本沒有鄰敵")
	}
	if err := a.startBattle(target, a.current); err != nil {
		t.Fatal(err)
	}
	a.battle.sim.Attacker[0].Current--
	a.battle.turn = 2
	a.cmdBudget.Spend(a.current)
	before := a.campaignState()
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(a.savePath)
	if err != nil {
		t.Fatal(err)
	}
	b := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
	if err := b.loadCampaign(data); err != nil {
		t.Fatal(err)
	}
	if b.screen != screenBattle || !reflect.DeepEqual(before, b.campaignState()) || a.battle.sim.Occ != b.battle.sim.Occ {
		t.Fatal("保存重置、重新部署或遺失戰鬥")
	}
	for _, u := range b.battle.sim.Attacker {
		if b.battle.sim.Unit(u.General) != u {
			t.Fatal("未重建戰鬥 ID 索引")
		}
	}
	// 已結算標記也原樣保存，不在讀取時再套用世界效果。
	a.battle.finished = true
	a.battle.settled = true
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	data, _ = os.ReadFile(a.savePath)
	if err := b.loadCampaign(data); err != nil {
		t.Fatal(err)
	}
	if !b.battle.finished || !b.battle.settled || !reflect.DeepEqual(a.campaignState(), b.campaignState()) {
		t.Fatal("結算標記未保留")
	}
}

func TestCampaignImportUnknownBytesAndFailedLoad(t *testing.T) {
	legacy := readStage1Save(t)
	current, player := stage1Player(t, legacy)
	sc, _ := game.ScenarioByStage(1)
	s, err := buildSession(legacy, sc, current, player)
	if err != nil {
		t.Fatal(err)
	}
	a := &app{m: campaignTestMap(t), dataRead: campaignTestRead, savePath: filepath.Join(t.TempDir(), "import.gesave"), stage: 1, provinceLimit: 36, rng: game.NewRand(1), campaign: &campaignIdentity{Mode: "original", Stage: 1, Player: player, SourceSHA256: map[string]string{"legacy.dt1": fmt.Sprintf("%064x", 1)}}}
	a.applySession(s)
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(a.savePath)
	var dto campaignSnapshot
	raw, err := game.DecodeCampaignSave(data, &dto)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, legacy) {
		t.Fatal("匯入來源的未知 bytes 改動")
	}
	before := a.campaignState()
	oldTable := a.tbl
	oldRNG := a.rng.Seed()
	if err := a.loadCampaign(data[:len(data)/2]); err == nil {
		t.Fatal("毀損容器被接受")
	}
	if a.tbl != oldTable || a.rng.Seed() != oldRNG || !reflect.DeepEqual(before, a.campaignState()) {
		t.Fatal("載入失敗部分替換現行遊戲")
	}
	if err := a.loadCampaign(data); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(a.origSave, legacy) {
		t.Fatal("重啟遺失原始 DT1")
	}
}

func TestCampaignEnemyInspectionKeepsPlayerAndBlocksCommands(t *testing.T) {
	a := &app{m: campaignTestMap(t), dataRead: campaignTestRead, initialSeed: 1}
	b := &app{m: a.m, dataRead: campaignTestRead, initialSeed: 1}
	if err := a.startCampaign(1, 1); err != nil {
		t.Fatal(err)
	}
	if err := b.startCampaign(1, 1); err != nil {
		t.Fatal(err)
	}
	for p := game.ProvinceID(1); p <= 36; p++ {
		if !b.ownsProvince(p) {
			b.current = p
			break
		}
	}
	called := false
	b.withBudget(b.current, func() string { called = true; return "changed" })
	if called || len(b.cmdBudget.Snapshot()) != 0 {
		t.Fatal("查看敵省時執行了修改或消耗預算")
	}
	a.endTurn()
	b.endTurn()
	left, right := a.campaignState(), b.campaignState()
	right.Current = left.Current
	if !reflect.DeepEqual(left, right) {
		t.Fatal("查看敵省交換了玩家與 AI 身份")
	}
}

func TestCampaignReorganizationPoolSurvivesQuitAndCompletion(t *testing.T) {
	a := &app{m: campaignTestMap(t), dataRead: campaignTestRead, initialSeed: 1, savePath: filepath.Join(t.TempDir(), "reorganize.gesave")}
	if err := a.startCampaign(1, 1); err != nil {
		t.Fatal(err)
	}
	r, err := a.world.BeginReorganization(a.current, game.BranchInfantry)
	if err != nil {
		t.Fatal(err)
	}
	a.reorganization = r
	id := r.Targets()[0]
	force := int(a.world.Strengths[int(id)-1].Force)
	if force == 0 || force > int(game.BranchFullStrength(game.BranchInfantry)) {
		t.Fatal("實際劇本的整編測試前提不成立")
	}
	budget := a.cmdBudget.Remaining(a.current)
	if err := r.Assign(id, 0); err != nil {
		t.Fatal(err)
	}
	a.syncReorganizationGenerals()
	if err := a.autosave(); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(a.savePath)
	b := &app{m: a.m, dataRead: campaignTestRead, savePath: a.savePath}
	if err := b.loadCampaign(data); err != nil {
		t.Fatal(err)
	}
	if b.screen != screenReorganizeTarget || b.reorganization.Remaining() != force || !reflect.DeepEqual(a.campaignState(), b.campaignState()) {
		t.Fatal("整編兵力池、武裝或候選在重啟後遺失")
	}
	if err := b.reorganization.Assign(id, force); err != nil {
		t.Fatal(err)
	}
	b.syncReorganizationGenerals()
	b.finishReorganization()
	if b.reorganization != nil || b.cmdBudget.Remaining(b.current) != budget-1 || int(b.generals[int(id)-1].Force) != force {
		t.Fatal("恢復後整編兵力不守恆或指令未正確扣除")
	}
	if err := b.autosave(); err != nil {
		t.Fatal(err)
	}
	good, _ := os.ReadFile(b.savePath)
	b.campaign.Mode = "unimplemented-mode"
	if err := b.autosave(); err == nil {
		t.Fatal("不合法模式被保存")
	}
	after, _ := os.ReadFile(b.savePath)
	if !bytes.Equal(good, after) {
		t.Fatal("保存驗證失敗覆蓋舊快照")
	}
}
