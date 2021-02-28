package gl2d

import (
	"fmt"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/mathgl/mgl32"
)

var (
	initialized               bool
	canvasWidth, canvasHeight int
	clipRect                  Quad

	projectionMatrix mgl32.Mat4
)

// Init initializes all OpenGL buffers and should be called once after OpenGL is initialized.
func Init() error {
	useShadersDefault()

	if err := initCircles(); err != nil {
		return fmt.Errorf("init gl2d circles: %s", err.Error())
	}
	if err := initImages(); err != nil {
		return fmt.Errorf("init gl2d images: %s", err.Error())
	}
	if err := initLines(); err != nil {
		return fmt.Errorf("init gl2d lines: %s", err.Error())
	}
	if err := initRectangles(); err != nil {
		return fmt.Errorf("init gl2d rectangles: %s", err.Error())
	}
	if err := initText(); err != nil {
		return fmt.Errorf("init gl2d text: %s", err.Error())
	}

	initialized = true
	return nil
}

// Terminate releases all OpenGL buffers and should be called when the application is about to exit.
func Terminate() {
	terminateLines()
	terminateCircles()
	terminateRectangles()
	terminateText()
}

// Begin starts rendering a single frame.
func Begin(width, height int) {
	if !initialized {
		panic("need to call gl2d.Init before gl2d.Begin")
	}

	canvasWidth = width
	canvasHeight = height

	projectionMatrix = mgl32.Ortho2D(0, float32(width), float32(height), 0)

	gl.Disable(gl.DEPTH_TEST)
	gl.Enable(gl.BLEND)
	gl.BlendFunc(gl.SRC_ALPHA, gl.ONE_MINUS_SRC_ALPHA)

	ResetClipRect()
}

// SetClipRect will skip rendering outside of the given clip rectangle.
func SetClipRect(q Quad) {
	clipRect = q
}

// ResetClipRect allows to render on the whole screen area.
func ResetClipRect() {
	clipRect = Quad{
		Left:   0,
		Right:  float32(canvasWidth - 1),
		Top:    0,
		Bottom: float32(canvasHeight - 1),
	}
}

// End ends the current 2D frame.
func End() {
	gl.UseProgram(0)
	gl.Disable(gl.BLEND)
}
