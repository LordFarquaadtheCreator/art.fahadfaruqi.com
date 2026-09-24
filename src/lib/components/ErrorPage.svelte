<script lang="ts">
	import { navLine } from '$lib/actions/nav-line';

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
	<p class="title-attention fade-in-animation">{status}</p>
	<p class="title fade-in-animation">{title}</p>
	<p class="subtitle fade-in-animation">{subtitle}</p>
	<a
		class="nav-button label error__back fade-in-animation"
		href="/"
		use:navLine
	>
		Back to the gallery <span aria-hidden="true">→</span>
	</a>
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
	
	.error__back {
		align-self: flex-start;
	}

	.error__back span {
		display: inline-block;
		transition: transform 0.25s ease;
	}

	.error__back:hover span,
	.error__back:focus-visible span {
		transform: translateX(0.3rem);
	}
</style>
