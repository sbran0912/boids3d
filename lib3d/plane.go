package lib3d

import "math"

type Plane struct {
	Normal   Vec3
	Distance float32
	Boundary []Vec3
}

func planeCreate(normal Vec3, distance float32) Plane {
	return Plane{Normal: normal, Distance: distance}
}

func planeFromFace(faceVerts []Vec3) Plane {
	edge1 := faceVerts[1].Sub(faceVerts[0])
	edge2 := faceVerts[2].Sub(faceVerts[0])
	normal := edge1.Cross(edge2).Normalize()
	distance := -normal.Dot(faceVerts[0])

	boundary := make([]Vec3, len(faceVerts))
	copy(boundary, faceVerts)

	return Plane{
		Normal:   normal,
		Distance: distance,
		Boundary: boundary,
	}
}

func (p *Plane) DistanceToPoint(pnt Vec3) (float32, bool) {
	signedDist := p.Normal.Dot(pnt) + p.Distance
	closestPoint := pnt.Sub(p.Normal.Scale(signedDist))
	withinBoundary := true
	if len(p.Boundary) > 0 {
		withinBoundary = p.ContainsPoint(closestPoint)
	}
	return float32(math.Abs(float64(signedDist))), withinBoundary
}

func (p *Plane) IntersectLine(p1, p2 Vec3) (Vec3, bool) {
	dir := p2.Sub(p1)
	denom := p.Normal.Dot(dir)

	if math.Abs(float64(denom)) < 1e-10 {
		return Vec3{}, false
	}

	t := -(p.Normal.Dot(p1) + p.Distance) / denom

	if t < 0 || t > 1 {
		return Vec3{}, false
	}

	hit := p1.Add(dir.Scale(t))

	if len(p.Boundary) > 0 && !p.ContainsPoint(hit) {
		return Vec3{}, false
	}

	return hit, true
}

func (p *Plane) ContainsPoint(pnt Vec3) bool {
	for i := 0; i < len(p.Boundary); i++ {
		a := p.Boundary[i]
		b := p.Boundary[(i+1)%len(p.Boundary)]
		edge := b.Sub(a)
		toPoint := pnt.Sub(a)
		if edge.Cross(toPoint).Dot(p.Normal) < 0 {
			return false
		}
	}
	return true
}
