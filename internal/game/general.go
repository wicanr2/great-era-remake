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
	// genOffExperience 是**經驗**（`docs/playtest/08` 的「經驗」列）。
	//
	// ⚠️ **在 `MAN(N).DAT` 裡 274 筆全是 0**，所以拿劇本檔怎麼掃都找不到它
	// ——實機那張畫面讀的是 `SAVE(1)`。改掃存檔的將領區一次命中
	// （吳佩孚 30、其餘四人 0，與畫面逐格相同）。
	//
	// 這也說明它是**遊戲中累積**的量，開局一律 0。累積規則未解。
	genOffExperience = 3
	genOffProvince   = 4  // u8，所屬省編號（1-based），0 = 無所屬
	genOffForce      = 17 // u16 little-endian

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
	// genOffF19 是**士兵戰技**、genOffF20 是**武裝程度**——
	// 兩個都由實機「查閱將領」畫面逐格定名（`docs/playtest/08` §2），
	// 名字保留 `F` 前綴只是因為戰力公式那邊沿用了舊名。
	genOffF19    = 19 // 士兵戰技（10..226）
	genOffF20    = 20 // 武裝程度，只有 5 個等級值 {10,15,40,50,100}
	genOffBranch = 21 // 兵種，值域 {1, 4, 5, 6}
	// genOffStamina 是**體力**（0..100）。
	//
	// 證據是畫面文字（`docs/re/27`）：`sub_241D0`（墾地）與 `sub_24535`（挖金礦）
	// 檢查這一格不足時，印的正是「**士兵體力不足**」。慰勞軍民 `sub_3412B`
	// 也對它 +10 並夾到 100，與畫面「體力」對得上（`docs/re/22`）。
	//
	// 它同時參與戰力公式（`sub_5A0B9`），所以舊名 `F29` 出現在那裡。
	genOffStamina = 29
	// genOffF30 是**士氣**（`docs/playtest/08` §4：存檔 65/59/36/47/25
	// 與畫面的「士氣」列逐格相同）。它同時是每回合衰減 20% 的那一格。
	genOffF30   = 30
	genOffRange = 31 // 遠程攻擊的參數，第一期 274 筆全是 1
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
// 33 bytes 裡已解欄位包含三項能力、經驗、所屬省、兵力、戰技、武裝、兵種、
// 番號四欄、體力、士氣與射程；其餘 byte 原樣保留在 Raw。存檔寫回是「改寫」
// 不是「重建」，未解區域一個 byte 都不能動（CLAUDE.md §9）。
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
	// **三個都已定名**（`docs/playtest/08`，實機逐格對照）：
	//
	//	F19 = 士兵戰技    F20 = 武裝程度    F30 = 士氣
	//
	// 保留 `F` 前綴是因為戰力公式那邊沿用舊名，不是因為還沒解。
	F19, F20, F30 uint8

	// Experience 是**經驗**（`+3`）。`MAN(N).DAT` 裡一律 0，
	// 只有存檔的將領區才有值——它是遊戲中累積的量，累積規則未解。
	Experience uint8

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
	g.Experience = rec[genOffExperience]
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
// 這是 CLAUDE.md §9 要求的「改寫而非重建」——未解 byte 原樣帶過去，
// 驗收標準是 byte-for-byte round-trip。
func (g *General) Bytes() [GeneralRecordSize]byte {
	out := g.Raw
	out[genOffAbilityA] = g.AbilityA
	out[genOffAbilityB] = g.AbilityB
	out[genOffAbilityC] = g.AbilityC
	out[genOffExperience] = g.Experience
	out[genOffProvince] = byte(g.Province)
	binary.LittleEndian.PutUint16(out[genOffForce:], g.Force)
	out[genOffF19] = g.F19
	out[genOffF20] = g.F20
	out[genOffBranch] = g.Branch
	out[genOffTitlePrefix] = g.TitlePrefix
	out[genOffTitleNumber] = g.TitleNumber
	out[genOffTitleSuffix] = g.TitleSuffix
	out[genOffFactionName] = g.FactionName
	out[genOffStamina] = g.Stamina
	out[genOffF30] = g.F30
	out[genOffRange] = g.Range
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

// SaveGeneralsOffset 是 `SAVE(N).DT1` 裡**將領區的起點**：5203。
//
// 這一格先前標「至今沒找到」（`docs/spec/02` 附錄）。找法很直接：
// 拿 `MAN(1).DAT` 的 `+21`（兵種）前 60 筆當指紋，在 `.DT1` 裡滑動比對，
// 只有 5203 這一處全中。
//
// ⭐ 有一個獨立的正對照：`general_test.go` 早就在用 **7084** 當
// 「湖北第一位將領（吳佩孚，第 58 位）」的位置，而
//
//	5203 + 57 × 33 = 7084
//
// 兩者對得上，所以 5203 不是湊出來的。
//
// ⚠️ 省份區到將領區之間還有 3,756 bytes 未解（1447..5202）。
const SaveGeneralsOffset = 5203

// ParseSaveGenerals 解 `SAVE(N).DT1` 的將領區。
//
// `count` 由劇本決定（原版 `word_6BC4A`，第一期 274）。
// 與 `MAN(N).DAT` 是同一個 33 bytes 佈局——差別只在存檔的欄位是
// **遊戲中的當下值**：經驗、體力、士氣都會變。
func ParseSaveGenerals(data []byte, count int) ([]General, error) {
	if len(data) < SaveGeneralsOffset {
		return nil, fmt.Errorf("game: .DT1 只有 %d bytes，放不下將領區（起點 %d）",
			len(data), SaveGeneralsOffset)
	}
	return ParseGenerals(data[SaveGeneralsOffset:], count)
}
