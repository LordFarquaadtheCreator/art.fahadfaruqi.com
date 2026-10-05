<script lang="ts">
	import Card from '../Card.svelte';
	import type { Section } from '../types';

	let { section }: { section: Section } = $props();

	// Its own knob: how far the card travels as it turns through.
	const rise = $derived(section.params.rise ?? 24);
</script>

<div class="rise" style="--rise: {rise}vh">
	<Card {section} />
</div>

<style>
	.rise {
		will-change: transform, opacity;
	}

	/* The default turn: the card rises from below, holds at the centre, leaves above. */
	:global(.reel[data-reel='on']) .rise {
		opacity: var(--gate, 1);
		transform: translateY(calc(var(--c, 0) * -1 * var(--rise, 24vh)))
			scale(calc(0.95 + 0.05 * var(--gate, 1)));
	}
</style>
