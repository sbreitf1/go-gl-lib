package gl2d

func useShadersDefault() {

	defaultVertexShader = `#version 120

uniform mat4 projectionMatrix;
varying vec2 screenPos;

void main() {
	gl_Position = projectionMatrix*gl_Vertex;
	screenPos = gl_Vertex.xy;
}`

	fillCircleFragmentShader = `#version 120

varying vec2 screenPos;
uniform vec2 center;
uniform float radius;
uniform float rBlend = 0.5;
uniform vec4 color;

void main() {
	float d = distance(screenPos.xy,center)-0.5;
	if (d >= radius+rBlend) {
		discard;
	}
	if (d <= radius-rBlend) {
		gl_FragColor = color;
	} else {
		float f = (2*rBlend+radius-rBlend-d)/(2*rBlend);
		gl_FragColor = vec4(color.rgb, f*color.a);
	}
}`
	drawCircleFragmentShader = `#version 120

varying vec2 screenPos;
uniform vec2 center;
uniform float radius;
uniform float halfLineWidth;
uniform float rBlend = 0.5;
uniform vec4 color;

void main() {
	float d = abs(radius-distance(screenPos.xy,center));
	if (d <= halfLineWidth-rBlend) {
		gl_FragColor = color;
	} else if (d >= halfLineWidth+rBlend) {
		discard;
	} else {
		float f = (2*rBlend+halfLineWidth-rBlend-d)/(2*rBlend);
		gl_FragColor = vec4(color.rgb, f*color.a);
	}
}`

	drawImageVertexShader = `#version 120

uniform mat4 projectionMatrix;

void main() {
	gl_Position = projectionMatrix*gl_Vertex;
	gl_TexCoord[0] = gl_MultiTexCoord0;
}`
	drawImageFragmentShader = `#version 120

varying vec2 screenPos;
uniform sampler2D tex;
varying vec2 uv;
uniform vec4 color;

void main() {
	gl_FragColor = vec4(color*texture2D(tex, gl_TexCoord[0].st));
}`

	drawLineFragmentShader = `#version 120

varying vec2 screenPos;
uniform vec2 lineOffspring;
uniform vec2 lineDir;
uniform float lineLength;
uniform float halfLineWidth;
uniform float rBlend = 0.5;
uniform vec4 color;

void main() {
	vec2 dir = screenPos.xy-lineOffspring;
	float p = dot(lineDir,dir);
	float d;
	if (p >= 0 && p <= lineLength) {
		d = distance(lineOffspring+p*lineDir, screenPos.xy);
	} else if (p < 0) {
		d = distance(lineOffspring, screenPos.xy);
	} else {
		d = distance(lineOffspring+lineLength*lineDir, screenPos.xy);
	}
	if (d <= halfLineWidth-rBlend) {
		gl_FragColor = color;
	} else if (d >= halfLineWidth+rBlend) {
		discard;
	} else {
		float f = (2*rBlend+halfLineWidth-rBlend-d)/(2*rBlend);
		gl_FragColor = vec4(color.rgb, f*color.a);
	}
}`

	fillRectangleFragmentShader = `#version 120

varying vec2 screenPos;
uniform float left, right, top, bottom;
uniform float rBlend = 0.5;
uniform vec4 color;

void main() {
	float d = max(max(left-screenPos.x, screenPos.x-right), max(top-screenPos.y, screenPos.y-bottom));
	if (d <= -rBlend) {
		gl_FragColor = color;
	} else if (d >= rBlend) {
		discard;
	} else {
		float f = (2*rBlend-rBlend-d)/(2*rBlend);
		gl_FragColor = vec4(color.rgb, f*color.a);
	}
}`
	drawRectangleFragmentShader = `#version 120

varying vec2 screenPos;
uniform float left, right, top, bottom;
uniform float halfLineWidth;
uniform float rBlend = 0.5;
uniform vec4 color;

void main() {
	float d = abs(max(max(left-screenPos.x, screenPos.x-right), max(top-screenPos.y, screenPos.y-bottom)));
	if (d <= halfLineWidth-rBlend) {
		gl_FragColor = color;
	} else if (d >= halfLineWidth+rBlend) {
		discard;
	} else {
		float f = (2*rBlend+halfLineWidth-rBlend-d)/(2*rBlend);
		gl_FragColor = vec4(color.rgb, f*color.a);
	}
}`

	drawStringVertexShader = `#version 120

uniform mat4 projectionMatrix;

void main() {
	gl_Position = projectionMatrix*gl_Vertex;
	gl_TexCoord[0] = gl_MultiTexCoord0;
}`
	drawStringFragmentShader = `#version 120

varying vec2 screenPos;
uniform sampler2D tex;
varying vec2 uv;
uniform vec4 color;

void main() {
	gl_FragColor = vec4(color*texture2D(tex, gl_TexCoord[0].st));
}`

}
