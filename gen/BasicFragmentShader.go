// This shader file was generated using go:generate go run codegen/gen_shaders.go shaders/fragment/basic.fragment.shader BasicFragmentShader
// 2022-08-01 17:58:19.041453461 -0400 EDT m=+0.006774795"
package shaders
const BasicFragmentShader = `#version 330
out vec4 outColor;
in vec3 fragmentColor;
void main()
{
	outColor = vec4(fragmentColor, 1.0);
}`+"\x00"