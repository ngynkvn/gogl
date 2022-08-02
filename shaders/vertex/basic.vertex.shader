#version 330
layout (location = 0) in vec3 Position;
layout(location = 1) in vec3 vertexColor;

// Model View Projection matrixes
uniform mat4 model;
uniform mat4 view;
uniform mat4 projection;

out vec3 fragmentColor;
void main()
{ 
    gl_Position = projection * view * model * vec4(Position, 1.0);
    fragmentColor = vertexColor;
}