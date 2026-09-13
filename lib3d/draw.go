package lib3d

import (
	"math"

	"github.com/go-gl/gl/v3.3-core/gl"
)

type DrawStyle int

const (
	Stroke DrawStyle = iota
	Fill
	Both
)

type EffectMode int

const (
	EffectFlat EffectMode = iota
	EffectGradient
	EffectPulse
)

func (r *Renderer) Point(x, y, z float32) {
	r.submit(gl.POINTS, []float32{x, y, z}, true)
}

func (r *Renderer) Line(x1, y1, z1, x2, y2, z2 float32) {
	r.submit(gl.LINES, []float32{x1, y1, z1, x2, y2, z2}, true)
}

func (r *Renderer) Triangle(x1, y1, z1, x2, y2, z2, x3, y3, z3 float32, style DrawStyle) {
	pts := []float32{x1, y1, z1, x2, y2, z2, x3, y3, z3}
	shapeMetrics3d(pts)

	if style == Fill || style == Both {
		r.submit(gl.TRIANGLES, pts, false)
	}
	if style == Stroke || style == Both {
		s := []float32{x1, y1, z1, x2, y2, z2, x2, y2, z2, x3, y3, z3, x3, y3, z3, x1, y1, z1}
		r.submit(gl.LINES, s, true)
	}
}

func (r *Renderer) Shape(x1, y1, z1, x2, y2, z2, x3, y3, z3, x4, y4, z4 float32, style DrawStyle) {
	pts := []float32{x1, y1, z1, x2, y2, z2, x3, y3, z3, x4, y4, z4}
	shapeMetrics3d(pts)

	if style == Fill || style == Both {
		f := []float32{x1, y1, z1, x2, y2, z2, x3, y3, z3, x1, y1, z1, x3, y3, z3, x4, y4, z4}
		r.submit(gl.TRIANGLES, f, false)
	}
	if style == Stroke || style == Both {
		s := []float32{x1, y1, z1, x2, y2, z2, x2, y2, z2, x3, y3, z3, x3, y3, z3, x4, y4, z4, x4, y4, z4, x1, y1, z1}
		r.submit(gl.LINES, s, true)
	}
}

func (r *Renderer) Rect(x, y, w, h float32, style DrawStyle, z float32) {
	hw := w / 2
	hh := h / 2
	r.Shape(x-hw, y-hh, z, x+hw, y-hh, z, x+hw, y+hh, z, x-hw, y+hh, z, style)
}

func (r *Renderer) Circle(x, y, z, radius float32, style DrawStyle, segments int) {
	tau := float32(2 * math.Pi)

	if style == Fill || style == Both {
		fill := make([]float32, segments*9)
		fi := 0
		for i := 0; i < segments; i++ {
			a0 := (float32(i) / float32(segments)) * tau
			a1 := (float32(i+1) / float32(segments)) * tau
			fill[fi] = x
			fill[fi+1] = y
			fill[fi+2] = z
			fi += 3
			fill[fi] = x + float32(math.Cos(float64(a0)))*radius
			fill[fi+1] = y + float32(math.Sin(float64(a0)))*radius
			fill[fi+2] = z
			fi += 3
			fill[fi] = x + float32(math.Cos(float64(a1)))*radius
			fill[fi+1] = y + float32(math.Sin(float64(a1)))*radius
			fill[fi+2] = z
			fi += 3
		}
		r.submit(gl.TRIANGLES, fill, false)
	}

	if style == Stroke || style == Both {
		stroke := make([]float32, segments*3)
		si := 0
		for i := 0; i < segments; i++ {
			a := (float32(i) / float32(segments)) * tau
			stroke[si] = x + float32(math.Cos(float64(a)))*radius
			stroke[si+1] = y + float32(math.Sin(float64(a)))*radius
			stroke[si+2] = z
			si += 3
		}
		r.submit(gl.LINE_LOOP, stroke, true)
	}
}

func (r *Renderer) Polygon(pts []float32, style DrawStyle) {
	if len(pts) < 4 {
		return
	}
	shapeMetrics3d(pts)

	if style == Fill || style == Both {
		tris := len(pts)/3 - 2
		fill := make([]float32, tris*9)
		fi := 0
		for i := 1; i < len(pts)/3-1; i++ {
			fill[fi] = pts[0]
			fill[fi+1] = pts[1]
			fill[fi+2] = pts[2]
			fi += 3
			fill[fi] = pts[i*3]
			fill[fi+1] = pts[i*3+1]
			fill[fi+2] = pts[i*3+2]
			fi += 3
			fill[fi] = pts[i*3+3]
			fill[fi+1] = pts[i*3+4]
			fill[fi+2] = pts[i*3+5]
			fi += 3
		}
		r.submit(gl.TRIANGLES, fill, false)
	}
	if style == Stroke || style == Both {
		r.submit(gl.LINE_LOOP, pts, true)
	}
}

// DrawSolid zeichnet die expandierten Kanten/Flächen eines Solids mit der
// übergebenen View-/Welt-Matrix. Solid bleibt damit reine Daten – das
// Zeichnen liegt bewusst beim Renderer statt in einem Solid.Draw, das auf
// einen globalen Renderer zugegriffen hat.
func (r *Renderer) DrawSolid(s *Solid, view, world *Mat4x4) {
	vw := view.Mult(world)
	r.SetModelview(&vw)
	if len(s.FlatFaces) > 0 {
		r.submit(gl.TRIANGLES, s.FlatFaces, false) // Flächen (Fill, beleuchtet)
	}
	r.submit(gl.LINES, s.FlatEdges, true) // Kanten (Stroke, unbelichtet)
}

// DrawBody setzt Farben/Linienbreite des Bodies und zeichnet dessen Solid.
func (r *Renderer) DrawBody(b *Body, view *Mat4x4) {
	world := b.modelMatrix()
	r.SetStrokeWidth(b.LineWidth)
	r.SetStrokeColorHex(b.Color)
	r.SetFillColorHex(b.Color)
	r.DrawSolid(b.Solid, view, &world)
}
