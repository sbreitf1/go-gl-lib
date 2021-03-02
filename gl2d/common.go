package gl2d

import (
	"math"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	defaultVertexShader         string
	fillCircleFragmentShader    string
	drawCircleFragmentShader    string
	drawImageVertexShader       string
	drawImageFragmentShader     string
	drawLineFragmentShader      string
	fillRectangleFragmentShader string
	drawRectangleFragmentShader string
	drawStringVertexShader      string
	drawStringFragmentShader    string
)
var (
	// Red denotes opaque color red.
	Red Color = [4]float32{1, 0, 0, 1}
	// Green denotes opaque color green.
	Green Color = [4]float32{0, 1, 0, 1}
	// Blue denotes opaque color blue.
	Blue Color = [4]float32{0, 0, 1, 1}
	// Black denotes opaque color black.
	Black Color = [4]float32{0, 0, 0, 1}
	// DarkGray denotes opaque color dark gray.
	DarkGray Color = [4]float32{0.25, 0.25, 0.25, 1}
	// Gray denotes opaque color gray.
	Gray Color = [4]float32{0.5, 0.5, 0.5, 1}
	// White denotes opaque color white.
	White Color = [4]float32{1, 1, 1, 1}
	// Yellow denotes opaque color yellow.
	Yellow Color = [4]float32{1, 1, 0, 1}
	// Cyan denotes opaque color cyan.
	Cyan Color = [4]float32{0, 1, 1, 1}
	// Magenta denotes opaque color magenta.
	Magenta Color = [4]float32{1, 0, 1, 1}
)

var (
	rBlend = float32(0.5)
	//TODO what to do when lineWidth < rBlend ?
)

// Color represents an RGBA color.
type Color [4]float32

// MulColor returns a new color where the RGB components are multiplied with factor.
func (c Color) MulColor(factor float32) Color {
	return [4]float32{c[0] * factor, c[1] * factor, c[2] * factor, c[3]}
}

// Alpha returns a new color with modified alpha value.
func (c Color) Alpha(a float32) Color {
	return [4]float32{c[0], c[1], c[2], a}
}

func round32(val float32) float32 { return float32(math.Round(float64(val))) }
func floor32(val float32) float32 { return float32(math.Floor(float64(val))) }
func ceil32(val float32) float32  { return float32(math.Ceil(float64(val))) }
func min32(val1, val2 float32) float32 {
	if val1 < val2 {
		return val1
	}
	return val2
}
func max32(val1, val2 float32) float32 {
	if val1 > val2 {
		return val1
	}
	return val2
}

func useProg(prog uint32) {
	gl.UseProgram(prog)
	gl.UniformMatrix4fv(gl.GetUniformLocation(prog, gl.Str("projectionMatrix\x00")), 1, false, &projectionMatrix[0])
}

// Quad denotes an axis aligned rectangle.
type Quad struct {
	Left, Right, Top, Bottom float32
}

// TopLeft returns the top-left position of this quad.
func (q Quad) TopLeft() mgl32.Vec2 {
	return [2]float32{q.Left, q.Top}
}

// BottomRight returns the bottom-right position of this quad.
func (q Quad) BottomRight() mgl32.Vec2 {
	return [2]float32{q.Right, q.Bottom}
}

// Size returns the size of this quad.
func (q Quad) Size() mgl32.Vec2 {
	return [2]float32{q.Right - q.Left, q.Bottom - q.Top}
}

// Contains returns true, when the given 2d location is inside the quad.
func (q Quad) Contains(pos mgl32.Vec2) bool {
	return pos[0] >= q.Left && pos[0] <= q.Right && pos[1] >= q.Top && pos[1] <= q.Bottom
}

// FullScreenQuad returns a quad that covers the current canvas completely.
func FullScreenQuad() Quad {
	return Quad{0, float32(canvasWidth) - 1, 0, float32(canvasHeight) - 1}
}

// ShrinkAndCenterInside returns the largest quad that is centered and completely inside input quad q and has the same aspect ratio defined by width / height.
func ShrinkAndCenterInside(aspectRatio float32, q Quad) Quad {
	qw := q.Right - q.Left
	qh := q.Bottom - q.Top
	var w, h float32
	if float32(qw)/float32(qh) < aspectRatio {
		w = qw
		h = w / aspectRatio
	} else {
		h = qh
		w = aspectRatio * h
	}
	x := q.Left + float32(qw-w)/2
	y := q.Top + float32(qh-h)/2
	return Quad{
		Left:   x,
		Right:  x + w,
		Top:    y,
		Bottom: y + h,
	}
}

// ScaleCentered scales the quad while the center point remains at the same position.
func ScaleCentered(q Quad, factor float32) Quad {
	dw := ((q.Right-q.Left)*factor - (q.Right - q.Left)) / 2.0
	dh := ((q.Bottom-q.Top)*factor - (q.Bottom - q.Top)) / 2.0
	return Quad{
		Left:   q.Left - dw,
		Right:  q.Right + dw,
		Top:    q.Top - dh,
		Bottom: q.Bottom + dh,
	}
}

func renderQuad(q Quad) {
	if q.Right < clipRect.Left || q.Bottom < clipRect.Top || q.Left > clipRect.Right || q.Top > clipRect.Bottom {
		// quad is not visible on screen, nothing to draw
		return
	}

	gl.Begin(gl.QUADS)
	gl.Vertex3f(max32(clipRect.Left, q.Left), max32(clipRect.Top, q.Top), 0)
	gl.Vertex3f(min32(clipRect.Right, q.Right), max32(clipRect.Top, q.Top), 0)
	gl.Vertex3f(min32(clipRect.Right, q.Right), min32(clipRect.Bottom, q.Bottom), 0)
	gl.Vertex3f(max32(clipRect.Left, q.Left), min32(clipRect.Bottom, q.Bottom), 0)
	gl.End()
}

func renderTexturedQuad(q Quad, uvTopLeft, uvBottomRight mgl32.Vec2) {
	if q.Right < clipRect.Left || q.Bottom < clipRect.Top || q.Left > clipRect.Right || q.Top > clipRect.Bottom {
		// quad is not visible on screen, nothing to draw
		return
	}

	uvLeft := uvTopLeft[0]
	if q.Left < clipRect.Left {
		uvLeft = uvTopLeft[0] + (clipRect.Left-q.Left)/(q.Right-q.Left)*(uvBottomRight[0]-uvTopLeft[0])
	}
	uvTop := uvTopLeft[1]
	if q.Top < clipRect.Top {
		uvTop = uvTopLeft[1] + (clipRect.Top-q.Top)/(q.Bottom-q.Top)*(uvBottomRight[1]-uvTopLeft[1])
	}
	uvRight := uvBottomRight[0]
	if q.Right > clipRect.Right {
		uvRight = uvTopLeft[0] + (clipRect.Right-q.Left)/(q.Right-q.Left)*(uvBottomRight[0]-uvTopLeft[0])
	}
	uvBottom := uvBottomRight[1]
	if q.Bottom > clipRect.Bottom {
		uvBottom = uvTopLeft[1] + (clipRect.Bottom-q.Top)/(q.Bottom-q.Top)*(uvBottomRight[1]-uvTopLeft[1])
	}

	gl.Begin(gl.QUADS)
	gl.TexCoord2f(uvLeft, uvTop)
	gl.Vertex3f(max32(clipRect.Left, q.Left), max32(clipRect.Top, q.Top), 0)
	gl.TexCoord2f(uvRight, uvTop)
	gl.Vertex3f(min32(clipRect.Right, q.Right), max32(clipRect.Top, q.Top), 0)
	gl.TexCoord2f(uvRight, uvBottom)
	gl.Vertex3f(min32(clipRect.Right, q.Right), min32(clipRect.Bottom, q.Bottom), 0)
	gl.TexCoord2f(uvLeft, uvBottom)
	gl.Vertex3f(max32(clipRect.Left, q.Left), min32(clipRect.Bottom, q.Bottom), 0)
	gl.End()
}
