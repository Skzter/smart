<script lang="ts">
    import Prompt from "./components/Prompt.svelte";
    import Box from "./components/Box.svelte";
    import { Spinner } from "flowbite-svelte";
    import { getChatResponse, validatePrompt } from "./lib/Api.ts";
    import { onMount } from "svelte";
    import { auth } from "./lib/authService";
    import LoginButton from "./components/LoginButton.svelte";
    import LogoutButton from "./components/LogoutButton.svelte";

    let prompt = $state("");
    let convo = $state<{ id: number; role: string, content: string }[]>([]);

    // get ConversationId for api calls from cookies
    // Note: We should ideally get this from a backend state associated with the user
    var conversationId = $state("");

    onMount(() => {
        const media = window.matchMedia("(prefers-color-scheme: dark)");

        const updateTheme = () => {
            document.documentElement.setAttribute(
                "data-theme",
                media.matches ? "dark" : "light",
            );
        };

        media.addEventListener("change", updateTheme);
        updateTheme();
    });

    onMount(async () => {
        await auth.initAuth();
    });

    let userId = $state<string | undefined>(undefined);
    $effect(() => {
        if ($auth.isAuthenticated && $auth.user) {
            userId = $auth.user.sub;
        } else {
            userId = undefined;
        }
    });

    let paramsChatRequest = $derived({
        prompt: "",
        userId: userId,
        conversationId: conversationId,
    });
    const chatUrl = "chat";

    let isLoading = $state(false);
    
    async function onclick() {
        if (!userId) {
            console.error("User is not authenticated.");
            return;
        }

        const userQuestion = prompt.trim();
        if (!userQuestion) {
            console.warn("Prompt is empty, skipping request.");
            return;
        }

        prompt = "";

        // Push user message
        convo.push({
            role: "user",
            content: userQuestion,
        });

        isLoading = true;
        try {
            // === VALIDATE PROMPT ===
            const valResp = await validatePrompt({
                userId,
                conversationId,
                prompt: userQuestion,
            });

            conversationId = valResp.data.chatId ?? "";

            if (valResp.data.message?.body) {
                convo.push({
                    role: "system",
                    content: valResp.data.message.body,
                });
            }

            if (valResp.data.message?.body !== "Prompt validated successfully!") {
                isLoading = false;
                return;
            }

            // === CALL MAIN CHAT ===
            // keep old paramsChatRequest style
            paramsChatRequest.prompt = userQuestion;
            paramsChatRequest.userId = userId;
            paramsChatRequest.conversationId = conversationId;

            const answer = await getChatResponse(paramsChatRequest, chatUrl);

            if (answer.data.message?.body) {
                convo.push({
                    role: "assistant",
                    content: answer.data.message.body,
                });
            }

            conversationId = answer.data.conversationId ?? "";

        } catch (err) {
            let errorMsg = "An unexpected error occurred.";
            if (err.isAxiosError && err.response?.data?.message) {
                errorMsg = err.response.data.message;
            }
            convo.push({
                role: "system",
                content: errorMsg,
            });
        } finally {
            isLoading = false;
        }
    }




    let container: HTMLElement | undefined = $state();
    // Effect to trigger scrolling on relevant changes
    $effect(() => {
        if (container && (isLoading || convo.length > 0)) {
            // Small timeout to ensure DOM updates are complete
            setTimeout(() => {
                container.scrollTo({
                    top: container.scrollHeight,
                    behavior: "smooth",
                });
            }, 50);
        }
    });
</script>

{#if $auth.isAuthenticated}
    <div class="fixed left-4 top-4">
        <LogoutButton />
    </div>
    <main class="flex w-screen justify-center">
        <div
            class="flex flex-col h-[calc(100vh-132px)] overflow-y-auto w-8/10 gap-2 px-4 pt-4"
            bind:this={container}
        >
            {#each convo as c}
                {#if c.role === "user"}
                    <Box msg={c.content} name="User" {userId} {conversationId} />
                {:else if c.role === "assistant"}
                    <Box
                        msg={c.content}
                        name="Bot"
                        {userId}
                        {conversationId}
                        isCode={c.content.startsWith("import")}
                    />
                {:else if c.role === "system"}
                    <Box msg={c.content} name="System" {userId} {conversationId} />
                {/if}
            {/each}
            {#if isLoading}
                <div class="flex justify-start p-4">
                    <Spinner color="blue" />
                </div>
            {/if}
        </div>
    </main>
    <footer class="mt-[0px] mb-[-10px] pb-[5px] h-[100px]">
        <Prompt bind:input={prompt} {onclick} />
    </footer>
{:else}
    <main
        class="flex h-screen w-screen items-center justify-center bg-gray-100 dark:bg-gray-900"
    >
        <div
            class="w-full max-w-sm rounded-lg bg-white p-8 text-center shadow-xl dark:bg-gray-800"
        >
            <h1 class="mb-4 text-2xl font-bold text-gray-900 dark:text-white">
                Welcome to Project Autotester
            </h1>
            <p class="mb-8 text-gray-500 dark:text-gray-400">
                Please log in to continue
            </p>
            <LoginButton />
        </div>
    </main>
{/if}
