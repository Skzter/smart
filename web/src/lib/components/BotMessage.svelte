<script lang="ts">
    import { Bot } from "@lucide/svelte";
    import TestButtons from "./TestButtons.svelte";
    import { Runner } from "$lib/runner.svelte";
    import MonacoEditor from "./MonacoEditor.svelte";
    import { type Message, chat, user } from "$lib/shared.svelte";

    let {
        msg,
    }: {
        msg: Message;
    } = $props();

    function isCode(msg: Message): boolean {
        return msg.t == "generation";
    }

    // treat message as code when it looks like code or contains Playwright imports/markers
    let message = $derived(msg.Message);
    let runner = $derived(
        user.id ? new Runner(chat.id, user.id, chat.lastTest ?? "") : null,
    );
</script>

<div class="flex justify-start gap-2 items-start">
    <div
        class="h-8 w-8 shrink-0 rounded-full bg-muted flex items-center justify-center"
    >
        <Bot class="h-5 w-5" />
    </div>

    <div class="bg-muted rounded-2xl rounded-tl-sm w-[80%] overflow-hidden">
        {#if runner}
            <TestButtons
                iscode={isCode(msg)}
                bind:message
                testRunner={runner}
            />
        {/if}

        {#if isCode(msg)}
            <MonacoEditor
                bind:value={message}
                class="w-full h-full min-h-[200px]"
                options={{ useTextHeight: true, maxHeight: 600 }}
            />
        {:else}
            <div class="px-4 py-2 wrap-break-word whitespace-pre-wrap">
                {message}
            </div>
        {/if}
    </div>
</div>
