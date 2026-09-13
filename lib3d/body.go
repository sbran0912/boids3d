package lib3d

type Body struct {
	Solid     *Solid
	Pos       Vec3
	Vel       Vec3
	RotX      float32
	RotY      float32
	RotZ      float32
	Color     string
	LineWidth float32

	// Faces ist die Topologie für Physik/Kollision (Weltkoordinaten-Ebenen).
	// Sie ist bewusst vom Solid getrennt: NewBody übernimmt standardmäßig
	// Solid.Faces, aber durch Setzen auf nil lässt sich Kollision pro Body
	// abschalten (z. B. für Kugeln, deren Pol-Quads degeneriert sind).
	Faces [][]int
}

type BodyConfig struct {
	Color     string
	LineWidth float32
	RotX      float32
	RotY      float32
	RotZ      float32
}

var BodyConfigDefault = BodyConfig{
	Color:     "#ffffff",
	LineWidth: 1.0,
}

type Line struct {
	P1        Vec3
	P2        Vec3
	Color     string
	LineWidth float32
}

func NewBody(solid *Solid, x, y, z float32, cfg BodyConfig) Body {
	color := cfg.Color
	if color == "" {
		color = "#ffffff"
	}

	b := Body{
		Solid:     solid,
		Pos:       NewVec3(x, y, z),
		Vel:       NewVec3(0, 0, 0),
		RotX:      cfg.RotX,
		RotY:      cfg.RotY,
		RotZ:      cfg.RotZ,
		Color:     color,
		LineWidth: cfg.LineWidth,
		// Single Source of Truth: Kollisions-Topologie = Solid-Topologie.
		Faces: solid.Faces,
	}
	return b
}

// modelMatrix liefert Translation×Rotation – dieselbe Matrix, die Draw und
// die Physik verwenden. Eine einzige Quelle für die Welt-Transformation,
// damit Rendering und Kollision nicht divergieren können.
func (b *Body) modelMatrix() Mat4x4 {
	t := mat4x4Translate(b.Pos.X, b.Pos.Y, b.Pos.Z)
	if b.RotX == 0 && b.RotY == 0 && b.RotZ == 0 {
		return t
	}
	rot := mat4x4Rotate(b.RotX, b.RotY, b.RotZ)
	return t.Mult(&rot)
}

// WorldVertices wendet die Modellmatrix auf alle Solid-Vertices an.
func (b *Body) WorldVertices() []Vec3 {
	m := b.modelMatrix()
	out := make([]Vec3, len(b.Solid.Vertices))
	for i := range b.Solid.Vertices {
		out[i] = b.Solid.Vertices[i].Transform(&m)
	}
	return out
}

// GetFacePlanes liefert die Kollisions-/Raycast-Ebenen dieses Körpers in
// Weltkoordinaten. Gelesen wird b.Faces (nicht Solid.Faces), sodass sich die
// Kollision pro Body über b.Faces = nil abschalten lässt. Bodies ohne Faces
// liefern nil.
func (b *Body) GetFacePlanes() []Plane {
	if len(b.Faces) == 0 {
		return nil
	}

	worldVerts := b.WorldVertices()
	planes := make([]Plane, len(b.Faces))
	for i, face := range b.Faces {
		faceVerts := make([]Vec3, len(face))
		for j, vi := range face {
			faceVerts[j] = worldVerts[vi]
		}
		planes[i] = planeFromFace(faceVerts)
	}
	return planes
}

func (a *Body) Distance(b *Body) float32 {
	return a.Pos.Distance(b.Pos)
}
