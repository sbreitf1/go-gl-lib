package gl2d

import (
	"github.com/sbreitf1/go-gl-lib/glutil"

	"github.com/go-gl/gl/v2.1/gl"
)

var (
	drawImageProg uint32
)

func initImages() error {
	var err error
	if drawImageProg, err = glutil.AssembleShaderFromSource(glutil.ShaderSource{
		Vertex:   drawImageVertexShader,
		Fragment: drawImageFragmentShader,
	}); err != nil {
		return err
	}
	return nil
}

func terminateImages() {
	gl.DeleteProgram(drawImageProg)
}

// DrawImage draws the full texture to the given quad and stretches the image.
func DrawImage(tex *glutil.Texture, dst Quad) {
	DrawColorizedImage(tex, dst, White)
}

// DrawColorizedImage draws the full texture to the given quad and stretches the image. Also allows to colorize the image.
func DrawColorizedImage(tex *glutil.Texture, dst Quad, color Color) {
	gl.UseProgram(drawImageProg)
	gl.UniformMatrix4fv(gl.GetUniformLocation(drawImageProg, gl.Str("projectionMatrix\x00")), 1, false, &projectionMatrix[0])
	gl.Uniform4fv(gl.GetUniformLocation(drawImageProg, gl.Str("color\x00")), 1, &color[0])

	gl.BindTexture(gl.TEXTURE_2D, tex.Tex)

	renderTexturedQuad(dst, [2]float32{0, 0}, [2]float32{1, 1})

	gl.BindTexture(gl.TEXTURE_2D, 0)
}

// DrawImageSrc draws a sub-rectangle of texture to the given quad and stretches the image.
func DrawImageSrc(tex *glutil.Texture, dst, srcUV Quad) {
	DrawColorizedImageSrc(tex, dst, srcUV, White)
}

// DrawColorizedImageSrc draws a sub-rectangle of texture to the given quad and stretches the image. Also allows to colorize the image.
func DrawColorizedImageSrc(tex *glutil.Texture, dst, srcUV Quad, color Color) {
	gl.UseProgram(drawImageProg)
	gl.UniformMatrix4fv(gl.GetUniformLocation(drawImageProg, gl.Str("projectionMatrix\x00")), 1, false, &projectionMatrix[0])
	gl.Uniform4fv(gl.GetUniformLocation(drawImageProg, gl.Str("color\x00")), 1, &color[0])

	gl.BindTexture(gl.TEXTURE_2D, tex.Tex)

	renderTexturedQuad(dst, [2]float32{srcUV.Left, srcUV.Top}, [2]float32{srcUV.Right, srcUV.Bottom})

	gl.BindTexture(gl.TEXTURE_2D, 0)
}

// DrawAnimation draws the currently visible from to the given quad and stretches the image.
func DrawAnimation(anim *glutil.Animation, dst Quad) {
	DrawColorizedAnimation(anim, dst, White)
}

// DrawColorizedAnimation draws the currently visible from to the given quad and stretches the image. Also allows to colorize the image.
func DrawColorizedAnimation(anim *glutil.Animation, dst Quad, color Color) {
	srcTopLeftUV, srcBottomRightUV := anim.GetSrcUV()
	DrawColorizedImageSrc(anim.Image, dst, Quad{
		Left:   srcTopLeftUV[0],
		Right:  srcBottomRightUV[0],
		Top:    srcTopLeftUV[1],
		Bottom: srcBottomRightUV[1],
	}, color)
}
