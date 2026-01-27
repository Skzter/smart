<script lang="ts">
    import { onMount } from "svelte";

    let videoEl: HTMLVideoElement | null = null;
    let containerEl: HTMLElement | null = null;

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
        <video
            bind:this={videoEl}
            controls
            src="https://archive.org/download/ElephantsDream/ed_1024_512kb.mp4"
        >
            <track kind="captions" />
        </video>
    </div>
</div>
