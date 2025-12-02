<script lang="ts">
    import {
        Copy,
        Edit,
        Play,
        Save,
        User,
        Bot,
        Plus,
        Send,
    } from "@lucide/svelte";
    import Button from "./ui/button/button.svelte";
    import * as ButtonGroup from "$lib/components/ui/button-group";
    import * as InputGroup from "$lib/components/ui/input-group";
    import type { Message } from "src/types/message";

    let isLoading = $state(false);
    let messages = $state<Message[]>([]);
    let input = $state("");

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
            <!-- CHAT MESSAGES -->
            {#if messages.length === 0}
                <div
                    class="flex items-center justify-center flex-1 text-muted-foreground"
                >
                    <p>Start a chat...</p>
                </div>
            {:else}
                <div class="flex-1"></div>
                <div class="flex flex-col gap-4">
                    {#each messages as message}
                        <!-- User message -->
                        <div class="flex justify-end gap-2 items-start">
                            <div
                                class="bg-primary text-primary-foreground rounded-2xl rounded-br-sm px-4 py-2 max-w-[80%] break-words whitespace-pre-wrap"
                            >
                                {message.question}
                            </div>
                            <div
                                class="h-8 w-8 shrink-0 rounded-full bg-primary flex items-center justify-center"
                            >
                                <User class="h-7 w-7 text-primary-foreground" />
                            </div>
                        </div>

                        <!-- Bot response -->
                        {#if message.answer}
                            <div class="flex justify-start gap-2 items-start">
                                <div
                                    class="h-8 w-8 shrink-0 rounded-full bg-muted flex items-center justify-center"
                                >
                                    <Bot class="h-5 w-5" />
                                </div>
                                <div
                                    class="bg-muted text-foreground rounded-2xl rounded-tl-sm max-w-[80%] overflow-hidden"
                                >
                                    <div
                                        class="flex justify-end gap-1 px-3 py-2 bg-muted/40 border-b border-border/50"
                                    >
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            class="h-7 gap-1.5 px-2"
                                        >
                                            <Copy class="h-3.5 w-3.5" />
                                            <span class="text-xs">Kopieren</span
                                            >
                                        </Button>
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            class="h-7 gap-1.5 px-2"
                                        >
                                            <Edit class="h-3.5 w-3.5" />
                                            <span class="text-xs"
                                                >Bearbeiten</span
                                            >
                                        </Button>
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            class="h-7 gap-1.5 px-2"
                                            onclick={() =>
                                                handleSaveTest(message.answer)}
                                        >
                                            <Save class="h-3.5 w-3.5" />
                                            <span class="text-xs"
                                                >Speichern</span
                                            >
                                        </Button>
                                        <Button
                                            variant="ghost"
                                            size="sm"
                                            class="h-7 gap-1.5 px-2"
                                        >
                                            <Play class="h-3.5 w-3.5" />
                                            <span class="text-xs"
                                                >Ausführen</span
                                            >
                                        </Button>
                                    </div>
                                    <div
                                        class="px-4 py-2 break-words whitespace-pre-wrap"
                                    >
                                        {message.answer}
                                    </div>
                                </div>
                            </div>
                        {/if}
                    {/each}
                    <!-- Loading indicator -->
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

<div class="p-4 pt-0 sticky bottom-0 bg-background z-10">
    <div class="flex items-center justify-center w-full">
        <div class="w-full max-w-6xl">
            <ButtonGroup.Root
                class="[--radius:1rem] w-full flex items-center justify-center gap-3"
            >
                <ButtonGroup.Root class="shrink-0">
                    <Button variant="outline" size="icon" class="w-10 h-10">
                        <Plus />
                    </Button>
                </ButtonGroup.Root>
                <ButtonGroup.Root class="flex-1">
                    <div class="flex w-full items-center gap-2">
                        <InputGroup.Root class="w-full">
                            <InputGroup.Textarea
                                bind:value={input}
                                class="w-full resize-none min-h-11"
                                placeholder="Send a message..."
                                rows={1}
                                disabled={isLoading}
                            />
                        </InputGroup.Root>
                        <Button
                            variant="default"
                            size="icon"
                            class="w-10 h-10"
                            disabled={isLoading}
                        >
                            <Send />
                        </Button>
                    </div>
                </ButtonGroup.Root>
            </ButtonGroup.Root>
        </div>
    </div>
</div>
