package gl2d

import (
	"encoding/json"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"strings"

	"github.com/sbreitf1/go-gl-lib/glutil"

	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/mathgl/mgl32"
	"github.com/sirupsen/logrus"
	"golang.org/x/image/font"
	"golang.org/x/image/font/basicfont"
	"golang.org/x/image/math/fixed"
)

const (
	defaultAlphabet = " abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZäöüÄÖÜ0123456789!\"§$%&/()=?,.-+#;:_*'<>|\\´`^°µ@€²³{}[]~"
)

var (
	defaultFont    *Font
	drawStringProg uint32
)

type runeRect struct {
	topLeft, bottomRight mgl32.Vec2
}

// BitmapFont denotes a font in main memory represented by pre-rendered rune images.
type BitmapFont struct {
	image      *image.RGBA
	runeRects  map[rune]runeRect
	lineHeight float32
	runeWidths map[rune]float32
	kernings   map[rune]map[rune]float32
}

type bitmapFontMetaData struct {
	RuneRects  map[rune][4]float32       `json:",omitempty"`
	LineHeight float32                   `json:",omitempty"`
	RuneWidths map[rune]float32          `json:",omitempty"`
	Kernings   map[rune]map[rune]float32 `json:",omitempty"`
}

// Export writes the font image and meta data to two separate files.
func (f *BitmapFont) Export(imageFile, metaFile string) error {
	fImage, err := os.Create(imageFile)
	if err != nil {
		return err
	}
	defer fImage.Close()

	if err := png.Encode(fImage, f.image); err != nil {
		return err
	}

	runeRects := make(map[rune][4]float32)
	for r, rect := range f.runeRects {
		runeRects[r] = [4]float32{rect.topLeft[0], rect.topLeft[1], rect.bottomRight[0], rect.bottomRight[1]}
	}
	data, err := json.Marshal(&bitmapFontMetaData{runeRects, f.lineHeight, f.runeWidths, f.kernings})
	if err != nil {
		return err
	}

	return ioutil.WriteFile(metaFile, data, os.ModePerm)
}

// Font denotes an OpenGL font represented by pre-rendered rune images.
type Font struct {
	texture    *glutil.Texture
	runeRects  map[rune]runeRect
	lineHeight float32
	runeWidths map[rune]float32
	kernings   map[rune]map[rune]float32
}

// Destroy released all ressources of this OpenGL font.
func (f *Font) Destroy() {
	gl.DeleteTextures(1, &f.texture.Tex)
}

func (f *Font) splitLines(str string) []string {
	//TODO remove or replace unknown chars (keep control chars \n and \t)
	return strings.Split(strings.ReplaceAll(strings.ReplaceAll(str, "\r\n", "\n"), "\r", "\n"), "\n")
}

// LineHeight denotes the vertical distance between two baselines
func (f *Font) LineHeight() float32 {
	return f.lineHeight
}

// RuneSize returns the size of the given rune.
func (f *Font) RuneSize(r rune) (float32, float32) {
	w, ok := f.runeWidths[r]
	if !ok {
		return 0, 0
	}
	return w, f.lineHeight
}

// Kern returns the spacing depending on the previous rune.
func (f *Font) Kern(r1, r2 rune) float32 {
	if f.kernings == nil {
		return 0
	}

	l1, ok := f.kernings[r1]
	if ok {
		if k, ok := l1[r2]; ok {
			return k
		}
	}
	return 0
}

// DefaultFont returns a default font.
func DefaultFont() *Font {
	return defaultFont
}

func initText() error {
	var err error
	if drawStringProg, err = glutil.AssembleShaderFromSource(glutil.ShaderSource{
		Vertex:   drawStringVertexShader,
		Fragment: drawStringFragmentShader,
	}); err != nil {
		return err
	}

	if defaultFont, err = NewFontFromFace(basicfont.Face7x13); err != nil {
		return err
	}
	return nil
}

func terminateText() {
	gl.DeleteProgram(drawStringProg)
	defaultFont.Destroy()
	defaultFont = nil
}

func int26ToFloat32(i fixed.Int26_6) float32 {
	return float32(float64(i) / 64)
}

func float32ToInt26(f float32) fixed.Int26_6 {
	return fixed.Int26_6(round32(f * 64))
}

// NewBitmapFontFromFace renders font face to a new BitmapFont.
func NewBitmapFontFromFace(face font.Face) (*BitmapFont, error) {
	m := face.Metrics()
	return generateBitmapFont(bitmapFontSource{
		LineHeight: int26ToFloat32(m.Height),
		RuneWidth: func(r rune) (float32, bool) {
			if w, ok := face.GlyphAdvance(r); ok {
				return int26ToFloat32(w), true
			}
			return 0, false
		},
		RenderRune: func(r rune, dst draw.Image, x, y float32) {
			d := &font.Drawer{
				Dst:  dst,
				Src:  image.NewUniform(color.RGBA{255, 255, 255, 255}),
				Face: face,
				Dot:  fixed.Point26_6{X: float32ToInt26(x), Y: float32ToInt26(y + int26ToFloat32(m.Ascent))},
			}
			d.DrawString(string([]rune{r}))
		},
		Kern: func(r1, r2 rune) float32 {
			return int26ToFloat32(face.Kern(r1, r2))
		},
	})
}

type bitmapFontSource struct {
	LineHeight float32
	RuneWidth  func(r rune) (float32, bool)
	RenderRune func(r rune, dst draw.Image, x, y float32)
	Kern       func(r1, r2 rune) float32
}

// NewBitmapFontFromFace renders font face to a new BitmapFont.
func generateBitmapFont(src bitmapFontSource) (*BitmapFont, error) {
	runeImages := make(map[rune]*image.RGBA)

	maxRuneHeight := src.LineHeight
	maxRuneWidth := float32(0)

	// measure runes
	for _, r := range defaultAlphabet {
		if width, ok := src.RuneWidth(r); ok {
			if width > maxRuneWidth {
				maxRuneWidth = width
			}

			//TODO draw rune directly to texture image in next loop
			img := image.NewRGBA(image.Rect(0, 0, int(ceil32(width)), int(ceil32(maxRuneHeight))))
			src.RenderRune(r, img, 0, 0)
			runeImages[r] = img
		} else {
			logrus.Warnf("rune %v not supported by selected font", r)
		}
	}

	// first guess of runes per line
	runesPerLine := int(math.Ceil(math.Sqrt(float64(len([]rune(defaultAlphabet))))))
	minTexWidth := 1 + (int(ceil32(maxRuneWidth+1)) * runesPerLine)
	texWidth := roundUpToPowerOfTwo(minTexWidth)
	// now refine by fitting maximum number of runes into one line (rounding up to power of two leads to unused space)
	runesPerLine = int(math.Floor(float64(texWidth-1) / float64(1+maxRuneWidth)))
	minTexWidth = 1 + (int(ceil32(maxRuneWidth+1)) * runesPerLine)
	// compute remaining parameters
	lineCount := int(math.Ceil(float64(len(runeImages)) / float64(runesPerLine)))
	minTexHeight := 1 + (int(ceil32(maxRuneHeight+1)) * lineCount)
	texHeight := roundUpToPowerOfTwo(minTexHeight)
	logrus.Debugf("font with %d supported runes (max rune size: %fx%f) represented in %dx%d grid -> %dx%d texture [%dx%d used]", len(runeImages), maxRuneWidth, maxRuneHeight, runesPerLine, lineCount, texWidth, texHeight, minTexWidth, minTexHeight)

	font := &BitmapFont{
		image:      image.NewRGBA(image.Rect(0, 0, texWidth, texHeight)),
		runeRects:  make(map[rune]runeRect),
		runeWidths: make(map[rune]float32),
		kernings:   make(map[rune]map[rune]float32),
	}
	// assemble texture image that holds all rune images at different locations
	line := 0
	linePos := 0
	for _, r := range defaultAlphabet {
		if img, ok := runeImages[r]; ok {
			// translate grid location to texture location
			top := 1 + line*int(ceil32(1+maxRuneHeight))
			left := 1 + linePos*int(ceil32(1+maxRuneWidth))
			bottom := top + img.Bounds().Size().Y
			right := left + img.Bounds().Size().X

			// draw rune image to texture image
			draw.Draw(font.image, image.Rect(left, top, right, bottom), img, image.Point{X: 0, Y: 0}, draw.Src)
			// remember normalized uv-coordinates of rune in texture image
			font.runeRects[r] = runeRect{
				topLeft:     [2]float32{float32(left) / float32(texWidth), float32(top) / float32(texHeight)},
				bottomRight: [2]float32{float32(right) / float32(texWidth), float32(bottom) / float32(texHeight)},
			}
			// and save rune metadata
			if w, ok := src.RuneWidth(r); ok {
				font.runeWidths[r] = w
			} else {
				font.runeWidths[r] = 0
			}
			// save rune pair specific kernings
			kernMap := map[rune]float32(nil)
			for _, rNext := range defaultAlphabet {
				k := src.Kern(r, rNext)
				if k != 0 {
					if kernMap == nil {
						kernMap = make(map[rune]float32)
					}
					kernMap[rNext] = k
				}
			}
			if kernMap != nil {
				font.kernings[r] = kernMap
			}

			// compute next rune coordinate in texture image
			linePos++
			if linePos >= runesPerLine {
				linePos = 0
				line++
			}
		} else {
			logrus.Debugf("unsupported rune '%v'", r)
		}
	}

	font.lineHeight = maxRuneHeight
	if len(font.kernings) == 0 {
		font.kernings = nil
	}

	// invisible pixels are often set to black which leads to ugly artifactes on interpolation
	// making the text nearly unreadable. set all color information to white so only the alpha
	// component represents the letters leading to much better results when scaling a font
	for y := 0; y < texHeight; y++ {
		for x := 0; x < texWidth; x++ {
			c := font.image.At(x, y)
			_, _, _, a := c.RGBA()
			font.image.Set(x, y, color.RGBA{R: 255, G: 255, B: 255, A: uint8(a)})
		}
	}

	return font, nil
}

// NewFontFromBitmapFont transfers the bitmap font image data to graphics memory.
func NewFontFromBitmapFont(font *BitmapFont) (*Font, error) {
	// transfer texture image to video memory
	tex, err := glutil.TextureFromRGBA(font.image, glutil.TextureParameters{
		WrapS:     glutil.TextureWrapRepeat,
		WrapT:     glutil.TextureWrapRepeat,
		MinFilter: glutil.TextureFilterLinear,
		MagFilter: glutil.TextureFilterLinear,
	})
	if err != nil {
		return nil, fmt.Errorf("generate font texture: %s", err.Error())
	}

	return &Font{
		texture:    tex,
		runeRects:  font.runeRects,
		lineHeight: font.lineHeight,
		runeWidths: font.runeWidths,
		kernings:   font.kernings,
	}, nil
}

// NewFontFromFace prepares an OpenGL font from the given font face.
func NewFontFromFace(face font.Face) (*Font, error) {
	bitmapFont, err := NewBitmapFontFromFace(face)
	if err != nil {
		return nil, err
	}

	return NewFontFromBitmapFont(bitmapFont)
}

func roundUpToPowerOfTwo(val int) int {
	if val == 0 {
		return 0
	}
	if val < 0 {
		panic("roundUpToPowerOfTwo called for negative number")
	}
	if val > 16777216 {
		panic("roundUpToPowerOfTwo for large number")
	}

	power := 1
	for {
		if power >= val {
			return power
		}
		power *= 2
	}
}

// DrawStringOptions can be used to overwrite string rendering behaviour.
type DrawStringOptions struct {
	//TODO center lines
	TabSpaces int
	Scale     float32
	RoundPos  bool
}

func getActualDrawStringOptions(opts *DrawStringOptions) DrawStringOptions {
	var actualOpts DrawStringOptions
	if opts != nil {
		actualOpts = *opts
	}
	if actualOpts.TabSpaces == 0 {
		actualOpts.TabSpaces = 8
	}
	if actualOpts.Scale == 0 {
		actualOpts.Scale = 1
	}
	return actualOpts
}

// DrawString renders the given string. Multiple lines are aligned left.
func DrawString(str string, pos mgl32.Vec2, font *Font, color Color, opts *DrawStringOptions) {
	actualOpts := getActualDrawStringOptions(opts)

	if actualOpts.RoundPos {
		pos[0] = round32(pos[0])
		pos[1] = round32(pos[1])
	}

	spaceWidth, _ := font.RuneSize(' ')
	tabWidth := (spaceWidth + 1) * float32(actualOpts.TabSpaces)
	//TODO fallback if space does not exist

	gl.UseProgram(drawStringProg)
	gl.UniformMatrix4fv(gl.GetUniformLocation(drawStringProg, gl.Str("projectionMatrix\x00")), 1, false, &projectionMatrix[0])
	gl.Uniform4fv(gl.GetUniformLocation(drawStringProg, gl.Str("color\x00")), 1, &color[0])

	gl.BindTexture(gl.TEXTURE_2D, font.texture.Tex)

	lineHeight := font.lineHeight

	// each line needs to be processed separately
	lines := font.splitLines(str)

	for i := range lines {
		y := pos[1] + actualOpts.Scale*float32(i)*lineHeight
		x := pos[0]
		previousRune := rune(0)
		for _, r := range lines[i] {
			if r == '\t' {
				// jump to next tab anchor instead of drawing
				x = float32(math.Floor(float64(x-pos[0])/float64(tabWidth)) + float64(tabWidth))
				previousRune = rune(0)

			} else {
				if rect, ok := font.runeRects[r]; ok {
					if previousRune != rune(0) {
						x += float32(font.Kern(previousRune, r))
					}
					w, h := font.RuneSize(r)
					renderTexturedQuad(Quad{
						Left:   x,
						Right:  x + actualOpts.Scale*float32(w),
						Top:    y,
						Bottom: y + actualOpts.Scale*float32(h),
					}, rect.topLeft, rect.bottomRight)
					x += actualOpts.Scale * float32(w)
				}
			}
		}
	}

	gl.BindTexture(gl.TEXTURE_2D, 0)
}

// MeasureString return the size of the given string rendered with a specific font.
func MeasureString(str string, font *Font, opts *DrawStringOptions) mgl32.Vec2 {
	actualOpts := getActualDrawStringOptions(opts)

	spaceWidth, _ := font.RuneSize(' ')
	tabWidth := (spaceWidth + 1) * float32(actualOpts.TabSpaces)
	//TODO fallback if space does not exist

	lines := font.splitLines(str)
	maxWidth := float32(0)
	for i := range lines {
		x := float32(0)
		previousRune := rune(0)
		for _, r := range lines[i] {
			if r == '\t' {
				// jump to next tab anchor instead of drawing
				x = float32(math.Floor(float64(x)/float64(tabWidth)) + float64(tabWidth))
				previousRune = rune(0)

			} else {
				if _, ok := font.runeRects[r]; ok {
					if previousRune != rune(0) {
						x += float32(font.Kern(previousRune, r))
					}
					w, _ := font.RuneSize(r)
					x += actualOpts.Scale * float32(w)
				}
			}
		}
		if x > maxWidth {
			maxWidth = x
		}
	}

	return [2]float32{maxWidth, actualOpts.Scale * float32(len(lines)) * font.lineHeight}
}
