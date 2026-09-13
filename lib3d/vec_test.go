package lib3d

import (
	"math"
	"testing"
)

const testEps = 1e-5

func almostEqual(a, b float32) bool {
	return float32(math.Abs(float64(a-b))) < testEps
}

func TestVec3Arithmetic(t *testing.T) {
	a := NewVec3(1, 2, 3)
	b := NewVec3(4, 5, 6)

	tests := []struct {
		name string
		got  Vec3
		want Vec3
	}{
		{"Add", a.Add(b), NewVec3(5, 7, 9)},
		{"Sub", a.Sub(b), NewVec3(-3, -3, -3)},
		{"Scale", a.Scale(2), NewVec3(2, 4, 6)},
		{"Negate", a.Negate(), NewVec3(-1, -2, -3)},
		{"Cross", a.Cross(b), NewVec3(-3, 6, -3)},
		{"Lerp", a.Lerp(b, 0.5), NewVec3(2.5, 3.5, 4.5)},
	}
	for _, tt := range tests {
		if tt.got != tt.want {
			t.Errorf("%s = %v, want %v", tt.name, tt.got, tt.want)
		}
	}
}

func TestVec3Magnitudes(t *testing.T) {
	a := NewVec3(1, 2, 3)
	if got := a.SquaredLength(); !almostEqual(got, 14) {
		t.Errorf("SquaredLength = %v, want 14", got)
	}
	if got := a.Length(); !almostEqual(got, float32(math.Sqrt(14))) {
		t.Errorf("Length = %v, want sqrt(14)", got)
	}
	if got := a.Dot(a); !almostEqual(got, 14) {
		t.Errorf("Dot = %v, want 14", got)
	}
	if got := a.SetMagnitude(5).Length(); !almostEqual(got, 5) {
		t.Errorf("SetMagnitude length = %v, want 5", got)
	}
	if got := a.Limit(2).Length(); !almostEqual(got, 2) {
		t.Errorf("Limit length = %v, want 2", got)
	}
	if got := a.Normalize(); !almostEqual(got.Length(), 1) {
		t.Errorf("Normalize length = %v, want 1", got)
	}
	if got := (Vec3{}).Normalize(); got != (Vec3{}) {
		t.Errorf("Normalize zero = %v, want zero", got)
	}
	if got := (Vec3{}).SetMagnitude(3); got != (Vec3{}) {
		t.Errorf("SetMagnitude zero = %v, want zero", got)
	}
}

func TestVec3DistanceAndEquals(t *testing.T) {
	a := NewVec3(0, 0, 0)
	b := NewVec3(3, 4, 0)
	if got := a.Distance(b); !almostEqual(got, 5) {
		t.Errorf("Distance = %v, want 5", got)
	}
	if !a.Equals(NewVec3(0, 0, 0)) {
		t.Error("Equals should be true for identical vectors")
	}
	if a.Equals(b) {
		t.Error("Equals should be false for different vectors")
	}
}

func TestVec3Transform(t *testing.T) {
	id := mat4x4Identity()
	p := NewVec3(1, 2, 3)
	if got := p.Transform(&id); got != p {
		t.Errorf("Transform(identity) = %v, want %v", got, p)
	}
	tr := mat4x4Translate(10, 20, 30)
	if got := p.Transform(&tr); got != NewVec3(11, 22, 33) {
		t.Errorf("Transform(translate) = %v, want (11,22,33)", got)
	}
	// TransformDir ignoriert die Translation.
	if got := p.TransformDir(&tr); got != p {
		t.Errorf("TransformDir(translate) = %v, want %v", got, p)
	}
}
