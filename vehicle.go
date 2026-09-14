package main

import (
	"math"

	"boids3d/lib3d"
)

// Vehicle ist ein autonomes Fahrzeug der Simulation.
type Vehicle struct {
	Body    lib3d.Body
	Accel   lib3d.Vec3
	Heading lib3d.Vec3
	Health  float32
	DNA     [4]float32
}

// NewVehicle erstellt ein neues Vehicle mit zufälliger DNA.
func NewVehicle(body lib3d.Body) Vehicle {
	return Vehicle{
		Body:    body,
		Accel:   lib3d.NewVec3(0, 0, 0),
		Heading: lib3d.NewVec3(0, 0, 0),
	}
}

// AlignToVelocity richtet den Body an der Geschwindigkeit aus.
func (v *Vehicle) AlignToVelocity() {
	vel := v.Body.Vel

	mag := float32(math.Sqrt(float64(vel.X*vel.X + vel.Y*vel.Y + vel.Z*vel.Z)))
	if mag < 0.0001 {
		return
	}

	magXZ := float32(math.Sqrt(float64(vel.X*vel.X + vel.Z*vel.Z)))

	// vel.Y/mag muss in [-1,1] liegen (Schutz vor NaN durch Float-Rundung)
	cosY := vel.Y / mag
	if cosY > 1 {
		cosY = 1
	} else if cosY < -1 {
		cosY = -1
	}
	v.Body.RotX = float32(math.Acos(float64(cosY)))
	if magXZ < 0.0001 {
		v.Body.RotY = 0
	} else {
		v.Body.RotY = float32(math.Atan2(float64(vel.X/magXZ), float64(vel.Z/magXZ)))
	}
	v.Body.RotZ = 0

	// heading als normalisierte Richtung ableiten
	v.Heading = lib3d.NewVec3(vel.X/mag, vel.Y/mag, vel.Z/mag)
}

// ApplyForce addiert eine Kraft zur Beschleunigung.
func (v *Vehicle) ApplyForce(force lib3d.Vec3) {
	v.Accel = v.Accel.Add(force)
}

// Update integriert die Bewegung und begrenzt die Geschwindigkeit.
func (v *Vehicle) Update() {
	v.Body.Vel = v.Body.Vel.Add(v.Accel)
	speed := v.Body.Vel.Length()
	v.Body.Vel = v.Body.Vel.Normalize().Scale(lib3d.ConstrainNum(speed, 0.5, 2.0))

	v.Accel = lib3d.NewVec3(0, 0, 0)
	v.Body.Pos = v.Body.Pos.Add(v.Body.Vel)
}

// Seek steuert das Fahrzeug in Richtung eines Ziels.
func (v *Vehicle) Seek(target lib3d.Vec3) {
	desired := target.Sub(v.Body.Pos).Limit(3.0)
	steer := desired.Sub(v.Body.Vel).Limit(2.0)
	v.ApplyForce(steer.Scale(0.2))
}

// Allign gleicht die eigene Geschwindigkeit an die der Nachbarn an
// (Flocking-Regel: Alignment).
func (v *Vehicle) Allign(vehics []*Vehicle) {
	const minDistance = float32(50)
	sumVel := lib3d.NewVec3(0, 0, 0)
	count := 0

	for _, other := range vehics {
		distance := v.Body.Pos.Distance(other.Body.Pos)
		if distance > 0 && distance < minDistance {
			sumVel = sumVel.Add(other.Body.Vel)
			count++
		}
	}

	if count > 0 {
		sumVel = sumVel.Scale(1 / float32(count)) // Durchschnittsgeschwindigkeit
		sumVel = sumVel.SetMagnitude(0.1)
		steer := sumVel.Sub(v.Body.Vel).Scale(0.1)
		v.ApplyForce(steer)
	}
}

func (v *Vehicle) Separate(vehics []*Vehicle) {
	const minDistance = float32(40)
	diffSum := lib3d.NewVec3(0, 0, 0)
	count := 0
	for _, other := range vehics {
		distance := v.Body.Pos.Distance(other.Body.Pos)
		if distance > 0 && distance < minDistance {
			diff := v.Body.Pos.Sub(other.Body.Pos).Normalize()
			diffSum = diffSum.Add(diff)
			count++
		}
	}
	if count > 0 {
		diffSum = diffSum.Scale(1 / float32(count)) // Durchschnitt
		diffSum = diffSum.SetMagnitude(0.1)
		steer := diffSum.Sub(v.Body.Vel).Scale(0.1)
		v.ApplyForce(steer)
	}
}

// Cohesion steuert in Richtung des Schwerpunkts der Nachbarn
// (Flocking-Regel: Kohäsion).
func (v *Vehicle) Cohesion(vehics []*Vehicle) {
	const minDistance = float32(50)
	sumPos := lib3d.NewVec3(0, 0, 0)
	count := 0

	for _, other := range vehics {
		distance := v.Body.Pos.Distance(other.Body.Pos)
		if distance > 0 && distance < minDistance {
			sumPos = sumPos.Add(other.Body.Pos)
			count++
		}
	}

	if count > 0 {
		sumPos = sumPos.Scale(1 / float32(count)) // Schwerpunkt der Nachbarn
		desired := sumPos.Sub(v.Body.Pos)         // Richtung zur Gruppenmitte
		desired = desired.SetMagnitude(0.1)
		steer := desired.Sub(v.Body.Vel).Scale(0.1)
		v.ApplyForce(steer)
	}
}

// ApplyBoundary reflektiert die Geschwindigkeit an den Weltgrenzen.
func (v *Vehicle) ApplyBoundary() {
	// Definiere deine Weltgrenzen
	const (
		minX = -200.0
		maxX = 200.0
		minY = -130.0
		maxY = 130.0
		minZ = -130.0
		maxZ = 130.0
	)

	// X-Achse
	if v.Body.Pos.X < minX || v.Body.Pos.X > maxX {
		v.Body.Vel.X *= -1.0
	}

	// Y-Achse (Höhe)
	if v.Body.Pos.Y < minY || v.Body.Pos.Y > maxY {
		v.Body.Vel.Y *= -1.0
	}

	// Z-Achse (Tiefe)
	if v.Body.Pos.Z < minZ || v.Body.Pos.Z > maxZ {
		v.Body.Vel.Z *= -1.0
	}
}

// AvoidObstacles hält das Vehicle mit einem Sicherheitsabstand (margin)
// außerhalb der schwebenden Boxen. Die Kollisionserkennung nutzt die konvexen
// Flächen-Ebenen des Hindernisses (Body.GetFacePlanes) und den bestehenden
// Abstands-Test Plane.DistanceToPoint.
//
// DistanceToPoint liefert den Abstand zur Fläche und meldet über withinBoundary,
// ob die Projektion des Punktes innerhalb der Fläche liegt. Damit lässt sich
// die echte Distanz zur nächsten Fläche bestimmen. Liegt der Punkt vor einer
// Kante/Ecke, trifft keine Fläche – dann dient der größte vorzeichenbehaftete
// Ebenenabstand als konservative Untergrenze des Abstands.
//
// Unterschreitet der Abstand margin, wird das Vehicle entlang der Außen-Normalen
// herausgeschoben und die Geschwindigkeitskomponente ins Hindernis entfernt
// (Gleitbewegung).
func (v *Vehicle) AvoidObstacles(obstacles []*FloatingBox, margin float32) {
	for _, o := range obstacles {
		// Schneller Vorab-Test (Bounding Sphere): spart das Erzeugen der
		// Ebenen für weit entfernte Boids.
		if v.Body.Pos.Distance(o.Body.Pos) > o.Radius+margin {
			continue
		}

		planes := o.Body.GetFacePlanes()
		if len(planes) == 0 {
			continue
		}

		// Nächste Fläche suchen, auf die der Punkt projiziert, und ihren
		// vorzeichenbehafteten Abstand merken. Parallel den größten
		// vorzeichenbehafteten Abstand als Fallback für Kanten/Ecken
		// bereithalten, wo keine Fläche getroffen wird.
		nearest := float32(math.Inf(1))
		signed := float32(0)
		normal := lib3d.NewVec3(0, 0, 0)
		fallback := float32(math.Inf(-1))
		fallbackNormal := lib3d.NewVec3(0, 0, 0)

		for _, p := range planes {
			// Signed distance zur Ebene (EINMAL berechnet)
			dist, within := p.DistanceToPoint(v.Body.Pos)

			// Fallback: größter signed distance (immer für alle Ebenen)
			if dist > fallback {
				fallback = dist
				fallbackNormal = p.Normal
			}

			// Nearest: kürzester Abstand zu einer Fläche, die das Fahrzeug enthält
			if within && dist < nearest {
				nearest = dist
				signed = dist // Wiederverwendung!
				normal = p.Normal
			}
		}

		if math.IsInf(float64(nearest), 1) {
			signed = fallback
			normal = fallbackNormal
		}

		// signed < 0 → innerhalb des Volumens, sonst Abstand zur Oberfläche.
		// In beiden Fällen stellt eine Verschiebung um (margin - signed)
		// entlang der Außen-Normalen den Sicherheitsabstand her.
		if signed >= margin {
			continue
		}

		//v.Body.Pos = v.Body.Pos.Add(normal.Scale(margin - signed))

		// Geschwindigkeit ins Hindernis hinein streichen und wegsteuern.
		if into := v.Body.Vel.Dot(normal); into < 0 {
			v.Body.Vel = v.Body.Vel.Sub(normal.Scale(into))
		}
		v.ApplyForce(normal.Scale(0.01))
	}
}
