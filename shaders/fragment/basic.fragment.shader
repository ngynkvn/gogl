#version 330
out vec4 outColor;
in vec3 fragmentColor;
void main()
{
	outColor = vec4(fragmentColor, 1.0);
}