// Package actions 定義與輸入裝置無關的玩家動作與命中區。
// 鍵盤、桌面滑鼠與 Android 觸控都必須先轉成 Action，才能進入 app。
package actions

import (
	"strconv"
	"strings"
)

type Action string

const (
	None    Action = ""
	Back    Action = "common.back"
	Confirm Action = "common.confirm"
	Cancel  Action = "common.cancel"

	OpenCommands Action = "command.open"
	PreviousPage Action = "page.previous"
	NextPage     Action = "page.next"
	DeleteDigit  Action = "input.delete"
	Submit       Action = "input.submit"
	BattleAttack Action = "battle.attack"

	Digit0 Action = "input.digit.0"
	Digit1 Action = "input.digit.1"
	Digit2 Action = "input.digit.2"
	Digit3 Action = "input.digit.3"
	Digit4 Action = "input.digit.4"
	Digit5 Action = "input.digit.5"
	Digit6 Action = "input.digit.6"
	Digit7 Action = "input.digit.7"
	Digit8 Action = "input.digit.8"
	Digit9 Action = "input.digit.9"

	Select1  Action = "select.1"
	Select2  Action = "select.2"
	Select3  Action = "select.3"
	Select4  Action = "select.4"
	Select5  Action = "select.5"
	Select6  Action = "select.6"
	Select7  Action = "select.7"
	Select8  Action = "select.8"
	Select9  Action = "select.9"
	Select10 Action = "select.10"
	Select11 Action = "select.11"
	Select12 Action = "select.12"
	Select13 Action = "select.13"
	Select14 Action = "select.14"
	Select15 Action = "select.15"
)

// BattleMove 把原版六角方向編號 1..6 包成裝置無關的戰鬥移動動作。
func BattleMove(direction int) Action {
	if direction < 1 || direction > 6 {
		return None
	}
	return Action("battle.move." + strconv.Itoa(direction))
}

// BattleMoveDirection 解出 BattleMove 所帶的六角方向編號。
func BattleMoveDirection(action Action) (int, bool) {
	raw := string(action)
	if !strings.HasPrefix(raw, "battle.move.") {
		return 0, false
	}
	d, err := strconv.Atoi(strings.TrimPrefix(raw, "battle.move."))
	return d, err == nil && d >= 1 && d <= 6
}

var selections = [...]Action{Select1, Select2, Select3, Select4, Select5, Select6, Select7, Select8, Select9,
	Select10, Select11, Select12, Select13, Select14, Select15}

var digits = [...]Action{Digit0, Digit1, Digit2, Digit3, Digit4, Digit5, Digit6, Digit7, Digit8, Digit9}

func Digit(n int) Action {
	if n < 0 || n > 9 {
		return None
	}
	return digits[n]
}

func DigitNumber(action Action) (int, bool) {
	for n, candidate := range digits {
		if action == candidate {
			return n, true
		}
	}
	return 0, false
}

func Selection(n int) Action {
	if n < 1 || n > 99 {
		return None
	}
	if n <= len(selections) {
		return selections[n-1]
	}
	return Action("select." + strconv.Itoa(n))
}

func SelectionNumber(action Action) (int, bool) {
	raw := string(action)
	if !strings.HasPrefix(raw, "select.") {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimPrefix(raw, "select."))
	return n, err == nil && n >= 1 && n <= 99
}

type Rect struct{ X, Y, W, H int }

func (r Rect) Contains(x, y int) bool {
	return r.W > 0 && r.H > 0 && x >= r.X && y >= r.Y && x < r.X+r.W && y < r.Y+r.H
}

type Target struct {
	Action Action
	Rect   Rect
}

// Hit 依 renderer 公告的順序回傳第一個命中動作。命中區不應重疊；
// 若後續版面必須重疊，這個順序就是可測的選擇優先序。
func Hit(targets []Target, x, y int) Action {
	for _, t := range targets {
		if t.Rect.Contains(x, y) {
			return t.Action
		}
	}
	return None
}

// SurfaceToLogical 將平台 Surface 座標反算到等比縮放後的邏輯畫布。
// 黑邊回 false；縮放取寬高中較小比例，不拉伸原始畫面。
func SurfaceToLogical(surfaceX, surfaceY, surfaceW, surfaceH, logicalW, logicalH int) (int, int, bool) {
	if surfaceW <= 0 || surfaceH <= 0 || logicalW <= 0 || logicalH <= 0 {
		return 0, 0, false
	}
	scaleX, scaleY := float64(surfaceW)/float64(logicalW), float64(surfaceH)/float64(logicalH)
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}
	drawnW, drawnH := float64(logicalW)*scale, float64(logicalH)*scale
	offX, offY := (float64(surfaceW)-drawnW)/2, (float64(surfaceH)-drawnH)/2
	x, y := float64(surfaceX), float64(surfaceY)
	if x < offX || y < offY || x >= offX+drawnW || y >= offY+drawnH {
		return 0, 0, false
	}
	return int((x - offX) / scale), int((y - offY) / scale), true
}
