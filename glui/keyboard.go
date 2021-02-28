package glui

import (
	"github.com/go-gl/glfw/v3.3/glfw"
)

const (
	KeyUp        = Key(glfw.KeyUp)
	KeyDown      = Key(glfw.KeyDown)
	KeyLeft      = Key(glfw.KeyLeft)
	KeyRight     = Key(glfw.KeyRight)
	KeyA         = Key(glfw.KeyA)
	KeyD         = Key(glfw.KeyD)
	KeyE         = Key(glfw.KeyE)
	KeyK         = Key(glfw.KeyK)
	KeyS         = Key(glfw.KeyS)
	KeyW         = Key(glfw.KeyW)
	KeyY         = Key(glfw.KeyY)
	KeyZ         = Key(glfw.KeyZ)
	KeyLeftShift = Key(glfw.KeyLeftShift)
	KeySpace     = Key(glfw.KeySpace)
	KeyEscape    = Key(glfw.KeyEscape)
	KeyEnter     = Key(glfw.KeyEnter)

	ModShift    = ModifierKey(glfw.ModShift)
	ModAlt      = ModifierKey(glfw.ModAlt)
	ModControl  = ModifierKey(glfw.ModControl)
	ModCapsLock = ModifierKey(glfw.ModCapsLock)
)

var (
	knownKeys = []Key{
		KeyEscape,
		KeyUp, KeyDown, KeyLeft, KeyRight,
		KeyA, KeyD, KeyE, KeyK, KeyS, KeyW, KeyY, KeyZ,
		KeySpace, KeyEnter,
		KeyLeftShift,
	}
)

// Key represents a keyboard key.
type Key glfw.Key

// ModifierKey represents a modifier key like Ctrl, Shift or Alt.
type ModifierKey glfw.ModifierKey

// Has returns true when the given modifier key is pressed.
func (mods ModifierKey) Has(key ModifierKey) bool {
	return (mods & key) == key
}
