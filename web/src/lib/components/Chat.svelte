<script lang="ts">
    import UserMessage from "./UserMessage.svelte";
    import BotMessage from "./BotMessage.svelte";
    import { chat, messages } from "$lib/shared.svelte";
    import Dots from "./Dots.svelte";

    let container: HTMLElement | undefined = $state();
    // Effect to trigger scrolling on relevant changes
    $effect(() => {
        if (container && (chat.isLoading || messages.length > 0)) {
            // Small timeout to ensure DOM updates are complete
            setTimeout(() => {
                container?.scrollTo({
                    top: container.scrollHeight,
                    behavior: "smooth",
                });
            }, 50);
        }
    });
</script>

<div class="flex flex-1 flex-col gap-4 p-4 pt-0 h-full">
    <div class="flex-1 flex items-start justify-center min-h-0">
        <div
            bind:this={container}
            class="w-full max-w-6xl bg-muted/50 h-full rounded-xl md:min-h-min overflow-auto p-6 min-h-0 flex flex-col"
        >
            {#if messages.length === 0}
                <div
                    class="flex items-center justify-center flex-1 text-muted-foreground"
                >
                    <p>Start a conversation...</p>
                </div>
            {:else}
                <div class="flex-1"></div>
                <div class="flex flex-col gap-4">
                    {#each messages as message}
                        <UserMessage message={message.question} />
                        {#if message.answer}
                            <BotMessage msg={message.answer} />
                        {/if}
                    {/each}
                    <!-- Loading indicator, auch eigener component und vllt auf vorgefertigte component zurück greifen -->
                    {#if chat.isLoading}
                        <Dots />
                    {/if}
                </div>
            {/if}
        </div>
    </div>
</div>
