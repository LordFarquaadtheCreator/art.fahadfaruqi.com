<script lang="ts">
	import { page } from '$app/state';
	import { navLine } from '$lib/actions/nav-line';

	// What every `+error.svelte` renders with nothing passed: the router already holds the
	// status, the message and the address. The overrides are for a failure that came from a
	// page instead — its own status, its own reason, and a retry rather than a link.
	let {
		status = page.status,
		detail = null,
		onAction = null,
		href = '/'
	}: {
		status?: number;
		detail?: string | null;
		onAction?: (() => void) | null;
		href?: string | null;
	} = $props();

	const path = $derived(page.url.pathname.substr(1));
	const message = $derived(page.error?.message ?? null);

	const copy = $derived.by(() => {
		switch (status) {
			case 404:
				return {
					title: 'You must be lost.',
					subtitle: path ? `${path} does not exist` : 'That address does not exist',
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

<section class="error">
	<p class="title large fade-in-animation">{status}</p>
	<div class="hairline error__copy">
		<p class="title fade-in-animation">{copy.title}</p>
		<p class="subtitle fade-in-animation">{copy.subtitle}</p>
		{#if detail}
			<p class="num error__detail fade-in-animation">{detail}</p>
		{/if}
		{#if onAction}
			<button
				class="nav-button label error__back fade-in-animation"
				type="button"
				onclick={() => onAction()}
				use:navLine
			>
				{copy.action}
				<span aria-hidden="true">→</span>
			</button>
		{:else if href}
			<a class="nav-button label error__back fade-in-animation" {href} use:navLine>
				{copy.action}
				<span aria-hidden="true">→</span>
			</a>
		{/if}
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

	.error__detail {
		margin: -0.5rem 0 0;
		font-size: 0.75rem;
		color: var(--muted);
		overflow-wrap: anywhere;
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
