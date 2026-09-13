package main

import (
	"runtime"

	"boids3d/lib3d"
)

func main() {
	runtime.LockOSThread()

	r := &lib3d.Renderer{}
	if !r.Init(1600, 1000) {
		return
	}
	camPos := lib3d.NewVec3(50, 20, 200)
	target := lib3d.NewVec3(0, 0, 0)
	up := lib3d.NewVec3(0, 1, 0)
	r.SetFog(100.0, 400.0, 0.25, 0.25, 0.25, 1.0)

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
	r.StartAnimation(func() {
		r.Background(40, 40, 40)

		view := lib3d.Mat4x4Lookat(camPos, target, up)
		proj := lib3d.Mat4x4Perspective(1.2, r.Aspect(), 0.1, 1000.0)
		r.SetProjection(&proj)

		sunDir := lib3d.NewVec3(0.5, 1.0, 0.3)
		camLight := sunDir.TransformDir(&view)
		r.SetLightDirection(camLight.X, camLight.Y, camLight.Z)

		r.DrawBody(&grid, &view)

		// Hindernisse animieren und zeichnen.
		now := r.Time()
		for _, o := range obstacles {
			o.Update(now)
			r.DrawBody(o.Body, &view)
		}

		for _, v := range vehics {
			if r.IsMouseDown() {
				v.Seek(lib3d.NewVec3(0, 0, 0))
			}
			v.Allign(vehics)
			v.Separate(vehics)
			v.Cohesion(vehics)
			v.ApplyBoundary()
			v.AvoidObstacles(obstacles, obstacleMargin)
			v.AlignToVelocity()
			v.Update()
			r.DrawBody(&v.Body, &view)
		}
	})
}
