package game

import "fmt"

// 商業活動（政略指令 12）：跟外界買賣資源。
//
// 十支函式的入口是兩個選單（`docs/re/27`）：
//
//	進口 `sub_1D703`   糧食 `sub_1CB96`／彈藥 `sub_1CF62`／燃料 `sub_1D32E`
//	出口 `sub_1ED95`   糧食 `sub_1DAA8`／彈藥 `sub_1DE6E`／煤礦 `sub_1E234`／
//	                   鐵礦 `sub_1E5FA`／燃料 `sub_1E9C0`
//
// 每一支的形狀完全相同，只差匯率與動到哪個欄位：
//
//	成本（或所得）= 數量 ÷ 匯率      ; Longint 除法
//	若結果為 0 → 取 1                ; 保底，買一點點也要付 1 塊
//	資源 ±= 數量
//	黃金 ∓= 成本
//
// 畫面上的話是「司令欲購買多少糧食？」「共用 N」／
// 「司令欲拋售多少糧食？」「共得 N」。

// 匯率：**多少單位的資源換 1 黃金**。全部出自 `@$brdiv` 前的 `mov cx, N`。
//
// 買賣同價——進口與出口用的是同一個數，**原版沒有價差**。
// 這一點對規則層很重要：資源可以無損地換來換去，只受上限與保底的影響。
const (
	TradeRateFood = 30 // 糧食：30 換 1 黃金
	TradeRateAmmo = 10 // 彈藥：10 換 1（最貴的一種）
	TradeRateFuel = 30 // 燃料：30 換 1
	TradeRateCoal = 20 // 煤礦：20 換 1，**只能賣**
	TradeRateIron = 20 // 鐵礦：20 換 1，**只能賣**
)

// TradeGood 是可交易的資源。
type TradeGood int

const (
	GoodFood TradeGood = iota
	GoodAmmo
	GoodFuel
	GoodCoal
	GoodIron
)

// tradeSpec 是每種資源的匯率與可交易方向。
var tradeSpec = map[TradeGood]struct {
	rate       int
	name       string
	importable bool
}{
	GoodFood: {TradeRateFood, "糧食", true},
	GoodAmmo: {TradeRateAmmo, "彈藥", true},
	GoodFuel: {TradeRateFuel, "燃料", true},
	// 煤礦與鐵礦**只能賣不能買**（進口選單裡沒有它們，`docs/re/27`）。
	// 它們是兵工廠的原料，只能靠本省產能——與「運補不能搬煤鐵」
	// 同一個設計（`40-economy.md` §9）。
	GoodCoal: {TradeRateCoal, "煤礦", false},
	GoodIron: {TradeRateIron, "鐵礦", false},
}

// TradePrice 回傳買賣 n 單位要花／能得多少黃金。
//
// **保底 1**：原版算完除法後檢查結果是不是 0，是的話取 1
// （`or ax, dx / jnz / mov var, 1`）。所以買 1 顆糧食也要付 1 塊，
// 賣 1 顆糧食也能拿 1 塊。
func TradePrice(good TradeGood, n int) int {
	spec, ok := tradeSpec[good]
	if !ok || n <= 0 {
		return 0
	}
	p := n / spec.rate
	if p == 0 {
		p = 1
	}
	return p
}

// resourceField 取某種資源在省份記錄裡的欄位指標。
func resourceField(p *Province, good TradeGood) *uint16 {
	switch good {
	case GoodFood:
		return &p.Food
	case GoodAmmo:
		return &p.Ammo
	case GoodFuel:
		return &p.Fuel
	case GoodCoal:
		return &p.Coal
	case GoodIron:
		return &p.Iron
	}
	return nil
}

// TradeResult 記錄一次買賣的結果。
type TradeResult struct {
	// Amount 是實際成交的資源量（可能因為上限而少於要求）。
	Amount int
	// Gold 是花掉（進口）或收到（出口）的黃金。
	Gold int
}

// Import 買進 n 單位的資源，扣黃金。
//
// ⚠️ 黃金不足時原版怎麼處理**還沒讀**——`sub_1FC1C` 那類函式有
// 「資金不足」的訊息，但進口這幾支的檢查點還沒對上。
// 這裡先直接擋下來並回錯，**標為 remake 行為**。
func (w *AIWorld) Import(p ProvinceID, good TradeGood, n int) (TradeResult, error) {
	spec, ok := tradeSpec[good]
	if !ok {
		return TradeResult{}, fmt.Errorf("game: 未知的商品 %d", good)
	}
	if !spec.importable {
		return TradeResult{}, fmt.Errorf("game: %s 不能進口，只能出口", spec.name)
	}
	prov, err := w.Table.At(p)
	if err != nil {
		return TradeResult{}, err
	}
	if n <= 0 {
		return TradeResult{}, nil
	}
	cost := TradePrice(good, n)
	if int(prov.Gold) < cost {
		return TradeResult{}, fmt.Errorf("game: 黃金 %d 不足 %d", prov.Gold, cost)
	}
	field := resourceField(prov, good)
	before := *field
	*field = AddResource(*field, uint16(n))
	prov.Gold -= uint16(cost)
	return TradeResult{Amount: int(*field - before), Gold: cost}, nil
}

// Export 賣出 n 單位的資源，收黃金。
//
// 資源不足時只賣手上有的——**這是 remake 行為**，原版的處理未讀。
func (w *AIWorld) Export(p ProvinceID, good TradeGood, n int) (TradeResult, error) {
	spec, ok := tradeSpec[good]
	if !ok {
		return TradeResult{}, fmt.Errorf("game: 未知的商品 %d", good)
	}
	prov, err := w.Table.At(p)
	if err != nil {
		return TradeResult{}, err
	}
	if n <= 0 {
		return TradeResult{}, nil
	}
	field := resourceField(prov, good)
	if n > int(*field) {
		n = int(*field)
	}
	if n == 0 {
		return TradeResult{}, fmt.Errorf("game: 沒有%s可賣", spec.name)
	}
	gain := TradePrice(good, n)
	*field -= uint16(n)
	before := prov.Gold
	prov.Gold = AddResource(prov.Gold, uint16(gain))
	return TradeResult{Amount: n, Gold: int(prov.Gold - before)}, nil
}

// TradeGoodName 回傳資源的原版用詞。
func TradeGoodName(good TradeGood) string {
	if spec, ok := tradeSpec[good]; ok {
		return spec.name
	}
	return "未知"
}

// Importable 回報某種資源能不能買。
func Importable(good TradeGood) bool { return tradeSpec[good].importable }
