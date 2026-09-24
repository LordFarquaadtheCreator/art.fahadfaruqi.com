<script lang="ts">
	import Atmosphere from '$lib/components/Atmosphere.svelte';
	import ThemeToggle from '$lib/components/ThemeToggle.svelte';

	interface Props {
		status: number;
		/** Null only in the states the error page should not be rendered in at all: the
		 *  readout then carries the status alone rather than a sentence about it. */
		message: string | null;
		/** The address that was asked for. It is the one fact about this page the visitor
		 *  cannot see anywhere else, and the only clue to which link is broken. */
		path: string;
	}

	let { status, message, path }: Props = $props();

	// A wrong address is the only way in here that is nobody's fault, so it gets the one
	// line written by hand. Anything else says what actually happened rather than dressing
	// a failure up.
	const line = $derived(status === 404 ? 'You must be lost.' : (message ?? ''));
</script>

<svelte:head>
	<title>{status} — Fahad Faruqi</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<div class="error">
	<header class="bar shell">
		<a class="bar__mark" href="/">Fahad Faruqi</a>
		<ThemeToggle />
	</header>

	<main class="miss shell">
		<p class="miss__code">{status}</p>
		{#if line}
			<p class="miss__line">{line}</p>
		{/if}
		<a class="miss__back label" href="/">Back to the gallery <span aria-hidden="true">→</span></a>
	</main>

	<footer class="readouts shell">
		<span class="label num readout">
			{#if message}{status} {message}{:else}{status}{/if}
		</span>
		<span class="label readout readouts__path">{path}</span>
	</footer>
</div>

<Atmosphere />

<style>
	/* The page is the whole document here, so it takes the viewport directly: the bar, the
	   empty middle, and the readout line, in the same order as the gallery. */
	.error {
		position: relative;
		z-index: 1;
		display: grid;
		grid-template-rows: auto 1fr auto;
		min-height: 100dvh;
	}

	.bar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1.5rem;
		padding-block: 0.85rem;
		border-bottom: 1px solid var(--line);
	}

	.bar__mark {
		position: relative;
		font-size: 0.8125rem;
		font-weight: 600;
		letter-spacing: 0.02em;
		white-space: nowrap;
	}

	/* The masthead's mark, without the index beside it: this is the same bar. */
	.bar__mark::after {
		content: '';
		position: absolute;
		left: 0;
		right: 0;
		bottom: -0.15rem;
		height: 1px;
		background: var(--line-2);
		transform: scaleX(0);
		transform-origin: left center;
		transition: transform 0.45s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	.bar__mark:hover::after,
	.bar__mark:focus-visible::after {
		transform: scaleX(1);
	}

	.miss {
		display: flex;
		flex-direction: column;
		justify-content: center;
		gap: clamp(1.25rem, 4vh, 2.5rem);
		padding-block: clamp(2rem, 8vh, 6rem);
	}

	.miss__code {
		margin: 0;
		font-size: clamp(4.5rem, 22vw, 18rem);
		font-weight: 600;
		line-height: 0.82;
		letter-spacing: -0.05em;
	}

	.miss__line {
		margin: 0;
		max-width: 34ch;
		font-size: var(--lead);
		letter-spacing: -0.02em;
	}

	/* The way back is the only control on the page, so it is the only thing that moves. */
	.miss__back {
		align-self: flex-start;
		padding-top: 0.75rem;
		border-top: 1px solid var(--line);
		color: var(--muted);
		transition:
			color 0.25s ease,
			border-color 0.25s ease;
	}

	.miss__back span {
		display: inline-block;
		transition: transform 0.25s ease;
	}

	.miss__back:hover,
	.miss__back:focus-visible {
		color: var(--fg);
		border-top-color: var(--accent);
	}

	.miss__back:hover span {
		transform: translateX(0.3rem);
	}

	.readouts {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 1.5rem;
		padding-block: clamp(1rem, 3vh, 1.75rem);
		border-top: 1px solid var(--line);
	}

	/* A path is not a label: it keeps its own case, and gives way first when the window
	   is narrow. */
	.readout {
		text-transform: none;
	}

	.readouts__path {
		min-width: 0;
		overflow: hidden;
		white-space: nowrap;
		text-overflow: ellipsis;
		color: var(--faint);
	}

	/* The page assembles itself in reading order. Like every other entrance here it lives
	   in a keyframes `from` frame, so a page that never runs script is unanimated, not
	   empty. */
	.bar > *,
	.miss > *,
	.readouts > * {
		animation: enter 0.75s cubic-bezier(0.16, 0.84, 0.28, 1) both;
	}

	.bar > :nth-child(2) {
		animation-delay: 0.08s;
	}

	/* Named rather than counted: the line is absent in a state that has no message, and the
	   order of what follows it should not move when it goes. */
	.miss__line {
		animation-delay: 0.2s;
	}

	.miss__back {
		animation-delay: 0.28s;
	}

	.readouts > * {
		animation-delay: 0.36s;
	}

	@keyframes enter {
		from {
			opacity: 0;
			transform: translate3d(0, 0.75rem, 0);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}
</style>
