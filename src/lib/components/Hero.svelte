<script lang="ts">
	import { count } from '$lib/actions/count';
	import SplitText from './SplitText.svelte';

	let {
		status,
		photoCount,
		setCount
	}: {
		status: 'loading' | 'ready' | 'error';
		photoCount: number;
		setCount: number;
	} = $props();
</script>

<section class="hero">
	<h1 class="title large" aria-label="Fahad Faruqi">
		<SplitText text="Fahad Faruqi" />
	</h1>

	<div class="hairline hero__meta">
		<p class="label fade-in-animation">Photography</p>
		<p class="label fade-in-animation">Nikon D3300</p>
		<p class="label fade-in-animation">Queens, New York</p>
		<p class="label fade-in-animation">
			{#if status === 'ready'}
				<span use:count={photoCount}></span> photographs &middot; {setCount} sets
			{:else if status === 'loading'}
				Loading photos
			{:else}
				Photos unavailable
			{/if}
		</p>
	</div>
</section>

<style>
	.hero {
		padding-block: clamp(3rem, 12vh, 9rem) clamp(1rem, 3vh, 3rem);
	}

	.hero__meta {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
		gap: 0.5rem 1.5rem;
	}

	.hero__meta p {
		margin: 0;
	}
</style>
