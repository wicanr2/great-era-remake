package layout

import "testing"

func TestBattleCommandButtonsShareCompactMenuGeometry(t *testing.T) {
	for i := 0; i < 5; i++ {
		p := BattleCommandButton(i)
		if p.HitW != 185 || p.HitH != 24 || p.HitY != 160+i*24 {
			t.Fatalf("命令 %d 幾何=%+v", i+1, p)
		}
	}
	if p := BattleCommandButton(-1); p.HitW != 0 || p.HitH != 0 {
		t.Fatalf("非法命令索引不應有命中區：%+v", p)
	}
}

func TestBattleControlButtonsAreLargeAndNonOverlapping(t *testing.T) {
	var previous Placement
	for i := 0; i < 3; i++ {
		p := BattleControlButton(i)
		if p.HitW != 56 || p.HitH != 48 || p.HitY != 286 {
			t.Fatalf("控制 %d 幾何=%+v", i, p)
		}
		if i > 0 && p.HitX <= previous.HitX+previous.HitW {
			t.Fatalf("控制按鈕重疊：前=%+v 後=%+v", previous, p)
		}
		previous = p
	}
}

func TestBattleRetreatKeypadIsThreeByFourAndTouchSized(t *testing.T) {
	var previous Placement
	for i := 0; i < 12; i++ {
		p := BattleRetreatKeypadButton(i)
		if p.HitW != 56 || p.HitH != 48 {
			t.Fatalf("撤退鍵盤 %d 幾何=%+v", i, p)
		}
		if i > 0 {
			// 同列按鈕不重疊；跨列的 y 差由 row gap 護欄另測。
			if i%3 != 0 && p.HitX <= previous.HitX+previous.HitW {
				t.Fatalf("撤退鍵盤同列重疊：前=%+v 後=%+v", previous, p)
			}
			if i%3 == 0 && p.HitY <= previous.HitY {
				t.Fatalf("撤退鍵盤列順序錯誤：前=%+v 後=%+v", previous, p)
			}
		}
		previous = p
	}
	if p := BattleRetreatKeypadButton(-1); p.HitW != 0 || p.HitH != 0 {
		t.Fatalf("非法撤退鍵盤索引不應有命中區：%+v", p)
	}
	if p := BattleRetreatKeypadButton(12); p.HitW != 0 || p.HitH != 0 {
		t.Fatalf("超界撤退鍵盤索引不應有命中區：%+v", p)
	}
}
