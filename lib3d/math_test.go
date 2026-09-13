package lib3d

import "testing"

func TestConstrainNum(t *testing.T) {
	if got := ConstrainNum(5, 0, 10); got != 5 {
		t.Errorf("ConstrainNum(5,0,10) = %v, want 5", got)
	}
	if got := ConstrainNum(-1, 0, 10); got != 0 {
		t.Errorf("ConstrainNum(-1,0,10) = %v, want 0", got)
	}
	if got := ConstrainNum(11, 0, 10); got != 10 {
		t.Errorf("ConstrainNum(11,0,10) = %v, want 10", got)
	}
}

func TestRandomFloatInRange(t *testing.T) {
	for i := 0; i < 100; i++ {
		v := RandomFloat(-2, 2)
		if v < -2 || v > 2 {
			t.Fatalf("RandomFloat(-2,2) = %v, außerhalb des Bereichs", v)
		}
	}
}
