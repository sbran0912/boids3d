package lib3d

import (
	"math"
	"testing"
)

func TestMat4x4IdentityMult(t *testing.T) {
	id := mat4x4Identity()
	m := mat4x4Translate(1, 2, 3)
	if got := id.Mult(&m); got != m {
		t.Errorf("identity*M = %v, want %v", got, m)
	}
	if got := m.Mult(&id); got != m {
		t.Errorf("M*identity = %v, want %v", got, m)
	}
}

func TestMat4x4Flatten(t *testing.T) {
	m := mat4x4Translate(1, 2, 3)
	flat := m.Flatten()
	for i := range flat {
		if flat[i] != m.M[i] {
			t.Errorf("Flatten[%d] = %v, want %v", i, flat[i], m.M[i])
		}
	}
}

func TestMat4x4Perspective(t *testing.T) {
	fov := float32(math.Pi / 2) // tan(fov/2) = 1 -> f = 1
	m := Mat4x4Perspective(fov, 1, 1, 10)
	if !almostEqual(m.M[0], 1) {
		t.Errorf("M[0] = %v, want 1", m.M[0])
	}
	if !almostEqual(m.M[5], 1) {
		t.Errorf("M[5] = %v, want 1", m.M[5])
	}
	if !almostEqual(m.M[10], -11.0/9.0) {
		t.Errorf("M[10] = %v, want -11/9", m.M[10])
	}
	if m.M[11] != -1 {
		t.Errorf("M[11] = %v, want -1", m.M[11])
	}
	if !almostEqual(m.M[14], -20.0/9.0) {
		t.Errorf("M[14] = %v, want -20/9", m.M[14])
	}
}

func TestMat4x4Lookat(t *testing.T) {
	cam := NewVec3(0, 0, 5)
	target := NewVec3(0, 0, 0)
	up := NewVec3(0, 1, 0)
	view := Mat4x4LookAt(cam, target, up)

	// Die Kamera selbst liegt im Ursprung des Kameraraums.
	if got := cam.Transform(&view); got.Length() > testEps {
		t.Errorf("camera in view space = %v, want origin", got)
	}
	// Das Ziel liegt auf der negativen Z-Achse (Blickrichtung).
	got := target.Transform(&view)
	want := NewVec3(0, 0, -5)
	if !almostEqual(got.X, want.X) || !almostEqual(got.Y, want.Y) || !almostEqual(got.Z, want.Z) {
		t.Errorf("target in view space = %v, want %v", got, want)
	}
}
