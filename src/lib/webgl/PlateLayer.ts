import {
	DataTexture,
	LinearFilter,
	Mesh,
	OrthographicCamera,
	PlaneGeometry,
	RedFormat,
	RepeatWrapping,
	SRGBColorSpace,
	Scene,
	ShaderMaterial,
	Texture,
	UnsignedByteType,
	WebGLRenderer
} from 'three';
import { plateFragment, plateVertex } from './shaders';

type Quad = {
	mesh: Mesh;
	material: ShaderMaterial;
	texture: Texture | null;
	image: HTMLImageElement;
	frame: HTMLElement;
	hover: number;
	hoverTarget: number;
	loading: boolean;
	onEnter: () => void;
	onLeave: () => void;
};

// Displacement and RGB split are switched on by the commit after this one, so a
// mis-registered quad shows up as a seam while the layer is still a pass-through.
const EFFECTS = true;

// The print: grade and halation, then grain over the top. Separate switches because
// they fail differently — a grade that is too strong flattens the photographs, grain
// that is too strong makes them look dirty, and anyone can turn either off here.
const GRADE = true;
const GRAIN = 0.05;

/** How far outside the viewport a plate keeps its texture resident, in CSS px. */
const MARGIN = 400;
/** Cap on resident textures: beyond this the furthest plates are released. */
const MAX_TEXTURES = 12;
const DPR_CAP = 2;

export class PlateLayer {
	private renderer: WebGLRenderer;
	private scene = new Scene();
	private camera: OrthographicCamera;
	private geometry = new PlaneGeometry(1, 1);
	private noise: Texture;
	private quads = new Map<HTMLElement, Quad>();
	private raf = 0;

	private width = 0;
	private height = 0;
	private disposed = false;
	private contextLost = false;
	private last = 0;

	constructor(private canvas: HTMLCanvasElement) {
		this.renderer = new WebGLRenderer({
			canvas,
			alpha: true,
			antialias: false,
			powerPreference: 'high-performance'
		});
		this.renderer.outputColorSpace = SRGBColorSpace;
		this.renderer.setClearAlpha(0);

		// World space is CSS pixels with the origin at the top-left, so quads can be
		// positioned straight from getBoundingClientRect().
		this.camera = new OrthographicCamera(0, 1, 0, 1, -1, 1);
		this.noise = makeNoiseTexture();
		this.resize();

		canvas.addEventListener('webglcontextlost', this.handleContextLost);
		window.addEventListener('resize', this.resize);
	}

	/** Adopt a plate element. Called from a Svelte action, so it runs per mount. */
	register(frame: HTMLElement): () => void {
		if (this.contextLost) return () => {};

		const image = frame.querySelector<HTMLImageElement>('.plate__image');
		if (!image) return () => {};

		const material = makeMaterial(this.noise);
		const mesh = new Mesh(this.geometry, material);
		mesh.visible = false;
		mesh.frustumCulled = false;

		const quad: Quad = {
			mesh,
			material,
			texture: null,
			image,
			frame,
			hover: 0,
			hoverTarget: 0,
			loading: false,
			onEnter: () => (quad.hoverTarget = 1),
			onLeave: () => (quad.hoverTarget = 0)
		};

		frame.addEventListener('pointerenter', quad.onEnter);
		frame.addEventListener('pointerleave', quad.onLeave);
		this.quads.set(frame, quad);
		this.scene.add(mesh);

		if (!this.raf) this.start();
		return () => this.release(frame);
	}

	start() {
		if (this.raf || this.disposed || this.contextLost) return;
		this.raf = requestAnimationFrame(this.tick);
	}

	dispose() {
		this.disposed = true;
		cancelAnimationFrame(this.raf);
		this.raf = 0;
		for (const frame of [...this.quads.keys()]) this.release(frame);
		this.geometry.dispose();
		this.noise.dispose();
		this.canvas.removeEventListener('webglcontextlost', this.handleContextLost);
		window.removeEventListener('resize', this.resize);
		this.renderer.dispose();
		this.canvas.remove();
	}

	private release(frame: HTMLElement) {
		const quad = this.quads.get(frame);
		if (!quad) return;

		frame.removeEventListener('pointerenter', quad.onEnter);
		frame.removeEventListener('pointerleave', quad.onLeave);
		this.scene.remove(quad.mesh);
		this.dropTexture(quad);
		quad.material.dispose();
		this.quads.delete(frame);
	}

	/** Release a plate's pixels and hand it back to the DOM image. */
	private dropTexture(quad: Quad) {
		const source = quad.texture?.image as ImageBitmap | undefined;
		if (source && typeof source.close === 'function') source.close();
		quad.texture?.dispose();
		quad.texture = null;
		quad.mesh.visible = false;
		quad.frame.removeAttribute('data-gl');
	}

	private handleContextLost = (event: Event) => {
		event.preventDefault();
		this.contextLost = true;
		delete document.documentElement.dataset.webgl;
		for (const quad of this.quads.values()) this.dropTexture(quad);
	};

	private resize = () => {
		this.width = window.innerWidth;
		this.height = window.innerHeight;
		this.renderer.setPixelRatio(Math.min(window.devicePixelRatio || 1, DPR_CAP));
		this.renderer.setSize(this.width, this.height, false);
		this.camera.right = this.width;
		this.camera.bottom = this.height;
		this.camera.updateProjectionMatrix();
		this.canvas.style.width = `${this.width}px`;
		this.canvas.style.height = `${this.height}px`;
	};

	private tick = (now: number) => {
		// A lost context — or the documented escape hatch — stops the loop for good:
		// continuing would re-upload textures onto a canvas that can never draw and keep
		// the DOM images hidden behind it.
		if (
			this.disposed ||
			this.contextLost ||
			document.documentElement.dataset.webgl === 'off'
		) {
			this.raf = 0;
			for (const quad of this.quads.values()) this.dropTexture(quad);
			return;
		}
		this.raf = requestAnimationFrame(this.tick);

		if (this.canvas.width !== Math.round(this.width * this.renderer.getPixelRatio())) {
			this.resize();
		}

		const time = now / 1000;
		const dt = Math.min(0.05, Math.max(0, (now - (this.last || now)) / 1000));
		this.last = now;
		const centre = this.height / 2;
		let drawable = 0;

		for (const quad of this.quads.values()) {
			const rect = quad.frame.getBoundingClientRect();
			const visible =
				rect.width > 0 && rect.bottom > -MARGIN && rect.top < this.height + MARGIN;

			if (!visible) {
				// Outside the window: give the pixels back rather than hold the whole set.
				if (quad.texture) this.dropTexture(quad);
				continue;
			}

			if (!quad.texture && !quad.loading && quad.image.complete && quad.image.naturalWidth > 0) {
				void this.uploadTexture(quad);
			}

			if (!quad.texture) continue;

			quad.mesh.scale.set(rect.width, rect.height, 1);
			quad.mesh.position.set(rect.left + rect.width / 2, rect.top + rect.height / 2, 0);
			quad.mesh.visible = true;

			const distance = Math.min(1, Math.abs(rect.top + rect.height / 2 - centre) / centre);
			quad.hover += (quad.hoverTarget - quad.hover) * (1 - Math.exp(-dt / 0.12));
			quad.material.uniforms.uTime.value = time;
			quad.material.uniforms.uProgress.value = EFFECTS ? distance * distance : 0;
			quad.material.uniforms.uHover.value = EFFECTS ? quad.hover : 0;
			// Grade is a mix factor in the shader, so switching it off leaves the
			// photograph untouched rather than half-graded.
			quad.material.uniforms.uGrade.value = GRADE ? 1 : 0;
			quad.material.uniforms.uGrain.value = GRAIN;

			// The DOM image stays visible until its quad is actually drawing pixels.
			quad.frame.dataset.gl = 'live';
			drawable++;
		}

		this.enforceTextureBudget();
		this.renderer.render(this.scene, this.camera);

		// The canvas stays hidden until a frame has actually been produced, so a layer
		// that never draws cannot cover the page it was meant to enhance.
		if (!this.canvas.dataset.ready) {
			this.canvas.dataset.ready = '';
		}

		const flag = drawable > 0 ? 'on' : 'off';
		if (document.documentElement.dataset.webgl !== flag) {
			document.documentElement.dataset.webgl = flag;
		}
	};

	/**
	 * Upload a plate's pixels. The bytes are fetched here rather than read off the DOM
	 * <img>, so that the gallery never depends on CORS to display a photograph: if this
	 * fails, the plate simply stays a plain image.
	 */
	private async uploadTexture(quad: Quad) {
		quad.loading = true;
		const url = quad.image.currentSrc || quad.image.src;

		try {
			const bitmap = await fetchBitmap(url);
			if (!bitmap || quad.mesh.parent === null) {
				bitmap?.close();
				return;
			}

			// Sampled raw and written straight through. An sRGB colour space here would have
			// the GPU decode the image to linear, and the photograph would come out visibly
			// darker than the same file drawn by the browser. Mipmaps and the default
			// filter chain stay on: that is what keeps a 1600px source crisp in a 793px quad.
			const texture = new Texture(bitmap);
			texture.needsUpdate = true;
			quad.texture = texture;
			quad.material.uniforms.uTexture.value = texture;
			quad.material.needsUpdate = true;
		} finally {
			quad.loading = false;
		}
	}

	private enforceTextureBudget() {
		const resident = [...this.quads.values()].filter((quad) => quad.texture);
		if (resident.length <= MAX_TEXTURES) return;

		const distance = (quad: Quad) => {
			const rect = quad.frame.getBoundingClientRect();
			return Math.abs(rect.top + rect.height / 2 - this.height / 2);
		};

		resident
			.sort((a, b) => distance(b) - distance(a))
			.slice(0, resident.length - MAX_TEXTURES)
			.forEach((quad) => this.dropTexture(quad));
	}
}

/**
 * Fetch a plate's bytes as an ImageBitmap. The second attempt bypasses the HTTP cache
 * on purpose: derivatives are served `immutable` for a year, so a browser that fetched
 * one before the bucket allowed this origin holds a copy that can never satisfy a CORS
 * request, and only a network refetch clears that.
 */
async function fetchBitmap(url: string): Promise<ImageBitmap | null> {
	for (const cache of ['default', 'reload'] as const) {
		try {
			const response = await fetch(url, { mode: 'cors', cache });
			if (!response.ok) continue;
			return await createImageBitmap(await response.blob());
		} catch {
			// Try the next cache mode, then give up and leave the plate to the DOM.
		}
	}
	return null;
}

function makeMaterial(noise: Texture): ShaderMaterial {
	return new ShaderMaterial({
		vertexShader: plateVertex,
		fragmentShader: plateFragment,
		uniforms: {
			uTexture: { value: null },
			uNoise: { value: noise },
			uTime: { value: 0 },
			uProgress: { value: 0 },
			uHover: { value: 0 },
			uGrade: { value: 0 },
			uGrain: { value: 0 }
		},
		depthTest: false,
		depthWrite: false
	});
}

/** Tiling value noise, generated here so the site ships no noise asset. */
function makeNoiseTexture(size = 256, lattice = 16): DataTexture {
	const data = new Uint8Array(size * size);
	const hash = (x: number, y: number) => {
		const value = Math.sin(x * 127.1 + y * 311.7) * 43758.5453;
		return value - Math.floor(value);
	};
	const wrap = (value: number) => ((value % lattice) + lattice) % lattice;

	for (let y = 0; y < size; y++) {
		for (let x = 0; x < size; x++) {
			const gx = (x / size) * lattice;
			const gy = (y / size) * lattice;
			const x0 = Math.floor(gx);
			const y0 = Math.floor(gy);
			const fx = gx - x0;
			const fy = gy - y0;
			const tx = fx * fx * (3 - 2 * fx);
			const ty = fy * fy * (3 - 2 * fy);

			const a = hash(wrap(x0), wrap(y0));
			const b = hash(wrap(x0 + 1), wrap(y0));
			const c = hash(wrap(x0), wrap(y0 + 1));
			const d = hash(wrap(x0 + 1), wrap(y0 + 1));
			const top = a + (b - a) * tx;
			const bottom = c + (d - c) * tx;
			data[y * size + x] = Math.round((top + (bottom - top) * ty) * 255);
		}
	}

	const texture = new DataTexture(data, size, size, RedFormat, UnsignedByteType);
	texture.wrapS = RepeatWrapping;
	texture.wrapT = RepeatWrapping;
	texture.minFilter = LinearFilter;
	texture.magFilter = LinearFilter;
	texture.needsUpdate = true;
	return texture;
}
