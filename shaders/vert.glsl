#version 330 core

layout(location = 0) in vec3 aPos;

uniform mat4 uModelView;
uniform mat4 uProjection;
uniform float uPointSize;

out vec3 vCamPos;

void main() {
    vec4 camPos = uModelView * vec4(aPos, 1.0);
    vCamPos = camPos.xyz;
    gl_Position = uProjection * camPos;
    gl_PointSize = uPointSize;
}
