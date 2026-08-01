package game

// 資源運算。依據：docs/mechanics/40-economy.md

// ResourceCap 是資源的上限，來自 WAR.EXE 的 `sub_5A467`——
// 那支是全遊戲通用的「累加並夾上限」，被 56 處呼叫，每次都傳 0xEA60。
//
// 資源欄位本身是 u16（能裝到 65535），但遊戲用 60000 當實際上限。
// 五個資料檔的所有資源值都在其下，最大的是彈藥 53,692——逼近但未超過。
const ResourceCap = 60000

// AddResource 把 delta 加到 cur，夾到 ResourceCap。
//
// 原版 `sub_5A467(上限, 加數, 目標)`：
//
//	v = *目標 + 加數        ; 32-bit 運算，不會中途溢位
//	if v > 上限: v = 上限
//	*目標 = v
//
// 用 uint32 中間值重現「先算完再夾」的行為——直接用 uint16 相加會先溢位，
// 結果完全不同（例如 50000 + 20000 在 u16 是 4464，夾完還是 4464，
// 但原版會給 60000）。
func AddResource(cur, delta uint16) uint16 {
	v := uint32(cur) + uint32(delta)
	if v > ResourceCap {
		return ResourceCap
	}
	return uint16(v)
}

// MergeResourcesAfterBattle 套用戰後的省份資源結算。
//
// 原版 `sub_546D1`（九步結算的第三步）：先把省份記錄的黃金、糧食、彈藥、
// 燃料歸零，再把雙方各自帶的量相加寫回，每項夾到 60000。
//
//	黃金 = min(A.黃金 + B.黃金, 60000)
//	糧食 = min(A.糧食 + B.糧食, 60000)
//	彈藥 = min(A.彈藥 + B.彈藥, 60000)
//	燃料 = min(A.燃料 + B.燃料, 60000)
//
// **煤礦與鐵礦不參與**——原版只歸零了 `+0`／`+2`／`+4`／`+8`，
// 沒有碰 `+6`（煤礦）與 `+10`（鐵礦）。那兩者是省份的礦藏，
// 不隨部隊移動，戰後維持原值。
func (p *Province) MergeResourcesAfterBattle(a, b BattleResources) {
	p.Gold = AddResource(0, AddResource(a.Gold, b.Gold))
	p.Food = AddResource(0, AddResource(a.Food, b.Food))
	p.Ammo = AddResource(0, AddResource(a.Ammo, b.Ammo))
	p.Fuel = AddResource(0, AddResource(a.Fuel, b.Fuel))
	// Coal 與 Iron 刻意不動。
}

// BattleResources 是一方在戰鬥中帶著的資源。
//
// 對應 `.DT2` 記錄開頭的四個 u16（`docs/re/05` §2）——載入後被搬到
// `word_64932`／`word_64936`／`word_6493A`／`word_6493E`，
// 而每個變數後面緊接著另一方的同名欄位（所以目標位址間隔是 4 不是 2）。
type BattleResources struct {
	Gold, Food, Ammo, Fuel uint16
}
