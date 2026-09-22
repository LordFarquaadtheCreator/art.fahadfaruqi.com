<script lang="ts">
	import '@fontsource-variable/inter';
	import '@fontsource/ibm-plex-mono/400.css';
	import '../app.css';

	let { children } = $props();
</script>

<svelte:head>
	<link rel="icon" href="/favicon.png" />
</svelte:head>

{@render children()}
