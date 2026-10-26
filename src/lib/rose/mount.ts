import { ROSE_H, ROSE_SCALE, ROSE_W, hash, type Rose, vnoise } from './build';
import RoseWorker from './rose.worker?worker&inline';

const DPR_CAP = 2;
// Each glow layer is the art canvas downscaled, then stretched back up under the crisp marks:
// three cheap blurs instead of a filter.
const GLOW_STEPS = [2, 4, 8];

/**
 * Runs the rose on a canvas in the backdrop. The geometry is built in a worker, so mounting
 * costs nothing on the main thread, and until it lands there is simply nothing to draw.
 */
export function mountRose(canvas: HTMLCanvasElement): () => void {
	const got = canvas.getContext('2d');
	if (!got) return () => {};
	const ctx = got;

	let disposed = false;
	let raf = 0;
	let rose: Rose | null = null;
	let scaleX = 1;
	let scaleY = 1;

	const art = document.createElement('canvas');
	const artCtx = art.getContext('2d')! as CanvasRenderingContext2D;
	let glow: { canvas: HTMLCanvasElement; ctx: CanvasRenderingContext2D }[] = [];

	const quiet = matchMedia('(prefers-reduced-motion: reduce)').matches;

	function layout(): boolean {
		const cssW = canvas.clientWidth;
		const cssH = canvas.clientHeight;
		if (cssW < 1 || cssH < 1) return false;
		const dpr = Math.min(window.devicePixelRatio || 1, DPR_CAP);
		const w = Math.round(cssW * dpr);
		const h = Math.round(cssH * dpr);
		if (canvas.width !== w || canvas.height !== h) {
			canvas.width = w;
			canvas.height = h;
		}
		if (art.width !== w || art.height !== h) {
			art.width = w;
			art.height = h;
		}
		scaleX = w / ROSE_W;
		scaleY = h / ROSE_H;
		glow = GLOW_STEPS.map((step) => {
			const c = document.createElement('canvas');
			c.width = Math.max(1, Math.ceil(w / step));
			c.height = Math.max(1, Math.ceil(h / step));
			return { canvas: c, ctx: c.getContext('2d')! };
		});
		return true;
	}

	function draw(ms: number) {
		const data = rose;
		if (!data || !layout()) return;

		// opening cycle, 8s: petals swing out from the shared base and come back to exactly where
		// they started; outer petals lead, inner ones follow
		const t = ms / 1000;
		const AMAX = 0.11;
		const OK: number[] = [];
		for (let q = 0; q < data.petals; q++) {
			const r = hash(q, 23),
				T = 9, // one shared period: the wave, not the petal, sets the timing
				ph = ((((t - 2.4 * (1 - q / (data.petals - 1)) - 0.1 * r) % T) + T) % T) / T,
				wv = 0.5 - 0.5 * Math.cos(2 * Math.PI * ph),
				e = wv * wv * (3 - 2 * wv);
			// smooth noise over layer and time: neighbours move alike, the bloom as a whole still
			// looks uneven
			OK.push(
				e *
					(0.75 + 0.5 * vnoise(q * 0.03 + 5, t * 0.25)) *
					(1 + 0.12 * Math.sin(t * 1.1 + q * 0.15))
			);
		}

		const paths: Path2D[] = [];
		for (let b = 0; b <= 16; b++) paths.push(new Path2D());
		for (const m of data.marks) {
			// the whole plant bends from the soil up; each petal flutters on its own beat
			const th = OK[m.k >= 0 ? m.k : 0] * m.w * AMAX,
				c1 = Math.cos(th) - 1,
				sn = Math.sin(th);
			const x = m.x + (m.ax * c1 + m.cx * sn) * ROSE_SCALE;
			const y = m.y - (m.ay * c1 + m.cy * sn) * ROSE_SCALE;
			const path = paths[m.b];
			if (m.t === 0 || m.t === 3) {
				path.moveTo(x - m.n, y);
				path.lineTo(x + m.n, y);
				path.moveTo(x, y - m.n);
				path.lineTo(x, y + m.n);
			} else {
				path.moveTo(x - m.n, y);
				path.lineTo(x + m.n, y);
			}
		}

		artCtx.setTransform(scaleX, 0, 0, scaleY, 0, 0);
		artCtx.clearRect(0, 0, ROSE_W, ROSE_H);
		artCtx.lineWidth = 1.3;
		for (let b = 1; b <= 15; b++) {
			artCtx.strokeStyle = `rgba(255,255,255,${(b / 15).toFixed(3)})`;
			artCtx.stroke(paths[b]);
		}
		artCtx.lineWidth = 1.9;
		artCtx.strokeStyle = '#fff';
		artCtx.stroke(paths[16]);

		for (let i = 0; i < glow.length; i++) {
			const source = i === 0 ? art : glow[i - 1].canvas;
			const { canvas: c, ctx: g } = glow[i];
			g.clearRect(0, 0, c.width, c.height);
			g.drawImage(source, 0, 0, c.width, c.height);
		}

		const w = canvas.width;
		const h = canvas.height;
		ctx.setTransform(1, 0, 0, 1, 0, 0);
		ctx.globalCompositeOperation = 'source-over';
		ctx.globalAlpha = 1;
		ctx.clearRect(0, 0, w, h);
		// the canvas is transparent, not black: the marks are added onto whatever the backdrop has
		ctx.globalCompositeOperation = 'lighter';
		ctx.imageSmoothingQuality = 'high';
		ctx.drawImage(glow[2].canvas, 0, 0, w, h);
		ctx.globalAlpha = 0.9;
		ctx.drawImage(glow[1].canvas, 0, 0, w, h);
		ctx.globalAlpha = 1;
		ctx.drawImage(art, 0, 0, w, h);
	}

	function tick(ms: number) {
		if (disposed) return;
		draw(ms);
		raf = requestAnimationFrame(tick);
	}

	const observer = new ResizeObserver(() => {
		if (!rose || quiet) draw(0);
	});
	observer.observe(canvas);

	const worker = new RoseWorker();
	worker.onmessage = (event: MessageEvent<Rose>) => {
		if (disposed) return;
		rose = event.data;
		// the drawing's own ink ends above the frame: pull the canvas down so it meets the corner
		canvas.style.setProperty('--rose-trim', String(rose.trim));
		worker.terminate();
		if (quiet) draw(0);
		else raf = requestAnimationFrame(tick);
	};
	worker.postMessage('build');

	return () => {
		disposed = true;
		cancelAnimationFrame(raf);
		observer.disconnect();
		worker.terminate();
	};
}
