<script lang="ts">
    import { Bot } from "@lucide/svelte";
    import Code from "./Code.svelte";
    import TestButtons from "./TestButtons.svelte";
    import { Runner } from "$lib/runner.svelte";

    let {
        msg,
    }: {
        msg: string;
    } = $props();

    // treat message as code when it looks like code or contains Playwright imports/markers
    const messageIsCode = $derived((msg || "").includes("@playwright"));
    let message = $derived(msg);

    let runner = new Runner();
</script>

<div class="flex justify-start gap-2 items-start">
    <div
        class="h-8 w-8 shrink-0 rounded-full bg-muted flex items-center justify-center"
    >
        <Bot class="h-5 w-5" />
    </div>

    <div class="bg-muted rounded-2xl rounded-tl-sm max-w-[80%] overflow-hidden">
        <TestButtons
            iscode={messageIsCode}
            bind:message
            bind:testRunner={runner}
        ></TestButtons>

        {#if messageIsCode}
            <Code bind:code={message} />
        {:else}
            <div class="px-4 py-2 wrap-break-word whitespace-pre-wrap">
                {message}
            </div>
        {/if}
    </div>
</div>
