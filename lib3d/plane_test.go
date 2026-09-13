package lib3d

import "testing"

func squareFace() []Vec3 {
	return []Vec3{
		NewVec3(0, 0, 0),
		NewVec3(1, 0, 0),
		NewVec3(1, 1, 0),
		NewVec3(0, 1, 0),
	}
}

func TestPlaneFromFace(t *testing.T) {
	p := planeFromFace(squareFace())
	if p.Normal != NewVec3(0, 0, 1) {
		t.Errorf("normal = %v, want (0,0,1)", p.Normal)
	}
	if !almostEqual(p.Distance, 0) {
		t.Errorf("distance = %v, want 0", p.Distance)
	}
	if len(p.Boundary) != 4 {
		t.Errorf("boundary len = %d, want 4", len(p.Boundary))
	}
}

func TestPlaneDistanceToPoint(t *testing.T) {
	p := planeFromFace(squareFace())

	dist, within := p.DistanceToPoint(NewVec3(0.5, 0.5, 2))
	if !almostEqual(dist, 2) || !within {
		t.Errorf("inside: dist=%v within=%v, want 2/true", dist, within)
	}

	dist, within = p.DistanceToPoint(NewVec3(2, 2, 1))
	if !almostEqual(dist, 1) || within {
		t.Errorf("outside: dist=%v within=%v, want 1/false", dist, within)
	}
}

func TestPlaneContainsPoint(t *testing.T) {
	p := planeFromFace(squareFace())
	if !p.ContainsPoint(NewVec3(0.5, 0.5, 0)) {
		t.Error("center should be contained")
	}
	if p.ContainsPoint(NewVec3(2, 2, 0)) {
		t.Error("far point should not be contained")
	}
}

func TestPlaneIntersectLine(t *testing.T) {
	p := planeFromFace(squareFace())

	hit, ok := p.IntersectLine(NewVec3(0.5, 0.5, -1), NewVec3(0.5, 0.5, 1))
	if !ok || hit != NewVec3(0.5, 0.5, 0) {
		t.Errorf("intersect = %v, %v; want (0.5,0.5,0), true", hit, ok)
	}

	if _, ok := p.IntersectLine(NewVec3(0.5, 0.5, 1), NewVec3(0.5, 0.5, 2)); ok {
		t.Error("line on one side must not intersect")
	}
}
