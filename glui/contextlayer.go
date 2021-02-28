package glui

// ContextLayer represents a single view that can be displayed in a main window.
type ContextLayer interface {
	Enter(w *MainWindow)
	Leave(w *MainWindow)

	MouseDown(w *MainWindow, x, y int, button MouseButton) bool
	MouseUp(w *MainWindow, x, y int, button MouseButton) bool
	MouseMove(w *MainWindow, x, y int) bool
	MouseMoveCaptured(w *MainWindow, dx, dy float64) bool

	KeyDown(w *MainWindow, key Key, mods ModifierKey) bool
	KeyPress(w *MainWindow, key Key, mods ModifierKey) bool
	KeyUp(w *MainWindow, key Key, mods ModifierKey) bool
	EnterRune(w *MainWindow, r rune) bool

	Update(w *MainWindow, dt float64)
	Render(w *MainWindow)
}

// ContextLayerWrapper implements ContextLayer and calls handlers if defined. Use this for rapid prototyping.
type ContextLayerWrapper struct {
	EnterHandler             func(w *MainWindow)
	LeaveHandler             func(w *MainWindow)
	MouseDownHandler         func(w *MainWindow, x, y int, button MouseButton) bool
	MouseUpHandler           func(w *MainWindow, x, y int, button MouseButton) bool
	MouseMoveHandler         func(w *MainWindow, x, y int) bool
	MouseMoveCapturedHandler func(w *MainWindow, dx, dy float64) bool
	KeyDownHandler           func(w *MainWindow, key Key, mods ModifierKey) bool
	KeyPressHandler          func(w *MainWindow, key Key, mods ModifierKey) bool
	KeyUpHandler             func(w *MainWindow, key Key, mods ModifierKey) bool
	EnterRuneHandler         func(w *MainWindow, r rune) bool
	UpdateHandler            func(w *MainWindow, dt float64)
	RenderHandler            func(w *MainWindow)
}

// Enter calls c.EnterHandler
func (c *ContextLayerWrapper) Enter(w *MainWindow) {
	if c.EnterHandler != nil {
		c.EnterHandler(w)
	}
}

// Leave calls c.LeaveHandler
func (c *ContextLayerWrapper) Leave(w *MainWindow) {
	if c.LeaveHandler != nil {
		c.LeaveHandler(w)
	}
}

// MouseDown calls c.MouseDownHandler
func (c *ContextLayerWrapper) MouseDown(w *MainWindow, x, y int, button MouseButton) bool {
	if c.MouseDownHandler != nil {
		c.MouseDownHandler(w, x, y, button)
	}
	return false
}

// MouseUp calls c.MouseUpHandler
func (c *ContextLayerWrapper) MouseUp(w *MainWindow, x, y int, button MouseButton) bool {
	if c.MouseUpHandler != nil {
		c.MouseUpHandler(w, x, y, button)
	}
	return false
}

// MouseMove calls c.MouseMoveHandler
func (c *ContextLayerWrapper) MouseMove(w *MainWindow, x, y int) bool {
	if c.MouseMoveHandler != nil {
		c.MouseMoveHandler(w, x, y)
	}
	return false
}

// MouseMoveCaptured calls c.MouseMoveCapturedHandler
func (c *ContextLayerWrapper) MouseMoveCaptured(w *MainWindow, dx, dy float64) bool {
	if c.MouseMoveCapturedHandler != nil {
		return c.MouseMoveCapturedHandler(w, dx, dy)
	}
	return false
}

// KeyDown calls c.KeyDownHandler
func (c *ContextLayerWrapper) KeyDown(w *MainWindow, key Key, mods ModifierKey) bool {
	if c.KeyDownHandler != nil {
		return c.KeyDownHandler(w, key, mods)
	}
	return false
}

// KeyPress calls c.KeyPressHandler
func (c *ContextLayerWrapper) KeyPress(w *MainWindow, key Key, mods ModifierKey) bool {
	if c.KeyPressHandler != nil {
		return c.KeyPressHandler(w, key, mods)
	}
	return false
}

// KeyUp calls c.KeyUpHandler
func (c *ContextLayerWrapper) KeyUp(w *MainWindow, key Key, mods ModifierKey) bool {
	if c.KeyUpHandler != nil {
		return c.KeyUpHandler(w, key, mods)
	}
	return false
}

// EnterRune calls c.EnterRuneHandler
func (c *ContextLayerWrapper) EnterRune(w *MainWindow, r rune) bool {
	if c.EnterRuneHandler != nil {
		return c.EnterRuneHandler(w, r)
	}
	return false
}

// Update calls c.UpdateHandler
func (c *ContextLayerWrapper) Update(w *MainWindow, dt float64) {
	if c.UpdateHandler != nil {
		c.UpdateHandler(w, dt)
	}
}

// Render calls c.RenderHandler
func (c *ContextLayerWrapper) Render(w *MainWindow) {
	if c.RenderHandler != nil {
		c.RenderHandler(w)
	}
}
