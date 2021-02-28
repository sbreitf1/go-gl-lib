package glui

import "github.com/go-gl/glfw/v3.3/glfw"

const (
	// MouseButtonLeft denotes the left mouse button.
	MouseButtonLeft = MouseButton(glfw.MouseButtonLeft)
	// MouseButtonRight denotes the right mouse button.
	MouseButtonRight = MouseButton(glfw.MouseButtonRight)
	// MouseButtonMiddle denotes the middle mouse button.
	MouseButtonMiddle = MouseButton(glfw.MouseButtonMiddle)
)

var (
	knownMouseButtons = []MouseButton{MouseButtonLeft, MouseButtonRight, MouseButtonMiddle}
)

// MouseButton represents a mouse button.
type MouseButton glfw.MouseButton
