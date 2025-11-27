<script lang="ts">
    import { onMount, afterUpdate } from 'svelte';
    import AppSidebar from "$lib/components/app-sidebar.svelte";
    import * as Breadcrumb from "$lib/components/ui/breadcrumb";
    import { Separator } from "$lib/components/ui/separator";
    import * as Sidebar from "$lib/components/ui/sidebar";
    import Send from "@lucide/svelte/icons/send";
    import Plus from "@lucide/svelte/icons/plus";
    import { Button } from "$lib/components/ui/button";
    import * as ButtonGroup from "$lib/components/ui/button-group";
    import * as InputGroup from "$lib/components/ui/input-group";

    let textareaWrapper: HTMLDivElement;
    let messages: Array<{ text: string; id: number }> = [];
    let messageId = 0;
    let messagesContainer: HTMLDivElement;
    let messagesEndRef: HTMLDivElement;
    let shouldScroll = false;

    function sendMessage() {
        const textarea = textareaWrapper?.querySelector('textarea');
        const text = textarea?.value.trim();
        if (!text) return;

        messages = [...messages, { text, id: messageId++ }];
        textarea.value = '';
        textarea.style.height = '44px'; // Reset to min-h-11 (2.75rem = 44px)
        shouldScroll = true;
    }

    function handleKeydown(event: KeyboardEvent) {
        if (event.key === 'Enter' && !event.shiftKey) {
            event.preventDefault();
            sendMessage();
        }
    }

    afterUpdate(() => {
        if (shouldScroll && messagesEndRef) {
            messagesEndRef.scrollIntoView({ behavior: 'smooth', block: 'end' });
            shouldScroll = false;
        }
    });

    onMount(() => {
        const textarea = textareaWrapper?.querySelector('textarea') as HTMLTextAreaElement;

        if (textarea) {
            // Set initial height
            textarea.style.height = '44px';
            textarea.style.overflowY = 'hidden';

            // Add input listener for dynamic resizing
            const handleInput = () => {
                textarea.style.height = '44px';
                const newHeight = Math.min(textarea.scrollHeight, 384);
                textarea.style.height = newHeight + 'px';
                textarea.style.overflowY = newHeight >= 384 ? 'auto' : 'hidden';
            };

            textarea.addEventListener('input', handleInput);
            textarea.addEventListener('keydown', handleKeydown);

            return () => {
                textarea.removeEventListener('input', handleInput);
                textarea.removeEventListener('keydown', handleKeydown);
            };
        }
    });
</script>

<Sidebar.Provider>
    <AppSidebar />
    <Sidebar.Inset>
        <header class="group-has-data-[collapsible=icon]/sidebar-wrapper:h-12 flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear">
            <div class="flex items-center gap-2 px-4">
                <Sidebar.Trigger class="-ms-1" />
                <Separator orientation="vertical" class="me-2 data-[orientation=vertical]:h-4" />
                <Breadcrumb.Root />
            </div>
        </header>

        <div class="flex flex-1 flex-col gap-4 p-4 pt-0 h-full">
            <div class="flex-1 flex items-start justify-center min-h-0">
                <div bind:this={messagesContainer} class="w-full max-w-6xl bg-muted/50 h-full rounded-xl md:min-h-min overflow-auto p-6 min-h-0 flex flex-col">
                    <!-- CHAT MESSAGES -->
                    {#if messages.length === 0}
                        <div class="flex items-center justify-center flex-1 text-muted-foreground">
                            <p>Start a conversation...</p>
                        </div>
                    {:else}
                        <div class="flex-1"></div>
                        <div class="flex flex-col gap-4">
                            {#each messages as message (message.id)}
                                <div class="flex justify-end">
                                    <div class="bg-primary text-primary-foreground rounded-2xl rounded-br-sm px-4 py-2 max-w-[80%] break-words whitespace-pre-wrap">
                                        {message.text}
                                    </div>
                                </div>
                            {/each}
                            <div bind:this={messagesEndRef}></div>
                        </div>
                    {/if}
                </div>
            </div>

            <div class="p-4 pt-0 sticky bottom-0 bg-background z-10">
                <div class="flex items-center justify-center w-full">
                    <div class="w-full max-w-6xl">
                        <ButtonGroup.Root class="[--radius:1rem] w-full flex items-center justify-center gap-3">
                            <ButtonGroup.Root class="shrink-0">
                                <Button variant="outline" size="icon" class="w-10 h-10">
                                    <Plus />
                                </Button>
                            </ButtonGroup.Root>
                            <ButtonGroup.Root class="flex-1">
                                <div bind:this={textareaWrapper} class="flex w-full items-center gap-2">
                                    <InputGroup.Root class="w-full">
                                        <InputGroup.Textarea
                                                class="w-full resize-none min-h-11"
                                                placeholder="Send a message..."
                                                rows={1}
                                        />
                                    </InputGroup.Root>
                                    <Button variant="default" size="icon" class="w-10 h-10" on:click={sendMessage}>
                                        <Send />
                                    </Button>
                                </div>
                            </ButtonGroup.Root>
                        </ButtonGroup.Root>
                    </div>
                </div>
            </div>
        </div>
    </Sidebar.Inset>
</Sidebar.Provider>