#version 330 core

uniform int   uMode;
uniform vec4  uColor;
uniform vec4  uColor2;
uniform float uTime;
uniform vec3  uShapeCenter;
uniform float uShapeRadius;

uniform float uFogNear;
uniform float uFogFar;
uniform vec4  uFogColor;

uniform int   uLighted;
uniform vec3  uLightDir;

in vec3 vCamPos;
out vec4 fragColor;

void main() {
    vec3 base = uColor.rgb;

    if (uMode == 1) {
        float d = distance(vCamPos, uShapeCenter);
        float t = clamp(d / max(uShapeRadius, 1.0), 0.0, 1.0);
        base = mix(uColor, uColor2, t).rgb;
    } else if (uMode == 2) {
        float brightness = 0.6 + 0.4 * sin(uTime * 3.0);
        base = uColor.rgb * brightness;
    }

    // Flächige Schattierung (Headlight): Licht kommt aus Kamerarichtung.
    // Die Facetten-Normale wird pro Pixel aus den Kamera-Raum-Ableitungen
    // rekonstruiert. Bewusst OHNE gl_FrontFacing-Flip, damit die Flächen
    // schattiert bleiben (zur Kamera zeigende Flächen hell, wegdrehende
    // dunkler) – statt alles gleichmäßig auszuleuchten.
    if (uLighted == 1) {
        vec3 n = normalize(cross(dFdx(vCamPos), dFdy(vCamPos)));

        float diff = max(dot(n, normalize(uLightDir)), 0.0);
        base *= 0.25 + 0.75 * diff;
    }

    float depth = -vCamPos.z;
    float fog_t = clamp((depth - uFogNear) / (uFogFar - uFogNear), 0.0, 1.0);
    fragColor = mix(vec4(base, uColor.a), uFogColor, fog_t);
}
