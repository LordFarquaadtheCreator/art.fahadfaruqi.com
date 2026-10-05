<script lang="ts">
	import TheTurn from '$lib/components/reel/TheTurn.svelte';
	import { turn } from '$lib/content/about-reel';

	const pad = (value: number) => String(value).padStart(2, '0');
</script>

<svelte:head>
	<title>About — Fahad Faruqi</title>
	<meta name="description" content="About Fahad Faruqi — his photographs, in one turn." />
</svelte:head>

<section class="screen screen--head">
	<div class="screen__pin">
		<p class="label">About — one turn</p>
		<h1 class="title large">Fahad Faruqi</h1>
		<div class="hairline screen__meta">
			<p class="label">Photography</p>
			<p class="label">Nikon D3300</p>
			<p class="label">Queens, New York</p>
			<p class="label">{turn.sections.length} turns</p>
		</div>
		<div class="screen__hint">
			<p class="label">Scroll to turn</p>
			<p class="num screen__proto">
				prototype — copy is draft; every turn lives in src/lib/content/about-reel.json
			</p>
		</div>
	</div>
</section>

<TheTurn />

<section class="screen screen--tail">
	<div class="screen__pin">
		<p class="label">End of turn — {pad(turn.sections.length)} beats</p>
		<h2 class="title medium">That's the turn.</h2>
		<p class="subtitle">The photographs are in the archive, from the beginning.</p>
		<a class="nav-button label screen__back" href="/">Full archive <span aria-hidden="true">→</span></a>
	</div>
</section>

<style>
	.screen {
		position: relative;
		height: 150vh;
	}

	.screen--tail {
		height: 140vh;
	}

	.screen__pin {
		position: sticky;
		top: 0;
		min-height: 100dvh;
		display: flex;
		flex-direction: column;
		justify-content: center;
		gap: clamp(1.5rem, 5vh, 3rem);
		padding-top: var(--header-h);
	}

	.screen__meta {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(12rem, 1fr));
		gap: 0.5rem 1.5rem;
	}

	.screen__meta p,
	.screen__hint p {
		margin: 0;
	}

	.screen__hint {
		display: grid;
		gap: 0.5rem;
	}

	.screen__proto {
		color: var(--faint);
		font-size: 0.7rem;
	}

	.screen__back {
		align-self: flex-start;
	}
</style>
