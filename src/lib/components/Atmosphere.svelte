<script lang="ts">
	import { browser } from '$app/environment';
	import { cursor } from '$lib/utils/cursor.svelte';

	// Step-animated so the grain crawls and flickers like film
	const coarseNoise =
		"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='180' height='180'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.8' numOctaves='4' stitchTiles='stitch'/%3E%3CfeComponentTransfer%3E%3CfeFuncA type='linear' slope='1.4' intercept='-0.2'/%3E%3C/feComponentTransfer%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E";
	const fineNoise =
		"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='120' height='120'%3E%3Cfilter id='f'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='1.5' numOctaves='2' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23f)'/%3E%3C/svg%3E";

	let blobBase = $state<HTMLElement | null>(null);
	let blobStretch = $state<HTMLElement | null>(null);
	let blobBaseCore = $state<HTMLElement | null>(null);
	let carrier = $state<HTMLElement | null>(null);

	// Mouse tracker onhover logic
	$effect(() => {
		if (!browser || !carrier) return;
		if (window.matchMedia('(hover: none)').matches) return;

		const node = carrier;

		function onMove(event: PointerEvent) {
			node.style.transform = `translate3d(${event.clientX}px, ${event.clientY}px, 0)`;
		}

		window.addEventListener('pointermove', onMove, { passive: true });
		return () => window.removeEventListener('pointermove', onMove);
	});

	// Light movement logic
	$effect(() => {
		if (!browser || !blobBase || !blobStretch || !blobBaseCore) return;

		const nBlobBase = blobBase;
		const nBlobStretch = blobStretch;
		const nBlobBaseCore = blobBaseCore;
		const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
		const coarsePointer = window.matchMedia('(hover: none)').matches;

		let radius = 0;
		let frameSize = 1;
		function measure() {
			radius = (nBlobBaseCore.offsetWidth + nBlobBaseCore.offsetHeight) / 4;
			frameSize = nBlobStretch.offsetWidth;
		}
		measure();
		window.addEventListener('resize', measure);

		if (reduced || coarsePointer) {
		    // TODO: redirect to a static background
			const parked = `translate3d(${window.innerWidth * 0.78}px, ${window.innerHeight * 0.82}px, 0)`;
			nBlobBase.style.transform = parked;
			nBlobStretch.style.transform = `${parked} scale(${(2 * radius) / frameSize})`;
			nBlobStretch.style.setProperty('--blob-fade', '0');
			nBlobBase.style.setProperty('--blob-fade', '1');
			return () => window.removeEventListener('resize', measure);
		}

		let targetX = window.innerWidth * 0.78;
		let targetY = window.innerHeight * 0.82;
		let x = targetX;
		let y = targetY;
		let sampledX = targetX;
		let sampledY = targetY;
		let lagX = targetX;
		let lagY = targetY;
		let reaction = 0;
		let frame = 0;
		let heading = 0; // the held direction from the pointer to the base, in radians

		const follow = 0.022;
		const lagFollow = 0.1; // the stretch's tip follows on a slower pass — its delay
		const attack = 0.2; // how fast the shape answers the pointer
		const release = 0.045; // and how long it stays disturbed once the pointer stops
		const speedFull = 34; // px the pointer covers in one frame at full reaction

		function onPointerMove(event: PointerEvent) {
			targetX = event.clientX;
			targetY = event.clientY;
		}

		function tick(time: number) {
			x += (targetX - x) * follow;
			y += (targetY - y) * follow;
			lagX += (targetX - lagX) * lagFollow;
			lagY += (targetY - lagY) * lagFollow;

			// Pointer velocity
			const travelX = targetX - sampledX;
			const travelY = targetY - sampledY;
			sampledX = targetX;
			sampledY = targetY;

			const wanted = Math.min(Math.hypot(travelX, travelY) / speedFull, 1);
			reaction += (wanted - reaction) * (wanted > reaction ? attack : release);

			const swayX = Math.sin(time * 0.00021) * 54 + Math.sin(time * 0.00057) * 20;
			const swayY = Math.cos(time * 0.00029) * 42 + Math.sin(time * 0.00043) * 16;

			const ox = x + swayX;
			const oy = y + swayY;
			nBlobBase.style.transform = `translate3d(${ox}px, ${oy}px, 0)`;

			// The stretch oval. With L the delayed tip, O the base's centre and R the radius, the
			// point of the circle farthest from L is O + R·û, û = (O − L)/|O − L| — the maximiser
			// of |X − L| over |X − O| = R. The oval's major axis spans L to that point; its width
			// is the light's own diameter, so it rests on the light and grows towards the pointer.
			const dx = ox - lagX;
			const dy = oy - lagY;
			const gap = Math.hypot(dx, dy);
			if (gap > 1) heading = Math.atan2(dy, dx); // hold the last heading when they coincide
			const midX = (lagX + ox + Math.cos(heading) * radius) / 2;
			const midY = (lagY + oy + Math.sin(heading) * radius) / 2;
			const span = gap + radius; // |L → P|
			nBlobStretch.style.transform =
				`translate3d(${midX}px, ${midY}px, 0) rotate(${(heading * 180) / Math.PI}deg) scale(${span / frameSize}, ${(2 * radius) / frameSize})`;

			// Opacity is the distance, and the base takes the inverse share, so the two blobs sum to 1.
			const fade = Math.min(Math.max((gap - radius) / radius, 0), 1);
			nBlobStretch.style.setProperty('--blob-fade', String(fade));
			nBlobBase.style.setProperty('--blob-fade', String(1 - fade));

			frame = requestAnimationFrame(tick);
		}

		window.addEventListener('pointermove', onPointerMove, { passive: true });
		frame = requestAnimationFrame(tick);

		return () => {
			window.removeEventListener('pointermove', onPointerMove);
			window.removeEventListener('resize', measure);
			cancelAnimationFrame(frame);
		};
	});
</script>

<div class="backdrop" aria-hidden="true">
	<div class="blob-base" bind:this={blobBase}>
		<div class="blob-base__core" bind:this={blobBaseCore}></div>
	</div>
	<div class="blob-stretch" bind:this={blobStretch}>
		<div class="blob-stretch__core"></div>
	</div>
	<div class="vignette vignette--dark"></div>
	<div class="vignette vignette--light"></div>
	<div class="mesh"></div>
</div>

<div class="overlay" aria-hidden="true">
	<div class="grain grain--coarse" style="background-image: url('{coarseNoise}')"></div>
	<div class="grain grain--fine" style="background-image: url('{fineNoise}')"></div>

	<div class="carrier" bind:this={carrier}>
		<span class="carrier__label label num" class:carrier__label--active={cursor.active}
			>{cursor.label}</span
		>
	</div>
</div>

<style>
	.backdrop {
		position: fixed;
		inset: 0;
		z-index: 0;
		background: var(--bg);
		pointer-events: none;
		overflow: hidden;
	}

	.overlay {
		position: fixed;
		inset: 0;
		z-index: 6;
		pointer-events: none;
		overflow: hidden;
	}

	.carrier {
		position: absolute;
		top: 0;
		left: 0;
		will-change: transform;
	}

	.carrier__label {
		display: block;
		margin: -1.6rem 0 0 1rem;
		color: var(--accent);
		opacity: 0;
		transform: scale(0.86);
		transition:
			opacity 0.3s ease,
			transform 0.4s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	.carrier__label--active {
		opacity: 1;
		transform: scale(1);
	}

	/* If no hover, disable carrier */
	@media (hover: none) {
		.carrier {
			display: none;
		}
	}

	.blob-base,
	.blob-stretch {
		position: absolute;
		top: 0;
		left: 0;
		width: 78vmax;
		height: 78vmax;
		margin: -39vmax 0 0 -39vmax;
		opacity: var(--glow-opacity);
		transition: opacity 0.6s ease;
		mix-blend-mode: var(--glow-blend);
		will-change: transform;
	}

	/* The base light: an organic blob, centred in its frame. Fill: --blob-base. */
	.blob-base__core {
		position: absolute;
		inset: 0;
		margin: auto;
		width: 43.5%;
		height: 39%;
		filter: blur(44px);
		border-radius: 47% 53% 41% 59% / 55% 44% 56% 45%;
		background: radial-gradient(
			closest-side,
			color-mix(in srgb, var(--blob-base) 28%, transparent),
			color-mix(in srgb, var(--blob-base) 20%, transparent) 60%,
			color-mix(in srgb, var(--blob-base) 10.5%, transparent) 100%
		);
		opacity: var(--blob-fade, 1);
	}

	/* The stretch: an ellipse filling its frame, which the loop scales to the computed span, so
	   the fill and the blur stretch with it. Fill: --blob-stretch. */
	.blob-stretch__core {
		position: absolute;
		inset: 0;
		filter: blur(44px);
		border-radius: 50%;
		background: radial-gradient(
			closest-side,
			color-mix(in srgb, var(--blob-stretch) 28%, transparent),
			color-mix(in srgb, var(--blob-stretch) 20%, transparent) 60%,
			color-mix(in srgb, var(--blob-stretch) 10.5%, transparent) 100%
		);
		opacity: var(--blob-fade, 1);
	}

	/* The lattice the light reveals. One fixed layer for the whole page — never scaled or moved
	   with a blob, so the lines stay a property of the screen. It sits above the blobs and
	   inverts what they do to the ground — multiply where the light screens, screen where it
	   multiplies — so the lines cut into the glow and barely register on the empty ground. */
	.mesh {
		position: absolute;
		inset: 0;
		--mesh-cell: 28px;
		background-image:
			linear-gradient(to right, var(--mesh-ink) 0 1px, transparent 1px),
			linear-gradient(to bottom, var(--mesh-ink) 0 1px, transparent 1px);
		background-size: var(--mesh-cell) var(--mesh-cell);
		mix-blend-mode: var(--mesh-blend);
	}

	.grain {
		position: absolute;
		inset: -40%;
		background-repeat: repeat;
		will-change: transform, opacity;
	}

	.grain--coarse {
		background-size: 180px 180px;
		opacity: var(--grain-opacity);
		mix-blend-mode: var(--grain-blend);
		animation:
			grain-crawl 4.2s steps(7) infinite,
			grain-flicker 0.6s steps(3) infinite alternate;
	}

	.grain--fine {
		background-size: 120px 120px;
		opacity: calc(var(--grain-opacity) * 0.62);
		mix-blend-mode: soft-light;
		animation:
			grain-crawl-fine 2.6s steps(9) infinite reverse,
			grain-flicker-fine 0.34s steps(2) infinite alternate;
	}

	.vignette {
		position: absolute;
		inset: 0;
		mix-blend-mode: multiply;
		transition: opacity 0.7s ease;
	}

	.vignette--dark {
		background: radial-gradient(125% 95% at 50% 42%, transparent 42%, rgba(0, 0, 0, 0.55) 100%);
		opacity: 1;
	}

	.vignette--light {
		background: radial-gradient(125% 95% at 50% 42%, transparent 42%, rgba(60, 52, 44, 0.18) 100%);
		opacity: 0;
	}

	:global(html[data-theme='light']) .vignette--dark {
		opacity: 0;
	}

	:global(html[data-theme='light']) .vignette--light {
		opacity: 1;
	}

	@keyframes grain-crawl {
		0% {
			transform: translate3d(0, 0, 0);
		}
		20% {
			transform: translate3d(-3%, 2%, 0);
		}
		40% {
			transform: translate3d(2%, -3%, 0);
		}
		60% {
			transform: translate3d(-2%, -2%, 0);
		}
		80% {
			transform: translate3d(3%, 1%, 0);
		}
		100% {
			transform: translate3d(0, 0, 0);
		}
	}

	@keyframes grain-crawl-fine {
		0% {
			transform: translate3d(0, 0, 0);
		}
		25% {
			transform: translate3d(2%, -1%, 0);
		}
		50% {
			transform: translate3d(-1%, 2%, 0);
		}
		75% {
			transform: translate3d(1%, 1%, 0);
		}
		100% {
			transform: translate3d(0, 0, 0);
		}
	}

	@keyframes grain-flicker {
		from {
			opacity: calc(var(--grain-opacity) * 0.72);
		}
		to {
			opacity: calc(var(--grain-opacity) * 1.28);
		}
	}

	@keyframes grain-flicker-fine {
		from {
			opacity: calc(var(--grain-opacity) * 0.45);
		}
		to {
			opacity: calc(var(--grain-opacity) * 0.95);
		}
	}
</style>
