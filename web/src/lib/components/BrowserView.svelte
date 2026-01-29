<script lang="ts">
    import type { Runner } from "$lib/runner.svelte";
    import { onMount } from "svelte";

    let {
        runner,
    }: {
        runner: Runner;
    } = $props();

    let videoEl = $state<HTMLElement | null>(null);
    let containerEl: HTMLElement | null = null;

    // Watch for test failures and fetch video URL
    $effect(() => {
        if (runner.model.summary.status === "failed") {
            runner.fetchMediaUrl();
        } else {
            runner.clearVideoUrl();
        }
    });

    onMount(() => {
        updateVideoSize();
        window.addEventListener("resize", updateVideoSize);
        return () => window.removeEventListener("resize", updateVideoSize);
    });

    function updateVideoSize() {
        if (!containerEl || !videoEl) return;

        const parent = containerEl.parentElement;
        if (!parent) return;

        const parentHeight = parent.clientHeight;
        const availableHeight = parentHeight - 50;
        const calculatedWidth = availableHeight * (16 / 9);

        console.log("parent height:", parentHeight);
        console.log("available height:", availableHeight);
        console.log("calculated width:", calculatedWidth);

        videoEl.style.width = `${calculatedWidth}px`;
        videoEl.style.height = `${availableHeight}px`;
    }
</script>

<div bind:this={containerEl} id="container" class="flex flex-col h-full">
    <div
        class="px-4 py-2 border-b bg-muted/50 text-sm font-medium flex items-center gap-2"
    >
        Vorschau
    </div>
    <div class="flex justify-center items-center flex-1 overflow-hidden">
        {#if runner.videoUrl != null}
            <video bind:this={videoEl} controls src={runner.videoUrl}>
                <track kind="captions" />
            </video>
        {:else}
            Video wird nach Fehlschlag angezeigt
        {/if}
    </div>
</div>
