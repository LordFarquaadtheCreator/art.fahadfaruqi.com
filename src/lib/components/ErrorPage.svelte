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
	<p class="title-attention fade-in-animation">{status}</p>
	<p class="title fade-in-animation">{title}</p>
	<p class="subtitle fade-in-animation">{subtitle}</p>
	<a class="nav-button label error__back fade-in-animation" href="/">Back to the gallery <span aria-hidden="true">→</span></a>
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

	/* The look lives in the shared .title-attention / .title / .subtitle / .nav-button; what
	   is left here is where this page puts them. */
	.error__back {
		align-self: flex-start;
	}
</style>
