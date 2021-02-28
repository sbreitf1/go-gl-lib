package glutil

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/sirupsen/logrus"
)

// ShaderSource defines the input data of a shader program.
type ShaderSource struct {
	Vertex   string
	Fragment string
}

// AssembleShaderFromFiles compiles and links a shader from input files.
func AssembleShaderFromFiles(src ShaderSource) (uint32, error) {
	var vertCode, fragCode string

	if len(src.Fragment) > 0 {
		vertData, err := ioutil.ReadFile(src.Vertex)
		if err != nil {
			return 0, fmt.Errorf("read vertex shader file: %s", err.Error())
		}
		vertCode = string(vertData)
	}

	if len(src.Fragment) > 0 {
		fragData, err := ioutil.ReadFile(src.Fragment)
		if err != nil {
			return 0, fmt.Errorf("read fragment shader file: %s", err.Error())
		}
		fragCode = string(fragData)
	}

	return AssembleShaderFromSource(ShaderSource{Vertex: vertCode, Fragment: fragCode})
}

// AssembleShaderFromSource compiles and links a shader from input sources.
func AssembleShaderFromSource(src ShaderSource) (uint32, error) {
	vertexShader, err := compileShader(src.Vertex, gl.VERTEX_SHADER)
	if err != nil {
		return 0, fmt.Errorf("vertex shader: %s", err.Error())
	}

	fragmentShader, err := compileShader(src.Fragment, gl.FRAGMENT_SHADER)
	if err != nil {
		gl.DeleteShader(vertexShader)
		return 0, fmt.Errorf("fragment shader: %s", err.Error())
	}

	prog := gl.CreateProgram()
	gl.AttachShader(prog, vertexShader)
	gl.AttachShader(prog, fragmentShader)
	gl.LinkProgram(prog)

	var linkStatus int32
	gl.GetProgramiv(prog, gl.LINK_STATUS, &linkStatus)
	if linkStatus == 0 {
		var logLength int32
		gl.GetProgramiv(prog, gl.INFO_LOG_LENGTH, &logLength)

		programLog := strings.Repeat("\x00", int(logLength+1))
		gl.GetProgramInfoLog(prog, logLength, nil, gl.Str(programLog))
		for _, line := range strings.Split(programLog, "\n") {
			line = strings.Trim(line, "\r\x00")
			if len(line) > 0 {
				logrus.Error(line)
			}
		}

		gl.DeleteProgram(prog)
		gl.DeleteShader(vertexShader)
		gl.DeleteShader(fragmentShader)

		return 0, fmt.Errorf("failed to link shader")
	}

	gl.DetachShader(prog, vertexShader)
	gl.DetachShader(prog, fragmentShader)
	gl.DeleteShader(vertexShader)
	gl.DeleteShader(fragmentShader)

	return prog, nil
}

func compileShader(source string, shaderType uint32) (uint32, error) {
	shader := gl.CreateShader(shaderType)

	csources, free := gl.Strs(source + "\x00")
	gl.ShaderSource(shader, 1, csources, nil)
	free()
	gl.CompileShader(shader)

	var status int32
	gl.GetShaderiv(shader, gl.COMPILE_STATUS, &status)
	if status == gl.FALSE {
		var logLength int32
		gl.GetShaderiv(shader, gl.INFO_LOG_LENGTH, &logLength)

		shaderLog := strings.Repeat("\x00", int(logLength+1))
		gl.GetShaderInfoLog(shader, logLength, nil, gl.Str(shaderLog))
		for _, line := range strings.Split(shaderLog, "\n") {
			line = strings.Trim(line, "\r\x00")
			if len(line) > 0 {
				logrus.Error(line)
			}
		}

		return 0, fmt.Errorf("shader compilation failed")
	}

	return shader, nil
}
