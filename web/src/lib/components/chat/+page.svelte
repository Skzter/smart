<script lang="ts">
    import { onMount } from 'svelte';
    import AppSidebar from "$lib/components/app-sidebar.svelte";
    import * as Breadcrumb from "$lib/components/ui/breadcrumb";
    import { Separator } from "$lib/components/ui/separator";
    import * as Sidebar from "$lib/components/ui/sidebar";
    import Send from "@lucide/svelte/icons/send";
    import Plus from "@lucide/svelte/icons/plus";
    import { Button } from "$lib/components/ui/button";
    import * as ButtonGroup from "$lib/components/ui/button-group";
    import * as InputGroup from "$lib/components/ui/input-group";
    import { getChatResponse } from "$lib/Api";
    import { auth } from "$lib/authService";
    import LogIn from "@lucide/svelte/icons/log-in";
    import LogOut from "@lucide/svelte/icons/log-out";

    let textareaWrapper: HTMLDivElement;
    let messages = $state<Array<{ question: string; answer: string; id: number }>>([]);
    let messageId = 0;
    let messagesContainer: HTMLDivElement;
    let messagesEndRef: HTMLDivElement;
    let isLoading = $state(false);

    // Auth state
    let userId = $state<string | undefined>(undefined);

    // Conversation ID from localStorage
    let conversationId = $state(localStorage.getItem("conversationId") || "");

    // Initialize auth
    onMount(async () => {
        await auth.initAuth();

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

    // Watch for auth changes
    $effect(() => {
        if ($auth.isAuthenticated && $auth.user) {
            userId = $auth.user.sub;
        } else {
            userId = undefined;
        }
    });

    // Auth handlers
    function handleLogin() {
        auth.login();
    }

    function handleLogout() {
        auth.logout();
    }

    const chatUrl = "/chat";

    async function sendMessage() {
        if (!userId) {
            console.error("User is not authenticated.");
            return;
        }

        const textarea = textareaWrapper?.querySelector('textarea');
        const text = textarea?.value.trim();
        if (!text) return;

        // Add user message
        const userQuestion = text;
        messages = [...messages, { question: userQuestion, answer: '', id: messageId++ }];

        // Clear and reset textarea
        textarea.value = '';
        textarea.style.height = '44px';

        isLoading = true;

        // Sanitize userId (remove auth provider prefix if present)
        const sanitizedUserId = userId.includes("|")
            ? userId.split("|")[1]
            : userId;

        const paramsChatRequest = {
            message: { data: userQuestion, agent: "user" },
            userId: userId,
            conversationId: conversationId || "",
        };

        console.log('Sending request:', JSON.stringify(paramsChatRequest, null, 2));
        console.log('Original userId:', userId);
        console.log('Sanitized userId:', sanitizedUserId);
        console.log('conversationId:', conversationId);

        try {
            const answer = await getChatResponse(paramsChatRequest, chatUrl);
            console.log('Received answer:', answer);
            messages[messages.length - 1].answer = answer.data.message.data;
        } catch (err: any) {
            if (err.isAxiosError) {
                messages[messages.length - 1].answer = err.response.data.message;
            } else {
                messages[messages.length - 1].answer = "An error occurred while sending your message.";
            }
        } finally {
            isLoading = false;
        }
    }

    function handleKeydown(event: KeyboardEvent) {
        if (event.key === 'Enter' && !event.shiftKey) {
            event.preventDefault();
            sendMessage();
        }
    }

    // Auto-scroll effect
    $effect(() => {
        if (messagesEndRef && (messages.length > 0 || isLoading)) {
            setTimeout(() => {
                messagesEndRef?.scrollIntoView({ behavior: 'smooth', block: 'end' });
            }, 50);
        }
    });
</script>

{#if $auth.isAuthenticated}
    <Sidebar.Provider>
        <AppSidebar />
        <Sidebar.Inset>
            <header class="group-has-data-[collapsible=icon]/sidebar-wrapper:h-12 flex h-16 shrink-0 items-center gap-2 transition-[width,height] ease-linear">
                <div class="flex items-center gap-2 px-4 flex-1">
                    <Sidebar.Trigger class="-ms-1" />
                    <Separator orientation="vertical" class="me-2 data-[orientation=vertical]:h-4" />
                    <Breadcrumb.Root />
                </div>
                <div class="px-4">
                    <Button variant="outline" size="sm" on:click={handleLogout}>
                        <LogOut class="h-4 w-4 mr-2" />
                        Logout
                    </Button>
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
                                    <!-- User message -->
                                    <div class="flex justify-end">
                                        <div class="bg-primary text-primary-foreground rounded-2xl rounded-br-sm px-4 py-2 max-w-[80%] break-words whitespace-pre-wrap">
                                            {message.question}
                                        </div>
                                    </div>

                                    <!-- Bot response -->
                                    {#if message.answer}
                                        <div class="flex justify-start">
                                            <div class="bg-muted text-foreground rounded-2xl rounded-bl-sm px-4 py-2 max-w-[80%] break-words whitespace-pre-wrap">
                                                {message.answer}
                                            </div>
                                        </div>
                                    {/if}
                                {/each}

                                <!-- Loading indicator -->
                                {#if isLoading}
                                    <div class="flex justify-start">
                                        <div class="bg-muted text-foreground rounded-2xl rounded-bl-sm px-4 py-2">
                                            <div class="flex gap-1">
                                                <span class="animate-bounce" style="animation-delay: 0ms;">●</span>
                                                <span class="animate-bounce" style="animation-delay: 150ms;">●</span>
                                                <span class="animate-bounce" style="animation-delay: 300ms;">●</span>
                                            </div>
                                        </div>
                                    </div>
                                {/if}

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
                                                    disabled={isLoading}
                                            />
                                        </InputGroup.Root>
                                        <Button
                                                variant="default"
                                                size="icon"
                                                class="w-10 h-10"
                                                on:click={sendMessage}
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
            </div>
        </Sidebar.Inset>
    </Sidebar.Provider>
{:else}
    <main class="flex h-screen w-screen items-center justify-center bg-gray-100 dark:bg-gray-900">
        <div class="w-full max-w-sm rounded-lg bg-white p-8 text-center shadow-xl dark:bg-gray-800">
            <h1 class="mb-4 text-2xl font-bold text-gray-900 dark:text-white">
                Welcome to Project Autotester
            </h1>
            <p class="mb-8 text-gray-500 dark:text-gray-400">
                Please log in to continue
            </p>
            <Button on:click={handleLogin} size="lg" class="w-full">
                <LogIn class="h-5 w-5 mr-2" />
                Log In
            </Button>
        </div>
    </main>
{/if}