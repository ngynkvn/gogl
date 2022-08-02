// This shader file was generated using go:generate go run codegen/gen_shaders.go shaders/fragment/basic.fragment.shader BasicFragmentShader
// 2022-08-01 21:07:37.588652581 -0400 EDT m=+0.003543807"
package shaders
const BasicFragmentShader = `#version 330
out vec4 outColor;
in vec3 fragmentColor;
void main()
{
	outColor = vec4(fragmentColor, 1.0);
}`+"\x00"