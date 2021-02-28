package glutil

import (
	"fmt"
	"image"
	"image/draw"
	"image/png"
	"os"

	"github.com/go-gl/gl/v2.1/gl"
)

const (
	// TextureWrapClampToBorder shows no texture data at all outside the normalized uv range.
	TextureWrapClampToBorder TextureWrap = gl.CLAMP_TO_BORDER
	// TextureWrapClampToEdge shows the nearest pixel inside the texture outside the normalized uv range.
	TextureWrapClampToEdge TextureWrap = gl.CLAMP_TO_EDGE
	// TextureWrapRepeat repeats the texture.
	TextureWrapRepeat TextureWrap = gl.REPEAT
	// TextureWrapMirroredRepeat repeats the texture but flips adjacent tiles.
	TextureWrapMirroredRepeat TextureWrap = gl.MIRRORED_REPEAT
	// TextureFilterNearest denotes nearest-neighbour interpolation.
	TextureFilterNearest TextureFilter = gl.NEAREST
	// TextureFilterLinear denotes linear interpolation.
	TextureFilterLinear TextureFilter = gl.LINEAR
)

// TextureWrap denotes the behaviour of texture interpolation when leaving the normalized uv range 0...1
type TextureWrap int

// TextureFilter denotes how to interpolate a texture.
type TextureFilter int

// TextureParameters combines typical texture options.
type TextureParameters struct {
	WrapS     TextureWrap
	WrapT     TextureWrap
	MinFilter TextureFilter
	MagFilter TextureFilter
}

// TexParam returns a TextureParameters object which sets the same parameters in both axes.
func TexParam(wrap TextureWrap, filter TextureFilter) TextureParameters {
	return TextureParameters{wrap, wrap, filter, filter}
}

// Texture represents a single texture loaded to graphics memory.
type Texture struct {
	Tex           uint32
	Width, Height int
}

// AspectRatio returns width divided by height.
func (t *Texture) AspectRatio() float32 {
	return float32(t.Width) / float32(t.Height)
}

// TextureFromFile generates a new OpenGL texture and fills it with image data from file.
func TextureFromFile(file string, params TextureParameters) (*Texture, error) {
	img, err := imageFromFile(file)
	if err != nil {
		return nil, err
	}

	return TextureFromImage(img, params)
}

func imageFromFile(file string) (image.Image, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	//TODO file extension switch
	img, err := png.Decode(f)
	if err != nil {
		return nil, err
	}

	return img, nil
}

// TextureFromImage generates a new OpenGL texture and fills it with image data from a generic image.
func TextureFromImage(img image.Image, params TextureParameters) (*Texture, error) {
	rgba := image.NewRGBA(img.Bounds())
	draw.Draw(rgba, rgba.Bounds(), img, image.Pt(0, 0), draw.Src)
	return TextureFromRGBA(rgba, params)
}

// TextureFromRGBA generates a new OpenGL texture and fills it with image data from an RGBA image.
func TextureFromRGBA(rgba *image.RGBA, params TextureParameters) (*Texture, error) {
	var tex uint32
	gl.GenTextures(1, &tex)
	if errNum := gl.GetError(); errNum != 0 {
		return nil, fmt.Errorf("generate texture buffer: %d", errNum)
	}

	gl.BindTexture(gl.TEXTURE_2D, tex)
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_S, int32(params.WrapS))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_WRAP_T, int32(params.WrapT))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MIN_FILTER, int32(params.MinFilter))
	gl.TexParameteri(gl.TEXTURE_2D, gl.TEXTURE_MAG_FILTER, int32(params.MagFilter))
	gl.TexImage2D(gl.TEXTURE_2D, 0, gl.RGBA, int32(rgba.Bounds().Dx()), int32(rgba.Bounds().Dy()), 0, gl.RGBA, gl.UNSIGNED_BYTE, gl.Ptr(rgba.Pix))
	if errNum := gl.GetError(); errNum != 0 {
		return nil, fmt.Errorf("write image data to graphics memory: %d", errNum)
	}
	gl.BindTexture(gl.TEXTURE_2D, 0)

	return &Texture{tex, rgba.Bounds().Size().X, rgba.Bounds().Size().Y}, nil
}
