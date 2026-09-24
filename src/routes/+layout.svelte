<script lang="ts">
	import '@fontsource-variable/inter';
	import '@fontsource/ibm-plex-mono/400.css';
	import '../app.css';
	import Atmosphere from '$lib/components/Atmosphere.svelte';
	import Footer from '$lib/components/Footer.svelte';
	import Header from '$lib/components/Header.svelte';
	import ToTop from '$lib/components/ToTop.svelte';

	let { children } = $props();

	let scrolled = $state(false);
	let progress = $state(0);

	// The document's scroll state
	$effect(() => {
		const read = () => {
			const y = window.scrollY;
			const span = document.documentElement.scrollHeight - window.innerHeight;
			scrolled = y > 24;
			progress = span > 0 ? Math.min(1, y / span) : 0;
		};

		read();
		window.addEventListener('scroll', read, { passive: true });
		window.addEventListener('resize', read, { passive: true });

		return () => {
			window.removeEventListener('scroll', read);
			window.removeEventListener('resize', read);
		};
	});
</script>

<svelte:head>
	<link rel="icon" href="/favicon.png" />
</svelte:head>

<div class="site">
	<Header {scrolled} {progress} />

	<main class="shell">
		{@render children()}
	</main>

	<Footer />
</div>

<ToTop {progress} />

<Atmosphere />

<style>
	.site {
		display: flex;
		flex-direction: column;
		min-height: 100dvh;
	}
	main {
		position: relative;
		z-index: 1;
		flex: 1;
		display: flex;
		flex-direction: column;
	}
</style>
