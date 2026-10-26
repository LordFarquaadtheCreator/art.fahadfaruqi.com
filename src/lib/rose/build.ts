// The rose's geometry and its mark vocabulary, ported from the standalone sketch of the same
// name. Everything below is arithmetic on typed arrays — no canvas, no DOM — so it runs in a
// worker and the page never waits for it.

export const ROSE_W = 900;
export const ROSE_H = 1040;
/** Camera units to canvas pixels. */
export const ROSE_SCALE = 380;

/** 0 cross, 1 dash, 2 dot, 3 bold cross. */
export type MarkKind = 0 | 1 | 2 | 3;

export interface RoseMark {
	t: MarkKind;
	x: number;
	y: number;
	/** Half the stroke's length. */
	n: number;
	/** Alpha bucket, 1..16 (16 is the bold register). */
	b: number;
	/** The petal the mark belongs to, -1 outside the bloom. */
	k: number;
	/** How far the mark swings when its petal opens. */
	w: number;
	ax: number;
	ay: number;
	cx: number;
	cy: number;
}

export interface Rose {
	markWidth: number;
	markHeight: number;
	/** How many petals the bloom has: the wind loop counts them. */
	petals: number;
	/** The empty share of the frame below the ink, as a fraction of its height. */
	trim: number;
	marks: RoseMark[];
}

function rng(seed: number) {
	let s = seed;
	return function () {
		s |= 0;
		s = (s + 0x6d2b79f5) | 0;
		let t = Math.imul(s ^ (s >>> 15), 1 | s);
		t = (t + Math.imul(t ^ (t >>> 7), 61 | t)) ^ t;
		return ((t ^ (t >>> 14)) >>> 0) / 4294967296;
	};
}

export function hash(i: number, j: number) {
	const s = Math.sin(i * 127.1 + j * 311.7) * 43758.5453;
	return s - Math.floor(s);
}

function lerp(p: number, q: number, t: number) {
	return p + (q - p) * t;
}

export function vnoise(x: number, y: number) {
	const i = Math.floor(x),
		j = Math.floor(y),
		fx = x - i,
		fy = y - j;
	const u = fx * fx * (3 - 2 * fx),
		v = fy * fy * (3 - 2 * fy);
	return lerp(lerp(hash(i, j), hash(i + 1, j), u), lerp(hash(i, j + 1), hash(i + 1, j + 1), u), v);
}

function clamp(x: number) {
	return Math.min(1, Math.max(0, x));
}

export function buildRose(): Rose {
	const W = ROSE_W,
		H = ROSE_H,
		CELL = 6,
		S = ROSE_SCALE,
		BX = 470,
		BY = 700,
		STEM_LENGTH = 0.45; // share of the drop from the bloom's base to the frame edge that the stem keeps
	const AL = 0.8,
		CA = Math.cos(AL),
		SA = Math.sin(AL),
		GM = 0.16,
		CG = Math.cos(GM),
		SG = Math.sin(GM);

	// ---- 1. geometry: ~30 curled petal shells wrapped around a vertical axis, golden-angle spaced
	const R = rng(7),
		N = 44,
		petals: {
			phi: number;
			w: number;
			r0: number;
			b0: number;
			b1: number;
			L: number;
			y0: number;
			tw: number;
			ph: number;
		}[] = [];
	for (let i = 0; i < N; i++) {
		const t = i / (N - 1),
			r0 = 0.02 + 0.08 * t,
			tt = Math.pow(t, 0.9);
		petals.push({
			phi: i * 2.39996 + R() * 0.2,
			w: 1.25 - 0.2 * t + R() * 0.15,
			r0,
			b0: 0.12 + 0.6 * tt + R() * 0.06,
			b1: -0.1 + 0.3 * tt + R() * 0.08, // shared base; tall, mostly vertical petals; outer ones flare a little and end rising
			L: 1.5 - 0.35 * t + R() * 0.12,
			y0: 0.02 * t,
			tw: R() - 0.5,
			ph: R() * 6
		});
	}
	type Petal = (typeof petals)[number];
	const P = [0, 0, 0],
		Pu = [0, 0, 0],
		Pv = [0, 0, 0];
	// camera space: x right, y up, z toward viewer
	function pos(p: Petal, u: number, v: number, o: number[]) {
		const vv = Math.max(v, 0),
			w = p.w * (0.5 + 0.5 * Math.sqrt(vv)),
			phi = p.phi + u * w + 0.25 * p.tw * v;
		let k = p.b1 - p.b0;
		if (Math.abs(k) < 1e-3) k = 1e-3;
		const bb = p.b0 + k * vv; // angle from vertical, grows along the petal
		const rho =
			p.r0 +
			(p.L * (Math.cos(p.b0) - Math.cos(bb))) / k +
			0.02 * v * Math.sin(u * 4 + p.ph) +
			0.012 * vv * Math.sin(u * 13 + p.ph * 2);
		const y = p.y0 + (p.L * (Math.sin(bb) - Math.sin(p.b0))) / k - 0.08 * u * u * vv;
		const sx = rho * Math.sin(phi),
			sz = rho * Math.cos(phi);
		const y2 = y * CA - sz * SA,
			z2 = sz * CA + y * SA; // tilt: we look slightly down into the bloom
		o[0] = sx * CG + y2 * SG;
		o[1] = -sx * SG + y2 * CG;
		o[2] = z2; // lean
	}
	const NU = 280,
		NV = 300;

	function gen(withNormal: boolean, cb: (p: number[], n: number[] | null, k: number, v: number) => void) {
		for (let k = 0; k < N; k++) {
			const p = petals[k];
			for (let iu = 0; iu <= NU; iu++) {
				const u = -1 + (2 * iu) / NU,
					vmax = 1 - 0.3 * u * u + 0.04 * Math.sin(u * 3 + p.ph);
				for (let iv = 0; iv <= NV; iv++) {
					const v = (iv / NV) * vmax;
					pos(p, u, v, P);
					if (!withNormal) {
						cb(P, null, k, v);
						continue;
					}
					pos(p, u + 0.01, v, Pu);
					pos(p, u, v + 0.01, Pv);
					const ax = Pu[0] - P[0],
						ay = Pu[1] - P[1],
						az = Pu[2] - P[2],
						bx = Pv[0] - P[0],
						by = Pv[1] - P[1],
						bz = Pv[2] - P[2];
					const nx = ay * bz - az * by,
						ny = az * bx - ax * bz,
						nz = ax * by - ay * bx;
					let l = Math.hypot(nx, ny, nz) || 1;
					if (nz < 0) l = -l;
					cb(P, [nx / l, ny / l, nz / l], k, v);
				}
			}
		}
	}

	// ---- 2. light + shadow map (light from upper left, in front)
	let Lx = -0.38,
		Ly = 0.3,
		Lz = 0.87;
	const ll = Math.hypot(Lx, Ly, Lz);
	Lx /= ll;
	Ly /= ll;
	Lz /= ll;
	let E1x = Lz,
		E1z = -Lx;
	const el = Math.hypot(E1x, E1z);
	E1x /= el;
	E1z /= el;
	const E2x = Ly * E1z,
		E2y = Lz * E1x - Lx * E1z,
		E2z = -Ly * E1x;
	const SMW = 1600,
		SMC = 800,
		SM = new Float32Array(SMW * SMW).fill(-1e9);
	gen(false, (p) => {
		const sx = Math.round(SMC + (p[0] * E1x + p[2] * E1z) * S - 0.5),
			sy = Math.round(SMC - (p[0] * E2x + p[1] * E2y + p[2] * E2z) * S - 0.5);
		const d = p[0] * Lx + p[1] * Ly + p[2] * Lz;
		for (let dy = 0; dy < 2; dy++)
			for (let dx = 0; dx < 2; dx++) {
				const j = (sy + dy) * SMW + sx + dx;
				if (SM[j] < d) SM[j] = d;
			}
	});
	const TAPS = [0, 0, 5, 0, -5, 0, 0, 5, 0, -5, 11, 0, -11, 0, 0, 11, 0, -11, 8, 8, -8, 8, 8, -8, -8, -8]; // 13 taps in three rings: a wide, soft penumbra
	function litFrac(x: number, y: number, z: number) {
		const sx = Math.round(SMC + (x * E1x + z * E1z) * S),
			sy = Math.round(SMC - (x * E2x + y * E2y + z * E2z) * S);
		const d = x * Lx + y * Ly + z * Lz + 0.03;
		let n = 0;
		for (let q = 0; q < 26; q += 2) if (SM[(sy + TAPS[q + 1]) * SMW + sx + TAPS[q]] <= d) n++;
		return n / 13;
	}

	// ---- 3. z-buffer from the camera
	const ZB = new Float32Array(W * H).fill(-1e9),
		ID = new Int16Array(W * H).fill(-1);
	const PX = new Float32Array(W * H),
		PY = new Float32Array(W * H),
		NX = new Float32Array(W * H),
		NY = new Float32Array(W * H),
		NZ = new Float32Array(W * H),
		PV = new Float32Array(W * H);
	gen(true, (p, n, k, v) => {
		const ix = Math.round(BX + p[0] * S - 0.5),
			iy = Math.round(BY - p[1] * S - 0.5);
		if (ix < 0 || iy < 0 || ix >= W - 1 || iy >= H - 1) return;
		for (let dy = 0; dy < 2; dy++)
			for (let dx = 0; dx < 2; dx++) {
				const j = (iy + dy) * W + ix + dx;
				if (ZB[j] < p[2]) {
					ZB[j] = p[2];
					ID[j] = k;
					PX[j] = p[0];
					PY[j] = p[1];
					PV[j] = v;
					NX[j] = n![0];
					NY[j] = n![1];
					NZ[j] = n![2];
				}
			}
	});
	function at(x: number, y: number) {
		x = Math.round(x);
		y = Math.round(y);
		return x < 0 || y < 0 || x >= W || y >= H ? -1 : y * W + x;
	}

	// the lowest covered pixel near the middle is where the stem begins
	let stemY = 0,
		stemX0 = BX;
	for (let yy = H - 1; yy > 0 && !stemY; yy--)
		for (let xx = BX - 80; xx < BX + 80; xx++)
			if (ID[yy * W + xx] >= 0) {
				stemY = yy;
				stemX0 = xx;
				break;
			}

	// ---- 4. mark vocabulary
	// marks are recorded, not drawn: each remembers its petal and how high up the plant it sits, so it can move later
	const MK: RoseMark[] = [];
	function rec(t: MarkKind, x: number, y: number, n: number, al: number) {
		const b = t === 3 ? 16 : Math.round(clamp(al) * 15);
		if (b < 1) return;
		const j = at(x, y),
			kk = j >= 0 ? ID[j] : -1,
			m: RoseMark = { t, x, y, n, b, k: kk, w: 0, ax: 0, ay: 0, cx: 0, cy: 0 };
		if (kk >= 0) {
			// hinge: how far this mark travels when its petal rotates outward about the shared base
			const px3 = PX[j],
				py3 = PY[j],
				pz3 = ZB[j],
				ux = CA * SG,
				uy = CA * CG,
				uz = SA,
				hh = px3 * ux + py3 * uy + pz3 * uz;
			let rx = px3 - hh * ux,
				ry = py3 - hh * uy,
				rz = pz3 - hh * uz;
			const rl = Math.hypot(rx, ry, rz);
			if (rl > 1e-3) {
				rx /= rl;
				ry /= rl;
				rz /= rl;
				const nx = uy * rz - uz * ry,
					ny = uz * rx - ux * rz,
					nz = ux * ry - uy * rx,
					nd = nx * px3 + ny * py3 + nz * pz3;
				m.ax = px3 - nx * nd;
				m.ay = py3 - ny * nd;
				m.cx = ny * pz3 - nz * py3;
				m.cy = nz * px3 - nx * pz3;
				m.w = (0.2 + (0.8 * kk) / (N - 1)) * Math.pow(Math.min(1, PV[j]), 1.2);
			} // outer petals and petal tips swing furthest
		}
		MK.push(m);
	}
	const plus = (x: number, y: number, n: number, al: number) => rec(0, x, y, n, al);
	const dash = (x: number, y: number, len: number, al: number) => rec(1, x, y, len / 2, al);
	const dot = (x: number, y: number, al: number) => rec(2, x, y, 0.7, al);
	const bold = (x: number, y: number, n: number) => rec(3, x, y, n, 1);

	// ---- 5. rules, one mark per cell, first match wins
	const stemEnd = stemY + STEM_LENGTH * (H - stemY);
	const D4 = [
		[1, 0],
		[-1, 0],
		[0, 1],
		[0, -1]
	];
	const hi: number[][] = [];
	for (let row = 0; row * CELL < H; row++)
		for (let col = 0; col * CELL < W; col++) {
			const px = col * CELL + CELL / 2 + (row % 2) * (CELL / 2),
				py = row * CELL + CELL / 2,
				h = hash(col, row),
				i0 = at(px, py);
			if (i0 < 0) continue;
			if (ID[i0] < 0) {
				// outside the bloom: stem = column of crosses
				if (py > stemY - 10 && py < stemEnd) {
					const sxn = stemX0 - 0.12 * (py - stemY) + 9 * Math.sin((py - stemY) / 60);
					if (Math.abs(px - sxn) < CELL / 2 + 0.01) plus(px, py, 2.6, 0.7 + 0.3 * Math.sin(py * 0.07));
					else if (px < sxn && px > sxn - CELL * 1.6 && row % 3 === 0) dash(px, py, CELL * 1.2, 0.35);
				}
				continue;
			}
			let sil = false,
				lip = false,
				occl = false;
			const zi = ZB[i0];
			for (let d = 0; d < 4; d++) {
				const j = at(px + D4[d][0] * CELL * 0.9, py + D4[d][1] * CELL * 0.9);
				if (j < 0 || ID[j] < 0) sil = true;
				else if (ID[j] !== ID[i0]) {
					const dz = zi - ZB[j];
					if (dz > 0.035) lip = true;
					else if (dz < -0.035) occl = true;
				}
			}
			const nl = NX[i0] * Lx + NY[i0] * Ly + NZ[i0] * Lz,
				lam =
					Math.pow(clamp((nl + 0.45) / 1.45), 1.2) * (0.9 + 0.2 * vnoise(px * 0.05, py * 0.05)), // wrap lighting: the light-to-dark transition is gradual
				fill = Math.max(0, NX[i0] * 0.5 + NY[i0] * 0.2 + NZ[i0] * 0.84); // weak fill light from the right
			const lit = litFrac(PX[i0], PY[i0], zi),
				rim = Math.pow(1 - NZ[i0], 3) * 0.2;
			let L = clamp(0.08 + 0.46 * lam * (0.4 + 0.6 * lit) + 0.1 * fill + rim);
			if (occl) L *= 0.6; // contact shadow under a petal in front
			if (sil) {
				plus(px, py, 2.6, 0.5 + 0.5 * Math.max(L, 0.35)); // silhouette: crosses
				continue;
			}
			if (lip) {
				plus(px, py, 2.8, 0.55 + 0.45 * L); // petal lip: crosses
				if (L > 0.55) hi.push([px, py]);
				continue;
			}
			if (L > 0.6) {
				hi.push([px, py]);
				dash(px, py, CELL * 3.4, 1); // highlight
				continue;
			}
			if (L > 0.55) {
				dash(px, py, CELL * (1.2 + 2.2 * L), 0.55 + 0.45 * L); // lit: long dashes
				continue;
			}
			if (L > 0.42) {
				dash(px, py, CELL * (0.9 + 1.4 * L), 0.3 + 0.7 * L); // mid: medium dashes
				continue;
			}
			if (L > 0.26) {
				dash(px, py, CELL * 0.8, 0.25 + 0.8 * L); // low-mid: short dashes with a stray dot
				if (h > 0.6) dot(px + CELL * 0.4, py + CELL * 0.45, 0.4);
				continue;
			}
			if (L > 0.1 && h < L * 3.2) dot(px, py, 0.2 + L); // shadow: sparse dots
		}
	for (let q = 0; q < hi.length; q++) if (hash(q, 3) > 0.4) bold(hi[q][0], hi[q][1], 4.4);

	// how much of the frame's height is empty below the ink: the runner pulls the canvas down by
	// this share so the stem's end, not the frame edge, sits on the corner
	let inkBottom = 0;
	for (const m of MK) inkBottom = Math.max(inkBottom, m.y + m.n);

	return { markWidth: W, markHeight: H, petals: N, trim: (H - inkBottom) / H, marks: MK };
}