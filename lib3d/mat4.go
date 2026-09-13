package lib3d

import "math"

type Mat4x4 struct {
	M [16]float32
}

func mat4x4Identity() Mat4x4 {
	return Mat4x4{[16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}}
}

func mat4x4Translate(dx, dy, dz float32) Mat4x4 {
	return Mat4x4{[16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		dx, dy, dz, 1,
	}}
}

func mat4x4Rotate(ax, ay, az float32) Mat4x4 {
	rx := Mat4x4{[16]float32{
		1, 0, 0, 0,
		0, float32(math.Cos(float64(ax))), float32(math.Sin(float64(ax))), 0,
		0, -float32(math.Sin(float64(ax))), float32(math.Cos(float64(ax))), 0,
		0, 0, 0, 1,
	}}

	ry := Mat4x4{[16]float32{
		float32(math.Cos(float64(ay))), 0, -float32(math.Sin(float64(ay))), 0,
		0, 1, 0, 0,
		float32(math.Sin(float64(ay))), 0, float32(math.Cos(float64(ay))), 0,
		0, 0, 0, 1,
	}}

	rz := Mat4x4{[16]float32{
		float32(math.Cos(float64(az))), float32(math.Sin(float64(az))), 0, 0,
		-float32(math.Sin(float64(az))), float32(math.Cos(float64(az))), 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}}

	ryRx := rz.Mult(&ry)
	return ryRx.Mult(&rx)
}

func (a *Mat4x4) Mult(b *Mat4x4) Mat4x4 {
	var r Mat4x4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				r.M[i+j*4] += a.M[i+k*4] * b.M[k+j*4]
			}
		}
	}
	return r
}

func Mat4x4Lookat(cameraPos, target, up Vec3) Mat4x4 {
	forward := target.Sub(cameraPos).Normalize()
	right := forward.Cross(up).Normalize()
	realUp := right.Cross(forward)

	return Mat4x4{[16]float32{
		right.X, realUp.X, -forward.X, 0,
		right.Y, realUp.Y, -forward.Y, 0,
		right.Z, realUp.Z, -forward.Z, 0,
		-right.Dot(cameraPos),
		-realUp.Dot(cameraPos),
		forward.Dot(cameraPos),
		1,
	}}
}

func Mat4x4Perspective(fovY, aspect, znear, zfar float32) Mat4x4 {
	f := 1 / float32(math.Tan(float64(fovY)/2))
	return Mat4x4{[16]float32{
		f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, -(zfar + znear) / (zfar - znear), -1,
		0, 0, -2 * znear * zfar / (zfar - znear), 0,
	}}
}

func (m *Mat4x4) Flatten() [16]float32 {
	var dst [16]float32
	for i := 0; i < 16; i++ {
		dst[i] = m.M[i]
	}
	return dst
}
