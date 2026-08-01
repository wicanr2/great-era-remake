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
	provNeighbourLen = 8
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
