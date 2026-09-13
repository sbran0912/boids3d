package lib3d

import (
	"math"

	"github.com/go-gl/gl/v3.3-core/gl"
)

type drawState struct {
	Fill   colorState
	Stroke colorState
	LineW  float32
	Effect EffectMode
	Grad2  colorState
}

// drawCmd ist ein gesammelter Zeichenbefehl für eine Immediate-Primitive.
// Alle Befehle eines Frames werden gepuffert und am Frame-Ende mit einem
// einzigen wiederverwendbaren VBO gezeichnet (Batched Drawing).
type drawCmd struct {
	mode      uint32
	first     int32
	count     int32
	col       colorState
	effect    EffectMode
	grad2     colorState
	pointSize float32
	lineW     float32
	center    [3]float32
	radius    float32
	mv        [16]float32
	proj      [16]float32
	lit       uint32
}

// canMerge prüft, ob zwei aufeinanderfolgende Befehle identischen Zustand
// haben und zu einem DrawArrays-Aufruf zusammengefasst werden können.
// LINE_LOOP wird nie zusammengefasst, weil ein Loop sich selbst schließt.
func (c *drawCmd) canMerge(o drawCmd) bool {
	return c.mode == o.mode &&
		c.mode != gl.LINE_LOOP &&
		c.col == o.col &&
		c.effect == o.effect &&
		c.grad2 == o.grad2 &&
		c.pointSize == o.pointSize &&
		c.lineW == o.lineW &&
		c.center == o.center &&
		c.radius == o.radius &&
		c.mv == o.mv &&
		c.proj == o.proj &&
		c.lit == o.lit
}

// submit sammelt eine Immediate-Primitive im Frame-Batch.
// Es wird noch nichts an die GPU geschickt; das passiert erst beim
// flushBatch (Frame-Ende bzw. vor jedem Mesh-Draw).
func (r *Renderer) submit(mode uint32, verts []float32, useStroke bool) {
	col := r.state.Fill
	var lit uint32
	if useStroke {
		col = r.state.Stroke
	} else {
		lit = 1 // gefüllte Flächen werden mit Flat Shading beleuchtet
	}

	cmd := drawCmd{
		mode:      mode,
		first:     int32(len(r.batchVerts) / 3),
		count:     int32(len(verts) / 3),
		col:       col,
		effect:    r.state.Effect,
		grad2:     r.state.Grad2,
		pointSize: r.pointSize,
		lineW:     r.state.LineW,
		center:    r.gradCenter,
		radius:    r.gradRadius,
		mv:        r.mvUniform,
		proj:      r.projUniform,
		lit:       lit,
	}
	r.batchVerts = append(r.batchVerts, verts...)

	// Aufeinanderfolgende Befehle mit identischem Zustand zu einem
	// DrawArrays-Aufruf zusammenfassen (weniger Draw-Calls).
	if n := len(r.batchCmds); n > 0 && r.batchCmds[n-1].canMerge(cmd) {
		r.batchCmds[n-1].count += cmd.count
		return
	}
	r.batchCmds = append(r.batchCmds, cmd)
}

// flushBatch zeichnet alle gesammelten Primitives des Frames mit einem
// einzigen wiederverwendbaren VBO. Es gibt kein GenBuffers/DeleteBuffers
// pro Aufruf mehr; der GPU-Speicher wird nur beim Wachsen neu allokiert.
func (r *Renderer) flushBatch() {
	if len(r.batchCmds) == 0 {
		return
	}

	if r.batchBuf == 0 {
		gl.GenBuffers(1, &r.batchBuf)
	}
	gl.BindBuffer(gl.ARRAY_BUFFER, r.batchBuf)

	byteLen := len(r.batchVerts) * 4
	if byteLen > r.batchCap {
		// Nur beim Wachsen neu allozieren, sonst in-place ersetzen.
		gl.BufferData(gl.ARRAY_BUFFER, byteLen, gl.Ptr(r.batchVerts), gl.DYNAMIC_DRAW)
		r.batchCap = byteLen
	} else {
		gl.BufferSubData(gl.ARRAY_BUFFER, 0, byteLen, gl.Ptr(r.batchVerts))
	}

	gl.EnableVertexAttribArray(uint32(r.locPos))
	gl.VertexAttribPointer(uint32(r.locPos), 3, gl.FLOAT, false, 0, nil)

	for i := range r.batchCmds {
		c := &r.batchCmds[i]
		gl.UniformMatrix4fv(r.locModelView, 1, false, &c.mv[0])
		gl.UniformMatrix4fv(r.locProjection, 1, false, &c.proj[0])
		gl.Uniform1i(r.locMode, int32(c.effect))
		gl.Uniform1i(r.locLighted, int32(c.lit))
		gl.Uniform4f(r.locColor, c.col.R, c.col.G, c.col.B, c.col.A)
		gl.Uniform4f(r.locColor2, c.grad2.R, c.grad2.G, c.grad2.B, c.grad2.A)
		gl.Uniform1f(r.locPointSize, c.pointSize)
		gl.Uniform3f(r.locCenter, c.center[0], c.center[1], c.center[2])
		gl.Uniform1f(r.locRadius, c.radius)
		gl.LineWidth(c.lineW)
		gl.DrawArrays(c.mode, c.first, c.count)
	}

	r.batchVerts = r.batchVerts[:0]
	r.batchCmds = r.batchCmds[:0]
}

func shapeMetrics3d(pts []float32) (cx, cy, cz, r float32) {
	n := len(pts) / 3
	for i := 0; i < len(pts); i += 3 {
		cx += pts[i]
		cy += pts[i+1]
		cz += pts[i+2]
	}
	cx /= float32(n)
	cy /= float32(n)
	cz /= float32(n)

	for i := 0; i < len(pts); i += 3 {
		dx := pts[i] - cx
		dy := pts[i+1] - cy
		dz := pts[i+2] - cz
		dist := float32(math.Sqrt(float64(dx*dx + dy*dy + dz*dz)))
		if dist > r {
			r = dist
		}
	}
	return
}
