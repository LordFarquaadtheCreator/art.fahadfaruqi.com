<script lang="ts">
	import { navLine } from '$lib/actions/nav-line';

	interface Props {
		status: number;
		message: string | null;
		path: string;
	}

	let { status, message, path }: Props = $props();

	const copy = $derived.by(() => {
		switch (status) {
			case 404:
				return {
					title: 'You must be lost.',
					subtitle: 'There\'s nothing here for you.',
					action: 'My bad, let me get out of here'
				};
			case 500:
				return {
					title: 'Something broke on my end.',
					subtitle: 'Nothing to do with you. Try again in a moment.',
					action: 'Try the gallery again'
				};
			default:
				return {
					title: message ?? 'Something went wrong.',
					subtitle: 'This is embarassing.',
					action: 'Back to the gallery'
				};
		}
	});
</script>

<svelte:head>
	<title>{status} — Fahad Faruqi</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<section class="error shell">
	<p class="title large fade-in-animation">{status}</p>
	<div class="hairline error__copy">
		<p class="title fade-in-animation">{copy.title}</p>
		<p class="subtitle fade-in-animation">{copy.subtitle}</p>
		<a
			class="nav-button label error__back fade-in-animation"
			href="/"
			use:navLine
		>
			{copy.action}
			<span aria-hidden="true">→</span>
		</a>
	</div>
</section>

<style>
	.error {
		flex: 1;
		display: flex;
		flex-direction: column;
		justify-content: center;
		padding-block: clamp(2rem, 8vh, 6rem);
	}

	.error__copy {
		display: flex;
		flex-direction: column;
		gap: clamp(1.25rem, 4vh, 2.5rem);
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
