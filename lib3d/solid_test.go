package lib3d

import "testing"

func TestSolidBoxCounts(t *testing.T) {
	s := SolidBox(2, 2, 2)
	if len(s.Vertices) != 8 {
		t.Errorf("vertices = %d, want 8", len(s.Vertices))
	}
	if len(s.Edges) != 12 {
		t.Errorf("edges = %d, want 12", len(s.Edges))
	}
	if len(s.Faces) != 6 {
		t.Errorf("faces = %d, want 6", len(s.Faces))
	}
	if len(s.FlatEdges) != 12*2*3 {
		t.Errorf("flat edges = %d, want %d", len(s.FlatEdges), 12*2*3)
	}
	// 6 Quad-Flächen -> je 2 Dreiecke -> 6 Vertices -> 18 Floats.
	if len(s.FlatFaces) != 6*18 {
		t.Errorf("flat faces = %d, want %d", len(s.FlatFaces), 6*18)
	}
}

func TestSolidPyramidCounts(t *testing.T) {
	s := SolidPyramid(2, 3)
	if len(s.Vertices) != 5 {
		t.Errorf("vertices = %d, want 5", len(s.Vertices))
	}
	if len(s.Edges) != 8 {
		t.Errorf("edges = %d, want 8", len(s.Edges))
	}
	if len(s.Faces) != 5 {
		t.Errorf("faces = %d, want 5", len(s.Faces))
	}
}

func TestSolidGridCounts(t *testing.T) {
	s := SolidGrid(10, 3)
	if len(s.Vertices) != 4*4 {
		t.Errorf("vertices = %d, want 16", len(s.Vertices))
	}
	if len(s.Edges) != 2*3*4 {
		t.Errorf("edges = %d, want 24", len(s.Edges))
	}
	if len(s.Faces) != 0 {
		t.Errorf("faces = %d, want 0 (reines Drahtgitter)", len(s.Faces))
	}
	if len(s.FlatFaces) != 0 {
		t.Errorf("flat faces = %d, want 0", len(s.FlatFaces))
	}
}

func TestSolidSphereCounts(t *testing.T) {
	const slices, stacks = 4, 3
	s := SolidSphere(1, slices, stacks)
	if want := (stacks + 1) * (slices + 1); len(s.Vertices) != want {
		t.Errorf("vertices = %d, want %d", len(s.Vertices), want)
	}
	if want := (slices+1)*stacks + (stacks+1)*slices; len(s.Edges) != want {
		t.Errorf("edges = %d, want %d", len(s.Edges), want)
	}
	if want := stacks * slices; len(s.Faces) != want {
		t.Errorf("faces = %d, want %d", len(s.Faces), want)
	}
}
