<script lang="ts">
    import type { Runner } from "$lib/runner.svelte";
    import { onMount } from "svelte";
    import * as UnderlineTabs from "$lib/components/ui/underline-tabs";
    import * as Dialog from "$lib/components/ui/dialog";
    import { ZoomIn, ZoomOut, Maximize2 } from "@lucide/svelte";

    let {
        runner,
    }: {
        runner: Runner;
    } = $props();

    let videoEl = $state<HTMLElement | null>(null);
    let containerEl = $state<HTMLElement | null>(null);
    let activePreviewTab = $state("screenshot");
    let screenshotLightboxOpen = $state(false);
    let screenshotZoom = $state(1);
    let screenshotNaturalWidth = $state(0);
    let screenshotNaturalHeight = $state(0);

    $effect(() => {
        const status = runner.model.summary.status;
        const hasTest = runner.getCurTest() !== "";
        if (status === "failed") {
            runner.fetchMediaUrl();
        } else if (status === "idle" && hasTest) {
            runner.fetchMediaUrl();
        } else {
            runner.clearVideoUrl();
        }
    });

    // Reset lightbox dimensions when screenshot URL changes so new image gets correct size
    $effect(() => {
        runner.screenshotUrl = null;
        runner.videoUrl = null;
        screenshotNaturalWidth = 0;
        screenshotNaturalHeight = 0;
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

        videoEl.style.width = `${calculatedWidth}px`;
        videoEl.style.height = `${availableHeight}px`;
    }

    function openScreenshotLightbox() {
        screenshotZoom = 1;
        screenshotLightboxOpen = true;
    }

    function onScreenshotLoad(e: Event) {
        const img = e.currentTarget as HTMLImageElement;
        if (img?.naturalWidth != null) {
            screenshotNaturalWidth = img.naturalWidth;
            screenshotNaturalHeight = img.naturalHeight;
        }
    }

    function zoomScreenshot(delta: number) {
        screenshotZoom = Math.max(0.25, Math.min(4, screenshotZoom + delta));
    }

    function handleScreenshotWheel(e: WheelEvent) {
        if (!screenshotLightboxOpen) return;
        e.preventDefault();
        zoomScreenshot(e.deltaY > 0 ? -0.1 : 0.1);
    }
</script>

<div bind:this={containerEl} id="container" class="flex flex-col h-full">
    <UnderlineTabs.Root
        value={activePreviewTab}
        onValueChange={(v) => (v && (activePreviewTab = v))}
        class="flex flex-col h-full"
    >
    <div
        class="px-4 py-2 border-b bg-muted/50 text-sm font-medium flex items-center gap-2 shrink-0"
    >
        <span>Vorschau</span>
        <UnderlineTabs.List class="border-0 p-0 min-h-0 h-auto gap-4 ml-2">
            <UnderlineTabs.Trigger value="screenshot"
                >Screenshot</UnderlineTabs.Trigger
            >
            <UnderlineTabs.Trigger value="video">Video</UnderlineTabs.Trigger>
        </UnderlineTabs.List>
    </div>

    <div class="flex justify-center items-center flex-1 overflow-hidden min-h-0">
        <UnderlineTabs.Content
            value="screenshot"
            class="flex-1 flex flex-col items-center justify-center overflow-auto p-2 data-[state=inactive]:hidden"
        >
            {#if runner.screenshotUrl != null}
                <div
                    class="relative max-w-full max-h-full flex items-center justify-center"
                >
                    <img
                        src={runner.screenshotUrl}
                        alt="Test screenshot"
                        class="max-w-full max-h-full object-contain"
                    />
                    <button
                        type="button"
                        onclick={openScreenshotLightbox}
                        class="absolute bottom-2 right-2 rounded-md bg-black/60 text-white p-2 hover:bg-black/80 transition-colors"
                        title="In größerem Fenster öffnen"
                    >
                        <Maximize2 size={18} />
                    </button>
                </div>
            {:else}
                <p class="text-muted-foreground text-sm">
                    Screenshot wird nach Fehlschlag angezeigt
                </p>
            {/if}
        </UnderlineTabs.Content>
        <UnderlineTabs.Content
            value="video"
            class="flex-1 flex items-center justify-center overflow-hidden w-full data-[state=inactive]:hidden"
        >
            {#if runner.videoUrl != null}
                <video bind:this={videoEl} controls src={runner.videoUrl}>
                    <track kind="captions" />
                </video>
            {:else}
                <p class="text-muted-foreground text-sm">
                    Video wird nach Fehlschlag angezeigt
                </p>
            {/if}
        </UnderlineTabs.Content>
    </div>
</UnderlineTabs.Root>
</div>

<!-- Screenshot lightbox with zoom -->
<Dialog.Root bind:open={screenshotLightboxOpen}>
    <Dialog.Content
        class="!max-w-[95vw] w-[95vw] max-h-[90vh] h-[90vh] flex flex-col p-4"
        showCloseButton={false}
    >
        <Dialog.Header class="flex flex-row items-center justify-between gap-4 shrink-0">
            <Dialog.Title>Screenshot</Dialog.Title>
            <div class="flex items-center gap-2">
                <button
                    type="button"
                    onclick={() => zoomScreenshot(-0.25)}
                    class="rounded-md border bg-background p-2 hover:bg-muted"
                    title="Verkleinern"
                >
                    <ZoomOut size={18} />
                </button>
                <span class="text-sm tabular-nums min-w-[3rem] text-center"
                    >{Math.round(screenshotZoom * 100)}%</span
                >
                <button
                    type="button"
                    onclick={() => zoomScreenshot(0.25)}
                    class="rounded-md border bg-background p-2 hover:bg-muted"
                    title="Vergrößern"
                >
                    <ZoomIn size={18} />
                </button>
            </div>
        </Dialog.Header>
        <div
            class="flex-1 overflow-auto min-h-0 p-4"
            onwheel={handleScreenshotWheel}
            role="presentation"
        >
            {#if runner.screenshotUrl != null}
                {@const w =
                    screenshotNaturalWidth > 0
                        ? screenshotNaturalWidth * screenshotZoom
                        : null}
                {@const h =
                    screenshotNaturalHeight > 0
                        ? screenshotNaturalHeight * screenshotZoom
                        : null}
                <div
                    class="min-w-full min-h-full"
                    style="width: {w != null ? `${w}px` : 'auto'}; height: {h != null ? `${h}px` : 'auto'};"
                >
                    <img
                        src={runner.screenshotUrl}
                        alt="Screenshot (Vergrößert)"
                        class="w-full h-full object-contain object-left-top"
                        onload={onScreenshotLoad}
                    />
                </div>
            {/if}
        </div>
        <Dialog.Footer class="shrink-0">
            <Dialog.Close>Schließen</Dialog.Close>
        </Dialog.Footer>
    </Dialog.Content>
</Dialog.Root>
