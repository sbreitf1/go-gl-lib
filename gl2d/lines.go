package gl2d

import (
	"github.com/sbreitf1/go-gl-lib/glutil"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	drawLineProg uint32
)

func initLines() error {
	var err error
	drawLineProg, err = glutil.AssembleShaderFromSource(glutil.ShaderSource{
		Vertex:   defaultVertexShader,
		Fragment: drawLineFragmentShader,
	})
	return err
}

func terminateLines() {
	gl.DeleteProgram(drawLineProg)
}

// DrawLine renders a single line.
func DrawLine(from, to mgl32.Vec2, lineWidth float32, color Color) {
	dir := to.Sub(from)
	length := dir.Len()
	dir = dir.Normalize()

	//TODO no line visible if from=to

	// respect pixel offset
	from = from.Add([2]float32{0.5, 0.5})

	useProg(drawLineProg)
	gl.Uniform2fv(gl.GetUniformLocation(drawLineProg, gl.Str("lineOffspring\x00")), 1, &from[0])
	gl.Uniform2fv(gl.GetUniformLocation(drawLineProg, gl.Str("lineDir\x00")), 1, &dir[0])
	gl.Uniform1f(gl.GetUniformLocation(drawLineProg, gl.Str("lineLength\x00")), length)
	gl.Uniform1f(gl.GetUniformLocation(drawLineProg, gl.Str("halfLineWidth\x00")), lineWidth/2.0)
	gl.Uniform1f(gl.GetUniformLocation(drawLineProg, gl.Str("rBlend\x00")), rBlend)
	gl.Uniform4fv(gl.GetUniformLocation(drawLineProg, gl.Str("color\x00")), 1, &color[0])

	renderQuad(Quad{
		Left:   floor32(min32(from.X(), to.X())) - ceil32(lineWidth/2.0) - ceil32(rBlend),
		Right:  ceil32(max32(from.X(), to.X())) + ceil32(lineWidth/2.0) + ceil32(rBlend),
		Top:    floor32(min32(from.Y(), to.Y())) - ceil32(lineWidth/2.0) - ceil32(rBlend),
		Bottom: ceil32(max32(from.Y(), to.Y())) + ceil32(lineWidth/2.0) + ceil32(rBlend),
	})
}
