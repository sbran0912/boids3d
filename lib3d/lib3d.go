package lib3d

import (
	"fmt"
	"math"
	"os"
	"strconv"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

const epsilon = 1e-10

type Vec3 struct {
	X, Y, Z float32
}

type Vec2 struct {
	X, Y, S float32
}

type Mat4x4 struct {
	M [16]float32
}

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

type Plane struct {
	Normal   Vec3
	Distance float32
	Boundary []Vec3
}

func NewVec3(x, y, z float32) Vec3 {
	return Vec3{x, y, z}
}

func (a Vec3) Add(b Vec3) Vec3 {
	return NewVec3(a.X+b.X, a.Y+b.Y, a.Z+b.Z)
}

func (a Vec3) Sub(b Vec3) Vec3 {
	return NewVec3(a.X-b.X, a.Y-b.Y, a.Z-b.Z)
}

func (a Vec3) Scale(s float32) Vec3 {
	return NewVec3(a.X*s, a.Y*s, a.Z*s)
}

func (a Vec3) SetMagnitude(m float32) Vec3 {
	l := a.Length()
	if l == 0 {
		return NewVec3(0, 0, 0)
	}
	return a.Scale(m / l)
}

func (a Vec3) Limit(max float32) Vec3 {
	mSq := a.SquaredLength()
	if mSq > max*max {
		return a.Scale(max / float32(math.Sqrt(float64(mSq))))
	}
	return a
}

func (a Vec3) Negate() Vec3 {
	return NewVec3(-a.X, -a.Y, -a.Z)
}

func (a Vec3) Dot(b Vec3) float32 {
	return a.X*b.X + a.Y*b.Y + a.Z*b.Z
}

func (a Vec3) Cross(b Vec3) Vec3 {
	return NewVec3(
		a.Y*b.Z-a.Z*b.Y,
		a.Z*b.X-a.X*b.Z,
		a.X*b.Y-a.Y*b.X,
	)
}

func (a Vec3) SquaredLength() float32 {
	return a.X*a.X + a.Y*a.Y + a.Z*a.Z
}

func (a Vec3) Length() float32 {
	return float32(math.Sqrt(float64(a.SquaredLength())))
}

func (a Vec3) Distance(b Vec3) float32 {
	return a.Sub(b).Length()
}

func (a Vec3) Normalize() Vec3 {
	l := a.Length()
	if l == 0 {
		return NewVec3(0, 0, 0)
	}
	return a.Scale(1 / l)
}

func (a Vec3) Lerp(b Vec3, t float32) Vec3 {
	return a.Add(b.Sub(a).Scale(t))
}

func (a Vec3) Clone() Vec3 {
	return a
}

func (a Vec3) Equals(b Vec3) bool {
	return math.Abs(float64(a.X-b.X)) < epsilon &&
		math.Abs(float64(a.Y-b.Y)) < epsilon &&
		math.Abs(float64(a.Z-b.Z)) < epsilon
}

func (v Vec3) Transform(m *Mat4x4) Vec3 {
	return NewVec3(
		m.M[0]*v.X+m.M[4]*v.Y+m.M[8]*v.Z+m.M[12],
		m.M[1]*v.X+m.M[5]*v.Y+m.M[9]*v.Z+m.M[13],
		m.M[2]*v.X+m.M[6]*v.Y+m.M[10]*v.Z+m.M[14],
	)
}

// TransformDir transformiert nur die Rotation eines Vektors (ohne
// Translation) – für Richtungen wie Licht-/Normalenvektoren.
// Wendet den Rotationsteil M·v an: für M = View-Matrix (deren Zeilen die
// Kamera-Basisvektoren sind) liefert das bereits Welt→Kamera; für eine
// Modell-Matrix ist es Modell→Welt.
func (v Vec3) TransformDir(m *Mat4x4) Vec3 {
	return NewVec3(
		m.M[0]*v.X+m.M[4]*v.Y+m.M[8]*v.Z,
		m.M[1]*v.X+m.M[5]*v.Y+m.M[9]*v.Z,
		m.M[2]*v.X+m.M[6]*v.Y+m.M[10]*v.Z,
	)
}

func mat4x4Identity() Mat4x4 {
	return Mat4x4{[16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}}
}

func mat4x4Translate(dx, dy, dz float32) Mat4x4 {
	return Mat4x4{[16]float32{
		1, 0, 0, 0,
		0, 1, 0, 0,
		0, 0, 1, 0,
		dx, dy, dz, 1,
	}}
}

func mat4x4Rotate(ax, ay, az float32) Mat4x4 {
	rx := Mat4x4{[16]float32{
		1, 0, 0, 0,
		0, float32(math.Cos(float64(ax))), float32(math.Sin(float64(ax))), 0,
		0, -float32(math.Sin(float64(ax))), float32(math.Cos(float64(ax))), 0,
		0, 0, 0, 1,
	}}

	ry := Mat4x4{[16]float32{
		float32(math.Cos(float64(ay))), 0, -float32(math.Sin(float64(ay))), 0,
		0, 1, 0, 0,
		float32(math.Sin(float64(ay))), 0, float32(math.Cos(float64(ay))), 0,
		0, 0, 0, 1,
	}}

	rz := Mat4x4{[16]float32{
		float32(math.Cos(float64(az))), float32(math.Sin(float64(az))), 0, 0,
		-float32(math.Sin(float64(az))), float32(math.Cos(float64(az))), 0, 0,
		0, 0, 1, 0,
		0, 0, 0, 1,
	}}

	ryRx := rz.Mult(&ry)
	return ryRx.Mult(&rx)
}

func (a *Mat4x4) Mult(b *Mat4x4) Mat4x4 {
	var r Mat4x4
	for i := 0; i < 4; i++ {
		for j := 0; j < 4; j++ {
			for k := 0; k < 4; k++ {
				r.M[i+j*4] += a.M[i+k*4] * b.M[k+j*4]
			}
		}
	}
	return r
}

func Mat4x4Lookat(cameraPos, target, up Vec3) Mat4x4 {
	forward := target.Sub(cameraPos).Normalize()
	right := forward.Cross(up).Normalize()
	realUp := right.Cross(forward)

	return Mat4x4{[16]float32{
		right.X, realUp.X, -forward.X, 0,
		right.Y, realUp.Y, -forward.Y, 0,
		right.Z, realUp.Z, -forward.Z, 0,
		-right.Dot(cameraPos),
		-realUp.Dot(cameraPos),
		forward.Dot(cameraPos),
		1,
	}}
}

func Mat4x4Perspective(fovY, aspect, znear, zfar float32) Mat4x4 {
	f := 1 / float32(math.Tan(float64(fovY)/2))
	return Mat4x4{[16]float32{
		f / aspect, 0, 0, 0,
		0, f, 0, 0,
		0, 0, -(zfar + znear) / (zfar - znear), -1,
		0, 0, -2 * znear * zfar / (zfar - znear), 0,
	}}
}

func (point Vec3) ToCamera(view, world *Mat4x4) Vec3 {
	vw := view.Mult(world)
	return point.Transform(&vw)
}

func (point Vec3) RotateAround(pivot Vec3, rotation *Mat4x4) Vec3 {
	rel := point.Sub(pivot)
	rotated := rel.Transform(rotation)
	return rotated.Add(pivot)
}

func (v Vec3) Project(fov float32) Vec2 {
	s := fov / (fov + v.Z)
	return Vec2{v.X * s, v.Y * s, s}
}

func planeCreate(normal Vec3, distance float32) Plane {
	return Plane{Normal: normal, Distance: distance}
}

func planeFromFace(faceVerts []Vec3) Plane {
	edge1 := faceVerts[1].Sub(faceVerts[0])
	edge2 := faceVerts[2].Sub(faceVerts[0])
	normal := edge1.Cross(edge2).Normalize()
	distance := -normal.Dot(faceVerts[0])

	boundary := make([]Vec3, len(faceVerts))
	copy(boundary, faceVerts)

	return Plane{
		Normal:   normal,
		Distance: distance,
		Boundary: boundary,
	}
}

func (p *Plane) DistanceToPoint(pnt Vec3) (float32, bool) {
	signedDist := p.Normal.Dot(pnt) + p.Distance
	closestPoint := pnt.Sub(p.Normal.Scale(signedDist))
	withinBoundary := true
	if len(p.Boundary) > 0 {
		withinBoundary = p.ContainsPoint(closestPoint)
	}
	return float32(math.Abs(float64(signedDist))), withinBoundary
}

func (p *Plane) IntersectLine(p1, p2 Vec3) (Vec3, bool) {
	dir := p2.Sub(p1)
	denom := p.Normal.Dot(dir)

	if math.Abs(float64(denom)) < 1e-10 {
		return Vec3{}, false
	}

	t := -(p.Normal.Dot(p1) + p.Distance) / denom

	if t < 0 || t > 1 {
		return Vec3{}, false
	}

	hit := p1.Add(dir.Scale(t))

	if len(p.Boundary) > 0 && !p.ContainsPoint(hit) {
		return Vec3{}, false
	}

	return hit, true
}

func (p *Plane) ContainsPoint(pnt Vec3) bool {
	for i := 0; i < len(p.Boundary); i++ {
		a := p.Boundary[i]
		b := p.Boundary[(i+1)%len(p.Boundary)]
		edge := b.Sub(a)
		toPoint := pnt.Sub(a)
		if edge.Cross(toPoint).Dot(p.Normal) < 0 {
			return false
		}
	}
	return true
}

type colorState struct {
	R, G, B, A float32
}

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

const maxStack = 64

// Renderer kapselt den gesamten OpenGL-/Fensterzustand.
type Renderer struct {
	window    *glfw.Window
	program   uint32
	screenW   float32
	screenH   float32
	startTime float64

	// Vollbild-Umschaltung (F11): Start im Fenstermodus,
	// Größe/Position merken für die Rückkehr zum Fenstermodus.
	isFullscreen       bool
	winPrevW, winPrevH int
	winPosX, winPosY   int

	mouseX      float32
	mouseY      float32
	mouseStatus int

	locPos        int32
	locModelView  int32
	locProjection int32
	locPointSize  int32
	locMode       int32
	locColor      int32
	locColor2     int32
	locTime       int32
	locCenter     int32
	locRadius     int32
	locFogNear    int32
	locFogFar     int32
	locFogColor   int32
	locLighted    int32
	locLightDir   int32

	stateStack    [maxStack]drawState
	stateStackTop int
	state         drawState

	// Batching: gesammelte Immediate-Primitives eines Frames.
	batchVerts []float32
	batchCmds  []drawCmd
	batchBuf   uint32
	batchCap   int

	// Zuletzt gesetzte Uniform-Werte (für den Batch-Snapshot pro Befehl).
	mvUniform   [16]float32
	projUniform [16]float32
	pointSize   float32
	gradCenter  [3]float32
	gradRadius  float32
}

var DefaultRenderer = &Renderer{}

func parseColorHex(hex string) colorState {
	c := colorState{1, 1, 1, 1}
	h := hex
	if len(h) > 0 && h[0] == '#' {
		h = h[1:]
	}

	if len(h) == 3 {
		buf := string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
		n, err := strconv.ParseUint(buf, 16, 32)
		if err != nil {
			return c
		}
		c.R = float32((n>>16)&0xff) / 255
		c.G = float32((n>>8)&0xff) / 255
		c.B = float32(n&0xff) / 255
		return c
	}

	if len(h) == 6 {
		n, err := strconv.ParseUint(h, 16, 32)
		if err != nil {
			return c
		}
		c.R = float32((n>>16)&0xff) / 255
		c.G = float32((n>>8)&0xff) / 255
		c.B = float32(n&0xff) / 255
		return c
	}

	return c
}

func parseColorRGB(r, g, b float32) colorState {
	return colorState{r / 255, g / 255, b / 255, 1}
}

func parseColorRGBA(r, g, b, a float32) colorState {
	return colorState{r / 255, g / 255, b / 255, a / 255}
}

func (r *Renderer) compileShader(shaderType uint32, src string) uint32 {
	shader := gl.CreateShader(shaderType)
	csources, free := gl.Strs(src + "\x00")
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var ok int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &ok)
	if ok == gl.FALSE {
		log := make([]byte, 1024)
		var length int32
		gl.GetShaderInfoLog(shader, 1024, &length, &log[0])
		fmt.Fprintf(os.Stderr, "Shader-Fehler: %s\n", string(log[:length]))
	}
	return shader
}

func (r *Renderer) createProgram(vertSrc, fragSrc string) uint32 {
	prog := gl.CreateProgram()
	vs := r.compileShader(gl.VERTEX_SHADER, vertSrc)
	fs := r.compileShader(gl.FRAGMENT_SHADER, fragSrc)
	gl.AttachShader(prog, vs)
	gl.AttachShader(prog, fs)
	gl.LinkProgram(prog)

	var ok int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &ok)
	if ok == gl.FALSE {
		log := make([]byte, 1024)
		var length int32
		gl.GetProgramInfoLog(prog, 1024, &length, &log[0])
		fmt.Fprintf(os.Stderr, "Programm-Fehler: %s\n", string(log[:length]))
	}

	gl.DeleteShader(vs)
	gl.DeleteShader(fs)
	return prog
}

func (m *Mat4x4) Flatten() [16]float32 {
	var dst [16]float32
	for i := 0; i < 16; i++ {
		dst[i] = m.M[i]
	}
	return dst
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

func (r *Renderer) cursorCallback(w *glfw.Window, xpos, ypos float64) {
	r.mouseX = float32(xpos) - r.screenW/2
	r.mouseY = -(float32(ypos) - r.screenH/2)
}

func (r *Renderer) mouseButtonCallback(w *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	if button == glfw.MouseButtonLeft {
		if action == glfw.Press {
			r.mouseStatus = 1
		}
		if action == glfw.Release {
			r.mouseStatus = 2
		}
	}
}

func (r *Renderer) keyCallback(w *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	if action != glfw.Press {
		return
	}

	if key == glfw.KeyEscape {
		w.SetShouldClose(true)
	}
	if key == glfw.KeyF11 {
		r.ToggleFullscreen()
	}
}

// framebufferSizeCallback wird bei jeder Größenänderung des Framebuffers
// aufgerufen (z. B. beim Umschalten Vollbild <-> Fenster oder beim Ziehen).
func (r *Renderer) framebufferSizeCallback(w *glfw.Window, width, height int) {
	r.screenW = float32(width)
	r.screenH = float32(height)
	gl.Viewport(0, 0, int32(width), int32(height))
}

func (r *Renderer) Init(w, h int) bool {
	if err := glfw.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "GLFW-Init fehlgeschlagen.")
		return false
	}

	glfw.WindowHint(glfw.ContextVersionMajor, 3)
	glfw.WindowHint(glfw.ContextVersionMinor, 3)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLCoreProfile)

	// Wir starten im Fenstermodus mit der gewünschten Größe w x h.
	// (Vollbild ist auf manchen Treibern/Wayland-Konfigurationen
	// problematisch beim Erstellen des GL-Kontexts.) Über F11 kann
	// später via ToggleFullscreen() in den Vollbildmodus gewechselt
	// werden.
	var err error
	r.window, err = glfw.CreateWindow(w, h, "lib3d_opengl_go", nil, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, "Fenster-Erstellung fehlgeschlagen.")
		glfw.Terminate()
		return false
	}

	r.window.MakeContextCurrent()
	if err := gl.Init(); err != nil {
		fmt.Fprintln(os.Stderr, "GL-Init fehlgeschlagen.")
		return false
	}
	r.window.SetCursorPosCallback(r.cursorCallback)
	r.window.SetMouseButtonCallback(r.mouseButtonCallback)
	r.window.SetFramebufferSizeCallback(r.framebufferSizeCallback)
	r.window.SetKeyCallback(r.keyCallback)

	fbW, fbH := r.window.GetFramebufferSize()
	r.framebufferSizeCallback(r.window, fbW, fbH)

	// Fenstergröße für die Rückkehr aus dem Vollbildmodus merken.
	r.isFullscreen = false
	r.winPrevW, r.winPrevH = w, h

	vertSrc, err := os.ReadFile("shaders/vert.glsl")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Shader-Dateien nicht gefunden.")
		return false
	}
	fragSrc, err := os.ReadFile("shaders/frag.glsl")
	if err != nil {
		fmt.Fprintln(os.Stderr, "Shader-Dateien nicht gefunden.")
		return false
	}

	r.program = r.createProgram(string(vertSrc), string(fragSrc))
	gl.UseProgram(r.program)

	r.locPos = gl.GetAttribLocation(r.program, gl.Str("aPos\x00"))
	r.locModelView = gl.GetUniformLocation(r.program, gl.Str("uModelView\x00"))
	r.locProjection = gl.GetUniformLocation(r.program, gl.Str("uProjection\x00"))
	r.locPointSize = gl.GetUniformLocation(r.program, gl.Str("uPointSize\x00"))
	r.locMode = gl.GetUniformLocation(r.program, gl.Str("uMode\x00"))
	r.locColor = gl.GetUniformLocation(r.program, gl.Str("uColor\x00"))
	r.locColor2 = gl.GetUniformLocation(r.program, gl.Str("uColor2\x00"))
	r.locTime = gl.GetUniformLocation(r.program, gl.Str("uTime\x00"))
	r.locCenter = gl.GetUniformLocation(r.program, gl.Str("uShapeCenter\x00"))
	r.locRadius = gl.GetUniformLocation(r.program, gl.Str("uShapeRadius\x00"))
	r.locFogNear = gl.GetUniformLocation(r.program, gl.Str("uFogNear\x00"))
	r.locFogFar = gl.GetUniformLocation(r.program, gl.Str("uFogFar\x00"))
	r.locFogColor = gl.GetUniformLocation(r.program, gl.Str("uFogColor\x00"))
	r.locLighted = gl.GetUniformLocation(r.program, gl.Str("uLighted\x00"))
	r.locLightDir = gl.GetUniformLocation(r.program, gl.Str("uLightDir\x00"))

	gl.Uniform1f(r.locPointSize, 4.0)
	gl.Uniform3f(r.locCenter, 0, 0, 0)
	gl.Uniform1f(r.locRadius, 1)
	gl.Uniform1i(r.locLighted, 0)
	gl.Uniform3f(r.locLightDir, 0, 0, 1) // Headlight: Licht aus Kamerarichtung (View-Space)

	identity := [16]float32{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
	gl.UniformMatrix4fv(r.locProjection, 1, false, &identity[0])
	gl.UniformMatrix4fv(r.locModelView, 1, false, &identity[0])

	var vao uint32
	gl.GenVertexArrays(1, &vao)
	gl.BindVertexArray(vao)

	gl.Enable(gl.DEPTH_TEST)
	gl.DepthFunc(gl.LEQUAL) // koplanare Kanten bleiben sichtbar (sonst Z-Fighting mit Flächen)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)
	gl.Enable(gl.PROGRAM_POINT_SIZE)

	gl.Viewport(0, 0, int32(r.screenW), int32(r.screenH))
	r.startTime = glfw.GetTime()
	r.stateStackTop = -1

	r.state.Fill = colorState{1, 1, 1, 1}
	r.state.Stroke = colorState{0, 0, 0, 1}
	r.state.LineW = 1
	r.state.Effect = EffectFlat
	r.state.Grad2 = colorState{0, 0, 0, 1}

	// Batch-Zustand initialisieren (entspricht den oben gesetzten Uniforms).
	r.mvUniform = [16]float32{1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1, 0, 0, 0, 0, 1}
	r.projUniform = r.mvUniform
	r.pointSize = 4.0
	r.gradCenter = [3]float32{0, 0, 0}
	r.gradRadius = 1
	r.batchBuf = 0
	r.batchCap = 0

	return true
}

// ToggleFullscreen schaltet zwischen Fenster- und Vollbildmodus um
// (Demo: Taste F11).
func (r *Renderer) ToggleFullscreen() {
	if r.window == nil {
		return
	}

	if !r.isFullscreen {
		monitor := glfw.GetPrimaryMonitor()
		if monitor == nil {
			return
		}
		mode := monitor.GetVideoMode()
		if mode == nil {
			return
		}

		// Aktuelle Fenstergröße/-position merken, damit zurückgeschaltet
		// werden kann.
		r.winPosX, r.winPosY = r.window.GetPos()
		r.winPrevW, r.winPrevH = r.window.GetSize()

		r.window.SetMonitor(monitor, 0, 0, mode.Width, mode.Height, mode.RefreshRate)
		r.isFullscreen = true
	} else {
		r.window.SetMonitor(nil, r.winPosX, r.winPosY, r.winPrevW, r.winPrevH, 0)
		r.isFullscreen = false
	}
}

func (r *Renderer) ShouldClose() bool {
	return r.window.ShouldClose()
}

func (r *Renderer) BeginFrame() {
	t := glfw.GetTime() - r.startTime
	gl.Uniform1f(r.locTime, float32(t))
}

func (r *Renderer) EndFrame() {
	r.flushBatch() // alle gesammelten Primitives des Frames zeichnen
	r.window.SwapBuffers()
	glfw.PollEvents()
}

// StartAnimation startet die Render-Schleife. draw wird in jedem Frame
// aufgerufen. Kehrt zurück, wenn das Fenster geschlossen wird (ESC/F11
// werden vom Renderer behandelt). Die Schleife kapselt BeginFrame/EndFrame,
// sodass der Aufrufer nur noch das Zeichnen (und ggf. die Simulation)
// bereitstellen muss.
func (r *Renderer) StartAnimation(draw func()) {
	for !r.window.ShouldClose() {
		r.BeginFrame()
		if draw != nil {
			draw()
		}
		r.EndFrame()
	}
	r.Close()
}

// Close beendet GLFW. Beim Beenden des GL-Kontexts werden alle GL-Objekte
// (Meshes, Shader, VAOs) automatisch freigegeben.
func (r *Renderer) Close() {
	r.flushBatch()
	if r.batchBuf != 0 {
		gl.DeleteBuffers(1, &r.batchBuf)
		r.batchBuf = 0
	}
	glfw.Terminate()
}

func (r *Renderer) SetFog(near, far, red, green, blue, alpha float32) {
	gl.Uniform1f(r.locFogNear, near)
	gl.Uniform1f(r.locFogFar, far)
	gl.Uniform4f(r.locFogColor, red, green, blue, alpha)
}

// SetLightDirection setzt die Lichtrichtung in Kameraraum.
// Muss pro Frame mit der aktuellen View-Matrix neu gesetzt werden,
// damit die Beleuchtung unabhängig von der Kameraausrichtung bleibt.
func (r *Renderer) SetLightDirection(x, y, z float32) {
	gl.Uniform3f(r.locLightDir, x, y, z)
}

func (r *Renderer) Width() int      { return int(r.screenW) }
func (r *Renderer) Height() int     { return int(r.screenH) }
func (r *Renderer) Aspect() float32 { return r.screenW / r.screenH }
func (r *Renderer) MouseX() float32 { return r.mouseX }
func (r *Renderer) MouseY() float32 { return r.mouseY }
func (r *Renderer) Time() float32   { return float32(glfw.GetTime() - r.startTime) }

func (r *Renderer) IsMouseDown() bool { return r.mouseStatus == 1 }

func (r *Renderer) IsMouseUp() bool {
	if r.mouseStatus == 2 {
		r.mouseStatus = 0
		return true
	}
	return false
}

func (r *Renderer) SetProjection(m *Mat4x4) {
	r.projUniform = m.Flatten()
	gl.UniformMatrix4fv(r.locProjection, 1, false, &r.projUniform[0])
}

func (r *Renderer) SetModelview(m *Mat4x4) {
	r.mvUniform = m.Flatten()
	gl.UniformMatrix4fv(r.locModelView, 1, false, &r.mvUniform[0])
}

func (r *Renderer) SetGradientCenter(cx, cy, cz, radius float32) {
	r.gradCenter = [3]float32{cx, cy, cz}
	r.gradRadius = radius
	gl.Uniform3f(r.locCenter, cx, cy, cz)
	gl.Uniform1f(r.locRadius, radius)
}

func (r *Renderer) Push() {
	if r.stateStackTop < maxStack-1 {
		r.stateStackTop++
		r.stateStack[r.stateStackTop] = r.state
	}
}

func (r *Renderer) Pop() {
	if r.stateStackTop >= 0 {
		r.state = r.stateStack[r.stateStackTop]
		r.stateStackTop--
	}
}

func (r *Renderer) SetFillColor(red, green, blue, alpha float32) {
	r.state.Fill = parseColorRGBA(red, green, blue, alpha)
}
func (r *Renderer) SetFillColorHex(hex string) { r.state.Fill = parseColorHex(hex) }
func (r *Renderer) SetStrokeColor(red, green, blue, alpha float32) {
	r.state.Stroke = parseColorRGBA(red, green, blue, alpha)
}
func (r *Renderer) SetStrokeColorHex(hex string) { r.state.Stroke = parseColorHex(hex) }
func (r *Renderer) SetStrokeWidth(width float32) { r.state.LineW = width }
func (r *Renderer) SetEffect(mode EffectMode)    { r.state.Effect = mode }
func (r *Renderer) SetGradient(red, green, blue, alpha float32) {
	r.state.Grad2 = parseColorRGBA(red, green, blue, alpha)
}
func (r *Renderer) SetGradientHex(hex string) { r.state.Grad2 = parseColorHex(hex) }

func (r *Renderer) Background(red, green, blue float32) {
	r.flushBatch()
	c := parseColorRGB(red, green, blue)
	gl.ClearColor(c.R, c.G, c.B, c.A)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (r *Renderer) BackgroundHex(hex string) {
	r.flushBatch()
	c := parseColorHex(hex)
	gl.ClearColor(c.R, c.G, c.B, c.A)
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (r *Renderer) SetPointSize(px float32) {
	r.pointSize = px
	gl.Uniform1f(r.locPointSize, px)
}

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

func (s *Solid) Draw(view, world *Mat4x4) {
	vw := view.Mult(world)
	DefaultRenderer.SetModelview(&vw)
	if len(s.FlatFaces) > 0 {
		DefaultRenderer.submit(gl.TRIANGLES, s.FlatFaces, false) // Flächen (Fill-Farbe, beleuchtet)
	}
	DefaultRenderer.submit(gl.LINES, s.FlatEdges, true) // Kanten (Stroke-Farbe, unbelichtet)
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
