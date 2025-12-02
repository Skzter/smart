<script lang="ts">
    import { Bot } from "@lucide/svelte";
    import type { Message } from "src/types/message";
    import UserMessage from "./UserMessage.svelte";
    import BotMessage from "./BotMessage.svelte";

    let { isLoading = $bindable() }: { isLoading: boolean } = $props();

    let conversationId = $state("");
    let userId = $state("auth0|687270280dca20b77cfdcf73");
    let messages = $state<Message[]>([]);
    messages = [
        {
            question: "What is the capital of France?",
            answer: "Paris",
        },
        {
            question: "What is 2 + 2?",
            answer: "4",
        },
        {
            question: "Who wrote '1984'?",
            answer: "George Orwell",
        },
    ];

    let container: HTMLElement | undefined = $state();
    // Effect to trigger scrolling on relevant changes
    $effect(() => {
        if (container && (isLoading || messages.length > 0)) {
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
                        <!-- Bot response -->
                        {#if message.answer}
                            <BotMessage
                                message={message.answer}
                                {userId}
                                bind:conversationId
                            />
                        {/if}
                    {/each}
                    <!-- Loading indicator, auch eigener component und vllt auf vorgefertigte component zurück greifen -->
                    {#if isLoading}
                        <div class="flex justify-start gap-2 items-start">
                            <div
                                class="h-8 w-8 shrink-0 rounded-full bg-muted flex items-center justify-center"
                            >
                                <Bot class="h-7 w-7" />
                            </div>
                            <div
                                class="bg-muted text-foreground rounded-2xl rounded-bl-sm px-4 py-2"
                            >
                                <div class="flex gap-1">
                                    <span
                                        class="animate-bounce"
                                        style="animation-delay: 0ms;">●</span
                                    >
                                    <span
                                        class="animate-bounce"
                                        style="animation-delay: 150ms;">●</span
                                    >
                                    <span
                                        class="animate-bounce"
                                        style="animation-delay: 300ms;">●</span
                                    >
                                </div>
                            </div>
                        </div>
                    {/if}
                </div>
            {/if}
        </div>
    </div>
</div>
