package game

// 戰鬥 AI 的決策鏈（`sub_3AB99`）與 13 種行動的分派（`sub_3F698`）。
//
// 完整反組譯見 `docs/re/31-battle-ai-chain.md`，13 種行動的語意表在 §41。
//
// 結構與政略決策鏈**完全同構**：優先序 + 一個「已決定」旗標，
// 第一個做出決定的那一步就是答案，後面的不再跑。
//
// ⚠️ **兩條鏈各管一方**（§32，原本標假說已確認）：
//
//	分支 A（原版 byte_6AA85）→ 第二方（0x764 陣列）
//	分支 B（原版 byte_6AA84）→ 第一方（0x750 陣列）
//
// 「第一方／第二方」**不等於**「攻方／守方」，與 `byte_64901` 的對應仍未驗。

// BattleAction 是決策鏈選出的行動編號。
//
// 分支 A 用 11–19，分支 B 用 1–4。兩條鏈的值域不重疊是原版的設計，
// 因為 `sub_3F698` 用同一張分派表。
type BattleAction int

// 分支 B（第一方）的四種行動 —— **全部有語意**。
const (
	// ActBDecisive 是 **1：必勝結算**。戰力差五倍時直接叫九步結算判勝負，
	// 不再逐格打（§16）。與分支 A 的 11 共用 `sub_3B19C`，只差參數。
	ActBDecisive BattleAction = 1
	// ActBDeploy 是 **2：佈防**。以首位單位所在格為中心，取周圍三圈 37 格
	// 依序發給還沒有去處的單位；已經站在城市上的不動（§30）。
	ActBDeploy BattleAction = 2
	// ActBTakeCity 是 **3：打城市**，也是分支 B 的**預設行動**（§19／§21）。
	// 挑距離最近、由敵方佔據、而且還沒有人去的城市。
	ActBTakeCity BattleAction = 3
	// ActBStrikeForce 是 **4：打敵方主力周邊**（§29）。
	// 骨架與 3 相同，只是中心點換成敵方主力所在格。
	ActBStrikeForce BattleAction = 4
)

// 分支 A（第二方）的九種行動 —— **全部有語意**（值 18 標部分）。
const (
	// ActADecisive 是 **11：必勝結算**（§16）。
	ActADecisive BattleAction = 11
	// ActAReset 是 **12：推倒重來**（§33）。先把單位打回待命，
	// 再重新挑一個「站著守方、而且走得通」的格。不看現有命令。
	ActAReset BattleAction = 12
	// ActADefault 是 **13：預設行動**（§32）。依單位現有的命令分流，
	// 保留狀態——與 12 的「推倒重來」正好相反。
	ActADefault BattleAction = 13
	// ActADecapitateKeepOne 是 **14：斬首，只留一個守城**（§34）。
	ActADecapitateKeepOne BattleAction = 14
	// ActADecapitateKeepAll 是 **15：斬首，所有駐守的都留**（§34）。
	//
	// ⚠️ **15 比 14 保守**，與編號直覺相反。原版是同一支函式，
	// 差別只在一個旗標設不設。
	ActADecapitateKeepAll BattleAction = 15
	// ActAStandbyOnly 是 **16：只處理待命與已出發的單位**（§35）。
	// 命令 2 無條件處理；命令 4／5 要 `byte_64900 ≥ 5` 才處理。
	ActAStandbyOnly BattleAction = 16
	// ActARecompute 是 **17：重算全軍的行動**（§14）。
	ActARecompute BattleAction = 17
	// ActAWeakest 是 **18：唯一比戰力的一支**（§38）。
	//
	// ⚠️ 語意標**部分**：排序方向與派工結構指向「挑最弱的打」，
	// 但被排序的清單來自兩段未讀完的統計。
	ActAWeakest BattleAction = 18
	// ActAEngageAll 是 **19：全面接戰**（§40）。把第二方全體設成「找目標」，
	// 目標池是第一方**所有在場**單位——與斬首（只打一個）相對。
	ActAEngageAll BattleAction = 19
)

// BattleDecision 是決策鏈跑完的結果。
type BattleDecision struct {
	// Action 是選中的行動。
	Action BattleAction
	// Step 是哪一步做的決定，給除錯與測試用。
	// `"預設"` 表示五步（或三步）都沒決定。
	Step string
}

// BattleAIInput 是決策鏈要看的外部狀態。
//
// ⚠️ 這裡只放**已經解出來**的輸入。原版每一步內部還有未讀的判斷
// （見 `Undecided`），補齊之前不要假裝這個結構已經完整。
type BattleAIInput struct {
	// SideStrength 是本方全部單位的攻擊力總和。
	SideStrength int
	// FoeStrength 是對方全部單位的攻擊力總和。
	FoeStrength int
	// FirstUnitStrength 是首位單位的攻擊力（`sub_58D4A` 用）。
	FirstUnitStrength int
	// RatioGateSelf / RatioGateFoe 是 §42 那組比率門檻的結果
	// （原版 `sub_3A63C` 與 `sub_3A730` 的 and，各對一方求值）。
	//
	//	門檻 = ( 比率 + byte_64900 ) < 15
	//
	// ⚠️ **比率的來源未解**——`word_64932/34/36/38` 是分子、`sub_3A4CE` 是分母，
	// 形狀像按比例衰減的累積量但沒有證據。所以這兩個布林由呼叫端傳，
	// 決策鏈本身不算。來源解出來之後只要改呼叫端，這裡的邏輯不動。
	RatioGateSelf bool
	RatioGateFoe  bool

	// DeployGateOpen 是分支 B 值 2 獨有的額外閘門（原版 `word_6493A == 0`，§43）。
	// ⚠️ 語意未解。
	DeployGateOpen bool

	// FoeLeaderOnField 是 `sub_56D49(0)`：**當前交戰省的司令本人
	// 在不在守方的隊伍裡**（§44）。
	//
	// 原版掃第二方的 10 個單位，看有沒有一個就是 `word_64944`（省司令）。
	// 四個決策點共用它——值 4、值 16、值 17、值 19。
	//
	// ⭐ **守方主帥親自坐鎮，是電腦改變打法的觸發點。**
	FoeLeaderOnField bool

	// Sub53619 是 `sub_53619(side)` 的結果——**三處分流**都靠它（§45）。
	//
	//	sub_3A817   == 0 → 11 必勝結算    != 0 → 12／17
	//	sub_3AA51   == 0 →  1 必勝結算    != 0 →  3 打城市
	//	sub_3A94E   == 0 → 16             != 0 → 17
	//
	// ⭐ **兩處「必勝結算」都要它 == 0**——戰力差五倍還不夠，
	// 還要這個條件成立才准直接判勝負。
	//
	// ⭐ **語意已解**（§47）：`sub_534FF` 找的是「當前交戰省的鄰省裡，
	// 屬於我方而且可用的省」，`sub_53619` 回「有沒有那種省」的反相。
	//
	//	Sub53619 == false  → 有可用的我方鄰省（**有後援**）
	//	Sub53619 == true   → 沒有
	//
	// 所以兩處必勝結算的完整條件是：
	// **戰力差五倍，而且有我方鄰省可以支援，才敢直接判勝負。**
	// 沒有後援就不冒險速戰，改走打城市或推倒重來。
	//
	// 命名保留 `Sub53619` 是因為它的**反相**語意容易讓人讀錯——
	// 叫 `HasSupport` 的話，`!HasSupport` 才是必勝結算的條件，更繞。
	Sub53619 bool

	// EnableLastSteps 是 `byte_6FFCA & 4`：**啟用後面幾步**。
	//
	// 與政略決策鏈的「啟用最後三步」是同一個位元、同一個手法
	// （`docs/mechanics/70-ai.md` §6d）——這是 bit 2 的第四個用途。
	EnableLastSteps bool
}

// DecideBattleB 跑分支 B（第一方）的決策鏈（`sub_3AB99` 的 `arg_0 != 0` 那半）。
//
//	sub_3A9F4                            → 2
//	if !決定: sub_3AA51                  → 1 / 3
//	if !決定 且 byte_6FFCA & 4: sub_3AABA → 4
//	if !決定: 預設 3
//
// ⚠️ **只有 `sub_3AA51` 的必勝判斷解出來了**（§16 的 5.0 倍率）。
// `sub_3A9F4`（何時佈防）與 `sub_3AABA`（何時打主力周邊）**都未讀**，
// 所以這裡跑出來的結果會偏向「必勝結算」與「預設打城市」兩種。
// 這是**已知的偏差**，不是 bug——補讀那兩支之前不要拿它對原版做行為驗收。
func DecideBattleB(in BattleAIInput) BattleDecision {
	// 第一步 `sub_3A9F4` → 值 2（佈防）。
	//
	// ⭐⭐ 條件與分支 A 的 `sub_3A885`（值 12）**完全相同**，
	// 只多一個 `word_6493A == 0`（§43）：
	//
	//	同一個局勢下，分支 A 選「推倒重來」、分支 B 選「佈防」——
	//	兩者都是重整型的行動，兩條鏈的反應是同構的。
	//
	// ⚠️ `word_6493A` 的語意未解，用 `DeployGateOpen` 讓呼叫端傳。
	if in.RatioGateSelf && in.RatioGateFoe && in.DeployGateOpen {
		return BattleDecision{Action: ActBDeploy, Step: "sub_3A9F4 兩方比率門檻"}
	}

	// 第二步 `sub_3AA51`（§45）：第一方被壓到第二方的五分之一以下之後，
	// 由 `sub_53619(1)` 二選一。
	if ForceRatioLE(in.SideStrength, in.FoeStrength,
		AIBattleRatioCollapseNum, AIBattleRatioCollapseDen) {
		if !in.Sub53619 {
			return BattleDecision{Action: ActBDecisive, Step: "sub_3AA51 必勝門檻"}
		}
		return BattleDecision{Action: ActBTakeCity, Step: "sub_3AA51 改打城市"}
	}

	// 第三步 `sub_3AABA` → 值 4（§42）：`sub_56D49(0)` 成立就設。
	// ⚠️ `sub_56D49` 未讀（81 行，三處共用），用 `Sub56D49` 讓呼叫端傳。
	if in.EnableLastSteps && in.FoeLeaderOnField {
		return BattleDecision{Action: ActBStrikeForce, Step: "sub_3AABA"}
	}

	return BattleDecision{Action: ActBTakeCity, Step: "預設"}
}

// DecideBattleA 跑分支 A（第二方）的決策鏈（`arg_0 == 0` 那半）。
//
//	sub_3A817                            → 11 / 12 / 16 / 17
//	if !決定: sub_3A885                  → 12
//	if !決定: sub_3A8C8                  → 19
//	if !決定 且 byte_6FFCA & 4: sub_3A8F7 → 14 / 15 / 18
//	if !決定 且 byte_6FFCA & 4: sub_3A94E → 16 / 17
//	if !決定: 預設 13
//
// ⚠️ 與分支 B 同樣的限制：五步裡**只有兩個倍率判斷解出來了**
// （`sub_3A817` 的 5.0、`sub_3A8F7` 的 0.67，§16）。
// 每一步內部「在多個值之間怎麼選」都還沒讀，所以這裡只實作
// 「哪一步會出手」，選不到細分值時取該步的第一個值。
func DecideBattleA(in BattleAIInput) BattleDecision {
	// 第一步 `sub_3A817`（§45）：第二方被壓到第一方的五分之一以下之後，
	// 還要看 `sub_53619` 與 `sub_56D49` 才知道給哪個值。
	if ForceRatioLE(in.SideStrength, in.FoeStrength,
		AIBattleRatioCollapseNum, AIBattleRatioCollapseDen) {
		if !in.Sub53619 {
			// 戰力差五倍**還不夠**，還要 `sub_53619 == 0` 才准直接判勝負。
			// 原版這裡另外設 `byte_6B968 = 1`（語意未解）。
			return BattleDecision{Action: ActADecisive, Step: "sub_3A817 必勝門檻"}
		}
		if !in.FoeLeaderOnField {
			return BattleDecision{Action: ActAReset, Step: "sub_3A817 守方領袖不在場"}
		}
		// ⛔ 原版這裡再問一次 `sub_53619(0)`，但外層已經確定它非 0，
		// 而 `sub_56D49` 是純查詢無副作用——所以**恆走 17，值 16 走不到**。
		// 那是原版的死碼，照抄（`CLAUDE.md` §9：原版行為就是規格）。
		// 值 16 本身不是死碼，`sub_3A94E` 走得到它。
		return BattleDecision{Action: ActARecompute, Step: "sub_3A817（值 16 是死碼）"}
	}

	// 第二步 `sub_3A885` → 值 12；第三步 `sub_3A8C8` → 值 19。
	//
	// 兩者**共用同一組比率門檻**（§42），只差看幾方：
	//
	//	兩方都成立 → 12（推倒重來）
	//	只有第二方 → 19（全面接戰）
	//
	// ⚠️ 比率的**來源**還沒解（`word_64932/34/36/38` 與 `sub_3A4CE`），
	// 所以這裡收 `RatioGateSelf`／`RatioGateFoe` 兩個布林讓呼叫端傳，
	// 而不是在這裡算——來源解出來之後只要改呼叫端。
	if in.RatioGateSelf && in.RatioGateFoe {
		return BattleDecision{Action: ActAReset, Step: "sub_3A885 兩方比率門檻"}
	}
	if in.RatioGateSelf {
		return BattleDecision{Action: ActAEngageAll, Step: "sub_3A8C8 我方比率門檻"}
	}

	// 第四步 `sub_3A8F7`：要 `byte_6FFCA & 4`。
	// 條件是**不成立**才往下走（§16 的表）——也就是「敵方 < 我方的 2/3」。
	if in.EnableLastSteps && !ForceRatioLE(in.SideStrength, in.FoeStrength,
		AIBattleRatioEvenNum, AIBattleRatioEvenDen) {
		// ⚠️ 這一步給 14／15／18 三個值，怎麼選**未讀**。
		// 取 14（斬首、只留一個守城）當代表——標為已知偏差。
		return BattleDecision{Action: ActADecapitateKeepOne, Step: "sub_3A8F7 勢均"}
	}

	// 第五步 `sub_3A94E`（§43）：`sub_56D49(0)` 是前置，
	// `sub_53619(0)` 決定 16 還是 17。
	if in.EnableLastSteps && in.FoeLeaderOnField {
		if in.Sub53619 {
			return BattleDecision{Action: ActARecompute, Step: "sub_3A94E"}
		}
		return BattleDecision{Action: ActAStandbyOnly, Step: "sub_3A94E"}
	}

	return BattleDecision{Action: ActADefault, Step: "預設"}
}

// UndecidedBattleSteps 列出決策鏈裡**還沒讀的步驟**。
//
// 放在程式碼裡而不是只放文件，因為呼叫端需要知道
// 「這條鏈現在只有部分行為」——測試也拿它來確認清單沒有偷偷縮水。
//
// 補完一支就從這裡移除一筆，並在 `docs/re/31` 補一節。
var UndecidedBattleSteps = []string{
	"sub_3A8F7 在 14／15／18 之間怎麼選（分支 A 第四步）",
	"比率門檻的來源：word_64932/34/36/38 與 sub_3A4CE（§42 挖到第五層停手）",
	"sub_534FF 的第二個輸出（決定 sub_53619）＋ word_6493A ＋ byte_6B968",
}

// BattleActionName 回傳行動的中文名稱，給紀錄與測試訊息用。
func BattleActionName(a BattleAction) string {
	switch a {
	case ActBDecisive, ActADecisive:
		return "必勝結算"
	case ActBDeploy:
		return "佈防"
	case ActBTakeCity:
		return "打城市"
	case ActBStrikeForce:
		return "打敵方主力周邊"
	case ActAReset:
		return "推倒重來"
	case ActADefault:
		return "預設分流"
	case ActADecapitateKeepOne:
		return "斬首（只留一個守城）"
	case ActADecapitateKeepAll:
		return "斬首（駐守的都留）"
	case ActAStandbyOnly:
		return "只處理待命與已出發"
	case ActARecompute:
		return "重算全軍"
	case ActAWeakest:
		return "挑最弱的打"
	case ActAEngageAll:
		return "全面接戰"
	}
	return "未知行動"
}
