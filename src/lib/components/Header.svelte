<script lang="ts">
	import { page } from '$app/state';
	import ThemeToggle from './ThemeToggle.svelte';

	let { scrolled, progress }: { scrolled: boolean; progress: number } = $props();

	const tabs = [
		{ href: '/', label: 'Photos' },
		{ href: '/about', label: 'About' }
	];

	const current = $derived(page.url.pathname);
</script>

<header class="header" class:header--scrolled={scrolled}>
	<div class="header__inner shell">
		<a class="header__mark" href="/">Fahad Faruqi</a>

		<nav class="tabs" aria-label="Sections">
			{#each tabs as tab (tab.href)}
				{@const active = current === tab.href}
				<a
					class="tabs__item label"
					class:tabs__item--active={active}
					href={tab.href}
					aria-current={active ? 'page' : undefined}
				>
					{tab.label}
				</a>
			{/each}
		</nav>

		<ThemeToggle />
	</div>

	<span class="header__progress" aria-hidden="true" style="--progress: {progress}"></span>
</header>

<style>
	.header {
		position: sticky;
		top: 0;
		z-index: 20;
		background: color-mix(in srgb, var(--bg) 86%, transparent);
		backdrop-filter: blur(10px);
		border-bottom: 1px solid var(--line);
		transition:
			background-color 0.45s ease,
			border-color 0.45s ease,
			backdrop-filter 0.45s ease;
	}

	.header--scrolled {
		background: color-mix(in srgb, var(--bg) 72%, transparent);
		backdrop-filter: blur(18px);
		border-bottom-color: var(--line-2);
	}

	.header__inner {
		position: relative;
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 1.5rem;
		padding-block: 0.85rem;
		transition: padding 0.45s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	.header--scrolled .header__inner {
		padding-block: 0.6rem;
	}

	/* One hairline of how far through the document you are. */
	.header__progress {
		position: absolute;
		left: 0;
		right: 0;
		bottom: 0;
		height: 1px;
		background: var(--accent);
		transform: scaleX(var(--progress, 0));
		transform-origin: left center;
	}

	.header__inner > * {
		animation: header-in 0.7s cubic-bezier(0.16, 0.84, 0.28, 1) both;
	}

	.header__inner > :nth-child(1) {
		animation-delay: 0.04s;
	}

	.header__inner > :nth-child(2) {
		animation-delay: 0.11s;
	}

	.header__inner > :nth-child(3) {
		animation-delay: 0.18s;
	}

	@keyframes header-in {
		from {
			opacity: 0;
			transform: translate3d(0, -0.5rem, 0);
		}
		to {
			opacity: 1;
			transform: none;
		}
	}

	.header__mark {
		position: relative;
		font-size: 0.8125rem;
		font-weight: 600;
		letter-spacing: 0.02em;
		white-space: nowrap;
	}

	/* The mark draws its own underline on approach. */
	.header__mark::after {
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

	.header__mark:hover::after,
	.header__mark:focus-visible::after {
		transform: scaleX(1);
	}

	/* -------------------------------------------------------------- the sections */

	.tabs {
		display: flex;
		gap: 1.75rem;
	}

	.tabs__item {
		position: relative;
		padding-block: 0.125rem;
		color: var(--muted);
		transition: color 0.25s ease;
	}

	.tabs__item::after {
		content: '';
		position: absolute;
		left: 0;
		right: 0;
		bottom: -0.2rem;
		height: 1px;
		background: var(--accent);
		transform: scaleX(0);
		transform-origin: left center;
		transition: transform 0.45s cubic-bezier(0.16, 0.84, 0.28, 1);
	}

	.tabs__item:hover {
		color: var(--fg);
	}

	.tabs__item--active {
		color: var(--fg);
	}

	.tabs__item--active::after {
		transform: scaleX(1);
	}
</style>
