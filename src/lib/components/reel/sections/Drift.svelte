<script lang="ts">
	import Card from '../Card.svelte';
	import type { Section } from '../types';

	let { section }: { section: Section } = $props();
</script>

<div class="drift">
	<Card {section} />
</div>

<style>
	.drift {
		will-change: transform, opacity;
	}

	/* Its own bit: sideways. The card drifts in from the right with a slight tilt, settles
	   at the centre, and drifts out to the left. */
	:global(.reel[data-reel='on']) .drift {
		opacity: var(--gate, 1);
		transform: translateX(calc(var(--c, 0) * -18vw)) rotate(calc(var(--c, 0) * 1.5deg));
	}
</style>
