package lib3d

import (
	_ "embed"
	"fmt"
	"os"

	"github.com/go-gl/gl/v3.3-core/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
)

//go:embed shaders/vert.glsl
var vertShaderSrc string

//go:embed shaders/frag.glsl
var fragShaderSrc string

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

	// Die Shader-Quellen sind per go:embed in das Binary eingebettet und
	// funktionieren damit unabhängig vom aktuellen Arbeitsverzeichnis.
	r.program = r.createProgram(vertShaderSrc, fragShaderSrc)
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
