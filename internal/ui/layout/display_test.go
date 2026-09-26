package layout

import "testing"

func TestDisplayOptionsGeometryIsSharedByBothInputAndRenderer(t *testing.T) {
	wording := DisplayWordingOption(1, 218, 0, 394)
	if wording.X != 218 || wording.Y != 164 || wording.HitX != 218 || wording.HitY != 149 ||
		wording.HitW != 394 || wording.HitH != 48 {
		t.Fatalf("用語版面=%+v", wording)
	}
	modern := DisplayThemeOption(1, 218, 0, 394)
	if modern.X != 419 || modern.Y != 226 || modern.HitX != 419 || modern.HitY != 211 ||
		modern.HitW != 193 || modern.HitH != 48 {
		t.Fatalf("主題版面=%+v", modern)
	}
}
