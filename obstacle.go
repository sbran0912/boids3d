package main

import (
	"math"

	"boids3d/lib3d"
)

// FloatingBox ist ein schwebendes Hindernis für die Boids: ein Body mit
// Grundposition, sanfter Auf-und-Ab-Bewegung und langsamer Rotation.
type FloatingBox struct {
	Body   *lib3d.Body
	Base   lib3d.Vec3
	Phase  float32
	Spin   float32
	Radius float32 // Bounding-Sphere-Radius für den schnellen Vorab-Test
}

// NewFloatingBox erstellt eine schwebende Hindernis-Box an Position pos.
func NewFloatingBox(mesh *lib3d.Solid, pos lib3d.Vec3, color string, phase, spin float32) *FloatingBox {
	radius := float32(0)
	for _, vert := range mesh.Vertices {
		if l := vert.Length(); l > radius {
			radius = l
		}
	}

	body := lib3d.NewBody(mesh, pos.X, pos.Y, pos.Z, lib3d.BodyConfig{Color: color, LineWidth: 1.5})
	return &FloatingBox{
		Body:   &body,
		Base:   pos,
		Phase:  phase,
		Spin:   spin,
		Radius: radius,
	}
}

// Update animiert die Box (Schweben + Rotation). t ist die Zeit in Sekunden.
func (f *FloatingBox) Update(t float32) {
	f.Body.Pos = lib3d.NewVec3(
		f.Base.X,
		f.Base.Y+float32(math.Sin(float64(t+f.Phase)))*6,
		f.Base.Z,
	)
	f.Body.RotY = t * f.Spin
	f.Body.RotX = float32(math.Sin(float64(t*0.6+f.Phase))) * 0.3
}
