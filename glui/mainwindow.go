package glui

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.3/glfw"
	"github.com/sirupsen/logrus"
)

var (
	mainWindow *MainWindow
	m          sync.Mutex
)

//TODO set window icon from .ico file

// MainWindow provides an interface to the main window.
type MainWindow struct {
	glfwWindow                     *glfw.Window
	glVersionMajor, glVersionMinor int

	layers []ContextLayer

	maxSimStep   time.Duration
	totalSimTime time.Duration

	keyStates         map[Key]bool
	mouseButtonStates map[MouseButton]bool
	mouseX, mouseY    float64
	mouseCaptured     bool

	FixedPreFrameSleep     time.Duration
	FixedPollEventsTimeout time.Duration
}

// Init initializes GLFW and OpenGL with the given main window properties.
func Init(windowWidth, windowHeight int, windowTitle string) (*MainWindow, error) {
	m.Lock()
	defer m.Unlock()

	if mainWindow != nil {
		panic("cannot initialize ui twice")
	}

	logrus.Infof("init GLFW v%d.%d.%d", glfw.VersionMajor, glfw.VersionMinor, glfw.VersionRevision)
	if err := glfw.Init(); err != nil {
		return nil, fmt.Errorf("init GLFW: %s", err.Error())
	}

	glVersionMajor, glVersionMinor := 2, 1

	glfw.WindowHint(glfw.Resizable, glfw.True)
	glfw.WindowHint(glfw.ContextVersionMajor, glVersionMajor)
	glfw.WindowHint(glfw.ContextVersionMinor, glVersionMinor)
	glfw.WindowHint(glfw.OpenGLProfile, glfw.OpenGLAnyProfile)
	glfw.WindowHint(glfw.OpenGLForwardCompatible, glfw.False)
	glfwWindow, err := glfw.CreateWindow(windowWidth, windowHeight, windowTitle, nil, nil)
	if err != nil {
		return nil, err
	}
	glfwWindow.MakeContextCurrent()

	if err = gl.Init(); err != nil {
		return nil, fmt.Errorf("init OpenGL: %s", err.Error())
	}

	glVersionStr := gl.GoStr(gl.GetString(gl.VERSION))
	logrus.Infof("using OpenGL v%d.%d [%s]", glVersionMajor, glVersionMinor, glVersionStr)

	gl.Viewport(0, 0, int32(windowWidth), int32(windowHeight))

	keyStates := make(map[Key]bool)
	for _, k := range knownKeys {
		keyStates[k] = false
	}
	mouseButtonStates := make(map[MouseButton]bool)
	for _, k := range knownMouseButtons {
		mouseButtonStates[k] = false
	}

	mainWindow = &MainWindow{
		glVersionMajor:    glVersionMajor,
		glVersionMinor:    glVersionMinor,
		glfwWindow:        glfwWindow,
		layers:            make([]ContextLayer, 0),
		keyStates:         keyStates,
		mouseButtonStates: mouseButtonStates,
	}
	mainWindow.setupCallbacks()

	logrus.Infof("ui is now initialized")
	return mainWindow, nil
}

// Terminate releases all GLFW and OpenGL related ressources.
func Terminate() {
	m.Lock()
	defer m.Unlock()

	if mainWindow == nil {
		panic("cannot terminate uninitialized ui")
	}

	mainWindow.glfwWindow.Destroy()
	glfw.Terminate()
	mainWindow = nil

	logrus.Infof("ui has been terminated")
}

// Run enters the main loop of the application.
func (w *MainWindow) Run() {
	// https://www.iditect.com/how-to/53890601.html
	glfw.SwapInterval(0)

	//TODO detect best sleeping times using the old frame time and desired refresh rate

	begin := time.Now()
	last := time.Duration(0)
	for !w.glfwWindow.ShouldClose() {
		time.Sleep(w.FixedPreFrameSleep)

		t := time.Since(begin)
		dt := (t - last)
		if w.maxSimStep > 0 && dt > w.maxSimStep {
			dt = w.maxSimStep
		}
		w.totalSimTime += dt
		last = t

		glfw.WaitEventsTimeout(w.FixedPollEventsTimeout.Seconds())

		gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)

		for _, c := range w.layers {
			c.Update(w, dt.Seconds())
		}

		for _, c := range w.layers {
			c.Render(w)
		}

		w.glfwWindow.SwapBuffers()
	}

	// now gracefully close remaining contexts:
	for len(w.layers) > 0 {
		w.LeaveUppermostLayer()
	}
}

// Close will close the window on next update.
func (w *MainWindow) Close() {
	w.glfwWindow.SetShouldClose(true)
}

// EnterLayer sets the uppermost layer.
func (w *MainWindow) EnterLayer(c ContextLayer) {
	w.layers = append(w.layers, c)
	c.Enter(w)
}

// IsUppermostLayer returns true when the given layer is the current uppermost layer.
func (w *MainWindow) IsUppermostLayer(c ContextLayer) bool {
	//TODO find correct answer
	return true
}

// LeaveUppermostLayer leaves the current uppermost layer.
func (w *MainWindow) LeaveUppermostLayer() {
	if len(w.layers) > 0 {
		w.layers[len(w.layers)-1].Leave(w)
		w.layers = w.layers[:len(w.layers)-1]
	}
}

func (w *MainWindow) setupCallbacks() {
	w.glfwWindow.SetSizeCallback(w.cbResize)
	w.glfwWindow.SetKeyCallback(w.cbKey)
	w.glfwWindow.SetCharCallback(w.cbChar)
	w.glfwWindow.SetMouseButtonCallback(w.cbMouseButton)
	w.glfwWindow.SetCursorPosCallback(w.cbMouseMove)
}

func (w *MainWindow) cbResize(_ *glfw.Window, width int, height int) {
	gl.Viewport(0, 0, int32(width), int32(height))
}

func (w *MainWindow) cbKey(_ *glfw.Window, key glfw.Key, scancode int, action glfw.Action, mods glfw.ModifierKey) {
	switch action {
	case glfw.Press:
		if _, ok := w.keyStates[Key(key)]; ok {
			w.keyStates[Key(key)] = true
		}
		for i := len(w.layers) - 1; i >= 0; i-- {
			if w.layers[i].KeyDown(w, Key(key), ModifierKey(mods)) {
				break
			}
		}

	case glfw.Repeat:
		for i := len(w.layers) - 1; i >= 0; i-- {
			if w.layers[i].KeyPress(w, Key(key), ModifierKey(mods)) {
				break
			}
		}

	case glfw.Release:
		if _, ok := w.keyStates[Key(key)]; ok {
			w.keyStates[Key(key)] = false
		}
		for i := len(w.layers) - 1; i >= 0; i-- {
			if w.layers[i].KeyUp(w, Key(key), ModifierKey(mods)) {
				break
			}
		}
	}
}

func (w *MainWindow) cbChar(_ *glfw.Window, char rune) {
	for i := len(w.layers) - 1; i >= 0; i-- {
		if w.layers[i].EnterRune(w, char) {
			break
		}
	}
}

func (w *MainWindow) cbMouseButton(_ *glfw.Window, button glfw.MouseButton, action glfw.Action, mods glfw.ModifierKey) {
	switch action {
	case glfw.Press:
		if _, ok := w.mouseButtonStates[MouseButton(button)]; ok {
			w.mouseButtonStates[MouseButton(button)] = true
		}
		for i := len(w.layers) - 1; i >= 0; i-- {
			if w.layers[i].MouseDown(w, int(w.mouseX), int(w.mouseY), MouseButton(button)) {
				break
			}
		}

	case glfw.Release:
		if _, ok := w.mouseButtonStates[MouseButton(button)]; ok {
			w.mouseButtonStates[MouseButton(button)] = false
		}
		for i := len(w.layers) - 1; i >= 0; i-- {
			if w.layers[i].MouseUp(w, int(w.mouseX), int(w.mouseY), MouseButton(button)) {
				break
			}
		}
	}
}

func (w *MainWindow) cbMouseMove(_ *glfw.Window, xpos float64, ypos float64) {
	w.mouseX = xpos
	w.mouseY = ypos

	if w.mouseCaptured {
		width, height := w.glfwWindow.GetSize()
		centerX := float64(width) / 2.0
		centerY := float64(height) / 2.0
		dx := xpos - centerX
		dy := ypos - centerY
		w.glfwWindow.SetCursorPos(centerX, centerY)
		for i := len(w.layers) - 1; i >= 0; i-- {
			if w.layers[i].MouseMoveCaptured(w, dx, dy) {
				break
			}
		}

	} else {
		for i := len(w.layers) - 1; i >= 0; i-- {
			if w.layers[i].MouseMove(w, int(xpos), int(ypos)) {
				break
			}
		}
	}
}

// IsKeyDown returns true when the key is currently beeing pressed. Will panic for unknown keys.
func (w *MainWindow) IsKeyDown(key Key) bool {
	isDown, ok := w.keyStates[key]
	if !ok {
		panic(fmt.Sprintf("key %v is not available", key))
	}
	return isDown
}

// MousePos returns the current mouse cursor location inside the client area of the main window. This method should not be used while captured mouse mode is activated.
func (w *MainWindow) MousePos() (int, int) {
	return int(w.mouseX), int(w.mouseY)
}

// IsMouseButtonDown returns true when the given mouse button is currently beeing pressed.
func (w *MainWindow) IsMouseButtonDown(button MouseButton) bool {
	isDown, ok := w.mouseButtonStates[button]
	if !ok {
		panic(fmt.Sprintf("mouse button %v is not available", button))
	}
	return isDown
}

// IsMouseCaptured returns true when the mouse is currently in captured mode.
func (w *MainWindow) IsMouseCaptured() bool {
	return w.mouseCaptured
}

// SetMouseCaptured activates or deactivates captured mouse mode.
//
// In captured mouse mode, the mouse cursor is not visible and the callback function MouseMoveCaptured instead of MouseMove is called.
func (w *MainWindow) SetMouseCaptured(capture bool) {
	if capture {
		w.glfwWindow.SetInputMode(glfw.CursorMode, glfw.CursorDisabled)
		if glfw.RawMouseMotionSupported() {
			w.glfwWindow.SetInputMode(glfw.RawMouseMotion, glfw.True)
		}
		width, height := w.glfwWindow.GetSize()
		w.glfwWindow.SetCursorPos(float64(width)/2.0, float64(height)/2.0)
		w.mouseCaptured = true

	} else {
		if glfw.RawMouseMotionSupported() {
			w.glfwWindow.SetInputMode(glfw.RawMouseMotion, glfw.False)
		}
		w.glfwWindow.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
		w.mouseCaptured = false
	}
}

// GetSize returns the current client area size of the main window.
func (w *MainWindow) GetSize() (int, int) {
	return w.glfwWindow.GetSize()
}

// MaxSimStep returns the largest time delta to be processed in a single frame.
func (w *MainWindow) MaxSimStep() time.Duration {
	return w.maxSimStep
}

// SetMaxSimStep limits the time delta to be processed in a single frame. This can be used to prevent simulation discontinuities when moving or resizing the main window.
func (w *MainWindow) SetMaxSimStep(dt time.Duration) {
	w.maxSimStep = dt
}

// TotalSimTime returns the total time that has passed in the simulation.
func (w *MainWindow) TotalSimTime() time.Duration {
	return w.totalSimTime
}
