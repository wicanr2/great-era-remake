package game

import "fmt"

// 授權自治（政略指令 15「其他選項」底下的一項），出自 `sub_22E25`（720 行）
// 與 `sub_23FF6`（選單，211 行）。
//
// 畫面上的話：「司令不在本省／司令無其他省份須授權自治／司令〈某〉？」

// ProvinceFlagAutonomous 是省份 `+32` 的 **bit 5**：這個省已授權自治。
//
//	sub_22E25：  旗標 & 20h == 20h → 旗標 ^= 20h   （取消自治）
//	             否則              → 旗標 |= 20h   （設為自治）
//
// 是個 **toggle**，同一個指令開也關。
const ProvinceFlagAutonomous = 0x20

// AutonomousCommandCap 是自治省的指令數上限。
//
// `sub_10193`（玩家的指令迴圈）每執行完一個指令就：
//
//	dec  剩餘指令數
//	若 省份 +32 & 20h == 20h → 剩餘指令數 = 0
//
// 也就是**自治的省一個月只能下一個指令**——下完就跳到下一個省。
// 這就是「自治」在玩法上的代價：省還是你的，但你管不動它。
const AutonomousCommandCap = 1

// Autonomous 回報這個省是不是已授權自治。
func (p *Province) Autonomous() bool {
	return p.Flags&ProvinceFlagAutonomous != 0
}

// ToggleAutonomy 切換某省的自治狀態，回傳切換後的值。
//
// 原版是 toggle 沒有分開的「設定／取消」，所以這裡照抄。
func (w *AIWorld) ToggleAutonomy(p ProvinceID) (bool, error) {
	prov, err := w.Table.At(p)
	if err != nil {
		return false, err
	}
	if prov.Commander == 0 {
		return false, fmt.Errorf("game: 省 %d 無主，談不上自治", p)
	}
	prov.Flags ^= ProvinceFlagAutonomous
	return prov.Autonomous(), nil
}

// CommandsRemainingAfter 算某省執行完一個指令之後還剩幾個指令數。
//
// 照 `sub_10193`：先減一，自治的省再歸零。
func (w *AIWorld) CommandsRemainingAfter(p ProvinceID, before int) int {
	n := before - 1
	if n < 0 {
		n = 0
	}
	prov, err := w.Table.At(p)
	if err == nil && prov.Autonomous() {
		return 0
	}
	return n
}
