package lib3d

import "math"

// Solid ist reine, zwischen Bodies teilbare Geometrie: Vertices, Kanten und
// – optional – die Flächen-Topologie. Faces hat zwei Konsumenten:
//
//  1. expandFaces() trianguliert daraus FlatFaces für das gefüllte,
//     beleuchtete Rendering (aus den Kanten allein nicht rekonstruierbar).
//  2. NewBody übernimmt sie als Kollisions-Topologie (Single Source of Truth).
//
// Die flachen Edge-/Face-Arrays werden einmalig in newSolid expandiert und pro
// Frame in den GPU-Batch hochgeladen. Faces darf nil sein (reines Drahtgitter,
// z. B. Grid).
//
// Design-Entscheidung (Variante A): Die Topologie bleibt am Solid, weil sie
// Geometrie-Wissen ist und von allen Bodies desselben Meshes geteilt wird.
type Solid struct {
	Vertices  []Vec3
	Edges     [][2]int // eine Kante = ein Index-Paar
	Faces     [][]int  // ein Face = konvexes Polygon (leer = kein Fill/Kollision)
	FlatEdges []float32
	FlatFaces []float32
}

// newSolid erstellt ein Solid und expandiert Kanten/Flächen einmalig in die
// flachen Float-Arrays fürs GPU-Batching. faces darf nil sein (reines
// Drahtgitter, z. B. Grid).
func newSolid(vertices []Vec3, edges [][2]int, faces [][]int) *Solid {
	s := &Solid{Vertices: vertices, Edges: edges, Faces: faces}
	s.expandEdges()
	s.expandFaces()
	return s
}

// expandEdges füllt FlatEdges (zwei Vertices pro Kante, x,y,z,…).
func (s *Solid) expandEdges() {
	s.FlatEdges = make([]float32, len(s.Edges)*2*3)
	idx := 0
	for _, e := range s.Edges {
		a := s.Vertices[e[0]]
		b := s.Vertices[e[1]]
		s.FlatEdges[idx] = a.X
		s.FlatEdges[idx+1] = a.Y
		s.FlatEdges[idx+2] = a.Z
		idx += 3
		s.FlatEdges[idx] = b.X
		s.FlatEdges[idx+1] = b.Y
		s.FlatEdges[idx+2] = b.Z
		idx += 3
	}
}

// expandFaces trianguliert jedes Face (Fan aus Index 0) und füllt FlatFaces.
// Für Solids ohne Face-Daten (z. B. Grid) bleibt FlatFaces leer.
func (s *Solid) expandFaces() {
	s.FlatFaces = nil
	for _, face := range s.Faces {
		if len(face) < 3 {
			continue
		}
		for j := 1; j < len(face)-1; j++ {
			for _, k := range [3]int{0, j, j + 1} {
				v := s.Vertices[face[k]]
				s.FlatFaces = append(s.FlatFaces, v.X, v.Y, v.Z)
			}
		}
	}
}

func SolidBox(w, h, d float32) *Solid {
	hw := w / 2
	hh := h / 2
	hd := d / 2

	verts := []Vec3{
		{-hw, -hh, -hd}, {hw, -hh, -hd}, {hw, hh, -hd}, {-hw, hh, -hd},
		{-hw, -hh, hd}, {hw, -hh, hd}, {hw, hh, hd}, {-hw, hh, hd},
	}

	edges := [][2]int{
		{0, 1}, {1, 2}, {2, 3}, {3, 0}, {4, 5}, {5, 6}, {6, 7}, {7, 4},
		{0, 4}, {1, 5}, {2, 6}, {3, 7},
	}

	// Flächen: je ein konvexes Quad (CCW von außen), identisch mit der
	// Body.Faces-Topologie (Kollision/Raycasting).
	faces := [][]int{
		{0, 3, 2, 1}, // vorne
		{4, 5, 6, 7}, // hinten
		{0, 4, 7, 3}, // links
		{1, 2, 6, 5}, // rechts
		{0, 1, 5, 4}, // unten
		{3, 7, 6, 2}, // oben
	}

	return newSolid(verts, edges, faces)
}

func SolidPyramid(base, height float32) *Solid {
	hb := base / 2
	verts := []Vec3{
		{-hb, -height / 2, -hb}, {hb, -height / 2, -hb}, {hb, -height / 2, hb},
		{-hb, -height / 2, hb}, {0, height / 2, 0},
	}
	edges := [][2]int{
		{0, 1}, {1, 2}, {2, 3}, {3, 0}, {0, 4}, {1, 4}, {2, 4}, {3, 4},
	}

	// Flächen: 4 Dreiecks-Seiten + Quadrat-Basis (CCW von außen).
	faces := [][]int{
		{0, 1, 4},    // vorne
		{1, 2, 4},    // rechts
		{2, 3, 4},    // hinten
		{3, 0, 4},    // links
		{3, 2, 1, 0}, // Basis
	}

	return newSolid(verts, edges, faces)
}

func SolidGrid(size float32, cells int) *Solid {
	half := size / 2
	step := size / float32(cells)
	stride := cells + 1

	verts := make([]Vec3, stride*stride)
	vi := 0
	for iz := 0; iz <= cells; iz++ {
		for ix := 0; ix <= cells; ix++ {
			verts[vi] = NewVec3(-half+float32(ix)*step, 0, -half+float32(iz)*step)
			vi++
		}
	}

	edges := make([][2]int, 0, cells*stride*2)

	for iz := 0; iz <= cells; iz++ {
		for ix := 0; ix < cells; ix++ {
			idx := iz*stride + ix
			edges = append(edges, [2]int{idx, idx + 1})
		}
	}
	for ix := 0; ix <= cells; ix++ {
		for iz := 0; iz < cells; iz++ {
			idx := iz*stride + ix
			edges = append(edges, [2]int{idx, idx + stride})
		}
	}

	return newSolid(verts, edges, nil)
}

func SolidSphere(radius float32, slices, stacks int) *Solid {
	stride := slices + 1
	vcount := (stacks + 1) * stride
	verts := make([]Vec3, vcount)
	vi := 0

	for i := 0; i <= stacks; i++ {
		theta := (float32(i) / float32(stacks)) * float32(math.Pi)
		y := radius * float32(math.Cos(float64(theta)))
		r := radius * float32(math.Sin(float64(theta)))
		for j := 0; j <= slices; j++ {
			phi := (float32(j) / float32(slices)) * 2 * float32(math.Pi)
			verts[vi] = NewVec3(r*float32(math.Cos(float64(phi))), y, r*float32(math.Sin(float64(phi))))
			vi++
		}
	}

	edges := make([][2]int, 0, (slices+1)*stacks+(stacks+1)*slices)

	for j := 0; j <= slices; j++ {
		for i := 0; i < stacks; i++ {
			a := i*stride + j
			b := (i+1)*stride + j
			edges = append(edges, [2]int{a, b})
		}
	}
	for i := 0; i <= stacks; i++ {
		for j := 0; j < slices; j++ {
			a := i*stride + j
			b := i*stride + j + 1
			edges = append(edges, [2]int{a, b})
		}
	}

	// Flächen: jedes Gitterzellen-Quad (stacks × slices) wird eine Face.
	// An den Polen sind zwei Vertices identisch -> degenerierte, aber
	// harmlose Quads.
	faces := make([][]int, 0, stacks*slices)
	for i := 0; i < stacks; i++ {
		for j := 0; j < slices; j++ {
			a := i*stride + j
			b := i*stride + j + 1
			c := (i+1)*stride + j + 1
			d := (i+1)*stride + j
			faces = append(faces, []int{a, b, c, d})
		}
	}

	return newSolid(verts, edges, faces)
}
