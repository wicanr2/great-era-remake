package mobile

import "testing"

func TestLifecycleGateDeduplicatesBackgroundAndResumeEffects(t *testing.T) {
	g := NewLifecycleGate()
	if g.Phase() != PhaseActive || !g.AcceptInput() {
		t.Fatal("新 gate 應處於 active")
	}
	paused := g.Apply(EventPause)
	if !paused.Changed || !paused.PauseAudio || !paused.ResetInput || g.AcceptInput() {
		t.Fatalf("pause effect=%+v phase=%v", paused, g.Phase())
	}
	if duplicate := g.Apply(EventPause); duplicate.Changed || duplicate.PauseAudio || duplicate.ResetInput {
		t.Fatalf("重複 pause 不應再產生副作用：%+v", duplicate)
	}
	resumed := g.Apply(EventResume)
	if !resumed.Changed || !resumed.ResumeAudio || !resumed.ResetInput || !g.AcceptInput() {
		t.Fatalf("resume effect=%+v phase=%v", resumed, g.Phase())
	}
	if duplicate := g.Apply(EventResume); duplicate.Changed || duplicate.ResumeAudio {
		t.Fatalf("重複 resume 不應再產生副作用：%+v", duplicate)
	}
}

func TestLifecycleGateDestroyIsTerminalAndClosesAudioOnce(t *testing.T) {
	g := NewLifecycleGate()
	effect := g.Apply(EventDestroy)
	if !effect.Changed || !effect.PauseAudio || !effect.CloseAudio || !effect.ResetInput {
		t.Fatalf("destroy effect=%+v", effect)
	}
	if g.Phase() != PhaseDestroyed || g.AcceptInput() {
		t.Fatalf("destroy 後 phase=%v input=%v", g.Phase(), g.AcceptInput())
	}
	for _, event := range []LifecycleEvent{EventDestroy, EventPause, EventResume, 99} {
		if got := g.Apply(event); got.Changed || got.CloseAudio || got.PauseAudio || got.ResumeAudio {
			t.Fatalf("terminal event %v 產生副作用：%+v", event, got)
		}
	}
}
