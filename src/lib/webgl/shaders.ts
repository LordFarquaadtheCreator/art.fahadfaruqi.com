// GLSL for the plate quads. The noise texture is generated in JS (see makeNoiseTexture),
// so nothing here ships as an asset.

export const plateVertex = /* glsl */ `
	varying vec2 vUv;

	void main() {
		vUv = uv;
		gl_Position = projectionMatrix * modelViewMatrix * vec4(position, 1.0);
	}
`;

export const plateFragment = /* glsl */ `
	uniform sampler2D uTexture;
	uniform sampler2D uNoise;
	uniform float uTime;
	uniform float uProgress;   // 0 at the viewport centre, 1 at the edges
	uniform float uHover;      // 0 → 1 while the pointer is on the plate

	varying vec2 vUv;

	void main() {
		vec2 uv = vUv;

		float warp_amount = 0.026 * uProgress + 0.05 * uHover;
		float split = 0.0035 * uProgress + 0.004 * uHover;

		vec2 warp = vec2(0.0);
		if (warp_amount > 0.0) {
			float n1 = texture2D(uNoise, uv * vec2(1.7, 1.0) + vec2(uTime * 0.012, uTime * 0.008)).r;
			float n2 = texture2D(uNoise, uv * vec2(2.3, 2.1) - vec2(uTime * 0.009, 0.0)).r;
			warp = (vec2(n1, n2) - 0.5) * warp_amount;
		}

		vec3 colour = vec3(
			texture2D(uTexture, uv + warp + vec2(split, 0.0)).r,
			texture2D(uTexture, uv + warp).g,
			texture2D(uTexture, uv + warp - vec2(split, 0.0)).b
		);

		gl_FragColor = vec4(colour, 1.0);
	}
`;
