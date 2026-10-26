<script lang="ts">
	import { onMount } from 'svelte';
	import { mountRose } from '$lib/rose/mount';

	let canvas: HTMLCanvasElement | undefined = $state();

	onMount(() => {
		if (!canvas) return;
		return mountRose(canvas);
	});
</script>

<canvas bind:this={canvas} class="rose" aria-hidden="true"></canvas>

<style>
	.rose {
		position: absolute;
		right: 0;
		bottom: calc(-1 * var(--rose-h) * var(--rose-trim, 0));
		height: var(--rose-h);
		aspect-ratio: 900 / 1040;
		/* the drawing's ink stops 14% inside the art's own right edge: pull it out to the corner */
		transform: translateX(14%);
		opacity: var(--rose-opacity);
		filter: invert(var(--rose-invert));
		pointer-events: none;
		display: block;
	}
</style>
