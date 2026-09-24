<script lang="ts">
	interface Props {
		status: number;
		message: string | null;
		path: string;
	}

	let { status, message, path }: Props = $props();

	// todo: wrap in a switch
	const title = $derived(status === 404 ? 'You must be lost.' : (message ?? ''));
	const subtitle = 'These are not the Jedi you are looking for'
</script>

<svelte:head>
	<title>{status} — Fahad Faruqi</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<section class="error shell">
	<p class="error__code">{status}</p>
	<p class="error__title">{title}</p>
	<p class="error__subtitle">{subtitle}</p>
	<a class="error__back label" href="/">Back to the gallery <span aria-hidden="true">→</span></a>
</section>

<style>
	.error {
		flex: 1;
		display: flex;
		flex-direction: column;
		justify-content: center;
		gap: clamp(1.25rem, 4vh, 2.5rem);
		padding-block: clamp(2rem, 8vh, 6rem);
	}

	.error__code {
		margin: 0;
		font-size: clamp(4.5rem, 22vw, 18rem);
		font-weight: 600;
		line-height: 0.82;
		letter-spacing: -0.05em;
	}

	.error__title {
		margin: 0;
		max-width: 34ch;
		font-size: var(--lead);
		letter-spacing: -0.02em;
	}

	.error__subtitle {
		margin: 0;
		max-width: 34ch;
		font-size: clamp(0.875rem, 1.5vw, 1.125rem);
		line-height: 1.5;
		color: var(--muted);
	}

	.error__back {
		align-self: flex-start;
		padding-top: 0.75rem;
		border-top: 1px solid var(--line);
		color: var(--muted);
		transition:
			color 0.25s ease,
			border-color 0.25s ease;
	}

	.error__back span {
		display: inline-block;
		transition: transform 0.25s ease;
	}

	.error__back:hover,
	.error__back:focus-visible {
		color: var(--fg);
		border-top-color: var(--accent);
	}
	.error__back:hover span {
		transform: translateX(0.3rem);
	}
	
	/* animation staggered by position. */
	.error > * {
		animation: enter 0.75s cubic-bezier(0.16, 0.84, 0.28, 1) both;
		animation-delay: calc((sibling-index() - 1) * 80ms);
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
