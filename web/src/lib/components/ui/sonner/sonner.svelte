<script lang="ts">
	import { Toaster as Sonner, type ToasterProps as SonnerProps } from "svelte-sonner";
	import { mode } from "mode-watcher";

	let { ...restProps }: SonnerProps = $props();
</script>

<Sonner
	theme={mode.current}
	class="toaster group text-lg"
	style="--normal-bg: #22c55e; --normal-text: #fff; --normal-border: #22c55e;"
	position="top-right" duration={2500}
	{...restProps}
/>
