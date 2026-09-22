// GLSL for the plate quads. The noise texture is generated in JS (see makeNoiseTexture),
// so nothing here ships as an asset.
//
// The fragment shader does four things to a photograph, in this order: the pointer
// zoom, the highlight bleed of film halation, a print-like grade, and grain. All of it
// is deliberately gentle — these are somebody's photographs, and the effect exists to
// make them read as prints rather than as files.

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
	uniform float uGrade;      // 0 → 1, master mix for the print grade
	uniform float uGrain;      // grain amplitude

	varying vec2 vUv;

	const vec3 LUMA = vec3(0.2126, 0.7152, 0.0722);

	// Lifted blacks, a shallow S-curve, and a little warmth in the highlights: what a
	// print does, not what a negative holds. Kept under 20% so nothing clips.
	vec3 printed(vec3 c) {
		c = c * 0.94 + 0.03;
		c = mix(c, c * c * (3.0 - 2.0 * c), 0.2);
		c *= vec3(1.02, 0.996, 0.97);
		return c;
	}

	void main() {
		// Zoom in UV space rather than by scaling the quad, which would spill over the
		// caption and the neighbouring plates.
		vec2 uv = (vUv - 0.5) * (1.0 - 0.025 * uHover) + 0.5;

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

		// Halation: light that passed through the emulsion and came back off its backing,
		// which reads as a warm bloom hugging the brightest edges. A four-tap cross is
		// enough at this radius — the point is the tint, not a real blur.
		float radius = 0.011;
		vec3 bleed = texture2D(uTexture, uv + vec2(radius, 0.0)).rgb
			+ texture2D(uTexture, uv - vec2(radius, 0.0)).rgb
			+ texture2D(uTexture, uv + vec2(0.0, radius * 0.75)).rgb
			+ texture2D(uTexture, uv - vec2(0.0, radius * 0.75)).rgb;
		float bloom = smoothstep(0.6, 1.0, dot(bleed * 0.25, LUMA));

		colour = mix(colour, printed(colour), uGrade);
		colour += vec3(1.0, 0.52, 0.3) * bloom * 0.3 * uGrade;

		// Grain last, and heavier in the midtones than in either end, which is where film
		// actually shows it.
		float n = texture2D(uNoise, uv * vec2(2.4) + vec2(uTime * 0.021, uTime * 0.014)).r
			+ texture2D(uNoise, uv * vec2(4.7) - vec2(uTime * 0.017, 0.0)).r;
		float weight = mix(0.32, 1.0, 1.0 - abs(dot(colour, LUMA) * 2.0 - 1.0));

		colour += (n - 1.0) * 0.5 * uGrain * weight;

		gl_FragColor = vec4(colour, 1.0);
	}
`;
