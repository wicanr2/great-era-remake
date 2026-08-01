package game

import (
	"encoding/binary"
	"fmt"
)

// 依據：docs/spec/03-provinces.md（READY）

// ProvinceRecordSize 是一筆省份記錄的大小。
const ProvinceRecordSize = 37

// 兩個檔案裝同一個結構，只差陣列起點。
//
// [雷] .DT1 的省份記錄比 TOWN 晚 4 bytes。直接用 k×37 切 .DT1 會拿到橫跨兩筆
// 記錄的碎片，而且看起來還很像一筆（每筆尾端固定 19 19 19 19，切錯照樣對齊）。
const (
	// TownArrayOffset 是 TOWN(N).DAT 的省份陣列起點。
	TownArrayOffset = 0
	// SaveArrayOffset 是 SAVE(N).DT1 的省份陣列起點，前面 4 bytes 是檔頭。
	SaveArrayOffset = 4
)

// 已定名欄位在記錄內的位移。
const (
	provOffGold      = 0  // u16
	provOffFood      = 2  // u16
	provOffAmmo      = 4  // u16
	provOffCoal      = 6  // u16 [雷] 資料是煤、燃料，畫面印的是燃料、煤
	provOffFuel      = 8  // u16
	provOffIron      = 10 // u16
	provOffPop       = 12 // u32，單位是人；畫面顯示 ÷10000 為「萬」
	provOffCities    = 16 // u8
	provOffLandValue = 17 // u8
	provOffArsenals  = 18 // u8
	provOffLoyalty   = 19 // u8
	provOffCommander = 20 // u16，將領 ID（1-based），0 = 無主
	provOffNeighbour = 22 // 8 × u8
	provOffGovernor  = 30 // u16，將領 ID
	provOffFlags     = 32 // u8，位元旗標
	provNeighbourLen = 8
)

// 省份記錄 +32 的位元旗標。
//
// 初始檔（TOWN(N).DAT）全 39 省都是 0，存檔裡才會設。
const (
	// ProvinceFlagActed 是 bit 2：**這個省本回合已經處理過** — confirmed。
	//
	// ⛔ 這一格原本標「假說：本月已徵過稅」，依據是 `4.15` 詞條 42
	// 有「已徵過稅」而 `SAVE(1)` 剛好 25 省設了它。**程式碼推翻了它**
	// （`docs/re/14` §1）：
	//
	//	sub_10193（玩家指令選單）    or  al, 4    ← 玩家下完命令
	//	sub_13D23（電腦回合初始化）  or  al, 4    ← 電腦開始跑這個省
	//	sub_10AB5（主迴圈）          and al, 4    ← 立起來就跳過
	//	sub_383EF                    and al, 0FBh ← 對全 39 省清掉
	//
	// **它同時是玩家與電腦的分流開關**：主迴圈看到 bit 2 就跳過玩家選單。
	// 存檔裡有 25 省設它，是因為存檔當下那一輪已經輪過那些省；
	// 「無主的省一律沒設」也因此成立——無主省兩條路徑都不進。
	ProvinceFlagActed = 0x04

	// ProvinceFlagTaxed 是 bit 7：**本月已徵過稅** — confirmed。
	//
	// `sub_2C6C6`（指令 4 徵稅）的尾段 `or al, 80h`。
	//
	// `CONTEXT.md` §5.16 推翻「bit 2 = 已徵過稅」時寫過
	// 「那個詞條仍然存在，只是對應的旗標在別處，還沒找到」——**在這裡**。
	// `4.15` 詞條 42「已徵過稅」終於有了歸屬。
	ProvinceFlagTaxed = 0x80

	// ProvinceFlagInBattle 是 bit 6：**這個省正在打仗**。
	//
	// 兩處證據合起來就清楚了：
	//   - `sub_15925` 收集候選鄰省時跳過設了它的省（不能挑正在打的）
	//   - `sub_54DAC`（戰後結算）用 `&= 0xBF` **清除**它
	//
	// 所以它是戰鬥期間的暫時旗標，戰鬥一結束就清掉——這也解釋了
	// 為什麼兩份存檔裡都沒有出現過（存檔時不在戰鬥中）。
	ProvinceFlagInBattle = 0x40
)

// SeaBorder 是鄰省表裡表示「海洋或境外」的哨兵值。
//
// 帶這個值的省份全部沿海或臨國境（遼寧、山東、江蘇、浙江、福建、廣東、
// 臺灣、緬甸…），所以語意很明確。**但它與 WARPOS.DAT 用的 40 是不是同一個
// 意思沒有驗證過**（docs/spec/03 §4），不要互相代換。
const SeaBorder = 255

// GeneralID 是將領編號，1-based，指向 MAN{期}15 的名表。0 表示無主。
type GeneralID uint16

// Valid 回報這個 ID 是不是指向某位將領。
func (g GeneralID) Valid() bool { return g != 0 }

// Province 是一個省的狀態。
//
// 37 bytes 裡解出 14 個欄位，**未解的部分原樣保留在 Raw**——存檔寫回是
// 「改寫」不是「重建」（CLAUDE.md §9）。
type Province struct {
	// 六種資源。順序照資料，不是照畫面。
	Gold, Food, Ammo, Coal, Fuel, Iron uint16

	// Population 是人口（人）。畫面顯示的是這個值除以 10000 的「萬」。
	Population uint32

	Cities    uint8 // 城市數
	LandValue uint8 // 地價
	Arsenals  uint8 // 兵工廠數
	Loyalty   uint8 // 人民忠誠度，0-100

	// Commander 是司令，實際上是勢力領袖——TOWN(1) 解出來就是 1926 年的
	// 割據圖（張作霖轄東北九省+河北+山東、孫傳芳轄蘇皖浙閩贛…）。
	// Governor 是省長，存檔裡會與司令分化成實際駐守的將領。
	Commander, Governor GeneralID

	// Neighbours 是鄰省編號（1-based），已濾掉填充的 0 與 SeaBorder。
	Neighbours []ProvinceID
	// Coastal 表示鄰省表裡帶 SeaBorder，也就是這個省臨海或臨國境。
	Coastal bool

	// Flags 是 +32 的位元旗標，目前解出四個位元：
	//
	//	bit 1 (02h)  語意未解，只知道跨年會清掉（docs/re/25 §3）
	//	bit 2 (04h)  本回合已處理過（玩家/電腦分流開關，docs/re/14）
	//	bit 6 (40h)  這個省正在打仗（`sub_15925` 挑目標時跳過、
	//	             `sub_54DAC` 戰後清除、`sub_534FF` 挑增援來源時也跳過）
	//	bit 7 (80h)  本月已徵稅（docs/re/18）
	//
	// ⛔ bit 2 一度被記成「已徵稅」，那是錯的——徵稅旗標是 bit 7。
	Flags uint8

	// Raw 是完整的 37 bytes，寫回時以它為基底只蓋已解欄位。
	Raw [ProvinceRecordSize]byte
}

// PopulationWan 回傳畫面上顯示的「萬」，直接截斷小數。
//
// 實機湖北 18,250,000 顯示「1825 萬」。多數省份的人口是整千，
// 但臺灣（7,511,116）與海南島（4,028,520）不是——那兩筆的尾數是資料本來就有的，
// 不是解析錯誤（高位 byte 是 0，沒有溢位）。
func (p *Province) PopulationWan() uint32 { return p.Population / 10000 }

// ParseProvince 解一筆 37 bytes 的省份記錄。
func ParseProvince(rec []byte) (Province, error) {
	var p Province
	if len(rec) < ProvinceRecordSize {
		return p, fmt.Errorf("game: 省份記錄需要 %d bytes，只有 %d",
			ProvinceRecordSize, len(rec))
	}
	copy(p.Raw[:], rec)
	u16 := func(off int) uint16 { return binary.LittleEndian.Uint16(rec[off:]) }

	p.Gold, p.Food, p.Ammo = u16(provOffGold), u16(provOffFood), u16(provOffAmmo)
	p.Coal, p.Fuel, p.Iron = u16(provOffCoal), u16(provOffFuel), u16(provOffIron)
	p.Population = binary.LittleEndian.Uint32(rec[provOffPop:])
	p.Cities = rec[provOffCities]
	p.LandValue = rec[provOffLandValue]
	p.Arsenals = rec[provOffArsenals]
	p.Loyalty = rec[provOffLoyalty]
	p.Commander = GeneralID(u16(provOffCommander))
	p.Governor = GeneralID(u16(provOffGovernor))
	p.Flags = rec[provOffFlags]

	for _, b := range rec[provOffNeighbour : provOffNeighbour+provNeighbourLen] {
		switch {
		case b == 0: // 填充
		case b == SeaBorder:
			p.Coastal = true
		default:
			p.Neighbours = append(p.Neighbours, ProvinceID(b))
		}
	}
	return p, nil
}

// Bytes 產生寫回用的 37 bytes：以原始 bytes 為基底，只蓋已解欄位。
//
// 鄰省表（+22）刻意不寫回——它是靜態地理資料，遊戲中不會變，
// 而且 Neighbours 已經濾掉了填充與 SeaBorder，重建會遺失原本的排列。
func (p *Province) Bytes() [ProvinceRecordSize]byte {
	out := p.Raw
	put := func(off int, v uint16) { binary.LittleEndian.PutUint16(out[off:], v) }

	put(provOffGold, p.Gold)
	put(provOffFood, p.Food)
	put(provOffAmmo, p.Ammo)
	put(provOffCoal, p.Coal)
	put(provOffFuel, p.Fuel)
	put(provOffIron, p.Iron)
	binary.LittleEndian.PutUint32(out[provOffPop:], p.Population)
	out[provOffCities] = p.Cities
	out[provOffLandValue] = p.LandValue
	out[provOffArsenals] = p.Arsenals
	out[provOffLoyalty] = p.Loyalty
	put(provOffCommander, uint16(p.Commander))
	put(provOffGovernor, uint16(p.Governor))
	out[provOffFlags] = p.Flags
	return out
}

// GameDate 是存檔的遊戲內日期。
type GameDate struct {
	Year  uint16 // 民國紀年
	Month uint8
	// Unknown3 是檔頭第 4 個 byte，兩份存檔都是 1。語意未解，原樣保留。
	Unknown3 uint8
}

// ProvinceTable 是 39 個省的狀態。
type ProvinceTable struct {
	// Date 只在從 .DT1 解出來時有值；TOWN(N).DAT 沒有檔頭。
	Date     *GameDate
	Province [ProvinceCount]Province
}

// At 取第 id 個省（1-based）。
func (t *ProvinceTable) At(id ProvinceID) (*Province, error) {
	if !id.Valid() {
		return nil, fmt.Errorf("game: 省編號 %d 超出 1..%d", id, ProvinceCount)
	}
	return &t.Province[id-1], nil
}

// ParseTownFile 解 TOWN(N).DAT：39 筆記錄，沒有檔頭。
func ParseTownFile(data []byte) (*ProvinceTable, error) {
	if want := ProvinceCount * ProvinceRecordSize; len(data) != want {
		return nil, fmt.Errorf("game: TOWN 檔應為 %d bytes（%d 省 × %d），實得 %d",
			want, ProvinceCount, ProvinceRecordSize, len(data))
	}
	return parseProvinceArray(data, TownArrayOffset, nil)
}

// ParseSaveProvinces 解 SAVE(N).DT1 的前段：4 bytes 檔頭 + 39 筆記錄。
//
// .DT1 後面還有 13,000 多 bytes（將領區與未解區域），這裡只碰前段。
func ParseSaveProvinces(data []byte) (*ProvinceTable, error) {
	need := SaveArrayOffset + ProvinceCount*ProvinceRecordSize
	if len(data) < need {
		return nil, fmt.Errorf("game: .DT1 前段需要 %d bytes，只有 %d", need, len(data))
	}
	d := &GameDate{
		Year:     binary.LittleEndian.Uint16(data[0:]),
		Month:    data[2],
		Unknown3: data[3],
	}
	return parseProvinceArray(data, SaveArrayOffset, d)
}

func parseProvinceArray(data []byte, base int, d *GameDate) (*ProvinceTable, error) {
	t := &ProvinceTable{Date: d}
	for i := 0; i < ProvinceCount; i++ {
		p, err := ParseProvince(data[base+i*ProvinceRecordSize:])
		if err != nil {
			return nil, fmt.Errorf("game: 第 %d 省: %w", i+1, err)
		}
		t.Province[i] = p
	}
	return t, nil
}

// Factions 依司令把省份分組，回傳每位勢力領袖轄下的省編號。
//
// 無主的省（Commander == 0）不列入。
func (t *ProvinceTable) Factions() map[GeneralID][]ProvinceID {
	out := make(map[GeneralID][]ProvinceID)
	for i := range t.Province {
		if c := t.Province[i].Commander; c.Valid() {
			out[c] = append(out[c], ProvinceID(i+1))
		}
	}
	return out
}

// AttackableFrom 回傳從 from 出發可攻打的鄰省，順序照記錄裡的鄰省表。
//
// 規則是從 WAR.EXE 的 sub_5B7DC 讀出來的（docs/mechanics/20-military.md）：
//
//	for j = 1 to 8:
//	    n = 記錄[省].鄰省[j]
//	    if n == 0 or n == 0xFF:            continue   ; 填充／海洋
//	    if 記錄[n].司令 == 0:               continue   ; 無主的省
//	    if 記錄[n].司令 == 記錄[省].司令:    continue   ; 同一勢力
//	    → 可攻打
//
// [雷] 判準是**司令不同且非無主**，不是「不在我控制的清單裡」。
// 無主的省（司令 == 0）不能攻打——這一條 SPEC-01 §2 原本沒有，
// 是實機樣本剛好都沒碰到（湖北與河南的鄰省全部有主）。
//
// [雷] 用的是**省份記錄裡的鄰省表**（+22），不是 WARPOS 的幾何鄰接。
// 兩者有 3 筆差異（隔海的那組，docs/spec/03 §4），程式走的是記錄。
func (t *ProvinceTable) AttackableFrom(from ProvinceID) ([]ProvinceID, error) {
	src, err := t.At(from)
	if err != nil {
		return nil, err
	}
	var out []ProvinceID
	for _, n := range src.Neighbours {
		tgt, err := t.At(n)
		if err != nil {
			continue // 鄰省表裡的值超出範圍，跳過
		}
		if !tgt.Commander.Valid() || tgt.Commander == src.Commander {
			continue
		}
		out = append(out, n)
	}
	return out, nil
}

// FirstAttackable 回傳第一個可攻打的鄰省，沒有就回 0。
//
// 這正是 sub_5B7DC 的回傳值。`sub_15F3C` 用它分流：
// 非 0 走「有敵鄰省」的流程，0 走另一條（docs/mechanics/70-ai.md）。
func (t *ProvinceTable) FirstAttackable(from ProvinceID) ProvinceID {
	ns, err := t.AttackableFrom(from)
	if err != nil || len(ns) == 0 {
		return 0
	}
	return ns[0]
}

// Acted 回報這個省本回合是否已經處理過（見 ProvinceFlagActed）。
func (p *Province) Acted() bool { return p.Flags&ProvinceFlagActed != 0 }

// Taxed 回報這個省本月是否已徵過稅（見 ProvinceFlagTaxed）。
func (p *Province) Taxed() bool { return p.Flags&ProvinceFlagTaxed != 0 }

// InBattle 回報這個省是不是正在打仗（bit 6）。
func (p *Province) InBattle() bool { return p.Flags&ProvinceFlagInBattle != 0 }

// Capture 套用戰後的佔領結算，順序照 sub_54DAC：
//
//	記錄[省].+32 &= 0xBF     ; 清除「正在打仗」
//	記錄[省].+20 = 勝方       ; 司令改成勝方
//
// 勝方為 0 時只清旗標、不改司令——原版那一步有前置條件
// （`[arg_0+8] == word_64942`），條件不成立時省份易主不會發生。
func (p *Province) Capture(winner GeneralID) {
	p.Flags &^= ProvinceFlagInBattle
	if winner.Valid() {
		p.Commander = winner
	}
}

// AvailableGenerals 數某省有幾個「效忠該省司令且可用」的將領，
// 語意照 `sub_5A881`（`docs/re/07` §9）：
//
//	for 每個將領:
//	    if 將領.所屬省 != 省:          continue
//	    if !(將領.+16 & 1):            continue   ← 不可用
//	    if 將領.效忠勢力 != 省.司令:    continue
//	    count++
//
// 三個條件缺一不可——**所屬省對了但效忠別人的將領不算**，
// 這是易主之後會出現的情況。
func (t *ProvinceTable) AvailableGenerals(p ProvinceID, units []CombatUnit) (int, error) {
	prov, err := t.At(p)
	if err != nil {
		return 0, err
	}
	n := 0
	for i := range units {
		u := &units[i]
		if u.General == 0 || u.Province != p || !u.Active {
			continue
		}
		if u.Faction != prov.Commander {
			continue
		}
		n++
	}
	return n, nil
}

// ReinforcementSources 收集可以對 defender 省提供增援的鄰省，
// 語意照 `sub_534FF`（`docs/re/07` §9）：
//
//	for k = 1..7:                       ← 鄰省表有 8 格，原版只掃 7 格
//	    省 = 鄰省表[k]
//	    if 省 == 0 或 255:               continue   ← 填充／海洋境外
//	    if 省.司令 == 0:                 停止掃描    ← 無主省直接收工
//	    if 省.司令 != 單位.效忠勢力:      continue
//	    if 省.+32 & 40h:                 continue   ← 那個省正在打仗
//	    if 該省可用將領數 >= 100:         continue
//	    收下
//
// ⚠️ 那道 `< 100` 的門檻**實質恆真**——第一期全遊戲才 274 位將領，
// 一個省不可能有 100 位。它比較像是上限保護，不是真正的條件。
// 照抄是為了忠實，不是因為它會生效。
func (t *ProvinceTable) ReinforcementSources(defender ProvinceID, faction GeneralID, units []CombatUnit) ([]ProvinceID, error) {
	prov, err := t.At(defender)
	if err != nil {
		return nil, err
	}
	var out []ProvinceID
	// 讀 Raw 而不是 Neighbours——原版對**原始排列**敏感（遇到無主省就
	// 停止掃描），而 Neighbours 已經濾掉填充與 SeaBorder，順序不一樣。
	for k := 0; k < 7; k++ { // 原版只掃 7 格，第 8 格不看
		b := prov.Raw[provOffNeighbour+k]
		if b == 0 || b == SeaBorder {
			continue
		}
		n := ProvinceID(b)
		np, err := t.At(n)
		if err != nil {
			continue
		}
		if np.Commander == 0 {
			break // 原版遇到無主省就結束掃描，不是 continue
		}
		if np.Commander != faction {
			continue
		}
		if np.Flags&ProvinceFlagInBattle != 0 {
			continue
		}
		cnt, err := t.AvailableGenerals(n, units)
		if err != nil || cnt >= 100 {
			continue
		}
		out = append(out, n)
	}
	return out, nil
}
