package game

import "fmt"

// 談判停火（政略指令 10），出自 `sub_211D5`（440 行）與它底下的
// `sub_20CF0`（同意判定）、`sub_21168`（取對手）、`sub_20E05`（套用結果）。
//
// 畫面上的話：「司令不在本省／欲在何省談判停火？／並無戰事／同意／拒絕」。

// 停火判定的兩檔門檻（`sub_20CF0`）。
//
//	roll = Random(10)
//	請求方佔上風 → roll ≥ 3 才同意   → 7/10
//	請求方居劣勢 → roll ≥ 8 才同意   → 2/10
//
// **你強對方才肯停火**——這條規則讀起來完全合理，
// 也是目前解出來最「有設計感」的一段 AI 判斷。
const (
	CeasefireRollRange = 10
	// CeasefireStrongMin 是佔上風時的最低骰值（含）。
	CeasefireStrongMin = 3
	// CeasefireWeakMin 是居劣勢時的最低骰值（含）。
	CeasefireWeakMin = 8
)

// 兩張與戰爭狀態有關的表，位址從 `sub_20E05`／`sub_21168` 讀到。
//
// ⚠️ **兩張都還沒完全解**，記位址是為了以後接得上。
const (
	// WarRecordAddr 是每省的戰爭記錄表：`ds:B346h` 起、**60 B/筆**，
	// 以省編號索引（`mul 3Ch`）。已知 `+0` 與 `+2` 是交戰的兩個勢力
	// （`sub_21168` 依「省的司令是不是我方」二選一）。
	//
	// 39 省 × 60 = 2,340 B，正好落在 `CeasefireStateAddr` 之前。
	WarRecordAddr = 0xB346
	// CeasefireStateAddr 是停火狀態表：`ds:BCA5h` 起、每省 1 byte。
	//
	//	值 > 0  → 停火中
	//	同意時  `inc byte ptr [di-435Bh]`
	//
	// 遞增而不是設成固定值，形狀像**剩餘停火月數**，但沒有證據。
	CeasefireStateAddr = 0xBCA5
)

// CeasefireResult 記錄一次停火談判的結果。
type CeasefireResult struct {
	Agreed bool
	Roll   int
	// RequesterStronger 是請求方在該省是不是佔上風。
	RequesterStronger bool
	// AttackForce / DefendForce 是雙方已部署單位的攻擊力總和。
	AttackForce, DefendForce int
}

// BattleForces 累加某省在場單位的攻守雙方戰力（`sub_20CF0` 的前半）。
//
// 原版的篩選條件三個，全部對得上 `docs/spec/02` 的欄位：
//
//	將領 +4  == 目標省          （所屬省）
//	將領 +16 bit 2 == 1         （已部署）
//	將領 +8  == 0 → 守方，否則攻方
//
// ⚠️ **remake 差異**：`CombatUnit` 沒有建模 `+8`（攻守旗標）與
// `+16` bit 2（已部署），所以這裡改用等價的判斷——
//
//	在場   → `Cell.Valid()`（`+5` != 0xFF）
//	守方   → 效忠對象 == 該省司令，其餘算攻方
//
// 在「某省正在交戰」這個前提下兩者等價（`docs/re/06` 的增援邏輯
// 就是拿 `+14` 與省份 `+20` 直接比）。要完全照抄得先把 `+8` 建模進來。
//
// 用的是攻擊力公式 `sub_5A0B9`（`strength.go` 的 `Strength`）——
// **AI 的評估函式與戰鬥用的是同一支**，`70-ai.md` §6e 已經記過，
// 這裡是第二個獨立的例子。
func (w *AIWorld) BattleForces(p ProvinceID) (attack, defend int) {
	prov, err := w.Table.At(p)
	if err != nil {
		return 0, 0
	}
	for i := range w.Units {
		u := &w.Units[i]
		if u.Province != p || !u.Cell.Valid() || i >= len(w.Strengths) {
			continue
		}
		f := Strength(w.Strengths[i], w.Opts)
		if u.Faction == prov.Commander {
			defend += f
		} else {
			attack += f
		}
	}
	return
}

// NegotiateCeasefire 在某省談停火（`sub_211D5` → `sub_20CF0`）。
//
// `requester` 是提出的一方（勢力領袖 ID）。判定照原版：
//
//	佔上風 = 請求方是守方且守方戰力 ≥ 攻方，或請求方是攻方且攻方 ≥ 守方
//	佔上風 → Random(10) ≥ 3 同意
//	劣勢   → Random(10) ≥ 8 同意
//
// ⚠️ 同意之後原版做的是 `inc` 停火狀態表（`ds:BCA5h`），
// **那張表的語意還沒解**（見 `CeasefireStateAddr`），
// 所以這裡只回報「談成了沒」，**不改任何狀態**。
func (w *AIWorld) NegotiateCeasefire(p ProvinceID, requester GeneralID,
	rng *Rand) (CeasefireResult, error) {
	prov, err := w.Table.At(p)
	if err != nil {
		return CeasefireResult{}, err
	}
	if prov.Commander == 0 {
		return CeasefireResult{}, fmt.Errorf("game: 省 %d 無主，談不了停火", p)
	}

	res := CeasefireResult{}
	res.Roll = rng.Int(CeasefireRollRange)
	res.AttackForce, res.DefendForce = w.BattleForces(p)

	// 請求方是這個省的司令 → 他是守方，比守方戰力；否則他是攻方。
	if prov.Commander == requester {
		res.RequesterStronger = res.DefendForce >= res.AttackForce
	} else {
		res.RequesterStronger = res.AttackForce >= res.DefendForce
	}

	min := CeasefireWeakMin
	if res.RequesterStronger {
		min = CeasefireStrongMin
	}
	res.Agreed = res.Roll >= min
	return res, nil
}
