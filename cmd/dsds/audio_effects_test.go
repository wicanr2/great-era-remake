package main

import (
	"testing"

	"github.com/wicanr2/great-era-remake/internal/ui/actions"
	uiaudio "github.com/wicanr2/great-era-remake/internal/ui/audio"
)

func TestModernEffectForActionCoversDesktopAndTouchEvents(t *testing.T) {
	for _, tc := range []struct {
		action actions.Action
		want   uiaudio.Effect
	}{
		{actions.Back, uiaudio.EffectCancel},
		{actions.Cancel, uiaudio.EffectCancel},
		{actions.DeleteDigit, uiaudio.EffectCancel},
		{actions.Submit, uiaudio.EffectConfirm},
		{actions.Confirm, uiaudio.EffectConfirm},
		{actions.OpenBiography, uiaudio.EffectConfirm},
		{actions.Digit7, uiaudio.EffectKey},
		{actions.Selection(12), uiaudio.EffectKey},
		{actions.BattleAttack, uiaudio.EffectBattleAttack},
		{actions.BattleAttackTarget(4), uiaudio.EffectBattleAttack},
		{actions.BattleMove(3), uiaudio.EffectBattleMove},
		{actions.BattleNextUnit, uiaudio.EffectBattleMove},
		{actions.BattleEndTurn, uiaudio.EffectBattleTurn},
		{actions.BattleCommand(2), uiaudio.EffectCommand},
	} {
		if got := modernEffectForAction(tc.action); got != tc.want {
			t.Errorf("action %q effect=%q，預期 %q", tc.action, got, tc.want)
		}
	}
}
