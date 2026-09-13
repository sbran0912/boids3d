package lib3d

import "math"

const epsilon = 1e-10

type Vec3 struct {
	X, Y, Z float32
}

type Vec2 struct {
	X, Y, S float32
}

func NewVec3(x, y, z float32) Vec3 {
	return Vec3{x, y, z}
}

func (a Vec3) Add(b Vec3) Vec3 {
	return NewVec3(a.X+b.X, a.Y+b.Y, a.Z+b.Z)
}

func (a Vec3) Sub(b Vec3) Vec3 {
	return NewVec3(a.X-b.X, a.Y-b.Y, a.Z-b.Z)
}

func (a Vec3) Scale(s float32) Vec3 {
	return NewVec3(a.X*s, a.Y*s, a.Z*s)
}

func (a Vec3) SetMagnitude(m float32) Vec3 {
	l := a.Length()
	if l == 0 {
		return NewVec3(0, 0, 0)
	}
	return a.Scale(m / l)
}

func (a Vec3) Limit(max float32) Vec3 {
	mSq := a.SquaredLength()
	if mSq > max*max {
		return a.Scale(max / float32(math.Sqrt(float64(mSq))))
	}
	return a
}

func (a Vec3) Negate() Vec3 {
	return NewVec3(-a.X, -a.Y, -a.Z)
}

func (a Vec3) Dot(b Vec3) float32 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func (a Vec3) Cross(b Vec3) Vec3 {
	return NewVec3(
		a.Y*b.Z-a.Z*b.Y,
		a.Z*b.X-a.X*b.Z,
		a.X*b.Y-a.Y*b.X,
	)
}

func (a Vec3) SquaredLength() float32 {
	return a.X*a.X + a.Y*a.Y + a.Z*a.Z
}

func (a Vec3) Length() float32 {
	return float32(math.Sqrt(float64(a.SquaredLength())))
}

func (a Vec3) Distance(b Vec3) float32 {
	return a.Sub(b).Length()
}

func (a Vec3) Normalize() Vec3 {
	l := a.Length()
	if l == 0 {
		return NewVec3(0, 0, 0)
	}
	return a.Scale(1 / l)
}

func (a Vec3) Lerp(b Vec3, t float32) Vec3 {
	return a.Add(b.Sub(a).Scale(t))
}

func (a Vec3) Clone() Vec3 {
	return a
}

func (a Vec3) Equals(b Vec3) bool {
	return math.Abs(float64(a.X-b.X)) < epsilon &&
		math.Abs(float64(a.Y-b.Y)) < epsilon &&
		math.Abs(float64(a.Z-b.Z)) < epsilon
}

func (v Vec3) Transform(m *Mat4x4) Vec3 {
	return NewVec3(
		m.M[0]*v.X+m.M[4]*v.Y+m.M[8]*v.Z+m.M[12],
		m.M[1]*v.X+m.M[5]*v.Y+m.M[9]*v.Z+m.M[13],
		m.M[2]*v.X+m.M[6]*v.Y+m.M[10]*v.Z+m.M[14],
	)
}

// TransformDir transformiert nur die Rotation eines Vektors (ohne
// Translation) – für Richtungen wie Licht-/Normalenvektoren.
// Wendet den Rotationsteil M·v an: für M = View-Matrix (deren Zeilen die
// Kamera-Basisvektoren sind) liefert das bereits Welt→Kamera; für eine
// Modell-Matrix ist es Modell→Welt.
func (v Vec3) TransformDir(m *Mat4x4) Vec3 {
	return NewVec3(
		m.M[0]*v.X+m.M[4]*v.Y+m.M[8]*v.Z,
		m.M[1]*v.X+m.M[5]*v.Y+m.M[9]*v.Z,
		m.M[2]*v.X+m.M[6]*v.Y+m.M[10]*v.Z,
	)
}

func (point Vec3) ToCamera(view, world *Mat4x4) Vec3 {
	vw := view.Mult(world)
	return point.Transform(&vw)
}

func (point Vec3) RotateAround(pivot Vec3, rotation *Mat4x4) Vec3 {
	rel := point.Sub(pivot)
	rotated := rel.Transform(rotation)
	return rotated.Add(pivot)
}

func (v Vec3) Project(fov float32) Vec2 {
	s := fov / (fov + v.Z)
	return Vec2{v.X * s, v.Y * s, s}
}
