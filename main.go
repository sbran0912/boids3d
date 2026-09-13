package main

import (
	"math"
	"runtime"

	"lib3d_group/lib3d"
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

func main() {
	runtime.LockOSThread()

	if !lib3d.DefaultRenderer.Init(1600, 1000) {
		return
	}
	camPos := lib3d.NewVec3(50, 20, 200)
	target := lib3d.NewVec3(0, 0, 0)
	up := lib3d.NewVec3(0, 1, 0)
	lib3d.DefaultRenderer.SetFog(100.0, 400.0, 0.25, 0.25, 0.25, 1.0)

	gridMesh := lib3d.SolidGrid(600, 24)
	grid := lib3d.NewBody(gridMesh, 0, 0, 0, lib3d.BodyConfig{Color: "#777774", LineWidth: 1.0})

	vehicMesh := lib3d.SolidPyramid(2, 6)
	vehics := make([]*Vehicle, 0, 10)
	for range 50 {
		vehic := NewVehicle(lib3d.NewBody(vehicMesh, 0, 20, 100, lib3d.BodyConfigDefault))
		vehic.Body.Vel = lib3d.NewVec3(lib3d.RandomFloat(-2, 2), lib3d.RandomFloat(-2, 2), lib3d.RandomFloat(-2, 2))
		vehics = append(vehics, &vehic)
	}

	// Fünf schwebende Boxen als Hindernisse für die Boids.
	boxMesh := lib3d.SolidBox(40, 40, 40)
	obstacles := []*FloatingBox{
		NewFloatingBox(boxMesh, lib3d.NewVec3(-90, 40, -50), "#e06c5a", 0.0, 0.35),
		NewFloatingBox(boxMesh, lib3d.NewVec3(60, 70, 20), "#5aa9e0", 1.7, -0.25),
		NewFloatingBox(boxMesh, lib3d.NewVec3(10, 60, -90), "#8ad06a", 3.1, 0.2),
		NewFloatingBox(boxMesh, lib3d.NewVec3(80, -50, 80), "#ff6b6b", 2.5, 0.15),
		NewFloatingBox(boxMesh, lib3d.NewVec3(-40, 100, 40), "#4ecdc4", 4.2, -0.3),
	}
	const obstacleMargin = float32(20.0)

	/*---------------------------------
	Render-Schleife (in der Library)
	---------------------------------*/
	lib3d.DefaultRenderer.StartAnimation(func() {
		lib3d.DefaultRenderer.Background(40, 40, 40)

		view := lib3d.Mat4x4Lookat(camPos, target, up)
		proj := lib3d.Mat4x4Perspective(1.2, lib3d.DefaultRenderer.Aspect(), 0.1, 1000.0)
		lib3d.DefaultRenderer.SetProjection(&proj)

		sunDir := lib3d.NewVec3(0.5, 1.0, 0.3)
		camLight := sunDir.TransformDir(&view)
		lib3d.DefaultRenderer.SetLightDirection(camLight.X, camLight.Y, camLight.Z)

		grid.Draw(&view)

		// Hindernisse animieren und zeichnen.
		now := lib3d.DefaultRenderer.Time()
		for _, o := range obstacles {
			o.Update(now)
			o.Body.Draw(&view)
		}

		for _, v := range vehics {
			if lib3d.DefaultRenderer.IsMouseDown() {
				v.Seek(lib3d.NewVec3(0, 0, 0))
			}
			v.Allign(vehics)
			v.Separate(vehics)
			v.Cohesion(vehics)
			v.ApplyBoundary()
			v.AvoidObstacles(obstacles, obstacleMargin)
			v.AlignToVelocity()
			v.Update()
			v.Body.Draw(&view)
		}
	})
}
