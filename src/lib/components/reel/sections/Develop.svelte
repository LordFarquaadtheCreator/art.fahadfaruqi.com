<script lang="ts">
	import Card from '../Card.svelte';
	import type { Section } from '../types';

	let { section }: { section: Section } = $props();
</script>

<div class="develop">
	<Card {section} />
</div>

<style>
	.develop {
		will-change: opacity, filter;
	}

	/* Its own bit: the card does not travel — it develops. The frame wipes open out of the
	   dark as the turn brings it to the centre, and goes latent again as it leaves. */
	:global(.reel[data-reel='on']) .develop {
		opacity: calc(0.35 + 0.65 * var(--gate, 1));
	}

	:global(.reel[data-reel='on']) .develop :global(.frame--image .frame__inner) {
		clip-path: inset(0 calc((1 - var(--gate, 1)) * 100%) 0 0);
		filter: brightness(calc(0.35 + 0.65 * var(--gate, 1)));
	}
</style>
