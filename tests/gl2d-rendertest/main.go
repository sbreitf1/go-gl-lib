package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"runtime"

	"github.com/sbreitf1/go-gl-lib/gl2d"
	"github.com/sbreitf1/go-gl-lib/glui"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/sirupsen/logrus"
)

const (
	windowWidth     = 800
	windowHeight    = 600
	tolerance       = 15
	maxAvgTolerance = 0.0001

	// activate this value to render new test cases
	createMissingImages = true
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
		{"colors", func() {
			gl2d.FillRectangle([2]float32{0, 0}, [2]float32{10, 10}, gl2d.Red)
			gl2d.FillRectangle([2]float32{10, 0}, [2]float32{10, 10}, gl2d.Green)
			gl2d.FillRectangle([2]float32{20, 0}, [2]float32{10, 10}, gl2d.Blue)
			gl2d.FillRectangle([2]float32{0, 10}, [2]float32{10, 10}, gl2d.Black)
			gl2d.FillRectangle([2]float32{10, 10}, [2]float32{10, 10}, gl2d.White)
			gl2d.FillRectangle([2]float32{20, 10}, [2]float32{10, 10}, gl2d.Yellow)
			gl2d.FillRectangle([2]float32{0, 20}, [2]float32{10, 10}, gl2d.Cyan)
			gl2d.FillRectangle([2]float32{10, 20}, [2]float32{10, 10}, gl2d.Magenta)
		}},
		{"all-shapes", func() {
			gl2d.FillCircle([2]float32{50, 50}, 40, gl2d.White)
			gl2d.DrawCircle([2]float32{150, 50}, 40, 1, gl2d.White)
			gl2d.DrawLine([2]float32{200, 20}, [2]float32{300, 80}, 1, gl2d.White)
			gl2d.FillRectangle([2]float32{20, 150}, [2]float32{80, 80}, gl2d.White)
			gl2d.DrawRectangle([2]float32{120, 150}, [2]float32{80, 80}, 1, gl2d.White)
		}},
		{"text", func() {
			gl2d.DrawString("foo bar", [2]float32{10, 10}, gl2d.DefaultFont(), gl2d.White, nil)
			gl2d.DrawString("bigger text", [2]float32{10, 40}, gl2d.DefaultFont(), gl2d.White, &gl2d.DrawStringOptions{Scale: 2})
			gl2d.DrawString("float pos", [2]float32{200.5, 10}, gl2d.DefaultFont(), gl2d.White, nil)
			gl2d.DrawString("float pos rounded", [2]float32{200.5, 40}, gl2d.DefaultFont(), gl2d.White, &gl2d.DrawStringOptions{RoundPos: true})
			//TODO tabs
			//TODO center
		}},
		{"complex-rgb", func() {
			gl2d.DrawCircle([2]float32{100, 100}, 50, 1, gl2d.Green)
			gl2d.FillCircle([2]float32{500, 500}, 50, gl2d.Blue)
			gl2d.FillRectangle([2]float32{100, 250}, [2]float32{300, 200}, gl2d.Red)
		}},
	}

	gl.ClearColor(0, 0, 0, 1)

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
		logrus.Fatalf("%d of %d tests have failed", errorCount, len(tests))
	} else {
		logrus.Infof("tests successful, exiting test application")
	}
}

func gl2dTest(w *glui.MainWindow, test testDefinition) error {
	test.RenderFunc()

	currentImage, err := getCurrentImageFromOpenGL(w)
	if err != nil {
		return fmt.Errorf("could not get current image: %s", err.Error())
	}

	expectedImgPath := "expected-" + test.Name + ".png"
	expectedImage, err := readPNG(expectedImgPath)
	if err != nil {
		if os.IsNotExist(err) {
			if err := writePNG(expectedImgPath, currentImage); err != nil {
				return err
			}
			// test should fail when the image was missing
			return fmt.Errorf("reference image was missing but has been created")
		}
		return err
	}

	maxDiff, avgDiff, diffImg, err := compareImages(expectedImage, currentImage)
	if err != nil {
		return err
	}

	diffImgPath := "diff-" + test.Name + ".png"
	if maxDiff > 0 || avgDiff > 0 {
		if err := writePNG(diffImgPath, diffImg); err != nil {
			return fmt.Errorf("export diff image: %s", err.Error())
		}
	} else {
		// remove old diff image if exists
		os.Remove(diffImgPath)
	}

	if maxDiff > tolerance || avgDiff > maxAvgTolerance {
		return fmt.Errorf("maxDiff: %f   ;   avgDiff: %f", maxDiff, avgDiff)
	}

	return nil
}

func getCurrentImageFromOpenGL(w *glui.MainWindow) (*image.RGBA, error) {
	width, height := w.GetSize()
	imgData := make([]uint8, 3*width*height)
	gl.ReadPixels(0, 0, int32(width), int32(height), gl.RGB, gl.UNSIGNED_BYTE, gl.Ptr(&imgData[0]))
	if glErr := gl.GetError(); glErr != gl.NO_ERROR {
		return nil, fmt.Errorf("gl.ReadPixels returned code %d", glErr)
	}

	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			pos := 3 * (x + y*width)
			img.Set(x, height-y-1, color.RGBA{
				R: imgData[pos+0],
				G: imgData[pos+1],
				B: imgData[pos+2],
				A: 255,
			})
		}
	}

	return img, nil
}

func writePNG(file string, img image.Image) error {
	f, err := os.Create(file)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

func readPNG(file string) (*image.RGBA, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	pngImage, err := png.Decode(f)
	if err != nil {
		return nil, err
	}

	img := image.NewRGBA(pngImage.Bounds())
	draw.Draw(img, pngImage.Bounds(), pngImage, image.Point{0, 0}, draw.Src)
	return img, nil
}

func compareImages(expected, actual *image.RGBA) (float64, float64, *image.RGBA, error) {
	if expected.Bounds() != actual.Bounds() {
		return 0, 0, nil, fmt.Errorf("expected image of size %v, but was %v", expected.Bounds().Size(), actual.Bounds().Size())
	}

	width := expected.Bounds().Size().X
	height := expected.Bounds().Size().X

	diff := image.NewRGBA(expected.Bounds())
	var maxDiff uint8
	var avgDiff int64
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			expectedColor := expected.At(x, y)
			actualColor := actual.At(x, y)
			diffColor := colorDiff(expectedColor, actualColor)
			diff.Set(x, y, diffColor)

			if diffColor.R > maxDiff {
				maxDiff = diffColor.R
			}
			if diffColor.G > maxDiff {
				maxDiff = diffColor.G
			}
			if diffColor.B > maxDiff {
				maxDiff = diffColor.B
			}
			avgDiff += int64(diffColor.R) + int64(diffColor.G) + int64(diffColor.B)
		}
	}
	return float64(maxDiff), float64(avgDiff) / float64(3*width*height), diff, nil
}

func colorDiff(c1, c2 color.Color) color.RGBA {
	r1, g1, b1, _ := c1.RGBA()
	r2, g2, b2, _ := c2.RGBA()
	abs := func(val int) uint8 {
		if val < 0 {
			return uint8(-val)
		}
		return uint8(val)
	}
	return color.RGBA{
		R: abs(int(r1) - int(r2)),
		G: abs(int(g1) - int(g2)),
		B: abs(int(b1) - int(b2)),
		A: 255,
	}
}
