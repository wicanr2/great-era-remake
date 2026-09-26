package game

import "fmt"

// 外交（政略指令 9）的另外兩項：請求外援與償還外債。
// 貸款在 `loan.go`。
//
//	外援      `sub_21D1D`（738 行）
//	償還外債  `sub_223ED`（556 行）

// CreditMax 是信用度的上限（`cmp byte ptr [di-4225h], 64h`）。
//
// 償還外債會加信用度，加完夾到 100。與貸款的扣是**完全對稱**的：
//
//	貸款      信用度 −= 額度 ÷ 500
//	償還外債  信用度 += 額度 ÷ 500，夾到 100
//
// 兩邊用同一個 `LoanUnit`（500），所以借多少還多少就回到原點。
const CreditMax = 100

// 外援的拒絕判定：`Random(10) ≤ 6` → **70% 拒絕**。
//
// `sub_21D1D` 在這個區間跳到 `loc_21EF5`，該區塊逐字畫出
// 「各國均拒絕提供援助」；只有大於 6 才落到 `loc_21FCA` 的資源寫入路徑。
const (
	AidRollRange   = 10
	AidRefusalMax  = 6
	AidApprovalMin = AidRefusalMax + 1
)

// AidGoldCap 是外援黃金的上限（`cmp [bp+var_4], 1770h` → 6000）。
//
// 只有黃金被夾——糧食／彈藥／燃料那三格沒有對應的檢查。
const AidGoldCap = 6000

// AidResourceBases 回傳原版請求外援時使用的四個亂數基準。
//
// `sub_21D1D` 不是直接用一個固定上限；它先對當前省的黃金、糧食、
// 彈藥、燃料各做 `60000 - 目前值`，再把這個差額交給 `Random`。因此
// 資源越少，可能取得的援助越多；資源已達上限時基準為 0。
// 順序與 `RequestAid` 的 base[4] 相同：黃金、糧食、彈藥、燃料。
func AidResourceBases(p *Province) [4]int {
	if p == nil {
		return [4]int{}
	}
	cap := int(ResourceCap)
	clamp := func(v uint16) int {
		b := cap - int(v)
		if b < 0 {
			return 0
		}
		return b
	}
	return [4]int{clamp(p.Gold), clamp(p.Food), clamp(p.Ammo), clamp(p.Fuel)}
}

// AidResourceBases 讀取某省的外援亂數基準，保留與規則方法相同的省份
// 驗證邊界。這個方法只讀世界狀態，不擲骰，也不改資源。
func (w *AIWorld) AidResourceBases(p ProvinceID) ([4]int, error) {
	prov, err := w.Table.At(p)
	if err != nil {
		return [4]int{}, err
	}
	return AidResourceBases(prov), nil
}

// 援助國的慷慨程度。原版用一個代碼（`var_14`）分流出**除數**，
// 援助量是 `Random(基準) ÷ 除數`，所以**除數越小給越多**。
//
//	代碼 99 (63h) 或 142 (8Eh) → 除數 1   最慷慨
//	代碼 100 (64h)             → 除數 3
//	其他（101、141…）          → 除數 5
//
// 那些代碼看起來是 `2.15` 的詞條索引（1-based 99／100／101 依序是
// 美國／英國／俄國，`50-diplomacy.md` §3）。若真是如此，
// **美國最慷慨**——與 `sub_37A81`「美國援助軍事物資中」那個事件呼應。
//
// ⚠️ **代碼與國家的對應標為假說**：數值對得上詞條位置，但沒有直接證據，
// 而且 141／142 對不上「法國」的詞條編號。
const (
	AidDonorGenerous = 99  // 除數 1
	AidDonorMedium   = 100 // 除數 3
	AidDonorGenerou2 = 142 // 除數 1
)

// AidDonorCode 是 `sub_21D1D` 的第一期援助國代碼選擇器。
//
// `factionCode` 對應 `.DT1` 區塊 7 的 1-based 將領→勢力反查值；
// `roll` 是同一支函式最前面的 `Random(10)` 結果。第二／三期在
// `22088h` 直接把代碼覆成 99，第一期才使用下列分支。
//
// 這裡只回傳原版數值代碼，不替 `2.15` 詞條索引命名成國家；
// 代碼與國名的對應仍是未閉合假說。
func AidDonorCode(stage uint8, factionCode, roll int) int {
	donor := roll
	switch factionCode {
	case 10:
		// `div 2` 後 `xchg ax,dx` 比的是原始 roll 的餘數。
		if roll%2 == 1 {
			donor = AidDonorGenerous
		} else {
			donor = AidDonorMedium
		}
	case 1:
		donor = AidDonorGenerou2
	case 5:
		donor = AidDonorMedium
	case 4:
		switch roll {
		case 6:
			donor = AidDonorMedium
		case 7:
			donor = 146
		default:
			donor = AidDonorGenerous
		}
	case 3, 2:
		donor = 101
	default:
		switch roll {
		case 6:
			donor = AidDonorGenerous
		case 7:
			donor = AidDonorMedium
		case 8:
			donor = 101
		case 9:
			donor = 141
		}
	}
	if stage != 1 {
		return AidDonorGenerous
	}
	return donor
}

// AidDivisor 回傳某個援助國代碼對應的除數。
func AidDivisor(donor int) int {
	switch donor {
	case AidDonorGenerous, AidDonorGenerou2:
		return 1
	case AidDonorMedium:
		return 3
	}
	return 5
}

// 外援的原版特殊分支：張作霖在民國 17 年 2–6 月會跳過隨機拒絕文字，
// 直接進入核准／資源寫入路徑。
//
//	cmp  word ptr [di-6221h], 0A6h   ; 省份司令 == 166（張作霖）
//	cmp  byte_6FE7D, 11h             ; 年 == 17
//	cmp  byte_6FE7E, 2 / 6           ; 月在 2..6
//	→ 直接跳到 `loc_21FCA`（核准路徑）
//
// 這裡只記錄控制流，不把它命名成「禁運」或替原版推定歷史動機；
// 原版的日期／人物語意與現代史實解釋仍需分開。
const (
	AidSpecialLeader     = 166
	AidSpecialYear       = 17
	AidSpecialMonthFirst = 2
	AidSpecialMonthLast  = 6
)

// AidResult 記錄一次外援請求的結果。
type AidResult struct {
	Approved bool
	// CommandCompleted 表示已通過援助指令的前置驗證並立起原版
	// `byte_6FE81`。原版在第一次 Random(10) 前就立旗標，所以隨機拒絕
	// 與特殊核准都會消耗一個指令；呼叫端錯誤或尚未進入本函式則不會立旗標。
	CommandCompleted bool
	Roll             int
	// Donor 是援助國代碼，Divisor 是它對應的除數。
	Donor, Divisor int
	// 四種資源實際入帳的量。
	Gold, Food, Ammo, Fuel int
	// SpecialOverride 表示命中 `sub_21D1D` 的日期／司令特殊分支；
	// 這個旗標只描述控制流，不替原版推定歷史名稱。
	SpecialOverride bool
}

// RequestAid 向列強請求援助（`sub_21D1D`）。
//
// `base` 是四種資源各自的亂數基準。IDA 已確認原版先算
// `ResourceCap - 目前值`（見 `AidResourceBases`），再由呼叫端把這四格
// 傳進來；保留參數是為了讓規則測試可以注入固定基準。
//
// 流程：
//
//	Random(10) ≤ 6 → 拒絕（70%）
//	Random(10) ≥ 7 → 同意（30%）
//	張作霖 + 民國 17 年 2–6 月 → 略過隨機拒絕，直接同意
//	同意：四種資源各 += Random(base[i]) ÷ 除數，黃金夾 6000
func (w *AIWorld) RequestAid(p ProvinceID, st GameState, donor int,
	base [4]int, rng *Rand) (AidResult, error) {
	return w.requestAid(p, st, base, rng, func(int) int { return donor })
}

// RequestAidForFaction 使用原版第一期的援助國代碼分支。
//
// `factionCode` 可直接來自 `FactionOfGeneral.SlotOf`；若目前 session 沒有
// `.DT1` 區塊 7，傳 0 會保留原版 default 分支，而不把缺資料誤當成某個勢力。
func (w *AIWorld) RequestAidForFaction(p ProvinceID, st GameState, factionCode int,
	base [4]int, rng *Rand) (AidResult, error) {
	return w.requestAid(p, st, base, rng, func(roll int) int {
		return AidDonorCode(st.Stage, factionCode, roll)
	})
}

func (w *AIWorld) requestAid(p ProvinceID, st GameState, base [4]int,
	rng *Rand, donorForRoll func(int) int) (AidResult, error) {
	prov, err := w.Table.At(p)
	if err != nil {
		return AidResult{}, err
	}

	res := AidResult{}
	// `sub_21D1D` 在援助判定的第一次亂數前立 byte_6FE81；不要把
	// Approved 誤當成「指令已完成」，所有結果分支都會帶著這個旗標回主迴圈。
	res.CommandCompleted = true
	res.Roll = rng.Int(AidRollRange)
	res.Donor = donorForRoll(res.Roll)
	res.Divisor = AidDivisor(res.Donor)
	if res.Roll <= AidRefusalMax {
		// 原版只有這個日期／司令組合會從拒絕文字分支跳回核准路徑。
		if prov.Commander != AidSpecialLeader || st.Year != AidSpecialYear ||
			st.Month < AidSpecialMonthFirst || st.Month > AidSpecialMonthLast {
			return res, nil
		}
		res.SpecialOverride = true
	}

	res.Approved = true
	give := func(field *uint16, b int, cap int) int {
		if b <= 0 {
			return 0
		}
		v := rng.Int(b) / res.Divisor
		if cap > 0 && v > cap {
			v = cap
		}
		before := *field
		*field = AddResource(*field, uint16(v))
		return int(*field - before)
	}
	res.Gold = give(&prov.Gold, base[0], AidGoldCap)
	res.Food = give(&prov.Food, base[1], 0)
	res.Ammo = give(&prov.Ammo, base[2], 0)
	res.Fuel = give(&prov.Fuel, base[3], 0)
	return res, nil
}

// RepayDebt 償還外債（`sub_223ED`）的省份／信用度部分。
//
// 與貸款完全對稱：每 500 黃金加 1 點信用度，加完夾到 100。
//
//	額度 ÷ 500 → dl
//	信用度 += dl
//	若 信用度 > 100 → 100
//
// 外債本身是執行期 `DiplomacyLedger.Debt` 的 32-bit 表；需要同時扣除
// 帳本時請使用 `DiplomacyLedger.RepayDebt`。保留這個純規則入口，讓既有
// 呼叫端可以只測「付錢 + 加信用度」而不隱藏帳本副作用。
func (w *AIWorld) RepayDebt(p ProvinceID, amount int, credit uint8) (int, uint8, error) {
	prov, err := w.Table.At(p)
	if err != nil {
		return 0, credit, err
	}
	if amount <= 0 {
		return 0, credit, fmt.Errorf("game: 償還額要為正（%d）", amount)
	}
	if int(prov.Gold) < amount {
		return 0, credit, fmt.Errorf("game: 黃金 %d 不足 %d", prov.Gold, amount)
	}
	prov.Gold -= uint16(amount)

	units := amount / LoanUnit
	v := int(credit) + units
	if v > CreditMax {
		v = CreditMax
	}
	return units, uint8(v), nil
}
