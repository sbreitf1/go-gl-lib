package main

import (
	"fmt"
	"runtime"

	"github.com/sbreitf1/go-gl-lib/gl2d"
	"github.com/sbreitf1/go-gl-lib/glui"

	"github.com/sirupsen/logrus"
)

const (
	windowWidth          = 800
	windowHeight         = 600
	updateExpectedImages = true
)

type testDefinition struct {
	Name       string
	RenderFunc func()
}

func main() {
	runtime.LockOSThread()

	mainWindow, err := glui.Init(windowWidth, windowHeight, "gl2d-rendertest")
	if err != nil {
		logrus.Fatalf("failed to init main window: %s", err.Error())
	}
	defer glui.Terminate()

	if err := gl2d.Init(); err != nil {
		logrus.Fatalf("failed to init gl2d: %s", err.Error())
	}
	defer gl2d.Terminate()

	tests := []testDefinition{
		{"empty", func() {}},
	}

	currentTest := 0
	errorCount := 0
	mainWindow.EnterLayer(&glui.ContextLayerWrapper{
		RenderHandler: func(w *glui.MainWindow) {
			if currentTest >= len(tests) {
				// no tests left
				w.Close()
				return
			}

			gl2d.Begin(w.GetSize())
			defer gl2d.End()

			if err := gl2dTest(w, tests[currentTest]); err != nil {
				logrus.Errorf("test %q failed: %s", tests[currentTest].Name, err.Error())
				errorCount++
			}
			currentTest++
		},
	})
	mainWindow.Run()

	if errorCount > 0 {
		gl2d.Terminate()
		glui.Terminate()
		logrus.Fatalf("%d of %d tests failed have failed", errorCount, len(tests))
	} else {
		logrus.Infof("tests successful, exiting test application")
	}
}

func gl2dTest(w *glui.MainWindow, test testDefinition) error {
	return fmt.Errorf("not implemented")
}
