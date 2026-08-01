package game

import (
	"encoding/binary"
	"fmt"
)

// 依據：docs/spec/02-generals.md（READY）

// GeneralRecordSize 是一筆將領記錄的大小。
//
// MAN(N).DAT 與存檔 .DT1 的將領區用同一個結構。
// MAN(1).DAT 的 9,042 ÷ 274 人 = 33 整除。
const GeneralRecordSize = 33

// 已定名欄位在記錄內的位移。
//
// **這些偏移與執行期單位記錄（基址 0x7A7D）完全相同**——`MAN(N).DAT`
// 的 33 bytes 是原地載入的，不是另一套佈局。三個哨兵值印證了這件事
// （`docs/spec/02` §3）：
//
//	+5  檔案裡 274 筆**全部是 255**  ← 執行期的「不在場上」
//	+8  全部是 0                    ← 攻守旗標，檔案裡都是守方
//	+16 只有 {0, 1, 32}             ← bit 0 = 可行動
const (
	genOffAbilityA = 0
	genOffAbilityB = 1
	genOffAbilityC = 2
	genOffProvince = 4  // u8，所屬省編號（1-based），0 = 無所屬
	genOffForce    = 17 // u16 little-endian

	// 番號四欄，全部是 `FAN(1).15` 的**槽位索引（1-based）**，0 = 沒有。
	// 名字來自實機「查閱將領」的「番號」列（`docs/playtest/08`），
	// 對應關係逐格驗過（`docs/spec/02` §番號）。
	//
	//	討賊軍第25師 = 勢力名[4] + 前綴[10] + 25 + 後綴[18]
	//	              = 討賊軍   + 第       + 25 + 師
	genOffTitlePrefix = 25 // 「第」「暫」「騎」「砲」「新」…
	genOffTitleNumber = 26 // 數字，0 = 沒有編號
	genOffTitleSuffix = 27 // 「師」「團」「旅」「軍」…
	genOffFactionName = 28 // 勢力名：1 國民革命軍、4 討賊軍、5 五省聯軍…
	genOffF19         = 19 // 戰力公式用，語意未解
	genOffF20         = 20 // 戰力公式用，只有 5 個等級值
	genOffBranch      = 21 // 兵種，值域 {1, 4, 5, 6}
	// genOffStamina 是**體力**（0..100）。
	//
	// 證據是畫面文字（`docs/re/27`）：`sub_241D0`（墾地）與 `sub_24535`（挖金礦）
	// 檢查這一格不足時，印的正是「**士兵體力不足**」。慰勞軍民 `sub_3412B`
	// 也對它 +10 並夾到 100，與畫面「體力」對得上（`docs/re/22`）。
	//
	// 它同時參與戰力公式（`sub_5A0B9`），所以舊名 `F29` 出現在那裡。
	genOffStamina = 29
	genOffF30     = 30 // 戰力公式用，也是每回合衰減 20% 的那一格
	genOffRange   = 31 // 遠程攻擊的參數，第一期 274 筆全是 1
)

// 兵種。**這四個名字是有證據的**（`docs/spec/02` §4）：
// 每個兵種的兵力上限與社群傳的滿員數完全一致，三期零例外。
const (
	BranchInfantry = Branch1 // 1，滿員 20,000
	BranchArtiller = Branch4 // 4，滿員 2,000。遠程兵種（docs/re/09）
	BranchArmour   = Branch5 // 5，滿員 200。全遊戲只有一個
	BranchCavalry  = Branch6 // 6，滿員 10,000
)

// branchFullStrength 是各兵種的**滿員數**。
//
// 這是**社群資料第一次被原版資料證實**：`CLAUDE.md` §1.5 記的
// 20000／10000／2000／200 全部對上第一期 `MAN(1).DAT` 的最大值，
// 而且四個一起對上，不是巧合。
//
// ⚠️ **這不是硬性上限。** 第三期的騎兵有一筆 12,000，超過 10,000。
// 三期的編制不同：
//
//	期別   步兵          砲兵     裝甲兵   騎兵
//	 1     500..20000   1000..2000  200   500..10000
//	 2    9501..20000   2000..2000   —    5000..7000
//	 3    9500..12000   2000..2000  100   5000..12000
//
// 徵兵／整編會不會夾到滿員數，**還沒在程式碼裡找到**。
// 在找到之前，這張表只當「第一期的滿員數」用，不要拿來當驗證條件。
var branchFullStrength = map[uint8]uint16{
	BranchInfantry: 20000,
	BranchArtiller: 2000,
	BranchArmour:   200,
	BranchCavalry:  10000,
}

// BranchFullStrength 回傳兵種的滿員數（第一期），未知兵種回 0。
//
// ⚠️ 不是硬性上限，見上。
func BranchFullStrength(branch uint8) uint16 { return branchFullStrength[branch] }

// BranchName 回傳兵種名。
func BranchName(branch uint8) string {
	switch branch {
	case BranchInfantry:
		return "步兵"
	case BranchArtiller:
		return "砲兵"
	case BranchArmour:
		return "裝甲兵"
	case BranchCavalry:
		return "騎兵"
	}
	return "未知"
}

// General 是一位將領。
//
// 33 bytes 裡目前解出五個欄位，**其餘 30 個 byte 原樣保留在 Raw**——
// 存檔寫回是「改寫」不是「重建」，未解區域一個 byte 都不能動
// （CLAUDE.md §9）。
type General struct {
	// 三個能力值，**名字全部來自實機「查閱將領」畫面**
	// （`docs/playtest/08` §2，五個將領逐格對照）：
	//
	//	AbilityA (+0) = 帶兵能力    進攻擊力公式與名將加成（docs/re/08 §4）
	//	AbilityB (+1) = 忠誠度      司令本人顯示「---」，資料裡是 100
	//	AbilityC (+2) = 政治手腕    「開發」指令派將時列的就是這一欄
	//
	// ⛔ `AbilityC` 一度被定名為「開發能力」——那是從「開發指令列出它」
	// 反推的**用途**，不是它的名字。「查閱將領」畫面直接寫著「政治手腕」。
	//
	// **不知道哪個是統率、武力、政治或忠誠**（docs/spec/02 §1），
	// 所以刻意不做語意化命名。蔣中正三項全 100 可當 sanity check。
	AbilityA, AbilityB, AbilityC uint8

	// Force 是兵力。滿員數依兵種而定，見 BranchFullStrength——
	// 四個數字（20000／10000／2000／200）在第一期全部對上，
	// 但**不是硬性上限**（第三期的騎兵有一筆 12,000）。
	Force uint16

	// TitlePrefix / TitleNumber / TitleSuffix / FactionName 組成部隊番號，
	// 前後綴與勢力名都是 `FAN(1).15` 的槽位索引（1-based，0 = 無）。
	TitlePrefix, TitleNumber, TitleSuffix, FactionName uint8

	// Branch 是兵種（`+21`），值域 {1, 4, 5, 6}。
	Branch uint8

	// Stamina 是**體力**（`+29`，0..100）。開發指令的門檻就是它
	// （墾地 ≥5、挖金礦 ≥20），不足時原版印「士兵體力不足」。
	// 它同時是戰力公式的一項。
	Stamina uint8

	// 戰力公式（`sub_5A0B9`，`docs/re/08` §4d）要用的另外三個欄位。
	// **語意未解**——公式用得到，但還沒對上畫面上的名字。
	// 候選見 `docs/mechanics/60-personnel.md` §6（經驗／士兵攻擊力／
	// 武裝程度／士兵戰技）。⚠️ 那是名字的集合不是順序，別對號入座。
	F19, F20, F30 uint8

	// Range 是遠程攻擊的參數（`+31`，`docs/re/09` §1）。
	// 第一期 274 筆全部是 1，所以看不出值域。
	Range uint8

	// Province 是所屬省編號（1-based），0 表示無所屬。
	//
	// 驗證方式是「十大勢力領袖的所屬省必須落在自己的轄區內」——
	// offset 4 對 10 個勢力**全部命中**。接著兩個省的將領數與兵力加總
	// 都對上實機：湖北 15 位／97,500、河南 25 位／140,500，零誤差。
	Province ProvinceID

	// Raw 是完整的 33 bytes，寫回時以它為基底只蓋已解欄位。
	Raw [GeneralRecordSize]byte
}

// ParseGeneral 解一筆 33 bytes 的將領記錄。
func ParseGeneral(rec []byte) (General, error) {
	var g General
	if len(rec) < GeneralRecordSize {
		return g, fmt.Errorf("game: 將領記錄需要 %d bytes，只有 %d",
			GeneralRecordSize, len(rec))
	}
	copy(g.Raw[:], rec)
	g.Branch = rec[genOffBranch]
	g.F19, g.F20 = rec[genOffF19], rec[genOffF20]
	g.Stamina, g.F30 = rec[genOffStamina], rec[genOffF30]
	g.Range = rec[genOffRange]
	g.AbilityA = rec[genOffAbilityA]
	g.AbilityB = rec[genOffAbilityB]
	g.AbilityC = rec[genOffAbilityC]
	g.Province = ProvinceID(rec[genOffProvince])
	g.Force = binary.LittleEndian.Uint16(rec[genOffForce:])
	g.TitlePrefix = rec[genOffTitlePrefix]
	g.TitleNumber = rec[genOffTitleNumber]
	g.TitleSuffix = rec[genOffTitleSuffix]
	g.FactionName = rec[genOffFactionName]
	return g, nil
}

// Bytes 產生寫回用的 33 bytes：以原始 bytes 為基底，只蓋已解欄位。
//
// 這是 CLAUDE.md §9 要求的「改寫而非重建」——未解的 27 個 byte 原樣帶過去，
// 驗收標準是 byte-for-byte round-trip。
func (g *General) Bytes() [GeneralRecordSize]byte {
	out := g.Raw
	out[genOffAbilityA] = g.AbilityA
	out[genOffAbilityB] = g.AbilityB
	out[genOffAbilityC] = g.AbilityC
	out[genOffProvince] = byte(g.Province)
	binary.LittleEndian.PutUint16(out[genOffForce:], g.Force)
	out[genOffTitlePrefix] = g.TitlePrefix
	out[genOffTitleNumber] = g.TitleNumber
	out[genOffTitleSuffix] = g.TitleSuffix
	out[genOffFactionName] = g.FactionName
	return out
}

// ParseGenerals 解一整段連續的將領記錄。
func ParseGenerals(data []byte, count int) ([]General, error) {
	if count < 0 {
		return nil, fmt.Errorf("game: 將領數不能為負（%d）", count)
	}
	need := count * GeneralRecordSize
	if len(data) < need {
		return nil, fmt.Errorf("game: %d 位將領需要 %d bytes，只有 %d",
			count, need, len(data))
	}
	out := make([]General, count)
	for i := 0; i < count; i++ {
		g, err := ParseGeneral(data[i*GeneralRecordSize:])
		if err != nil {
			return nil, err
		}
		out[i] = g
	}
	return out, nil
}

// TotalForce 是一組將領的兵力總和。
//
// 政略畫面的「兵力」就是這個值，**不是存檔裡的單一欄位**——
// 這也解釋了它為什麼會超過 u16（湖北 97,500）。實機驗證見 docs/spec/02 §2。
//
// 回傳 uint32，因為總和會超過 u16。
func TotalForce(gs []General) uint32 {
	var sum uint32
	for i := range gs {
		sum += uint32(gs[i].Force)
	}
	return sum
}

// GeneralsOf 挑出屬於某個省的將領，保留原本的順序。
//
// 順序就是實機派將列表的順序——湖北的第一位是吳佩孚（20000），
// 與截圖一致（docs/playtest/02 §5A.4）。
func GeneralsOf(gs []General, p ProvinceID) []General {
	var out []General
	for i := range gs {
		if gs[i].Province == p {
			out = append(out, gs[i])
		}
	}
	return out
}

// ForceOf 回傳某個省的兵力，也就是該省將領的兵力加總。
//
// 這就是政略畫面「兵力」欄位的值——**不是存檔裡的單一欄位**，
// 所以會超過 u16（湖北 97,500）。
func ForceOf(gs []General, p ProvinceID) uint32 {
	var sum uint32
	for i := range gs {
		if gs[i].Province == p {
			sum += uint32(gs[i].Force)
		}
	}
	return sum
}

// CountOf 回傳某個省的將領數，也就是畫面上的「將領數」欄位。
func CountOf(gs []General, p ProvinceID) int {
	n := 0
	for i := range gs {
		if gs[i].Province == p {
			n++
		}
	}
	return n
}

// GeneralNameSlotWidth 是 MAN{N}15 名表的槽寬：每個人名佔 3 個字模格。
//
// 將領筆數要用名表算（`len(Glyphs) / 3`），**不能用「檔案大小 ÷ 33」**——
// 三個 MAN(N).DAT 都是 9,042 B，但第二三期只有 106 位，其餘是殘料。
const GeneralNameSlotWidth = 3
