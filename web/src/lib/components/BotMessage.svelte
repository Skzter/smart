<script lang="ts">
    import { Bot } from "@lucide/svelte";
    import RunWindow from "./RunWindow.svelte";
    import SaveButton from "./SaveButton.svelte";
    import EditButton from "./EditButton.svelte";
    import CopyButton from "./CopyButton.svelte";
    import Code from "./Code.svelte";

    let {
        message,
    }: {
        message: string;
    } = $props();

    // treat message as code when it looks like code or contains Playwright imports/markers
    const messageIsCode = $derived(
        (message || "").includes("@playwright")
    );
</script>

<div class="flex justify-start gap-2 items-start">
    <div class="h-8 w-8 shrink-0 rounded-full bg-muted flex items-center justify-center">
        <Bot class="h-5 w-5" />
    </div>

    <div class="bg-muted text-foreground rounded-2xl rounded-tl-sm max-w-[80%] overflow-hidden">
        <div class="flex justify-end gap-1 px-3 py-2 bg-muted/40 border-b border-border/50">
            <CopyButton code={message} />
            {#if messageIsCode}
                <EditButton />
                <SaveButton code={message} />
                <RunWindow code={message} />
            {/if}
        </div>

        {#if messageIsCode}
            <Code code={message} />
        {:else}
            <div class="px-4 py-2 wrap-break-word whitespace-pre-wrap">{message}</div>
        {/if}
    </div>
</div>
