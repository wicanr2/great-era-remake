package main

import (
	"strings"

	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	uiaudio "github.com/wicanr2/great-era-remake/internal/ui/audio"
)

// playInputEffect 是輸入層唯一的 Modern 效果音接點。它只消費已確認的
// Action，不觀察或修改規則狀態；鍵盤、滑鼠與觸控因此共用同一套音效語意。
func (a *app) playInputEffect(action actions.Action) {
	if a == nil || a.audio == nil || a.audioMode != uiaudio.ModeModern || action == actions.None {
		return
	}
	effect := modernEffectForAction(action)
	// 音訊是外殼 polish；裝置初始化或 player 建立失敗不能阻斷玩家輸入。
	_ = a.audio.PlayEffect(effect, 0.52)
}

func modernEffectForAction(action actions.Action) uiaudio.Effect {
	effect := uiaudio.EffectConfirm
	switch {
	case action == actions.Back || action == actions.Cancel:
		effect = uiaudio.EffectCancel
	case action == actions.DeleteDigit:
		effect = uiaudio.EffectCancel
	case action == actions.Submit || action == actions.Confirm ||
		action == actions.OpenCommands || action == actions.OpenNarrative || action == actions.OpenBiography:
		effect = uiaudio.EffectConfirm
	case action == actions.BattleAttack || strings.HasPrefix(string(action), uiaudioBattleAttackPrefix):
		effect = uiaudio.EffectBattleAttack
	case action == actions.BattleEndTurn:
		effect = uiaudio.EffectBattleTurn
	case action == actions.BattleNextUnit:
		effect = uiaudio.EffectBattleMove
	case strings.HasPrefix(string(action), "battle.move."):
		effect = uiaudio.EffectBattleMove
	case strings.HasPrefix(string(action), "battle.command."):
		effect = uiaudio.EffectCommand
	case strings.HasPrefix(string(action), "input.digit.") || strings.HasPrefix(string(action), "select."):
		effect = uiaudio.EffectKey
	}
	return effect
}

const uiaudioBattleAttackPrefix = "battle.attack-target."
