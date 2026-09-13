package lib3d

import "testing"

func TestNewBodyWorldVertices(t *testing.T) {
	box := SolidBox(2, 2, 2)
	b := NewBody(box, 10, 0, 0, BodyConfigDefault)
	verts := b.WorldVertices()
	if len(verts) != len(box.Vertices) {
		t.Fatalf("world vertices = %d, want %d", len(verts), len(box.Vertices))
	}
	// Der erste Vertex (-1,-1,-1) wird um x=10 verschoben.
	if verts[0] != NewVec3(9, -1, -1) {
		t.Errorf("vertex[0] = %v, want (9,-1,-1)", verts[0])
	}
}

func TestBodyGetFacePlanes(t *testing.T) {
	box := SolidBox(2, 2, 2)
	b := NewBody(box, 0, 0, 0, BodyConfigDefault)
	planes := b.GetFacePlanes()
	if len(planes) != 6 {
		t.Fatalf("planes = %d, want 6", len(planes))
	}
	// Jede Boxfläche hat den Abstand 1 vom Ursprung.
	for i, p := range planes {
		dist, _ := p.DistanceToPoint(NewVec3(0, 0, 0))
		if !almostEqual(dist, 1) {
			t.Errorf("plane %d distance to center = %v, want 1", i, dist)
		}
	}
}

func TestBodyWithoutFaces(t *testing.T) {
	grid := SolidGrid(10, 2)
	b := NewBody(grid, 0, 0, 0, BodyConfigDefault)
	if b.Faces != nil {
		t.Error("grid body should have nil faces")
	}
	if planes := b.GetFacePlanes(); planes != nil {
		t.Errorf("planes = %v, want nil", planes)
	}
}
