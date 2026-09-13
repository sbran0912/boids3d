package lib3d

import (
	"fmt"
	"os"

	"github.com/go-gl/gl/v3.3-core/gl"
)

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
