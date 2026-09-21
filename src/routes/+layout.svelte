<script lang="ts">
	import '@fontsource-variable/inter';
	import '../app.css';

	let { children } = $props();
</script>

<svelte:head>
	<!-- Same mark as the main site at fahadfaruqi.com. -->
	<link rel="icon" href="/favicon.png" />
</svelte:head>

{@render children()}
