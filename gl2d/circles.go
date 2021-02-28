package gl2d

import (
	"github.com/sbreitf1/go-gl-lib/glutil"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	fillCircleProg uint32
	drawCircleProg uint32
)

func initCircles() error {
	var err error
	if fillCircleProg, err = glutil.AssembleShaderFromSource(glutil.ShaderSource{
		Vertex:   defaultVertexShader,
		Fragment: fillCircleFragmentShader,
	}); err != nil {
		return err
	}

	drawCircleProg, err = glutil.AssembleShaderFromSource(glutil.ShaderSource{
		Vertex:   defaultVertexShader,
		Fragment: drawCircleFragmentShader,
	})
	return err
}

func terminateCircles() {
	gl.DeleteProgram(fillCircleProg)
	gl.DeleteProgram(drawCircleProg)
}

// FillCircle renders a filled circle with given radius.
func FillCircle(center mgl32.Vec2, radius float32, color Color) {
	// respect pixel offset
	center = center.Add([2]float32{0.5, 0.5})

	setupCircleProgram(fillCircleProg, center, radius, color)

	renderQuad(Quad{
		Left:   floor32(center.X()) - ceil32(radius) - ceil32(rBlend),
		Right:  ceil32(center.X()) + ceil32(radius) + ceil32(rBlend),
		Top:    floor32(center.Y()) - ceil32(radius) - ceil32(rBlend),
		Bottom: ceil32(center.Y()) + ceil32(radius) + ceil32(rBlend),
	})
}

// DrawCircle renders an outlined circle. The radius denotes the midpoint of the outer ring and thus the outermost visible radius is radius+lineWidth/2.
func DrawCircle(center mgl32.Vec2, radius float32, lineWidth float32, color Color) {
	// respect pixel offset
	center = center.Add([2]float32{0.5, 0.5})

	setupCircleProgram(drawCircleProg, center, radius, color)
	gl.Uniform1f(gl.GetUniformLocation(drawCircleProg, gl.Str("halfLineWidth\x00")), lineWidth/2.0)

	renderQuad(Quad{
		Left:   floor32(center.X()) - ceil32(lineWidth/2.0) - ceil32(radius) - ceil32(rBlend),
		Right:  ceil32(center.X()) + ceil32(lineWidth/2.0) + ceil32(radius) + ceil32(rBlend),
		Top:    floor32(center.Y()) - ceil32(lineWidth/2.0) - ceil32(radius) - ceil32(rBlend),
		Bottom: ceil32(center.Y()) + ceil32(lineWidth/2.0) + ceil32(radius) + ceil32(rBlend),
	})
}

func setupCircleProgram(prog uint32, center mgl32.Vec2, radius float32, color Color) {
	useProg(prog)
	gl.Uniform2fv(gl.GetUniformLocation(prog, gl.Str("center\x00")), 1, &center[0])
	gl.Uniform1f(gl.GetUniformLocation(prog, gl.Str("radius\x00")), radius)
	gl.Uniform1f(gl.GetUniformLocation(prog, gl.Str("rBlend\x00")), rBlend)
	gl.Uniform4fv(gl.GetUniformLocation(prog, gl.Str("color\x00")), 1, &color[0])
}
