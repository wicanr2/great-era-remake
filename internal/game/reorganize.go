package game

import "fmt"

// Reorganization 是指令 5「重新整編」的一次互動工作階段。
//
// 原版四條主函式 `sub_25B2B`／`sub_26D69`／`sub_27DCA`／`sub_28FF7`
// 都維護一個「目前剩餘兵力」池。玩家選一位將領後，該將現有兵力先進池，
// 再輸入要配回給他的數量；池歸零才可離開。
type Reorganization struct {
	w       *AIWorld
	branch  uint8
	targets []GeneralID
	pool    int
	arms    int // 步兵／騎兵的武裝總量（兵力 × 武裝程度）。
	changed bool
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
	return r, nil
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
