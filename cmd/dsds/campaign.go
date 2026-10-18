package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"

	"github.com/wicanr2/great-era-remake/internal/assets"
	"github.com/wicanr2/great-era-remake/internal/game"
)

type campaignIdentity struct {
	Mode         string
	Stage        uint8
	Player       game.GeneralID
	SourceSHA256 map[string]string
}

type campaignSnapshot struct {
	Identity       campaignIdentity
	State          game.DT1State
	Current        game.ProvinceID
	Year           uint16
	Month          uint8
	RNG            uint32
	Budget         map[game.ProvinceID]int
	Units          []game.CombatUnit
	Strengths      []game.StrengthInput
	FactionBonus   uint8
	EnableExtra    bool
	StrategyTurn   *strategyTurn `json:",omitempty"`
	Battle         *campaignBattle
	Reorganization *game.ReorganizationSnapshot
	DT2, MemWar    []byte
}

type campaignExpansion struct {
	Base   campaignSnapshot
	Phase  string
	Armory *game.Armory
	View   armoryViewSnapshot
}

// v3 separates the live rules switch from the campaign's historical identity.
// The armory may remain present while disabled so no equipment is discarded.
type campaignOptions struct {
	Base   campaignSnapshot
	Rules  game.MilitaryRules
	Armory *game.Armory
	View   armoryViewSnapshot
}

type campaignBattle struct {
	Paused                         bool `json:",omitempty"`
	HumanDefender                  bool `json:",omitempty"`
	JapaneseBlitz                  bool `json:",omitempty"`
	ExternalInvasion               bool `json:",omitempty"`
	At, From                       game.ProvinceID
	AtCommander                    game.GeneralID
	Attacker, Defender             []*game.Combatant
	FactionBonus                   uint8
	Selected                       int
	Log                            string
	Finished                       bool
	Settlement                     game.BattleSettlement
	Settled, ForceDirty            bool
	Mode                           battleMode
	RetreatInput                   uint32
	RetreatTargets                 []game.ProvinceID
	Turn, TurnCap                  int
	AILog                          string
	AIAction, AIMoves, AIFights    int
	Leader                         game.GeneralID
	SupplyAttacker, SupplyDefender game.BattleSupply
	Units                          []game.CombatUnit
	EnableLastSteps                bool
}

func newCampaignSession(read func(string) ([]byte, error), stage uint8, player game.GeneralID) (*gameSession, *campaignIdentity, error) {
	sc, err := game.ScenarioByStage(stage)
	if err != nil {
		return nil, nil, err
	}
	source := make(map[string]string)
	load := func(name string) ([]byte, error) {
		b, e := read(name)
		if e == nil {
			source[name] = fmt.Sprintf("%x", sha256.Sum256(b))
		}
		return b, e
	}
	townBytes, err := load(fmt.Sprintf("TOWN(%d).DAT", stage))
	if err != nil {
		return nil, nil, err
	}
	manBytes, err := load(fmt.Sprintf("MAN(%d).DAT", stage))
	if err != nil {
		return nil, nil, err
	}
	town, err := game.ParseTownFile(townBytes)
	if err != nil {
		return nil, nil, err
	}
	gens, err := game.ParseGenerals(manBytes, sc.Generals)
	if err != nil {
		return nil, nil, err
	}
	state, home, err := game.NewCampaignState(stage, town, gens, player)
	if err != nil {
		return nil, nil, err
	}
	s := sessionFromState(state, sc, home, player)
	identity := &campaignIdentity{Mode: "original", Stage: stage, Player: player, SourceSHA256: source}
	return s, identity, nil
}

func sessionFromState(state game.DT1State, sc game.Scenario, current game.ProvinceID, player game.GeneralID) *gameSession {
	s := &gameSession{tbl: state.Provinces, generals: state.Generals, factions: state.Factions, leaders: state.FactionLeaders, factionOf: state.FactionOfGeneral, warRecords: state.WarRecords, majorPowerLeaders: state.MajorPowerLeaders, ledger: state.Ledger, current: current, playerCommander: player}
	s.world = buildWorld(s.tbl, s.generals, &s.factions, sc.Stage)
	s.world.CeasefireState = state.CeasefireStates
	s.cmdBudget = game.NewCommandBudget(s.world)
	if s.tbl.Date != nil {
		s.year, s.month = s.tbl.Date.Year, s.tbl.Date.Month
	}
	return s
}

func (a *app) campaignState() campaignSnapshot {
	s := campaignSnapshot{Identity: *a.campaign, Current: a.current, Year: a.year, Month: a.month, RNG: a.rng.Seed(), Budget: a.cmdBudget.Snapshot(), Units: a.world.Units, Strengths: a.world.Strengths, FactionBonus: a.world.Opts.FactionBonus, EnableExtra: a.world.EnableExtra, DT2: a.battleDT2, MemWar: a.battleMemWar}
	s.StrategyTurn = a.strategyTurn
	s.State = game.DT1State{Provinces: a.tbl, Generals: a.generals, Factions: a.factions, FactionLeaders: a.factionLeaders, FactionOfGeneral: a.factionOf, WarRecords: a.warRecords, CeasefireStates: a.world.CeasefireState, MajorPowerLeaders: a.majorPowerLeaders, Ledger: a.ledger}
	table := *a.tbl
	date := game.GameDate{Year: a.year, Month: a.month}
	if table.Date != nil {
		date.Unknown3 = table.Date.Unknown3
	}
	table.Date = &date
	s.State.Provinces = &table
	s.Reorganization = a.reorganization.Snapshot()
	if b := a.battle; b != nil && b.sim != nil {
		s.Battle = &campaignBattle{At: b.sim.At, From: b.sim.From, AtCommander: b.sim.AtCommander, Attacker: b.sim.Attacker, Defender: b.sim.Defender, FactionBonus: b.sim.Opts.FactionBonus, Selected: b.sel, Log: b.log, Finished: b.finished, Settlement: b.settlement, Settled: b.settled, ForceDirty: b.forceDirty, Mode: b.mode, RetreatInput: b.retreatInput, RetreatTargets: b.retreatTargets, Turn: b.turn, TurnCap: b.turnCap, AILog: b.aiLog, AIAction: b.aiAction, AIMoves: b.aiMoves, AIFights: b.aiFights, Leader: b.leader, SupplyAttacker: b.supAtk, SupplyDefender: b.supDef, Units: b.units, EnableLastSteps: b.enableLastSteps}
		s.Battle.Paused = b.paused
		s.Battle.HumanDefender = b.humanDefender
		s.Battle.JapaneseBlitz = b.japaneseBlitz
		s.Battle.ExternalInvasion = b.externalInvasion
	}
	return s
}

func (a *app) saveCampaign() error {
	if a.savePath == "" || a.world == nil || a.rng == nil || a.cmdBudget == nil {
		return fmt.Errorf("戰役保存缺少必要狀態或路徑")
	}
	// 不重跑戰後結算、不寫 DT2 旁檔，也不改匯入的原始 DT1。
	if a.armory != nil {
		if err := a.syncBattleForces(); err != nil {
			return err
		}
		if err := a.armory.Sync(a.world); err != nil {
			return err
		}
	}
	s := a.campaignState()
	v4, err := a.campaignV4State(s)
	if err != nil {
		return err
	}
	b, err := game.EncodeCampaignSaveVersion(v4, a.origSave, game.CampaignSaveCurrentVersion)
	if err != nil {
		return err
	}
	// 自己產生的快照也走語意驗證，避免寫出無法繼續的進度。
	if _, _, _, err = a.decodeCampaign(b); err != nil {
		return err
	}
	if err = os.MkdirAll(filepath.Dir(a.savePath), 0o755); err != nil {
		return err
	}
	return writeFileAtomic(a.savePath, b, 0o600)
}

func (a *app) decodeCampaign(data []byte) (*gameSession, *campaignSnapshot, *battleState, error) {
	var saved campaignSnapshot
	version, err := game.CampaignSaveVersion(data)
	if err != nil {
		return nil, nil, nil, err
	}
	var legacy []byte
	var expansion *campaignExpansion
	var options *campaignOptions
	var v4 *campaignV4
	if version == 1 {
		legacy, err = game.DecodeCampaignSave(data, &saved)
	} else if version == 2 {
		expansion = &campaignExpansion{}
		legacy, err = game.DecodeCampaignSaveVersion(data, expansion, 2)
		saved = expansion.Base
	} else if version == 3 {
		options = &campaignOptions{}
		legacy, err = game.DecodeCampaignSaveVersion(data, options, 3)
		saved = options.Base
	} else {
		v4 = &campaignV4{}
		legacy, err = game.DecodeCampaignSaveVersion(data, v4, 4)
		saved = v4.Base
		options = &campaignOptions{Base: saved, Rules: v4.Rules, Armory: v4.Armory, View: v4.View}
	}
	if err != nil {
		return nil, nil, nil, err
	}
	sc, err := game.ScenarioByStage(saved.Identity.Stage)
	if err != nil {
		return nil, nil, nil, err
	}
	bad := func() (*gameSession, *campaignSnapshot, *battleState, error) {
		return nil, nil, nil, fmt.Errorf("戰役保存的世界或身份資料不合法")
	}
	validMode := version == 1 && saved.Identity.Mode == "original" || version == 2 && saved.Identity.Mode == "unification-resistance" && !(saved.Identity.Stage > 1 && saved.Identity.Player == 107) || version >= 3 && (saved.Identity.Mode == "original" || saved.Identity.Mode == "unification-resistance")
	if !validMode || saved.Identity.Player == 0 || int(saved.Identity.Player) > sc.Generals || saved.State.Provinces == nil || len(saved.State.Generals) != sc.Generals || len(saved.Units) != sc.Generals || len(saved.Strengths) != sc.Generals || !saved.Current.Valid() || int(saved.Current) > sc.Provinces || saved.Year == 0 || saved.Month < 1 || saved.Month > 12 {
		return bad()
	}
	for _, hash := range saved.Identity.SourceSHA256 {
		decoded, err := hex.DecodeString(hash)
		if err != nil || len(decoded) != 32 {
			return bad()
		}
	}
	if len(saved.Identity.SourceSHA256) == 0 {
		return bad()
	}
	seen := map[game.GeneralID]bool{}
	var actorGenerals []game.General
	var actorUnits []game.CombatUnit
	var actorStrengths []game.StrengthInput
	var actorKeys map[game.GeneralID]string
	if v4 != nil {
		actorGenerals, actorUnits, actorStrengths, actorKeys, err = projectCampaignActors(v4, sc)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	validRuntimeID := func(id game.GeneralID) bool { return id > 0 && (int(id) <= sc.Generals || actorKeys[id] != "") }
	validFactionID := func(id game.GeneralID) bool {
		return int(id) <= sc.Generals || id == 300 && sc.Stage == 1 && v4 != nil && v4.Declaration != nil && v4.Declaration.Controller == 1
	}
	for i, id := range saved.State.FactionLeaders {
		if id == 0 {
			continue
		}
		if int(id) > sc.Generals || seen[id] || !saved.State.Factions.HasFaction(i) || saved.State.Factions[i].Leader != id || saved.State.FactionOfGeneral.SlotOf(id, saved.State.FactionLeaders) != uint8(i+1) {
			return bad()
		}
		seen[id] = true
	}
	for _, p := range saved.State.Provinces.Province[:sc.Provinces] {
		if !validFactionID(p.Commander) {
			return bad()
		}
	}
	for i, u := range saved.Units {
		if int(u.General) != i+1 || !validFactionID(u.Faction) || (!u.Cell.Valid() && u.Cell != game.NoCell) || (!u.NextCell.Valid() && u.NextCell != game.NoCell) || u.TargetUnit != 0 && !validRuntimeID(u.TargetUnit) {
			return bad()
		}
	}
	if len(legacy) > 0 {
		if _, err := game.ParseDT1(legacy, sc.Generals); err != nil {
			return nil, nil, nil, err
		}
	}
	s := sessionFromState(saved.State, sc, saved.Current, saved.Identity.Player)
	s.origSave = legacy
	s.year, s.month = saved.Year, saved.Month
	s.world.Units, s.world.Strengths = saved.Units, saved.Strengths
	if v4 != nil {
		s.generals, s.world.Units, s.world.Strengths = actorGenerals, actorUnits, actorStrengths
		s.campaignActors, s.declaration = actorKeys, v4.Declaration
	}
	s.world.Opts.FactionBonus, s.world.EnableExtra = saved.FactionBonus, saved.EnableExtra
	if options != nil {
		month := uint32(saved.Year)*12 + uint32(saved.Month)
		if !options.Rules.Valid(month) || options.Rules.Enabled && saved.Identity.Stage > 1 && saved.Identity.Player == 107 || options.Rules.Enabled && options.Armory == nil || !options.Rules.Enabled && options.View.Open {
			return bad()
		}
		s.militaryRules = &options.Rules
		if options.Armory == nil {
			if options.View != (armoryViewSnapshot{}) {
				return bad()
			}
		} else {
			expansion = &campaignExpansion{Base: saved, Phase: "preparation", Armory: options.Armory, View: options.View}
		}
	}
	if expansion != nil {
		if expansion.Phase != "preparation" || expansion.Armory == nil || expansion.Armory.Provinces != sc.Provinces || expansion.Armory.LastMonth != uint32(saved.Year)*12+uint32(saved.Month) || !validArmoryView(expansion.View, expansion.Armory, saved.Identity.Player) {
			return bad()
		}
		if err := expansion.Armory.Validate(s.world); err != nil {
			return nil, nil, nil, err
		}
		for i, input := range s.world.Strengths {
			u := expansion.Armory.Units[game.GeneralID(i+1)]
			if u == nil || u.Faction != s.world.Units[i].Faction || u.Branch != input.Branch || u.Force != input.Force {
				return bad()
			}
		}
		s.armory = expansion.Armory
		s.world.Opts.Armory = s.armory
		s.pendingArmoryView = &expansion.View
		if s.militaryRules == nil {
			s.militaryRules = &game.MilitaryRules{Enabled: true}
		}
	}
	if v4 != nil && (!validCampaignLanding(v4) || !validCampaignController(v4, sc)) {
		return bad()
	}
	runtimeScenario := sc
	runtimeScenario.Generals = len(s.world.Units)
	runtimeSaved := saved
	runtimeSaved.Units, runtimeSaved.Strengths = s.world.Units, s.world.Strengths
	if !validStrategyTurn(saved.StrategyTurn, &runtimeSaved, options, runtimeScenario) {
		return bad()
	}
	s.strategyTurn = saved.StrategyTurn
	if err := s.cmdBudget.Restore(saved.Budget, sc.Provinces); err != nil {
		return nil, nil, nil, err
	}
	if saved.Reorganization != nil {
		p, _ := s.tbl.At(saved.Current)
		if p.Commander != saved.Identity.Player || saved.Battle != nil {
			return bad()
		}
		s.pendingReorganization, err = s.world.RestoreReorganization(saved.Current, saved.Reorganization)
		if err != nil {
			return nil, nil, nil, err
		}
	}
	var battle *battleState
	if b := saved.Battle; b != nil {
		selectable := len(b.Attacker)
		if b.HumanDefender {
			if b.AtCommander != saved.Identity.Player {
				return bad()
			}
			province, err := saved.State.Provinces.At(b.At)
			if err != nil || !b.Settled && province.Commander != saved.Identity.Player {
				return bad()
			}
			if !b.Finished {
				candidates := game.DefenderRetreatTargets(saved.State.Provinces, b.At, saved.Identity.Player)
				if len(candidates) != len(b.RetreatTargets) {
					return bad()
				}
				for i, id := range candidates {
					if b.RetreatTargets[i] != id {
						return bad()
					}
				}
				if b.Mode == battleModeRetreat && len(candidates) == 0 {
					return bad()
				}
			} else if len(b.RetreatTargets) != 0 {
				return bad()
			}
			for _, unit := range b.Defender {
				if unit == nil || unit.Faction != saved.Identity.Player {
					return bad()
				}
			}
			for _, unit := range b.Attacker {
				if unit == nil || unit.Faction == 0 || unit.Faction == saved.Identity.Player {
					return bad()
				}
			}
			selectable = len(b.Defender)
		}
		external := v4 != nil && b.ExternalInvasion && b.From == game.ExternalInvasionSource
		if b.ExternalInvasion && !external || int(b.At) > sc.Provinces || int(b.From) > sc.Provinces && !external || b.Selected < 0 || b.Selected >= max(1, selectable) || b.Mode > battleModeInspect || b.Turn < 0 || b.TurnCap < 1 || b.TurnCap > 16 || len(b.Units) != runtimeScenario.Generals {
			return bad()
		}
		if v4 != nil {
			for i, unit := range b.Units {
				if unit != s.world.Units[i] {
					return bad()
				}
			}
		}
		for _, side := range [][]*game.Combatant{b.Attacker, b.Defender} {
			for _, unit := range side {
				if unit == nil || !validRuntimeID(unit.General) || unit.TargetUnit != 0 && !validRuntimeID(unit.TargetUnit) {
					return bad()
				}
			}
		}
		opt := s.world.Opts
		opt.FactionBonus = b.FactionBonus
		var sim *game.BattleSim
		if external {
			sim, err = game.RestoreExternalBattleSim(a.m, b.At, b.AtCommander, b.Attacker, b.Defender, opt, runtimeScenario.Generals)
		} else {
			sim, err = game.RestoreBattleSim(a.m, b.At, b.From, b.AtCommander, b.Attacker, b.Defender, opt, runtimeScenario.Generals)
		}
		if err != nil {
			return nil, nil, nil, err
		}
		battle = &battleState{sim: sim, sel: b.Selected, log: b.Log, finished: b.Finished, settlement: b.Settlement, settled: b.Settled, forceDirty: b.ForceDirty, mode: b.Mode, retreatInput: b.RetreatInput, retreatTargets: b.RetreatTargets, turn: b.Turn, turnCap: b.TurnCap, aiLog: b.AILog, aiAction: b.AIAction, aiMoves: b.AIMoves, aiFights: b.AIFights, leader: b.Leader, tbl: s.tbl, supAtk: b.SupplyAttacker, supDef: b.SupplyDefender, units: b.Units, enableLastSteps: b.EnableLastSteps}
		battle.paused = b.Paused
		battle.humanDefender = b.HumanDefender
		battle.japaneseBlitz = b.JapaneseBlitz
		battle.externalInvasion = b.ExternalInvasion
		// Runtime support queries share the live strategic units before saving.
		// Restore that relationship so withdrawal cannot leave a stale copy.
		battle.units = s.world.Units
		if _, err := (&app{tbl: s.tbl}).prepareJapaneseSupply(battle); err != nil {
			return bad()
		}
	}
	for _, raw := range [][]byte{saved.DT2, saved.MemWar} {
		if len(raw) > 0 {
			if _, err := game.ParseBattleStates(raw); err != nil {
				return nil, nil, nil, err
			}
		}
	}
	if (len(saved.DT2) == 0) != (len(saved.MemWar) == 0) {
		return bad()
	}
	return s, &saved, battle, nil
}

func (a *app) loadCampaign(data []byte) error {
	s, saved, battle, err := a.decodeCampaign(data)
	if err != nil {
		return err
	}
	// 期別字模亦先完成，避免核心替換後才發現必要素材缺失。
	gen, fan, err := a.campaignFonts(saved.Identity.Stage)
	if err != nil {
		return err
	}
	a.applySession(s)
	a.stage = int(saved.Identity.Stage)
	sc, _ := game.ScenarioByStage(saved.Identity.Stage)
	a.provinceLimit = sc.Provinces
	a.campaign = &saved.Identity
	a.rng = game.NewRand(saved.RNG)
	a.battle = battle
	a.fonts.Gen, a.fan = gen, fan
	a.battleDT2, a.battleMemWar = saved.DT2, saved.MemWar
	a.battleDT2Path, a.battleMemWarPath = "", ""
	if len(saved.DT2) > 0 {
		a.battleDT2States, _ = game.ParseBattleStates(saved.DT2)
		a.battleMemStates, _ = game.ParseBattleStates(saved.MemWar)
	}
	a.screen = screenMap
	if battle != nil {
		if !battle.paused {
			a.screen = screenBattle
		}
	} else if a.reorganization != nil {
		a.screen = screenReorganizeTarget
	}
	if view := s.pendingArmoryView; view != nil {
		a.armoryView = view.View
		if view.Open && battle == nil && a.reorganization == nil {
			a.screen = screenArmory
		}
	}
	a.opening = nil
	a.dirty = true
	return nil
}

func (a *app) campaignFonts(stage uint8) (*assets.GlyphFile, *assets.GlyphFile, error) {
	if a.dataRead == nil {
		return a.fonts.Gen, a.fan, nil
	}
	g, err := a.dataRead(fmt.Sprintf("MAN%d15", stage))
	if err != nil {
		return nil, nil, err
	}
	f, err := a.dataRead(fmt.Sprintf("FAN(%d).15", stage))
	if err != nil {
		return nil, nil, err
	}
	gen, err := assets.ParseGlyphFile(g)
	if err != nil {
		return nil, nil, err
	}
	fan, err := assets.ParseGlyphFile(f)
	return gen, fan, err
}

func (a *app) startCampaign(stage uint8, player game.GeneralID) error {
	return a.startCampaignMode(stage, player, "original")
}

func (a *app) startCampaignMode(stage uint8, player game.GeneralID, mode string) error {
	if mode != "original" && mode != "unification-resistance" || mode == "unification-resistance" && stage > 1 && player == 107 {
		return fmt.Errorf("invalid campaign mode or player")
	}
	s, identity, err := newCampaignSession(a.dataRead, stage, player)
	if err != nil {
		return err
	}
	s.militaryRules = &game.MilitaryRules{Enabled: mode == "unification-resistance"}
	gen, fan, err := a.campaignFonts(stage)
	if err != nil {
		return err
	}
	if mode == "unification-resistance" {
		sc, _ := game.ScenarioByStage(stage)
		s.armory, err = game.NewArmory(s.world, sc.Provinces, uint32(s.year)*12+uint32(s.month))
		if err != nil {
			return err
		}
		s.world.Opts.Armory = s.armory
		identity.Mode = mode
	}
	a.applySession(s)
	a.campaign = identity
	a.stage = int(stage)
	sc, _ := game.ScenarioByStage(stage)
	a.provinceLimit = sc.Provinces
	a.fonts.Gen, a.fan = gen, fan
	a.rng = game.NewRand(a.initialSeed)
	a.battleDT2, a.battleMemWar = nil, nil
	a.battleDT2Path, a.battleMemWarPath = "", ""
	a.battleDT2States = [game.ProvinceCount]game.BattleState{}
	a.battleMemStates = [game.ProvinceCount]game.BattleState{}
	a.opening = nil
	a.screen = screenMap
	a.msg = ""
	a.dirty = true
	return nil
}

func (a *app) ownsProvince(p game.ProvinceID) bool {
	if a.campaign == nil {
		return true
	} // 既有 DT1 入口保留原操作，僅新戰役固定身份。
	province, err := a.tbl.At(p)
	return err == nil && int(p) <= a.provinceLimit && province.Commander == a.playerCommander
}

func isCampaignSave(data []byte) bool { return bytes.HasPrefix(data, []byte("PK\x03\x04")) }
