package glutil

import (
	"fmt"
	"image"
	"image/draw"
	"math"
	"path/filepath"

	"github.com/go-gl/mathgl/mgl32"
	"github.com/sirupsen/logrus"
)

// Animation wraps a Texture and manages moving uv-coordinate frames to display an animation.
type Animation struct {
	Image         *Texture
	Width, Height int
	gridW, gridH  int
	frameCount    int
	duration      float64
	currentTime   float64
	Loop          bool
	FlipX         bool
	uvSize        mgl32.Vec2
}

// AnimationFromFileSequence reads an animation from multiple files like 00.png, 01.png, ...
func AnimationFromFileSequence(dir, format string, first, last int, duration float64, params TextureParameters) (*Animation, error) {
	frameCount := last - first + 1
	if frameCount < 1 {
		return nil, fmt.Errorf("invalid frame count %d", frameCount)
	}
	if duration <= 0 {
		return nil, fmt.Errorf("invalid duration %f", duration)
	}

	images := make([]image.Image, frameCount)
	for i := first; i <= last; i++ {
		img, err := imageFromFile(filepath.Join(dir, fmt.Sprintf(format, i)))
		if err != nil {
			return nil, err
		}
		if i > first {
			if img.Bounds() != images[0].Bounds() {
				return nil, fmt.Errorf("image %d has different dimensions", i)
			}
		}
		images[i-first] = img
	}

	imgW := images[0].Bounds().Size().X
	imgH := images[0].Bounds().Size().Y
	gridW := int(math.Round(math.Sqrt(float64(frameCount))))
	gridH := int(math.Ceil(float64(frameCount) / float64(gridW)))
	texW := gridW * imgW
	texH := gridH * imgH
	logrus.Debugf("arrange animation with %d frames in %dx%d grid -> %dx%d texture", frameCount, gridW, gridH, texW, texH)

	rgba := image.NewRGBA(image.Rect(0, 0, texW, texH))
	for i := 0; i < frameCount; i++ {
		posX := (i % gridW) * imgW
		posY := (i / gridW) * imgH
		draw.Draw(rgba, image.Rect(posX, posY, posX+imgW, posY+imgH), images[i], image.Pt(0, 0), draw.Src)
	}

	tex, err := TextureFromRGBA(rgba, params)
	if err != nil {
		return nil, err
	}

	return &Animation{
		Image:      tex,
		Width:      imgW,
		Height:     imgH,
		gridW:      gridW,
		gridH:      gridH,
		frameCount: frameCount,
		duration:   duration,
		uvSize:     [2]float32{1 / float32(gridW), 1 / float32(gridH)},
	}, nil
}

// AnimationFromLinewiseGridFile reads an animation from a single file where frames are arranged line by line.
func AnimationFromLinewiseGridFile(file string, frameCount int, duration float64, imgW, imgH, gridW, gridH int, params TextureParameters) (*Animation, error) {
	if gridW*gridH < frameCount {
		return nil, fmt.Errorf("animation grid of size %dx%d is too small to hold %d frames", gridW, gridH, frameCount)
	}

	tex, err := TextureFromFile(file, params)
	if err != nil {
		return nil, err
	}

	if tex.Width != (imgW*gridW) || tex.Height != (imgH*gridH) {
		return nil, fmt.Errorf("texture size %dx%d does not match expected texture size %dx%d based on grid", tex.Width, tex.Height, imgW*gridW, imgH*gridH)
	}

	return &Animation{
		Image:      tex,
		Width:      imgW,
		Height:     imgH,
		gridW:      gridW,
		gridH:      gridH,
		frameCount: frameCount,
		duration:   duration,
		uvSize:     [2]float32{1 / float32(gridW), 1 / float32(gridH)},
	}, nil
}

// CurrentTime returns the current animation time in seconds.
func (a *Animation) CurrentTime() float64 {
	return a.currentTime
}

// CurrentFrame returns the currently visible frame index.
func (a *Animation) CurrentFrame() int {
	currentFrame := int(math.Floor(float64(a.frameCount) * a.currentTime / a.duration))
	if currentFrame < 0 {
		currentFrame = 0
	}
	if currentFrame >= a.frameCount {
		currentFrame = a.frameCount - 1
	}
	return currentFrame
}

// Reset sets animation time to 0.
func (a *Animation) Reset() {
	a.currentTime = 0
}

// ToEnd sets animation time to end.
func (a *Animation) ToEnd() {
	a.currentTime = a.duration
}

// SetCurrentTime sets the current animation time.
func (a *Animation) SetCurrentTime(t float64) {
	a.currentTime = t
	if a.Loop {
		a.currentTime -= a.duration * math.Floor(t/a.duration)
	}
}

// SetCurrentFrame sets the current time to the beginning of a given frame index.
func (a *Animation) SetCurrentFrame(f int) {
	a.currentTime = float64(f) * a.duration / float64(a.frameCount)
}

// Update simulation a time step and sets animation parameters accordingly.
func (a *Animation) Update(dt float64) {
	a.currentTime += dt
	if a.Loop && a.currentTime >= a.duration {
		a.Repeat()
	}
}

// Repeat loops the animation once if it has reached it's end.
func (a *Animation) Repeat() {
	// reset animation, but keep time overflow
	for a.currentTime >= a.duration {
		a.currentTime -= a.duration
	}
}

// IsFinished returns whether the animation has reached it's end.
func (a *Animation) IsFinished() bool {
	return !a.Loop && a.currentTime >= a.duration
}

// GetSrcUV returns the uv-coordinates of the current frame.
func (a *Animation) GetSrcUV() (topLeft, bottomRight mgl32.Vec2) {
	currentFrame := a.CurrentFrame()
	topLeft = [2]float32{float32(currentFrame%a.gridW) / float32(a.gridW), float32(currentFrame/a.gridW) / float32(a.gridH)}
	bottomRight = [2]float32{topLeft[0] + a.uvSize[0], topLeft[1] + a.uvSize[1]}
	if a.FlipX {
		tmp := topLeft[0]
		topLeft[0] = bottomRight[0]
		bottomRight[0] = tmp
	}
	return
}
