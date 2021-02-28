package gl2d

import (
	"github.com/sbreitf1/go-gl-lib/glutil"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	fillRectangleProg uint32
	drawRectangleProg uint32
)

func initRectangles() error {
	var err error
	if fillRectangleProg, err = glutil.AssembleShaderFromSource(glutil.ShaderSource{
		Vertex:   defaultVertexShader,
		Fragment: fillRectangleFragmentShader,
	}); err != nil {
		return err
	}

	drawRectangleProg, err = glutil.AssembleShaderFromSource(glutil.ShaderSource{
		Vertex:   defaultVertexShader,
		Fragment: drawRectangleFragmentShader,
	})
	return err
}

func terminateRectangles() {
	gl.DeleteProgram(fillRectangleProg)
	gl.DeleteProgram(drawRectangleProg)
}

// FillRectangle renders a filled rectangle.
func FillRectangle(topLeft, size mgl32.Vec2, color Color) {
	setupRectangleProgram(fillRectangleProg, topLeft, size, color)

	renderQuad(Quad{
		Left:   floor32(topLeft[0]) - ceil32(rBlend),
		Right:  ceil32(topLeft[0]+size[0]) + ceil32(rBlend),
		Top:    floor32(topLeft[1]) - ceil32(rBlend),
		Bottom: ceil32(topLeft[1]+size[1]) + ceil32(rBlend),
	})
}

// DrawRectangle renders an outlined rectangle. The outline center exactly represents the rectangle defined by size, thus the visible outermost rectangle size is increased by lineWidth in both dimensions.
func DrawRectangle(topLeft, size mgl32.Vec2, lineWidth float32, color Color) {
	// respect pixel offset
	topLeft = topLeft.Add([2]float32{0.5, 0.5})
	size = size.Sub([2]float32{1, 1})

	setupRectangleProgram(drawRectangleProg, topLeft, size, color)
	gl.Uniform1f(gl.GetUniformLocation(drawRectangleProg, gl.Str("halfLineWidth\x00")), lineWidth/2.0)

	renderQuad(Quad{
		Left:   floor32(topLeft[0]) - ceil32(lineWidth/2.0) - ceil32(rBlend),
		Right:  ceil32(topLeft[0]+size[0]) + ceil32(lineWidth/2.0) + ceil32(rBlend),
		Top:    floor32(topLeft[1]) - ceil32(lineWidth/2.0) - ceil32(rBlend),
		Bottom: ceil32(topLeft[1]+size[1]) + ceil32(lineWidth/2.0) + ceil32(rBlend),
	})
}

func setupRectangleProgram(prog uint32, topLeft, size mgl32.Vec2, color Color) {
	left := topLeft[0]
	right := left + size[0]
	top := topLeft[1]
	bottom := top + size[1]

	useProg(prog)
	gl.Uniform1fv(gl.GetUniformLocation(prog, gl.Str("left\x00")), 1, &left)
	gl.Uniform1fv(gl.GetUniformLocation(prog, gl.Str("right\x00")), 1, &right)
	gl.Uniform1fv(gl.GetUniformLocation(prog, gl.Str("top\x00")), 1, &top)
	gl.Uniform1fv(gl.GetUniformLocation(prog, gl.Str("bottom\x00")), 1, &bottom)
	gl.Uniform1f(gl.GetUniformLocation(prog, gl.Str("rBlend\x00")), rBlend)
	gl.Uniform4fv(gl.GetUniformLocation(prog, gl.Str("color\x00")), 1, &color[0])
}
