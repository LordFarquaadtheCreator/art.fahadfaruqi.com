<script lang="ts">
	import { browser } from '$app/environment';
	import { cursor } from '$lib/utils/cursor.svelte';

	// Two noise plates at different scales, plus a vignette. The plates are step-animated
	// on a transform and an opacity so the grain crawls and flickers like film rather than
	// sitting there as a static texture.
	const coarseNoise =
		"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='180' height='180'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.8' numOctaves='4' stitchTiles='stitch'/%3E%3CfeComponentTransfer%3E%3CfeFuncA type='linear' slope='1.4' intercept='-0.2'/%3E%3C/feComponentTransfer%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E";
	const fineNoise =
		"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='120' height='120'%3E%3Cfilter id='f'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='1.5' numOctaves='2' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23f)'/%3E%3C/svg%3E";

	let glow = $state<HTMLElement | null>(null);
	let carrier = $state<HTMLElement | null>(null);

	// The number the pointer is carrying over the grid. No easing here: a cursor label
	// that trails reads as lag, and the show/hide motion is CSS.
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

	// The light leak follows the pointer, but only slowly: the rendered position eases
	// toward the cursor at a low factor, wobbles on two slow sine terms, and stretches
	// along its own direction of travel — so the light trails and sways instead of
	// snapping to the cursor.
	$effect(() => {
		if (!browser || !glow) return;

		const node = glow;
		const reduced = window.matchMedia('(prefers-reduced-motion: reduce)').matches;
		const coarsePointer = window.matchMedia('(hover: none)').matches;

		if (reduced || coarsePointer) {
			node.style.transform = `translate3d(${window.innerWidth * 0.78}px, ${window.innerHeight * 0.82}px, 0)`;
			return;
		}

		let targetX = window.innerWidth * 0.78;
		let targetY = window.innerHeight * 0.82;
		let x = targetX;
		let y = targetY;
		let previousX = x;
		let previousY = y;
		let frame = 0;

		const follow = 0.022;

		function onPointerMove(event: PointerEvent) {
			targetX = event.clientX;
			targetY = event.clientY;
		}

		function tick(time: number) {
			x += (targetX - x) * follow;
			y += (targetY - y) * follow;

			const swayX = Math.sin(time * 0.00021) * 54 + Math.sin(time * 0.00057) * 20;
			const swayY = Math.cos(time * 0.00029) * 42 + Math.sin(time * 0.00043) * 16;

			const velocityX = x - previousX;
			const velocityY = y - previousY;
			const speed = Math.min(Math.hypot(velocityX, velocityY), 40);

			// Travel direction, mapped into a rotation, and a stretch along that axis.
			const angle = (Math.atan2(velocityY, velocityX) * 180) / Math.PI;
			const stretch = 1 + speed * 0.012;

			node.style.transform = `translate3d(${x + swayX}px, ${y + swayY}px, 0) rotate(${angle}deg) scale(${stretch}, ${2 - stretch})`;

			previousX = x;
			previousY = y;
			frame = requestAnimationFrame(tick);
		}

		window.addEventListener('pointermove', onPointerMove, { passive: true });
		frame = requestAnimationFrame(tick);

		return () => {
			window.removeEventListener('pointermove', onPointerMove);
			cancelAnimationFrame(frame);
		};
	});
</script>

<div class="backdrop" aria-hidden="true">
	<div class="glow" bind:this={glow}></div>
	<div class="vignette vignette--dark"></div>
	<div class="vignette vignette--light"></div>
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
	/* Behind the content: the page's own light source. */
	.backdrop {
		position: fixed;
		inset: 0;
		z-index: 0;
		pointer-events: none;
		overflow: hidden;
	}

	/* Over the content: film grain, the only layer that sits on top of the photographs. */
	.overlay {
		position: fixed;
		inset: 0;
		z-index: 6;
		pointer-events: none;
		overflow: hidden;
	}

	/* ------------------------------------------------------------ cursor index */

	/* Sits above the photographs, carrying the plate's number with the pointer. */
	.carrier {
		position: absolute;
		top: 0;
		left: 0;
		will-change: transform;
	}

	/* The number the pointer is carrying is the one live readout on the grid. */
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

	/* A finger has no hover, and a label pinned under the finger is just in the way. */
	@media (hover: none) {
		.carrier {
			display: none;
		}
	}

	/* ------------------------------------------------------------ light leak */

	/* Hue 36 and 31, against the accent's 35.8: the light and the signal are one warm
	   family. The stops used to sit at 22 and 13 — a redder orange that read as a second
	   colour the moment the accent existed. */
	.glow {
		position: absolute;
		top: 0;
		left: 0;
		width: 78vmax;
		height: 78vmax;
		margin: -39vmax 0 0 -39vmax;
		background: radial-gradient(
			closest-side,
			rgba(255, 168, 38, 0.3),
			rgba(255, 144, 25, 0.12) 42%,
			transparent 70%
		);
		filter: blur(60px);
		opacity: var(--glow-opacity);
		/* So the light eases when the theme changes instead of snapping. */
		transition: opacity 0.6s ease;
		mix-blend-mode: var(--glow-blend);
		will-change: transform;
	}

	/* ------------------------------------------------------------ grain */

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

	/* ------------------------------------------------------------ vignette */

	/* Two vignettes, one per theme, cross-faded by opacity: a gradient's colours cannot
	   interpolate, so the only way to ease this on a theme switch is to swap layers. */
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
