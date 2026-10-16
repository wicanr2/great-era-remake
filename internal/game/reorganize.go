package game

import "fmt"

// Reorganization 是指令 5「重新整編」的一次互動工作階段。
//
// 原版四條主函式 `sub_25B2B`／`sub_26D69`／`sub_27DCA`／`sub_28FF7`
// 都維護一個「目前剩餘兵力」池。玩家選一位將領後，該將現有兵力先進池，
// 再輸入要配回給他的數量；池歸零才可離開。
type Reorganization struct {
	w                *AIWorld
	branch           uint8
	targets          []GeneralID
	pool             int
	arms             int // 步兵／騎兵的武裝總量（兵力 × 武裝程度）。
	changed          bool
	initial          []ReorganizationMember
	initialEquipment *Armory
}

type ReorganizationMember struct {
	General GeneralID
	Force   uint16
	Arms    uint8
}

// ReorganizationSnapshot 保存已離開將領記錄、仍在整編池中的兵力。
type ReorganizationSnapshot struct {
	Branch           uint8
	Targets          []GeneralID
	Pool, Arms       int
	Changed          bool
	Initial          []ReorganizationMember `json:",omitempty"`
	InitialEquipment *Armory                `json:",omitempty"`
}

func (r *Reorganization) Snapshot() *ReorganizationSnapshot {
	if r == nil {
		return nil
	}
	return &ReorganizationSnapshot{Branch: r.branch, Targets: r.Targets(), Pool: r.pool, Arms: r.arms, Changed: r.changed,
		Initial: append([]ReorganizationMember(nil), r.initial...), InitialEquipment: r.initialEquipment.Clone()}
}

func (w *AIWorld) RestoreReorganization(p ProvinceID, state *ReorganizationSnapshot) (*Reorganization, error) {
	if state == nil {
		return nil, nil
	}
	r, err := w.BeginReorganization(p, state.Branch)
	if err != nil {
		return nil, err
	}
	if len(r.targets) != len(state.Targets) || state.Pool < 0 || int64(state.Pool) > int64(len(w.Units))*65535 || state.Arms < 0 || int64(state.Arms) > int64(state.Pool)*255+65535 {
		return nil, fmt.Errorf("game: 整編池或候選不合法")
	}
	for i, id := range r.targets {
		if state.Targets[i] != id {
			return nil, fmt.Errorf("game: 整編候選與世界不一致")
		}
	}
	r.pool, r.arms, r.changed = state.Pool, state.Arms, state.Changed
	r.initial = append([]ReorganizationMember(nil), state.Initial...)
	r.initialEquipment = state.InitialEquipment.Clone()
	if len(r.initial) > 0 {
		if _, err := r.rollbackWorld(); err != nil {
			return nil, err
		}
	} else if r.initialEquipment != nil {
		return nil, fmt.Errorf("game: 整編裝備快照缺少初始人員")
	}
	return r, nil
}

// BeginReorganization 建立某省、某兵種的整編工作階段。
// 候選照原版掃描順序（將領 ID 由小到大），不套 RosterOf 的戰力排序。
func (w *AIWorld) BeginReorganization(p ProvinceID, branch uint8) (*Reorganization, error) {
	prov, err := w.Table.At(p)
	if err != nil {
		return nil, err
	}
	if BranchFullStrength(branch) == 0 {
		return nil, fmt.Errorf("game: 未知兵種 %d", branch)
	}
	r := &Reorganization{w: w, branch: branch}
	for i := range w.Units {
		u := &w.Units[i]
		if !u.Active || u.Province != p || u.Faction != prov.Commander ||
			i >= len(w.Strengths) || w.Strengths[i].Branch != branch {
			continue
		}
		r.targets = append(r.targets, u.General)
	}
	if len(r.targets) == 0 {
		return nil, fmt.Errorf("game: 省 %d 沒有可整編的%s", p, BranchName(branch))
	}
	for _, id := range r.targets {
		i := r.indexOf(id)
		s := w.Strengths[i]
		r.initial = append(r.initial, ReorganizationMember{General: id, Force: s.Force, Arms: s.F20})
	}
	if w.Opts.Armory != nil {
		r.initialEquipment = w.Opts.Armory.Clone()
		if err := r.initialEquipment.Sync(w); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// CanCancel is false for historical saves that never recorded a starting state.
func (r *Reorganization) CanCancel() bool { return r != nil && len(r.initial) > 0 }

func (r *Reorganization) rollbackWorld() (*AIWorld, error) {
	if !r.CanCancel() || len(r.initial) != len(r.targets) {
		return nil, fmt.Errorf("game: 沒有完整的整編前配置")
	}
	world := *r.w
	world.Strengths = append([]StrengthInput(nil), r.w.Strengths...)
	before, now := uint64(0), uint64(r.pool)
	for j, member := range r.initial {
		i := r.indexOf(member.General)
		if member.General != r.targets[j] || i < 0 || i >= len(world.Strengths) || world.Strengths[i].Branch != r.branch {
			return nil, fmt.Errorf("game: 整編初始候選不一致")
		}
		before += uint64(member.Force)
		now += uint64(world.Strengths[i].Force)
		world.Strengths[i].Force, world.Strengths[i].F20 = member.Force, member.Arms
	}
	if before != now {
		return nil, fmt.Errorf("game: 整編兵力快照不守恆")
	}
	if (r.w.Opts.Armory != nil) != (r.initialEquipment != nil) {
		return nil, fmt.Errorf("game: 整編缺少裝備快照")
	}
	if r.initialEquipment != nil {
		if err := r.initialEquipment.Validate(&world); err != nil {
			return nil, err
		}
		for i, u := range world.Units {
			e := r.initialEquipment.Units[u.General]
			if e == nil || e.Faction != u.Faction || e.Force != world.Strengths[i].Force || e.Branch != world.Strengths[i].Branch {
				return nil, fmt.Errorf("game: 整編初始裝備與人員不一致")
			}
		}
	}
	return &world, nil
}

// Cancel restores a fully validated start state before changing any live data.
// Equipment bookkeeping caused by saving a draft is rolled back with personnel.
func (r *Reorganization) Cancel() error {
	world, err := r.rollbackWorld()
	if err != nil {
		return err
	}
	for _, member := range r.initial {
		i := r.indexOf(member.General)
		r.w.Strengths[i].Force, r.w.Strengths[i].F20 = world.Strengths[i].Force, world.Strengths[i].F20
	}
	if r.initialEquipment != nil {
		*r.w.Opts.Armory = *r.initialEquipment.Clone()
	}
	r.pool, r.arms, r.changed = 0, 0, false
	return nil
}

// Targets 回傳原版順序的候選將領 ID 副本。
func (r *Reorganization) Targets() []GeneralID {
	return append([]GeneralID(nil), r.targets...)
}

// Remaining 是尚未指派的兵力。非零時原版不允許離開整編畫面。
func (r *Reorganization) Remaining() int { return r.pool }

// CanFinish 回報是否已把兵力池全部配完。
func (r *Reorganization) CanFinish() bool { return r.pool == 0 }

// Changed 回報這次工作階段是否真的改過兵力或武裝程度。
func (r *Reorganization) Changed() bool { return r.changed }

func (r *Reorganization) indexOf(id GeneralID) int {
	for i := range r.w.Units {
		if r.w.Units[i].General == id {
			for _, target := range r.targets {
				if target == id {
					return i
				}
			}
			return -1
		}
	}
	return -1
}

// Limit 回傳選中該將後可指派的最大兵力：池中兵力加上該將目前兵力，
// 再夾到兵種滿員上限。
func (r *Reorganization) Limit(id GeneralID) int {
	i := r.indexOf(id)
	if i < 0 || i >= len(r.w.Strengths) {
		return 0
	}
	limit := r.pool + int(r.w.Strengths[i].Force)
	if full := int(BranchFullStrength(r.branch)); limit > full {
		limit = full
	}
	return limit
}

// Assign 選中一位將領並指派兵力，語意照四支 `分配兵力` 函式。
//
// 步兵 `sub_25527` 與騎兵 `sub_289F3` 會把池中與該將的武裝總量合併，
// 依兵力加權平均後寫回武裝程度；砲兵與裝甲兵的對應函式不碰該欄位。
func (r *Reorganization) Assign(id GeneralID, force int) error {
	i := r.indexOf(id)
	if i < 0 || i >= len(r.w.Strengths) {
		return fmt.Errorf("game: 將領 %d 不在整編候選中", id)
	}
	if force < 0 || force > r.Limit(id) {
		return fmt.Errorf("game: 將領 %d 可指派 0..%d，要求 %d", id, r.Limit(id), force)
	}
	s := &r.w.Strengths[i]
	oldForce, oldArms := s.Force, s.F20
	total := r.pool + int(s.Force)

	if r.branch == BranchInfantry || r.branch == BranchCavalry {
		totalArms := r.arms + int(s.Force)*int(s.F20)
		armsLevel := 10
		usedArms := 0
		if total > 0 {
			armsLevel = roundPositive(totalArms, total)
			usedArms = roundPositive(force*totalArms, total)
		}
		if armsLevel > 100 {
			armsLevel = 100
		}
		s.F20 = uint8(armsLevel)
		r.arms = totalArms - usedArms
		if r.arms < 0 {
			r.arms = 0
		}
	}

	s.Force = uint16(force)
	r.pool = total - force
	if s.Force != oldForce || s.F20 != oldArms {
		r.changed = true
	}
	return nil
}

// roundPositive 等價於本處 Turbo Pascal Real 的正數 Round。
func roundPositive(n, d int) int {
	if d <= 0 || n <= 0 {
		return 0
	}
	return (n*2 + d) / (2 * d)
}
